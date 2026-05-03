<!--
  MermaidZoom - sits invisibly in the layout shell. After each route
  change, scans the DOM for mermaid diagrams and makes them
  click-to-fullscreen. Opens a modal overlay containing the SVG cloned
  to viewport size; Escape or background click closes.
-->
<script setup lang="ts">
import { useRoute } from 'vitepress'
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const route = useRoute()
const zoomedSvg = ref<string | null>(null)

const SELECTOR = '.mermaid, [id^="mermaid-"]'

function close() {
  zoomedSvg.value = null
}

function enhance() {
  const els = document.querySelectorAll<HTMLElement>(SELECTOR)
  els.forEach((el) => {
    if (el.dataset.zoomEnhanced === '1') return
    if (!el.querySelector('svg')) return
    el.dataset.zoomEnhanced = '1'
    el.classList.add('glacier-mermaid')
    el.setAttribute('role', 'button')
    el.setAttribute('tabindex', '0')
    el.setAttribute('title', 'Click to expand')

    const open = () => {
      const svg = el.querySelector('svg')
      if (!svg) return
      // Mermaid SVGs ship with explicit width/height/style that constrain
      // them inside the modal. Clone, strip those attributes so the SVG
      // scales to fill our viewport-sized container.
      const clone = svg.cloneNode(true) as SVGElement
      clone.removeAttribute('width')
      clone.removeAttribute('height')
      clone.removeAttribute('style')
      // Ensure preserveAspectRatio so it scales nicely.
      if (!clone.getAttribute('preserveAspectRatio')) {
        clone.setAttribute('preserveAspectRatio', 'xMidYMid meet')
      }
      zoomedSvg.value = clone.outerHTML
    }

    el.addEventListener('click', open)
    el.addEventListener('keydown', (e: KeyboardEvent) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault()
        open()
      }
    })
  })
}

function scheduleEnhance() {
  // Mermaid renders asynchronously after Vue mounts; allow a beat.
  setTimeout(enhance, 80)
  setTimeout(enhance, 400)
  setTimeout(enhance, 1200)
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && zoomedSvg.value !== null) {
    close()
  }
}

onMounted(() => {
  scheduleEnhance()
  window.addEventListener('keydown', onKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
})

watch(
  () => route.path,
  () => {
    nextTick(scheduleEnhance)
  },
)
</script>

<template>
  <Teleport to="body">
    <Transition name="glacier-mermaid-fade">
      <div
        v-if="zoomedSvg !== null"
        class="glacier-mermaid-modal"
        role="dialog"
        aria-modal="true"
        aria-label="Diagram, click outside or press Escape to close"
        @click="close"
      >
        <button
          class="glacier-mermaid-modal__close"
          type="button"
          aria-label="Close diagram"
          @click.stop="close"
        >
          &times;
        </button>
        <div
          class="glacier-mermaid-modal__svg"
          @click.stop
          v-html="zoomedSvg"
        ></div>
      </div>
    </Transition>
  </Teleport>
</template>

<style>
/* Inline-rendered mermaid block: hint that it's clickable */
.glacier-mermaid {
  cursor: zoom-in;
  border-radius: 10px;
  border: 1px solid var(--mg-border);
  background: var(--mg-surface);
  padding: 1rem;
  margin: 1.5rem auto;
  max-width: 1100px;
  transition: border-color 150ms ease, transform 150ms ease;
}

.glacier-mermaid:hover,
.glacier-mermaid:focus-visible {
  border-color: var(--mg-cyan);
  transform: translateY(-1px);
  outline: none;
}

.glacier-mermaid svg {
  display: block;
  margin: 0 auto;
  max-width: 100%;
  height: auto;
}

/* Fullscreen modal */
.glacier-mermaid-modal {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(14, 17, 22, 0.92);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem 2rem;
  cursor: zoom-out;
}

.glacier-mermaid-modal__svg {
  width: min(1800px, 96vw);
  height: min(1100px, 90vh);
  background: var(--mg-bg);
  border: 1px solid var(--mg-border);
  border-radius: 14px;
  padding: 2rem;
  overflow: auto;
  cursor: default;
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
}

.glacier-mermaid-modal__svg svg {
  width: 100% !important;
  height: 100% !important;
  max-width: 100% !important;
  max-height: 100% !important;
  display: block;
}

.glacier-mermaid-modal__close {
  position: absolute;
  top: 1.2rem;
  right: 1.5rem;
  width: 44px;
  height: 44px;
  border-radius: 999px;
  background: var(--mg-surface);
  border: 1px solid var(--mg-border);
  color: var(--mg-text);
  font-size: 1.5rem;
  font-family: var(--mg-font-display);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  transition: border-color 150ms ease, color 150ms ease, transform 150ms ease;
}

.glacier-mermaid-modal__close:hover {
  border-color: var(--mg-cyan);
  color: var(--mg-cyan);
  transform: rotate(90deg);
}

.glacier-mermaid-fade-enter-active,
.glacier-mermaid-fade-leave-active {
  transition: opacity 200ms ease;
}

.glacier-mermaid-fade-enter-from,
.glacier-mermaid-fade-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .glacier-mermaid {
    transition: none;
  }
  .glacier-mermaid-modal__close {
    transition: none;
  }
  .glacier-mermaid-fade-enter-active,
  .glacier-mermaid-fade-leave-active {
    transition: none;
  }
}
</style>
