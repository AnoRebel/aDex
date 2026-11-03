// tests/composables/use-update-checker.spec.ts
//
// The composable decides when to check and mirrors the Go updater's events
// into state. The bindings and the runtime are mocked; each test re-imports
// the module so its module-scope state starts fresh.

import { beforeEach, describe, expect, it, vi } from 'vitest'

const svc = vi.hoisted(() => ({
  Info: vi.fn(),
  Check: vi.fn(),
  Install: vi.fn(),
  Restart: vi.fn(),
}))
vi.mock('../../bindings/aDex/internal/services/updater/service.js', () => svc)

const DAY = 24 * 60 * 60 * 1000
const release = { version: '2.0.0', name: 'aDex 2.0.0', notes: 'n', publishedAt: '', url: 'https://example.test/v2.0.0' }

async function load(opts: { canSelfUpdate?: boolean } = {}) {
  svc.Info.mockResolvedValue({
    available: true, currentVersion: '1.0.0', state: 'idle',
    canSelfUpdate: opts.canSelfUpdate ?? true, selfUpdateReason: opts.canSelfUpdate === false ? 'not running from an AppImage' : '',
  })
  vi.resetModules()
  const runtime = await import('@wailsio/runtime')
  const handlers = new Map<string, (e: { data?: unknown }) => void>()
  vi.mocked(runtime.Events.On).mockImplementation(((name: string, cb: (e: { data?: unknown }) => void) => {
    handlers.set(name, cb)
    return () => {}
  }) as never)
  const { useUpdateChecker } = await import('../../app/composables/useUpdateChecker')
  const u = useUpdateChecker()
  await vi.waitFor(() => expect(u.info.value).not.toBeNull())
  const fire = (name: string, data?: unknown) => handlers.get(`wails:updater:${name}`)?.({ data })
  return { u, fire, runtime }
}

beforeEach(() => {
  localStorage.clear()
  vi.clearAllMocks()
  svc.Check.mockResolvedValue(release)
  svc.Install.mockResolvedValue(undefined)
  svc.Restart.mockResolvedValue(undefined)
})

describe('useUpdateChecker — scheduling', () => {
  it('does nothing when automatic checks are disabled', async () => {
    const { u } = await load()
    u.settings.value.enabled = false
    await u.maybeCheckOnLaunch()
    expect(svc.Check).not.toHaveBeenCalled()
  })

  it('does nothing in manual mode', async () => {
    const { u } = await load()
    u.settings.value.interval = 'manual'
    await u.maybeCheckOnLaunch()
    expect(svc.Check).not.toHaveBeenCalled()
  })

  it('checks on every launch in launch mode', async () => {
    const { u } = await load()
    u.state.value.lastChecked = new Date().toISOString()
    await u.maybeCheckOnLaunch()
    expect(svc.Check).toHaveBeenCalledOnce()
  })

  it.each([
    ['daily', 23 * 60 * 60 * 1000, false],
    ['daily', DAY + 1000, true],
    ['weekly', 6 * DAY, false],
    ['weekly', 7 * DAY + 1000, true],
  ] as const)('%s mode, last check %d ms ago → checks: %s', async (interval, ago, expected) => {
    const { u } = await load()
    u.settings.value.interval = interval
    u.state.value.lastChecked = new Date(Date.now() - ago).toISOString()
    await u.maybeCheckOnLaunch()
    expect(svc.Check).toHaveBeenCalledTimes(expected ? 1 : 0)
  })

  it('a manual check runs regardless of the frequency setting', async () => {
    const { u } = await load()
    u.settings.value.enabled = false
    expect(await u.check()).toEqual(release)
    expect(u.phase.value).toBe('available')
    expect(u.state.value.lastChecked).not.toBe('')
  })

  it('records a failed check as an error, not an update', async () => {
    const { u } = await load()
    svc.Check.mockRejectedValue(new Error('offline'))
    expect(await u.check()).toBeNull()
    expect(u.phase.value).toBe('error')
    expect(u.state.value.lastError).toBe('offline')
    expect(u.hasUpdate.value).toBe(false)
  })

  it('reports up to date when the service returns no release', async () => {
    const { u } = await load()
    svc.Check.mockResolvedValue(null)
    await u.check()
    expect(u.phase.value).toBe('up-to-date')
    expect(u.hasUpdate.value).toBe(false)
  })
})

describe('useUpdateChecker — skip list', () => {
  it('hides a skipped version but not a newer one', async () => {
    const { u } = await load()
    await u.check()
    u.skipVersion('2.0.0')
    expect(u.hasUpdate.value).toBe(false)

    svc.Check.mockResolvedValue({ ...release, version: '2.1.0' })
    await u.check()
    expect(u.hasUpdate.value).toBe(true)

    u.unskipAll()
    expect(u.settings.value.skippedVersions).toEqual([])
  })
})

describe('useUpdateChecker — install flow', () => {
  it('follows the updater events through to ready', async () => {
    const { u, fire } = await load()
    await u.check()
    await u.install()
    expect(svc.Install).toHaveBeenCalledOnce()

    fire('download-started')
    expect(u.phase.value).toBe('downloading')
    fire('download-progress', { written: 50, total: 100, rate: 10 })
    expect(u.progress.value).toEqual({ written: 50, total: 100, rate: 10 })
    fire('verifying')
    expect(u.phase.value).toBe('verifying')
    fire('installing')
    expect(u.phase.value).toBe('installing')
    fire('update-ready')
    expect(u.phase.value).toBe('ready')

    await u.restart()
    expect(svc.Restart).toHaveBeenCalledOnce()
  })

  it('surfaces the failed stage from the error event', async () => {
    const { u, fire } = await load()
    await u.check()
    fire('error', { stage: 'verify', message: 'digest mismatch' })
    expect(u.phase.value).toBe('error')
    expect(u.error.value).toEqual({ stage: 'verify', message: 'digest mismatch' })
  })

  it('never calls Install when this copy cannot replace itself', async () => {
    const { u, runtime } = await load({ canSelfUpdate: false })
    await u.check()
    expect(u.canSelfUpdate.value).toBe(false)
    await u.install()
    expect(svc.Install).not.toHaveBeenCalled()

    u.openReleasePage()
    expect(runtime.Browser.OpenURL).toHaveBeenCalledWith(release.url)
  })
})
