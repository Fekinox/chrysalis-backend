-- name: CreateTask :one
INSERT INTO tasks (form_version_id, client_id, task_name, task_summary, slug)
  VALUES ($1, $2, $3, $4, $5)
RETURNING
  id, client_id, form_version_id, task_name, task_summary, slug, created_at;

-- name: GetOutboundTasks :many
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
  client.username = sqlc.arg ('client_username')
ORDER BY
  ts.status ASC,
  ts.idx ASC;

-- name: GetInboundTasks :many
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
  creator.username = sqlc.arg ('creator_username')
ORDER BY
  ts.status ASC,
  ts.idx ASC;

-- name: GetServiceTasksBySlug :many
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
  forms.slug = sqlc.arg ('form_slug')
  AND users.username = sqlc.arg ('creator_username')
ORDER BY
  ts.status ASC,
  ts.idx ASC;

-- name: GetServiceTasksWithStatus :many
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
  forms.slug = sqlc.arg ('service')
  AND users.username = sqlc.arg ('username')
  AND ts.status = sqlc.arg ('status')
ORDER BY
  ts.idx ASC;

-- name: GetTaskHeader :one
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
  AND forms.slug = sqlc.arg ('form_slug')
  AND tasks.slug = sqlc.arg ('task_slug');

-- name: AddFilledFieldToTask :one
INSERT INTO filled_form_fields (task_id, idx, ftype, filled)
  VALUES ($1, $2, $3, $4)
RETURNING
  task_id, idx, ftype, filled;

-- name: AddCheckboxFieldToTask :one
INSERT INTO filled_checkbox_fields (task_id, idx, selected_options)
  VALUES ($1, $2, $3)
RETURNING
  task_id, idx, selected_options;

-- name: AddRadioFieldToTask :one
INSERT INTO filled_radio_fields (task_id, idx, selected_option)
  VALUES ($1, $2, $3)
RETURNING
  task_id, idx, selected_option;

-- name: AddTextFieldToTask :one
INSERT INTO filled_text_fields (task_id, idx, content)
  VALUES ($1, $2, $3)
RETURNING
  task_id, idx, content;

-- name: UpdateTaskStatus :many
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
  creator.username = sqlc.arg ('creator')
  AND forms.slug = sqlc.arg ('form_slug')
  AND tasks.slug = sqlc.arg ('task_slug')
  AND tasks.id = task_states.task_id
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
  fv.id = sqlc.arg ('form_version_id')
  AND tk.slug = sqlc.arg ('task_slug')
ORDER BY
  ffs.idx;

-- name: CreateTaskState :one
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
  status;

-- name: SwapTasks :exec
UPDATE
  task_states
SET
  idx = CASE WHEN task_id = sqlc.arg ('task_id_1') THEN
  (
    SELECT
      idx
    FROM
      task_states
    WHERE
      task_id = sqlc.arg ('task_id_2'))
  WHEN task_id = sqlc.arg ('task_id_2') THEN
  (
    SELECT
      idx
    FROM
      task_states
    WHERE
      task_id = sqlc.arg ('task_id_1'))
ELSE
  idx
  END
WHERE
  task_id = sqlc.arg ('task_id_1')
  OR task_id = sqlc.arg ('task_id_2');

-- name: RemoveTask :exec
UPDATE
  task_states
SET
  idx = - 1
WHERE
  task_id = sqlc.arg ('task_id_1');

-- name: InsertTask :exec
UPDATE
  task_states
SET
  idx = CASE WHEN task_id = sqlc.arg ('task_id') THEN
    sqlc.arg ('new_index')
  WHEN status = sqlc.arg ('status')
    AND idx >= sqlc.arg ('new_index') THEN
    idx + 1
  ELSE
    idx
  END,
  status = sqlc.arg ('status')
WHERE (status = sqlc.arg ('status')
  AND idx >= sqlc.arg ('new_index'))
  OR task_id = sqlc.arg ('task_id');

-- name: ReorderTaskStatuses :exec
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
  creators.username = sqlc.arg ('creator_username')
  AND forms.slug = sqlc.arg ('form_slug')
  AND task_states.task_id = new_indices.task_id
  AND idx <> - 1;

-- name: GetTaskByStatusAndIndex :one
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
  forms.slug = sqlc.arg ('form_slug')
  AND users.username = sqlc.arg ('creator_username')
  AND ts.status = sqlc.arg ('status')
  AND ts.idx = sqlc.arg ('idx');

-- name: GetTaskCounts :many
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
  creators.username = sqlc.arg ('username')
  AND fms.slug = sqlc.arg ('service')
GROUP BY
  ts.status;
