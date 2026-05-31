import { useState, useEffect, useCallback } from 'react'
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

  const reload = useCallback(async () => {
    const [tracks, dir] = await Promise.all([GetFavorites(), GetFavoritesDir()])
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
      if (isFavorite(track.id)) {
        await RemoveFavorite(track.id)
      } else {
        await AddFavorite(track)
      }
      await reload()
    },
    [isFavorite, reload],
  )

  const pickDir = useCallback(async () => {
    const dir = await PickFavoritesDir()
    if (dir) setFavoritesDir(dir)
  }, [])

  return { favorites, favoritesDir, isFavorite, toggleFavorite, pickDir, reload }
}
