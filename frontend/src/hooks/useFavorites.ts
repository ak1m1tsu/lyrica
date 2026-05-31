import { useState, useEffect, useCallback, useRef } from 'react'
import {
  GetFavorites,
  AddFavorite,
  RemoveFavorite,
  GetFavoritesDir,
  PickFavoritesDir,
} from '../../wailsjs/go/main/App'
import { Track } from '../components/TrackCard'

export function useFavorites() {
  const [favorites, setFavorites] = useState<Track[]>([])
  const [favoritesDir, setFavoritesDir] = useState('')
  // Mirror of `favorites` read synchronously inside toggleFavorite. A state
  // updater is not guaranteed to run before setFavorites() returns, so we
  // cannot rely on a functional updater to decide the add/remove branch for
  // rapid successive toggles — the ref gives us a consistent, immediate view.
  const favoritesRef = useRef<Track[]>([])
  useEffect(() => { favoritesRef.current = favorites }, [favorites])

  const reload = useCallback(async () => {
    const [tracks, dir] = await Promise.all([GetFavorites(), GetFavoritesDir()])
    favoritesRef.current = tracks ?? []
    setFavorites(tracks ?? [])
    setFavoritesDir(dir)
  }, [])

  useEffect(() => { reload() }, [reload])

  const isFavorite = useCallback(
    (id: number) => favorites.some(t => t.id === id),
    [favorites],
  )

  const toggleFavorite = useCallback(
    async (track: Track) => {
      // Decide the branch from the synchronous ref so back-to-back toggles
      // before a re-render still alternate correctly instead of both adding.
      const wasFavorite = favoritesRef.current.some(t => t.id === track.id)
      const optimistic = wasFavorite
        ? favoritesRef.current.filter(t => t.id !== track.id)
        : [...favoritesRef.current, track]
      favoritesRef.current = optimistic
      // Functional updater dedups on insert as a second line of defense.
      setFavorites(prev =>
        wasFavorite
          ? prev.filter(t => t.id !== track.id)
          : prev.some(t => t.id === track.id)
            ? prev
            : [...prev, track],
      )
      try {
        if (wasFavorite) {
          await RemoveFavorite(track.id)
        } else {
          await AddFavorite(track)
        }
      } catch (e) {
        // Roll back the optimistic change on failure.
        const reverted = wasFavorite
          ? favoritesRef.current.some(t => t.id === track.id)
            ? favoritesRef.current
            : [...favoritesRef.current, track]
          : favoritesRef.current.filter(t => t.id !== track.id)
        favoritesRef.current = reverted
        setFavorites(prev =>
          wasFavorite
            ? prev.some(t => t.id === track.id) ? prev : [...prev, track]
            : prev.filter(t => t.id !== track.id),
        )
        throw e
      }
    },
    [],
  )

  const pickDir = useCallback(async () => {
    const dir = await PickFavoritesDir()
    if (dir) setFavoritesDir(dir)
  }, [])

  return { favorites, favoritesDir, isFavorite, toggleFavorite, pickDir, reload }
}
