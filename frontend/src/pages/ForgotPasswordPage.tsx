import { useState } from 'react'
import { Link } from 'react-router-dom'
import { isAxiosError } from 'axios'
import toast from 'react-hot-toast'
import api from '../services/api'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitted, setSubmitted] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!email) {
      toast.error('Por favor ingresa tu email')
      return
    }

    try {
      setLoading(true)
      await api.post('/users/forgot-password', { email })
      
      setSubmitted(true)
      toast.success('Instrucciones enviadas! Revisa tu email')
    } catch (err) {
      let message = 'Error al solicitar recuperación de contraseña'
      if (isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined
        message = data?.error ?? err.message
      } else if (err instanceof Error) {
        message = err.message
      }
      toast.error(message)
    } finally {
      setLoading(false)
    }
  }

  if (submitted) {
    return (
      <div className="min-h-screen bg-neutral-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-2xl shadow-soft-lg border border-neutral-200 p-8">
          <div className="text-center">
            <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-2xl bg-emerald-100 mb-4">
              <svg className="h-8 w-8 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            
            <h2 className="text-2xl font-display font-bold text-neutral-900 mb-2">
              Revisa tu email
            </h2>
            
            <p className="text-neutral-600 mb-6">
              Si existe una cuenta con el email <strong>{email}</strong>, recibirás instrucciones para restablecer tu contraseña.
            </p>
            
            <div className="bg-amber-50 border border-amber-200 rounded-xl p-4 mb-6">
              <p className="text-sm text-amber-800">
                <strong>Importante:</strong> El enlace expirará en 1 hora. Si no ves el email, revisa tu carpeta de spam.
              </p>
            </div>
            
            <Link 
              to="/login" 
              className="inline-block w-full py-3 px-4 bg-primary-600 text-white rounded-xl hover:bg-primary-700 transition-colors font-medium shadow-soft"
            >
              Volver al inicio de sesión
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-neutral-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-2xl shadow-soft-lg border border-neutral-200 p-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-display font-bold text-neutral-900 mb-2">
            ¿Olvidaste tu contraseña?
          </h1>
          <p className="text-neutral-600">
            No te preocupes, te enviaremos instrucciones para restablecerla.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-neutral-700 mb-2">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoFocus
              placeholder="tu@email.com"
              className="w-full px-4 py-2.5 rounded-xl border border-neutral-300 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 px-4 bg-primary-600 text-white rounded-xl hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium shadow-soft"
          >
            {loading ? 'Enviando...' : 'Enviar instrucciones'}
          </button>
        </form>

        <div className="mt-6 text-center space-y-2">
          <Link 
            to="/login" 
            className="text-sm text-primary-600 hover:text-primary-700 block"
          >
            ← Volver al inicio de sesión
          </Link>
          
          <p className="text-sm text-neutral-600">
            ¿No tienes cuenta?{' '}
            <Link to="/register" className="text-primary-600 hover:text-primary-700 font-medium">
              Regístrate aquí
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
