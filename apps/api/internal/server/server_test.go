package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/config"
	"github.com/Qurioth/chara-q/apps/api/internal/graph"
	"github.com/Qurioth/chara-q/apps/api/internal/server"
)

// DB を使わずに、受け付ける HTTP リクエストの条件を確かめる（FR-005、research R4）。
// Cookie を付けないため、認証のミドルウェアは DB に触れない。
func newHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return server.New(config.Config{}, &graph.Resolver{Auth: auth.NewService(nil, logger)}, logger)
}

func TestGraphQLTransport(t *testing.T) {
	h := newHandler()
	mutation := `{"query":"mutation { logout }"}`

	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		target      string
		wantStatus  int
		wantBody    string
	}{
		{
			name:       "GET のクエリは受け付ける",
			method:     http.MethodGet,
			target:     "/graphql?query=" + url.QueryEscape("{ __typename }"),
			wantStatus: http.StatusOK,
			wantBody:   `"__typename":"Query"`,
		},
		{
			name:       "GET のミューテーションは処理しない",
			method:     http.MethodGet,
			target:     "/graphql?query=" + url.QueryEscape("mutation { logout }"),
			wantStatus: http.StatusNotAcceptable,
			wantBody:   "GET requests only allow query operations",
		},
		{
			name:        "POST の application/json は受け付ける",
			method:      http.MethodPost,
			contentType: "application/json; charset=utf-8",
			body:        mutation,
			wantStatus:  http.StatusOK,
			wantBody:    `"logout":true`,
		},
		{
			name:        "POST の text/plain は拒否する",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        mutation,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "POST のフォームは拒否する",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			body:        "query=" + url.QueryEscape("mutation { logout }"),
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "POST の multipart は拒否する",
			method:      http.MethodPost,
			contentType: "multipart/form-data; boundary=x",
			body:        "--x\r\nContent-Disposition: form-data; name=\"operations\"\r\n\r\n" + mutation + "\r\n--x--\r\n",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:       "Content-Type なしの POST は拒否する",
			method:     http.MethodPost,
			body:       mutation,
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := tt.target
			if target == "" {
				target = "/graphql"
			}
			req := httptest.NewRequestWithContext(context.Background(), tt.method, target, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			req.Header.Set("Origin", "https://evil.example")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d（%s）", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body に %q がない: %s", tt.wantBody, rec.Body.String())
			}
			if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
				t.Errorf("CORS の許可を返してはいけない: %q", v)
			}
		})
	}
}

func TestNoCORSOnPreflight(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/graphql", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, req)
	for _, h := range []string{"Access-Control-Allow-Origin", "Access-Control-Allow-Methods", "Access-Control-Allow-Headers"} {
		if v := rec.Header().Get(h); v != "" {
			t.Errorf("%s を返してはいけない: %q", h, v)
		}
	}
}
