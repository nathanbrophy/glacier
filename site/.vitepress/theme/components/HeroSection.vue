<!--
  HeroSection — landing-hero composition.
  Stacks: wordmark, tagline, (mascot + terminal side-by-side), CTAs.
  Mascot state is driven by the terminal via provide/inject (HERO_MASCOT_STATE).
-->
<script setup lang="ts">
import HeroTerminal from './HeroTerminal.vue'
import MascotKaomoji from './MascotKaomoji.vue'
import WordmarkSVG from './WordmarkSVG.vue'
import { heroScript } from '../../../tests/fixtures/hero-script.fixture'
</script>

<template>
  <section class="glacier-hero-section">
    <WordmarkSVG class="glacier-hero__wordmark" size="hero" />

    <p class="glacier-hero__tagline">Less plumbing. More Go.</p>
    <p class="glacier-hero__lede">
      A curated Go framework. Idiomatic handler code stays on your side. The
      generic plumbing (flags, config, lifecycle, mocks, transport, terminal,
      observability) stays on Glacier&rsquo;s.
    </p>

    <div class="glacier-hero">
      <div class="glacier-hero__mascot-col">
        <MascotKaomoji />
      </div>
      <div class="glacier-hero__terminal-col">
        <HeroTerminal :script="heroScript" />
      </div>
    </div>

    <div class="glacier-hero__cta">
      <a class="primary" href="/glacier/docs/">Read the docs &rarr;</a>
      <a class="secondary" href="https://github.com/nathanbrophy/glacier" target="_blank" rel="noopener">
        View on GitHub
      </a>
    </div>
  </section>
</template>

<style scoped>
.glacier-hero-section {
  max-width: 1200px;
  margin: 0 auto;
  padding: 4rem 1.5rem 5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.25rem;
  text-align: center;
}

@media (min-width: 768px) {
  .glacier-hero-section {
    padding: 5rem 2rem 6rem;
  }
}
</style>
