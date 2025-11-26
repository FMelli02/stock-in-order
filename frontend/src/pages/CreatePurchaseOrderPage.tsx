import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
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
  const [selectedLote, setSelectedLote] = useState<string>('')
  const [selectedExpiry, setSelectedExpiry] = useState<string>('')
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
          api.get('/suppliers?page=1&page_size=1000'),
          api.get('/products?page=1&page_size=1000'),
        ])
        if (mounted) {
          // Las APIs retornan objetos paginados con estructura { items: [], metadata: {} }
          setSuppliers(sRes.data.items || sRes.data || [])
          setProducts(pRes.data.items || pRes.data || [])
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

  // Calcular el total de la orden
  const orderTotal = useMemo(() => {
    return orderItems.reduce((sum, item) => {
      return sum + (item.quantity * item.unitCost)
    }, 0)
  }, [orderItems])

  const addItem = () => {
    const pid = Number(selectedProductId)
    if (!pid || selectedQty <= 0 || selectedCost < 0) return
    setOrderItems((prev) => {
      const i = prev.findIndex((it) => it.productId === pid)
      if (i >= 0) {
        const copy = [...prev]
        copy[i] = { 
          ...copy[i], 
          quantity: copy[i].quantity + selectedQty, 
          unitCost: selectedCost,
          loteNumber: selectedLote || undefined,
          expiryDate: selectedExpiry || undefined
        }
        return copy
      }
      return [...prev, { 
        productId: pid, 
        quantity: selectedQty, 
        unitCost: selectedCost,
        loteNumber: selectedLote || undefined,
        expiryDate: selectedExpiry || undefined
      }]
    })
    setSelectedProductId('')
    setSelectedQty(1)
    setSelectedCost(0)
    setSelectedLote('')
    setSelectedExpiry('')
  }

  const removeItem = (pid: number) => {
    setOrderItems((prev) => prev.filter((it) => it.productId !== pid))
  }

  const handleSubmit = async () => {
    const supplierIdNum = Number(selectedSupplierId)
    if (!supplierIdNum) {
      toast.error('Seleccioná un proveedor')
      return
    }
    if (orderItems.length === 0) {
      toast.error('Agregá al menos un ítem a la orden')
      return
    }
    try {
      setSubmitting(true)
      setError(null)
      const dto = {
        supplier_id: supplierIdNum,
        items: orderItems.map((it) => ({ 
          product_id: it.productId, 
          quantity: it.quantity, 
          unit_cost: it.unitCost,
          lote_number: it.loteNumber || '',
          expiry_date: it.expiryDate || null
        })),
      }
      await api.post('/purchase-orders', dto)
      toast.success('Orden de compra creada correctamente')
      navigate('/purchase-orders')
    } catch (e) {
      console.error(e)
      toast.error('No se pudo guardar la orden de compra')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-display font-bold mb-4 text-neutral-900">Nueva Orden de Compra</h1>
      {loading && <p className="text-neutral-600">Cargando...</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && (
        <div className="space-y-6">
          {/* Paso 1: Seleccionar Proveedor */}
          <div className="bg-white rounded-2xl shadow-soft border border-neutral-200 p-6">
            <h2 className="font-display font-semibold mb-4 text-neutral-900">Seleccionar Proveedor</h2>
            <select
              className="mt-1 block w-full rounded-xl border-neutral-300 px-4 py-2.5 shadow-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500"
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
          <div className="bg-white rounded-2xl shadow-soft border border-neutral-200 p-6">
            <h2 className="font-display font-semibold mb-4 text-neutral-900">Añadir Ítem</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end mb-4">
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-2">Producto</label>
                <select
                  className="mt-1 block w-full rounded-xl border-neutral-300 px-4 py-2.5 shadow-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500"
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
                <label className="block text-sm font-medium text-neutral-700 mb-2">Cantidad</label>
                <input
                  type="number"
                  min={1}
                  className="mt-1 block w-full rounded-xl border-neutral-300 px-4 py-2.5 shadow-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500"
                  value={selectedQty}
                  onChange={(e) => setSelectedQty(Number(e.target.value))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-2">Costo Unitario</label>
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  className="mt-1 block w-full rounded-xl border-neutral-300 px-4 py-2.5 shadow-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500"
                  value={selectedCost}
                  onChange={(e) => setSelectedCost(Number(e.target.value))}
                />
              </div>
            </div>
            
            {/* Campos opcionales de lote */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-2">
                  Número de Lote <span className="text-neutral-500 text-xs">(opcional)</span>
                </label>
                <input
                  type="text"
                  className="mt-1 block w-full rounded-xl border-neutral-300 px-4 py-2.5 shadow-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500"
                  value={selectedLote}
                  onChange={(e) => setSelectedLote(e.target.value)}
                  placeholder="Ej: LOTE-2024-001"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-2">
                  Fecha de Vencimiento <span className="text-neutral-500 text-xs">(opcional)</span>
                </label>
                <input
                  type="date"
                  className="mt-1 block w-full rounded-xl border-neutral-300 px-4 py-2.5 shadow-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500"
                  value={selectedExpiry}
                  onChange={(e) => setSelectedExpiry(e.target.value)}
                />
              </div>
              <div>
                <button
                  type="button"
                  onClick={addItem}
                  className="w-full px-4 py-2.5 bg-primary-600 text-white rounded-xl hover:bg-primary-700 shadow-soft"
                >
                  Añadir a la Orden
                </button>
              </div>
            </div>
          </div>

          {/* Paso 3: Resumen */}
          <div className="bg-white rounded-2xl shadow-soft border border-neutral-200 p-6">
            <h2 className="font-display font-semibold mb-4 text-neutral-900">Resumen de la Orden</h2>
            <div className="overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead className="bg-neutral-50">
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
                          className="px-3 py-1.5 text-sm bg-red-600 text-white rounded-lg hover:bg-red-700 shadow-soft"
                        >
                          Eliminar
                        </button>
                      </td>
                    </tr>
                  ))}
                  {orderItems.length === 0 && (
                    <tr>
                      <td className="px-4 py-4 text-center text-neutral-500" colSpan={4}>
                        Todavía no agregaste ítems.
                      </td>
                    </tr>
                  )}
                </tbody>
                <tfoot className="bg-neutral-50 font-semibold">
                  <tr>
                    <td className="px-4 py-2 text-right" colSpan={3}>Total:</td>
                    <td className="px-4 py-2">${orderTotal.toFixed(2)}</td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>

          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={() => navigate('/purchase-orders')}
              className="px-6 py-2.5 rounded-xl border border-neutral-300 hover:bg-neutral-50 transition"
            >
              Cancelar
            </button>
            <button
              type="button"
              disabled={submitting}
              onClick={handleSubmit}
              className="px-6 py-2.5 bg-primary-600 text-white rounded-xl hover:bg-primary-700 disabled:opacity-50 shadow-soft transition"
            >
              {submitting ? 'Guardando...' : 'Guardar Orden de Compra'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
