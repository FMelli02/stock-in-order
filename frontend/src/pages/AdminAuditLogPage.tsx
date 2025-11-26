import React, { useEffect, useState } from 'react';
import api from '../services/api';
import { useAuth } from '../contexts/AuthContext';

interface AuditLog {
  id: number;
  user_id: number | null;
  user_email: string | null;
  user_role: string | null;
  action: string;
  entity_type: string;
  entity_id: number | null;
  details: string | null;
  timestamp: string;
}

interface AuditLogsResponse {
  logs: AuditLog[];
  total: number;
  limit: number;
  offset: number;
}

const AdminAuditLogPage: React.FC = () => {
  const { user } = useAuth();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [limit] = useState(50);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Redirigir si no es admin
  useEffect(() => {
    if (user && user.role !== 'admin') {
      window.location.href = '/';
    }
  }, [user]);

  const fetchAuditLogs = async (currentOffset: number) => {
    setLoading(true);
    setError('');
    try {
      const response = await api.get<AuditLogsResponse>('/admin/audit-logs', {
        params: {
          limit,
          offset: currentOffset,
        },
      });
      setLogs(response.data.logs || []);
      setTotal(response.data.total);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Error al cargar logs de auditoría');
      console.error('Error fetching audit logs:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAuditLogs(offset);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [offset]);

  const handlePreviousPage = () => {
    if (offset > 0) {
      setOffset(Math.max(0, offset - limit));
    }
  };

  const handleNextPage = () => {
    if (offset + limit < total) {
      setOffset(offset + limit);
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleString('es-AR', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const getActionBadgeClass = (action: string) => {
    switch (action) {
      case 'CREATE':
        return 'bg-green-100 text-green-800';
      case 'UPDATE':
        return 'bg-blue-100 text-blue-800';
      case 'DELETE':
        return 'bg-red-100 text-red-800';
      case 'ADJUST_STOCK':
        return 'bg-purple-100 text-purple-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getActionLabel = (action: string) => {
    switch (action) {
      case 'CREATE':
        return 'Crear';
      case 'UPDATE':
        return 'Actualizar';
      case 'DELETE':
        return 'Eliminar';
      case 'ADJUST_STOCK':
        return 'Ajustar Stock';
      default:
        return action;
    }
  };

  const getEntityTypeLabel = (entityType: string) => {
    switch (entityType) {
      case 'PRODUCT':
        return 'Producto';
      case 'CUSTOMER':
        return 'Cliente';
      case 'SUPPLIER':
        return 'Proveedor';
      case 'SALES_ORDER':
        return 'Orden de Venta';
      case 'PURCHASE_ORDER':
        return 'Orden de Compra';
      case 'USER':
        return 'Usuario';
      default:
        return entityType;
    }
  };

  const currentPage = Math.floor(offset / limit) + 1;
  const totalPages = Math.ceil(total / limit);

  if (user && user.role !== 'admin') {
    return null;
  }

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-display font-bold text-neutral-900">Registro de Auditoría</h1>
        <p className="text-neutral-600 mt-2">
          Visualiza todas las acciones realizadas en el sistema
        </p>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded-xl mb-4">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex justify-center items-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : (
        <>
          <div className="bg-white shadow-soft rounded-2xl border border-neutral-200 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-neutral-200">
                <thead className="bg-neutral-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">
                      Fecha y Hora
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">
                      Usuario
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">
                      Rol
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">
                      Acción
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">
                      Entidad
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">
                      ID Entidad
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">
                      Detalles
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-neutral-200">
                  {logs.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="px-6 py-4 text-center text-neutral-500">
                        No hay registros de auditoría
                      </td>
                    </tr>
                  ) : (
                    logs.map((log) => (
                      <tr key={log.id} className="hover:bg-neutral-50">
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-neutral-900">
                          {formatTimestamp(log.timestamp)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-neutral-900">
                          {log.user_email || 'Sistema'}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                          <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-primary-100 text-primary-800">
                            {log.user_role || 'N/A'}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                          <span
                            className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getActionBadgeClass(
                              log.action
                            )}`}
                          >
                            {getActionLabel(log.action)}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-neutral-900">
                          {getEntityTypeLabel(log.entity_type)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-neutral-900">
                          {log.entity_id || 'N/A'}
                        </td>
                        <td className="px-6 py-4 text-sm text-neutral-900 max-w-xs truncate">
                          {log.details || 'Sin detalles'}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* Paginación */}
          <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-neutral-200 sm:px-6 mt-4 rounded-2xl shadow-soft">
            <div className="flex-1 flex justify-between sm:hidden">
              <button
                onClick={handlePreviousPage}
                disabled={offset === 0}
                className={`relative inline-flex items-center px-4 py-2 border border-neutral-300 text-sm font-medium rounded-xl ${
                  offset === 0
                    ? 'bg-neutral-100 text-neutral-400 cursor-not-allowed'
                    : 'bg-white text-neutral-700 hover:bg-neutral-50'
                }`}
              >
                Anterior
              </button>
                <button
                  onClick={handleNextPage}
                  disabled={offset + limit >= total}
                  className={`ml-3 relative inline-flex items-center px-4 py-2 border border-neutral-300 text-sm font-medium rounded-xl ${
                    offset + limit >= total
                      ? 'bg-neutral-100 text-neutral-400 cursor-not-allowed'
                      : 'bg-white text-neutral-700 hover:bg-neutral-50'
                }`}
              >
                Siguiente
              </button>
            </div>
            <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-neutral-700">
                  Mostrando <span className="font-medium">{offset + 1}</span> a{' '}
                  <span className="font-medium">
                    {Math.min(offset + limit, total)}
                  </span>{' '}
                  de <span className="font-medium">{total}</span> registros
                </p>
              </div>
              <div>
                <nav
                  className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px"
                  aria-label="Pagination"
                >
                  <button
                    onClick={handlePreviousPage}
                    disabled={offset === 0}
                    className={`relative inline-flex items-center px-2 py-2 rounded-l-xl border border-neutral-300 text-sm font-medium ${
                      offset === 0
                        ? 'bg-neutral-100 text-neutral-400 cursor-not-allowed'
                        : 'bg-white text-neutral-500 hover:bg-neutral-50'
                    }`}
                  >
                    <span className="sr-only">Anterior</span>
                    <svg
                      className="h-5 w-5"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        fillRule="evenodd"
                        d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </button>
                  <span className="relative inline-flex items-center px-4 py-2 border border-neutral-300 bg-white text-sm font-medium text-neutral-700">
                    Página {currentPage} de {totalPages}
                  </span>
                  <button
                    onClick={handleNextPage}
                    disabled={offset + limit >= total}
                    className={`relative inline-flex items-center px-2 py-2 rounded-r-xl border border-neutral-300 text-sm font-medium ${
                      offset + limit >= total
                        ? 'bg-neutral-100 text-neutral-400 cursor-not-allowed'
                        : 'bg-white text-neutral-500 hover:bg-neutral-50'
                    }`}
                  >
                    <span className="sr-only">Siguiente</span>
                    <svg
                      className="h-5 w-5"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        fillRule="evenodd"
                        d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </button>
                </nav>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default AdminAuditLogPage;
