// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: user.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
    username,
    password
) VALUES (
    $1,
    $2
) RETURNING
    id,
    username,
    password
`

type CreateUserParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Username, arg.Password)
	var i User
	err := row.Scan(&i.ID, &i.Username, &i.Password)
	return &i, err
}

const getUserByUUID = `-- name: GetUserByUUID :one
SELECT
    id,
    username,
    password
FROM
    users
WHERE
    id = $1
`

func (q *Queries) GetUserByUUID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByUUID, id)
	var i User
	err := row.Scan(&i.ID, &i.Username, &i.Password)
	return &i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT
    id,
    username,
    password
FROM
    users
WHERE
    username = $1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	row := q.db.QueryRow(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(&i.ID, &i.Username, &i.Password)
	return &i, err
}
