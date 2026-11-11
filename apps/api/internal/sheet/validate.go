package sheet

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
)

// 入力の失敗の項目。値は GraphQL の InputField と同じ。失敗は auth.InputErrors で返し、
// resolver の変換（internal/graph の inputErrors）を認証と共通にする（research R11）。
const (
	FieldCharacterName auth.Field = "CHARACTER_NAME"
	FieldPlayerName    auth.Field = "PLAYER_NAME"
	FieldAge           auth.Field = "AGE"
	FieldSex           auth.Field = "SEX"
	FieldMemo          auth.Field = "MEMO"
	FieldData          auth.Field = "DATA"
)

// 入力の失敗の種類のうち、シートで足したもの。値は GraphQL の InputErrorCode と同じ。
const (
	CodeNotObject        auth.Code = "NOT_OBJECT"
	CodeTooLarge         auth.Code = "TOO_LARGE"
	CodeRevisionConflict auth.Code = "REVISION_CONFLICT"
)

// 共通の項目の規則（spec の FR-006、FR-007）。文字数はコードポイントで数える（Postgres の char_length と同じ）。
const (
	CharacterNameMaxLen = 100
	PlayerNameMaxLen    = 50
	AgeMaxLen           = 20
	SexMaxLen           = 20
	MemoMaxLen          = 10000
	DataMaxBytes        = 256 * 1024
)

// Fields は、作成と更新で共通の入力。Data は JSON の値そのもの。
type Fields struct {
	CharacterName string
	PlayerName    string
	Age           string
	Sex           string
	Memo          string
	Data          []byte
}

// Validate は入力を確かめ、保存に使う値（前後の空白を除いた文字列、\n にそろえたメモ）を返す。
// 1 つの項目に複数の失敗があっても、返すのは最初の 1 つ（REQUIRED → TOO_LONG → INVALID_CHARACTERS）。
func Validate(f Fields) (Fields, error) {
	var errs auth.InputErrors
	add := func(field auth.Field, code auth.Code) {
		errs = append(errs, auth.InputError{Field: field, Code: code})
	}

	out := Fields{
		CharacterName: strings.TrimSpace(f.CharacterName),
		PlayerName:    strings.TrimSpace(f.PlayerName),
		Age:           strings.TrimSpace(f.Age),
		Sex:           strings.TrimSpace(f.Sex),
		Memo:          normalizeNewlines(f.Memo),
		Data:          f.Data,
	}

	if out.CharacterName == "" {
		add(FieldCharacterName, auth.CodeRequired)
	} else if code, ok := checkLine(out.CharacterName, CharacterNameMaxLen); !ok {
		add(FieldCharacterName, code)
	}
	for _, item := range []struct {
		field auth.Field
		value string
		max   int
	}{
		{FieldPlayerName, out.PlayerName, PlayerNameMaxLen},
		{FieldAge, out.Age, AgeMaxLen},
		{FieldSex, out.Sex, SexMaxLen},
	} {
		if code, ok := checkLine(item.value, item.max); !ok {
			add(item.field, code)
		}
	}
	if code, ok := checkMemo(out.Memo); !ok {
		add(FieldMemo, code)
	}
	if code, ok := checkData(out.Data); !ok {
		add(FieldData, code)
	}

	if len(errs) > 0 {
		return out, errs
	}
	return out, nil
}

// checkLine は 1 行の文字列（キャラクター名・プレイヤー名・年齢・性別）を確かめる。
func checkLine(s string, maxLen int) (auth.Code, bool) {
	switch {
	case utf8.RuneCountInString(s) > maxLen:
		return auth.CodeTooLong, false
	case strings.IndexFunc(s, isForbiddenInLine) >= 0:
		return auth.CodeInvalidCharacters, false
	}
	return "", true
}

func checkMemo(s string) (auth.Code, bool) {
	switch {
	case utf8.RuneCountInString(s) > MemoMaxLen:
		return auth.CodeTooLong, false
	case strings.IndexFunc(s, isForbiddenInMemo) >= 0:
		return auth.CodeInvalidCharacters, false
	}
	return "", true
}

func checkData(b []byte) (auth.Code, bool) {
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		return CodeNotObject, false
	}
	if len(b) > DataMaxBytes {
		return CodeTooLarge, false
	}
	return "", true
}

// normalizeNewlines は \r\n と \r を \n にそろえる。
func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func isForbiddenInLine(r rune) bool {
	return unicode.IsControl(r) || r == ' ' || r == ' '
}

// isForbiddenInMemo は、改行とタブ以外の制御文字を拒む（DB の CHECK 制約と同じ）。
func isForbiddenInMemo(r rune) bool {
	if r == '\n' || r == '\t' {
		return false
	}
	return isForbiddenInLine(r)
}
