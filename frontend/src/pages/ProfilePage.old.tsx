import { useEffect, useState } from 'react';
import api from '../services/api';
import { toast } from 'react-hot-toast';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

interface UserProfile {
  id: number;
  email: string;
  name: string;
  role: string;
  created_at: string;
}

interface Subscription {
  id: number;
  user_id: number;
  plan_id: string;
  status: string;
  current_period_start?: string;
  current_period_end?: string;
  created_at: string;
}

interface PlanFeatures {
  max_products: number;
  max_orders: number;
  max_users: number;
  advanced_reports: boolean;
  email_notifications: boolean;
  api_access: boolean;
  priority_support: boolean;
  batch_tracking: boolean;
  expiry_management: boolean;
  multi_warehouse: boolean;
  custom_integrations: boolean;
}

interface UsageStats {
  products_count: number;
  orders_this_month: number;
  users_count: number;
}

const PLAN_NAMES: Record<string, string> = {
  plan_free: '🆓 Gratuito',
  plan_basico: '📦 Básico',
  plan_pro: '⭐ Pro',
  plan_enterprise: '🚀 Enterprise',
};

const PLAN_PRICES: Record<string, string> = {
  plan_free: '$0 ARS/mes',
  plan_basico: '$5,000 ARS/mes',
  plan_pro: '$15,000 ARS/mes',
  plan_enterprise: '$40,000 ARS/mes',
};

const ROLE_NAMES: Record<string, string> = {
  admin: '👑 Administrador',
  repositor: '📦 Repositor',
  vendedor: '🤝 Vendedor',
};

const STATUS_BADGES: Record<string, { bg: string; text: string; label: string }> = {
  active: { bg: 'bg-green-100', text: 'text-green-800', label: '✓ Activa' },
  inactive: { bg: 'bg-gray-100', text: 'text-gray-800', label: 'Inactiva' },
  past_due: { bg: 'bg-orange-100', text: 'text-orange-800', label: '⚠ Vencida' },
  canceled: { bg: 'bg-red-100', text: 'text-red-800', label: '✗ Cancelada' },
};

