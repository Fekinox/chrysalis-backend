// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: taskupdate.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const acknowledgeAllUpdatesForTask = `-- name: AcknowledgeAllUpdatesForTask :exec
UPDATE task_updates
  SET acknowledged = TRUE
  WHERE task_updates.task_id = $1
`

func (q *Queries) AcknowledgeAllUpdatesForTask(ctx context.Context, taskID int64) error {
	_, err := q.db.Exec(ctx, acknowledgeAllUpdatesForTask, taskID)
	return err
}

const acknowledgeAllUpdatesForUser = `-- name: AcknowledgeAllUpdatesForUser :exec
UPDATE task_updates
  SET acknowledged = TRUE
  FROM
    tasks
    INNER JOIN users AS clients ON clients.id = tasks.client_id
  WHERE
    clients.username = $1 AND
    task_updates.task_id = tasks.id
`

func (q *Queries) AcknowledgeAllUpdatesForUser(ctx context.Context, username string) error {
	_, err := q.db.Exec(ctx, acknowledgeAllUpdatesForUser, username)
	return err
}

const acknowledgeUpdate = `-- name: AcknowledgeUpdate :exec
UPDATE task_updates
  SET acknowledged = TRUE
  WHERE id = $1
`

func (q *Queries) AcknowledgeUpdate(ctx context.Context, taskUpdateID int64) error {
	_, err := q.db.Exec(ctx, acknowledgeUpdate, taskUpdateID)
	return err
}

const allNewUpdatesForUser = `-- name: AllNewUpdatesForUser :many
SELECT
  task_updates.id AS task_update_id,
  task_updates.created_at AS updated_at,
  task_updates.old_position,
  task_updates.old_status,
  task_updates.new_position,
  task_updates.new_status,
  tasks.task_name,
  tasks.task_summary,
  tasks.slug AS task_identifier,
  forms.slug AS form_identifier,
  form_versions.name AS form_name,
  creators.username AS form_creator
FROM
  task_updates
  INNER JOIN tasks ON tasks.id = task_updates.task_id
  INNER JOIN users AS clients ON clients.id = tasks.client_id
  INNER JOIN form_versions ON form_versions.id = tasks.form_version_id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users AS creators ON creators.id = forms.creator_id
WHERE
  clients.username = $1 AND
  NOT task_updates.acknowledged
ORDER BY task_updates.created_at ASC
`

type AllNewUpdatesForUserRow struct {
	TaskUpdateID   int64              `json:"task_update_id"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
	OldPosition    int32              `json:"old_position"`
	OldStatus      TaskStatus         `json:"old_status"`
	NewPosition    int32              `json:"new_position"`
	NewStatus      TaskStatus         `json:"new_status"`
	TaskName       string             `json:"task_name"`
	TaskSummary    string             `json:"task_summary"`
	TaskIdentifier string             `json:"task_identifier"`
	FormIdentifier string             `json:"form_identifier"`
	FormName       string             `json:"form_name"`
	FormCreator    string             `json:"form_creator"`
}

func (q *Queries) AllNewUpdatesForUser(ctx context.Context, username string) ([]*AllNewUpdatesForUserRow, error) {
	rows, err := q.db.Query(ctx, allNewUpdatesForUser, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*AllNewUpdatesForUserRow
	for rows.Next() {
		var i AllNewUpdatesForUserRow
		if err := rows.Scan(
			&i.TaskUpdateID,
			&i.UpdatedAt,
			&i.OldPosition,
			&i.OldStatus,
			&i.NewPosition,
			&i.NewStatus,
			&i.TaskName,
			&i.TaskSummary,
			&i.TaskIdentifier,
			&i.FormIdentifier,
			&i.FormName,
			&i.FormCreator,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createUpdate = `-- name: CreateUpdate :one
INSERT INTO task_updates (
  task_id,
  old_position,
  old_status,
  new_position,
  new_status
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING
  id,
  task_id,
  created_at,
  old_position,
  old_status,
  new_position,
  new_status,
  acknowledged
`

type CreateUpdateParams struct {
	TaskID      int64      `json:"task_id"`
	OldPosition int32      `json:"old_position"`
	OldStatus   TaskStatus `json:"old_status"`
	NewPosition int32      `json:"new_position"`
	NewStatus   TaskStatus `json:"new_status"`
}

func (q *Queries) CreateUpdate(ctx context.Context, arg CreateUpdateParams) (*TaskUpdate, error) {
	row := q.db.QueryRow(ctx, createUpdate,
		arg.TaskID,
		arg.OldPosition,
		arg.OldStatus,
		arg.NewPosition,
		arg.NewStatus,
	)
	var i TaskUpdate
	err := row.Scan(
		&i.ID,
		&i.TaskID,
		&i.CreatedAt,
		&i.OldPosition,
		&i.OldStatus,
		&i.NewPosition,
		&i.NewStatus,
		&i.Acknowledged,
	)
	return &i, err
}

const lastUpdateTime = `-- name: LastUpdateTime :one
SELECT
  created_at
FROM (
  SELECT
    created_at,
    row_number() OVER (PARTITION BY task_id ORDER BY created_at DESC) AS rn
  FROM
    task_updates
  WHERE
    task_id = $1) AS sub
WHERE
  sub.rn = 1
`

func (q *Queries) LastUpdateTime(ctx context.Context, taskID int64) (pgtype.Timestamptz, error) {
	row := q.db.QueryRow(ctx, lastUpdateTime, taskID)
	var created_at pgtype.Timestamptz
	err := row.Scan(&created_at)
	return created_at, err
}