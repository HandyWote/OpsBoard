<script setup>
import { computed, onMounted, reactive, ref, watch, watchEffect } from 'vue'
import { useRouter } from 'vue-router'
import { useCurrentUser } from '../composables/useCurrentUser.js'
import { changePassword, fetchCurrentUser, updateProfile as updateProfileRequest } from '../services/users.js'
import { isAuthenticated } from '../services/http.js'

const router = useRouter()
const { profile, updateProfile, hydrate } = useCurrentUser()

const profileForm = reactive({
  name: '',
  headline: '',
  bio: ''
})

const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const savingProfile = ref(false)
const profileSuccess = ref(false)
const profileError = ref('')
const savingSecurity = ref(false)
const securityMessage = ref('')
const securityError = ref('')

watch(
  () => [profile.name, profile.headline, profile.bio],
  ([name, headline, bio]) => {
    profileForm.name = name ?? ''
    profileForm.headline = headline ?? ''
    profileForm.bio = bio ?? ''
  },
  { immediate: true }
)

const initials = computed(() => profile.avatarText || (profile.name || '').slice(0, 1) || '用')

const roleLabel = computed(() => {
  if (profile.role === 'admin') return '管理员'
  if (profile.role === 'member') return '成员'
  return profile.role || '访客'
})

const canSubmitProfile = computed(() => {
  return Boolean(profileForm.name.trim()) && !savingProfile.value
})

const canSubmitPassword = computed(() => {
  return (
    passwordForm.currentPassword.trim() &&
    passwordForm.newPassword.trim() &&
    passwordForm.newPassword === passwordForm.confirmPassword &&
    !savingSecurity.value
  )
})

const handleProfileSubmit = async () => {
  if (!canSubmitProfile.value) return

  savingProfile.value = true
  profileSuccess.value = false
  profileError.value = ''
  const payload = {
    displayName: profileForm.name.trim(),
    headline: profileForm.headline.trim(),
    bio: profileForm.bio.trim()
  }

  try {
    const updated = await updateProfileRequest(payload)
    if (updated) {
      hydrate(updated)
      profileSuccess.value = true
      setTimeout(() => {
        profileSuccess.value = false
      }, 2400)
    }
  } catch (error) {
    profileSuccess.value = false
    profileError.value = error.message || '资料保存失败'
  } finally {
    savingProfile.value = false
  }
}

const resetProfileForm = () => {
  profileForm.name = profile.name ?? ''
  profileForm.headline = profile.headline ?? ''
  profileForm.bio = profile.bio ?? ''
  profileError.value = ''
}

const handleSecuritySubmit = async () => {
  if (!canSubmitPassword.value) {
    securityError.value = '请填写完整并确保新密码两次输入一致。'
    return
  }

  savingSecurity.value = true
  securityError.value = ''
  securityMessage.value = ''

  try {
    await changePassword({
      currentPassword: passwordForm.currentPassword,
      newPassword: passwordForm.newPassword
    })
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    securityMessage.value = '密码已更新'
    setTimeout(() => {
      securityMessage.value = ''
    }, 2400)
  } catch (error) {
    securityError.value = error.message || '密码更新失败'
  } finally {
    savingSecurity.value = false
  }
}

const handleCancelSecurity = () => {
  passwordForm.currentPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  securityError.value = ''
  securityMessage.value = ''
}

onMounted(async () => {
  try {
    const latest = await fetchCurrentUser()
    if (latest) {
      hydrate(latest)
      resetProfileForm()
    }
  } catch (error) {
    /* ignore */
  }
})

