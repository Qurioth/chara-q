package sheet

// System はシートのゲームシステム。値は DB の character_sheets.system と同じ（data-model.md）。
// GraphQL の GameSystem との変換は internal/graph で行う（sheet は GraphQL に依存しない）。
type System string

// System の値。
const (
	SystemYuyakekoyake  System = "yuyakekoyake"
	SystemBeginningIdol System = "beginning_idol"
)

// Valid は知っているシステムかを返す。
func (s System) Valid() bool {
	switch s {
	case SystemYuyakekoyake, SystemBeginningIdol:
		return true
	}
	return false
}
