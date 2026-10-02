// Shared fixtures and the engine stubs for tests/*.test.ts (not itself a test file).
// Mirrors the sibling mod's support: everything beneath the plugin in a test is
// stubbed per test (M-3); the test `$` raises the engine's own events at the hooks.
import { mock, type test } from 'claude-code/testing'

export type On = Parameters<Parameters<typeof test>[1]>[1]

// Doctor box rows exactly as the CLI prints them (plan §B.4, C5 re-measure):
// the row is `    <status>    <check name>  <message>  ` inside the box borders.
const box = (rows: string[]): string =>
  ['╭──────────────────────────────────────────╮', '│  System Diagnostics                      │', '│    STATUS  CHECK        MESSAGE          │', ...rows.map(r => `│  ${r}  │`), '╰──────────────────────────────────────────╯'].join('\n')

export const BINARY_BEHIND = box(['  warn    Binary Freshness  binary is behind source tree (binary: 802a72235, HEAD: c7b72b430)'])
export const BINARY_FRESH = box(['  ok      Binary Freshness  binary matches source HEAD (802a72235)'])
export const MCP_OK = box(['  ok      MCP Server Version  no running moai MCP server recorded'])
export const MCP_STALE = box(['  warn    MCP Server Version  running MCP server is stale (pid 4242: 0a1b2c3d4; binary: 802a72235)'])

const store = (dir: string, files: number, cap: number) => ({
  store: { dir, origin: 'test' },
  exists: true,
  topic_files: files,
  cap,
  index_lines: 3,
  findings: null,
})

export const MEMORY_OK = JSON.stringify([store('/a', 12, 50), store('/b', 49, 50)])
export const MEMORY_OVER = JSON.stringify([store('/a', 1422, 50), store('/b', 60, 50), store('/c', 12, 50)])

export type Stub = {
  calls: string[][]
  inits: (number | undefined)[]
  clock: ReturnType<typeof mock.clock>
  status: { lines: (string | undefined)[] }
  toasts: { lines: string[] }
  /** What the next process answers: 'ok' (fixtures), 'reject' (cannot start), 'exit2', 'garbage'. */
  mode: { value: 'ok' | 'reject' | 'exit2' | 'garbage' }
  /** The next `Binary Freshness` check blocks on this promise; one use only. */
  hold: { wait: Promise<void> | undefined }
}

export const setup = (on: On): Stub => {
  const calls: string[][] = []
  const inits: (number | undefined)[] = []
  const status = { lines: [] as (string | undefined)[] }
  const toasts = { lines: [] as string[] }
  const mode = { value: 'ok' } as Stub['mode']
  const hold: Stub['hold'] = { wait: undefined }
  on('session.start', (_$, e) => ({ cwd: e.cwd }))
  // The engine implements none of the mod's event nouns (M-3): the test answers
  // the raised events with their core echo shapes.
  on('session.measure', (_$, e) => ({ changed: e.changed }))
  on('session.receive', (_$, e) => ({ text: e.text }))
  on('ui.status', (_$, e) => {
    status.lines.push(e.text)
    return { value: undefined }
  })
  on('ui.toast', (_$, e) => {
    toasts.lines.push(e.text)
    return { value: undefined }
  })
  on('process.run', async (_$, e) => {
    calls.push([...e.argv])
    inits.push(e.init?.timeoutMs)
    if (mode.value === 'reject') throw new Error('spawn moai ENOENT')
    const ran = (stdout: string, exitCode = 0) => ({
      value: { exitCode, stdout, stderr: '', isStdoutTruncated: false, isStderrTruncated: false },
    })
    if (mode.value === 'exit2') return ran('', 2)
    if (mode.value === 'garbage') return ran('<<< nothing parseable here >>>')
    const joined = e.argv.join(' ')
    if (joined === 'moai doctor --check Binary Freshness') {
      if (hold.wait !== undefined) {
        const wait = hold.wait
        hold.wait = undefined
        await wait
      }
      return ran(BINARY_BEHIND)
    }
    if (joined === 'moai doctor --check MCP Server Version') return ran(MCP_OK)
    if (joined === 'moai memory doctor --json') return ran(MEMORY_OK)
    return ran('')
  })
  return { calls, inits, clock: mock.clock(on), status, toasts, mode, hold }
}

/** A warned context measure on a 200K window (soft 90): 91% warns. */
export const warnMeasure = () => ({ context: { window: 200_000, percent: 91 }, rateLimits: [], changed: ['context'] })

/** A quiet measure: 31% on the same window, nothing warned. */
export const quietMeasure = () => ({ context: { window: 200_000, percent: 31 }, rateLimits: [], changed: ['context'] })
