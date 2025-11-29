-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hash_password)
VALUES (
    gen_random_uuid(),Now(),Now(), $1, $2
)
RETURNING *;

-- name: GetUserWithEmail :one
SELECT * FROM users where email = $1;

-- name: UpdateUserEmail :exec
UPDATE users SET email = $2, hash_password = $3, updated_at = Now() where id = $1;

