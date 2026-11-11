package sheet

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// PageSize は一覧の 1 ページの件数（FR-012）。
const PageSize = 20

// Order は一覧の並び順（FR-011、research R9）。
type Order int

// Order の値。
const (
	OrderUpdatedDesc Order = iota
	OrderUpdatedAsc
	OrderCreatedDesc
	OrderCreatedAsc
	OrderCharacterNameAsc
	OrderCharacterNameDesc
)

// Filter は一覧の絞り込み（FR-010）。
type Filter struct {
	// nil ならすべてのシステム
	System *System
	// 前後の空白を除いて空なら条件なし
	Keyword string
	// true なら、viewer のシートだけ。viewer が nil（ログインしていない）なら無視する
	Mine bool
}

// Page は一覧の 1 ページ。
type Page struct {
	Rows       []db.CharacterSheet
	TotalCount int
	// 範囲の外を丸めた後のページ
	Page       int
	PageSize   int
	TotalPages int
}

// List は削除されていないシートの一覧を返す。page は 1〜全ページ数に丸める（0 件なら 1）。
func (s *Service) List(ctx context.Context, f Filter, order Order, page int, viewer *uuid.UUID) (Page, error) {
	var system *string
	if f.System != nil {
		v := string(*f.System)
		system = &v
	}
	var keyword *string
	if k := strings.TrimSpace(f.Keyword); k != "" {
		keyword = &k
	}
	var owner *uuid.UUID
	if f.Mine && viewer != nil {
		owner = viewer
	}

	count, err := s.q.CountCharacterSheets(ctx, db.CountCharacterSheetsParams{System: system, OwnerID: owner, Keyword: keyword})
	if err != nil {
		return Page{}, fmt.Errorf("シートの件数を数えられません: %w", err)
	}
	total := int(count)
	totalPages := max(1, (total+PageSize-1)/PageSize)
	page = min(max(page, 1), totalPages)

	limit := int32(PageSize)
	offset, ok := toInt32((page - 1) * PageSize)
	if !ok {
		return Page{}, fmt.Errorf("ページが大きすぎます: %d", page)
	}

	var rows []db.CharacterSheet
	switch order {
	case OrderCreatedDesc:
		rows, err = s.q.ListCharacterSheetsByCreated(ctx, db.ListCharacterSheetsByCreatedParams{
			System: system, OwnerID: owner, Keyword: keyword, RowLimit: limit, RowOffset: offset,
		})
	case OrderCreatedAsc:
		rows, err = s.q.ListCharacterSheetsByCreatedAsc(ctx, db.ListCharacterSheetsByCreatedAscParams{
			System: system, OwnerID: owner, Keyword: keyword, RowLimit: limit, RowOffset: offset,
		})
	case OrderCharacterNameAsc:
		rows, err = s.q.ListCharacterSheetsByName(ctx, db.ListCharacterSheetsByNameParams{
			System: system, OwnerID: owner, Keyword: keyword, RowLimit: limit, RowOffset: offset,
		})
	case OrderCharacterNameDesc:
		rows, err = s.q.ListCharacterSheetsByNameDesc(ctx, db.ListCharacterSheetsByNameDescParams{
			System: system, OwnerID: owner, Keyword: keyword, RowLimit: limit, RowOffset: offset,
		})
	case OrderUpdatedAsc:
		rows, err = s.q.ListCharacterSheetsByUpdatedAsc(ctx, db.ListCharacterSheetsByUpdatedAscParams{
			System: system, OwnerID: owner, Keyword: keyword, RowLimit: limit, RowOffset: offset,
		})
	default:
		rows, err = s.q.ListCharacterSheetsByUpdated(ctx, db.ListCharacterSheetsByUpdatedParams{
			System: system, OwnerID: owner, Keyword: keyword, RowLimit: limit, RowOffset: offset,
		})
	}
	if err != nil {
		return Page{}, fmt.Errorf("シートの一覧を読めません: %w", err)
	}
	return Page{Rows: rows, TotalCount: total, Page: page, PageSize: PageSize, TotalPages: totalPages}, nil
}
