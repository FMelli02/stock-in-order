import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function ProtectedRoute() {
  const { isAuthenticated, token, loading } = useAuth()

  console.log('🛡️ ProtectedRoute check:', { isAuthenticated, hasToken: !!token, loading })

  // Esperar a que el AuthContext termine de cargar
  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-neutral-50">
        <div className="text-center">
          <div className="w-12 h-12 border-3 border-neutral-200 border-t-primary-600 rounded-full animate-spin mb-4"></div>
          <p className="text-sm font-medium text-neutral-600">Cargando...</p>
        </div>
      </div>
    )
  }

  // Si no está autenticado, redirigir al login
  if (!isAuthenticated || !token) {
    console.log('❌ No autenticado, redirigiendo a /login')
    return <Navigate to="/login" replace />
  }

  console.log('✅ Autenticado, mostrando contenido protegido')
  return <Outlet />
}

