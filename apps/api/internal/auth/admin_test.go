package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func TestCreateAdmin(t *testing.T) {
	env := authtest.New(t)
	ctx := context.Background()

	u, err := env.Service.CreateAdmin(ctx, " Admin ", " 管理者 ", "password1")
	if err != nil {
		t.Fatalf("CreateAdmin: %v", err)
	}
	if !u.IsAdmin || u.LoginID != "Admin" || u.DisplayName != "管理者" {
		t.Errorf("got %+v", u)
	}
	if _, err := env.Service.Login(ctx, "admin", "password1", nil); err != nil {
		t.Errorf("作った管理者でログインできるはず: %v", err)
	}

	tests := []struct {
		name                           string
		loginID, displayName, password string
		field                          auth.Field
		code                           auth.Code
	}{
		{"使われているログイン ID", "admin", "別の人", "password2", auth.FieldLoginID, auth.CodeLoginIDTaken},
		{"大文字だけ違うログイン ID", "ADMIN", "別の人", "password2", auth.FieldLoginID, auth.CodeLoginIDTaken},
		{"規則に合わないログイン ID", "a", "x", "password1", auth.FieldLoginID, auth.CodeTooShort},
		{"短いパスワード", "admin2", "x", "short", auth.FieldPassword, auth.CodeTooShort},
		{"空の表示名", "admin2", " ", "password1", auth.FieldDisplayName, auth.CodeRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := env.Service.CreateAdmin(ctx, tt.loginID, tt.displayName, tt.password)
			wantInputError(t, err, tt.field, tt.code)
		})
	}

	// 失敗しても既存のユーザーは変わらない
	if _, err := env.Service.Login(ctx, "admin", "password1", nil); err != nil {
		t.Errorf("既存のユーザーは変わらないはず: %v", err)
	}
	if n := env.Count(t, "users"); n != 1 {
		t.Errorf("ユーザーは 1 人のはず: %d", n)
	}
}

func TestListUsers(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "first", "password1", true)
	env.Clock.Advance(time.Minute)
	env.CreateUser(t, "second", "password1", false)

	users, err := env.Service.ListUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 || users[0].LoginID != "first" || users[1].LoginID != "second" {
		t.Errorf("作成日時の順のはず: %+v", users)
	}
}
