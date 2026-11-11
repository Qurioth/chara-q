package graph_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// シートの GraphQL のテストで共通に使うクエリと補助。

const sheetFields = `
  __typename id system characterName playerName age sex memo revision createdAt updatedAt
  owner { id displayName }
  ... on YuyakekoyakeSheet { data }
  ... on BeginningIdolSheet { data }
`

const createSheetMutation = `mutation($input: CreateCharacterSheetInput!) {
  createCharacterSheet(input: $input) {
    characterSheet {` + sheetFields + `}
    errors { field code }
  }
}`

const updateSheetMutation = `mutation($input: UpdateCharacterSheetInput!) {
  updateCharacterSheet(input: $input) {
    characterSheet {` + sheetFields + `}
    errors { field code }
  }
}`

const copySheetMutation = `mutation($id: ID!) {
  copyCharacterSheet(id: $id) {
    characterSheet {` + sheetFields + `}
    errors { field code }
  }
}`

const deleteSheetMutation = `mutation($id: ID!) { deleteCharacterSheet(id: $id) { deletedId } }`

const getSheetQuery = `query($id: ID!) { characterSheet(id: $id) {` + sheetFields + `} }`

type sheetJSON struct {
	Typename      string          `json:"__typename"`
	ID            string          `json:"id"`
	System        string          `json:"system"`
	CharacterName string          `json:"characterName"`
	PlayerName    string          `json:"playerName"`
	Revision      int             `json:"revision"`
	Data          json.RawMessage `json:"data"`
	Owner         struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	} `json:"owner"`
}

type inputErrorJSON struct {
	Field *string `json:"field"`
	Code  string  `json:"code"`
}

type sheetPayloadJSON struct {
	CharacterSheet *sheetJSON       `json:"characterSheet"`
	Errors         []inputErrorJSON `json:"errors"`
}

// payload は、data の key の payload を読む。
func (r gqlResponse) sheetPayload(t *testing.T, key string) sheetPayloadJSON {
	t.Helper()
	var p sheetPayloadJSON
	if err := json.Unmarshal(r.Data[key], &p); err != nil {
		t.Fatalf("%s: %v: %s", key, err, r.raw)
	}
	return p
}

func sheetInput(system, name string) map[string]any {
	return map[string]any{"system": system, "fields": map[string]any{"characterName": name}}
}

// createSheet は、c のユーザーでシートを作り、ID を返す。
func createSheet(t *testing.T, c *client, system, name string) sheetJSON {
	t.Helper()
	resp := c.do(createSheetMutation, map[string]any{"input": sheetInput(system, name)})
	p := resp.sheetPayload(t, "createCharacterSheet")
	if p.CharacterSheet == nil {
		t.Fatalf("シートを作れません: %s", resp.raw)
	}
	return *p.CharacterSheet
}

// assertNoLoginID は、応答の本文にログイン ID が含まれないことを確かめる（SC-005）。
func assertNoLoginID(t *testing.T, resp gqlResponse, loginIDs ...string) {
	t.Helper()
	for _, id := range loginIDs {
		if strings.Contains(resp.raw, id) {
			t.Errorf("応答にログイン ID %q が含まれる: %s", id, resp.raw)
		}
	}
}
