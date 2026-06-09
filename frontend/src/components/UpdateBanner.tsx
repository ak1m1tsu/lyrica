import { useState } from 'react'
import type { UpdateInfo, DownloadProgress } from '../hooks/useUpdater'

interface UpdateBannerProps {
  updateInfo: UpdateInfo | null
  dismissed: boolean
  downloading: boolean
  progress: DownloadProgress | null
  onInstall: () => void
  onSkip: (version: string) => void
  onDismiss: () => void
}

export function UpdateBanner({
  updateInfo,
  dismissed,
  downloading,
  progress,
  onInstall,
  onSkip,
  onDismiss,
}: UpdateBannerProps) {
  const [confirmingSkip, setConfirmingSkip] = useState(false)

  if (!updateInfo?.available || dismissed) return null

  const percent = progress && progress.total > 0
    ? Math.round((progress.received / progress.total) * 100)
    : null

  const receivedMB = progress ? Math.round(progress.received / 1024 / 1024 * 10) / 10 : 0
  const totalMB = progress && progress.total > 0 ? Math.round(progress.total / 1024 / 1024 * 10) / 10 : 0

  if (downloading) {
    return (
      <div className="fixed bottom-0 inset-x-0 z-40 flex items-center gap-3 px-4 py-2.5 dark:bg-[#161920] bg-[#f5f3ff] border-t dark:border-white/10 border-[#9B84D1]/20 select-none">
        <span className="text-xs dark:text-white/60 text-[#0f1117]/60 shrink-0">Downloading update…</span>
        <div className="flex-1 h-1 rounded-full dark:bg-white/10 bg-[#0f1117]/10 overflow-hidden">
          {percent !== null ? (
            <div
              className="h-full rounded-full bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] transition-all duration-150"
              style={{ width: `${percent}%` }}
            />
          ) : (
            <div className="h-full w-1/3 rounded-full bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] animate-progress" />
          )}
        </div>
        {percent !== null && (
          <span className="text-xs dark:text-white/40 text-[#0f1117]/40 shrink-0 tabular-nums">
            {percent}% &nbsp;{receivedMB} / {totalMB} MB
          </span>
        )}
      </div>
    )
  }

  if (confirmingSkip) {
    return (
      <div className="fixed bottom-0 inset-x-0 z-40 flex items-center justify-between gap-3 px-4 py-2.5 dark:bg-[#161920] bg-[#f5f3ff] border-t dark:border-white/10 border-[#9B84D1]/20 select-none">
        <span className="text-xs dark:text-white/60 text-[#0f1117]/60">
          Skip v{updateInfo.latestVersion}?
        </span>
        <div className="flex items-center gap-2">
          <button
            onClick={() => onSkip(updateInfo.latestVersion)}
            className="rounded px-2.5 py-1 text-xs dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20 transition-colors"
          >
            Skip this version
          </button>
          <button
            onClick={() => { setConfirmingSkip(false); onDismiss() }}
            className="rounded px-2.5 py-1 text-xs dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20 transition-colors"
          >
            Remind me later
          </button>
          <button
            onClick={() => setConfirmingSkip(false)}
            aria-label="Cancel"
            className="flex h-5 w-5 items-center justify-center rounded dark:text-white/40 text-[#0f1117]/40 hover:dark:text-white hover:text-[#0f1117] transition-colors"
          >
            <svg width="8" height="8" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
              <line x1="1" y1="1" x2="9" y2="9" />
              <line x1="9" y1="1" x2="1" y2="9" />
            </svg>
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed bottom-0 inset-x-0 z-40 flex items-center justify-between gap-3 px-4 py-2.5 dark:bg-[#161920] bg-[#f5f3ff] border-t dark:border-white/10 border-[#9B84D1]/20 select-none">
      <span className="text-xs dark:text-white/60 text-[#0f1117]/60">
        Lyrica <span className="font-semibold text-[#9B84D1]">v{updateInfo.latestVersion}</span> is available
      </span>
      <div className="flex items-center gap-2">
        <button
          onClick={onInstall}
          className="rounded px-2.5 py-1 text-xs text-white bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] hover:opacity-90 transition-opacity"
        >
          Update
        </button>
        <button
          onClick={() => setConfirmingSkip(true)}
          className="rounded px-2.5 py-1 text-xs dark:bg-white/10 bg-[#0f1117]/10 dark:text-white/70 text-[#0f1117]/70 hover:dark:bg-white/20 hover:bg-[#0f1117]/20 transition-colors"
        >
          Skip
        </button>
        <button
          onClick={onDismiss}
          aria-label="Dismiss"
          className="flex h-5 w-5 items-center justify-center rounded dark:text-white/40 text-[#0f1117]/40 hover:dark:text-white hover:text-[#0f1117] transition-colors"
        >
          <svg width="8" height="8" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <line x1="1" y1="1" x2="9" y2="9" />
            <line x1="9" y1="1" x2="1" y2="9" />
          </svg>
        </button>
      </div>
    </div>
  )
}
