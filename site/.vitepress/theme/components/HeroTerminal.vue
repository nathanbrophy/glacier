<!--
  HeroTerminal — animated typing of a TerminalLine[] script.

  - Provides a reactive mascot state via `HERO_MASCOT_STATE` so a sibling
    <MascotKaomoji /> reflects the current line's expression.
  - Renders a polite aria-live region accumulating the output text for
    screen readers.
  - Honors prefers-reduced-motion: renders the end state immediately and
    sets `data-state="complete"` so e2e can assert end-state without
    waiting for animation.
-->
<script setup lang="ts">
import { onBeforeUnmount, onMounted, provide, ref, shallowRef } from 'vue'
import {
  HERO_MASCOT_STATE,
  type MascotState,
  type TerminalLine,
} from '../../types'

const props = defineProps<{
  script: readonly TerminalLine[]
}>()

interface RenderedLine {
  id: number
  kind: 'cmd' | 'out'
  text: string
  done: boolean
}

const lines = ref<RenderedLine[]>([])
const animationState = ref<'idle' | 'running' | 'complete'>('idle')
const currentMascotState = shallowRef<MascotState>('calm')

provide(HERO_MASCOT_STATE, currentMascotState)

const TYPE_DELAY_CMD = 32   // ms per character for command lines
const TYPE_DELAY_OUT = 14   // ms per character for output lines
const LINE_PAUSE = 360      // ms pause between lines

let timer: ReturnType<typeof setTimeout> | null = null
let cancelled = false

function clearTimer() {
  if (timer !== null) {
    clearTimeout(timer)
    timer = null
  }
}

function jumpToEnd() {
  cancelled = true
  clearTimer()
  lines.value = props.script.map((line, i) => ({
    id: i,
    kind: line.kind,
    text: line.text,
    done: true,
  }))
  const last = props.script[props.script.length - 1]
  if (last?.mascotState) {
    currentMascotState.value = last.mascotState
  }
  animationState.value = 'complete'
}

async function runScript() {
  if (cancelled) return
  animationState.value = 'running'

  for (let i = 0; i < props.script.length; i++) {
    if (cancelled) return
    const source = props.script[i]
    const rendered: RenderedLine = {
      id: i,
      kind: source.kind,
      text: '',
      done: false,
    }
    lines.value = [...lines.value, rendered]

    if (source.mascotState) {
      currentMascotState.value = source.mascotState
    }

    const delay = source.kind === 'cmd' ? TYPE_DELAY_CMD : TYPE_DELAY_OUT
    for (let c = 0; c < source.text.length; c++) {
      if (cancelled) return
      await new Promise<void>((resolve) => {
        timer = setTimeout(resolve, delay)
      })
      rendered.text = source.text.slice(0, c + 1)
      lines.value = [...lines.value]
    }
    rendered.done = true
    lines.value = [...lines.value]

    if (i < props.script.length - 1) {
      await new Promise<void>((resolve) => {
        timer = setTimeout(resolve, LINE_PAUSE)
      })
    }
  }

  if (!cancelled) {
    animationState.value = 'complete'
  }
}

onMounted(() => {
  const reduced =
    typeof window !== 'undefined' &&
    window.matchMedia &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (reduced) {
    jumpToEnd()
    return
  }
  runScript()
})

onBeforeUnmount(() => {
  cancelled = true
  clearTimer()
})

// Accumulated output text for the aria-live region.
function liveText() {
  return lines.value
    .filter((l) => l.kind === 'out')
    .map((l) => l.text)
    .join('\n')
}
</script>

<template>
  <div class="glacier-terminal" :data-state="animationState">
    <div class="glacier-terminal__chrome">
      <span class="glacier-terminal__dot" data-color="r"></span>
      <span class="glacier-terminal__dot" data-color="y"></span>
      <span class="glacier-terminal__dot" data-color="g"></span>
      <span class="glacier-terminal__title">glacier</span>
    </div>
    <div class="glacier-terminal__body">
      <div
        v-for="line in lines"
        :key="line.id"
        class="glacier-terminal__line"
        :class="[`glacier-terminal__line--${line.kind}`]"
      >
        <span v-if="line.kind === 'cmd'" class="glacier-terminal__prompt">$</span>
        <span class="glacier-terminal__text">{{ line.text }}</span>
        <span
          v-if="!line.done && animationState === 'running'"
          class="glacier-terminal__cursor"
          aria-hidden="true"
        ></span>
      </div>
      <div
        v-if="animationState === 'complete'"
        class="glacier-terminal__line"
        aria-hidden="true"
      >
        <span class="glacier-terminal__prompt">$</span>
        <span class="glacier-terminal__cursor glacier-terminal__cursor--idle"></span>
      </div>
    </div>
    <span class="glacier-terminal__sr" aria-live="polite">{{ liveText() }}</span>
  </div>
</template>

<style scoped>
.glacier-terminal {
  width: 100%;
  border: 1px solid var(--mg-border);
  background: var(--mg-surface);
  border-radius: 12px;
  overflow: hidden;
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.04) inset,
    0 24px 60px rgba(0, 0, 0, 0.45);
  font-family: var(--mg-font-mono);
}

.glacier-terminal__chrome {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 0.9rem;
  background: var(--mg-surface-2);
  border-bottom: 1px solid var(--mg-border);
}

.glacier-terminal__dot {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: #444;
}

.glacier-terminal__dot[data-color="r"] { background: #FF5F56; }
.glacier-terminal__dot[data-color="y"] { background: #FFBD2E; }
.glacier-terminal__dot[data-color="g"] { background: #27C93F; }

.glacier-terminal__title {
  margin-left: 0.5rem;
  font-size: 0.78rem;
  color: var(--mg-text-muted);
  letter-spacing: 0.02em;
}

.glacier-terminal__body {
  padding: 1rem 1.1rem 1.2rem;
  font-size: clamp(0.85rem, 1.4vw, 0.98rem);
  line-height: 1.55;
  min-height: 11rem;
}

.glacier-terminal__line {
  display: flex;
  gap: 0.5rem;
  white-space: pre;
  color: var(--mg-text);
}

.glacier-terminal__line--cmd {
  color: var(--mg-text);
}

.glacier-terminal__line--out {
  color: var(--mg-cyan);
}

.glacier-terminal__prompt {
  color: var(--mg-text-faint);
  user-select: none;
}

.glacier-terminal__text {
  flex: 1 1 auto;
}

.glacier-terminal__cursor {
  display: inline-block;
  width: 0.55ch;
  height: 1.2em;
  background: var(--mg-cyan);
  vertical-align: -0.15em;
  animation: terminal-blink 1.05s steps(2, end) infinite;
}

.glacier-terminal__cursor--idle {
  background: var(--mg-cyan);
}

@keyframes terminal-blink {
  50% { opacity: 0; }
}

@media (prefers-reduced-motion: reduce) {
  .glacier-terminal__cursor {
    animation: none;
  }
}

/* Visually hidden but read by screen readers (per a11y best practice). */
.glacier-terminal__sr {
  position: absolute;
  left: -10000px;
  width: 1px;
  height: 1px;
  overflow: hidden;
}
</style>
