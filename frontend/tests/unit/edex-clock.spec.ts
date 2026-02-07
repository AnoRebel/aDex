import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import EdexClock from '~/app/components/edex/EdexClock.vue'

describe('EdexClock Component', () => {
  let wrapper: VueWrapper

  beforeEach(() => {
    vi.useFakeTimers()
    // Fix a known date: 2026-02-05 14:30:45.200
    vi.setSystemTime(new Date(2026, 1, 5, 14, 30, 45, 200))
  })

  afterEach(() => {
    wrapper?.unmount()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  describe('Rendering', () => {
    it('renders with mod-clock class', () => {
      wrapper = mount(EdexClock)
      expect(wrapper.find('.mod-clock').exists()).toBe(true)
    })

    it('renders inside a mod-panel wrapper', () => {
      wrapper = mount(EdexClock)
      expect(wrapper.find('.mod-panel').exists()).toBe(true)
    })

    it('displays time as individual characters with span and em elements', () => {
      wrapper = mount(EdexClock)
      // Time is 14:30:45 => characters: 1,4,:,3,0,:,4,5
      // Digits should be in <span>, colons in <em>
      const spans = wrapper.findAll('.mod-clock > span')
      const ems = wrapper.findAll('.mod-clock > em')

      // 6 digit spans + 2 colon ems = 8 total children
      expect(spans.length).toBe(6)
      expect(ems.length).toBe(2)
    })

    it('displays time in HH:MM:SS format (6 digits + 2 colons)', async () => {
      wrapper = mount(EdexClock)
      await wrapper.vm.$nextTick()
      const text = wrapper.find('.mod-clock').text()
      // Should contain the time characters: 14:30:45
      // The text content joins all the span/em text
      expect(text).toContain('1')
      expect(text).toContain('4')
      expect(text).toContain(':')
      expect(text).toContain('3')
      expect(text).toContain('0')
      expect(text).toContain('4')
      expect(text).toContain('5')
    })

    it('each digit is wrapped in a <span> element', async () => {
      wrapper = mount(EdexClock)
      await wrapper.vm.$nextTick()
      const spans = wrapper.findAll('.mod-clock > span')
      // For 14:30:45, digits are 1, 4, 3, 0, 4, 5
      const digitTexts = spans.map(s => s.text())
      expect(digitTexts).toEqual(['1', '4', '3', '0', '4', '5'])
    })

    it('each colon is wrapped in an <em> element', () => {
      wrapper = mount(EdexClock)
      const ems = wrapper.findAll('.mod-clock > em')
      expect(ems.length).toBe(2)
      expect(ems[0].text()).toBe(':')
      expect(ems[1].text()).toBe(':')
    })
  })

  describe('24-hour format (default)', () => {
    it('uses 24-hour format by default', async () => {
      wrapper = mount(EdexClock)
      await wrapper.vm.$nextTick()
      const spans = wrapper.findAll('.mod-clock > span')
      const hourDigits = spans[0].text() + spans[1].text()
      // 14:30:45 in 24h => hour is "14"
      expect(hourDigits).toBe('14')
    })

    it('shows hour 0 as 00 in 24-hour format', () => {
      vi.setSystemTime(new Date(2026, 1, 5, 0, 5, 9, 200))
      wrapper = mount(EdexClock)
      const spans = wrapper.findAll('.mod-clock > span')
      const hourDigits = spans[0].text() + spans[1].text()
      expect(hourDigits).toBe('00')
    })

    it('shows hour 23 correctly in 24-hour format', async () => {
      vi.setSystemTime(new Date(2026, 1, 5, 23, 59, 59, 200))
      wrapper = mount(EdexClock)
      await wrapper.vm.$nextTick()
      const spans = wrapper.findAll('.mod-clock > span')
      const hourDigits = spans[0].text() + spans[1].text()
      expect(hourDigits).toBe('23')
    })
  })

  describe('12-hour format', () => {
    it('supports 12-hour format when use24Hour=false', async () => {
      vi.setSystemTime(new Date(2026, 1, 5, 14, 30, 45, 200))
      wrapper = mount(EdexClock, {
        props: { use24Hour: false }
      })
      await wrapper.vm.$nextTick()
      const spans = wrapper.findAll('.mod-clock > span')
      const hourDigits = spans[0].text() + spans[1].text()
      // 14:00 in 12h => 2 PM => "02"
      expect(hourDigits).toBe('02')
    })

    it('shows 12 for noon in 12-hour format', async () => {
      vi.setSystemTime(new Date(2026, 1, 5, 12, 0, 0, 200))
      wrapper = mount(EdexClock, {
        props: { use24Hour: false }
      })
      await wrapper.vm.$nextTick()
      const spans = wrapper.findAll('.mod-clock > span')
      const hourDigits = spans[0].text() + spans[1].text()
      expect(hourDigits).toBe('12')
    })

    it('shows 12 for midnight in 12-hour format', async () => {
      vi.setSystemTime(new Date(2026, 1, 5, 0, 0, 0, 200))
      wrapper = mount(EdexClock, {
        props: { use24Hour: false }
      })
      await wrapper.vm.$nextTick()
      const spans = wrapper.findAll('.mod-clock > span')
      const hourDigits = spans[0].text() + spans[1].text()
      expect(hourDigits).toBe('12')
    })

    it('shows correct hour for morning times in 12-hour format', async () => {
      vi.setSystemTime(new Date(2026, 1, 5, 9, 15, 30, 200))
      wrapper = mount(EdexClock, {
        props: { use24Hour: false }
      })
      await wrapper.vm.$nextTick()
      const spans = wrapper.findAll('.mod-clock > span')
      const hourDigits = spans[0].text() + spans[1].text()
      expect(hourDigits).toBe('09')
    })
  })

  describe('Timer and updates', () => {
    it('creates a setInterval on mount', () => {
      const setIntervalSpy = vi.spyOn(global, 'setInterval')
      wrapper = mount(EdexClock)
      // setInterval is called with tick function and 1000ms
      expect(setIntervalSpy).toHaveBeenCalledWith(expect.any(Function), 1000)
    })

    it('updates the time when the timer fires', async () => {
      vi.setSystemTime(new Date(2026, 1, 5, 14, 30, 45, 200))
      wrapper = mount(EdexClock)
      await wrapper.vm.$nextTick()

      // Verify initial time
      let spans = wrapper.findAll('.mod-clock > span')
      expect(spans[4].text() + spans[5].text()).toBe('45') // seconds

      // Advance time by 1 second and trigger the interval
      vi.advanceTimersByTime(1000)
      await wrapper.vm.$nextTick()

      spans = wrapper.findAll('.mod-clock > span')
      expect(spans[4].text() + spans[5].text()).toBe('46')
    })

    it('updates minutes correctly when seconds roll over', async () => {
      vi.setSystemTime(new Date(2026, 1, 5, 14, 30, 59, 200))
      wrapper = mount(EdexClock)
      await wrapper.vm.$nextTick()

      // Advance to next minute
      vi.advanceTimersByTime(1000)
      await wrapper.vm.$nextTick()

      const spans = wrapper.findAll('.mod-clock > span')
      const minutes = spans[2].text() + spans[3].text()
      const seconds = spans[4].text() + spans[5].text()
      expect(minutes).toBe('31')
      expect(seconds).toBe('00')
    })

    it('clears the interval on unmount', () => {
      const clearIntervalSpy = vi.spyOn(global, 'clearInterval')
      wrapper = mount(EdexClock)
      wrapper.unmount()
      expect(clearIntervalSpy).toHaveBeenCalled()
    })
  })

  describe('Colon blinking', () => {
    it('has clock-blink class when colonVisible is true (ms < 500)', () => {
      vi.setSystemTime(new Date(2026, 1, 5, 14, 30, 45, 200)) // ms=200 < 500
      wrapper = mount(EdexClock)
      expect(wrapper.find('.mod-clock').classes()).toContain('clock-blink')
    })

    it('does not have clock-blink class when colonVisible is false (ms >= 500)', async () => {
      vi.setSystemTime(new Date(2026, 1, 5, 14, 30, 45, 600)) // ms=600 >= 500
      wrapper = mount(EdexClock)
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.mod-clock').classes()).not.toContain('clock-blink')
    })
  })
})
