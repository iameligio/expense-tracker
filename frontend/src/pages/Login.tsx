import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { ApiError } from '../api/client'

export default function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      await login(email, password)
      navigate('/dashboard')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Something went wrong')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-screen">
      <form className="auth-card" onSubmit={handleSubmit}>
        <div className="auth-brand"><span className="brand-mark">₱</span> Expense Tracker</div>
        <h1>Welcome back</h1>
        <p className="muted">Log in to manage your monthly budget.</p>

        {error && <div className="alert alert-error">{error}</div>}

        <label>Email
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)}
            required autoFocus autoComplete="email" inputMode="email" maxLength={255} />
        </label>
        <label>Password
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)}
            required autoComplete="current-password" maxLength={72} />
        </label>

        <button className="btn btn-primary" type="submit" disabled={busy}>
          {busy ? 'Logging in…' : 'Log in'}
        </button>
        <p className="auth-switch">No account? <Link to="/register">Create one</Link></p>
      </form>
    </div>
  )
}
