import { useEffect, useState, type FormEvent } from 'react'
import { incomeApi, type IncomeInput } from '../api/endpoints'
import { currentMonth, monthLabel, peso } from '../format'
import { INCOME_SOURCES, SOURCE_LABELS, type Income, type IncomeSource } from '../types'
import { ApiError } from '../api/client'

const today = () => new Date().toISOString().slice(0, 10)

export default function Income() {
  const [month, setMonth] = useState(currentMonth())
  const [incomes, setIncomes] = useState<Income[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  // Form state (also used for editing when editingId is set).
  const [editingId, setEditingId] = useState<string | null>(null)
  const [amount, setAmount] = useState('')
  const [source, setSource] = useState<IncomeSource>('salary')
  const [receivedOn, setReceivedOn] = useState(today())
  const [note, setNote] = useState('')

  async function loadIncomes() {
    setLoading(true)
    setError('')
    try {
      setIncomes(await incomeApi.list(month))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load income')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadIncomes()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [month])

  function resetForm() {
    setEditingId(null)
    setAmount('')
    setSource('salary')
    setReceivedOn(today())
    setNote('')
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    const input: IncomeInput = {
      amount,
      source,
      receivedOn,
      note: note.trim() ? note.trim() : null,
    }
    try {
      if (editingId) {
        await incomeApi.update(editingId, input)
      } else {
        await incomeApi.create(input)
      }
      resetForm()
      loadIncomes()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to save income')
    }
  }

  function startEdit(inc: Income) {
    setEditingId(inc.id)
    setAmount(inc.amount)
    setSource(inc.source)
    setReceivedOn(inc.receivedOn.slice(0, 10))
    setNote(inc.note ?? '')
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this income entry?')) return
    try {
      await incomeApi.remove(id)
      if (editingId === id) resetForm()
      loadIncomes()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to delete income')
    }
  }

  const total = incomes.reduce((s, i) => s + parseFloat(i.amount), 0)

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1>Income</h1>
          <p className="muted">{monthLabel(month)} · {incomes.length} entries · {peso(total)}</p>
        </div>
        <input type="month" className="month-picker" value={month} onChange={(e) => setMonth(e.target.value)} />
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="card">
        <div className="card-head"><h2>{editingId ? 'Edit income' : 'Add income'}</h2></div>
        <form className="expense-form" onSubmit={handleSubmit}>
          <label>Amount (₱)
            <input type="number" min="0.01" step="0.01" value={amount} onChange={(e) => setAmount(e.target.value)} required />
          </label>
          <label>Source
            <select value={source} onChange={(e) => setSource(e.target.value as IncomeSource)}>
              {INCOME_SOURCES.map((s) => (
                <option key={s} value={s}>{SOURCE_LABELS[s]}</option>
              ))}
            </select>
          </label>
          <label>Date
            <input type="date" value={receivedOn} onChange={(e) => setReceivedOn(e.target.value)} required />
          </label>
          <label>Note (optional)
            <input type="text" value={note} onChange={(e) => setNote(e.target.value)} maxLength={255} />
          </label>
          <div className="form-actions">
            <button className="btn btn-primary" type="submit">{editingId ? 'Save changes' : 'Add income'}</button>
            {editingId && <button className="btn btn-ghost" type="button" onClick={resetForm}>Cancel</button>}
          </div>
        </form>
      </div>

      <div className="card">
        <div className="card-head"><h2>This month</h2></div>
        {loading ? (
          <div className="spinner" />
        ) : incomes.length === 0 ? (
          <p className="muted">No income logged for {monthLabel(month)}.</p>
        ) : (
          <table className="table">
            <thead>
              <tr><th>Date</th><th>Source</th><th>Note</th><th className="right">Amount</th><th></th></tr>
            </thead>
            <tbody>
              {incomes.map((i) => (
                <tr key={i.id}>
                  <td>{i.receivedOn.slice(0, 10)}</td>
                  <td>
                    <span className={`pill pill-${i.source}`}>{SOURCE_LABELS[i.source]}</span>
                  </td>
                  <td className="muted">{i.note ?? ''}</td>
                  <td className="right">{peso(i.amount)}</td>
                  <td className="right nowrap">
                    <button className="btn btn-ghost btn-sm" onClick={() => startEdit(i)}>Edit</button>
                    <button className="btn btn-danger btn-sm" onClick={() => handleDelete(i.id)}>Delete</button>
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
