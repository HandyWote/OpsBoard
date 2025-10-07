<script setup>
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'

const route = useRoute()

const contentClass = computed(() => {
  if (route.name === 'home') {
    return 'app-shell__content app-shell__content--wide'
  }
  return 'app-shell__content app-shell__content--compact'
})
</script>

<template>
  <div class="app-shell">
    <div class="app-shell__background" />
    <main :class="contentClass">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(circle at 20% 20%, rgba(148, 187, 233, 0.65), rgba(238, 174, 202, 0.4)),
    linear-gradient(135deg, #5b86e5, #36d1dc);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  overflow: hidden;
  padding: 32px 16px;
}

.app-shell__background {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 80% 20%, rgba(255, 255, 255, 0.35), transparent 55%),
    radial-gradient(circle at 10% 90%, rgba(255, 255, 255, 0.25), transparent 60%);
  filter: blur(40px);
  opacity: 0.8;
  pointer-events: none;
  transform: scale(1.1);
}

.app-shell__content {
  position: relative;
  width: 100%;
  display: flex;
  justify-content: center;
  z-index: 1;
  transition: max-width 0.4s ease;
}

.app-shell__content--compact {
  max-width: 420px;
}

.app-shell__content--wide {
  max-width: 1080px;
}
</style>
