import { useState, useEffect } from 'react'

interface Props {
  onSearch: (query: string) => void
  loading: boolean
}

export function SearchBar({ onSearch, loading }: Props) {
  const [value, setValue] = useState('')

  useEffect(() => {
    if (!value.trim()) {
      onSearch('')
      return
    }
    const timer = setTimeout(() => onSearch(value), 300)
    return () => clearTimeout(timer)
  }, [value])

  return (
    <div className="relative w-full max-w-xl">
      <input
        type="search"
        value={value}
        onChange={e => setValue(e.target.value)}
        placeholder="Search for a song..."
        className="w-full rounded-lg bg-white/10 px-4 py-3 text-white placeholder-gray-500 outline-none ring-1 ring-white/10 focus:ring-white/30 transition-all"
      />
      {loading && (
        <div className="absolute right-3 top-1/2 -translate-y-1/2">
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-white/20 border-t-white" />
        </div>
      )}
    </div>
  )
}
