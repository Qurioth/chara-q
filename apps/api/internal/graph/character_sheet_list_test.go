package graph_test

import (
	"encoding/json"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
)

const listSheetsQuery = `query($filter: CharacterSheetFilter, $orderBy: CharacterSheetOrder!, $page: Int!) {
  characterSheets(filter: $filter, orderBy: $orderBy, page: $page) {
    nodes { id characterName owner { id displayName } }
    totalCount page pageSize totalPages
  }
}`

type sheetPageJSON struct {
	Nodes []struct {
		ID            string `json:"id"`
		CharacterName string `json:"characterName"`
		Owner         struct {
			ID string `json:"id"`
		} `json:"owner"`
	} `json:"nodes"`
	TotalCount int `json:"totalCount"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

func listSheets(t *testing.T, c *client, filter map[string]any) (sheetPageJSON, gqlResponse) {
	t.Helper()
	resp := c.do(listSheetsQuery, map[string]any{"filter": filter, "orderBy": "UPDATED_AT_DESC", "page": 1})
	if len(resp.Errors) != 0 {
		t.Fatalf("エラー: %s", resp.raw)
	}
	var p sheetPageJSON
	if err := json.Unmarshal(resp.Data["characterSheets"], &p); err != nil {
		t.Fatalf("json: %v", err)
	}
	return p, resp
}

func TestCharacterSheetsQuery(t *testing.T) {
	env := authtest.New(t)
	alice, a := loginAs(t, env, "alice-login", false)
	bob, _ := loginAs(t, env, "bob-login", false)
	createSheet(t, alice, "YUYAKEKOYAKE", "アリスのシート")
	createSheet(t, bob, "BEGINNING_IDOL", "ボブのシート")
	anon := newClient(t, env)

	p, resp := listSheets(t, anon, nil)
	if p.TotalCount != 2 || len(p.Nodes) != 2 || p.PageSize != 20 || p.Page != 1 || p.TotalPages != 1 {
		t.Fatalf("ログインなしで一覧が読めるはず: %s", resp.raw)
	}
	assertNoLoginID(t, resp, "alice-login", "bob-login")

	if p, _ := listSheets(t, anon, map[string]any{"mine": true}); p.TotalCount != 2 {
		t.Errorf("ログインなしでは mine を無視するはず: %d", p.TotalCount)
	}
	p, _ = listSheets(t, alice, map[string]any{"mine": true})
	if p.TotalCount != 1 || p.Nodes[0].Owner.ID != a.ID.String() {
		t.Errorf("mine は本人のシートだけのはず: %+v", p)
	}
	if p, _ := listSheets(t, anon, map[string]any{"system": "BEGINNING_IDOL", "keyword": "ボブ"}); p.TotalCount != 1 {
		t.Errorf("システムとキーワードの組み合わせ: %+v", p)
	}
}
