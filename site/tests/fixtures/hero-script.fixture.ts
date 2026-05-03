// Canonical hero terminal script — shared by HeroTerminal at runtime AND by
// future component / e2e tests. See specs/0031-public-site.md §Examples.

import type { TerminalLine } from '../../.vitepress/types'

export const heroScript: readonly TerminalLine[] = [
  { kind: 'cmd', text: 'go get github.com/nathanbrophy/glacier' },
  { kind: 'out', text: 'go: added github.com/nathanbrophy/glacier' },
  { kind: 'cmd', text: 'go run ./cmd/example' },
  { kind: 'out', text: 'ʕ•ᴥ•ʔ glacier: ready',         mascotState: 'calm' },
  { kind: 'out', text: 'ʕ⌐■-■ʔ serving on :8080',      mascotState: 'confident' },
] as const
