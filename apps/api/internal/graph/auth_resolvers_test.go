package graph_test

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
	"github.com/Qurioth/chara-q/apps/api/internal/graph"
	"github.com/Qurioth/chara-q/apps/api/internal/server"
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
	h := server.New(config.Config{}, &graph.Resolver{Queries: env.Queries, Auth: env.Service}, logger)
	return &client{t: t, h: h}
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

const viewerQuery = `{ viewer { id loginId displayName isAdmin } }`

const loginMutation = `mutation($loginId: String!, $password: String!) {
  login(input: { loginId: $loginId, password: $password }) {
    viewer { loginId displayName isAdmin }
    errors { field code }
    lockedUntil
  }
}`

func TestLoginLogoutViewer(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	c := newClient(t, env)

	if got := string(c.do(viewerQuery, nil).Data["viewer"]); got != "null" {
		t.Errorf("ログインしていなければ null: %s", got)
	}

	resp := c.do(loginMutation, map[string]any{"loginId": "alice", "password": "wrong-password"})
	if !strings.Contains(resp.raw, `"errors":[{"field":null,"code":"INVALID_CREDENTIALS"}]`) {
		t.Errorf("INVALID_CREDENTIALS: %s", resp.raw)
	}

	resp = c.do(loginMutation, map[string]any{"loginId": "alice", "password": "password1"})
	for _, want := range []string{"__Host-chara_q_session=", "Path=/", "HttpOnly", "Secure", "SameSite=Lax"} {
		if !strings.Contains(resp.setCookie, want) {
			t.Errorf("Set-Cookie に %q がない: %s", want, resp.setCookie)
		}
	}
	if strings.Contains(strings.ToLower(resp.setCookie), "domain=") {
		t.Errorf("Domain を付けない: %s", resp.setCookie)
	}

	resp = c.do(viewerQuery, nil)
	if !strings.Contains(resp.raw, `"loginId":"alice","displayName":"表示名 alice","isAdmin":false`) {
		t.Errorf("本人の情報: %s", resp.raw)
	}

	token := c.token
	resp = c.do(`mutation { logout }`, nil)
	if !strings.Contains(resp.setCookie, "Max-Age=0") {
		t.Errorf("Cookie を消す: %s", resp.setCookie)
	}
	c.token = token // ログアウト前の Cookie を手で戻しても使えない
	if got := string(c.do(viewerQuery, nil).Data["viewer"]); got != "null" {
		t.Errorf("ログアウトしたセッションは使えない: %s", got)
	}
}

func TestLoginLocked(t *testing.T) {
	env := authtest.New(t)
	c := newClient(t, env)
	var resp gqlResponse
	for range 5 {
		resp = c.do(loginMutation, map[string]any{"loginId": "nobody", "password": "wrong-password"})
	}
	if !strings.Contains(resp.raw, `"code":"LOCKED"`) || strings.Contains(resp.raw, `"lockedUntil":null`) {
		t.Errorf("LOCKED と lockedUntil: %s", resp.raw)
	}
}

const registerMutation = `mutation($code: String!, $loginId: String!) {
  register(input: { invitationCode: $code, loginId: $loginId, password: "password1", displayName: "たま" }) {
    viewer { loginId displayName }
    errors { field code }
  }
}`

