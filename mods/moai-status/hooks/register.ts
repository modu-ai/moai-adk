// moai-status hooks module. The only file that spells `$.…`: helper modules
// (data.ts, health.ts) stay `$`-free and receive functions or resolved element
// tables as arguments — the engine refuses `$` passed into a function imported
// from another file (sibling M-17). Strictly additive observer (spec.md §2):
// every handler passes its event through with next(e), every hook fails soft,
// and the only child-process calls are the fixed argv table's three diagnostics.
import type { EngineInterface, Register } from 'claude-code'
import { CMD_TIMEOUT_MS, DEFAULT_BAND, classifyMeasure, sameUsage, stripLine, suffixMarker, toastLine, type MoaiStatusHealth, type MoaiStatusUsage, type RunResult } from './data'

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

export const register: Register = on => {
  // The module's only child-process call site (REQ-MSM-002): every argv that
  // reaches it comes from the fixed table in data.ts (wired by the health
  // timer at M4). Helpers receive it as a function; `$` never crosses an import.
  const runDiag = (argv: readonly string[]): Promise<RunResult> => $.process.run(argv, { timeoutMs: CMD_TIMEOUT_MS })
  void runDiag

  on('session.start', async ($, e, next) => {
    await soft($, async () => {
      // First-write initialization so every later read is defined. The refs are
      // spelled per call: validate lists only literal references (M-13).
      if ((await $.state.get(usageRef)).value === undefined) await $.state.set(usageRef, EMPTY_USAGE)
      if ((await $.state.get(healthRef)).value === undefined) await $.state.set(healthRef, EMPTY_HEALTH)
      if ((await $.state.get(noticeRef)).value === undefined) await $.state.set(noticeRef, '')
    })
    return next(e)
  })

  on('session.end', async ($, e, next) => {
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
