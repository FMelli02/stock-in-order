import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate', // Auto-actualización cuando hay cambios
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'masked-icon.svg'],
      manifest: {
        name: 'Stock In Order',
        short_name: 'StockApp',
        description: 'Gestión de inventario inteligente para PyMEs',
        theme_color: '#1f2937', // Color del navbar (Tailwind bg-gray-800)
        background_color: '#ffffff',
        display: 'standalone', // Abre sin barra del navegador
        orientation: 'portrait',
        icons: [
          {
            src: 'pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png'
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable'
          }
        ]
      },
      workbox: {
        // Estrategia de caché para funcionamiento offline
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        runtimeCaching: [{
          urlPattern: ({ url }) => url.pathname.startsWith('/api'),
          handler: 'NetworkFirst', // API: intenta internet primero
          options: {
            cacheName: 'api-cache',
            expiration: {
              maxEntries: 10,
              maxAgeSeconds: 60 * 60 * 24 // 1 día
            },
            networkTimeoutSeconds: 10 // Timeout de 10 segundos
          }
        }]
      }
    })
  ],
})
