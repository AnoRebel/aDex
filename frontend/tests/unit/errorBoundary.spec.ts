import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'
import ErrorBoundary from '../../app/components/ui/ErrorBoundary.vue'

const Exploding = defineComponent({
  props: { shouldThrow: { type: Boolean, default: true } },
  setup(props) {
    return () => {
      if (props.shouldThrow) throw new Error('panel exploded')
      return h('div', { class: 'recovered' }, 'ok')
    }
  },
})

describe('ErrorBoundary', () => {
  it('contains a child error instead of propagating it', async () => {
    // Mounting must NOT throw: that is what stops one panel taking down the app.
    const wrapper = mount(ErrorBoundary, {
      props: { label: 'CPU' },
      slots: { default: () => h(Exploding) },
    })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('CPU UNAVAILABLE')
    expect(wrapper.text()).toContain('panel exploded')
  })

  it('offers a retry affordance', async () => {
    const wrapper = mount(ErrorBoundary, {
      props: { label: 'CPU' },
      slots: { default: () => h(Exploding) },
    })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.eb-retry').exists()).toBe(true)
  })

  it('recovers when the underlying fault is gone', async () => {
    const shouldThrow = ref(true)
    const wrapper = mount(ErrorBoundary, {
      props: { label: 'CPU' },
      slots: { default: () => h(Exploding, { shouldThrow: shouldThrow.value }) },
    })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('CPU UNAVAILABLE')

    // Clear the fault, then retry: the boundary should render the child again.
    shouldThrow.value = false
    await wrapper.find('.eb-retry').trigger('click')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).not.toContain('UNAVAILABLE')
  })

  it('renders its slot normally when nothing throws', () => {
    const wrapper = mount(ErrorBoundary, {
      props: { label: 'CPU' },
      slots: { default: () => h('div', { class: 'fine' }, 'all good') },
    })
    expect(wrapper.find('.fine').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('UNAVAILABLE')
  })
})
