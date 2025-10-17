import axios from 'axios'

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

// Auto-logout on 401 responses
api.interceptors.response.use(
  (response) => response,
  (error) => {
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
