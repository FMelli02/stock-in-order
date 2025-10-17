import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../services/api'
import type { PurchaseOrderItemInput } from '../types/purchaseOrder'
import type { Product } from '../types/product'
import type { Supplier } from '../types/supplier'

export default function CreatePurchaseOrderPage() {
  const navigate = useNavigate()
  const [suppliers, setSuppliers] = useState<Supplier[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [selectedSupplierId, setSelectedSupplierId] = useState<string>('')
  const [selectedProductId, setSelectedProductId] = useState<string>('')
  const [selectedQty, setSelectedQty] = useState<number>(1)
  const [selectedCost, setSelectedCost] = useState<number>(0)
  const [orderItems, setOrderItems] = useState<PurchaseOrderItemInput[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    let mounted = true
    const load = async () => {
      try {
        setLoading(true)
        const [sRes, pRes] = await Promise.all([
          api.get<Supplier[]>('/suppliers'),
          api.get<Product[]>('/products'),
        ])
        if (mounted) {
          setSuppliers(sRes.data)
          setProducts(pRes.data)
        }
      } catch (e) {
        console.error(e)
        if (mounted) setError('No se pudieron cargar proveedores/productos')
      } finally {
        if (mounted) setLoading(false)
      }
    }
    load()
    return () => {
      mounted = false
    }
  }, [])

  const productsById = useMemo(() => {
    const m = new Map<number, Product>()
    for (const p of products) m.set(p.id, p)
    return m
  }, [products])

  const addItem = () => {
    const pid = Number(selectedProductId)
    if (!pid || selectedQty <= 0 || selectedCost < 0) return
    setOrderItems((prev) => {
      const i = prev.findIndex((it) => it.productId === pid)
      if (i >= 0) {
        const copy = [...prev]
        copy[i] = { ...copy[i], quantity: copy[i].quantity + selectedQty, unitCost: selectedCost }
        return copy
      }
      return [...prev, { productId: pid, quantity: selectedQty, unitCost: selectedCost }]
    })
    setSelectedProductId('')
    setSelectedQty(1)
    setSelectedCost(0)
  }

  const removeItem = (pid: number) => {
    setOrderItems((prev) => prev.filter((it) => it.productId !== pid))
  }

  const handleSubmit = async () => {
    const supplierIdNum = Number(selectedSupplierId)
    if (!supplierIdNum) {
      alert('Seleccioná un proveedor')
      return
    }
    if (orderItems.length === 0) {
      alert('Agregá al menos un ítem a la orden')
      return
    }
    try {
      setSubmitting(true)
      setError(null)
      const dto = {
        supplier_id: supplierIdNum,
        items: orderItems.map((it) => ({ product_id: it.productId, quantity: it.quantity, unit_cost: it.unitCost })),
      }
      await api.post('/purchase-orders', dto)
      navigate('/purchase-orders')
    } catch (e) {
      console.error(e)
      alert('No se pudo guardar la orden de compra')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Nueva Orden de Compra</h1>
      {loading && <p>Cargando...</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && (
        <div className="space-y-6">
          {/* Paso 1: Seleccionar Proveedor */}
          <div className="bg-white rounded shadow p-4">
            <h2 className="font-semibold mb-2">Seleccionar Proveedor</h2>
            <select
              className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              value={selectedSupplierId}
              onChange={(e) => setSelectedSupplierId(e.target.value)}
            >
              <option value="">-- Elegí un proveedor --</option>
              {suppliers.map((s) => (
                <option key={s.id} value={s.id}>{s.name} (#{s.id})</option>
              ))}
            </select>
          </div>

          {/* Paso 2: Añadir Ítem */}
          <div className="bg-white rounded shadow p-4">
            <h2 className="font-semibold mb-2">Añadir Ítem</h2>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
              <div>
                <label className="block text-sm font-medium text-gray-700">Producto</label>
                <select
                  className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  value={selectedProductId}
                  onChange={(e) => setSelectedProductId(e.target.value)}
                >
                  <option value="">-- Elegí un producto --</option>
                  {products.map((p) => (
                    <option key={p.id} value={p.id}>{p.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Cantidad</label>
                <input
                  type="number"
                  min={1}
                  className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  value={selectedQty}
                  onChange={(e) => setSelectedQty(Number(e.target.value))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Costo Unitario</label>
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  value={selectedCost}
                  onChange={(e) => setSelectedCost(Number(e.target.value))}
                />
              </div>
              <div>
                <button
                  type="button"
                  onClick={addItem}
                  className="w-full px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                >
                  Añadir a la Orden
                </button>
              </div>
            </div>
          </div>

          {/* Paso 3: Resumen */}
          <div className="bg-white rounded shadow p-4">
            <h2 className="font-semibold mb-2">Resumen de la Orden</h2>
            <div className="overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead className="bg-gray-100">
                  <tr>
                    <th className="px-4 py-2 text-left">Producto</th>
                    <th className="px-4 py-2 text-left">Cantidad</th>
                    <th className="px-4 py-2 text-left">Costo Unitario</th>
                    <th className="px-4 py-2 text-left">Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {orderItems.map((it) => (
                    <tr key={it.productId} className="border-t">
                      <td className="px-4 py-2">{productsById.get(it.productId)?.name ?? `#${it.productId}`}</td>
                      <td className="px-4 py-2">{it.quantity}</td>
                      <td className="px-4 py-2">{it.unitCost}</td>
                      <td className="px-4 py-2">
                        <button
                          type="button"
                          onClick={() => removeItem(it.productId)}
                          className="px-2 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
                        >
                          Eliminar
                        </button>
                      </td>
                    </tr>
                  ))}
                  {orderItems.length === 0 && (
                    <tr>
                      <td className="px-4 py-4 text-center text-gray-500" colSpan={4}>
                        Todavía no agregaste ítems.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => navigate('/purchase-orders')}
              className="px-4 py-2 rounded border border-gray-300"
            >
              Cancelar
            </button>
            <button
              type="button"
              disabled={submitting}
              onClick={handleSubmit}
              className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
            >
              {submitting ? 'Guardando...' : 'Guardar Orden de Compra'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
