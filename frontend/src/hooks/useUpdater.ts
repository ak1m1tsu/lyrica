import { useState, useEffect, useCallback } from 'react'
import { CheckForUpdates, DownloadAndInstall, GetUpdateAvailable } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { main } from '../../wailsjs/go/models'

export type UpdateInfo = main.UpdateResult

export interface DownloadProgress {
  received: number
  total: number
}

export function useUpdater() {
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null)
  const [checking, setChecking] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [progress, setProgress] = useState<DownloadProgress | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    // Seed badge state without a network call; if update is already known,
    // also fetch full details for display.
    GetUpdateAvailable().then(available => {
      if (available) {
        CheckForUpdates().then(result => {
          if (result?.available) setUpdateInfo(result)
        }).catch(() => {})
      }
    }).catch(() => {})

    // Listen for the startup background check result.
    const offAvailable = EventsOn('update:available', (result: UpdateInfo) => {
      if (result?.available) setUpdateInfo(result)
    })

    // Listen for download progress.
    const offProgress = EventsOn('update:progress', (data: DownloadProgress) => {
      setProgress(data)
    })

    return () => {
      offAvailable()
      offProgress()
    }
  }, [])

  const check = useCallback(async () => {
    setChecking(true)
    setError(null)
    try {
      const result = await CheckForUpdates()
      setUpdateInfo(result ?? null)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setChecking(false)
    }
  }, [])

  const install = useCallback(async () => {
    setDownloading(true)
    setProgress(null)
    setError(null)
    try {
      await DownloadAndInstall()
      // App quits on success — no cleanup needed.
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e))
      setDownloading(false)
      setProgress(null)
    }
  }, [])

  return { updateInfo, checking, downloading, progress, error, check, install }
}
