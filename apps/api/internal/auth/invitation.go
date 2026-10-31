package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// IssueInvitation は招待コードを発行する。平文のコードは戻り値だけで返し、DB にはハッシュだけを保存する。
func (s *Service) IssueInvitation(ctx context.Context, issuer Viewer) (code string, expiresAt time.Time, err error) {
	if !issuer.IsAdmin {
		return "", time.Time{}, ErrForbidden
	}
	now := s.now()
	token := NewToken()
	expiresAt = now.Add(InvitationTTL)
	if err := s.q.CreateInvitation(ctx, db.CreateInvitationParams{
		CodeHash:  HashToken(token),
		IssuedBy:  issuer.UserID,
		Now:       now,
		ExpiresAt: expiresAt,
	}); err != nil {
		return "", time.Time{}, fmt.Errorf("招待コードを作れません: %w", err)
	}
	return FormatCode(token), expiresAt, nil
}

// RegisterInput は登録の入力。
type RegisterInput struct {
	InvitationCode string
	LoginID        string
	Password       string
	DisplayName    string
}

// Register は招待コードで登録し、新しいセッションを返す（research R8）。
func (s *Service) Register(ctx context.Context, in RegisterInput) (Session, error) {
	code := NormalizeCode(in.InvitationCode)
	var codeErr *InputError
	if code == "" {
		codeErr = &InputError{Field: FieldInvitationCode, Code: CodeRequired}
	}
	loginID, loginErr := ValidateLoginID(in.LoginID)
	displayName, nameErr := ValidateDisplayName(in.DisplayName)
	if err := collect(codeErr, loginErr, ValidatePassword(FieldPassword, in.Password), nameErr); err != nil {
		return Session{}, err
	}
	// Argon2id は遅いため、トランザクションの外で計算する
	hash, err := HashPassword(in.Password)
	if err != nil {
		return Session{}, err
	}

	now := s.now()
	var sess Session
	err = s.withTx(ctx, func(q *db.Queries) error {
		// 招待コードを先に確かめ、使えない招待コードの持ち主にログイン ID の有無を知らせない
		invitationID, err := q.LockUsableInvitation(ctx, db.LockUsableInvitationParams{CodeHash: HashToken(code), Now: now})
		if errors.Is(err, pgx.ErrNoRows) {
			return fail(FieldInvitationCode, CodeInvitationUnavailable)
		}
		if err != nil {
			return fmt.Errorf("招待コードを読めません: %w", err)
		}
		user, err := q.CreateUser(ctx, db.CreateUserParams{
			LoginID:      loginID,
			DisplayName:  displayName,
			PasswordHash: hash,
			IsAdmin:      false,
			Now:          now,
		})
		if isLoginIDTaken(err) {
			// ロールバックし、招待コードは未使用のまま残す（FR-014）
			return fail(FieldLoginID, CodeLoginIDTaken)
		}
		if err != nil {
			return fmt.Errorf("ユーザーを作れません: %w", err)
		}
		if err := q.MarkInvitationUsed(ctx, db.MarkInvitationUsedParams{ID: invitationID, Now: now, UsedBy: user.ID}); err != nil {
			return fmt.Errorf("招待コードを使用済みにできません: %w", err)
		}
		sess, err = createSession(ctx, q, user, now)
		return err
	})
	if err != nil {
		return Session{}, err
	}
	return sess, nil
}
