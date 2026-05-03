<!--
  Custom 404 page. Replaces VitePress's default NotFound layout.
  Uses the wave companion sprite per spec 0031 §Architecture / Visual
  system / Mascot.
-->
<script setup lang="ts">
import { withBase } from 'vitepress'
import MascotSprite from './MascotSprite.vue'

const homeHref = withBase('/')
const docsHref = withBase('/docs/')
</script>

<template>
  <div class="glacier-404">
    <MascotSprite state="wave" :size="160" />
    <h1 class="glacier-404__code">404</h1>
    <p class="glacier-404__lede">
      You took a turn into the snow. The page you were after is not here.
    </p>
    <div class="glacier-404__cta">
      <a class="primary" :href="homeHref">Back to landing</a>
      <a class="secondary" :href="docsHref">Browse the docs</a>
    </div>
  </div>
</template>

<style scoped>
.glacier-404 {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.25rem;
  padding: 5rem 1.5rem 6rem;
  min-height: 60vh;
  text-align: center;
}

.glacier-404__code {
  margin: 0;
  font-family: var(--mg-font-mono);
  font-weight: 700;
  font-size: clamp(3.5rem, 8vw, 5.5rem);
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
  line-height: 1;
}

.glacier-404__lede {
  margin: 0;
  font-family: var(--mg-font-display);
  font-weight: 500;
  font-size: clamp(1rem, 2vw, 1.18rem);
  color: var(--mg-text-muted);
  max-width: 44ch;
  line-height: 1.5;
}

.glacier-404__cta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.glacier-404__cta a {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1.1rem;
  border-radius: 8px;
  font-family: var(--mg-font-body);
  font-weight: 600;
  font-size: 0.95rem;
  text-decoration: none;
  transition: transform 120ms ease, background 120ms ease, color 120ms ease;
}

.glacier-404__cta .primary {
  background: var(--mg-cyan);
  color: var(--mg-bg);
}

.glacier-404__cta .primary:hover {
  background: var(--mg-teal);
  transform: translateY(-1px);
}

.glacier-404__cta .secondary {
  background: transparent;
  color: var(--mg-text);
  border: 1px solid var(--mg-border);
}

.glacier-404__cta .secondary:hover {
  border-color: var(--mg-cyan);
  color: var(--mg-cyan);
}
</style>
