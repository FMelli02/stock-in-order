import { useEffect, useState } from 'react'
import api from '../services/api'
import type { DashboardMetrics } from '../types/dashboard'
import MetricCard from '../components/MetricCard'

export default function DashboardPage() {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let mounted = true
    const load = async () => {
      try {
        setLoading(true)
        const res = await api.get<DashboardMetrics>('/dashboard/metrics')
        if (mounted) setMetrics(res.data)
      } catch (e) {
        console.error(e)
        if (mounted) setError('No se pudieron cargar las métricas')
      } finally {
        if (mounted) setLoading(false)
      }
    }
    load()
    return () => {
      mounted = false
    }
  }, [])

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Resumen del Negocio</h1>
      {loading && <p>Cargando...</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && metrics && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <MetricCard title="Productos Totales" value={metrics.total_products} />
          <MetricCard title="Clientes Totales" value={metrics.total_customers} />
          <MetricCard title="Proveedores Totales" value={metrics.total_suppliers} />
          <MetricCard title="Ventas Pendientes" value={metrics.pending_sales_orders} color="border border-yellow-400" />
          <MetricCard title="Productos con Bajo Stock" value={metrics.products_low_stock} color="border border-red-500" />
        </div>
      )}
    </div>
  )
}
