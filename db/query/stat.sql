-- name: GetChrysalisStats :one
SELECT
	COUNT(users) AS count_users,
	COUNT(forms) AS num_forms,
    COUNT(tasks) AS num_tasks
FROM
    users,
    forms,
    tasks;

