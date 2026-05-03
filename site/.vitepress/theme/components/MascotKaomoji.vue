<!--
  MascotKaomoji — renders the canonical kaomoji at hero or inline scale,
  with state shifts mapped one-to-one to spec 0001 D45.

  When inside a HeroTerminal, the parent provides a reactive `HERO_MASCOT_STATE`
  ref; this component prefers the injected value over the prop, so the
  terminal can drive expression changes as it advances.

  Accessibility per spec 0031 §Architecture / Visual system / Mascot:
    - aria-label set per state (human-readable, e.g. "Glacier mascot, calm")
    - raw kaomoji span aria-hidden so screen readers don't mangle codepoints.
-->
<script setup lang="ts">
import { computed, inject, type Ref } from 'vue'
import {
  HERO_MASCOT_STATE,
  MASCOT_KAOMOJI,
  MASCOT_LABEL,
  type MascotState,
} from '../../types'

const props = withDefaults(
  defineProps<{
    state?: MascotState
    scale?: number
  }>(),
  {
    state: 'calm',
    scale: 1,
  },
)

const injected = inject<Ref<MascotState> | null>(HERO_MASCOT_STATE, null)

const effectiveState = computed<MascotState>(() => {
  return injected?.value ?? props.state
})

const kaomoji = computed(() => MASCOT_KAOMOJI[effectiveState.value])
const label = computed(() => MASCOT_LABEL[effectiveState.value])
</script>

<template>
  <div
    class="glacier-mascot"
    role="img"
    :aria-label="label"
    :data-state="effectiveState"
    :style="{ '--mascot-scale': scale }"
  >
    <span class="glacier-mascot__face" aria-hidden="true">{{ kaomoji }}</span>
  </div>
</template>

<style scoped>
.glacier-mascot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: var(--mg-font-mono);
  line-height: 1;
  user-select: none;
}

.glacier-mascot__face {
  font-size: clamp(64px, 8vw, 120px);
  font-weight: 500;
  letter-spacing: -0.02em;
  background: linear-gradient(
    180deg,
    var(--mg-cyan-100) 0%,
    var(--mg-cyan)     50%,
    var(--mg-teal)     100%
  );
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  filter: drop-shadow(0 0 18px rgba(34, 211, 238, 0.18));
  transition: transform 240ms cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* Subtle expression-shift animation when the state changes. */
.glacier-mascot[data-state="confident"] .glacier-mascot__face {
  transform: scale(1.04) rotate(-2deg);
}

.glacier-mascot[data-state="thinking"] .glacier-mascot__face {
  transform: translateY(-2px) rotate(2deg);
}

.glacier-mascot[data-state="alarmed"] .glacier-mascot__face {
  transform: scale(1.08);
  filter: drop-shadow(0 0 18px rgba(251, 191, 36, 0.28));
}

.glacier-mascot[data-state="error"] .glacier-mascot__face {
  filter: drop-shadow(0 0 18px rgba(248, 113, 113, 0.28));
}

@media (prefers-reduced-motion: reduce) {
  .glacier-mascot__face {
    transition: none;
  }
}
</style>
