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
      <label class="filters__select-label">排序</label>
      <select :value="sortKey" class="filters__select-control" @change="handleSortChange">
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
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 22px;
  padding: 0 18px;
  backdrop-filter: blur(16px);
  color: #fff;
  transition: border 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.filters__search input {
  width: 100%;
  height: 48px;
  border: none;
  background: transparent;
  color: inherit;
  font-size: 15px;
  padding-left: 10px;
}

.filters__search input::placeholder {
  color: rgba(255, 255, 255, 0.6);
}

.filters__search input:focus {
  outline: none;
}

.filters__search:focus-within {
  border-color: rgba(189, 224, 254, 0.85);
  background: rgba(255, 255, 255, 0.16);
  box-shadow: 0 16px 34px rgba(15, 23, 42, 0.24);
}

.filters__icon {
  font-size: 18px;
}

.filters__select {
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.16);
  padding: 0 18px;
  border-radius: 22px;
  height: 48px;
  backdrop-filter: blur(14px);
  color: #fff;
  position: relative;
  transition: border 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.filters__select-label {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
}

.filters__select::after {
  content: '▾';
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
}

.filters__select-control {
  background: transparent;
  border: none;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  padding-right: 28px;
  cursor: pointer;
  appearance: none;
}

.filters__select-control option {
  color: #1f2937;
}

.filters__select:focus-within {
  border-color: rgba(189, 224, 254, 0.85);
  background: rgba(255, 255, 255, 0.16);
  box-shadow: 0 16px 34px rgba(15, 23, 42, 0.24);
}

.filters__select-control:focus {
  outline: none;
}

@media (max-width: 768px) {
  .filters {
    flex-direction: column;
    align-items: stretch;
  }

  .filters__select {
    justify-content: space-between;
    width: 100%;
  }
}
</style>
