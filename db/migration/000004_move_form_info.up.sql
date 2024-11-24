BEGIN;

ALTER TABLE forms
  ADD COLUMN IF NOT EXISTS slug text UNIQUE;

UPDATE
  forms
SET
  slug = (
    SELECT
      form_versions.slug
    FROM
      form_versions
      INNER JOIN current_form_versions AS cfv ON cfv.form_version_id = form_versions.id
    WHERE
      cfv.form_id = forms.id);

ALTER TABLE form_versions
  DROP COLUMN IF EXISTS slug;

ALTER TABLE forms
  ALTER COLUMN slug SET NOT NULL;

COMMIT;
