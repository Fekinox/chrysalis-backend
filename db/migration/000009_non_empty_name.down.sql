BEGIN;

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS nonempty_task_name;

COMMIT;