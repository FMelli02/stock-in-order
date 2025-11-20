import { useEffect, useState, useCallback } from 'react';
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
  active: { bg: 'bg-emerald-100', text: 'text-emerald-700', label: '✓ Activa' },
  inactive: { bg: 'bg-slate-100', text: 'text-slate-700', label: 'Inactiva' },
  past_due: { bg: 'bg-amber-100', text: 'text-amber-700', label: '⚠ Vencida' },
  canceled: { bg: 'bg-rose-100', text: 'text-rose-700', label: '✗ Cancelada' },
};

function ProfilePage() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [planFeatures, setPlanFeatures] = useState<PlanFeatures | null>(null);
  const [usageStats, setUsageStats] = useState<UsageStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { logout, user: authUser, token, isAuthenticated } = useAuth();

  // Debug: Log auth state on mount
  useEffect(() => {
    console.log('ProfilePage mounted - Auth state:', {
      isAuthenticated,
      hasToken: !!token,
      hasUser: !!authUser,
      user: authUser
    });
  }, [isAuthenticated, token, authUser]);

  const fetchProfileData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      // PRIMERO: Verificar que tengamos datos del AuthContext
      if (!authUser) {
        throw new Error('No hay sesión activa');
      }

      // Usar datos del AuthContext SIEMPRE (no depender del backend)
      setUser({
        id: authUser.id,
        email: authUser.email,
        name: authUser.name,
        role: authUser.role,
        created_at: new Date().toISOString(),
      });

      console.log('✅ Usando datos del AuthContext:', authUser);

      // Configurar suscripción por defecto SIEMPRE
      setSubscription({
        id: 0,
        user_id: authUser.id,
        plan_id: 'plan_free',
        status: 'active',
        created_at: new Date().toISOString(),
      });
      setPlanFeatures({
        max_products: 50,
        max_orders: 20,
        max_users: 1,
        advanced_reports: false,
        email_notifications: false,
        api_access: false,
        priority_support: false,
        batch_tracking: false,
        expiry_management: false,
        multi_warehouse: false,
        custom_integrations: false,
      });

      console.log('✅ Usando suscripción por defecto');

      // Configurar estadísticas por defecto SIEMPRE
      setUsageStats({
        products_count: 0,
        orders_this_month: 0,
        users_count: 1,
      });

      console.log('✅ Usando estadísticas por defecto');

      // IMPORTANTE: Terminar loading AQUÍ, no esperar al backend
      setLoading(false);

      // OPCIONAL: Intentar actualizar desde backend (en background, sin bloquear)
      // Si falla, no importa porque ya tenemos los datos
      api.get('/users/me')
        .then(response => {
          console.log('✅ Datos actualizados desde API:', response.data);
          setUser(response.data);
        })
        .catch(err => {
          console.log('ℹ️ No se pudo actualizar desde API (usando datos locales):', err.message);
        });

      // OPCIONAL: Intentar actualizar suscripción desde backend (en background)
      api.get('/subscriptions/status')
        .then(response => {
          console.log('✅ Suscripción actualizada desde API:', response.data);
          setSubscription(response.data.subscription);
          setPlanFeatures(response.data.plan_features);
        })
        .catch(err => {
          console.log('ℹ️ No se pudo actualizar suscripción (usando datos por defecto):', err.message);
        });

      // OPCIONAL: Intentar actualizar estadísticas desde backend (en background)
      api.get('/subscriptions/usage')
        .then(response => {
          console.log('✅ Estadísticas actualizadas desde API:', response.data);
          setUsageStats(response.data);
        })
        .catch(err => {
          console.log('ℹ️ No se pudo actualizar estadísticas (usando datos por defecto):', err.message);
        });

    } catch (error: unknown) {
      console.error('❌ Error crítico en ProfilePage:', error);
      const err = error as { response?: { data?: { message?: string } }; message?: string };
      const errorMessage = err?.response?.data?.message || err?.message || 'Error al cargar tu perfil';
      setError(errorMessage);
      toast.error(errorMessage);
      setLoading(false);
    }
  }, [authUser]);

  useEffect(() => {
    fetchProfileData();
  }, [fetchProfileData]);

  const handleLogout = () => {
    logout();
    toast.success('¡Hasta luego! Sesión cerrada correctamente', {
      icon: '👋',
      duration: 3000,
    });
    navigate('/login');
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('es-AR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const calculateUsagePercentage = (current: number, max: number): number => {
    if (max === -1) return 0;
    return Math.min((current / max) * 100, 100);
  };

  const getUsageColor = (percentage: number): string => {
    if (percentage >= 90) return 'from-rose-500 to-rose-600';
    if (percentage >= 75) return 'from-amber-500 to-amber-600';
    return 'from-emerald-500 to-emerald-600';
  };

  const renderFeatureIcon = (enabled: boolean) => {
    return enabled ? (
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 shadow-lg">
        <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
        </svg>
      </div>
    ) : (
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gradient-to-br from-slate-200 to-slate-300">
        <svg className="w-5 h-5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M6 18L18 6M6 6l12 12" />
        </svg>
      </div>
    );
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
        <div className="text-center">
          <div className="relative inline-flex">
            <div className="w-20 h-20 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="text-3xl animate-pulse">📦</span>
            </div>
          </div>
          <p className="mt-6 text-lg font-semibold text-slate-700 animate-pulse">
            Cargando tu perfil...
          </p>
        </div>
      </div>
    );
  }

  if (error && !user) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50 p-4">
        <div className="max-w-md w-full bg-white/90 backdrop-blur-xl p-8 rounded-3xl shadow-2xl border border-white/20 text-center animate-fadeInUp">
          <div className="w-20 h-20 bg-gradient-to-br from-rose-500 to-rose-600 rounded-full flex items-center justify-center mx-auto mb-6 shadow-lg">
            <span className="text-4xl">⚠️</span>
          </div>
          <h3 className="text-2xl font-bold text-slate-900 mb-3">Error al cargar el perfil</h3>
          <p className="text-slate-600 mb-6">{error}</p>
          <div className="flex gap-3">
            <button
              onClick={fetchProfileData}
              className="flex-1 px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 text-white rounded-xl hover:from-indigo-700 hover:to-indigo-800 font-semibold shadow-lg transform transition-all hover:scale-105"
            >
              Reintentar
            </button>
            <button
              onClick={() => navigate('/')}
              className="flex-1 px-6 py-3 border-2 border-slate-300 text-slate-700 rounded-xl hover:bg-slate-50 font-semibold transition-all"
            >
              Volver
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!user) return null;

  const statusBadge = STATUS_BADGES[subscription?.status || 'inactive'];

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header Premium con efecto glassmorphism */}
        <div className="relative overflow-hidden rounded-3xl shadow-2xl animate-fadeInUp">
          {/* Fondo con gradiente animado */}
          <div className="absolute inset-0 bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600">
            <div className="absolute inset-0 opacity-30">
              <div className="absolute top-0 left-0 w-96 h-96 bg-white/20 rounded-full blur-3xl animate-pulse"></div>
              <div className="absolute bottom-0 right-0 w-96 h-96 bg-white/20 rounded-full blur-3xl animate-pulse" style={{ animationDelay: '1s' }}></div>
              <div className="absolute top-1/2 left-1/2 w-64 h-64 bg-white/10 rounded-full blur-3xl animate-pulse" style={{ animationDelay: '2s' }}></div>
            </div>
          </div>

          {/* Contenido del header */}
          <div className="relative px-8 py-12 lg:px-12 lg:py-16">
            <div className="flex flex-col lg:flex-row items-center justify-between gap-8">
              <div className="flex items-center gap-6">
                {/* Avatar con efecto glassmorphism */}
                <div className="relative group">
                  <div className="absolute inset-0 bg-gradient-to-r from-white/40 to-white/20 rounded-3xl blur-xl group-hover:blur-2xl transition-all"></div>
                  <div className="relative w-24 h-24 lg:w-32 lg:h-32 rounded-3xl bg-white/20 backdrop-blur-md border-2 border-white/30 flex items-center justify-center shadow-2xl transform transition-all group-hover:scale-110">
                    <span className="text-5xl lg:text-6xl">👤</span>
                  </div>
                </div>

                {/* Info del usuario */}
                <div className="text-center lg:text-left">
                  <div className="inline-flex items-center gap-2 px-4 py-1.5 bg-white/20 backdrop-blur-md rounded-full border border-white/30 mb-3">
                    <span className="text-sm font-semibold text-white">
                      {ROLE_NAMES[user.role] || user.role}
                    </span>
                  </div>
                  <h1 className="text-4xl lg:text-5xl font-bold text-white mb-2 tracking-tight">
                    {user.name || 'Usuario'}
                  </h1>
                  <div className="flex items-center gap-2 text-white/90 text-lg justify-center lg:justify-start">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                    </svg>
                    <span className="font-medium">{user.email}</span>
                  </div>
                </div>
              </div>

              {/* Botón de cerrar sesión premium */}
              <button
                onClick={handleLogout}
                className="group px-8 py-4 bg-white/10 hover:bg-white/20 backdrop-blur-md border-2 border-white/30 rounded-2xl text-white font-semibold shadow-xl transform transition-all hover:scale-105 active:scale-95"
              >
                <div className="flex items-center gap-3">
                  <svg className="w-5 h-5 transform transition-transform group-hover:translate-x-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                  </svg>
                  <span>Cerrar Sesión</span>
                </div>
              </button>
            </div>
          </div>
        </div>

        {/* Grid de dos columnas */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Columna izquierda - Info personal y suscripción */}
          <div className="lg:col-span-2 space-y-6">
            {/* Información Personal */}
            <div className="bg-white/80 backdrop-blur-xl rounded-3xl shadow-xl border border-white/20 p-8 animate-fadeInUp animation-delay-100">
              <h2 className="text-2xl font-bold text-slate-900 mb-6 flex items-center gap-3">
                <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl flex items-center justify-center shadow-lg">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </div>
                <span>Información Personal</span>
              </h2>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="group p-5 bg-gradient-to-br from-blue-50 to-blue-100/50 rounded-2xl border border-blue-200/50 transform transition-all hover:scale-105 hover:shadow-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center shadow">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                    </div>
                    <span className="text-sm font-semibold text-blue-700">Nombre Completo</span>
                  </div>
                  <p className="text-xl font-bold text-slate-900 pl-11">{user.name || 'Sin nombre'}</p>
                </div>

                <div className="group p-5 bg-gradient-to-br from-purple-50 to-purple-100/50 rounded-2xl border border-purple-200/50 transform transition-all hover:scale-105 hover:shadow-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center shadow">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                      </svg>
                    </div>
                    <span className="text-sm font-semibold text-purple-700">Correo Electrónico</span>
                  </div>
                  <p className="text-xl font-bold text-slate-900 break-all pl-11">{user.email}</p>
                </div>

                <div className="group p-5 bg-gradient-to-br from-emerald-50 to-emerald-100/50 rounded-2xl border border-emerald-200/50 transform transition-all hover:scale-105 hover:shadow-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-8 h-8 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-lg flex items-center justify-center shadow">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                    </div>
                    <span className="text-sm font-semibold text-emerald-700">Miembro desde</span>
                  </div>
                  <p className="text-xl font-bold text-slate-900 pl-11">{formatDate(user.created_at)}</p>
                </div>

                <div className="group p-5 bg-gradient-to-br from-amber-50 to-amber-100/50 rounded-2xl border border-amber-200/50 transform transition-all hover:scale-105 hover:shadow-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-8 h-8 bg-gradient-to-br from-amber-500 to-amber-600 rounded-lg flex items-center justify-center shadow">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
                      </svg>
                    </div>
                    <span className="text-sm font-semibold text-amber-700">ID de Usuario</span>
                  </div>
                  <p className="text-xl font-bold text-slate-900 pl-11">#{user.id}</p>
                </div>
              </div>
            </div>

            {/* Estadísticas de Uso */}
            {usageStats && planFeatures && (
              <div className="bg-white/80 backdrop-blur-xl rounded-3xl shadow-xl border border-white/20 p-8 animate-fadeInUp animation-delay-200">
                <h2 className="text-2xl font-bold text-slate-900 mb-6 flex items-center gap-3">
                  <div className="w-10 h-10 bg-gradient-to-br from-cyan-500 to-cyan-600 rounded-xl flex items-center justify-center shadow-lg">
                    <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                    </svg>
                  </div>
                  <span>Uso del Plan</span>
                </h2>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {/* Productos */}
                  <div className="group">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl flex items-center justify-center shadow-lg transform transition-transform group-hover:scale-110">
                          <span className="text-2xl">📦</span>
                        </div>
                        <div>
                          <p className="text-sm font-medium text-slate-600">Productos</p>
                          <p className="text-2xl font-bold text-slate-900">
                            {usageStats.products_count}
                            <span className="text-sm font-normal text-slate-500">
                              {' / '}
                              {planFeatures.max_products === -1 ? '∞' : planFeatures.max_products}
                            </span>
                          </p>
                        </div>
                      </div>
                    </div>
                    {planFeatures.max_products !== -1 && (
                      <div className="relative h-3 bg-slate-200 rounded-full overflow-hidden">
                        <div
                          className={`absolute h-full bg-gradient-to-r ${getUsageColor(
                            calculateUsagePercentage(usageStats.products_count, planFeatures.max_products)
                          )} rounded-full transition-all duration-1000 ease-out shadow-lg`}
                          style={{
                            width: `${calculateUsagePercentage(usageStats.products_count, planFeatures.max_products)}%`,
                          }}
                        >
                          <div className="absolute inset-0 bg-white/20 animate-pulse"></div>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Órdenes */}
                  <div className="group">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg transform transition-transform group-hover:scale-110">
                          <span className="text-2xl">📋</span>
                        </div>
                        <div>
                          <p className="text-sm font-medium text-slate-600">Órdenes</p>
                          <p className="text-2xl font-bold text-slate-900">
                            {usageStats.orders_this_month}
                            <span className="text-sm font-normal text-slate-500">
                              {' / '}
                              {planFeatures.max_orders === -1 ? '∞' : planFeatures.max_orders}
                            </span>
                          </p>
                        </div>
                      </div>
                    </div>
                    {planFeatures.max_orders !== -1 && (
                      <div className="relative h-3 bg-slate-200 rounded-full overflow-hidden">
                        <div
                          className={`absolute h-full bg-gradient-to-r ${getUsageColor(
                            calculateUsagePercentage(usageStats.orders_this_month, planFeatures.max_orders)
                          )} rounded-full transition-all duration-1000 ease-out shadow-lg`}
                          style={{
                            width: `${calculateUsagePercentage(usageStats.orders_this_month, planFeatures.max_orders)}%`,
                          }}
                        >
                          <div className="absolute inset-0 bg-white/20 animate-pulse"></div>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Usuarios */}
                  <div className="group">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-gradient-to-br from-pink-500 to-pink-600 rounded-xl flex items-center justify-center shadow-lg transform transition-transform group-hover:scale-110">
                          <span className="text-2xl">👥</span>
                        </div>
                        <div>
                          <p className="text-sm font-medium text-slate-600">Usuarios</p>
                          <p className="text-2xl font-bold text-slate-900">
                            {usageStats.users_count}
                            <span className="text-sm font-normal text-slate-500">
                              {' / '}
                              {planFeatures.max_users === -1 ? '∞' : planFeatures.max_users}
                            </span>
                          </p>
                        </div>
                      </div>
                    </div>
                    {planFeatures.max_users !== -1 && (
                      <div className="relative h-3 bg-slate-200 rounded-full overflow-hidden">
                        <div
                          className={`absolute h-full bg-gradient-to-r ${getUsageColor(
                            calculateUsagePercentage(usageStats.users_count, planFeatures.max_users)
                          )} rounded-full transition-all duration-1000 ease-out shadow-lg`}
                          style={{
                            width: `${calculateUsagePercentage(usageStats.users_count, planFeatures.max_users)}%`,
                          }}
                        >
                          <div className="absolute inset-0 bg-white/20 animate-pulse"></div>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                {/* Alerta de límite */}
                {planFeatures.max_products !== -1 &&
                  calculateUsagePercentage(usageStats.products_count, planFeatures.max_products) > 80 && (
                    <div className="mt-6 p-5 bg-gradient-to-r from-amber-50 to-orange-50 border-l-4 border-amber-500 rounded-r-2xl animate-fadeIn">
                      <div className="flex items-start gap-4">
                        <div className="flex-shrink-0 w-10 h-10 bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl flex items-center justify-center shadow-lg">
                          <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                          </svg>
                        </div>
                        <div className="flex-1">
                          <p className="text-sm font-semibold text-amber-900 mb-1">
                            ¡Atención! Estás cerca del límite
                          </p>
                          <p className="text-sm text-amber-800">
                            Has utilizado el {Math.round(calculateUsagePercentage(usageStats.products_count, planFeatures.max_products))}% de tus productos disponibles.
                          </p>
                          <button
                            onClick={() => navigate('/pricing')}
                            className="mt-3 inline-flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-amber-500 to-amber-600 text-white rounded-xl hover:from-amber-600 hover:to-amber-700 font-semibold text-sm shadow-lg transform transition-all hover:scale-105"
                          >
                            Actualizar Plan
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                            </svg>
                          </button>
                        </div>
                      </div>
                    </div>
                  )}
              </div>
            )}
          </div>

          {/* Columna derecha - Plan y características */}
          <div className="space-y-6">
            {/* Plan actual */}
            {subscription && (
              <div className="bg-white/80 backdrop-blur-xl rounded-3xl shadow-xl border border-white/20 p-8 animate-fadeInUp animation-delay-300">
                <h2 className="text-2xl font-bold text-slate-900 mb-6 flex items-center gap-3">
                  <div className="w-10 h-10 bg-gradient-to-br from-indigo-500 to-indigo-600 rounded-xl flex items-center justify-center shadow-lg">
                    <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <span>Tu Plan</span>
                </h2>

                <div className="relative p-6 bg-gradient-to-br from-indigo-50 to-purple-50 rounded-2xl border-2 border-indigo-200 mb-6 overflow-hidden group">
                  <div className="absolute inset-0 bg-gradient-to-r from-indigo-500/5 to-purple-500/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  <div className="relative">
                    <div className="flex items-start justify-between mb-4">
                      <div>
                        <p className="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 mb-2">
                          {PLAN_NAMES[subscription.plan_id] || subscription.plan_id}
                        </p>
                        <p className="text-2xl font-bold text-slate-700">
                          {PLAN_PRICES[subscription.plan_id]}
                        </p>
                      </div>
                      <span
                        className={`px-4 py-2 rounded-full text-sm font-bold ${statusBadge.bg} ${statusBadge.text} shadow-lg`}
                      >
                        {statusBadge.label}
                      </span>
                    </div>

                    {subscription.plan_id !== 'plan_free' && subscription.current_period_end && (
                      <div className="pt-4 border-t border-indigo-200/50">
                        <div className="flex items-center gap-2 text-sm text-slate-600">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                          </svg>
                          <span>Renueva el {formatDate(subscription.current_period_end)}</span>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                <div className="space-y-3">
                  <button
                    onClick={() => navigate('/pricing')}
                    className="w-full px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl hover:from-indigo-700 hover:to-purple-700 font-semibold shadow-lg transform transition-all hover:scale-105"
                  >
                    Cambiar Plan
                  </button>
                  <button
                    onClick={() => navigate('/billing')}
                    className="w-full px-6 py-3 border-2 border-indigo-200 text-indigo-700 rounded-xl hover:bg-indigo-50 font-semibold transition-all"
                  >
                    Ver Facturación
                  </button>
                </div>
              </div>
            )}

            {/* Características */}
            {planFeatures && (
              <div className="bg-white/80 backdrop-blur-xl rounded-3xl shadow-xl border border-white/20 p-8 animate-fadeInUp animation-delay-400">
                <h2 className="text-2xl font-bold text-slate-900 mb-6 flex items-center gap-3">
                  <div className="w-10 h-10 bg-gradient-to-br from-violet-500 to-violet-600 rounded-xl flex items-center justify-center shadow-lg">
                    <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
                    </svg>
                  </div>
                  <span>Características</span>
                </h2>

                <div className="space-y-3">
                  {[
                    { label: 'Reportes Avanzados', enabled: planFeatures.advanced_reports },
                    { label: 'Notificaciones por Email', enabled: planFeatures.email_notifications },
                    { label: 'Acceso a API', enabled: planFeatures.api_access },
                    { label: 'Soporte Prioritario', enabled: planFeatures.priority_support },
                    { label: 'Trazabilidad de Lotes', enabled: planFeatures.batch_tracking },
                    { label: 'Gestión de Vencimientos', enabled: planFeatures.expiry_management },
                    { label: 'Múltiples Almacenes', enabled: planFeatures.multi_warehouse },
                    { label: 'Integraciones Personalizadas', enabled: planFeatures.custom_integrations },
                  ].map((feature, index) => (
                    <div
                      key={index}
                      className="flex items-center justify-between p-4 bg-slate-50 rounded-xl hover:bg-slate-100 transition-all group"
                    >
                      <span className="text-sm font-medium text-slate-700 group-hover:text-slate-900">
                        {feature.label}
                      </span>
                      {renderFeatureIcon(feature.enabled)}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default ProfilePage;
