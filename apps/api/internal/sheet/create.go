package sheet

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// ErrUnknownSystem は、知らないシステムを指定したときのエラー。GraphQL の enum で弾かれるため、通常は起きない。
var ErrUnknownSystem = errors.New("知らないシステムです")

// Create はシートを作る。所有者は引数だけから決め、入力からは受け取らない（FR-021）。
func (s *Service) Create(ctx context.Context, owner uuid.UUID, system System, f Fields) (db.CharacterSheet, error) {
	if !system.Valid() {
		return db.CharacterSheet{}, ErrUnknownSystem
	}
	v, err := Validate(f)
	if err != nil {
		return db.CharacterSheet{}, err
	}
	row, err := s.q.CreateCharacterSheet(ctx, db.CreateCharacterSheetParams{
		System:        string(system),
		OwnerID:       owner,
		CharacterName: v.CharacterName,
		PlayerName:    v.PlayerName,
		Age:           v.Age,
		Sex:           v.Sex,
		Memo:          v.Memo,
		Data:          v.Data,
		Now:           s.now(),
	})
	if err != nil {
		return db.CharacterSheet{}, fmt.Errorf("シートを作れません: %w", err)
	}
	return row, nil
}
