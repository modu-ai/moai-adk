// Pure tests for hooks/data.ts, run with `bun test` (developer-local evidence,
// spec.md G-13). Named *.spec.ts so the engine runner's *.test.ts glob skips them.
import { expect, test } from 'bun:test'
import {
  ARGV,
  buildPickArgv,
  canPick,
  clampInterval,
  createFlightGate,
  createQueueReaders,
  describeFailure,
  emptyFeed,
  executePick,
  feedOk,
  feedStatus,
  finishQueueFallback,
  groupLanes,
  isConfirmed,
  isStillQueued,
  laneCardText,
  parseQueueText,
  pickPrefix,
  planQueue,
  readLanes,
  readQueue,
  sessionText,
  summaryLine,
  type Outcome,
  type Run,
} from '../../hooks/data'

const NOW = Date.parse('2026-10-02T12:00:00Z')
const SEP = ' · '

const card = (over: Record<string, unknown> = {}) => ({
  id: 't1436',
  state: 'queued' as const,
  text: 'moai-board: read-only queue side panel for Claude Code',
  addedAt: '',
  specId: '',
  ...over,
})

const ok = (stdout: string, isTruncated = false): Outcome => ({ kind: 'ok', stdout, isTruncated })
const reply = (stdout: string, exitCode = 0, stderr = ''): Run => async () => ({
  exitCode,
  stdout,
  stderr,
  isStdoutTruncated: false,
})

// A queue payload shaped like spec.md M-2: items + findings + archived + runtime.
const queueJson = (counts = { picked: 17, queued: 44, hold: 2, dropped: 89 }, archivedRows = 0) => {
  const items: Record<string, unknown>[] = []
  let n = 1
  for (const [state, count] of Object.entries(counts))
    for (let i = 0; i < count; i++, n++)
      items.push({ id: `t${n}`, text: `card ${n} ${state} text`, state, added_at: '2026-09-01T00:00:00Z', spec_id: '', card_uuid: 'u' })
  const archived = Array.from({ length: archivedRows }, (_, i) => ({ id: `a${i}`, text: 'ARCHIVED-MARKER ' + 'z'.repeat(360) }))
  return JSON.stringify({
    project_uuid: 'p',
    version: 1,
    last_seq: 9,
    items,
    findings: [{ a: 'FINDING-MARKER' }],
    archived,
    runtime: { r: 'RUNTIME-MARKER' },
    unknown_key: 'UNKNOWN-MARKER',
  })
}

// ---- pick ----------------------------------------------------------------------

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

test('pick-pure: non-zero exit shows 400 chars of stderr and does not retry', async () => {
  let calls = 0
  const run: Run = async () => {
    calls++
    return { exitCode: 2, stdout: '', stderr: 'x'.repeat(600), isStdoutTruncated: false }
  }
  expect(await executePick(run, ['moai', 'gtd', 'next', 't1', '--expect', 'p'])).toEqual({
    kind: 'failed',
    text: 'x'.repeat(400),
  })
  expect(calls).toBe(1)
  const cannotStart: Run = async () => {
    calls++
    throw new Error('spawn moai ENOENT')
  }
  expect(await executePick(cannotStart, ['moai'])).toEqual({ kind: 'failed', text: 'spawn moai ENOENT' })
  expect(calls).toBe(2)
  expect(await executePick(reply(''), ['moai'])).toEqual({ kind: 'ok' })
})

test('pick-pure: exit 0 with card still queued reads unconfirmed', () => {
  expect(isStillQueued([card()], 't1436')).toBe(true)
  expect(isStillQueued([card({ state: 'picked' })], 't1436')).toBe(false)
  expect(isStillQueued([card({ id: 't9' })], 't1436')).toBe(false)
})

// ---- polling ---------------------------------------------------------------------

test('poll-pure: interval is clamped to the 15000 floor', () => {
  for (const requested of [0, -5, 1, 14_999, Number.NaN]) expect(clampInterval(requested)).toBe(15_000)
  expect(clampInterval(15_000)).toBe(15_000)
  expect(clampInterval(20_000)).toBe(20_000)
})

