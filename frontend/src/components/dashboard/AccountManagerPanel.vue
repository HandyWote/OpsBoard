<script setup>
const props = defineProps({
  accounts: {
    type: Array,
    default: () => []
  },
  currentUserId: {
    type: String,
    default: ''
  },
  adminCount: {
    type: Number,
    default: 1
  }
})

const emit = defineEmits(['close', 'toggle-admin'])

const roleLabel = (role) => (role === 'admin' ? '管理员' : '成员')

const roleTone = (role) => (role === 'admin' ? 'rgba(255, 255, 255, 0.22)' : 'rgba(255, 255, 255, 0.16)')

const isLastAdmin = (account) => account.role === 'admin' && props.adminCount <= 1
</script>

<template>
  <div class="manager-overlay" @click.self="emit('close')">
    <section class="manager-panel">
      <header class="manager-header">
        <div>
          <h2>账号管理面板</h2>
          <p>查看账号详情，并在此授予或撤销管理员权限。</p>
        </div>
        <button type="button" class="manager-header__close" @click="emit('close')">✕</button>
      </header>

      <ul class="account-list">
        <li v-for="account in accounts" :key="account.id">
          <div
            class="account-row"
            :class="{ 'account-row--self': account.id === currentUserId }"
          >
            <div class="account-row__main">
              <span class="account-row__avatar">{{ account.name ? account.name.slice(0, 1) : '用' }}</span>
              <div class="account-row__info">
                <strong>{{ account.name }}</strong>
                <span class="account-row__email">{{ account.email }}</span>
                <span v-if="account.teams?.length" class="account-row__teams">
                  {{ account.teams.join(' · ') }}
                </span>
              </div>
            </div>
            <div class="account-row__meta">
              <span class="account-row__role" :style="{ background: roleTone(account.role) }">
                {{ roleLabel(account.role) }}
              </span>
              <span v-if="account.id === currentUserId" class="account-row__self-tag">当前账号</span>
              <button
                type="button"
                class="account-row__toggle"
                :class="{ 'account-row__toggle--danger': account.role === 'admin' }"
                :disabled="isLastAdmin(account)"
                :aria-disabled="isLastAdmin(account)"
                :title="isLastAdmin(account) ? '至少保留一名管理员' : ''"
                @click="emit('toggle-admin', account.id)"
              >
                {{ account.role === 'admin' ? '撤销管理员' : '授予管理员' }}
              </button>
            </div>
          </div>
        </li>
      </ul>

      <footer class="manager-footer">
        <button type="button" class="manager-footer__close" @click="emit('close')">关闭</button>
      </footer>
    </section>
  </div>
</template>

<style scoped>
.manager-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.5);
  backdrop-filter: blur(12px);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 24px;
  z-index: 10;
}

.manager-panel {
  width: min(580px, 100%);
  max-height: 100%;
  padding: 32px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid var(--frost-border-soft);
  color: #fff;
  box-shadow: 0 30px 60px var(--frost-shadow-strong);
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.manager-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.manager-header h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
}

.manager-header p {
  margin: 6px 0 0;
  font-size: 14px;
  color: var(--frost-text-secondary);
}

.manager-header__close {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  cursor: pointer;
}

.manager-header__close:hover {
  background: rgba(255, 255, 255, 0.26);
}

.account-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
}

.account-row {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 18px 20px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.08);
  color: inherit;
  transition: box-shadow 0.2s ease, border 0.2s ease, background 0.2s ease;
}

.account-row:hover {
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.22);
  border-color: rgba(255, 255, 255, 0.28);
  background: rgba(255, 255, 255, 0.12);
}

.account-row__avatar {
  width: 46px;
  height: 46px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.9);
  color: #1f2937;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 600;
}

.account-row__main {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  min-width: 0;
}

.account-row__info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: left;
  font-size: 14px;
  min-width: 0;
}

.account-row__info strong {
  font-size: 15px;
  letter-spacing: 0.3px;
}

.account-row__email {
  color: var(--frost-text-secondary);
}

.account-row__teams {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.76);
}

.account-row__meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.account-row__role {
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.4px;
  color: #0f172a;
  background: rgba(255, 255, 255, 0.16);
}

.account-row__self-tag {
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  color: rgba(31, 41, 55, 0.88);
  background: rgba(255, 255, 255, 0.4);
}

.account-row__toggle {
  height: 34px;
  padding: 0 14px;
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.24);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.4px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease, border 0.2s ease;
}

.account-row__toggle:hover {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.18);
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.18);
}

.account-row__toggle--danger {
  border-color: rgba(255, 118, 132, 0.42);
  background: rgba(255, 118, 132, 0.2);
}

.account-row__toggle--danger:hover {
  background: rgba(255, 118, 132, 0.32);
  box-shadow: 0 10px 24px rgba(255, 118, 132, 0.32);
}

.account-row__toggle:disabled {
  cursor: not-allowed;
  opacity: 0.6;
  transform: none;
  box-shadow: none;
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.12);
}

.account-row--self {
  border-color: rgba(255, 255, 255, 0.48);
  box-shadow: 0 18px 36px rgba(15, 23, 42, 0.32);
}

.manager-footer {
  display: flex;
  justify-content: flex-end;
}

.manager-footer__close {
  height: 42px;
  padding: 0 20px;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.28);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  cursor: pointer;
}

.manager-footer__close:hover {
  background: rgba(255, 255, 255, 0.18);
}

@media (max-width: 720px) {
  .manager-panel {
    padding: 24px;
    gap: 20px;
  }

  .account-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .account-row__role {
    align-self: flex-start;
  }

  .account-row__meta {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
