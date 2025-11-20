import axios from 'axios'
import * as Sentry from '@sentry/react'

// Vite exposes env vars on import.meta.env
const { VITE_API_URL } = import.meta.env as { VITE_API_URL?: string }
const baseURL = VITE_API_URL ?? 'http://localhost:8080/api/v1'

const api = axios.create({
  baseURL,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  },
})

// Attach Authorization header with JWT if available
api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('authToken')
    if (token) {
      config.headers = config.headers ?? {}
      ;(config.headers as Record<string, string>)['Authorization'] = `Bearer ${token}`
      console.log('API Request:', {
        url: config.url,
        method: config.method,
        hasToken: !!token,
        tokenPreview: token ? `${token.substring(0, 20)}...` : 'none'
      })
    } else {
      console.warn('API Request without token:', config.url)
    }
  }
  return config
})

// Auto-logout on 401 responses and capture errors in Sentry
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Capture API errors in Sentry
    if (error?.response) {
      // Server responded with error status
      Sentry.captureException(error, {
        tags: {
          api_endpoint: error.config?.url,
          http_method: error.config?.method,
          status_code: error.response.status,
        },
        extra: {
          response_data: error.response.data,
          request_data: error.config?.data,
        },
      })
    } else if (error?.request) {
      // Request was made but no response received (network error)
      Sentry.captureException(error, {
        tags: {
          error_type: 'network_error',
          api_endpoint: error.config?.url,
        },
      })
    }

    // Auto-logout on 401 responses
    if (error?.response?.status === 401 && typeof window !== 'undefined') {
      const currentPath = window.location.pathname
      
      // NUNCA redirigir en /profile - dejar que el componente use AuthContext
      if (currentPath.includes('/profile') || currentPath.includes('/mi-perfil')) {
        console.warn('⚠️ 401 en /profile - ignorando (usando AuthContext)')
        return Promise.reject(error)
      }
      
      // En otras páginas protegidas, hacer logout y redirigir
      if (!['/login', '/register', '/pricing'].includes(currentPath)) {
        console.error('❌ 401 Unauthorized - Cerrando sesión y redirigiendo')
        localStorage.removeItem('authToken')
        localStorage.removeItem('user')
        sessionStorage.setItem('redirectAfterLogin', currentPath)
        window.location.href = '/login'
      }
    }

    // Redirect to billing on 402 Payment Required
    if (error?.response?.status === 402 && typeof window !== 'undefined') {
      // Store the error data for display
      const paymentError = {
        message: error.response.data?.message || 'Suscripción requerida',
        upgrade_url: error.response.data?.upgrade_url || '/billing',
      }
      sessionStorage.setItem('paymentRequired', JSON.stringify(paymentError))
      
      // Only redirect if not already on billing/pricing pages
      if (!['/billing', '/pricing'].includes(window.location.pathname)) {
        window.location.href = '/billing'
      }
    }

    return Promise.reject(error)
  }
)

export default api
