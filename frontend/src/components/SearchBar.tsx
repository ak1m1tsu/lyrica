import { RefObject } from 'react'

interface Props {
  onSearch: (query: string) => void
  inputRef?: RefObject<HTMLInputElement>
}

export function SearchBar({ onSearch, inputRef }: Props) {
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const query = (e.currentTarget.elements.namedItem('q') as HTMLInputElement).value
    onSearch(query)
  }

  return (
    <form onSubmit={handleSubmit} className="relative w-full max-w-xl flex gap-2">
      <input
        type="search"
        name="q"
        aria-label="Search for a song"
        placeholder="Search for a song..."
        ref={inputRef}
        className="flex-1 rounded-lg bg-[var(--color-card-05)] px-4 py-3 text-[var(--color-text)] placeholder-[var(--color-text-40)] outline-none ring-1 ring-[var(--color-card-20)] focus:ring-[var(--color-accent)] transition-all"
      />
      <button
        type="submit"
        className="rounded-lg bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] hover:from-[var(--color-accent-lt-hover)] hover:to-[var(--color-accent-hover)] px-4 py-3 text-white transition-colors"
        aria-label="Search"
      >
        <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
        </svg>
      </button>
    </form>
  )
}
