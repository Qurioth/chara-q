-- name: CreateInvitation :exec
INSERT INTO invitations (code_hash, issued_by, created_at, expires_at)
VALUES (@code_hash, @issued_by, @now, @expires_at);

-- name: LockUsableInvitation :one
-- 未使用で期限内の招待コードの行を押さえる。同時に使われても、2 つ目は使用済みの行を見て何も返さない。
SELECT id FROM invitations
WHERE code_hash = @code_hash
  AND used_at IS NULL
  AND expires_at > @now
FOR UPDATE;

-- name: MarkInvitationUsed :exec
UPDATE invitations SET used_at = sqlc.arg(now)::timestamptz, used_by = sqlc.arg(used_by)::uuid
WHERE id = @id;