watchEffect(() => {
  if (!isAuthenticated()) {
    router.replace({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
  }
})
</script>

<template>
  <div class="profile-page">
    <header class="profile-page__header">
      <button type="button" class="profile-page__back" @click="router.push({ name: 'home' })">
        ← 返回任务大厅
      </button>
      <div>
        <h1>我的资料</h1>
        <p class="profile-page__subtitle">更新昵称、自我介绍与安全设置，保持账户信息最新。</p>
      </div>
    </header>

    <div class="profile-page__grid">
      <aside class="profile-summary">
        <section class="summary-card">
          <div class="summary-card__avatar">{{ initials }}</div>
          <h2 class="summary-card__name">{{ profile.name }}</h2>
          <p v-if="profile.headline" class="summary-card__headline">{{ profile.headline }}</p>
          <div class="summary-card__chips">
            <span class="summary-chip">角色：{{ roleLabel }}</span>
            <span v-if="profile.teams?.length" class="summary-chip summary-chip--ghost">
              {{ profile.teams.join(' / ') }}
            </span>
          </div>
          <p v-if="profile.bio" class="summary-card__bio">{{ profile.bio }}</p>
        </section>

        <section class="summary-card summary-card--muted">
          <h3>安全提示</h3>
          <ul>
            <li>定期更新密码，避免与其他平台重复。</li>
            <li>如需调整权限，请联系当前管理员确认。</li>
          </ul>
        </section>
      </aside>

      <section class="profile-settings">
        <form class="settings-panel" @submit.prevent="handleProfileSubmit">
          <div class="settings-panel__header">
            <div>
              <p class="settings-panel__eyebrow">展示信息</p>
              <h2>基础资料</h2>
            </div>
            <transition name="fade">
              <span v-if="profileSuccess" class="status-pill status-pill--success">已保存</span>
            </transition>
            <transition name="fade">
              <span v-if="profileError" class="status-pill status-pill--danger">{{ profileError }}</span>
            </transition>
          </div>

          <div class="form-grid">
            <label class="form-field">
              <span>昵称</span>
              <input
                v-model="profileForm.name"
                type="text"
                name="nickname"
                placeholder="请输入展示昵称"
                autocomplete="off"
              />
            </label>
            <label class="form-field">
              <span>签名</span>
              <input
                v-model="profileForm.headline"
                type="text"
                name="headline"
                placeholder="一句话介绍自己"
                autocomplete="off"
              />
            </label>
          </div>

          <label class="form-field">
            <span>自我介绍</span>
            <textarea
              v-model="profileForm.bio"
              name="bio"
              rows="5"
              placeholder="用于任务大厅侧边栏展示，推荐写下负责范围与兴趣方向。"
            ></textarea>
          </label>

          <div class="form-actions">
            <button type="button" class="btn btn--ghost" @click="resetProfileForm">重置</button>
            <button type="submit" class="btn" :disabled="!canSubmitProfile">
              <span v-if="!savingProfile">保存资料</span>
              <span v-else class="btn__spinner" aria-hidden="true"></span>
            </button>
          </div>
        </form>

        <form class="settings-panel settings-panel--secondary" @submit.prevent="handleSecuritySubmit">
          <div class="settings-panel__header">
            <div>
              <p class="settings-panel__eyebrow">账户安全</p>
              <h2>修改密码</h2>
            </div>
            <transition name="fade">
              <span v-if="securityMessage" class="status-pill status-pill--success">{{ securityMessage }}</span>
            </transition>
            <transition name="fade">
              <span v-if="securityError" class="status-pill status-pill--danger">{{ securityError }}</span>
            </transition>
          </div>

          <div class="form-grid form-grid--stacked">
            <label class="form-field">
              <span>当前密码</span>
              <input
                v-model="passwordForm.currentPassword"
                type="password"
                name="currentPassword"
                placeholder="请输入当前密码"
                autocomplete="current-password"
              />
            </label>
            <label class="form-field">
              <span>新密码</span>
              <input
                v-model="passwordForm.newPassword"
                type="password"
                name="newPassword"
                placeholder="至少 8 位，包含数字与字母"
                autocomplete="new-password"
              />
            </label>
            <label class="form-field">
              <span>确认新密码</span>
              <input
                v-model="passwordForm.confirmPassword"
                type="password"
                name="confirmPassword"
                placeholder="再次输入新密码"
                autocomplete="new-password"
              />
            </label>
          </div>

          <div class="form-actions">
            <button type="button" class="btn btn--ghost" @click="handleCancelSecurity">取消</button>
            <button type="submit" class="btn" :disabled="!canSubmitPassword">
              <span v-if="!savingSecurity">更新密码</span>
              <span v-else class="btn__spinner" aria-hidden="true"></span>
            </button>
          </div>
        </form>
      </section>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 32px;
  color: #fff;
}

.profile-page__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.profile-page__back {
  border: none;
  border-radius: 999px;
  padding: 10px 18px;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-weight: 500;
  cursor: pointer;
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.18);
  transition: transform 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.profile-page__back:hover {
  transform: translateY(-2px);
  background: rgba(255, 255, 255, 0.22);
  box-shadow: 0 18px 36px rgba(15, 23, 42, 0.24);
}

.profile-page__subtitle {
  margin: 8px 0 0;
  color: var(--frost-text-secondary);
  font-size: 14px;
}

.profile-page__grid {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 32px;
}

.profile-summary {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.summary-card {
  position: relative;
  padding: 28px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid var(--frost-border-strong);
  backdrop-filter: blur(18px);
  box-shadow: 0 20px 48px var(--frost-shadow-light);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary-card__avatar {
  width: 72px;
  height: 72px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.92);
  color: #1f2937;
  font-size: 32px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.2);
}

.summary-card__name {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
}

.summary-card__headline {
  margin: 0;
  font-size: 14px;
  color: var(--frost-text-secondary);
}

.summary-card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.summary-chip {
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
  font-size: 12px;
  letter-spacing: 0.2px;
}

.summary-chip--ghost {
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid rgba(255, 255, 255, 0.22);
}

.summary-card__bio {
  margin: 10px 0 0;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.84);
}

