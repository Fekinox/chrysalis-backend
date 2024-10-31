-- name: CreateTask :one
INSERT INTO tasks (
    client_id,
    slug
) VALUES (
    sqlc.arg('client_id'),
    sqlc.arg('slug')
) RETURNING
    id,
    client_id,
    status,
    slug,
    created_at;

-- name: AddFormToTask :one
INSERT INTO filled_forms (
    task_id,
    form_version_id
) VALUES (
    sqlc.arg('task_id'),
    sqlc.arg('form_version_id')
) RETURNING
    task_id,
    form_version_id;

-- name: GetClientTasks :many
SELECT
    id,
    client_id,
    status,
    slug,
    created_at
FROM
    tasks
WHERE
    client_id = sqlc.arg('client_id');

-- name: GetServiceTasks :many
SELECT
    tasks.id,
    tasks.client_id,
    status,
    tasks.slug,
    tasks.created_at
FROM
    tasks
    INNER JOIN filled_forms ON filled_forms.task_id = tasks.id
    INNER JOIN form_versions ON filled_forms.form_version_id = form_versions.id
WHERE
    form_versions.form_id = sqlc.arg('form_id');

-- name: AddFilledFieldToTask :one
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
    filled;

-- name: AddCheckboxFieldToTask :one
INSERT INTO filled_checkbox_fields (
    task_id,
    idx,
    selected_options 
) VALUES (
    $1, $2, $3
) RETURNING
    task_id,
    idx,
    selected_options;

-- name: AddRadioFieldToTask :one
INSERT INTO filled_radio_fields (
    task_id,
    idx,
    selected_option
) VALUES (
    $1, $2, $3
) RETURNING
    task_id,
    idx,
    selected_option;
-- name: AddTextFieldToTask :one
INSERT INTO filled_text_fields (
    task_id,
    idx,
    content
) VALUES (
    $1, $2, $3
) RETURNING
    task_id,
    idx,
    content;
