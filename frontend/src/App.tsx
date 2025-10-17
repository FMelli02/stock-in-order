import { createBrowserRouter, RouterProvider } from 'react-router-dom'
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
        ],
      },
    ],
  },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
])

function App() {
  return <RouterProvider router={router} />
}

export default App
