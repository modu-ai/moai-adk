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

export const SPEC_LIST = [
  'SPEC-ID                        Status          Modified',
  '-'.repeat(60),
  'SPEC-A-001                     draft           2026-10-02 17:28',
  'SPEC-B-001                     completed       2026-10-02 17:28',
  '',
].join('\n')

export const PANE = { title: 'moai-board', isFocused: false, bodyColumns: 80, placement: 'inline' } as const

export type Stub = {
  calls: string[][]
  clock: ReturnType<typeof mock.clock>
  /** What the next process answers: 'ok' (fixtures), 'reject' (cannot start), 'exit2', 'garbage'. */
  mode: { value: 'ok' | 'reject' | 'exit2' | 'garbage' }
  queue: { value: string }
  /** The next `gtd list --json` blocks on this promise (its answer is fixed when it starts); one use only. */
  hold: { wait: Promise<void> | undefined }
  /** When set, a pick argv (`next`) replaces the queue payload with this one. */
  afterPick: { value: string | undefined }
}

// Everything beneath the plugin in a test: the engine's own nouns the module calls, a
// process stub that records every argv list and answers by command, and a mocked clock.
export const setup = (on: On, queue = queueFixture()): Stub => {
  const calls: string[][] = []
  const mode: Stub['mode'] = { value: 'ok' }
  const q = { value: queue }
  const hold: Stub['hold'] = { wait: undefined }
  const afterPick: Stub['afterPick'] = { value: undefined }
  on('session.start', (_$, e) => ({ cwd: e.cwd }))
  on('command.register', (_$, e) => ({ value: { command: e.name } }))
  on('ui.open', () => ({ value: { isPlaced: true } }))
  on('ui.close', () => ({ value: undefined }))
  on('process.run', async (_$, e) => {
    calls.push([...e.argv])
    if (mode.value === 'reject') throw new Error('spawn moai ENOENT')
    const ran = (stdout: string, exitCode = 0, stderr = '') => ({
      value: { exitCode, stdout, stderr, isStdoutTruncated: false, isStderrTruncated: false },
    })
    if (mode.value === 'exit2') return ran('', 2, 'moai: something broke')
    if (mode.value === 'garbage') return ran('<<< not json >>>')
    const joined = e.argv.join(' ')
    if (e.argv.includes('next') && afterPick.value !== undefined) q.value = afterPick.value
    if (joined === 'moai gtd list --json' && hold.wait !== undefined) {
      const answer = q.value
      const wait = hold.wait
      hold.wait = undefined
      await wait
      return ran(answer)
    }
    return joined.includes('gtd list')
      ? ran(q.value)
      : joined.includes('factory status')
        ? ran(FACTORY)
        : joined.includes('session list')
          ? ran('[]')
          : joined.includes('spec status')
            ? ran(SPEC_LIST)
            : ran('')
  })
  return { calls, clock: mock.clock(on), mode, queue: q, hold, afterPick }
}

/** Answers the engine's AskUserQuestion dialog (what `$.ui.ask` raises) with a label, or dismisses it. */
export const answerAsk = (on: On, answer: string | undefined) =>
  on('tool.call', { tool: 'AskUserQuestion' }, (_$, e) =>
    answer === undefined
      ? { deny: 'dismissed' }
      : { result: { questions: e.questions, answers: { [e.questions[0]?.question ?? '']: answer } } },
  )

export const picks = (calls: string[][]): string[][] => calls.filter(argv => argv.includes('next'))

/**
 * The file nouns beneath the plugin for the SPEC tab: the session root, `stat` (with the real path a
 * resolve asks for) and `read`. `files` maps an asked-for path to where it really lands and its text.
 */
export const stubSpecFiles = (on: On, root: string, files: Record<string, { realPath: string; text: string }>) => {
  const reads: string[] = []
  const stats: Record<string, { kind: 'file' | 'dir'; realPath: string }> = {
    [`${root}/.moai/specs`]: { kind: 'dir', realPath: `/real${root}/.moai/specs` },
  }
  for (const [path, f] of Object.entries(files)) stats[path] = { kind: 'file', realPath: f.realPath }
  on('session.root', () => ({ value: root }))
  on('fs.stat', (_$, e) => {
    const found = stats[e.path]
    if (found === undefined) throw new Error(`ENOENT ${e.path}`)
    return { value: { ...found, size: 1, mtimeMs: 0, isLink: false } }
  })
  on('fs.read', (_$, e) => {
    reads.push(e.path)
    const text = Object.values(files).find(f => f.realPath === e.path)?.text
    if (text === undefined) throw new Error(`ENOENT ${e.path}`)
    return { value: text }
  })
  return reads
}
