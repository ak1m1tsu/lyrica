import { useEffect, useRef } from 'react'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface SpotifyTrackEvent {
  trackName: string
  artistName: string
}

export function useSpotify(onTrack: (trackName: string, artistName: string) => void) {
  // Always call the latest callback without re-registering the event listener.
  const onTrackRef = useRef(onTrack)
  onTrackRef.current = onTrack

  useEffect(() => {
    const off = EventsOn('spotify:track', (data: SpotifyTrackEvent) => {
      if (data?.trackName) {
        onTrackRef.current(data.trackName, data.artistName ?? '')
      }
    })
    return () => { off() }
  }, [])
}
