// Package sheettest は、DB を使うシートのテストの補助を置く。
// 表は authtest.New の TRUNCATE users ... CASCADE で空になる。
package sheettest

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
)

// NewService は、authtest の DB と時計を使う sheet.Service を返す。
func NewService(env *authtest.Env) *sheet.Service {
	s := sheet.NewService(env.Pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.SetClock(env.Clock.Now)
	return s
}

// CreateSheet はテスト用のシートを作る。日時は env の時計の時刻。
func CreateSheet(t *testing.T, env *authtest.Env, owner uuid.UUID, system sheet.System, name, playerName string) db.CharacterSheet {
	t.Helper()
	row, err := env.Queries.CreateCharacterSheet(context.Background(), db.CreateCharacterSheetParams{
		System:        string(system),
		OwnerID:       owner,
		CharacterName: name,
		PlayerName:    playerName,
		Data:          []byte(`{}`),
		Now:           env.Clock.Now(),
	})
	if err != nil {
		t.Fatalf("CreateCharacterSheet: %v", err)
	}
	return row
}

// SetTimes は、並び順のテストのために作成日時と更新日時を直接書き換える。
func SetTimes(t *testing.T, env *authtest.Env, id uuid.UUID, createdAt, updatedAt time.Time) {
	t.Helper()
	if _, err := env.Pool.Exec(context.Background(),
		"UPDATE character_sheets SET created_at = $2, updated_at = $3 WHERE id = $1", id, createdAt, updatedAt); err != nil {
		t.Fatalf("SetTimes: %v", err)
	}
}

// MarkDeleted は、行を削除済みにする（所有者の確認なし）。
func MarkDeleted(t *testing.T, env *authtest.Env, id uuid.UUID) {
	t.Helper()
	if _, err := env.Pool.Exec(context.Background(),
		"UPDATE character_sheets SET deleted_at = $2 WHERE id = $1", id, env.Clock.Now()); err != nil {
		t.Fatalf("MarkDeleted: %v", err)
	}
}

// Restore は、data-model.md の「削除の取り消し」と同じ SQL で削除を取り消す。
func Restore(t *testing.T, env *authtest.Env, id uuid.UUID) {
	t.Helper()
	tag, err := env.Pool.Exec(context.Background(),
		"UPDATE character_sheets SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL", id)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("Restore: %d 行が変わった（1 行のはず）", tag.RowsAffected())
	}
}

// Row は、削除済みを含めて行を読む。
func Row(t *testing.T, env *authtest.Env, id uuid.UUID) db.CharacterSheet {
	t.Helper()
	var r db.CharacterSheet
	err := env.Pool.QueryRow(context.Background(),
		`SELECT id, system, owner_id, character_name, player_name, age, sex, memo, data, revision,
		        created_at, updated_at, deleted_at
		 FROM character_sheets WHERE id = $1`, id).
		Scan(&r.ID, &r.System, &r.OwnerID, &r.CharacterName, &r.PlayerName, &r.Age, &r.Sex, &r.Memo, &r.Data,
			&r.Revision, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)
	if err != nil {
		t.Fatalf("Row: %v", err)
	}
	return r
}
