import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import api from '../services/api'

// Simple SVG Icons
const CheckCircleIcon = ({ className }: { className?: string }) => (
  <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
)

const XCircleIcon = ({ className }: { className?: string }) => (
  <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
)

const CreditCardIcon = ({ className }: { className?: string }) => (
  <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
  </svg>
)

const SparklesIcon = ({ className }: { className?: string }) => (
  <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
  </svg>
)

const RocketLaunchIcon = ({ className }: { className?: string }) => (
  <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
  </svg>
)

const BuildingOfficeIcon = ({ className }: { className?: string }) => (
  <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
  </svg>
)

interface Plan {
  id: string
  name: string
  price: number
  currency: string
  billing_cycle: 'monthly' | 'yearly'
  description: string
  features: {
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
  highlighted?: boolean
  icon?: React.ComponentType<{ className?: string }>
}

const plans: Plan[] = [
  {
    id: 'basico',
    name: 'Básico',
    price: 5000,
    currency: 'ARS',
    billing_cycle: 'monthly',
    description: 'Perfecto para pequeños negocios que comienzan',
    icon: SparklesIcon,
    features: {
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
  },
  {
    id: 'pro',
    name: 'Pro',
    price: 15000,
    currency: 'ARS',
    billing_cycle: 'monthly',
    description: 'Para negocios en crecimiento que necesitan más',
    icon: RocketLaunchIcon,
    highlighted: true,
    features: {
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
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: 40000,
    currency: 'ARS',
    billing_cycle: 'monthly',
    description: 'Solución completa para empresas grandes',
    icon: BuildingOfficeIcon,
    features: {
      max_products: -1, // Unlimited
      max_orders: -1, // Unlimited
      max_users: -1, // Unlimited
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
  },
]

const featureLabels: Record<keyof Plan['features'], string> = {
  max_products: 'Productos',
  max_orders: 'Órdenes mensuales',
  max_users: 'Usuarios',
  reports: 'Reportes avanzados (XLSX, Email)',
  api_access: 'Acceso a API REST',
  multi_warehouse: 'Multi-almacén',
  advanced_analytics: 'Analítica avanzada',
  integrations: 'Integraciones (MercadoLibre, etc.)',
  priority_support: 'Soporte prioritario',
  custom_reports: 'Reportes personalizados',
  automations: 'Automatizaciones',
  bulk_operations: 'Operaciones masivas',
}

export default function PricingPage() {
  const navigate = useNavigate()
  const [loadingPlan, setLoadingPlan] = useState<string | null>(null)

  const handleSubscribe = async (planId: string) => {
    setLoadingPlan(planId)

    try {
      // Verificar si el usuario está autenticado
      const token = localStorage.getItem('authToken')
      if (!token) {
        toast.error('Debes iniciar sesión para suscribirte')
        navigate('/login')
        return
      }

      // Crear checkout de MercadoPago
      const response = await api.post('/subscriptions/create-checkout', {
        plan_type: planId,
        billing_cycle: 'monthly',
      })

      if (response.data.checkout_url) {
        // Redirigir a MercadoPago
        toast.success('Redirigiendo a MercadoPago...')
        window.location.href = response.data.checkout_url
      } else {
        toast.error('Error al crear el checkout')
      }
    } catch (error) {
      console.error('Error al crear checkout:', error)
      
      const apiError = error as { response?: { status?: number; data?: { error?: string } } }
      
      if (apiError?.response?.status === 401) {
        toast.error('Sesión expirada. Por favor, inicia sesión nuevamente.')
        navigate('/login')
      } else if (apiError?.response?.data?.error) {
        toast.error(apiError.response.data.error)
      } else {
        toast.error('Error al procesar el pago. Intenta nuevamente.')
      }
    } finally {
      setLoadingPlan(null)
    }
  }

  const formatPrice = (price: number, currency: string) => {
    return new Intl.NumberFormat('es-AR', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(price)
  }

  const renderFeatureValue = (value: boolean | number) => {
    if (typeof value === 'boolean') {
      return value ? (
        <CheckCircleIcon className="w-5 h-5 text-green-500 flex-shrink-0" />
      ) : (
        <XCircleIcon className="w-5 h-5 text-gray-300 flex-shrink-0" />
      )
    }

    if (value === -1) {
      return <span className="font-semibold text-indigo-600">Ilimitado</span>
    }

    return <span className="font-semibold text-gray-900">{value.toLocaleString()}</span>
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-purple-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-16">
          <h1 className="text-4xl font-extrabold text-gray-900 sm:text-5xl md:text-6xl">
            Elige el plan perfecto para tu negocio
          </h1>
          <p className="mt-4 text-xl text-gray-600 max-w-3xl mx-auto">
            Gestiona tu inventario, órdenes de venta y compra con Stock In Order. 
            Sin costos ocultos, cancela cuando quieras.
          </p>
          <div className="mt-6 flex justify-center gap-4">
            <button
              onClick={() => navigate('/login')}
              className="px-6 py-2 text-sm font-medium text-indigo-600 bg-white border border-indigo-600 rounded-lg hover:bg-indigo-50 transition"
            >
              Iniciar Sesión
            </button>
            <button
              onClick={() => navigate('/register')}
              className="px-6 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-lg hover:bg-indigo-700 transition"
            >
              Crear Cuenta Gratis
            </button>
          </div>
        </div>

        {/* Pricing Cards */}
        <div className="grid grid-cols-1 gap-8 lg:grid-cols-3">
          {plans.map((plan) => {
            const Icon = plan.icon
            const isLoading = loadingPlan === plan.id

            return (
              <div
                key={plan.id}
                className={`relative rounded-2xl shadow-xl overflow-hidden transition-transform hover:scale-105 ${
                  plan.highlighted
                    ? 'border-4 border-indigo-600 bg-white'
                    : 'border border-gray-200 bg-white'
                }`}
              >
                {/* Highlighted Badge */}
                {plan.highlighted && (
                  <div className="absolute top-0 right-0 bg-indigo-600 text-white px-4 py-1 text-xs font-bold uppercase rounded-bl-lg">
                    Más Popular
                  </div>
                )}

                {/* Card Header */}
                <div className={`p-8 ${plan.highlighted ? 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white' : 'bg-gray-50'}`}>
                  <div className="flex items-center gap-3 mb-4">
                    {Icon && <Icon className="w-8 h-8" />}
                    <h3 className="text-2xl font-bold">{plan.name}</h3>
                  </div>
                  <p className={`text-sm ${plan.highlighted ? 'text-indigo-100' : 'text-gray-600'}`}>
                    {plan.description}
                  </p>
                  <div className="mt-6">
                    <span className="text-5xl font-extrabold">
                      {formatPrice(plan.price, plan.currency)}
                    </span>
                    <span className={`text-lg ${plan.highlighted ? 'text-indigo-200' : 'text-gray-500'}`}>
                      /mes
                    </span>
                  </div>
                </div>

                {/* Features List */}
                <div className="p-8">
                  <ul className="space-y-4">
                    {Object.entries(plan.features).map(([key, value]) => (
                      <li key={key} className="flex items-center gap-3">
                        {renderFeatureValue(value)}
                        <span className="text-sm text-gray-700">
                          {featureLabels[key as keyof Plan['features']]}
                        </span>
                      </li>
                    ))}
                  </ul>

                  {/* Subscribe Button */}
                  <button
                    onClick={() => handleSubscribe(plan.id)}
                    disabled={isLoading}
                    className={`mt-8 w-full flex items-center justify-center gap-2 px-6 py-3 rounded-lg font-semibold transition ${
                      plan.highlighted
                        ? 'bg-indigo-600 text-white hover:bg-indigo-700 disabled:bg-indigo-400'
                        : 'bg-gray-900 text-white hover:bg-gray-800 disabled:bg-gray-400'
                    }`}
                  >
                    {isLoading ? (
                      <>
                        <svg
                          className="animate-spin h-5 w-5 text-white"
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 24 24"
                        >
                          <circle
                            className="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            strokeWidth="4"
                          ></circle>
                          <path
                            className="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                          ></path>
                        </svg>
                        Procesando...
                      </>
                    ) : (
                      <>
                        <CreditCardIcon className="w-5 h-5" />
                        Suscribirme
                      </>
                    )}
                  </button>
                </div>
              </div>
            )
          })}
        </div>

        {/* FAQ Section */}
        <div className="mt-16 bg-white rounded-2xl shadow-lg p-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">Preguntas Frecuentes</h2>
          <div className="space-y-6">
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">¿Puedo cambiar de plan en cualquier momento?</h3>
              <p className="text-gray-600">
                Sí, puedes actualizar o degradar tu plan cuando lo necesites. Los cambios se reflejarán en tu próxima facturación.
              </p>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">¿Qué métodos de pago aceptan?</h3>
              <p className="text-gray-600">
                Aceptamos todos los métodos de pago disponibles en MercadoPago: tarjetas de crédito, débito, efectivo (RapiPago, PagoFácil) y más.
              </p>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">¿Hay costos adicionales?</h3>
              <p className="text-gray-600">
                No, el precio que ves es el precio que pagas. Sin costos ocultos ni cargos adicionales.
              </p>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">¿Puedo cancelar mi suscripción?</h3>
              <p className="text-gray-600">
                Por supuesto. Puedes cancelar tu suscripción en cualquier momento desde tu página de facturación. No hay multas ni penalizaciones.
              </p>
            </div>
          </div>
        </div>

        {/* CTA Section */}
        <div className="mt-16 bg-gradient-to-r from-indigo-600 to-purple-600 rounded-2xl shadow-2xl p-12 text-center text-white">
          <h2 className="text-3xl font-bold mb-4">¿Tienes un negocio grande?</h2>
          <p className="text-xl text-indigo-100 mb-8 max-w-2xl mx-auto">
            Contacta con nuestro equipo de ventas para obtener una solución personalizada con precios especiales.
          </p>
          <button
            onClick={() => window.open('mailto:sales@stockinorder.com', '_blank')}
            className="bg-white text-indigo-600 px-8 py-4 rounded-lg font-semibold hover:bg-indigo-50 transition"
          >
            Contactar Ventas
          </button>
        </div>
      </div>
    </div>
  )
}
