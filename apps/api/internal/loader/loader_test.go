package loader_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/config"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
	"github.com/Qurioth/chara-q/apps/api/internal/graph"
	"github.com/Qurioth/chara-q/apps/api/internal/server"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

// userQueryCounter は、所有者をまとめて引く SQL が実行された回数を数える pgx の QueryTracer。
type userQueryCounter struct{ n atomic.Int32 }

func (c *userQueryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if strings.Contains(data.SQL, "FROM users WHERE id = ANY") {
		c.n.Add(1)
	}
	return ctx
}

func (c *userQueryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

// TestOwnerIsBatched は、一覧の件数を増やしても、所有者の問い合わせが 1 回のままであることを確かめる（SC-003）。
func TestOwnerIsBatched(t *testing.T) {
	for _, n := range []int{1, 20} {
		t.Run(fmt.Sprintf("%d 件", n), func(t *testing.T) {
			env := authtest.New(t)
			// 所有者がすべて違うシートを n 件作る
			for i := range n {
				u := env.CreateUser(t, fmt.Sprintf("user%02d", i), "password-1234", false)
				env.Clock.Advance(time.Second)
				sheettest.CreateSheet(t, env, u.ID, sheet.SystemYuyakekoyake, fmt.Sprintf("シート %02d", i), "")
			}

			counter := &userQueryCounter{}
			cfg, err := pgxpool.ParseConfig(os.Getenv("TEST_DATABASE_URL"))
			if err != nil {
				t.Fatalf("ParseConfig: %v", err)
			}
			cfg.ConnConfig.Tracer = counter
			pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
			if err != nil {
				t.Fatalf("NewWithConfig: %v", err)
			}
			t.Cleanup(pool.Close)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			h := server.New(config.Config{}, &graph.Resolver{
				Queries: db.New(pool),
				Auth:    auth.NewService(pool, logger),
				Sheets:  sheet.NewService(pool, logger),
			}, logger)

			body, _ := json.Marshal(map[string]any{
				"query": `{ characterSheets { nodes { owner { displayName } } } }`,
			})
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/graphql", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			var resp struct {
				Data struct {
					CharacterSheets struct {
						Nodes []struct {
							Owner struct {
								DisplayName string `json:"displayName"`
							} `json:"owner"`
						} `json:"nodes"`
					} `json:"characterSheets"`
				} `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil || len(resp.Data.CharacterSheets.Nodes) != n {
				t.Fatalf("一覧が読めない: %s", rec.Body.String())
			}
			for _, node := range resp.Data.CharacterSheets.Nodes {
				if node.Owner.DisplayName == "" {
					t.Fatalf("所有者の表示名が空: %s", rec.Body.String())
				}
			}
			if got := counter.n.Load(); got != 1 {
				t.Fatalf("所有者の問い合わせは 1 回のはず: %d 回", got)
			}
		})
	}
}
