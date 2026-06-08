import { useEffect, useState } from 'react'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import { GetVersion, GetAppIcon } from '../../wailsjs/go/main/App'

interface AboutModalProps {
  open: boolean
  onClose: () => void
}

export function AboutModal({ open, onClose }: AboutModalProps) {
  const [version, setVersion] = useState('')
  const [icon, setIcon] = useState('')

  useEffect(() => {
    if (open) {
      GetVersion().then(setVersion)
      GetAppIcon().then(setIcon)
    }
  }, [open])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={onClose}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div
        className="relative z-10 w-72 rounded-lg bg-[var(--color-surface)] shadow-xl border border-[var(--color-accent-20)] p-6"
        onClick={e => e.stopPropagation()}
      >
        <button
          onClick={onClose}
          aria-label="Close"
          className="absolute top-3 right-3 flex h-6 w-6 items-center justify-center rounded text-[var(--color-text-40)] hover:text-[var(--color-text)] transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <line x1="1" y1="1" x2="9" y2="9" />
            <line x1="9" y1="1" x2="1" y2="9" />
          </svg>
        </button>

        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-3">
            {icon && <img src={icon} alt="Lyrica" className="w-12 h-12 rounded-xl" />}
            <div>
              <h2 className="text-lg font-bold tracking-tight text-[var(--color-text)]">Lyrica</h2>
              <p className="text-xs text-[var(--color-text-60)] mt-0.5">Version {version}</p>
            </div>
          </div>

          <div className="flex flex-col gap-1.5 text-sm">
            <div className="flex justify-between">
              <span className="text-[var(--color-text-60)]">Author</span>
              <span className="text-[var(--color-text)] font-medium">ak1m1tsu</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-[var(--color-text-60)]">Source</span>
              <button
                onClick={() => BrowserOpenURL('https://github.com/ak1m1tsu/lyrica')}
                className="text-[var(--color-accent-on-bg)] hover:text-[var(--color-accent-on-bg-hover)] font-medium transition-colors"
              >
                github.com/ak1m1tsu/lyrica
              </button>
            </div>
          </div>

          <p className="text-xs text-[var(--color-text-40)]">
            Lyrics browser powered by LRCLib.net
          </p>
        </div>
      </div>
    </div>
  )
}
