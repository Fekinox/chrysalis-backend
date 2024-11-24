BEGIN;

ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS status task_status NOT NULL DEFAULT 'pending';

UPDATE
  tasks
SET
  status = (
    SELECT
      ts.status
    FROM
      task_states AS ts
      INNER JOIN tasks ON tasks.id = ts.task_id);

DROP TABLE IF EXISTS task_states;

COMMIT;
