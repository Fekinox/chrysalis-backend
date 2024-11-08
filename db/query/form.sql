-- name: CreateForm :one
INSERT INTO forms (creator_id, slug)
  VALUES (sqlc.arg ('creator_id'), sqlc.arg('slug'))
RETURNING id, creator_id, slug;

-- name: CreateFormVersion :one
INSERT INTO form_versions (
    form_id,
    name,
    description
) VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetUserFormHeaders :many
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
    forms.creator_id = sqlc.arg('creator_id')
ORDER BY updated_at DESC;


-- name: GetFormHeaderBySlug :one
SELECT
    forms.id,
    forms.slug,
    forms.creator_id,
    forms.created_at
FROM
    forms
WHERE
    forms.slug = sqlc.arg('slug') AND 
    forms.creator_id = sqlc.arg('creator_id');

-- name: AssignCurrentFormVersion :one
INSERT INTO current_form_versions (form_id, form_version_id)
  VALUES (sqlc.arg ('form_id'), sqlc.arg ('form_version_id'))
ON CONFLICT (form_id)
  DO UPDATE SET
    form_version_id = EXCLUDED.form_version_id
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

-- name: GetCurrentFormVersionBySlug :one
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
  forms.slug = sqlc.arg ('slug') AND
  forms.creator_id = sqlc.arg('creator_id');

-- name: GetFormFields :many
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
  fv.id = sqlc.arg ('form_version_id')::bigint
ORDER BY
  ffs.idx;

-- name: DeleteForm :exec
DELETE FROM forms WHERE forms.slug = $1 AND forms.creator_id = $2;

-- name: FindDuplicates :many
SELECT
    r_fv.id
FROM
    form_versions AS fv
    INNER JOIN form_versions AS r_fv USING (form_id, name, description)
WHERE
    fv.id = $1 AND
    r_fv.id <> $1 AND
    (
        SELECT COUNT(*) FROM form_fields
        WHERE form_fields.form_version_id = fv.id
    ) =
    (
        SELECT COUNT(*)
        FROM
            form_fields AS l_ffs
            LEFT JOIN checkbox_fields AS l_ch_fs USING (form_version_id, idx)
            LEFT JOIN radio_fields AS l_r_fs USING (form_version_id, idx)
            LEFT JOIN text_fields AS l_t_fs USING (form_version_id, idx)
            FULL OUTER JOIN
                form_fields AS r_ffs
                LEFT JOIN checkbox_fields AS r_ch_fs USING (form_version_id, idx)
                LEFT JOIN radio_fields AS r_r_fs USING (form_version_id, idx)
                LEFT JOIN text_fields AS r_t_fs USING (form_version_id, idx)
            USING (idx, ftype, prompt, required)
        WHERE
            l_ffs.form_version_id = fv.id AND
            r_ffs.form_version_id = r_fv.id AND
            CASE l_ffs.ftype
                WHEN 'checkbox' THEN
                    l_ch_fs.options = r_ch_fs.options          
                WHEN 'radio' THEN
                    l_r_fs.options = r_r_fs.options          
                ELSE
                    l_t_fs.paragraph = r_t_fs.paragraph
            END
    );
