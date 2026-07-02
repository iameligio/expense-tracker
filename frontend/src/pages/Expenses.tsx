import { useEffect, useState, type FormEvent } from 'react'
import { categoryApi, expenseApi, type ExpenseInput } from '../api/endpoints'
import { currentMonth, monthLabel, peso } from '../format'
import { TYPE_LABELS, type Category, type Expense } from '../types'
import { ApiError } from '../api/client'

const today = () => new Date().toISOString().slice(0, 10)

export default function Expenses() {
  const [month, setMonth] = useState(currentMonth())
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  // Form state (also used for editing when editingId is set).
  const [editingId, setEditingId] = useState<string | null>(null)
  const [amount, setAmount] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [spentOn, setSpentOn] = useState(today())
  const [note, setNote] = useState('')

  async function loadCategories() {
    try {
      const cats = await categoryApi.list()
      setCategories(cats)
      if (!categoryId && cats.length) setCategoryId(cats[0].id)
    } catch {
      /* handled elsewhere */
    }
  }

  async function loadExpenses() {
    setLoading(true)
    setError('')
    try {
      setExpenses(await expenseApi.list(month))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load expenses')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadCategories()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    loadExpenses()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [month])

  function resetForm() {
    setEditingId(null)
    setAmount('')
    setSpentOn(today())
    setNote('')
    if (categories.length) setCategoryId(categories[0].id)
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    const input: ExpenseInput = {
      amount,
      categoryId,
      spentOn,
      note: note.trim() ? note.trim() : null,
    }
    try {
      if (editingId) {
        await expenseApi.update(editingId, input)
      } else {
        await expenseApi.create(input)
      }
      resetForm()
      loadExpenses()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to save expense')
    }
  }

  function startEdit(exp: Expense) {
    setEditingId(exp.id)
    setAmount(exp.amount)
    setCategoryId(exp.categoryId)
    setSpentOn(exp.spentOn.slice(0, 10))
    setNote(exp.note ?? '')
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this expense?')) return
    try {
      await expenseApi.remove(id)
      if (editingId === id) resetForm()
      loadExpenses()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to delete expense')
    }
  }

  const total = expenses.reduce((s, e) => s + parseFloat(e.amount), 0)

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1>Expenses</h1>
          <p className="muted">{monthLabel(month)} · {expenses.length} entries · {peso(total)}</p>
        </div>
        <input type="month" className="month-picker" value={month} onChange={(e) => setMonth(e.target.value)} />
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="card">
        <div className="card-head"><h2>{editingId ? 'Edit expense' : 'Add expense'}</h2></div>
        {categories.length === 0 ? (
          <p className="muted">Create a category first on the Categories page.</p>
        ) : (
          <form className="expense-form" onSubmit={handleSubmit}>
            <label>Amount (₱)
              <input type="number" min="0.01" step="0.01" value={amount} onChange={(e) => setAmount(e.target.value)} required />
            </label>
            <label>Category
              <select value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
                {categories.map((c) => (
                  <option key={c.id} value={c.id}>{c.name} · {TYPE_LABELS[c.type]}</option>
                ))}
              </select>
            </label>
            <label>Date
              <input type="date" value={spentOn} onChange={(e) => setSpentOn(e.target.value)} required />
            </label>
            <label>Note (optional)
              <input type="text" value={note} onChange={(e) => setNote(e.target.value)} maxLength={255} />
            </label>
            <div className="form-actions">
              <button className="btn btn-primary" type="submit">{editingId ? 'Save changes' : 'Add expense'}</button>
              {editingId && <button className="btn btn-ghost" type="button" onClick={resetForm}>Cancel</button>}
            </div>
          </form>
        )}
      </div>

      <div className="card">
        <div className="card-head"><h2>This month</h2></div>
        {loading ? (
          <div className="spinner" />
        ) : expenses.length === 0 ? (
          <p className="muted">No expenses logged for {monthLabel(month)}.</p>
        ) : (
          <table className="table">
            <thead>
              <tr><th>Date</th><th>Category</th><th>Note</th><th className="right">Amount</th><th></th></tr>
            </thead>
            <tbody>
              {expenses.map((e) => (
                <tr key={e.id}>
                  <td>{e.spentOn.slice(0, 10)}</td>
                  <td>
                    <span className={`pill pill-${e.category?.type ?? 'variable'}`}>
                      {e.category?.name ?? '—'}
                    </span>
                  </td>
                  <td className="muted">{e.note ?? ''}</td>
                  <td className="right">{peso(e.amount)}</td>
                  <td className="right nowrap">
                    <button className="btn btn-ghost btn-sm" onClick={() => startEdit(e)}>Edit</button>
                    <button className="btn btn-danger btn-sm" onClick={() => handleDelete(e.id)}>Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
