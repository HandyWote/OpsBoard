<script setup>
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { login } from '../services/auth.js'

const form = reactive({
  username: '',
  password: ''
})

const touched = reactive({
  username: false,
  password: false
})

const focused = reactive({
  username: false,
  password: false
})

const errors = reactive({
  username: '',
  password: ''
})

const feedback = reactive({
  type: '',
  message: ''
})

const loading = ref(false)
const visible = ref(false)

const canSubmit = computed(() => !loading.value && !errors.username && !errors.password && form.username && form.password)

const validateField = (field) => {
  if (!form[field]?.trim()) {
    errors[field] = field === 'username' ? '请输入用户名' : '请输入密码'
  } else {
    errors[field] = ''
  }
}

const handleFocus = (field) => {
  focused[field] = true
}

const handleBlur = (field) => {
  focused[field] = false
  touched[field] = true
  validateField(field)
}

const handleInput = (field) => {
  if (touched[field]) {
    validateField(field)
  }
}

const resetFeedback = () => {
  feedback.type = ''
  feedback.message = ''
}

const submit = async () => {
  touched.username = true
  touched.password = true
  validateField('username')
  validateField('password')

  if (errors.username || errors.password) {
    return
  }

  loading.value = true
  resetFeedback()

  try {
    const response = await login({
      username: form.username.trim(),
      password: form.password
    })

    if (response.success) {
      feedback.type = 'success'
      feedback.message = response.message || '登录成功，正在为您跳转...'
      await nextTick()
      setTimeout(() => {
        window.location.href = '/'
      }, 1200)
    } else {
      feedback.type = 'error'
      feedback.message = response.message || '用户名或密码错误'
    }
  } catch (error) {
    feedback.type = 'error'
    feedback.message = error.message || '服务暂时不可用，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  requestAnimationFrame(() => {
    visible.value = true
  })
})
</script>

<template>
  <section class="login" :class="{ 'login--visible': visible }" aria-live="polite">
    <header class="login__header">
      <h1 class="login__title">欢迎回来</h1>
      <p class="login__subtitle">请使用您的账户登录</p>
    </header>

    <form class="login__form" @submit.prevent="submit" novalidate>
      <div
        class="field"
        :class="{
          'field--active': focused.username || !!form.username,
          'field--error': !!errors.username
        }"
      >
        <label class="field__label" for="username">用户名</label>
        <input
          id="username"
          v-model="form.username"
          class="field__input"
          type="text"
          name="username"
          autocomplete="username"
          :disabled="loading"
          @focus="handleFocus('username')"
          @blur="handleBlur('username')"
          @input="handleInput('username')"
        />
        <span v-if="errors.username" class="field__message">{{ errors.username }}</span>
      </div>

      <div
        class="field"
        :class="{
          'field--active': focused.password || !!form.password,
          'field--error': !!errors.password
        }"
      >
        <label class="field__label" for="password">密码</label>
        <input
          id="password"
          v-model="form.password"
          class="field__input"
          type="password"
          name="password"
          autocomplete="current-password"
          :disabled="loading"
          @focus="handleFocus('password')"
          @blur="handleBlur('password')"
          @input="handleInput('password')"
        />
        <span v-if="errors.password" class="field__message">{{ errors.password }}</span>
      </div>

      <button class="login__submit" type="submit" :disabled="!canSubmit">
        <span v-if="!loading">登录</span>
        <span v-else class="login__spinner" aria-hidden="true"></span>
      </button>
    </form>

    <p v-if="feedback.message" class="login__feedback" :class="`login__feedback--${feedback.type}`">
      {{ feedback.message }}
    </p>
  </section>
</template>

<style scoped>
.login {
  width: 100%;
  max-width: 320px;
  border-radius: 24px;
  padding: 32px 28px;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(15px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  color: #fff;
  display: flex;
  flex-direction: column;
  gap: 28px;
  opacity: 0;
  transform: translateY(24px);
  transition: opacity 0.6s ease, transform 0.6s ease;
}

.login--visible {
  opacity: 1;
  transform: translateY(0);
}

.login__header {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.login__title {
  font-size: 28px;
  font-weight: 600;
  letter-spacing: -0.5px;
  margin: 0;
}

.login__subtitle {
  margin: 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.75);
}

.login__form {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.field {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field__label {
  position: absolute;
  inset-inline-start: 16px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 15px;
  color: rgba(255, 255, 255, 0.65);
  pointer-events: none;
  transition: all 0.3s ease;
}

.field--active .field__label {
  top: 10px;
  font-size: 12px;
  transform: translateY(0);
}

.field__input {
  width: 100%;
  height: 52px;
  padding: 20px 16px 8px;
  font-size: 16px;
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
  transition: border-color 0.3s ease, background 0.3s ease, box-shadow 0.3s ease;
  box-sizing: border-box;
}

.field__input:focus {
  border-color: rgba(189, 224, 254, 0.85);
  background: rgba(255, 255, 255, 0.2);
  outline: none;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.12);
}

.field--error .field__input {
  border-color: rgba(255, 114, 118, 0.95);
}

.field--error .field__label {
  color: rgba(255, 114, 118, 0.95);
}

.field__message {
  font-size: 12px;
  color: rgba(255, 114, 118, 0.95);
  margin-inline-start: 4px;
}

.login__submit {
  position: relative;
  width: 100%;
  height: 52px;
  border: none;
  border-radius: 16px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.85), rgba(255, 255, 255, 0.65));
  color: #1f2937;
  font-weight: 600;
  font-size: 16px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.3s ease;
}

.login__submit:hover:not(:disabled) {
  transform: scale(1.05);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.75));
  box-shadow: 0 12px 30px rgba(31, 45, 72, 0.2);
}

.login__submit:active:not(:disabled) {
  transform: scale(0.95);
}

.login__submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login__spinner {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2px solid rgba(31, 45, 72, 0.4);
  border-top-color: rgba(31, 45, 72, 0.9);
  display: inline-block;
  animation: spin 0.9s linear infinite;
}

.login__feedback {
  font-size: 14px;
  text-align: center;
  margin: -10px 0 0;
}

.login__feedback--error {
  color: rgba(255, 114, 118, 0.95);
}

.login__feedback--success {
  color: rgba(189, 224, 254, 0.95);
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 480px) {
  .login {
    padding: 28px 22px;
    gap: 22px;
  }

  .login__title {
    font-size: 24px;
  }
}
</style>
