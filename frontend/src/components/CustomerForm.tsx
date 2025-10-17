import { useEffect, useState } from 'react'
import api from '../services/api'
import type { Customer } from '../types/customer'

type Props = {
  customerToEdit: Customer | null
  onSuccess: () => void
}

export default function CustomerForm({ customerToEdit, onSuccess }: Props) {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [address, setAddress] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (customerToEdit) {
      setName(customerToEdit.name)
      setEmail(customerToEdit.email)
      setPhone(customerToEdit.phone)
      setAddress(customerToEdit.address)
    } else {
      setName('')
      setEmail('')
      setPhone('')
      setAddress('')
    }
  }, [customerToEdit])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const payload = { name, email, phone, address }
      if (customerToEdit) {
        await api.put(`/customers/${customerToEdit.id}`, payload)
      } else {
        await api.post('/customers', payload)
      }
      onSuccess()
    } catch (err) {
      console.error(err)
      setError('No se pudo guardar el cliente')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h2 className="text-lg font-semibold">{customerToEdit ? 'Editar Cliente' : 'Crear Cliente'}</h2>
      {error && <p className="text-red-600 text-sm">{error}</p>}
      <div>
        <label className="block text-sm font-medium text-gray-700">Nombre</label>
        <input
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">Email</label>
        <input
          type="email"
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">Teléfono</label>
        <input
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
          required
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">Dirección</label>
        <input
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
          value={address}
          onChange={(e) => setAddress(e.target.value)}
          required
        />
      </div>
      <div className="flex justify-end gap-2">
        <button
          type="submit"
          disabled={submitting}
          className="px-4 py-2 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {submitting ? 'Guardando...' : 'Guardar'}
        </button>
      </div>
    </form>
  )
}
