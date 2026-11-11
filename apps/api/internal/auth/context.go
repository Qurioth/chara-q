package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Viewer はログインしている本人。
type Viewer struct {
	UserID      uuid.UUID
	LoginID     string
	DisplayName string
	IsAdmin     bool
}

// requestState はリクエストごとの認証の状態。ミドルウェアが context に入れ、resolver が使う。
type requestState struct {
	viewer    *Viewer
	tokenHash []byte
	w         http.ResponseWriter
}

type contextKey struct{}

func withRequestState(ctx context.Context, st *requestState) context.Context {
	return context.WithValue(ctx, contextKey{}, st)
}

func stateFrom(ctx context.Context) *requestState {
	st, _ := ctx.Value(contextKey{}).(*requestState)
	return st
}

// ViewerFrom はログインしている本人を返す。ログインしていなければ nil。
func ViewerFrom(ctx context.Context) *Viewer {
	if st := stateFrom(ctx); st != nil {
		return st.viewer
	}
	return nil
}

// SessionTokenHashFrom は、ブラウザが送ってきたセッション ID のハッシュを返す（なければ nil）。
// セッションが有効でなくても返す（ログインのときに前のセッションを消すため）。
func SessionTokenHashFrom(ctx context.Context) []byte {
	if st := stateFrom(ctx); st != nil {
		return st.tokenHash
	}
	return nil
}
