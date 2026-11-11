package graph

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
	"github.com/Qurioth/chara-q/apps/api/internal/graph/model"
	"github.com/Qurioth/chara-q/apps/api/internal/graph/scalar"
	"github.com/Qurioth/chara-q/apps/api/internal/loader"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
)

// シートの DB の値と GraphQL の値の変換（specs/003-character-sheet-base の research R3）。
// DB の行からどちらのシステムの型を作るかは、ここの toSheetModel だけで決める。

func toSheetSystem(s model.GameSystem) (sheet.System, error) {
	switch s {
	case model.GameSystemYuyakekoyake:
		return sheet.SystemYuyakekoyake, nil
	case model.GameSystemBeginningIdol:
		return sheet.SystemBeginningIdol, nil
	}
	return "", fmt.Errorf("知らないシステムです: %s", s)
}

func toGameSystem(s sheet.System) (model.GameSystem, error) {
	switch s {
	case sheet.SystemYuyakekoyake:
		return model.GameSystemYuyakekoyake, nil
	case sheet.SystemBeginningIdol:
		return model.GameSystemBeginningIdol, nil
	}
	return "", fmt.Errorf("知らないシステムです: %s", s)
}

// toSheetModel は、DB の行をシステムに合った GraphQL の型にする。
func toSheetModel(row db.CharacterSheet) (model.CharacterSheet, error) {
	system, err := toGameSystem(sheet.System(row.System))
	if err != nil {
		return nil, err
	}
	common := model.SheetCommon{
		ID:            row.ID.String(),
		System:        system,
		OwnerID:       row.OwnerID,
		CharacterName: row.CharacterName,
		PlayerName:    row.PlayerName,
		Age:           row.Age,
		Sex:           row.Sex,
		Memo:          row.Memo,
		Revision:      int(row.Revision),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	switch system {
	case model.GameSystemYuyakekoyake:
		return &model.YuyakekoyakeSheet{SheetCommon: common, Data: scalar.JSON(row.Data)}, nil
	case model.GameSystemBeginningIdol:
		return &model.BeginningIdolSheet{SheetCommon: common, Data: scalar.JSON(row.Data)}, nil
	}
	return nil, fmt.Errorf("知らないシステムです: %s", system)
}

func toSheetFields(in *model.CharacterSheetFieldsInput) sheet.Fields {
	return sheet.Fields{
		CharacterName: in.CharacterName,
		PlayerName:    in.PlayerName,
		Age:           in.Age,
		Sex:           in.Sex,
		Memo:          in.Memo,
		Data:          []byte(in.Data),
	}
}

func toSheetOrder(o model.CharacterSheetOrder) sheet.Order {
	switch o {
	case model.CharacterSheetOrderUpdatedAtAsc:
		return sheet.OrderUpdatedAsc
	case model.CharacterSheetOrderCreatedAtDesc:
		return sheet.OrderCreatedDesc
	case model.CharacterSheetOrderCreatedAtAsc:
		return sheet.OrderCreatedAsc
	case model.CharacterSheetOrderCharacterNameAsc:
		return sheet.OrderCharacterNameAsc
	case model.CharacterSheetOrderCharacterNameDesc:
		return sheet.OrderCharacterNameDesc
	}
	return sheet.OrderUpdatedDesc
}

// loadOwner は、シートの所有者を DataLoader でまとめて引く（constitution V）。ログイン ID は返さない。
func loadOwner(ctx context.Context, id uuid.UUID) (*model.User, error) {
	u, err := loader.For(ctx).UserByID.Load(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.User{ID: u.ID.String(), DisplayName: u.DisplayName}, nil
}

// sheetPayload は、作成・更新・複製の結果を payload にする。入力の失敗（REVISION_CONFLICT を含む）は
// payload の errors に、見つからない・所有者でないは GraphQL のエラーにする（contracts/graphql.md）。
func sheetPayload(row db.CharacterSheet, err error) (*model.CharacterSheetPayload, error) {
	if errs, ok := inputErrors(err); ok {
		return &model.CharacterSheetPayload{Errors: errs}, nil
	}
	if err := sheetError(err); err != nil {
		return nil, err
	}
	s, err := toSheetModel(row)
	if err != nil {
		return nil, err
	}
	return &model.CharacterSheetPayload{CharacterSheet: s, Errors: noErrors()}, nil
}

// sheetError は、sheet のエラーを GraphQL のエラーにする。
func sheetError(err error) error {
	switch {
	case errors.Is(err, sheet.ErrNotFound):
		return errNotFound()
	case errors.Is(err, sheet.ErrForbidden):
		return errForbidden()
	}
	return err
}

// viewerID は、ログインしている本人の ID を返す。ログインしていなければ nil。
// resolver のファイル（gqlgen が import を並べ直す）で外部のパッケージを読み込まないよう、ここに置く。
func viewerID(ctx context.Context) *uuid.UUID {
	if v := auth.ViewerFrom(ctx); v != nil {
		return &v.UserID
	}
	return nil
}
