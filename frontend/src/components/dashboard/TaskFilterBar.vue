<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps({
  keyword: {
    type: String,
    default: ''
  },
  sortKey: {
    type: String,
    default: 'status_priority'
  }
})

const emit = defineEmits(['update:keyword', 'update:sortKey'])

const handleKeyword = (event) => {
  emit('update:keyword', event.target.value)
}

const sortOptions = [
  { value: 'status_priority', label: '状态 + 优先级' },
  { value: 'priority', label: '仅优先级' },
  { value: 'deadline', label: '截止时间' },
  { value: 'created_desc', label: '最新发布' },
  { value: 'bounty_desc', label: '赏金最高' }
]

const isMenuOpen = ref(false)
const dropdownRef = ref(null)

const currentSortLabel = computed(() => {
  const option = sortOptions.find((item) => item.value === props.sortKey)
  return option ? option.label : '状态 + 优先级'
})

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
}

const selectOption = (value) => {
  if (value !== props.sortKey) {
    emit('update:sortKey', value)
  }
  isMenuOpen.value = false
}

const handleKeydown = (event) => {
  if (!dropdownRef.value) return
  const isInsideDropdown = dropdownRef.value.contains(event.target)

  if (event.key === 'Escape') {
    isMenuOpen.value = false
  }

  if (!isInsideDropdown) return

  if ((event.key === 'Enter' || event.key === ' ') && !isMenuOpen.value) {
    event.preventDefault()
    isMenuOpen.value = true
  }

  if (event.key === 'ArrowDown' && !isMenuOpen.value) {
    event.preventDefault()
    isMenuOpen.value = true
  }
}

const handleClickOutside = (event) => {
  if (!dropdownRef.value) return
  if (dropdownRef.value.contains(event.target)) return
  isMenuOpen.value = false
}

onMounted(() => {
  window.addEventListener('click', handleClickOutside)
  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('click', handleClickOutside)
  window.removeEventListener('keydown', handleKeydown)
})

watch(
  () => props.sortKey,
  () => {
    isMenuOpen.value = false
  }
)
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

    <div ref="dropdownRef" class="filters__select">
      <label class="filters__select-label">排序</label>
      <button
        class="filters__select-trigger"
        type="button"
        :aria-expanded="isMenuOpen"
        aria-haspopup="listbox"
        @click.stop="toggleMenu"
      >
        <span>{{ currentSortLabel }}</span>
        <span class="filters__select-arrow" aria-hidden="true">⌄</span>
      </button>

      <transition name="filters-menu">
        <ul v-if="isMenuOpen" class="filters__menu" role="listbox" :aria-activedescendant="`sort-${sortKey}`">
          <li
            v-for="option in sortOptions"
            :id="`sort-${option.value}`"
            :key="option.value"
            class="filters__menu-item"
            :class="{ 'filters__menu-item--active': option.value === sortKey }"
            role="option"
            :aria-selected="option.value === sortKey"
            @click.stop="selectOption(option.value)"
          >
            <span class="filters__menu-label">{{ option.label }}</span>
            <span v-if="option.value === sortKey" class="filters__menu-check" aria-hidden="true">●</span>
          </li>
        </ul>
      </transition>
    </div>
  </section>
</template>

<style scoped>
.filters {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  position: relative;
  z-index: 3;
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
  padding: 0 18px 0 18px;
  border-radius: 22px;
  height: 48px;
  backdrop-filter: blur(14px);
  color: #fff;
  position: relative;
  transition: border 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
  z-index: 5;
}

.filters__select-label {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
}

.filters__select-trigger {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  height: 38px;
  padding: 0 38px 0 16px;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(15, 23, 42, 0.24);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.05);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: border 0.2s ease, background 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.filters__select-trigger:hover {
  background: rgba(15, 23, 42, 0.32);
  box-shadow: inset 0 0 0 1px rgba(189, 224, 254, 0.18), 0 10px 22px rgba(15, 23, 42, 0.28);
}

.filters__select-trigger:focus {
  outline: none;
}

.filters__select-trigger[aria-expanded='true'] {
  border-color: rgba(189, 224, 254, 0.85);
  background: rgba(15, 23, 42, 0.36);
  box-shadow: inset 0 0 0 1px rgba(189, 224, 254, 0.36), 0 14px 28px rgba(15, 23, 42, 0.32);
}

.filters__select-arrow {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.04em;
  transition: transform 0.2s ease;
}

.filters__select-trigger[aria-expanded='true'] .filters__select-arrow {
  transform: translateY(-50%) rotate(180deg);
}

.filters__menu {
  position: absolute;
  right: 14px;
  top: calc(100% + 14px);
  width: 180px;
  padding: 10px;
  margin: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
  background: linear-gradient(160deg, rgba(255, 255, 255, 0.16), rgba(148, 187, 233, 0.18));
  border: 1px solid rgba(255, 255, 255, 0.26);
  border-radius: 18px;
  box-shadow: 0 18px 38px rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(18px);
  z-index: 12;
}

.filters__menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border-radius: 14px;
  cursor: pointer;
  font-size: 14px;
  color: rgba(15, 23, 42, 0.78);
  transition: background 0.2s ease, color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
}

.filters__menu-item:hover {
  background: rgba(189, 224, 254, 0.24);
  color: #0f172a;
  box-shadow: inset 0 0 0 1px rgba(189, 224, 254, 0.3);
}

.filters__menu-item--active {
  background: rgba(255, 255, 255, 0.6);
  color: #0f172a;
  box-shadow: inset 0 0 0 1px rgba(189, 224, 254, 0.42);
}

.filters__menu-label {
  flex: 1;
}

.filters__menu-check {
  font-size: 12px;
  color: rgba(15, 23, 42, 0.7);
}

.filters-menu-enter-active,
.filters-menu-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.filters-menu-enter-from,
.filters-menu-leave-to {
  opacity: 0;
  transform: translateY(-8px);
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

  .filters__menu {
    right: 20px;
  }
}
</style>
