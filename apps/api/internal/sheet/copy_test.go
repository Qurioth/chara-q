package sheet_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

func TestCopy(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	alice := env.CreateUser(t, "alice", "password-1234", false)
	bob := env.CreateUser(t, "bob", "password-1234", false)
	src, err := svc.Create(context.Background(), alice.ID, sheet.SystemBeginningIdol, sheet.Fields{
		CharacterName: "元", PlayerName: "P", Age: "17", Sex: "女", Memo: "メモ", Data: []byte(`{"x":1}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	env.Clock.Advance(time.Hour)

	cp, err := svc.Copy(context.Background(), bob.ID, src.ID.String())
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if cp.ID == src.ID || cp.OwnerID != bob.ID || cp.System != src.System || cp.Revision != 1 {
		t.Errorf("別の ID・複製した人の所有・同じシステム・revision 1 のはず: %+v", cp)
	}
	if cp.CharacterName != "元（コピー）" || cp.PlayerName != "P" || cp.Age != "17" || cp.Sex != "女" || cp.Memo != "メモ" || string(cp.Data) != string(src.Data) {
		t.Errorf("内容が同じ・名前の末尾に（コピー）のはず: %+v", cp)
	}
	if !cp.CreatedAt.Equal(env.Clock.Now()) || !cp.UpdatedAt.Equal(env.Clock.Now()) {
		t.Errorf("日時は複製した時刻のはず: %v %v", cp.CreatedAt, cp.UpdatedAt)
	}

	// 複製を更新しても元は変わらない
	if _, err := svc.Update(context.Background(), bob.ID, cp.ID.String(), 1, sheet.Fields{CharacterName: "変えた", Data: []byte(`{}`)}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := sheettest.Row(t, env, src.ID); got.CharacterName != "元" || got.Revision != 1 {
		t.Errorf("元のシートが変わった: %+v", got)
	}

	// 自分のシートも複製できる
	if _, err := svc.Copy(context.Background(), alice.ID, src.ID.String()); err != nil {
		t.Errorf("自分のシートの複製: %v", err)
	}
}

func TestCopyName(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"95 文字はそのまま", strings.Repeat("あ", 95), strings.Repeat("あ", 95) + "（コピー）"},
		{"100 文字は先頭 95 文字", strings.Repeat("い", 100), strings.Repeat("い", 95) + "（コピー）"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, tt.src, "")
			cp, err := svc.Copy(context.Background(), owner.ID, src.ID.String())
			if err != nil {
				t.Fatalf("Copy: %v", err)
			}
			if cp.CharacterName != tt.want || utf8.RuneCountInString(cp.CharacterName) > sheet.CharacterNameMaxLen {
				t.Errorf("got %q (%d 文字)", cp.CharacterName, utf8.RuneCountInString(cp.CharacterName))
			}
		})
	}
}

func TestCopyNotFound(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)
	deleted := sheettest.CreateSheet(t, env, owner.ID, sheet.SystemYuyakekoyake, "削除済み", "")
	sheettest.MarkDeleted(t, env, deleted.ID)

	for _, id := range []string{deleted.ID.String(), "abc", "01900000-0000-7000-8000-000000000000"} {
		if _, err := svc.Copy(context.Background(), owner.ID, id); !errors.Is(err, sheet.ErrNotFound) {
			t.Errorf("%s: ErrNotFound のはず: %v", id, err)
		}
	}
}
