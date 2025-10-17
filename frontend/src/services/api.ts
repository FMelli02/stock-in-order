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
      localStorage.removeItem('authToken')
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export default api
