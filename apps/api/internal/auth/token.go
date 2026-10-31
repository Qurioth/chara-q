package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"strings"
	"unicode"
)

// NewToken は招待コード・再設定コード・セッション ID に使うランダムな値を返す
// （128 ビット、base32 の 26 文字。research R7）。
func NewToken() string {
	return rand.Text()
}

// HashToken はトークンの SHA-256 を返す。DB にはこのハッシュだけを保存する（FR-008）。
// 128 ビットの乱数は総当たりできないため、ソルトや遅いハッシュはいらない。
func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// NormalizeCode は、利用者が入力した招待コード・再設定コードを、発行したときの値の形にそろえる。
// 空白とハイフンを取り除き、大文字にする。
func NormalizeCode(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) || r == '-' {
			continue
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

// FormatCode は、コードを読みやすいように 4 文字ごとにハイフンで区切る。
func FormatCode(code string) string {
	var b strings.Builder
	for i, r := range code {
		if i > 0 && i%4 == 0 {
			b.WriteByte('-')
		}
		b.WriteRune(r)
	}
	return b.String()
}
