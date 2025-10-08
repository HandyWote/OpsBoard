import { requestJSON } from './http.js'

export async function fetchTasks(params = {}) {
  const searchParams = new URLSearchParams()
  if (params.keyword) searchParams.set('keyword', params.keyword)
  if (params.sort) searchParams.set('sort', params.sort)
  if (params.page) searchParams.set('page', String(params.page))
  if (params.pageSize) searchParams.set('pageSize', String(params.pageSize))
  if (params.status) searchParams.set('status', params.status)

  const query = searchParams.toString()
  return requestJSON(`/api/v1/tasks${query ? `?${query}` : ''}`)
}

export async function createTask(payload) {
  return requestJSON('/api/v1/tasks', {
    method: 'POST',
    body: payload
  })
}

export async function claimTask(taskId) {
  return requestJSON(`/api/v1/tasks/${taskId}/claim`, {
    method: 'POST'
  })
}

export async function releaseTask(taskId) {
  return requestJSON(`/api/v1/tasks/${taskId}/release`, {
    method: 'POST'
  })
}

export async function publishTask(taskId) {
  return requestJSON(`/api/v1/tasks/${taskId}/publish`, {
    method: 'POST'
  })
}
