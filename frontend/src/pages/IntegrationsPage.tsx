import { useState, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { isAxiosError } from "axios";
import toast from "react-hot-toast";
import api from "../services/api";

interface Integration {
  id: number;
  platform: string;
  external_user_id?: string;
  expires_at: string;
  is_expired: boolean;
  created_at: string;
}

export default function IntegrationsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState<number | null>(null);
  const [showShopifyModal, setShowShopifyModal] = useState(false);
  const [showWooModal, setShowWooModal] = useState(false);

  // Verificar si venimos del callback de OAuth
  useEffect(() => {
    const success = searchParams.get("success");
    const error = searchParams.get("error");

    if (success === "true") {
      toast.success("¡Conexión exitosa! Tu cuenta de Mercado Libre ha sido conectada.");
      setSearchParams({});
    } else if (success === "false" && error) {
      let errorMessage = "Hubo un problema al conectar con Mercado Libre.";
      
      switch (error) {
        case "denied":
          errorMessage = "Rechazaste la autorización. Intenta nuevamente si cambias de opinión.";
          break;
        case "invalid_params":
          errorMessage = "Parámetros inválidos en el callback.";
          break;
        case "invalid_state":
          errorMessage = "Estado inválido. Por favor, intenta nuevamente.";
          break;
        case "token_exchange_failed":
          errorMessage = "No se pudieron obtener los tokens de acceso. Intenta nuevamente.";
          break;
        case "database_error":
          errorMessage = "Error al guardar la integración. Contacta a soporte.";
          break;
      }

      toast.error(errorMessage);
      setSearchParams({});
    }
  }, [searchParams, setSearchParams]);

  // Cargar integraciones del usuario
  useEffect(() => {
    fetchIntegrations();
  }, []);

  const fetchIntegrations = async () => {
    try {
      const response = await api.get<Integration[]>("/integrations");
      setIntegrations(response.data || []);
    } catch (err: unknown) {
      let message = "Error al cargar integraciones";
      if (isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined;
        message = data?.error || err.message;
      } else if (err instanceof Error) {
        message = err.message;
      }
      toast.error(message);
    } finally {
      setLoading(false);
    }
  };

  const handleConnectMercadoLibre = () => {
    // Obtener el token del localStorage
    const token = localStorage.getItem("token");
    if (!token) {
      toast.error("No estás autenticado");
      return;
    }
    
    // Redirigir al endpoint del backend que inicia el flujo OAuth, pasando el token
    window.location.href = `http://localhost:8080/api/v1/integrations/mercadolibre/connect?token=${token}`;
  };

  const handleDeleteIntegration = async (platform: string, integrationId: number) => {
    if (!confirm(`¿Estás seguro de que deseas desconectar ${getPlatformName(platform)}?`)) {
      return;
    }

    setDeleting(integrationId);
    try {
      await api.delete(`/integrations/${platform}`);
      toast.success(`La integración con ${getPlatformName(platform)} ha sido eliminada.`);
      fetchIntegrations();
    } catch (err: unknown) {
      let message = "Error al eliminar la integración";
      if (isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined;
        message = data?.error || err.message;
      } else if (err instanceof Error) {
        message = err.message;
      }
      toast.error(message);
    } finally {
      setDeleting(null);
    }
  };

  const getPlatformName = (platform: string) => {
    const names: Record<string, string> = {
      mercadolibre: "Mercado Libre",
      shopify: "Shopify",
      woocommerce: "WooCommerce",
    };
    return names[platform] || platform;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  const mercadolibreIntegration = integrations.find(
    (i) => i.platform === "mercadolibre"
  );

  const handleConnectShopify = async (shopName: string, apiKey: string, password: string) => {
    try {
      await api.post("/integrations/shopify/connect", {
        shop_name: shopName,
        api_key: apiKey,
        password: password,
      });
      toast.success("¡Shopify conectado exitosamente!");
      setShowShopifyModal(false);
      fetchIntegrations();
    } catch (err: unknown) {
      let message = "Error al conectar Shopify";
      if (isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined;
        message = data?.error || err.message;
      }
      toast.error(message);
    }
  };

  const handleConnectWooCommerce = async (siteUrl: string, consumerKey: string, consumerSecret: string) => {
    try {
      await api.post("/integrations/woocommerce/connect", {
        site_url: siteUrl,
        consumer_key: consumerKey,
        consumer_secret: consumerSecret,
      });
      toast.success("¡WooCommerce conectado exitosamente!");
      setShowWooModal(false);
      fetchIntegrations();
    } catch (err: unknown) {
      let message = "Error al conectar WooCommerce";
      if (isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined;
        message = data?.error || err.message;
      }
      toast.error(message);
    }
  };

  return (
    <>
      {/* Modales */}
      {showShopifyModal && (
        <ShopifyModal onClose={() => setShowShopifyModal(false)} onConnect={handleConnectShopify} />
      )}
      {showWooModal && (
        <WooCommerceModal onClose={() => setShowWooModal(false)} onConnect={handleConnectWooCommerce} />
      )}

      <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-display font-bold text-neutral-900">Integraciones</h1>
        <p className="text-neutral-600 mt-2">
          Conecta tu cuenta con plataformas de venta para sincronizar productos e inventario.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {/* Mercado Libre Card */}
        <div className="bg-white rounded-2xl shadow-soft overflow-hidden border border-neutral-200 hover:shadow-soft-lg transition-shadow">
          {/* Header con icono */}
          <div className="bg-gradient-to-r from-yellow-400 to-yellow-500 p-6">
            <div className="flex items-center gap-3 text-white">
              <svg
                className="h-10 w-10"
                fill="currentColor"
                viewBox="0 0 24 24"
              >
                <path d="M12 2L2 7v10c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V7l-10-5zm0 18c-3.87-.96-7-5.3-7-10V8.3l7-3.5 7 3.5V10c0 4.7-3.13 9.04-7 10z" />
              </svg>
              <div>
                <h3 className="text-xl font-bold">Mercado Libre</h3>
                <p className="text-sm text-yellow-100">Marketplace #1 de LATAM</p>
              </div>
            </div>
          </div>

          {/* Content */}
          <div className="p-6">
            {mercadolibreIntegration ? (
              <div className="space-y-4">
                {/* Estado */}
                <div className="flex items-center gap-2">
                  <span
                    className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold ${
                      mercadolibreIntegration.is_expired
                        ? "bg-red-100 text-red-800"
                        : "bg-green-100 text-green-800"
                    }`}
                  >
                    {mercadolibreIntegration.is_expired ? "⚠️ Token Expirado" : "✅ Conectado"}
                  </span>
                </div>

                {/* Info */}
                {mercadolibreIntegration.external_user_id && (
                  <div className="text-sm">
                    <span className="text-gray-600">ID Usuario:</span>{" "}
                    <span className="font-mono font-medium">
                      {mercadolibreIntegration.external_user_id}
                    </span>
                  </div>
                )}

                <div className="text-sm">
                  <span className="text-gray-600">Conectado:</span>{" "}
                  <span className="font-medium">
                    {new Date(mercadolibreIntegration.created_at).toLocaleDateString()}
                  </span>
                </div>

                <div className="text-sm">
                  <span className="text-gray-600">Expira:</span>{" "}
                  <span className="font-medium">
                    {new Date(mercadolibreIntegration.expires_at).toLocaleDateString()}
                  </span>
                </div>

                {/* Botones */}
                <div className="flex gap-2 pt-4">
                  {mercadolibreIntegration.is_expired && (
                  <button
                    onClick={handleConnectMercadoLibre}
                    className="flex-1 bg-primary-600 text-white px-4 py-2 rounded-xl hover:bg-primary-700 transition-colors font-medium flex items-center justify-center gap-2 shadow-soft"
                    >
                      <svg
                        className="h-4 w-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                        />
                      </svg>
                      Reconectar
                    </button>
                  )}
                  
                  <button
                    onClick={() =>
                      handleDeleteIntegration(
                        "mercadolibre",
                        mercadolibreIntegration.id
                      )
                    }
                    disabled={deleting === mercadolibreIntegration.id}
                    className={`${
                      mercadolibreIntegration.is_expired ? "" : "flex-1"
                    } bg-red-600 text-white px-4 py-2 rounded-xl hover:bg-red-700 transition-colors font-medium flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed shadow-soft`}
                  >
                    {deleting === mercadolibreIntegration.id ? (
                      <>
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                        Desconectando...
                      </>
                    ) : (
                      <>
                        <svg
                          className="h-4 w-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                        Desconectar
                      </>
                    )}
                  </button>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <p className="text-sm text-gray-600">
                  Conecta tu cuenta de Mercado Libre para sincronizar tus publicaciones
                  y gestionar el inventario automáticamente.
                </p>
                
                <button
                  onClick={handleConnectMercadoLibre}
                  className="w-full bg-primary-600 text-white px-4 py-3 rounded-xl hover:bg-primary-700 transition-colors font-medium flex items-center justify-center gap-2 shadow-soft"
                >
                  <svg
                    className="h-5 w-5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                    />
                  </svg>
                  Conectar con Mercado Libre
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Shopify */}
        <ShopifyCard 
          integrations={integrations}
          deleting={deleting}
          onConnect={() => setShowShopifyModal(true)}
          onDelete={handleDeleteIntegration}
        />

        {/* WooCommerce */}
        <WooCommerceCard 
          integrations={integrations}
          deleting={deleting}
          onConnect={() => setShowWooModal(true)}
          onDelete={handleDeleteIntegration}
        />
      </div>
    </div>
    </>
  );
}

// Componente para tarjeta de Shopify
function ShopifyCard({ integrations, deleting, onConnect, onDelete }: {
  integrations: Integration[];
  deleting: number | null;
  onConnect: () => void;
  onDelete: (platform: string, id: number) => void;
}) {
  const shopifyIntegration = integrations.find((i) => i.platform === "shopify");

  return (
    <div className="bg-white rounded-2xl shadow-soft overflow-hidden border border-neutral-200 hover:shadow-soft-lg transition-shadow">
      <div className="bg-gradient-to-r from-green-500 to-green-600 p-6">
        <div className="flex items-center gap-3 text-white">
          <svg className="h-10 w-10" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2L2 7v10c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V7l-10-5z" />
          </svg>
          <div>
            <h3 className="text-xl font-bold">Shopify</h3>
            <p className="text-sm text-green-100">E-commerce líder mundial</p>
          </div>
        </div>
      </div>

      <div className="p-6">
        {shopifyIntegration ? (
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-green-100 text-green-800">
                ✅ Conectado
              </span>
            </div>

            {shopifyIntegration.external_user_id && (
              <div className="text-sm">
                <span className="text-gray-600">Tienda:</span>{" "}
                <span className="font-medium">{shopifyIntegration.external_user_id}</span>
              </div>
            )}

            <div className="text-sm">
              <span className="text-gray-600">Conectado:</span>{" "}
              <span className="font-medium">
                {new Date(shopifyIntegration.created_at).toLocaleDateString()}
              </span>
            </div>

            <button
              onClick={() => onDelete("shopify", shopifyIntegration.id)}
              disabled={deleting === shopifyIntegration.id}
              className="w-full bg-red-600 text-white px-4 py-2 rounded-xl hover:bg-red-700 transition-colors font-medium flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed shadow-soft"
            >
              {deleting === shopifyIntegration.id ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                  Desconectando...
                </>
              ) : (
                <>
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                  Desconectar
                </>
              )}
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              Conecta tu tienda Shopify para sincronizar productos e inventario automáticamente.
            </p>
            <button
              onClick={onConnect}
              className="w-full bg-primary-600 text-white px-4 py-3 rounded-xl hover:bg-primary-700 transition-colors font-medium flex items-center justify-center gap-2 shadow-soft"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
              Conectar con Shopify
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

// Componente para tarjeta de WooCommerce
function WooCommerceCard({ integrations, deleting, onConnect, onDelete }: {
  integrations: Integration[];
  deleting: number | null;
  onConnect: () => void;
  onDelete: (platform: string, id: number) => void;
}) {
  const wooIntegration = integrations.find((i) => i.platform === "woocommerce");

  return (
    <div className="bg-white rounded-2xl shadow-soft overflow-hidden border border-neutral-200 hover:shadow-soft-lg transition-shadow">
      <div className="bg-gradient-to-r from-purple-500 to-purple-600 p-6">
        <div className="flex items-center gap-3 text-white">
          <svg className="h-10 w-10" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2L2 7v10c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V7l-10-5z" />
          </svg>
          <div>
            <h3 className="text-xl font-bold">WooCommerce</h3>
            <p className="text-sm text-purple-100">Plugin WordPress #1</p>
          </div>
        </div>
      </div>

      <div className="p-6">
        {wooIntegration ? (
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-purple-100 text-purple-800">
                ✅ Conectado
              </span>
            </div>

            {wooIntegration.external_user_id && (
              <div className="text-sm">
                <span className="text-gray-600">Sitio:</span>{" "}
                <span className="font-medium">{wooIntegration.external_user_id}</span>
              </div>
            )}

            <div className="text-sm">
              <span className="text-gray-600">Conectado:</span>{" "}
              <span className="font-medium">
                {new Date(wooIntegration.created_at).toLocaleDateString()}
              </span>
            </div>

            <button
              onClick={() => onDelete("woocommerce", wooIntegration.id)}
              disabled={deleting === wooIntegration.id}
              className="w-full bg-red-600 text-white px-4 py-2 rounded-xl hover:bg-red-700 transition-colors font-medium flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed shadow-soft"
            >
              {deleting === wooIntegration.id ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                  Desconectando...
                </>
              ) : (
                <>
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                  Desconectar
                </>
              )}
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              Conecta tu tienda WooCommerce para sincronizar productos e inventario automáticamente.
            </p>
            <button
              onClick={onConnect}
              className="w-full bg-primary-600 text-white px-4 py-3 rounded-xl hover:bg-primary-700 transition-colors font-medium flex items-center justify-center gap-2 shadow-soft"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
              Conectar con WooCommerce
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

// Modal para conectar Shopify
function ShopifyModal({ onClose, onConnect }: {
  onClose: () => void;
  onConnect: (shopName: string, apiKey: string, password: string) => void;
}) {
  const [shopName, setShopName] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    await onConnect(shopName, apiKey, password);
    setLoading(false);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-xl max-w-md w-full p-6">
        <h3 className="text-2xl font-bold text-neutral-900 mb-4">Conectar Shopify</h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">
              Nombre de la tienda
            </label>
            <input
              type="text"
              value={shopName}
              onChange={(e) => setShopName(e.target.value)}
              placeholder="example.myshopify.com"
              required
              className="w-full px-4 py-2 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
            <p className="text-xs text-neutral-500 mt-1">Tu URL de Shopify sin https://</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">
              API Key
            </label>
            <input
              type="text"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              required
              className="w-full px-4 py-2 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">
              API Password / Access Token
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full px-4 py-2 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          <div className="flex gap-3 mt-6">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-neutral-300 rounded-xl text-neutral-700 hover:bg-neutral-50 transition-colors font-medium"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-green-600 text-white px-4 py-2 rounded-xl hover:bg-green-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? "Conectando..." : "Conectar"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// Modal para conectar WooCommerce
function WooCommerceModal({ onClose, onConnect }: {
  onClose: () => void;
  onConnect: (siteUrl: string, consumerKey: string, consumerSecret: string) => void;
}) {
  const [siteUrl, setSiteUrl] = useState("");
  const [consumerKey, setConsumerKey] = useState("");
  const [consumerSecret, setConsumerSecret] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    await onConnect(siteUrl, consumerKey, consumerSecret);
    setLoading(false);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-xl max-w-md w-full p-6">
        <h3 className="text-2xl font-bold text-neutral-900 mb-4">Conectar WooCommerce</h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">
              URL del sitio
            </label>
            <input
              type="url"
              value={siteUrl}
              onChange={(e) => setSiteUrl(e.target.value)}
              placeholder="https://tusitio.com"
              required
              className="w-full px-4 py-2 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">
              Consumer Key
            </label>
            <input
              type="text"
              value={consumerKey}
              onChange={(e) => setConsumerKey(e.target.value)}
              placeholder="ck_xxxxx"
              required
              className="w-full px-4 py-2 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">
              Consumer Secret
            </label>
            <input
              type="password"
              value={consumerSecret}
              onChange={(e) => setConsumerSecret(e.target.value)}
              placeholder="cs_xxxxx"
              required
              className="w-full px-4 py-2 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          <div className="flex gap-3 mt-6">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-neutral-300 rounded-xl text-neutral-700 hover:bg-neutral-50 transition-colors font-medium"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-purple-600 text-white px-4 py-2 rounded-xl hover:bg-purple-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? "Conectando..." : "Conectar"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
