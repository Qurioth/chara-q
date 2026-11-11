package auth

import (
	"context"
	"errors"
	"net/http"
)

// SessionCookieName はセッションの Cookie の名前。
// __Host- 接頭辞で、ブラウザに Secure・Path=/・Domain なしを強制させる（research R2）。
const SessionCookieName = "__Host-chara_q_session"

// sessionCookieMaxAge は Cookie の寿命（400 日。Chrome の上限）。
// 有効かどうかは DB のセッションの期限で決める。Server Components から api を呼んだ応答の
// Set-Cookie はブラウザに届かないため、Cookie を DB の期限に合わせて付け直すことはできない（research R2）。
const sessionCookieMaxAge = 400 * 24 * 60 * 60

var errNoResponseWriter = errors.New("セッションの Cookie を書けません（認証のミドルウェアを通っていません）")

func newSessionCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

// SetSessionCookie はセッション ID の Cookie を応答に付ける。
func SetSessionCookie(ctx context.Context, token string) error {
	st := stateFrom(ctx)
	if st == nil || st.w == nil {
		return errNoResponseWriter
	}
	http.SetCookie(st.w, newSessionCookie(token, sessionCookieMaxAge))
	return nil
}

// ClearSessionCookie はセッションの Cookie を消す応答を付ける。
func ClearSessionCookie(ctx context.Context) error {
	st := stateFrom(ctx)
	if st == nil || st.w == nil {
		return errNoResponseWriter
	}
	clearSessionCookie(st.w)
	return nil
}

func clearSessionCookie(w http.ResponseWriter) {
	// MaxAge が負なら、Go は Max-Age=0 を書く
	http.SetCookie(w, newSessionCookie("", -1))
}
