import { useEffect, useState, type FormEvent } from 'react'
import { adminApi } from '../api/endpoints'
import { useAuth } from '../auth/AuthContext'
import { peso } from '../format'
import { type AppSetting, type SavingsTargetType, type User, type UserStatus } from '../types'
import { ApiError } from '../api/client'

const STATUS_PILL: Record<UserStatus, string> = {
  active: 'pill-variable',
  suspended: 'pill-wants',
  banned: 'pill-debts',
}

export default function Admin() {
  const { user: me } = useAuth()
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

  async function changeStatus(u: User, next: UserStatus, verb: string) {
    if (next !== 'active' && !confirm(`${verb} ${u.email}? This will end their active sessions.`)) return
    setError('')
    setStatus('')
    try {
      const updated = await adminApi.setUserStatus(u.id, next)
      setUsers((list) => list.map((x) => (x.id === u.id ? updated : x)))
      setStatus(`${u.email} is now ${next}.`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update user status')
    }
  }

  async function removeUser(u: User) {
    if (!confirm(`Permanently delete ${u.email} and ALL their data? This cannot be undone.`)) return
    setError('')
    setStatus('')
    try {
      await adminApi.deleteUser(u.id)
      setUsers((list) => list.filter((x) => x.id !== u.id))
      setStatus(`${u.email} has been deleted.`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to delete user')
    }
  }

  if (loading) return <div className="page"><div className="spinner" /></div>

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1>Admin</h1>
          <p className="muted">Manage members and the global savings-target policy.</p>
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
          <thead>
            <tr><th>Email</th><th>Role</th><th>Status</th><th className="right">Monthly income</th><th>Actions</th></tr>
          </thead>
          <tbody>
            {users.map((u) => {
              const isSelf = u.id === me?.id
              return (
                <tr key={u.id}>
                  <td>{u.email}{isSelf && <span className="muted"> (you)</span>}</td>
                  <td><span className={`pill ${u.role === 'admin' ? 'pill-debts' : 'pill-variable'}`}>{u.role}</span></td>
                  <td><span className={`pill ${STATUS_PILL[u.status]}`}>{u.status}</span></td>
                  <td className="right">{peso(u.monthlyIncome)}</td>
                  <td className="nowrap">
                    {isSelf ? (
                      <span className="muted">—</span>
                    ) : (
                      <div className="action-row">
                        {u.status === 'active' ? (
                          <button className="btn btn-ghost btn-sm" onClick={() => changeStatus(u, 'suspended', 'Suspend')}>Suspend</button>
                        ) : (
                          <button className="btn btn-ghost btn-sm" onClick={() => changeStatus(u, 'active', 'Reactivate')}>Reactivate</button>
                        )}
                        {u.status !== 'banned' && (
                          <button className="btn btn-danger btn-sm" onClick={() => changeStatus(u, 'banned', 'Ban')}>Ban</button>
                        )}
                        <button className="btn btn-danger btn-sm" onClick={() => removeUser(u)}>Delete</button>
                      </div>
                    )}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
