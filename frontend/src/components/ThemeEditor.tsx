import { useState } from 'react'
import type { ThemeDefinition, ThemeColors } from '../types/theme'

interface Props {
  initial?: ThemeDefinition
  onSave: (theme: ThemeDefinition) => Promise<void>
  onClose: () => void
}

function nameToId(name: string): string {
  return name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '').slice(0, 64) || 'custom'
}

const COLOR_FIELDS: { key: keyof ThemeColors; label: string }[] = [
  { key: 'background', label: 'Background' },
  { key: 'surface',    label: 'Surface' },
  { key: 'text',       label: 'Text' },
  { key: 'accent',     label: 'Accent' },
  { key: 'accentLight',label: 'Accent Light' },
]

export function ThemeEditor({ initial, onSave, onClose }: Props) {
  const isNew = !initial
  const [name, setName] = useState(initial?.name ?? '')
  const [colors, setColors] = useState<ThemeColors>(
    initial?.colors ?? {
      background: '#0f1117',
      surface: '#161920',
      text: '#ffffff',
      accent: '#9B84D1',
      accentLight: '#C8B1F3',
    }
  )
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const id = initial?.id ?? nameToId(name)

  const handleSave = async () => {
    if (!name.trim()) { setError('Name is required.'); return }
    const generatedId = isNew ? nameToId(name) : initial!.id
    if (!generatedId) { setError('Name must produce a valid ID.'); return }
    setSaving(true)
    setError('')
    try {
      await onSave({ id: generatedId, name: name.trim(), colors })
      onClose()
    } catch (e: any) {
      setError(typeof e === 'string' ? e : (e?.message ?? 'Failed to save theme.'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" onClick={onClose}>
      <div className="absolute inset-0 bg-black/50" />
      <div
        className="relative z-10 w-80 rounded-lg bg-[var(--color-surface)] shadow-xl border border-[var(--color-accent-20)] p-6 flex flex-col gap-4"
        onClick={e => e.stopPropagation()}
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-[var(--color-text)]">
            {isNew ? 'New theme' : 'Edit theme'}
          </h2>
          <button
            onClick={onClose}
            aria-label="Close"
            className="flex h-6 w-6 items-center justify-center rounded text-[var(--color-text-40)] hover:text-[var(--color-text)] transition-colors"
          >
            <svg width="10" height="10" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
              <line x1="1" y1="1" x2="9" y2="9" /><line x1="9" y1="1" x2="1" y2="9" />
            </svg>
          </button>
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-xs text-[var(--color-text-60)]">Name</label>
          <input
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="My theme"
            className="rounded-md bg-[var(--color-card-05)] px-3 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-text-40)] outline-none ring-1 ring-[var(--color-border)] focus:ring-[var(--color-accent)] transition-all"
          />
          {isNew && name && (
            <p className="text-xs text-[var(--color-text-40)]">ID: {nameToId(name)}</p>
          )}
        </div>

        <div className="flex flex-col gap-2">
          <p className="text-xs text-[var(--color-text-60)]">Colors</p>
          {COLOR_FIELDS.map(({ key, label }) => (
            <div key={key} className="flex items-center justify-between">
              <span className="text-sm text-[var(--color-text-80)]">{label}</span>
              <div className="flex items-center gap-2">
                <span className="text-xs text-[var(--color-text-50)] font-mono">{colors[key]}</span>
                <input
                  type="color"
                  value={colors[key]}
                  onChange={e => setColors(prev => ({ ...prev, [key]: e.target.value }))}
                  className="h-7 w-10 cursor-pointer rounded border border-[var(--color-border)] bg-transparent"
                />
              </div>
            </div>
          ))}
        </div>

        <div
          className="rounded-md p-3 flex items-center gap-3"
          style={{ backgroundColor: colors.surface }}
        >
          <div className="w-4 h-4 rounded-full" style={{ background: `linear-gradient(to right, ${colors.accentLight}, ${colors.accent})` }} />
          <span className="text-sm font-medium" style={{ color: colors.text }}>Preview text</span>
          <span className="text-xs ml-auto" style={{ color: colors.accentLight }}>accent</span>
        </div>

        {error && <p className="text-xs text-red-400">{error}</p>}

        <div className="flex gap-2">
          <button
            onClick={onClose}
            className="flex-1 rounded-md py-1.5 text-sm bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)] transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className={`flex-1 rounded-md py-1.5 text-sm text-white transition-colors ${
              saving
                ? 'bg-[var(--color-accent-20)] cursor-not-allowed'
                : 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:from-[var(--color-accent-lt-hover)] hover:to-[var(--color-accent-hover)]'
            }`}
          >
            {saving ? 'Saving…' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  )
}
