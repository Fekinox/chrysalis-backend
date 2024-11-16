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
INSERT INTO filled_checkbox_fields (
    task_id,
    idx,
    selected_options 
) VALUES (
    $1, $2, $3
) RETURNING
    task_id,
    idx,
    selected_options
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
INSERT INTO filled_form_fields (
    task_id,
    idx,
    ftype,
    filled
) VALUES (
    $1, $2, $3, $4
) RETURNING
    task_id,
    idx,
    ftype,
    filled
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
INSERT INTO filled_radio_fields (
    task_id,
    idx,
    selected_option
) VALUES (
    $1, $2, $3
) RETURNING
    task_id,
    idx,
    selected_option
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
INSERT INTO filled_text_fields (
    task_id,
    idx,
    content
) VALUES (
    $1, $2, $3
) RETURNING
    task_id,
    idx,
    content
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
INSERT INTO tasks (
    form_version_id,
    client_id,
    slug
) VALUES (
    $1, $2, $3
) RETURNING
    id,
    client_id,
    form_version_id,
    status,
    slug,
    created_at
`

type CreateTaskParams struct {
	FormVersionID int64     `json:"form_version_id"`
	ClientID      uuid.UUID `json:"client_id"`
	Slug          string    `json:"slug"`
}

type CreateTaskRow struct {
	ID            int64              `json:"id"`
	ClientID      uuid.UUID          `json:"client_id"`
	FormVersionID int64              `json:"form_version_id"`
	Status        TaskStatus         `json:"status"`
	Slug          string             `json:"slug"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (*CreateTaskRow, error) {
	row := q.db.QueryRow(ctx, createTask, arg.FormVersionID, arg.ClientID, arg.Slug)
	var i CreateTaskRow
	err := row.Scan(
		&i.ID,
		&i.ClientID,
		&i.FormVersionID,
		&i.Status,
		&i.Slug,
		&i.CreatedAt,
	)
	return &i, err
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
    fv.id = $1 AND
    tk.slug = $2
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
    status,
    tasks.slug AS task_slug,
    tasks.created_at
FROM
    tasks
    INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
    INNER JOIN forms ON forms.id = form_versions.form_id
    INNER JOIN users AS creator ON forms.creator_id = creator.id
WHERE
    creator.username = $1
`

type GetInboundTasksRow struct {
	FormID        int64              `json:"form_id"`
	Username      string             `json:"username"`
	FormVersionID int64              `json:"form_version_id"`
	Name          string             `json:"name"`
	FormSlug      string             `json:"form_slug"`
	TaskID        int64              `json:"task_id"`
	ClientID      uuid.UUID          `json:"client_id"`
	Status        TaskStatus         `json:"status"`
	TaskSlug      string             `json:"task_slug"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
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
			&i.Status,
			&i.TaskSlug,
			&i.CreatedAt,
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
    status,
    tasks.slug AS task_slug,
    tasks.created_at
FROM
    tasks
    INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
    INNER JOIN forms ON forms.id = form_versions.form_id
    INNER JOIN users AS client ON tasks.client_id = client.id
    INNER JOIN users AS creator ON forms.creator_id = creator.id
WHERE
    client.username = $1
`

type GetOutboundTasksRow struct {
	FormID        int64              `json:"form_id"`
	CreatorID     uuid.UUID          `json:"creator_id"`
	FormVersionID int64              `json:"form_version_id"`
	Name          string             `json:"name"`
	FormSlug      string             `json:"form_slug"`
	TaskID        int64              `json:"task_id"`
	ClientID      uuid.UUID          `json:"client_id"`
	Status        TaskStatus         `json:"status"`
	TaskSlug      string             `json:"task_slug"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
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
			&i.Status,
			&i.TaskSlug,
			&i.CreatedAt,
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
    status,
    tasks.slug,
    tasks.created_at
FROM
    tasks
    INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
    INNER JOIN forms ON forms.id = form_versions.form_id
    INNER JOIN users ON forms.creator_id = users.id
WHERE
    forms.slug = $1 AND
    users.username = $2
`

type GetServiceTasksBySlugParams struct {
	FormSlug        string `json:"form_slug"`
	CreatorUsername string `json:"creator_username"`
}

type GetServiceTasksBySlugRow struct {
	FormVersionID int64              `json:"form_version_id"`
	ID            int64              `json:"id"`
	ClientID      uuid.UUID          `json:"client_id"`
	Status        TaskStatus         `json:"status"`
	Slug          string             `json:"slug"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
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
			&i.Status,
			&i.Slug,
			&i.CreatedAt,
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

const getTaskFields = `-- name: GetTaskFields :one
SELECT 1 FROM tasks
`

func (q *Queries) GetTaskFields(ctx context.Context) (int32, error) {
	row := q.db.QueryRow(ctx, getTaskFields)
	var column_1 int32
	err := row.Scan(&column_1)
	return column_1, err
}

const getTaskHeader = `-- name: GetTaskHeader :one
SELECT
    tasks.form_version_id,
    tasks.id,
    tasks.client_id,
    status,
    tasks.created_at
FROM
    tasks
    INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
    INNER JOIN forms ON forms.id = form_versions.form_id
    INNER JOIN users ON forms.creator_id = users.id
WHERE
    users.username = $1 AND
    forms.slug = $2 AND
    tasks.slug = $3
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
	Status        TaskStatus         `json:"status"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
}

func (q *Queries) GetTaskHeader(ctx context.Context, arg GetTaskHeaderParams) (*GetTaskHeaderRow, error) {
	row := q.db.QueryRow(ctx, getTaskHeader, arg.Username, arg.FormSlug, arg.TaskSlug)
	var i GetTaskHeaderRow
	err := row.Scan(
		&i.FormVersionID,
		&i.ID,
		&i.ClientID,
		&i.Status,
		&i.CreatedAt,
	)
	return &i, err
}

const updateTaskStatus = `-- name: UpdateTaskStatus :many
UPDATE tasks
    SET status = $1
    FROM
        form_versions
        JOIN forms ON form_versions.form_id = forms.id
        JOIN users AS creator ON forms.creator_id = creator.id
    WHERE
        form_versions.id = tasks.form_version_id AND
        creator.username = $2 AND
        forms.slug = $3 AND
        tasks.slug = $4
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
