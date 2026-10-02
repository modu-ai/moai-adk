// moai-status data layer: `$`-free. The fixed argv table (REQ-MSM-002), the
// measure classifier and the strip/suffix line builders (M2), the toast line
// builder (M3). Pure: bun-tested under tests/pure/. argv lists only, never
// shell strings; the table is the ONLY argv built anywhere.
import type { MoaiStatusBand, MoaiStatusFigure, MoaiStatusMeasureInput, MoaiStatusUsage } from '../types'

/** What the engine's process.run resolves with (the laid typings' ProcessRunResult). */
export type RunResult = {
  exitCode: number
  stdout: string
  stderr: string
  isStdoutTruncated: boolean
  isStderrTruncated: boolean
}

/** The run function the hooks module hands the helpers; the single process.run site wraps it. */
export type Run = (argv: readonly string[]) => Promise<RunResult>

// plan §G parameters. HEALTH_POLL_MS above the floor is Q3's provisional value;
// the tests assert only the floor (REQ-MSM-007).
export const HEALTH_POLL_MIN_MS = 15_000
export const HEALTH_POLL_MS = 60_000
export const CMD_TIMEOUT_MS = 20_000
export const TOAST_EXCERPT_CP = 80
export const STATUS_LINE_MAX = 200

// The fixed argv table (plan §B.3) — the only argv built anywhere, pinned in
// double quotes for the AC-MSM-002(ii) structural check. The check names are
// the in-tree check identifiers (internal/cli/doctor.go:210,
// internal/cli/doctor_mcp_version.go).
export const ARGV = {
  binary: ["moai", "doctor", "--check", "Binary Freshness"],
  mcp: ["moai", "doctor", "--check", "MCP Server Version"],
  memory: ["moai", "memory", "doctor", "--json"],
} as const

// The strip thresholds mirror the in-repo gates' DEFAULT configuration (M-7,
// spec.md D-3): the t1442 context band and the t1347 quota-gate holds. The
// gates are runtime-configurable; the mod shows its own frozen values (G-12).
export const DEFAULT_BAND: MoaiStatusBand = {
  softLargePct: 50,
  softStandardPct: 90,
  largeWindowCutoff: 500_000,
  autoCompactPct: 85,
  hardMarginPct: 10,
  hardCapPct: 95,
  fiveHourHoldPct: 90,
  sevenDayHoldPct: 95,
}

// soft = 50 when the window is at least 500,000 tokens, else 90 (renderer.go).
export const softPct = (window: number | undefined, band: MoaiStatusBand): number =>
  window !== undefined && window >= band.largeWindowCutoff ? band.softLargePct : band.softStandardPct

// hard = min(hardCap, autoCompact + margin), clamped up to soft.
export const hardPct = (window: number | undefined, band: MoaiStatusBand): number =>
  Math.max(softPct(window, band), Math.min(band.hardCapPct, band.autoCompactPct + band.hardMarginPct))

// The pure classifier of plan §B.4: a function of the pushed figure and the
// explicit band. Absent figures are no reading, never zero (REQ-MSM-003);
// only the two gate kinds warn, at their holds; the output carries warn and
// critical figures only, so the render hooks draw from state without
// re-classifying.
export const classifyMeasure = (input: MoaiStatusMeasureInput, band: MoaiStatusBand): MoaiStatusUsage => {
  const figures: MoaiStatusFigure[] = []
  const percent = input.context?.percent
  const window = input.context?.window
  if (typeof percent === 'number' && typeof window === 'number' && window > 0) {
    const soft = softPct(window, band)
    const hard = hardPct(window, band)
    if (percent >= hard)
      figures.push({ kind: 'context', level: 'critical', percent, softPct: soft, hardPct: hard })
    else if (percent >= soft)
      figures.push({ kind: 'context', level: 'warn', percent, softPct: soft, hardPct: hard })
  }
  for (const entry of input.rateLimits ?? []) {
    if (entry.kind === 'five_hour' && entry.percentUsed >= band.fiveHourHoldPct)
      figures.push({ kind: 'rate', window: 'five_hour', percent: entry.percentUsed, holdPct: band.fiveHourHoldPct })
    else if (entry.kind === 'seven_day' && entry.percentUsed >= band.sevenDayHoldPct)
      figures.push({ kind: 'rate', window: 'seven_day', percent: entry.percentUsed, holdPct: band.sevenDayHoldPct })
  }
  return { figures }
}

// State is written only when the classification moved (REQ-MSM-003), so a
// measure burst cannot storm the subscribed renders.
export const sameUsage = (a: MoaiStatusUsage | undefined, b: MoaiStatusUsage): boolean =>
  a !== undefined && JSON.stringify(a) === JSON.stringify(b)

// The strip line: one line naming each warned figure and its threshold
// (REQ-MSM-004). Empty when nothing is warned — the hook passes then.
export const stripLine = (usage: MoaiStatusUsage | undefined): string => {
  const parts: string[] = []
  for (const figure of usage?.figures ?? []) {
    if (figure.kind === 'context')
      parts.push(
        figure.level === 'critical'
          ? `ctx ${figure.percent}% (critical at ${figure.hardPct})`
          : `ctx ${figure.percent}% (warn at ${figure.softPct})`,
      )
    else if (figure.window === 'five_hour') parts.push(`5h quota ${figure.percent}% (hold ${figure.holdPct})`)
    else parts.push(`7d quota ${figure.percent}% (hold ${figure.holdPct})`)
  }
  return parts.length === 0 ? '' : `moai-status: ${parts.join(' · ')}`
}

// The spinner marker: the shortest form of the same classification — context
// first, else the first warned rate window; '' when there is nothing to say
// (REQ-MSM-005). The hook appends it to the incoming suffix verbatim.
export const suffixMarker = (usage: MoaiStatusUsage | undefined): string => {
  for (const figure of usage?.figures ?? []) {
    if (figure.kind === 'context') return ` · ctx ${figure.percent}%`
  }
  for (const figure of usage?.figures ?? []) {
    if (figure.kind === 'rate')
      return figure.window === 'five_hour' ? ` · 5h ${figure.percent}%` : ` · 7d ${figure.percent}%`
  }
  return ''
}

// The toast line (REQ-MSM-006): the origin kind, then the first kept code
// points of the text's first line, control characters dropped — one line,
// always.
export const toastLine = (originKind: string, text: string): string => {
  const first = text.split('\n', 1)[0] ?? ''
  let excerpt = ''
  let kept = 0
  for (const ch of first) {
    const cp = ch.codePointAt(0) ?? 0
    if (cp < 0x20 || cp === 0x7f) continue
    if (kept >= TOAST_EXCERPT_CP) break
    excerpt += ch
    kept++
  }
  return `${originKind}: ${excerpt}`
}
