BEGIN;

CREATE TYPE task_status AS ENUM (
  'pending',
  'approved',
  'in progress',
  'delayed',
  'complete',
  'cancelled'
);

ALTER TABLE IF EXISTS tasks
  ALTER COLUMN status SET DATA TYPE task_status USING CASE status
  WHEN 0 THEN
    'pending'::task_status
  WHEN 1 THEN
    'approved'::task_status
  WHEN 2 THEN
    'in progress'::task_status
  WHEN 3 THEN
    'delayed'::task_status
  WHEN 4 THEN
    'complete'::task_status
  ELSE
    'cancelled'::task_status
  END;

COMMIT;
