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
import { computed } from 'vue'
import { useNow, useDateFormat } from '@vueuse/core'

const props = withDefaults(defineProps<{
  use24Hour?: boolean
}>(), {
  use24Hour: true,
})

// VueUse's `useNow` gives us a reactive Date that re-evaluates every
// second. `useDateFormat` then renders it through date-fns' tokens
// (HH 24h, hh 12h, etc) without us touching getHours/getMinutes by
// hand. Replaces the previous setInterval(tick, 1000) scaffold.
const now = useNow({ interval: 1000 })
const timeStr = useDateFormat(now, computed(() => (props.use24Hour ? 'HH:mm:ss' : 'hh:mm:ss')))

const displayChars = computed(() => timeStr.value.split(''))

// Colon blinks once per second — derived from millisecond offset of
// the live `now` ref so we don't need a separate timer.
const colonVisible = computed(() => now.value.getMilliseconds() < 500)
</script>
