import { computed, reactive, ref } from 'vue'

const priorityOrder = {
  critical: 1,
  high: 2,
  medium: 3,
  low: 4
}

const priorityMeta = {
  critical: { label: '特急', tone: 'var(--danger)' },
  high: { label: '高', tone: 'var(--warning)' },
  medium: { label: '中', tone: 'var(--info)' },
  low: { label: '低', tone: 'var(--muted)' }
}

const initialTasks = [
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
]

export function useTaskBoard(user = { name: '', role: 'member' }) {
  const currentUser = reactive({ ...user })
  const sortKey = ref('priority')
  const keyword = ref('')
  const showPublishPanel = ref(false)
  const submitting = ref(false)
  const tasks = reactive(initialTasks.slice())

  const publishForm = reactive({
    title: '',
    reward: '',
    deadline: '',
    tags: '',
    description: ''
  })

  const isAdmin = computed(() => currentUser.role === 'admin')

  const myPendingTasks = computed(() =>
    tasks
      .filter((task) => task.status === 'claimed' && task.assignee === currentUser.name)
      .slice()
      .sort((a, b) => new Date(a.deadline).getTime() - new Date(b.deadline).getTime())
  )

  const availableTasks = computed(() =>
    tasks
      .filter((task) => task.status === 'available')
      .slice()
      .sort((a, b) => new Date(a.deadline).getTime() - new Date(b.deadline).getTime())
  )

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

  const togglePublishPanel = () => {
    if (!isAdmin.value && !showPublishPanel.value) {
      return
    }
    showPublishPanel.value = !showPublishPanel.value
  }

  const resetForm = () => {
    publishForm.title = ''
    publishForm.reward = ''
    publishForm.deadline = ''
    publishForm.tags = ''
    publishForm.description = ''
  }

  const updateFormField = (field, value) => {
    if (Object.prototype.hasOwnProperty.call(publishForm, field)) {
      publishForm[field] = value
    }
  }

  const updateFormDescription = (value) => {
    publishForm.description = value
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

  const submitTask = () => {
    const plainText = publishForm.description.replace(/<[^>]+>/g, '').trim()
    if (!publishForm.title.trim() || !plainText) {
      return
    }

    submitting.value = true
    setTimeout(() => {
      tasks.unshift({
        id: `T-${Math.floor(Math.random() * 9000 + 1000)}`,
        title: publishForm.title.trim(),
        summary: plainText.slice(0, 140) || '新任务',
        reward: Number(publishForm.reward) || 0,
        deadline: publishForm.deadline || new Date().toISOString(),
        priority: 'medium',
        tags: publishForm.tags
          .split(',')
          .map((tag) => tag.trim())
          .filter(Boolean),
        status: 'available',
        assignee: null
      })

      resetForm()
      submitting.value = false
      showPublishPanel.value = false
    }, 450)
  }

  return {
    currentUser,
    isAdmin,
    sortKey,
    keyword,
    showPublishPanel,
    submitting,
    publishForm,
    filteredTasks,
    myPendingTasks,
    availableTasks,
    priorityMeta,
    togglePublishPanel,
    updateFormField,
    updateFormDescription,
    handleAccept,
    handleRelease,
    submitTask
  }
}
