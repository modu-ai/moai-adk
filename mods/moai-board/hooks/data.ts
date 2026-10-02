// moai-board data layer: `$`-free. Everything here is a pure function or takes a
// `run` function from hooks/register.tsx, because the engine refuses `$` passed
// into an imported function (spec.md §4). Tested with bun (tests/pure/).
import type {
  MoaiBoardCard,
  MoaiBoardCardState,
  MoaiBoardFeed,
  MoaiBoardLaneCard,
  MoaiBoardLanes,
  MoaiBoardQueue,
  MoaiBoardSession,
} from '../types'

// ---- constants (plan.md §G) ------------------------------------------------
export const POLL_MIN_MS = 15_000
export const POLL_INTERVAL_MS = 15_000
export const CMD_TIMEOUT_MS = 20_000
export const STDERR_SHOW = 400
export const EXPECT_PREFIX_LEN = 40
export const CONFIRM_LABEL = 'Pick'
export const CANCEL_LABEL = 'Cancel'

// ---- the process boundary --------------------------------------------------
export type RunResult = {
  exitCode: number
  stdout: string
  stderr: string
  isStdoutTruncated: boolean
}
/** What register.tsx hands over in place of `$`: one argv list in, one result out (rejects when it cannot run). */
export type Run = (argv: readonly string[]) => Promise<RunResult>

// ---- the fixed argv table (REQ-MBM-002, REQ-MBM-013) ------------------------
/** Read-only commands. The one write-capable argv is built by the pick builder below, nowhere else. */
export const ARGV = {
  queueJson: ['moai', 'gtd', 'list', '--json'],
  queueText: ['moai', 'gtd', 'list', '--limit', '0'],
  lanes: ['moai', 'factory', 'status', '--json'],
  sessions: ['moai', 'session', 'list', '--json'],
  specList: ['moai', 'spec', 'status', '--list'],
} as const satisfies Record<string, readonly string[]>

// ---- pick (plan.md §B.3, §B.8) ----------------------------------------------
const ID_RE = /^[A-Za-z0-9][A-Za-z0-9_-]{0,31}$/

export const isCardId = (id: string): boolean => ID_RE.test(id)

/** The first 40 code points of the card text as polled. */
export const pickPrefix = (text: string): string => Array.from(text).slice(0, EXPECT_PREFIX_LEN).join('')

const isPrefixOk = (prefix: string): boolean => prefix.length > 0 && !prefix.startsWith('-')

// @MX:NOTE: [AUTO] the one write-capable argv in the mod. Only the confirmed press handler in
// register.tsx may call it (SPEC-MOAI-BOARD-MOD-001 REQ-MBM-002, REQ-MBM-003).
/** The one write-capable argv. Undefined when the id or the prefix fails its check. */
export const buildPickArgv = (id: string, prefix: string): readonly string[] | undefined =>
  isCardId(id) && isPrefixOk(prefix) ? ['moai', 'gtd', 'next', id, '--expect', prefix] : undefined

/** A card gets a pick button only when it is queued and its id and prefix pass the checks. */
export const canPick = (card: Pick<MoaiBoardCard, 'id' | 'state' | 'text'>): boolean =>
  card.state === 'queued' && isCardId(card.id) && isPrefixOk(pickPrefix(card.text))

/** Only the exact confirm label confirms: other text typed under "Other", a dismissal, or Cancel do not. */
export const isConfirmed = (answer: unknown): boolean => answer === CONFIRM_LABEL

// ---- small text helpers ------------------------------------------------------
const SEP = ' · '

/** Cut to n code points (never inside a surrogate pair). */
export const cutChars = (text: string, n: number): string => {
  const chars = Array.from(text)
  return chars.length <= n ? text : chars.slice(0, n).join('')
}

const oneLine = (text: string, n: number): string => cutChars(text.replace(/\s+/g, ' ').trim(), n)

export const firstLine = (text: string): string => text.split('\n', 1)[0] ?? ''

const isRecord = (v: unknown): v is Record<string, unknown> => typeof v === 'object' && v !== null && !Array.isArray(v)

