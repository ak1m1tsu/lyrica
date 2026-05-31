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

function formatDuration(seconds: number): string {
  const s = Math.floor(seconds)
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
}

export function TrackCard({ track, onSelect }: Props) {
  const badge = track.instrumental
    ? { label: 'Instrumental', className: 'bg-gray-700 text-gray-300' }
    : track.syncedLyrics
    ? { label: 'Synced', className: 'bg-blue-900 text-blue-300' }
    : { label: 'Plain', className: 'bg-gray-700 text-gray-300' }

  return (
    <div
      className="flex items-center justify-between rounded-lg bg-white/5 px-4 py-3 hover:bg-white/10 cursor-pointer transition-colors"
      onClick={() => onSelect(track)}
    >
      <div className="min-w-0 flex-1">
        <p className="truncate font-medium text-white">{track.trackName}</p>
        <p className="truncate text-sm text-gray-400">
          {track.artistName}{track.albumName ? ` — ${track.albumName}` : ''}
        </p>
      </div>
      <div className="ml-4 flex shrink-0 items-center gap-2">
        <span className={`rounded px-2 py-0.5 text-xs font-medium ${badge.className}`}>
          {badge.label}
        </span>
        <span className="text-xs text-gray-500">{formatDuration(track.duration)}</span>
      </div>
    </div>
  )
}

export type { Track }
