let sortables = new Map()

function NewSortable(key, s) {
  if (key in sortables) {
    sortables.get(key).destroy()
    sortables.delete(key)
  }
  sortables.set(key, s)
  console.log(key)
}

function refreshCurrentTab() {
  htmx.trigger(document.querySelector("#dashboardcontent [aria-selected=true]"), "refresh")
}

function refreshColumn(status) {
  htmx.trigger(document.querySelector(`#dashboardcontent [data-dashboard-column][data-status=${status}]`), 'refresh')
}

async function updateStatus({ username, service, task, status }) {
  await api.put(`/app/${username}/services/${service}/tasks/${task}?status=${status}`);
  refreshCurrentTab();
}

function setHighlight(target) {
  currentTab = document.querySelector("#dashboardtabs [aria-selected]")
  currentTab.setAttribute("aria-selected", "false")

  target.setAttribute("aria-selected", "true")
}

function loadSortables(username, service, target) {
  setHighlight(target)
  if (target.dataset.tab == "tabs") {
    attachSortable(username, service)
  }
}

function attachSortable(username, service) {
  sts = document.querySelector('#dashboardcontent [aria-selected=true]').dataset.tab
  NewSortable("dashboard", Sortable.create(document.getElementById('tasktable'), {
    animation: 150,
    onEnd: async (ev) => {
      console.log(ev);
      params = new URLSearchParams({
        srcStatus: sts,
        dstStatus: sts,
        srcIndex: ev.oldIndex,
        dstIndex: ev.newIndex,
      })
      await api.post(`/api/users/${username}/services/${service}/move?${params}`);
      refreshCurrentTab();
    },
  }));
}

function attachColumnSortable(username, service, status) {
  column = document.querySelector(`#dashboardcontent [data-dashboard-column][data-status=${status}]`)
  NewSortable(`columns-${column.dataset.status}`, Sortable.create(column, {
    animation: 150,
    group: 'columns',
    onEnd: async (ev) => {
      console.log({
        to: ev.to.dataset.status,
        from: ev.from.dataset.status,
        oldIndex: ev.oldIndex,
        newIndex: ev.newIndex,
      })

      params = new URLSearchParams({
        srcStatus: ev.from.dataset.status,
        srcIndex: ev.oldIndex,
        dstStatus: ev.to.dataset.status,
        dstIndex: ev.newIndex
      })
      await api.post(`/api/users/${username}/services/${service}/move?${params}`);

      refreshColumn(ev.to.dataset.status)
      refreshColumn(ev.from.dataset.status)
    }
  }))
}
