// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: task.sql

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const addCheckboxFieldToTask = `-- name: AddCheckboxFieldToTask :one
INSERT INTO filled_checkbox_fields (task_id, idx, selected_options)
  VALUES ($1, $2, $3)
RETURNING
  task_id, idx, selected_options
`

type AddCheckboxFieldToTaskParams struct {
	TaskID          int64    `json:"task_id"`
	Idx             int32    `json:"idx"`
	SelectedOptions []string `json:"selected_options"`
}

func (q *Queries) AddCheckboxFieldToTask(ctx context.Context, arg AddCheckboxFieldToTaskParams) (*FilledCheckboxField, error) {
	row := q.db.QueryRow(ctx, addCheckboxFieldToTask, arg.TaskID, arg.Idx, arg.SelectedOptions)
	var i FilledCheckboxField
	err := row.Scan(&i.TaskID, &i.Idx, &i.SelectedOptions)
	return &i, err
}

const addFilledFieldToTask = `-- name: AddFilledFieldToTask :one
INSERT INTO filled_form_fields (task_id, idx, ftype, filled)
  VALUES ($1, $2, $3, $4)
RETURNING
  task_id, idx, ftype, filled
`

type AddFilledFieldToTaskParams struct {
	TaskID int64     `json:"task_id"`
	Idx    int32     `json:"idx"`
	Ftype  FieldType `json:"ftype"`
	Filled bool      `json:"filled"`
}

func (q *Queries) AddFilledFieldToTask(ctx context.Context, arg AddFilledFieldToTaskParams) (*FilledFormField, error) {
	row := q.db.QueryRow(ctx, addFilledFieldToTask,
		arg.TaskID,
		arg.Idx,
		arg.Ftype,
		arg.Filled,
	)
	var i FilledFormField
	err := row.Scan(
		&i.TaskID,
		&i.Idx,
		&i.Ftype,
		&i.Filled,
	)
	return &i, err
}

const addRadioFieldToTask = `-- name: AddRadioFieldToTask :one
INSERT INTO filled_radio_fields (task_id, idx, selected_option)
  VALUES ($1, $2, $3)
RETURNING
  task_id, idx, selected_option
`

type AddRadioFieldToTaskParams struct {
	TaskID         int64   `json:"task_id"`
	Idx            int32   `json:"idx"`
	SelectedOption *string `json:"selected_option"`
}

func (q *Queries) AddRadioFieldToTask(ctx context.Context, arg AddRadioFieldToTaskParams) (*FilledRadioField, error) {
	row := q.db.QueryRow(ctx, addRadioFieldToTask, arg.TaskID, arg.Idx, arg.SelectedOption)
	var i FilledRadioField
	err := row.Scan(&i.TaskID, &i.Idx, &i.SelectedOption)
	return &i, err
}

const addTextFieldToTask = `-- name: AddTextFieldToTask :one
INSERT INTO filled_text_fields (task_id, idx, content)
  VALUES ($1, $2, $3)
RETURNING
  task_id, idx, content
`

type AddTextFieldToTaskParams struct {
	TaskID  int64   `json:"task_id"`
	Idx     int32   `json:"idx"`
	Content *string `json:"content"`
}

func (q *Queries) AddTextFieldToTask(ctx context.Context, arg AddTextFieldToTaskParams) (*FilledTextField, error) {
	row := q.db.QueryRow(ctx, addTextFieldToTask, arg.TaskID, arg.Idx, arg.Content)
	var i FilledTextField
	err := row.Scan(&i.TaskID, &i.Idx, &i.Content)
	return &i, err
}

const createTask = `-- name: CreateTask :one
INSERT INTO tasks (form_version_id, client_id, task_name, task_summary, slug)
  VALUES ($1, $2, $3, $4, $5)
RETURNING
  id, client_id, form_version_id, task_name, task_summary, slug, created_at
`

