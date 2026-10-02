// moai-status health layer: `$`-free. The doctor box-row parser (row-first,
// REQ-MSM-012), the memory-doctor JSON parser and the health-line composer
// (REQ-MSM-008/-009). Pure: bun-tested under tests/pure/. The message
// spellings are pinned from the in-tree sources (M-9: internal/cli/doctor.go,
// internal/cli/doctor_mcp_version.go); an unrecognized warn still warns with
// the bounded raw message — never healthy by accident.
import type { MoaiStatusGoodVerdict, MoaiStatusHealth, MoaiStatusSourceVerdict } from '../types'
import { STATUS_LINE_MAX } from './data'

const SHA = '[0-9a-f]+'
const ROW = /^(ok|warn|fail)\s+(\S.*)$/
const BEHIND = new RegExp(`^binary is behind source tree \\(binary: (${SHA}), HEAD: (${SHA})\\)$`)
const NEWER = new RegExp(`^binary is newer than this tree — freshness undetermined \\(binary: (${SHA}), HEAD: (${SHA})\\)$`)
const STALE = /^running MCP server is stale \((.+)\)$/

const bound = (text: string, maxCp: number): string =>
  Array.from(text).length <= maxCp ? text : `${Array.from(text).slice(0, maxCp - 1).join('')}…`

const warnFromMessage = (message: string): MoaiStatusSourceVerdict => {
  const behind = BEHIND.exec(message)
  if (behind) return { state: 'warn', segment: `${behind[1]} behind ${behind[2]}` }
  const newer = NEWER.exec(message)
  if (newer) return { state: 'warn', segment: `${newer[1]} newer than ${newer[2]} (undetermined)` }
  const stale = STALE.exec(message)
  if (stale) return { state: 'warn', segment: `stale (${stale[1]})` }
  return { state: 'warn', segment: bound(message, 80) }
}

// Row-first precedence (REQ-MSM-012): the first box row FOR THE ASKED CHECK
// decides; the parser sees stdout alone, so a parseable row always wins and a
// non-zero exit means unknown only on the no-row path (the doctor single checks
// exit 0 on warn anyway, M-9). Any STATUS token other than ok/warn — `fail`
// included, unreachable from the fixed argv table — is not a verdict: unknown,
// never healthy. A summary line (`0 ok, 1 warn, 0 fail`) starts with a digit and
// never matches the row shape.
export const parseDoctorCheck = (stdout: string, checkName: string): MoaiStatusSourceVerdict => {
  for (const raw of stdout.split('\n')) {
    let line = raw.trim()
    if (line.startsWith('│')) line = line.slice(1).trim()
    if (line.endsWith('│')) line = line.slice(0, -1).trim()
    const match = ROW.exec(line)
    if (match === null) continue
    const token = match[1]
    const rest = match[2]
    if (!rest.startsWith(checkName)) continue
    const after = rest.slice(checkName.length)
    if (after !== '' && after[0] !== ' ') continue
    const message = after.trim()
    if (token === 'ok') return { state: 'ok' }
    if (token === 'warn') return warnFromMessage(message)
    return { state: 'unknown' }
  }
  return { state: 'unknown' }
}

type MemoryStore = { topic_files?: unknown; cap?: unknown }

// The retention signal (M-8): the worst over-cap store of `moai memory doctor
// --json`. Unparseable output, a non-array, and a payload with no judgeable
// store are unknown — never healthy (REQ-MSM-009).
export const parseMemoryDoctor = (stdout: string): MoaiStatusSourceVerdict => {
  let parsed: unknown
  try {
    parsed = JSON.parse(stdout)
  } catch {
    return { state: 'unknown' }
  }
  if (!Array.isArray(parsed)) return { state: 'unknown' }
  let worst: { files: number; cap: number; ratio: number } | undefined
  let judged = 0
  for (const item of parsed as MemoryStore[]) {
    const files = item?.topic_files
    const cap = item?.cap
    if (typeof files !== 'number' || typeof cap !== 'number' || !Number.isFinite(files) || !Number.isFinite(cap) || cap <= 0)
      continue
    judged++
    if (files > cap) {
      const ratio = files / cap
      if (worst === undefined || ratio > worst.ratio) worst = { files, cap, ratio }
    }
  }
  if (judged === 0) return { state: 'unknown' }
  if (worst === undefined) return { state: 'ok' }
  return { state: 'warn', segment: `${worst.files}/${worst.cap} files` }
}

export type FreshVerdicts = {
  binary: MoaiStatusSourceVerdict
  mcp: MoaiStatusSourceVerdict
  memory: MoaiStatusSourceVerdict
}

const LABEL = { binary: 'binary', mcp: 'mcp', memory: 'memory' } as const

// Warnings only (Q4): warn segments join with ` · `; unknown sources show as
// `name ?`; a clean cycle composes '' — the caller clears the line
// (REQ-MSM-008). The line is bounded to STATUS_LINE_MAX code points.
export const composeHealthLine = (fresh: FreshVerdicts, maxCp: number = STATUS_LINE_MAX): string => {
  const parts: string[] = []
  for (const source of ['binary', 'mcp', 'memory'] as const) {
    const verdict = fresh[source]
    if (verdict.state === 'warn') parts.push(`${LABEL[source]} ${verdict.segment}`)
    else if (verdict.state === 'unknown') parts.push(`${LABEL[source]} ?`)
  }
  if (parts.length === 0) return ''
  return bound(parts.join(' · '), maxCp)
}

const good = (previous: MoaiStatusGoodVerdict | undefined, verdict: MoaiStatusSourceVerdict): MoaiStatusGoodVerdict => {
  if (verdict.state === 'unknown') return previous ?? { state: 'ok' } // the last good stands
  return verdict.state === 'ok' ? { state: 'ok' } : { state: 'warn', segment: verdict.segment }
}

// The state shape keeps only good classifications; a source reading unknown
// keeps its previous slot (REQ-MSM-009 — the last good classification stands).
export const mergeGoodHealth = (previous: MoaiStatusHealth | undefined, fresh: FreshVerdicts): MoaiStatusHealth => ({
  binary: good(previous?.binary, fresh.binary),
  mcp: good(previous?.mcp, fresh.mcp),
  memory: good(previous?.memory, fresh.memory),
})
