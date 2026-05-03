<!--
  UsedInTasksBadges - on a package page, lists task pages that compose this
  package. Required at the top of every package page per spec 0031.
  Falls back to "Used in tasks: none" when no current task page composes it.
-->
<script setup lang="ts">
import { withBase } from 'vitepress'
import { computed } from 'vue'
import { tasks, type TaskMeta } from '../../data/sidebar'

const props = defineProps<{
  packageName: string
}>()

const items = computed<readonly TaskMeta[]>(() => {
  return tasks.filter((t) => t.packagesUsed.includes(props.packageName))
})
</script>

<template>
  <p class="glacier-task-row">
    <span class="glacier-task-row__label">Used in tasks:</span>
    <template v-if="items.length > 0">
      <a
        v-for="task in items"
        :key="task.slug"
        class="glacier-task-row__chip"
        :href="withBase(`/docs/${task.slug}`)"
      >
        {{ task.title }}
      </a>
    </template>
    <span v-else class="glacier-task-row__none">none</span>
  </p>
</template>

<style scoped>
.glacier-task-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.6rem;
  margin: 0.4rem 0 1.6rem;
  font-size: 0.92rem;
}

.glacier-task-row__label {
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 0.76rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--mg-text-muted);
  margin-right: 0.2rem;
}

.glacier-task-row__chip {
  font-family: var(--mg-font-body);
  font-size: 0.86rem;
  font-weight: 500;
  padding: 0.25rem 0.7rem;
  border-radius: 999px;
  border: 1px solid var(--mg-border);
  background: var(--mg-surface);
  color: var(--mg-cyan);
  text-decoration: none !important;
  transition: border-color 120ms ease, transform 120ms ease;
}

.glacier-task-row__chip:hover {
  transform: translateY(-1px);
  border-color: var(--mg-cyan);
}

.glacier-task-row__none {
  font-family: var(--mg-font-body);
  font-size: 0.86rem;
  color: var(--mg-text-faint);
  font-style: italic;
}
</style>
