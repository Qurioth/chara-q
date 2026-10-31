package auth

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionCookie(t *testing.T) {
	tests := []struct {
		name  string
		write func(ctx context.Context) error
		want  []string
	}{
		{
			name:  "設定",
			write: func(ctx context.Context) error { return SetSessionCookie(ctx, "TOKEN") },
			want:  []string{"__Host-chara_q_session=TOKEN", "Path=/", "Max-Age=34560000", "HttpOnly", "Secure", "SameSite=Lax"},
		},
		{
			name:  "削除",
			write: ClearSessionCookie,
			want:  []string{"__Host-chara_q_session=", "Path=/", "Max-Age=0", "HttpOnly", "Secure", "SameSite=Lax"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx := withRequestState(context.Background(), &requestState{w: rec})
			if err := tt.write(ctx); err != nil {
				t.Fatal(err)
			}
			got := rec.Header().Get("Set-Cookie")
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("%q がない: %s", w, got)
				}
			}
			if strings.Contains(strings.ToLower(got), "domain=") {
				t.Errorf("Domain を付けてはいけない: %s", got)
			}
		})
	}
}

func TestSessionCookieWithoutMiddleware(t *testing.T) {
	if err := SetSessionCookie(context.Background(), "x"); err == nil {
		t.Error("ミドルウェアを通っていなければエラー")
	}
}
