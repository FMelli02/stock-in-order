import { useEffect, useState } from 'react'
import api from '../services/api'
import type { Supplier } from '../types/supplier'

type Props = {
  supplierToEdit: Supplier | null
  onSuccess: () => void
}

export default function SupplierForm({ supplierToEdit, onSuccess }: Props) {
  const [name, setName] = useState('')
  const [contactPerson, setContactPerson] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (supplierToEdit) {
      setName(supplierToEdit.name)
      setContactPerson(supplierToEdit.contact_person)
      setEmail(supplierToEdit.email)
      setPhone(supplierToEdit.phone)
    } else {
      setName('')
      setContactPerson('')
      setEmail('')
      setPhone('')
    }
  }, [supplierToEdit])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const payload = { name, contact_person: contactPerson, email, phone }
      if (supplierToEdit) {
        await api.put(`/suppliers/${supplierToEdit.id}`, payload)
      } else {
        await api.post('/suppliers', payload)
      }
      onSuccess()
    } catch (err) {
      console.error(err)
      setError('No se pudo guardar el proveedor')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h2 className="text-lg font-semibold">
        {supplierToEdit ? 'Editar Proveedor' : 'Crear Proveedor'}
      </h2>
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
        <label className="block text-sm font-medium text-gray-700">Persona de Contacto</label>
        <input
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2"
          value={contactPerson}
          onChange={(e) => setContactPerson(e.target.value)}
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
