import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'

// Mock the axios API client to avoid importing Vite-specific env usage
jest.mock('../services/api', () => ({
  __esModule: true,
  default: {
    post: jest.fn(),
  },
}))

import LoginPage from './LoginPage'

describe('LoginPage', () => {
  it('debe renderizar el formulario de login correctamente', () => {
    render(
      <BrowserRouter>
        <LoginPage />
      </BrowserRouter>
    )

    // Campo de email
    expect(screen.getByRole('textbox', { name: /email/i })).toBeInTheDocument()
    // Campo de contraseña (es un input type=password, accesible por label)
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    // Botón de envío
    expect(screen.getByRole('button', { name: /ingresar/i })).toBeInTheDocument()
  })
})
