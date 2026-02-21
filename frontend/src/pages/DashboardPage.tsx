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
  BarChart,
  Bar,
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
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-display font-semibold text-neutral-900 mb-2">Dashboard</h1>
        <p className="text-neutral-600 text-sm">Resumen del negocio y métricas clave</p>
      </div>
      
      {loading && (
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <div className="w-10 h-10 border-3 border-neutral-200 border-t-primary-600 rounded-full animate-spin mb-3 mx-auto"></div>
            <p className="text-sm text-neutral-600">Cargando dashboard...</p>
          </div>
        </div>
      )}
      
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-xl text-sm">
          {error}
        </div>
      )}
      
      {!loading && !error && kpis && chartData && (
        <>
          {/* KPIs Section */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <MetricCard 
              title="Total de Productos" 
              value={kpis.total_products} 
            />
            <MetricCard 
              title="Productos Bajo Stock" 
              value={kpis.low_stock_products} 
              variant="danger"
            />
            <MetricCard 
              title="Ventas del Mes" 
              value={formatCurrency(kpis.current_month_sales)} 
              variant="success"
            />
            <MetricCard 
              title="Órdenes Pendientes" 
              value={kpis.pending_sales_orders} 
              variant="warning"
            />
          </div>

          {/* Charts Section - Primera fila */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            {/* Top 5 Productos Más Vendidos */}
            <div className="bg-white rounded-2xl border border-neutral-200 shadow-soft p-6">
              <h3 className="text-lg font-display font-semibold text-neutral-900 mb-6">
                Top 5 Productos Más Vendidos
              </h3>
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
                <div className="h-300 flex items-center justify-center text-neutral-500 text-sm">
                  No hay datos de ventas disponibles
                </div>
              )}
            </div>

            {/* Evolución de Ventas (Últimos 30 días) */}
            <div className="bg-white rounded-2xl border border-neutral-200 shadow-soft p-6">
              <h3 className="text-lg font-display font-semibold text-neutral-900 mb-6">
                Ventas en los Últimos 30 Días
              </h3>
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
                <div className="h-300 flex items-center justify-center text-neutral-500 text-sm">
                  No hay datos de evolución de ventas
                </div>
              )}
            </div>
          </div>

          {/* Segunda fila - Ingresos mensuales y Stock Bajo */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Ingresos Mensuales - Ventas vs Compras */}
            <div className="lg:col-span-2 bg-white rounded-2xl border border-neutral-200 shadow-soft p-6">
              <h3 className="text-lg font-display font-semibold text-neutral-900 mb-6">
                Ingresos Mensuales - Últimos 6 Meses
              </h3>
              {chartData.monthly_revenue.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={chartData.monthly_revenue}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis 
                      dataKey="month" 
                      tick={{ fontSize: 12 }}
                    />
                    <YAxis 
                      tickFormatter={(value) => `$${value}`}
                      tick={{ fontSize: 12 }}
                    />
                    <Tooltip 
                      formatter={(value: number) => formatCurrency(value)}
                    />
                    <Legend />
                    <Bar 
                      dataKey="sales" 
                      fill="#10b981" 
                      name="Ventas"
                      radius={[8, 8, 0, 0]}
                    />
                    <Bar 
                      dataKey="purchases" 
                      fill="#ef4444" 
                      name="Compras"
                      radius={[8, 8, 0, 0]}
                    />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <div className="h-300 flex items-center justify-center text-neutral-500 text-sm">
                  No hay datos de ingresos mensuales
                </div>
              )}
            </div>

            {/* Productos con Stock Bajo */}
            <div className="bg-white rounded-2xl border border-neutral-200 shadow-soft p-6">
              <h3 className="text-lg font-display font-semibold text-neutral-900 mb-4">
                ⚠️ Stock Crítico
              </h3>
              {chartData.low_stock_products.length > 0 ? (
                <div className="space-y-3 max-h-[300px] overflow-y-auto">
                  {chartData.low_stock_products.map((product) => (
                    <div 
                      key={product.product_id}
                      className="p-3 bg-red-50 border border-red-200 rounded-xl"
                    >
                      <div className="font-medium text-sm text-neutral-900 mb-1">
                        {product.product_name}
                      </div>
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-neutral-600">
                          Stock: <span className="font-semibold text-red-600">{product.current_stock}</span>
                        </span>
                        <span className="text-neutral-500">
                          Mín: {product.min_stock}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="flex items-center justify-center h-[200px] text-neutral-500 text-sm">
                  <div className="text-center">
                    <div className="text-4xl mb-2">✅</div>
                    <p>No hay productos con stock bajo</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
