// moai-status hooks module. The only file that spells `$.…`: helper modules
// (data.ts, health.ts) stay `$`-free and receive functions or resolved element
// tables as arguments — the engine refuses `$` passed into a function imported
// from another file (sibling M-17). Strictly additive observer (spec.md §2):
// every handler passes its event through with next(e), every hook fails soft,
// and the only child-process calls are the fixed argv table's three diagnostics.
import type { EngineInterface, Register } from 'claude-code'
import {
  ARGV,
  CMD_TIMEOUT_MS,
  DEFAULT_BAND,
  HEALTH_POLL_MS,
  classifyMeasure,
  sameUsage,
  stripLine,
  suffixMarker,
  toastLine,
  type MoaiStatusHealth,
  type MoaiStatusSourceVerdict,
  type MoaiStatusUsage,
  type Run,
  type RunResult,
} from './data'
import { composeHealthLine, mergeGoodHealth, parseDoctorCheck, parseMemoryDoctor } from './health'

// Typed references: plugin and key are literals (the shape validate enforces, M-13).
const usageRef = { plugin: 'moai-status', key: 'usage' } as const
const healthRef = { plugin: 'moai-status', key: 'health' } as const
const noticeRef = { plugin: 'moai-status', key: 'notice' } as const

const EMPTY_USAGE: MoaiStatusUsage = { figures: [] }
const EMPTY_HEALTH: MoaiStatusHealth = { binary: { state: 'ok' }, mcp: { state: 'ok' }, memory: { state: 'ok' } }

const errorText = (err: unknown): string => (err instanceof Error ? err.message : String(err))

// Fail-soft (REQ-MSM-009): a guarded body that throws leaves a notice and no
// exception ever leaves a hook — the session continues unaffected. notice's
// only writer is this guard; the next clean cycle clears it (plan §B.1).
const setNotice = async ($: EngineInterface, text: string): Promise<void> => {
  const current = (await $.state.get(noticeRef)).value ?? ''
  if (current !== text) await $.state.set(noticeRef, text)
}

const soft = async ($: EngineInterface, body: () => Promise<void>): Promise<void> => {
  try {
    await body()
    await setNotice($, '')
  } catch (err) {
    try {
      await setNotice($, `moai-status: ${errorText(err)}`)
    } catch {
      // nothing left to try; the session continues unaffected
    }
  }
}

// The health timer (REQ-MSM-007): one $.clock.every timer at session.start,
// cancelled at session.end. Module variables hold only what a hot reload may
// lose at the cost of one timer restart (REQ-MSM-010); the single-flight gate
// drops a tick that finds one cycle running regardless of which instance owns it.
// The cancel handle is defensive over both shapes the engine shows: the laid
// typings name a bare function, the test clock hands back { cancel }.
let cancelTimer: unknown
let inFlight = false

const stopHandle = (handle: unknown): void => {
  if (typeof handle === 'function') (handle as () => void)()
  else if (handle !== null && typeof handle === 'object' && typeof (handle as { cancel?: unknown }).cancel === 'function')
    (handle as { cancel: () => void }).cancel()
}

const stopHealth = (): void => {
  stopHandle(cancelTimer)
  cancelTimer = undefined
}

// Row-first precedence lives in the parser; a rejected run (cannot start,
// timeout — the call rejects) is unknown: never healthy, never a session
// failure (REQ-MSM-009/-012).
const readDoctorSource = async (run: Run, argv: readonly string[], checkName: string): Promise<MoaiStatusSourceVerdict> => {
  try {
    const result = await run(argv)
    return parseDoctorCheck(result.stdout, checkName)
  } catch {
    return { state: 'unknown' }
  }
}

const readMemorySource = async (run: Run): Promise<MoaiStatusSourceVerdict> => {
  try {
    const result = await run([...ARGV.memory])
    return parseMemoryDoctor(result.stdout)
  } catch {
    return { state: 'unknown' }
  }
}

// One health cycle (REQ-MSM-007/-008): the three table argvs one at a time,
// at most one cycle in flight; the composed line pins via $.ui.status (or
// clears it), and state keeps the last good classification per source.
const runHealthCycle = async ($: EngineInterface, run: Run): Promise<void> => {
  if (inFlight) return
  inFlight = true
  try {
    const fresh = {
      binary: await readDoctorSource(run, [...ARGV.binary], 'Binary Freshness'),
      mcp: await readDoctorSource(run, [...ARGV.mcp], 'MCP Server Version'),
      memory: await readMemorySource(run),
    }
    const line = composeHealthLine(fresh)
    $.ui.status(line === '' ? undefined : `moai-status: ${line}`)
    const previous = (await $.state.get(healthRef)).value
    const merged = mergeGoodHealth(previous, fresh)
    if (JSON.stringify(previous) !== JSON.stringify(merged)) await $.state.set(healthRef, merged)
  } catch (err) {
    await setNotice($, `moai-status: ${errorText(err)}`)
  } finally {
    inFlight = false
  }
}

