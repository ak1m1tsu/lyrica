import { useEffect, useState } from 'react'
import {
  GetCloseToTray, SetCloseToTray,
  GetDiscordPresence, SetDiscordPresence,
  GetSpotifyEnabled, ConnectSpotify, DisconnectSpotify,
} from '../../wailsjs/go/main/App'

interface SettingsModalProps {
  open: boolean
  onClose: () => void
}

export function SettingsModal({ open, onClose }: SettingsModalProps) {
  const [tab, setTab] = useState<'general' | 'connections'>('general')
  const [closeToTray, setCloseToTray] = useState(false)
  const [discordPresence, setDiscordPresence] = useState(false)
  const [spotifyEnabled, setSpotifyEnabled] = useState(false)
  const [spotifyConnecting, setSpotifyConnecting] = useState(false)
  const [spotifyError, setSpotifyError] = useState('')

  useEffect(() => {
    if (open) {
      GetCloseToTray().then(setCloseToTray)
      GetDiscordPresence().then(setDiscordPresence)
      GetSpotifyEnabled().then(setSpotifyEnabled)
    }
  }, [open])

  if (!open) return null

  async function handleToggle() {
    const next = !closeToTray
    setCloseToTray(next)
    await SetCloseToTray(next)
  }

  async function handleDiscordToggle() {
    const next = !discordPresence
    setDiscordPresence(next)
    await SetDiscordPresence(next)
  }

  async function handleConnect() {
    setSpotifyConnecting(true)
    setSpotifyError('')
    try {
      await ConnectSpotify()
      setSpotifyEnabled(true)
    } catch (e: any) {
      setSpotifyError(e?.message ?? 'Connection failed')
    } finally {
      setSpotifyConnecting(false)
    }
  }

  async function handleDisconnect() {
    await DisconnectSpotify()
    setSpotifyEnabled(false)
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={onClose}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div
        className="relative z-10 w-80 rounded-lg dark:bg-[#161920] bg-[#f5f3ff] shadow-xl border dark:border-white/10 border-[#9B84D1]/20 p-6"
        onClick={e => e.stopPropagation()}
      >
        <button
          onClick={onClose}
          aria-label="Close"
          className="absolute top-3 right-3 flex h-6 w-6 items-center justify-center rounded text-[#0f1117]/40 dark:text-white/40 hover:text-[#0f1117] dark:hover:text-white transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <line x1="1" y1="1" x2="9" y2="9" />
            <line x1="9" y1="1" x2="1" y2="9" />
          </svg>
        </button>

        <h2 className="text-lg font-bold tracking-tight dark:text-white text-[#0f1117]">Settings</h2>

        <div className="flex gap-1 border-b dark:border-white/10 border-[#0f1117]/10 mb-4 mt-3">
          {(['general', 'connections'] as const).map(t => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-3 py-1.5 text-sm capitalize transition-colors ${
                tab === t
                  ? 'dark:text-white text-[#0f1117] border-b-2 border-[#9B84D1] -mb-px'
                  : 'dark:text-white/40 text-[#0f1117]/40 hover:dark:text-white/70 hover:text-[#0f1117]/70'
              }`}
            >
              {t}
            </button>
          ))}
        </div>

        {tab === 'general' && (
          <div className="flex flex-col gap-4">
            <label className="flex items-center justify-between gap-4 cursor-pointer select-none">
              <span className="text-sm dark:text-white/80 text-[#0f1117]/80">Close to system tray</span>
              <button
                role="switch"
                aria-checked={closeToTray}
                onClick={handleToggle}
                className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none ${
                  closeToTray
                    ? 'bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1]'
                    : 'dark:bg-white/20 bg-[#0f1117]/20'
                }`}
              >
                <span
                  className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform duration-200 ${
                    closeToTray ? 'translate-x-4' : 'translate-x-0'
                  }`}
                />
              </button>
            </label>

            <p className="text-xs dark:text-white/40 text-[#0f1117]/40">
              When enabled, closing the window keeps Lyrica running in the system tray.
            </p>
          </div>
        )}

        {tab === 'connections' && (
          <div className="flex flex-col gap-4">
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
                <span
                  className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform duration-200 ${
                    discordPresence ? 'translate-x-4' : 'translate-x-0'
                  }`}
                />
              </button>
            </label>

            <p className="text-xs dark:text-white/40 text-[#0f1117]/40">
              Show current track on your Discord profile.
            </p>

            <hr className="dark:border-white/10 border-[#0f1117]/10" />

            {/* Spotify header */}
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium dark:text-white/80 text-[#0f1117]/80">Spotify</span>
              <div className="flex items-center gap-1.5">
                <span className={`h-2 w-2 rounded-full ${spotifyEnabled ? 'bg-green-400' : 'dark:bg-white/20 bg-[#0f1117]/20'}`} />
                <span className="text-xs dark:text-white/40 text-[#0f1117]/40">
                  {spotifyEnabled ? 'Connected' : 'Not connected'}
                </span>
              </div>
            </div>

            {/* Connect/Disconnect button */}
            {spotifyEnabled ? (
              <button
                onClick={handleDisconnect}
                className="w-full rounded-md py-1.5 text-sm dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20 transition-colors"
              >
                Disconnect
              </button>
            ) : (
              <button
                onClick={handleConnect}
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

            {spotifyError && (
              <p className="text-xs text-red-400">{spotifyError}</p>
            )}

          </div>
        )}
      </div>
    </div>
  )
}
