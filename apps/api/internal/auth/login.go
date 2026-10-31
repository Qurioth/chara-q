package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// invalidCredentials は、ログイン ID が存在しない・パスワードが違う・ユーザーが無効の、どの場合にも
// 返す失敗（FR-018）。
var invalidCredentials = fail(FieldNone, CodeInvalidCredentials)

// Login はログイン ID とパスワードでログインし、新しいセッションを返す。
// previousTokenHash はブラウザが送ってきた前のセッションで、あれば消す（FR-019）。
func (s *Service) Login(ctx context.Context, loginID, password string, previousTokenHash []byte) (Session, error) {
	now := s.now()
	if !ValidLoginID(loginID) {
		// 規則に合わないログイン ID は存在し得ないため、失敗の記録はしない
		VerifyDummy(password)
		return Session{}, invalidCredentials
	}
	key := LoginIDKey(loginID)
	if err := s.checkLock(ctx, purposeLogin, key, now); err != nil {
		return Session{}, err
	}

	user, err := s.findUser(ctx, loginID)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return Session{}, err
	}
	ok := false
	if err == nil {
		ok, err = VerifyPassword(password, user.PasswordHash)
		if err != nil {
			return Session{}, fmt.Errorf("パスワードを照合できません: %w", err)
		}
	} else {
		// 応答時間からログイン ID の有無がわからないようにする
		VerifyDummy(password)
	}
	if !ok || user.DisabledAt != nil {
		if err := s.recordFailure(ctx, purposeLogin, key, now); err != nil {
			return Session{}, err
		}
		return Session{}, invalidCredentials
	}

	var sess Session
	err = s.withTx(ctx, func(q *db.Queries) error {
		if err := clearFailures(ctx, q, purposeLogin, key); err != nil {
			return err
		}
		if previousTokenHash != nil {
			if err := q.DeleteSession(ctx, previousTokenHash); err != nil {
				return fmt.Errorf("前のセッションを消せません: %w", err)
			}
		}
		sess, err = createSession(ctx, q, user, now)
		return err
	})
	if err != nil {
		return Session{}, err
	}
	s.cleanup(ctx, now)
	return sess, nil
}

// Logout はセッションを消す。
func (s *Service) Logout(ctx context.Context, tokenHash []byte) error {
	if tokenHash == nil {
		return nil
	}
	if err := s.q.DeleteSession(ctx, tokenHash); err != nil {
		return fmt.Errorf("セッションを消せません: %w", err)
	}
	return nil
}
