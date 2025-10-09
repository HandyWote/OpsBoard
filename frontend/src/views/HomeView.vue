<script setup>
import { ref, watchEffect } from 'vue'
import { useRouter } from 'vue-router'
import WorkspaceTopbar from '../components/layout/WorkspaceTopbar.vue'
import HeroBanner from '../components/dashboard/HeroBanner.vue'
import TaskFilterBar from '../components/dashboard/TaskFilterBar.vue'
import TaskBoard from '../components/dashboard/TaskBoard.vue'
import PublishPanel from '../components/dashboard/PublishPanel.vue'
import UserSidebar from '../components/dashboard/UserSidebar.vue'
import AccountManagerPanel from '../components/dashboard/AccountManagerPanel.vue'
import { useTaskBoard } from '../composables/useTaskBoard.js'
import { isAuthenticated } from '../services/http.js'

const {
  currentUser,
  isAdmin,
  sortKey,
  keyword,
  showPublishPanel,
  panelMode,
  submitting,
  publishForm,
  tasks: boardTasks,
  myPendingTasks,
  availableTasks,
  myCompletedTasks,
  earnedPoints,
  accounts,
  adminCount,
  priorityMeta,
  openPublishPanel,
  closePublishPanel,
  updateFormField,
  updateFormDescription,
  handleAccept,
  handleRelease,
  handleSubmitCompletion,
  handleVerifyCompletion,
  submitTask,
  toggleAdminForAccount,
  startEditTask,
  removeTask
} = useTaskBoard()

const showAccountManager = ref(false)
const router = useRouter()

const handleFieldUpdate = ({ field, value }) => {
  updateFormField(field, value)
}

const handleDescriptionUpdate = (value) => {
  updateFormDescription(value)
}

const handleOpenAdminPanel = () => {
  if (!isAdmin.value) return
  showAccountManager.value = true
}

const handleToggleAdmin = async (accountId) => {
  const nextRole = await toggleAdminForAccount(accountId)
  if (accountId === currentUser.id && nextRole !== 'admin') {
    showAccountManager.value = false
  }
}

const handleOpenProfile = () => {
  router.push({ name: 'profile' })
}

watchEffect(() => {
  if (!isAuthenticated()) {
    router.replace({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
  }
})

const handleCreateTask = () => {
  if (showPublishPanel.value && panelMode.value === 'create') {
    closePublishPanel()
  } else {
    openPublishPanel()
  }
}

const handleEditTask = (task) => {
  startEditTask(task)
}

const handleDeleteTask = (task) => {
  removeTask(task)
}
</script>

<template>
  <div class="workspace">
    <WorkspaceTopbar
      :user="currentUser"
      :is-admin="isAdmin"
      @toggle-publish="handleCreateTask"
      @open-admin="handleOpenAdminPanel"
      @open-profile="handleOpenProfile"
    />

    <div class="workspace-body">
      <div class="primary-pane">
        <HeroBanner :task-count="boardTasks.length" />

        <TaskFilterBar
          :keyword="keyword"
          :sort-key="sortKey"
          @update:keyword="keyword = $event"
          @update:sortKey="sortKey = $event"
        />

        <TaskBoard
          :tasks="boardTasks"
          :priority-meta="priorityMeta"
          :current-user-name="currentUser.name"
          :is-admin="isAdmin"
          @accept="handleAccept"
          @release="handleRelease"
          @edit="handleEditTask"
          @delete="handleDeleteTask"
        />

        <PublishPanel
          v-if="showPublishPanel"
          :form="publishForm"
          :mode="panelMode"
          :submitting="submitting"
          @close="closePublishPanel"
          @submit="submitTask"
          @update:field="handleFieldUpdate"
          @update:description="handleDescriptionUpdate"
        />
      </div>

      <UserSidebar
        :user="currentUser"
        :pending-tasks="myPendingTasks"
        :available-tasks="availableTasks"
        :completed-tasks="myCompletedTasks"
        :earned-points="earnedPoints"
        @submit-task="handleSubmitCompletion"
        @verify-task="handleVerifyCompletion"
      />
    </div>

    <AccountManagerPanel
      v-if="showAccountManager"
      :accounts="accounts"
      :current-user-id="currentUser.id"
      :admin-count="adminCount"
      @close="showAccountManager = false"
      @toggle-admin="handleToggleAdmin"
    />
  </div>
</template>

<style scoped>
.workspace {
  width: 100%;
  max-width: 1200px;
  display: flex;
  flex-direction: column;
  gap: 32px;
  z-index: 1;
}

.workspace-body {
  width: 100%;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 32px;
  align-items: start;
}

.primary-pane {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

@media (max-width: 1100px) {
  .workspace-body {
    grid-template-columns: 1fr;
  }

  .primary-pane {
    order: 1;
  }
}
</style>
