<!--
  Custom layout — wraps DefaultTheme.Layout and renders AuroraBackdrop
  behind everything (positioned via .glacier-shell).
-->
<script setup lang="ts">
import DefaultTheme from 'vitepress/theme'
import AuroraBackdrop from './components/AuroraBackdrop.vue'
import MascotSprite from './components/MascotSprite.vue'
import MermaidZoom from './components/MermaidZoom.vue'

const { Layout } = DefaultTheme
</script>

<template>
  <div class="glacier-shell">
    <AuroraBackdrop />
    <Layout>
      <template #layout-bottom>
        <div class="glacier-footer-companion" aria-hidden="true">
          <MascotSprite state="idle" :size="56" />
        </div>
      </template>
    </Layout>
    <MermaidZoom />
  </div>
</template>

<style scoped>
.glacier-footer-companion {
  display: flex;
  justify-content: center;
  padding: 1rem 0 2rem;
  opacity: 0.85;
}
</style>
