package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("correct horse")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Errorf("PHC 形式になっていない: %s", hash)
	}
	other, err := HashPassword("correct horse")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == other {
		t.Error("同じパスワードでもソルトが違えばハッシュが違うはず")
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"同じパスワード", "correct horse", true},
		{"違うパスワード", "correct horses", false},
		{"空", "", false},
		{"前後の空白は別のパスワード", " correct horse", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyPassword(tt.password, hash)
			if err != nil {
				t.Fatalf("VerifyPassword: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
	}{
		{"空", ""},
		{"別のアルゴリズム", "$argon2i$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"},
		{"バージョン違い", "$argon2id$v=16$m=19456,t=2,p=1$c2FsdA$aGFzaA"},
		{"パラメーターが壊れている", "$argon2id$v=19$m=x$c2FsdA$aGFzaA"},
		{"base64 が壊れている", "$argon2id$v=19$m=19456,t=2,p=1$!!!$aGFzaA"},
		{"区切りが足りない", "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := VerifyPassword("x", tt.encoded); err == nil {
				t.Error("エラーになるはず")
			}
		})
	}
}