.summary-card--muted {
  background: rgba(255, 255, 255, 0.08);
  border: 1px dashed rgba(255, 255, 255, 0.28);
  box-shadow: none;
}

.summary-card--muted h3 {
  margin: 0 0 12px;
  font-size: 15px;
  font-weight: 600;
}

.summary-card--muted ul {
  margin: 0;
  padding-left: 20px;
  color: rgba(255, 255, 255, 0.78);
  font-size: 13px;
  line-height: 1.7;
}

.profile-settings {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.settings-panel {
  position: relative;
  padding: 32px;
  border-radius: 26px;
  background: var(--frost-card-bg);
  border: 1px solid var(--frost-border-soft);
  box-shadow: 0 20px 52px var(--frost-shadow-light);
  display: flex;
  flex-direction: column;
  gap: 24px;
  backdrop-filter: blur(18px);
}

.settings-panel--secondary {
  background: rgba(255, 255, 255, 0.1);
}

.settings-panel__header {
  display: flex;
  align-items: center;
  gap: 16px;
  justify-content: space-between;
}

.settings-panel__eyebrow {
  margin: 0 0 6px;
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
  letter-spacing: 1.2px;
  text-transform: uppercase;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.form-grid--stacked {
  grid-template-columns: 1fr;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 10px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.82);
}

.form-field span {
  font-weight: 500;
}

.form-field input,
.form-field textarea {
  width: 100%;
  padding: 14px 16px;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.24);
  background: rgba(27, 39, 63, 0.45);
  color: #fff;
  font-size: 14px;
  line-height: 1.5;
  transition: border 0.2s ease, box-shadow 0.2s ease;
}

.form-field input::placeholder,
.form-field textarea::placeholder {
  color: rgba(255, 255, 255, 0.52);
}

.form-field input:focus,
.form-field textarea:focus {
  outline: none;
  border-color: rgba(255, 255, 255, 0.6);
  box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.18);
}

.form-field textarea {
  min-height: 140px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn {
  min-width: 120px;
  height: 44px;
  border-radius: 16px;
  border: none;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.92), rgba(255, 255, 255, 0.7));
  color: #1f2937;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, filter 0.2s ease;
}

.btn:hover:enabled {
  transform: translateY(-1px);
  box-shadow: 0 12px 28px rgba(255, 255, 255, 0.35);
}

.btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
  box-shadow: none;
}

.btn--ghost {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.26);
}

.btn__spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(31, 41, 55, 0.4);
  border-top-color: rgba(31, 41, 55, 0.9);
  border-radius: 50%;
  display: inline-block;
  animation: spin 0.6s linear infinite;
}

.status-pill {
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.status-pill--success {
  background: rgba(153, 229, 255, 0.18);
  color: rgba(222, 249, 255, 0.95);
  border: 1px solid rgba(153, 229, 255, 0.4);
}

.status-pill--danger {
  background: rgba(255, 118, 132, 0.18);
  color: rgba(255, 214, 219, 0.95);
  border: 1px solid rgba(255, 118, 132, 0.4);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1024px) {
  .profile-page__grid {
    grid-template-columns: 1fr;
  }

  .profile-summary {
    flex-direction: row;
    flex-wrap: wrap;
  }

  .summary-card {
    flex: 1 1 260px;
  }
}

@media (max-width: 768px) {
  .profile-page__header {
    flex-direction: column-reverse;
    align-items: flex-start;
  }

  .profile-page__grid {
    gap: 24px;
  }

  .settings-panel {
    padding: 24px;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