const str = (v: unknown): string => (typeof v === 'string' ? v : '')

const safeJson = (parse: (s: string) => unknown, raw: string): unknown => {
  try {
    return parse(raw)
  } catch {
    return undefined
  }
}

/** "5m ago" style age; the board states ages only, never a verdict about a session. */
export const ageText = (ms: number): string => {
  if (!(ms >= 60_000)) return 'just now'
  const minutes = Math.floor(ms / 60_000)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  return hours < 48 ? `${hours}h ago` : `${Math.floor(hours / 24)}d ago`
}

// ---- run outcomes and fail-soft causes (REQ-MBM-010) -----------------------------
export type Outcome =
  | { kind: 'ok'; stdout: string; isTruncated: boolean }
  | { kind: 'exit'; code: number; stderr: string }
  | { kind: 'timeout' }
  | { kind: 'cannot-start'; message: string }

export type Failure = Exclude<Outcome, { kind: 'ok' }> | { kind: 'unparseable' }

const TIMEOUT_RE = /time(d)?[ -]?out|still running/i

/** Never throws: a rejected run becomes timeout or cannot-start, any exit code a value. */
export const runOutcome = async (run: Run, argv: readonly string[]): Promise<Outcome> => {
  try {
    const r = await run(argv)
    return r.exitCode === 0
      ? { kind: 'ok', stdout: r.stdout, isTruncated: r.isStdoutTruncated }
      : { kind: 'exit', code: r.exitCode, stderr: r.stderr }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    return TIMEOUT_RE.test(message) ? { kind: 'timeout' } : { kind: 'cannot-start', message }
  }
}

export type Source = { cmd: string; what: string }
export const SRC_QUEUE: Source = { cmd: 'moai gtd list', what: 'the queue' }
export const SRC_LANES: Source = { cmd: 'moai factory status', what: 'the lanes' }
export const SRC_SESSIONS: Source = { cmd: 'moai session list', what: 'the sessions' }
export const SRC_SPECS: Source = { cmd: 'moai spec status', what: 'the SPEC list' }

/** One plain-language line for the affected tab. */
export const describeFailure = (src: Source, failure: Failure): string => {
  switch (failure.kind) {
    case 'cannot-start':
      return `moai was not found on PATH, so the board cannot read ${src.what}.`
    case 'exit': {
      const detail = oneLine(failure.stderr, 200)
      return `\`${src.cmd}\` failed (exit ${failure.code})${detail === '' ? '.' : `: ${detail}`}`
    }
    case 'timeout':
      return `\`${src.cmd}\` took longer than ${CMD_TIMEOUT_MS / 1000} s and was stopped.`
    case 'unparseable':
      return `\`${src.cmd}\` returned output the board could not read.`
  }
}

// ---- feeds: last good data + latest failure --------------------------------------
export const emptyFeed = <T>(): MoaiBoardFeed<T> => ({ data: undefined, at: 0, error: '' })
export const feedOk = <T>(data: T, now: number): MoaiBoardFeed<T> => ({ data, at: now, error: '' })
/** A failure keeps the last good data (dimmed by the view) and records the cause. */
export const feedFail = <T>(feed: MoaiBoardFeed<T>, cause: string): MoaiBoardFeed<T> => ({ ...feed, error: cause })
export const sameFeed = <T>(a: MoaiBoardFeed<T>, b: MoaiBoardFeed<T>): boolean =>
  a.error === b.error && JSON.stringify(a.data) === JSON.stringify(b.data)

/** Data is dimmed while its latest read failed; `at` is the last time the data changed. */
export const feedStatus = <T>(feed: MoaiBoardFeed<T>, now: number): { isDim: boolean; age: string; cause: string } => ({
  isDim: feed.data !== undefined && feed.error !== '',
  age: feed.data === undefined ? '' : `last updated ${ageText(now - feed.at)}`,
  cause: feed.error,
})

// ---- polling policy (REQ-MBM-006) ---------------------------------------------------
/** Never below the 15,000 ms floor; NaN and negatives read as the floor. */
export const clampInterval = (ms: number): number => (Number.isFinite(ms) ? Math.max(POLL_MIN_MS, ms) : POLL_MIN_MS)

