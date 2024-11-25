BEGIN;

CREATE TABLE IF NOT EXISTS task_updates (
    id bigserial PRIMARY KEY,
    task_id bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    old_position integer NOT NULL,
    old_status task_status NOT NULL,
    new_position integer NOT NULL,
    new_status task_status NOT NULL,

    CONSTRAINT fk_task_id FOREIGN KEY (task_id) REFERENCES tasks ON DELETE CASCADE ON UPDATE CASCADE
);

COMMIT;
