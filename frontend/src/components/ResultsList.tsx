import { Track, TrackCard } from './TrackCard'
import { EmptyState } from './EmptyState'

interface Props {
  tracks: Track[]
  onSelect: (track: Track) => void
}

export function ResultsList({ tracks, onSelect }: Props) {
  if (tracks.length === 0) {
    return <EmptyState message="No results found." />
  }
  return (
    <div className="flex flex-col gap-2">
      {tracks.map(t => (
        <TrackCard key={t.id} track={t} onSelect={onSelect} />
      ))}
    </div>
  )
}
