<!--
  PillarCard - one of the four "why Glacier" pillars.
  Used on / and /why.
-->
<script setup lang="ts">
defineProps<{
  title: string
  headline: string
}>()
</script>

<template>
  <article class="glacier-pillar">
    <h3 class="glacier-pillar__title">{{ title }}</h3>
    <p class="glacier-pillar__headline">{{ headline }}</p>
    <div class="glacier-pillar__details">
      <slot />
    </div>
  </article>
</template>

<style scoped>
.glacier-pillar {
  background: var(--mg-surface);
  border: 1px solid var(--mg-border);
  border-radius: 12px;
  padding: 1.5rem 1.4rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  transition: border-color 160ms ease, transform 160ms ease;
  height: 100%;
}

.glacier-pillar:hover {
  border-color: var(--mg-cyan);
  transform: translateY(-2px);
}

.glacier-pillar__title {
  margin: 0;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 1.05rem;
  letter-spacing: -0.01em;
  color: var(--mg-cyan);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-size: 0.78rem;
}

.glacier-pillar__headline {
  margin: 0;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 1.18rem;
  letter-spacing: -0.01em;
  color: var(--mg-text);
  line-height: 1.3;
}

.glacier-pillar__details {
  font-size: 0.95rem;
  color: var(--mg-text-muted);
  line-height: 1.55;
}

.glacier-pillar__details :deep(p) {
  margin: 0;
}

.glacier-pillar__details :deep(p + p) {
  margin-top: 0.5rem;
}

.glacier-pillar__details :deep(code) {
  font-family: var(--mg-font-mono);
  font-size: 0.85em;
  background: var(--mg-surface-2);
  padding: 0.05em 0.3em;
  border-radius: 4px;
  color: var(--mg-cyan);
}
</style>
