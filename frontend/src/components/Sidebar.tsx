import { Link, NavLink } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import InstallPWA from './InstallPWA'
import { 
  Home, Package, Building2, Users, ShoppingCart, 
  ShoppingBag, Link as LinkIcon, Camera, CreditCard,
  Gem, UserCog, FileText, Settings
} from 'lucide-react'

export default function Sidebar() {
  const { user } = useAuth()
  const base = 'flex items-center gap-3 px-4 py-2.5 rounded-xl transition-all duration-200 text-sm font-medium group'
  const active = 'bg-primary-600 text-white shadow-soft'
  const inactive = 'text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900'
  
  return (
    <aside className="w-64 min-h-screen bg-white border-r border-neutral-200 flex flex-col">
      {/* Logo y Título */}
      <Link to="/" className="px-6 py-6 border-b border-neutral-200">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-primary-600 flex items-center justify-center shadow-soft">
            <Package className="w-5 h-5 text-white" strokeWidth={2.5} />
          </div>
          <div>
            <h2 className="text-lg font-display font-semibold text-neutral-900">
              Stock in Order
            </h2>
            <p className="text-xs text-neutral-500 font-light">Gestión integral</p>
          </div>
        </div>
      </Link>
      
      {/* Perfil de Usuario */}
      {user && (
        <div className="px-4 py-4 border-b border-neutral-200">
          <div className="flex items-center gap-3 p-3 rounded-xl bg-neutral-50">
            <div className="w-10 h-10 rounded-full bg-primary-100 flex items-center justify-center flex-shrink-0">
              <span className="text-sm font-semibold text-primary-700">
                {user.name.charAt(0).toUpperCase()}
              </span>
            </div>
            <div className="flex-1 min-w-0">
              <div className="font-medium text-sm text-neutral-900 truncate">{user.name}</div>
              <div className="text-xs text-neutral-500 truncate">{user.email}</div>
            </div>
          </div>
          <div className="mt-2">
            <span className="inline-flex items-center px-2.5 py-1 bg-primary-50 text-primary-700 rounded-lg text-xs font-medium">
              {user.role === 'admin' ? 'Administrador' : user.role === 'repositor' ? 'Repositor' : 'Vendedor'}
            </span>
          </div>
        </div>
      )}
      
      <nav className="flex flex-col gap-1 flex-1 p-4 overflow-y-auto">
        <div className="text-xs font-semibold text-neutral-400 px-4 mb-2 uppercase tracking-wider">Principal</div>
        
        <NavLink to="/" end className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <Home className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Dashboard</span>
        </NavLink>
        
        <NavLink to="/products" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <Package className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Productos</span>
        </NavLink>
        
        <NavLink to="/suppliers" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <Building2 className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Proveedores</span>
        </NavLink>
        
        <NavLink to="/customers" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <Users className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Clientes</span>
        </NavLink>
        
        <div className="border-t border-neutral-200 my-3"></div>
        <div className="text-xs font-semibold text-neutral-400 px-4 mb-2 uppercase tracking-wider">Órdenes</div>
        
        <NavLink to="/sales-orders" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <ShoppingCart className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Ventas</span>
        </NavLink>
        
        <NavLink to="/purchase-orders" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <ShoppingBag className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Compras</span>
        </NavLink>
        
        <div className="border-t border-neutral-200 my-3"></div>
        <div className="text-xs font-semibold text-neutral-400 px-4 mb-2 uppercase tracking-wider">Herramientas</div>
        
        <NavLink to="/integrations" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <LinkIcon className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Integraciones</span>
        </NavLink>
        
        <NavLink to="/scanner" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <Camera className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Escanear</span>
        </NavLink>
        
        <div className="border-t border-neutral-200 my-3"></div>
        <div className="text-xs font-semibold text-neutral-400 px-4 mb-2 uppercase tracking-wider">Cuenta</div>
        
        <NavLink to="/billing" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <CreditCard className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Mi Suscripción</span>
        </NavLink>
        
        <NavLink to="/pricing" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
          <Gem className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
          <span>Ver Planes</span>
        </NavLink>
        
        {/* Admin-only section */}
        {user?.role === 'admin' && (
          <>
            <div className="border-t border-neutral-200 my-3"></div>
            <div className="text-xs font-semibold text-neutral-400 px-4 mb-2 uppercase tracking-wider">Administración</div>
            
            <NavLink to="/admin/users" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
              <UserCog className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
              <span>Usuarios</span>
            </NavLink>
            
            <NavLink to="/admin/audit-logs" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
              <FileText className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
              <span>Auditoría</span>
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
        <div className="border-t border-neutral-200 pt-3">
          <NavLink to="/profile" className={({ isActive }) => `${base} ${isActive ? active : inactive}`}>
            <Settings className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
            <span>Mi Perfil</span>
          </NavLink>
        </div>
      </nav>
    </aside>
  )
}

