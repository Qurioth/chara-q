package auth

import (
	"context"
	"fmt"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// UpdateDisplayName は本人の表示名を変える。
func (s *Service) UpdateDisplayName(ctx context.Context, v Viewer, displayName string) (Viewer, error) {
	displayName, nameErr := ValidateDisplayName(displayName)
	if err := collect(nameErr); err != nil {
		return Viewer{}, err
	}
	u, err := s.q.UpdateDisplayName(ctx, db.UpdateDisplayNameParams{
		DisplayName: displayName,
		Now:         s.now(),
		ID:          v.UserID,
	})
	if err != nil {
		return Viewer{}, fmt.Errorf("表示名を変えられません: %w", err)
	}
	return viewerOf(u), nil
}

// ChangePassword は本人のパスワードを変える。そのユーザーのセッションをすべて消し、
// 変更したブラウザ用の新しいセッションを返す（FR-033）。
func (s *Service) ChangePassword(ctx context.Context, v Viewer, current, next string) (Session, error) {
	var currentErr *InputError
	if current == "" {
		currentErr = &InputError{Field: FieldCurrentPassword, Code: CodeRequired}
	}
	if err := collect(currentErr, ValidatePassword(FieldNewPassword, next)); err != nil {
		return Session{}, err
	}
	user, err := s.findUser(ctx, v.LoginID)
	if err != nil {
		return Session{}, err
	}
	ok, err := VerifyPassword(current, user.PasswordHash)
	if err != nil {
		return Session{}, fmt.Errorf("パスワードを照合できません: %w", err)
	}
	if !ok {
		return Session{}, fail(FieldCurrentPassword, CodeWrongCurrentPassword)
	}
	hash, err := HashPassword(next)
	if err != nil {
		return Session{}, err
	}

	now := s.now()
	var sess Session
	err = s.withTx(ctx, func(q *db.Queries) error {
		if err := q.UpdatePasswordHash(ctx, db.UpdatePasswordHashParams{PasswordHash: hash, Now: now, ID: user.ID}); err != nil {
			return fmt.Errorf("パスワードを変えられません: %w", err)
		}
		if _, err := q.DeleteUserSessions(ctx, user.ID); err != nil {
			return fmt.Errorf("セッションを消せません: %w", err)
		}
		sess, err = createSession(ctx, q, user, now)
		return err
	})
	if err != nil {
		return Session{}, err
	}
	return sess, nil
}
