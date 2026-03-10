import { clsx } from 'clsx'
import type { InputHTMLAttributes } from 'react'

interface Props extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export function Input({ label, error, className, ...props }: Props) {
  return (
    <div className="flex flex-col gap-1">
      {label && <label className="text-sm text-slate-400">{label}</label>}
      <input
        className={clsx(
          'bg-slate-700 border border-slate-600 rounded-lg px-3 py-2 text-slate-100 placeholder-slate-500 focus:outline-none focus:border-blue-500 transition-colors',
          error && 'border-red-500',
          className,
        )}
        {...props}
      />
      {error && <span className="text-sm text-red-400">{error}</span>}
    </div>
  )
}
