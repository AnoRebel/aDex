import { Chart, registerables } from 'chart.js'

// Register Chart.js components
Chart.register(...registerables)

// Default configuration
Chart.defaults.font.family = 'system-ui, -apple-system, sans-serif'
Chart.defaults.color = '#374151' // gray-700

export default defineNuxtPlugin(() => {
  // Chart.js is now available globally
  return {
    provide: {
      chart: Chart
    }
  }
})