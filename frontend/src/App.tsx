import { useState, useRef } from 'react'
import { GetByID, UpdatePresenceIdle, UpdatePresenceSearching, UpdatePresenceTrack } from '../wailsjs/go/main/App'
import { SearchBar } from './components/SearchBar'
import { ResultsList } from './components/ResultsList'
import { LyricsView } from './components/LyricsView'
import { ProgressBar } from './components/ProgressBar'
import { TitleBar } from './components/TitleBar'
import { Track } from './components/TrackCard'
import { useTheme } from './hooks/useTheme'
import { useFavorites } from './hooks/useFavorites'
import { useSearch } from './hooks/useSearch'
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts'
import { FavoritesPanel } from './components/FavoritesPanel'
import { AboutModal } from './components/AboutModal'
import { SettingsModal } from './components/SettingsModal'

type View = 'home' | 'lyrics'

export default function App() {
  const { theme, toggle } = useTheme()
  const { favorites, favoritesDir, isFavorite, toggleFavorite, pickDir } = useFavorites()
  const { results, loading: searchLoading, error: searchError, search: handleSearchBase } = useSearch()
  const [favPanelOpen, setFavPanelOpen] = useState(false)
  const [aboutOpen, setAboutOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [view, setView] = useState<View>('home')
  const [selectedTrack, setSelectedTrack] = useState<Track | null>(null)
  const [lyricsLoading, setLyricsLoading] = useState(false)
  const [lyricsError, setLyricsError] = useState<string | null>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)

  const handleSelect = async (track: Track) => {
    setLyricsLoading(true)
    setLyricsError(null)
    setSelectedTrack(track)
    setView('lyrics')
    try {
      const full = await GetByID(track.id)
      if (full) {
        setSelectedTrack(full)
        UpdatePresenceTrack(full.trackName, full.artistName, !!full.syncedLyrics)
      }
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
    UpdatePresenceIdle()
  }

  const handleSearch = (query: string) => {
    if (query.trim()) {
      UpdatePresenceSearching(query.trim())
    } else {
      UpdatePresenceIdle()
    }
    handleSearchBase(query)
  }

  useKeyboardShortcuts({
    view,
    favPanelOpen,
    aboutOpen,
    settingsOpen,
    searchInputRef,
    onBack: handleBack,
    onToggleFavPanel: () => setFavPanelOpen(o => !o),
    onOpenSettings: () => setSettingsOpen(true),
    onCloseSettings: () => setSettingsOpen(false),
    onCloseAbout: () => setAboutOpen(false),
  })

  return (
    <div className="flex h-screen flex-col dark:bg-[#0f1117] bg-white dark:text-white text-[#0f1117] overflow-hidden">
      <TitleBar theme={theme} toggle={toggle} onFavorites={() => setFavPanelOpen(true)} onAbout={() => setAboutOpen(true)} onSettings={() => setSettingsOpen(true)} />
      <div className="flex-1 overflow-y-auto">
        {view === 'home' && (
          <div className="flex flex-col items-center justify-start px-4 pt-16 gap-6">
            <button onClick={() => setView('home')} className="text-2xl font-bold tracking-tight hover:opacity-80 transition-opacity">
              Lyrica
            </button>
            <SearchBar onSearch={handleSearch} inputRef={searchInputRef} />
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
            isFavorite={isFavorite(selectedTrack.id)}
            onToggleFavorite={() => toggleFavorite(selectedTrack)}
          />
        )}
      </div>
      <FavoritesPanel
        open={favPanelOpen}
        favorites={favorites}
        favoritesDir={favoritesDir}
        onClose={() => setFavPanelOpen(false)}
        onSelect={handleSelect}
        onRemove={toggleFavorite}
        onPickDir={pickDir}
      />
      <AboutModal open={aboutOpen} onClose={() => setAboutOpen(false)} />
      <SettingsModal open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </div>
  )
}
