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
          <div className="flex items-center gap-3">
            {icon && <img src={icon} alt="Lyrica" className="w-12 h-12 rounded-xl" />}
            <div>
              <h2 className="text-lg font-bold tracking-tight dark:text-white text-[#0f1117]">Lyrica</h2>
              <p className="text-xs dark:text-white/60 text-[#0f1117]/60 mt-0.5">Version {version}</p>
            </div>
          </div>

          <div className="flex flex-col gap-1.5 text-sm">
            <div className="flex justify-between">
              <span className="dark:text-white/60 text-[#0f1117]/60">Author</span>
              <span className="dark:text-white text-[#0f1117] font-medium">ak1m1tsu</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="dark:text-white/60 text-[#0f1117]/60">Source</span>
              <button
                onClick={() => BrowserOpenURL('https://github.com/ak1m1tsu/lyrica')}
                className="text-[#9B84D1] dark:text-[#C8B1F3] hover:text-[#7B64B1] dark:hover:text-[#D4C0F5] font-medium transition-colors"
              >
                github.com/ak1m1tsu/lyrica
              </button>
            </div>
          </div>

          <p className="text-xs dark:text-white/40 text-[#0f1117]/40">
            Lyrics browser powered by LRCLib.net
          </p>
        </div>
      </div>
    </div>
  )
}
