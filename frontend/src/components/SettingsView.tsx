import { useEffect, useState } from 'react'
import {
  GetCloseToTray, SetCloseToTray,
  GetDiscordPresence, SetDiscordPresence,
  GetSpotifyEnabled, ConnectSpotify, DisconnectSpotify,
  GetSpotifyAutoSearch, SetSpotifyAutoSearch,
} from '../../wailsjs/go/main/App'
import type { UpdateInfo, DownloadProgress } from '../hooks/useUpdater'
import type { useGoogleDriveSync } from '../hooks/useGoogleDriveSync'
import type { ThemeDefinition } from '../types/theme'
import { BUILT_IN_IDS } from '../types/theme'
import { ThemeList } from './ThemeList'
import { ThemeEditor } from './ThemeEditor'
import { Select } from './Select'

const SECTIONS = ['appearance', 'general', 'connections', 'updates', 'sync'] as const
type Section = typeof SECTIONS[number]

interface Props {
  onBack: () => void
  hasUpdate?: boolean
  updateInfo?: UpdateInfo | null
  checking?: boolean
  downloading?: boolean
  downloadProgress?: DownloadProgress | null
  updateError?: string | null
  autoUpdate?: boolean
  checkUpdates?: boolean
  onCheckUpdate?: () => void
  onInstallUpdate?: () => void
  onSetAutoUpdate?: (enabled: boolean) => void
  onSetCheckUpdates?: (enabled: boolean) => void
  googleDrive?: ReturnType<typeof useGoogleDriveSync>
  spotifyTokenExpired?: boolean
  themeId: string
  themes: ThemeDefinition[]
  onSetTheme: (id: string) => Promise<void>
  onAddTheme: (theme: ThemeDefinition) => Promise<void>
  onUpdateTheme: (theme: ThemeDefinition) => Promise<void>
  onRemoveTheme: (id: string) => Promise<void>
  onExportTheme: (id: string) => Promise<void>
  onImportTheme: () => Promise<void>
}

