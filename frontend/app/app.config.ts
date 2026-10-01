// Nuxt UI v4 global theme overrides.
//
// We tweak just the components that read out-of-place against the
// sci-fi terminal aesthetic. Everything else inherits the @nuxt/ui
// defaults. Targets:
//   - selectMenu : trigger button reads as a styled trigger with the
//                  same chrome as our `.settings-input` text inputs,
//                  not the default underlined-text look that landed
//                  unstyled when we wrapped USelectMenu without an
//                  app-level config (the `.settings-select` class on
//                  the root only styled the wrapper div, not the
//                  inner button slot).
//   - select     : same shape for the simpler USelect variant.
//
// Tailwind utility classes match the rest of the modal: monospace
// font, semi-transparent terminal-green border that brightens on
// focus, full-width inside the field column.
//
// MUST be wrapped in defineAppConfig(). Nuxt only registers app config
// exported through that macro; a bare `export default {...}` is silently
// ignored, which is exactly what happened here — every override below was
// dead, so the select controls rendered with stock Nuxt UI styling while the
// hand-styled inputs beside them looked correct. Every example in the Nuxt UI
// theming docs uses defineAppConfig.
export default defineAppConfig({
  ui: {
    selectMenu: {
      slots: {
        // `base` is the trigger button. Replace the default rounded
        // look with our flat terminal-style chrome.
        base: [
          'w-full inline-flex items-center justify-between gap-2',
          'px-2.5 py-1.5',
          '!bg-[var(--color_light_black,#05080d)]',
          '!border !border-[var(--color_accent_dimmed,rgba(170,207,209,0.35))]',
          '!text-[var(--color_accent,rgb(170,207,209))]',
          '!font-mono !rounded-none',
          '!ring-0 !shadow-none',
          'cursor-pointer outline-none transition-colors',
          'hover:border-[var(--color_accent,rgb(170,207,209))]',
          'focus:border-[var(--color_accent,rgb(170,207,209))]',
          'disabled:opacity-30 disabled:cursor-not-allowed',
          'data-[state=open]:border-[var(--color_accent,rgb(170,207,209))]',
        ].join(' '),
        value: 'truncate text-left',
        placeholder: 'truncate opacity-50',
        trailingIcon: 'shrink-0 size-4 text-[var(--color_accent,rgb(170,207,209))] group-data-[state=open]:rotate-180 transition-transform',
        // Dropdown content panel — keep terminal palette.
        content: [
          'w-(--reka-select-trigger-width) max-h-60',
          '!bg-[var(--color_black,#000)]',
          '!border !border-[var(--color_accent,rgb(170,207,209))]',
          '!rounded-none !ring-0 shadow-lg overflow-hidden',
          'flex flex-col',
          'data-[state=open]:animate-[scale-in_100ms_ease-out]',
          'data-[state=closed]:animate-[scale-out_100ms_ease-in]',
          'pointer-events-auto',
        ].join(' '),
        viewport: 'relative scroll-py-1 overflow-y-auto flex-1',
        item: [
          'group relative w-full flex items-center select-none outline-none',
          'px-2.5 py-1.5 text-sm',
          '!text-[var(--color_accent,rgb(170,207,209))]',
          '!rounded-none cursor-pointer',
          'data-[highlighted]:!bg-[var(--color_accent_glow,rgba(170,207,209,0.15))]',
          'data-[state=checked]:text-[var(--color_accent,rgb(170,207,209))]',
          'data-[state=checked]:font-bold',
          'data-disabled:cursor-not-allowed data-disabled:opacity-50',
        ].join(' '),
        itemLabel: 'truncate',
        input: 'border-b border-[var(--color_accent_dimmed,rgba(170,207,209,0.35))] bg-transparent px-2 py-1 text-[var(--color_accent,rgb(170,207,209))] outline-none w-full',
      },
    },
    select: {
      slots: {
        base: [
          'w-full inline-flex items-center justify-between gap-2',
          'px-2.5 py-1.5',
          '!bg-[var(--color_light_black,#05080d)]',
          '!border !border-[var(--color_accent_dimmed,rgba(170,207,209,0.35))]',
          '!text-[var(--color_accent,rgb(170,207,209))]',
          '!font-mono !rounded-none',
          '!ring-0 !shadow-none',
          'cursor-pointer outline-none transition-colors',
          'hover:border-[var(--color_accent,rgb(170,207,209))]',
          'focus:border-[var(--color_accent,rgb(170,207,209))]',
          'disabled:opacity-30 disabled:cursor-not-allowed',
        ].join(' '),
      },
    },
  },
})
