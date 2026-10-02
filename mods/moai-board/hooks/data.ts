// moai-board data layer: `$`-free. Everything here is a pure function or takes a
// `run` function from hooks/register.tsx, because the engine refuses `$` passed
// into an imported function (spec.md §4). Tested with bun (tests/pure/).
import type { MoaiBoardCard } from '../types'

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
/** Read-only commands. The pick command is built only by buildPickArgv. */
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

/** The one write-capable argv. Undefined when the id or the prefix fails its check. */
export const buildPickArgv = (id: string, prefix: string): readonly string[] | undefined =>
  isCardId(id) && isPrefixOk(prefix) ? ['moai', 'gtd', 'next', id, '--expect', prefix] : undefined

/** A card gets a pick button only when it is queued and its id and prefix pass the checks. */
export const canPick = (card: Pick<MoaiBoardCard, 'id' | 'state' | 'text'>): boolean =>
  card.state === 'queued' && isCardId(card.id) && isPrefixOk(pickPrefix(card.text))

/** Only the exact confirm label confirms: other text typed under "Other", a dismissal, or Cancel do not. */
export const isConfirmed = (answer: unknown): boolean => answer === CONFIRM_LABEL
