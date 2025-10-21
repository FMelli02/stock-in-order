import { useEffect, useState } from 'react'
import { isAxiosError } from 'axios'
import toast from 'react-hot-toast'
import api from '../services/api'
import Modal from '../components/Modal'
import SupplierForm from '../components/SupplierForm'
import type { Supplier } from '../types/supplier'

export default function SuppliersPage() {
  const [suppliers, setSuppliers] = useState<Supplier[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [selectedSupplier, setSelectedSupplier] = useState<Supplier | null>(null)

  const fetchSuppliers = async () => {
    try {
      setLoading(true)
      setError(null)
      const res = await api.get<Supplier[]>('/suppliers')
      setSuppliers(res.data)
    } catch (err: unknown) {
      let message = 'Error al cargar proveedores'
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
    fetchSuppliers()
  }, [])

  const openCreateModal = () => {
    setSelectedSupplier(null)
    setIsModalOpen(true)
  }

  const openEditModal = (supplier: Supplier) => {
    setSelectedSupplier(supplier)
    setIsModalOpen(true)
  }

  const closeModalAndRefresh = async () => {
    setIsModalOpen(false)
    setSelectedSupplier(null)
    await fetchSuppliers()
  }

  const handleDelete = async (supplier: Supplier) => {
    const confirmed = await new Promise<boolean>((resolve) => {
      toast((t) => (
        <div>
          <p className="font-medium mb-2">¿Estás seguro de eliminar este proveedor?</p>
          <div className="flex gap-2">
            <button
              onClick={() => {
                toast.dismiss(t.id)
                resolve(true)
              }}
              className="px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700"
            >
              Eliminar
            </button>
            <button
              onClick={() => {
                toast.dismiss(t.id)
                resolve(false)
              }}
              className="px-3 py-1 bg-gray-300 text-gray-800 rounded hover:bg-gray-400"
            >
              Cancelar
            </button>
          </div>
        </div>
      ), { duration: Infinity })
    })
    
    if (!confirmed) return
    
    try {
      await api.delete(`/suppliers/${supplier.id}`)
      toast.success('Proveedor eliminado correctamente')
      await fetchSuppliers()
    } catch (err) {
      console.error(err)
      toast.error('No se pudo eliminar el proveedor')
    }
  }

  const handleRequestReportByEmail = async () => {
    try {
      const response = await api.post<{ message: string }>('/reports/suppliers/email')
      toast.success(response.data.message || '¡Listo! Te estamos mandando el reporte por mail.')
    } catch (err) {
      console.error(err)
      toast.error('No se pudo solicitar el reporte')
    }
  }

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Proveedores</h1>
        <div className="flex gap-2">
          <button
            onClick={handleRequestReportByEmail}
            className="px-4 py-2 rounded bg-purple-600 text-white hover:bg-purple-700 flex items-center gap-2"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z" />
              <path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z" />
            </svg>
            Recibir por Email
          </button>
          <button onClick={openCreateModal} className="px-4 py-2 rounded bg-green-600 text-white hover:bg-green-700">
            Crear Nuevo Proveedor
          </button>
        </div>
      </div>

      {loading && <p className="text-gray-600">Cargando proveedores...</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nombre</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Persona de Contacto</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Teléfono</th>
                <th className="px-6 py-3" />
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {suppliers.map((s) => (
                <tr key={s.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{s.name}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{s.contact_person}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{s.email}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{s.phone}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right">
                    <button
                      onClick={() => openEditModal(s)}
                      className="mr-2 px-3 py-1 rounded bg-blue-600 text-white hover:bg-blue-700"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => handleDelete(s)}
                      className="px-3 py-1 rounded bg-red-600 text-white hover:bg-red-700"
                    >
                      Eliminar
                    </button>
                  </td>
                </tr>
              ))}
              {suppliers.length === 0 && (
                <tr>
                  <td className="px-6 py-4 text-sm text-gray-500" colSpan={5}>
                    No hay proveedores para mostrar.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <SupplierForm supplierToEdit={selectedSupplier} onSuccess={closeModalAndRefresh} />
      </Modal>
    </div>
  )
}
