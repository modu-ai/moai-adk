// Shared fixtures and the engine stubs for tests/*.test.ts (not itself a test file).
import { mock, type test } from 'claude-code/testing'

export type On = Parameters<Parameters<typeof test>[1]>[1]

/** A queue payload shaped like spec.md M-2: items plus archived and runtime, which the board drops. */
export const queueFixture = (counts = { picked: 1, queued: 1, hold: 1, dropped: 1 }): string => {
  const items: Record<string, unknown>[] = []
  let n = 1
  for (const [state, count] of Object.entries(counts))
    for (let i = 0; i < count; i++, n++)
      items.push({
        id: `t${n}`,
        text: n === 2 ? 'second card text' : `card ${n} ${state} text`,
        state,
        added_at: '2026-09-02T00:00:00Z',
        spec_id: '',
      })
  return JSON.stringify({ items, archived: [{ id: 'a1', text: 'archived' }], runtime: { x: 1 } })
}

export const FACTORY = JSON.stringify({
  run: 'r',
  cards: [
    { card_id: 't1344', state: 'assigned', legacy: false, stage: '', owner: 'lane-1', spec_id: '', lease_expired: false },
    { card_id: 't1399', state: 'plan', legacy: false, stage: 'plan', owner: 'lane-3', spec_id: 'SPEC-X-001', lease_expired: true },
  ],
  unavailable: [],
})

export const PANE = { title: 'moai-board', isFocused: false, bodyColumns: 80, placement: 'inline' } as const

export type Stub = {
  calls: string[][]
  clock: ReturnType<typeof mock.clock>
  /** What the next process answers: 'ok' (fixtures), 'reject' (cannot start), 'exit2', 'garbage'. */
  mode: { value: 'ok' | 'reject' | 'exit2' | 'garbage' }
  queue: { value: string }
}

// Everything beneath the plugin in a test: the engine's own nouns the module calls, a
// process stub that records every argv list and answers by command, and a mocked clock.
export const setup = (on: On, queue = queueFixture()): Stub => {
  const calls: string[][] = []
  const mode: Stub['mode'] = { value: 'ok' }
  const q = { value: queue }
  on('session.start', (_$, e) => ({ cwd: e.cwd }))
  on('command.register', (_$, e) => ({ value: { command: e.name } }))
  on('ui.open', () => ({ value: { isPlaced: true } }))
  on('ui.close', () => ({ value: undefined }))
  on('process.run', (_$, e) => {
    calls.push([...e.argv])
    if (mode.value === 'reject') throw new Error('spawn moai ENOENT')
    const ran = (stdout: string, exitCode = 0, stderr = '') => ({
      value: { exitCode, stdout, stderr, isStdoutTruncated: false, isStderrTruncated: false },
    })
    if (mode.value === 'exit2') return ran('', 2, 'moai: something broke')
    if (mode.value === 'garbage') return ran('<<< not json >>>')
    const joined = e.argv.join(' ')
    return joined.includes('gtd list')
      ? ran(q.value)
      : joined.includes('factory status')
        ? ran(FACTORY)
        : joined.includes('session list')
          ? ran('[]')
          : ran('')
  })
  return { calls, clock: mock.clock(on), mode, queue: q }
}

/** Answers the engine's AskUserQuestion dialog (what `$.ui.ask` raises) with a label, or dismisses it. */
export const answerAsk = (on: On, answer: string | undefined) =>
  on('tool.call', { tool: 'AskUserQuestion' }, (_$, e) =>
    answer === undefined
      ? { deny: 'dismissed' }
      : { result: { questions: e.questions, answers: { [e.questions[0]?.question ?? '']: answer } } },
  )

export const picks = (calls: string[][]): string[][] => calls.filter(argv => argv.includes('next'))