function ProfilePage() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [planFeatures, setPlanFeatures] = useState<PlanFeatures | null>(null);
  const [usageStats, setUsageStats] = useState<UsageStats | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { logout } = useAuth();

  useEffect(() => {
    fetchProfileData();
  }, []);

  const fetchProfileData = async () => {
    try {
      setLoading(true);

      // Obtener información del usuario
      const userResponse = await api.get('/users/me');
      setUser(userResponse.data);

      // Obtener información de suscripción
      const subResponse = await api.get('/subscriptions/status');
      setSubscription(subResponse.data.subscription);
      setPlanFeatures(subResponse.data.plan_features);

      // Intentar obtener estadísticas de uso (opcional)
      try {
        const statsResponse = await api.get('/subscriptions/usage');
        setUsageStats(statsResponse.data);
      } catch (statsError) {
        // Si no hay endpoint de stats, no es crítico
        console.log('Stats endpoint not available:', statsError);
      }
    } catch (error) {
      console.error('Error cargando datos del perfil:', error);
      toast.error('Error al cargar tu perfil');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    logout();
    toast.success('Sesión cerrada exitosamente');
    navigate('/login');
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('es-AR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const renderFeatureIcon = (enabled: boolean) => {
    return enabled ? (
      <span className="text-green-500 text-xl">✓</span>
    ) : (
      <span className="text-gray-400 text-xl">✗</span>
    );
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-indigo-50 via-white to-purple-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-16 w-16 border-b-4 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600 font-medium">Cargando tu perfil...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-indigo-50 via-white to-purple-50">
        <div className="text-center bg-white p-8 rounded-2xl shadow-xl">
          <div className="text-6xl mb-4">😕</div>
          <p className="text-gray-600 mb-6 text-lg">No se pudo cargar tu perfil</p>
          <button
            onClick={() => navigate('/')}
            className="px-6 py-3 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 font-semibold transition-all transform hover:scale-105"
          >
            Volver al Dashboard
          </button>
        </div>
      </div>
    );
  }

  const statusBadge = STATUS_BADGES[subscription?.status || 'inactive'];

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-purple-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-6xl mx-auto">
        {/* Header con Avatar */}
        <div className="bg-white rounded-3xl shadow-2xl overflow-hidden mb-8 border border-gray-100 transform transition-all hover:shadow-3xl animate-fadeInUp">
          <div className="bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 px-8 py-16 text-white relative overflow-hidden">
            {/* Decoración de fondo */}
            <div className="absolute top-0 right-0 w-64 h-64 bg-white/10 rounded-full -mr-32 -mt-32 animate-pulse"></div>
            <div className="absolute bottom-0 left-0 w-48 h-48 bg-white/10 rounded-full -ml-24 -mb-24 animate-pulse animation-delay-2000"></div>
            <div className="absolute top-1/2 left-1/4 w-32 h-32 bg-white/5 rounded-full animate-pulse animation-delay-4000"></div>
            
            <div className="relative flex flex-col md:flex-row items-center justify-between gap-6">
              <div className="text-center md:text-left">
                <div className="inline-block px-4 py-1 bg-white/20 rounded-full text-sm font-medium mb-4 backdrop-blur-sm">
                  {ROLE_NAMES[user.role] || user.role}
                </div>
                <h1 className="text-5xl font-bold mb-3">{user.name || 'Usuario'}</h1>
                <p className="text-indigo-100 text-lg flex items-center gap-2 justify-center md:justify-start">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  {user.email}
                </p>
              </div>
              <div className="flex-shrink-0">
                <div className="w-32 h-32 rounded-full bg-gradient-to-br from-white/30 to-white/10 backdrop-blur-md flex items-center justify-center text-7xl border-4 border-white/30 shadow-2xl">
                  👤
                </div>
              </div>
            </div>
          </div>

          {/* Información del Usuario */}
          <div className="px-8 py-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
              <span className="text-3xl">📋</span>
              Información Personal
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-gradient-to-br from-blue-50 to-blue-100/50 p-5 rounded-2xl border border-blue-200/50 transform transition-all hover:scale-105 hover:shadow-lg">
                <label className="block text-sm font-medium text-blue-600 mb-2 flex items-center gap-1">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  Nombre Completo
                </label>
                <p className="text-xl font-bold text-gray-900">{user.name || 'Sin nombre'}</p>
              </div>
              <div className="bg-gradient-to-br from-purple-50 to-purple-100/50 p-5 rounded-2xl border border-purple-200/50 transform transition-all hover:scale-105 hover:shadow-lg">
                <label className="block text-sm font-medium text-purple-600 mb-2 flex items-center gap-1">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  Correo Electrónico
                </label>
                <p className="text-xl font-bold text-gray-900 break-all">{user.email}</p>
              </div>
              <div className="bg-gradient-to-br from-green-50 to-green-100/50 p-5 rounded-2xl border border-green-200/50 transform transition-all hover:scale-105 hover:shadow-lg">
                <label className="block text-sm font-medium text-green-600 mb-2 flex items-center gap-1">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  Miembro desde
                </label>
                <p className="text-xl font-bold text-gray-900">{formatDate(user.created_at)}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Estadísticas de Uso */}
        {usageStats && planFeatures && (
          <div className="bg-white rounded-3xl shadow-2xl overflow-hidden mb-8 border border-gray-100 transform transition-all hover:shadow-3xl animate-fadeInUp animation-delay-200">
            <div className="px-8 py-6 bg-gradient-to-r from-cyan-50 to-blue-50 border-b border-gray-200">
              <h2 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                <span className="text-3xl">📊</span>
                Uso del Plan
              </h2>
            </div>
            <div className="px-8 py-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Productos */}
                <div className="relative">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg">
                        <span className="text-xl">📦</span>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-600">Productos</p>
                        <p className="text-2xl font-bold text-gray-900">
                          {usageStats.products_count}
                          <span className="text-sm font-normal text-gray-500">
                            {' / '}
                            {planFeatures.max_products === -1 ? '∞' : planFeatures.max_products}
                          </span>
                        </p>
                      </div>
                    </div>
                  </div>
                  {planFeatures.max_products !== -1 && (
                    <div className="relative h-3 bg-gray-200 rounded-full overflow-hidden">
                      <div
                        className="absolute h-full bg-gradient-to-r from-blue-500 to-blue-600 rounded-full transition-all duration-500"
                        style={{
                          width: `${Math.min((usageStats.products_count / planFeatures.max_products) * 100, 100)}%`,
                        }}
                      ></div>
                    </div>
                  )}
                </div>

                {/* Órdenes */}
                <div className="relative">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-lg">
                        <span className="text-xl">📋</span>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-600">Órdenes (mes)</p>
                        <p className="text-2xl font-bold text-gray-900">
                          {usageStats.orders_this_month}
                          <span className="text-sm font-normal text-gray-500">
                            {' / '}
                            {planFeatures.max_orders === -1 ? '∞' : planFeatures.max_orders}
                          </span>
                        </p>
                      </div>
                    </div>
                  </div>
                  {planFeatures.max_orders !== -1 && (
                    <div className="relative h-3 bg-gray-200 rounded-full overflow-hidden">
                      <div
                        className="absolute h-full bg-gradient-to-r from-purple-500 to-purple-600 rounded-full transition-all duration-500"
                        style={{
                          width: `${Math.min((usageStats.orders_this_month / planFeatures.max_orders) * 100, 100)}%`,
                        }}
                      ></div>
                    </div>
                  )}
                </div>

                {/* Usuarios */}
                <div className="relative">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500 to-pink-600 flex items-center justify-center shadow-lg">
                        <span className="text-xl">👥</span>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-600">Usuarios</p>
                        <p className="text-2xl font-bold text-gray-900">
                          {usageStats.users_count}
                          <span className="text-sm font-normal text-gray-500">
                            {' / '}
                            {planFeatures.max_users === -1 ? '∞' : planFeatures.max_users}
                          </span>
                        </p>
                      </div>
                    </div>
                  </div>
                  {planFeatures.max_users !== -1 && (
                    <div className="relative h-3 bg-gray-200 rounded-full overflow-hidden">
                      <div
                        className="absolute h-full bg-gradient-to-r from-pink-500 to-pink-600 rounded-full transition-all duration-500"
                        style={{
                          width: `${Math.min((usageStats.users_count / planFeatures.max_users) * 100, 100)}%`,
                        }}
                      ></div>
                    </div>
                  )}
                </div>
              </div>

              {/* Mensaje de alerta si está cerca del límite */}
              {planFeatures.max_products !== -1 && 
               usageStats.products_count / planFeatures.max_products > 0.8 && (
                <div className="mt-6 bg-amber-50 border-l-4 border-amber-500 p-4 rounded-r-xl">
                  <div className="flex items-start gap-3">
                    <span className="text-2xl">⚠️</span>
                    <div>
                      <p className="text-sm text-amber-800 font-medium">
                        Estás cerca del límite de productos de tu plan actual.
                      </p>
                      <button
                        onClick={() => navigate('/pricing')}
                        className="mt-2 text-sm text-amber-700 font-semibold hover:text-amber-900 underline"
                      >
                        Actualizar plan →
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Información de Suscripción */}
        <div className="bg-white rounded-3xl shadow-2xl overflow-hidden mb-8 border border-gray-100 transform transition-all hover:shadow-3xl animate-fadeInUp animation-delay-300">
          <div className="px-8 py-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
              <span className="text-3xl">💳</span>
              Suscripción Actual
            </h2>
            {subscription ? (
              <div className="space-y-6">
                <div className="bg-gradient-to-br from-indigo-50 to-purple-50 p-6 rounded-2xl border-2 border-indigo-200">
                  <div className="flex flex-col md:flex-row items-center justify-between gap-4">
                    <div>
                      <p className="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600">
                        {PLAN_NAMES[subscription.plan_id] || subscription.plan_id}
                      </p>
                      <p className="text-2xl text-gray-700 mt-2 font-semibold">
                        {PLAN_PRICES[subscription.plan_id]}
                      </p>
                    </div>
                    <span
                      className={`px-5 py-2 rounded-full text-base font-bold ${statusBadge.bg} ${statusBadge.text} shadow-md`}
                    >
                      {statusBadge.label}
                    </span>
                  </div>

                  {subscription.plan_id !== 'plan_free' && subscription.current_period_start && subscription.current_period_end && (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-6">
                      <div className="bg-white p-4 rounded-xl border border-gray-200">
                        <label className="block text-sm font-semibold text-indigo-600 mb-2 flex items-center gap-1">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                          </svg>
                          Inicio del periodo
                        </label>
                        <p className="text-lg font-bold text-gray-900">
                          {formatDate(subscription.current_period_start)}
                        </p>
                      </div>
                      <div className="bg-white p-4 rounded-xl border border-gray-200">
                        <label className="block text-sm font-semibold text-purple-600 mb-2 flex items-center gap-1">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                          </svg>
                          Próxima renovación
                        </label>
                        <p className="text-lg font-bold text-gray-900">
                          {formatDate(subscription.current_period_end)}
                        </p>
                      </div>
                    </div>
                  )}
                </div>

                {subscription.plan_id === 'plan_free' && (
                  <div className="bg-gradient-to-r from-amber-50 to-orange-50 border-l-4 border-amber-500 p-5 rounded-r-2xl">
                    <div className="flex items-start gap-3">
                      <span className="text-2xl">💡</span>
                      <div>
                        <p className="text-sm text-gray-800 font-medium mb-2">
                          <strong>¿Sabías que?</strong> Con un plan pago puedes acceder a más productos, órdenes y características avanzadas.
                        </p>
                        <button
                          onClick={() => navigate('/pricing')}
                          className="px-5 py-2 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl hover:from-indigo-700 hover:to-purple-700 transition-all text-sm font-semibold shadow-lg transform hover:scale-105"
                        >
                          ✨ Ver Planes Premium
                        </button>
                      </div>
                    </div>
                  </div>
                )}

                {subscription.plan_id !== 'plan_free' && (
                  <div className="flex flex-col sm:flex-row gap-3">
                    <button
                      onClick={() => navigate('/pricing')}
                      className="flex-1 px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl hover:from-indigo-700 hover:to-purple-700 transition-all font-semibold shadow-lg transform hover:scale-105"
                    >
                      🔄 Cambiar Plan
                    </button>
                    <button
                      onClick={() => navigate('/billing')}
                      className="flex-1 px-6 py-3 border-2 border-indigo-600 text-indigo-600 rounded-xl hover:bg-indigo-50 transition-all font-semibold transform hover:scale-105"
                    >
                      📄 Ver Facturación
                    </button>
                  </div>
                )}
              </div>
            ) : (
              <div className="text-center py-12 bg-gradient-to-br from-gray-50 to-gray-100 rounded-2xl">
                <div className="text-6xl mb-4">📦</div>
                <p className="text-gray-600 mb-6 text-lg">No tienes una suscripción activa</p>
                <button
                  onClick={() => navigate('/pricing')}
                  className="px-8 py-4 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl hover:from-indigo-700 hover:to-purple-700 font-semibold shadow-lg transform hover:scale-105 transition-all"
                >
                  Ver Planes Disponibles
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Características del Plan */}
        {planFeatures && (
          <div className="bg-white rounded-3xl shadow-2xl overflow-hidden mb-8 border border-gray-100 transform transition-all hover:shadow-3xl">
            <div className="px-8 py-6 bg-gradient-to-r from-indigo-50 to-purple-50 border-b border-gray-200">
              <h2 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                <span className="text-3xl">✨</span>
                Características de tu Plan
              </h2>
            </div>
            <div className="px-8 py-6">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <div className="flex items-center justify-between p-4 bg-gradient-to-br from-indigo-50 to-indigo-100/50 rounded-xl border border-indigo-200 transform transition-all hover:scale-105">
                  <span className="text-gray-700 font-medium">Productos máximos</span>
                  <span className="font-bold text-indigo-600 text-xl">
                    {planFeatures.max_products === -1 ? '∞' : planFeatures.max_products}
                  </span>
                </div>
                <div className="flex items-center justify-between p-4 bg-gradient-to-br from-purple-50 to-purple-100/50 rounded-xl border border-purple-200 transform transition-all hover:scale-105">
                  <span className="text-gray-700 font-medium">Órdenes por mes</span>
                  <span className="font-bold text-purple-600 text-xl">
                    {planFeatures.max_orders === -1 ? '∞' : planFeatures.max_orders}
                  </span>
                </div>
                <div className="flex items-center justify-between p-4 bg-gradient-to-br from-pink-50 to-pink-100/50 rounded-xl border border-pink-200 transform transition-all hover:scale-105">
                  <span className="text-gray-700 font-medium">Usuarios permitidos</span>
                  <span className="font-bold text-pink-600 text-xl">
                    {planFeatures.max_users === -1 ? '∞' : planFeatures.max_users}
                  </span>
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-200">
                  <span className="text-gray-700 font-medium">Reportes avanzados</span>
                  {renderFeatureIcon(planFeatures.advanced_reports)}
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-200">
                  <span className="text-gray-700 font-medium">Notificaciones por email</span>
                  {renderFeatureIcon(planFeatures.email_notifications)}
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-200">
                  <span className="text-gray-700 font-medium">Acceso a API</span>
                  {renderFeatureIcon(planFeatures.api_access)}
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-200">
                  <span className="text-gray-700 font-medium">Soporte prioritario</span>
                  {renderFeatureIcon(planFeatures.priority_support)}
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-200">
                  <span className="text-gray-700 font-medium">Trazabilidad de lotes</span>
                  {renderFeatureIcon(planFeatures.batch_tracking)}
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-200">
                  <span className="text-gray-700 font-medium">Gestión de vencimientos</span>
                  {renderFeatureIcon(planFeatures.expiry_management)}
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-200">
                  <span className="text-gray-700 font-medium">Múltiples almacenes</span>
                  {renderFeatureIcon(planFeatures.multi_warehouse)}
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-200">
                  <span className="text-gray-700 font-medium">Integraciones personalizadas</span>
                  {renderFeatureIcon(planFeatures.custom_integrations)}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Botón de Cerrar Sesión */}
        <div className="bg-white rounded-3xl shadow-2xl overflow-hidden border border-gray-100 transform transition-all hover:shadow-3xl">
          <div className="px-8 py-6">
            <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
              <div>
                <h3 className="text-xl font-bold text-gray-900 flex items-center gap-2">
                  <span className="text-2xl">🚪</span>
                  Cerrar Sesión
                </h3>
                <p className="text-sm text-gray-600 mt-1">Salir de tu cuenta de forma segura</p>
              </div>
              <button
                onClick={handleLogout}
                className="px-8 py-3 bg-gradient-to-r from-red-600 to-red-700 text-white rounded-xl hover:from-red-700 hover:to-red-800 transition-all font-semibold shadow-lg transform hover:scale-105"
              >
                🚪 Cerrar Sesión
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ProfilePage;
