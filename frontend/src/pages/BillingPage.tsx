import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import api from '../services/api'

interface Subscription {
  id: number
  user_id: number
  plan_type: string
  status: string
  start_date: string
  end_date: string | null
  auto_renew: boolean
  mp_subscription_id: string | null
  mp_payment_id: string | null
  mp_preapproval_id: string | null
  created_at: string
  updated_at: string
}

interface PlanFeatures {
  max_products: number
  max_orders: number
  max_users: number
  reports: boolean
  api_access: boolean
  multi_warehouse: boolean
  advanced_analytics: boolean
  integrations: boolean
  priority_support: boolean
  custom_reports: boolean
  automations: boolean
  bulk_operations: boolean
}

const planPrices: Record<string, number> = {
  free: 0,
  basico: 5000,
  pro: 15000,
  enterprise: 40000,
}

const planFeatures: Record<string, PlanFeatures> = {
  free: {
    max_products: 50,
    max_orders: 20,
    max_users: 1,
    reports: false,
    api_access: false,
    multi_warehouse: false,
    advanced_analytics: false,
    integrations: false,
    priority_support: false,
    custom_reports: false,
    automations: false,
    bulk_operations: false,
  },
  basico: {
    max_products: 200,
    max_orders: 100,
    max_users: 3,
    reports: true,
    api_access: false,
    multi_warehouse: false,
    advanced_analytics: false,
    integrations: true,
    priority_support: false,
    custom_reports: false,
    automations: false,
    bulk_operations: true,
  },
  pro: {
    max_products: 1000,
    max_orders: 500,
    max_users: 10,
    reports: true,
    api_access: true,
    multi_warehouse: true,
    advanced_analytics: true,
    integrations: true,
    priority_support: true,
    custom_reports: true,
    automations: true,
    bulk_operations: true,
  },
  enterprise: {
    max_products: -1,
    max_orders: -1,
    max_users: -1,
    reports: true,
    api_access: true,
    multi_warehouse: true,
    advanced_analytics: true,
    integrations: true,
    priority_support: true,
    custom_reports: true,
    automations: true,
    bulk_operations: true,
  },
}

