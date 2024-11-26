BEGIN;

CREATE TABLE IF NOT EXISTS task_updates (
  id bigserial PRIMARY KEY,
  task_id bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  old_position integer NOT NULL,
  old_status task_status NOT NULL,
  new_position integer NOT NULL,
  new_status task_status NOT NULL,
  CONSTRAINT fk_task_id FOREIGN KEY (task_id) REFERENCES tasks ON DELETE
    CASCADE ON UPDATE CASCADE
);

ALTER TABLE IF EXISTS users
  DROP CONSTRAINT min_length_username,
  ADD CONSTRAINT nonempty_username CHECK (length(username) >= 1);

ALTER TABLE IF EXISTS forms
  DROP CONSTRAINT min_length_slug,
  ADD CONSTRAINT nonempty_slug CHECK (length(slug) >= 1);

COMMIT;