const startHealth = ($: EngineInterface, run: Run): void => {
  stopHealth()
  cancelTimer = $.clock.every(HEALTH_POLL_MS, () => {
    void runHealthCycle($, run)
  })
}

// The module's only child-process call site (REQ-MSM-002): every argv that
// reaches it comes from the fixed table in data.ts. `$` is the dispatch's own,
// threaded from the handler into this top-of-file helper (the engine's $-flow
// analysis requires that shape); `$` never crosses an import.
// @MX:ANCHOR: [AUTO] the module's only child-process call site - every diagnostic argv routes through this single $.process.run seam
// @MX:REASON: the fixed argv table in hooks/data.ts is the only argv source; a second call site would escape the fail-soft boundary audit (AC-MSM-002 counts exactly this one line) and the $-flow invariant ($ never crosses an import)
// @MX:SPEC: SPEC-MOAI-STATUS-MOD-001
const runDiag = ($: EngineInterface, argv: readonly string[]): Promise<RunResult> =>
  $.process.run(argv, { timeoutMs: CMD_TIMEOUT_MS })

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await soft($, async () => {
      // First-write initialization so every later read is defined. The refs are
      // spelled per call: validate lists only literal references (M-13).
      if ((await $.state.get(usageRef)).value === undefined) await $.state.set(usageRef, EMPTY_USAGE)
      if ((await $.state.get(healthRef)).value === undefined) await $.state.set(healthRef, EMPTY_HEALTH)
      if ((await $.state.get(noticeRef)).value === undefined) await $.state.set(noticeRef, '')
    })
    try {
      startHealth($, argv => runDiag($, argv))
    } catch {
      // no timer this session; the session continues unaffected
    }
    return next(e)
  })

  on('session.end', async ($, e, next) => {
    stopHealth()
    return next(e)
  })

  on('session.measure', async ($, e, next) => {
    // The engine pushes the figures (D-2); the mod classifies and holds them
    // in state, writing only when the classification moved (REQ-MSM-003).
    await soft($, async () => {
      const previous = (await $.state.get(usageRef)).value
      const classified = classifyMeasure(e, DEFAULT_BAND)
      if (!sameUsage(previous, classified)) await $.state.set(usageRef, classified)
    })
    return next(e)
  })

  on('session.receive', async ($, e, next) => {
    try {
      // The toast is attempted before the delivery passes; its failure never
      // holds, rewrites or consumes the delivery (REQ-MSM-006). The handler
      // returns next(e) on every path — the `{ consumed }` shape is never
      // produced.
      $.ui.toast(toastLine(e.origin.kind, e.text))
    } catch {
      // the delivery passes regardless
    }
    return next(e)
  })

  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    try {
      // A survey holds the band; the mod yields (REQ-MSM-004).
      if (e.props.hasSurvey) return next(e)
      const usage = (await $.state.get(usageRef)).value
      const line = stripLine(usage)
      if (line === '') return next(e)
      const T = $.ui.resolve(e)
      const upstream = await next(e)
      // Compose: the strip leads, the upstream tree still draws (REQ-MSM-004).
      // Text takes no key — the findable line sits in a keyed Box (sibling craft).
      return T.Box({
        key: 'moai-status-strip',
        flexDirection: 'column',
        children: [T.Box({ key: 'moai-status-strip-line', children: T.Text({ children: line }) }), upstream],
      })
    } catch {
      // A broken strip never blanks the band for later mods (plan §B.5).
      return next(e)
    }
  })

  on('ui.render', { component: 'Spinner' }, async ($, e, next) => {
    try {
      const usage = (await $.state.get(usageRef)).value
      const marker = suffixMarker(usage)
      if (marker === '') return next(e)
      // Only the suffix prop changes; word, message and mode pass untouched
      // (REQ-MSM-005). The incoming suffix is preserved, the marker appended.
      return next({ ...e, props: { ...e.props, suffix: e.props.suffix + marker } })
    } catch {
      return next(e)
    }
  })
}
