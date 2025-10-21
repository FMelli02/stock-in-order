import { useEffect, useState } from 'react'
import api from '../services/api'
import type { DashboardKPIs, ChartData } from '../types/dashboard'
import MetricCard from '../components/MetricCard'
import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Tooltip,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Legend,
} from 'recharts'

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8']

export default function DashboardPage() {
  const [kpis, setKpis] = useState<DashboardKPIs | null>(null)
  const [chartData, setChartData] = useState<ChartData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let mounted = true
    const load = async () => {
      try {
        setLoading(true)
        
        // Cargar KPIs y datos de gráficos en paralelo
        const [kpisRes, chartsRes] = await Promise.all([
          api.get<DashboardKPIs>('/dashboard/kpis'),
          api.get<ChartData>('/dashboard/charts'),
        ])
        
        if (mounted) {
          setKpis(kpisRes.data)
          setChartData(chartsRes.data)
        }
      } catch (e) {
        console.error(e)
        if (mounted) setError('No se pudieron cargar las métricas del dashboard')
      } finally {
        if (mounted) setLoading(false)
      }
    }
    load()
    return () => {
      mounted = false
    }
  }, [])

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('es-AR', {
      style: 'currency',
      currency: 'ARS',
    }).format(value)
  }

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold mb-6">Dashboard - Resumen del Negocio</h1>
      
      {loading && (
        <div className="flex items-center justify-center h-64">
          <p className="text-lg text-gray-600">Cargando dashboard...</p>
        </div>
      )}
      
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}
      
      {!loading && !error && kpis && chartData && (
        <>
          {/* KPIs Section */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <MetricCard 
              title="Total de Productos" 
              value={kpis.total_products} 
            />
            <MetricCard 
              title="Productos Bajo Stock" 
              value={kpis.low_stock_products} 
              color="border-2 border-red-500" 
            />
            <MetricCard 
              title="Ventas del Mes" 
              value={formatCurrency(kpis.current_month_sales)} 
            />
            <MetricCard 
              title="Órdenes Pendientes" 
              value={kpis.pending_sales_orders} 
              color="border-2 border-yellow-400" 
            />
          </div>

          {/* Charts Section */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Top 5 Productos Más Vendidos */}
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-xl font-bold mb-4 text-gray-800">
                Top 5 Productos Más Vendidos
              </h2>
              {chartData.top_selling_products.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <PieChart>
                    <Pie
                      data={chartData.top_selling_products as any}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({ product_name, total_sold }) =>
                        `${product_name}: ${total_sold}`
                      }
                      outerRadius={100}
                      fill="#8884d8"
                      dataKey="total_sold"
                    >
                      {chartData.top_selling_products.map((_, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              ) : (
                <div className="h-300 flex items-center justify-center text-gray-500">
                  No hay datos de ventas disponibles
                </div>
              )}
            </div>

            {/* Evolución de Ventas (Últimos 30 días) */}
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-xl font-bold mb-4 text-gray-800">
                Ventas en los Últimos 30 Días
              </h2>
              {chartData.sales_evolution.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={chartData.sales_evolution}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis 
                      dataKey="date" 
                      angle={-45}
                      textAnchor="end"
                      height={80}
                      tick={{ fontSize: 12 }}
                    />
                    <YAxis 
                      tickFormatter={(value) => `$${value}`}
                      tick={{ fontSize: 12 }}
                    />
                    <Tooltip 
                      formatter={(value: number) => formatCurrency(value)}
                      labelFormatter={(label) => `Fecha: ${label}`}
                    />
                    <Legend />
                    <Line 
                      type="monotone" 
                      dataKey="total" 
                      stroke="#8884d8" 
                      strokeWidth={2}
                      name="Ventas"
                      dot={{ r: 3 }}
                      activeDot={{ r: 5 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="h-300 flex items-center justify-center text-gray-500">
                  No hay datos de evolución de ventas
                </div>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
