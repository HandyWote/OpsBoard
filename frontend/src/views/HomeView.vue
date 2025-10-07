<script setup>
import { computed, reactive, ref } from 'vue'

const currentUser = reactive({
  name: '林运维',
  role: 'admin'
})

const sortKey = ref('priority')
const keyword = ref('')
const showPublishPanel = ref(false)
const submitting = ref(false)

const priorityOrder = {
  critical: 1,
  high: 2,
  medium: 3,
  low: 4
}

const tasks = reactive([
  {
    id: 'T-1024',
    title: '校园网节点流量突增排查',
    summary: '定位晚高峰教室区域流量异常，补充监控指标并撰写复盘。',
    reward: 220,
    deadline: '2025-02-18T18:00:00+08:00',
    priority: 'critical',
    tags: ['网络', '应急'],
    status: 'available',
    assignee: null
  },
  {
    id: 'T-1018',
    title: '公共机房巡检自动化脚本优化',
    summary: '更新巡检脚本以兼容新批次服务器，补充告警推送配置。',
    reward: 150,
    deadline: '2025-02-25T12:00:00+08:00',
    priority: 'high',
    tags: ['自动化', '脚本'],
    status: 'claimed',
    assignee: 'Jerry'
  },
  {
    id: 'T-0991',
    title: '知识库：应急通信回落流程整理',
    summary: '整理 2024 年紧急回落流程并绘制流程图，更新到知识库。',
    reward: 120,
    deadline: '2025-03-02T09:00:00+08:00',
    priority: 'medium',
    tags: ['文档', '知识库'],
    status: 'available',
    assignee: null
  },
  {
    id: 'T-0977',
    title: 'Nginx 配置管理策略梳理',
    summary: '收敛 Nginx 配置，输出灰度规范，并同步给发布系统。',
    reward: 180,
    deadline: '2025-02-23T20:00:00+08:00',
    priority: 'high',
    tags: ['发布', '架构'],
    status: 'available',
    assignee: null
  }
])

const publishForm = reactive({
  title: '',
  reward: '',
  deadline: '',
  tags: '',
  description: ''
})

const isAdmin = computed(() => currentUser.role === 'admin')

const filteredTasks = computed(() => {
  const term = keyword.value.trim().toLowerCase()
  const copy = tasks
    .filter((task) => {
      if (!term) return true
      return `${task.id} ${task.title} ${task.summary}`.toLowerCase().includes(term)
    })
    .slice()

  const availableFirst = (task) => (task.status === 'available' ? 0 : task.status === 'claimed' ? 1 : 2)

  if (sortKey.value === 'priority') {
    return copy.sort((a, b) => {
      const statusCompare = availableFirst(a) - availableFirst(b)
      if (statusCompare !== 0) return statusCompare
      return (priorityOrder[a.priority] || 5) - (priorityOrder[b.priority] || 5)
    })
  }

  if (sortKey.value === 'deadline') {
    return copy.sort((a, b) => {
      const statusCompare = availableFirst(a) - availableFirst(b)
      if (statusCompare !== 0) return statusCompare
      return new Date(a.deadline).getTime() - new Date(b.deadline).getTime()
    })
  }

  return copy
})

const priorityMeta = {
  critical: { label: '特急', tone: 'var(--danger)' },
  high: { label: '高', tone: 'var(--warning)' },
  medium: { label: '中', tone: 'var(--info)' },
  low: { label: '低', tone: 'var(--muted)' }
}

