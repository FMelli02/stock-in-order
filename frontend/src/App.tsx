import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import * as Sentry from '@sentry/react'
import MainLayout from './layouts/MainLayout'
import DashboardPage from './pages/DashboardPage'
import ProductsPage from './pages/ProductsPage'
import LoginPage from './pages/LoginPage'
import ProtectedRoute from './components/ProtectedRoute'
import SuppliersPage from './pages/SuppliersPage'
import CustomersPage from './pages/CustomersPage'
import RegisterPage from './pages/RegisterPage'
import SalesOrdersPage from './pages/SalesOrdersPage'
import CreateSalesOrderPage from './pages/CreateSalesOrderPage'
import PurchaseOrdersPage from './pages/PurchaseOrdersPage'
import CreatePurchaseOrderPage from './pages/CreatePurchaseOrderPage'
import SalesOrderDetailPage from './pages/SalesOrderDetailPage'
import PurchaseOrderDetailPage from './pages/PurchaseOrderDetailPage'
import SentryTestPage from './pages/SentryTestPage'

const router = createBrowserRouter([
  {
    path: '/',
    element: <ProtectedRoute />,
    children: [
      {
        element: <MainLayout />,
        children: [
          { index: true, element: <DashboardPage /> },
          { path: 'products', element: <ProductsPage /> },
          { path: 'suppliers', element: <SuppliersPage /> },
          { path: 'customers', element: <CustomersPage /> },
          { path: 'sales-orders', element: <SalesOrdersPage /> },
          { path: 'sales-orders/new', element: <CreateSalesOrderPage /> },
          { path: 'sales-orders/:id', element: <SalesOrderDetailPage /> },
          { path: 'purchase-orders', element: <PurchaseOrdersPage /> },
          { path: 'purchase-orders/new', element: <CreatePurchaseOrderPage /> },
          { path: 'purchase-orders/:id', element: <PurchaseOrderDetailPage /> },
          { path: 'sentry-test', element: <SentryTestPage /> }, // Testing page
        ],
      },
    ],
  },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
])

function App() {
  return (
    <Sentry.ErrorBoundary 
      fallback={({ error, resetError }) => (
        <div style={{ 
          padding: '2rem', 
          textAlign: 'center',
          minHeight: '100vh',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          alignItems: 'center',
          backgroundColor: '#f5f5f5'
        }}>
          <h1 style={{ color: '#d32f2f', marginBottom: '1rem' }}>
            ⚠️ Algo salió mal
          </h1>
          <p style={{ color: '#666', marginBottom: '2rem' }}>
            Lo sentimos, ha ocurrido un error inesperado.
          </p>
          <details style={{ marginBottom: '2rem', maxWidth: '600px', textAlign: 'left' }}>
            <summary style={{ cursor: 'pointer', fontWeight: 'bold', marginBottom: '0.5rem' }}>
              Detalles del error
            </summary>
            <pre style={{ 
              backgroundColor: '#fff', 
              padding: '1rem', 
              borderRadius: '4px',
              overflow: 'auto',
              fontSize: '0.875rem'
            }}>
              {error instanceof Error ? error.message : String(error)}
            </pre>
          </details>
          <button 
            onClick={resetError}
            style={{
              padding: '0.75rem 1.5rem',
              backgroundColor: '#1976d2',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontSize: '1rem'
            }}
          >
            Reintentar
          </button>
        </div>
      )}
      showDialog={false}
    >
      <Toaster position="top-right" />
      <RouterProvider router={router} />
    </Sentry.ErrorBoundary>
  )
}

export default App
