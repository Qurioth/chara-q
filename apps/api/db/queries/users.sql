-- name: CreateUser :one
INSERT INTO users (login_id, display_name, password_hash, is_admin, created_at, updated_at)
VALUES (@login_id, @display_name, @password_hash, @is_admin, @now, @now)
RETURNING *;

-- name: GetUserByLoginID :one
-- ログイン ID は大文字と小文字を区別しない（users_login_id_key と同じ式で引く）。
SELECT * FROM users WHERE lower(login_id) = lower(@login_id);

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at, id;

-- name: UpdateDisplayName :one
UPDATE users SET display_name = @display_name, updated_at = @now
WHERE id = @id
RETURNING *;

-- name: UpdatePasswordHash :exec
UPDATE users SET password_hash = @password_hash, updated_at = @now
WHERE id = @id;

-- name: DisableUser :execrows
UPDATE users SET disabled_at = sqlc.arg(now)::timestamptz, updated_at = sqlc.arg(now)::timestamptz
WHERE id = @id AND disabled_at IS NULL;

-- name: EnableUser :execrows
UPDATE users SET disabled_at = NULL, updated_at = @now
WHERE id = @id AND disabled_at IS NOT NULL;
