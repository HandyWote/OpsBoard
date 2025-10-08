const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, '')
const STORAGE_KEY = 'opsboard.tokens.v1'

let accessToken = ''
let refreshToken = ''

loadPersistedTokens()

function loadPersistedTokens() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    accessToken = typeof parsed.accessToken === 'string' ? parsed.accessToken : ''
    refreshToken = typeof parsed.refreshToken === 'string' ? parsed.refreshToken : ''
  } catch (error) {
    // ignore malformed storage
    accessToken = ''
    refreshToken = ''
  }
}

function persistTokens() {
  if (!accessToken && !refreshToken) {
    localStorage.removeItem(STORAGE_KEY)
    return
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify({ accessToken, refreshToken }))
}

export function setTokens(tokens = {}) {
  accessToken = tokens.accessToken || ''
  refreshToken = tokens.refreshToken || ''
  persistTokens()
}

export function clearTokens() {
  accessToken = ''
  refreshToken = ''
  persistTokens()
}

export function getAccessToken() {
  return accessToken
}

export function getRefreshToken() {
  return refreshToken
}

export function isAuthenticated() {
  return Boolean(accessToken && refreshToken)
}

let refreshingPromise = null

async function refreshTokens() {
  if (!refreshToken) return false
  if (refreshingPromise) return refreshingPromise

  const payload = {
    refreshToken
  }

  refreshingPromise = fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
    .then(async (response) => {
      let data = null
      try {
        data = await response.json()
      } catch (error) {
        /* ignore parse failure */
      }

      if (!response.ok) {
        clearTokens()
        return false
      }

      const payloadData = data && typeof data === 'object' && 'data' in data ? data.data : data
      if (!payloadData?.accessToken || !payloadData?.refreshToken) {
        clearTokens()
        return false
      }

      setTokens(payloadData)
      return true
    })
    .catch(() => {
      clearTokens()
      return false
    })
    .finally(() => {
      refreshingPromise = null
    })

  return refreshingPromise
}

export async function request(path, options = {}) {
  const {
    method = 'GET',
    headers = {},
    body,
    auth = true,
    signal
  } = options

  const finalHeaders = new Headers(headers)
  let payload = body

  if (body && typeof body === 'object' && !(body instanceof FormData)) {
    if (!finalHeaders.has('Content-Type')) {
      finalHeaders.set('Content-Type', 'application/json')
    }
    payload = JSON.stringify(body)
  }

  if (auth && accessToken && !finalHeaders.has('Authorization')) {
    finalHeaders.set('Authorization', `Bearer ${accessToken}`)
  }

  const url = `${API_BASE_URL}${path}`
  let response = await fetch(url, {
    method,
    headers: finalHeaders,
    body: payload,
    signal
  })

  if (response.status === 401 && auth && refreshToken) {
    const refreshed = await refreshTokens()
    if (refreshed && accessToken) {
      finalHeaders.set('Authorization', `Bearer ${accessToken}`)
      response = await fetch(url, {
        method,
        headers: finalHeaders,
        body: payload,
        signal
      })
    }
  }

  return response
}

export async function requestJSON(path, options = {}) {
  const response = await request(path, options)
  let payload = null

  try {
    payload = await response.json()
  } catch (error) {
    /* ignore parse failure */
  }

  if (!response.ok) {
    const message =
      payload?.error?.message ||
      payload?.message ||
      payload?.error ||
      response.statusText ||
      '服务暂时不可用'
    const err = new Error(message)
    err.status = response.status
    throw err
  }

  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data
  }

  return payload
}
