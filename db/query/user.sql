-- name: CreateUser :one
INSERT INTO users (username, PASSWORD)
  VALUES (sqlc.arg ('username'), sqlc.arg ('password'))
RETURNING
  id, username, PASSWORD;

-- name: GetUserByUUID :one
SELECT
  id,
  username,
  PASSWORD
FROM
  users
WHERE
  id = sqlc.arg ('id');

-- name: GetUserByUsername :one
SELECT
  id,
  username,
  PASSWORD
FROM
  users
WHERE
  username = sqlc.arg ('username');
