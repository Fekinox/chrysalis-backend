BEGIN;

CREATE TABLE IF NOT EXISTS task_states (
    task_id bigint,
    idx integer NOT NULL,
    status task_status NOT NULL DEFAULT 'pending',

    CONSTRAINT pk_task_state_task_id PRIMARY KEY (task_id),
    CONSTRAINT fk_tasks_task_id FOREIGN KEY (task_id) REFERENCES tasks ON DELETE CASCADE ON UPDATE CASCADE
);

INSERT INTO task_states (
    task_id, idx, status 
)
SELECT
    tasks.id,
    (row_number() OVER (ORDER BY tasks.created_at ASC))-1,
    tasks.status
FROM
    tasks;

ALTER TABLE tasks
    DROP COLUMN IF EXISTS status;

COMMIT;