type CreateTaskParams struct {
	FormVersionID int64     `json:"form_version_id"`
	ClientID      uuid.UUID `json:"client_id"`
	TaskName      string    `json:"task_name"`
	TaskSummary   string    `json:"task_summary"`
	Slug          string    `json:"slug"`
}

type CreateTaskRow struct {
	ID            int64              `json:"id"`
	ClientID      uuid.UUID          `json:"client_id"`
	FormVersionID int64              `json:"form_version_id"`
	TaskName      string             `json:"task_name"`
	TaskSummary   string             `json:"task_summary"`
	Slug          string             `json:"slug"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (*CreateTaskRow, error) {
	row := q.db.QueryRow(ctx, createTask,
		arg.FormVersionID,
		arg.ClientID,
		arg.TaskName,
		arg.TaskSummary,
		arg.Slug,
	)
	var i CreateTaskRow
	err := row.Scan(
		&i.ID,
		&i.ClientID,
		&i.FormVersionID,
		&i.TaskName,
		&i.TaskSummary,
		&i.Slug,
		&i.CreatedAt,
	)
	return &i, err
}

const createTaskState = `-- name: CreateTaskState :one
INSERT INTO task_states (task_id, idx)
  VALUES ($1, (
      SELECT
        COUNT(*)
      FROM
        task_states AS ts
        INNER JOIN tasks AS tks ON ts.task_id = tks.id
        INNER JOIN form_versions AS fvs ON fvs.id = tks.form_version_id
        INNER JOIN forms AS fms ON fms.id = fvs.form_id,
        tasks AS cur_task
        INNER JOIN form_versions AS cur_fv ON cur_fv.id = cur_task.form_version_id
        INNER JOIN forms AS cur_fms ON cur_fms.id = cur_fv.form_id
      WHERE
        cur_task.id = $1
        AND fms.id = cur_fms.id))
RETURNING
  task_id,
  idx,
  status
`

func (q *Queries) CreateTaskState(ctx context.Context, taskID int64) (*TaskState, error) {
	row := q.db.QueryRow(ctx, createTaskState, taskID)
	var i TaskState
	err := row.Scan(&i.TaskID, &i.Idx, &i.Status)
	return &i, err
}

const findDiscrepancies = `-- name: FindDiscrepancies :many
SELECT
  task_states.task_id,
  tasks.task_name,
  task_states.idx AS actual_index,
  expected_indices.idx AS expected_idx
FROM
  task_states
  INNER JOIN (
    SELECT
      task_id,
      (row_number() OVER (PARTITION BY status ORDER BY idx ASC) - 1) AS idx
    FROM
      task_states
      INNER JOIN tasks ON task_states.task_id = tasks.id
      INNER JOIN form_versions ON form_versions.id = tasks.form_version_id
      INNER JOIN forms ON forms.id = form_versions.form_id
      INNER JOIN users AS creators ON creators.id = forms.creator_id
    WHERE
      creators.username = $1
      AND forms.slug = $2
    ) AS expected_indices ON expected_indices.task_id = task_states.task_id
    INNER JOIN tasks ON task_states.task_id = tasks.id
WHERE
  expected_indices.idx <> task_states.idx
`

type FindDiscrepanciesParams struct {
	CreatorUsername string `json:"creator_username"`
	ServiceName     string `json:"service_name"`
}

type FindDiscrepanciesRow struct {
	TaskID      int64  `json:"task_id"`
	TaskName    string `json:"task_name"`
	ActualIndex int32  `json:"actual_index"`
	ExpectedIdx int64  `json:"expected_idx"`
}

func (q *Queries) FindDiscrepancies(ctx context.Context, arg FindDiscrepanciesParams) ([]*FindDiscrepanciesRow, error) {
	rows, err := q.db.Query(ctx, findDiscrepancies, arg.CreatorUsername, arg.ServiceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*FindDiscrepanciesRow
	for rows.Next() {
		var i FindDiscrepanciesRow
		if err := rows.Scan(
			&i.TaskID,
			&i.TaskName,
			&i.ActualIndex,
			&i.ExpectedIdx,
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

const getFilledFormFields = `-- name: GetFilledFormFields :many
SELECT
  ffs.ftype,
  ffs.filled,
  ch_fs.selected_options AS "checkbox_options",
  r_fs.selected_option AS "radio_option",
  t_fs.content AS "text_content"
FROM
  tasks AS tk
  INNER JOIN filled_form_fields AS ffs ON tk.id = ffs.task_id
  LEFT JOIN filled_checkbox_fields AS ch_fs USING (task_id, idx)
  LEFT JOIN filled_radio_fields AS r_fs USING (task_id, idx)
  LEFT JOIN filled_text_fields AS t_fs USING (task_id, idx)
  INNER JOIN form_versions AS fv ON tk.form_version_id = fv.id
WHERE
  fv.id = $1
  AND tk.slug = $2
ORDER BY
  ffs.idx
`

type GetFilledFormFieldsParams struct {
	FormVersionID int64  `json:"form_version_id"`
	TaskSlug      string `json:"task_slug"`
}

type GetFilledFormFieldsRow struct {
	Ftype           FieldType `json:"ftype"`
	Filled          bool      `json:"filled"`
	CheckboxOptions []string  `json:"checkbox_options"`
	RadioOption     *string   `json:"radio_option"`
	TextContent     *string   `json:"text_content"`
}

func (q *Queries) GetFilledFormFields(ctx context.Context, arg GetFilledFormFieldsParams) ([]*GetFilledFormFieldsRow, error) {
	rows, err := q.db.Query(ctx, getFilledFormFields, arg.FormVersionID, arg.TaskSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetFilledFormFieldsRow
	for rows.Next() {
		var i GetFilledFormFieldsRow
		if err := rows.Scan(
			&i.Ftype,
			&i.Filled,
			&i.CheckboxOptions,
			&i.RadioOption,
			&i.TextContent,
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

const getInboundTasks = `-- name: GetInboundTasks :many
SELECT
  forms.id AS form_id,
  creator.username,
  tasks.form_version_id,
  form_versions.name,
  forms.slug AS form_slug,
  tasks.id AS task_id,
  tasks.client_id,
  tasks.slug AS task_slug,
  tasks.created_at,
  client.username AS client_username,
  ts.status,
  ts.idx
FROM
  tasks
  INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users AS creator ON forms.creator_id = creator.id
  INNER JOIN users AS client ON tasks.client_id = client.id
  INNER JOIN task_states AS ts ON ts.task_id = tasks.id
WHERE
  creator.username = $1
ORDER BY
  ts.status ASC,
  ts.idx ASC
`

type GetInboundTasksRow struct {
	FormID         int64              `json:"form_id"`
	Username       string             `json:"username"`
	FormVersionID  int64              `json:"form_version_id"`
	Name           string             `json:"name"`
	FormSlug       string             `json:"form_slug"`
	TaskID         int64              `json:"task_id"`
	ClientID       uuid.UUID          `json:"client_id"`
	TaskSlug       string             `json:"task_slug"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	ClientUsername string             `json:"client_username"`
	Status         TaskStatus         `json:"status"`
	Idx            int32              `json:"idx"`
}

