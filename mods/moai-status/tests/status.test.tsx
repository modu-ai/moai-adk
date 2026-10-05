// Engine tests that mount the AbovePrompt band and the Spinner through the
// plugin (never a surface's paint). The test registers its own ui.render hook
// for the same component as the chain's TERMINAL: the test env's core draws
// nothing for these components, so the last hook answers with a drawing of its
// own. The terminal records that it ran — the mod resolves it through next(e)
// when it composes, and a mod that drops the upstream never lets it run.
// Test-side hooks register before the test's first $ call (the engine requires it).
import { expect, test } from 'claude-code/testing'
import { quietMeasure, setup, warnMeasure } from './support'

const ABOVE_PROPS = { hasSurvey: false, isWorking: false, maxRows: 10, bodyColumns: 80 }
const SPINNER_PROPS = { word: 'Sauteing', message: null, suffix: '…', mode: 'thinking' } as const

type On = Parameters<Parameters<typeof test>[1]>[1]
type Test$ = Parameters<Parameters<typeof test>[1]>[0]

// The terminal drawing for an AbovePrompt mount; records each run in `ran`.
const sentinelHook = (on: On, ran: number[]) =>
  on('ui.render', { component: 'AbovePrompt' }, async ($, e) => {
    ran.push(1)
    const T = $.ui.resolve(e)
    return T.Box({ key: 'upstream-wrap', children: T.Text({ key: 'upstream-sentinel', children: 'UPSTREAM-SENTINEL' }) })
  })

const mountAbove = async ($: Test$, props: Record<string, unknown>) =>
  $.ui.mount({ plugin: 'moai-status', surface: 'terminal', component: 'AbovePrompt', props: props as never })

// ---- strip (AC-MSM-004 engine) ------------------------------------------------------

test('strip: warn draws a one-line band', async ($, on) => {
  const ran: number[] = []
  sentinelHook(on, ran)
  setup(on)
  await $.session.measure(warnMeasure())
  const ui = await mountAbove($, ABOVE_PROPS)
  const strip = await ui.find({ key: 'moai-status-strip-line' })
  expect(strip?.text).toContain('moai-status:')
  expect(strip?.text).toContain('ctx 91% (warn at 90)')
  // The mod resolved the upstream through next(e): the terminal ran, so the
  // upstream drawing survives beside the strip (REQ-MSM-004's compose rule —
  // a mod that dropped the upstream would never have let the terminal run).
  expect(ran.length).toBe(1)
  await ui.unmount()
})

test('strip: hasSurvey passes through', async ($, on) => {
  const ran: number[] = []
  sentinelHook(on, ran)
  setup(on)
  await $.session.measure(warnMeasure())
  const ui = await mountAbove($, { ...ABOVE_PROPS, hasSurvey: true })
  expect(await ui.find({ key: 'moai-status-strip-line' })).toBeUndefined()
  // The pass-through still resolved the terminal: the band keeps drawing.
  expect(ran.length).toBe(1)
  await ui.unmount()
})

test('strip: quiet classification passes through', async ($, on) => {
  const ran: number[] = []
  sentinelHook(on, ran)
  setup(on)
  await $.session.measure(quietMeasure())
  const ui = await mountAbove($, ABOVE_PROPS)
  expect(await ui.find({ key: 'moai-status-strip-line' })).toBeUndefined()
  expect(ran.length).toBe(1)
  await ui.unmount()
})

// ---- spinner suffix (AC-MSM-005 engine) ---------------------------------------------

test('suffix: warn appends to the incoming suffix', async ($, on) => {
  // The observer is the chain's terminator: it records the event it received —
  // carrying the mod's rewrite when the mod rewrote props via next — and draws it.
  const seen: { suffix: string; word: string; message: string | null }[] = []
  on('ui.render', { component: 'Spinner' }, async ($, e) => {
    const T = $.ui.resolve(e)
    seen.push({ suffix: e.props.suffix, word: e.props.word, message: e.props.message })
    return T.Box({ key: 'spinner-observed', children: T.Text({ key: 'spinner-line', children: `${e.props.word}${e.props.suffix}` }) })
  })
  setup(on)
  await $.session.measure(warnMeasure())
  const ui = await $.ui.mount({ plugin: 'moai-status', surface: 'terminal', component: 'Spinner', props: SPINNER_PROPS as never })
  await ui.unmount()
  expect(seen.length).toBe(1)
  // The rewritten props carry the incoming ellipsis plus the marker; word and
  // message are never touched (REQ-MSM-005).
  expect(seen[0]?.suffix).toBe('… · ctx 91%')
  expect(seen[0]?.suffix.startsWith('…')).toBe(true)
  expect(seen[0]?.word).toBe('Sauteing')
  expect(seen[0]?.message).toBeNull()
})

test('suffix: quiet passes the event through', async ($, on) => {
  const seen: { suffix: string }[] = []
  on('ui.render', { component: 'Spinner' }, async ($, e) => {
    const T = $.ui.resolve(e)
    seen.push({ suffix: e.props.suffix })
    return T.Box({ key: 'spinner-observed', children: T.Text({ key: 'spinner-line', children: `${e.props.word}${e.props.suffix}` }) })
  })
  setup(on)
  await $.session.measure(quietMeasure())
  const ui = await $.ui.mount({ plugin: 'moai-status', surface: 'terminal', component: 'Spinner', props: SPINNER_PROPS as never })
  await ui.unmount()
  expect(seen.length).toBe(1)
  expect(seen[0]?.suffix).toBe('…')
})
