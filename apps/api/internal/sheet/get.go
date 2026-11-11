package sheet

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// Get は有効なシートを返す。存在しない・削除済み・形の正しくない ID なら nil（エラーにしない、FR-018）。
func (s *Service) Get(ctx context.Context, id string) (*db.CharacterSheet, error) {
	uid, err := parseID(id)
	if err != nil {
		return nil, nil
	}
	row, err := s.q.GetCharacterSheet(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("シートを読めません: %w", err)
	}
	return &row, nil
}
