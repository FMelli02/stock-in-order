import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'react-hot-toast'
import api from '../services/api'
import PaginationControls from '../components/PaginationControls'
import type { SalesOrder } from '../types/salesOrder'

interface PaginationMetadata {
  current_page: number
  page_size: number
  first_page: number
  last_page: number
  total_records: number
}

interface PaginatedResponse {
  items: SalesOrder[]
  metadata: PaginationMetadata
}

export default function SalesOrdersPage() {
  const [orders, setOrders] = useState<SalesOrder[]>([])
  const [metadata, setMetadata] = useState<PaginationMetadata>({
    current_page: 1,
    page_size: 20,
    first_page: 1,
    last_page: 1,
    total_records: 0
  })
  const [, setCurrentPage] = useState<number>(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  // Filtros para exportación
  const [filterDateFrom, setFilterDateFrom] = useState('')
  const [filterDateTo, setFilterDateTo] = useState('')
  const [filterStatus, setFilterStatus] = useState('')

  const fetchOrders = async (page: number = 1) => {
    try {
      setLoading(true)
      const res = await api.get<PaginatedResponse>(`/sales-orders?page=${page}&page_size=20`)
      setOrders(res.data.items)
      setMetadata(res.data.metadata)
      setCurrentPage(page)
    } catch (e) {
      console.error(e)
      setError('No se pudieron cargar las órdenes')
    } finally {
      setLoading(false)
    }
  }

  const handlePageChange = (page: number) => {
    fetchOrders(page)
  }

  useEffect(() => {
    fetchOrders()
  }, [])

  const handleExportExcel = async () => {
    try {
      // Construir URL con filtros
      const params = new URLSearchParams()
      if (filterDateFrom) params.append('date_from', filterDateFrom)
      if (filterDateTo) params.append('date_to', filterDateTo)
      if (filterStatus) params.append('status', filterStatus)

      const url = `/reports/sales-orders/xlsx${params.toString() ? `?${params.toString()}` : ''}`
      
      const response = await api.get(url, {
        responseType: 'blob',
      })

      const blob = new Blob([response.data], { 
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
      })
      const downloadUrl = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = downloadUrl
      link.setAttribute('download', 'ventas.xlsx')
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(downloadUrl)

      toast.success('Órdenes de venta exportadas a Excel correctamente')
    } catch (err) {
      console.error(err)
      toast.error('Error al exportar órdenes')
    }
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-display font-semibold text-neutral-900 mb-2">Órdenes de Venta</h1>
          <p className="text-neutral-600 text-sm">Gestiona tus ventas y pedidos</p>
        </div>
        <Link to="/sales-orders/new" className="px-4 py-2.5 rounded-xl bg-primary-600 hover:bg-primary-700 text-white font-medium transition-all shadow-soft hover:shadow-soft-lg text-sm">
          Crear Nueva Venta
        </Link>
      </div>

      {/* Sección de Filtros y Exportación */}
      <div className="bg-white rounded-2xl border border-neutral-200 shadow-soft p-6 mb-6">
        <h2 className="text-lg font-display font-semibold text-neutral-900 mb-4">Exportar a Excel</h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-2">Fecha Desde</label>
            <input
              type="date"
              value={filterDateFrom}
              onChange={(e) => setFilterDateFrom(e.target.value)}
              className="w-full px-4 py-2.5 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all text-sm"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-2">Fecha Hasta</label>
            <input
              type="date"
              value={filterDateTo}
              onChange={(e) => setFilterDateTo(e.target.value)}
              className="w-full px-4 py-2.5 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all text-sm"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-2">Estado</label>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="w-full px-4 py-2.5 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all text-sm"
            >
              <option value="">Todos</option>
              <option value="pending">Pendiente</option>
              <option value="completed">Completado</option>
              <option value="cancelled">Cancelado</option>
            </select>
          </div>
          <div className="flex items-end">
            <button
              onClick={handleExportExcel}
              className="w-full px-4 py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-medium transition-all shadow-soft flex items-center justify-center gap-2 text-sm"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
              Exportar Excel
            </button>
          </div>
        </div>
      </div>
      {loading && (
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <div className="w-10 h-10 border-3 border-neutral-200 border-t-primary-600 rounded-full animate-spin mb-3 mx-auto"></div>
            <p className="text-sm text-neutral-600">Cargando órdenes...</p>
          </div>
        </div>
      )}
      {error && <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-xl text-sm">{error}</div>}
      {!loading && !error && (
        <div className="bg-white rounded-2xl border border-neutral-200 shadow-soft overflow-hidden">
          <table className="min-w-full">
            <thead className="bg-neutral-50">
              <tr>
                <th className="px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider">ID de Orden</th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider">Cliente</th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider">Fecha de Orden</th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider">Estado</th>
                <th className="px-4 py-2 text-left">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((o) => (
                <tr key={o.id} className="border-t">
                  <td className="px-4 py-2">
                    <Link to={`/sales-orders/${o.id}`} className="text-indigo-600 hover:underline">{o.id}</Link>
                  </td>
                  <td className="px-4 py-2">{o.customer_name ?? '-'}</td>
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

          <PaginationControls 
            metadata={metadata} 
            onPageChange={handlePageChange} 
          />
        </div>
      )}
    </div>
  )
}
