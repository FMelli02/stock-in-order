// Utilidades de estilos consistentes para toda la aplicación

export const buttonStyles = {
  primary: 'px-4 py-2.5 bg-primary-600 hover:bg-primary-700 text-white rounded-xl font-medium transition-all shadow-soft hover:shadow-soft-lg disabled:opacity-50 disabled:cursor-not-allowed text-sm',
  secondary: 'px-4 py-2.5 bg-neutral-100 hover:bg-neutral-200 text-neutral-900 rounded-xl font-medium transition-all border border-neutral-300 disabled:opacity-50 disabled:cursor-not-allowed text-sm',
  danger: 'px-4 py-2.5 bg-red-600 hover:bg-red-700 text-white rounded-xl font-medium transition-all shadow-soft hover:shadow-soft-lg disabled:opacity-50 disabled:cursor-not-allowed text-sm',
  success: 'px-4 py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-medium transition-all shadow-soft hover:shadow-soft-lg disabled:opacity-50 disabled:cursor-not-allowed text-sm',
  ghost: 'px-4 py-2.5 hover:bg-neutral-100 text-neutral-700 rounded-xl font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed text-sm',
  link: 'text-primary-600 hover:text-primary-700 font-medium transition-colors underline-offset-4 hover:underline text-sm',
}

export const inputStyles = {
  base: 'w-full px-4 py-2.5 border border-neutral-300 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all text-sm placeholder:text-neutral-400',
  error: 'w-full px-4 py-2.5 border border-red-300 rounded-xl focus:ring-2 focus:ring-red-500 focus:border-red-500 outline-none transition-all text-sm placeholder:text-neutral-400 bg-red-50',
}

export const cardStyles = {
  base: 'bg-white rounded-2xl border border-neutral-200 shadow-soft p-6',
  interactive: 'bg-white rounded-2xl border border-neutral-200 shadow-soft p-6 transition-all hover:shadow-soft-lg cursor-pointer',
}

export const tableStyles = {
  container: 'bg-white rounded-2xl border border-neutral-200 shadow-soft overflow-hidden',
  header: 'bg-neutral-50 border-b border-neutral-200',
  headerCell: 'px-6 py-4 text-left text-xs font-semibold text-neutral-700 uppercase tracking-wider',
  row: 'border-b border-neutral-100 hover:bg-neutral-50 transition-colors',
  cell: 'px-6 py-4 text-sm text-neutral-900',
}

export const badgeStyles = {
  default: 'inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-medium bg-neutral-100 text-neutral-700',
  primary: 'inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-medium bg-primary-100 text-primary-700',
  success: 'inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-medium bg-emerald-100 text-emerald-700',
  warning: 'inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-medium bg-amber-100 text-amber-700',
  danger: 'inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-medium bg-red-100 text-red-700',
}
