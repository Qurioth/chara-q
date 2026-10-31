// Package graph は gqlgen の resolver と生成コードを置く。
package graph

//go:generate go tool gqlgen generate

import (
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
	"github.com/Qurioth/chara-q/apps/api/internal/graph/model"
)

// Resolver は resolver が使う依存をまとめる。
type Resolver struct {
	Queries *db.Queries
	Auth    *auth.Service
}

// graphqlTime はスキーマの Time に対応する Go の型。
type graphqlTime = time.Time

func toViewer(v auth.Viewer) *model.Viewer {
	return &model.Viewer{
		ID:          v.UserID.String(),
		LoginID:     v.LoginID,
		DisplayName: v.DisplayName,
		IsAdmin:     v.IsAdmin,
	}
}
