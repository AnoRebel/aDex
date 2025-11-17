<template>
  <div v-if="isOpen" class="modal-overlay" @click="handleOverlayClick">
    <div class="modal-container" :class="typeClasses">
      <div class="modal-header">
        <h3 class="modal-title">{{ title }}</h3>
        <button 
          v-if="showCloseButton" 
          @click="closeModal" 
          class="close-button"
          aria-label="Close modal"
        >
          ✕
        </button>
      </div>
      
      <div class="modal-body">
        <p class="modal-content">{{ content }}</p>
        <slot></slot>
      </div>
      
      <div class="modal-footer">
        <button @click="closeModal" class="btn btn-primary">
          OK
        </button>
        <button 
          v-if="!persistent" 
          @click="closeModal" 
          class="btn btn-secondary"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useModalStore } from '~/stores/modal'

const modalStore = useModalStore()

const isOpen = computed(() => modalStore.isOpen)
const title = computed(() => modalStore.config.title || 'Notice')
const content = computed(() => modalStore.config.content || '')
const type = computed(() => modalStore.config.type || 'info')
const showCloseButton = computed(() => modalStore.config.showCloseButton !== false)
const persistent = computed(() => modalStore.config.persistent || false)

const typeClasses = computed(() => ({
  'modal-info': type.value === 'info',
  'modal-warning': type.value === 'warning',
  'modal-error': type.value === 'error',
  'modal-success': type.value === 'success'
}))

const closeModal = () => {
  modalStore.closeModal()
}

const handleOverlayClick = (event: MouseEvent) => {
  if (event.target === event.currentTarget && !persistent.value) {
    closeModal()
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-container {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5);
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow: hidden;
  animation: slideIn 0.3s ease-out;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
}

.modal-title {
  margin: 0;
  color: var(--text-primary);
  font-size: 1.25rem;
}

.close-button {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.25rem;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.25rem;
  transition: all 0.3s ease;
}

.close-button:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.modal-body {
  padding: 1rem;
}

.modal-content {
  margin: 0;
  color: var(--text-primary);
  line-height: 1.5;
}

.modal-footer {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  padding: 1rem;
  border-top: 1px solid var(--border-color);
}

.btn {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.3s ease;
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.btn:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-primary);
}

.btn-primary {
  background: var(--accent-primary);
  color: var(--bg-primary);
  border-color: var(--accent-primary);
}

.btn-primary:hover {
  background: var(--accent-secondary);
  border-color: var(--accent-secondary);
}

/* Modal type styles */
.modal-info {
  border-left: 4px solid #0099ff;
}

.modal-warning {
  border-left: 4px solid #ffaa00;
}

.modal-error {
  border-left: 4px solid #ff3333;
}

.modal-success {
  border-left: 4px solid #00ff00;
}

@keyframes slideIn {
  from {
    transform: translateY(-20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
</style>
