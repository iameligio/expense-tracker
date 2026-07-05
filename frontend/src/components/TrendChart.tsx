import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from 'chart.js'
import { Chart } from 'react-chartjs-2'
import { peso, shortMonth } from '../format'

ChartJS.register(BarElement, LineElement, PointElement, CategoryScale, LinearScale, Tooltip, Legend)

export interface TrendChartPoint {
  month: string
  income: number
  expenses: number
  net: number
}

export default function TrendChart({ points }: { points: TrendChartPoint[] }) {
  if (points.length === 0) {
    return <p className="muted">No income or expenses in this range yet.</p>
  }

  const labels = points.map((p) => shortMonth(p.month))
  const data = {
    labels,
    datasets: [
      {
        type: 'bar' as const,
        label: 'Income',
        data: points.map((p) => p.income),
        backgroundColor: '#22c55e',
        borderRadius: 4,
        order: 2,
      },
      {
        type: 'bar' as const,
        label: 'Expenses',
        data: points.map((p) => p.expenses),
        backgroundColor: '#ef4444',
        borderRadius: 4,
        order: 2,
      },
      {
        type: 'line' as const,
        label: 'Net savings',
        data: points.map((p) => p.net),
        borderColor: '#6366f1',
        backgroundColor: '#6366f1',
        tension: 0.3,
        pointRadius: 3,
        order: 1,
      },
    ],
  }

  return (
    <Chart
      type="bar"
      data={data}
      options={{
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: { position: 'bottom', labels: { color: '#cbd5e1', padding: 14 } },
          tooltip: {
            callbacks: {
              label: (ctx) => `${ctx.dataset.label}: ${peso(ctx.parsed.y ?? 0)}`,
            },
          },
        },
        scales: {
          x: { ticks: { color: '#94a3b8' }, grid: { display: false } },
          y: {
            beginAtZero: true,
            ticks: { color: '#94a3b8', callback: (v) => peso(Number(v)) },
            grid: { color: 'rgba(148,163,184,0.12)' },
          },
        },
      }}
    />
  )
}
