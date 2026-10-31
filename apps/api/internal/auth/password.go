package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id のパラメーター（OWASP の推奨の最低値。research R6）。
// 照合は保存された文字列のパラメーターで行うため、後で値を変えても古いハッシュを照合できる。
const (
	argonMemoryKiB = 19456
	argonTime      = 2
	argonThreads   = 1
	argonSaltLen   = 16
	argonKeyLen    = 32
)

var errInvalidHash = errors.New("パスワードのハッシュの形式が不正です")

// HashPassword はパスワードを Argon2id でハッシュにし、PHC 形式の文字列で返す。
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("ソルトを作れません: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemoryKiB, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword はパスワードが PHC 形式のハッシュと一致するかを返す。
func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	// "", "argon2id", "v=19", "m=…,t=…,p=…", salt, hash
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, errInvalidHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, errInvalidHash
	}
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, errInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, errInvalidHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) == 0 {
		return false, errInvalidHash
	}
	got := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// dummyHash は、存在しないログイン ID でも照合と同じ時間をかけるための捨てのハッシュ。
var dummyHash = mustHashPassword("chara-q-dummy-password")

func mustHashPassword(password string) string {
	h, err := HashPassword(password)
	if err != nil {
		panic(err)
	}
	return h
}

// VerifyDummy は捨てのハッシュで照合を 1 回行う。応答時間からログイン ID の有無がわからないようにする（FR-018）。
func VerifyDummy(password string) {
	_, _ = VerifyPassword(password, dummyHash)
}
