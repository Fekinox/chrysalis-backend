-- name: LastUpdateTime :one
SELECT
  created_at
FROM (
  SELECT
    created_at,
    row_number() OVER (PARTITION BY task_id ORDER BY created_at DESC) AS rn
  FROM
    task_updates
  WHERE
    task_id = sqlc.arg ('task_id')) AS sub
WHERE
  sub.rn = 1;

-- name: AllNewUpdatesForUser :many
SELECT
  task_updates.id AS task_update_id,
  task_updates.created_at AS updated_at,
  task_updates.old_position,
  task_updates.old_status,
  task_updates.new_position,
  task_updates.new_status,
  tasks.task_name,
  tasks.task_summary,
  tasks.slug AS task_identifier,
  forms.slug AS form_identifier,
  form_versions.name AS form_name,
  creators.username AS form_creator
FROM
  task_updates
  INNER JOIN tasks ON tasks.id = task_updates.task_id
  INNER JOIN users AS clients ON clients.id = tasks.client_id
  INNER JOIN form_versions ON form_versions.id = tasks.form_version_id
  INNER JOIN forms ON forms.id = form_versions.form_id
  INNER JOIN users AS creators ON creators.id = forms.creator_id
WHERE
  clients.username = sqlc.arg ('username')
  AND NOT task_updates.acknowledged
ORDER BY
  task_updates.created_at ASC;

-- name: CreateUpdate :one
INSERT INTO task_updates (task_id, old_position, old_status, new_position, new_status)
  VALUES ($1, $2, $3, $4, $5)
RETURNING
  id, task_id, created_at, old_position, old_status, new_position, new_status,
    acknowledged;

-- name: AcknowledgeUpdate :exec
UPDATE
  task_updates
SET
  acknowledged = TRUE
WHERE
  id = sqlc.arg ('task_update_id');

-- name: AcknowledgeAllUpdatesForUser :exec
UPDATE
  task_updates
SET
  acknowledged = TRUE
FROM
  tasks
  INNER JOIN users AS clients ON clients.id = tasks.client_id
WHERE
  clients.username = sqlc.arg ('username')
  AND task_updates.task_id = tasks.id;

-- name: AcknowledgeAllUpdatesForTask :exec
UPDATE
  task_updates
SET
  acknowledged = TRUE
WHERE
  task_updates.task_id = sqlc.arg ('task_id');
