import { useEffect, useState } from 'react'
import { isAxiosError } from 'axios'
import toast from 'react-hot-toast'
import api from '../services/api'
import Modal from '../components/Modal'
import ProductForm from '../components/ProductForm'
import type { Product } from '../types/product'

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null)

  const fetchProducts = async () => {
    try {
      setLoading(true)
      setError(null)
      const res = await api.get<Product[]>('/products')
      setProducts(res.data)
    } catch (err: unknown) {
      let message = 'Error al cargar productos'
      if (isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined
        if (data?.error) {
          message = data.error
        } else {
          message = err.message
        }
      } else if (err instanceof Error) {
        message = err.message
      }
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchProducts()
  }, [])

  const openCreateModal = () => {
    setSelectedProduct(null)
    setIsModalOpen(true)
  }

  const openEditModal = (product: Product) => {
    setSelectedProduct(product)
    setIsModalOpen(true)
  }

  const closeModalAndRefresh = async () => {
    setIsModalOpen(false)
    setSelectedProduct(null)
    await fetchProducts()
  }

  const handleDelete = async (product: Product) => {
    const confirmed = await new Promise<boolean>((resolve) => {
      toast((t) => (
        <div>
          <p className="font-medium mb-2">¿Estás seguro de eliminar este producto?</p>
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
      await api.delete(`/products/${product.id}`)
      toast.success('Producto eliminado correctamente')
      await fetchProducts()
    } catch (err: any) {
      console.error(err)
      const errorMessage = err.response?.data?.error || 'No se pudo eliminar el producto'
      toast.error(errorMessage)
    }
  }

  const handleExportExcel = async () => {
    try {
      const response = await api.get('/reports/products/xlsx', {
        responseType: 'blob',
      })

      // Crear un blob y un link de descarga para archivo Excel
      const blob = new Blob([response.data], { 
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
      })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', 'productos.xlsx')
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)

      toast.success('Productos exportados a Excel correctamente')
    } catch (err) {
      console.error(err)
      toast.error('No se pudo exportar los productos')
    }
  }

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Productos</h1>
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
          <button
            onClick={openCreateModal}
            className="px-4 py-2 rounded bg-green-600 text-white hover:bg-green-700"
          >
            Crear Nuevo Producto
          </button>
        </div>
      </div>
      {loading && <p className="text-gray-600">Cargando productos...</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nombre</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">SKU</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Cantidad en Stock</th>
                <th className="px-6 py-3" />
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {products.map((p) => (
                <tr key={p.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{p.name}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{p.sku}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{p.quantity}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right">
                    <button
                      onClick={() => openEditModal(p)}
                      className="mr-2 px-3 py-1 rounded bg-blue-600 text-white hover:bg-blue-700"
                    >
                      Editar
                    </button>
                    <button
                      onClick={() => handleDelete(p)}
                      className="px-3 py-1 rounded bg-red-600 text-white hover:bg-red-700"
                    >
                      Eliminar
                    </button>
                  </td>
                </tr>
              ))}
              {products.length === 0 && (
                <tr>
                  <td className="px-6 py-4 text-sm text-gray-500" colSpan={3}>
                    No hay productos para mostrar.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <ProductForm productToEdit={selectedProduct} onSuccess={closeModalAndRefresh} />
      </Modal>
    </div>
  )
}
