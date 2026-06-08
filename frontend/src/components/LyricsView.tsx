import { useState } from 'react'
import { Track } from './TrackCard'
import { EmptyState } from './EmptyState'
import { ErrorBlock } from './ErrorBlock'
import { ProgressBar } from './ProgressBar'
import { formatDuration } from '../utils/formatting'
import { FavoriteButton } from './FavoriteButton'
import { ExportMenu } from './ExportMenu'

interface Props {
  track: Track
  onBack: () => void
  error?: string | null
  loading?: boolean
  isFavorite?: boolean
  onToggleFavorite?: () => Promise<void>
}

function extractTimestamp(line: string): string {
  if (line.startsWith('[')) {
    const end = line.indexOf(']')
    if (end > 1) return line.slice(0, end + 1)
  }
  return ''
}

function extractLyricText(line: string): string {
  if (line.startsWith('[')) {
    const end = line.indexOf(']')
    if (end > 0 && end + 1 < line.length) return line.slice(end + 1).trim()
  }
  return line
}

export function LyricsView({ track, onBack, error, loading, isFavorite = false, onToggleFavorite }: Props) {
  const [plain, setPlain] = useState(false)

  const renderLyrics = () => {
    if (loading) return <ProgressBar />
    if (error) return <ErrorBlock message={error} />
    if (track.instrumental) return <EmptyState message="This track is instrumental." />

    const text = plain ? track.plainLyrics : track.syncedLyrics
    if (!text) {
      if (!plain && track.plainLyrics) {
        return (
          <div>
            <p className="mb-3 text-sm text-yellow-600 dark:text-yellow-500">Synced lyrics unavailable — showing plain.</p>
            {renderPlainText(track.plainLyrics)}
          </div>
        )
      }
      return <EmptyState message="No lyrics available for this track." />
    }

    return plain ? renderPlainText(text) : renderSyncedText(text)
  }

  const renderPlainText = (text: string) => (
    <div className="space-y-1">
      {text.split('\n').map((line, i) => (
        <p key={i} className="text-[var(--color-text-90)]">{line || <br />}</p>
      ))}
    </div>
  )

  const renderSyncedText = (text: string) => (
    <div className="space-y-1">
      {text.split('\n').map((line, i) => {
        const ts = extractTimestamp(line)
        const lyric = extractLyricText(line)
        return (
          <div key={i} className="flex gap-3">
            {ts && <span className="shrink-0 text-right text-xs text-[var(--color-accent-on-bg)] w-16 pt-0.5">{ts}</span>}
            <span className="text-[var(--color-text-90)]">{lyric}</span>
          </div>
        )
      })}
    </div>
  )

  return (
    <div className="flex flex-col gap-4 px-4 py-6 max-w-2xl mx-auto">
      <button
        onClick={onBack}
        className="flex items-center gap-1 text-sm text-[var(--color-text-60)] hover:text-[var(--color-text)] transition-colors w-fit"
      >
        ← Back to search
      </button>

      <div>
        <div className="flex items-start justify-between gap-2">
          <h1 className="text-xl font-semibold text-[var(--color-text)]">{track.trackName}</h1>
          <div className="flex shrink-0 items-center gap-1">
            {onToggleFavorite && (
              <FavoriteButton isFavorite={isFavorite} onToggle={onToggleFavorite} />
            )}
            <ExportMenu track={track} />
          </div>
        </div>
        <p className="text-[var(--color-text-60)]">
          {track.artistName}
          {track.albumName ? ` — ${track.albumName}` : ''}
          {track.duration ? ` · ${formatDuration(track.duration)}` : ''}
        </p>
      </div>

      {!track.instrumental && (
        <div className="flex gap-2">
          <button
            aria-pressed={!plain}
            onClick={() => setPlain(false)}
            className={`rounded px-3 py-1 text-sm font-medium transition-colors ${
              !plain
                ? 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] text-white'
                : 'bg-[var(--color-card-10)] text-[var(--color-text-60)] hover:bg-[var(--color-card-20)]'
            }`}
          >
            Synced
          </button>
          <button
            aria-pressed={plain}
            onClick={() => setPlain(true)}
            className={`rounded px-3 py-1 text-sm font-medium transition-colors ${
              plain
                ? 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] text-white'
                : 'bg-[var(--color-card-10)] text-[var(--color-text-60)] hover:bg-[var(--color-card-20)]'
            }`}
          >
            Plain
          </button>
        </div>
      )}

      <div className="lyrics-scroll max-h-[60vh] overflow-y-auto rounded-lg bg-[var(--color-card-05)] p-4 font-mono text-sm">
        {renderLyrics()}
      </div>
    </div>
  )
}
