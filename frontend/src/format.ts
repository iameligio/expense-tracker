// Formats a numeric string as Philippine Peso.
const pesoFmt = new Intl.NumberFormat('en-PH', {
  style: 'currency',
  currency: 'PHP',
  minimumFractionDigits: 2,
})

export function peso(value: string | number): string {
  const n = typeof value === 'string' ? parseFloat(value) : value
  if (Number.isNaN(n)) return '₱0.00'
  return pesoFmt.format(n)
}

// Current month as YYYY-MM (local time).
export function currentMonth(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

// Human-readable month label, e.g. "July 2026".
export function monthLabel(month: string): string {
  const [y, m] = month.split('-').map(Number)
  if (!y || !m) return month
  return new Date(y, m - 1, 1).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
}

// Compact month label for chart axes, e.g. "Jul '26".
export function shortMonth(month: string): string {
  const [y, m] = month.split('-').map(Number)
  if (!y || !m) return month
  const mon = new Date(y, m - 1, 1).toLocaleDateString('en-US', { month: 'short' })
  return `${mon} '${String(y).slice(2)}`
}

// Shift a YYYY-MM month string by delta months (delta may be negative).
export function shiftMonth(month: string, delta: number): string {
  const [y, m] = month.split('-').map(Number)
  const d = new Date(y, m - 1 + delta, 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}
