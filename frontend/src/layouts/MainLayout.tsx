import { Outlet } from 'react-router-dom'
import Sidebar from '../components/Sidebar'
import ScannerFAB from '../components/ScannerFAB'

export default function MainLayout() {
  return (
    <div className="min-h-screen flex">
      <Sidebar />
      <main className="flex-1 p-6">
        <Outlet />
      </main>
      {/* Botón flotante para acceso rápido al scanner en móviles */}
      <ScannerFAB />
    </div>
  )
}