export function SettingsView({
  onBack,
  hasUpdate, updateInfo, checking, downloading, downloadProgress, updateError,
  autoUpdate, checkUpdates, onSetAutoUpdate, onSetCheckUpdates,
  onCheckUpdate, onInstallUpdate,
  googleDrive,
  spotifyTokenExpired,
  themeId, themes, onSetTheme, onAddTheme, onUpdateTheme, onRemoveTheme, onExportTheme, onImportTheme,
}: Props) {
  const [section, setSection] = useState<Section>('appearance')
  const [closeToTray, setCloseToTray] = useState(false)
  const [discordPresence, setDiscordPresence] = useState(false)
  const [spotifyEnabled, setSpotifyEnabled] = useState(false)
  const [spotifyConnecting, setSpotifyConnecting] = useState(false)
  const [spotifyError, setSpotifyError] = useState('')
  const [spotifyAutoSearch, setSpotifyAutoSearch] = useState(true)
  const [editorOpen, setEditorOpen] = useState(false)
  const [editingTheme, setEditingTheme] = useState<ThemeDefinition | undefined>()

  useEffect(() => {
    GetCloseToTray().then(setCloseToTray)
    GetDiscordPresence().then(setDiscordPresence)
    GetSpotifyEnabled().then(setSpotifyEnabled)
    GetSpotifyAutoSearch().then(setSpotifyAutoSearch)
  }, [])

  async function handleCloseToTrayToggle() {
    const next = !closeToTray
    setCloseToTray(next)
    await SetCloseToTray(next)
  }

  async function handleDiscordToggle() {
    const next = !discordPresence
    setDiscordPresence(next)
    await SetDiscordPresence(next)
  }

  async function handleSpotifyAutoSearchToggle() {
    const next = !spotifyAutoSearch
    setSpotifyAutoSearch(next)
    await SetSpotifyAutoSearch(next)
  }

  async function handleSpotifyConnect() {
    setSpotifyConnecting(true)
    setSpotifyError('')
    try {
      await ConnectSpotify()
      setSpotifyEnabled(true)
    } catch (e: any) {
      setSpotifyError(typeof e === 'string' ? e : (e?.message ?? 'Connection failed'))
    } finally {
      setSpotifyConnecting(false)
    }
  }

  async function handleSpotifyDisconnect() {
    await DisconnectSpotify()
    setSpotifyEnabled(false)
  }

  function openNewTheme() {
    setEditingTheme(undefined)
    setEditorOpen(true)
  }

  function openEditTheme(theme: ThemeDefinition) {
    setEditingTheme(theme)
    setEditorOpen(true)
  }

  async function handleSaveTheme(theme: ThemeDefinition) {
    if (editingTheme) {
      await onUpdateTheme(theme)
    } else {
      await onAddTheme(theme)
    }
  }

  async function handleDeleteTheme(id: string) {
    await onRemoveTheme(id)
    if (themeId === id) await onSetTheme('dark')
  }

  const toggle = (checked: boolean, onChange: () => void) => (
    <button
      role="switch"
      aria-checked={checked}
      onClick={onChange}
      className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none ${
        checked
          ? 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)]'
          : 'bg-[var(--color-card-20)]'
      }`}
    >
      <span className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform duration-200 ${checked ? 'translate-x-4' : 'translate-x-0'}`} />
    </button>
  )

  return (
    <>
      <div className="flex flex-col gap-6 px-4 py-6 max-w-xl mx-auto w-full">
        <button
          onClick={onBack}
          className="flex items-center gap-1 text-sm text-[var(--color-text-60)] hover:text-[var(--color-text)] transition-colors w-fit"
        >
          ← Back
        </button>

        <div className="flex gap-8">
          {/* Sidebar nav */}
          <nav className="w-32 shrink-0 flex flex-col gap-0.5">
            {SECTIONS.map(s => (
              <button
                key={s}
                onClick={() => setSection(s)}
                className={`relative w-full px-3 py-2 text-left text-sm capitalize rounded-md transition-colors ${
                  section === s
                    ? 'bg-[var(--color-card-05)] text-[var(--color-text)]'
                    : 'text-[var(--color-text-40)] hover:text-[var(--color-text-60)]'
                }`}
              >
                {section === s && (
                  <span className="absolute left-0 inset-y-2 w-0.5 rounded-r bg-[var(--color-accent)]" />
                )}
                <span className="relative">
                  {s}
                  {s === 'updates' && hasUpdate && section !== 'updates' && (
                    <span className="absolute -top-0.5 -right-3 h-1.5 w-1.5 rounded-full bg-[var(--color-accent)]" />
                  )}
                </span>
              </button>
            ))}
          </nav>

          {/* Content */}
          <div className="flex-1 min-w-0 flex flex-col gap-4">

            {section === 'appearance' && (
              <>
                <div className="flex items-center justify-between gap-4">
                  <span className="text-sm text-[var(--color-text-80)]">Active theme</span>
                  <Select
                    value={themeId}
                    options={themes.map(t => ({
                      value: t.id,
                      label: t.name + (BUILT_IN_IDS.has(t.id) ? ' (built-in)' : ''),
                    }))}
                    onChange={onSetTheme}
                  />
                </div>

                <hr className="border-[var(--color-border)]" />

                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-[var(--color-text-80)]">Custom themes</span>
                </div>

                <ThemeList
                  themes={themes}
                  activeId={themeId}
                  onEdit={openEditTheme}
                  onExport={onExportTheme}
                  onDelete={handleDeleteTheme}
                />

                <div className="flex gap-2">
                  <button
                    onClick={openNewTheme}
                    className="flex-1 rounded-md py-1.5 text-sm text-white bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:from-[var(--color-accent-lt-hover)] hover:to-[var(--color-accent-hover)] transition-colors"
                  >
                    + New Theme
                  </button>
                  <button
                    onClick={onImportTheme}
                    className="flex-1 rounded-md py-1.5 text-sm bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)] transition-colors"
                  >
                    Import Theme
                  </button>
                </div>
              </>
            )}

            {section === 'general' && (
              <>
                <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
                  <span className="text-sm text-[var(--color-text-80)]">Close to system tray</span>
                  {toggle(closeToTray, handleCloseToTrayToggle)}
                </label>
                <p className="text-xs text-[var(--color-text-40)]">
                  When enabled, closing the window keeps Lyrica running in the system tray.
                </p>
              </>
            )}

            {section === 'connections' && (
              <>
                <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
                  <span className="text-sm text-[var(--color-text-80)]">Discord Rich Presence</span>
                  {toggle(discordPresence, handleDiscordToggle)}
                </label>
                <p className="text-xs text-[var(--color-text-40)]">Show current track on your Discord profile.</p>

                <hr className="border-[var(--color-border)]" />

                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-[var(--color-text-80)]">Spotify</span>
                  <div className="flex items-center gap-1.5">
                    <span className={`h-2 w-2 rounded-full ${
                      spotifyEnabled && spotifyTokenExpired ? 'bg-amber-400' :
                      spotifyEnabled ? 'bg-green-400' : 'bg-[var(--color-card-20)]'
                    }`} />
                    <span className="text-xs text-[var(--color-text-40)]">
                      {spotifyEnabled && spotifyTokenExpired ? 'Session expired' :
                       spotifyEnabled ? 'Connected' : 'Not connected'}
                    </span>
                  </div>
                </div>

                {spotifyEnabled && spotifyTokenExpired ? (
                  <>
                    <div className="rounded-md bg-amber-500/10 px-3 py-2.5 flex flex-col gap-1">
                      <p className="text-xs text-amber-400 font-medium">Session expired</p>
                      <p className="text-xs text-amber-400/70">Your Spotify authorization has expired. Reconnect to continue auto-searching.</p>
                    </div>
                    <button
                      onClick={handleSpotifyConnect}
                      disabled={spotifyConnecting}
                      className={`w-full rounded-md py-1.5 text-sm text-white transition-colors ${
                        spotifyConnecting ? 'bg-[var(--color-accent-20)] cursor-not-allowed' : 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:opacity-90'
                      }`}
                    >
                      {spotifyConnecting ? 'Reconnecting…' : 'Reconnect to Spotify'}
                    </button>
                    <hr className="border-[var(--color-border)]" />
                    <button
                      onClick={handleSpotifyDisconnect}
                      className="w-full rounded-md py-1.5 text-sm bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)] transition-colors"
                    >
                      Disconnect
                    </button>
                  </>
                ) : spotifyEnabled ? (
                  <>
                    <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
                      <span className="text-sm text-[var(--color-text-80)]">Auto-search on track change</span>
                      {toggle(spotifyAutoSearch, handleSpotifyAutoSearchToggle)}
                    </label>
                    <p className="text-xs text-[var(--color-text-40)]">
                      Automatically search for lyrics when the track changes in Spotify.
                    </p>
                    <hr className="border-[var(--color-border)]" />
                    <button
                      onClick={handleSpotifyDisconnect}
                      className="w-full rounded-md py-1.5 text-sm bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)] transition-colors"
                    >
                      Disconnect
                    </button>
                  </>
                ) : (
                  <button
                    onClick={handleSpotifyConnect}
                    disabled={spotifyConnecting}
                    className={`w-full rounded-md py-1.5 text-sm text-white transition-colors ${
                      spotifyConnecting
                        ? 'bg-[var(--color-accent-20)] cursor-not-allowed'
                        : 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:opacity-90'
                    }`}
                  >
                    {spotifyConnecting ? 'Connecting…' : 'Connect to Spotify'}
                  </button>
                )}
                {spotifyError && <p className="text-xs text-red-400">{spotifyError}</p>}
              </>
            )}

            {section === 'updates' && (
              <>
                <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
                  <span className="text-sm text-[var(--color-text-80)]">Auto-update</span>
                  {toggle(autoUpdate ?? true, () => onSetAutoUpdate?.(!(autoUpdate ?? true)))}
                </label>
                <p className="text-xs text-[var(--color-text-40)] -mt-2">
                  Automatically download and apply updates on startup.
                </p>

                <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
                  <span className="text-sm text-[var(--color-text-80)]">Check for updates</span>
                  {toggle(checkUpdates ?? true, () => onSetCheckUpdates?.(!(checkUpdates ?? true)))}
                </label>
                <p className="text-xs text-[var(--color-text-40)] -mt-2">
                  Show a notification when a new version is available.
                </p>

                <hr className="border-[var(--color-border)]" />

                {downloading ? (
                  <>
                    <p className="text-sm text-[var(--color-text-80)]">Downloading update…</p>
                    <div className="w-full h-1.5 rounded-full bg-[var(--color-card-10)] overflow-hidden">
                      {downloadProgress && downloadProgress.total > 0 ? (
                        <div
                          className="h-full rounded-full bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] transition-all duration-150"
                          style={{ width: `${Math.round((downloadProgress.received / downloadProgress.total) * 100)}%` }}
                        />
                      ) : (
                        <div className="h-full w-1/3 bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] animate-progress" />
                      )}
                    </div>
                    {downloadProgress && downloadProgress.total > 0 && (
                      <p className="text-xs text-[var(--color-text-40)]">
                        {Math.round(downloadProgress.received / 1024 / 1024 * 10) / 10} MB / {Math.round(downloadProgress.total / 1024 / 1024 * 10) / 10} MB
                      </p>
                    )}
                  </>
                ) : updateError ? (
                  <>
                    <p className="text-sm text-red-400">{updateError}</p>
                    <button onClick={onCheckUpdate} className="w-full rounded-md py-1.5 text-sm bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)] transition-colors">
                      Retry
                    </button>
                  </>
                ) : updateInfo?.available ? (
                  <>
                    <p className="text-sm text-[var(--color-text-80)]">
                      Lyrica <span className="font-semibold text-[var(--color-accent-on-bg)]">v{updateInfo.latestVersion}</span> is available
                    </p>
                    <button onClick={onInstallUpdate} className="w-full rounded-md py-1.5 text-sm text-white bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:opacity-90 transition-opacity">
                      Download &amp; Install
                    </button>
                  </>
                ) : (
                  <>
                    <p className="text-sm text-[var(--color-text-60)]">
                      {updateInfo ? "You're up to date." : 'Check for the latest version of Lyrica.'}
                    </p>
                    <button
                      onClick={onCheckUpdate}
                      disabled={checking}
                      className={`w-full rounded-md py-1.5 text-sm transition-colors ${
                        checking
                          ? 'bg-[var(--color-card-05)] text-[var(--color-text-30)] cursor-not-allowed'
                          : 'bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)]'
                      }`}
                    >
                      {checking ? 'Checking…' : 'Check for Updates'}
                    </button>
                  </>
                )}
              </>
            )}

            {section === 'sync' && googleDrive && (
              <>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-[var(--color-text-80)]">Google Drive</span>
                  <div className="flex items-center gap-1.5">
                    <span className={`h-2 w-2 rounded-full ${
                      googleDrive.enabled && googleDrive.tokenExpired ? 'bg-amber-400' :
                      googleDrive.enabled ? 'bg-green-400' : 'bg-[var(--color-card-20)]'
                    }`} />
                    <span className="text-xs text-[var(--color-text-40)]">
                      {googleDrive.enabled && googleDrive.tokenExpired ? 'Session expired' :
                       googleDrive.enabled ? 'Connected' : 'Not connected'}
                    </span>
                  </div>
                </div>

                {googleDrive.enabled && googleDrive.tokenExpired ? (
                  <>
                    <div className="rounded-md bg-amber-500/10 px-3 py-2.5 flex flex-col gap-1">
                      <p className="text-xs text-amber-400 font-medium">Session expired</p>
                      <p className="text-xs text-amber-400/70">Your Google authorization has been revoked or expired. Reconnect to continue syncing.</p>
                    </div>
                    <button
                      onClick={googleDrive.connect}
                      disabled={googleDrive.connecting}
                      className={`w-full rounded-md py-1.5 text-sm text-white transition-colors ${
                        googleDrive.connecting ? 'bg-[var(--color-accent-20)] cursor-not-allowed' : 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:opacity-90'
                      }`}
                    >
                      {googleDrive.connecting ? 'Reconnecting…' : 'Reconnect to Google Drive'}
                    </button>
                    <hr className="border-[var(--color-border)]" />
                    <button
                      onClick={googleDrive.disconnect}
                      className="w-full rounded-md py-1.5 text-sm bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)] transition-colors"
                    >
                      Disconnect
                    </button>
                  </>
                ) : googleDrive.enabled ? (
                  <>
                    <button
                      onClick={googleDrive.syncTo}
                      disabled={googleDrive.syncing}
                      className={`w-full rounded-md py-1.5 text-sm text-white transition-colors ${
                        googleDrive.syncing ? 'bg-[var(--color-accent-20)] cursor-not-allowed' : 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:opacity-90'
                      }`}
                    >
                      {googleDrive.syncing ? 'Syncing…' : 'Upload to Drive'}
                    </button>
                    <button
                      onClick={googleDrive.syncFrom}
                      disabled={googleDrive.syncing}
                      className={`w-full rounded-md py-1.5 text-sm transition-colors ${
                        googleDrive.syncing
                          ? 'bg-[var(--color-card-05)] text-[var(--color-text-30)] cursor-not-allowed'
                          : 'bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)]'
                      }`}
                    >
                      {googleDrive.syncing ? 'Syncing…' : 'Download from Drive'}
                    </button>
                    {googleDrive.lastSyncTime && (
                      <p className="text-xs text-[var(--color-text-40)]">
                        Last synced: {new Date(googleDrive.lastSyncTime).toLocaleString()}
                      </p>
                    )}
                    <hr className="border-[var(--color-border)]" />
                    <button
                      onClick={googleDrive.disconnect}
                      className="w-full rounded-md py-1.5 text-sm bg-[var(--color-card-10)] text-[var(--color-text-70)] hover:bg-[var(--color-card-20)] transition-colors"
                    >
                      Disconnect
                    </button>
                  </>
                ) : (
                  <>
                    <p className="text-xs text-[var(--color-text-40)]">
                      Sync your favorites list across devices using your Google Drive app data folder.
                    </p>
                    <button
                      onClick={googleDrive.connect}
                      disabled={googleDrive.connecting}
                      className={`w-full rounded-md py-1.5 text-sm text-white transition-colors ${
                        googleDrive.connecting ? 'bg-[var(--color-accent-20)] cursor-not-allowed' : 'bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:opacity-90'
                      }`}
                    >
                      {googleDrive.connecting ? 'Connecting…' : 'Connect to Google Drive'}
                    </button>
                  </>
                )}

                {googleDrive.error && (
                  <p className="rounded-md bg-red-500/10 px-3 py-2 text-xs text-red-400">{googleDrive.error}</p>
                )}
              </>
            )}

          </div>
        </div>
      </div>

      {editorOpen && (
        <ThemeEditor
          initial={editingTheme}
          onSave={handleSaveTheme}
          onClose={() => setEditorOpen(false)}
        />
      )}
    </>
  )
}
