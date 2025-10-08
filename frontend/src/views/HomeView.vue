<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import WorkspaceTopbar from '../components/layout/WorkspaceTopbar.vue'
import HeroBanner from '../components/dashboard/HeroBanner.vue'
import TaskFilterBar from '../components/dashboard/TaskFilterBar.vue'
import TaskBoard from '../components/dashboard/TaskBoard.vue'
import PublishPanel from '../components/dashboard/PublishPanel.vue'
import UserSidebar from '../components/dashboard/UserSidebar.vue'
import AccountManagerPanel from '../components/dashboard/AccountManagerPanel.vue'
import { useTaskBoard } from '../composables/useTaskBoard.js'

const {
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
  accounts,
  adminCount,
  priorityMeta,
  togglePublishPanel,
  updateFormField,
  updateFormDescription,
  handleAccept,
  handleRelease,
  submitTask,
  toggleAdminForAccount
} = useTaskBoard({ name: '林运维', role: 'admin' })

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

const handleToggleAdmin = (accountId) => {
  const nextRole = toggleAdminForAccount(accountId)
  if (accountId === currentUser.id && nextRole !== 'admin') {
    showAccountManager.value = false
  }
}

const handleOpenProfile = () => {
  router.push({ name: 'profile' })
}
</script>

<template>
  <div class="workspace">
    <WorkspaceTopbar
      :user="currentUser"
      :is-admin="isAdmin"
      @toggle-publish="togglePublishPanel"
      @open-admin="handleOpenAdminPanel"
      @open-profile="handleOpenProfile"
    />

    <div class="workspace-body">
      <div class="primary-pane">
        <HeroBanner :task-count="filteredTasks.length" />

        <TaskFilterBar
          :keyword="keyword"
          :sort-key="sortKey"
          @update:keyword="keyword = $event"
          @update:sortKey="sortKey = $event"
        />

        <TaskBoard
          :tasks="filteredTasks"
          :priority-meta="priorityMeta"
          :current-user-name="currentUser.name"
          @accept="handleAccept"
          @release="handleRelease"
        />

        <PublishPanel
          v-if="showPublishPanel"
          :form="publishForm"
          :submitting="submitting"
          @close="togglePublishPanel"
          @submit="submitTask"
          @update:field="handleFieldUpdate"
          @update:description="handleDescriptionUpdate"
        />
      </div>

      <UserSidebar
        :user="currentUser"
        :pending-tasks="myPendingTasks"
        :available-tasks="availableTasks"
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
