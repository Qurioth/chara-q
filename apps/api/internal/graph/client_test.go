package graph_test

// GraphQL のテストで共通に使うクライアント（002-auth と 003-character-sheet-base のテスト）。

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/config"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
	"github.com/Qurioth/chara-q/apps/api/internal/graph"
	"github.com/Qurioth/chara-q/apps/api/internal/server"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
)

type client struct {
	t     *testing.T
	h     http.Handler
	token string
}

type gqlResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message    string         `json:"message"`
		Extensions map[string]any `json:"extensions"`
	} `json:"errors"`
	setCookie string
	raw       string
}

func newClient(t *testing.T, env *authtest.Env) *client {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sheets := sheet.NewService(env.Pool, logger)
	sheets.SetClock(env.Clock.Now)
	h := server.New(config.Config{}, &graph.Resolver{Queries: env.Queries, Auth: env.Service, Sheets: sheets}, logger)
	return &client{t: t, h: h}
}

// loginAs はユーザーを作り、そのユーザーでログインしたクライアントを返す。
// 応答にログイン ID が含まれないことを確かめられるよう、表示名はログイン ID を含まないものにする
// （authtest.CreateUser の表示名は「表示名 <ログイン ID>」）。
func loginAs(t *testing.T, env *authtest.Env, loginID string, isAdmin bool) (*client, db.User) {
	t.Helper()
	const password = "password-1234"
	created := env.CreateUser(t, loginID, password, isAdmin)
	u, err := env.Queries.UpdateDisplayName(context.Background(), db.UpdateDisplayNameParams{
		ID:          created.ID,
		DisplayName: "ユーザー " + created.ID.String()[len(created.ID.String())-6:],
		Now:         env.Clock.Now(),
	})
	if err != nil {
		t.Fatalf("UpdateDisplayName: %v", err)
	}
	c := newClient(t, env)
	resp := c.do(loginMutation, map[string]any{"loginId": loginID, "password": password})
	if c.token == "" {
		t.Fatalf("ログインできません: %s", resp.raw)
	}
	return c, u
}

// do は /graphql に POST する。応答の Set-Cookie のセッション ID を、次のリクエストの Cookie にする
// （httptest は https ではないため、Secure の Cookie を手で受け渡す）。
func (c *client) do(query string, vars map[string]any) gqlResponse {
	c.t.Helper()
	body, _ := json.Marshal(map[string]any{"query": query, "variables": vars})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/graphql", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: c.token})
	}
	rec := httptest.NewRecorder()
	c.h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		c.t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var resp gqlResponse
	resp.raw = rec.Body.String()
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		c.t.Fatalf("json: %v", err)
	}
	resp.setCookie = rec.Header().Get("Set-Cookie")
	for _, sc := range rec.Result().Cookies() {
		if sc.Name == auth.SessionCookieName {
			c.token = sc.Value
		}
	}
	return resp
}

func (r gqlResponse) errorCode() string {
	if len(r.Errors) == 0 {
		return ""
	}
	code, _ := r.Errors[0].Extensions["code"].(string)
	return code
}
