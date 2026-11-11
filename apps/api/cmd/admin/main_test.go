package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func TestAdminCommands(t *testing.T) {
	env := authtest.New(t) // 表を空にする
	url := os.Getenv("TEST_DATABASE_URL")
	getenv := func(k string) string {
		if k == "DATABASE_URL" {
			return url
		}
		return ""
	}
	exec := func(stdin string, args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		code := run(context.Background(), args, strings.NewReader(stdin), &out, &errOut, getenv)
		return code, out.String(), errOut.String()
	}

	t.Run("管理者を作る", func(t *testing.T) {
		code, out, errOut := exec("password1\npassword1\n", "create-admin", "-login-id", "Admin", "-display-name", "管理者")
		if code != 0 || !strings.Contains(out, "管理者を作りました: Admin") {
			t.Fatalf("code=%d out=%q err=%q", code, out, errOut)
		}
		if !strings.Contains(errOut, "接続先: ") || strings.Contains(errOut, "chara_q:chara_q") {
			t.Errorf("接続先を出し、認証情報は出さない: %q", errOut)
		}
		if strings.Contains(out+errOut, "password1") {
			t.Error("パスワードを出力してはいけない")
		}
		if _, err := env.Service.Login(context.Background(), "admin", "password1", nil); err != nil {
			t.Errorf("作った管理者でログインできる: %v", err)
		}
	})

	t.Run("作れない場合", func(t *testing.T) {
		tests := []struct {
			name  string
			stdin string
			args  []string
			want  string
		}{
			{"パスワードが一致しない", "password1\npassword2\n", []string{"create-admin", "-login-id", "other", "-display-name", "x"}, "パスワードが一致しません"},
			{"使われているログイン ID", "password1\npassword1\n", []string{"create-admin", "-login-id", "ADMIN", "-display-name", "x"}, "ログイン ID が既に使われています"},
			{"規則に合わないログイン ID", "password1\npassword1\n", []string{"create-admin", "-login-id", "a!", "-display-name", "x"}, "半角英数字"},
			{"引数がない", "", []string{"create-admin"}, "-login-id と -display-name"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				code, _, errOut := exec(tt.stdin, tt.args...)
				if code != 1 || !strings.Contains(errOut, tt.want) {
					t.Errorf("code=%d err=%q", code, errOut)
				}
			})
		}
	})

	env.CreateUser(t, "member", "password1", false)

	t.Run("一覧", func(t *testing.T) {
		code, out, _ := exec("", "users")
		if code != 0 || !strings.Contains(out, "Admin") || !strings.Contains(out, "member") || !strings.Contains(out, "有効") {
			t.Errorf("code=%d out=%q", code, out)
		}
	})

	t.Run("再設定コード", func(t *testing.T) {
		code, out, _ := exec("", "reset-code", "member")
		if code != 0 || !strings.Contains(out, "再設定コード: ") || !strings.Contains(out, "有効期限: ") {
			t.Errorf("code=%d out=%q", code, out)
		}
		code, _, errOut := exec("", "reset-code", "nobody")
		if code != 1 || !strings.Contains(errOut, "ユーザーが見つかりません: nobody") {
			t.Errorf("code=%d err=%q", code, errOut)
		}
	})

	t.Run("無効化と有効化", func(t *testing.T) {
		if _, err := env.Service.Login(context.Background(), "member", "password1", nil); err != nil {
			t.Fatal(err)
		}
		code, out, _ := exec("", "disable", "member")
		if code != 0 || !strings.Contains(out, "無効にしました: member（切ったセッション: 1）") {
			t.Errorf("code=%d out=%q", code, out)
		}
		code, out, _ = exec("", "disable", "member")
		if code != 0 || !strings.Contains(out, "既に無効です") {
			t.Errorf("code=%d out=%q", code, out)
		}
		code, _, errOut := exec("", "reset-code", "member")
		if code != 1 || !strings.Contains(errOut, "無効なユーザーには再設定コードを発行できません") {
			t.Errorf("code=%d err=%q", code, errOut)
		}
		code, out, _ = exec("", "enable", "member")
		if code != 0 || !strings.Contains(out, "有効にしました: member") {
			t.Errorf("code=%d out=%q", code, out)
		}
	})

	t.Run("不明なサブコマンド", func(t *testing.T) {
		if code, _, errOut := exec("", "unknown"); code != 1 || !strings.Contains(errOut, "使い方") {
			t.Errorf("code=%d err=%q", code, errOut)
		}
	})
}
