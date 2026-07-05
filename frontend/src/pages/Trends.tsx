import { useEffect, useMemo, useState } from 'react'
import { trendsApi } from '../api/endpoints'
import TrendChart, { type TrendChartPoint } from '../components/TrendChart'
import { currentMonth, monthLabel, peso, shiftMonth } from '../format'
import { type Trend } from '../types'
import { ApiError } from '../api/client'

type Preset = '6m' | '1y' | 'ytd' | 'custom'

const PRESETS: { key: Preset; label: string }[] = [
  { key: '6m', label: '6 months' },
  { key: '1y', label: '1 year' },
  { key: 'ytd', label: 'Year to date' },
  { key: 'custom', label: 'Custom' },
]

// Resolve a preset (plus custom pickers) into an inclusive [from, to] month pair.
function resolveRange(preset: Preset, customFrom: string, customTo: string): { from: string; to: string } {
  const cur = currentMonth()
  switch (preset) {
    case '6m':
      return { from: shiftMonth(cur, -5), to: cur }
    case '1y':
      return { from: shiftMonth(cur, -11), to: cur }
    case 'ytd':
      return { from: `${cur.slice(0, 4)}-01`, to: cur }
    case 'custom':
      return { from: customFrom, to: customTo }
  }
}

export default function Trends() {
  const cur = currentMonth()
  const [preset, setPreset] = useState<Preset>('6m')
  const [customFrom, setCustomFrom] = useState(shiftMonth(cur, -5))
  const [customTo, setCustomTo] = useState(cur)
  const [data, setData] = useState<Trend | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const { from, to } = resolveRange(preset, customFrom, customTo)

  useEffect(() => {
    let cancelled = false
    async function load() {
      setLoading(true)
      setError('')
      try {
        const trend = await trendsApi.get(from, to)
        if (!cancelled) setData(trend)
      } catch (err) {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Failed to load trends')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [from, to])

  const points: TrendChartPoint[] = useMemo(
    () =>
      (data?.months ?? []).map((m) => ({
        month: m.month,
        income: parseFloat(m.income),
        expenses: parseFloat(m.expenses),
        net: parseFloat(m.net),
      })),
    [data],
  )

  const totals = useMemo(() => {
    return points.reduce(
      (acc, p) => ({
        income: acc.income + p.income,
        expenses: acc.expenses + p.expenses,
        net: acc.net + p.net,
      }),
      { income: 0, expenses: 0, net: 0 },
    )
  }, [points])

  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1>Trends</h1>
          <p className="muted">Income vs expenses · {monthLabel(from)} – {monthLabel(to)}</p>
        </div>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="card">
        <div className="card-head">
          <h2>Income vs Expenses</h2>
          <div className="toggle">
            {PRESETS.map((p) => (
              <button
                key={p.key}
                className={preset === p.key ? 'active' : ''}
                onClick={() => setPreset(p.key)}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>

        {preset === 'custom' && (
          <div className="range-custom">
            <label>From
              <input type="month" className="month-picker" value={customFrom} max={customTo} onChange={(e) => setCustomFrom(e.target.value)} />
            </label>
            <label>To
              <input type="month" className="month-picker" value={customTo} min={customFrom} max={cur} onChange={(e) => setCustomTo(e.target.value)} />
            </label>
          </div>
        )}

        <div className="kpi-grid">
          <div className="card kpi kpi-good">
            <span className="kpi-label">Total Income</span>
            <span className="kpi-value">{peso(totals.income)}</span>
          </div>
          <div className="card kpi kpi-spend">
            <span className="kpi-label">Total Expenses</span>
            <span className="kpi-value">{peso(totals.expenses)}</span>
          </div>
          <div className={`card kpi ${totals.net >= 0 ? 'kpi-good' : 'kpi-bad'}`}>
            <span className="kpi-label">Net Savings</span>
            <span className="kpi-value">{peso(totals.net)}</span>
          </div>
        </div>

        <div className="chart-wrap chart-wrap-tall">
          {loading ? <div className="spinner" /> : <TrendChart points={points} />}
        </div>
      </div>
    </div>
  )
}
