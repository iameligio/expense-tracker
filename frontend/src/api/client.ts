import type { AuthResponse } from '../types'

// Base URL for the API. Empty by default: the SPA calls same-origin /api paths,
// proxied to the Go backend by Vite (dev) or the reverse proxy (prod). Set
// VITE_API_BASE_URL only if the API lives on a different host (e.g. a subdomain).
const API_BASE = import.meta.env.VITE_API_BASE_URL ?? ''

function url(path: string) {
  return API_BASE + path
}

// The access token lives only in memory (never localStorage) to limit XSS blast
// radius. The refresh token is an HttpOnly cookie the browser sends automatically.
let accessToken: string | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function getAccessToken() {
  return accessToken
}

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

// Single-flight refresh so concurrent 401s don't fire multiple refreshes.
let refreshing: Promise<boolean> | null = null

async function tryRefresh(): Promise<boolean> {
  if (!refreshing) {
    refreshing = (async () => {
      try {
        const res = await fetch(url('/api/auth/refresh'), {
          method: 'POST',
          credentials: 'include',
        })
        if (!res.ok) return false
        const data: AuthResponse = await res.json()
        setAccessToken(data.accessToken)
        return true
      } catch {
        return false
      } finally {
        // Cleared on next tick so awaiters share this result.
        setTimeout(() => {
          refreshing = null
        }, 0)
      }
    })()
  }
  return refreshing
}

interface RequestOptions {
  method?: string
  body?: unknown
  retryOnAuthFail?: boolean
}

// request is the core fetch wrapper: attaches the bearer token, parses JSON,
// throws ApiError on failure, and transparently refreshes once on a 401.
export async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, retryOnAuthFail = true } = opts

  const headers: Record<string, string> = {}
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`

  const res = await fetch(url(path), {
    method,
    headers,
    credentials: 'include',
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401 && retryOnAuthFail && !path.startsWith('/api/auth/')) {
    const ok = await tryRefresh()
    if (ok) return request<T>(path, { ...opts, retryOnAuthFail: false })
  }

  const isJSON = res.headers.get('content-type')?.includes('application/json')
  const data = isJSON ? await res.json().catch(() => null) : null

  if (!res.ok) {
    const msg = (data && (data.error as string)) || `Request failed (${res.status})`
    throw new ApiError(res.status, msg)
  }
  return data as T
}

export { tryRefresh }
