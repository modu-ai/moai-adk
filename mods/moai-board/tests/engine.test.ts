// Engine tests: run with `CLAUDE_CONFIG_DIR=<empty dir> claude plugin test mods/moai-board`.
// They exercise hook dispatch, timers and the pick flow against the engine itself
// (never a surface's paint). Pure parsers are covered by tests/pure/ under bun.
import { expect, mock, test } from 'claude-code/testing'

type On = Parameters<Parameters<typeof test>[1]>[1]

const QUEUE = JSON.stringify({
  items: [
    { id: 't1', text: 'first card text', state: 'picked', added_at: '2026-09-01T00:00:00Z', spec_id: '' },
    { id: 't2', text: 'second card text', state: 'queued', added_at: '2026-09-02T00:00:00Z', spec_id: '' },
    { id: 't3', text: 'dropped card', state: 'dropped', added_at: '', spec_id: '' },
  ],
  archived: [],
  runtime: {},
})
const FACTORY = JSON.stringify({ run: 'r', cards: [], unavailable: [] })

// Everything beneath the plugin in a test: the engine's own nouns the module calls,
// a process.run stub that records every argv list and answers by command, and a mocked clock.
const setup = (on: On) => {
  const calls: string[][] = []
  on('session.start', (_$, e) => ({ cwd: e.cwd }))
  on('command.register', (_$, e) => ({ value: { command: e.name } }))
  on('ui.open', () => ({ value: { isPlaced: true } }))
  on('ui.close', () => ({ value: undefined }))
  on('process.run', (_$, e) => {
    calls.push([...e.argv])
    const joined = e.argv.join(' ')
    const stdout = joined.includes('gtd list')
      ? QUEUE
      : joined.includes('factory status')
        ? FACTORY
        : joined.includes('session list')
          ? '[]'
          : ''
    return { value: { exitCode: 0, stdout, stderr: '', isStdoutTruncated: false, isStderrTruncated: false } }
  })
  return { calls, clock: mock.clock(on) }
}

const PANE = { title: 'moai-board', isFocused: false, bodyColumns: 80, placement: 'inline' } as const

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
  // origin plugin), which is the event the module's hook answers by cancelling the timer.
  const ui = await $.ui.mount({
    plugin: 'moai-board',
    surface: 'terminal',
    component: 'Pane',
    requestId: 'moai-board',
    props: PANE,
  })
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
  const ui = await $.ui.mount({
    plugin: 'moai-board',
    surface: 'terminal',
    component: 'Pane',
    requestId: 'moai-board',
    props: PANE,
  })
  await ui.unmount()
  expect(calls.length).toBeGreaterThan(0)
  for (const argv of calls) expect(argv.includes('next')).toBe(false)
})
