import { useEffect, useState } from 'react'
import { isAxiosError } from 'axios'
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
    const ok = window.confirm('¿Estás seguro?')
    if (!ok) return
    try {
      await api.delete(`/products/${product.id}`)
      await fetchProducts()
    } catch (err) {
      console.error(err)
      alert('No se pudo eliminar el producto')
    }
  }

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Productos</h1>
        <button
          onClick={openCreateModal}
          className="px-4 py-2 rounded bg-green-600 text-white hover:bg-green-700"
        >
          Crear Nuevo Producto
        </button>
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
