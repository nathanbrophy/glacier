// Shared TypeScript schema for Glacier site components.
// See specs/0031-public-site.md §Schema.

export type MascotState =
  | 'calm'
  | 'confident'
  | 'thinking'
  | 'alarmed'
  | 'error'

// Map to canonical kaomoji forms — spec 0001 D43, D45.
export const MASCOT_KAOMOJI: Record<MascotState, string> = {
  calm: 'ʕ•ᴥ•ʔ',
  confident: 'ʕ⌐■-■ʔ',
  thinking: 'ʕ•_•ʔ',
  alarmed: 'ʕ◉_◉ʔ',
  error: 'ʕ× ×ʔ',
}

export const MASCOT_LABEL: Record<MascotState, string> = {
  calm: 'Glacier mascot, calm',
  confident: 'Glacier mascot, confident',
  thinking: 'Glacier mascot, thinking',
  alarmed: 'Glacier mascot, alarmed',
  error: 'Glacier mascot, error',
}

export type CompanionState = 'idle' | 'wave' | 'thinking'

export type Tier = 'kernel' | 'mid' | 'leaf'

export interface TerminalLine {
  kind: 'cmd' | 'out'
  text: string
  mascotState?: MascotState
}

// Vue inject key used by HeroTerminal to publish current mascot state to a sibling MascotKaomoji.
export const HERO_MASCOT_STATE = Symbol('hero-mascot-state')
