// Package authtest は、DB を使う認証のテストの補助を置く。
//
// テストのたびに認証の表を空にするため、DB を使うテストでは t.Parallel() を使わず、
// パッケージも並列に実行しない（go test -p 1。Taskfile の test:api と CI で指定している）。
package authtest

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// Clock はテストで進められる時計。
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

// Now は現在の時刻を返す。
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance は時計を進める。
func (c *Clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// Env は DB を使うテストの環境。
type Env struct {
	Service *auth.Service
	Pool    *pgxpool.Pool
	Queries *db.Queries
	Clock   *Clock
}

// New は、認証の表を空にした DB に接続した Service を返す。
// TEST_DATABASE_URL が未設定ならテストをスキップする（CI では必ず設定する）。
func New(t *testing.T) *Env {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL が設定されていません")
	}
	if url == os.Getenv("DATABASE_URL") {
		t.Fatal("TEST_DATABASE_URL が開発用の DATABASE_URL と同じです。テストは表を空にするため、別のデータベースにしてください（task db:test:prepare）")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("DB の接続設定が不正です: %v", err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, "TRUNCATE users, sessions, invitations, password_reset_codes, auth_failures CASCADE"); err != nil {
		t.Fatalf("表を空にできません（task db:test:prepare を実行しましたか）: %v", err)
	}

	// DB の時刻の精度（マイクロ秒）にそろえる
	clock := &Clock{now: time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)}
	svc := auth.NewService(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	svc.SetClock(clock.Now)
	return &Env{Service: svc, Pool: pool, Queries: db.New(pool), Clock: clock}
}

// CreateUser はテスト用のユーザーを作る。
func (e *Env) CreateUser(t *testing.T, loginID, password string, isAdmin bool) db.User {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	u, err := e.Queries.CreateUser(context.Background(), db.CreateUserParams{
		LoginID:      loginID,
		DisplayName:  "表示名 " + loginID,
		PasswordHash: hash,
		IsAdmin:      isAdmin,
		Now:          e.Clock.Now(),
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return u
}

// Count は表の行数を返す。
func (e *Env) Count(t *testing.T, table string) int {
	t.Helper()
	var n int
	// table はテストのコードが渡す固定の名前だけ
	if err := e.Pool.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}
