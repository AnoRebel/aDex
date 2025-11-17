<template>
  <div class="notification-container">
    <TransitionGroup name="notification" tag="div">
      <div
        v-for="notification in notifications"
        :key="notification.id"
        :class="['notification', `notification-${notification.type}`]"
        @click="removeNotification(notification.id)"
      >
        <div class="notification-icon">
          {{ getIcon(notification.type) }}
        </div>
        <div class="notification-content">
          <h4 class="notification-title">{{ notification.title }}</h4>
          <p class="notification-message">{{ notification.message }}</p>
        </div>
        <button class="notification-close" @click.stop="removeNotification(notification.id)">
          ✕
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Notification {
  id: string
  title: string
  message: string
  type: 'info' | 'success' | 'warning' | 'error'
  duration?: number
}

const notifications = ref<Notification[]>([])

const addNotification = (notification: Omit<Notification, 'id'>) => {
  const id = Date.now().toString()
  const newNotification = {
    ...notification,
    id,
    duration: notification.duration || 5000
  }
  
  notifications.value.push(newNotification)
  
  // Auto-remove after duration
  if (newNotification.duration > 0) {
    setTimeout(() => {
      removeNotification(id)
    }, newNotification.duration)
  }
}

const removeNotification = (id: string) => {
  const index = notifications.value.findIndex(n => n.id === id)
  if (index > -1) {
    notifications.value.splice(index, 1)
  }
}

const getIcon = (type: Notification['type']) => {
  const icons = {
    info: 'ℹ️',
    success: '✅',
    warning: '⚠️',
    error: '❌'
  }
  return icons[type]
}

// Expose method globally for other components
onMounted(() => {
  // Make it available globally via app.config.globalProperties if needed
})
</script>

<style scoped>
.notification-container {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 1001;
  pointer-events: none;
}

.notification {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  padding: 1rem;
  margin-bottom: 0.5rem;
  min-width: 300px;
  max-width: 400px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  pointer-events: all;
  cursor: pointer;
  transition: all 0.3s ease;
}

.notification:hover {
  transform: translateX(-4px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

.notification-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.notification-content {
  flex: 1;
}

.notification-title {
  margin: 0 0 0.25rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
}

.notification-message {
  margin: 0;
  font-size: 0.75rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.notification-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.25rem;
  font-size: 0.75rem;
  border-radius: 0.25rem;
  transition: all 0.3s ease;
}

.notification-close:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

/* Notification type styles */
.notification-info {
  border-left: 4px solid #0099ff;
}

.notification-success {
  border-left: 4px solid #00ff00;
}

.notification-warning {
  border-left: 4px solid #ffaa00;
}

.notification-error {
  border-left: 4px solid #ff3333;
}

/* Transitions */
.notification-enter-active,
.notification-leave-active {
  transition: all 0.3s ease;
}

.notification-enter-from {
  transform: translateX(100%);
  opacity: 0;
}

.notification-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

.notification-move {
  transition: transform 0.3s ease;
}

@media (max-width: 768px) {
  .notification-container {
    top: auto;
    bottom: 1rem;
    right: 1rem;
    left: 1rem;
  }
  
  .notification {
    min-width: auto;
    max-width: none;
  }
}
</style>
