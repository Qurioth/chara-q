package sheet

import (
	"errors"
	"strings"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
)

func validFields() Fields {
	return Fields{CharacterName: "名前", Data: []byte(`{}`)}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		modify func(f *Fields)
		// 失敗しないなら空
		field auth.Field
		code  auth.Code
	}{
		{"最小の入力", func(*Fields) {}, "", ""},
		{"キャラクター名が空", func(f *Fields) { f.CharacterName = "" }, FieldCharacterName, auth.CodeRequired},
		{"キャラクター名が空白だけ", func(f *Fields) { f.CharacterName = " 　\t" }, FieldCharacterName, auth.CodeRequired},
		{"キャラクター名が 1 文字", func(f *Fields) { f.CharacterName = "あ" }, "", ""},
		{"キャラクター名が 100 文字", func(f *Fields) { f.CharacterName = strings.Repeat("あ", 100) }, "", ""},
		{"キャラクター名が 101 文字", func(f *Fields) { f.CharacterName = strings.Repeat("あ", 101) }, FieldCharacterName, auth.CodeTooLong},
		{"キャラクター名の前後の空白は上限に数えない", func(f *Fields) { f.CharacterName = "  " + strings.Repeat("a", 100) + "  " }, "", ""},
		{"キャラクター名に改行", func(f *Fields) { f.CharacterName = "あ\nい" }, FieldCharacterName, auth.CodeInvalidCharacters},
		{"キャラクター名にタブ", func(f *Fields) { f.CharacterName = "あ\tい" }, FieldCharacterName, auth.CodeInvalidCharacters},
		{"キャラクター名に U+2028", func(f *Fields) { f.CharacterName = "あ い" }, FieldCharacterName, auth.CodeInvalidCharacters},
		{"キャラクター名に絵文字", func(f *Fields) { f.CharacterName = "🦊きつね" }, "", ""},
		{"長すぎて制御文字もあるときは TOO_LONG", func(f *Fields) { f.CharacterName = strings.Repeat("あ", 101) + "\n" }, FieldCharacterName, auth.CodeTooLong},
		{"プレイヤー名が 50 文字", func(f *Fields) { f.PlayerName = strings.Repeat("a", 50) }, "", ""},
		{"プレイヤー名が 51 文字", func(f *Fields) { f.PlayerName = strings.Repeat("a", 51) }, FieldPlayerName, auth.CodeTooLong},
		{"プレイヤー名に制御文字", func(f *Fields) { f.PlayerName = "a\x00b" }, FieldPlayerName, auth.CodeInvalidCharacters},
		{"年齢が 20 文字", func(f *Fields) { f.Age = strings.Repeat("1", 20) }, "", ""},
		{"年齢が 21 文字", func(f *Fields) { f.Age = strings.Repeat("1", 21) }, FieldAge, auth.CodeTooLong},
		{"性別が 21 文字", func(f *Fields) { f.Sex = strings.Repeat("女", 21) }, FieldSex, auth.CodeTooLong},
		{"メモが 10,000 文字", func(f *Fields) { f.Memo = strings.Repeat("あ", 10000) }, "", ""},
		{"メモが 10,001 文字", func(f *Fields) { f.Memo = strings.Repeat("あ", 10001) }, FieldMemo, auth.CodeTooLong},
		{"メモに改行とタブ", func(f *Fields) { f.Memo = "1 行目\n\t2 行目" }, "", ""},
		{"メモに改行以外の制御文字", func(f *Fields) { f.Memo = "あ\x07い" }, FieldMemo, auth.CodeInvalidCharacters},
		{"data が配列", func(f *Fields) { f.Data = []byte(`[]`) }, FieldData, CodeNotObject},
		{"data が文字列", func(f *Fields) { f.Data = []byte(`"a"`) }, FieldData, CodeNotObject},
		{"data が null", func(f *Fields) { f.Data = []byte(`null`) }, FieldData, CodeNotObject},
		{"data が空", func(f *Fields) { f.Data = nil }, FieldData, CodeNotObject},
		{"data がちょうど 256KB", func(f *Fields) { f.Data = objectOfSize(DataMaxBytes) }, "", ""},
		{"data が 256KB を 1 バイト超える", func(f *Fields) { f.Data = objectOfSize(DataMaxBytes + 1) }, FieldData, CodeTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := validFields()
			tt.modify(&f)
			_, err := Validate(f)
			if tt.field == "" {
				if err != nil {
					t.Fatalf("失敗しないはず: %v", err)
				}
				return
			}
			var ie auth.InputErrors
			if !errors.As(err, &ie) {
				t.Fatalf("InputErrors のはず: %v", err)
			}
			if len(ie) != 1 || ie[0].Field != tt.field || ie[0].Code != tt.code {
				t.Fatalf("got %v, want %s:%s", ie, tt.field, tt.code)
			}
		})
	}
}

func TestValidateNormalizes(t *testing.T) {
	got, err := Validate(Fields{
		CharacterName: "  名前  ",
		PlayerName:    " プレイヤー ",
		Age:           " 17 歳 ",
		Sex:           " 女 ",
		Memo:          "  1 行目\r\n2 行目\r3 行目  ",
		Data:          []byte(`{"a":1}`),
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	want := Fields{
		CharacterName: "名前",
		PlayerName:    "プレイヤー",
		Age:           "17 歳",
		Sex:           "女",
		// メモの前後の空白は取り除かず、改行だけをそろえる
		Memo: "  1 行目\n2 行目\n3 行目  ",
		Data: []byte(`{"a":1}`),
	}
	if got.CharacterName != want.CharacterName || got.PlayerName != want.PlayerName || got.Age != want.Age ||
		got.Sex != want.Sex || got.Memo != want.Memo || string(got.Data) != string(want.Data) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestValidateReportsEachField(t *testing.T) {
	_, err := Validate(Fields{
		CharacterName: "",
		PlayerName:    strings.Repeat("a", 51),
		Data:          []byte(`[]`),
	})
	var ie auth.InputErrors
	if !errors.As(err, &ie) {
		t.Fatalf("InputErrors のはず: %v", err)
	}
	if len(ie) != 3 {
		t.Fatalf("項目ごとに 1 つずつ、3 つのはず: %v", ie)
	}
}

// objectOfSize は、ちょうど n バイトの JSON のオブジェクトを返す。
func objectOfSize(n int) []byte {
	const prefix, suffix = `{"a":"`, `"}`
	return []byte(prefix + strings.Repeat("x", n-len(prefix)-len(suffix)) + suffix)
}
