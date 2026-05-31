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
    ? { label: 'Instrumental', className: 'bg-gray-200 text-gray-600 dark:bg-gray-700 dark:text-gray-300' }
    : track.syncedLyrics
    ? { label: 'Synced', className: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300' }
    : { label: 'Plain', className: 'bg-gray-200 text-gray-600 dark:bg-gray-700 dark:text-gray-300' }

  return (
    <button
      className="flex w-full items-center justify-between rounded-lg bg-gray-100 dark:bg-white/5 px-4 py-3 hover:bg-gray-200 dark:hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors text-left"
      onClick={() => onSelect(track)}
    >
      <div className="min-w-0 flex-1">
        <p className="truncate font-medium text-gray-900 dark:text-white">{track.trackName}</p>
        <p className="truncate text-sm text-gray-600 dark:text-gray-400">
          {track.artistName}{track.albumName ? ` — ${track.albumName}` : ''}
        </p>
      </div>
      <div className="ml-4 flex shrink-0 items-center gap-2">
        <span className={`rounded px-2 py-0.5 text-xs font-medium ${badge.className}`}>
          {badge.label}
        </span>
        <span className="text-xs text-gray-500">{formatDuration(track.duration)}</span>
      </div>
    </button>
  )
}

export type { Track }
