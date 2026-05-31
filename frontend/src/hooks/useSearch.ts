import { useState, useRef, useCallback, useEffect } from 'react'
import { Search } from '../../wailsjs/go/main/App'
import { Track } from '../components/TrackCard'

const DEBOUNCE_MS = 300
const CACHE_MAX = 50

export function useSearch() {
  const [results, setResults] = useState<Track[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // In-memory cache keyed by the trimmed query string.
  const cacheRef = useRef<Map<string, Track[]>>(new Map())
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  // Guards against out-of-order responses overwriting newer results.
  const requestIdRef = useRef(0)

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [])

  const search = useCallback((query: string) => {
    const trimmed = query.trim()

    // Cancel any pending debounced request.
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }

    if (!trimmed) {
      requestIdRef.current++
      setResults([])
      setError(null)
      setLoading(false)
      return
    }

    // Cache hit — return immediately without firing an RPC.
    const cached = cacheRef.current.get(trimmed)
    if (cached) {
      // Mark as most-recently-used by reinserting (Map preserves insert order).
      cacheRef.current.delete(trimmed)
      cacheRef.current.set(trimmed, cached)
      requestIdRef.current++
      setResults(cached)
      setError(null)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)

    const reqId = ++requestIdRef.current
    timerRef.current = setTimeout(async () => {
      timerRef.current = null
      try {
        const tracks = (await Search(trimmed)) ?? []
        // Bounded LRU insert: evict the oldest entry once over capacity.
        cacheRef.current.delete(trimmed)
        cacheRef.current.set(trimmed, tracks)
        if (cacheRef.current.size > CACHE_MAX) {
          const oldest = cacheRef.current.keys().next().value
          if (oldest !== undefined) cacheRef.current.delete(oldest)
        }
        if (reqId === requestIdRef.current) setResults(tracks)
      } catch (e: unknown) {
        if (reqId === requestIdRef.current) {
          setError(e instanceof Error ? e.message : String(e))
        }
      } finally {
        if (reqId === requestIdRef.current) setLoading(false)
      }
    }, DEBOUNCE_MS)
  }, [])

  return { results, loading, error, search }
}
