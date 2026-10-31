-- name: CreateSession :exec
INSERT INTO sessions (token_hash, user_id, created_at, last_used_at, expires_at)
VALUES (@token_hash, @user_id, @now, @now, @expires_at);

-- name: GetSessionViewer :one
-- 有効なセッションと、そのユーザーを 1 つのクエリで読む（無効なユーザーのセッションは返さない）。
SELECT
    s.last_used_at,
    u.id AS user_id,
    u.login_id,
    u.display_name,
    u.is_admin
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token_hash = @token_hash
  AND s.expires_at > @now
  AND u.disabled_at IS NULL;

-- name: TouchSession :exec
UPDATE sessions SET last_used_at = @now, expires_at = @expires_at
WHERE token_hash = @token_hash;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = @token_hash;

-- name: DeleteUserSessions :execrows
DELETE FROM sessions WHERE user_id = @user_id;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= @now;
