// Package loader は、GraphQL の関連データをまとめて引く DataLoader を置く（constitution V、
// specs/003-character-sheet-base の research R4）。ローダーはリクエストごとに作り、別のリクエストの結果を混ぜない。
package loader

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// User は公開してよいユーザーの情報（ID と表示名だけ。ログイン ID は持たない）。
type User struct {
	ID          uuid.UUID
	DisplayName string
}

// Loaders はリクエストごとのローダー。
type Loaders struct {
	// UserByID は、シートの所有者などのユーザーを ID でまとめて引く。無効なユーザーも返す
	UserByID *dataloadgen.Loader[uuid.UUID, User]
}

// New はローダーを作る。
func New(queries *db.Queries) *Loaders {
	return &Loaders{
		UserByID: dataloadgen.NewMappedLoader(func(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]User, error) {
			rows, err := queries.ListUsersByIDs(ctx, ids)
			if err != nil {
				return nil, fmt.Errorf("ユーザーを読めません: %w", err)
			}
			users := make(map[uuid.UUID]User, len(rows))
			for _, r := range rows {
				users[r.ID] = User{ID: r.ID, DisplayName: r.DisplayName}
			}
			return users, nil
		}),
	}
}

type contextKey struct{}

// Middleware は、リクエストごとに新しいローダーを context に入れる。
func Middleware(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), contextKey{}, New(queries))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// For は context のローダーを返す。Middleware を通っていなければ panic する（設定の誤り）。
func For(ctx context.Context) *Loaders {
	l, ok := ctx.Value(contextKey{}).(*Loaders)
	if !ok {
		panic("loader.Middleware を通っていません")
	}
	return l
}
