package sheet

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// Update は所有者のシートの共通の項目と data を置き換える。所有者とリビジョンの確認と更新は
// 1 文で行い、同時に送られても成功は 1 回だけ（research R6）。
// 失敗は、入力の規則 → ErrNotFound → ErrForbidden → ErrRevisionConflict の順に判定する。
func (s *Service) Update(ctx context.Context, owner uuid.UUID, id string, revision int, f Fields) (db.CharacterSheet, error) {
	v, err := Validate(f)
	if err != nil {
		return db.CharacterSheet{}, err
	}
	uid, err := parseID(id)
	if err != nil {
		return db.CharacterSheet{}, err
	}
	rev, ok := toInt32(revision)
	if !ok {
		// どの行のリビジョンとも合わない
		return db.CharacterSheet{}, s.stateError(ctx, uid, owner, true)
	}
	row, err := s.q.UpdateCharacterSheet(ctx, db.UpdateCharacterSheetParams{
		ID:            uid,
		OwnerID:       owner,
		Revision:      rev,
		CharacterName: v.CharacterName,
		PlayerName:    v.PlayerName,
		Age:           v.Age,
		Sex:           v.Sex,
		Memo:          v.Memo,
		Data:          v.Data,
		Now:           s.now(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.CharacterSheet{}, s.stateError(ctx, uid, owner, true)
	}
	if err != nil {
		return db.CharacterSheet{}, fmt.Errorf("シートを更新できません: %w", err)
	}
	return row, nil
}
