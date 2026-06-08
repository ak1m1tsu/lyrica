import { useEffect, useRef, useState } from 'react'

interface Option {
  value: string
  label: string
}

interface Props {
  value: string
  options: Option[]
  onChange: (value: string) => void
  className?: string
}

export function Select({ value, options, onChange, className = '' }: Props) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const selected = options.find(o => o.value === value)

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  return (
    <div ref={ref} className={`relative ${className}`}>
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="flex items-center justify-between gap-2 min-w-[140px] rounded-md bg-[var(--color-card-05)] px-2 py-1 text-sm text-[var(--color-text)] ring-1 ring-[var(--color-border)] hover:bg-[var(--color-card-10)] focus:outline-none focus:ring-[var(--color-accent)] transition-all"
      >
        <span>{selected?.label ?? value}</span>
        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-1 z-50 min-w-full rounded-lg border border-[var(--color-accent-20)] bg-[var(--color-surface)] shadow-lg py-1 max-h-48 overflow-y-auto">
          {options.map(opt => (
            <button
              key={opt.value}
              type="button"
              onClick={() => { onChange(opt.value); setOpen(false) }}
              className={`flex w-full items-center gap-2 px-3 py-2 text-sm transition-colors ${
                opt.value === value
                  ? 'text-[var(--color-accent-on-bg)] bg-[var(--color-card-05)]'
                  : 'text-[var(--color-text-80)] hover:bg-[var(--color-card-10)]'
              }`}
            >
              {opt.value === value && (
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
              )}
              {opt.value !== value && <span className="w-[10px]" />}
              {opt.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
