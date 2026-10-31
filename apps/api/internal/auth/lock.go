package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Qurioth/chara-q/apps/api/internal/db"
)

// checkLock はロック中なら LockedError を返す（research R9）。
func (s *Service) checkLock(ctx context.Context, purpose, key string, now time.Time) error {
	until, err := s.q.GetLockedUntil(ctx, db.GetLockedUntilParams{Purpose: purpose, LoginIDKey: key, Now: now})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("ロックの状態を読めません: %w", err)
	}
	return &LockedError{Until: until}
}

// recordFailure は失敗を 1 回数え、ロックしたら LockedError を返す。
func (s *Service) recordFailure(ctx context.Context, purpose, key string, now time.Time) error {
	row, err := s.q.RecordFailure(ctx, db.RecordFailureParams{
		Purpose:     purpose,
		LoginIDKey:  key,
		Now:         now,
		MaxFailures: MaxFailures,
		LockUntil:   now.Add(LockDuration),
	})
	if err != nil {
		return fmt.Errorf("失敗を記録できません: %w", err)
	}
	if row.LockedUntil != nil {
		return &LockedError{Until: *row.LockedUntil}
	}
	return nil
}

func clearFailures(ctx context.Context, q *db.Queries, purpose, key string) error {
	if err := q.ClearFailures(ctx, db.ClearFailuresParams{Purpose: purpose, LoginIDKey: key}); err != nil {
		return fmt.Errorf("失敗の記録を消せません: %w", err)
	}
	return nil
}

// cleanup は期限切れのセッションと、古い失敗の記録を消す。常時起動の仕組みを置かないため、
// ログインの成功のたびに行う（research R9）。失敗してもログインは成功させる。
func (s *Service) cleanup(ctx context.Context, now time.Time) {
	if err := s.q.DeleteExpiredSessions(ctx, now); err != nil {
		s.logger.ErrorContext(ctx, "期限切れのセッションを消せません", "error", err)
	}
	if err := s.q.DeleteStaleFailures(ctx, db.DeleteStaleFailuresParams{
		StaleBefore: now.Add(-staleFailureAge),
		Now:         now,
	}); err != nil {
		s.logger.ErrorContext(ctx, "古い失敗の記録を消せません", "error", err)
	}
}
