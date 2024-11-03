-- name: CreateUser :one
INSERT INTO users (
    username,
    password
) VALUES (
    sqlc.arg('username'),
    sqlc.arg('password')
) RETURNING
    id,
    username,
    password;

-- name: GetUserByUUID :one
SELECT
    id,
    username,
    password
FROM
    users
WHERE
    id = sqlc.arg('id');

-- name: GetUserByUsername :one
SELECT
    id,
    username,
    password
FROM
    users
WHERE
    username = sqlc.arg('username');
