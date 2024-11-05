// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: form.sql

package db

import (
	"context"

	"github.com/google/uuid"
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
	Prompt        string    `json:"prompt"`
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

const addRadioFieldToForm = `-- name: AddRadioFieldToForm :one
INSERT INTO radio_fields (
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

type AddRadioFieldToFormParams struct {
	FormVersionID int64    `json:"form_version_id"`
	Idx           int64    `json:"idx"`
	Options       []string `json:"options"`
}

func (q *Queries) AddRadioFieldToForm(ctx context.Context, arg AddRadioFieldToFormParams) (*RadioField, error) {
	row := q.db.QueryRow(ctx, addRadioFieldToForm, arg.FormVersionID, arg.Idx, arg.Options)
	var i RadioField
	err := row.Scan(&i.FormVersionID, &i.Idx, &i.Options)
	return &i, err
}

const addTextFieldToForm = `-- name: AddTextFieldToForm :one
INSERT INTO text_fields (
    form_version_id,
    idx,
    paragraph
) VALUES (
    $1,
    $2,
    $3
) RETURNING
    form_version_id,
    idx,
    paragraph
`

type AddTextFieldToFormParams struct {
	FormVersionID int64 `json:"form_version_id"`
	Idx           int64 `json:"idx"`
	Paragraph     bool  `json:"paragraph"`
}

func (q *Queries) AddTextFieldToForm(ctx context.Context, arg AddTextFieldToFormParams) (*TextField, error) {
	row := q.db.QueryRow(ctx, addTextFieldToForm, arg.FormVersionID, arg.Idx, arg.Paragraph)
	var i TextField
	err := row.Scan(&i.FormVersionID, &i.Idx, &i.Paragraph)
	return &i, err
}

const assignCurrentFormVersion = `-- name: AssignCurrentFormVersion :one
INSERT INTO current_form_versions (form_id, form_version_id)
  VALUES ($1, $2)
ON CONFLICT (form_id)
  DO UPDATE SET
    form_version_id = EXCLUDED.form_version_id
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
INSERT INTO forms (creator_id, slug)
  VALUES ($1, $2)
RETURNING id, creator_id, slug
`

type CreateFormParams struct {
	CreatorID uuid.UUID `json:"creator_id"`
	Slug      string    `json:"slug"`
}

type CreateFormRow struct {
	ID        int64     `json:"id"`
	CreatorID uuid.UUID `json:"creator_id"`
	Slug      string    `json:"slug"`
}

func (q *Queries) CreateForm(ctx context.Context, arg CreateFormParams) (*CreateFormRow, error) {
	row := q.db.QueryRow(ctx, createForm, arg.CreatorID, arg.Slug)
	var i CreateFormRow
	err := row.Scan(&i.ID, &i.CreatorID, &i.Slug)
	return &i, err
}

const createFormVersion = `-- name: CreateFormVersion :one
INSERT INTO form_versions (
    form_id,
    name,
    description
) VALUES (
    $1,
    $2,
    $3
)
RETURNING id, name, description, created_at, form_id
`

type CreateFormVersionParams struct {
	FormID      int64  `json:"form_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (q *Queries) CreateFormVersion(ctx context.Context, arg CreateFormVersionParams) (*FormVersion, error) {
	row := q.db.QueryRow(ctx, createFormVersion, arg.FormID, arg.Name, arg.Description)
	var i FormVersion
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.CreatedAt,
		&i.FormID,
	)
	return &i, err
}

const deleteForm = `-- name: DeleteForm :exec
DELETE FROM forms WHERE forms.slug = $1 AND forms.creator_id = $2
`

type DeleteFormParams struct {
	Slug      string    `json:"slug"`
	CreatorID uuid.UUID `json:"creator_id"`
}

func (q *Queries) DeleteForm(ctx context.Context, arg DeleteFormParams) error {
	_, err := q.db.Exec(ctx, deleteForm, arg.Slug, arg.CreatorID)
	return err
}

const getCurrentFormVersionBySlug = `-- name: GetCurrentFormVersionBySlug :one
SELECT
  forms.id,
  forms.creator_id,
  forms.slug,
  fv.id AS form_version_id,
  fv.name,
  fv.description,
  forms.created_at,
  fv.created_at AS updated_at
FROM
  forms
  INNER JOIN current_form_versions AS cfv ON forms.id = cfv.form_id
  INNER JOIN form_versions AS fv ON fv.id = cfv.form_version_id
WHERE
  forms.slug = $1 AND
  forms.creator_id = $2
`

