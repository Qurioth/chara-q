package auth

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Field は入力の失敗がどの項目のものかを表す。値は GraphQL の InputField と同じ。
type Field string

// Field の値。FieldNone はフォーム全体の失敗。
const (
	FieldNone            Field = ""
	FieldInvitationCode  Field = "INVITATION_CODE"
	FieldLoginID         Field = "LOGIN_ID"
	FieldPassword        Field = "PASSWORD"
	FieldCurrentPassword Field = "CURRENT_PASSWORD"
	FieldNewPassword     Field = "NEW_PASSWORD"
	FieldDisplayName     Field = "DISPLAY_NAME"
	FieldResetCode       Field = "RESET_CODE"
)

// Code は入力の失敗の種類を表す。値は GraphQL の InputErrorCode と同じ。
type Code string

// Code の値。
const (
	CodeRequired              Code = "REQUIRED"
	CodeTooShort              Code = "TOO_SHORT"
	CodeTooLong               Code = "TOO_LONG"
	CodeInvalidCharacters     Code = "INVALID_CHARACTERS"
	CodeInvitationUnavailable Code = "INVITATION_UNAVAILABLE"
	CodeLoginIDTaken          Code = "LOGIN_ID_TAKEN"
	CodeInvalidCredentials    Code = "INVALID_CREDENTIALS"
	CodeLocked                Code = "LOCKED"
	CodeWrongCurrentPassword  Code = "WRONG_CURRENT_PASSWORD"
	CodeResetCodeUnavailable  Code = "RESET_CODE_UNAVAILABLE"
)

// InputError は、利用者が入力を直せば解決する失敗。
type InputError struct {
	Field Field
	Code  Code
}

// InputErrors は 1 つ以上の InputError。error として返す。
type InputErrors []InputError

func (e InputErrors) Error() string {
	parts := make([]string, len(e))
	for i, ie := range e {
		parts[i] = string(ie.Field) + ":" + string(ie.Code)
	}
	return "入力の失敗: " + strings.Join(parts, ", ")
}

func fail(field Field, code Code) InputErrors {
	return InputErrors{{Field: field, Code: code}}
}

// 入力の規則（Clarifications Q1、research R11）。web の src/lib/auth/rules.ts と同じにする。
const (
	LoginIDMinLen     = 3
	LoginIDMaxLen     = 32
	PasswordMinLen    = 8
	PasswordMaxLen    = 128
	DisplayNameMaxLen = 32
)

var loginIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// NormalizeLoginID は前後の空白を取り除く。
func NormalizeLoginID(s string) string {
	return strings.TrimSpace(s)
}

// LoginIDKey は、大文字と小文字を区別しない比較やロックの記録に使う値を返す。
func LoginIDKey(s string) string {
	return strings.ToLower(NormalizeLoginID(s))
}

// ValidLoginID はログイン ID が規則に合うかを返す（失敗の種類は問わない）。
func ValidLoginID(s string) bool {
	_, ok := checkLoginID(NormalizeLoginID(s))
	return ok
}

// ValidateLoginID は前後の空白を除いたログイン ID が規則に合うかを確かめる。
func ValidateLoginID(s string) (string, *InputError) {
	s = NormalizeLoginID(s)
	if code, ok := checkLoginID(s); !ok {
		return s, &InputError{Field: FieldLoginID, Code: code}
	}
	return s, nil
}

func checkLoginID(s string) (Code, bool) {
	switch {
	case s == "":
		return CodeRequired, false
	case !loginIDPattern.MatchString(s):
		return CodeInvalidCharacters, false
	case len(s) < LoginIDMinLen:
		return CodeTooShort, false
	case len(s) > LoginIDMaxLen:
		return CodeTooLong, false
	}
	return "", true
}

// ValidatePassword はパスワードが規則に合うかを確かめる。前後の空白は取り除かない。
func ValidatePassword(field Field, s string) *InputError {
	n := utf8.RuneCountInString(s)
	switch {
	case s == "":
		return &InputError{Field: field, Code: CodeRequired}
	case n < PasswordMinLen:
		return &InputError{Field: field, Code: CodeTooShort}
	case n > PasswordMaxLen:
		return &InputError{Field: field, Code: CodeTooLong}
	}
	return nil
}

// ValidateDisplayName は前後の空白を除いた表示名が規則に合うかを確かめる。
func ValidateDisplayName(s string) (string, *InputError) {
	s = strings.TrimSpace(s)
	switch {
	case s == "":
		return s, &InputError{Field: FieldDisplayName, Code: CodeRequired}
	case strings.IndexFunc(s, isForbiddenInDisplayName) >= 0:
		return s, &InputError{Field: FieldDisplayName, Code: CodeInvalidCharacters}
	case utf8.RuneCountInString(s) > DisplayNameMaxLen:
		return s, &InputError{Field: FieldDisplayName, Code: CodeTooLong}
	}
	return s, nil
}

func isForbiddenInDisplayName(r rune) bool {
	return unicode.IsControl(r) || r == '\u2028' || r == '\u2029'
}

// collect は nil でない失敗をまとめる。失敗がなければ nil を返す。
func collect(errs ...*InputError) error {
	var out InputErrors
	for _, e := range errs {
		if e != nil {
			out = append(out, *e)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
