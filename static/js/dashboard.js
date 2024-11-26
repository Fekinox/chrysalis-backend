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

document.addEventListener('alpine:init', () => {
  console.log('initting')
  Alpine.data('board', () => ({
    updates: [],

    constructor() {
      console.log('hi')
    },

    reset() {
      this.updates = []
      document.querySelectorAll('#dashboardcontent [data-dashboard-column]').forEach((col) => {
        col.dispatchEvent(new Event('refresh'))
      })
    },

    async commit(username, service) {
      await api.post(`/api/users/${username}/services/${service}/update`, 
        this.updates.map((update) => ({
          "task": update["task_identifier"],
          "new_index": update["new_index"],
          "new_status": update["new_status"]
        }))
      );
      this.reset()
    },

    attachColumnSortable(username, service, status) {
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

          this.checkUpdates()
        }
      }));
    },

    checkUpdates() {
      this.updates = []
      cols = document.querySelectorAll('#dashboardcontent [data-dashboard-column]');
      for (col of cols) {
        tasks = col.querySelectorAll('[data-task]')
        for (let i = 0; i < tasks.length; i++) {
          task = tasks[i]
          if (task.dataset.taskStatus == col.dataset.status && Number(task.dataset.taskIdx) == i) {
            continue
          }
          this.updates.push({
            'task_identifier': task.dataset.taskIdentifier,
            'old_status': task.dataset.taskStatus,
            'old_index': Number(task.dataset.taskIdx),
            'new_status': col.dataset.status,
            'new_index': i,
          })
        }
      }

      console.log(this.updates)
    }
  }));
});
