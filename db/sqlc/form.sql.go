// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: form.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addCheckboxFieldToForm = `-- name: AddCheckboxFieldToForm :one
INSERT INTO checkbox_fields (
    form_version_id,
    idx,
    options
) VALUES (
    $1,
    $2,
    $3
) RETURNING
    form_version_id,
    idx,
    options
`

type AddCheckboxFieldToFormParams struct {
	FormVersionID int64    `json:"form_version_id"`
	Idx           int64    `json:"idx"`
	Options       []string `json:"options"`
}

func (q *Queries) AddCheckboxFieldToForm(ctx context.Context, arg AddCheckboxFieldToFormParams) (*CheckboxField, error) {
	row := q.db.QueryRow(ctx, addCheckboxFieldToForm, arg.FormVersionID, arg.Idx, arg.Options)
	var i CheckboxField
	err := row.Scan(&i.FormVersionID, &i.Idx, &i.Options)
	return &i, err
}

const addFormFieldToForm = `-- name: AddFormFieldToForm :one
INSERT INTO form_fields (
    form_version_id,
    idx,
    ftype,
    prompt,
    required
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING
    form_version_id,
    idx,
    ftype,
    prompt,
    required
`

type AddFormFieldToFormParams struct {
	FormVersionID int64     `json:"form_version_id"`
	Idx           int64     `json:"idx"`
	Ftype         FieldType `json:"ftype"`
	Prompt        *string   `json:"prompt"`
	Required      bool      `json:"required"`
}

func (q *Queries) AddFormFieldToForm(ctx context.Context, arg AddFormFieldToFormParams) (*FormField, error) {
	row := q.db.QueryRow(ctx, addFormFieldToForm,
		arg.FormVersionID,
		arg.Idx,
		arg.Ftype,
		arg.Prompt,
		arg.Required,
	)
	var i FormField
	err := row.Scan(
		&i.FormVersionID,
		&i.Idx,
		&i.Ftype,
		&i.Prompt,
		&i.Required,
	)
	return &i, err
}

const assignCurrentFormVersion = `-- name: AssignCurrentFormVersion :one
INSERT INTO current_form_versions (form_id, form_version_id)
  VALUES ($1, $2)
ON CONFLICT (form_id)
  DO UPDATE SET
    form_version_id = EXCLUDED.form_id
  RETURNING form_id, form_version_id
`

type AssignCurrentFormVersionParams struct {
	FormID        int64 `json:"form_id"`
	FormVersionID int64 `json:"form_version_id"`
}

func (q *Queries) AssignCurrentFormVersion(ctx context.Context, arg AssignCurrentFormVersionParams) (*CurrentFormVersion, error) {
	row := q.db.QueryRow(ctx, assignCurrentFormVersion, arg.FormID, arg.FormVersionID)
	var i CurrentFormVersion
	err := row.Scan(&i.FormID, &i.FormVersionID)
	return &i, err
}

const createForm = `-- name: CreateForm :one
INSERT INTO forms (creator_id)
  VALUES ($1)
RETURNING id, creator_id
`

func (q *Queries) CreateForm(ctx context.Context, creatorID pgtype.UUID) (*Form, error) {
	row := q.db.QueryRow(ctx, createForm, creatorID)
	var i Form
	err := row.Scan(&i.ID, &i.CreatorID)
	return &i, err
}

const createFormVersion = `-- name: CreateFormVersion :one
INSERT INTO form_versions (form_id)
  VALUES ($1)
RETURNING id, created_at, form_id
`

type CreateFormVersionRow struct {
	ID        int64              `json:"id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	FormID    int64              `json:"form_id"`
}

func (q *Queries) CreateFormVersion(ctx context.Context, formID int64) (*CreateFormVersionRow, error) {
	row := q.db.QueryRow(ctx, createFormVersion, formID)
	var i CreateFormVersionRow
	err := row.Scan(&i.ID, &i.CreatedAt, &i.FormID)
	return &i, err
}

const getCurrentFormVersion = `-- name: GetCurrentFormVersion :many
SELECT
  forms.id,
  cfv.form_version_id,
  fv.created_at,
  ffs.ftype,
  COALESCE(ffs.prompt, ''),
  ffs.required,
  ch_fs.options AS "checkbox_options",
  r_fs.options AS "radio_options",
  t_fs.paragraph AS "text_paragraph"
FROM
  forms
  INNER JOIN current_form_versions AS cfv ON forms.id = cfv.form_id
  INNER JOIN form_versions AS fv ON fv.id = cfv.form_version_id
  INNER JOIN form_fields AS ffs USING (form_version_id)
  LEFT JOIN checkbox_fields AS ch_fs USING (form_version_id, idx)
  LEFT JOIN radio_fields AS r_fs USING (form_version_id, idx)
  LEFT JOIN text_fields AS t_fs USING (form_version_id, idx)
WHERE
  forms.id = $1::bigint
ORDER BY
  ffs.idx
`

type GetCurrentFormVersionRow struct {
	ID              int64              `json:"id"`
	FormVersionID   int64              `json:"form_version_id"`
	CreatedAt       pgtype.Timestamptz `json:"created_at"`
	Ftype           FieldType          `json:"ftype"`
	Prompt          string             `json:"prompt"`
	Required        bool               `json:"required"`
	CheckboxOptions []string           `json:"checkbox_options"`
	RadioOptions    []string           `json:"radio_options"`
	TextParagraph   *bool              `json:"text_paragraph"`
}

func (q *Queries) GetCurrentFormVersion(ctx context.Context, formID int64) ([]*GetCurrentFormVersionRow, error) {
	rows, err := q.db.Query(ctx, getCurrentFormVersion, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetCurrentFormVersionRow
	for rows.Next() {
		var i GetCurrentFormVersionRow
		if err := rows.Scan(
			&i.ID,
			&i.FormVersionID,
			&i.CreatedAt,
			&i.Ftype,
			&i.Prompt,
			&i.Required,
			&i.CheckboxOptions,
			&i.RadioOptions,
			&i.TextParagraph,
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

const numTasksOnVersion = `-- name: NumTasksOnVersion :one
SELECT
  COUNT(filled_forms.task_id)
FROM
  form_versions
  INNER JOIN filled_forms ON form_versions.id = filled_forms.form_version_id
WHERE
  form_versions.id = $1
`

func (q *Queries) NumTasksOnVersion(ctx context.Context, formVersionID int64) (int64, error) {
	row := q.db.QueryRow(ctx, numTasksOnVersion, formVersionID)
	var count int64
	err := row.Scan(&count)
	return count, err
}
