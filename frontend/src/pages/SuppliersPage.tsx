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

  const handleExportExcel = async () => {
    try {
      const response = await api.get('/reports/suppliers/xlsx', {
        responseType: 'blob',
      })

      const blob = new Blob([response.data], { 
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
      })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', 'proveedores.xlsx')
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)

      toast.success('Proveedores exportados a Excel correctamente')
    } catch (err) {
      console.error(err)
      toast.error('Error al exportar proveedores')
    }
  }

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Proveedores</h1>
        <div className="flex gap-2">
          <button
            onClick={handleExportExcel}
            className="px-4 py-2 rounded bg-blue-600 text-white hover:bg-blue-700 flex items-center gap-2"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
            Exportar Excel
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
