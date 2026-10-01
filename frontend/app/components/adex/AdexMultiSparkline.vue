<template>
  <div ref="containerRef" class="multi-sparkline">
    <!-- See AdexSparkline for why only `height` is passed. -->
    <Chart
      v-if="chartHeight > 8"
      :definition="definition"
      :height="chartHeight"
      class="multi-sparkline-chart"
      :aria-label="ariaLabel"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * Multi-series sparkline — several rolling buffers on shared axes.
 *
 * Used for the per-core CPU panel, where each core is its own line over the
 * same time window. The series are distinguished by opacity rather than hue,
 * matching the single-accent look of the rest of the interface: later cores
 * are drawn more opaque, so the band reads as one texture instead of a dozen
 * competing colours.
 *
 * Rendering is SVG — see AdexSparkline for why the canvas renderer is not
 * available through TanStack's Vue host.
 */
import { ref, computed } from 'vue'
import { useElementSize } from '@vueuse/core'
import { defineChart, lineY } from '@tanstack/charts'
import { scaleLinear } from '@tanstack/charts/scales/linear'
import { Chart } from '@tanstack/charts/vue'

const props = withDefaults(defineProps<{
  /** One rolling buffer per series, oldest sample first. */
  series: number[][]
  /** Fixed upper bound of the y scale (CPU percentages: 100). */
  scaleMax?: number
  /** `r, g, b` triple for the theme accent. */
  accent?: string
  /** Window width in samples. Series shorter than this are right-aligned, so
   *  a freshly-started buffer grows in from the right rather than stretching. */
  windowSize?: number
  ariaLabel?: string
}>(), {
  scaleMax: 100,
  accent: '170, 207, 209',
  windowSize: 60,
  ariaLabel: 'Per-core activity over time',
})

const containerRef = ref<HTMLElement | null>(null)
const { height } = useElementSize(containerRef)
const chartHeight = computed(() => Math.round(height.value))

/**
 * Flatten every series into one row set tagged with its series index.
 *
 * `z` is the grouping channel: one line is drawn per distinct value, which is
 * what keeps this a single mark rather than one mark per core.
 */
const rows = computed(() => {
  const out: Array<{ index: number; value: number; series: number }> = []
  const window = Math.max(2, props.windowSize)

  props.series.forEach((buffer, series) => {
    if (!buffer || buffer.length < 2) return
    // Right-align: a buffer with fewer samples than the window starts partway
    // across rather than being stretched to fill it.
    const offset = window - buffer.length
    buffer.forEach((value, i) => {
      out.push({ index: offset + i, value, series })
    })
  })
  return out
})

/** Later series are more opaque, giving the stack depth without extra hues. */
function seriesAlpha(series: number): number {
  const count = props.series.length
  if (count <= 1) return 0.9
  return 0.3 + 0.7 * (series / (count - 1))
}

const definition = computed(() => defineChart({
  margin: 0,
  guides: false,
  scales: {
    x: { scale: scaleLinear, domain: [0, Math.max(1, props.windowSize - 1)] },
    y: { scale: scaleLinear, domain: [0, props.scaleMax > 0 ? props.scaleMax : 100] },
  },
  marks: [
    lineY(rows.value, {
      x: 'index',
      y: 'value',
      z: 'series',
      stroke: (row: { series: number }) =>
        `rgba(${props.accent}, ${seriesAlpha(row.series).toFixed(2)})`,
      strokeWidth: 1,
    }),
  ],
}))
</script>

<style scoped>
/* Absolute fill: see AdexSparkline — a percentage height against a
   flex-grown parent resolves to 0 and leaves the chart unmounted. */
.multi-sparkline {
  position: absolute;
  inset: 0;
}

.multi-sparkline-chart {
  width: 100%;
  height: 100%;
  display: block;
}
</style>
