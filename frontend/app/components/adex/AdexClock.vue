<template>
  <div class="mod-panel">
    <div class="mod-clock" :class="{ 'clock-blink': colonVisible }">
      <template v-for="(char, index) in displayChars" :key="index">
        <em v-if="char === ':'">:</em>
        <span v-else>{{ char }}</span>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'

const props = withDefaults(defineProps<{
  use24Hour?: boolean
}>(), {
  use24Hour: true
})

const hours = ref('00')
const minutes = ref('00')
const seconds = ref('00')
const colonVisible = ref(true)

let tickTimer: ReturnType<typeof setInterval> | null = null

const displayChars = computed(() => {
  const timeStr = `${hours.value}:${minutes.value}:${seconds.value}`
  return timeStr.split('')
})

function tick() {
  const now = new Date()
  let h = now.getHours()

  if (!props.use24Hour) {
    h = h % 12 || 12
  }

  hours.value = String(h).padStart(2, '0')
  minutes.value = String(now.getMinutes()).padStart(2, '0')
  seconds.value = String(now.getSeconds()).padStart(2, '0')
  colonVisible.value = now.getMilliseconds() < 500
}

onMounted(() => {
  tick()
  tickTimer = setInterval(tick, 1000)
})

onBeforeUnmount(() => {
  if (tickTimer) {
    clearInterval(tickTimer)
    tickTimer = null
  }
})
</script>
