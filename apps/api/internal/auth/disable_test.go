package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func TestDisableAndEnableUser(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)
	a := login(t, env, "alice", "password1")
	b := login(t, env, "alice", "password1")
	ctx := context.Background()

	changed, sessions, err := env.Service.DisableUser(ctx, "ALICE")
	if err != nil {
		t.Fatal(err)
	}
	if !changed || sessions != 2 {
		t.Errorf("無効にして 2 つのセッションを消すはず: changed=%v sessions=%d", changed, sessions)
	}
	for _, s := range []auth.Session{a, b} {
		if v, _ := serve(env, s.Token); v != nil {
			t.Error("セッションはすぐに切れる")
		}
	}
	_, err = env.Service.Login(ctx, "alice", "password1", nil)
	wantInputError(t, err, auth.FieldNone, auth.CodeInvalidCredentials)

	changed, _, err = env.Service.DisableUser(ctx, "alice")
	if err != nil || changed {
		t.Errorf("既に無効なら何もせず成功する: changed=%v err=%v", changed, err)
	}

	changed, err = env.Service.EnableUser(ctx, "alice")
	if err != nil || !changed {
		t.Fatalf("有効に戻す: changed=%v err=%v", changed, err)
	}
	login(t, env, "alice", "password1")

	changed, err = env.Service.EnableUser(ctx, "alice")
	if err != nil || changed {
		t.Errorf("既に有効なら何もせず成功する: changed=%v err=%v", changed, err)
	}

	if _, _, err := env.Service.DisableUser(ctx, "nobody"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Errorf("存在しないユーザー: %v", err)
	}
	if _, err := env.Service.EnableUser(ctx, "nobody"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Errorf("存在しないユーザー: %v", err)
	}
}
