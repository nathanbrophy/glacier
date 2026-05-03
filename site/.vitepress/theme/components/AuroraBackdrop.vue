<!--
  AuroraBackdrop — slow-drifting cool gradient field behind the hero.
  Spec 0031 §Architecture / Visual system: 30s loop, opacity ≤ 0.18,
  reduced-motion freezes the animation.
-->
<template>
  <div class="glacier-aurora" aria-hidden="true">
    <div class="glacier-aurora__layer glacier-aurora__layer--a"></div>
    <div class="glacier-aurora__layer glacier-aurora__layer--b"></div>
  </div>
</template>

<style scoped>
.glacier-aurora {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  background: var(--mg-bg);
}

.glacier-aurora__layer {
  position: absolute;
  inset: -25%;
  filter: blur(80px);
  opacity: 0.16;
  will-change: transform;
}

.glacier-aurora__layer--a {
  background:
    radial-gradient(40% 30% at 25% 30%, var(--mg-cyan)   0%, transparent 60%),
    radial-gradient(35% 25% at 70% 60%, var(--mg-cyan-100) 0%, transparent 65%);
  animation: aurora-drift-a 32s ease-in-out infinite alternate;
}

.glacier-aurora__layer--b {
  background:
    radial-gradient(45% 35% at 80% 25%, var(--mg-teal)    0%, transparent 60%),
    radial-gradient(35% 25% at 20% 70%, var(--mg-cyan-700) 0%, transparent 65%);
  animation: aurora-drift-b 38s ease-in-out infinite alternate;
  mix-blend-mode: screen;
}

@keyframes aurora-drift-a {
  0%   { transform: translate3d(-3%, -2%, 0) rotate(-2deg); }
  100% { transform: translate3d( 4%,  3%, 0) rotate( 3deg); }
}

@keyframes aurora-drift-b {
  0%   { transform: translate3d( 3%,  2%, 0) rotate( 2deg); }
  100% { transform: translate3d(-4%, -3%, 0) rotate(-3deg); }
}

@media (prefers-reduced-motion: reduce) {
  .glacier-aurora__layer {
    animation: none !important;
  }
}
</style>
