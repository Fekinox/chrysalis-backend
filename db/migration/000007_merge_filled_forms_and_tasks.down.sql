BEGIN;

TRUNCATE tasks CASCADE;

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS tasks_slug_unique,
    ADD CONSTRAINT tasks_slug_key UNIQUE (slug),
    DROP CONSTRAINT tasks_fvid_fk,
    DROP COLUMN IF EXISTS form_version_id;

CREATE TABLE IF NOT EXISTS filled_forms (
  task_id bigint,
  form_version_id bigint,
  CONSTRAINT pk_task_fv_id PRIMARY KEY (task_id),
  CONSTRAINT fk_task_id FOREIGN KEY (task_id) REFERENCES tasks ON DELETE CASCADE,
  CONSTRAINT fk_form_version_id FOREIGN KEY (form_version_id) REFERENCES
    form_versions ON DELETE CASCADE
);


COMMIT;
