// Canonical hero terminal script — shared by HeroTerminal at runtime AND by
// future component / e2e tests. See specs/0031-public-site.md §Examples.

import type { TerminalLine } from '../../.vitepress/types'

export const heroScript: readonly TerminalLine[] = [
  { kind: 'cmd', text: 'go install github.com/nathanbrophy/glacier/cmd/glacier@latest' },
  { kind: 'out', text: 'go: downloading glacier v0',                       mascotState: 'thinking' },
  { kind: 'cmd', text: 'glacier init my-app --yes' },
  { kind: 'out', text: 'ʕ•ᴥ•ʔ  scaffolding my-app',                        mascotState: 'calm' },
  { kind: 'out', text: 'ʕ⌐■-■ʔ  cli, mock, httpmock, otel wired',          mascotState: 'confident' },
  { kind: 'out', text: 'ʕ⌐■-■ʔ  all set. cd my-app && glacier test',       mascotState: 'confident' },
] as const