func (q *Queries) GetInboundTasks(ctx context.Context, creatorUsername string) ([]*GetInboundTasksRow, error) {
	rows, err := q.db.Query(ctx, getInboundTasks, creatorUsername)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetInboundTasksRow
	for rows.Next() {
		var i GetInboundTasksRow
		if err := rows.Scan(
			&i.FormID,
			&i.Username,
			&i.FormVersionID,
			&i.Name,
			&i.FormSlug,
			&i.TaskID,
			&i.ClientID,
			&i.TaskSlug,
			&i.CreatedAt,
			&i.ClientUsername,
			&i.Status,
			&i.Idx,
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

const getOutboundTasks = `-- name: GetOutboundTasks :many
SELECT
  forms.id AS form_id,
  forms.creator_id,
  tasks.form_version_id,
  form_versions.name,
  forms.slug AS form_slug,
  tasks.id AS task_id,
  tasks.client_id,
  tasks.slug AS task_slug,
  tasks.created_at,
  client.username AS client_username,
  creator.username AS client_username,
  ts.status,
  ts.idx
FROM
  tasks
  INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users AS client ON tasks.client_id = client.id
  INNER JOIN users AS creator ON forms.creator_id = creator.id
  INNER JOIN task_states AS ts ON ts.task_id = tasks.id
WHERE
  client.username = $1
ORDER BY
  ts.status ASC,
  ts.idx ASC
`

type GetOutboundTasksRow struct {
	FormID           int64              `json:"form_id"`
	CreatorID        uuid.UUID          `json:"creator_id"`
	FormVersionID    int64              `json:"form_version_id"`
	Name             string             `json:"name"`
	FormSlug         string             `json:"form_slug"`
	TaskID           int64              `json:"task_id"`
	ClientID         uuid.UUID          `json:"client_id"`
	TaskSlug         string             `json:"task_slug"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	ClientUsername   string             `json:"client_username"`
	ClientUsername_2 string             `json:"client_username_2"`
	Status           TaskStatus         `json:"status"`
	Idx              int32              `json:"idx"`
}

func (q *Queries) GetOutboundTasks(ctx context.Context, clientUsername string) ([]*GetOutboundTasksRow, error) {
	rows, err := q.db.Query(ctx, getOutboundTasks, clientUsername)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetOutboundTasksRow
	for rows.Next() {
		var i GetOutboundTasksRow
		if err := rows.Scan(
			&i.FormID,
			&i.CreatorID,
			&i.FormVersionID,
			&i.Name,
			&i.FormSlug,
			&i.TaskID,
			&i.ClientID,
			&i.TaskSlug,
			&i.CreatedAt,
			&i.ClientUsername,
			&i.ClientUsername_2,
			&i.Status,
			&i.Idx,
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

const getServiceTasksBySlug = `-- name: GetServiceTasksBySlug :many
SELECT
  tasks.form_version_id,
  tasks.id,
  tasks.client_id,
  clients.username AS client_username,
  tasks.slug,
  tasks.created_at,
  tasks.task_name,
  tasks.task_summary,
  ts.status,
  ts.idx
FROM
  tasks
  INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users ON forms.creator_id = users.id
  INNER JOIN users AS clients ON tasks.client_id = clients.id
  INNER JOIN task_states AS ts ON ts.task_id = tasks.id
WHERE
  forms.slug = $1
  AND users.username = $2
ORDER BY
  ts.status ASC,
  ts.idx ASC
`

type GetServiceTasksBySlugParams struct {
	FormSlug        string `json:"form_slug"`
	CreatorUsername string `json:"creator_username"`
}

type GetServiceTasksBySlugRow struct {
	FormVersionID  int64              `json:"form_version_id"`
	ID             int64              `json:"id"`
	ClientID       uuid.UUID          `json:"client_id"`
	ClientUsername string             `json:"client_username"`
	Slug           string             `json:"slug"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	TaskName       string             `json:"task_name"`
	TaskSummary    string             `json:"task_summary"`
	Status         TaskStatus         `json:"status"`
	Idx            int32              `json:"idx"`
}

func (q *Queries) GetServiceTasksBySlug(ctx context.Context, arg GetServiceTasksBySlugParams) ([]*GetServiceTasksBySlugRow, error) {
	rows, err := q.db.Query(ctx, getServiceTasksBySlug, arg.FormSlug, arg.CreatorUsername)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetServiceTasksBySlugRow
	for rows.Next() {
		var i GetServiceTasksBySlugRow
		if err := rows.Scan(
			&i.FormVersionID,
			&i.ID,
			&i.ClientID,
			&i.ClientUsername,
			&i.Slug,
			&i.CreatedAt,
			&i.TaskName,
			&i.TaskSummary,
			&i.Status,
			&i.Idx,
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

const getServiceTasksWithStatus = `-- name: GetServiceTasksWithStatus :many
SELECT
  tasks.form_version_id,
  tasks.id,
  tasks.client_id,
  clients.username AS client_username,
  tasks.slug,
  tasks.created_at,
  tasks.task_name,
  tasks.task_summary,
  ts.status,
  ts.idx
FROM
  tasks
  INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users ON forms.creator_id = users.id
  INNER JOIN users AS clients ON tasks.client_id = clients.id
  INNER JOIN task_states AS ts ON ts.task_id = tasks.id
WHERE
  forms.slug = $1
  AND users.username = $2
  AND ts.status = $3
ORDER BY
  ts.idx ASC
`

type GetServiceTasksWithStatusParams struct {
	Service  string     `json:"service"`
	Username string     `json:"username"`
	Status   TaskStatus `json:"status"`
}

type GetServiceTasksWithStatusRow struct {
	FormVersionID  int64              `json:"form_version_id"`
	ID             int64              `json:"id"`
	ClientID       uuid.UUID          `json:"client_id"`
	ClientUsername string             `json:"client_username"`
	Slug           string             `json:"slug"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	TaskName       string             `json:"task_name"`
	TaskSummary    string             `json:"task_summary"`
	Status         TaskStatus         `json:"status"`
	Idx            int32              `json:"idx"`
}

func (q *Queries) GetServiceTasksWithStatus(ctx context.Context, arg GetServiceTasksWithStatusParams) ([]*GetServiceTasksWithStatusRow, error) {
	rows, err := q.db.Query(ctx, getServiceTasksWithStatus, arg.Service, arg.Username, arg.Status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetServiceTasksWithStatusRow
	for rows.Next() {
		var i GetServiceTasksWithStatusRow
		if err := rows.Scan(
			&i.FormVersionID,
			&i.ID,
			&i.ClientID,
			&i.ClientUsername,
			&i.Slug,
			&i.CreatedAt,
			&i.TaskName,
			&i.TaskSummary,
			&i.Status,
			&i.Idx,
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

const getTaskByStatusAndIndex = `-- name: GetTaskByStatusAndIndex :one
SELECT
  tasks.form_version_id,
  tasks.id,
  tasks.client_id,
  clients.username AS client_username,
  tasks.slug,
  tasks.created_at,
  tasks.task_name,
  tasks.task_summary,
  ts.status,
  ts.idx
FROM
  tasks
  INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users ON forms.creator_id = users.id
  INNER JOIN users AS clients ON tasks.client_id = clients.id
  INNER JOIN task_states AS ts ON ts.task_id = tasks.id
WHERE
  forms.slug = $1
  AND users.username = $2
  AND ts.status = $3
  AND ts.idx = $4
`

type GetTaskByStatusAndIndexParams struct {
	FormSlug        string     `json:"form_slug"`
	CreatorUsername string     `json:"creator_username"`
	Status          TaskStatus `json:"status"`
	Idx             int32      `json:"idx"`
}

type GetTaskByStatusAndIndexRow struct {
	FormVersionID  int64              `json:"form_version_id"`
	ID             int64              `json:"id"`
	ClientID       uuid.UUID          `json:"client_id"`
	ClientUsername string             `json:"client_username"`
	Slug           string             `json:"slug"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	TaskName       string             `json:"task_name"`
	TaskSummary    string             `json:"task_summary"`
	Status         TaskStatus         `json:"status"`
	Idx            int32              `json:"idx"`
}

func (q *Queries) GetTaskByStatusAndIndex(ctx context.Context, arg GetTaskByStatusAndIndexParams) (*GetTaskByStatusAndIndexRow, error) {
	row := q.db.QueryRow(ctx, getTaskByStatusAndIndex,
		arg.FormSlug,
		arg.CreatorUsername,
		arg.Status,
		arg.Idx,
	)
	var i GetTaskByStatusAndIndexRow
	err := row.Scan(
		&i.FormVersionID,
		&i.ID,
		&i.ClientID,
		&i.ClientUsername,
		&i.Slug,
		&i.CreatedAt,
		&i.TaskName,
		&i.TaskSummary,
		&i.Status,
		&i.Idx,
	)
	return &i, err
}

const getTaskCounts = `-- name: GetTaskCounts :many
SELECT
  ts.status AS status,
  COUNT(ts.status) AS count
FROM
  task_states AS ts
  INNER JOIN tasks AS tks ON ts.task_id = tks.id
  INNER JOIN form_versions AS fvs ON fvs.id = tks.form_version_id
  INNER JOIN forms AS fms ON fms.id = fvs.form_id
  INNER JOIN users AS creators ON creators.id = fms.creator_id
WHERE
  creators.username = $1
  AND fms.slug = $2
GROUP BY
  ts.status
`

type GetTaskCountsParams struct {
	Username string `json:"username"`
	Service  string `json:"service"`
}

type GetTaskCountsRow struct {
	Status TaskStatus `json:"status"`
	Count  int64      `json:"count"`
}

func (q *Queries) GetTaskCounts(ctx context.Context, arg GetTaskCountsParams) ([]*GetTaskCountsRow, error) {
	rows, err := q.db.Query(ctx, getTaskCounts, arg.Username, arg.Service)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetTaskCountsRow
	for rows.Next() {
		var i GetTaskCountsRow
		if err := rows.Scan(&i.Status, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTaskHeader = `-- name: GetTaskHeader :one
SELECT
  tasks.form_version_id,
  tasks.id,
  tasks.client_id,
  tasks.task_name,
  tasks.task_summary,
  tasks.created_at,
  ts.status,
  ts.idx
FROM
  tasks
  INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users ON forms.creator_id = users.id
  INNER JOIN task_states AS ts ON ts.task_id = tasks.id
WHERE
  users.username = $1
  AND forms.slug = $2
  AND tasks.slug = $3
`

type GetTaskHeaderParams struct {
	Username string `json:"username"`
	FormSlug string `json:"form_slug"`
	TaskSlug string `json:"task_slug"`
}

type GetTaskHeaderRow struct {
	FormVersionID int64              `json:"form_version_id"`
	ID            int64              `json:"id"`
	ClientID      uuid.UUID          `json:"client_id"`
	TaskName      string             `json:"task_name"`
	TaskSummary   string             `json:"task_summary"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	Status        TaskStatus         `json:"status"`
	Idx           int32              `json:"idx"`
}

func (q *Queries) GetTaskHeader(ctx context.Context, arg GetTaskHeaderParams) (*GetTaskHeaderRow, error) {
	row := q.db.QueryRow(ctx, getTaskHeader, arg.Username, arg.FormSlug, arg.TaskSlug)
	var i GetTaskHeaderRow
	err := row.Scan(
		&i.FormVersionID,
		&i.ID,
		&i.ClientID,
		&i.TaskName,
		&i.TaskSummary,
		&i.CreatedAt,
		&i.Status,
		&i.Idx,
	)
	return &i, err
}

const insertTask = `-- name: InsertTask :exec
UPDATE
  task_states
SET
  idx = CASE WHEN task_id = $1 THEN
    $2
  WHEN status = $3
    AND idx >= $2 THEN
    idx + 1
  ELSE
    idx
  END,
  status = $3
WHERE (status = $3
  AND idx >= $2)
  OR task_id = $1
`

type InsertTaskParams struct {
	TaskID   int64      `json:"task_id"`
	NewIndex int32      `json:"new_index"`
	Status   TaskStatus `json:"status"`
}

func (q *Queries) InsertTask(ctx context.Context, arg InsertTaskParams) error {
	_, err := q.db.Exec(ctx, insertTask, arg.TaskID, arg.NewIndex, arg.Status)
	return err
}

const removeTask = `-- name: RemoveTask :exec
UPDATE
  task_states
SET
  idx = - 1
WHERE
  task_id = $1
`

func (q *Queries) RemoveTask(ctx context.Context, taskID1 int64) error {
	_, err := q.db.Exec(ctx, removeTask, taskID1)
	return err
}

const reorderTaskStatuses = `-- name: ReorderTaskStatuses :exec
UPDATE
  task_states
SET
  idx = new_indices.value
FROM (
  SELECT
    task_id,
    (row_number() OVER (PARTITION BY status ORDER BY idx ASC) - 1) AS value
  FROM
    task_states
  WHERE
    idx <> - 1) AS new_indices
  INNER JOIN tasks ON new_indices.task_id = tasks.id
  INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users AS creators ON creators.id = forms.creator_id
WHERE
  creators.username = $1
  AND forms.slug = $2
  AND task_states.task_id = new_indices.task_id
  AND idx <> - 1
`

type ReorderTaskStatusesParams struct {
	CreatorUsername string `json:"creator_username"`
	FormSlug        string `json:"form_slug"`
}

func (q *Queries) ReorderTaskStatuses(ctx context.Context, arg ReorderTaskStatusesParams) error {
	_, err := q.db.Exec(ctx, reorderTaskStatuses, arg.CreatorUsername, arg.FormSlug)
	return err
}

const swapTasks = `-- name: SwapTasks :exec
UPDATE
  task_states
SET
  idx = CASE WHEN task_id = $1 THEN
  (
    SELECT
      idx
    FROM
      task_states
    WHERE
      task_id = $2)
  WHEN task_id = $2 THEN
  (
    SELECT
      idx
    FROM
      task_states
    WHERE
      task_id = $1)
ELSE
  idx
  END
WHERE
  task_id = $1
  OR task_id = $2
`

type SwapTasksParams struct {
	TaskID1 *int64 `json:"task_id_1"`
	TaskID2 *int64 `json:"task_id_2"`
}

func (q *Queries) SwapTasks(ctx context.Context, arg SwapTasksParams) error {
	_, err := q.db.Exec(ctx, swapTasks, arg.TaskID1, arg.TaskID2)
	return err
}

const updatePositionAndStatus = `-- name: UpdatePositionAndStatus :exec
UPDATE
  task_states
SET
  idx = $1,
  status = $2
WHERE
  task_id = $3
`

type UpdatePositionAndStatusParams struct {
	Idx    int32      `json:"idx"`
	Status TaskStatus `json:"status"`
	ID     int64      `json:"id"`
}

func (q *Queries) UpdatePositionAndStatus(ctx context.Context, arg UpdatePositionAndStatusParams) error {
	_, err := q.db.Exec(ctx, updatePositionAndStatus, arg.Idx, arg.Status, arg.ID)
	return err
}

const updateTaskStatus = `-- name: UpdateTaskStatus :many
UPDATE
  task_states
SET
  status = $1
FROM
  form_versions
  JOIN forms ON form_versions.form_id = forms.id
  JOIN users AS creator ON forms.creator_id = creator.id
  INNER JOIN tasks ON tasks.form_version_id = form_versions.id
WHERE
  creator.username = $2
  AND forms.slug = $3
  AND tasks.slug = $4
  AND tasks.id = task_states.task_id
RETURNING
  1
`

type UpdateTaskStatusParams struct {
	Status   TaskStatus `json:"status"`
	Creator  string     `json:"creator"`
	FormSlug string     `json:"form_slug"`
	TaskSlug string     `json:"task_slug"`
}

func (q *Queries) UpdateTaskStatus(ctx context.Context, arg UpdateTaskStatusParams) ([]int32, error) {
	rows, err := q.db.Query(ctx, updateTaskStatus,
		arg.Status,
		arg.Creator,
		arg.FormSlug,
		arg.TaskSlug,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var column_1 int32
		if err := rows.Scan(&column_1); err != nil {
			return nil, err
		}
		items = append(items, column_1)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
