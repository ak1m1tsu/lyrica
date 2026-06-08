import { useState, useRef } from 'react'
import { GetByID, UpdatePresenceIdle, UpdatePresenceSearching, UpdatePresenceTrack } from '../wailsjs/go/main/App'
import { SearchBar } from './components/SearchBar'
import { ResultsList } from './components/ResultsList'
import { LyricsView } from './components/LyricsView'
import { SettingsView } from './components/SettingsView'
import { ProgressBar } from './components/ProgressBar'
import { TitleBar } from './components/TitleBar'
import { Track } from './components/TrackCard'
import { useTheme } from './hooks/useTheme'
import { useFavorites } from './hooks/useFavorites'
import { useSearch } from './hooks/useSearch'
import { useSpotify } from './hooks/useSpotify'
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts'
import { useUpdater } from './hooks/useUpdater'
import { useGoogleDriveSync } from './hooks/useGoogleDriveSync'
import { FavoritesPanel } from './components/FavoritesPanel'
import { AboutModal } from './components/AboutModal'

type View = 'home' | 'lyrics' | 'settings'

export default function App() {
  const { themeId, themes, setThemeId, addTheme, updateTheme, removeTheme, exportTheme, importTheme } = useTheme()
  const { updateInfo, checking, downloading, progress, error: updateError, check: checkUpdate, install: installUpdate } = useUpdater()
  const { favorites, favoritesDir, isFavorite, toggleFavorite, pickDir, reload: reloadFavorites } = useFavorites()
  const googleDriveSync = useGoogleDriveSync(reloadFavorites)
  const { results, loading: searchLoading, error: searchError, search: handleSearchBase } = useSearch()
  useSpotify((trackName, artistName) => {
    setView('home')
    setSelectedTrack(null)
    handleSearch(`${artistName} ${trackName}`.trim())
  })
  const [favPanelOpen, setFavPanelOpen] = useState(false)
  const [aboutOpen, setAboutOpen] = useState(false)
  const [view, setView] = useState<View>('home')
  const [selectedTrack, setSelectedTrack] = useState<Track | null>(null)
  const [lyricsLoading, setLyricsLoading] = useState(false)
  const [lyricsError, setLyricsError] = useState<string | null>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)
  const previousViewRef = useRef<'home' | 'lyrics'>('home')

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

  const handleBackFromLyrics = () => {
    setView('home')
    setSelectedTrack(null)
    setLyricsError(null)
    UpdatePresenceIdle()
  }

  const handleOpenSettings = () => {
    if (view !== 'settings') {
      previousViewRef.current = view as 'home' | 'lyrics'
    }
    setView('settings')
  }

  const handleBackFromSettings = () => {
    setView(previousViewRef.current)
  }

  const handleBack = () => {
    if (view === 'settings') {
      handleBackFromSettings()
    } else {
      handleBackFromLyrics()
    }
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
    searchInputRef,
    onBack: handleBack,
    onToggleFavPanel: () => setFavPanelOpen(o => !o),
    onOpenSettings: handleOpenSettings,
    onCloseAbout: () => setAboutOpen(false),
  })

  return (
    <div className="flex h-screen flex-col bg-[var(--color-bg)] text-[var(--color-text)] overflow-hidden">
      <TitleBar onFavorites={() => setFavPanelOpen(true)} onAbout={() => setAboutOpen(true)} onSettings={handleOpenSettings} hasUpdate={updateInfo?.available ?? false} />
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
            onBack={handleBackFromLyrics}
            error={lyricsError}
            loading={lyricsLoading}
            isFavorite={isFavorite(selectedTrack.id)}
            onToggleFavorite={() => toggleFavorite(selectedTrack)}
          />
        )}

        {view === 'settings' && (
          <SettingsView
            onBack={handleBackFromSettings}
            hasUpdate={updateInfo?.available ?? false}
            updateInfo={updateInfo}
            checking={checking}
            downloading={downloading}
            downloadProgress={progress}
            updateError={updateError}
            onCheckUpdate={checkUpdate}
            onInstallUpdate={installUpdate}
            googleDrive={googleDriveSync}
            themeId={themeId}
            themes={themes}
            onSetTheme={setThemeId}
            onAddTheme={addTheme}
            onUpdateTheme={updateTheme}
            onRemoveTheme={removeTheme}
            onExportTheme={exportTheme}
            onImportTheme={importTheme}
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
    </div>
  )
}
