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
    // Configuración del scanner
    const config: ScannerConfig = {
      fps: 10, // Frames por segundo
      qrbox: { width: 250, height: 250 }, // Área de escaneo
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
      <div className="bg-white rounded-lg shadow-md p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className="bg-blue-100 p-3 rounded-lg">
              <Camera className="w-6 h-6 text-blue-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">
                Escanear Código de Barras
              </h1>
              <p className="text-gray-600">
                Apunta la cámara al código de barras del producto
              </p>
            </div>
          </div>
          <button
            onClick={handleStop}
            className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
          >
            <X className="w-5 h-5" />
            Detener
          </button>
        </div>

        {/* Instrucciones */}
        <div className="mb-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <Barcode className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <div>
              <h3 className="font-semibold text-blue-900 mb-2">
                Instrucciones:
              </h3>
              <ul className="text-sm text-blue-800 space-y-1">
                <li>• Permite el acceso a la cámara cuando el navegador lo solicite</li>
                <li>• Coloca el código de barras dentro del cuadro de escaneo</li>
                <li>• Mantén el código estable y bien iluminado</li>
                <li>• El escaneo es automático, no necesitas presionar ningún botón</li>
                <li>• Serás redirigido automáticamente al producto cuando se detecte</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Área del Scanner */}
        <div className="scanner-container">
          <div id="barcode-scanner" className="w-full"></div>
        </div>

        {/* Estado */}
        {isScanning && (
          <div className="mt-4 flex items-center justify-center gap-2 text-green-600">
            <div className="animate-pulse w-2 h-2 bg-green-600 rounded-full"></div>
            <span className="text-sm font-medium">Scanner activo - Listo para escanear</span>
          </div>
        )}
      </div>

      {/* Estilos personalizados para el scanner */}
      <style>{`
        #barcode-scanner {
          border-radius: 8px;
          overflow: hidden;
        }

        #barcode-scanner video {
          border-radius: 8px;
          width: 100% !important;
        }

        #barcode-scanner button {
          margin-top: 1rem;
          padding: 0.5rem 1rem;
          background-color: #3B82F6;
          color: white;
          border: none;
          border-radius: 0.375rem;
          cursor: pointer;
          font-weight: 500;
          transition: background-color 0.2s;
        }

        #barcode-scanner button:hover {
          background-color: #2563EB;
        }

        #barcode-scanner select {
          margin: 0.5rem 0;
          padding: 0.5rem;
          border: 1px solid #D1D5DB;
          border-radius: 0.375rem;
          background-color: white;
          cursor: pointer;
        }

        /* Ocultar el mensaje de error por defecto */
        #barcode-scanner__dashboard_section_csr {
          margin-top: 1rem;
        }
      `}</style>
    </div>
  );
};

export default ScannerPage;
