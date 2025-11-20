import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'

interface User {
  id: number
  name: string
  email: string
  role: 'admin' | 'vendedor' | 'repositor'
}

interface AuthContextType {
  user: User | null
  token: string | null
  login: (token: string, userData: User) => void
  logout: () => void
  isAuthenticated: boolean
  loading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  // Initialize from localStorage on mount
  useEffect(() => {
    console.log('🔄 Inicializando AuthContext...')
    const storedToken = localStorage.getItem('authToken')
    const storedUser = localStorage.getItem('user')
    
    console.log('📦 localStorage:', {
      hasToken: !!storedToken,
      hasUser: !!storedUser,
      tokenPreview: storedToken ? storedToken.substring(0, 20) + '...' : 'none'
    })
    
    if (storedToken && storedUser) {
      try {
        const userData = JSON.parse(storedUser) as User
        setToken(storedToken)
        setUser(userData)
        console.log('✅ Auth cargado desde localStorage:', userData)
      } catch (err) {
        console.error('❌ Error parseando usuario:', err)
        localStorage.removeItem('authToken')
        localStorage.removeItem('user')
      }
    } else {
      console.log('⚠️ No hay datos de autenticación en localStorage')
    }
    
    setLoading(false)
  }, [])

  const login = (newToken: string, userData: User) => {
    localStorage.setItem('authToken', newToken)
    localStorage.setItem('user', JSON.stringify(userData))
    setToken(newToken)
    setUser(userData)
  }

  const logout = () => {
    localStorage.removeItem('authToken')
    localStorage.removeItem('user')
    setToken(null)
    setUser(null)
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        login,
        logout,
        isAuthenticated: !!token && !!user,
        loading,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
