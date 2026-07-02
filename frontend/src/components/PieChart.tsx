import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Pie } from 'react-chartjs-2'
import { peso } from '../format'

ChartJS.register(ArcElement, Tooltip, Legend)

export interface PieDatum {
  label: string
  value: number
}

const PALETTE = [
  '#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#06b6d4',
  '#a855f7', '#ec4899', '#14b8a6', '#f97316', '#84cc16',
]

export default function PieChart({ data }: { data: PieDatum[] }) {
  if (data.length === 0) {
    return <p className="muted">No spending recorded for this month yet.</p>
  }

  const chartData = {
    labels: data.map((d) => d.label),
    datasets: [
      {
        data: data.map((d) => d.value),
        backgroundColor: data.map((_, i) => PALETTE[i % PALETTE.length]),
        borderColor: '#0f172a',
        borderWidth: 2,
      },
    ],
  }

  return (
    <Pie
      data={chartData}
      options={{
        plugins: {
          legend: { position: 'bottom', labels: { color: '#cbd5e1', padding: 14 } },
          tooltip: {
            callbacks: {
              label: (ctx) => {
                const total = data.reduce((s, d) => s + d.value, 0)
                const val = ctx.parsed
                const pct = total ? ((val / total) * 100).toFixed(1) : '0'
                return `${ctx.label}: ${peso(val)} (${pct}%)`
              },
            },
          },
        },
      }}
    />
  )
}
