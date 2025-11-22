import { useEffect, useState } from 'react'
import api from '../services/api'
import { isAxiosError } from 'axios'

interface ProductBatch {
  id: number
  product_id: number
  user_id: number
  lote_number: string | null
  quantity: number
  expiry_date: string | null
  created_at: string
}

interface ProductBatchesTableProps {
  productId: number
}

const ProductBatchesTable: React.FC<ProductBatchesTableProps> = ({ productId }) => {
  const [batches, setBatches] = useState<ProductBatch[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchBatches = async () => {
      try {
        setLoading(true)
        setError(null)
        const response = await api.get<ProductBatch[]>(`/products/${productId}/batches`)
        setBatches(response.data)
      } catch (err) {
        let message = 'Error al cargar lotes'
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

    fetchBatches()
  }, [productId])

  const getDaysUntilExpiry = (expiryDate: string | null): number | null => {
    if (!expiryDate) return null
    const expiry = new Date(expiryDate)
    const today = new Date()
    const diffTime = expiry.getTime() - today.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    return diffDays
  }

  const getRowClass = (expiryDate: string | null): string => {
    const daysUntilExpiry = getDaysUntilExpiry(expiryDate)
    if (daysUntilExpiry === null) return ''
    
    if (daysUntilExpiry < 0) return 'bg-red-100' // Vencido
    if (daysUntilExpiry <= 15) return 'bg-red-50' // Vence en 15 días o menos
    if (daysUntilExpiry <= 30) return 'bg-yellow-50' // Vence en 30 días o menos
    
    return ''
  }

  const formatDate = (dateString: string | null): string => {
    if (!dateString) return 'Sin vencimiento'
    const date = new Date(dateString)
    return date.toLocaleDateString('es-AR', { 
      year: 'numeric', 
      month: '2-digit', 
      day: '2-digit' 
    })
  }

  const getExpiryLabel = (expiryDate: string | null): React.ReactElement | null => {
    const daysUntilExpiry = getDaysUntilExpiry(expiryDate)
    if (daysUntilExpiry === null) return null
    
    if (daysUntilExpiry < 0) {
      return (
        <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800">
          ❌ Vencido ({Math.abs(daysUntilExpiry)} días)
        </span>
      )
    }
    
    if (daysUntilExpiry <= 15) {
      return (
        <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800">
          ⚠️ Vence en {daysUntilExpiry} días
        </span>
      )
    }
    
    if (daysUntilExpiry <= 30) {
      return (
        <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
          ⚠ Vence en {daysUntilExpiry} días
        </span>
      )
    }
    
    return null
  }

  if (loading) {
    return <div className="text-gray-600 py-4">Cargando lotes...</div>
  }

  if (error) {
    return <div className="text-red-600 py-4">{error}</div>
  }

  if (batches.length === 0) {
    return (
      <div className="bg-gray-50 rounded-lg p-6 text-center">
        <p className="text-gray-600">No hay lotes activos para este producto</p>
        <p className="text-sm text-gray-500 mt-1">
          Los lotes se crean automáticamente al recibir órdenes de compra
        </p>
      </div>
    )
  }

  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <div className="px-4 py-3 bg-gray-50 border-b border-gray-200">
        <h3 className="text-lg font-semibold text-gray-900">
          Lotes Activos ({batches.length})
        </h3>
        <p className="text-sm text-gray-600 mt-1">
          Stock total: {batches.reduce((sum, b) => sum + b.quantity, 0)} unidades
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Nro Lote
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Cantidad
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Vencimiento
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Ingreso
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {batches.map((batch) => (
              <tr key={batch.id} className={getRowClass(batch.expiry_date)}>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {batch.lote_number || 'N/A'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {batch.quantity}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {formatDate(batch.expiry_date)}
                  {getExpiryLabel(batch.expiry_date)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {formatDate(batch.created_at)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="px-4 py-3 bg-gray-50 border-t border-gray-200">
        <div className="flex items-center gap-4 text-xs text-gray-600">
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 bg-red-100 rounded"></div>
            <span>Vencido o próximo a vencer (&lt;15 días)</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 bg-yellow-50 rounded border border-yellow-200"></div>
            <span>Vence pronto (15-30 días)</span>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ProductBatchesTable
