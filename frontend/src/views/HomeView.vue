<script setup>
import WorkspaceTopbar from '../components/layout/WorkspaceTopbar.vue'
import HeroBanner from '../components/dashboard/HeroBanner.vue'
import TaskFilterBar from '../components/dashboard/TaskFilterBar.vue'
import TaskBoard from '../components/dashboard/TaskBoard.vue'
import PublishPanel from '../components/dashboard/PublishPanel.vue'
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
  priorityMeta,
  togglePublishPanel,
  updateFormField,
  updateFormDescription,
  handleAccept,
  handleRelease,
  submitTask
} = useTaskBoard({ name: '林运维', role: 'admin' })

const handleFieldUpdate = ({ field, value }) => {
  updateFormField(field, value)
}

const handleDescriptionUpdate = (value) => {
  updateFormDescription(value)
}
</script>

<template>
  <div class="workspace">
    <WorkspaceTopbar :user="currentUser" :is-admin="isAdmin" @toggle-publish="togglePublishPanel" />

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
</style>
