import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { isAxiosError } from 'axios'
import toast from 'react-hot-toast'
import api from '../services/api'
import Modal from '../components/Modal'
import ProductForm from '../components/ProductForm'
import type { Product } from '../types/product'

export default function ProductsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null)
  const [searchTerm, setSearchTerm] = useState<string>('')

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
    
    // Si hay un parámetro 'search' en la URL, aplicar filtro automáticamente
    const searchFromUrl = searchParams.get('search')
    if (searchFromUrl) {
      setSearchTerm(searchFromUrl)
      // Limpiar el parámetro de la URL después de usarlo
      setSearchParams({})
    }
  }, [searchParams, setSearchParams])

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
    } catch (err: unknown) {
      console.error(err)
      let errorMessage = 'No se pudo eliminar el producto'
      if (isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined
        errorMessage = data?.error || errorMessage
      }
      toast.error(errorMessage)
    }
  }

  const handleRequestReportByEmail = async () => {
    try {
      const response = await api.post<{ message: string }>('/reports/products/email')
      toast.success(response.data.message || '¡Listo! Te estamos mandando el reporte por mail.')
    } catch (err) {
      console.error(err)
      toast.error('No se pudo solicitar el reporte')
    }
  }

  // Filtrar productos según el término de búsqueda
  const filteredProducts = products.filter((product) => {
    if (!searchTerm) return true
    
    const term = searchTerm.toLowerCase()
    return (
      product.name.toLowerCase().includes(term) ||
      product.sku.toLowerCase().includes(term)
    )
  })

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Productos</h1>
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
          <button
            onClick={openCreateModal}
            className="px-4 py-2 rounded bg-green-600 text-white hover:bg-green-700"
          >
            Crear Nuevo Producto
          </button>
        </div>
      </div>

      {/* Campo de búsqueda */}
      <div className="mb-4">
        <div className="relative">
          <input
            type="text"
            placeholder="Buscar por nombre o SKU..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full px-4 py-2 pl-10 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <svg
            className="absolute left-3 top-2.5 h-5 w-5 text-gray-400"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clipRule="evenodd"
            />
          </svg>
          {searchTerm && (
            <button
              onClick={() => setSearchTerm('')}
              className="absolute right-3 top-2.5 text-gray-400 hover:text-gray-600"
            >
              <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          )}
        </div>
        {searchTerm && (
          <p className="mt-2 text-sm text-gray-600">
            Mostrando {filteredProducts.length} de {products.length} productos
          </p>
        )}
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
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Stock Actual</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Stock Mínimo</th>
                <th className="px-6 py-3" />
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredProducts.map((p) => {
                const isLowStock = p.quantity <= p.stock_minimo
                return (
                  <tr key={p.id} className={isLowStock ? 'bg-red-50' : ''}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {p.name}
                      {isLowStock && (
                        <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800">
                          ⚠️ Stock Bajo
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{p.sku}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <span className={isLowStock ? 'font-bold text-red-600' : 'text-gray-900'}>
                        {p.quantity}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{p.stock_minimo}</td>
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
                )
              })}
              {filteredProducts.length === 0 && (
                <tr>
                  <td className="px-6 py-4 text-sm text-gray-500" colSpan={5}>
                    {searchTerm ? 'No se encontraron productos con ese término de búsqueda.' : 'No hay productos para mostrar.'}
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
