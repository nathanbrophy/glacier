<!--
  CodeCompare - side-by-side "without Glacier" vs "with Glacier" code blocks.
  Used on / to demonstrate the framework's value concretely.
-->
<script setup lang="ts">
withDefaults(
  defineProps<{
    withTitle?: string
    withoutTitle?: string
  }>(),
  {
    withTitle: 'With Glacier',
    withoutTitle: 'Without Glacier',
  },
)
</script>

<template>
  <section class="glacier-compare">
    <div class="glacier-compare__col glacier-compare__col--without">
      <header>
        <span class="glacier-compare__kicker">{{ withoutTitle }}</span>
      </header>
      <div class="glacier-compare__code">
        <slot name="without" />
      </div>
    </div>
    <div class="glacier-compare__col glacier-compare__col--with">
      <header>
        <span class="glacier-compare__kicker">{{ withTitle }}</span>
      </header>
      <div class="glacier-compare__code">
        <slot name="with" />
      </div>
    </div>
  </section>
</template>

<style scoped>
.glacier-compare {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.2rem;
}

@media (min-width: 900px) {
  .glacier-compare {
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
  }
}

.glacier-compare__col {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  background: var(--mg-surface);
  border: 1px solid var(--mg-border);
  border-radius: 10px;
  overflow: hidden;
}

.glacier-compare__col header {
  padding: 0.7rem 1rem;
  border-bottom: 1px solid var(--mg-border);
  background: var(--mg-surface-2);
}

.glacier-compare__kicker {
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 0.74rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--mg-text-muted);
}

.glacier-compare__col--with .glacier-compare__kicker {
  color: var(--mg-cyan);
}

.glacier-compare__col--with {
  border-color: rgba(34, 211, 238, 0.4);
}

.glacier-compare__code :deep(div[class*='language-']) {
  margin: 0;
  border-radius: 0;
  background: var(--mg-bg);
}

.glacier-compare__code :deep(pre) {
  background: var(--mg-bg) !important;
  margin: 0;
  padding: 1rem 1.2rem;
}
</style>
