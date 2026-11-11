package auth

import (
	"context"
	"fmt"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// CreateAdmin は管理者ユーザーを作る（管理者のコマンドの admin:create）。
func (s *Service) CreateAdmin(ctx context.Context, loginID, displayName, password string) (db.User, error) {
	loginID, loginErr := ValidateLoginID(loginID)
	displayName, nameErr := ValidateDisplayName(displayName)
	if err := collect(loginErr, nameErr, ValidatePassword(FieldPassword, password)); err != nil {
		return db.User{}, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return db.User{}, err
	}
	now := s.now()
	u, err := s.q.CreateUser(ctx, db.CreateUserParams{
		LoginID:      loginID,
		DisplayName:  displayName,
		PasswordHash: hash,
		IsAdmin:      true,
		Now:          now,
	})
	if isLoginIDTaken(err) {
		return db.User{}, fail(FieldLoginID, CodeLoginIDTaken)
	}
	if err != nil {
		return db.User{}, fmt.Errorf("ユーザーを作れません: %w", err)
	}
	return u, nil
}

// ListUsers はユーザーを作成日時の順に返す。
func (s *Service) ListUsers(ctx context.Context) ([]db.User, error) {
	users, err := s.q.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("ユーザーの一覧を読めません: %w", err)
	}
	return users, nil
}
