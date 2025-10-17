import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import api from '../services/api'
import type { SalesOrder } from '../types/salesOrder'

export default function SalesOrdersPage() {
  const [orders, setOrders] = useState<SalesOrder[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let mounted = true
    const fetchOrders = async () => {
      try {
        setLoading(true)
        const res = await api.get<SalesOrder[]>('/sales-orders')
        if (mounted) setOrders(res.data)
      } catch (e) {
        console.error(e)
        if (mounted) setError('No se pudieron cargar las órdenes')
      } finally {
        if (mounted) setLoading(false)
      }
    }
    fetchOrders()
    return () => {
      mounted = false
    }
  }, [])

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Órdenes de Venta</h1>
        <Link to="/sales-orders/new" className="inline-block px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">
          Crear Nueva Venta
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
                <th className="px-4 py-2 text-left">ID de Cliente</th>
                <th className="px-4 py-2 text-left">Fecha de Orden</th>
                <th className="px-4 py-2 text-left">Estado</th>
                <th className="px-4 py-2 text-left">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((o) => (
                <tr key={o.id} className="border-t">
                  <td className="px-4 py-2">
                    <Link to={`/sales-orders/${o.id}`} className="text-indigo-600 hover:underline">{o.id}</Link>
                  </td>
                  <td className="px-4 py-2">{o.customer_id ?? '-'}</td>
                  <td className="px-4 py-2">{new Date(o.order_date).toLocaleString()}</td>
                  <td className="px-4 py-2">{o.status}</td>
                  <td className="px-4 py-2">
                    <Link to={`/sales-orders/${o.id}`} className="px-3 py-1 bg-gray-100 text-gray-700 rounded hover:bg-gray-200">
                      Ver Detalles
                    </Link>
                  </td>
                </tr>
              ))}
              {orders.length === 0 && (
                <tr>
                  <td className="px-4 py-4 text-center text-gray-500" colSpan={5}>
                    No hay órdenes de venta.
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
