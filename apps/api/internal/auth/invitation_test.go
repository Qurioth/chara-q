package auth_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func issue(t *testing.T, env *authtest.Env, admin auth.Viewer) string {
	t.Helper()
	code, _, err := env.Service.IssueInvitation(context.Background(), admin)
	if err != nil {
		t.Fatalf("IssueInvitation: %v", err)
	}
	return code
}

func adminViewer(t *testing.T, env *authtest.Env) auth.Viewer {
	t.Helper()
	u := env.CreateUser(t, "admin", "password1", true)
	return auth.Viewer{UserID: u.ID, LoginID: u.LoginID, DisplayName: u.DisplayName, IsAdmin: true}
}

func registerInput(code, loginID string) auth.RegisterInput {
	return auth.RegisterInput{InvitationCode: code, LoginID: loginID, Password: "password1", DisplayName: "たま"}
}

func TestIssueInvitation(t *testing.T) {
	env := authtest.New(t)
	admin := adminViewer(t, env)
	ctx := context.Background()

	code, expiresAt, err := env.Service.IssueInvitation(ctx, admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(auth.NormalizeCode(code)) != 26 || strings.Count(code, "-") != 6 {
		t.Errorf("4 文字ごとにハイフンで区切った 26 文字のはず: %q", code)
	}
	if want := env.Clock.Now().Add(7 * 24 * time.Hour); !expiresAt.Equal(want) {
		t.Errorf("期限は 7 日後: got %v, want %v", expiresAt, want)
	}

	// DB にはハッシュだけがある（SC-007）
	var hash []byte
	var row string
	if err := env.Pool.QueryRow(ctx, "SELECT code_hash, i::text FROM invitations i").Scan(&hash, &row); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(hash, auth.HashToken(auth.NormalizeCode(code))) {
		t.Error("code_hash は正規化したコードの SHA-256")
	}
	if strings.Contains(row, auth.NormalizeCode(code)) {
		t.Error("平文のコードを保存してはいけない")
	}

	if _, _, err := env.Service.IssueInvitation(ctx, auth.Viewer{UserID: admin.UserID}); !errors.Is(err, auth.ErrForbidden) {
		t.Errorf("管理者でなければ発行できない: %v", err)
	}
}

func TestRegister(t *testing.T) {
	env := authtest.New(t)
	admin := adminViewer(t, env)
	ctx := context.Background()

	t.Run("成功してログインした状態になる", func(t *testing.T) {
		code := issue(t, env, admin)
		// 小文字・空白入りで入力しても使える
		in := registerInput(" "+strings.ToLower(code)+" ", "tama")
		sess, err := env.Service.Register(ctx, in)
		if err != nil {
			t.Fatalf("Register: %v", err)
		}
		if sess.Viewer.LoginID != "tama" || sess.Viewer.DisplayName != "たま" || sess.Viewer.IsAdmin {
			t.Errorf("got %+v", sess.Viewer)
		}
		if v, _ := serve(env, sess.Token); v == nil {
			t.Error("セッションが使えるはず")
		}

		_, err = env.Service.Register(ctx, registerInput(code, "tama2"))
		wantInputError(t, err, auth.FieldInvitationCode, auth.CodeInvitationUnavailable)
	})

	t.Run("期限切れ・存在しないコードは同じ失敗", func(t *testing.T) {
		code := issue(t, env, admin)
		env.Clock.Advance(7*24*time.Hour + time.Second)
		_, err := env.Service.Register(ctx, registerInput(code, "late"))
		wantInputError(t, err, auth.FieldInvitationCode, auth.CodeInvitationUnavailable)

		_, err = env.Service.Register(ctx, registerInput("NOSUCHCODE", "late"))
		wantInputError(t, err, auth.FieldInvitationCode, auth.CodeInvitationUnavailable)
	})

	t.Run("使えないコードなら、ログイン ID が使われていても答えない", func(t *testing.T) {
		_, err := env.Service.Register(ctx, registerInput("NOSUCHCODE", "admin"))
		errs := inputCodes(t, err)
		if len(errs) != 1 || errs[0].Code != auth.CodeInvitationUnavailable {
			t.Errorf("INVITATION_UNAVAILABLE だけのはず: %v", errs)
		}
	})

	t.Run("ログイン ID が使われていればコードは未使用のまま", func(t *testing.T) {
		code := issue(t, env, admin)
		_, err := env.Service.Register(ctx, registerInput(code, "ADMIN"))
		wantInputError(t, err, auth.FieldLoginID, auth.CodeLoginIDTaken)
		if _, err := env.Service.Register(ctx, registerInput(code, "fresh")); err != nil {
			t.Errorf("コードはまだ使えるはず（FR-014）: %v", err)
		}
	})

	t.Run("入力の規則に合わなければコードは未使用のまま", func(t *testing.T) {
		code := issue(t, env, admin)
		_, err := env.Service.Register(ctx, auth.RegisterInput{InvitationCode: code, LoginID: "x", Password: "short", DisplayName: ""})
		wantInputError(t, err, auth.FieldLoginID, auth.CodeTooShort)
		wantInputError(t, err, auth.FieldPassword, auth.CodeTooShort)
		wantInputError(t, err, auth.FieldDisplayName, auth.CodeRequired)
		if _, err := env.Service.Register(ctx, registerInput(code, "fresh2")); err != nil {
			t.Errorf("コードはまだ使えるはず: %v", err)
		}
	})

	t.Run("招待コードが空", func(t *testing.T) {
		_, err := env.Service.Register(ctx, registerInput(" - ", "empty"))
		wantInputError(t, err, auth.FieldInvitationCode, auth.CodeRequired)
	})
}

// 同じ招待コードで 10 人が同時に登録しても、成功は 1 回だけ（SC-005）。
func TestRegisterConcurrently(t *testing.T) {
	env := authtest.New(t)
	code := issue(t, env, adminViewer(t, env))

	const n = 10
	var wg sync.WaitGroup
	results := make([]error, n)
	for i := range n {
		wg.Go(func() {
			_, results[i] = env.Service.Register(context.Background(), registerInput(code, "user"+string(rune('a'+i))))
		})
	}
	wg.Wait()

	success := 0
	for _, err := range results {
		if err == nil {
			success++
			continue
		}
		wantInputError(t, err, auth.FieldInvitationCode, auth.CodeInvitationUnavailable)
	}
	if success != 1 {
		t.Errorf("成功は 1 回だけのはず: %d", success)
	}
}
