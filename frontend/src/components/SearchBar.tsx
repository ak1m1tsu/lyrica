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
        className="flex-1 rounded-lg bg-[#0f1117]/5 dark:bg-white/5 px-4 py-3 text-[#0f1117] dark:text-white placeholder-[#0f1117]/40 dark:placeholder-white/30 outline-none ring-1 ring-[#0f1117]/20 dark:ring-white/10 focus:ring-[#9B84D1] dark:focus:ring-white/30 transition-all"
      />
      <button
        type="submit"
        className="rounded-lg bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] hover:from-[#D4C0F5] hover:to-[#A892D8] px-4 py-3 text-white transition-colors"
        aria-label="Search"
      >
        <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
        </svg>
      </button>
    </form>
  )
}
