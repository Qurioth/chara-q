package auth_test

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func issueReset(t *testing.T, env *authtest.Env, loginID string) string {
	t.Helper()
	code, _, err := env.Service.IssueResetCode(context.Background(), loginID)
	if err != nil {
		t.Fatalf("IssueResetCode: %v", err)
	}
	return code
}

func TestIssueResetCode(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	ctx := context.Background()

	code, expiresAt, err := env.Service.IssueResetCode(ctx, "ALICE")
	if err != nil {
		t.Fatal(err)
	}
	if want := env.Clock.Now().Add(24 * time.Hour); !expiresAt.Equal(want) {
		t.Errorf("期限は 24 時間後: got %v, want %v", expiresAt, want)
	}
	var hash []byte
	if err := env.Pool.QueryRow(ctx, "SELECT code_hash FROM password_reset_codes").Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(hash, auth.HashToken(auth.NormalizeCode(code))) {
		t.Error("DB にはハッシュだけを保存する")
	}

	if _, _, err := env.Service.IssueResetCode(ctx, "nobody"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Errorf("存在しないユーザー: %v", err)
	}
	if _, _, err := env.Service.DisableUser(ctx, "alice"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := env.Service.IssueResetCode(ctx, "alice"); !errors.Is(err, auth.ErrUserDisabled) {
		t.Errorf("無効なユーザーには発行しない: %v", err)
	}
}

func TestResetPassword(t *testing.T) {
	ctx := context.Background()

	t.Run("成功するとセッションと失敗の記録が消え、新しいパスワードでログインできる", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		sess := login(t, env, "alice", "password1")
		_ = failLogin(t, env, "alice", 5) // ロックする
		code := issueReset(t, env, "alice")

		if err := env.Service.ResetPassword(ctx, "Alice", code, "password2"); err != nil {
			t.Fatalf("ResetPassword: %v", err)
		}
		if v, _ := serve(env, sess.Token); v != nil {
			t.Error("既存のセッションは切れる")
		}
		if n := env.Count(t, "auth_failures"); n != 0 {
			t.Errorf("失敗の記録は消える: %d", n)
		}
		login(t, env, "alice", "password2")
		_, err := env.Service.Login(ctx, "alice", "password1", nil)
		wantInputError(t, err, auth.FieldNone, auth.CodeInvalidCredentials)

		// 同じコードは 2 回使えない
		err = env.Service.ResetPassword(ctx, "alice", code, "password3")
		wantInputError(t, err, auth.FieldResetCode, auth.CodeResetCodeUnavailable)
	})

	t.Run("使えないコードはどれも同じ失敗", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		env.CreateUser(t, "bob", "password1", false)
		first := issueReset(t, env, "alice")
		second := issueReset(t, env, "alice") // 1 回目は使えなくなる
		bobCode := issueReset(t, env, "bob")

		tests := []struct {
			name, loginID, code string
		}{
			{"前に発行したコード", "alice", first},
			{"別のユーザー用のコード", "alice", bobCode},
			{"存在しないコード", "alice", "NOSUCHCODE"},
			{"存在しないユーザー", "nobody", second},
			{"形式に合わないログイン ID", "a!", second},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := env.Service.ResetPassword(ctx, tt.loginID, tt.code, "password2")
				wantInputError(t, err, auth.FieldResetCode, auth.CodeResetCodeUnavailable)
			})
		}

		env.Clock.Advance(24*time.Hour + time.Second)
		err := env.Service.ResetPassword(ctx, "alice", second, "password2")
		wantInputError(t, err, auth.FieldResetCode, auth.CodeResetCodeUnavailable)
	})

	t.Run("無効なユーザーは再設定できない", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		code := issueReset(t, env, "alice")
		if _, _, err := env.Service.DisableUser(ctx, "alice"); err != nil {
			t.Fatal(err)
		}
		err := env.Service.ResetPassword(ctx, "alice", code, "password2")
		wantInputError(t, err, auth.FieldResetCode, auth.CodeResetCodeUnavailable)
	})

	t.Run("新しいパスワードが規則に合わなければコードは未使用のまま", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		code := issueReset(t, env, "alice")
		err := env.Service.ResetPassword(ctx, "alice", code, "short")
		wantInputError(t, err, auth.FieldNewPassword, auth.CodeTooShort)
		err = env.Service.ResetPassword(ctx, "", "", "password2")
		wantInputError(t, err, auth.FieldLoginID, auth.CodeRequired)
		wantInputError(t, err, auth.FieldResetCode, auth.CodeRequired)
		if err := env.Service.ResetPassword(ctx, "alice", code, "password2"); err != nil {
			t.Errorf("コードはまだ使える: %v", err)
		}
	})

	t.Run("5 回続けて失敗するとロックし、ログインのロックとは別", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		var err error
		for range 5 {
			err = env.Service.ResetPassword(ctx, "alice", "NOSUCHCODE", "password2")
		}
		wantLocked(t, err, env.Clock.Now().Add(15*time.Minute))
		code := issueReset(t, env, "alice")
		err = env.Service.ResetPassword(ctx, "alice", code, "password2")
		wantLocked(t, err, env.Clock.Now().Add(15*time.Minute))
		login(t, env, "alice", "password1")
	})
}

// 同じ再設定コードを 10 回同時に使っても、成功は 1 回だけ（SC-005）。
func TestResetPasswordConcurrently(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	code := issueReset(t, env, "alice")

	const n = 10
	var wg sync.WaitGroup
	results := make([]error, n)
	for i := range n {
		wg.Go(func() {
			results[i] = env.Service.ResetPassword(context.Background(), "alice", code, "password2")
		})
	}
	wg.Wait()

	success := 0
	for _, err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Errorf("成功は 1 回だけのはず: %d（%v）", success, results)
	}
}
