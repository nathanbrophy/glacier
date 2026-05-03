<!--
  PromiseSection - the four spec 0001 §The Promise statements rendered verbatim.
  Anchored at #promise so spec 0001 §Test Matrix "Promise satisfaction" has a
  committed home (per spec 0031 D-S20).
-->
<script setup lang="ts">
const promises: readonly { kaomoji: string; text: string }[] = [
  { kaomoji: 'ʕ•ᴥ•ʔ',  text: "I'm only writing what's mine."             },
  { kaomoji: 'ʕ⌐■-■ʔ', text: 'I trust the defaults.'                      },
  { kaomoji: 'ʕ•_•ʔ',  text: 'The error tells me what to do next.'        },
  { kaomoji: 'ʕ◉_◉ʔ',  text: 'Tests are easy because the framework helps.'},
]
</script>

<template>
  <section id="promise" class="glacier-promise">
    <header class="glacier-promise__head">
      <p class="glacier-promise__kicker">The Promise</p>
      <h2 class="glacier-promise__heading">
        Four sentences a Glacier user should say truthfully.
      </h2>
    </header>
    <ul class="glacier-promise__list">
      <li v-for="p in promises" :key="p.text" class="glacier-promise__item">
        <span class="glacier-promise__kaomoji" aria-hidden="true">{{ p.kaomoji }}</span>
        <span class="glacier-promise__quote">{{ p.text }}</span>
      </li>
    </ul>
    <p class="glacier-promise__caption">
      Every component in Glacier is reviewed against this list. If a design
      stops a user from saying one of these, the design is wrong.
    </p>
  </section>
</template>

<style scoped>
.glacier-promise {
  max-width: 1100px;
  margin: 4rem auto;
  padding: 3rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

@media (min-width: 768px) {
  .glacier-promise {
    padding: 4rem 2rem;
  }
}

.glacier-promise__head {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.glacier-promise__kicker {
  margin: 0;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 0.78rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--mg-cyan);
}

.glacier-promise__heading {
  margin: 0;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: clamp(1.5rem, 3vw, 2.1rem);
  letter-spacing: -0.01em;
  color: var(--mg-text);
  max-width: 36ch;
  line-height: 1.2;
}

.glacier-promise__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.9rem;
  grid-template-columns: 1fr;
}

@media (min-width: 768px) {
  .glacier-promise__list {
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
}

.glacier-promise__item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.2rem;
  background: var(--mg-surface);
  border: 1px solid var(--mg-border);
  border-left: 2px solid var(--mg-cyan);
  border-radius: 10px;
}

.glacier-promise__kaomoji {
  font-family: var(--mg-font-mono);
  font-size: 1.4rem;
  color: var(--mg-cyan);
  flex: 0 0 auto;
  min-width: 4ch;
  text-align: center;
}

.glacier-promise__quote {
  font-family: var(--mg-font-display);
  font-weight: 500;
  font-size: 1.05rem;
  color: var(--mg-text);
  line-height: 1.4;
}

.glacier-promise__caption {
  margin: 0;
  font-size: 0.92rem;
  color: var(--mg-text-muted);
  line-height: 1.6;
  max-width: 60ch;
}
</style>
