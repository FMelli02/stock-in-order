import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import * as Sentry from '@sentry/react'
import { registerSW } from 'virtual:pwa-register'

// Initialize Sentry for error tracking
const sentryDsn = import.meta.env.VITE_SENTRY_DSN
if (sentryDsn) {
  Sentry.init({
    dsn: sentryDsn,
    environment: import.meta.env.MODE, // 'development' or 'production'
    integrations: [
      Sentry.browserTracingIntegration(),
      Sentry.replayIntegration({
        maskAllText: false,
        blockAllMedia: false,
      }),
    ],
    // Performance Monitoring
    tracesSampleRate: 1.0, // Capture 100% of transactions in development
    // Session Replay
    replaysSessionSampleRate: 0.1, // 10% of sessions
    replaysOnErrorSampleRate: 1.0, // 100% of sessions with errors
  })
  console.log('Sentry initialized for frontend')
} else {
  console.warn('VITE_SENTRY_DSN not configured, error monitoring disabled')
}

// Register service worker for PWA functionality
registerSW({
  onNeedRefresh() {
    console.log('Nueva versión disponible, recargando...')
  },
  onOfflineReady() {
    console.log('App lista para funcionar offline')
  },
  onRegistered(registration) {
    console.log('Service Worker registrado:', registration)
  },
  onRegisterError(error) {
    console.error('Error registrando Service Worker:', error)
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
