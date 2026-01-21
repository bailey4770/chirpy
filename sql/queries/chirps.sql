-- name: CreateChirp :one
INSERT INTO chirps (created_at, updated_at, body, user_id)
VALUES (NOW(), NOW(), $1, $2)
RETURNING *;

-- name: FetchChirpsByAge :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: FetchChirpsFromUserByAge :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: FetchChirpByID :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;
