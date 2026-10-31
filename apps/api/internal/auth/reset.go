package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// IssueResetCode はユーザー専用の再設定コードを発行する（管理者のコマンドの admin:reset-code）。
// 同じユーザーの未使用のコードは使えなくなる（FR-043）。
func (s *Service) IssueResetCode(ctx context.Context, loginID string) (code string, expiresAt time.Time, err error) {
	user, err := s.findUser(ctx, loginID)
	if err != nil {
		return "", time.Time{}, err
	}
	if user.DisabledAt != nil {
		return "", time.Time{}, ErrUserDisabled
	}
	now := s.now()
	token := NewToken()
	expiresAt = now.Add(ResetCodeTTL)
	err = s.withTx(ctx, func(q *db.Queries) error {
		if err := q.DeleteUnusedResetCodes(ctx, user.ID); err != nil {
			return fmt.Errorf("前の再設定コードを消せません: %w", err)
		}
		if err := q.CreateResetCode(ctx, db.CreateResetCodeParams{
			UserID:    user.ID,
			CodeHash:  HashToken(token),
			Now:       now,
			ExpiresAt: expiresAt,
		}); err != nil {
			return fmt.Errorf("再設定コードを作れません: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return FormatCode(token), expiresAt, nil
}

// resetCodeUnavailable は、再設定コードが使えないどの場合にも返す失敗（FR-038）。
var resetCodeUnavailable = fail(FieldResetCode, CodeResetCodeUnavailable)

// errResetCodeNotUsable はトランザクションの中で再設定コードが使えなかったことを表す
// （InputErrors はスライスで errors.Is で比べられないため、別の値にする）。
var errResetCodeNotUsable = errors.New("再設定コードが使えません")

// ResetPassword は再設定コードでパスワードを設定し直す。そのユーザーのセッションと、ログイン・再設定の
// 失敗の記録をすべて消す（FR-030、FR-039）。ログインはさせない。
func (s *Service) ResetPassword(ctx context.Context, loginID, code, newPassword string) error {
	code = NormalizeCode(code)
	var loginErr, codeErr *InputError
	if NormalizeLoginID(loginID) == "" {
		loginErr = &InputError{Field: FieldLoginID, Code: CodeRequired}
	}
	if code == "" {
		codeErr = &InputError{Field: FieldResetCode, Code: CodeRequired}
	}
	if err := collect(loginErr, codeErr, ValidatePassword(FieldNewPassword, newPassword)); err != nil {
		return err
	}
	if !ValidLoginID(loginID) {
		// 規則に合わないログイン ID のユーザーはいないため、失敗の記録はしない
		return resetCodeUnavailable
	}

	now := s.now()
	key := LoginIDKey(loginID)
	if err := s.checkLock(ctx, purposePasswordReset, key, now); err != nil {
		return err
	}
	failed := func() error {
		if err := s.recordFailure(ctx, purposePasswordReset, key, now); err != nil {
			return err
		}
		return resetCodeUnavailable
	}

	user, err := s.findUser(ctx, loginID)
	if errors.Is(err, ErrUserNotFound) {
		return failed()
	}
	if err != nil {
		return err
	}
	if user.DisabledAt != nil {
		return failed()
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	err = s.withTx(ctx, func(q *db.Queries) error {
		id, err := q.LockUsableResetCode(ctx, db.LockUsableResetCodeParams{
			UserID:   user.ID,
			CodeHash: HashToken(code),
			Now:      now,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return errResetCodeNotUsable
		}
		if err != nil {
			return fmt.Errorf("再設定コードを読めません: %w", err)
		}
		if err := q.UpdatePasswordHash(ctx, db.UpdatePasswordHashParams{PasswordHash: hash, Now: now, ID: user.ID}); err != nil {
			return fmt.Errorf("パスワードを変えられません: %w", err)
		}
		if err := q.MarkResetCodeUsed(ctx, db.MarkResetCodeUsedParams{Now: now, ID: id}); err != nil {
			return fmt.Errorf("再設定コードを使用済みにできません: %w", err)
		}
		if _, err := q.DeleteUserSessions(ctx, user.ID); err != nil {
			return fmt.Errorf("セッションを消せません: %w", err)
		}
		if err := clearFailures(ctx, q, purposeLogin, key); err != nil {
			return err
		}
		return clearFailures(ctx, q, purposePasswordReset, key)
	})
	if errors.Is(err, errResetCodeNotUsable) {
		return failed()
	}
	return err
}
