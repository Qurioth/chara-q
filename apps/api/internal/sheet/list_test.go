package sheet_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/db"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

func names(rows []db.CharacterSheet) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.CharacterName
	}
	return out
}

func sameNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestListFilters(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	alice := env.CreateUser(t, "alice", "password-1234", false)
	bob := env.CreateUser(t, "bob", "password-1234", false)

	// 更新日時の新しい順に並ぶよう、1 件ごとに時計を進める
	create := func(owner uuid.UUID, system sheet.System, name, player string) db.CharacterSheet {
		env.Clock.Advance(time.Minute)
		return sheettest.CreateSheet(t, env, owner, system, name, player)
	}
	create(alice.ID, sheet.SystemYuyakekoyake, "ＦＯＸ きつね", "Taro")
	create(alice.ID, sheet.SystemBeginningIdol, "アイドル 100%", "花子")
	create(bob.ID, sheet.SystemYuyakekoyake, "たぬき_丸", "jiro")
	create(bob.ID, sheet.SystemBeginningIdol, `back\slash`, "")
	deleted := create(bob.ID, sheet.SystemYuyakekoyake, "削除した fox", "")
	sheettest.MarkDeleted(t, env, deleted.ID)

	yuyake := sheet.SystemYuyakekoyake
	tests := []struct {
		name   string
		filter sheet.Filter
		viewer *uuid.UUID
		want   []string
	}{
		{"条件なし（削除済みを含まない）", sheet.Filter{}, nil, []string{`back\slash`, "たぬき_丸", "アイドル 100%", "ＦＯＸ きつね"}},
		{"システム", sheet.Filter{System: &yuyake}, nil, []string{"たぬき_丸", "ＦＯＸ きつね"}},
		{"半角の小文字で全角の大文字が見つかる", sheet.Filter{Keyword: "fox"}, nil, []string{"ＦＯＸ きつね"}},
		{"プレイヤー名でも見つかる", sheet.Filter{Keyword: "TARO"}, nil, []string{"ＦＯＸ きつね"}},
		{"% は文字どおり", sheet.Filter{Keyword: "%"}, nil, []string{"アイドル 100%"}},
		{"全角の ％ も文字どおり", sheet.Filter{Keyword: "％"}, nil, []string{"アイドル 100%"}},
		{"_ は文字どおり", sheet.Filter{Keyword: "_"}, nil, []string{"たぬき_丸"}},
		{`\ は文字どおり`, sheet.Filter{Keyword: `\`}, nil, []string{`back\slash`}},
		{"空白だけのキーワードは条件なし", sheet.Filter{Keyword: "  "}, nil, []string{`back\slash`, "たぬき_丸", "アイドル 100%", "ＦＯＸ きつね"}},
		{"自分のシートだけ", sheet.Filter{Mine: true}, &alice.ID, []string{"アイドル 100%", "ＦＯＸ きつね"}},
		{"ログインしていなければ自分のシートだけは無視", sheet.Filter{Mine: true}, nil, []string{`back\slash`, "たぬき_丸", "アイドル 100%", "ＦＯＸ きつね"}},
		{"組み合わせ", sheet.Filter{System: &yuyake, Mine: true}, &bob.ID, []string{"たぬき_丸"}},
		{"合うものがない", sheet.Filter{Keyword: "存在しない"}, nil, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := svc.List(context.Background(), tt.filter, sheet.OrderUpdatedDesc, 1, tt.viewer)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if got := names(p.Rows); !sameNames(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if p.TotalCount != len(tt.want) {
				t.Errorf("TotalCount: got %d, want %d", p.TotalCount, len(tt.want))
			}
		})
	}
}

func TestListOrders(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)

	base := env.Clock.Now()
	a := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "かきくけこ", "")
	b := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "アイウエオ", "")
	c := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "さしすせそ", "")
	sheettest.SetTimes(t, env, a.ID, base.Add(3*time.Minute), base.Add(1*time.Minute))
	sheettest.SetTimes(t, env, b.ID, base.Add(1*time.Minute), base.Add(3*time.Minute))
	sheettest.SetTimes(t, env, c.ID, base.Add(2*time.Minute), base.Add(2*time.Minute))

	tests := []struct {
		name  string
		order sheet.Order
		want  []string
	}{
		{"更新日時の新しい順", sheet.OrderUpdatedDesc, []string{"アイウエオ", "さしすせそ", "かきくけこ"}},
		{"作成日時の新しい順", sheet.OrderCreatedDesc, []string{"かきくけこ", "さしすせそ", "アイウエオ"}},
		// カタカナとひらがなを五十音の順に並べる（ja-x-icu）
		{"キャラクター名の順", sheet.OrderCharacterNameAsc, []string{"アイウエオ", "かきくけこ", "さしすせそ"}},
		{"更新日時の古い順", sheet.OrderUpdatedAsc, []string{"かきくけこ", "さしすせそ", "アイウエオ"}},
		{"作成日時の古い順", sheet.OrderCreatedAsc, []string{"アイウエオ", "さしすせそ", "かきくけこ"}},
		{"キャラクター名の逆の順", sheet.OrderCharacterNameDesc, []string{"さしすせそ", "かきくけこ", "アイウエオ"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := svc.List(context.Background(), sheet.Filter{}, tt.order, 1, nil)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if got := names(p.Rows); !sameNames(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestListTiesAreStable(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)

	// 日時も名前も同じシート。ID の順で並ぶ
	var ids []uuid.UUID
	for range 3 {
		ids = append(ids, sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "同じ名前", "").ID)
	}
	allOrders := []sheet.Order{
		sheet.OrderUpdatedDesc, sheet.OrderUpdatedAsc,
		sheet.OrderCreatedDesc, sheet.OrderCreatedAsc,
		sheet.OrderCharacterNameAsc, sheet.OrderCharacterNameDesc,
	}
	for _, order := range allOrders {
		first, _ := svc.List(context.Background(), sheet.Filter{}, order, 1, nil)
		for range 3 {
			again, _ := svc.List(context.Background(), sheet.Filter{}, order, 1, nil)
			for i := range first.Rows {
				if first.Rows[i].ID != again.Rows[i].ID {
					t.Fatalf("並び順 %d: 毎回同じ順のはず", order)
				}
			}
		}
	}
	asc, _ := svc.List(context.Background(), sheet.Filter{}, sheet.OrderCharacterNameAsc, 1, nil)
	desc, _ := svc.List(context.Background(), sheet.Filter{}, sheet.OrderUpdatedDesc, 1, nil)
	for i := range ids {
		if asc.Rows[i].ID != desc.Rows[len(ids)-1-i].ID {
			t.Fatalf("同じ値のときは、名前の順は ID の昇順、日時の順は ID の降順のはず")
		}
	}
	// 逆の順は、同じ値のときも含めて、ちょうど反対に並ぶ（ID の向きも反転する）
	for _, pair := range [][2]sheet.Order{
		{sheet.OrderUpdatedDesc, sheet.OrderUpdatedAsc},
		{sheet.OrderCreatedDesc, sheet.OrderCreatedAsc},
		{sheet.OrderCharacterNameAsc, sheet.OrderCharacterNameDesc},
	} {
		fwd, _ := svc.List(context.Background(), sheet.Filter{}, pair[0], 1, nil)
		rev, _ := svc.List(context.Background(), sheet.Filter{}, pair[1], 1, nil)
		for i := range ids {
			if fwd.Rows[i].ID != rev.Rows[len(ids)-1-i].ID {
				t.Fatalf("並び順 %d と %d: ちょうど反対に並ぶはず", pair[0], pair[1])
			}
		}
	}
}

func TestListPages(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)

	empty, err := svc.List(context.Background(), sheet.Filter{}, sheet.OrderUpdatedDesc, 5, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if empty.Page != 1 || empty.TotalPages != 1 || empty.TotalCount != 0 || len(empty.Rows) != 0 {
		t.Fatalf("0 件なら 1 ページ目・全 1 ページのはず: %+v", empty)
	}

	for i := range 45 {
		env.Clock.Advance(time.Second)
		sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, fmt.Sprintf("シート %02d", i), "")
	}
	tests := []struct {
		name     string
		page     int
		wantPage int
		wantLen  int
	}{
		{"1 ページ目", 1, 1, 20},
		{"最後のページ", 3, 3, 5},
		{"0 は 1 ページ目", 0, 1, 20},
		{"負の数は 1 ページ目", -3, 1, 20},
		{"範囲の外は最後のページ", 999, 3, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := svc.List(context.Background(), sheet.Filter{}, sheet.OrderUpdatedDesc, tt.page, nil)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if p.Page != tt.wantPage || len(p.Rows) != tt.wantLen || p.TotalCount != 45 || p.TotalPages != 3 || p.PageSize != 20 {
				t.Errorf("got page=%d len=%d total=%d pages=%d size=%d", p.Page, len(p.Rows), p.TotalCount, p.TotalPages, p.PageSize)
			}
		})
	}

	// ページを最後まで送っても、重複も抜けもない
	seen := map[uuid.UUID]bool{}
	for page := 1; page <= 3; page++ {
		p, _ := svc.List(context.Background(), sheet.Filter{}, sheet.OrderUpdatedDesc, page, nil)
		for _, r := range p.Rows {
			if seen[r.ID] {
				t.Fatalf("重複: %s", r.CharacterName)
			}
			seen[r.ID] = true
		}
	}
	if len(seen) != 45 {
		t.Fatalf("抜けがある: %d 件", len(seen))
	}
}
