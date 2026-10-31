package auth

import (
	"context"
	"fmt"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// DisableUser はユーザーを無効にし、そのユーザーのセッションをすべて消す（FR-044）。
// 既に無効なら何もしない。changed は状態を変えたか、sessions は消したセッションの件数。
func (s *Service) DisableUser(ctx context.Context, loginID string) (changed bool, sessions int64, err error) {
	user, err := s.findUser(ctx, loginID)
	if err != nil {
		return false, 0, err
	}
	err = s.withTx(ctx, func(q *db.Queries) error {
		n, err := q.DisableUser(ctx, db.DisableUserParams{Now: s.now(), ID: user.ID})
		if err != nil {
			return fmt.Errorf("ユーザーを無効にできません: %w", err)
		}
		changed = n > 0
		sessions, err = q.DeleteUserSessions(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("セッションを消せません: %w", err)
		}
		return nil
	})
	if err != nil {
		return false, 0, err
	}
	return changed, sessions, nil
}

// EnableUser は無効にしたユーザーを有効に戻す。既に有効なら何もしない。パスワードは変わらない。
func (s *Service) EnableUser(ctx context.Context, loginID string) (changed bool, err error) {
	user, err := s.findUser(ctx, loginID)
	if err != nil {
		return false, err
	}
	n, err := s.q.EnableUser(ctx, db.EnableUserParams{Now: s.now(), ID: user.ID})
	if err != nil {
		return false, fmt.Errorf("ユーザーを有効にできません: %w", err)
	}
	return n > 0, nil
}
