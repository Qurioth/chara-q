package graph_test

import (
	"encoding/json"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func TestDeleteCharacterSheet(t *testing.T) {
	env := authtest.New(t)
	owner, _ := loginAs(t, env, "owner", false)
	admin, _ := loginAs(t, env, "admin", true)
	anon := newClient(t, env)
	s := createSheet(t, owner, "YUYAKEKOYAKE", "消すシート")

	if code := admin.do(deleteSheetMutation, map[string]any{"id": s.ID}).errorCode(); code != "FORBIDDEN" {
		t.Errorf("所有者でない管理者は FORBIDDEN のはず: %s", code)
	}
	if code := anon.do(deleteSheetMutation, map[string]any{"id": s.ID}).errorCode(); code != "UNAUTHENTICATED" {
		t.Errorf("ログインなしは UNAUTHENTICATED のはず: %s", code)
	}

	resp := owner.do(deleteSheetMutation, map[string]any{"id": s.ID})
	var p struct {
		DeletedID string `json:"deletedId"`
	}
	if err := json.Unmarshal(resp.Data["deleteCharacterSheet"], &p); err != nil || p.DeletedID != s.ID {
		t.Fatalf("所有者は削除できるはず: %s", resp.raw)
	}
	if got, _ := getSheet(t, anon, s.ID); got != nil {
		t.Errorf("削除したシートは null のはず: %+v", got)
	}
	if code := owner.do(deleteSheetMutation, map[string]any{"id": s.ID}).errorCode(); code != "NOT_FOUND" {
		t.Errorf("もう一度の削除は NOT_FOUND のはず: %s", code)
	}
}
