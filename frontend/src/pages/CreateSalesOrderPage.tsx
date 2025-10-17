import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import api from '../services/api'
import type { OrderItemInput } from '../types/salesOrder'
import type { Product } from '../types/product'
import type { Customer } from '../types/customer'

export default function CreateSalesOrderPage() {
  const navigate = useNavigate()
  const [customers, setCustomers] = useState<Customer[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [selectedCustomerId, setSelectedCustomerId] = useState<string>('')
  const [selectedProductId, setSelectedProductId] = useState<string>('')
  const [selectedQty, setSelectedQty] = useState<number>(1)
  const [qtyError, setQtyError] = useState<string>('')
  const [orderItems, setOrderItems] = useState<OrderItemInput[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    let mounted = true
    const load = async () => {
      try {
        setLoading(true)
        const [cRes, pRes] = await Promise.all([
          api.get<Customer[]>('/customers'),
          api.get<Product[]>('/products'),
        ])
        if (mounted) {
          setCustomers(cRes.data)
          setProducts(pRes.data)
        }
      } catch (e) {
        console.error(e)
        if (mounted) setError('No se pudieron cargar clientes/productos')
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
      const product = productsById.get(item.productId)
      if (product) {
        return sum + (item.quantity * product.quantity)
      }
      return sum
    }, 0)
  }, [orderItems, productsById])

  // Obtener el producto seleccionado actualmente
  const selectedProduct = useMemo(() => {
    const pid = Number(selectedProductId)
    return pid ? productsById.get(pid) : null
  }, [selectedProductId, productsById])

  // Validar cantidad contra stock disponible
  const handleQtyChange = (value: number) => {
    setSelectedQty(value)
    if (selectedProduct && value > selectedProduct.quantity) {
      setQtyError(`La cantidad no puede superar el stock disponible (${selectedProduct.quantity})`)
    } else if (value <= 0) {
      setQtyError('La cantidad debe ser mayor a 0')
    } else {
      setQtyError('')
    }
  }

  const canAddItem = useMemo(() => {
    if (!selectedProductId || selectedQty <= 0) return false
    if (selectedProduct && selectedQty > selectedProduct.quantity) return false
    return true
  }, [selectedProductId, selectedQty, selectedProduct])

  const addItem = () => {
    const pid = Number(selectedProductId)
    if (!pid || selectedQty <= 0) return
    setOrderItems((prev) => {
      const i = prev.findIndex((it) => it.productId === pid)
      if (i >= 0) {
        const copy = [...prev]
        copy[i] = { ...copy[i], quantity: copy[i].quantity + selectedQty }
        return copy
      }
      return [...prev, { productId: pid, quantity: selectedQty }]
    })
    setSelectedProductId('')
    setSelectedQty(1)
  }

  const removeItem = (pid: number) => {
    setOrderItems((prev) => prev.filter((it) => it.productId !== pid))
  }

  const handleSubmit = async () => {
    const customerIdNum = Number(selectedCustomerId)
    if (!customerIdNum) {
      toast.error('Seleccioná un cliente')
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
        customer_id: customerIdNum,
        items: orderItems.map((it) => ({ product_id: it.productId, quantity: it.quantity })),
      }
      await api.post('/sales-orders', dto)
      toast.success('Orden de venta creada correctamente')
      navigate('/sales-orders')
    } catch (e) {
      console.error(e)
      toast.error('No se pudo guardar la orden')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Nueva Orden de Venta</h1>
      {loading && <p>Cargando...</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && (
        <div className="space-y-6">
          {/* Paso 1: Seleccionar Cliente */}
          <div className="bg-white rounded shadow p-4">
            <h2 className="font-semibold mb-2">Seleccionar Cliente</h2>
            <select
              className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              value={selectedCustomerId}
              onChange={(e) => setSelectedCustomerId(e.target.value)}
            >
              <option value="">-- Elegí un cliente --</option>
              {customers.map((c) => (
                <option key={c.id} value={c.id}>{c.name} (#{c.id})</option>
              ))}
            </select>
          </div>

          {/* Paso 2: Añadir Ítem */}
          <div className="bg-white rounded shadow p-4">
            <h2 className="font-semibold mb-2">Añadir Ítem</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
              <div>
                <label className="block text-sm font-medium text-gray-700">Producto</label>
                <select
                  className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  value={selectedProductId}
                  onChange={(e) => setSelectedProductId(e.target.value)}
                >
                  <option value="">-- Elegí un producto --</option>
                  {products.map((p) => (
                    <option key={p.id} value={p.id}>{p.name} (Stock: {p.quantity})</option>
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
                  onChange={(e) => handleQtyChange(Number(e.target.value))}
                />
                {qtyError && <p className="text-xs text-red-600 mt-1">{qtyError}</p>}
              </div>
              <div>
                <button
                  type="button"
                  onClick={addItem}
                  disabled={!canAddItem}
                  className="w-full px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
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
                    <th className="px-4 py-2 text-left">Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {orderItems.map((it) => (
                    <tr key={it.productId} className="border-t">
                      <td className="px-4 py-2">{productsById.get(it.productId)?.name ?? `#${it.productId}`}</td>
                      <td className="px-4 py-2">{it.quantity}</td>
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
                      <td className="px-4 py-4 text-center text-gray-500" colSpan={3}>
                        Todavía no agregaste ítems.
                      </td>
                    </tr>
                  )}
                </tbody>
                <tfoot className="bg-gray-50 font-semibold">
                  <tr>
                    <td className="px-4 py-2 text-right" colSpan={2}>Total:</td>
                    <td className="px-4 py-2">${orderTotal.toFixed(2)}</td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => navigate('/sales-orders')}
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
              {submitting ? 'Guardando...' : 'Guardar Orden'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
