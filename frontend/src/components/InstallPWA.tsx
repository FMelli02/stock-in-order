import { useEffect, useState } from 'react';

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

const InstallPWA = () => {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const handleBeforeInstallPrompt = (e: Event) => {
      // Prevenir que el navegador muestre el banner automático
      e.preventDefault();
      
      // Guardar el evento para usarlo después
      setDeferredPrompt(e as BeforeInstallPromptEvent);
      
      // Mostrar el botón de instalación
      setIsVisible(true);
      
      console.log('📱 PWA instalable detectada');
    };

    // Escuchar el evento beforeinstallprompt
    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt);

    // Verificar si la app ya está instalada
    if (window.matchMedia('(display-mode: standalone)').matches) {
      console.log('✅ PWA ya instalada');
      setIsVisible(false);
    }

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
    };
  }, []);

  const handleInstallClick = async () => {
    if (!deferredPrompt) {
      console.log('❌ No hay evento de instalación disponible');
      return;
    }

    // Mostrar el prompt de instalación
    await deferredPrompt.prompt();

    // Esperar a que el usuario responda al prompt
    const choiceResult = await deferredPrompt.userChoice;

    if (choiceResult.outcome === 'accepted') {
      console.log('✅ Usuario aceptó instalar la PWA');
    } else {
      console.log('❌ Usuario rechazó instalar la PWA');
    }

    // Limpiar el prompt guardado
    setDeferredPrompt(null);
    setIsVisible(false);
  };

  // No renderizar nada si no está disponible la instalación
  if (!isVisible) {
    return null;
  }

  return (
    <button
      onClick={handleInstallClick}
      className="flex items-center gap-3 px-4 py-3 w-full text-left bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 text-white rounded-lg transition-all duration-200 shadow-md hover:shadow-lg transform hover:scale-[1.02]"
      aria-label="Instalar aplicación"
    >
      <span className="text-2xl">⬇️</span>
      <div className="flex flex-col">
        <span className="font-semibold text-sm">Instalar App</span>
        <span className="text-xs opacity-90">Acceso rápido offline</span>
      </div>
    </button>
  );
};

export default InstallPWA;