const humanDeadline = (iso) => {
  const date = new Date(iso)
  return date.toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const handleAccept = (task) => {
  if (task.status !== 'available') return
  task.status = 'claimed'
  task.assignee = currentUser.name
}

const handleRelease = (task) => {
  if (task.assignee !== currentUser.name) return
  task.status = 'available'
  task.assignee = null
}

const togglePublishPanel = () => {
  if (!isAdmin.value) return
  showPublishPanel.value = !showPublishPanel.value
}

const resetForm = () => {
  publishForm.title = ''
  publishForm.reward = ''
  publishForm.deadline = ''
  publishForm.tags = ''
  publishForm.description = ''
}

const handleDescriptionInput = (event) => {
  publishForm.description = event.currentTarget.innerHTML
}

const submitTask = () => {
  if (!publishForm.title.trim() || !publishForm.description.trim()) {
    return
  }

  submitting.value = true
  setTimeout(() => {
    tasks.unshift({
      id: `T-${Math.floor(Math.random() * 9000 + 1000)}`,
      title: publishForm.title.trim(),
      summary: publishForm.description.replace(/<[^>]+>/g, '').slice(0, 140) || '新任务',
      reward: Number(publishForm.reward) || 0,
      deadline: publishForm.deadline || new Date().toISOString(),
      priority: 'medium',
      tags: publishForm.tags.split(',').map((tag) => tag.trim()).filter(Boolean),
      status: 'available',
      assignee: null
    })

    resetForm()
    submitting.value = false
    showPublishPanel.value = false
  }, 450)
}
</script>

<template>
  <div class="workspace">
    <header class="workspace__topbar">
      <div class="workspace__brand">
        <span class="workspace__logo">OpsBoard</span>
        <span class="workspace__divider" />
        <span class="workspace__section">任务大厅</span>
      </div>

      <div class="workspace__actions">
        <button
          class="workspace__publish"
          :class="{ 'workspace__publish--disabled': !isAdmin }"
          type="button"
          @click="togglePublishPanel"
        >
          <span>发布任务</span>
          <small v-if="!isAdmin">仅管理员可发布</small>
        </button>

        <div class="workspace__user">
          <span class="workspace__avatar">{{ currentUser.name.slice(0, 1) }}</span>
          <div class="workspace__user-info">
            <strong>{{ currentUser.name }}</strong>
            <small>{{ isAdmin ? '管理员' : '成员' }}</small>
          </div>
        </div>
      </div>
    </header>

    <section class="workspace__hero">
      <div class="hero-card">
        <div>
          <h1>今日共有 {{ filteredTasks.length }} 个任务等待认领</h1>
          <p>按优先级和截止时间排序，指派后请按 SLA 内完成并同步复盘。</p>
        </div>
        <div class="hero-card__cta">
          <button class="hero-card__primary" type="button">快速筛选高优任务</button>
          <button class="hero-card__secondary" type="button">查看执行准则</button>
        </div>
      </div>
    </section>

    <section class="workspace__filters">
      <div class="search">
        <span class="search__icon">🔍</span>
        <input v-model.trim="keyword" type="search" placeholder="搜索任务编号、标题或描述" />
      </div>

      <div class="filters">
        <label class="filters__label">排序</label>
        <select v-model="sortKey">
          <option value="priority">优先级</option>
          <option value="deadline">截止时间</option>
        </select>
      </div>
    </section>

    <section class="task-board">
      <article v-for="task in filteredTasks" :key="task.id" class="task-card" :class="`task-card--${task.status}`">
        <header class="task-card__header">
          <span class="task-card__id">#{{ task.id }}</span>
          <span class="task-card__priority" :style="{ '--priority-tone': priorityMeta[task.priority]?.tone }">
            {{ priorityMeta[task.priority]?.label || '普通' }}
          </span>
        </header>

        <h2 class="task-card__title">{{ task.title }}</h2>
        <p class="task-card__summary">{{ task.summary }}</p>

        <footer class="task-card__meta">
          <div class="task-card__chips">
            <span class="chip chip--reward">赏金 ¥{{ task.reward }}</span>
            <span class="chip chip--deadline">截止 {{ humanDeadline(task.deadline) }}</span>
            <span v-for="tag in task.tags" :key="tag" class="chip chip--tag">#{{ tag }}</span>
          </div>

          <div class="task-card__controls">
            <div v-if="task.status === 'claimed'" class="task-card__assignee">
              <span class="assignee__label">执行人</span>
              <span class="assignee__name">{{ task.assignee }}</span>
            </div>

            <button
              v-if="task.status === 'available'"
              class="task-card__btn"
              type="button"
              @click="handleAccept(task)"
            >
              认领任务
            </button>

            <button
              v-else-if="task.status === 'claimed' && task.assignee === currentUser.name"
              class="task-card__btn task-card__btn--ghost"
              type="button"
              @click="handleRelease(task)"
            >
              释放任务
            </button>

            <span v-else class="task-card__status">{{ task.assignee ? `${task.assignee} 已认领` : '进行中' }}</span>
          </div>
        </footer>
      </article>
    </section>

    <transition name="publish">
      <aside v-if="showPublishPanel" class="publish-panel">
        <header class="publish-panel__header">
          <h2>发布新任务</h2>
          <button type="button" class="publish-panel__close" @click="togglePublishPanel">✕</button>
        </header>

        <form class="publish-form" @submit.prevent="submitTask">
          <label class="publish-form__field">
            <span>任务标题</span>
            <input v-model="publishForm.title" type="text" placeholder="例如：机房巡检脚本修复" required />
          </label>

          <div class="publish-form__grid">
            <label class="publish-form__field">
              <span>赏金 (¥)</span>
              <input v-model="publishForm.reward" type="number" min="0" step="10" placeholder="200" />
            </label>

            <label class="publish-form__field">
              <span>截止时间</span>
              <input v-model="publishForm.deadline" type="datetime-local" />
            </label>
          </div>

          <label class="publish-form__field">
            <span>标签 (逗号分隔)</span>
            <input v-model="publishForm.tags" type="text" placeholder="网络, 自动化" />
          </label>

          <label class="publish-form__field">
            <span>任务说明</span>
            <div
              class="publish-form__editor"
              contenteditable
              placeholder="请输入任务背景、目标、交付标准，可粘贴截图或拖拽图片。"
              @input="handleDescriptionInput"
            ></div>
          </label>

          <div class="publish-form__toolbar">
            <button type="button" @click="document.execCommand('bold', false)">
              <span>加粗</span>
            </button>
            <button type="button" @click="document.execCommand('italic', false)">
              <span>斜体</span>
            </button>
            <button type="button" @click="document.execCommand('insertUnorderedList', false)">
              <span>列表</span>
            </button>
            <label class="publish-form__upload">
              <span>上传附件</span>
              <input type="file" hidden multiple />
            </label>
          </div>

          <div class="publish-form__actions">
            <button type="button" class="publish-form__btn publish-form__btn--ghost" @click="togglePublishPanel">
              取消
            </button>
            <button type="submit" class="publish-form__btn" :disabled="submitting">
              <span v-if="!submitting">发布任务</span>
              <span v-else>发布中...</span>
            </button>
          </div>
        </form>
      </aside>
    </transition>
  </div>
</template>

<style scoped>
:global(:root) {
  --danger: rgba(255, 118, 132, 0.9);
  --warning: rgba(255, 201, 99, 0.92);
  --info: rgba(153, 229, 255, 0.9);
  --muted: rgba(255, 255, 255, 0.45);
}

.workspace {
  width: 100%;
  max-width: 1200px;
  display: flex;
  flex-direction: column;
  gap: 32px;
  z-index: 1;
}

.workspace__topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid rgba(255, 255, 255, 0.24);
  border-radius: 28px;
  padding: 18px 28px;
  backdrop-filter: blur(16px);
  box-shadow: 0 15px 40px rgba(17, 25, 40, 0.18);
  color: #fff;
}

