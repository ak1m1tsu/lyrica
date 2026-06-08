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
    ? { label: 'Instrumental', className: 'bg-[var(--color-card-10)] text-[var(--color-text-70)]' }
    : track.syncedLyrics
    ? { label: 'Synced', className: 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] text-white' }
    : { label: 'Plain', className: 'bg-[var(--color-card-10)] text-[var(--color-text-70)]' }

  return (
    <button
      className="flex w-full items-center justify-between rounded-lg bg-[var(--color-card-05)] px-4 py-3 hover:bg-[var(--color-card-10)] focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-accent)] transition-colors text-left"
      onClick={() => onSelect(track)}
    >
      <div className="min-w-0 flex-1">
        <p className="truncate font-medium text-[var(--color-text)]">{track.trackName}</p>
        <p className="truncate text-sm text-[var(--color-text-60)]">
          {track.artistName}{track.albumName ? ` — ${track.albumName}` : ''}
        </p>
      </div>
      <div className="ml-4 flex shrink-0 items-center gap-2">
        <span className={`rounded px-2 py-0.5 text-xs font-medium ${badge.className}`}>
          {badge.label}
        </span>
        <span className="text-xs text-[var(--color-text-50)]">{formatDuration(track.duration)}</span>
      </div>
    </button>
  )
}

export type { Track }
