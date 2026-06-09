import { useState, useEffect, useCallback } from 'react'
import {
  CheckForUpdates,
  DownloadAndInstall,
  GetUpdateAvailable,
  GetAutoUpdateEnabled,
  GetCheckUpdatesEnabled,
  SetAutoUpdateEnabled,
  SetCheckUpdatesEnabled,
  SkipUpdateVersion,
} from '../../wailsjs/go/main/App'
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
  const [dismissed, setDismissed] = useState(false)
  const [autoUpdate, setAutoUpdateState] = useState(true)
  const [checkUpdates, setCheckUpdatesState] = useState(true)

  useEffect(() => {
    // Seed badge state without a network call.
    GetUpdateAvailable().then(available => {
      if (available) {
        CheckForUpdates().then(result => {
          if (result?.available) setUpdateInfo(result)
        }).catch(() => {})
      }
    }).catch(() => {})

    // Seed toggle states from persisted config.
    GetAutoUpdateEnabled().then(setAutoUpdateState).catch(() => {})
    GetCheckUpdatesEnabled().then(setCheckUpdatesState).catch(() => {})

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
      // App relaunches on success — no cleanup needed.
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e))
      setDownloading(false)
      setProgress(null)
    }
  }, [])

  const skipVersion = useCallback(async (version: string) => {
    try {
      await SkipUpdateVersion(version)
      setUpdateInfo(null)
      setDismissed(false)
    } catch {
      // Best-effort; ignore errors.
    }
  }, [])

  const dismiss = useCallback(() => {
    setDismissed(true)
  }, [])

  const setAutoUpdate = useCallback(async (enabled: boolean) => {
    setAutoUpdateState(enabled)
    await SetAutoUpdateEnabled(enabled)
  }, [])

  const setCheckUpdates = useCallback(async (enabled: boolean) => {
    setCheckUpdatesState(enabled)
    await SetCheckUpdatesEnabled(enabled)
  }, [])

  return {
    updateInfo,
    checking,
    downloading,
    progress,
    error,
    dismissed,
    autoUpdate,
    checkUpdates,
    check,
    install,
    skipVersion,
    dismiss,
    setAutoUpdate,
    setCheckUpdates,
  }
}
