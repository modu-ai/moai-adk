// Engine tests: run with `CLAUDE_CONFIG_DIR=<empty dir> claude plugin test mods/moai-board`.
// They exercise hook dispatch, timers and the pick flow against the engine itself
// (never a surface's paint). Pure parsers are covered by tests/pure/ under bun.
import { expect, test } from 'claude-code/testing'
import { PANE, answerAsk, picks, queueFixture, setup, stubSpecFiles } from './support'

const mount = ($: Parameters<Parameters<typeof test>[1]>[0]) =>
  $.ui.mount({ plugin: 'moai-board', surface: 'terminal', component: 'Pane', requestId: 'moai-board', props: PANE })

// ---- polling and dispatch -------------------------------------------------------------

test('poll: no process while the pane is closed', async ($, on) => {
  const { calls, clock } = setup(on)
  await $.session.start({ cwd: '/work', surface: 'terminal', isInteractive: true })
  await clock.advance(120_000)
  expect(calls.length).toBe(0)
})

test('poll: timer cancelled on ui.close', async ($, on) => {
  const { calls, clock } = setup(on)
  await $.command.run({ command: 'moai-board' })
  const opened = calls.length
  expect(opened).toBeGreaterThan(0)
  await clock.advance(15_000)
  expect(calls.length).toBeGreaterThan(opened)
  // The kit cannot raise ui.close itself; the pane's own close button does ($.ui.close,
  // origin plugin), and the module stops its timer there.
  const ui = await mount($)
  await ui.press({ key: 'close' })
  const closed = calls.length
  await clock.advance(120_000)
  expect(calls.length).toBe(closed)
})

test('poll: only read argv across 10 ticks', async ($, on) => {
  const { calls, clock } = setup(on)
  await $.command.run({ command: 'moai-board' })
  for (let i = 0; i < 10; i++) await clock.advance(15_000)
  expect(calls.length).toBeGreaterThanOrEqual(10)
  const allowed = new Set([
    'moai gtd list --json',
    'moai gtd list --limit 0',
    'moai factory status --json',
    'moai session list --json',
    'moai spec status --list',
  ])
  for (const argv of calls) {
    expect(allowed.has(argv.join(' '))).toBe(true)
    expect(argv.includes('next')).toBe(false)
  }
})

test('dispatch: no pick argv from session.start, command.run or render', async ($, on) => {
  const { calls } = setup(on)
  await $.session.start({ cwd: '/work', surface: 'terminal', isInteractive: true })
  await $.command.run({ command: 'moai-board' })
  const ui = await mount($)
  await ui.unmount()
  expect(calls.length).toBeGreaterThan(0)
  for (const argv of calls) expect(argv.includes('next')).toBe(false)
})

// ---- pick: confirmation, no-op paths -------------------------------------------------------

const pressPick = async ($: Parameters<Parameters<typeof test>[1]>[0]) => {
  await $.command.run({ command: 'moai-board' })
  const ui = await mount($)
  await ui.press({ key: 'open:t2' })
  await ui.press({ key: 'pick:t2' })
  return ui
}

test('pick: confirm label runs exactly one argv', async ($, on) => {
  const { calls } = setup(on)
  answerAsk(on, 'Pick')
  await pressPick($)
  expect(picks(calls)).toEqual([['moai', 'gtd', 'next', 't2', '--expect', 'second card text']])
})

test('pick: cancel runs no process', async ($, on) => {
  const { calls } = setup(on)
  answerAsk(on, 'Cancel')
  await pressPick($)
  expect(picks(calls).length).toBe(0)
})

test('pick: other text runs no process', async ($, on) => {
  const { calls } = setup(on)
  answerAsk(on, 'pick')
  await pressPick($)
  expect(picks(calls).length).toBe(0)
})

test('pick: rejected ask runs no process', async ($, on) => {
  const { calls } = setup(on)
  answerAsk(on, undefined)
  await pressPick($)
  expect(picks(calls).length).toBe(0)
})

// ---- pick: the confirmation reads the queue as it is after the pick (REQ-MBM-005) ------------------

const PICKED_T2 = queueFixture({ picked: 2, queued: 0, hold: 1, dropped: 1 })

const settledNotice = async (ui: Awaited<ReturnType<typeof mount>>) => {
  for (let i = 0; i < 50; i++) {
    const found = await ui.find({ key: 'notice' })
    if (found !== undefined) return found.text
  }
  return undefined
}

