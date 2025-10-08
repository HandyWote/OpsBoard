<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

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
const activeMarks = ref(new Set())

onMounted(() => {
  if (editorRef.value) {
    editorRef.value.innerHTML = props.form.description || ''
  }

  document.addEventListener('selectionchange', syncActiveMarks)
  syncActiveMarks()
})

onBeforeUnmount(() => {
  document.removeEventListener('selectionchange', syncActiveMarks)
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
  syncActiveMarks()
}

const toolbarButtons = [
  { label: 'B', hint: '加粗 (Ctrl+B)', command: 'bold' },
  { label: 'I', hint: '斜体 (Ctrl+I)', command: 'italic' },
  { label: 'U', hint: '下划线 (Ctrl+U)', command: 'underline' },
  { label: '•', hint: '项目符号', command: 'insertUnorderedList' },
  { label: '1.', hint: '编号列表', command: 'insertOrderedList' },
  { label: '“”', hint: '引用', command: 'formatBlock', value: 'blockquote' },
  { label: '</>', hint: '等宽文本', command: 'formatBlock', value: 'pre' }
]

const syncActiveMarks = () => {
  const el = editorRef.value
  const selection = document.getSelection()
  if (!el || !selection || !selection.anchorNode || !el.contains(selection.anchorNode)) {
    activeMarks.value = new Set()
    return
  }

  const next = new Set()
  toolbarButtons.forEach((btn) => {
    try {
      const state = document.queryCommandState(btn.command)
      if (state) {
        next.add(btn.command + (btn.value ? `:${btn.value}` : ''))
      }
    } catch (err) {
      /* 静默失败，部分浏览器未实现 */
    }
  })
  activeMarks.value = next
}

const exec = (command, value = null) => {
  document.execCommand(command, false, value)
  syncActiveMarks()
}

const isActive = (btn) => activeMarks.value.has(btn.command + (btn.value ? `:${btn.value}` : ''))

const handleBlur = () => {
  activeMarks.value = new Set()
}

const handleKeydown = (event) => {
  if (!event.ctrlKey && !event.metaKey) return

  switch (event.key.toLowerCase()) {
    case 'b':
      event.preventDefault()
      exec('bold')
      break
    case 'i':
      event.preventDefault()
      exec('italic')
      break
    case 'u':
      event.preventDefault()
      exec('underline')
      break
    default:
      break
  }
}

const handlePaste = (event) => {
  event.preventDefault()
  const text = event.clipboardData?.getData('text/plain') ?? ''
  document.execCommand('insertText', false, text)
}
</script>

<template>
  <transition name="publish-overlay" appear>
    <div class="panel" role="dialog" aria-modal="true" aria-label="发布新任务">
      <transition name="publish-card" appear>
        <section class="panel__content">
          <div class="panel__glow" aria-hidden="true"></div>
          <header class="panel__header">
            <div>
              <p class="panel__eyebrow">新的运维任务</p>
              <h2>发布任务</h2>
              <p class="panel__subtitle">完善基础信息并通过富文本描述执行标准。</p>
            </div>
            <button type="button" class="panel__close" @click="emit('close')" aria-label="关闭发布面板">✕</button>
          </header>

          <form class="form" @submit.prevent="emit('submit')">
            <label class="form__field">
              <span>任务标题</span>
              <input
                :value="form.title"
                type="text"
                placeholder="例如：机房巡检脚本修复"
                autocomplete="off"
                @input="handleField('title', $event)"
              />
            </label>

            <div class="form__grid">
              <label class="form__field">
                <span>赏金 (¥)</span>
                <input
                  :value="form.reward"
                  type="number"
                  min="0"
                  step="10"
                  placeholder="200"
                  @input="handleField('reward', $event)"
                />
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

            <div class="editor">
              <label class="form__field">
                <span>任务说明</span>
                <div
                  ref="editorRef"
                  class="editor__surface"
                  contenteditable
                  role="textbox"
                  aria-multiline="true"
                  placeholder="请输入任务背景、目标、交付标准，可直接粘贴或拖拽截图。"
                  @keydown="handleKeydown"
                  @paste="handlePaste"
                  @focus="syncActiveMarks"
                  @blur="handleBlur"
                  @input="handleDescriptionInput"
                  @dragover.prevent
                  @drop.prevent
                ></div>
              </label>

              <div class="editor__toolbar" role="toolbar" aria-label="富文本工具栏">
                <button
                  v-for="btn in toolbarButtons"
                  :key="btn.command + (btn.value || '')"
                  type="button"
                  class="editor__tool"
                  :class="{ 'editor__tool--active': isActive(btn) }"
                  :title="btn.hint"
                  @mousedown.prevent
                  @click="exec(btn.command, btn.value)"
                >
                  {{ btn.label }}
                </button>

                <label class="editor__upload" title="上传附件">
                  <span>附件</span>
                  <input type="file" hidden multiple />
                </label>
              </div>
            </div>

            <div class="form__actions">
              <button type="button" class="form__btn form__btn--ghost" @click="emit('close')">取消</button>
              <button type="submit" class="form__btn" :disabled="submitting">
                <span v-if="!submitting">发布任务</span>
                <span v-else>发布中...</span>
              </button>
            </div>
          </form>
        </section>
      </transition>
    </div>
  </transition>
</template>

<style scoped>
.publish-overlay-enter-active,
.publish-overlay-leave-active {
  transition: opacity 0.36s cubic-bezier(0.22, 0.61, 0.36, 1);
}

.publish-overlay-enter-from,
.publish-overlay-leave-to {
  opacity: 0;
}

.publish-card-enter-active,
.publish-card-leave-active {
  transition: transform 0.38s cubic-bezier(0.22, 0.61, 0.36, 1),
    opacity 0.34s cubic-bezier(0.22, 0.61, 0.36, 1);
}

.publish-card-enter-from,
.publish-card-leave-to {
  opacity: 0;
  transform: translate3d(28px, -12px, 0) scale(0.9);
}

.panel {
  position: fixed;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(12px);
  padding: 48px 24px;
  z-index: 5;
}

.panel__content {
  position: relative;
  width: min(640px, 100%);
  background: linear-gradient(160deg, rgba(255, 255, 255, 0.16), rgba(148, 187, 233, 0.18));
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 32px;
  padding: 40px 44px;
  color: #fff;
  overflow-y: auto;
  box-shadow: 0 24px 68px rgba(15, 23, 42, 0.34);
  backdrop-filter: blur(18px);
  transform-origin: top right;
}

.panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 32px;
}

