<script setup>
const props = defineProps({
  user: {
    type: Object,
    required: true
  },
  isAdmin: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['toggle-publish', 'open-admin'])

const handlePublish = () => {
  if (!props.isAdmin) return
  emit('toggle-publish')
}
</script>

<template>
  <header class="topbar">
    <div class="topbar__brand">
      <span class="topbar__logo">OpsBoard</span>
      <span class="topbar__divider" />
      <span class="topbar__section">任务大厅</span>
    </div>

    <div class="topbar__actions">
      <button
        class="topbar__publish"
        type="button"
        :class="{ 'topbar__publish--disabled': !isAdmin }"
        @click="handlePublish"
      >
        <span>发布任务</span>
        <small v-if="!isAdmin">仅管理员可发布</small>
      </button>

      <button
        class="topbar__manage"
        type="button"
        :class="{ 'topbar__manage--disabled': !isAdmin }"
        @click="isAdmin ? emit('open-admin') : null"
      >
        <span class="topbar__manage-label">管理面板</span>
        <small v-if="isAdmin" class="topbar__manage-meta">当前：{{ user.name }}</small>
        <small v-else class="topbar__manage-meta">需管理员权限</small>
      </button>
    </div>
  </header>
</template>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid var(--frost-border-strong);
  border-radius: 28px;
  padding: 18px 28px;
  backdrop-filter: blur(16px);
  box-shadow: 0 15px 40px var(--frost-shadow-strong);
  color: #fff;
}

.topbar__brand {
  display: flex;
  align-items: center;
  gap: 14px;
  font-weight: 600;
  font-size: 18px;
}

.topbar__logo {
  padding: 6px 12px;
  border-radius: 999px;
  background: var(--frost-highlight);
  color: #1f2937;
}

.topbar__divider {
  width: 1px;
  height: 18px;
  background: rgba(255, 255, 255, 0.4);
}

.topbar__section {
  color: rgba(255, 255, 255, 0.8);
}

.topbar__actions {
  display: flex;
  align-items: center;
  gap: 20px;
}

.topbar__publish {
  position: relative;
  height: 48px;
  padding: 0 24px;
  border-radius: 16px;
  border: none;
  font-weight: 600;
  font-size: 15px;
  cursor: pointer;
  color: #1f2937;
  background: var(--frost-highlight);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.topbar__publish small {
  display: block;
  font-size: 11px;
  font-weight: 500;
  color: rgba(31, 41, 55, 0.7);
}

.topbar__publish:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 32px rgba(255, 255, 255, 0.28);
}

.topbar__publish--disabled {
  background: rgba(255, 255, 255, 0.22);
  color: rgba(31, 41, 55, 0.6);
  cursor: not-allowed;
}

.topbar__manage {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 4px;
  height: 52px;
  padding: 8px 20px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease, border 0.2s ease;
}

.topbar__manage:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.18);
  background: rgba(255, 255, 255, 0.18);
}

.topbar__manage--disabled {
  cursor: not-allowed;
  pointer-events: none;
  opacity: 0.65;
  transform: none;
  box-shadow: none;
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.12);
}

.topbar__manage-label {
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.4px;
}

.topbar__manage-meta {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.78);
}

@media (max-width: 768px) {
  .topbar {
    flex-direction: column;
    align-items: stretch;
    gap: 18px;
    padding: 20px;
  }

  .topbar__actions {
    justify-content: space-between;
  }
}
</style>
