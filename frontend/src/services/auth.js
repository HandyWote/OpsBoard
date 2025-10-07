const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

export async function login(payload) {
  const response = await fetch(`${API_BASE_URL}/api/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })

  let data = { success: false }
  try {
    data = await response.json()
  } catch (error) {
    // ignore json parse error and fall back to default message handling
  }

  if (!response.ok) {
    const message = data.message || '登录失败，请稍后重试'
    throw new Error(message)
  }

  return data
}
