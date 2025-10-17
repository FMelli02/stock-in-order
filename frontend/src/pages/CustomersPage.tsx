import { useEffect, useState } from 'react'
import { isAxiosError } from 'axios'
import api from '../services/api'
import Modal from '../components/Modal'
import CustomerForm from '../components/CustomerForm'
import type { Customer } from '../types/customer'

export default function CustomersPage() {
  const [customers, setCustomers] = useState<Customer[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)

  const fetchCustomers = async () => {
    try {
      setLoading(true)
      setError(null)
      const res = await api.get<Customer[]>('/customers')
      setCustomers(res.data)
    } catch (err: unknown) {
      let message = 'Error al cargar clientes'
      if (isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined
        message = data?.error ?? err.message
      } else if (err instanceof Error) {
        message = err.message
      }
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchCustomers()
  }, [])

  const openCreateModal = () => {
    setSelectedCustomer(null)
    setIsModalOpen(true)
  }

  const openEditModal = (customer: Customer) => {
    setSelectedCustomer(customer)
    setIsModalOpen(true)
  }

  const closeModalAndRefresh = async () => {
    setIsModalOpen(false)
    setSelectedCustomer(null)
    await fetchCustomers()
  }

  const handleDelete = async (customer: Customer) => {
    if (!window.confirm('¿Estás seguro?')) return
    try {
      await api.delete(`/customers/${customer.id}`)
      await fetchCustomers()
    } catch (err) {
      console.error(err)
      alert('No se pudo eliminar el cliente')
    }
  }

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Clientes</h1>
        <button onClick={openCreateModal} className="px-4 py-2 rounded bg-green-600 text-white hover:bg-green-700">
          Crear Nuevo Cliente
        </button>
      </div>

      {loading && <p className="text-gray-600">Cargando clientes...</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nombre</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Teléfono</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Dirección</th>
                <th className="px-6 py-3" />
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {customers.map((c) => (
                <tr key={c.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{c.name}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{c.email}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{c.phone}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{c.address}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right">
                    <button
                      onClick={() => openEditModal(c)}
                      className="mr-2 px-3 py-1 rounded bg-blue-600 text-white hover:bg-blue-700"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => handleDelete(c)}
                      className="px-3 py-1 rounded bg-red-600 text-white hover:bg-red-700"
                    >
                      Eliminar
                    </button>
                  </td>
                </tr>
              ))}
              {customers.length === 0 && (
                <tr>
                  <td className="px-6 py-4 text-sm text-gray-500" colSpan={5}>
                    No hay clientes para mostrar.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <CustomerForm customerToEdit={selectedCustomer} onSuccess={closeModalAndRefresh} />
      </Modal>
    </div>
  )
}
