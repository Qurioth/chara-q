package graph_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

func getSheet(t *testing.T, c *client, id string) (*sheetJSON, gqlResponse) {
	t.Helper()
	resp := c.do(getSheetQuery, map[string]any{"id": id})
	if len(resp.Errors) != 0 {
		t.Fatalf("エラーにしないはず: %s", resp.raw)
	}
	var s *sheetJSON
	if err := json.Unmarshal(resp.Data["characterSheet"], &s); err != nil {
		t.Fatalf("json: %v", err)
	}
	return s, resp
}

func TestCharacterSheetQuery(t *testing.T) {
	env := authtest.New(t)
	owner, u := loginAs(t, env, "owner-login", false)
	created := createSheet(t, owner, "YUYAKEKOYAKE", "きつね")
	anon := newClient(t, env)

	s, resp := getSheet(t, anon, created.ID)
	if s == nil || s.CharacterName != "きつね" || s.Owner.DisplayName != u.DisplayName {
		t.Fatalf("ログインなしで読めるはず: %s", resp.raw)
	}
	assertNoLoginID(t, resp, "owner-login")

	if s, _ := getSheet(t, anon, strings.ToUpper(created.ID)); s == nil || s.ID != created.ID {
		t.Errorf("大文字の UUID でも同じシートのはず: %+v", s)
	}

	// 所有者が無効にされても、同じように読める
	if _, err := env.Queries.DisableUser(context.Background(), db.DisableUserParams{ID: u.ID, Now: env.Clock.Now()}); err != nil {
		t.Fatalf("DisableUser: %v", err)
	}
	if s, _ := getSheet(t, anon, created.ID); s == nil || s.Owner.DisplayName != u.DisplayName {
		t.Errorf("無効な所有者のシートも読めるはず: %+v", s)
	}

	deleted := sheettest.CreateSheet(t, env, u.ID, "yuyakekoyake", "削除済み", "")
	sheettest.MarkDeleted(t, env, deleted.ID)
	for _, id := range []string{uuid.NewString(), "abc", deleted.ID.String()} {
		if s, _ := getSheet(t, anon, id); s != nil {
			t.Errorf("%s: null のはず: %+v", id, s)
		}
	}
}
