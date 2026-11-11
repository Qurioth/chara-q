package sheet_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

func updatedFields(name string) sheet.Fields {
	return sheet.Fields{CharacterName: name, Age: "17 歳", Data: []byte(`{"a":1}`)}
}

func TestUpdate(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)
	row := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "元の名前", "")
	env.Clock.Advance(time.Hour)

	got, err := svc.Update(context.Background(), owner.ID, row.ID.String(), 1, updatedFields("新しい名前"))
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.CharacterName != "新しい名前" || got.Age != "17 歳" || got.Revision != 2 {
		t.Errorf("内容と revision: %+v", got)
	}
	if !got.UpdatedAt.Equal(env.Clock.Now()) || !got.CreatedAt.Equal(row.CreatedAt) {
		t.Errorf("updated_at は時計の時刻、created_at は変わらないはず: %v %v", got.UpdatedAt, got.CreatedAt)
	}
	if got.OwnerID != owner.ID || got.System != row.System {
		t.Errorf("所有者とシステムは変わらないはず: %+v", got)
	}
}

func TestUpdateFailures(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)
	admin := env.CreateUser(t, "admin", "password-1234", true)
	row := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "名前", "")
	deleted := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "削除済み", "")
	sheettest.MarkDeleted(t, env, deleted.ID)

	tests := []struct {
		name     string
		id       string
		revision int
		wantErr  error
	}{
		{"所有者でない（管理者）", row.ID.String(), 1, sheet.ErrForbidden},
		{"存在しない", "01900000-0000-7000-8000-000000000000", 1, sheet.ErrNotFound},
		{"形の正しくない ID", "abc", 1, sheet.ErrNotFound},
		{"削除済み", deleted.ID.String(), 1, sheet.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Update(context.Background(), admin.ID, tt.id, tt.revision, updatedFields("変えた"))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}

	// 所有者でも、古い revision なら REVISION_CONFLICT
	if _, err := svc.Update(context.Background(), owner.ID, row.ID.String(), 1, updatedFields("1 回目")); err != nil {
		t.Fatalf("1 回目: %v", err)
	}
	_, err := svc.Update(context.Background(), owner.ID, row.ID.String(), 1, updatedFields("2 回目"))
	var ie auth.InputErrors
	if !errors.As(err, &ie) || ie[0].Code != sheet.CodeRevisionConflict || ie[0].Field != auth.FieldNone {
		t.Fatalf("REVISION_CONFLICT のはず: %v", err)
	}
	// 判定の順: 所有者でなく、revision も古いときは FORBIDDEN
	if _, err := svc.Update(context.Background(), admin.ID, row.ID.String(), 1, updatedFields("x")); !errors.Is(err, sheet.ErrForbidden) {
		t.Fatalf("FORBIDDEN のはず: %v", err)
	}
	if got := sheettest.Row(t, env, row.ID); got.CharacterName != "1 回目" || got.Revision != 2 {
		t.Fatalf("失敗した更新で内容が変わってはいけない: %+v", got)
	}
}

func TestUpdateConcurrent(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)
	row := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "名前", "")

	const n = 10
	var wg sync.WaitGroup
	results := make([]error, n)
	for i := range n {
		wg.Go(func() {
			_, results[i] = svc.Update(context.Background(), owner.ID, row.ID.String(), 1, updatedFields("同時"))
		})
	}
	wg.Wait()
	ok := 0
	for _, err := range results {
		if err == nil {
			ok++
		}
	}
	if ok != 1 {
		t.Fatalf("同じ revision の同時の更新は、成功が 1 回だけのはず: %d 回", ok)
	}
	if got := sheettest.Row(t, env, row.ID); got.Revision != 2 {
		t.Fatalf("revision は 2 のはず: %d", got.Revision)
	}
}
