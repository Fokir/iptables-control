import type { ApiResponse } from '../types'

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  })

  if (res.status === 401) {
    if (!window.location.pathname.startsWith('/login')) {
      window.location.href = '/login'
    }
    throw new Error('Unauthorized')
  }

  const body: ApiResponse<T> = await res.json()
  if (!body.success) {
    throw new Error(body.error || 'Unknown error')
  }
  return body.data as T
}

export const api = {
  get: <T>(url: string) => request<T>(url),
  post: <T>(url: string, data?: unknown) =>
    request<T>(url, { method: 'POST', body: data ? JSON.stringify(data) : undefined }),
  put: <T>(url: string, data: unknown) =>
    request<T>(url, { method: 'PUT', body: JSON.stringify(data) }),
  delete: <T>(url: string) => request<T>(url, { method: 'DELETE' }),
}
