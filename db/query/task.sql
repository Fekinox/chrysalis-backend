-- name: CreateTask :one
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
    created_at;

-- name: GetOutboundTasks :many
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
    client.username = sqlc.arg('client_username');

-- name: GetInboundTasks :many
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
    creator.username = sqlc.arg('creator_username');
    

-- name: GetServiceTasksBySlug :many
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
    forms.slug = sqlc.arg('form_slug') AND
    users.username = sqlc.arg('creator_username');

-- name: GetTaskHeader :one
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
    forms.slug = sqlc.arg('form_slug') AND
    tasks.slug = sqlc.arg('task_slug');

-- name: GetTaskFields :one
SELECT 1 FROM tasks;

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
