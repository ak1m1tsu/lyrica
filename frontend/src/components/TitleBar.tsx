import { WindowMinimise, WindowToggleMaximise, Quit } from '../../wailsjs/runtime/runtime'
import { ThemeToggle } from './ThemeToggle'

interface TitleBarProps {
  theme: 'dark' | 'light'
  toggle: () => void
  onFavorites: () => void
  onAbout: () => void
}

export function TitleBar({ theme, toggle, onFavorites, onAbout }: TitleBarProps) {
  return (
    <div
      className="flex h-9 shrink-0 items-center justify-between dark:bg-[#0f1117] bg-gray-100 px-3 select-none"
      style={{ '--wails-draggable': 'drag' } as React.CSSProperties}
    >
      <span className="text-xs font-semibold tracking-widest dark:text-gray-600 text-gray-400 uppercase">lrclib</span>

      <div
        className="flex items-center"
        style={{ '--wails-draggable': 'no-drag' } as React.CSSProperties}
      >
        <button
          onClick={onFavorites}
          aria-label="Open favorites"
          className="flex h-9 w-11 items-center justify-center text-gray-500 hover:bg-gray-200 dark:hover:bg-white/10 hover:text-red-500 dark:hover:text-red-400 transition-colors"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
          </svg>
        </button>
        <ThemeToggle theme={theme} toggle={toggle} />

        <button
          onClick={onAbout}
          aria-label="About"
          className="flex h-9 w-11 items-center justify-center text-gray-500 dark:hover:bg-white/10 hover:bg-gray-200 hover:text-white transition-colors text-xs font-bold"
        >
          ?
        </button>

        <button
          onClick={WindowMinimise}
          aria-label="Minimize"
          className="flex h-9 w-11 items-center justify-center text-gray-500 dark:hover:bg-white/10 hover:bg-gray-200 hover:text-white transition-colors"
        >
          <svg width="10" height="1" viewBox="0 0 10 1" fill="currentColor">
            <rect width="10" height="1" />
          </svg>
        </button>

        <button
          onClick={WindowToggleMaximise}
          aria-label="Maximize"
          className="flex h-9 w-11 items-center justify-center text-gray-500 dark:hover:bg-white/10 hover:bg-gray-200 hover:text-white transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="1">
            <rect x="0.5" y="0.5" width="9" height="9" />
          </svg>
        </button>

        <button
          onClick={Quit}
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
