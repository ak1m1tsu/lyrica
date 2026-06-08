import { useCallback, useEffect, useState } from 'react'
import type { ThemeDefinition } from '../types/theme'
import { BUILT_IN_IDS, BUILT_IN_THEMES } from '../types/theme'
import { applyTheme } from '../utils/themeUtils'
import {
  DeleteCustomTheme,
  ExportTheme,
  GetCurrentTheme,
  GetCustomThemes,
  ImportTheme,
  SaveCustomTheme,
  SetCurrentTheme,
} from '../../wailsjs/go/main/App'
import { domain } from '../../wailsjs/go/models'

export interface UseThemeResult {
  themeId: string
  themes: ThemeDefinition[]
  setThemeId: (id: string) => Promise<void>
  addTheme: (theme: ThemeDefinition) => Promise<void>
  updateTheme: (theme: ThemeDefinition) => Promise<void>
  removeTheme: (id: string) => Promise<void>
  exportTheme: (id: string) => Promise<void>
  importTheme: () => Promise<void>
}

export function useTheme(): UseThemeResult {
  const [themeId, setThemeIdState] = useState<string>('dark')
  const [customThemes, setCustomThemes] = useState<ThemeDefinition[]>([])

  const allThemes: ThemeDefinition[] = [...BUILT_IN_THEMES, ...customThemes]

  const applyById = useCallback((id: string, list: ThemeDefinition[]) => {
    const found = list.find(t => t.id === id)
    applyTheme((found ?? BUILT_IN_THEMES[0]).colors)
    document.documentElement.classList.remove('dark')
  }, [])

  useEffect(() => {
    Promise.all([GetCurrentTheme(), GetCustomThemes()])
      .then(([id, custom]) => {
        const customs = custom ?? []
        setCustomThemes(customs)
        const resolved = id || 'dark'
        setThemeIdState(resolved)
        applyById(resolved, [...BUILT_IN_THEMES, ...customs])
      })
      .catch(() => applyById('dark', BUILT_IN_THEMES))
  }, [applyById])

  const setThemeId = useCallback(async (id: string) => {
    setThemeIdState(id)
    applyById(id, allThemes)
    await SetCurrentTheme(id)
  }, [allThemes, applyById])

  const toWailsTheme = (t: ThemeDefinition): domain.Theme =>
    new domain.Theme({ id: t.id, name: t.name, colors: t.colors })

  const addTheme = useCallback(async (theme: ThemeDefinition) => {
    await SaveCustomTheme(toWailsTheme(theme))
    setCustomThemes(prev => [...prev.filter(t => t.id !== theme.id), theme])
  }, [])

  const updateTheme = useCallback(async (theme: ThemeDefinition) => {
    await SaveCustomTheme(toWailsTheme(theme))
    setCustomThemes(prev => prev.map(t => t.id === theme.id ? theme : t))
    if (themeId === theme.id) applyTheme(theme.colors)
  }, [themeId])

  const removeTheme = useCallback(async (id: string) => {
    await DeleteCustomTheme(id)
    setCustomThemes(prev => prev.filter(t => t.id !== id))
  }, [])

  const exportThemeFn = useCallback(async (id: string) => {
    await ExportTheme(id)
  }, [])

  const importThemeFn = useCallback(async () => {
    const imported = await ImportTheme()
    if (imported) {
      setCustomThemes(prev => [...prev.filter(t => t.id !== imported.id), imported])
      setThemeIdState(imported.id)
      applyTheme(imported.colors)
    }
  }, [])

  return {
    themeId,
    themes: allThemes,
    setThemeId,
    addTheme,
    updateTheme,
    removeTheme,
    exportTheme: exportThemeFn,
    importTheme: importThemeFn,
  }
}
