BEGIN;

ALTER TABLE tasks
    ADD CONSTRAINT nonempty_task_name CHECK (task_name <> '');

COMMIT;
