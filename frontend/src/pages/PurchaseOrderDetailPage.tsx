import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import api from '../services/api'
import type { PurchaseOrder, PurchaseOrderItem } from '../types/purchaseOrder'
import type { Product } from '../types/product'

type PurchaseOrderDetailResponse = {
  order: PurchaseOrder
  items: PurchaseOrderItem[]
}

export default function PurchaseOrderDetailPage() {
  const { id } = useParams()
  const orderId = Number(id)
  const [order, setOrder] = useState<PurchaseOrder | null>(null)
  const [items, setItems] = useState<PurchaseOrderItem[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let mounted = true
    const fetchAll = async () => {
      if (!orderId) {
        setError('ID de orden inválido')
        setLoading(false)
        return
      }
      try {
        setLoading(true)
        setError(null)
        const [orderRes, productsRes] = await Promise.all([
          api.get<PurchaseOrderDetailResponse>(`/purchase-orders/${orderId}`),
          api.get<Product[]>('/products'),
        ])
        if (!mounted) return
        setOrder(orderRes.data.order)
        setItems(orderRes.data.items)
        setProducts(productsRes.data)
      } catch (e) {
        console.error(e)
        if (mounted) setError('No se pudo cargar la orden de compra')
      } finally {
        if (mounted) setLoading(false)
      }
    }
    fetchAll()
    return () => {
      mounted = false
    }
  }, [orderId])

  const productsById = useMemo(() => {
    const m = new Map<number, Product>()
    for (const p of products) m.set(p.id, p)
    return m
  }, [products])

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Detalle de Orden de Compra</h1>
        <Link to="/purchase-orders" className="text-indigo-600 hover:underline">Volver a Órdenes</Link>
      </div>

      {loading && <p>Cargando...</p>}
      {error && <p className="text-red-600">{error}</p>}

      {!loading && !error && order && (
        <div className="space-y-6">
          <div className="bg-white rounded shadow p-4">
            <h2 className="font-semibold mb-2">Información de la Orden</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2 text-sm">
              <div><span className="text-gray-500">ID:</span> <span className="font-medium">#{order.id}</span></div>
              <div><span className="text-gray-500">Fecha:</span> <span className="font-medium">{new Date(order.order_date).toLocaleString()}</span></div>
              <div><span className="text-gray-500">Estado:</span> <span className="font-medium">{order.status}</span></div>
              <div><span className="text-gray-500">Proveedor:</span> <span className="font-medium">{order.supplier_id ?? '-'}</span></div>
            </div>
          </div>

          <div className="bg-white rounded shadow p-4">
            <h2 className="font-semibold mb-2">Ítems de la Orden</h2>
            <div className="overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead className="bg-gray-100">
                  <tr>
                    <th className="px-4 py-2 text-left">Producto (Nombre/SKU)</th>
                    <th className="px-4 py-2 text-left">Cantidad</th>
                    <th className="px-4 py-2 text-left">Costo Unitario</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((it) => {
                    const p = productsById.get(Number(it.product_id))
                    const label = p ? `${p.name} (${p.sku})` : `#${it.product_id}`
                    return (
                      <tr key={it.id} className="border-t">
                        <td className="px-4 py-2">{label}</td>
                        <td className="px-4 py-2">{it.quantity}</td>
                        <td className="px-4 py-2">{it.unit_cost}</td>
                      </tr>
                    )
                  })}
                  {items.length === 0 && (
                    <tr>
                      <td className="px-4 py-4 text-center text-gray-500" colSpan={3}>
                        Esta orden no tiene ítems.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