.workspace__brand {
  display: flex;
  align-items: center;
  gap: 14px;
  font-weight: 600;
  font-size: 18px;
}

.workspace__logo {
  padding: 6px 12px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(255, 255, 255, 0.65));
  color: #1f2937;
}

.workspace__divider {
  width: 1px;
  height: 18px;
  background: rgba(255, 255, 255, 0.4);
}

.workspace__section {
  color: rgba(255, 255, 255, 0.8);
}

.workspace__actions {
  display: flex;
  align-items: center;
  gap: 20px;
}

.workspace__publish {
  position: relative;
  height: 48px;
  padding: 0 24px;
  border-radius: 16px;
  border: none;
  font-weight: 600;
  font-size: 15px;
  cursor: pointer;
  color: #1f2937;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.7));
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.workspace__publish small {
  display: block;
  font-size: 11px;
  font-weight: 500;
  color: rgba(31, 41, 55, 0.7);
}

.workspace__publish:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 32px rgba(255, 255, 255, 0.28);
}

.workspace__publish--disabled {
  background: rgba(255, 255, 255, 0.22);
  color: rgba(31, 41, 55, 0.6);
  cursor: not-allowed;
}

.workspace__user {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
}

.workspace__avatar {
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

.workspace__user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 13px;
}

.workspace__user-info strong {
  font-size: 14px;
}

.workspace__hero {
  display: flex;
}

.hero-card {
  width: 100%;
  padding: 36px 40px;
  border-radius: 32px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.22);
  backdrop-filter: blur(18px);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  box-shadow: 0 18px 50px rgba(15, 23, 42, 0.18);
}

.hero-card h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.hero-card p {
  margin: 8px 0 0;
  font-size: 15px;
  color: rgba(255, 255, 255, 0.75);
}

.hero-card__cta {
  display: flex;
  gap: 16px;
}

.hero-card__primary,
.hero-card__secondary {
  height: 46px;
  padding: 0 24px;
  border-radius: 16px;
  border: none;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.hero-card__primary {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.7));
  color: #1f2937;
}

.hero-card__primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 32px rgba(255, 255, 255, 0.24);
}

.hero-card__secondary {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.18);
}

.hero-card__secondary:hover {
  transform: translateY(-2px);
  background: rgba(255, 255, 255, 0.2);
}

