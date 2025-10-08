<script setup>
const props = defineProps({
  keyword: {
    type: String,
    default: ''
  },
  sortKey: {
    type: String,
    default: 'priority'
  }
})

const emit = defineEmits(['update:keyword', 'update:sortKey'])

const handleKeyword = (event) => {
  emit('update:keyword', event.target.value)
}

const handleSortChange = (event) => {
  emit('update:sortKey', event.target.value)
}
</script>

<template>
  <section class="filters">
    <div class="filters__search">
      <span class="filters__icon">🔍</span>
      <input
        :value="keyword"
        type="search"
        placeholder="搜索任务编号、标题或描述"
        @input="handleKeyword"
      />
    </div>

    <div class="filters__select">
      <label>排序</label>
      <select :value="sortKey" @change="handleSortChange">
        <option value="priority">优先级</option>
        <option value="deadline">截止时间</option>
      </select>
    </div>
  </section>
</template>

<style scoped>
.filters {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.filters__search {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid var(--frost-border-soft);
  border-radius: 18px;
  padding: 0 16px;
  backdrop-filter: blur(14px);
  color: #fff;
}

.filters__search input {
  width: 100%;
  height: 46px;
  border: none;
  background: transparent;
  color: inherit;
  font-size: 15px;
  padding-left: 8px;
}

.filters__search input::placeholder {
  color: rgba(255, 255, 255, 0.6);
}

.filters__search input:focus {
  outline: none;
}

.filters__icon {
  font-size: 16px;
}

.filters__select {
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid var(--frost-border-soft);
  padding: 0 12px;
  border-radius: 16px;
  height: 46px;
  backdrop-filter: blur(12px);
  color: #fff;
}

.filters__select select {
  background: transparent;
  border: none;
  color: #fff;
  font-size: 14px;
}

.filters__select select:focus {
  outline: none;
}

@media (max-width: 768px) {
  .filters {
    flex-direction: column;
    align-items: stretch;
  }

  .filters__select {
    justify-content: space-between;
  }
}
</style>
