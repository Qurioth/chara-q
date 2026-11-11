// Package sheet はキャラクターシートの共通の処理（作成、閲覧、一覧、更新、複製、削除）を扱う
// （specs/003-character-sheet-base）。所有者の確認は、画面ではなくここと SQL の条件で行う（ADR 0005）。
package sheet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

var (
	// ErrNotFound は、シートが存在しない・削除済み・形の正しくない ID のときのエラー。
	ErrNotFound = errors.New("シートが見つかりません")
	// ErrForbidden は、所有者でないユーザーが更新・削除しようとしたときのエラー。
	ErrForbidden = errors.New("この操作は行えません")
	// ErrRevisionConflict は、読み込んだ後に別の場所で更新されていたときのエラー。
	// resolver は payload の errors（REVISION_CONFLICT）として返す。
	ErrRevisionConflict = auth.InputErrors{{Field: auth.FieldNone, Code: CodeRevisionConflict}}
)

// Service はシートの処理をまとめる。
type Service struct {
	q      *db.Queries
	logger *slog.Logger
	now    func() time.Time
}

// NewService は Service を作る。
func NewService(pool *pgxpool.Pool, logger *slog.Logger) *Service {
	return &Service{q: db.New(pool), logger: logger, now: time.Now}
}

// SetClock は現在時刻の関数を差し替える（テスト用）。
func (s *Service) SetClock(now func() time.Time) {
	s.now = now
}

// parseID は URL や入力の ID を UUID にする。形が正しくなければ ErrNotFound（research R5）。
func parseID(id string) (uuid.UUID, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, ErrNotFound
	}
	return u, nil
}

// stateError は、更新・削除が 0 行だったときに理由を判定する。
// revisionChecked が false（削除）なら、所有者が合っていても ErrNotFound にする
// （削除の直前に、別の場所で削除されたときだけ起きる）。
func (s *Service) stateError(ctx context.Context, id, owner uuid.UUID, revisionChecked bool) error {
	st, err := s.q.GetCharacterSheetState(ctx, id)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNotFound
	case err != nil:
		return fmt.Errorf("シートの状態を読めません: %w", err)
	case st.DeletedAt != nil:
		return ErrNotFound
	case st.OwnerID != owner:
		return ErrForbidden
	case !revisionChecked:
		return ErrNotFound
	}
	return ErrRevisionConflict
}

// toInt32 は int を int32 に収める。範囲の外なら ok は false。
func toInt32(n int) (int32, bool) {
	if n < math.MinInt32 || n > math.MaxInt32 {
		return 0, false
	}
	return int32(n), true
}
