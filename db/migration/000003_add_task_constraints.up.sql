BEGIN;

ALTER TABLE filled_form_fields
    ALTER COLUMN filled SET NOT NULL;

COMMIT;
