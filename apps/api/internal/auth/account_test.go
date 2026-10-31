package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func login(t *testing.T, env *authtest.Env, loginID, password string) auth.Session {
	t.Helper()
	sess, err := env.Service.Login(context.Background(), loginID, password, nil)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	return sess
}

func TestUpdateDisplayName(t *testing.T) {
	env := authtest.New(t)
	u := env.CreateUser(t, "alice", "password1", false)
	v := login(t, env, "alice", "password1").Viewer
	ctx := context.Background()

	env.Clock.Advance(time.Minute)
	got, err := env.Service.UpdateDisplayName(ctx, v, "  ありす ")
	if err != nil {
		t.Fatal(err)
	}
	if got.DisplayName != "ありす" {
		t.Errorf("got %q", got.DisplayName)
	}
	var updatedAt time.Time
	if err := env.Pool.QueryRow(ctx, "SELECT updated_at FROM users WHERE id = $1", u.ID).Scan(&updatedAt); err != nil {
		t.Fatal(err)
	}
	if !updatedAt.Equal(env.Clock.Now()) {
		t.Errorf("updated_at が進むはず: %v", updatedAt)
	}

	_, err = env.Service.UpdateDisplayName(ctx, v, "改行\nあり")
	wantInputError(t, err, auth.FieldDisplayName, auth.CodeInvalidCharacters)
}

func TestChangePassword(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	a := login(t, env, "alice", "password1")
	b := login(t, env, "alice", "password1")
	ctx := context.Background()

	t.Run("現在のパスワードが違えば何も変わらない", func(t *testing.T) {
		_, err := env.Service.ChangePassword(ctx, a.Viewer, "wrong-password", "password2")
		wantInputError(t, err, auth.FieldCurrentPassword, auth.CodeWrongCurrentPassword)
		if v, _ := serve(env, b.Token); v == nil {
			t.Error("セッションは残るはず")
		}
		login(t, env, "alice", "password1")
	})

	t.Run("新しいパスワードが規則に合わない", func(t *testing.T) {
		_, err := env.Service.ChangePassword(ctx, a.Viewer, "password1", "short")
		wantInputError(t, err, auth.FieldNewPassword, auth.CodeTooShort)
		_, err = env.Service.ChangePassword(ctx, a.Viewer, "", "password2")
		wantInputError(t, err, auth.FieldCurrentPassword, auth.CodeRequired)
	})

	t.Run("成功するとほかのセッションがすべて切れる", func(t *testing.T) {
		sess, err := env.Service.ChangePassword(ctx, a.Viewer, "password1", "password2")
		if err != nil {
			t.Fatal(err)
		}
		if v, _ := serve(env, a.Token); v != nil {
			t.Error("変更前のセッションは切れる")
		}
		if v, _ := serve(env, b.Token); v != nil {
			t.Error("ほかのセッションは切れる")
		}
		if v, _ := serve(env, sess.Token); v == nil {
			t.Error("新しいセッションは使える")
		}
		if n := env.Count(t, "sessions"); n != 1 {
			t.Errorf("セッションは 1 つだけ: %d", n)
		}
		login(t, env, "alice", "password2")
		_, err = env.Service.Login(ctx, "alice", "password1", nil)
		wantInputError(t, err, auth.FieldNone, auth.CodeInvalidCredentials)
	})
}
