package sheet

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// Copy は削除されていないシートを複製し、owner の所有にする。名前の末尾に「（コピー）」を付ける（research R12）。
func (s *Service) Copy(ctx context.Context, owner uuid.UUID, id string) (db.CharacterSheet, error) {
	uid, err := parseID(id)
	if err != nil {
		return db.CharacterSheet{}, err
	}
	row, err := s.q.CopyCharacterSheet(ctx, db.CopyCharacterSheetParams{ID: uid, OwnerID: owner, Now: s.now()})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.CharacterSheet{}, ErrNotFound
	}
	if err != nil {
		return db.CharacterSheet{}, fmt.Errorf("シートを複製できません: %w", err)
	}
	return row, nil
}
