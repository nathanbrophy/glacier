<!--
  SdkSpotlight - prominent landing-page block introducing the Glacier SDK
  binary. Pairs with the framework story: framework is the library suite,
  the SDK is the CLI that ships every framework feature in one binary.
-->
<script setup lang="ts">
import { withBase } from 'vitepress'

const installCmd = 'go install github.com/nathanbrophy/glacier/cmd/glacier@latest'
const initCmd = 'glacier init my-app --yes'

const sdkLink = withBase('/sdk/')
const installLink = withBase('/sdk/install')
</script>

<template>
  <section class="sdk-spotlight">
    <header class="sdk-spotlight__head">
      <p class="sdk-spotlight__kicker">The Glacier SDK</p>
      <h2 class="sdk-spotlight__heading">
        One binary. Nine commands. Every framework package, dogfooded.
      </h2>
      <p class="sdk-spotlight__lede">
        The <code>glacier</code> CLI is the public face of the framework. It
        scaffolds projects, runs the three code generators, lints, tests with
        a streaming summary and bench gating, and explains every marker, exit
        code, and config key to you as you go.
      </p>
    </header>

    <div class="sdk-spotlight__grid">
      <div class="sdk-spotlight__card sdk-spotlight__card--install">
        <p class="sdk-spotlight__card-kicker">Install</p>
        <pre class="sdk-spotlight__cmd"><span class="sdk-spotlight__prompt">$</span> {{ installCmd }}</pre>
        <p class="sdk-spotlight__card-kicker">Scaffold</p>
        <pre class="sdk-spotlight__cmd"><span class="sdk-spotlight__prompt">$</span> {{ initCmd }}</pre>
        <p class="sdk-spotlight__card-out">
          <span class="sdk-spotlight__kaomoji">ʕ⌐■-■ʔ</span>
          all set. cd my-app &amp;&amp; glacier test
        </p>
      </div>

      <div class="sdk-spotlight__card sdk-spotlight__card--commands">
        <p class="sdk-spotlight__card-kicker">CREATE</p>
        <ul class="sdk-spotlight__cmds">
          <li><code>glacier init</code> &mdash; scaffold a new project, signals + banner + version + completions wired</li>
          <li><code>glacier new</code> &mdash; add a package, command, or option to an existing project</li>
        </ul>
        <p class="sdk-spotlight__card-kicker">DEVELOP</p>
        <ul class="sdk-spotlight__cmds">
          <li><code>glacier generate</code> &mdash; run cli, mock, httpmock generators concurrently</li>
          <li><code>glacier lint</code> &mdash; gofmt + vet + staticcheck + 6 Glacier lints with auto-fix</li>
          <li><code>glacier test</code> &mdash; <code>go test -json</code> + streaming summary + bench gating</li>
        </ul>
        <p class="sdk-spotlight__card-kicker">INSPECT &amp; UTILITY</p>
        <ul class="sdk-spotlight__cmds">
          <li><code>glacier explain</code> &middot; <code>version</code> &middot; <code>vibe</code> &middot; <code>completions</code></li>
        </ul>
      </div>
    </div>

    <div class="sdk-spotlight__cta">
      <a class="sdk-spotlight__primary" :href="sdkLink">Read the SDK docs &rarr;</a>
      <a class="sdk-spotlight__secondary" :href="installLink">Install instructions</a>
    </div>

    <p class="sdk-spotlight__caption">
      Every byte of the SDK is built on Glacier framework packages. If a package
      isn&rsquo;t ergonomic enough for the SDK, that surfaces as a paper cut against
      the framework, not as private SDK plumbing.
    </p>
  </section>
</template>

<style scoped>
.sdk-spotlight {
  max-width: 1100px;
  margin: 4rem auto;
  padding: 3rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  position: relative;
}

@media (min-width: 768px) {
  .sdk-spotlight {
    padding: 4rem 2rem;
  }
}

