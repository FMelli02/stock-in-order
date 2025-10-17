import { Link, NavLink } from 'react-router-dom'

export default function Sidebar() {
  const base = 'block px-4 py-2 rounded hover:bg-gray-700'
  const active = 'bg-gray-700 font-semibold'
  const handleLogout = () => {
    try {
      localStorage.removeItem('authToken')
    } finally {
      // Force a full reload to clear in-memory state
      window.location.href = '/login'
    }
  }
  return (
    <aside className="w-56 min-h-screen p-4 bg-gray-800 text-white flex flex-col">
      <h2 className="text-xl font-bold mb-4">
        <Link to="/">Stock In Order</Link>
      </h2>
      <nav className="flex flex-col gap-2">
        <NavLink to="/" end className={({ isActive }) => `${base} ${isActive ? active : ''}`}>
          Dashboard
        </NavLink>
        <NavLink to="/products" className={({ isActive }) => `${base} ${isActive ? active : ''}`}>
          Productos
        </NavLink>
        <NavLink to="/suppliers" className={({ isActive }) => `${base} ${isActive ? active : ''}`}>
          Proveedores
        </NavLink>
        <NavLink to="/customers" className={({ isActive }) => `${base} ${isActive ? active : ''}`}>
          Clientes
        </NavLink>
        <NavLink to="/sales-orders" className={({ isActive }) => `${base} ${isActive ? active : ''}`}>
          Ventas
        </NavLink>
        <NavLink to="/purchase-orders" className={({ isActive }) => `${base} ${isActive ? active : ''}`}>
          Compras
        </NavLink>
        <NavLink to="/login" className={({ isActive }) => `${base} ${isActive ? active : ''}`}>
          Login
        </NavLink>
      </nav>
      <div className="mt-auto pt-4">
        <button onClick={handleLogout} className="w-full px-4 py-2 rounded bg-red-600 hover:bg-red-700">
          Cerrar Sesión
        </button>
      </div>
    </aside>
  )
}
