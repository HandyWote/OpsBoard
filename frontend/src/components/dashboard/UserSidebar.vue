<script setup>
import { computed } from 'vue'

const props = defineProps({
  user: {
    type: Object,
    required: true
  },
  pendingTasks: {
    type: Array,
    default: () => []
  },
  availableTasks: {
    type: Array,
    default: () => []
  }
})

const initials = computed(() => (props.user.name ? props.user.name.slice(0, 1) : '用'))
const pendingPreview = computed(() => props.pendingTasks.slice(0, 3))
const availablePreview = computed(() => props.availableTasks.slice(0, 3))

const roleLabel = computed(() => {
  if (props.user.role === 'admin') return '管理员'
  if (props.user.role === 'member') return '成员'
  return props.user.role || '访客'
})

const formatDeadline = (deadline) => {
  try {
    return new Date(deadline).toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch (err) {
    return deadline
  }
}
</script>

<template>
  <aside class="user-sidebar">
    <section class="profile-card">
      <div class="avatar">{{ initials }}</div>
      <div class="profile-meta">
        <p class="greeting">你好，{{ user.name }}</p>
        <p class="role">角色：{{ roleLabel }}</p>
      </div>
    </section>

    <section class="quick-stats">
      <div class="stat">
        <span class="label">待执行任务</span>
        <span class="value">{{ pendingTasks.length }}</span>
      </div>
      <div class="stat">
        <span class="label">可领取任务</span>
        <span class="value">{{ availableTasks.length }}</span>
      </div>
    </section>

    <section class="task-block">
      <h3>我的待办</h3>
      <ol v-if="pendingTasks.length" class="task-list">
        <li v-for="task in pendingPreview" :key="task.id">
          <p class="task-title">{{ task.title }}</p>
          <p class="task-deadline">截止：{{ formatDeadline(task.deadline) }}</p>
        </li>
      </ol>
      <p v-else class="empty">当前没有待执行任务</p>
      <p v-if="pendingTasks.length > pendingPreview.length" class="more">
        还有 {{ pendingTasks.length - pendingPreview.length }} 项待处理
      </p>
    </section>

    <section v-if="availablePreview.length" class="task-block">
      <h3>推荐任务</h3>
      <ul class="task-list">
        <li v-for="task in availablePreview" :key="task.id">
          <p class="task-title">{{ task.title }}</p>
          <p class="task-deadline">截止：{{ formatDeadline(task.deadline) }}</p>
        </li>
      </ul>
    </section>
  </aside>
</template>

<style scoped>
.user-sidebar {
  min-width: 280px;
  max-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 24px;
  border-radius: 24px;
  background: var(--frost-card-bg);
  border: 1px solid var(--frost-border-soft);
  backdrop-filter: blur(18px);
  box-shadow: 0 18px 40px var(--frost-shadow-light);
}

.profile-card {
  display: flex;
  align-items: center;
  gap: 16px;
}

.avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.9);
  color: #1f2937;
  font-size: 24px;
  font-weight: 600;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.18);
}

.profile-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.greeting {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: -0.2px;
}

.role {
  margin: 0;
  font-size: 14px;
  color: var(--frost-text-secondary);
}

.quick-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.stat {
  padding: 14px;
  border-radius: 18px;
  background: var(--frost-muted-strong);
  border: 1px solid var(--frost-border-soft);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.label {
  font-size: 12px;
  color: var(--frost-text-secondary);
  letter-spacing: 0.2px;
}

.value {
  font-size: 22px;
  font-weight: 600;
}

.task-block {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-block h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.task-list {
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  list-style: none;
}

.task-title {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
}

.task-deadline {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--frost-text-secondary);
}

.empty {
  margin: 0;
  font-size: 14px;
  color: var(--frost-text-secondary);
}

.more {
  margin: 0;
  font-size: 12px;
  color: var(--frost-text-secondary);
}

@media (max-width: 1100px) {
  .user-sidebar {
    max-width: none;
    width: 100%;
  }
}
</style>
