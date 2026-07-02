import { useEffect, useState, type FormEvent } from 'react'
import { adminApi } from '../api/endpoints'
import { peso } from '../format'
import { type AppSetting, type SavingsTargetType, type User } from '../types'
import { ApiError } from '../api/client'

export default function Admin() {
  const [users, setUsers] = useState<User[]>([])
  const [setting, setSetting] = useState<AppSetting | null>(null)
  const [error, setError] = useState('')
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(true)

  const [targetType, setTargetType] = useState<SavingsTargetType>('percent')
  const [targetValue, setTargetValue] = useState('20')

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [u, s] = await Promise.all([adminApi.listUsers(), adminApi.getSettings()])
      setUsers(u)
      setSetting(s)
      setTargetType(s.savingsTargetType)
      setTargetValue(s.savingsTargetValue)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load admin data')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function saveSettings(e: FormEvent) {
    e.preventDefault()
    setError('')
    setStatus('')
    try {
      const updated = await adminApi.updateSettings(targetType, targetValue)
      setSetting(updated)
      setStatus('Savings target updated. It applies to every member’s dashboard.')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update settings')
    }
  }

  if (loading) return <div className="page"><div className="spinner" /></div>

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1>Admin</h1>
          <p className="muted">Manage the global savings-target policy and view members.</p>
        </div>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {status && <div className="alert alert-success">{status}</div>}

      <div className="card">
        <div className="card-head"><h2>Savings target policy</h2></div>
        <p className="muted">
          Current: {setting?.savingsTargetType === 'percent'
            ? `${setting.savingsTargetValue}% of income`
            : `${peso(setting?.savingsTargetValue ?? '0')} fixed`}
        </p>
        <form className="settings-form" onSubmit={saveSettings}>
          <label>Type
            <select value={targetType} onChange={(e) => setTargetType(e.target.value as SavingsTargetType)}>
              <option value="percent">Percent of income</option>
              <option value="fixed">Fixed amount (₱)</option>
            </select>
          </label>
          <label>{targetType === 'percent' ? 'Percent (0–100)' : 'Amount (₱)'}
            <input
              type="number"
              min="0"
              max={targetType === 'percent' ? '100' : undefined}
              step="0.01"
              value={targetValue}
              onChange={(e) => setTargetValue(e.target.value)}
              required
            />
          </label>
          <div className="form-actions">
            <button className="btn btn-primary" type="submit">Save policy</button>
          </div>
        </form>
      </div>

      <div className="card">
        <div className="card-head"><h2>Members ({users.length})</h2></div>
        <table className="table">
          <thead><tr><th>Email</th><th>Role</th><th className="right">Monthly income</th></tr></thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id}>
                <td>{u.email}</td>
                <td><span className={`pill ${u.role === 'admin' ? 'pill-debts' : 'pill-variable'}`}>{u.role}</span></td>
                <td className="right">{peso(u.monthlyIncome)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
