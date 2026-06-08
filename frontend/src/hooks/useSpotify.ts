import { useEffect, useRef, useState } from 'react'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface SpotifyTrackEvent {
  trackName: string
  artistName: string
}

export function useSpotify(onTrack: (trackName: string, artistName: string) => void) {
  const onTrackRef = useRef(onTrack)
  onTrackRef.current = onTrack
  const [tokenExpired, setTokenExpired] = useState(false)

  useEffect(() => {
    const offTrack = EventsOn('spotify:track', (data: SpotifyTrackEvent) => {
      if (data?.trackName) {
        onTrackRef.current(data.trackName, data.artistName ?? '')
      }
    })
    const offExpired = EventsOn('spotify:token_expired', () => setTokenExpired(true))
    const offConnected = EventsOn('spotify:connected', () => setTokenExpired(false))
    return () => { offTrack(); offExpired(); offConnected() }
  }, [])

  return { tokenExpired }
}
