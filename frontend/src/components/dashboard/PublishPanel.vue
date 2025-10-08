<script setup>
import { onMounted, ref, watch } from 'vue'

const props = defineProps({
  form: {
    type: Object,
    required: true
  },
  submitting: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'submit', 'update:field', 'update:description'])

const editorRef = ref(null)

onMounted(() => {
  if (editorRef.value) {
    editorRef.value.innerHTML = props.form.description || ''
  }
})

watch(
  () => props.form.description,
  (value) => {
    if (editorRef.value && editorRef.value.innerHTML !== value) {
      editorRef.value.innerHTML = value || ''
    }
  }
)

const handleField = (field, event) => {
  emit('update:field', { field, value: event.target.value })
}

const handleDescriptionInput = (event) => {
  emit('update:description', event.currentTarget.innerHTML)
}

const exec = (command) => {
  document.execCommand(command, false)
}
</script>

<template>
  <transition name="publish">
    <aside class="panel">
      <div class="panel__content">
        <header class="panel__header">
          <h2>发布新任务</h2>
          <button type="button" class="panel__close" @click="emit('close')">✕</button>
        </header>

        <form class="form" @submit.prevent="emit('submit')">
          <label class="form__field">
            <span>任务标题</span>
            <input :value="form.title" type="text" placeholder="例如：机房巡检脚本修复" @input="handleField('title', $event)" />
          </label>

          <div class="form__grid">
            <label class="form__field">
              <span>赏金 (¥)</span>
              <input :value="form.reward" type="number" min="0" step="10" placeholder="200" @input="handleField('reward', $event)" />
            </label>

            <label class="form__field">
              <span>截止时间</span>
              <input :value="form.deadline" type="datetime-local" @input="handleField('deadline', $event)" />
            </label>
          </div>

          <label class="form__field">
            <span>标签 (逗号分隔)</span>
            <input :value="form.tags" type="text" placeholder="网络, 自动化" @input="handleField('tags', $event)" />
          </label>

          <label class="form__field">
            <span>任务说明</span>
            <div
              ref="editorRef"
              class="form__editor"
              contenteditable
              placeholder="请输入任务背景、目标、交付标准，可粘贴截图或拖拽图片。"
              @input="handleDescriptionInput"
            ></div>
          </label>

          <div class="form__toolbar">
            <button type="button" @click="exec('bold')"><span>加粗</span></button>
            <button type="button" @click="exec('italic')"><span>斜体</span></button>
            <button type="button" @click="exec('insertUnorderedList')"><span>列表</span></button>
            <label class="form__upload">
              <span>上传附件</span>
              <input type="file" hidden multiple />
            </label>
          </div>

          <div class="form__actions">
            <button type="button" class="form__btn form__btn--ghost" @click="emit('close')">取消</button>
            <button type="submit" class="form__btn" :disabled="submitting">
              <span v-if="!submitting">发布任务</span>
              <span v-else>发布中...</span>
            </button>
          </div>
        </form>
      </div>
    </aside>
  </transition>
</template>

<style scoped>
.publish-enter-active,
.publish-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.publish-enter-from,
.publish-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

.panel {
  position: fixed;
  inset: 0;
  display: flex;
  justify-content: flex-end;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(12px);
  padding: 48px;
  z-index: 5;
}

.panel__content {
  width: min(420px, 100%);
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid var(--frost-border-strong);
  border-radius: 28px;
  padding: 32px;
  color: #fff;
  overflow-y: auto;
  box-shadow: 0 30px 60px var(--frost-shadow-strong);
}

.panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
}

.panel__header h2 {
  margin: 0;
  font-size: 22px;
}

.panel__close {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  cursor: pointer;
}

.panel__close:hover {
  background: rgba(255, 255, 255, 0.26);
}

.form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form__field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 14px;
}

.form__field span {
  color: var(--frost-text-secondary);
}

.form__field input {
  height: 44px;
  border-radius: 14px;
  border: 1px solid var(--frost-border-soft);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  padding: 0 14px;
  font-size: 14px;
}

.form__field input:focus {
  outline: none;
  border-color: rgba(189, 224, 254, 0.7);
  background: rgba(255, 255, 255, 0.12);
}

.form__grid {
  display: flex;
  gap: 16px;
}

.form__editor {
  min-height: 160px;
  border-radius: 18px;
  padding: 18px;
  border: 1px solid var(--frost-border-soft);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  font-size: 14px;
  line-height: 1.6;
  overflow-y: auto;
}

.form__editor:focus {
  outline: none;
  border-color: rgba(189, 224, 254, 0.74);
  background: rgba(255, 255, 255, 0.12);
}

.form__editor:empty:before {
  content: attr(placeholder);
  color: rgba(255, 255, 255, 0.4);
}

.form__toolbar {
  display: flex;
  gap: 12px;
}

.form__toolbar button,
.form__upload {
  height: 40px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid var(--frost-border-soft);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  font-size: 13px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.form__toolbar button:hover,
.form__upload:hover {
  background: rgba(255, 255, 255, 0.2);
}

.form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
}

.form__btn {
  height: 44px;
  padding: 0 22px;
  border-radius: 14px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  color: #1f2937;
  background: var(--frost-highlight);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.form__btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form__btn--ghost {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.form__btn--ghost:hover {
  background: rgba(255, 255, 255, 0.2);
}

@media (max-width: 1024px) {
  .panel {
    padding: 32px 16px;
  }
}

@media (max-width: 768px) {
  .panel {
    padding: 12px;
  }

  .panel__content {
    border-radius: 20px;
    padding: 24px 18px;
  }

  .form__grid {
    flex-direction: column;
  }
}
</style>
