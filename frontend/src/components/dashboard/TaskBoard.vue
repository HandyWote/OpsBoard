<script setup>
import TaskCard from './TaskCard.vue'

defineProps({
  tasks: {
    type: Array,
    default: () => []
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
</script>

<template>
  <section class="board">
    <TaskCard
      v-for="task in tasks"
      :key="task.id"
      :task="task"
      :priority-meta="priorityMeta"
      :current-user-name="currentUserName"
      @accept="emit('accept', $event)"
      @release="emit('release', $event)"
    />

    <p v-if="!tasks.length" class="board__empty">暂无符合条件的任务，试试更换筛选条件。</p>
  </section>
</template>

<style scoped>
.board {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.board__empty {
  margin: 0;
  padding: 28px;
  border-radius: 24px;
  text-align: center;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid var(--frost-border-soft);
  color: var(--frost-text-secondary);
}
</style>
