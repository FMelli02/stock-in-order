import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Camera } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

const ScannerFAB = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    // Detectar si es dispositivo móvil
    const checkMobile = () => {
      const mobile = window.innerWidth < 768 || 
                     /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
      setIsMobile(mobile);
    };

    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // No mostrar en la página del scanner o si no es móvil o si no hay usuario
  if (!isMobile || location.pathname === '/scanner' || !user) {
    return null;
  }

  // Solo mostrar para admin y repositor (roles que gestionan inventario)
  if (user.role !== 'admin' && user.role !== 'repositor') {
    return null;
  }

  const handleClick = () => {
    navigate('/scanner');
  };

  return (
    <>
      {/* Botón flotante - Fixed position en la esquina inferior derecha */}
      <button
        onClick={handleClick}
        className="fixed bottom-6 right-6 z-50 bg-gradient-to-r from-blue-500 to-indigo-600 hover:from-blue-600 hover:to-indigo-700 text-white rounded-full shadow-2xl hover:shadow-xl transition-all duration-300 transform hover:scale-110 active:scale-95 p-4 min-w-[56px] min-h-[56px] flex items-center justify-center touch-manipulation"
        aria-label="Escanear código de barras"
      >
        <Camera className="w-6 h-6" strokeWidth={2.5} />
      </button>

      {/* Tooltip opcional que aparece brevemente */}
      <div className="fixed bottom-24 right-6 z-40 pointer-events-none opacity-0 animate-fade-in-out">
        <div className="bg-gray-900 text-white text-xs px-3 py-2 rounded-lg shadow-lg whitespace-nowrap">
          Escanear producto
        </div>
      </div>

      <style>{`
        @keyframes fade-in-out {
          0%, 100% { opacity: 0; transform: translateY(10px); }
          10%, 90% { opacity: 1; transform: translateY(0); }
        }
        
        .animate-fade-in-out {
          animation: fade-in-out 3s ease-in-out;
        }

        /* Asegurar que el FAB esté por encima de todo */
        .z-50 {
          z-index: 9999;
        }

        /* Mejorar el área de toque en móviles */
        @media (hover: none) {
          button[aria-label="Escanear código de barras"] {
            padding: 1.25rem;
          }
        }
      `}</style>
    </>
  );
};

export default ScannerFAB;
