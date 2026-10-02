// State contract of the moai-status plugin (Claude Code function hooks, 2.1.287).
// Self-contained on purpose: the engine requires a contract with no import, its
// exported names led by the plugin's PascalCase name, and `PluginState` declared
// for the plugin's own name. `plugin.json` names this file under "types".

/** The context band and the per-window holds the classifier classifies against (spec.md D-3, plan §G). */
export type MoaiStatusBand = {
  softLargePct: number
  softStandardPct: number
  largeWindowCutoff: number
  autoCompactPct: number
  hardMarginPct: number
  hardCapPct: number
  fiveHourHoldPct: number
  sevenDayHoldPct: number
}

/** The engine's pushed `session.measure` figure (the laid typings' SessionMeasureInput, structurally). */
export type MoaiStatusMeasureInput = {
  context?: { tokens?: number; window?: number; percent?: number }
  rateLimits?: { kind: string; percentUsed: number; resetsAt?: string }[]
  changed?: string[]
}

/** One warned figure of the strip classification; only warn and critical are stored. */
export type MoaiStatusContextFigure = {
  kind: 'context'
  level: 'warn' | 'critical'
  percent: number
  softPct: number
  hardPct: number
}

export type MoaiStatusRateFigure = {
  kind: 'rate'
  window: 'five_hour' | 'seven_day'
  percent: number
  holdPct: number
}

export type MoaiStatusFigure = MoaiStatusContextFigure | MoaiStatusRateFigure

/** The strip classification held in state; `figures` empty means nothing to say. */
export type MoaiStatusUsage = { figures: MoaiStatusFigure[] }

/** One health source's last good verdict; a source that reads unknown keeps its previous slot. */
export type MoaiStatusGoodVerdict = { state: 'ok' } | { state: 'warn'; segment: string }

export type MoaiStatusHealth = {
  binary: MoaiStatusGoodVerdict
  mcp: MoaiStatusGoodVerdict
  memory: MoaiStatusGoodVerdict
}

/** One cycle's fresh verdict per source; `unknown` shows as `?` and never becomes healthy. */
export type MoaiStatusSourceVerdict = { state: 'ok' } | { state: 'warn'; segment: string } | { state: 'unknown' }

declare module 'claude-code' {
  interface PluginState {
    'moai-status': {
      /** The last strip classification (warn and critical figures only). */
      usage: MoaiStatusUsage
      /** The last good health classification, per source. */
      health: MoaiStatusHealth
      /** The fail-soft guard's one-line failure notice; '' when the last cycle was clean. */
      notice: string
    }
  }
}
