import { useEffect, useState } from 'react'
import {
  GetCloseToTray, SetCloseToTray,
  GetDiscordPresence, SetDiscordPresence,
  GetSpotifyEnabled, ConnectSpotify, DisconnectSpotify,
} from '../../wailsjs/go/main/App'
import type { UpdateInfo, DownloadProgress } from '../hooks/useUpdater'
import type { useGoogleDriveSync } from '../hooks/useGoogleDriveSync'

const SECTIONS = ['general', 'connections', 'updates', 'sync'] as const
type Section = typeof SECTIONS[number]

interface Props {
  onBack: () => void
  hasUpdate?: boolean
  updateInfo?: UpdateInfo | null
  checking?: boolean
  downloading?: boolean
  downloadProgress?: DownloadProgress | null
  updateError?: string | null
  onCheckUpdate?: () => void
  onInstallUpdate?: () => void
  googleDrive?: ReturnType<typeof useGoogleDriveSync>
}

export function SettingsView({
  onBack,
  hasUpdate, updateInfo, checking, downloading, downloadProgress, updateError,
  onCheckUpdate, onInstallUpdate,
  googleDrive,
}: Props) {
  const [section, setSection] = useState<Section>('general')
  const [closeToTray, setCloseToTray] = useState(false)
  const [discordPresence, setDiscordPresence] = useState(false)
  const [spotifyEnabled, setSpotifyEnabled] = useState(false)
  const [spotifyConnecting, setSpotifyConnecting] = useState(false)
  const [spotifyError, setSpotifyError] = useState('')

  useEffect(() => {
    GetCloseToTray().then(setCloseToTray)
    GetDiscordPresence().then(setDiscordPresence)
    GetSpotifyEnabled().then(setSpotifyEnabled)
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

  return (
    <div className="flex flex-col gap-6 px-4 py-6 max-w-xl mx-auto w-full">
      <button
        onClick={onBack}
        className="flex items-center gap-1 text-sm text-[#0f1117]/60 dark:text-white/60 hover:text-[#0f1117] dark:hover:text-white transition-colors w-fit"
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
                  ? 'dark:bg-white/5 bg-[#0f1117]/5 dark:text-white text-[#0f1117]'
                  : 'dark:text-white/40 text-[#0f1117]/40 hover:dark:text-white/60 hover:text-[#0f1117]/60'
              }`}
            >
              {section === s && (
                <span className="absolute left-0 inset-y-2 w-0.5 rounded-r bg-[#9B84D1]" />
              )}
              <span className="relative">
                {s}
                {s === 'updates' && hasUpdate && section !== 'updates' && (
                  <span className="absolute -top-0.5 -right-3 h-1.5 w-1.5 rounded-full bg-[#9B84D1]" />
                )}
              </span>
            </button>
          ))}
        </nav>

        {/* Content */}
        <div className="flex-1 min-w-0 flex flex-col gap-4">

          {section === 'general' && (
            <>
              <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
                <span className="text-sm dark:text-white/80 text-[#0f1117]/80">Close to system tray</span>
                <button
                  role="switch"
                  aria-checked={closeToTray}
                  onClick={handleCloseToTrayToggle}
                  className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none ${
                    closeToTray
                      ? 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1]'
                      : 'dark:bg-white/20 bg-[#0f1117]/20'
                  }`}
                >
                  <span className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform duration-200 ${closeToTray ? 'translate-x-4' : 'translate-x-0'}`} />
                </button>
              </label>
              <p className="text-xs dark:text-white/40 text-[#0f1117]/40">
                When enabled, closing the window keeps Lyrica running in the system tray.
              </p>
            </>
          )}

          {section === 'connections' && (
            <>
              <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
                <span className="text-sm dark:text-white/80 text-[#0f1117]/80">Discord Rich Presence</span>
                <button
                  role="switch"
                  aria-checked={discordPresence}
                  onClick={handleDiscordToggle}
                  className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none ${
                    discordPresence
                      ? 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1]'
                      : 'dark:bg-white/20 bg-[#0f1117]/20'
                  }`}
                >
                  <span className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform duration-200 ${discordPresence ? 'translate-x-4' : 'translate-x-0'}`} />
                </button>
              </label>
              <p className="text-xs dark:text-white/40 text-[#0f1117]/40">Show current track on your Discord profile.</p>

              <hr className="dark:border-white/10 border-[#0f1117]/10" />

              <div className="flex items-center justify-between">
                <span className="text-sm font-medium dark:text-white/80 text-[#0f1117]/80">Spotify</span>
                <div className="flex items-center gap-1.5">
                  <span className={`h-2 w-2 rounded-full ${spotifyEnabled ? 'bg-green-400' : 'dark:bg-white/20 bg-[#0f1117]/20'}`} />
                  <span className="text-xs dark:text-white/40 text-[#0f1117]/40">
                    {spotifyEnabled ? 'Connected' : 'Not connected'}
                  </span>
                </div>
              </div>

              {spotifyEnabled ? (
                <button
                  onClick={handleSpotifyDisconnect}
                  className="w-full rounded-md py-1.5 text-sm dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20 transition-colors"
                >
                  Disconnect
                </button>
              ) : (
                <button
                  onClick={handleSpotifyConnect}
                  disabled={spotifyConnecting}
                  className={`w-full rounded-md py-1.5 text-sm text-white transition-colors ${
                    spotifyConnecting
                      ? 'bg-[#9B84D1]/50 cursor-not-allowed'
                      : 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] hover:opacity-90'
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
              {downloading ? (
                <>
                  <p className="text-sm dark:text-white/80 text-[#0f1117]/80">Downloading update…</p>
                  <div className="w-full h-1.5 rounded-full dark:bg-white/10 bg-[#0f1117]/10 overflow-hidden">
                    {downloadProgress && downloadProgress.total > 0 ? (
                      <div
                        className="h-full rounded-full bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] transition-all duration-150"
                        style={{ width: `${Math.round((downloadProgress.received / downloadProgress.total) * 100)}%` }}
                      />
                    ) : (
                      <div className="h-full w-1/3 bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] animate-progress" />
                    )}
                  </div>
                  {downloadProgress && downloadProgress.total > 0 && (
                    <p className="text-xs dark:text-white/40 text-[#0f1117]/40">
                      {Math.round(downloadProgress.received / 1024 / 1024 * 10) / 10} MB / {Math.round(downloadProgress.total / 1024 / 1024 * 10) / 10} MB
                    </p>
                  )}
                </>
              ) : updateError ? (
                <>
                  <p className="text-sm text-red-400">{updateError}</p>
                  <button onClick={onCheckUpdate} className="w-full rounded-md py-1.5 text-sm dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20 transition-colors">
                    Retry
                  </button>
                </>
              ) : updateInfo?.available ? (
                <>
                  <p className="text-sm dark:text-white/80 text-[#0f1117]/80">
                    Lyrica <span className="font-semibold text-[#9B84D1]">v{updateInfo.latestVersion}</span> is available
                  </p>
                  {updateInfo.releaseNotes && (
                    <pre className="text-xs dark:text-white/50 text-[#0f1117]/50 whitespace-pre-wrap overflow-y-auto max-h-28 rounded dark:bg-white/5 bg-[#0f1117]/5 p-2">
                      {updateInfo.releaseNotes}
                    </pre>
                  )}
                  <button onClick={onInstallUpdate} className="w-full rounded-md py-1.5 text-sm text-white bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] hover:opacity-90 transition-opacity">
                    Download &amp; Install
                  </button>
                </>
              ) : (
                <>
                  <p className="text-sm dark:text-white/60 text-[#0f1117]/60">
                    {updateInfo ? "You're up to date." : 'Check for the latest version of Lyrica.'}
                  </p>
                  <button
                    onClick={onCheckUpdate}
                    disabled={checking}
                    className={`w-full rounded-md py-1.5 text-sm transition-colors ${
                      checking
                        ? 'dark:bg-white/5 bg-[#0f1117]/5 dark:text-white/30 text-[#0f1117]/30 cursor-not-allowed'
                        : 'dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20'
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
                <span className="text-sm font-medium dark:text-white/80 text-[#0f1117]/80">Google Drive</span>
                <div className="flex items-center gap-1.5">
                  <span className={`h-2 w-2 rounded-full ${
                    googleDrive.enabled && googleDrive.tokenExpired ? 'bg-amber-400' :
                    googleDrive.enabled ? 'bg-green-400' : 'dark:bg-white/20 bg-[#0f1117]/20'
                  }`} />
                  <span className="text-xs dark:text-white/40 text-[#0f1117]/40">
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
                      googleDrive.connecting ? 'bg-[#9B84D1]/50 cursor-not-allowed' : 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] hover:opacity-90'
                    }`}
                  >
                    {googleDrive.connecting ? 'Reconnecting…' : 'Reconnect to Google Drive'}
                  </button>
                  <hr className="dark:border-white/10 border-[#0f1117]/10" />
                  <button
                    onClick={googleDrive.disconnect}
                    className="w-full rounded-md py-1.5 text-sm dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20 transition-colors"
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
                      googleDrive.syncing ? 'bg-[#9B84D1]/50 cursor-not-allowed' : 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] hover:opacity-90'
                    }`}
                  >
                    {googleDrive.syncing ? 'Syncing…' : 'Upload to Drive'}
                  </button>
                  <button
                    onClick={googleDrive.syncFrom}
                    disabled={googleDrive.syncing}
                    className={`w-full rounded-md py-1.5 text-sm transition-colors ${
                      googleDrive.syncing
                        ? 'dark:bg-white/5 bg-[#0f1117]/5 dark:text-white/30 text-[#0f1117]/30 cursor-not-allowed'
                        : 'dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20'
                    }`}
                  >
                    {googleDrive.syncing ? 'Syncing…' : 'Download from Drive'}
                  </button>
                  {googleDrive.lastSyncTime && (
                    <p className="text-xs dark:text-white/40 text-[#0f1117]/40">
                      Last synced: {new Date(googleDrive.lastSyncTime).toLocaleString()}
                    </p>
                  )}
                  <hr className="dark:border-white/10 border-[#0f1117]/10" />
                  <button
                    onClick={googleDrive.disconnect}
                    className="w-full rounded-md py-1.5 text-sm dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20 transition-colors"
                  >
                    Disconnect
                  </button>
                </>
              ) : (
                <>
                  <p className="text-xs dark:text-white/40 text-[#0f1117]/40">
                    Sync your favorites list across devices using your Google Drive app data folder.
                  </p>
                  <button
                    onClick={googleDrive.connect}
                    disabled={googleDrive.connecting}
                    className={`w-full rounded-md py-1.5 text-sm text-white transition-colors ${
                      googleDrive.connecting ? 'bg-[#9B84D1]/50 cursor-not-allowed' : 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] hover:opacity-90'
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
  )
}
