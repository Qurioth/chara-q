package graph_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

func TestCopyCharacterSheet(t *testing.T) {
	env := authtest.New(t)
	alice, _ := loginAs(t, env, "alice", false)
	bob, b := loginAs(t, env, "bob", false)
	anon := newClient(t, env)
	s := createSheet(t, alice, "BEGINNING_IDOL", "アイドル")

	resp := bob.do(copySheetMutation, map[string]any{"id": s.ID})
	p := resp.sheetPayload(t, "copyCharacterSheet")
	if p.CharacterSheet == nil || p.CharacterSheet.Owner.ID != b.ID.String() || p.CharacterSheet.CharacterName != "アイドル（コピー）" ||
		p.CharacterSheet.Typename != "BeginningIdolSheet" {
		t.Fatalf("複製は複製した人の所有・名前の末尾に（コピー）のはず: %s", resp.raw)
	}

	if code := anon.do(copySheetMutation, map[string]any{"id": s.ID}).errorCode(); code != "UNAUTHENTICATED" {
		t.Errorf("ログインなしは UNAUTHENTICATED のはず: %s", code)
	}
	sheettest.MarkDeleted(t, env, uuid.MustParse(s.ID))
	for _, id := range []string{s.ID, uuid.NewString()} {
		if code := bob.do(copySheetMutation, map[string]any{"id": id}).errorCode(); code != "NOT_FOUND" {
			t.Errorf("%s: NOT_FOUND のはず: %s", id, code)
		}
	}
}
