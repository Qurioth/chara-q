// Package graph は gqlgen の resolver と生成コードを置く。
package graph

//go:generate go tool gqlgen generate

import "github.com/Qurioth/chara-q/apps/api/internal/db"

// Resolver は resolver が使う依存をまとめる。
type Resolver struct {
	Queries *db.Queries
}
