import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { isAxiosError } from 'axios'
import toast from 'react-hot-toast'
import api from '../services/api'
import Modal from '../components/Modal'
import ProductForm from '../components/ProductForm'
import PaginationControls from '../components/PaginationControls'
import type { Product } from '../types/product'

interface PaginationMetadata {
  current_page: number
  page_size: number
  first_page: number
  last_page: number
  total_records: number
}

interface PaginatedResponse {
  items: Product[]
  metadata: PaginationMetadata
}

export default function ProductsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [products, setProducts] = useState<Product[]>([])
  const [metadata, setMetadata] = useState<PaginationMetadata>({
    current_page: 1,
    page_size: 20,
    first_page: 1,
    last_page: 1,
    total_records: 0
  })
  const [, setCurrentPage] = useState<number>(1)
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null)
  const [searchTerm, setSearchTerm] = useState<string>('')

  const fetchProducts = async (page: number = 1) => {
    try {
      setLoading(true)
      setError(null)
      const res = await api.get<PaginatedResponse>(`/products?page=${page}&page_size=20`)
      setProducts(res.data.items)
      setMetadata(res.data.metadata)
      setCurrentPage(page)
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

  const handlePageChange = (page: number) => {
    fetchProducts(page)
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
    <div className="p-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-display font-semibold text-neutral-900 mb-2">Productos</h1>
          <p className="text-neutral-600 text-sm">Gestiona tu inventario y catálogo</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={handleRequestReportByEmail}
            className="px-4 py-2.5 rounded-xl bg-neutral-100 hover:bg-neutral-200 text-neutral-900 font-medium transition-all border border-neutral-300 flex items-center gap-2 text-sm"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z" />
              <path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z" />
            </svg>
            Recibir por Email
          </button>
          <button
            onClick={openCreateModal}
            className="px-4 py-2.5 rounded-xl bg-primary-600 hover:bg-primary-700 text-white font-medium transition-all shadow-soft hover:shadow-soft-lg text-sm"
          >
            Crear Nuevo Producto
          </button>
        </div>
      </div>

      {/* Campo de búsqueda */}
      <div className="mb-6">
        <div className="relative">
          <input
            type="text"
            placeholder="Buscar por nombre o SKU..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full px-4 py-3 pl-11 border border-neutral-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-all text-sm placeholder:text-neutral-400"
          />
          <svg
            className="absolute left-4 top-3.5 h-5 w-5 text-neutral-400"
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
              className="absolute right-4 top-3.5 text-neutral-400 hover:text-neutral-600 transition-colors"
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
          <p className="mt-2 text-sm text-neutral-600">
            Mostrando {filteredProducts.length} de {products.length} productos
          </p>
        )}
      </div>

      {loading && (
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <div className="w-10 h-10 border-3 border-neutral-200 border-t-primary-600 rounded-full animate-spin mb-3 mx-auto"></div>
            <p className="text-sm text-neutral-600">Cargando productos...</p>
          </div>
        </div>
      )}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-xl text-sm">
          {error}
        </div>
      )}
      {!loading && !error && (
        <div className="bg-white rounded-2xl border border-neutral-200 shadow-soft overflow-hidden">
          <table className="min-w-full divide-y divide-neutral-200">
            <thead className="bg-neutral-50">
              <tr>
                <th className="px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider">Nombre</th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider">SKU</th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider">Stock Actual</th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider">Stock Mínimo</th>
                <th className="px-6 py-4" />
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-100">
              {filteredProducts.map((p) => {
                const isLowStock = p.quantity <= p.stock_minimo
                const isOutOfStock = p.quantity === 0
                
                // Determinar el color de fondo según el nivel de stock
                let rowBgClass = 'hover:bg-neutral-50 transition-colors'
                if (isOutOfStock) {
                  rowBgClass = 'bg-red-50/50 hover:bg-red-50'
                } else if (isLowStock) {
                  rowBgClass = 'bg-amber-50/30 hover:bg-amber-50'
                }
                
                return (
                  <tr key={p.id} className={rowBgClass}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-neutral-900">
                      {p.name}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-neutral-600">{p.sku}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <span className={isLowStock ? 'font-semibold text-red-600' : 'text-neutral-900'}>
                        {p.quantity}
                        {isLowStock && (
                          <span 
                            className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-700"
                            title="Stock por debajo del mínimo definido"
                          >
                            Bajo stock
                          </span>
                        )}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-neutral-600">{p.stock_minimo}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-right">
                      <button
                        onClick={() => openEditModal(p)}
                        className="mr-2 px-3 py-1.5 rounded-lg bg-primary-600 text-white hover:bg-primary-700 font-medium transition-all text-xs"
                      >
                        Editar
                      </button>
                      <button
                        onClick={() => handleDelete(p)}
                        className="px-3 py-1.5 rounded-lg bg-red-600 text-white hover:bg-red-700 font-medium transition-all text-xs"
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

          <PaginationControls 
            metadata={metadata} 
            onPageChange={handlePageChange} 
          />
        </div>
      )}

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <ProductForm productToEdit={selectedProduct} onSuccess={closeModalAndRefresh} />
      </Modal>
    </div>
  )
}
