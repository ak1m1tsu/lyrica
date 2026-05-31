import { useState } from 'react'
import { Search, GetByID } from '../wailsjs/go/main/App'
import { SearchBar } from './components/SearchBar'
import { ResultsList } from './components/ResultsList'
import { LyricsView } from './components/LyricsView'
import { Track } from './components/TrackCard'

type View = 'home' | 'lyrics'

export default function App() {
  const [view, setView] = useState<View>('home')
  const [results, setResults] = useState<Track[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [searchError, setSearchError] = useState<string | null>(null)
  const [selectedTrack, setSelectedTrack] = useState<Track | null>(null)
  const [lyricsLoading, setLyricsLoading] = useState(false)
  const [lyricsError, setLyricsError] = useState<string | null>(null)

  const handleSearch = async (query: string) => {
    if (!query.trim()) {
      setResults([])
      setSearchError(null)
      return
    }
    setSearchLoading(true)
    setSearchError(null)
    try {
      const tracks = await Search(query)
      setResults(tracks ?? [])
    } catch (e: any) {
      setSearchError(e?.toString() ?? 'Search failed.')
    } finally {
      setSearchLoading(false)
    }
  }

  const handleSelect = async (track: Track) => {
    setLyricsLoading(true)
    setLyricsError(null)
    setSelectedTrack(track)
    setView('lyrics')
    try {
      const full = await GetByID(track.id)
      if (full) setSelectedTrack(full)
    } catch (e: any) {
      setLyricsError(e?.toString() ?? 'Failed to load lyrics.')
    } finally {
      setLyricsLoading(false)
    }
  }

  const handleBack = () => {
    setView('home')
    setSelectedTrack(null)
    setLyricsError(null)
  }

  return (
    <div className="min-h-screen bg-[#0f1117] text-white">
      {view === 'home' && (
        <div className="flex flex-col items-center justify-start px-4 pt-16 gap-6">
          <a href="#" onClick={e => { e.preventDefault(); setView('home') }} className="text-2xl font-bold tracking-tight">
            lrclib
          </a>
          <SearchBar onSearch={handleSearch} loading={searchLoading} />
          {searchError && (
            <p className="text-sm text-red-400">{searchError}</p>
          )}
          <div className="w-full max-w-xl">
            <ResultsList tracks={results} onSelect={handleSelect} />
          </div>
        </div>
      )}

      {view === 'lyrics' && selectedTrack && (
        lyricsLoading ? (
          <div className="flex items-center justify-center min-h-screen">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-white/20 border-t-white" />
          </div>
        ) : (
          <LyricsView
            track={selectedTrack}
            onBack={handleBack}
            error={lyricsError}
          />
        )
      )}
    </div>
  )
}
