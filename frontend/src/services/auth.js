import { clearTokens, getRefreshToken, requestJSON, setTokens } from './http.js'
import { request } from './http.js'

export async function login(credentials) {
  const data = await requestJSON('/api/v1/auth/login', {
    method: 'POST',
    body: credentials,
    auth: false
  })

  if (data?.accessToken && data?.refreshToken) {
    setTokens({
      accessToken: data.accessToken,
      refreshToken: data.refreshToken
    })
  }

  return data
}

export async function logout() {
  const token = getRefreshToken()
  try {
    await request('/api/v1/auth/logout', {
      method: 'POST',
      body: { refreshToken: token },
      auth: false
    })
  } catch (error) {
    // ignore network errors during logout
  } finally {
    clearTokens()
  }
}
