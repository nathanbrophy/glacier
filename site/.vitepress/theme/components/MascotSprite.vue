<!--
  MascotSprite - small illustrated polar-bear companion. Per spec 0001
  amendment A (in spec 0031), this asset class is allowed only in:
    - page footer
    - 404 page
    - sidebar empty states
    - scroll-to-top button
  Never replaces the kaomoji at hero scale.
-->
<script setup lang="ts">
import { withBase } from 'vitepress'
import { computed } from 'vue'
import type { CompanionState } from '../../types'

const props = withDefaults(
  defineProps<{
    state?: CompanionState
    size?: number
  }>(),
  { state: 'idle', size: 64 },
)

const STATE_LABEL: Record<CompanionState, string> = {
  idle:     'Glacier companion bear',
  wave:     'Glacier companion bear, waving hello',
  thinking: 'Glacier companion bear, thinking',
}

const src = computed(() => withBase(`/mascot/companion-${props.state}.svg`))
const label = computed(() => STATE_LABEL[props.state])
</script>

<template>
  <img
    class="glacier-companion"
    :src="src"
    :alt="label"
    role="img"
    :width="size"
    :height="size"
    decoding="async"
    loading="lazy"
  />
</template>

<style scoped>
.glacier-companion {
  display: inline-block;
  vertical-align: middle;
  user-select: none;
  transition: transform 200ms cubic-bezier(0.34, 1.56, 0.64, 1);
}

.glacier-companion:hover {
  transform: translateY(-2px) rotate(-3deg);
}

@media (prefers-reduced-motion: reduce) {
  .glacier-companion {
    transition: none;
  }
}
</style>