/** At most one poll in flight. */
export const createFlightGate = () => {
  let isBusy = false
  return {
    tryStart: (): boolean => {
      if (isBusy) return false
      isBusy = true
      return true
    },
    finish: (): void => {
      isBusy = false
    },
  }
}

// ---- queue (REQ-MBM-007, REQ-MBM-008) --------------------------------------------------
const STATES: readonly MoaiBoardCardState[] = ['picked', 'queued', 'hold']

const asCardState = (v: unknown): MoaiBoardCardState | undefined => STATES.find(s => s === v)

/** Non-dropped items only; archived, runtime, findings and unknown keys never leave this function. */
export const reduceQueue = (parsed: unknown): MoaiBoardCard[] | undefined => {
  if (!isRecord(parsed) || !Array.isArray(parsed['items'])) return undefined
  const cards: MoaiBoardCard[] = []
  for (const item of parsed['items']) {
    if (!isRecord(item)) continue
    const state = asCardState(item['state'])
    const id = str(item['id'])
    if (state === undefined || id === '') continue
    cards.push({ id, state, text: str(item['text']), addedAt: str(item['added_at']), specId: str(item['spec_id']) })
  }
  return cards
}

const QUEUE_ROW_RE = /^([A-Za-z0-9][A-Za-z0-9_-]*)\t(picked|queued|hold)\t(.*)$/
const QUEUE_META_RE = /^(?:by|lease)=[^\t]*\t/

/** The `moai gtd list` text form: `id<TAB>state<TAB>[by=… <TAB>lease=… <TAB>]text`; continuation lines and the footer are not rows. */
export const parseQueueText = (text: string): MoaiBoardCard[] => {
  const cards: MoaiBoardCard[] = []
  for (const line of text.split('\n')) {
    const m = QUEUE_ROW_RE.exec(line)
    const state = asCardState(m?.[2])
    if (m === null || state === undefined) continue
    let rest = m[3] ?? ''
    for (let g = QUEUE_META_RE.exec(rest); g !== null; g = QUEUE_META_RE.exec(rest)) rest = rest.slice(g[0].length)
    cards.push({ id: m[1] ?? '', state, text: rest, addedAt: '', specId: '' })
  }
  return cards
}

type Reader<C> = ((raw: string) => { isChanged: boolean; cards: C }) & { forget: () => void }

/** A raw payload equal to the previous one is neither re-parsed nor re-reduced. */
const createReader = <C>(derive: (raw: string) => C): Reader<C> => {
  let lastRaw: string | undefined
  let last: C | undefined
  const read = (raw: string) => {
    if (raw === lastRaw) return { isChanged: false, cards: last as C }
    lastRaw = raw
    last = derive(raw)
    return { isChanged: true, cards: last }
  }
  return Object.assign(read, {
    forget: () => {
      lastRaw = undefined
      last = undefined
    },
  })
}

export type QueueReaders = {
  json: Reader<MoaiBoardCard[] | undefined>
  text: Reader<MoaiBoardCard[]>
}

export const createQueueReaders = (parse: (s: string) => unknown = JSON.parse): QueueReaders => ({
  json: createReader(raw => reduceQueue(safeJson(parse, raw))),
  text: createReader(parseQueueText),
})

export const REDUCED_NOTICE =
  'Showing a reduced list: the full queue data could not be read in one piece, so detail columns are limited.'
export const DEGRADED_CAUSE = 'The queue data could not be read.'

export type QueuePlan =
  | { step: 'unchanged' }
  | { step: 'done'; cards: MoaiBoardCard[]; isReduced: boolean }
  | { step: 'fallback'; argv: readonly string[]; notice: string }
  | { step: 'failed'; cause: string }

/** Decide what the JSON read means: use it, read the text list instead, or fail. */
export const planQueue = (readers: QueueReaders, out: Outcome): QueuePlan => {
  if (out.kind !== 'ok') return { step: 'failed', cause: describeFailure(SRC_QUEUE, out) }
  if (out.isTruncated) return { step: 'fallback', argv: ARGV.queueText, notice: REDUCED_NOTICE }
  const r = readers.json(out.stdout)
  if (r.cards === undefined) return { step: 'fallback', argv: ARGV.queueText, notice: REDUCED_NOTICE }
  return r.isChanged ? { step: 'done', cards: r.cards, isReduced: false } : { step: 'unchanged' }
}

