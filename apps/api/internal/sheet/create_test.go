package sheet_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/auth/authtest"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet"
	"github.com/Qurioth/chara-q/apps/api/internal/sheet/sheettest"
)

func TestCreate(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)

	row, err := svc.Create(context.Background(), owner.ID, sheet.SystemBeginningIdol, sheet.Fields{
		CharacterName: "  アイドル  ",
		PlayerName:    " プレイヤー ",
		Memo:          "1 行目\r\n2 行目",
		Data:          []byte(`{"voice":3}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if row.OwnerID != owner.ID || row.System != string(sheet.SystemBeginningIdol) || row.Revision != 1 {
		t.Errorf("所有者・システム・revision が違う: %+v", row)
	}
	if row.CharacterName != "アイドル" || row.PlayerName != "プレイヤー" || row.Memo != "1 行目\n2 行目" {
		t.Errorf("前後の空白を除いた値・そろえた改行で保存するはず: %+v", row)
	}
	if !row.CreatedAt.Equal(env.Clock.Now()) || !row.UpdatedAt.Equal(env.Clock.Now()) {
		t.Errorf("日時は時計の時刻のはず: %v %v", row.CreatedAt, row.UpdatedAt)
	}
	if string(row.Data) != `{"voice": 3}` {
		t.Errorf("data がそのまま保存されていない: %s", row.Data)
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	env := authtest.New(t)
	svc := sheettest.NewService(env)
	owner := env.CreateUser(t, "owner", "password-1234", false)

	_, err := svc.Create(context.Background(), owner.ID, sheet.SystemYuyakekoyake, sheet.Fields{Data: []byte(`{}`)})
	var ie auth.InputErrors
	if !errors.As(err, &ie) || ie[0].Field != sheet.FieldCharacterName {
		t.Fatalf("キャラクター名の失敗のはず: %v", err)
	}
	if _, err := svc.Create(context.Background(), owner.ID, sheet.System("unknown"), sheet.Fields{CharacterName: "a", Data: []byte(`{}`)}); !errors.Is(err, sheet.ErrUnknownSystem) {
		t.Fatalf("知らないシステムは失敗するはず: %v", err)
	}
	if n := env.Count(t, "character_sheets"); n != 0 {
		t.Fatalf("行が作られてはいけない: %d", n)
	}
}
