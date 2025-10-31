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
    // Redirigir al endpoint del backend que inicia el flujo OAuth
    window.location.href = "http://localhost:8080/api/v1/integrations/mercadolibre/connect";
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
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  const mercadolibreIntegration = integrations.find(
    (i) => i.platform === "mercadolibre"
  );

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Integraciones</h1>
        <p className="text-gray-600 mt-2">
          Conecta tu cuenta con plataformas de venta para sincronizar productos e inventario.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {/* Mercado Libre Card */}
        <div className="bg-white rounded-lg shadow-md overflow-hidden border border-gray-200 hover:shadow-lg transition-shadow">
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
                      className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors font-medium flex items-center justify-center gap-2"
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
                    } bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors font-medium flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed`}
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
                  className="w-full bg-blue-600 text-white px-4 py-3 rounded-lg hover:bg-blue-700 transition-colors font-medium flex items-center justify-center gap-2"
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

        {/* Shopify - Próximamente */}
        <div className="bg-white rounded-lg shadow-md overflow-hidden border border-gray-200 opacity-60">
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
            <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-gray-100 text-gray-800">
              Próximamente
            </span>
            <p className="text-sm text-gray-600 mt-3">
              La integración con Shopify estará disponible pronto.
            </p>
          </div>
        </div>

        {/* WooCommerce - Próximamente */}
        <div className="bg-white rounded-lg shadow-md overflow-hidden border border-gray-200 opacity-60">
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
            <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-gray-100 text-gray-800">
              Próximamente
            </span>
            <p className="text-sm text-gray-600 mt-3">
              La integración con WooCommerce estará disponible pronto.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
