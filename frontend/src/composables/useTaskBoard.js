import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useCurrentUser } from './useCurrentUser.js'
import { fetchCurrentUser, listAccounts, toggleAdmin } from '../services/users.js'
import { claimTask, createTask, fetchTasks, releaseTask } from '../services/tasks.js'

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

  const isAdmin = computed(() => currentUser.role === 'admin')

  const adminCount = computed(() => accounts.value.filter((account) => account.role === 'admin').length)

  const myPendingTasks = computed(() =>
    tasks.value
      .filter((task) => task.status === 'claimed' && task.assignee === currentUser.name)
      .slice()
      .sort((a, b) => new Date(a.deadline || 0) - new Date(b.deadline || 0))
  )

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

  const togglePublishPanel = () => {
    if (!isAdmin.value && !showPublishPanel.value) return
    showPublishPanel.value = !showPublishPanel.value
  }

  const resetPublishForm = () => {
    publishForm.title = ''
    publishForm.reward = ''
    publishForm.deadline = ''
    publishForm.tags = ''
    publishForm.description = ''
  }

  const mapTaskFromApi = (item) => {
    const status = item.status === 'published' ? 'available' : item.status
    const assigneeName = item.currentAssignee?.username || ''
    return {
      id: item.id,
      title: item.title,
      summary: item.descriptionPlain || '',
      reward: item.bounty ?? 0,
      deadline: item.deadline,
      priority: item.priority || 'medium',
      tags: Array.isArray(item.tags) ? item.tags.slice() : [],
      status,
      assignee: assigneeName,
      assigneeId: item.currentAssignee?.userId || '',
      createdAt: item.createdAt,
      updatedAt: item.updatedAt
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
    await loadAccounts()
  }

  const submitTask = async () => {
    if (!publishForm.title.trim() || !publishForm.description.trim()) {
      return
    }

    submitting.value = true
    try {
      const payload = {
        title: publishForm.title.trim(),
        descriptionHtml: publishForm.description,
        bounty: Number(publishForm.reward) || 0,
        priority: 'medium',
        deadline: publishForm.deadline ? new Date(publishForm.deadline).toISOString() : null,
        tags: publishForm.tags
          .split(',')
          .map((tag) => tag.trim())
          .filter(Boolean),
        publish: true
      }

      await createTask(payload)
      resetPublishForm()
      showPublishPanel.value = false
      await loadTasks()
    } catch (error) {
      console.error('发布任务失败', error)
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
    submitting,
    publishForm,
    myPendingTasks,
    availableTasks,
    accounts,
    adminCount,
    updateFormField: updatePublishForm,
    updateFormDescription: (value) => updatePublishForm('description', value),
    togglePublishPanel,
    submitTask,
    handleAccept,
    handleRelease,
    toggleAdminForAccount,
    initialize
  }
}
