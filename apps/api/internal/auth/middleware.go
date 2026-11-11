package auth

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// Middleware はセッションの Cookie から本人を読み、context に入れる。
// Cookie がない・セッションが無効のときはログインしていない扱いにし、エラーにはしない（FR-023）。
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := &requestState{w: w}
		ctx := withRequestState(r.Context(), st)

		c, err := r.Cookie(SessionCookieName)
		if err != nil || c.Value == "" {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		st.tokenHash = HashToken(c.Value)
		now := s.now()
		row, err := s.q.GetSessionViewer(ctx, db.GetSessionViewerParams{TokenHash: st.tokenHash, Now: now})
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			// 期限切れ・存在しない・無効なユーザーのセッション。Cookie も消す
			clearSessionCookie(w)
		case err != nil:
			// Cookie の値はログに出さない（FR-010）
			s.logger.ErrorContext(ctx, "セッションを読めません", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		default:
			st.viewer = &Viewer{
				UserID:      row.UserID,
				LoginID:     row.LoginID,
				DisplayName: row.DisplayName,
				IsAdmin:     row.IsAdmin,
			}
			// 最後に使ってから 30 日で切れる。書き込みを減らすため、1 時間ごとにだけ延ばす（research R10）
			if now.Sub(row.LastUsedAt) >= sessionTouchInterval {
				if err := s.q.TouchSession(ctx, db.TouchSessionParams{
					Now:       now,
					ExpiresAt: now.Add(SessionTTL),
					TokenHash: st.tokenHash,
				}); err != nil {
					s.logger.ErrorContext(ctx, "セッションを延長できません", "error", err)
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
