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
    clients.username AS client_username,
    status,
    tasks.slug,
    tasks.created_at
FROM
    tasks
    INNER JOIN form_versions ON tasks.form_version_id = form_versions.id
    INNER JOIN forms ON forms.id = form_versions.form_id
    INNER JOIN users ON forms.creator_id = users.id
    INNER JOIN users AS clients ON tasks.client_id = clients.id
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

-- name: UpdateTaskStatus :many
UPDATE tasks
    SET status = $1
    FROM
        form_versions
        JOIN forms ON form_versions.form_id = forms.id
        JOIN users AS creator ON forms.creator_id = creator.id
    WHERE
        form_versions.id = tasks.form_version_id AND
        creator.username = sqlc.arg('creator') AND
        forms.slug = sqlc.arg('form_slug') AND
        tasks.slug = sqlc.arg('task_slug')
    RETURNING
        1;

-- name: GetFilledFormFields :many
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
    fv.id = sqlc.arg('form_version_id') AND
    tk.slug = sqlc.arg ('task_slug')
ORDER BY
    ffs.idx;
    
