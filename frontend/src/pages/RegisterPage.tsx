import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
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
      if (token) {
        localStorage.setItem('authToken', token)
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
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="w-full max-w-sm bg-white p-6 rounded-lg shadow">
        <h1 className="text-2xl font-bold mb-4 text-center">Crear Cuenta</h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && <p className="text-sm text-red-600">{error}</p>}
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700">Nombre</label>
            <input
              id="name"
              className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
            {nameError && <p className="text-xs text-red-600 mt-1">{nameError}</p>}
          </div>
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700">Email</label>
            <input
              id="email"
              type="email"
              className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
            {emailError && <p className="text-xs text-red-600 mt-1">{emailError}</p>}
          </div>
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700">Password</label>
            <input
              id="password"
              type="password"
              className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
            {passwordError && <p className="text-xs text-red-600 mt-1">{passwordError}</p>}
          </div>
          <div>
            <label htmlFor="confirm" className="block text-sm font-medium text-gray-700">Confirmar Password</label>
            <input
              id="confirm"
              type="password"
              className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
            {confirmError && <p className="text-xs text-red-600 mt-1">{confirmError}</p>}
          </div>
          <button type="submit" disabled={submitting} className="w-full py-2 px-4 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50">
            {submitting ? 'Creando...' : 'Crear Cuenta'}
          </button>
        </form>
      </div>
    </div>
  )
}
