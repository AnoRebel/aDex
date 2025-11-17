import { defineStore } from 'pinia'

export interface ModalConfig {
  title?: string
  content?: string
  type?: 'info' | 'warning' | 'error' | 'success'
  showCloseButton?: boolean
  persistent?: boolean
}

export const useModalStore = defineStore('modal', {
  state: () => ({
    isOpen: false,
    config: {
      title: '',
      content: '',
      type: 'info' as const,
      showCloseButton: true,
      persistent: false
    } as ModalConfig
  }),

  actions: {
    openModal(config: Partial<ModalConfig> = {}) {
      this.config = { ...this.config, ...config }
      this.isOpen = true
    },

    closeModal() {
      if (!this.config.persistent) {
        this.isOpen = false
      }
    },

    showModal(title: string, content: string, type: ModalConfig['type'] = 'info') {
      this.openModal({ title, content, type })
    },

    showError(title: string, content: string) {
      this.showModal(title, content, 'error')
    },

    showWarning(title: string, content: string) {
      this.showModal(title, content, 'warning')
    },

    showSuccess(title: string, content: string) {
      this.showModal(title, content, 'success')
    }
  }
})
