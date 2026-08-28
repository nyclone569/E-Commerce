-- name: CreateUser :one
INSERT INTO users (id, email, normalized_email, password_hash, display_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, normalized_email, password_hash, display_name, status, created_at, updated_at;

-- name: CreateSession :one
INSERT INTO sessions (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at;

-- name: GetUserCredentialByNormalizedEmail :one
SELECT id, email, normalized_email, password_hash, display_name, status, created_at, updated_at
FROM users
WHERE normalized_email = $1;

-- name: GetUserByActiveSession :one
SELECT u.id, u.email, u.display_name, u.status, u.created_at, u.updated_at
FROM sessions AS s
JOIN users AS u ON u.id = s.user_id
WHERE s.token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > $2
  AND u.status = 'active';

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = $2
WHERE token_hash = $1
  AND revoked_at IS NULL;
