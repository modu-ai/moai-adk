// Engine tests: run with `CLAUDE_CONFIG_DIR=<empty dir> claude plugin test mods/moai-status`.
// They exercise hook dispatch, timers and the render path against the engine
// itself (never a surface's paint). Pure parsers and builders are covered by
// tests/pure/ under bun. The test `$` exposes no $.state — classifications in
// state are observed through the engine's own read path, the render hooks.
import { expect, test } from 'claude-code/testing'
import { HEALTH_POLL_MIN_MS, HEALTH_POLL_MS } from '../hooks/data'
import { BINARY_FRESH, MEMORY_OVER, MCP_STALE, setup, warnMeasure } from './support'

// ---- measure: classify the pushed figures into state (AC-MSM-006) ------------------

test('measure: warn classification reaches state', async ($, on) => {
  setup(on)
  // The chain terminator for the mount: the test env's core draws nothing for
  // AbovePrompt, so the last hook answers with a plain drawing.
  on('ui.render', { component: 'AbovePrompt' }, async ($, e) => {
    const T = $.ui.resolve(e)
    return T.Box({ key: 'upstream-wrap', children: T.Text({ key: 'upstream-core', children: 'CORE-DRAWING' }) })
  })
  await $.session.measure(warnMeasure())
  // The mod's render hook reads the classification from $.state (the engine's
  // own subscription path): the strip drawing carries the warned level.
  const ui = await $.ui.mount({
    plugin: 'moai-status',
    surface: 'terminal',
    component: 'AbovePrompt',
    props: { hasSurvey: false, isWorking: false, maxRows: 10, bodyColumns: 80 },
  })
  const strip = await ui.find({ key: 'moai-status-strip-line' })
  expect(strip?.text).toContain('ctx 91% (warn at 90)')
  await ui.unmount()
})

test('measure: next(e) is always returned', async ($, on) => {
  setup(on)
  const result = await $.session.measure(warnMeasure())
  expect(result).toEqual({ changed: ['context'] })
  const quiet = await $.session.measure({ context: { window: 200_000, percent: 31 }, rateLimits: [], changed: ['context'] })
  expect(quiet).toEqual({ changed: ['context'] })
})

// ---- toast: a pure observer of the inbound delivery (AC-MSM-007) -------------------

test('toast: delivery is toasted before passing', async ($, on) => {
  const { toasts } = setup(on)
  const result = await $.session.receive({ origin: { kind: 'peer' }, text: 'lane says hi\nsecond line' })
  // One line: the origin kind plus the bounded first-line excerpt.
  expect(toasts.lines).toEqual(['peer: lane says hi'])
  expect(result).toEqual({ text: 'lane says hi\nsecond line' })
})

test('toast: toast failure still passes the delivery', async ($, on) => {
  const stub = setup(on)
  stub.toastThrows.value = true
  const result = await $.session.receive({ origin: { kind: 'coordinator' }, text: 'urgent dispatch' })
  expect(result).toEqual({ text: 'urgent dispatch' })
})

test('toast: consumed is never produced', async ($, on) => {
  const stub = setup(on)
  const first = await $.session.receive({ origin: { kind: 'peer' }, text: 'one' })
  stub.toastThrows.value = true
  const second = await $.session.receive({ origin: { kind: 'task-notification' }, text: 'two' })
  // No path answers the delivery: every result is the queued shape, `{ text }`.
  expect('consumed' in first).toBe(false)
  expect('consumed' in second).toBe(false)
  expect(first).toEqual({ text: 'one' })
  expect(second).toEqual({ text: 'two' })
})

// ---- health cycle (AC-MSM-009) --------------------------------------------------------

const START = { cwd: '/work', surface: 'terminal', isInteractive: true } as const

test('health-cycle: only table argv across ticks', async ($, on) => {
  const { calls, clock } = setup(on)
  await $.session.start(START)
  await clock.advance(HEALTH_POLL_MS * 3)
  expect(calls.length).toBeGreaterThanOrEqual(3)
  const allowed = new Set([
    'moai doctor --check Binary Freshness',
    'moai doctor --check MCP Server Version',
    'moai memory doctor --json',
  ])
  for (const argv of calls) expect(allowed.has(argv.join(' '))).toBe(true)
})

