type Props = {
  title: string
  value: number
  color?: string // Tailwind classes for emphasis (bg/border)
}

export default function MetricCard({ title, value, color }: Props) {
  const base = 'rounded-lg shadow p-4 bg-white'
  const emphasis = color ? `${color}` : ''
  return (
    <div className={`${base} ${emphasis}`.trim()}>
      <div className="text-sm text-gray-500">{title}</div>
      <div className="mt-1 text-3xl font-bold">{value}</div>
    </div>
  )
}
