BEGIN;

ALTER TABLE forms
    ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP;

COMMIT;