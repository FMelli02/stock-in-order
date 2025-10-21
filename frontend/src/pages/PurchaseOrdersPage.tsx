import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import toast from 'react-hot-toast'
import api from '../services/api'
import type { PurchaseOrder } from '../types/purchaseOrder'

export default function PurchaseOrdersPage() {
  const [orders, setOrders] = useState<PurchaseOrder[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [updating, setUpdating] = useState<number | null>(null)

  const load = async () => {
    try {
      setLoading(true)
      const res = await api.get<PurchaseOrder[]>('/purchase-orders')
      setOrders(res.data)
    } catch (e) {
      console.error(e)
      setError('No se pudieron cargar las órdenes de compra')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const markReceived = async (id: number) => {
    try {
      setUpdating(id)
      await api.put(`/purchase-orders/${id}/status`, { status: 'completed' })
      toast.success('Estado actualizado a completado')
      await load()
    } catch (e) {
      console.error(e)
      toast.error('No se pudo actualizar el estado')
    } finally {
      setUpdating(null)
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Órdenes de Compra</h1>
        <Link to="/purchase-orders/new" className="inline-block px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">
          Crear Nueva Compra
        </Link>
      </div>
      {loading && <p>Cargando...</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && (
        <div className="overflow-x-auto bg-white rounded shadow">
          <table className="min-w-full text-sm">
            <thead className="bg-gray-100">
              <tr>
                <th className="px-4 py-2 text-left">ID de Orden</th>
                <th className="px-4 py-2 text-left">Proveedor</th>
                <th className="px-4 py-2 text-left">Fecha de Orden</th>
                <th className="px-4 py-2 text-left">Estado</th>
                <th className="px-4 py-2 text-left">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((o) => (
                <tr key={o.id} className="border-t">
                  <td className="px-4 py-2">
                    <Link to={`/purchase-orders/${o.id}`} className="text-indigo-600 hover:underline">{o.id}</Link>
                  </td>
                  <td className="px-4 py-2">{o.supplier_name ?? '-'}</td>
                  <td className="px-4 py-2">{o.order_date ? new Date(o.order_date).toLocaleDateString() : '-'}</td>
                  <td className="px-4 py-2">{o.status}</td>
                  <td className="px-4 py-2">
                    <div className="flex gap-2 items-center">
                      <Link to={`/purchase-orders/${o.id}`} className="px-3 py-1 bg-gray-100 text-gray-700 rounded hover:bg-gray-200">
                        Ver Detalles
                      </Link>
                      {o.status === 'pending' ? (
                      <button
                        type="button"
                        onClick={() => markReceived(o.id)}
                        disabled={updating === o.id}
                        className="px-3 py-1 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
                      >
                        {updating === o.id ? 'Actualizando...' : 'Marcar como Recibida'}
                      </button>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {orders.length === 0 && (
                <tr>
                  <td className="px-4 py-4 text-center text-gray-500" colSpan={5}>
                    No hay órdenes de compra.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
