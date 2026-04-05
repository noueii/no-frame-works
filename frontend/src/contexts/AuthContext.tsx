import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'

interface AuthContextValue {
  sessionToken: string
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(
    () => localStorage.getItem('session_token')
  )
  const navigate = useNavigate()

  useEffect(() => {
    if (!token) {
      navigate('/login')
    }
  }, [token, navigate])

  const logout = async () => {
    await fetch('/api/v1/auth/logout', {
      method: 'POST',
      credentials: 'include',
    })
    localStorage.removeItem('session_token')
    setToken(null)
  }

  if (!token) return null

  return (
    <AuthContext.Provider value={{ sessionToken: token, logout }}>
      {children}
    </AuthContext.Provider>
  )
}
