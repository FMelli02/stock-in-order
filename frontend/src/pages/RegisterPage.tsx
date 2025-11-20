import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import api from '../services/api'

export default function RegisterPage() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [nameError, setNameError] = useState<string | null>(null)
  const [emailError, setEmailError] = useState<string | null>(null)
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [confirmError, setConfirmError] = useState<string | null>(null)
  const navigate = useNavigate()
  const { login } = useAuth()

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  const minPasswordLen = 8

  const validate = () => {
    let valid = true
    const n = name.trim()
    const e = email.trim()
    const p = password
    const c = confirmPassword

    setNameError(null)
    setEmailError(null)
    setPasswordError(null)
    setConfirmError(null)
    setError(null)

    if (!n) {
      setNameError('El nombre es obligatorio')
      valid = false
    }
    if (!e) {
      setEmailError('El email es obligatorio')
      valid = false
    } else if (!emailRegex.test(e)) {
      setEmailError('Ingresá un email válido')
      valid = false
    }
    if (!p) {
      setPasswordError('La contraseña es obligatoria')
      valid = false
    } else if (p.length < minPasswordLen) {
      setPasswordError(`La contraseña debe tener al menos ${minPasswordLen} caracteres`)
      valid = false
    }
    if (!c) {
      setConfirmError('Confirmá la contraseña')
      valid = false
    } else if (p !== c) {
      setConfirmError('Las contraseñas no coinciden')
      valid = false
    }

    return valid
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!validate()) return
    try {
      setSubmitting(true)
      setError(null)
      await api.post('/users/register', { name: name.trim(), email: email.trim(), password })
      // luego de registrar, iniciar sesión automáticamente
      const res = await api.post('/users/login', { email: email.trim(), password })
      const token = res.data?.token as string | undefined
      const user = res.data?.user
      
      if (token && user) {
        login(token, user)
        navigate('/')
      } else {
        navigate('/login')
      }
    } catch (err: unknown) {
      console.error(err)
      let message = 'No se pudo registrar el usuario'
      if (typeof err === 'object' && err !== null) {
        type Axiosish = { response?: { data?: { error?: string; details?: string } | unknown; status?: number }; request?: unknown }
        const maybeAxios = err as Axiosish
        const data = maybeAxios.response?.data as { error?: string; details?: string } | undefined
        if (maybeAxios.response?.status === 409) {
          message = 'El email ya está registrado'
        } else if (maybeAxios.response?.status === 400 && data?.details) {
          message = data.details
        } else if (typeof data === 'object' && data && 'error' in data && typeof data.error === 'string') {
          message = data.error
        } else if (!maybeAxios.response && maybeAxios.request) {
          message = 'No se pudo contactar al servidor'
        }
      }
      setError(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-100 via-purple-50 to-pink-100 py-12 px-4 sm:px-6 lg:px-8">
      {/* Círculos decorativos de fondo */}
      <div className="absolute top-0 left-0 w-72 h-72 bg-indigo-300 rounded-full mix-blend-multiply filter blur-xl opacity-30 animate-pulse"></div>
      <div className="absolute top-0 right-0 w-72 h-72 bg-purple-300 rounded-full mix-blend-multiply filter blur-xl opacity-30 animate-pulse animation-delay-2000"></div>
      <div className="absolute bottom-0 left-1/2 w-72 h-72 bg-pink-300 rounded-full mix-blend-multiply filter blur-xl opacity-30 animate-pulse animation-delay-4000"></div>
      
      <div className="relative w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-3xl shadow-2xl mb-4 transform hover:scale-110 transition-transform">
            <span className="text-4xl">📦</span>
          </div>
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Stock In Order</h1>
          <p className="text-gray-600">Crea tu cuenta y comienza a gestionar tu inventario</p>
        </div>

        {/* Tarjeta de Registro */}
        <div className="bg-white/80 backdrop-blur-xl p-8 rounded-3xl shadow-2xl border border-white/20">
          <div className="mb-6">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">Crear Cuenta</h2>
            <p className="text-gray-600">Completa tus datos para comenzar</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="bg-red-50 border-l-4 border-red-500 p-4 rounded-r-xl">
                <div className="flex items-center gap-2">
                  <span className="text-red-500 text-xl">⚠️</span>
                  <p className="text-sm text-red-700 font-medium">{error}</p>
                </div>
              </div>
            )}
            
            <div>
              <label htmlFor="name" className="block text-sm font-semibold text-gray-700 mb-2">
                Nombre Completo
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </div>
                <input
                  id="name"
                  placeholder="Tu nombre"
                  className="pl-10 block w-full rounded-xl border-2 border-gray-200 py-3 px-4 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100 transition-all"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                />
              </div>
              {nameError && <p className="text-xs text-red-600 mt-1 ml-1">{nameError}</p>}
            </div>
            
            <div>
              <label htmlFor="email" className="block text-sm font-semibold text-gray-700 mb-2">
                Correo Electrónico
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                </div>
                <input
                  id="email"
                  type="email"
                  placeholder="tu@email.com"
                  className="pl-10 block w-full rounded-xl border-2 border-gray-200 py-3 px-4 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100 transition-all"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
              {emailError && <p className="text-xs text-red-600 mt-1 ml-1">{emailError}</p>}
            </div>
            
            <div>
              <label htmlFor="password" className="block text-sm font-semibold text-gray-700 mb-2">
                Contraseña
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                </div>
                <input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  className="pl-10 block w-full rounded-xl border-2 border-gray-200 py-3 px-4 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100 transition-all"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              {passwordError && <p className="text-xs text-red-600 mt-1 ml-1">{passwordError}</p>}
            </div>
            
            <div>
              <label htmlFor="confirm" className="block text-sm font-semibold text-gray-700 mb-2">
                Confirmar Contraseña
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <input
                  id="confirm"
                  type="password"
                  placeholder="••••••••"
                  className="pl-10 block w-full rounded-xl border-2 border-gray-200 py-3 px-4 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100 transition-all"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                />
              </div>
              {confirmError && <p className="text-xs text-red-600 mt-1 ml-1">{confirmError}</p>}
            </div>

            <button 
              type="submit" 
              disabled={submitting}
              className="w-full py-3 px-4 bg-gradient-to-r from-indigo-600 to-purple-600 text-white font-semibold rounded-xl hover:from-indigo-700 hover:to-purple-700 focus:outline-none focus:ring-4 focus:ring-indigo-200 shadow-lg transform transition-all hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
            >
              {submitting ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Creando cuenta...
                </span>
              ) : (
                'Crear Cuenta'
              )}
            </button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              ¿Ya tienes cuenta?{' '}
              <Link to="/login" className="font-semibold text-indigo-600 hover:text-indigo-700 transition-colors">
                Iniciar sesión
              </Link>
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-sm text-gray-600">
            Al registrarte, aceptas nuestros{' '}
            <a href="#" className="font-medium text-indigo-600 hover:text-indigo-700">
              términos y condiciones
            </a>
          </p>
        </div>
      </div>
    </div>
  )
}

