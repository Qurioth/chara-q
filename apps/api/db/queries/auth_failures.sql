-- name: GetLockedUntil :one
-- ロック中なら、ロックが解ける日時を返す（ロック中でなければ行を返さない）。
SELECT locked_until::timestamptz AS locked_until FROM auth_failures
WHERE purpose = @purpose
  AND login_id_key = @login_id_key
  AND locked_until > sqlc.arg(now)::timestamptz;

-- name: RecordFailure :one
-- 失敗を 1 回数える。ロックの時間が過ぎていれば 1 回目から数え直し、max_failures 回になったらロックする。
-- 同時の失敗でも数え漏れないよう、1 文で行う。
INSERT INTO auth_failures AS f (purpose, login_id_key, failure_count, last_failed_at, locked_until)
VALUES (
    @purpose,
    @login_id_key,
    1,
    @now,
    CASE WHEN 1 >= sqlc.arg(max_failures)::int THEN sqlc.arg(lock_until)::timestamptz END
)
ON CONFLICT (purpose, login_id_key) DO UPDATE SET
    failure_count = CASE
        WHEN f.locked_until IS NOT NULL AND f.locked_until <= excluded.last_failed_at THEN 1
        ELSE f.failure_count + 1
    END,
    last_failed_at = excluded.last_failed_at,
    locked_until = CASE
        WHEN (CASE
            WHEN f.locked_until IS NOT NULL AND f.locked_until <= excluded.last_failed_at THEN 1
            ELSE f.failure_count + 1
        END) >= sqlc.arg(max_failures)::int THEN sqlc.arg(lock_until)::timestamptz
    END
RETURNING failure_count, locked_until;

-- name: ClearFailures :exec
DELETE FROM auth_failures WHERE purpose = @purpose AND login_id_key = @login_id_key;

-- name: DeleteStaleFailures :exec
-- 最後の失敗から時間がたち、ロック中でない記録を消す（存在しないログイン ID の行が増え続けないようにする）。
DELETE FROM auth_failures
WHERE last_failed_at < @stale_before
  AND (locked_until IS NULL OR locked_until <= sqlc.arg(now)::timestamptz);