test('poll-pure: a second poll while one is in flight is refused', () => {
  const gate = createFlightGate()
  expect(gate.tryStart()).toBe(true)
  expect(gate.tryStart()).toBe(false)
  gate.finish()
  expect(gate.tryStart()).toBe(true)
})

// ---- queue reduction ---------------------------------------------------------------

test('queue: drops dropped, archived, runtime', () => {
  const raw = queueJson(undefined, 20)
  const { cards } = createQueueReaders().json(raw)
  expect(cards?.length).toBe(63)
  expect(cards?.some(c => (c.state as string) === 'dropped')).toBe(false)
  const kept = JSON.stringify(cards)
  for (const marker of ['ARCHIVED-MARKER', 'RUNTIME-MARKER', 'FINDING-MARKER', 'UNKNOWN-MARKER'])
    expect(kept.includes(marker)).toBe(false)
  expect(cards?.[0]).toEqual({
    id: 't1',
    state: 'picked',
    text: 'card 1 picked text',
    addedAt: '2026-09-01T00:00:00Z',
    specId: '',
  })
})

test('queue: summary counts', () => {
  const { cards } = createQueueReaders().json(queueJson())
  expect(summaryLine(cards ?? [])).toBe(`Picked 17${SEP}Queued 44${SEP}Held 2`)
  expect(summaryLine([])).toBe(`Picked 0${SEP}Queued 0${SEP}Held 0`)
  expect(summaryLine(cards ?? []).includes('In progress')).toBe(false)
})

test('queue: unchanged raw is not re-parsed', () => {
  let parses = 0
  const readers = createQueueReaders(s => {
    parses++
    return JSON.parse(s)
  })
  const raw = queueJson()
  const first = readers.json(raw)
  const second = readers.json(raw)
  expect(parses).toBe(1)
  expect(first.isChanged).toBe(true)
  expect(second.isChanged).toBe(false)
  expect(second.cards?.length).toBe(63)
  readers.json(queueJson({ picked: 1, queued: 1, hold: 0, dropped: 0 }))
  expect(parses).toBe(2)
})

test('queue: 2 MB payload reduces', () => {
  const raw = queueJson(undefined, 5000)
  expect(raw.length).toBeGreaterThanOrEqual(1_900_000)
  const { cards } = createQueueReaders().json(raw)
  expect(cards?.length).toBe(63)
  expect(summaryLine(cards ?? [])).toBe(`Picked 17${SEP}Queued 44${SEP}Held 2`)
})

// ---- truncation and invalid-JSON fallback ---------------------------------------------

const textList = () => {
  const rows: string[] = []
  const states = ['picked', 'queued', 'hold']
  for (let i = 1; i <= 63; i++) {
    const state = states[i % 3]
    const prefix = i % 7 === 0 ? 'by=lane-3\tlease=2026-10-01T16:36:02Z\t' : ''
    rows.push(`t${i}\t${state}\t${prefix}row ${i} text`)
    if (i % 5 === 0) rows.push(`\t↳ near-duplicate t${i + 1000} (jev, model signal p=0.84)`)
  }
  rows.push('89 dropped (hidden — see: moai todo list --dropped)')
  return rows.join('\n') + '\n'
}

const readers = () => createQueueReaders()

test('queue-fallback: truncated stdout', () => {
  const plan = planQueue(readers(), ok('{"items":[', true))
  expect(plan).toEqual({ step: 'fallback', argv: ['moai', 'gtd', 'list', '--limit', '0'], notice: expect.stringContaining('reduced list') })
})

test('queue-fallback: invalid json', () => {
  const plan = planQueue(readers(), ok('{ this is not json'))
  expect(plan.step).toBe('fallback')
  expect((plan as { argv: readonly string[] }).argv).toEqual(['moai', 'gtd', 'list', '--limit', '0'])
})

test('queue-fallback: no items array', () => {
  for (const stdout of ['{"findings":[]}', '{"items":"nope"}', '[]', 'null']) {
    const plan = planQueue(readers(), ok(stdout))
    expect(plan.step).toBe('fallback')
  }
})

