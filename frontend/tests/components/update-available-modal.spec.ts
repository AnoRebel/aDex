// tests/components/update-available-modal.spec.ts
//
// Renders the update modal in each flow state and checks which actions it
// offers. useUpdateChecker is replaced by plain refs so each state can be set
// directly.

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'

const checker = vi.hoisted(() => ({ current: null as any }))

vi.mock('~/composables/useAdexAudio', () => ({ useAdexAudio: () => ({ playCue: vi.fn() }) }))
vi.mock('~/composables/useUpdateChecker', () => ({ useUpdateChecker: () => checker.current }))

import UpdateAvailableModal from '~/components/ui/UpdateAvailableModal.vue'

function fakeChecker(phase: string, opts: { canSelfUpdate?: boolean; error?: any; progress?: any } = {}) {
  return {
    release: ref({ version: '2.0.0', name: 'aDex 2.0.0', notes: 'Fixes', publishedAt: '2026-10-01T00:00:00Z', url: 'https://example.test' }),
    phase: ref(phase),
    progress: ref(opts.progress ?? null),
    error: ref(opts.error ?? null),
    info: ref({ selfUpdateReason: opts.canSelfUpdate === false ? 'not running from an AppImage' : '' }),
    currentVersion: ref('1.0.0'),
    canSelfUpdate: ref(opts.canSelfUpdate ?? true),
    install: vi.fn(),
    restart: vi.fn(),
    check: vi.fn(),
    openReleasePage: vi.fn(),
    skipVersion: vi.fn(),
  }
}

function render(phase: string, opts?: Parameters<typeof fakeChecker>[1]) {
  checker.current = fakeChecker(phase, opts)
  const w = mount(UpdateAvailableModal, {
    props: { modelValue: true },
    global: { stubs: { Teleport: true, Transition: false } },
  })
  const buttons = () => w.findAll('.ua-actions button').map(b => b.text())
  const click = (label: string) => w.findAll('.ua-actions button').find(b => b.text() === label)!.trigger('click')
  return { w, buttons, click, c: checker.current }
}

describe('UpdateAvailableModal', () => {
  beforeEach(() => { checker.current = null })

  it('available: offers install, skip and later', async () => {
    const { w, buttons, click, c } = render('available')
    expect(w.text()).toContain('UPDATE AVAILABLE')
    expect(w.text()).toContain('Fixes')
    expect(buttons()).toEqual(['SKIP THIS VERSION', 'REMIND ME LATER', 'INSTALL'])
    await click('INSTALL')
    expect(c.install).toHaveBeenCalledOnce()
    await click('SKIP THIS VERSION')
    expect(c.skipVersion).toHaveBeenCalledWith('2.0.0')
  })

  it('focuses the main action on open and when the step changes', async () => {
    checker.current = fakeChecker('available')
    const w = mount(UpdateAvailableModal, {
      props: { modelValue: true },
      global: { stubs: { Teleport: true, Transition: false } },
      attachTo: document.body,
    })
    await flushPromises()
    expect(document.activeElement?.textContent).toBe('INSTALL')
    checker.current.phase.value = 'ready'
    await flushPromises()
    expect(document.activeElement?.textContent).toBe('RESTART & APPLY')
    w.unmount()
  })

  it('notify-only: offers the release page instead of install', async () => {
    const { w, buttons, click, c } = render('available', { canSelfUpdate: false })
    expect(buttons()).toEqual(['SKIP THIS VERSION', 'REMIND ME LATER', 'OPEN RELEASE PAGE'])
    expect(w.find('[data-testid="ua-notify-only"]').text()).toContain('not running from an AppImage')
    await click('OPEN RELEASE PAGE')
    expect(c.openReleasePage).toHaveBeenCalledOnce()
    expect(c.install).not.toHaveBeenCalled()
  })

  it('downloading: shows progress and only a hide action', () => {
    const { w, buttons } = render('downloading', { progress: { written: 512 * 1024, total: 1024 * 1024, rate: 2048 } })
    expect(w.find('[role="progressbar"]').attributes('aria-valuenow')).toBe('50')
    expect(w.text()).toContain('512.0 KB / 1.0 MB')
    expect(buttons()).toEqual(['HIDE'])
  })

  it.each(['verifying', 'installing'])('%s: shows status only', (phase) => {
    const { w, buttons } = render(phase)
    expect(w.text()).toContain(phase === 'verifying' ? 'VERIFYING SIGNATURE' : 'INSTALLING')
    expect(buttons()).toEqual(['HIDE'])
  })

  it('ready: offers restart & apply and later', async () => {
    const { buttons, click, c } = render('ready')
    expect(buttons()).toEqual(['LATER', 'RESTART & APPLY'])
    await click('RESTART & APPLY')
    expect(c.restart).toHaveBeenCalledOnce()
  })

  it('error: shows the stage and message, and retries the failed step', async () => {
    const { w, buttons, click, c } = render('error', { error: { stage: 'verify', message: 'digest mismatch' } })
    expect(w.text()).toContain('VERIFY FAILED')
    expect(w.text()).toContain('digest mismatch')
    expect(buttons()).toEqual(['CLOSE', 'RETRY'])
    await click('RETRY')
    expect(c.install).toHaveBeenCalledOnce()
    expect(c.check).not.toHaveBeenCalled()
  })

  it('closing emits update:modelValue false', async () => {
    const { w, click } = render('ready')
    await click('LATER')
    expect(w.emitted('update:modelValue')?.[0]).toEqual([false])
  })
})
