import { Track } from './TrackCard'
import { FavoriteButton } from './FavoriteButton'
import { formatDuration } from '../utils/formatting'

interface Props {
  open: boolean
  favorites: Track[]
  favoritesDir: string
  onClose: () => void
  onSelect: (track: Track) => void
  onRemove: (track: Track) => Promise<void>
  onPickDir: () => Promise<void>
}

export function FavoritesPanel({ open, favorites, favoritesDir, onClose, onSelect, onRemove, onPickDir }: Props) {
  return (
    <>
      {open && (
        <div
          className="fixed inset-0 z-40 bg-black/40"
          onClick={onClose}
        />
      )}

      <div
        className={`fixed top-0 right-0 z-50 flex h-full w-80 flex-col
          bg-[var(--color-surface)] border-l border-[var(--color-accent-20)]
          shadow-2xl transition-transform duration-200
          ${open ? 'translate-x-0' : 'translate-x-full'}`}
      >
        <div className="flex items-center justify-between border-b border-[var(--color-accent-20)] px-4 py-3">
          <span className="font-semibold text-[var(--color-text)]">Favorites</span>
          <button
            onClick={onClose}
            aria-label="Close favorites"
            className="flex h-7 w-7 items-center justify-center rounded text-[var(--color-text-40)] hover:bg-[var(--color-card-10)] hover:text-[var(--color-text)] transition-colors"
          >
            <svg width="14" height="14" viewBox="0 0 10 10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
              <line x1="1" y1="1" x2="9" y2="9" />
              <line x1="9" y1="1" x2="1" y2="9" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto">
          {favorites.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-2 py-16 text-[var(--color-text-30)]">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
              </svg>
              <p className="text-sm">No favorites yet</p>
            </div>
          ) : (
            <ul className="divide-y divide-[var(--color-border)]">
              {favorites.map(track => (
                <li key={track.id} className="flex items-center gap-2 px-3 py-2.5">
                  <button
                    onClick={() => { onClose(); onSelect(track) }}
                    className="min-w-0 flex-1 text-left"
                  >
                    <p className="truncate text-sm font-medium text-[var(--color-text)]">{track.trackName}</p>
                    <p className="truncate text-xs text-[var(--color-text-60)]">
                      {track.artistName}{track.albumName ? ` — ${track.albumName}` : ''} · {formatDuration(track.duration)}
                    </p>
                  </button>
                  <FavoriteButton isFavorite size="sm" onToggle={() => onRemove(track)} />
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="border-t border-[var(--color-accent-20)] px-4 py-3">
          <p className="mb-1 text-xs text-[var(--color-text-40)]">Storage folder</p>
          <p className="mb-2 truncate text-xs text-[var(--color-text-70)]" title={favoritesDir}>{favoritesDir || '—'}</p>
          <button
            onClick={onPickDir}
            className="w-full rounded px-3 py-1.5 text-xs font-medium
              bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)]
              hover:from-[var(--color-accent-lt-hover)] hover:to-[var(--color-accent-hover)]
              text-white transition-colors"
          >
            Change folder…
          </button>
        </div>
      </div>
    </>
  )
}
