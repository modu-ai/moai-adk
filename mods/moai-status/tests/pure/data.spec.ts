// Pure tests for hooks/data.ts, run with `bun test` (developer-local evidence,
// spec.md G-11). Named *.spec.ts so the engine runner's *.test.ts glob skips them.
import { expect, test } from 'bun:test'
import { ARGV, DEFAULT_BAND, classifyMeasure, sameUsage, stripLine, suffixMarker, toastLine } from '../../hooks/data'
import type { MoaiStatusBand, MoaiStatusMeasureInput, MoaiStatusUsage } from '../../types'

test('argv: table carries the three diagnostic commands only', () => {
  expect(Object.keys(ARGV).sort()).toEqual(['binary', 'mcp', 'memory'])
  expect([...ARGV.binary]).toEqual(['moai', 'doctor', '--check', 'Binary Freshness'])
  expect([...ARGV.mcp]).toEqual(['moai', 'doctor', '--check', 'MCP Server Version'])
  expect([...ARGV.memory]).toEqual(['moai', 'memory', 'doctor', '--json'])
})

const measure = (over: Partial<MoaiStatusMeasureInput>): MoaiStatusMeasureInput => ({ ...over })

// ---- classify (AC-MSM-003) --------------------------------------------------------

test('classify: context soft follows the window band', () => {
  // 200K window: soft is the standard 90 — 89 is quiet, 91 warns.
  const small = measure({ context: { window: 200_000, percent: 89 }, rateLimits: [] })
  expect(classifyMeasure(small, DEFAULT_BAND).figures).toEqual([])
  const smallWarn = measure({ context: { window: 200_000, percent: 91 }, rateLimits: [] })
  expect(classifyMeasure(smallWarn, DEFAULT_BAND).figures).toEqual([
    { kind: 'context', level: 'warn', percent: 91, softPct: 90, hardPct: 95 },
  ])
  // 1M window: soft drops to 50 — 49 is quiet, 51 warns.
  const large = measure({ context: { window: 1_000_000, percent: 49 }, rateLimits: [] })
  expect(classifyMeasure(large, DEFAULT_BAND).figures).toEqual([])
  const largeWarn = measure({ context: { window: 1_000_000, percent: 51 }, rateLimits: [] })
  expect(classifyMeasure(largeWarn, DEFAULT_BAND).figures).toEqual([
    { kind: 'context', level: 'warn', percent: 51, softPct: 50, hardPct: 95 },
  ])
})

test('classify: context critical at the hard ceiling', () => {
  // Defaults: hard = min(95, 85 + 10) = 95; soft 90 < hard, no clamp.
  const atHard = measure({ context: { window: 200_000, percent: 95 }, rateLimits: [] })
  expect(classifyMeasure(atHard, DEFAULT_BAND).figures).toEqual([
    { kind: 'context', level: 'critical', percent: 95, softPct: 90, hardPct: 95 },
  ])
  const belowHard = measure({ context: { window: 200_000, percent: 92 }, rateLimits: [] })
  expect(classifyMeasure(belowHard, DEFAULT_BAND).figures).toEqual([
    { kind: 'context', level: 'warn', percent: 92, softPct: 90, hardPct: 95 },
  ])
  // The clamp branch is reachable only because the band is an input: an
  // aggressive autoCompact of 60 makes the naive hard 70, clamped up to soft 90 —
  // so 92 is CRITICAL at 90, not warn at the frozen 95.
  const clampBand: MoaiStatusBand = { ...DEFAULT_BAND, autoCompactPct: 60 }
  const clamped = measure({ context: { window: 200_000, percent: 92 }, rateLimits: [] })
  expect(classifyMeasure(clamped, clampBand).figures).toEqual([
    { kind: 'context', level: 'critical', percent: 92, softPct: 90, hardPct: 90 },
  ])
})

test('classify: absent percent is no reading, never zero', () => {
  const noPercent = measure({ context: { window: 200_000 }, rateLimits: [] })
  expect(classifyMeasure(noPercent, DEFAULT_BAND).figures).toEqual([])
  const noWindow = measure({ context: { percent: 91 }, rateLimits: [] })
  expect(classifyMeasure(noWindow, DEFAULT_BAND).figures).toEqual([])
  const zeroWindow = measure({ context: { window: 0, percent: 91 }, rateLimits: [] })
  expect(classifyMeasure(zeroWindow, DEFAULT_BAND).figures).toEqual([])
  const noContext = measure({ rateLimits: [] })
  expect(classifyMeasure(noContext, DEFAULT_BAND).figures).toEqual([])
})