/** The text-list read after a fallback; when it fails too the tab is degraded. */
export const finishQueueFallback = (readers: QueueReaders, out: Outcome): QueuePlan => {
  if (out.kind !== 'ok' || out.isTruncated) return { step: 'failed', cause: DEGRADED_CAUSE }
  const r = readers.text(out.stdout)
  return r.isChanged ? { step: 'done', cards: r.cards, isReduced: true } : { step: 'unchanged' }
}

type QueueRead = { feed: MoaiBoardFeed<MoaiBoardQueue>; isChanged: boolean }

const applyQueuePlan = (prev: MoaiBoardFeed<MoaiBoardQueue>, plan: QueuePlan, now: number): QueueRead => {
  switch (plan.step) {
    case 'done':
      return { feed: feedOk({ cards: plan.cards, isReduced: plan.isReduced }, now), isChanged: true }
    case 'failed':
      return { feed: feedFail(prev, plan.cause), isChanged: prev.error !== plan.cause }
    case 'unchanged':
      return prev.error === '' ? { feed: prev, isChanged: false } : { feed: { ...prev, error: '' }, isChanged: true }
    case 'fallback':
      return { feed: feedFail(prev, DEGRADED_CAUSE), isChanged: prev.error !== DEGRADED_CAUSE }
  }
}

/** One queue poll: JSON first, the text list only when the JSON cannot be used. */
export const readQueue = async (
  run: Run,
  readers: QueueReaders,
  prev: MoaiBoardFeed<MoaiBoardQueue>,
  now: number,
): Promise<QueueRead> => {
  let plan = planQueue(readers, await runOutcome(run, ARGV.queueJson))
  if (plan.step === 'fallback') {
    readers.json.forget()
    plan = finishQueueFallback(readers, await runOutcome(run, plan.argv))
  } else if (plan.step === 'done') {
    readers.text.forget()
  }
  return applyQueuePlan(prev, plan, now)
}

/** `Picked N · Queued N · Held N`; Picked is an operator promotion, not evidence that anyone works on the card. */
export const summaryLine = (cards: readonly MoaiBoardCard[]): string => {
  const n = (state: MoaiBoardCardState) => cards.filter(c => c.state === state).length
  return `Picked ${n('picked')}${SEP}Queued ${n('queued')}${SEP}Held ${n('hold')}`
}

// ---- Lanes tab (REQ-MBM-009) --------------------------------------------------------------
export const HEARTBEAT_WINDOW_MS = 24 * 3_600_000
export const SESSION_CAP = 20

export const parseLanes = (stdout: string): MoaiBoardLaneCard[] | undefined => {
  const parsed = safeJson(JSON.parse, stdout)
  if (!isRecord(parsed) || !Array.isArray(parsed['cards'])) return undefined
  const cards: MoaiBoardLaneCard[] = []
  for (const c of parsed['cards']) {
    if (!isRecord(c) || c['legacy'] === true || c['state'] === 'completed') continue
    cards.push({
      id: str(c['card_id']),
      owner: str(c['owner']),
      state: str(c['state']),
      stage: str(c['stage']),
      specId: str(c['spec_id']),
      isLeaseExpired: c['lease_expired'] === true,
    })
  }
  return cards
}

/** Sessions with a heartbeat inside the window, newest first, capped. No verdict on any session. */
export const parseSessions = (stdout: string, now: number): MoaiBoardSession[] | undefined => {
  const parsed = safeJson(JSON.parse, stdout)
  if (!Array.isArray(parsed)) return undefined
  const sessions: MoaiBoardSession[] = []
  for (const s of parsed) {
    if (!isRecord(s)) continue
    const heartbeatMs = Date.parse(str(s['last_heartbeat']))
    if (Number.isNaN(heartbeatMs) || now - heartbeatMs > HEARTBEAT_WINDOW_MS) continue
    sessions.push({
      sessionId: str(s['session_id']),
      specId: str(s['spec_id']),
      phase: str(s['phase']),
      heartbeatMs,
    })
  }
  return sessions.sort((a, b) => b.heartbeatMs - a.heartbeatMs).slice(0, SESSION_CAP)
}

