BEGIN;

DO $$ BEGIN
    CREATE TYPE field_type AS ENUM (
        'checkbox',
        'radio',
        'text'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- All users (both creators and clients) in the database
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
  name varchar(256) NOT NULL
    CONSTRAINT user_name_unique UNIQUE,
  pass_hash varchar(256) NOT NULL
);

-- A form created by a specific user for other users to fill out
CREATE TABLE IF NOT EXISTS forms (
  id bigserial PRIMARY KEY,
  creator_id uuid NOT NULL,
  CONSTRAINT fk_creator_id
    FOREIGN KEY (creator_id)
    REFERENCES users
);

-- A specific version of a form
CREATE TABLE IF NOT EXISTS form_versions (
  id bigserial PRIMARY KEY,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  form_id bigint,
  CONSTRAINT fk_form_id
    FOREIGN KEY (form_id)
    REFERENCES forms
);

-- Relation identifying a form with its current version
CREATE TABLE IF NOT EXISTS current_form_versions (
  form_id bigint,
  form_version_id bigint,
  CONSTRAINT form_id_unique UNIQUE (form_id),
  CONSTRAINT form_v_unique UNIQUE (form_version_id),
  CONSTRAINT fk_form_id
    FOREIGN KEY (form_id)
    REFERENCES forms,
  CONSTRAINT fk_form_version_id
    FOREIGN KEY (form_version_id)
    REFERENCES form_versions
);

-- A field within a form
CREATE TABLE IF NOT EXISTS form_fields (
  id bigserial PRIMARY KEY,
  form_version_id bigint REFERENCES form_versions,
  idx bigint NOT NULL,

  type field_type NOT NULL,
  prompt text,
  required boolean,

  CONSTRAINT unique_idx UNIQUE (form_version_id, idx)
);

-- Checkbox field: select multiple of the given options
CREATE TABLE IF NOT EXISTS checkbox_fields (
  ff_id bigint NOT NULL,
  options text[],
  CONSTRAINT fk_ff_id
    FOREIGN KEY (ff_id)
    REFERENCES form_fields
);

-- Radio field: select one option
CREATE TABLE IF NOT EXISTS radio_fields (
  ff_id bigint NOT NULL,
  options text[],
  CONSTRAINT fk_ff_id
    FOREIGN KEY (ff_id)
    REFERENCES form_fields
);

-- Text field: either short answer (single line) or long answer (multiline)
CREATE TABLE IF NOT EXISTS text_fields (
  ff_id bigint NOT NULL,
  paragraph boolean DEFAULT false,
  CONSTRAINT fk_ff_id
    FOREIGN KEY (ff_id)
    REFERENCES form_fields
);

-- Task: Created by a client who fills out a form
CREATE TABLE IF NOT EXISTS tasks (
  id bigserial PRIMARY KEY,
  client_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status integer,
  CONSTRAINT fk_client_id
    FOREIGN KEY (client_id)
    REFERENCES users
);

-- Filled Form: The form that a given client has filled out
CREATE TABLE IF NOT EXISTS filled_forms (
  task_id bigint,
  form_version_id bigint,
  CONSTRAINT pk_task_fv_id PRIMARY KEY (task_id),
  CONSTRAINT fk_task_id
    FOREIGN KEY (task_id)
    REFERENCES tasks,
  CONSTRAINT fk_form_version_id
    FOREIGN KEY (form_version_id)
    REFERENCES form_versions
);

CREATE TABLE IF NOT EXISTS filled_form_fields (
  id bigserial PRIMARY KEY,
  task_id bigint,
  idx integer,

  type field_type NOT NULL,
  filled boolean
);

CREATE TABLE IF NOT EXISTS filled_checkbox_fields (
  ff_id bigint PRIMARY KEY,
  selected_options text[], 
  CONSTRAINT fk_ff_id
    FOREIGN KEY (ff_id)
    REFERENCES filled_form_fields
);

CREATE TABLE IF NOT EXISTS filled_radio_fields (
  ff_id bigint PRIMARY KEY,
  selected_option text,
  CONSTRAINT fk_ff_id
    FOREIGN KEY (ff_id)
    REFERENCES filled_form_fields
);

CREATE TABLE IF NOT EXISTS filled_text_fields (
  ff_id bigint PRIMARY KEY,
  content text,
  CONSTRAINT fk_ff_id
    FOREIGN KEY (ff_id)
    REFERENCES filled_form_fields
);

COMMIT;
