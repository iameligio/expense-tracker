import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'

export default function Layout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/login')
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">
          <span className="brand-mark">₱</span>
          <span>Expense Tracker</span>
        </div>
        <nav className="nav">
          <NavLink to="/dashboard">Dashboard</NavLink>
          <NavLink to="/income">Income</NavLink>
          <NavLink to="/expenses">Expenses</NavLink>
          <NavLink to="/trends">Trends</NavLink>
          <NavLink to="/categories">Categories</NavLink>
          {user?.role === 'admin' && <NavLink to="/admin">Admin</NavLink>}
        </nav>
        <div className="user-menu">
          <span className="user-email">{user?.email}</span>
          <button className="btn btn-ghost" onClick={handleLogout}>Log out</button>
        </div>
      </header>
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}
