// Package scalar は GraphQL のカスタムスカラーを定義する。
package scalar

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

// JSON は GraphQL の JSON スカラー。JSONB の値をデコードせずにそのまま受け渡す。
type JSON json.RawMessage

// MarshalGQL は値をそのまま JSON として書き出す。
func (j JSON) MarshalGQL(w io.Writer) {
	if len(j) == 0 {
		_, _ = io.WriteString(w, "null")
		return
	}
	_, _ = w.Write(j)
}

// UnmarshalGQL は GraphQL の入力値を JSON に変換する。
func (j *JSON) UnmarshalGQL(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("JSON スカラーに変換できません: %w", err)
	}
	*j = b
	return nil
}

var (
	_ graphql.Marshaler   = JSON(nil)
	_ graphql.Unmarshaler = (*JSON)(nil)
)
