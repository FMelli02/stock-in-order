import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function ProtectedRoute() {
  const { isAuthenticated, token, loading } = useAuth()

  console.log('🛡️ ProtectedRoute check:', { isAuthenticated, hasToken: !!token, loading })

  // Esperar a que el AuthContext termine de cargar
  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
        <div className="text-center">
          <div className="relative inline-flex">
            <div className="w-16 h-16 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="text-2xl">📦</span>
            </div>
          </div>
          <p className="mt-4 text-lg font-semibold text-slate-700">Cargando...</p>
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

