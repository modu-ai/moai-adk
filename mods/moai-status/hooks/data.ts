// moai-status data layer: `$`-free. The fixed argv table (REQ-MSM-002) and the
// shared constants live here; the measure classifier and the line builders join
// in the milestones whose tests define them (M2, M3). Pure: bun-tested under
// tests/pure/. argv lists only, never shell strings; the table is the ONLY argv
// built anywhere.
/** What `$.process.run` resolves with (the laid typings' ProcessRunResult). */
export type RunResult = {
  exitCode: number
  stdout: string
  stderr: string
  isStdoutTruncated: boolean
  isStderrTruncated: boolean
}

/** The run function the hooks module hands the helpers; the single `$.process.run` site wraps it. */
export type Run = (argv: readonly string[]) => Promise<RunResult>

// plan §G parameters. HEALTH_POLL_MS above the floor is Q3's provisional value;
// the tests assert only the floor (REQ-MSM-007).
export const HEALTH_POLL_MIN_MS = 15_000
export const HEALTH_POLL_MS = 60_000
export const CMD_TIMEOUT_MS = 20_000
export const TOAST_EXCERPT_CP = 80
export const STATUS_LINE_MAX = 200

// The fixed argv table (plan §B.3). The check names are the in-tree check
// identifiers (internal/cli/doctor.go:210, internal/cli/doctor_mcp_version.go).
export const ARGV = {
  binary: ['moai', 'doctor', '--check', 'Binary Freshness'],
  mcp: ['moai', 'doctor', '--check', 'MCP Server Version'],
  memory: ['moai', 'memory', 'doctor', '--json'],
} as const
