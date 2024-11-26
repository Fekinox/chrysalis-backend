BEGIN;

DROP TABLE IF EXISTS task_updates;

ALTER TABLE IF EXISTS forms
  DROP CONSTRAINT nonempty_slug,
  ADD CONSTRAINT min_length_slug CHECK (length(slug) >= 6);

ALTER TABLE IF EXISTS users
  DROP CONSTRAINT nonempty_username,
  ADD CONSTRAINT min_length_username CHECK (length(username) >= 6);

COMMIT;
