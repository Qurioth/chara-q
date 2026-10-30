package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// TEST_DATABASE_URL が未設定のときはスキップする（CI では必ず設定する）。
func newTestQueries(t *testing.T) *db.Queries {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL が設定されていません")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("DB の接続設定が不正です: %v", err)
	}
	t.Cleanup(pool.Close)
	return db.New(pool)
}

func TestPing(t *testing.T) {
	q := newTestQueries(t)
	got, err := q.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}
