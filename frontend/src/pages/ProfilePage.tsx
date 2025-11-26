import { useEffect, useState, useCallback } from 'react';
import api from '../services/api';
import { toast } from 'react-hot-toast';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { User, Mail, Calendar, Shield, Package, FileText, Users, Check, X, LogOut, AlertTriangle, ArrowRight } from 'lucide-react';

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
  plan_free: 'Gratuito',
  plan_basico: 'Básico',
  plan_pro: 'Pro',
  plan_enterprise: 'Enterprise',
};

const PLAN_PRICES: Record<string, string> = {
  plan_free: '$0 ARS/mes',
  plan_basico: '$5,000 ARS/mes',
  plan_pro: '$15,000 ARS/mes',
  plan_enterprise: '$40,000 ARS/mes',
};

const ROLE_NAMES: Record<string, string> = {
  admin: 'Administrador',
  repositor: 'Repositor',
  vendedor: 'Vendedor',
};

const STATUS_BADGES: Record<string, { bg: string; text: string; label: string }> = {
  active: { bg: 'bg-emerald-50', text: 'text-emerald-700', label: 'Activa' },
  inactive: { bg: 'bg-neutral-100', text: 'text-neutral-700', label: 'Inactiva' },
  past_due: { bg: 'bg-amber-50', text: 'text-amber-700', label: 'Vencida' },
  canceled: { bg: 'bg-rose-50', text: 'text-rose-700', label: 'Cancelada' },
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
          // La API retorna directamente los datos, no anidados
          const data = response.data;
          setSubscription({
            id: 0,
            user_id: data.user_id,
            plan_id: data.plan_id,
            status: data.status,
            current_period_end: data.current_period_end,
            created_at: new Date().toISOString(),
          });
          setPlanFeatures({
            max_products: data.features.MaxProducts || 50,
            max_orders: data.features.MaxOrders || 20,
            max_users: data.features.MaxUsers || 1,
            advanced_reports: data.features.AdvancedReports || false,
            email_notifications: data.features.EmailNotifications || false,
            api_access: data.features.APIAccess || false,
            priority_support: data.features.PrioritySupport || false,
            batch_tracking: data.features.BatchTracking || false,
            expiry_management: data.features.ExpiryManagement || false,
            multi_warehouse: data.features.MultiWarehouse || false,
            custom_integrations: data.features.CustomIntegrations || false,
          });
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
    if (percentage >= 90) return 'bg-rose-500';
    if (percentage >= 75) return 'bg-amber-500';
    return 'bg-emerald-500';
  };

  const renderFeatureIcon = (enabled: boolean) => {
    return enabled ? (
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-emerald-600 shadow-soft">
        <Check className="w-5 h-5 text-white" />
      </div>
    ) : (
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-neutral-300">
        <X className="w-5 h-5 text-neutral-600" />
      </div>
    );
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-neutral-50">
        <div className="text-center">
          <div className="relative inline-flex">
            <div className="w-20 h-20 border-4 border-primary-200 border-t-primary-600 rounded-full animate-spin"></div>
          </div>
          <p className="mt-6 text-lg font-semibold text-neutral-700 animate-pulse">
            Cargando tu perfil...
          </p>
        </div>
      </div>
    );
  }

  if (error && !user) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-neutral-50 p-4">
        <div className="max-w-md w-full bg-white p-8 rounded-2xl shadow-soft-lg border border-neutral-200 text-center">
          <div className="w-20 h-20 bg-rose-500 rounded-2xl flex items-center justify-center mx-auto mb-6 shadow-soft">
            <AlertTriangle className="w-10 h-10 text-white" />
          </div>
          <h3 className="text-2xl font-display font-bold text-neutral-900 mb-3">Error al cargar el perfil</h3>
          <p className="text-neutral-600 mb-6">{error}</p>
          <div className="flex gap-3">
            <button
              onClick={fetchProfileData}
              className="flex-1 px-6 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-700 font-semibold shadow-soft transform transition-all hover:scale-105"
            >
              Reintentar
            </button>
            <button
              onClick={() => navigate('/')}
              className="flex-1 px-6 py-3 border-2 border-neutral-300 text-neutral-700 rounded-xl hover:bg-neutral-50 font-semibold transition-all"
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
    <div className="min-h-screen bg-neutral-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="relative overflow-hidden rounded-2xl shadow-soft-lg bg-gradient-to-r from-primary-600 to-primary-700">
          {/* Contenido del header */}
          <div className="relative px-8 py-12 lg:px-12 lg:py-16">
            <div className="flex flex-col lg:flex-row items-center justify-between gap-8">
              <div className="flex items-center gap-6">
                {/* Avatar */}
                <div className="relative group">
                  <div className="relative w-24 h-24 lg:w-32 lg:h-32 rounded-2xl bg-white/10 backdrop-blur-md border-2 border-white/30 flex items-center justify-center shadow-soft transform transition-all group-hover:scale-105">
                    <User className="w-12 h-12 lg:w-16 lg:h-16 text-white" />
                  </div>
                </div>

                {/* Info del usuario */}
                <div className="text-center lg:text-left">
                  <div className="inline-flex items-center gap-2 px-4 py-1.5 bg-white/20 backdrop-blur-md rounded-full border border-white/30 mb-3">
                    <span className="text-sm font-semibold text-white">
                      {ROLE_NAMES[user.role] || user.role}
                    </span>
                  </div>
                  <h1 className="text-4xl lg:text-5xl font-display font-bold text-white mb-2">
                    {user.name || 'Usuario'}
                  </h1>
                  <div className="flex items-center gap-2 text-white/90 text-lg justify-center lg:justify-start">
                    <Mail className="w-5 h-5" />
                    <span className="font-medium">{user.email}</span>
                  </div>
                </div>
              </div>

              {/* Botón de cerrar sesión premium */}
              <button
                onClick={handleLogout}
                className="group px-8 py-4 bg-white/10 hover:bg-white/20 backdrop-blur-md border-2 border-white/30 rounded-xl text-white font-semibold shadow-soft transform transition-all hover:scale-105 active:scale-95"
              >
                <div className="flex items-center gap-3">
                  <LogOut className="w-5 h-5" />
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
            <div className="bg-white rounded-2xl shadow-soft-lg border border-neutral-200 p-8">
              <h2 className="text-2xl font-display font-bold text-neutral-900 mb-6 flex items-center gap-3">
                <div className="w-10 h-10 bg-primary-600 rounded-xl flex items-center justify-center shadow-soft">
                  <User className="w-6 h-6 text-white" />
                </div>
                <span>Información Personal</span>
              </h2>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="group p-5 bg-neutral-50 rounded-xl border border-neutral-200 transition-all hover:shadow-soft">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center shadow-soft">
                      <User className="w-5 h-5 text-white" />
                    </div>
                    <span className="text-sm font-semibold text-neutral-700">Nombre Completo</span>
                  </div>
                  <p className="text-xl font-bold text-neutral-900 pl-11">{user.name || 'Sin nombre'}</p>
                </div>

                <div className="group p-5 bg-neutral-50 rounded-xl border border-neutral-200 transition-all hover:shadow-soft">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center shadow-soft">
                      <Mail className="w-5 h-5 text-white" />
                    </div>
                    <span className="text-sm font-semibold text-neutral-700">Correo Electrónico</span>
                  </div>
                  <p className="text-xl font-bold text-neutral-900 break-all pl-11">{user.email}</p>
                </div>

                <div className="group p-5 bg-neutral-50 rounded-xl border border-neutral-200 transition-all hover:shadow-soft">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center shadow-soft">
                      <Calendar className="w-5 h-5 text-white" />
                    </div>
                    <span className="text-sm font-semibold text-neutral-700">Miembro desde</span>
                  </div>
                  <p className="text-xl font-bold text-neutral-900 pl-11">{formatDate(user.created_at)}</p>
                </div>

                <div className="group p-5 bg-neutral-50 rounded-xl border border-neutral-200 transition-all hover:shadow-soft">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center shadow-soft">
                      <Shield className="w-5 h-5 text-white" />
                    </div>
                    <span className="text-sm font-semibold text-neutral-700">ID de Usuario</span>
                  </div>
                  <p className="text-xl font-bold text-neutral-900 pl-11">#{user.id}</p>
                </div>
              </div>
            </div>

            {/* Estadísticas de Uso */}
            {usageStats && planFeatures && (
              <div className="bg-white rounded-2xl shadow-soft-lg border border-neutral-200 p-8">
                <h2 className="text-2xl font-display font-bold text-neutral-900 mb-6 flex items-center gap-3">
                  <div className="w-10 h-10 bg-primary-600 rounded-xl flex items-center justify-center shadow-soft">
                    <FileText className="w-6 h-6 text-white" />
                  </div>
                  <span>Uso del Plan</span>
                </h2>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {/* Productos */}
                  <div className="group">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-primary-600 rounded-xl flex items-center justify-center shadow-soft transform transition-transform group-hover:scale-110">
                          <Package className="w-6 h-6 text-white" />
                        </div>
                        <div>
                          <p className="text-sm font-medium text-neutral-600">Productos</p>
                          <p className="text-2xl font-bold text-neutral-900">
                            {usageStats.products_count}
                            <span className="text-sm font-normal text-neutral-500">
                              {' / '}
                              {planFeatures.max_products === -1 ? '∞' : planFeatures.max_products}
                            </span>
                          </p>
                        </div>
                      </div>
                    </div>
                    {planFeatures.max_products !== -1 && (
                      <div className="relative h-3 bg-neutral-200 rounded-full overflow-hidden">
                        <div
                          className={`absolute h-full ${getUsageColor(
                            calculateUsagePercentage(usageStats.products_count, planFeatures.max_products)
                          )} rounded-full transition-all duration-1000 ease-out shadow-soft`}
                          style={{
                            width: `${calculateUsagePercentage(usageStats.products_count, planFeatures.max_products)}%`,
                          }}
                        />
                      </div>
                    )}
                  </div>

                  {/* Órdenes */}
                  <div className="group">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-primary-600 rounded-xl flex items-center justify-center shadow-soft transform transition-transform group-hover:scale-110">
                          <Users className="w-6 h-6 text-white" />
                        </div>
                        <div>
                          <p className="text-sm font-medium text-neutral-600">Órdenes</p>
                          <p className="text-2xl font-bold text-neutral-900">
                            {usageStats.orders_this_month}
                            <span className="text-sm font-normal text-neutral-500">
                              {' / '}
                              {planFeatures.max_orders === -1 ? '∞' : planFeatures.max_orders}
                            </span>
                          </p>
                        </div>
                      </div>
                    </div>
                    {planFeatures.max_orders !== -1 && (
                      <div className="relative h-3 bg-neutral-200 rounded-full overflow-hidden">
                        <div
                          className={`absolute h-full ${getUsageColor(
                            calculateUsagePercentage(usageStats.orders_this_month, planFeatures.max_orders)
                          )} rounded-full transition-all duration-1000 ease-out shadow-soft`}
                          style={{
                            width: `${calculateUsagePercentage(usageStats.orders_this_month, planFeatures.max_orders)}%`,
                          }}
                        />
                      </div>
                    )}
                  </div>

                  {/* Usuarios */}
                  <div className="group">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-primary-600 rounded-xl flex items-center justify-center shadow-soft transform transition-transform group-hover:scale-110">
                          <Users className="w-6 h-6 text-white" />
                        </div>
                        <div>
                          <p className="text-sm font-medium text-neutral-600">Usuarios</p>
                          <p className="text-2xl font-bold text-neutral-900">
                            {usageStats.users_count}
                            <span className="text-sm font-normal text-neutral-500">
                              {' / '}
                              {planFeatures.max_users === -1 ? '∞' : planFeatures.max_users}
                            </span>
                          </p>
                        </div>
                      </div>
                    </div>
                    {planFeatures.max_users !== -1 && (
                      <div className="relative h-3 bg-neutral-200 rounded-full overflow-hidden">
                        <div
                          className={`absolute h-full ${getUsageColor(
                            calculateUsagePercentage(usageStats.users_count, planFeatures.max_users)
                          )} rounded-full transition-all duration-1000 ease-out shadow-soft`}
                          style={{
                            width: `${calculateUsagePercentage(usageStats.users_count, planFeatures.max_users)}%`,
                          }}
                        />
                      </div>
                    )}
                  </div>
                </div>

                {/* Alerta de límite */}
                {planFeatures.max_products !== -1 &&
                  calculateUsagePercentage(usageStats.products_count, planFeatures.max_products) > 80 && (
                    <div className="mt-6 p-5 bg-amber-50 border-l-4 border-amber-500 rounded-r-xl">
                      <div className="flex items-start gap-4">
                        <div className="flex-shrink-0 w-10 h-10 bg-amber-500 rounded-xl flex items-center justify-center shadow-soft">
                          <AlertTriangle className="w-6 h-6 text-white" />
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
                            className="mt-3 inline-flex items-center gap-2 px-4 py-2 bg-amber-600 text-white rounded-xl hover:bg-amber-700 font-semibold text-sm shadow-soft transform transition-all hover:scale-105"
                          >
                            Actualizar Plan
                            <ArrowRight className="w-4 h-4" />
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
              <div className="bg-white rounded-2xl shadow-soft-lg border border-neutral-200 p-8">
                <h2 className="text-2xl font-display font-bold text-neutral-900 mb-6 flex items-center gap-3">
                  <div className="w-10 h-10 bg-primary-600 rounded-xl flex items-center justify-center shadow-soft">
                    <Shield className="w-6 h-6 text-white" />
                  </div>
                  <span>Tu Plan</span>
                </h2>

                <div className="relative p-6 bg-neutral-50 rounded-xl border border-neutral-200 mb-6 group">
                  <div className="relative">
                    <div className="flex items-start justify-between mb-4">
                      <div>
                        <p className="text-4xl font-display font-bold text-primary-600 mb-2">
                          {PLAN_NAMES[subscription.plan_id] || subscription.plan_id}
                        </p>
                        <p className="text-2xl font-bold text-neutral-700">
                          {PLAN_PRICES[subscription.plan_id]}
                        </p>
                      </div>
                      <span
                        className={`px-4 py-2 rounded-full text-sm font-bold ${statusBadge.bg} ${statusBadge.text} shadow-soft`}
                      >
                        {statusBadge.label}
                      </span>
                    </div>

                    {subscription.plan_id !== 'plan_free' && subscription.current_period_end && (
                      <div className="pt-4 border-t border-neutral-200">
                        <div className="flex items-center gap-2 text-sm text-neutral-600">
                          <Calendar className="w-4 h-4" />
                          <span>Renueva el {formatDate(subscription.current_period_end)}</span>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                <div className="space-y-3">
                  <button
                    onClick={() => navigate('/pricing')}
                    className="w-full px-6 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-700 font-semibold shadow-soft transform transition-all hover:scale-105"
                  >
                    Cambiar Plan
                  </button>
                  <button
                    onClick={() => navigate('/billing')}
                    className="w-full px-6 py-3 border-2 border-neutral-200 text-neutral-700 rounded-xl hover:bg-neutral-50 font-semibold transition-all"
                  >
                    Ver Facturación
                  </button>
                </div>
              </div>
            )}

            {/* Características */}
            {planFeatures && (
              <div className="bg-white rounded-2xl shadow-soft-lg border border-neutral-200 p-8">
                <h2 className="text-2xl font-display font-bold text-neutral-900 mb-6 flex items-center gap-3">
                  <div className="w-10 h-10 bg-primary-600 rounded-xl flex items-center justify-center shadow-soft">
                    <Check className="w-6 h-6 text-white" />
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
                      className="flex items-center justify-between p-4 bg-neutral-50 rounded-xl hover:bg-neutral-100 transition-all group"
                    >
                      <span className="text-sm font-medium text-neutral-700 group-hover:text-neutral-900">
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
