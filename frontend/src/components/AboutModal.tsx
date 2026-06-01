import { useEffect, useState } from 'react'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import { GetVersion } from '../../wailsjs/go/main/App'

interface AboutModalProps {
  open: boolean
  onClose: () => void
}

export function AboutModal({ open, onClose }: AboutModalProps) {
  const [version, setVersion] = useState('')

  useEffect(() => {
    if (open) GetVersion().then(setVersion)
  }, [open])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={onClose}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div
        className="relative z-10 w-72 rounded-lg dark:bg-[#1a1d27] bg-white shadow-xl border dark:border-white/10 border-gray-200 p-6"
        onClick={e => e.stopPropagation()}
      >
        <button
          onClick={onClose}
          aria-label="Close"
          className="absolute top-3 right-3 flex h-6 w-6 items-center justify-center rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <line x1="1" y1="1" x2="9" y2="9" />
            <line x1="9" y1="1" x2="1" y2="9" />
          </svg>
        </button>

        <div className="flex flex-col gap-4">
          <div>
            <h2 className="text-lg font-bold tracking-tight dark:text-white text-gray-900">Lyrica</h2>
            <p className="text-xs dark:text-gray-400 text-gray-500 mt-0.5">Version {version}</p>
          </div>

          <div className="flex flex-col gap-1.5 text-sm">
            <div className="flex justify-between">
              <span className="dark:text-gray-400 text-gray-500">Author</span>
              <span className="dark:text-gray-200 text-gray-700 font-medium">ak1m1tsu</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="dark:text-gray-400 text-gray-500">Source</span>
              <button
                onClick={() => BrowserOpenURL('https://github.com/ak1m1tsu/lyrica')}
                className="text-blue-500 hover:text-blue-400 font-medium transition-colors"
              >
                github.com/ak1m1tsu/lyrica
              </button>
            </div>
          </div>

          <p className="text-xs dark:text-gray-500 text-gray-400">
            Lyrics browser powered by LRCLib.net
          </p>
        </div>
      </div>
    </div>
  )
}
