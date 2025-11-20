import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Html5QrcodeScanner } from 'html5-qrcode';
import { Camera, X, Barcode } from 'lucide-react';

interface ScannerConfig {
  fps: number;
  qrbox: { width: number; height: number };
  aspectRatio: number;
  disableFlip: boolean;
}

const ScannerPage: React.FC = () => {
  const navigate = useNavigate();
  const [scanner, setScanner] = useState<Html5QrcodeScanner | null>(null);
  const [isScanning, setIsScanning] = useState(false);

  useEffect(() => {
    // Detectar si es móvil para ajustar configuración
    const isMobile = window.innerWidth < 768;
    
    // Configuración del scanner optimizada para móviles
    const config: ScannerConfig = {
      fps: 10, // Frames por segundo
      qrbox: { 
        width: isMobile ? Math.min(250, window.innerWidth - 40) : 250, 
        height: isMobile ? Math.min(250, window.innerWidth - 40) : 250 
      }, // Área de escaneo adaptativa
      aspectRatio: 1.0,
      disableFlip: false, // Permitir voltear imagen
    };

    // Función que se ejecuta al escanear exitosamente
    const onScanSuccess = (decodedText: string) => {
      console.log('✅ Código escaneado:', decodedText);
      
      // Detener el scanner
      if (scannerRef) {
        scannerRef.clear();
      }
      
      // Redirigir a la página de productos con el SKU
      navigate(`/products?search=${encodeURIComponent(decodedText)}`);
    };

    // Función que se ejecuta en caso de error (opcional, no es crítico)
    const onScanError = () => {
      // No mostramos errores en consola ya que son muy frecuentes mientras escanea
    };

    // Inicializar el scanner
    const html5QrcodeScanner = new Html5QrcodeScanner(
      'barcode-scanner', // ID del elemento HTML
      config,
      false // verbose (false para menos logs)
    );

    html5QrcodeScanner.render(onScanSuccess, onScanError);
    const scannerRef = html5QrcodeScanner;
    setScanner(html5QrcodeScanner);
    setIsScanning(true);

    // Cleanup: detener el scanner cuando el componente se desmonte
    return () => {
      if (scannerRef) {
        scannerRef.clear().catch((error) => {
          console.error('Error al limpiar el scanner:', error);
        });
      }
    };
  }, [navigate]); // Dependencia necesaria para navigate

  const handleStop = () => {
    if (scanner) {
      scanner.clear().then(() => {
        setIsScanning(false);
        navigate('/products');
      }).catch((error) => {
        console.error('Error al detener el scanner:', error);
      });
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="bg-white rounded-lg shadow-md p-3 sm:p-6">
        {/* Header - Optimizado para móviles */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
          <div className="flex items-center gap-3">
            <div className="bg-blue-100 p-2 sm:p-3 rounded-lg">
              <Camera className="w-5 h-5 sm:w-6 sm:h-6 text-blue-600" />
            </div>
            <div>
              <h1 className="text-xl sm:text-2xl font-bold text-gray-900">
                Escanear Código
              </h1>
              <p className="text-sm sm:text-base text-gray-600">
                Apunta al código de barras
              </p>
            </div>
          </div>
          {/* Botón grande para touch en móviles - mínimo 44px de altura */}
          <button
            onClick={handleStop}
            className="flex items-center justify-center gap-2 w-full sm:w-auto px-6 py-3 sm:py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-semibold text-base sm:text-sm min-h-[44px] touch-manipulation"
            aria-label="Detener scanner"
          >
            <X className="w-5 h-5" />
            <span>Detener Scanner</span>
          </button>
        </div>

        {/* Instrucciones - Colapsables en móvil */}
        <div className="mb-4 sm:mb-6 bg-blue-50 border border-blue-200 rounded-lg p-3 sm:p-4">
          <div className="flex items-start gap-2 sm:gap-3">
            <Barcode className="w-4 h-4 sm:w-5 sm:h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <div>
              <h3 className="font-semibold text-blue-900 mb-1 sm:mb-2 text-sm sm:text-base">
                Instrucciones:
              </h3>
              <ul className="text-xs sm:text-sm text-blue-800 space-y-1">
                <li>• Permite el acceso a la cámara</li>
                <li>• Coloca el código dentro del cuadro</li>
                <li>• Mantén el código estable e iluminado</li>
                <li className="hidden sm:list-item">• El escaneo es automático</li>
                <li className="hidden sm:list-item">• Serás redirigido al producto automáticamente</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Área del Scanner - Ocupa más espacio en móviles */}
        <div className="scanner-container">
          <div id="barcode-scanner" className="w-full min-h-[280px] sm:min-h-[320px]"></div>
        </div>

        {/* Estado */}
        {isScanning && (
          <div className="mt-4 flex items-center justify-center gap-2 text-green-600 py-2">
            <div className="animate-pulse w-2 h-2 bg-green-600 rounded-full"></div>
            <span className="text-xs sm:text-sm font-medium">Scanner activo</span>
          </div>
        )}
      </div>

      {/* Estilos personalizados para el scanner - Optimizado para móviles */}
      <style>{`
        #barcode-scanner {
          border-radius: 8px;
          overflow: hidden;
        }

        #barcode-scanner video {
          border-radius: 8px;
          width: 100% !important;
          min-height: 280px !important;
          object-fit: cover;
        }
        
        @media (min-width: 640px) {
          #barcode-scanner video {
            min-height: 320px !important;
          }
        }

        /* Botones del scanner - Touch friendly */
        #barcode-scanner button {
          margin-top: 1rem;
          padding: 0.75rem 1.25rem;
          min-height: 44px;
          background-color: #3B82F6;
          color: white;
          border: none;
          border-radius: 0.5rem;
          cursor: pointer;
          font-weight: 600;
          font-size: 1rem;
          transition: all 0.2s;
          touch-action: manipulation;
        }

        #barcode-scanner button:hover {
          background-color: #2563EB;
          transform: scale(1.02);
        }

        #barcode-scanner button:active {
          transform: scale(0.98);
        }

        #barcode-scanner select {
          margin: 0.5rem 0;
          padding: 0.75rem;
          min-height: 44px;
          border: 1px solid #D1D5DB;
          border-radius: 0.5rem;
          background-color: white;
          cursor: pointer;
          font-size: 1rem;
          touch-action: manipulation;
        }

        /* Ocultar el mensaje de error por defecto */
        #barcode-scanner__dashboard_section_csr {
          margin-top: 1rem;
        }
        
        /* Mejorar espaciado en móviles */
        @media (max-width: 640px) {
          #barcode-scanner {
            padding: 0;
          }
          
          #barcode-scanner > div {
            padding: 0 !important;
          }
        }
      `}</style>
    </div>
  );
};

export default ScannerPage;
