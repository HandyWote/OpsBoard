<script setup>
import { onMounted, ref } from 'vue'

const visible = ref(false)

const spotlightCards = [
  {
    title: '任务总览',
    description: '实时掌握运维任务进度，快速响应高优事件。',
    accent: 'rgba(255, 255, 255, 0.6)'
  },
  {
    title: '待办提醒',
    description: '聚合个人待处理事项，减少上下文切换。',
    accent: 'rgba(186, 230, 253, 0.65)'
  },
  {
    title: '知识库',
    description: '沉淀最佳实践与应急手册，团队协作更高效。',
    accent: 'rgba(221, 214, 254, 0.65)'
  }
]

onMounted(() => {
  requestAnimationFrame(() => {
    visible.value = true
  })
})
</script>

<template>
  <section class="home" :class="{ 'home--visible': visible }">
    <header class="home__header">
      <h1 class="home__title">OpsBoard 控制台</h1>
      <p class="home__subtitle">欢迎回来，立即开始管理你的任务与协作。</p>
      <div class="home__actions">
        <button type="button" class="home__btn home__btn--primary">发布新任务</button>
        <button type="button" class="home__btn home__btn--ghost">查看任务池</button>
      </div>
    </header>

    <section class="grid">
      <article v-for="card in spotlightCards" :key="card.title" class="grid__card">
        <header class="grid__card-header">
          <span class="grid__indicator" :style="{ background: card.accent }"></span>
          <h2 class="grid__card-title">{{ card.title }}</h2>
        </header>
        <p class="grid__card-desc">
          {{ card.description }}
        </p>
        <button type="button" class="grid__card-action">了解详情</button>
      </article>
    </section>
  </section>
</template>

<style scoped>
.home {
  width: 100%;
  max-width: 960px;
  padding: 48px;
  border-radius: 36px;
  background: rgba(255, 255, 255, 0.14);
  backdrop-filter: blur(18px);
  border: 1px solid rgba(255, 255, 255, 0.25);
  box-shadow: 0 20px 60px rgba(15, 23, 42, 0.18);
  color: #fff;
  display: flex;
  flex-direction: column;
  gap: 40px;
  opacity: 0;
  transform: translateY(30px);
  transition: opacity 0.65s ease, transform 0.65s ease;
}

.home--visible {
  opacity: 1;
  transform: translateY(0);
}

.home__header {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.home__title {
  margin: 0;
  font-size: 36px;
  font-weight: 700;
  letter-spacing: -0.8px;
}

.home__subtitle {
  margin: 0;
  font-size: 18px;
  color: rgba(255, 255, 255, 0.75);
  max-width: 520px;
}

.home__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.home__btn {
  height: 48px;
  padding: 0 24px;
  border-radius: 16px;
  border: none;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.3s ease;
}

.home__btn--primary {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.7));
  color: #1f2937;
  box-shadow: 0 14px 30px rgba(255, 255, 255, 0.18);
}

.home__btn--primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 18px 40px rgba(255, 255, 255, 0.2);
}

.home__btn--ghost {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.25);
}

.home__btn--ghost:hover {
  transform: translateY(-2px);
  background: rgba(255, 255, 255, 0.18);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 24px;
}

.grid__card {
  padding: 28px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.15), 0 12px 32px rgba(15, 23, 42, 0.16);
  display: flex;
  flex-direction: column;
  gap: 18px;
  transition: transform 0.3s ease, box-shadow 0.3s ease, background 0.3s ease;
}

.grid__card:hover {
  background: rgba(255, 255, 255, 0.18);
  transform: translateY(-6px);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.22);
}

.grid__card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.grid__indicator {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  box-shadow: 0 0 20px currentColor;
}

.grid__card-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.grid__card-desc {
  margin: 0;
  font-size: 15px;
  color: rgba(255, 255, 255, 0.75);
  line-height: 1.6;
}

.grid__card-action {
  align-self: flex-start;
  padding: 10px 18px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  background: rgba(255, 255, 255, 0.85);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.grid__card:hover .grid__card-action {
  transform: translateX(6px);
  box-shadow: 0 12px 24px rgba(255, 255, 255, 0.2);
}

@media (max-width: 1024px) {
  .home {
    padding: 36px;
  }

  .home__title {
    font-size: 32px;
  }
}

@media (max-width: 640px) {
  .home {
    padding: 28px;
    gap: 32px;
  }

  .home__title {
    font-size: 28px;
  }

  .home__actions {
    flex-direction: column;
  }

  .grid {
    gap: 16px;
  }
}
</style>
