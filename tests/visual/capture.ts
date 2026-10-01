/**
 * Visual verification harness.
 *
 * Loads the built aDex-UI frontend in an attached browser and captures
 * screenshots of the states that matter for the Wails v3 migration:
 * the boot screen, the main interface, and a terminal pane.
 *
 * Usage:
 *   bun tests/visual/capture.ts [--url <url>] [--out <dir>] [--tag <name>]
 *
 * Defaults to the dev server on port 9245 (`bun run dev` in frontend/).
 * Screenshots land in docs/evidence/<tag>/.
 *
 * Note on scope: driving the frontend in a browser exercises the UI, the
 * generated bindings' call shapes, and event wiring, but it is NOT the
 * packaged desktop shell — a browser has no Wails IPC peer, so backend
 * calls fail there by design. Treat these as UI-render evidence; the
 * packaged-shell evidence comes from running bin/aDex-UI directly.
 */

import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { CDPSession, readDevToolsEndpoint } from './cdp.ts'

interface Args { url: string; out: string; tag: string; wait: number }

function parseArgs(argv: string[]): Args {
  const get = (flag: string, fallback: string) => {
    const i = argv.indexOf(flag)
    return i !== -1 && argv[i + 1] ? argv[i + 1]! : fallback
  }
  const tag = get('--tag', 'wails-v3-migration')
  return {
    url: get('--url', 'http://127.0.0.1:9245/'),
    out: get('--out', join('docs', 'evidence', tag)),
    tag,
    wait: Number(get('--wait', '6000')),
  }
}

const sleep = (ms: number) => new Promise(r => setTimeout(r, ms))

export async function capture(args: Args) {
  mkdirSync(args.out, { recursive: true })

  const endpoint = readDevToolsEndpoint()
  console.log(`[visual] attaching: ${endpoint}`)
  // Helium may raise an in-browser approval prompt for a new debugging
  // session, so allow a generous window for the handshake.
  const browser = await CDPSession.connect(endpoint, 60_000)

  // Open a dedicated tab so we never disturb the developer's own tabs.
  const { targetId } = await browser.send<{ targetId: string }>('Target.createTarget', {
    url: 'about:blank',
  })
  const { sessionId } = await browser.send<{ sessionId: string }>('Target.attachToTarget', {
    targetId,
    flatten: true,
  })

  const call = (method: string, params: Record<string, unknown> = {}) =>
    browser.send(method, params, sessionId)

  const consoleErrors: string[] = []
  browser.on('Runtime.consoleAPICalled', (p: any) => {
    if (p?.type === 'error') {
      consoleErrors.push((p.args ?? []).map((a: any) => a?.value ?? a?.description ?? '').join(' '))
    }
  })

  await call('Page.enable')
  await call('Runtime.enable')
  await call('Emulation.setDeviceMetricsOverride', {
    width: 1600, height: 1000, deviceScaleFactor: 1, mobile: false,
  })

  /** Poll until the Vue app has rendered something, or time out. */
  const waitForRender = async (minChars: number, timeoutMs: number) => {
    const deadline = Date.now() + timeoutMs
    while (Date.now() < deadline) {
      const r = await call('Runtime.evaluate', {
        expression: 'document.getElementById("__nuxt")?.innerHTML.length ?? 0',
        returnByValue: true,
      }) as any
      if ((r?.result?.value ?? 0) >= minChars) return true
      await sleep(400)
    }
    return false
  }

  const shot = async (name: string) => {
    const { data } = await call('Page.captureScreenshot', { format: 'png' }) as { data: string }
    const file = join(args.out, `${name}.png`)
    writeFileSync(file, Buffer.from(data, 'base64'))
    console.log(`[visual] captured ${file}`)
    return file
  }

  console.log(`[visual] navigating: ${args.url}`)
  await call('Page.navigate', { url: args.url })

  // Boot screen: wait for the app to mount, then capture while the boot
  // sequence is still running.
  const rendered = await waitForRender(1000, 30_000)
  if (!rendered) {
    throw new Error(
      `Nothing rendered at ${args.url} within 30s. Is the dev server running ` +
        '(bun run dev in frontend/) or the URL correct?',
    )
  }
  await shot('01-boot-screen')

  // Main interface: after the boot sequence has had time to finish.
  await sleep(args.wait)
  await shot('02-main-interface')

  // Terminal pane.
  await call('Runtime.evaluate', {
    expression: `(() => {
      const el = document.querySelector('.xterm, [class*="terminal"], [data-terminal]')
      if (el) el.scrollIntoView({ block: 'center' })
      return !!el
    })()`,
    returnByValue: true,
  })
  await sleep(1200)
  await shot('03-terminal')

  const title = await call('Runtime.evaluate', {
    expression: 'document.title', returnByValue: true,
  }) as any

  await browser.send('Target.closeTarget', { targetId })
  browser.close()

  return { title: title?.result?.value, consoleErrors }
}

if (import.meta.main) {
  const args = parseArgs(Bun.argv.slice(2))
  const result = await capture(args)
  console.log(`\n[visual] document.title: ${result.title}`)
  if (result.consoleErrors.length) {
    console.log(`[visual] console errors (${result.consoleErrors.length}):`)
    for (const e of result.consoleErrors.slice(0, 10)) console.log(`  - ${e}`)
  } else {
    console.log('[visual] no console errors')
  }
  console.log(`[visual] evidence written to ${args.out}/`)
}
