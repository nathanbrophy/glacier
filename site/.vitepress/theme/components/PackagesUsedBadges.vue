<!--
  PackagesUsedBadges - on a task page, lists the packages composed by that
  task. Required at the top of every task page per spec 0031 §Architecture.
-->
<script setup lang="ts">
import { withBase } from 'vitepress'
import { computed } from 'vue'
import { packages } from '../../data/packages'

const props = defineProps<{
  packageNames: readonly string[]
}>()

const items = computed(() => {
  return props.packageNames
    .map((name) => packages.find((p) => p.name === name))
    .filter((p): p is NonNullable<typeof p> => Boolean(p))
})
</script>

<template>
  <p class="glacier-pkg-row" v-if="items.length > 0">
    <span class="glacier-pkg-row__label">Packages used:</span>
    <a
      v-for="pkg in items"
      :key="pkg.name"
      class="glacier-pkg-row__chip"
      :data-tier="pkg.tier"
      :href="withBase(`/docs/packages/${pkg.slug}`)"
    >
      {{ pkg.name }}
    </a>
  </p>
</template>

<style scoped>
.glacier-pkg-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.6rem;
  margin: 0.5rem 0 1.6rem;
  font-size: 0.92rem;
}

.glacier-pkg-row__label {
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 0.76rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--mg-text-muted);
  margin-right: 0.2rem;
}

.glacier-pkg-row__chip {
  font-family: var(--mg-font-mono);
  font-size: 0.86rem;
  font-weight: 600;
  padding: 0.25rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--mg-border);
  background: var(--mg-surface);
  text-decoration: none !important;
  transition: border-color 120ms ease, transform 120ms ease;
}

.glacier-pkg-row__chip[data-tier='kernel'] {
  color: var(--mg-cyan-100);
}

.glacier-pkg-row__chip[data-tier='mid'] {
  color: var(--mg-cyan);
}

.glacier-pkg-row__chip[data-tier='leaf'] {
  color: var(--mg-teal);
}

.glacier-pkg-row__chip:hover {
  transform: translateY(-1px);
  border-color: currentColor;
}
</style>