.sdk-spotlight__head {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.sdk-spotlight__kicker {
  margin: 0;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 0.78rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--mg-cyan);
}

.sdk-spotlight__heading {
  margin: 0;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: clamp(1.5rem, 3vw, 2.1rem);
  letter-spacing: -0.01em;
  color: var(--mg-text);
  max-width: 38ch;
  line-height: 1.2;
}

.sdk-spotlight__lede {
  margin: 0;
  color: var(--mg-text-muted);
  font-size: 1rem;
  line-height: 1.6;
  max-width: 64ch;
}

.sdk-spotlight__lede code {
  font-family: var(--mg-font-mono);
  font-size: 0.9em;
  background: var(--mg-surface-2);
  padding: 0.05em 0.3em;
  border-radius: 4px;
  color: var(--mg-cyan);
}

.sdk-spotlight__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.2rem;
}

@media (min-width: 900px) {
  .sdk-spotlight__grid {
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
  }
}

.sdk-spotlight__card {
  background: var(--mg-surface);
  border: 1px solid var(--mg-border);
  border-radius: 12px;
  padding: 1.4rem 1.4rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.sdk-spotlight__card--install {
  border-left: 2px solid var(--mg-cyan);
}

.sdk-spotlight__card-kicker {
  margin: 0.4rem 0 0.2rem;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 0.7rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--mg-cyan);
}

.sdk-spotlight__card-kicker:first-child {
  margin-top: 0;
}

.sdk-spotlight__cmd {
  margin: 0;
  font-family: var(--mg-font-mono);
  font-size: 0.92rem;
  color: var(--mg-text);
  background: var(--mg-bg);
  border: 1px solid var(--mg-border);
  border-radius: 8px;
  padding: 0.7rem 0.9rem;
  overflow-x: auto;
  white-space: pre;
}

.sdk-spotlight__prompt {
  color: var(--mg-text-faint);
  user-select: none;
  margin-right: 0.4rem;
}

.sdk-spotlight__card-out {
  margin: 0.4rem 0 0;
  font-family: var(--mg-font-mono);
  font-size: 0.92rem;
  color: var(--mg-cyan);
}

.sdk-spotlight__kaomoji {
  margin-right: 0.5rem;
}

.sdk-spotlight__cmds {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  font-size: 0.95rem;
  color: var(--mg-text-muted);
  line-height: 1.5;
}

.sdk-spotlight__cmds code {
  font-family: var(--mg-font-mono);
  font-size: 0.88em;
  background: var(--mg-surface-2);
  padding: 0.05em 0.3em;
  border-radius: 4px;
  color: var(--mg-cyan);
}

.sdk-spotlight__cta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.8rem;
  margin-top: 0.5rem;
}

.sdk-spotlight__primary,
.sdk-spotlight__secondary {
  display: inline-flex;
  align-items: center;
  padding: 0.65rem 1.1rem;
  font-family: var(--mg-font-display);
  font-weight: 600;
  font-size: 0.95rem;
  border-radius: 8px;
  text-decoration: none;
  transition: transform 160ms ease, border-color 160ms ease, background 160ms ease;
}

.sdk-spotlight__primary {
  background: var(--mg-cyan);
  color: var(--mg-bg);
  border: 1px solid var(--mg-cyan);
}

.sdk-spotlight__primary:hover {
  transform: translateY(-1px);
  background: #5fdaee;
  border-color: #5fdaee;
}

.sdk-spotlight__secondary {
  background: transparent;
  color: var(--mg-text);
  border: 1px solid var(--mg-border);
}

.sdk-spotlight__secondary:hover {
  border-color: var(--mg-cyan);
  color: var(--mg-cyan);
}

.sdk-spotlight__caption {
  margin: 0;
  font-size: 0.92rem;
  color: var(--mg-text-muted);
  line-height: 1.6;
  max-width: 60ch;
}
</style>