export default function BillingPage() {
  const navigate = useNavigate()
  const [subscription, setSubscription] = useState<Subscription | null>(null)
  const [loading, setLoading] = useState(true)
  const [canceling, setCanceling] = useState(false)

  useEffect(() => {
    fetchSubscription()
    
    // Check if redirected due to payment required (402)
    const paymentRequiredData = sessionStorage.getItem('paymentRequired')
    if (paymentRequiredData) {
      try {
        const data = JSON.parse(paymentRequiredData)
        toast.error(data.message || 'Necesitas una suscripción activa para acceder a esta funcionalidad', {
          duration: 5000,
        })
        sessionStorage.removeItem('paymentRequired')
      } catch {
        // Ignore parsing errors
      }
    }
  }, [])

  const fetchSubscription = async () => {
    try {
      const response = await api.get('/subscriptions/status')
      setSubscription(response.data)
    } catch (error) {
      console.error('Error al obtener suscripción:', error)
      const apiError = error as { response?: { status?: number } }
      if (apiError?.response?.status === 404) {
        toast.error('No tienes una suscripción activa')
      } else {
        toast.error('Error al cargar la suscripción')
      }
    } finally {
      setLoading(false)
    }
  }

  const handleCancelSubscription = async () => {
    if (!subscription) return

    const confirmed = window.confirm(
      '¿Estás seguro de que deseas cancelar tu suscripción? ' +
      'Perderás el acceso a las funcionalidades premium al finalizar el período actual.'
    )

    if (!confirmed) return

    setCanceling(true)

    try {
      await api.post('/subscriptions/cancel', {
        subscription_id: subscription.id,
      })

      toast.success('Suscripción cancelada exitosamente')
      await fetchSubscription() // Refresh data
    } catch (error) {
      console.error('Error al cancelar suscripción:', error)
      const apiError = error as { response?: { data?: { error?: string } } }
      toast.error(apiError?.response?.data?.error || 'Error al cancelar la suscripción')
    } finally {
      setCanceling(false)
    }
  }

  const handleUpgrade = () => {
    navigate('/pricing')
  }

  const getStatusBadge = (status: string) => {
    const badges: Record<string, { bg: string; text: string; label: string }> = {
      active: { bg: 'bg-green-100', text: 'text-green-800', label: 'Activa' },
      cancelled: { bg: 'bg-yellow-100', text: 'text-yellow-800', label: 'Cancelada' },
      expired: { bg: 'bg-red-100', text: 'text-red-800', label: 'Expirada' },
      pending: { bg: 'bg-blue-100', text: 'text-blue-800', label: 'Pendiente' },
    }

    const badge = badges[status] || { bg: 'bg-gray-100', text: 'text-gray-800', label: status }

    return (
      <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${badge.bg} ${badge.text}`}>
        {badge.label}
      </span>
    )
  }

  const formatDate = (dateString: string | null) => {
    if (!dateString) return 'N/A'
    return new Intl.DateTimeFormat('es-AR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    }).format(new Date(dateString))
  }

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('es-AR', {
      style: 'currency',
      currency: 'ARS',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(price)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    )
  }

  if (!subscription) {
    return (
      <div className="max-w-4xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        <div className="bg-white shadow rounded-lg p-8 text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <h3 className="mt-4 text-lg font-medium text-gray-900">No tienes una suscripción activa</h3>
          <p className="mt-2 text-sm text-gray-500">
            Suscríbete a un plan para desbloquear todas las funcionalidades premium.
          </p>
          <button
            onClick={() => navigate('/pricing')}
            className="mt-6 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
          >
            Ver Planes
          </button>
        </div>
      </div>
    )
  }

  const features = planFeatures[subscription.plan_type] || planFeatures.free
  const price = planPrices[subscription.plan_type] || 0

  return (
    <div className="max-w-6xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Facturación y Suscripción</h1>
        <p className="mt-2 text-sm text-gray-600">
          Administra tu suscripción y métodos de pago
        </p>
      </div>

      <div className="grid grid-cols-1 gap-8 lg:grid-cols-3">
        {/* Current Plan Card */}
        <div className="lg:col-span-2">
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <div className="bg-gradient-to-r from-indigo-600 to-purple-600 px-6 py-8 text-white">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-2xl font-bold capitalize">Plan {subscription.plan_type}</h2>
                  <p className="mt-2 text-indigo-100">
                    {formatPrice(price)} / mes
                  </p>
                </div>
                <div className="text-right">
                  {getStatusBadge(subscription.status)}
                </div>
              </div>
            </div>

            <div className="px-6 py-6 space-y-6">
              {/* Subscription Details */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-gray-500">Fecha de inicio</p>
                  <p className="mt-1 text-sm text-gray-900">{formatDate(subscription.start_date)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">Fecha de renovación</p>
                  <p className="mt-1 text-sm text-gray-900">{formatDate(subscription.end_date)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">Renovación automática</p>
                  <p className="mt-1 text-sm text-gray-900">
                    {subscription.auto_renew ? 'Activada' : 'Desactivada'}
                  </p>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500">ID de suscripción</p>
                  <p className="mt-1 text-xs text-gray-600 font-mono">
                    {subscription.mp_preapproval_id || subscription.mp_subscription_id || 'N/A'}
                  </p>
                </div>
              </div>

              {/* Features List */}
              <div className="border-t pt-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Características incluidas</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div className="flex items-center gap-2">
                    <span className="text-indigo-600">✓</span>
                    <span className="text-sm text-gray-700">
                      {features.max_products === -1 ? 'Productos ilimitados' : `${features.max_products} productos`}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-indigo-600">✓</span>
                    <span className="text-sm text-gray-700">
                      {features.max_orders === -1 ? 'Órdenes ilimitadas' : `${features.max_orders} órdenes/mes`}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-indigo-600">✓</span>
                    <span className="text-sm text-gray-700">
                      {features.max_users === -1 ? 'Usuarios ilimitados' : `${features.max_users} usuarios`}
                    </span>
                  </div>
                  {features.reports && (
                    <div className="flex items-center gap-2">
                      <span className="text-indigo-600">✓</span>
                      <span className="text-sm text-gray-700">Reportes avanzados</span>
                    </div>
                  )}
                  {features.api_access && (
                    <div className="flex items-center gap-2">
                      <span className="text-indigo-600">✓</span>
                      <span className="text-sm text-gray-700">Acceso a API</span>
                    </div>
                  )}
                  {features.integrations && (
                    <div className="flex items-center gap-2">
                      <span className="text-indigo-600">✓</span>
                      <span className="text-sm text-gray-700">Integraciones</span>
                    </div>
                  )}
                  {features.priority_support && (
                    <div className="flex items-center gap-2">
                      <span className="text-indigo-600">✓</span>
                      <span className="text-sm text-gray-700">Soporte prioritario</span>
                    </div>
                  )}
                  {features.bulk_operations && (
                    <div className="flex items-center gap-2">
                      <span className="text-indigo-600">✓</span>
                      <span className="text-sm text-gray-700">Operaciones masivas</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Actions */}
              <div className="border-t pt-6 flex gap-4">
                <button
                  onClick={handleUpgrade}
                  className="flex-1 bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition font-medium"
                >
                  Actualizar Plan
                </button>
                {subscription.status === 'active' && (
                  <button
                    onClick={handleCancelSubscription}
                    disabled={canceling}
                    className="flex-1 bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition font-medium disabled:bg-red-400"
                  >
                    {canceling ? 'Cancelando...' : 'Cancelar Suscripción'}
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Payment Method */}
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Método de pago</h3>
            <div className="flex items-center gap-3 p-4 bg-gray-50 rounded-lg">
              <svg className="h-8 w-8 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
                />
              </svg>
              <div>
                <p className="text-sm font-medium text-gray-900">MercadoPago</p>
                <p className="text-xs text-gray-500">Gestiona tus métodos de pago en MercadoPago</p>
              </div>
            </div>
          </div>

          {/* Help Card */}
          <div className="bg-indigo-50 border border-indigo-200 rounded-lg p-6">
            <h3 className="text-lg font-medium text-indigo-900 mb-2">¿Necesitas ayuda?</h3>
            <p className="text-sm text-indigo-700 mb-4">
              Si tienes alguna pregunta sobre tu suscripción o facturación, contáctanos.
            </p>
            <button
              onClick={() => window.open('mailto:support@stockinorder.com', '_blank')}
              className="w-full bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition text-sm font-medium"
            >
              Contactar Soporte
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
