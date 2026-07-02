import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { ApiError } from '../api/client'

export default function Register() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    if (password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }
    if (password.length > 72) {
      setError('Password must be at most 72 characters')
      return
    }
    const local = email.split('@')[0]?.toLowerCase() ?? ''
    if (local.length >= 4 && password.toLowerCase().includes(local)) {
      setError('Password must not contain your email name')
      return
    }
    setBusy(true)
    try {
      await register(email, password)
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
        <h1>Create your account</h1>
        <p className="muted">Start tracking where your money goes each month.</p>

        {error && <div className="alert alert-error">{error}</div>}

        <label>Email
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)}
            required autoFocus autoComplete="email" inputMode="email" maxLength={255} />
        </label>
        <label>Password
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)}
            required minLength={8} maxLength={72} autoComplete="new-password" placeholder="At least 8 characters" />
        </label>

        <button className="btn btn-primary" type="submit" disabled={busy}>
          {busy ? 'Creating…' : 'Create account'}
        </button>
        <p className="auth-switch">Already registered? <Link to="/login">Log in</Link></p>
      </form>
    </div>
  )
}
