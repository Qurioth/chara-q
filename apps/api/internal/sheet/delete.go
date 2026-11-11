package sheet

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// Delete は所有者のシートを論理削除する（deleted_at を記録し、データは残す、research R7）。
// updated_at と revision は変えない。取り消しは管理者の DB 操作で行う（data-model.md の「削除の取り消し」）。
func (s *Service) Delete(ctx context.Context, owner uuid.UUID, id string) error {
	uid, err := parseID(id)
	if err != nil {
		return err
	}
	n, err := s.q.DeleteCharacterSheet(ctx, db.DeleteCharacterSheetParams{ID: uid, OwnerID: owner, Now: s.now()})
	if err != nil {
		return fmt.Errorf("シートを削除できません: %w", err)
	}
	if n == 0 {
		return s.stateError(ctx, uid, owner, false)
	}
	return nil
}