test('queue-fallback: both fail is degraded', async () => {
  const r = readers()
  expect(planQueue(r, ok('{"items":[', true)).step).toBe('fallback')
  const failed: Outcome = { kind: 'exit', code: 1, stderr: 'boom' }
  expect(finishQueueFallback(r, failed)).toEqual({ step: 'failed', cause: 'The queue data could not be read.' })
  // the whole read: JSON truncated, text list fails too -> degraded feed keeps the last good data dimmed
  const prev = feedOk({ cards: [card()], isReduced: false }, NOW - 120_000)
  let calls = 0
  const run: Run = async argv => {
    calls++
    return argv.includes('--json')
      ? { exitCode: 0, stdout: '{"items":[', stderr: '', isStdoutTruncated: true }
      : { exitCode: 1, stdout: '', stderr: 'text list broke', isStdoutTruncated: false }
  }
  const { feed, isChanged } = await readQueue(run, readers(), prev, NOW)
  expect(calls).toBe(2)
  expect(isChanged).toBe(true)
  expect(feed.error).toBe('The queue data could not be read.')
  expect(feed.data).toEqual(prev.data)
})

test('queue-text: rows, continuation lines, footer', () => {
  const cards = parseQueueText(textList())
  expect(cards.length).toBe(63)
  expect(cards.map(c => c.id)).toEqual(Array.from({ length: 63 }, (_, i) => `t${i + 1}`))
  expect(cards[6]).toEqual({ id: 't7', state: 'queued', text: 'row 7 text', addedAt: '', specId: '' })
  expect(cards.some(c => c.text.includes('near-duplicate') || c.text.includes('dropped'))).toBe(false)
  const text = createQueueReaders().text(textList())
  expect(text.cards?.length).toBe(63)
})

// ---- Lanes tab -----------------------------------------------------------------------------

const factoryJson = () =>
  JSON.stringify({
    run: 'tm3yoq',
    cards: [
      { card_id: 't587', state: 'completed', legacy: true, stage: '', owner: 'lead', spec_id: '', lease_expired: false },
      { card_id: 't600', state: 'completed', legacy: false, stage: '', owner: 'lane-1', spec_id: '', lease_expired: false },
      { card_id: 't1344', state: 'assigned', legacy: false, stage: '', owner: 'lane-1', spec_id: '', lease_expired: false },
      { card_id: 't1345', state: 'assigned', legacy: false, stage: '', owner: 'worker-64', spec_id: '', lease_expired: false },
      { card_id: 't1399', state: 'plan', legacy: false, stage: 'plan', owner: 'lane-3', spec_id: 'SPEC-X-001', lease_expired: true },
      { card_id: 't1400', state: 'plan', legacy: false, stage: 'plan', owner: 'lane-3', spec_id: '', lease_expired: false },
    ],
    unavailable: [],
  })

const sessionJson = () => {
  const mk = (id: string, minutesAgo: number, extra: Record<string, unknown> = {}) => ({
    session_id: `${id}-aaaa-bbbb`,
    spec_id: '(none)',
    phase: '(none)',
    started_at: '2026-09-26T07:04:06Z',
    last_heartbeat: new Date(NOW - minutesAgo * 60_000).toISOString(),
    pid: 40177,
    host: 'h',
    cwd: '/x',
    ...extra,
  })
  const sessions = [mk('old00000', 60 * 30), mk('mid00000', 90)]
  for (let i = 0; i < 25; i++) sessions.push(mk(`s${String(i).padStart(7, '0')}`, 5 + i))
  return JSON.stringify(sessions)
}

test('lanes: groups by owner, drops legacy and completed', async () => {
  const feed = await readLanes(async argv => ({ exitCode: 0, stdout: argv.includes('factory') ? factoryJson() : '[]', stderr: '', isStdoutTruncated: false }), emptyFeed(), NOW)
  const groups = groupLanes(feed.data?.cards ?? [])
  expect(groups.map(g => g.owner)).toEqual(['lane-1', 'worker-64', 'lane-3'])
  expect(groups.map(g => g.cards.map(c => c.id))).toEqual([['t1344'], ['t1345'], ['t1399', 't1400']])
})

