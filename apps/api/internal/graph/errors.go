package graph

import (
	"context"
	"errors"
	"log/slog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/Qurioth/chara-q/apps/api/internal/auth"
	"github.com/Qurioth/chara-q/apps/api/internal/graph/model"
)

// GraphQL のエラーの extensions.code（contracts/graphql.md）。
const (
	codeUnauthenticated = "UNAUTHENTICATED"
	codeForbidden       = "FORBIDDEN"
)

func errUnauthenticated() error {
	return &gqlerror.Error{Message: "ログインが必要です", Extensions: map[string]any{"code": codeUnauthenticated}}
}

func errForbidden() error {
	return &gqlerror.Error{Message: "管理者だけが行えます", Extensions: map[string]any{"code": codeForbidden}}
}

// noErrors は payload の errors に入れる空の一覧。
func noErrors() []*model.InputError {
	return []*model.InputError{}
}

// inputErrors は auth の入力の失敗を、payload の errors に変換する。入力の失敗でなければ ok は false。
func inputErrors(err error) (errs []*model.InputError, ok bool) {
	var ie auth.InputErrors
	if !errors.As(err, &ie) {
		return nil, false
	}
	errs = make([]*model.InputError, len(ie))
	for i, e := range ie {
		out := &model.InputError{Code: model.InputErrorCode(e.Code)}
		if e.Field != auth.FieldNone {
			f := model.InputField(e.Field)
			out.Field = &f
		}
		errs[i] = out
	}
	return errs, true
}

// lockedErrors はロック中の失敗を返す。ロック中でなければ ok は false。
func lockedErrors(err error) (errs []*model.InputError, lockedUntil *graphqlTime, ok bool) {
	var le *auth.LockedError
	if !errors.As(err, &le) {
		return nil, nil, false
	}
	until := le.Until
	return []*model.InputError{{Code: model.InputErrorCodeLocked}}, &until, true
}

// ErrorPresenter は想定外のエラーの本文を汎用の文言にし、詳細はログにだけ出す。
// 入力の値（パスワードやコード）が応答に含まれないようにする（FR-010）。
func ErrorPresenter(logger *slog.Logger) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		var gqlErr *gqlerror.Error
		if errors.As(err, &gqlErr) && gqlErr.Err == nil {
			// スキーマの検証の失敗や、errUnauthenticated などの意図したエラー
			return graphql.DefaultErrorPresenter(ctx, err)
		}
		logger.ErrorContext(ctx, "GraphQL の処理に失敗しました", "error", err)
		presented := graphql.DefaultErrorPresenter(ctx, err)
		presented.Message = "サーバーでエラーが起きました"
		presented.Extensions = map[string]any{"code": "INTERNAL_SERVER_ERROR"}
		return presented
	}
}
