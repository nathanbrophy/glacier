// Custom Glacier theme. Extends VitePress default theme with:
//   - dark-only brand tokens (tokens.css)
//   - landing-hero styles (site.css)
//   - AuroraBackdrop wrapping the whole shell (Layout.vue)
//   - global registration of hero / pillar / package / badge components so
//     they can be used from Markdown without per-page imports.

import type { Theme } from 'vitepress'
import DefaultTheme from 'vitepress/theme'

import Layout from './Layout.vue'
import NotFound from './components/NotFound.vue'

import AuroraBackdrop from './components/AuroraBackdrop.vue'
import CodeCompare from './components/CodeCompare.vue'
import HeroSection from './components/HeroSection.vue'
import HeroTerminal from './components/HeroTerminal.vue'
import MascotKaomoji from './components/MascotKaomoji.vue'
import MascotSprite from './components/MascotSprite.vue'
import PackageGrid from './components/PackageGrid.vue'
import PackagesUsedBadges from './components/PackagesUsedBadges.vue'
import PillarCard from './components/PillarCard.vue'
import PromiseSection from './components/PromiseSection.vue'
import TierBadge from './components/TierBadge.vue'
import UsedInTasksBadges from './components/UsedInTasksBadges.vue'
import WordmarkSVG from './components/WordmarkSVG.vue'

import './styles/tokens.css'
import './styles/site.css'

const theme: Theme = {
  extends: DefaultTheme,
  Layout,
  NotFound,
  enhanceApp({ app }) {
    app.component('AuroraBackdrop', AuroraBackdrop)
    app.component('CodeCompare', CodeCompare)
    app.component('HeroSection', HeroSection)
    app.component('HeroTerminal', HeroTerminal)
    app.component('MascotKaomoji', MascotKaomoji)
    app.component('MascotSprite', MascotSprite)
    app.component('PackageGrid', PackageGrid)
    app.component('PackagesUsedBadges', PackagesUsedBadges)
    app.component('PillarCard', PillarCard)
    app.component('PromiseSection', PromiseSection)
    app.component('TierBadge', TierBadge)
    app.component('UsedInTasksBadges', UsedInTasksBadges)
    app.component('WordmarkSVG', WordmarkSVG)
  },
}

export default theme
