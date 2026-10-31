// Package server は HTTP ルーターを組み立てる。
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/Qurioth/chara-q/apps/api/internal/config"
	"github.com/Qurioth/chara-q/apps/api/internal/graph"
)

// New は API のルーターを返す。
func New(cfg config.Config, resolver *graph.Resolver, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(requestLogger(logger))
	r.Use(middleware.Recoverer)
	// ブラウザは web とだけ通信し、api へは rewrites で中継するため、CORS の許可は置かない（ADR 0006）

	// /healthz はプロセスが動いているかだけを返す（DB の確認は GraphQL の health で行う）
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	// セッションの Cookie から本人を読む（ADR 0006）
	r.With(resolver.Auth.Middleware).Handle("/graphql", newGraphQLHandler(cfg, resolver, logger))
	if cfg.Development {
		r.Handle("/playground", playground.Handler("chara-q GraphQL", "/graphql"))
	}
	return r
}

func newGraphQLHandler(cfg config.Config, resolver *graph.Resolver, logger *slog.Logger) http.Handler {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	// CSRF の対策: ミューテーションは POST かつ application/json だけで受け付ける（FR-005、research R4）。
	// gqlgen の POST は application/json だけを、GET はクエリだけを受け付ける。
	// フォームや multipart の transport は追加しない（server_test.go で確かめる）
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.SetErrorPresenter(graph.ErrorPresenter(logger))
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	if cfg.Development {
		srv.Use(extension.Introspection{})
	}
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})
	return srv
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}