test('health-cycle: warn cycle pins one line', async ($, on) => {
  const { status, clock } = setup(on)
  await $.session.start(START)
  await clock.advance(HEALTH_POLL_MS)
  expect(status.lines).toEqual(['moai-status: binary 802a72235 behind c7b72b430'])
})

test('health-cycle: healthy cycle clears', async ($, on) => {
  const stub = setup(on)
  stub.out.binary = BINARY_FRESH
  await $.session.start(START)
  await stub.clock.advance(HEALTH_POLL_MS)
  expect(stub.status.lines.length).toBe(1)
  expect(stub.status.lines[0]).toBeUndefined()
})

test('health-cycle: a tick while one runs is dropped', async ($, on) => {
  const stub = setup(on)
  stub.hold.wait = new Promise<void>(resolve => {
    stub.hold.release = resolve
  })
  await $.session.start(START)
  await stub.clock.advance(HEALTH_POLL_MS) // tick 1 starts and blocks on the binary check
  await stub.clock.settle()
  const inFlight = stub.calls.length
  expect(inFlight).toBe(1)
  await stub.clock.advance(HEALTH_POLL_MS) // tick 2 finds the gate closed: dropped
  expect(stub.calls.length).toBe(inFlight)
  stub.hold.release()
  await stub.clock.settle()
  // Exactly one round ran: the three table argvs and nothing more.
  expect(stub.calls.length).toBe(3)
})

test('health-cycle: timer cancelled on session.end', async ($, on) => {
  const { calls, clock } = setup(on)
  await $.session.start(START)
  await $.session.end({ reason: 'other', sessionId: 's1', resume: { id: 's1' } })
  await clock.advance(HEALTH_POLL_MS * 2)
  expect(calls.length).toBe(0)
})

test('health-cycle: interval respects the floor', async ($, on) => {
  // The clock records no interval (debt N-5), so the floor is asserted
  // behaviorally: nothing fires inside HEALTH_POLL_MIN_MS, a cycle fires after.
  const { calls, clock } = setup(on)
  await $.session.start(START)
  await clock.advance(HEALTH_POLL_MIN_MS - 1)
  expect(calls.length).toBe(0)
  await clock.advance(HEALTH_POLL_MS)
  expect(calls.length).toBeGreaterThanOrEqual(3)
})

test('health-cycle: run carries the 20s timeout', async ($, on) => {
  const { inits, clock } = setup(on)
  await $.session.start(START)
  await clock.advance(HEALTH_POLL_MS)
  expect(inits.length).toBeGreaterThanOrEqual(3)
  for (const timeoutMs of inits) expect(timeoutMs).toBe(20_000)
})

// ---- fail-soft (AC-MSM-010) -------------------------------------------------------------

test('failsoft: a rejected source shows unknown and keeps last good', async ($, on) => {
  const stub = setup(on)
  await $.session.start(START)
  await stub.clock.advance(HEALTH_POLL_MS)
  // Cycle 1: the binary check warns with both SHAs.
  expect(stub.status.lines[0]).toContain('binary 802a72235 behind c7b72b430')
  stub.mode.value = 'reject'
  await stub.clock.advance(HEALTH_POLL_MS)
  // Cycle 2: every source is unknown — shown as `?`, never healthy, never a
  // session failure. The last good classification stands in state (judged purely
  // by health-merge: in tests/pure/health.spec.ts; the test `$` exposes no state).
  const last = stub.status.lines[stub.status.lines.length - 1]
  expect(last).toContain('binary ?')
  expect(last).toContain('mcp ?')
  expect(last).toContain('memory ?')
})

test('failsoft: a throwing hook does not reach the session', async ($, on) => {
  // A broken hook further down the chain, narrowed to the peer origin, throws
  // mid-dispatch BEFORE setup's answering hook: the engine skips it (one line
  // naming mod, event and reason — the M-3 shape) and the dispatch still
  // completes; the mod's own observer work already happened.
  on(
    'session.receive',
    { origin: 'peer' },
    () => {
      throw new Error('broken sibling hook')
    },
  )
  const stub = setup(on)
  const result = await $.session.receive({ origin: { kind: 'peer' }, text: 'hello' })
  expect(result).toEqual({ text: 'hello' })
  expect(stub.toasts.lines).toEqual(['peer: hello'])
})

