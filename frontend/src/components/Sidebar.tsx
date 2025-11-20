import { Link, NavLink } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import InstallPWA from './InstallPWA'

export default function Sidebar() {
  const { user } = useAuth()
  const base = 'block px-4 py-3 rounded-xl hover:bg-gray-700/50 transition-all duration-200 font-medium'
  const active = 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white shadow-lg'
  
  return (
    <aside className="w-64 min-h-screen p-4 bg-gradient-to-b from-gray-900 via-gray-800 to-gray-900 text-white flex flex-col shadow-2xl">
      {/* Logo y Título */}
      <Link to="/" className="mb-6 group">
        <div className="bg-gradient-to-r from-indigo-600 to-purple-600 p-4 rounded-2xl shadow-lg transform transition-all group-hover:scale-105">
          <h2 className="text-xl font-bold text-center">
            📦 Stock In Order
          </h2>
        </div>
      </Link>
      
      {/* Perfil de Usuario */}
      {user && (
        <div className="mb-6 p-4 bg-gray-800/50 backdrop-blur rounded-2xl border border-gray-700 shadow-lg">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-xl shadow-lg">
              👤
            </div>
            <div className="flex-1 min-w-0">
              <div className="font-semibold text-sm truncate">{user.name}</div>
              <div className="text-xs text-gray-400 truncate">{user.email}</div>
            </div>
          </div>
          <div className="mt-2 pt-2 border-t border-gray-700">
            <span className="inline-block px-2 py-1 bg-indigo-600/20 text-indigo-300 rounded-lg text-xs font-medium">
              {user.role === 'admin' ? '👑 Admin' : user.role === 'repositor' ? '📦 Repositor' : '🤝 Vendedor'}
            </span>
          </div>
        </div>
      )}
      
      <nav className="flex flex-col gap-2 flex-1">
        <div className="text-xs font-semibold text-gray-400 px-2 mb-1 uppercase tracking-wider">Principal</div>
        
        <NavLink to="/" end className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">🏠</span>
            <span>Dashboard</span>
          </div>
        </NavLink>
        
        <NavLink to="/products" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">📦</span>
            <span>Productos</span>
          </div>
        </NavLink>
        
        <NavLink to="/suppliers" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">🏭</span>
            <span>Proveedores</span>
          </div>
        </NavLink>
        
        <NavLink to="/customers" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">👥</span>
            <span>Clientes</span>
          </div>
        </NavLink>
        
        <div className="border-t border-gray-700 my-3"></div>
        <div className="text-xs font-semibold text-gray-400 px-2 mb-1 uppercase tracking-wider">Órdenes</div>
        
        <NavLink to="/sales-orders" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">🛒</span>
            <span>Ventas</span>
          </div>
        </NavLink>
        
        <NavLink to="/purchase-orders" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">📥</span>
            <span>Compras</span>
          </div>
        </NavLink>
        
        <div className="border-t border-gray-700 my-3"></div>
        <div className="text-xs font-semibold text-gray-400 px-2 mb-1 uppercase tracking-wider">Herramientas</div>
        
        <NavLink to="/integrations" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">🔗</span>
            <span>Integraciones</span>
          </div>
        </NavLink>
        
        <NavLink to="/scanner" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">📷</span>
            <span>Escanear</span>
          </div>
        </NavLink>
        
        <div className="border-t border-gray-700 my-3"></div>
        <div className="text-xs font-semibold text-gray-400 px-2 mb-1 uppercase tracking-wider">Cuenta</div>
        
        <NavLink to="/billing" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">💳</span>
            <span>Mi Suscripción</span>
          </div>
        </NavLink>
        
        <NavLink to="/pricing" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
          <div className="flex items-center gap-3">
            <span className="text-lg">💎</span>
            <span>Ver Planes</span>
          </div>
        </NavLink>
        
        {/* Admin-only section */}
        {user?.role === 'admin' && (
          <>
            <div className="border-t border-gray-700 my-3"></div>
            <div className="text-xs font-semibold text-gray-400 px-2 mb-1 uppercase tracking-wider">Administración</div>
            
            <NavLink to="/admin/users" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
              <div className="flex items-center gap-3">
                <span className="text-lg">👥</span>
                <span>Usuarios</span>
              </div>
            </NavLink>
            
            <NavLink to="/admin/audit-logs" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
              <div className="flex items-center gap-3">
                <span className="text-lg">📋</span>
                <span>Auditoría</span>
              </div>
            </NavLink>
          </>
        )}
        
        {/* Spacer */}
        <div className="flex-1"></div>
        
        {/* Install PWA Button */}
        <div className="mb-3">
          <InstallPWA />
        </div>
        
        {/* Profile at bottom */}
        <div className="border-t border-gray-700 pt-3">
          <NavLink to="/profile" className={({ isActive }) => `${base} ${isActive ? active : 'hover:bg-gray-700'}`}>
            <div className="flex items-center gap-3">
              <span className="text-lg">⚙️</span>
              <span>Mi Perfil</span>
            </div>
          </NavLink>
        </div>
      </nav>
    </aside>
  )
}

