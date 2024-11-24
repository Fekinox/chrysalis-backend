BEGIN;

ALTER TABLE form_versions
  ADD COLUMN IF NOT EXISTS slug text UNIQUE;

UPDATE
  form_versions
SET
  slug = (
    SELECT
      forms.slug || '-' || form_versions.id
    FROM
      forms
      INNER JOIN current_form_versions AS cfv ON cfv.form_id = forms.id
    WHERE
      cfv.form_version_id = form_versions.id);

ALTER TABLE forms
  DROP COLUMN IF EXISTS slug;

ALTER TABLE form_versions
  ALTER COLUMN slug SET NOT NULL;

COMMIT;
