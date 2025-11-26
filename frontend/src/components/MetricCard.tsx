type Props = {
  title: string
  value: number | string
  variant?: 'default' | 'warning' | 'danger' | 'success'
}

export default function MetricCard({ title, value, variant = 'default' }: Props) {
  const variants = {
    default: 'border-neutral-200',
    warning: 'border-amber-200 bg-amber-50/50',
    danger: 'border-red-200 bg-red-50/50',
    success: 'border-emerald-200 bg-emerald-50/50',
  }
  
  return (
    <div className={`rounded-2xl border bg-white p-6 shadow-soft transition-all hover:shadow-soft-lg ${variants[variant]}`}>
      <div className="text-sm text-neutral-500 font-medium mb-2">{title}</div>
      <div className="text-3xl font-display font-semibold text-neutral-900">{value}</div>
    </div>
  )
}