type GetCurrentFormVersionBySlugParams struct {
	Slug      string    `json:"slug"`
	CreatorID uuid.UUID `json:"creator_id"`
}

type GetCurrentFormVersionBySlugRow struct {
	ID            int64              `json:"id"`
	CreatorID     uuid.UUID          `json:"creator_id"`
	Slug          string             `json:"slug"`
	FormVersionID int64              `json:"form_version_id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetCurrentFormVersionBySlug(ctx context.Context, arg GetCurrentFormVersionBySlugParams) (*GetCurrentFormVersionBySlugRow, error) {
	row := q.db.QueryRow(ctx, getCurrentFormVersionBySlug, arg.Slug, arg.CreatorID)
	var i GetCurrentFormVersionBySlugRow
	err := row.Scan(
		&i.ID,
		&i.CreatorID,
		&i.Slug,
		&i.FormVersionID,
		&i.Name,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const getFormFields = `-- name: GetFormFields :many
SELECT
  ffs.ftype,
  ffs.prompt,
  ffs.required,
  ch_fs.options AS "checkbox_options",
  r_fs.options AS "radio_options",
  t_fs.paragraph AS "text_paragraph"
FROM
  form_versions AS fv
  INNER JOIN form_fields AS ffs ON fv.id = ffs.form_version_id
  LEFT JOIN checkbox_fields AS ch_fs USING (form_version_id, idx)
  LEFT JOIN radio_fields AS r_fs USING (form_version_id, idx)
  LEFT JOIN text_fields AS t_fs USING (form_version_id, idx)
WHERE
  fv.id = $1::bigint
ORDER BY
  ffs.idx
`

type GetFormFieldsRow struct {
	Ftype           FieldType `json:"ftype"`
	Prompt          string    `json:"prompt"`
	Required        bool      `json:"required"`
	CheckboxOptions []string  `json:"checkbox_options"`
	RadioOptions    []string  `json:"radio_options"`
	TextParagraph   *bool     `json:"text_paragraph"`
}

func (q *Queries) GetFormFields(ctx context.Context, formVersionID int64) ([]*GetFormFieldsRow, error) {
	rows, err := q.db.Query(ctx, getFormFields, formVersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetFormFieldsRow
	for rows.Next() {
		var i GetFormFieldsRow
		if err := rows.Scan(
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

const getFormHeaderBySlug = `-- name: GetFormHeaderBySlug :one
SELECT
    forms.id,
    forms.slug,
    forms.creator_id
FROM
    forms
WHERE
    forms.slug = $1 AND 
    forms.creator_id = $2
`

type GetFormHeaderBySlugParams struct {
	Slug      string    `json:"slug"`
	CreatorID uuid.UUID `json:"creator_id"`
}

type GetFormHeaderBySlugRow struct {
	ID        int64     `json:"id"`
	Slug      string    `json:"slug"`
	CreatorID uuid.UUID `json:"creator_id"`
}

func (q *Queries) GetFormHeaderBySlug(ctx context.Context, arg GetFormHeaderBySlugParams) (*GetFormHeaderBySlugRow, error) {
	row := q.db.QueryRow(ctx, getFormHeaderBySlug, arg.Slug, arg.CreatorID)
	var i GetFormHeaderBySlugRow
	err := row.Scan(&i.ID, &i.Slug, &i.CreatorID)
	return &i, err
}

const getUserFormHeaders = `-- name: GetUserFormHeaders :many
SELECT
    forms.id,
    forms.slug,
    forms.creator_id,
    fv.name,
    fv.description,
    forms.created_at,
    fv.created_at AS updated_at
FROM
    forms
    INNER JOIN current_form_versions AS cfv ON cfv.form_id = forms.id
    INNER JOIN form_versions AS fv ON cfv.form_version_id = fv.id
WHERE
    forms.creator_id = $1
ORDER BY updated_at DESC
`

type GetUserFormHeadersRow struct {
	ID          int64              `json:"id"`
	Slug        string             `json:"slug"`
	CreatorID   uuid.UUID          `json:"creator_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUserFormHeaders(ctx context.Context, creatorID uuid.UUID) ([]*GetUserFormHeadersRow, error) {
	rows, err := q.db.Query(ctx, getUserFormHeaders, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetUserFormHeadersRow
	for rows.Next() {
		var i GetUserFormHeadersRow
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.CreatorID,
			&i.Name,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
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