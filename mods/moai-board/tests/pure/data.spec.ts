// Pure tests for hooks/data.ts, run with `bun test` (developer-local evidence,
// spec.md G-13). Named *.spec.ts so the engine runner's *.test.ts glob skips them.
import { expect, test } from 'bun:test'
import { ARGV, buildPickArgv, canPick, isConfirmed, pickPrefix } from '../../hooks/data'

const card = (over: Record<string, unknown> = {}) => ({
  id: 't1436',
  state: 'queued' as const,
  text: 'moai-board: read-only queue side panel for Claude Code',
  addedAt: '',
  specId: '',
  ...over,
})

test('pick-pure: argv is moai gtd next id expect prefix', () => {
  expect(buildPickArgv('t1436', 'moai-board: read-only')).toEqual([
    'moai', 'gtd', 'next', 't1436', '--expect', 'moai-board: read-only',
  ])
  // the read-only table never carries the write verb
  for (const argv of Object.values(ARGV)) expect(argv).not.toContain('next')
  // the prefix is the first 40 code points of the polled text
  expect(Array.from(pickPrefix('a'.repeat(100))).length).toBe(40)
})

test('pick-pure: id or prefix failing the check offers no pick', () => {
  for (const id of ['', '-x', 'a b', 'a;rm', 'x'.repeat(33), '../t1', 't1\n'])
    expect(buildPickArgv(id, 'ok')).toBeUndefined()
  expect(buildPickArgv('t1', '--force')).toBeUndefined()
  expect(buildPickArgv('t1', '')).toBeUndefined()
  expect(canPick(card({ id: 'bad id' }))).toBe(false)
  expect(canPick(card({ text: '-looks like a flag' }))).toBe(false)
  expect(canPick(card({ state: 'picked' }))).toBe(false)
  expect(canPick(card({ state: 'hold' }))).toBe(false)
  expect(canPick(card())).toBe(true)
})

test('pick-pure: only the confirm label confirms', () => {
  expect(isConfirmed('Pick')).toBe(true)
  for (const answer of ['Cancel', 'pick', ' Pick', 'Pick ', '', 'Pick,Cancel', undefined])
    expect(isConfirmed(answer)).toBe(false)
})
