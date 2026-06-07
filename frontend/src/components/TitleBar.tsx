import { WindowMinimise, WindowToggleMaximise } from '../../wailsjs/runtime/runtime'
import { CloseApp } from '../../wailsjs/go/main/App'
import { ThemeToggle } from './ThemeToggle'

interface TitleBarProps {
  theme: 'dark' | 'light'
  toggle: () => void
  onFavorites: () => void
  onAbout: () => void
  onSettings: () => void
  hasUpdate?: boolean
}

export function TitleBar({ theme, toggle, onFavorites, onAbout, onSettings, hasUpdate }: TitleBarProps) {
  return (
    <div
      className="flex h-9 shrink-0 items-center justify-between dark:bg-[#0f1117] bg-white px-3 select-none"
      style={{ '--wails-draggable': 'drag' } as React.CSSProperties}
    >
      <span className="text-xs font-semibold tracking-widest dark:text-white/30 text-[#0f1117]/40 uppercase">Lyrica</span>

      <div
        className="flex items-center"
        style={{ '--wails-draggable': 'no-drag' } as React.CSSProperties}
      >
        <button
          onClick={onFavorites}
          aria-label="Open favorites"
          className="flex h-9 w-11 items-center justify-center text-[#0f1117]/40 dark:text-white/40 hover:bg-[#0f1117]/10 dark:hover:bg-white/10 hover:text-red-500 dark:hover:text-red-400 transition-colors"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
          </svg>
        </button>
        <ThemeToggle theme={theme} toggle={toggle} />

        <button
          onClick={onSettings}
          aria-label="Settings"
          className="relative flex h-9 w-11 items-center justify-center text-[#0f1117]/40 dark:text-white/40 dark:hover:bg-white/10 hover:bg-[#0f1117]/10 dark:hover:text-white hover:text-[#0f1117] transition-colors"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
          {hasUpdate && (
            <span className="absolute top-2 right-2 h-1.5 w-1.5 rounded-full bg-[#9B84D1]" />
          )}
        </button>

        <button
          onClick={onAbout}
          aria-label="About"
          className="flex h-9 w-11 items-center justify-center text-[#0f1117]/40 dark:text-white/40 dark:hover:bg-white/10 hover:bg-[#0f1117]/10 dark:hover:text-white hover:text-[#0f1117] transition-colors text-xs font-bold"
        >
          ?
        </button>

        <button
          onClick={WindowMinimise}
          aria-label="Minimize"
          className="flex h-9 w-11 items-center justify-center text-[#0f1117]/40 dark:text-white/40 dark:hover:bg-white/10 hover:bg-[#0f1117]/10 dark:hover:text-white hover:text-[#0f1117] transition-colors"
        >
          <svg width="10" height="1" viewBox="0 0 10 1" fill="currentColor">
            <rect width="10" height="1" />
          </svg>
        </button>

        <button
          onClick={WindowToggleMaximise}
          aria-label="Maximize"
          className="flex h-9 w-11 items-center justify-center text-[#0f1117]/40 dark:text-white/40 dark:hover:bg-white/10 hover:bg-[#0f1117]/10 dark:hover:text-white hover:text-[#0f1117] transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="1">
            <rect x="0.5" y="0.5" width="9" height="9" />
          </svg>
        </button>

        <button
          onClick={CloseApp}
          aria-label="Close"
          className="flex h-9 w-11 items-center justify-center text-gray-500 hover:bg-red-600 hover:text-white transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round">
            <line x1="1" y1="1" x2="9" y2="9" />
            <line x1="9" y1="1" x2="1" y2="9" />
          </svg>
        </button>
      </div>
    </div>
  )
}