test('pick: no poll in flight, the post-pick read confirms', async ($, on) => {
  const stub = setup(on)
  stub.afterPick.value = PICKED_T2
  answerAsk(on, 'Pick')
  const ui = await pressPick($)
  expect(await settledNotice(ui)).toBe('Picked t2.')
})

test('pick: poll in flight when the dialog is answered, the post-pick read confirms', async ($, on) => {
  const stub = setup(on)
  stub.afterPick.value = PICKED_T2
  answerAsk(on, 'Pick')
  await $.command.run({ command: 'moai-board' })
  const ui = await mount($)
  await ui.press({ key: 'open:t2' })
  let release: () => void = () => {}
  stub.hold.wait = new Promise<void>(resolve => {
    release = resolve
  })
  await stub.clock.advance(15_000) // a timer poll starts here and blocks, its answer still showing t2 queued
  await ui.press({ key: 'pick:t2' })
  release()
  expect(await settledNotice(ui)).toBe('Picked t2.')
  expect(picks(stub.calls).length).toBe(1)
})

// ---- fail-soft ------------------------------------------------------------------------------

test('failsoft: no hook rejects', async ($, on) => {
  const stub = setup(on)
  for (const mode of ['reject', 'exit2', 'garbage'] as const) {
    stub.mode.value = mode
    await $.session.start({ cwd: '/work', surface: 'terminal', isInteractive: true })
    await $.command.run({ command: 'moai-board' })
    const ui = await mount($)
    await ui.press({ key: 'refresh' })
    await ui.press({ key: 'tab-lanes' })
    await ui.press({ key: 'refresh' })
    await stub.clock.advance(15_000)
    await ui.unmount()
  }
  expect(stub.calls.length).toBeGreaterThan(3)
})

test('failsoft: last good data stays visible dimmed', async ($, on) => {
  const stub = setup(on)
  await $.command.run({ command: 'moai-board' })
  const ui = await mount($)
  const before = await ui.find({ key: 'open:t2' })
  expect(before).toBeDefined()
  expect(before?.props['dimColor']).not.toBe(true)
  stub.mode.value = 'reject'
  await ui.press({ key: 'refresh' })
  const after = await ui.find({ key: 'open:t2' })
  expect(after).toBeDefined()
  expect(after?.props['dimColor']).toBe(true)
  const status = await ui.find({ key: 'status' })
  expect(status?.text).toContain('not found on PATH')
})

// ---- SPEC tab: the file calls are wired in the module, the guard is covered under bun ----------

test('spec-tab: list, open and read through the guard', async ($, on) => {
  const { calls } = setup(on)
  const reads = stubSpecFiles(on, '/work', {
    '/work/.moai/specs/SPEC-A-001/spec.md': { realPath: '/real/work/.moai/specs/SPEC-A-001/spec.md', text: '# Spec body text' },
  })
  await $.command.run({ command: 'moai-board' })
  const ui = await mount($)
  await ui.press({ key: 'tab-spec' })
  expect(calls.map(argv => argv.join(' '))).toContain('moai spec status --list')
  expect(await ui.find({ key: 'spec:SPEC-A-001' })).toBeDefined()
  expect(await ui.find({ key: 'spec:SPEC-B-001' })).toBeUndefined() // completed: hidden by the default filter
  await ui.press({ key: 'spec:SPEC-A-001' })
  await ui.press({ key: 'file:spec.md' })
  expect((await ui.find({ key: 'doc-0' }))?.text).toContain('Spec body text')
  expect(reads).toEqual(['/real/work/.moai/specs/SPEC-A-001/spec.md'])
  await ui.press({ key: 'back' })
  expect(await ui.find({ key: 'spec:SPEC-A-001' })).toBeDefined()
})

test('spec-tab: a file resolving outside the folder is not read', async ($, on) => {
  setup(on)
  const reads = stubSpecFiles(on, '/work', {
    '/work/.moai/specs/SPEC-A-001/plan.md': { realPath: '/etc/secret.md', text: 'SECRET' },
  })
  await $.command.run({ command: 'moai-board' })
  const ui = await mount($)
  await ui.press({ key: 'tab-spec' })
  await ui.press({ key: 'spec:SPEC-A-001' })
  await ui.press({ key: 'file:plan.md' })
  expect(reads.length).toBe(0)
  expect((await ui.find({ key: 'doc-notice' }))?.text).toContain('outside')
  expect(await ui.find({ key: 'doc-0' })).toBeUndefined()
})
