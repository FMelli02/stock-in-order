import { useState, useEffect } from 'react'
import toast from 'react-hot-toast'
import api from '../services/api'

interface CreateUserFormData {
  name: string
  email: string
  password: string
  role: 'admin' | 'vendedor' | 'repositor'
}

interface User {
  id: number
  name: string
  email: string
  role: string
  created_at: string
}

export default function AdminUsersPage() {
  const [formData, setFormData] = useState<CreateUserFormData>({
    name: '',
    email: '',
    password: '',
    role: 'vendedor',
  })
  const [loading, setLoading] = useState(false)
  const [users, setUsers] = useState<User[]>([])
  const [loadingUsers, setLoadingUsers] = useState(true)
  const [deletingUserId, setDeletingUserId] = useState<number | null>(null)

  // Fetch users on component mount
  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    try {
      setLoadingUsers(true)
      const response = await api.get('/admin/users')
      setUsers(response.data)
    } catch (err) {
      toast.error('Error al cargar usuarios')
      console.error('Error fetching users:', err)
    } finally {
      setLoadingUsers(false)
    }
  }

  const handleDelete = async (userId: number, userName: string) => {
    if (!confirm(`¿Estás seguro de que deseas eliminar al usuario "${userName}"?\n\nEsta acción no se puede deshacer.`)) {
      return
    }

    try {
      setDeletingUserId(userId)
      await api.delete(`/admin/users/${userId}`)
      toast.success(`Usuario "${userName}" eliminado exitosamente`)
      
      // Refresh users list
      fetchUsers()
    } catch (err) {
      let errorMsg = 'Error al eliminar usuario'
      if (err && typeof err === 'object' && 'response' in err) {
        const response = (err as { response?: { data?: { error?: string } } }).response
        if (response?.data?.error) {
          errorMsg = response.data.error
        }
      }
      toast.error(errorMsg)
    } finally {
      setDeletingUserId(null)
    }
  }

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
      
      // Refresh users list
      fetchUsers()
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
      <h1 className="text-3xl font-display font-bold mb-6 text-neutral-900">Gestión de Usuarios</h1>
      
      <div className="bg-white shadow-soft rounded-2xl border border-neutral-200 p-6">
        <h2 className="text-xl font-display font-semibold mb-4 text-neutral-900">Crear Nuevo Usuario</h2>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Nombre */}
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-neutral-700 mb-1">
              Nombre Completo
            </label>
            <input
              id="name"
              name="name"
              type="text"
              required
              value={formData.name}
              onChange={handleChange}
              className="w-full px-4 py-2.5 border border-neutral-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              placeholder="Ej: Juan Pérez"
            />
          </div>

          {/* Email */}
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-neutral-700 mb-1">
              Email
            </label>
            <input
              id="email"
              name="email"
              type="email"
              required
              value={formData.email}
              onChange={handleChange}
              className="w-full px-4 py-2.5 border border-neutral-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              placeholder="juan@example.com"
            />
          </div>

          {/* Password */}
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-neutral-700 mb-1">
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
              className="w-full px-4 py-2.5 border border-neutral-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              placeholder="Mínimo 8 caracteres"
            />
            <p className="text-xs text-gray-500 mt-1">La contraseña debe tener al menos 8 caracteres</p>
          </div>

          {/* Rol */}
          <div>
            <label htmlFor="role" className="block text-sm font-medium text-neutral-700 mb-1">
              Rol del Usuario
            </label>
            <select
              id="role"
              name="role"
              required
              value={formData.role}
              onChange={handleChange}
              className="w-full px-4 py-2.5 border border-neutral-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            >
              <option value="vendedor">Vendedor (Área Comercial)</option>
              <option value="repositor">Repositor (Logística e Inventario)</option>
              <option value="admin">Admin (Acceso Total)</option>
            </select>
            <p className="text-xs text-neutral-500 mt-1">
              {formData.role === 'admin' && 'Acceso total al sistema'}
              {formData.role === 'vendedor' && 'Gestión de clientes y ventas'}
              {formData.role === 'repositor' && 'Gestión de proveedores, compras e inventario'}
            </p>
          </div>

          {/* Submit Button */}
          <div className="flex gap-3 pt-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-primary-600 text-white py-2.5 px-4 rounded-xl hover:bg-primary-700 disabled:bg-neutral-400 disabled:cursor-not-allowed transition-colors shadow-soft"
            >
              {loading ? 'Creando...' : 'Crear Usuario'}
            </button>
            
            <button
              type="button"
              onClick={() => setFormData({ name: '', email: '', password: '', role: 'vendedor' })}
              className="px-4 py-2.5 border border-neutral-300 rounded-xl hover:bg-neutral-50 transition-colors"
            >
              Limpiar
            </button>
          </div>
        </form>

        {/* Info Card */}
        <div className="mt-6 p-4 bg-primary-50 rounded-xl border border-primary-200">
          <h3 className="text-sm font-semibold text-primary-900 mb-2">Información sobre Roles</h3>
          <ul className="text-sm text-primary-800 space-y-1">
            <li>• <strong>Admin:</strong> Puede crear usuarios, eliminar registros y acceder a todo</li>
            <li>• <strong>Vendedor:</strong> Gestiona clientes, ventas y consulta productos</li>
            <li>• <strong>Repositor:</strong> Gestiona proveedores, compras y ajustes de stock</li>
          </ul>
        </div>
      </div>

      {/* Users List Table */}
      <div className="bg-white shadow-soft rounded-2xl border border-neutral-200 p-6 mt-6">
        <h2 className="text-xl font-display font-semibold mb-4 text-neutral-900">Usuarios Existentes</h2>
        
        {loadingUsers ? (
          <div className="text-center py-8">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
            <p className="mt-2 text-neutral-600">Cargando usuarios...</p>
          </div>
        ) : users.length === 0 ? (
          <div className="text-center py-8 text-neutral-500">
            <p>No hay usuarios registrados en tu organización.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-neutral-200">
              <thead className="bg-neutral-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Nombre
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Rol
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Fecha de Creación
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Acciones
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-neutral-200">
                {users.map((user) => (
                  <tr key={user.id} className="hover:bg-neutral-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">{user.name}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-500">{user.email}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        user.role === 'admin' ? 'bg-purple-100 text-purple-800' :
                        user.role === 'vendedor' ? 'bg-blue-100 text-blue-800' :
                        'bg-emerald-100 text-emerald-800'
                      }`}>
                        {user.role === 'admin' && 'Admin'}
                        {user.role === 'vendedor' && 'Vendedor'}
                        {user.role === 'repositor' && 'Repositor'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(user.created_at).toLocaleDateString('es-AR', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit'
                      })}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => handleDelete(user.id, user.name)}
                        disabled={deletingUserId === user.id}
                        className="text-red-600 hover:text-red-900 disabled:text-neutral-400 disabled:cursor-not-allowed transition-colors"
                        title="Eliminar usuario"
                      >
                        {deletingUserId === user.id ? (
                          <span className="inline-block animate-spin">Eliminando...</span>
                        ) : (
                          'Eliminar'
                        )}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
