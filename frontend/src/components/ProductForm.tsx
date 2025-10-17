import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import api from '../services/api'
import type { Product } from '../types/product'

type Props = {
  productToEdit: Product | null
  onSuccess: () => void
}

export default function ProductForm({ productToEdit, onSuccess }: Props) {
  const [name, setName] = useState('')
  const [sku, setSku] = useState('')
  const [quantity, setQuantity] = useState<number>(0)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (productToEdit) {
      setName(productToEdit.name)
      setSku(productToEdit.sku)
      setQuantity(productToEdit.quantity)
    } else {
      setName('')
      setSku('')
      setQuantity(0)
    }
  }, [productToEdit])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      if (productToEdit) {
        await api.put(`/products/${productToEdit.id}`, { name, sku, quantity })
        toast.success('Producto actualizado correctamente')
      } else {
        await api.post('/products', { name, sku, quantity })
        toast.success('Producto creado correctamente')
      }
      onSuccess()
    } catch (err) {
      console.error(err)
      toast.error('No se pudo guardar el producto')
      setError('No se pudo guardar el producto')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h2 className="text-lg font-semibold">
        {productToEdit ? 'Editar Producto' : 'Crear Producto'}
      </h2>
      {error && <p className="text-red-600 text-sm">{error}</p>}
      <div>
        <label className="block text-sm font-medium text-gray-700">Nombre</label>
        <input
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">SKU</label>
        <input
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
          value={sku}
          onChange={(e) => setSku(e.target.value)}
          required
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">Cantidad</label>
        <input
          type="number"
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
          value={quantity}
          onChange={(e) => setQuantity(Number(e.target.value))}
          required
          min={0}
        />
      </div>
      <div className="flex justify-end gap-2">
        <button
          type="submit"
          disabled={submitting}
          className="px-4 py-2 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {submitting ? 'Guardando...' : 'Guardar'}
        </button>
      </div>
    </form>
  )
}
