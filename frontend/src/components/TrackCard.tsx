import { formatDuration } from '../utils/formatting'

interface Track {
  id: number
  trackName: string
  artistName: string
  albumName: string
  duration: number
  instrumental: boolean
  plainLyrics: string
  syncedLyrics: string
}

interface Props {
  track: Track
  onSelect: (track: Track) => void
}

export function TrackCard({ track, onSelect }: Props) {
  const badge = track.instrumental
    ? { label: 'Instrumental', className: 'bg-[#0f1117]/10 text-[#0f1117]/70 dark:bg-white/10 dark:text-white/70' }
    : track.syncedLyrics
    ? { label: 'Synced', className: 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] text-white' }
    : { label: 'Plain', className: 'bg-[#0f1117]/10 text-[#0f1117]/70 dark:bg-white/10 dark:text-white/70' }

  return (
    <button
      className="flex w-full items-center justify-between rounded-lg bg-[#0f1117]/5 dark:bg-white/5 px-4 py-3 hover:bg-[#0f1117]/10 dark:hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-[#9B84D1] transition-colors text-left"
      onClick={() => onSelect(track)}
    >
      <div className="min-w-0 flex-1">
        <p className="truncate font-medium text-[#0f1117] dark:text-white">{track.trackName}</p>
        <p className="truncate text-sm text-[#0f1117]/60 dark:text-white/60">
          {track.artistName}{track.albumName ? ` — ${track.albumName}` : ''}
        </p>
      </div>
      <div className="ml-4 flex shrink-0 items-center gap-2">
        <span className={`rounded px-2 py-0.5 text-xs font-medium ${badge.className}`}>
          {badge.label}
        </span>
        <span className="text-xs text-[#0f1117]/50 dark:text-white/50">{formatDuration(track.duration)}</span>
      </div>
    </button>
  )
}

export type { Track }
