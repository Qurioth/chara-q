// Package model は GraphQL の型を置く。models_gen.go は gqlgen の生成コード。
package model

// このファイルは gqlgen の生成コードではない。gqlgen.yml の models で、スキーマの
// YuyakekoyakeSheet と BeginningIdolSheet をここの型に結び付けている（specs/003-character-sheet-base の research R3）。
// 所有者は ID だけを持ち、owner の項目は resolver が DataLoader で引く（constitution V）。

import (
	"time"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/graph/scalar"
)

// SheetCommon は、interface CharacterSheet の共通の項目。
type SheetCommon struct {
	ID            string
	System        GameSystem
	OwnerID       uuid.UUID
	CharacterName string
	PlayerName    string
	Age           string
	Sex           string
	Memo          string
	Revision      int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// YuyakekoyakeSheet は、ふしぎもののけRPG ゆうやけこやけ のシート。
type YuyakekoyakeSheet struct {
	SheetCommon
	// システム固有の部分（JSON のオブジェクト）
	Data scalar.JSON
}

// IsCharacterSheet は、gqlgen の CharacterSheet インターフェースを満たす。
func (YuyakekoyakeSheet) IsCharacterSheet() {}

// BeginningIdolSheet は、駆け出しアイドルRPG ビギニングアイドル のシート。
type BeginningIdolSheet struct {
	SheetCommon
	// システム固有の部分（JSON のオブジェクト）
	Data scalar.JSON
}

// IsCharacterSheet は、gqlgen の CharacterSheet インターフェースを満たす。
func (BeginningIdolSheet) IsCharacterSheet() {}