.panel__eyebrow {
  margin: 0;
  font-size: 13px;
  letter-spacing: 0.32em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.6);
}

.panel__header h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.4px;
}

.panel__subtitle {
  margin: 10px 0 0;
  color: rgba(255, 255, 255, 0.74);
  font-size: 14px;
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
  transform: rotate(90deg);
}

.panel__glow {
  position: absolute;
  inset: -120px 10% auto;
  height: 220px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(227, 232, 238, 0.45), transparent 70%);
  opacity: 0.6;
  pointer-events: none;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 22px;
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
  height: 48px;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(15, 23, 42, 0.26);
  color: #fff;
  padding: 0 16px;
  font-size: 14px;
  transition: border 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.form__field input:focus {
  outline: none;
  border-color: rgba(189, 224, 254, 0.85);
  background: rgba(15, 23, 42, 0.32);
  box-shadow: 0 10px 24px rgba(148, 187, 233, 0.35);
}

.form__field input:hover {
  background: rgba(15, 23, 42, 0.32);
}

.form__field input::placeholder {
  color: rgba(255, 255, 255, 0.52);
}

.form__grid {
  display: flex;
  gap: 16px;
}

.editor {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.editor__surface {
  min-height: 160px;
  border-radius: 22px;
  padding: 20px 22px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(15, 23, 42, 0.18);
  color: #fff;
  font-size: 15px;
  line-height: 1.6;
  overflow-y: auto;
  transition: border 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
  backdrop-filter: blur(12px);
}

.editor__surface:focus {
  outline: none;
  border-color: rgba(189, 224, 254, 0.74);
  background: rgba(15, 23, 42, 0.26);
  box-shadow: inset 0 0 0 1px rgba(189, 224, 254, 0.22), 0 16px 34px rgba(15, 23, 42, 0.32);
}

.editor__surface:empty:before {
  content: attr(placeholder);
  color: rgba(255, 255, 255, 0.4);
}

.editor__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 8px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.14);
  backdrop-filter: blur(12px);
}

.editor__tool,
.editor__upload {
  height: 38px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  font-size: 13px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease, border 0.2s ease;
}

.editor__tool:hover,
.editor__upload:hover {
  background: rgba(255, 255, 255, 0.2);
  box-shadow: 0 10px 18px rgba(15, 23, 42, 0.22);
}

.editor__tool--active {
  border-color: rgba(189, 224, 254, 0.9);
  background: rgba(189, 224, 254, 0.24);
  color: #0f172a;
  font-weight: 600;
}

.form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
}

.form__btn {
  height: 48px;
  padding: 0 26px;
  border-radius: 16px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  color: #1f2937;
  background: var(--frost-highlight);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.form__btn:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 18px 34px rgba(255, 255, 255, 0.32);
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

  .panel__content {
    padding: 32px 28px;
  }
}

@media (max-width: 768px) {
  .panel {
    padding: 12px;
  }

  .panel__content {
    border-radius: 20px;
    padding: 26px 18px;
  }

  .form__grid {
    flex-direction: column;
  }

  .panel__header {
    flex-direction: column;
    align-items: flex-start;
  }

  .panel__close {
    align-self: flex-end;
  }
}
</style>
