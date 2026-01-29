-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password_hash)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET updated_at = NOW(), email = $2, password_hash = $3
WHERE id = $1
RETURNING *;

-- name: UpgradeChirpyRed :exec
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1;

-- name: DowngradeChirpyRed :exec
UPDATE users
SET is_chirpy_red = FALSE
WHERE id = $1;