func TestRegisterAndIssueInvitation(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "admin", "password1", true)
	env.CreateUser(t, "member", "password1", false)

	anonymous := newClient(t, env)
	if code := anonymous.do(`mutation { issueInvitation { code } }`, nil).errorCode(); code != "UNAUTHENTICATED" {
		t.Errorf("ログインしていなければ UNAUTHENTICATED: %s", code)
	}

	member := newClient(t, env)
	member.do(loginMutation, map[string]any{"loginId": "member", "password": "password1"})
	if code := member.do(`mutation { issueInvitation { code } }`, nil).errorCode(); code != "FORBIDDEN" {
		t.Errorf("管理者でなければ FORBIDDEN: %s", code)
	}

	admin := newClient(t, env)
	admin.do(loginMutation, map[string]any{"loginId": "admin", "password": "password1"})
	resp := admin.do(`mutation { issueInvitation { code expiresAt } }`, nil)
	var issued struct {
		Code      string `json:"code"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := json.Unmarshal(resp.Data["issueInvitation"], &issued); err != nil || issued.Code == "" || issued.ExpiresAt == "" {
		t.Fatalf("招待コード: %s", resp.raw)
	}

	newcomer := newClient(t, env)
	resp = newcomer.do(registerMutation, map[string]any{"code": issued.Code, "loginId": "tama"})
	if !strings.Contains(resp.raw, `"viewer":{"loginId":"tama","displayName":"たま"}`) || !strings.Contains(resp.setCookie, auth.SessionCookieName) {
		t.Errorf("登録してログインした状態になる: %s", resp.raw)
	}
	if !strings.Contains(newcomer.do(viewerQuery, nil).raw, `"loginId":"tama"`) {
		t.Error("登録した人の viewer")
	}

	resp = newClient(t, env).do(registerMutation, map[string]any{"code": issued.Code, "loginId": "tama2"})
	if !strings.Contains(resp.raw, `{"field":"INVITATION_CODE","code":"INVITATION_UNAVAILABLE"}`) {
		t.Errorf("使用済み: %s", resp.raw)
	}

	// 無効にした管理者のセッションでは発行できない
	if _, _, err := env.Service.DisableUser(context.Background(), "admin"); err != nil {
		t.Fatal(err)
	}
	if code := admin.do(`mutation { issueInvitation { code } }`, nil).errorCode(); code != "UNAUTHENTICATED" {
		t.Errorf("無効にした管理者は発行できない: %s", code)
	}
}

func TestAccountMutations(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	env.CreateUser(t, "bob", "password1", false)

	anonymous := newClient(t, env)
	for _, q := range []string{
		`mutation { updateDisplayName(input: { displayName: "x" }) { errors { code } } }`,
		`mutation { changePassword(input: { currentPassword: "password1", newPassword: "password2" }) { errors { code } } }`,
	} {
		if code := anonymous.do(q, nil).errorCode(); code != "UNAUTHENTICATED" {
			t.Errorf("ログインしていなければ UNAUTHENTICATED: %s", code)
		}
	}

	a := newClient(t, env)
	a.do(loginMutation, map[string]any{"loginId": "alice", "password": "password1"})
	other := newClient(t, env)
	other.do(loginMutation, map[string]any{"loginId": "alice", "password": "password1"})

	resp := a.do(`mutation { updateDisplayName(input: { displayName: "ありす" }) { viewer { displayName } errors { field code } } }`, nil)
	if !strings.Contains(resp.raw, `"displayName":"ありす"`) {
		t.Errorf("表示名を変える: %s", resp.raw)
	}

	before := a.token
	resp = a.do(`mutation { changePassword(input: { currentPassword: "password1", newPassword: "password2" }) { viewer { loginId } errors { field code } } }`, nil)
	if a.token == before || !strings.Contains(resp.setCookie, auth.SessionCookieName) {
		t.Errorf("新しいセッションの Cookie を付ける: %s", resp.setCookie)
	}
	if !strings.Contains(a.do(viewerQuery, nil).raw, `"loginId":"alice"`) {
		t.Error("変更したブラウザはログインしたまま")
	}
	if got := string(other.do(viewerQuery, nil).Data["viewer"]); got != "null" {
		t.Errorf("ほかのブラウザはログインしていない扱い: %s", got)
	}

	// 他のユーザーのログイン ID は、どの応答にも含まれない（SC-008）
	bob := newClient(t, env)
	bob.do(loginMutation, map[string]any{"loginId": "bob", "password": "password1"})
	if raw := bob.do(viewerQuery, nil).raw; strings.Contains(raw, "alice") {
		t.Errorf("他のユーザーのログイン ID が含まれる: %s", raw)
	}
}

func TestResetPasswordMutation(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	code, _, err := env.Service.IssueResetCode(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	c := newClient(t, env)
	const q = `mutation($code: String!) {
  resetPassword(input: { loginId: "alice", resetCode: $code, newPassword: "password2" }) { ok errors { field code } lockedUntil }
}`
	resp := c.do(q, map[string]any{"code": code})
	if !strings.Contains(resp.raw, `"ok":true`) || resp.setCookie != "" {
		t.Errorf("再設定し、ログインはさせない: %s / %s", resp.raw, resp.setCookie)
	}
	resp = c.do(q, map[string]any{"code": code})
	if !strings.Contains(resp.raw, `{"field":"RESET_CODE","code":"RESET_CODE_UNAVAILABLE"}`) {
		t.Errorf("2 回目は使えない: %s", resp.raw)
	}
}
