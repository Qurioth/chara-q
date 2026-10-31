-- name: DeleteUnusedResetCodes :exec
-- 新しい再設定コードを発行するときに、前の未使用のコードを使えなくする。
DELETE FROM password_reset_codes WHERE user_id = @user_id AND used_at IS NULL;

-- name: CreateResetCode :exec
INSERT INTO password_reset_codes (user_id, code_hash, created_at, expires_at)
VALUES (@user_id, @code_hash, @now, @expires_at);

-- name: LockUsableResetCode :one
-- そのユーザー用の、未使用で期限内の再設定コードの行を押さえる。
SELECT id FROM password_reset_codes
WHERE user_id = @user_id
  AND code_hash = @code_hash
  AND used_at IS NULL
  AND expires_at > @now
FOR UPDATE;

-- name: MarkResetCodeUsed :exec
UPDATE password_reset_codes SET used_at = sqlc.arg(now)::timestamptz WHERE id = @id;
