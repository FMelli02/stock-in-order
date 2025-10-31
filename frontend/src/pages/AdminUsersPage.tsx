import { useState } from 'react'
import toast from 'react-hot-toast'
import api from '../services/api'

interface CreateUserFormData {
  name: string
  email: string
  password: string
  role: 'admin' | 'vendedor' | 'repositor'
}

export default function AdminUsersPage() {
  const [formData, setFormData] = useState<CreateUserFormData>({
    name: '',
    email: '',
    password: '',
    role: 'vendedor',
  })
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      await api.post('/admin/users', formData)
      toast.success(`Usuario ${formData.name} creado exitosamente con rol ${formData.role}`)
      
      // Reset form
      setFormData({
        name: '',
        email: '',
        password: '',
        role: 'vendedor',
      })
    } catch (err) {
      let errorMsg = 'Error al crear usuario'
      if (err && typeof err === 'object' && 'response' in err) {
        const response = (err as { response?: { data?: { error?: string } } }).response
        if (response?.data?.error) {
          errorMsg = response.data.error
        }
      }
      toast.error(errorMsg)
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    })
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">👥 Gestión de Usuarios</h1>
      
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-xl font-semibold mb-4">Crear Nuevo Usuario</h2>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Nombre */}
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              Nombre Completo
            </label>
            <input
              id="name"
              name="name"
              type="text"
              required
              value={formData.name}
              onChange={handleChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Ej: Juan Pérez"
            />
          </div>

          {/* Email */}
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              Email
            </label>
            <input
              id="email"
              name="email"
              type="email"
              required
              value={formData.email}
              onChange={handleChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="juan@example.com"
            />
          </div>

          {/* Password */}
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              Contraseña
            </label>
            <input
              id="password"
              name="password"
              type="password"
              required
              minLength={8}
              value={formData.password}
              onChange={handleChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Mínimo 8 caracteres"
            />
            <p className="text-xs text-gray-500 mt-1">La contraseña debe tener al menos 8 caracteres</p>
          </div>

          {/* Rol */}
          <div>
            <label htmlFor="role" className="block text-sm font-medium text-gray-700 mb-1">
              Rol del Usuario
            </label>
            <select
              id="role"
              name="role"
              required
              value={formData.role}
              onChange={handleChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="vendedor">🛒 Vendedor (Área Comercial)</option>
              <option value="repositor">📦 Repositor (Logística e Inventario)</option>
              <option value="admin">👑 Admin (Acceso Total)</option>
            </select>
            <p className="text-xs text-gray-500 mt-1">
              {formData.role === 'admin' && '👑 Acceso total al sistema'}
              {formData.role === 'vendedor' && '🛒 Gestión de clientes y ventas'}
              {formData.role === 'repositor' && '📦 Gestión de proveedores, compras e inventario'}
            </p>
          </div>

          {/* Submit Button */}
          <div className="flex gap-3 pt-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? 'Creando...' : '✅ Crear Usuario'}
            </button>
            
            <button
              type="button"
              onClick={() => setFormData({ name: '', email: '', password: '', role: 'vendedor' })}
              className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
            >
              Limpiar
            </button>
          </div>
        </form>

        {/* Info Card */}
        <div className="mt-6 p-4 bg-blue-50 rounded-md border border-blue-200">
          <h3 className="text-sm font-semibold text-blue-900 mb-2">ℹ️ Información sobre Roles</h3>
          <ul className="text-sm text-blue-800 space-y-1">
            <li>• <strong>Admin:</strong> Puede crear usuarios, eliminar registros y acceder a todo</li>
            <li>• <strong>Vendedor:</strong> Gestiona clientes, ventas y consulta productos</li>
            <li>• <strong>Repositor:</strong> Gestiona proveedores, compras y ajustes de stock</li>
          </ul>
        </div>
      </div>
    </div>
  )
}
