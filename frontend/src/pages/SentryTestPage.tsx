import { useState } from 'react'
import * as Sentry from '@sentry/react'

/**
 * Página de prueba para verificar el funcionamiento de Sentry
 * Esta página permite generar diferentes tipos de errores para testing
 */
export default function SentryTestPage() {
  const [count, setCount] = useState(0)

  // Test 1: Error de React (capturado por ErrorBoundary)
  const throwReactError = () => {
    throw new Error('🔴 Test: React Component Error')
  }

  // Test 2: Error asíncrono (capturado por Sentry global)
  const throwAsyncError = async () => {
    setTimeout(() => {
      throw new Error('🔴 Test: Async Error')
    }, 100)
  }

  // Test 3: Error manual capturado
  const captureManualError = () => {
    try {
      throw new Error('🔴 Test: Manual Captured Error')
    } catch (error) {
      Sentry.captureException(error, {
        tags: {
          test_type: 'manual',
          page: 'sentry-test',
        },
        extra: {
          count,
          timestamp: new Date().toISOString(),
        },
      })
      alert('Error capturado manualmente y enviado a Sentry!')
    }
  }

  // Test 4: Mensaje custom en Sentry
  const sendCustomMessage = () => {
    Sentry.captureMessage('✅ Test: Custom Message from Sentry Test Page', {
      level: 'info',
      tags: {
        test_type: 'message',
      },
      extra: {
        count,
      },
    })
    alert('Mensaje enviado a Sentry!')
  }

  // Test 5: Configurar contexto de usuario
  const setUserContext = () => {
    Sentry.setUser({
      id: '12345',
      email: 'test@example.com',
      username: 'test_user',
    })
    Sentry.setTag('environment', 'testing')
    alert('Contexto de usuario configurado en Sentry!')
  }

  return (
    <div style={{ padding: '2rem', maxWidth: '800px', margin: '0 auto' }}>
      <h1>🧪 Sentry Testing Page</h1>
      <p>
        Esta página permite probar diferentes tipos de captura de errores con Sentry.
        <br />
        <strong>Nota:</strong> Verifica la consola del navegador y el dashboard de Sentry.
      </p>

      <div style={{ 
        backgroundColor: '#f5f5f5', 
        padding: '1rem', 
        borderRadius: '8px',
        marginBottom: '2rem'
      }}>
        <p><strong>Contador de prueba:</strong> {count}</p>
        <button 
          onClick={() => setCount(count + 1)}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: '#4caf50',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Incrementar
        </button>
      </div>

      <div style={{ display: 'grid', gap: '1rem', marginBottom: '2rem' }}>
        <button 
          onClick={throwReactError}
          style={{
            padding: '1rem',
            backgroundColor: '#f44336',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontWeight: 'bold',
          }}
        >
          🔴 Test 1: Throw React Error (ErrorBoundary)
        </button>

        <button 
          onClick={throwAsyncError}
          style={{
            padding: '1rem',
            backgroundColor: '#ff9800',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontWeight: 'bold',
          }}
        >
          🟠 Test 2: Throw Async Error
        </button>

        <button 
          onClick={captureManualError}
          style={{
            padding: '1rem',
            backgroundColor: '#2196f3',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontWeight: 'bold',
          }}
        >
          🔵 Test 3: Capture Manual Error
        </button>

        <button 
          onClick={sendCustomMessage}
          style={{
            padding: '1rem',
            backgroundColor: '#9c27b0',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontWeight: 'bold',
          }}
        >
          🟣 Test 4: Send Custom Message
        </button>

        <button 
          onClick={setUserContext}
          style={{
            padding: '1rem',
            backgroundColor: '#607d8b',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontWeight: 'bold',
          }}
        >
          ⚙️ Test 5: Set User Context
        </button>
      </div>

      <div style={{ 
        backgroundColor: '#fff3cd', 
        padding: '1rem', 
        borderRadius: '8px',
        border: '1px solid #ffc107'
      }}>
        <h3>⚠️ Resultados Esperados:</h3>
        <ul style={{ marginLeft: '1.5rem' }}>
          <li><strong>Test 1:</strong> ErrorBoundary mostrará fallback UI</li>
          <li><strong>Test 2:</strong> Error en consola + capturado por Sentry global</li>
          <li><strong>Test 3:</strong> Error capturado manualmente con contexto</li>
          <li><strong>Test 4:</strong> Mensaje info en Sentry (no es error)</li>
          <li><strong>Test 5:</strong> Contexto de usuario añadido a futuros eventos</li>
        </ul>
      </div>

      <div style={{ 
        marginTop: '2rem',
        padding: '1rem',
        backgroundColor: '#e3f2fd',
        borderRadius: '8px',
        border: '1px solid #2196f3'
      }}>
        <h3>📊 Verificar en Sentry:</h3>
        <ol style={{ marginLeft: '1.5rem' }}>
          <li>Ir a Sentry Dashboard</li>
          <li>Navegar a Issues → All Issues</li>
          <li>Ver los errores capturados con contexto completo</li>
          <li>Revisar breadcrumbs y session replay (si está configurado)</li>
        </ol>
      </div>
    </div>
  )
}
