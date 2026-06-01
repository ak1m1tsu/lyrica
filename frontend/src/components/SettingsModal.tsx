import { useEffect, useState } from 'react'
import { GetCloseToTray, SetCloseToTray } from '../../wailsjs/go/main/App'

interface SettingsModalProps {
  open: boolean
  onClose: () => void
}

export function SettingsModal({ open, onClose }: SettingsModalProps) {
  const [closeToTray, setCloseToTray] = useState(false)

  useEffect(() => {
    if (open) GetCloseToTray().then(setCloseToTray)
  }, [open])

  if (!open) return null

  async function handleToggle() {
    const next = !closeToTray
    setCloseToTray(next)
    await SetCloseToTray(next)
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={onClose}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div
        className="relative z-10 w-72 rounded-lg dark:bg-[#161920] bg-[#f5f3ff] shadow-xl border dark:border-white/10 border-[#9B84D1]/20 p-6"
        onClick={e => e.stopPropagation()}
      >
        <button
          onClick={onClose}
          aria-label="Close"
          className="absolute top-3 right-3 flex h-6 w-6 items-center justify-center rounded text-[#0f1117]/40 dark:text-white/40 hover:text-[#0f1117] dark:hover:text-white transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <line x1="1" y1="1" x2="9" y2="9" />
            <line x1="9" y1="1" x2="1" y2="9" />
          </svg>
        </button>

        <div className="flex flex-col gap-4">
          <h2 className="text-lg font-bold tracking-tight dark:text-white text-[#0f1117]">Settings</h2>

          <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
            <span className="text-sm dark:text-white/80 text-[#0f1117]/80">Close to system tray</span>
            <button
              role="switch"
              aria-checked={closeToTray}
              onClick={handleToggle}
              className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none ${
                closeToTray
                  ? 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1]'
                  : 'dark:bg-white/20 bg-[#0f1117]/20'
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform duration-200 ${
                  closeToTray ? 'translate-x-4' : 'translate-x-0'
                }`}
              />
            </button>
          </label>

          <p className="text-xs dark:text-white/40 text-[#0f1117]/40">
            When enabled, closing the window keeps Lyrica running in the system tray.
          </p>
        </div>
      </div>
    </div>
  )
}
