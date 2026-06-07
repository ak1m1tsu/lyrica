import { useEffect, useState } from 'react'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import {
  GetGoogleDriveEnabled,
  GetLastSyncTime,
  ConnectGoogleDrive,
  DisconnectGoogleDrive,
  SyncToGoogleDrive,
  SyncFromGoogleDrive,
} from '../../wailsjs/go/main/App'

export function useGoogleDriveSync(onPulled?: () => void) {
  const [enabled, setEnabled] = useState(false)
  const [connecting, setConnecting] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [lastSyncTime, setLastSyncTime] = useState('')
  const [error, setError] = useState('')
  const [tokenExpired, setTokenExpired] = useState(false)

  useEffect(() => {
    GetGoogleDriveEnabled().then(setEnabled)
    GetLastSyncTime().then(setLastSyncTime)
    EventsOn('gdrive:synced', (t: string) => setLastSyncTime(t))
    return () => { EventsOff('gdrive:synced') }
  }, [])

  function handleSyncError(e: any, fallback: string) {
    const msg: string = typeof e === 'string' ? e : (e?.message ?? fallback)
    if (msg === 'token_expired') {
      setTokenExpired(true)
    } else {
      setError(msg || fallback)
    }
  }

  async function connect() {
    setConnecting(true)
    setError('')
    try {
      await ConnectGoogleDrive()
      setEnabled(true)
      setTokenExpired(false)
    } catch (e: any) {
      setError(typeof e === 'string' ? e : (e?.message ?? 'Connection failed'))
    } finally {
      setConnecting(false)
    }
  }

  async function disconnect() {
    setError('')
    try {
      await DisconnectGoogleDrive()
      setEnabled(false)
      setTokenExpired(false)
      setLastSyncTime('')
    } catch (e: any) {
      setError(typeof e === 'string' ? e : (e?.message ?? 'Disconnect failed'))
    }
  }

  async function syncTo() {
    setSyncing(true)
    setError('')
    try {
      await SyncToGoogleDrive()
    } catch (e: any) {
      handleSyncError(e, 'Upload failed')
    } finally {
      setSyncing(false)
    }
  }

  async function syncFrom() {
    setSyncing(true)
    setError('')
    try {
      await SyncFromGoogleDrive()
      onPulled?.()
    } catch (e: any) {
      handleSyncError(e, 'Download failed')
    } finally {
      setSyncing(false)
    }
  }

  return { enabled, connecting, syncing, lastSyncTime, error, tokenExpired, connect, disconnect, syncTo, syncFrom }
}
