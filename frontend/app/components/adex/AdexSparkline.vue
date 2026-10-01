<template>
  <div ref="containerRef" class="sparkline">
    <!-- Only `height` is passed, deliberately. The Vue host writes these
         props straight into inline styles, so a numeric `width` becomes
         `width: 300` — invalid CSS — and the box collapses; omitting it
         yields `width: 100%`, which is what a flex-sized panel wants anyway.
         `height` is the one dimension that must be given: left unset it
         defaults to 320px and overflows these short panels. -->
    <Chart
      v-if="chartHeight > 8"
      :definition="definition"
      :height="chartHeight"
      class="sparkline-chart"
      :aria-label="ariaLabel"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * Rolling sparkline for the live metric panels (network traffic, CPU load).
 *
 * Built on TanStack Charts rather than hand-rolled canvas so the panels share
 * one charting grammar instead of each reimplementing scales, grids and fills.
 *
 * Two deliberate choices:
 *
 * - **SVG rendering.** TanStack's canvas renderer is not an option here: the
 *   Vue `Chart` host is SVG-only (it hardcodes `createSvgChartRenderer`), so a
 *   per-mark `canvasChartRenderer` silently renders nothing. The series are
 *   small — three marks over 60 points — so an SVG path per mark is cheap.
 *
 * - **Theme colours are read from CSS custom properties** and passed in as
 *   real values. The chart cannot resolve `var(--color_r)` itself, and these
 *   panels must restyle when the user switches theme, so the colour is a
 *   reactive input rather than a static string.
 */
import { ref, computed } from 'vue'
import { useElementSize } from '@vueuse/core'
import { defineChart, areaY, lineY } from '@tanstack/charts'
import { dot } from '@tanstack/charts/dot'
import { scaleLinear } from '@tanstack/charts/scales/linear'
import { Chart } from '@tanstack/charts/vue'

const props = withDefaults(defineProps<{
  /** Oldest-to-newest samples. Rendered against their index on x. */
  data: number[]
  /** Upper bound of the y scale. Held steady by the caller so the line does
   *  not rescale on every sample. */
  scaleMax?: number
  /** `r, g, b` triple for the theme accent. */
  accent?: string
  ariaLabel?: string
}>(), {
  scaleMax: 1,
  accent: '170, 207, 209',
  ariaLabel: 'Activity over time',
})

const containerRef = ref<HTMLElement | null>(null)

// The chart needs a real pixel height; width it can resolve itself from the
// container. useElementSize keeps this current as the column reflows.
const { height } = useElementSize(containerRef)
const chartHeight = computed(() => Math.round(height.value))

const rows = computed(() => props.data.map((value, index) => ({ index, value })))

const stroke = computed(() => `rgba(${props.accent}, 0.9)`)
const fill = computed(() => `rgba(${props.accent}, 0.18)`)

/** Just the newest sample, so the leading edge keeps its marker dot. */
const head = computed(() => {
  const last = rows.value[rows.value.length - 1]
  return last ? [last] : []
})

const definition = computed(() => defineChart({
  // Zero padding and no guides: these are sparklines inside an existing panel
  // frame, which already supplies its own labels and scale readout.
  margin: 0,
  guides: false,
  // Scales live under `scales`, and each needs an explicit scale factory —
  // a bare `{ domain }` is rejected with "Chart scales must define resolved
  // 'x' and 'y' entries". Fixed domains (rather than inferred ones) keep the
  // line from rescaling on every sample.
  scales: {
    x: { scale: scaleLinear, domain: [0, Math.max(1, props.data.length - 1)] },
    y: { scale: scaleLinear, domain: [0, props.scaleMax > 0 ? props.scaleMax : 1] },
  },
  marks: [
    areaY(rows.value, {
      x: 'index',
      y: 'value',
      fill: fill.value,
    }),
    lineY(rows.value, {
      x: 'index',
      y: 'value',
      stroke: stroke.value,
      strokeWidth: 1.5,
    }),
    dot(head.value, {
      x: 'index',
      y: 'value',
      fill: `rgba(${props.accent}, 1)`,
      r: 2,
    }),
  ],
}))
</script>

<style scoped>
/* Fill the parent by absolute positioning rather than `height: 100%`.
   The parent is a flex item whose height comes from `flex-grow`, and a
   percentage height against that resolved to 0 — which left the chart
   unmounted and the panel blank. Insetting to the parent's padding box
   gives a real measured height regardless of how the parent got its size.
   (The parent is already `position: relative; overflow: hidden`.) */
.sparkline {
  position: absolute;
  inset: 0;
}

.sparkline-chart {
  width: 100%;
  height: 100%;
  display: block;
}
</style>
