BEGIN;

ALTER TABLE filled_form_fields
    ALTER COLUMN filled DROP NOT NULL;

COMMIT;