test('lanes: lease expired marker', async () => {
  const feed = await readLanes(async argv => ({ exitCode: 0, stdout: argv.includes('factory') ? factoryJson() : '[]', stderr: '', isStdoutTruncated: false }), emptyFeed(), NOW)
  const cards = feed.data?.cards ?? []
  const expired = cards.find(c => c.id === 't1399')
  const fine = cards.find(c => c.id === 't1400')
  expect(laneCardText(expired!)).toContain('lease expired')
  expect(laneCardText(expired!)).toContain('SPEC-X-001')
  expect(laneCardText(fine!)).not.toContain('lease expired')
})

test('lanes: sessions window, order, cap', async () => {
  const feed = await readLanes(async argv => ({ exitCode: 0, stdout: argv.includes('session') ? sessionJson() : '{"cards":[]}', stderr: '', isStdoutTruncated: false }), emptyFeed(), NOW)
  const sessions = feed.data?.sessions ?? []
  expect(sessions.length).toBe(20)
  expect(sessions.some(s => s.sessionId.startsWith('old00000'))).toBe(false)
  const beats = sessions.map(s => s.heartbeatMs)
  expect(beats).toEqual([...beats].sort((a, b) => b - a))
  expect(sessionText(sessions[0]!, NOW)).toContain('heartbeat 5m ago')
})

test('lanes: never says alive or dead', async () => {
  const feed = await readLanes(async argv => ({ exitCode: 0, stdout: argv.includes('factory') ? factoryJson() : sessionJson(), stderr: '', isStdoutTruncated: false }), emptyFeed(), NOW)
  const lines = [
    ...(feed.data?.cards ?? []).map(laneCardText),
    ...(feed.data?.sessions ?? []).map(s => sessionText(s, NOW)),
  ]
  expect(lines.length).toBeGreaterThan(5)
  for (const line of lines) expect(/alive|dead|\blive\b/i.test(line)).toBe(false)
})

// ---- fail-soft --------------------------------------------------------------------------------

const SRC = { cmd: 'moai gtd list', what: 'the queue' }

test('failsoft-pure: moai not found', async () => {
  const rejecting: Run = async () => {
    throw new Error('spawn moai ENOENT')
  }
  const feed = await readQueue(rejecting, readers(), feedOk({ cards: [], isReduced: false }, NOW), NOW)
  expect(feed.feed.error).toBe('moai was not found on PATH, so the board cannot read the queue.')
  expect(describeFailure(SRC, { kind: 'cannot-start', message: 'x' })).not.toContain('\n')
})

test('failsoft-pure: non-zero exit', () => {
  const cause = describeFailure(SRC, { kind: 'exit', code: 2, stderr: 'line one\nline two ' + 'y'.repeat(300) })
  expect(cause.startsWith('`moai gtd list` failed (exit 2): line one line two')).toBe(true)
  expect(cause).not.toContain('\n')
  expect(cause.length).toBeLessThan(260)
})

test('failsoft-pure: timeout', async () => {
  const slow: Run = async () => {
    throw new Error('process timed out after 20000 ms')
  }
  const { feed } = await readQueue(slow, readers(), emptyFeed(), NOW)
  expect(feed.error).toBe('`moai gtd list` took longer than 20 s and was stopped.')
})

test('failsoft-pure: unparseable output', async () => {
  const prev = feedOk({ cards: [], isReduced: false }, NOW)
  const good = { cards: [card({ id: 'l1' })], sessions: [] } as never
  const prevLanes = feedOk(good, NOW - 3 * 60_000)
  const feed = await readLanes(reply('<<< not json >>>'), prevLanes, NOW)
  expect(feed.error).toBe('`moai factory status` returned output the board could not read.')
  expect(feed.data).toEqual(prevLanes.data)
  const status = feedStatus(feed, NOW)
  expect(status.isDim).toBe(true)
  expect(status.age).toBe('last updated 3m ago')
  expect(feedStatus(prev, NOW).isDim).toBe(false)
})
