import { useState } from 'react'
import { Search, GetByID } from '../wailsjs/go/main/App'
import { SearchBar } from './components/SearchBar'
import { ResultsList } from './components/ResultsList'
import { LyricsView } from './components/LyricsView'
import { ProgressBar } from './components/ProgressBar'
import { TitleBar } from './components/TitleBar'
import { Track } from './components/TrackCard'
import { useTheme } from './hooks/useTheme'

type View = 'home' | 'lyrics'

export default function App() {
  const { theme, toggle } = useTheme()
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
    } catch (e: unknown) {
      setSearchError(e instanceof Error ? e.message : String(e))
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
    } catch (e: unknown) {
      setLyricsError(e instanceof Error ? e.message : String(e))
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
    <div className="flex h-screen flex-col dark:bg-[#0f1117] bg-gray-50 dark:text-white text-gray-900 overflow-hidden">
      <TitleBar theme={theme} toggle={toggle} />
      <div className="flex-1 overflow-y-auto">
        {view === 'home' && (
          <div className="flex flex-col items-center justify-start px-4 pt-16 gap-6">
            <button onClick={() => setView('home')} className="text-2xl font-bold tracking-tight hover:opacity-80 transition-opacity">
              lrclib
            </button>
            <SearchBar onSearch={handleSearch} />
            {searchLoading && <ProgressBar />}
            {searchError && (
              <p className="text-sm text-red-400">{searchError}</p>
            )}
            <div className="w-full max-w-xl">
              <ResultsList tracks={results} onSelect={handleSelect} />
            </div>
          </div>
        )}

        {view === 'lyrics' && selectedTrack && (
          <LyricsView
            track={selectedTrack}
            onBack={handleBack}
            error={lyricsError}
            loading={lyricsLoading}
          />
        )}
      </div>
    </div>
  )
}
