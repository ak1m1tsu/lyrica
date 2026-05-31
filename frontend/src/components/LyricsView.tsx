import { useState } from 'react'
import { Track } from './TrackCard'
import { EmptyState } from './EmptyState'
import { ErrorBlock } from './ErrorBlock'

interface Props {
  track: Track
  onBack: () => void
  error?: string | null
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

function formatDuration(seconds: number): string {
  const s = Math.floor(seconds)
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
}

export function LyricsView({ track, onBack, error }: Props) {
  const [plain, setPlain] = useState(false)

  const renderLyrics = () => {
    if (error) return <ErrorBlock message={error} />
    if (track.instrumental) return <EmptyState message="This track is instrumental." />

    const text = plain ? track.plainLyrics : track.syncedLyrics
    if (!text) {
      if (!plain && track.plainLyrics) {
        return (
          <div>
            <p className="mb-3 text-sm text-yellow-500">Synced lyrics unavailable — showing plain.</p>
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
        <p key={i} className="text-gray-200">{line || <br />}</p>
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
            {ts && <span className="shrink-0 text-right text-xs text-blue-400 w-16 pt-0.5">{ts}</span>}
            <span className="text-gray-200">{lyric}</span>
          </div>
        )
      })}
    </div>
  )

  return (
    <div className="flex flex-col gap-4 px-4 py-6 max-w-2xl mx-auto">
      <button
        onClick={onBack}
        className="flex items-center gap-1 text-sm text-gray-400 hover:text-white transition-colors w-fit"
      >
        ← Back to search
      </button>

      <div>
        <h1 className="text-xl font-semibold text-white">{track.trackName}</h1>
        <p className="text-gray-400">
          {track.artistName}
          {track.albumName ? ` — ${track.albumName}` : ''}
          {track.duration ? ` · ${formatDuration(track.duration)}` : ''}
        </p>
      </div>

      {!track.instrumental && (
        <div className="flex gap-2">
          <button
            onClick={() => setPlain(false)}
            className={`rounded px-3 py-1 text-sm font-medium transition-colors ${
              !plain ? 'bg-blue-600 text-white' : 'bg-white/10 text-gray-400 hover:bg-white/20'
            }`}
          >
            Synced
          </button>
          <button
            onClick={() => setPlain(true)}
            className={`rounded px-3 py-1 text-sm font-medium transition-colors ${
              plain ? 'bg-blue-600 text-white' : 'bg-white/10 text-gray-400 hover:bg-white/20'
            }`}
          >
            Plain
          </button>
        </div>
      )}

      <div className="max-h-[60vh] overflow-y-auto rounded-lg bg-white/5 p-4 font-mono text-sm">
        {renderLyrics()}
      </div>
    </div>
  )
}
