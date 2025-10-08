<script setup>
defineProps({
  taskCount: {
    type: Number,
    default: 0
  }
})
</script>

<template>
  <section class="hero">
    <div>
      <h1>今日共有 {{ taskCount }} 个任务等待认领</h1>
      <p>按优先级和截止时间排序，指派后请在 SLA 内完成并回传复盘。</p>
    </div>
    <div class="hero__cta">
      <button class="hero__primary" type="button">快速筛选高优任务</button>
      <button class="hero__secondary" type="button">查看执行准则</button>
    </div>
  </section>
</template>

<style scoped>
.hero {
  width: 100%;
  padding: 36px 40px;
  border-radius: 32px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid var(--frost-border-soft);
  backdrop-filter: blur(18px);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  box-shadow: 0 18px 50px var(--frost-shadow-strong);
}

.hero h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.hero p {
  margin: 8px 0 0;
  font-size: 15px;
  color: var(--frost-text-secondary);
}

.hero__cta {
  display: flex;
  gap: 16px;
}

.hero__primary,
.hero__secondary {
  height: 46px;
  padding: 0 24px;
  border-radius: 16px;
  border: none;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.hero__primary {
  background: var(--frost-highlight);
  color: #1f2937;
}

.hero__primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 32px rgba(255, 255, 255, 0.24);
}

.hero__secondary {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  border: 1px solid var(--frost-border-soft);
}

.hero__secondary:hover {
  transform: translateY(-2px);
  background: rgba(255, 255, 255, 0.2);
}

@media (max-width: 1024px) {
  .hero {
    flex-direction: column;
    align-items: flex-start;
  }

  .hero__cta {
    width: 100%;
    justify-content: flex-start;
  }
}

@media (max-width: 768px) {
  .hero {
    border-radius: 28px;
    padding: 20px;
  }
}
</style>
