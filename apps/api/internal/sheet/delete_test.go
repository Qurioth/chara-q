package sheet_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

func TestDelete(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)
	other := env.CreateUser(t, "other", "password-1234", true)
	sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "古い", "")
	env.Clock.Advance(time.Minute)
	row := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "消す", "")
	env.Clock.Advance(time.Minute)
	sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "新しい", "")
	env.Clock.Advance(time.Hour)
	ctx := context.Background()

	if err := svc.Delete(ctx, other.ID, row.ID.String()); !errors.Is(err, sheet.ErrForbidden) {
		t.Fatalf("所有者でない（管理者）は ErrForbidden のはず: %v", err)
	}
	if err := svc.Delete(ctx, owner.ID, row.ID.String()); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got := sheettest.Row(t, env, row.ID)
	if got.DeletedAt == nil || !got.DeletedAt.Equal(env.Clock.Now()) {
		t.Errorf("deleted_at は時計の時刻のはず: %v", got.DeletedAt)
	}
	if !got.UpdatedAt.Equal(row.UpdatedAt) || got.Revision != row.Revision || got.CharacterName != row.CharacterName {
		t.Errorf("updated_at・revision・内容は変わらないはず: %+v", got)
	}

	// 削除の後は、どの操作の対象にもならない
	if err := svc.Delete(ctx, owner.ID, row.ID.String()); !errors.Is(err, sheet.ErrNotFound) {
		t.Errorf("もう一度の削除は ErrNotFound のはず: %v", err)
	}
	if r, err := svc.Get(ctx, row.ID.String()); err != nil || r != nil {
		t.Errorf("Get は nil のはず: %v %v", r, err)
	}
	if _, err := svc.Update(ctx, owner.ID, row.ID.String(), 1, sheet.Fields{CharacterName: "x", Data: []byte(`{}`)}); !errors.Is(err, sheet.ErrNotFound) {
		t.Errorf("Update は ErrNotFound のはず: %v", err)
	}
	if _, err := svc.Copy(ctx, owner.ID, row.ID.String()); !errors.Is(err, sheet.ErrNotFound) {
		t.Errorf("Copy は ErrNotFound のはず: %v", err)
	}
	p, _ := svc.List(ctx, sheet.Filter{}, sheet.OrderUpdatedDesc, 1, nil)
	if got := names(p.Rows); !sameNames(got, []string{"新しい", "古い"}) {
		t.Errorf("一覧に削除したシートが含まれる: %q", got)
	}

	// data-model.md の取り消しの SQL で、同じ内容・元の位置に戻る
	sheettest.Restore(t, env, row.ID)
	if r, err := svc.Get(ctx, row.ID.String()); err != nil || r == nil || r.CharacterName != "消す" {
		t.Errorf("取り消すと Get で読めるはず: %v %v", r, err)
	}
	p, _ = svc.List(ctx, sheet.Filter{}, sheet.OrderUpdatedDesc, 1, nil)
	if got := names(p.Rows); !sameNames(got, []string{"新しい", "消す", "古い"}) {
		t.Errorf("取り消すと元の位置に戻るはず: %q", got)
	}
}

func TestDeleteNotFound(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)
	for _, id := range []string{"abc", "01900000-0000-7000-8000-000000000000"} {
		if err := svc.Delete(context.Background(), owner.ID, id); !errors.Is(err, sheet.ErrNotFound) {
			t.Errorf("%s: ErrNotFound のはず: %v", id, err)
		}
	}
}
