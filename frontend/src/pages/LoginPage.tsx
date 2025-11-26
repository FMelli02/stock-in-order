import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import api from '../services/api'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { login } = useAuth()

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError(null)
    setLoading(true)
    
    try {
      const res = await api.post('/users/login', { email, password })
      const token = res.data?.token as string | undefined
      const user = res.data?.user
      
      if (token && user) {
        login(token, user)
        navigate('/')
      } else {
        setError('Credenciales incorrectas')
      }
    } catch {
      setError('Credenciales incorrectas. Por favor, verifica tus datos.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-50 py-12 px-4 sm:px-6 lg:px-8">
      
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-2xl shadow-soft-lg mb-4">
            <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          </div>
          <h1 className="text-3xl font-display font-semibold text-neutral-900 mb-2">Stock in Order</h1>
          <p className="text-neutral-600 text-sm font-light">Gestión integral de inventario</p>
        </div>

        {/* Tarjeta de Login */}
        <div className="bg-white p-8 rounded-2xl shadow-soft-lg border border-neutral-200">
          <div className="mb-8">
            <h2 className="text-2xl font-display font-semibold text-neutral-900 mb-2">Bienvenido</h2>
            <p className="text-neutral-600 text-sm">Ingresa tus credenciales para continuar</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="bg-red-50 border border-red-200 p-4 rounded-xl">
                <p className="text-sm text-red-700 font-medium">{error}</p>
              </div>
            )}
            
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-neutral-700 mb-2">
                Correo Electrónico
              </label>
              <input
                id="email"
                type="email"
                placeholder="tu@email.com"
                className="w-full px-4 py-3 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all text-sm placeholder:text-neutral-400"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-neutral-700 mb-2">
                Contraseña
              </label>
              <input
                id="password"
                type="password"
                placeholder="••••••••"
                className="w-full px-4 py-3 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all text-sm placeholder:text-neutral-400"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>

            <div className="flex items-center justify-end">
              <Link 
                to="/forgot-password" 
                className="text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors"
              >
                ¿Olvidaste tu contraseña?
              </Link>
            </div>

            <button 
              type="submit" 
              disabled={loading}
              className="w-full py-3 px-4 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-xl transition-all shadow-soft hover:shadow-soft-lg disabled:opacity-50 disabled:cursor-not-allowed text-sm"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Ingresando...
                </span>
              ) : (
                'Ingresar'
              )}
            </button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-neutral-600">
              ¿No tienes cuenta?{' '}
              <Link to="/register" className="font-medium text-primary-600 hover:text-primary-700 transition-colors">
                Crear cuenta
              </Link>
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-sm text-neutral-500">
            ¿Necesitas ayuda? Contacta a{' '}
            <a href="mailto:soporte@stockinorder.com" className="font-medium text-primary-600 hover:text-primary-700">
              soporte@stockinorder.com
            </a>
          </p>
        </div>
      </div>
    </div>
  )
}

