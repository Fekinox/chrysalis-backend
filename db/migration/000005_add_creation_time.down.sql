BEGIN;

ALTER TABLE forms
  DROP COLUMN IF EXISTS created_at;

COMMIT;
