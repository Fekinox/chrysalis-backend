BEGIN;

ALTER TABLE IF EXISTS tasks
    ALTER COLUMN status SET DATA TYPE integer
    USING
        CASE status
            WHEN 'pending' THEN
                0
            WHEN 'approved' THEN
                1
            WHEN 'in progress' THEN
                2
            WHEN 'delayed' THEN
                3
            WHEN 'complete' THEN
                4
            ELSE
                5
        END;

DROP TYPE IF EXISTS task_status;

COMMIT;
