<script setup>
import { computed } from 'vue'

const props = defineProps({
  task: {
    type: Object,
    required: true
  },
  priorityMeta: {
    type: Object,
    default: () => ({})
  },
  currentUserName: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['accept', 'release'])

const meta = computed(() => props.priorityMeta[props.task.priority] || { label: '普通', tone: 'rgba(255,255,255,0.85)' })

const deadlineLabel = computed(() => {
  const date = new Date(props.task.deadline)
  if (Number.isNaN(date.getTime())) return '未设定'
  return date.toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
})

const isOwner = computed(() => props.task.assignee === props.currentUserName)

const handleAccept = () => {
  if (props.task.status !== 'available') return
  emit('accept', props.task)
}

const handleRelease = () => {
  if (!isOwner.value) return
  emit('release', props.task)
}
</script>

<template>
  <article class="card" :class="`card--${task.status}`">
    <header class="card__header">
      <span class="card__id">#{{ task.id }}</span>
      <span class="card__priority" :style="{ '--priority-tone': meta.tone }">
        {{ meta.label }}
      </span>
    </header>

    <h2 class="card__title">{{ task.title }}</h2>
    <p class="card__summary">{{ task.summary }}</p>

    <footer class="card__meta">
      <div class="card__chips">
        <span class="chip chip--reward">赏金 ¥{{ task.reward }}</span>
        <span class="chip chip--deadline">截止 {{ deadlineLabel }}</span>
        <span v-for="tag in task.tags" :key="tag" class="chip chip--tag">#{{ tag }}</span>
      </div>

      <div class="card__controls">
        <div v-if="task.status === 'claimed'" class="card__assignee">
          <span class="assignee__label">执行人</span>
          <span class="assignee__name">{{ task.assignee }}</span>
        </div>

        <button v-if="task.status === 'available'" class="card__btn" type="button" @click="handleAccept">
          认领任务
        </button>

        <button
          v-else-if="task.status === 'claimed' && isOwner"
          class="card__btn card__btn--ghost"
          type="button"
          @click="handleRelease"
        >
          释放任务
        </button>

        <span v-else class="card__status">{{ task.assignee ? `${task.assignee} 已认领` : '进行中' }}</span>
      </div>
    </footer>
  </article>
</template>

<style scoped>
.card {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 24px;
  border-radius: 26px;
  background: var(--frost-card-bg);
  border: 1px solid var(--frost-border-soft);
  backdrop-filter: blur(12px);
  color: #fff;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.12), 0 12px 30px var(--frost-shadow-light);
  transition: transform 0.25s ease, box-shadow 0.25s ease, background 0.25s ease;
}

.card:hover {
  transform: translateY(-6px);
  background: var(--frost-card-hover);
  box-shadow: 0 18px 42px var(--frost-shadow-strong);
}

.card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
}

.card__id {
  color: rgba(255, 255, 255, 0.7);
}

.card__priority {
  padding: 4px 10px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--priority-tone);
  font-weight: 600;
  font-size: 12px;
}

.card__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.card__summary {
  margin: 0;
  font-size: 14px;
  color: var(--frost-text-secondary);
  line-height: 1.6;
}

.card__meta {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 14px;
  font-size: 13px;
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid rgba(255, 255, 255, 0.18);
}

.chip--reward {
  color: rgba(255, 255, 255, 0.9);
}

.chip--deadline {
  color: rgba(189, 224, 254, 0.95);
}

.chip--tag {
  color: rgba(255, 255, 255, 0.7);
}

.card__controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.card__assignee {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 13px;
}

.assignee__label {
  color: rgba(255, 255, 255, 0.55);
}

.assignee__name {
  font-weight: 600;
}

.card__btn {
  height: 42px;
  padding: 0 20px;
  border-radius: 14px;
  border: none;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  color: #1f2937;
  background: var(--frost-highlight);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.card__btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 28px rgba(255, 255, 255, 0.26);
}

.card__btn--ghost {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.28);
}

.card__status {
  font-size: 13px;
  color: var(--frost-text-secondary);
}

@media (max-width: 768px) {
  .card {
    border-radius: 22px;
    padding: 20px;
  }
}
</style>
