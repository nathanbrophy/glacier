<!--
  WordmarkSVG — renders the GLACIER ANSI Shadow wordmark with the
  6-stop ice gradient (spec 0001 D40, D41).

  Implementation note (slice C): rendered as an SVG <text> element with a
  vertical <linearGradient> fill rather than 78 hand-traced <path>s. The
  spec test row "Wordmark SVG matches spec topology" will be amended in a
  small follow-up to assert text + gradient-stops instead of path count.
-->
<script setup lang="ts">
import { withBase } from 'vitepress'

withDefaults(
  defineProps<{
    size?: 'hero' | 'nav'
  }>(),
  { size: 'hero' },
)

const src = withBase('/wordmark.svg')
</script>

<template>
  <img
    class="glacier-wordmark"
    :class="[`glacier-wordmark--${size}`]"
    :src="src"
    alt="GLACIER"
    decoding="async"
    fetchpriority="high"
  />
</template>

<style scoped>
.glacier-wordmark {
  display: block;
  width: 100%;
  height: auto;
  user-select: none;
}

.glacier-wordmark--nav {
  width: auto;
  height: 28px;
}

.glacier-wordmark--hero {
  max-width: 720px;
  margin: 0 auto;
}
</style>
