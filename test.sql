SELECT
  task_states.task_id,
  tasks.task_name,
  task_states.idx AS actual_index,
  expected_indices.idx AS expected_idx
FROM
  task_states
  INNER JOIN (
    SELECT
      task_id,
      (row_number() OVER (PARTITION BY status ORDER BY idx ASC) - 1) AS idx
    FROM
      task_states
      INNER JOIN tasks ON task_states.task_id = tasks.id
      INNER JOIN form_versions ON form_versions.id = tasks.form_version_id
      INNER JOIN forms ON forms.id = form_versions.form_id
      INNER JOIN users AS creators ON creators.id = forms.creator_id
    WHERE
      creators.username = 'bar' AND forms.slug = 'foobar'
    ) AS expected_indices ON expected_indices.task_id = task_states.task_id
  INNER JOIN tasks ON task_states.task_id = tasks.id;
