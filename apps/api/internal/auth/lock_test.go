package auth_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func wantLocked(t *testing.T, err error, until time.Time) {
	t.Helper()
	var le *auth.LockedError
	if !errors.As(err, &le) {
		t.Fatalf("LockedError のはず: %v", err)
	}
	if !le.Until.Equal(until) {
		t.Errorf("ロックが解ける日時: got %v, want %v", le.Until, until)
	}
}

func failLogin(t *testing.T, env *authtest.Env, loginID string, times int) error {
	t.Helper()
	var err error
	for range times {
		_, err = env.Service.Login(context.Background(), loginID, "wrong-password", nil)
	}
	return err
}

func failureCount(t *testing.T, env *authtest.Env, key string) int {
	t.Helper()
	var n int
	err := env.Pool.QueryRow(context.Background(),
		"SELECT COALESCE((SELECT failure_count FROM auth_failures WHERE purpose = 'login' AND login_id_key = $1), 0)", key).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestLoginLock(t *testing.T) {
	ctx := context.Background()

	t.Run("4 回の失敗の後の成功で回数が 0 に戻る", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		wantInputError(t, failLogin(t, env, "alice", 4), auth.FieldNone, auth.CodeInvalidCredentials)
		if _, err := env.Service.Login(ctx, "alice", "password1", nil); err != nil {
			t.Fatal(err)
		}
		if n := failureCount(t, env, "alice"); n != 0 {
			t.Errorf("回数は 0 のはず: %d", n)
		}
	})

	t.Run("5 回目でロックし、15 分後に解ける", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		_ = failLogin(t, env, "alice", 4)
		env.Clock.Advance(time.Minute)
		lockedAt := env.Clock.Now()
		wantLocked(t, failLogin(t, env, "ALICE", 1), lockedAt.Add(15*time.Minute))

		// ロック中は正しいパスワードでもログインできず、回数もロックの時間も変わらない
		env.Clock.Advance(14 * time.Minute)
		_, err := env.Service.Login(ctx, "alice", "password1", nil)
		wantLocked(t, err, lockedAt.Add(15*time.Minute))
		wantLocked(t, failLogin(t, env, "alice", 1), lockedAt.Add(15*time.Minute))
		if n := failureCount(t, env, "alice"); n != 5 {
			t.Errorf("回数は 5 のまま: %d", n)
		}

		env.Clock.Advance(time.Minute)
		if _, err := env.Service.Login(ctx, "alice", "password1", nil); err != nil {
			t.Errorf("15 分後にはログインできる: %v", err)
		}
		if n := env.Count(t, "auth_failures"); n != 0 {
			t.Errorf("成功したら記録は消える: %d", n)
		}
	})

	t.Run("ロックが過ぎた後の最初の失敗は 1 回目", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		_ = failLogin(t, env, "alice", 5)
		env.Clock.Advance(15 * time.Minute)
		wantInputError(t, failLogin(t, env, "alice", 1), auth.FieldNone, auth.CodeInvalidCredentials)
		if n := failureCount(t, env, "alice"); n != 1 {
			t.Errorf("1 回目から数え直す: %d", n)
		}
	})

	t.Run("存在しないログイン ID も同じようにロックする", func(t *testing.T) {
		env := authtest.New(t)
		_ = failLogin(t, env, "nobody", 4)
		wantLocked(t, failLogin(t, env, "nobody", 1), env.Clock.Now().Add(15*time.Minute))
	})

	t.Run("形式に合わないログイン ID は記録しない", func(t *testing.T) {
		env := authtest.New(t)
		_ = failLogin(t, env, "no!", 6)
		if n := env.Count(t, "auth_failures"); n != 0 {
			t.Errorf("記録しないはず: %d", n)
		}
	})

	t.Run("別のユーザーには影響しない", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		env.CreateUser(t, "bob", "password1", false)
		_ = failLogin(t, env, "alice", 5)
		if _, err := env.Service.Login(ctx, "bob", "password1", nil); err != nil {
			t.Errorf("bob はログインできる: %v", err)
		}
	})

	t.Run("ログインの成功で期限切れのセッションと古い失敗の記録を消す", func(t *testing.T) {
		env := authtest.New(t)
		env.CreateUser(t, "alice", "password1", false)
		if _, err := env.Service.Login(ctx, "alice", "password1", nil); err != nil {
			t.Fatal(err)
		}
		_ = failLogin(t, env, "old", 1)
		_ = failLogin(t, env, "locked", 5)
		env.Clock.Advance(auth.SessionTTL + time.Second)
		_ = failLogin(t, env, "recent", 1)
		_ = failLogin(t, env, "stilllocked", 5)
		if _, err := env.Service.Login(ctx, "alice", "password1", nil); err != nil {
			t.Fatal(err)
		}
		if n := env.Count(t, "sessions"); n != 1 {
			t.Errorf("期限切れのセッションは消え、新しいものだけ残る: %d", n)
		}
		if n := env.Count(t, "auth_failures"); n != 2 {
			t.Errorf("recent と stilllocked だけ残る: %d", n)
		}
	})
}

// 同じログイン ID で同時に 5 回失敗しても、数え漏れずにロックする。
func TestLoginLockConcurrently(t *testing.T) {
	env := authtest.New(t)
	env.CreateUser(t, "alice", "password1", false)

	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			_, _ = env.Service.Login(context.Background(), "alice", "wrong-password", nil)
		})
	}
	wg.Wait()

	if n := failureCount(t, env, "alice"); n != 5 {
		t.Errorf("回数は 5 のはず: %d", n)
	}
	_, err := env.Service.Login(context.Background(), "alice", "password1", nil)
	wantLocked(t, err, env.Clock.Now().Add(15*time.Minute))
}
