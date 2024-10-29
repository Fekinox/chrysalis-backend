-- name: CreateForm :one
INSERT INTO
  forms (creator_id)
VALUES
  (sqlc.arg ('creator_id')) RETURNING (id, creator_id);

-- name: CreateFormVersion :one
INSERT INTO
  form_versions (form_id)
VALUES
  (sqlc.arg ('form_id')) RETURNING (id, created_at, form_id);

-- name: NumTasksOnVersion :one
SELECT
  COUNT(filled_forms.task_id)
FROM
  form_versions
  INNER JOIN filled_forms ON form_versions.id = filled_forms.form_version_id
  WHERE form_versions.id = sqlc.arg('form_version_id');


-- name: AssignCurrentFormVersion :one
INSERT INTO
  current_form_versions (form_id, form_version_id)
VALUES
  (sqlc.arg ('form_id'), sqlc.arg ('form_version_id')
ON CONFLICT (form_id) DO UPDATE SET form_version_id = EXCLUDED.form_id
RETURNING (form_id, form_version_id);

-- name: AddCheckboxFieldToFormVersion :one
WITH
  ff AS (
    INSERT INTO
      form_fields (form_version_id, idx, type, prompt, required)
    VALUES
      (
        sqlc.arg ('form_version_id'),
        sqlc.arg ('idx'),
        'checkbox' sqlc.arg ('prompt'),
        sqlc.arg ('required')
      ) RETURNING (id, form_version_id, idx, type, prompt, required)
  )
INSERT INTO
  checkbox_fields (ff_id, options)
VALUES
  (ff.id, sqlc.arg ('options')) RETURNING (
  ff.id, ff.form_version_id,
  ff.idx,
  ff.type,
  ff.prompt,
  ff.required,
  options
);

-- name: AddRadioFieldToFormVersion :one
WITH
  ff AS (
    INSERT INTO
      form_fields (form_version_id, idx, type, prompt, required)
    VALUES
      (
        sqlc.arg ('form_version_id'),
        sqlc.arg ('idx'),
        'radio',
        sqlc.arg ('prompt'),
        sqlc.arg ('required')
      ) RETURNING (id, form_version_id, idx, type, prompt, required)
  )
INSERT INTO
  radio_fields (ff_id, options)
VALUES
  (ff.id, sqlc.arg ('options')) RETURNING (
  ff.id, ff.form_version_id,
  ff.idx,
  ff.type,
  ff.prompt,
  ff.required,
  options
);

-- name: AddTextFieldToFormVersion :one
WITH
  ff AS (
    INSERT INTO
      form_fields (form_version_id, idx, type, prompt, required)
    VALUES
      (
        sqlc.arg ('form_version_id'),
        sqlc.arg ('idx'),
        'text',
        sqlc.arg ('prompt'),
        sqlc.arg ('required')
      ) RETURNING (id, form_version_id, idx, type, prompt, required)
  )
INSERT INTO
  text_fields (ff_id, paragraph)
VALUES
  (ff.id, sqlc.arg ('paragraph')) RETURNING (
  ff.id, ff.form_version_id,
  ff.idx,
  ff.type,
  ff.prompt,
  ff.required,
  paragraph 
);

-- name: GetCurrentFormVersion :many
SELECT (
  forms.id,
  cfv.form_version_id,
  fv.created_at,
  ffs.type,
  ffs.prompt,
  ffs.required,
  ch_fs.options,
  r_fs.options,
  t_fs.paragraph
)
FROM
  forms
  INNER JOIN current_form_versions AS cfv ON forms.id = cfv.form_id
  INNER JOIN form_versions AS fv ON fv.id = cfv.form_version_id
  INNER JOIN form_fields AS ffs ON cfv.form_version_id = ffs.form_version_id
  LEFT JOIN checkbox_fields AS ch_fs ON ch_fs.ff_id = ffs.id
  LEFT JOIN radio_fields AS r_fs ON r_fs.ff_id = ffs.id
  LEFT JOIN text_fields AS t_fs ON t_fs.ff_id = ffs.id
WHERE forms.id = sqlc.arg('form_id')
ORDER BY ffs.idx;
