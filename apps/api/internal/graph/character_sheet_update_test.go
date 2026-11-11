package graph_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

func updateInput(id string, revision int, name string) map[string]any {
	return map[string]any{"input": map[string]any{
		"id": id, "revision": revision, "fields": map[string]any{"characterName": name},
	}}
}

func TestUpdateCharacterSheet(t *testing.T) {
	env := authtest.New(t)
	owner, _ := loginAs(t, env, "owner", false)
	admin, _ := loginAs(t, env, "admin", true)
	anon := newClient(t, env)
	s := createSheet(t, owner, "YUYAKEKOYAKE", "元の名前")

	resp := owner.do(updateSheetMutation, updateInput(s.ID, 1, "新しい名前"))
	p := resp.sheetPayload(t, "updateCharacterSheet")
	if p.CharacterSheet == nil || p.CharacterSheet.CharacterName != "新しい名前" || p.CharacterSheet.Revision != 2 {
		t.Fatalf("所有者は更新できるはず: %s", resp.raw)
	}

	// 古い revision は REVISION_CONFLICT（field は null）
	resp = owner.do(updateSheetMutation, updateInput(s.ID, 1, "古い内容から"))
	p = resp.sheetPayload(t, "updateCharacterSheet")
	if p.CharacterSheet != nil || len(p.Errors) != 1 || p.Errors[0].Code != "REVISION_CONFLICT" || p.Errors[0].Field != nil {
		t.Fatalf("REVISION_CONFLICT のはず: %s", resp.raw)
	}

	if code := admin.do(updateSheetMutation, updateInput(s.ID, 2, "管理者")).errorCode(); code != "FORBIDDEN" {
		t.Errorf("所有者でない管理者は FORBIDDEN のはず: %s", code)
	}
	if code := anon.do(updateSheetMutation, updateInput(s.ID, 2, "匿名")).errorCode(); code != "UNAUTHENTICATED" {
		t.Errorf("ログインなしは UNAUTHENTICATED のはず: %s", code)
	}
	if code := owner.do(updateSheetMutation, updateInput(uuid.NewString(), 1, "x")).errorCode(); code != "NOT_FOUND" {
		t.Errorf("存在しないシートは NOT_FOUND のはず: %s", code)
	}
	if got := sheettest.Row(t, env, uuid.MustParse(s.ID)); got.CharacterName != "新しい名前" || got.Revision != 2 {
		t.Errorf("失敗した更新で内容が変わってはいけない: %+v", got)
	}
}
