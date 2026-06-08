import { useState, useRef, useEffect } from 'react'
import { ExportLyrics } from '../../wailsjs/go/main/App'
import { Track } from './TrackCard'

interface Props {
  track: Track
}

export function ExportMenu({ track }: Props) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  const handle = async (ext: '.lrc' | '.txt') => {
    setOpen(false)
    const text = ext === '.lrc' ? track.syncedLyrics : track.plainLyrics
    const name = `${track.artistName} - ${track.trackName}`
    await ExportLyrics(name, text, ext)
  }

  const hasLrc = Boolean(track.syncedLyrics)
  const hasTxt = Boolean(track.plainLyrics)
  if (!hasLrc && !hasTxt) return null

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(o => !o)}
        className="flex items-center gap-1 rounded px-3 py-1.5 text-sm font-medium
          bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)]
          hover:from-[var(--color-accent-lt-hover)] hover:to-[var(--color-accent-hover)]
          text-white transition-colors"
      >
        Export
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-1 z-50 min-w-[140px] rounded-lg border
          border-[var(--color-accent-20)] bg-[var(--color-surface)] shadow-lg py-1">
          {hasLrc && (
            <button
              onClick={() => handle('.lrc')}
              className="flex w-full items-center gap-2 px-3 py-2 text-sm
                text-[var(--color-text-80)] hover:bg-[var(--color-card-10)] transition-colors"
            >
              <span className="text-[var(--color-accent-on-bg)] font-mono text-xs font-bold">.lrc</span>
              Synced lyrics
            </button>
          )}
          {hasTxt && (
            <button
              onClick={() => handle('.txt')}
              className="flex w-full items-center gap-2 px-3 py-2 text-sm
                text-[var(--color-text-80)] hover:bg-[var(--color-card-10)] transition-colors"
            >
              <span className="text-[var(--color-text-40)] font-mono text-xs font-bold">.txt</span>
              Plain lyrics
            </button>
          )}
        </div>
      )}
    </div>
  )
}
