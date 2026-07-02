import { useEffect, useMemo, useState } from 'react'
import { dashboardApi, meApi } from '../api/endpoints'
import { useAuth } from '../auth/AuthContext'
import PieChart, { type PieDatum } from '../components/PieChart'
import { currentMonth, monthLabel, peso } from '../format'
import { TYPE_LABELS, type Dashboard as DashboardData } from '../types'
import { ApiError } from '../api/client'

export default function Dashboard() {
  const { user, refreshUser } = useAuth()
  const [month, setMonth] = useState(currentMonth())
  const [data, setData] = useState<DashboardData | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [chartMode, setChartMode] = useState<'category' | 'type'>('category')

  const [editingIncome, setEditingIncome] = useState(false)
  const [incomeInput, setIncomeInput] = useState('')

  async function load() {
    setLoading(true)
    setError('')
    try {
      setData(await dashboardApi.get(month))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load dashboard')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [month])

  async function saveIncome() {
    try {
      await meApi.setIncome(incomeInput || '0')
      await refreshUser()
      setEditingIncome(false)
      load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update income')
    }
  }

  const pieData: PieDatum[] = useMemo(() => {
    if (!data) return []
    if (chartMode === 'type') {
      return data.typeBreakdown.map((t) => ({ label: TYPE_LABELS[t.type], value: parseFloat(t.total) }))
    }
    return data.categoryBreakdown.map((c) => ({ label: c.name, value: parseFloat(c.total) }))
  }, [data, chartMode])

  const summary = data?.summary

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1>Dashboard</h1>
          <p className="muted">{monthLabel(month)}</p>
        </div>
        <input
          type="month"
          className="month-picker"
          value={month}
          onChange={(e) => setMonth(e.target.value)}
        />
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="income-bar">
        <span>Monthly income</span>
        {editingIncome ? (
          <div className="income-edit">
            <input
              type="number"
              min="0"
              step="0.01"
              value={incomeInput}
              onChange={(e) => setIncomeInput(e.target.value)}
              autoFocus
            />
            <button className="btn btn-primary btn-sm" onClick={saveIncome}>Save</button>
            <button className="btn btn-ghost btn-sm" onClick={() => setEditingIncome(false)}>Cancel</button>
          </div>
        ) : (
          <div className="income-display">
            <strong>{peso(user?.monthlyIncome ?? '0')}</strong>
            <button
              className="btn btn-ghost btn-sm"
              onClick={() => {
                setIncomeInput(user?.monthlyIncome ?? '')
                setEditingIncome(true)
              }}
            >
              Edit
            </button>
          </div>
        )}
      </div>

      {loading ? (
        <div className="center-screen"><div className="spinner" /></div>
      ) : summary ? (
        <>
          <div className="kpi-grid">
            <KpiCard label="Total Income" value={peso(summary.income)} tone="neutral" />
            <KpiCard label="Total Expenses" value={peso(summary.totalExpenses)} tone="spend" />
            <KpiCard
              label="Savings Target"
              value={peso(summary.savingsTarget)}
              tone="neutral"
            />
            <KpiCard
              label="Actual Savings"
              value={peso(summary.actualSavings)}
              tone={summary.targetMet ? 'good' : 'warn'}
              badge={summary.targetMet ? 'On track' : 'Below target'}
            />
            <KpiCard
              label="Remaining Budget"
              value={peso(summary.remainingBudget)}
              tone={parseFloat(summary.remainingBudget) >= 0 ? 'good' : 'bad'}
            />
          </div>

          <div className="card chart-card">
            <div className="card-head">
              <h2>Where your money goes</h2>
              <div className="toggle">
                <button
                  className={chartMode === 'category' ? 'active' : ''}
                  onClick={() => setChartMode('category')}
                >
                  By category
                </button>
                <button
                  className={chartMode === 'type' ? 'active' : ''}
                  onClick={() => setChartMode('type')}
                >
                  By bucket
                </button>
              </div>
            </div>
            <div className="chart-wrap">
              <PieChart data={pieData} />
            </div>
          </div>
        </>
      ) : null}
    </div>
  )
}

function KpiCard({
  label,
  value,
  tone,
  badge,
}: {
  label: string
  value: string
  tone: 'neutral' | 'good' | 'bad' | 'warn' | 'spend'
  badge?: string
}) {
  return (
    <div className={`card kpi kpi-${tone}`}>
      <span className="kpi-label">{label}</span>
      <span className="kpi-value">{value}</span>
      {badge && <span className="kpi-badge">{badge}</span>}
    </div>
  )
}
