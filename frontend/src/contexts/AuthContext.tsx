import { createContext, useContext, type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import {
  useGetAuthMeQuery,
  usePostAuthLogoutMutation,
  type AuthUser,
} from '../services/api/api'

interface AuthContextValue {
  user: AuthUser
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

// Wraps protected routes — redirects to /login if not authenticated.
export function AuthProvider({ children }: { children: ReactNode }) {
  const { data: user, isLoading, isError } = useGetAuthMeQuery()
  const [logoutMutation] = usePostAuthLogoutMutation()

  if (isLoading) return null
  if (isError || !user) return <Navigate to="/login" replace />

  const logout = async () => {
    await fetch('/api/v1/auth/logout', { method: 'POST', credentials: 'include', redirect: 'follow' })
    window.location.href = '/login'
  }

  return (
    <AuthContext.Provider value={{ user, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

// Wraps public routes (login, register) — redirects to / if already authenticated.
export function GuestGuard({ children }: { children: ReactNode }) {
  const { data: user, isLoading } = useGetAuthMeQuery()

  if (isLoading) return null
  if (user) return <Navigate to="/" replace />

  return <>{children}</>
}