test('classify: rate windows warn at their gate holds', () => {
  const justBelow = measure({
    context: { window: 200_000 },
    rateLimits: [
      { kind: 'five_hour', percentUsed: 89.9 },
      { kind: 'seven_day', percentUsed: 94.9 },
    ],
  })
  expect(classifyMeasure(justBelow, DEFAULT_BAND).figures).toEqual([])
  const atHold = measure({
    context: { window: 200_000 },
    rateLimits: [
      { kind: 'five_hour', percentUsed: 90 },
      { kind: 'seven_day', percentUsed: 95 },
    ],
  })
  expect(classifyMeasure(atHold, DEFAULT_BAND).figures).toEqual([
    { kind: 'rate', window: 'five_hour', percent: 90, holdPct: 90 },
    { kind: 'rate', window: 'seven_day', percent: 95, holdPct: 95 },
  ])
})

test('classify: unknown kinds never warn', () => {
  const other = measure({ context: { window: 200_000 }, rateLimits: [{ kind: 'spend_limit', percentUsed: 99 }] })
  expect(classifyMeasure(other, DEFAULT_BAND).figures).toEqual([])
})

test('classify: unchanged classification is not rewritten', () => {
  const warned: MoaiStatusUsage = { figures: [{ kind: 'context', level: 'warn', percent: 91, softPct: 90, hardPct: 95 }] }
  expect(sameUsage(warned, { figures: [{ kind: 'context', level: 'warn', percent: 91, softPct: 90, hardPct: 95 }] })).toBe(true)
  expect(sameUsage(undefined, warned)).toBe(false)
  expect(sameUsage({ figures: [] }, warned)).toBe(false)
  expect(sameUsage(warned, { figures: [] })).toBe(false)
})

// ---- strip line (AC-MSM-004 pure) --------------------------------------------------

test('strip-pure: warn line names the figure and the threshold', () => {
  const line = stripLine({ figures: [{ kind: 'context', level: 'warn', percent: 91, softPct: 90, hardPct: 95 }] })
  expect(line).toContain('moai-status:')
  expect(line).toContain('ctx 91% (warn at 90)')
  const quota = stripLine({ figures: [{ kind: 'rate', window: 'five_hour', percent: 92, holdPct: 90 }] })
  expect(quota).toContain('5h quota 92%')
  const critical = stripLine({ figures: [{ kind: 'context', level: 'critical', percent: 96, softPct: 90, hardPct: 95 }] })
  expect(critical).toContain('ctx 96% (critical at 95)')
  const both = stripLine({
    figures: [
      { kind: 'context', level: 'warn', percent: 91, softPct: 90, hardPct: 95 },
      { kind: 'rate', window: 'seven_day', percent: 96, holdPct: 95 },
    ],
  })
  expect(both).toContain('ctx 91% (warn at 90)')
  expect(both).toContain('7d quota 96%')
})

test('strip-pure: nothing above info yields an empty line', () => {
  expect(stripLine({ figures: [] })).toBe('')
  expect(stripLine(undefined)).toBe('')
})

// ---- suffix marker (AC-MSM-005 pure) -----------------------------------------------

test('suffix-pure: warn context yields ctx marker', () => {
  expect(suffixMarker({ figures: [{ kind: 'context', level: 'warn', percent: 91, softPct: 90, hardPct: 95 }] })).toBe(' · ctx 91%')
  expect(suffixMarker({ figures: [{ kind: 'rate', window: 'five_hour', percent: 92, holdPct: 90 }] })).toBe(' · 5h 92%')
  expect(suffixMarker({ figures: [{ kind: 'rate', window: 'seven_day', percent: 96, holdPct: 95 }] })).toBe(' · 7d 96%')
})

test('suffix-pure: nothing to say yields no marker', () => {
  expect(suffixMarker({ figures: [] })).toBe('')
  expect(suffixMarker(undefined)).toBe('')
})

// ---- toast line (AC-MSM-007 pure) ---------------------------------------------------

test('toast-pure: one line from origin and excerpt', () => {
  const line = toastLine('peer', 'first line\nsecond line')
  expect(line).toBe('peer: first line')
  expect(line.includes('\n')).toBe(false)
  const long = toastLine('task-notification', 'x'.repeat(200))
  expect(Array.from(long).length).toBe('task-notification: '.length + 80)
  expect(toastLine('peer', '')).toBe('peer: ')
})

test('toast-pure: control characters stripped', () => {
  expect(toastLine('peer', 'a\tb\rc\x07d')).toBe('peer: abcd')
})

