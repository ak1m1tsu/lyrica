import type { ThemeDefinition } from '../types/theme'
import { BUILT_IN_IDS } from '../types/theme'

interface Props {
  themes: ThemeDefinition[]
  activeId: string
  onEdit: (theme: ThemeDefinition) => void
  onExport: (id: string) => void
  onDelete: (id: string) => void
}

export function ThemeList({ themes, activeId, onEdit, onExport, onDelete }: Props) {
  const customThemes = themes.filter(t => !BUILT_IN_IDS.has(t.id))

  if (customThemes.length === 0) {
    return (
      <p className="text-xs text-[var(--color-text-40)]">No custom themes yet. Create one below.</p>
    )
  }

  return (
    <ul className="flex flex-col gap-1">
      {customThemes.map(theme => (
        <li
          key={theme.id}
          className={`flex items-center gap-2 rounded-md px-2 py-2 ${
            activeId === theme.id ? 'bg-[var(--color-card-10)]' : ''
          }`}
        >
          <div className="flex items-center gap-1.5 shrink-0">
            {(['background', 'surface', 'accent', 'accentLight', 'text'] as const).map(key => (
              <span
                key={key}
                className="h-3 w-3 rounded-full border border-[var(--color-border)]"
                style={{ backgroundColor: theme.colors[key] }}
                title={key}
              />
            ))}
          </div>
          <span className="flex-1 min-w-0 truncate text-sm text-[var(--color-text-80)]">{theme.name}</span>
          <div className="flex items-center gap-1 shrink-0">
            <button
              onClick={() => onEdit(theme)}
              className="rounded px-1.5 py-0.5 text-xs text-[var(--color-text-50)] hover:bg-[var(--color-card-10)] hover:text-[var(--color-text)] transition-colors"
            >
              Edit
            </button>
            <button
              onClick={() => onExport(theme.id)}
              className="rounded px-1.5 py-0.5 text-xs text-[var(--color-text-50)] hover:bg-[var(--color-card-10)] hover:text-[var(--color-text)] transition-colors"
            >
              Export
            </button>
            <button
              onClick={() => onDelete(theme.id)}
              className="rounded px-1.5 py-0.5 text-xs text-[var(--color-text-50)] hover:bg-[var(--color-card-10)] hover:text-red-400 transition-colors"
            >
              Delete
            </button>
          </div>
        </li>
      ))}
    </ul>
  )
}
