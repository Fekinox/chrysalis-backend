BEGIN;

ALTER TABLE IF EXISTS forms
  ADD CONSTRAINT min_length_slug CHECK (length(slug) >= 6);

ALTER TABLE IF EXISTS users
  ADD CONSTRAINT min_length_username CHECK (length(username) >= 6);

ALTER TABLE IF EXISTS form_versions
  ADD CONSTRAINT nonempty_name CHECK (length(name) >= 1);

COMMIT;