.workspace__filters {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.search {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 18px;
  padding: 0 16px;
  backdrop-filter: blur(14px);
  color: #fff;
}

.search input {
  width: 100%;
  height: 46px;
  border: none;
  background: transparent;
  color: inherit;
  font-size: 15px;
  padding-left: 8px;
}

.search input::placeholder {
  color: rgba(255, 255, 255, 0.6);
}

.search input:focus {
  outline: none;
}

.search__icon {
  font-size: 16px;
}

.filters {
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 0 12px;
  border-radius: 16px;
  height: 46px;
  backdrop-filter: blur(12px);
}

.filters select {
  background: transparent;
  border: none;
  color: #fff;
  font-size: 14px;
}

.filters select:focus {
  outline: none;
}

  .task-board {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

.task-card {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 24px;
  border-radius: 26px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(12px);
  color: #fff;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.12), 0 12px 30px rgba(15, 23, 42, 0.16);
  transition: transform 0.25s ease, box-shadow 0.25s ease, background 0.25s ease;
}

.task-card:hover {
  transform: translateY(-6px);
  background: rgba(255, 255, 255, 0.18);
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.22);
}

.task-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
}

.task-card__id {
  color: rgba(255, 255, 255, 0.7);
}

.task-card__priority {
  padding: 4px 10px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.14);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--priority-tone, rgba(255, 255, 255, 0.85));
  font-weight: 600;
  font-size: 12px;
}

.task-card__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.task-card__summary {
  margin: 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.75);
  line-height: 1.6;
}

.task-card__meta {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.task-card__chips {
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

.task-card__controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.task-card__assignee {
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

.task-card__btn {
  height: 42px;
  padding: 0 20px;
  border-radius: 14px;
  border: none;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  color: #1f2937;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.7));
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.task-card__btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 28px rgba(255, 255, 255, 0.26);
}

.task-card__btn--ghost {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.28);
}

.task-card__btn--ghost:hover {
  box-shadow: 0 12px 24px rgba(255, 255, 255, 0.22);
}

.task-card__status {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.75);
}

.publish-enter-active,
.publish-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.publish-enter-from,
.publish-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

.publish-panel {
  position: fixed;
  inset: 0;
  display: flex;
  justify-content: flex-end;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(12px);
  padding: 48px;
  z-index: 5;
}

.publish-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
}

.publish-panel__header h2 {
  margin: 0;
  font-size: 22px;
}

.publish-panel__close {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  cursor: pointer;
}

.publish-panel__close:hover {
  background: rgba(255, 255, 255, 0.26);
}

.publish-panel > * {
  width: min(420px, 100%);
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.24);
  border-radius: 28px;
  padding: 32px;
  color: #fff;
  overflow-y: auto;
  box-shadow: 0 30px 60px rgba(15, 23, 42, 0.28);
}

.publish-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.publish-form__field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 14px;
}

.publish-form__field span {
  color: rgba(255, 255, 255, 0.75);
}

.publish-form__field input {
  height: 44px;
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  padding: 0 14px;
  font-size: 14px;
}

.publish-form__field input:focus {
  outline: none;
  border-color: rgba(189, 224, 254, 0.7);
  background: rgba(255, 255, 255, 0.12);
}

.publish-form__grid {
  display: flex;
  gap: 16px;
}

.publish-form__editor {
  min-height: 160px;
  border-radius: 18px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  font-size: 14px;
  line-height: 1.6;
  overflow-y: auto;
}

.publish-form__editor:focus {
  outline: none;
  border-color: rgba(189, 224, 254, 0.74);
  background: rgba(255, 255, 255, 0.12);
}

.publish-form__editor:empty:before {
  content: attr(placeholder);
  color: rgba(255, 255, 255, 0.4);
}

.publish-form__toolbar {
  display: flex;
  gap: 12px;
}

.publish-form__toolbar button,
.publish-form__upload {
  height: 40px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  font-size: 13px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.publish-form__toolbar button:hover,
.publish-form__upload:hover {
  background: rgba(255, 255, 255, 0.2);
}

.publish-form__toolbar input {
  display: none;
}

.publish-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
}

.publish-form__btn {
  height: 44px;
  padding: 0 22px;
  border-radius: 14px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  color: #1f2937;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.7));
}

.publish-form__btn--ghost {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.publish-form__btn[disabled] {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 1024px) {
  .hero-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .hero-card__cta {
    width: 100%;
    justify-content: flex-start;
  }

  .publish-panel {
    padding: 32px 16px;
  }
}

@media (max-width: 768px) {
  .workspace__topbar,
  .hero-card,
  .task-card,
  .workspace__filters {
    padding: 20px;
  }

  .workspace__topbar,
  .workspace__filters {
    flex-direction: column;
    align-items: stretch;
    gap: 18px;
  }

  .workspace__actions {
    justify-content: space-between;
  }

  .task-board {
    grid-template-columns: 1fr;
  }

  .publish-panel {
    padding: 12px;
  }

  .publish-panel > * {
    border-radius: 20px;
    padding: 24px 18px;
  }
}
</style>
