import { useEffect, useState, type FormEvent } from 'react'
import { categoryApi } from '../api/endpoints'
import { CATEGORY_TYPES, TYPE_LABELS, type Category, type CategoryType } from '../types'
import { ApiError } from '../api/client'

export default function Categories() {
  const [categories, setCategories] = useState<Category[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const [editingId, setEditingId] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [type, setType] = useState<CategoryType>('fixed')

  async function load() {
    setLoading(true)
    setError('')
    try {
      setCategories(await categoryApi.list())
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load categories')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  function resetForm() {
    setEditingId(null)
    setName('')
    setType('fixed')
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    try {
      if (editingId) await categoryApi.update(editingId, name.trim(), type)
      else await categoryApi.create(name.trim(), type)
      resetForm()
      load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to save category')
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this category?')) return
    try {
      await categoryApi.remove(id)
      if (editingId === id) resetForm()
      load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to delete category')
    }
  }

  const grouped = CATEGORY_TYPES.map((t) => ({
    type: t,
    items: categories.filter((c) => c.type === t),
  }))

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1>Categories</h1>
          <p className="muted">Organize spending into Fixed, Variable, Wants, and Debts.</p>
        </div>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="card">
        <div className="card-head"><h2>{editingId ? 'Edit category' : 'Add category'}</h2></div>
        <form className="category-form" onSubmit={handleSubmit}>
          <label>Name
            <input type="text" value={name} onChange={(e) => setName(e.target.value)} maxLength={100} required />
          </label>
          <label>Bucket
            <select value={type} onChange={(e) => setType(e.target.value as CategoryType)}>
              {CATEGORY_TYPES.map((t) => <option key={t} value={t}>{TYPE_LABELS[t]}</option>)}
            </select>
          </label>
          <div className="form-actions">
            <button className="btn btn-primary" type="submit">{editingId ? 'Save' : 'Add'}</button>
            {editingId && <button className="btn btn-ghost" type="button" onClick={resetForm}>Cancel</button>}
          </div>
        </form>
      </div>

      {loading ? (
        <div className="spinner" />
      ) : (
        <div className="category-groups">
          {grouped.map((g) => (
            <div className="card" key={g.type}>
              <div className="card-head">
                <h2><span className={`pill pill-${g.type}`}>{TYPE_LABELS[g.type]}</span></h2>
              </div>
              {g.items.length === 0 ? (
                <p className="muted">None yet.</p>
              ) : (
                <ul className="cat-list">
                  {g.items.map((c) => (
                    <li key={c.id}>
                      <span>{c.name}</span>
                      <span className="nowrap">
                        <button className="btn btn-ghost btn-sm" onClick={() => { setEditingId(c.id); setName(c.name); setType(c.type) }}>Edit</button>
                        <button className="btn btn-danger btn-sm" onClick={() => handleDelete(c.id)}>Delete</button>
                      </span>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
