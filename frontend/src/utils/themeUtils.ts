import type { ThemeColors } from '../types/theme'

function hexToRgba(hex: string, alpha: number): string {
  const h = hex.replace('#', '')
  const r = parseInt(h.slice(0, 2), 16)
  const g = parseInt(h.slice(2, 4), 16)
  const b = parseInt(h.slice(4, 6), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

function clamp(v: number): number {
  return Math.max(0, Math.min(255, v))
}

function adjustBrightness(hex: string, amount: number): string {
  const h = hex.replace('#', '')
  const r = clamp(parseInt(h.slice(0, 2), 16) + amount)
  const g = clamp(parseInt(h.slice(2, 4), 16) + amount)
  const b = clamp(parseInt(h.slice(4, 6), 16) + amount)
  return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`
}

// Perceived luminance (0=black, 1=white) used to pick contrast-safe accent text.
function luminance(hex: string): number {
  const h = hex.replace('#', '')
  const toLinear = (c: number) => {
    const s = c / 255
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4)
  }
  const r = toLinear(parseInt(h.slice(0, 2), 16))
  const g = toLinear(parseInt(h.slice(2, 4), 16))
  const b = toLinear(parseInt(h.slice(4, 6), 16))
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

export function applyTheme(colors: ThemeColors): void {
  const root = document.documentElement
  const set = (name: string, value: string) => root.style.setProperty(name, value)

  set('--color-bg', colors.background)
  set('--color-surface', colors.surface)
  set('--color-text', colors.text)

  for (const [suffix, alpha] of [['90', 0.9], ['80', 0.8], ['70', 0.7], ['60', 0.6], ['50', 0.5], ['40', 0.4], ['30', 0.3]] as [string, number][]) {
    set(`--color-text-${suffix}`, hexToRgba(colors.text, alpha))
  }

  // Card/surface overlays: text color at low opacity for subtle backgrounds
  set('--color-card-20', hexToRgba(colors.text, 0.2))
  set('--color-card-10', hexToRgba(colors.text, 0.1))
  set('--color-card-05', hexToRgba(colors.text, 0.05))

  set('--color-border', hexToRgba(colors.text, 0.1))

  set('--color-accent', colors.accent)
  set('--color-accent-hover', adjustBrightness(colors.accent, -13))
  set('--color-accent-lt', colors.accentLight)
  set('--color-accent-lt-hover', adjustBrightness(colors.accentLight, 13))
  set('--color-accent-20', hexToRgba(colors.accent, 0.2))
  set('--color-accent-10', hexToRgba(colors.accent, 0.1))

  // Contrast-safe accent for text/links: lighter on dark bg, darker on light bg
  const bgLum = luminance(colors.background)
  const accentOnBg = bgLum < 0.5 ? colors.accentLight : colors.accent
  const accentOnBgHover = bgLum < 0.5
    ? adjustBrightness(colors.accentLight, 13)
    : adjustBrightness(colors.accent, -13)
  set('--color-accent-on-bg', accentOnBg)
  set('--color-accent-on-bg-hover', accentOnBgHover)

  set('--color-scroll-thumb', colors.accent)
  set('--color-scroll-thumb-hover', colors.accentLight)
}
