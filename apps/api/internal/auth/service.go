// Package auth は認証（登録、ログイン、セッション、ロック、パスワードの再設定、ユーザーの無効化）を扱う。
// GraphQL の resolver と管理者のコマンドの両方から使う（specs/002-auth）。
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// 期限と回数（Clarifications Q2）。
const (
	SessionTTL           = 30 * 24 * time.Hour
	sessionTouchInterval = time.Hour
	InvitationTTL        = 7 * 24 * time.Hour
	ResetCodeTTL         = 24 * time.Hour
	MaxFailures          = 5
	LockDuration         = 15 * time.Minute
	staleFailureAge      = 24 * time.Hour
)

// 失敗を数える目的（auth_failures.purpose）。
const (
	purposeLogin         = "login"
	purposePasswordReset = "password_reset"
)

var (
	// ErrForbidden は管理者だけの操作を、管理者でないユーザーが行おうとしたときのエラー。
	ErrForbidden = errors.New("管理者だけが行えます")
	// ErrUserNotFound は指定したログイン ID のユーザーがいないときのエラー。
	ErrUserNotFound = errors.New("ユーザーが見つかりません")
	// ErrUserDisabled は無効なユーザーに対して行えない操作のエラー。
	ErrUserDisabled = errors.New("ユーザーが無効です")
)

// LockedError は、続けて失敗したため一時的にロックしているときのエラー。
type LockedError struct {
	Until time.Time
}

func (e *LockedError) Error() string {
	return "続けて失敗したため " + e.Until.Format(time.RFC3339) + " までロックしています"
}

// Session は発行したセッション。Token はブラウザの Cookie に入れる値で、DB には保存しない。
type Session struct {
	Token  string
	Viewer Viewer
}

// Service は認証の処理をまとめる。
type Service struct {
	pool   *pgxpool.Pool
	q      *db.Queries
	logger *slog.Logger
	now    func() time.Time
}

// NewService は Service を作る。
func NewService(pool *pgxpool.Pool, logger *slog.Logger) *Service {
	return &Service{pool: pool, q: db.New(pool), logger: logger, now: time.Now}
}

// SetClock は現在時刻の関数を差し替える（テスト用）。
func (s *Service) SetClock(now func() time.Time) {
	s.now = now
}

func (s *Service) withTx(ctx context.Context, fn func(q *db.Queries) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return fn(s.q.WithTx(tx))
	})
}

func viewerOf(u db.User) Viewer {
	return Viewer{UserID: u.ID, LoginID: u.LoginID, DisplayName: u.DisplayName, IsAdmin: u.IsAdmin}
}

// createSession は新しいセッションを作り、ブラウザに渡すトークンを返す。
func createSession(ctx context.Context, q *db.Queries, user db.User, now time.Time) (Session, error) {
	token := NewToken()
	if err := q.CreateSession(ctx, db.CreateSessionParams{
		TokenHash: HashToken(token),
		UserID:    user.ID,
		Now:       now,
		ExpiresAt: now.Add(SessionTTL),
	}); err != nil {
		return Session{}, fmt.Errorf("セッションを作れません: %w", err)
	}
	return Session{Token: token, Viewer: viewerOf(user)}, nil
}

// isLoginIDTaken は、ログイン ID の一意の制約の違反かを返す。
func isLoginIDTaken(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_login_id_key"
}

// findUser はログイン ID でユーザーを引く。見つからなければ ErrUserNotFound。
func (s *Service) findUser(ctx context.Context, loginID string) (db.User, error) {
	u, err := s.q.GetUserByLoginID(ctx, NormalizeLoginID(loginID))
	if errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, ErrUserNotFound
	}
	if err != nil {
		return db.User{}, fmt.Errorf("ユーザーを読めません: %w", err)
	}
	return u, nil
}
