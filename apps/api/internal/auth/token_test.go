package auth

import (
	"bytes"
	"testing"
)

func TestNewToken(t *testing.T) {
	a, b := NewToken(), NewToken()
	if len(a) != 26 {
		t.Errorf("26 文字のはず: %d", len(a))
	}
	if a == b {
		t.Error("毎回違う値のはず")
	}
	if NormalizeCode(a) != a {
		t.Error("生成した値は正規化しても変わらないはず")
	}
}

func TestHashToken(t *testing.T) {
	if len(HashToken("x")) != 32 {
		t.Error("SHA-256 は 32 バイト")
	}
	if !bytes.Equal(HashToken("ABC"), HashToken("ABC")) {
		t.Error("同じ値は同じハッシュ")
	}
	if bytes.Equal(HashToken("ABC"), HashToken("ABD")) {
		t.Error("違う値は違うハッシュ")
	}
}

func TestNormalizeCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"そのまま", "ABCDEFGH", "ABCDEFGH"},
		{"小文字", "abcdefgh", "ABCDEFGH"},
		{"ハイフン区切り", "ABCD-EFGH", "ABCDEFGH"},
		{"前後と途中の空白", "  abcd efgh\t", "ABCDEFGH"},
		{"全角の空白", "ABCD　EFGH", "ABCDEFGH"},
		{"空", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeCode(tt.in); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCode(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"ABC", "ABC"},
		{"ABCD", "ABCD"},
		{"ABCDE", "ABCD-E"},
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ", "ABCD-EFGH-IJKL-MNOP-QRST-UVWX-YZ"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := FormatCode(tt.in)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if NormalizeCode(got) != tt.in {
				t.Errorf("NormalizeCode(FormatCode(%q)) が元に戻らない", tt.in)
			}
		})
	}
}
