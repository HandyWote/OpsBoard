import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useCurrentUser } from './useCurrentUser.js'
import { fetchCurrentUser, listAccounts, toggleAdmin } from '../services/users.js'
import {
  claimTask,
  createTask,
  deleteTask,
  fetchTasks,
  releaseTask,
  submitTaskProgress,
  updateTask,
  verifyTaskCompletion,
  rejectTaskSubmission
} from '../services/tasks.js'
import { mapTaskFromApi } from '../utils/mapTask.js'

const priorityMeta = {
  critical: { label: '特急', tone: 'var(--danger)' },
  high: { label: '高', tone: 'var(--warning)' },
  medium: { label: '中', tone: 'var(--info)' },
  low: { label: '低', tone: 'var(--muted)' }
}

export function useTaskBoard() {
  const { mutableProfile: currentUser, hydrate, setRole } = useCurrentUser()

  const tasks = ref([])
  const totalTasks = ref(0)
  const loadingTasks = ref(false)
  const loadingUser = ref(false)
  const accounts = ref([])
  const loadingAccounts = ref(false)
  const completedTasks = ref([])

  const sortKey = ref('priority')
  const keyword = ref('')

  const showPublishPanel = ref(false)
  const submitting = ref(false)

  const publishForm = reactive({
    title: '',
    reward: '',
    deadline: '',
    tags: '',
    description: ''
  })
  const panelMode = ref('create')
  const editingTaskId = ref('')

  const isAdmin = computed(() => currentUser.role === 'admin')

  const adminCount = computed(() => accounts.value.filter((account) => account.role === 'admin').length)

  const toInputDeadline = (value) => {
    if (!value) return ''
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ''
    const offsetDate = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
    return offsetDate.toISOString().slice(0, 16)
  }

  const toDeadlinePayload = (value) => {
    if (!value) return null
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return null
    return date.toISOString()
  }

  const completionTime = (task) => {
    if (task.completedAt) {
      const completedTs = new Date(task.completedAt).getTime()
      if (!Number.isNaN(completedTs)) return completedTs
    }
    if (task.updatedAt) {
      const updatedTs = new Date(task.updatedAt).getTime()
      if (!Number.isNaN(updatedTs)) return updatedTs
    }
    return 0
  }

  const myCompletedTasks = computed(() =>
    completedTasks.value.slice().sort((a, b) => completionTime(b) - completionTime(a))
  )

  const earnedPoints = computed(() =>
    completedTasks.value.reduce((sum, task) => sum + (Number(task.reward) || 0), 0)
  )

  const myPendingTasks = computed(() => {
    const currentId = currentUser.id
    if (!currentId) return []

    const sortTime = (task) => {
      if (task.deadline) {
        const deadline = new Date(task.deadline).getTime()
        if (!Number.isNaN(deadline)) return deadline
      }
      if (task.updatedAt) {
        const updated = new Date(task.updatedAt).getTime()
        if (!Number.isNaN(updated)) return updated
      }
      return Number.MAX_SAFE_INTEGER
    }

    const executing = tasks.value
      .filter((task) => task.status === 'claimed' && task.assigneeId === currentId)
      .map((task) => ({ ...task, pendingKind: 'execute' }))

    const reviewing = tasks.value
      .filter((task) => {
        if (task.status !== 'submitted') return false
        if (task.ownerId === currentId) return true
        return isAdmin.value
      })
      .map((task) => ({ ...task, pendingKind: 'review' }))

    return [...executing, ...reviewing].sort((a, b) => sortTime(a) - sortTime(b))
  })

  const availableTasks = computed(() =>
    tasks.value
      .filter((task) => task.status === 'available')
      .slice()
      .sort((a, b) => new Date(a.deadline || 0) - new Date(b.deadline || 0))
  )

  const filteredTasks = computed(() => {
    const term = keyword.value.trim().toLowerCase()
    if (!term) return tasks.value
    return tasks.value.filter((task) => `${task.id} ${task.title} ${task.summary}`.toLowerCase().includes(term))
  })

  const updatePublishForm = (field, value) => {
    if (Object.prototype.hasOwnProperty.call(publishForm, field)) {
      publishForm[field] = value
    }
  }

  const resetPublishForm = () => {
    publishForm.title = ''
    publishForm.reward = ''
    publishForm.deadline = ''
    publishForm.tags = ''
    publishForm.description = ''
  }

  const closePublishPanel = () => {
    showPublishPanel.value = false
    panelMode.value = 'create'
    editingTaskId.value = ''
    resetPublishForm()
  }

  const openPublishPanel = () => {
    if (!isAdmin.value) return
    panelMode.value = 'create'
    editingTaskId.value = ''
    resetPublishForm()
    showPublishPanel.value = true
  }

  const togglePublishPanel = () => {
    if (!isAdmin.value) return
    if (showPublishPanel.value) {
      closePublishPanel()
    } else {
      openPublishPanel()
    }
  }

  const fillFormFromTask = (task) => {
    publishForm.title = task.title || ''
    publishForm.reward = task.reward ? String(task.reward) : ''
    publishForm.deadline = toInputDeadline(task.deadline)
    publishForm.tags = Array.isArray(task.tags) ? task.tags.join(', ') : ''
    publishForm.description = task.descriptionHtml || ''
  }

  const startEditTask = (task) => {
    if (!isAdmin.value || !task) return
    if (task.status === 'completed') {
      console.warn('已完成的任务不可编辑')
      return
    }
    panelMode.value = 'edit'
    editingTaskId.value = task.id
    fillFormFromTask(task)
    showPublishPanel.value = true
  }

  const removeTask = async (task) => {
    if (!isAdmin.value || !task) return
    const confirmed = window.confirm(`确定删除任务「${task.title}」吗？该操作不可撤销。`)
    if (!confirmed) return

    try {
      await deleteTask(task.id)
      if (editingTaskId.value === task.id) {
        closePublishPanel()
      }
      await Promise.all([loadTasks(), loadCompletedTasks()])
    } catch (error) {
      console.error('删除任务失败', error)
    }
  }

  const loadTasks = async () => {
    loadingTasks.value = true
    try {
      const data = await fetchTasks({
        keyword: keyword.value.trim(),
        sort: sortKey.value
      })
      const items = Array.isArray(data?.items) ? data.items : []
      tasks.value = items.map(mapTaskFromApi)
      totalTasks.value = data?.total ?? items.length
    } catch (error) {
      console.error('加载任务失败', error)
    } finally {
      loadingTasks.value = false
    }
  }

  const loadCompletedTasks = async () => {
    if (!currentUser.id) {
      completedTasks.value = []
      return
    }
    try {
      const data = await fetchTasks({
        status: 'completed',
        assignee: 'me',
        pageSize: 100
      })
      const items = Array.isArray(data?.items) ? data.items : []
      completedTasks.value = items.map(mapTaskFromApi)
    } catch (error) {
      console.error('加载完成任务失败', error)
    }
  }

  const loadCurrentUser = async () => {
    loadingUser.value = true
    try {
      const data = await fetchCurrentUser()
      hydrate(data || {})
    } catch (error) {
      console.error('获取用户信息失败', error)
    } finally {
      loadingUser.value = false
    }
  }

  const loadAccounts = async () => {
    if (!isAdmin.value) {
      accounts.value = []
      return
    }

    loadingAccounts.value = true
    try {
      const data = await listAccounts({ page: 1, pageSize: 50 })
      const items = Array.isArray(data?.items) ? data.items : []
      accounts.value = items.map((item) => ({
        id: item.id,
        name: item.displayName || item.name || item.username || '',
        role: item.roles?.includes('admin') ? 'admin' : 'member',
        email: item.email || '',
        teams: item.teams || []
      }))
    } catch (error) {
      console.error('加载账号列表失败', error)
    } finally {
      loadingAccounts.value = false
    }
  }

  const initialize = async () => {
    await Promise.all([loadCurrentUser(), loadTasks()])
    await Promise.all([loadAccounts(), loadCompletedTasks()])
  }

  const submitTask = async () => {
    if (!publishForm.title.trim() || !publishForm.description.trim()) {
      return
    }

    const currentMode = panelMode.value
    submitting.value = true
    try {
      const basePayload = {
        title: publishForm.title.trim(),
        descriptionHtml: publishForm.description,
        bounty: Number(publishForm.reward) || 0,
        priority: 'medium',
        deadline: toDeadlinePayload(publishForm.deadline),
        tags: publishForm.tags
          .split(',')
          .map((tag) => tag.trim())
          .filter(Boolean)
      }

      if (currentMode === 'edit' && editingTaskId.value) {
        await updateTask(editingTaskId.value, {
          title: basePayload.title,
          descriptionHtml: basePayload.descriptionHtml,
          bounty: basePayload.bounty,
          priority: basePayload.priority,
          deadline: basePayload.deadline,
          tags: basePayload.tags
        })
      } else {
        await createTask({ ...basePayload, publish: true })
      }

      panelMode.value = 'create'
      editingTaskId.value = ''
      resetPublishForm()

      await loadTasks()
      closePublishPanel()
    } catch (error) {
      const message = currentMode === 'edit' ? '更新任务失败' : '发布任务失败'
      console.error(message, error)
    } finally {
      submitting.value = false
    }
  }

  const handleAccept = async (task) => {
    try {
      await claimTask(task.id)
      await loadTasks()
    } catch (error) {
      console.error('认领任务失败', error)
    }
  }

  const handleRelease = async (task) => {
    try {
      await releaseTask(task.id)
      await loadTasks()
    } catch (error) {
      console.error('释放任务失败', error)
    }
  }

  const handleSubmitCompletion = async (task) => {
    try {
      await submitTaskProgress(task.id)
      await loadTasks()
    } catch (error) {
      console.error('提交任务失败', error)
    }
  }

  const handleVerifyCompletion = async (task) => {
    try {
      await verifyTaskCompletion(task.id)
      await loadTasks()
      await loadCompletedTasks()
    } catch (error) {
      console.error('验收任务失败', error)
    }
  }

  const handleRejectCompletion = async (task) => {
    try {
      await rejectTaskSubmission(task.id)
      await loadTasks()
    } catch (error) {
      console.error('拒绝任务失败', error)
    }
  }

  const toggleAdminForAccount = async (accountId) => {
    const target = accounts.value.find((item) => item.id === accountId)
    if (!target) return currentUser.role

    const grant = target.role !== 'admin'
    try {
      await toggleAdmin(accountId, grant)
      target.role = grant ? 'admin' : 'member'
      if (currentUser.id === accountId) {
        setRole(target.role)
      }
      await loadAccounts()
      return target.role
    } catch (error) {
      console.error('切换管理员权限失败', error)
      return target.role
    }
  }

  let keywordDebounce = null
  watch(
    () => keyword.value,
    () => {
      if (keywordDebounce) {
        clearTimeout(keywordDebounce)
      }
      keywordDebounce = setTimeout(() => {
        loadTasks()
      }, 350)
    }
  )

  watch(
    () => sortKey.value,
    () => {
      loadTasks()
    }
  )

  watch(
    () => isAdmin.value,
    (value) => {
      if (value) {
        loadAccounts()
      } else {
        accounts.value = []
      }
    }
  )

  onMounted(() => {
    initialize()
  })

  return {
    currentUser,
    isAdmin,
    tasks: filteredTasks,
    priorityMeta,
    sortKey,
    keyword,
    totalTasks,
    loadingTasks,
    loadingUser,
    loadingAccounts,
    showPublishPanel,
    panelMode,
    submitting,
    publishForm,
    myPendingTasks,
    availableTasks,
    myCompletedTasks,
    earnedPoints,
    accounts,
    adminCount,
    updateFormField: updatePublishForm,
    updateFormDescription: (value) => updatePublishForm('description', value),
    openPublishPanel,
    closePublishPanel,
    startEditTask,
    removeTask,
    submitTask,
    handleAccept,
    handleRelease,
    handleSubmitCompletion,
    handleVerifyCompletion,
    handleRejectCompletion,
    toggleAdminForAccount,
    initialize
  }
}
