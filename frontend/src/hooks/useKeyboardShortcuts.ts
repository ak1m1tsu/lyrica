import { RefObject, useEffect, useRef } from 'react'

export interface KeyboardShortcutsOptions {
  view: 'home' | 'lyrics' | 'settings'
  favPanelOpen: boolean
  aboutOpen: boolean
  searchInputRef: RefObject<HTMLInputElement>
  onBack: () => void
  onToggleFavPanel: () => void
  onOpenSettings: () => void
  onCloseAbout: () => void
}

export function useKeyboardShortcuts(options: KeyboardShortcutsOptions): void {
  const optsRef = useRef(options)
  useEffect(() => { optsRef.current = options })

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const opts = optsRef.current
      const inInput = e.target instanceof HTMLElement &&
        (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA')

      if (e.key === 'Escape') {
        if (opts.aboutOpen) {
          opts.onCloseAbout()
          e.preventDefault()
        } else if (opts.favPanelOpen) {
          opts.onToggleFavPanel()
          e.preventDefault()
        } else if (opts.view === 'lyrics' || opts.view === 'settings') {
          opts.onBack()
          e.preventDefault()
        }
        return
      }

      if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
        opts.searchInputRef.current?.focus()
        opts.searchInputRef.current?.select()
        e.preventDefault()
        return
      }

      if (inInput) return

      if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
        opts.onToggleFavPanel()
        e.preventDefault()
        return
      }

      if ((e.ctrlKey || e.metaKey) && e.key === ',') {
        if (opts.view !== 'settings') {
          opts.onOpenSettings()
          e.preventDefault()
        }
        return
      }

      if (opts.view === 'lyrics') {
        if ((e.altKey && e.key === 'ArrowLeft') ||
            (e.key === 'Backspace' && !e.altKey && !e.ctrlKey && !e.metaKey)) {
          opts.onBack()
          e.preventDefault()
        }
      }
    }

    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [])
}
