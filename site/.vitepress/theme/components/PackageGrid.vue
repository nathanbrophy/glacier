<!--
  PackageGrid - the 15-package suite rendered as three tier-grouped grids.
  Used on / and /features.
-->
<script setup lang="ts">
import { withBase } from 'vitepress'
import { computed } from 'vue'
import { packages } from '../../data/packages'
import type { Tier } from '../../types'

const TIER_ORDER: readonly Tier[] = ['kernel', 'mid', 'leaf']

const TIER_LABEL: Record<Tier, string> = {
  kernel: 'Kernel',
  mid: 'Mid',
  leaf: 'Leaves',
}

const TIER_KICKER: Record<Tier, string> = {
  kernel: 'Tier 0. Universal: every Glacier consumer transitively depends on these.',
  mid: 'Tier 1. Independent of each other; depend only on kernel.',
  leaf: 'Tier 2. Large enough to deserve isolation; never import each other.',
}

const grouped = computed(() => {
  return TIER_ORDER.map((tier) => ({
    tier,
    label: TIER_LABEL[tier],
    kicker: TIER_KICKER[tier],
    items: packages.filter((p) => p.tier === tier),
  }))
})

function packageHref(slug: string): string {
  return withBase(`/docs/packages/${slug}`)
}
</script>

<template>
  <section class="glacier-grid">
    <div v-for="group in grouped" :key="group.tier" class="glacier-grid__tier">
      <header class="glacier-grid__tier-head">
        <span class="glacier-grid__tier-label" :data-tier="group.tier">
          {{ group.label }}
        </span>
        <p class="glacier-grid__tier-kicker">{{ group.kicker }}</p>
      </header>
      <div class="glacier-grid__cards">
        <a
          v-for="pkg in group.items"
          :key="pkg.name"
          :href="packageHref(pkg.slug)"
          class="glacier-grid__card"
          :data-tier="pkg.tier"
        >
          <h4 class="glacier-grid__name">{{ pkg.name }}</h4>
          <p class="glacier-grid__teaser">{{ pkg.teaser }}</p>
          <span class="glacier-grid__chevron" aria-hidden="true">&rarr;</span>
        </a>
      </div>
    </div>
  </section>
</template>

<style scoped>
.glacier-grid {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.glacier-grid__tier {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.glacier-grid__tier-head {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.glacier-grid__tier-label {
  align-self: flex-start;
  font-family: var(--mg-font-display);
  font-weight: 700;
  font-size: 0.72rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  padding: 0.3rem 0.7rem;
  border-radius: 999px;
  border: 1px solid var(--mg-border);
}

.glacier-grid__tier-label[data-tier='kernel'] {
  color: var(--mg-cyan-100);
  background: rgba(165, 243, 252, 0.06);
  border-color: rgba(165, 243, 252, 0.3);
}

.glacier-grid__tier-label[data-tier='mid'] {
  color: var(--mg-cyan);
  background: rgba(34, 211, 238, 0.06);
  border-color: rgba(34, 211, 238, 0.3);
}

.glacier-grid__tier-label[data-tier='leaf'] {
  color: var(--mg-teal);
  background: rgba(45, 212, 191, 0.06);
  border-color: rgba(45, 212, 191, 0.3);
}

.glacier-grid__tier-kicker {
  margin: 0;
  font-size: 0.92rem;
  color: var(--mg-text-muted);
  max-width: 70ch;
}

.glacier-grid__cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 0.8rem;
}

.glacier-grid__card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 1.1rem 1.2rem 1.4rem;
  background: var(--mg-surface);
  border: 1px solid var(--mg-border);
  border-radius: 10px;
  text-decoration: none;
  transition: border-color 160ms ease, transform 160ms ease, background 160ms ease;
}

.glacier-grid__card:hover,
.glacier-grid__card:focus-visible {
  border-color: var(--mg-cyan);
  transform: translateY(-2px);
}

.glacier-grid__card[data-tier='leaf']:hover,
.glacier-grid__card[data-tier='leaf']:focus-visible {
  border-color: var(--mg-teal);
}

.glacier-grid__name {
  margin: 0;
  font-family: var(--mg-font-mono);
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--mg-cyan);
}

.glacier-grid__card[data-tier='leaf'] .glacier-grid__name {
  color: var(--mg-teal);
}

.glacier-grid__teaser {
  margin: 0;
  font-size: 0.92rem;
  line-height: 1.5;
  color: var(--mg-text-muted);
}

.glacier-grid__chevron {
  position: absolute;
  top: 1.1rem;
  right: 1.1rem;
  color: var(--mg-text-faint);
  transition: transform 160ms ease, color 160ms ease;
}

.glacier-grid__card:hover .glacier-grid__chevron {
  color: var(--mg-cyan);
  transform: translateX(2px);
}
</style>
