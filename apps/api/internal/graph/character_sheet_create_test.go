package graph_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

func TestCreateCharacterSheet(t *testing.T) {
	env := authtest.New(t)
	c, u := loginAs(t, env, "creator-login", false)

	for _, tt := range []struct{ system, typename string }{
		{"YUYAKEKOYAKE", "YuyakekoyakeSheet"},
		{"BEGINNING_IDOL", "BeginningIdolSheet"},
	} {
		resp := c.do(createSheetMutation, map[string]any{"input": map[string]any{
			"system": tt.system,
			"fields": map[string]any{"characterName": "名前", "data": map[string]any{"a": 1}},
		}})
		p := resp.sheetPayload(t, "createCharacterSheet")
		if p.CharacterSheet == nil || len(p.Errors) != 0 {
			t.Fatalf("%s: 作れるはず: %s", tt.system, resp.raw)
		}
		s := p.CharacterSheet
		if s.Typename != tt.typename || s.System != tt.system || s.Revision != 1 {
			t.Errorf("型・システム・revision: %+v", s)
		}
		if s.Owner.ID != u.ID.String() || s.Owner.DisplayName != u.DisplayName {
			t.Errorf("所有者は本人のはず: %+v", s.Owner)
		}
		var data map[string]any
		if err := json.Unmarshal(s.Data, &data); err != nil || len(data) != 1 || data["a"] != float64(1) {
			t.Errorf("data は入力と同じ JSON のはず: %s", s.Data)
		}
		assertNoLoginID(t, resp, "creator-login")
	}

	// data を省くと {} になる
	resp := c.do(createSheetMutation, map[string]any{"input": sheetInput("YUYAKEKOYAKE", "data なし")})
	if p := resp.sheetPayload(t, "createCharacterSheet"); p.CharacterSheet == nil || string(p.CharacterSheet.Data) != `{}` {
		t.Errorf("data の既定値は {} のはず: %s", resp.raw)
	}
}

func TestCreateCharacterSheetFailures(t *testing.T) {
	env := authtest.New(t)
	c, _ := loginAs(t, env, "creator", false)

	tests := []struct {
		name   string
		fields map[string]any
		field  string
		code   string
	}{
		{"キャラクター名が空", map[string]any{"characterName": "  "}, "CHARACTER_NAME", "REQUIRED"},
		{"data が配列", map[string]any{"characterName": "a", "data": []any{1}}, "DATA", "NOT_OBJECT"},
		{"プレイヤー名が長すぎる", map[string]any{"characterName": "a", "playerName": strings.Repeat("a", 51)}, "PLAYER_NAME", "TOO_LONG"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := c.do(createSheetMutation, map[string]any{"input": map[string]any{"system": "YUYAKEKOYAKE", "fields": tt.fields}})
			p := resp.sheetPayload(t, "createCharacterSheet")
			if p.CharacterSheet != nil || len(p.Errors) != 1 || p.Errors[0].Field == nil || *p.Errors[0].Field != tt.field {
				t.Fatalf("%s の失敗のはず: %s", tt.field, resp.raw)
			}
			if tt.code != "" && p.Errors[0].Code != tt.code {
				t.Fatalf("code: got %s, want %s", p.Errors[0].Code, tt.code)
			}
		})
	}

	anon := newClient(t, env)
	if code := anon.do(createSheetMutation, map[string]any{"input": sheetInput("YUYAKEKOYAKE", "a")}).errorCode(); code != "UNAUTHENTICATED" {
		t.Errorf("ログインなしは UNAUTHENTICATED のはず: %s", code)
	}
	if n := env.Count(t, "character_sheets"); n != 0 {
		t.Errorf("行が作られてはいけない: %d", n)
	}
}