type Settled<T> = { value: T | undefined; cause: string }

const settle = <T>(src: Source, out: Outcome, parse: (stdout: string) => T | undefined): Settled<T> => {
  if (out.kind !== 'ok') return { value: undefined, cause: describeFailure(src, out) }
  const value = parse(out.stdout)
  return value === undefined
    ? { value: undefined, cause: describeFailure(src, { kind: 'unparseable' }) }
    : { value, cause: '' }
}

/** One Lanes poll: factory cards and sessions. Any failure keeps the previous data and names the first cause. */
export const readLanes = async (
  run: Run,
  prev: MoaiBoardFeed<MoaiBoardLanes>,
  now: number,
): Promise<MoaiBoardFeed<MoaiBoardLanes>> => {
  const cards = settle(SRC_LANES, await runOutcome(run, ARGV.lanes), parseLanes)
  const sessions = settle(SRC_SESSIONS, await runOutcome(run, ARGV.sessions), out => parseSessions(out, now))
  const cause = cards.cause || sessions.cause
  if (cause !== '' || cards.value === undefined || sessions.value === undefined) return feedFail(prev, cause)
  return feedOk({ cards: cards.value, sessions: sessions.value }, now)
}

export const groupLanes = (cards: readonly MoaiBoardLaneCard[]): { owner: string; cards: MoaiBoardLaneCard[] }[] => {
  const groups: { owner: string; cards: MoaiBoardLaneCard[] }[] = []
  for (const card of cards) {
    const owner = card.owner === '' ? '(no owner)' : card.owner
    const group = groups.find(g => g.owner === owner)
    if (group === undefined) groups.push({ owner, cards: [card] })
    else group.cards.push(card)
  }
  return groups
}

export const laneCardText = (c: MoaiBoardLaneCard): string =>
  `${c.id}  ${c.state}${c.stage !== '' && c.stage !== c.state ? ` / ${c.stage}` : ''}` +
  `${c.specId === '' ? '' : SEP + c.specId}${c.isLeaseExpired ? `${SEP}lease expired` : ''}`

const noneToEmpty = (v: string): string => (v === '(none)' ? '' : v)

export const sessionText = (s: MoaiBoardSession, now: number): string => {
  const detail = [noneToEmpty(s.phase), noneToEmpty(s.specId)].filter(v => v !== '').join(' ')
  return `${s.sessionId.slice(0, 8)}${detail === '' ? '' : `  ${detail}`}${SEP}heartbeat ${ageText(now - s.heartbeatMs)}`
}

// ---- pick outcome (REQ-MBM-005) ---------------------------------------------------------------
export type PickResult = { kind: 'ok' } | { kind: 'failed'; text: string }

/** Runs the pick argv exactly once, never retries; a failure reads as the first 400 characters of stderr (or the start error). */
export const executePick = async (run: Run, argv: readonly string[]): Promise<PickResult> => {
  const out = await runOutcome(run, argv)
  switch (out.kind) {
    case 'ok':
      return { kind: 'ok' }
    case 'exit':
      return {
        kind: 'failed',
        text: cutChars(out.stderr.trim() === '' ? `moai exited with code ${out.code}` : out.stderr, STDERR_SHOW),
      }
    case 'timeout':
      return { kind: 'failed', text: `The pick command took longer than ${CMD_TIMEOUT_MS / 1000} s and was stopped.` }
    case 'cannot-start':
      return { kind: 'failed', text: cutChars(out.message, STDERR_SHOW) }
  }
}

/** After a pick that exited 0: the card still reading `queued` means the pick is unconfirmed. */
export const isStillQueued = (cards: readonly MoaiBoardCard[], id: string): boolean =>
  cards.some(c => c.id === id && c.state === 'queued')
