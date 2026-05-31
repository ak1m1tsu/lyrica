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
    <ul className="flex flex-col gap-2">
      {tracks.map(t => (
        <li key={t.id}>
          <TrackCard track={t} onSelect={onSelect} />
        </li>
      ))}
    </ul>
  )
}
