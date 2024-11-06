BEGIN;

DROP TABLE IF EXISTS filled_forms;

TRUNCATE tasks CASCADE;

ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS form_version_id bigint NOT NULL,
    ADD CONSTRAINT tasks_fvid_fk FOREIGN KEY (form_version_id) REFERENCES
    form_versions ON DELETE CASCADE,
    DROP CONSTRAINT IF EXISTS tasks_slug_key,
    ADD CONSTRAINT tasks_slug_unique UNIQUE (form_version_id, slug);

COMMIT;
