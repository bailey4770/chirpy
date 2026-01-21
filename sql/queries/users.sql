-- name: CreateUser :one
INSERT INTO users (created_at, updated_at, email, hashed_password)
VALUES (NOW(), NOW(), $1, $2)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE $1=email;

-- name: UpdateEmailAndPassword :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: MakeUserRed :exec
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1;
