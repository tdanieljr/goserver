-- name: InsertToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
  $1,Now(),Now(), $2, $3
)
RETURNING *;

-- name: GetUserFromToken :one
SELECT user_id, revoked_at from refresh_tokens where token = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens SET updated_at = Now(), revoked_at = Now() where token = $1;
