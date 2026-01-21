-- name: CreateChirp :one
INSERT INTO chirps (created_at, updated_at, body, user_id)
VALUES (NOW(), NOW(), $1, $2)
RETURNING *;

-- name: FetchChirpsWithOptionalParams :many
SELECT * FROM chirps
WHERE (NULLIF($1::uuid, '00000000-0000-0000-0000-000000000000') IS NULL
       OR user_id = $1)
ORDER BY
  CASE WHEN $2 = 'asc'  THEN created_at END ASC,
  CASE WHEN $2 = 'desc' THEN created_at END DESC;

-- name: FetchChirpByID :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;
