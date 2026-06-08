export interface ThemeColors {
  background: string
  surface: string
  text: string
  accent: string
  accentLight: string
}

export interface ThemeDefinition {
  id: string
  name: string
  colors: ThemeColors
}

export const BUILT_IN_IDS = new Set(['light', 'dark'])

export const BUILT_IN_THEMES: ThemeDefinition[] = [
  {
    id: 'dark',
    name: 'Dark',
    colors: {
      background: '#0f1117',
      surface: '#161920',
      text: '#ffffff',
      accent: '#9B84D1',
      accentLight: '#C8B1F3',
    },
  },
  {
    id: 'light',
    name: 'Light',
    colors: {
      background: '#ffffff',
      surface: '#f5f3ff',
      text: '#0f1117',
      accent: '#9B84D1',
      accentLight: '#C8B1F3',
    },
  },
]
