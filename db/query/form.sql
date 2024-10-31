-- name: CreateForm :one
INSERT INTO forms (creator_id)
  VALUES (sqlc.arg ('creator_id'))
RETURNING id, creator_id;

-- name: CreateFormVersion :one
INSERT INTO form_versions (
    form_id,
    name,
    slug,
    description
) VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: NumTasksOnVersion :one
SELECT
  COUNT(filled_forms.task_id)
FROM
  form_versions
  INNER JOIN filled_forms ON form_versions.id = filled_forms.form_version_id
WHERE
  form_versions.id = sqlc.arg ('form_version_id');

-- name: AssignCurrentFormVersion :one
INSERT INTO current_form_versions (form_id, form_version_id)
  VALUES (sqlc.arg ('form_id'), sqlc.arg ('form_version_id'))
ON CONFLICT (form_id)
  DO UPDATE SET
    form_version_id = EXCLUDED.form_id
  RETURNING form_id, form_version_id;

-- name: AddFormFieldToForm :one
INSERT INTO form_fields (
    form_version_id,
    idx,
    ftype,
    prompt,
    required
) VALUES (
    sqlc.arg('form_version_id'),
    sqlc.arg('idx'),
    sqlc.arg('ftype'),
    sqlc.arg('prompt'),
    sqlc.arg('required')
) RETURNING
    form_version_id,
    idx,
    ftype,
    prompt,
    required;

-- name: AddCheckboxFieldToForm :one
INSERT INTO checkbox_fields (
    form_version_id,
    idx,
    options
) VALUES (
    sqlc.arg('form_version_id'),
    sqlc.arg('idx'),
    sqlc.arg('options')
) RETURNING
    form_version_id,
    idx,
    options;

-- name: AddRadioFieldToForm :one
INSERT INTO radio_fields (
    form_version_id,
    idx,
    options
) VALUES (
    sqlc.arg('form_version_id'),
    sqlc.arg('idx'),
    sqlc.arg('options')
) RETURNING
    form_version_id,
    idx,
    options;

-- name: AddTextFieldToForm :one
INSERT INTO text_fields (
    form_version_id,
    idx,
    paragraph
) VALUES (
    sqlc.arg('form_version_id'),
    sqlc.arg('idx'),
    sqlc.arg('paragraph')
) RETURNING
    form_version_id,
    idx,
    paragraph;

-- name: GetCurrentFormVersion :many
SELECT
  forms.id,
  cfv.form_version_id,
  fv.created_at,
  ffs.ftype,
  ffs.prompt,
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
  forms.id = sqlc.arg ('form_id')::bigint
ORDER BY
  ffs.idx;
