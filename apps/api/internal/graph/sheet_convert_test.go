package graph

import (
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/graph/model"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
)

func TestToSheetOrderCoversAllValues(t *testing.T) {
	// GraphQL の並び順のすべての値が、それぞれ別の並び順になる（足し忘れると既定の順に落ちる）
	seen := map[sheet.Order]model.CharacterSheetOrder{}
	for _, o := range model.AllCharacterSheetOrder {
		got := toSheetOrder(o)
		if prev, ok := seen[got]; ok {
			t.Errorf("%s と %s が同じ並び順 %d になる", prev, o, got)
		}
		seen[got] = o
	}
}
