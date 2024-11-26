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
  sub.rn = 1
