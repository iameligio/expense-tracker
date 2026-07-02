import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { authApi, meApi } from '../api/endpoints'
import { setAccessToken, tryRefresh } from '../api/client'
import type { User } from '../types'

interface AuthState {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthState | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  // On first load, attempt a silent refresh (using the HttpOnly cookie) so a
  // returning user stays logged in across page reloads.
  useEffect(() => {
    ;(async () => {
      const ok = await tryRefresh()
      if (ok) {
        try {
          setUser(await meApi.get())
        } catch {
          setUser(null)
        }
      }
      setLoading(false)
    })()
  }, [])

  async function login(email: string, password: string) {
    const res = await authApi.login(email, password)
    setAccessToken(res.accessToken)
    setUser(res.user)
  }

  async function register(email: string, password: string) {
    const res = await authApi.register(email, password)
    setAccessToken(res.accessToken)
    setUser(res.user)
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      setAccessToken(null)
      setUser(null)
    }
  }

  async function refreshUser() {
    setUser(await meApi.get())
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
