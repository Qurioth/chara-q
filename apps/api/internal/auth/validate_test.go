package auth

import (
	"errors"
	"strings"
	"testing"
)

func codeOf(e *InputError) Code {
	if e == nil {
		return ""
	}
	return e.Code
}

func TestValidateLoginID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Code
		norm string
	}{
		{"3 文字", "abc", "", "abc"},
		{"32 文字", strings.Repeat("a", 32), "", strings.Repeat("a", 32)},
		{"英数字・_・-", "Ab_9-z", "", "Ab_9-z"},
		{"前後の空白は取り除く", "  admin ", "", "admin"},
		{"空", "", CodeRequired, ""},
		{"空白だけ", "   ", CodeRequired, ""},
		{"2 文字", "ab", CodeTooShort, "ab"},
		{"33 文字", strings.Repeat("a", 33), CodeTooLong, strings.Repeat("a", 33)},
		{"記号", "ab!c", CodeInvalidCharacters, "ab!c"},
		{"途中の空白", "ab c", CodeInvalidCharacters, "ab c"},
		{"日本語", "あいう", CodeInvalidCharacters, "あいう"},
		{"全角英字", "ａｂｃ", CodeInvalidCharacters, "ａｂｃ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			norm, err := ValidateLoginID(tt.in)
			if got := codeOf(err); got != tt.want {
				t.Errorf("code: got %q, want %q", got, tt.want)
			}
			if norm != tt.norm {
				t.Errorf("norm: got %q, want %q", norm, tt.norm)
			}
			if ValidLoginID(tt.in) != (tt.want == "") {
				t.Error("ValidLoginID が ValidateLoginID と一致しない")
			}
		})
	}
}

func TestLoginIDKey(t *testing.T) {
	if LoginIDKey(" Admin ") != "admin" {
		t.Error("前後の空白を除いて小文字にするはず")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Code
	}{
		{"7 文字", strings.Repeat("a", 7), CodeTooShort},
		{"8 文字", strings.Repeat("a", 8), ""},
		{"128 文字", strings.Repeat("a", 128), ""},
		{"129 文字", strings.Repeat("a", 129), CodeTooLong},
		{"空", "", CodeRequired},
		{"日本語 8 文字", "あいうえおかきく", ""},
		{"空白を含めて 8 文字", " abcdef ", ""},
		{"文字の種類は問わない", "aaaaaaaa", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(FieldPassword, tt.in)
			if got := codeOf(err); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if err != nil && err.Field != FieldPassword {
				t.Errorf("field: got %q", err.Field)
			}
		})
	}
}

func TestValidateDisplayName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Code
		norm string
	}{
		{"1 文字", "a", "", "a"},
		{"日本語 32 文字", strings.Repeat("あ", 32), "", strings.Repeat("あ", 32)},
		{"日本語 33 文字", strings.Repeat("あ", 33), CodeTooLong, strings.Repeat("あ", 33)},
		{"絵文字", "🐱たま", "", "🐱たま"},
		{"前後の空白は取り除く", "　たま ", "", "たま"},
		{"途中の空白はよい", "たま ねこ", "", "たま ねこ"},
		{"空", "", CodeRequired, ""},
		{"空白だけ", " 　 ", CodeRequired, ""},
		{"改行", "たま\nねこ", CodeInvalidCharacters, "たま\nねこ"},
		{"タブ", "たま\tねこ", CodeInvalidCharacters, "たま\tねこ"},
		{"行区切り", "たま\u2028ねこ", CodeInvalidCharacters, "たま\u2028ねこ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			norm, err := ValidateDisplayName(tt.in)
			if got := codeOf(err); got != tt.want {
				t.Errorf("code: got %q, want %q", got, tt.want)
			}
			if norm != tt.norm {
				t.Errorf("norm: got %q, want %q", norm, tt.norm)
			}
		})
	}
}

func TestCollect(t *testing.T) {
	if collect(nil, nil) != nil {
		t.Error("失敗がなければ nil")
	}
	err := collect(&InputError{Field: FieldLoginID, Code: CodeRequired}, nil, &InputError{Field: FieldPassword, Code: CodeTooShort})
	var errs InputErrors
	if !errors.As(err, &errs) || len(errs) != 2 {
		t.Fatalf("2 件の InputErrors のはず: %v", err)
	}
}
