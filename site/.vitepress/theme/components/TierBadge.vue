<!--
  TierBadge - small pill that names the package's tier and links to the
  /concepts page anchor explaining that tier. Required on every package
  reference page per spec 0031 §Architecture / Sidebar cross-reference.
-->
<script setup lang="ts">
import { withBase } from 'vitepress'
import { computed } from 'vue'
import type { Tier } from '../../types'

const props = defineProps<{
  tier: Tier
}>()

const TIER_LABEL: Record<Tier, string> = {
  kernel: 'Kernel · Tier 0',
  mid: 'Mid · Tier 1',
  leaf: 'Leaf · Tier 2',
}

const label = computed(() => TIER_LABEL[props.tier])
const href = computed(() => withBase(`/concepts#tier-${props.tier}`))
</script>

<template>
  <a class="glacier-tier-badge" :data-tier="tier" :href="href">
    {{ label }}
  </a>
</template>

<style scoped>
.glacier-tier-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 0.72rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  padding: 0.3rem 0.75rem;
  border-radius: 999px;
  border: 1px solid var(--mg-border);
  background: var(--mg-surface);
  text-decoration: none !important;
  transition: border-color 120ms ease, color 120ms ease, transform 120ms ease;
}

.glacier-tier-badge[data-tier='kernel'] {
  color: var(--mg-cyan-100);
  border-color: rgba(165, 243, 252, 0.35);
}

.glacier-tier-badge[data-tier='mid'] {
  color: var(--mg-cyan);
  border-color: rgba(34, 211, 238, 0.35);
}

.glacier-tier-badge[data-tier='leaf'] {
  color: var(--mg-teal);
  border-color: rgba(45, 212, 191, 0.35);
}

.glacier-tier-badge:hover {
  transform: translateY(-1px);
}
</style>
