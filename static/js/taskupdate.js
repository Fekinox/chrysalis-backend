function readElement(el) {
  el.classList.add('read')

  el.querySelectorAll('button, a').forEach((button) => {
    button.setAttribute('disabled', 'true')
  })
}

async function markAsRead(username, service, task) {
  await api.post('/app/dashboard/mark-as-read', {
    username: username,
    service: service,
    task: task,
  });
  el = document.querySelector(`[data-task-update][data-creator='${username}'][data-service='${service}'][data-task='${task}']`)
  readElement(el)
}

async function markAllAsRead() {
  await api.post('/app/dashboard/mark-all-as-read');

  document.querySelectorAll('[data-task-update]').forEach(readElement)
}

async function markAllAsReadOnService(username, service) {
  document.querySelectorAll('[data-task-update]').forEach(readElement)
}
