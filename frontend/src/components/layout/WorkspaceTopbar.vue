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

const emit = defineEmits(['toggle-publish'])

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

      <div class="topbar__user">
        <span class="topbar__avatar">{{ user.name?.slice(0, 1) ?? 'U' }}</span>
        <div class="topbar__user-info">
          <strong>{{ user.name }}</strong>
          <small>{{ isAdmin ? '管理员' : '成员' }}</small>
        </div>
      </div>
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

.topbar__user {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
}

.topbar__avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.85);
  color: #1f2937;
  font-weight: 700;
}

.topbar__user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 13px;
}

.topbar__user-info strong {
  font-size: 14px;
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
