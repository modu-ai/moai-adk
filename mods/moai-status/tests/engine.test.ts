// Engine tests: run with `CLAUDE_CONFIG_DIR=<empty dir> claude plugin test mods/moai-status`.
// They exercise hook dispatch and the render path against the engine itself
// (never a surface's paint). Pure parsers and builders are covered by tests/pure/
// under bun. The test `$` exposes no $.state — the classification in state is
// observed through the engine's own read path, the AbovePrompt render.
import { expect, test } from 'claude-code/testing'
import { setup, warnMeasure } from './support'

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
