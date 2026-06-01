interface Props {
  isFavorite: boolean
  onToggle: () => Promise<void>
  size?: 'sm' | 'md'
}

export function FavoriteButton({ isFavorite, onToggle, size = 'md' }: Props) {
  const dim = size === 'sm' ? 14 : 18

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onToggle()
  }

  return (
    <button
      onClick={handleClick}
      aria-label={isFavorite ? 'Remove from favorites' : 'Add to favorites'}
      aria-pressed={isFavorite}
      className={`flex items-center justify-center rounded transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-[#9B84D1]
        ${size === 'sm' ? 'h-6 w-6' : 'h-8 w-8'}
        ${isFavorite
          ? 'text-[#C8B1F3] hover:text-[#D4C0F5] dark:text-[#C8B1F3] dark:hover:text-[#D4C0F5]'
          : 'text-[#0f1117]/40 dark:text-white/40 hover:text-[#9B84D1] dark:hover:text-[#C8B1F3]'
        }`}
    >
      {isFavorite ? (
        <svg width={dim} height={dim} viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z" />
        </svg>
      ) : (
        <svg width={dim} height={dim} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
        </svg>
      )}
    </button>
  )
}
