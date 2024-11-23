let sortables = new Map()

function NewSortable(key, s) {
  if (key in sortables) {
    sortables.get(key).destroy()
    sortables.delete(key)
  }
  sortables.set(key, s)
}

function refreshCurrentTab() {
  htmx.trigger(document.querySelector("#tabarea [aria-selected=true]"), "refresh")
}
async function updateStatus({ username, service, task, status }) {
  await fetch(`/app/${username}/services/${service}/tasks/${task}?status=${status}`, {
    method: "PUT",
  });
  refreshCurrentTab();
}

function attachSortable(username, service) {
  document.getElementById('tabarea').addEventListener('htmx:afterOnLoad', () => {
    sts = document.querySelector('#tabarea [aria-selected=true]').dataset.tab
    NewSortable("dashboard", Sortable.create(document.getElementById('tasktable'), {
      disable: true,
      animation: 150,
      onEnd: async (ev) => {
        console.log(ev);
        params = new URLSearchParams({
          status: sts,
          src: ev.oldIndex,
          dest: ev.newIndex,
        })
        await fetch(`/api/users/${username}/services/${service}/move?${params}`, {
          method: "POST",
        });
        refreshCurrentTab();
      },
    }));
  })
}
