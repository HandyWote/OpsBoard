import { requestJSON } from './http.js'

export async function fetchCurrentUser() {
  return requestJSON('/api/v1/users/me')
}

export async function updateProfile(payload) {
  return requestJSON('/api/v1/users/me/profile', {
    method: 'PATCH',
    body: payload
  })
}

export async function changePassword(payload) {
  return requestJSON('/api/v1/users/me/password', {
    method: 'PATCH',
    body: payload
  })
}

export async function listAccounts(params = {}) {
  const searchParams = new URLSearchParams()
  if (params.keyword) searchParams.set('keyword', params.keyword)
  if (params.page) searchParams.set('page', String(params.page))
  if (params.pageSize) searchParams.set('pageSize', String(params.pageSize))

  const query = searchParams.toString()
  return requestJSON(`/api/v1/users${query ? `?${query}` : ''}`)
}

export async function toggleAdmin(accountId, grant) {
  return requestJSON(`/api/v1/users/${accountId}/toggle-admin`, {
    method: 'POST',
    body: { grant }
  })
}
