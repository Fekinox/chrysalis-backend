CREATE TYPE field_type AS ENUM (
    'checkbox',
    'radio',
    'text'
);

BEGIN;

-- All users (both creators and clients) in the database
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  username varchar(256) NOT NULL,
  password varchar(256) NOT NULL,
  CONSTRAINT user_name_unique UNIQUE (username)
);

-- A form created by a specific user for other users to fill out
CREATE TABLE IF NOT EXISTS forms (
  id bigserial PRIMARY KEY,
  creator_id uuid NOT NULL,
  CONSTRAINT fk_creator_id FOREIGN KEY (creator_id) REFERENCES users ON DELETE CASCADE
);

-- A specific version of a form
CREATE TABLE IF NOT EXISTS form_versions (
  id bigserial PRIMARY KEY,
  name text NOT NULL,
  slug varchar(256) NOT NULL UNIQUE,
  description text,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  form_id bigint NOT NULL,
  CONSTRAINT fk_form_id FOREIGN KEY (form_id) REFERENCES forms ON DELETE CASCADE
);

-- Relation identifying a form with its current version
CREATE TABLE IF NOT EXISTS current_form_versions (
  form_id bigint NOT NULL,
  form_version_id bigint NOT NULL,
  CONSTRAINT form_id_unique UNIQUE (form_id),
  CONSTRAINT form_v_unique UNIQUE (form_version_id),
  CONSTRAINT fk_form_id FOREIGN KEY (form_id) REFERENCES forms ON DELETE CASCADE,
  CONSTRAINT fk_form_version_id FOREIGN KEY (form_version_id) REFERENCES
    form_versions ON DELETE CASCADE
);

-- A field within a form
CREATE TABLE IF NOT EXISTS form_fields (
  form_version_id bigint,
  idx bigint,
  ftype field_type NOT NULL,
  prompt text NOT NULL,
  required boolean NOT NULL,
  CONSTRAINT pk_form_field PRIMARY KEY (form_version_id, idx),
  CONSTRAINT fk_form_version_id FOREIGN KEY (form_version_id) REFERENCES
    form_versions ON DELETE CASCADE
);

-- Checkbox field: select multiple of the given options
CREATE TABLE IF NOT EXISTS checkbox_fields (
  form_version_id bigint,
  idx bigint,
  options text[],
  CONSTRAINT pk_checkbox_field PRIMARY KEY (form_version_id, idx),
  CONSTRAINT fk_ff_id FOREIGN KEY (form_version_id, idx) REFERENCES form_fields ON DELETE CASCADE
);

-- Radio field: select one option
CREATE TABLE IF NOT EXISTS radio_fields (
  form_version_id bigint,
  idx bigint,
  options text[],
  CONSTRAINT pk_radio_field PRIMARY KEY (form_version_id, idx),
  CONSTRAINT fk_ff_id FOREIGN KEY (form_version_id, idx) REFERENCES form_fields ON DELETE CASCADE
);

-- Text field: either short answer (single line) or long answer (multiline)
CREATE TABLE IF NOT EXISTS text_fields (
  form_version_id bigint,
  idx bigint,
  paragraph boolean DEFAULT FALSE,
  CONSTRAINT pk_text_field PRIMARY KEY (form_version_id, idx),
  CONSTRAINT fk_ff_id FOREIGN KEY (form_version_id, idx) REFERENCES form_fields ON DELETE CASCADE
);

-- Task: Created by a client who fills out a form
CREATE TABLE IF NOT EXISTS tasks (
  id bigserial PRIMARY KEY,
  client_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status integer,
  slug varchar(256) NOT NULL UNIQUE,
  CONSTRAINT fk_client_id FOREIGN KEY (client_id) REFERENCES users ON DELETE CASCADE
);

-- Filled Form: The form that a given client has filled out
CREATE TABLE IF NOT EXISTS filled_forms (
  task_id bigint,
  form_version_id bigint,
  CONSTRAINT pk_task_fv_id PRIMARY KEY (task_id),
  CONSTRAINT fk_task_id FOREIGN KEY (task_id) REFERENCES tasks ON DELETE CASCADE,
  CONSTRAINT fk_form_version_id FOREIGN KEY (form_version_id) REFERENCES
    form_versions ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS filled_form_fields (
  id bigserial PRIMARY KEY,
  task_id bigint,
  idx integer,
  ftype field_type NOT NULL,
  filled boolean,
  CONSTRAINT fk_task_id FOREIGN KEY (task_id) REFERENCES tasks ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS filled_checkbox_fields (
  ff_id bigint PRIMARY KEY,
  selected_options text[],
  CONSTRAINT fk_ff_id FOREIGN KEY (ff_id) REFERENCES filled_form_fields ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS filled_radio_fields (
  ff_id bigint PRIMARY KEY,
  selected_option text,
  CONSTRAINT fk_ff_id FOREIGN KEY (ff_id) REFERENCES filled_form_fields ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS filled_text_fields (
  ff_id bigint PRIMARY KEY,
  content text,
  CONSTRAINT fk_ff_id FOREIGN KEY (ff_id) REFERENCES filled_form_fields ON DELETE CASCADE
);

COMMIT;
