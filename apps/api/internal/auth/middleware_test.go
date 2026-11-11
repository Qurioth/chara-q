package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// serve はミドルウェアを通したリクエストを送り、ハンドラーが見た Viewer と応答を返す。
func serve(env *authtest.Env, token string) (*auth.Viewer, *httptest.ResponseRecorder) {
	var got *auth.Viewer
	h := env.Service.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = auth.ViewerFrom(r.Context())
	}))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/graphql", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return got, rec
}

func sessionExpiresAt(t *testing.T, env *authtest.Env, token string) time.Time {
	t.Helper()
	var exp time.Time
	if err := env.Pool.QueryRow(context.Background(), "SELECT expires_at FROM sessions WHERE token_hash = $1", auth.HashToken(token)).Scan(&exp); err != nil {
		t.Fatalf("expires_at: %v", err)
	}
	return exp
}

func TestMiddleware(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	sess, err := env.Service.Login(context.Background(), "alice", "password1", nil)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	t.Run("Cookie なしならログインしていない", func(t *testing.T) {
		v, rec := serve(env, "")
		if v != nil {
			t.Error("Viewer はないはず")
		}
		if rec.Header().Get("Set-Cookie") != "" {
			t.Error("Cookie を書かないはず")
		}
	})

	t.Run("有効なセッションなら本人", func(t *testing.T) {
		v, _ := serve(env, sess.Token)
		if v == nil || v.LoginID != "alice" || v.DisplayName != "表示名 alice" || v.IsAdmin {
			t.Errorf("got %+v", v)
		}
	})

	t.Run("存在しないセッションなら Cookie を消す", func(t *testing.T) {
		v, rec := serve(env, "NOSUCHTOKEN")
		if v != nil {
			t.Error("Viewer はないはず")
		}
		if !strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0") {
			t.Errorf("Cookie を消すはず: %q", rec.Header().Get("Set-Cookie"))
		}
	})

	t.Run("1 時間以内なら延長しない、1 時間を過ぎたら延長する", func(t *testing.T) {
		before := sessionExpiresAt(t, env, sess.Token)
		env.Clock.Advance(59 * time.Minute)
		serve(env, sess.Token)
		if got := sessionExpiresAt(t, env, sess.Token); !got.Equal(before) {
			t.Errorf("延長しないはず: %v → %v", before, got)
		}
		env.Clock.Advance(2 * time.Minute)
		serve(env, sess.Token)
		want := env.Clock.Now().Add(auth.SessionTTL)
		if got := sessionExpiresAt(t, env, sess.Token); !got.Equal(want) {
			t.Errorf("延長するはず: got %v, want %v", got, want)
		}
	})

	t.Run("期限切れならログインしていない", func(t *testing.T) {
		env.Clock.Advance(auth.SessionTTL + time.Second)
		v, rec := serve(env, sess.Token)
		if v != nil {
			t.Error("期限切れのセッションは使えない")
		}
		if !strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0") {
			t.Error("Cookie を消すはず")
		}
	})
}

func TestMiddlewareDisabledUser(t *testing.T) {
	env := authtest.New(t)
	u := env.CreateUser(t, "bob", "password1", false)
	sess, err := env.Service.Login(context.Background(), "bob", "password1", nil)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	// セッションを残したまま無効にしても、ユーザーが無効なら使えない
	if _, err := env.Queries.DisableUser(context.Background(), db.DisableUserParams{Now: env.Clock.Now(), ID: u.ID}); err != nil {
		t.Fatal(err)
	}
	if v, _ := serve(env, sess.Token); v != nil {
		t.Error("無効なユーザーのセッションは使えない")
	}
}
