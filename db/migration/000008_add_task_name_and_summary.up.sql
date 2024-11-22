BEGIN;

ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS task_name text NOT NULL DEFAULT 'Task Name',
    ADD COLUMN IF NOT EXISTS task_summary text NOT NULL DEFAULT 'Task Summary';

COMMIT;
