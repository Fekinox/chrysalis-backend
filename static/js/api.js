const csrf = document.querySelector("meta[name='csrf-token']").getAttribute("content")

const api = axios.create({
  timeout: 1000,
  maxRedirects: 0,
})

api.defaults.headers.post['X-CSRF-Token'] = csrf
api.defaults.headers.put['X-CSRF-Token'] = csrf
api.defaults.headers.delete['X-CSRF-Token'] = csrf
api.defaults.headers.patch['X-CSRF-Token'] = csrf

api.defaults.headers.post['X-CSRF-Protection'] = "1"
api.defaults.headers.put['X-CSRF-Protection'] = "1"
api.defaults.headers.delete['X-CSRF-Protection'] = "1"
api.defaults.headers.patch['X-CSRF-Protection'] = "1"

document.body.addEventListener('htmx:configRequest', (event) => {
  event.detail.headers['X-CSRF-Token'] = csrf
  event.detail.headers['X-CSRF-Protection'] = "1"
})
