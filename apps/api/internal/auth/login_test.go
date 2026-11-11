package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

// inputCodes は error が InputErrors なら、その (field, code) の一覧を返す。
func inputCodes(t *testing.T, err error) []auth.InputError {
	t.Helper()
	var ie auth.InputErrors
	if !errors.As(err, &ie) {
		t.Fatalf("InputErrors のはず: %v", err)
	}
	return ie
}

func wantInputError(t *testing.T, err error, field auth.Field, code auth.Code) {
	t.Helper()
	for _, e := range inputCodes(t, err) {
		if e.Field == field && e.Code == code {
			return
		}
	}
	t.Errorf("%s:%s がない: %v", field, code, err)
}

func TestLogin(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "Alice", "password1", false)
	ctx := context.Background()

	t.Run("成功", func(t *testing.T) {
		tests := []struct {
			name    string
			loginID string
		}{
			{"登録したとおり", "Alice"},
			{"大文字と小文字の違いは同じ", "aLICE"},
			{"前後の空白は取り除く", "  alice "},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sess, err := env.Service.Login(ctx, tt.loginID, "password1", nil)
				if err != nil {
					t.Fatalf("Login: %v", err)
				}
				if len(sess.Token) != 26 || sess.Viewer.LoginID != "Alice" {
					t.Errorf("got %+v", sess)
				}
			})
		}
	})

	t.Run("失敗はどれも同じ", func(t *testing.T) {
		tests := []struct {
			name, loginID, password string
		}{
			{"誤ったパスワード", "alice", "password2"},
			{"存在しないログイン ID", "nobody", "password1"},
			{"形式に合わないログイン ID", "a!", "password1"},
			{"空", "", ""},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := env.Service.Login(ctx, tt.loginID, tt.password, nil)
				errs := inputCodes(t, err)
				if len(errs) != 1 || errs[0].Field != auth.FieldNone || errs[0].Code != auth.CodeInvalidCredentials {
					t.Errorf("INVALID_CREDENTIALS だけのはず: %v", errs)
				}
			})
		}
	})
}

func TestLoginReplacesPreviousSession(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	ctx := context.Background()

	first, err := env.Service.Login(ctx, "alice", "password1", nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := env.Service.Login(ctx, "alice", "password1", auth.HashToken(first.Token))
	if err != nil {
		t.Fatal(err)
	}
	if first.Token == second.Token {
		t.Error("新しいセッション ID を発行するはず（FR-019）")
	}
	if v, _ := serve(env, first.Token); v != nil {
		t.Error("前のセッションは消えるはず")
	}
	if v, _ := serve(env, second.Token); v == nil {
		t.Error("新しいセッションは使えるはず")
	}
}

func TestLogout(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	ctx := context.Background()
	sess, err := env.Service.Login(ctx, "alice", "password1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.Service.Logout(ctx, auth.HashToken(sess.Token)); err != nil {
		t.Fatal(err)
	}
	if v, _ := serve(env, sess.Token); v != nil {
		t.Error("ログアウトしたセッションは使えない")
	}
	if err := env.Service.Logout(ctx, nil); err != nil {
		t.Errorf("ログインしていなくても成功する: %v", err)
	}
}
