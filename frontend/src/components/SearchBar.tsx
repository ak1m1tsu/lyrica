interface Props {
  onSearch: (query: string) => void
}

export function SearchBar({ onSearch }: Props) {
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
        className="flex-1 rounded-lg bg-gray-100 dark:bg-white/10 px-4 py-3 text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 outline-none ring-1 ring-gray-300 dark:ring-white/10 focus:ring-gray-400 dark:focus:ring-white/30 transition-all"
      />
      <button
        type="submit"
        className="rounded-lg bg-gray-100 dark:bg-white/10 px-4 py-3 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-white/20 hover:text-gray-900 dark:hover:text-white transition-colors ring-1 ring-gray-300 dark:ring-white/10"
        aria-label="Search"
      >
        <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
        </svg>
      </button>
    </form>
  )
}
