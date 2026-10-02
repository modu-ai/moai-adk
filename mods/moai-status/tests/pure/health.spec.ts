// Pure tests for hooks/health.ts, run with `bun test` (developer-local evidence,
// spec.md G-11). Named *.spec.ts so the engine runner's *.test.ts glob skips them.
import { expect, test } from 'bun:test'
import { composeHealthLine, mergeGoodHealth, parseDoctorCheck, parseMemoryDoctor } from '../../hooks/health'

// Box rows exactly as the CLI prints them (C5 re-measure): the row sits inside
// the box borders as `    <status>    <check name>  <message>  `.
const row = (status: string, name: string, message: string): string =>
  [
    '╭──────────────────────────────────────────╮',
    '│  System Diagnostics                      │',
    '│    STATUS  CHECK        MESSAGE          │',
    `│    ${status}    ${name}  ${message}  │`,
    '╰──────────────────────────────────────────╯',
  ].join('\n')

const BEHIND = row('warn', 'Binary Freshness', 'binary is behind source tree (binary: 802a72235, HEAD: 58dad3055)')
const FRESH = row('ok', 'Binary Freshness', 'binary matches source HEAD (802a72235)')
const DIFFERENT = row('ok', 'Binary Freshness', 'binary from a different branch (binary: 802a72235, HEAD: 0a1b2c3d4)')
const NEWER = row('warn', 'Binary Freshness', 'binary is newer than this tree — freshness undetermined (binary: 802a72235, HEAD: 58dad3055)')
const NO_SERVER = row('ok', 'MCP Server Version', 'no running moai MCP server recorded')
const MCP_MATCH = row('ok', 'MCP Server Version', '1 running MCP server(s) match the installed binary (802a72235)')
const STALE = row('warn', 'MCP Server Version', 'running MCP server is stale (pid 4242: 0a1b2c3d4; binary: 802a72235)')
const FAIL_ROW = row('fail', 'Binary Freshness', 'some exotic condition')

test('health: behind row names both shas', () => {
  expect(parseDoctorCheck(BEHIND, 'Binary Freshness')).toEqual({ state: 'warn', segment: '802a72235 behind 58dad3055' })
})

test('health: fresh and different-branch rows are healthy', () => {
  expect(parseDoctorCheck(FRESH, 'Binary Freshness')).toEqual({ state: 'ok' })
  expect(parseDoctorCheck(DIFFERENT, 'Binary Freshness')).toEqual({ state: 'ok' })
})

test('health: newer-than-tree row is warn undetermined', () => {
  expect(parseDoctorCheck(NEWER, 'Binary Freshness')).toEqual({ state: 'warn', segment: '802a72235 newer than 58dad3055 (undetermined)' })
})

test('health: stale mcp row names pid and commits', () => {
  const verdict = parseDoctorCheck(STALE, 'MCP Server Version')
  expect(verdict).toEqual({ state: 'warn', segment: 'stale (pid 4242: 0a1b2c3d4; binary: 802a72235)' })
})

test('health: no-server and match rows are healthy', () => {
  expect(parseDoctorCheck(NO_SERVER, 'MCP Server Version')).toEqual({ state: 'ok' })
  expect(parseDoctorCheck(MCP_MATCH, 'MCP Server Version')).toEqual({ state: 'ok' })
})

test('health: summary line is not a verdict', () => {
  // The summary line names the same words but is not a box row for the check.
  expect(parseDoctorCheck('0 ok, 1 warn, 0 fail', 'Binary Freshness')).toEqual({ state: 'unknown' })
  // A row whose STATUS token is not ok/warn — `fail` included, unreachable from
  // the asked checks (REQ-MSM-012) — is not a verdict either: unknown, never healthy.
  expect(parseDoctorCheck(FAIL_ROW, 'Binary Freshness')).toEqual({ state: 'unknown' })
  // A missing row is unknown.
  expect(parseDoctorCheck('', 'Binary Freshness')).toEqual({ state: 'unknown' })
  // Another check's row is not this check's verdict.
  expect(parseDoctorCheck(NO_SERVER, 'Binary Freshness')).toEqual({ state: 'unknown' })
})

test('health-rowfirst: non-zero exit with a parseable row still classifies from the row', () => {
  // Row-first precedence (REQ-MSM-012, debt N-4): the parser sees stdout alone,
  // so a row always decides; the exit code is advisory (M-9: the single checks
  // already exit 0 on warn). A non-zero exit means unknown only when no
  // parseable row for the asked check is present — the no-row path above.
  expect(parseDoctorCheck(BEHIND, 'Binary Freshness')).toEqual({ state: 'warn', segment: '802a72235 behind 58dad3055' })
})

test('health: memory json worst over-cap store', () => {
  const worst = JSON.stringify([
    { store: { dir: '/a' }, exists: true, topic_files: 60, cap: 50, index_lines: 1, findings: [] },
    { store: { dir: '/b' }, exists: true, topic_files: 1422, cap: 50, index_lines: 1, findings: [{ Code: 'MEMORY_TOPIC_COUNT_OVER_CAP' }] },
    { store: { dir: '/c' }, exists: true, topic_files: 12, cap: 50, index_lines: 1, findings: null },
    { store: { dir: '/d' }, exists: false, topic_files: 0, cap: 50, index_lines: 0, findings: null },
  ])
  expect(parseMemoryDoctor(worst)).toEqual({ state: 'warn', segment: '1422/50 files' })
  expect(parseMemoryDoctor(JSON.stringify([{ store: { dir: '/a' }, exists: true, topic_files: 49, cap: 50, index_lines: 1, findings: null }]))).toEqual({ state: 'ok' })
  // Unparseable output, a non-array, and stores with no judgeable fields are
  // unknown — never healthy (REQ-MSM-009).
  expect(parseMemoryDoctor('<<< not json >>>')).toEqual({ state: 'unknown' })
  expect(parseMemoryDoctor('{"not":"an array"}')).toEqual({ state: 'unknown' })
  expect(parseMemoryDoctor(JSON.stringify([{ store: { dir: '/a' }, exists: true }]))).toEqual({ state: 'unknown' })
})

test('health: composer joins warns and clears when empty', () => {
  const line = composeHealthLine({
    binary: { state: 'warn', segment: '802a72235 behind 58dad3055' },
    mcp: { state: 'unknown' },
    memory: { state: 'warn', segment: '1422/50 files' },
  })
  expect(line).toBe('binary 802a72235 behind 58dad3055 · mcp ? · memory 1422/50 files')
  expect(composeHealthLine({ binary: { state: 'ok' }, mcp: { state: 'ok' }, memory: { state: 'ok' } })).toBe('')
})

test('health-merge: unknown keeps the previous good verdict', () => {
  const previous = {
    binary: { state: 'warn', segment: '802a72235 behind 58dad3055' } as const,
    mcp: { state: 'ok' } as const,
    memory: { state: 'ok' } as const,
  }
  const merged = mergeGoodHealth(previous, {
    binary: { state: 'unknown' },
    mcp: { state: 'ok' },
    memory: { state: 'warn', segment: '1422/50 files' },
  })
  // The unknown source keeps the last good classification; a fresh good
  // verdict replaces the slot.
  expect(merged.binary).toEqual({ state: 'warn', segment: '802a72235 behind 58dad3055' })
  expect(merged.memory).toEqual({ state: 'warn', segment: '1422/50 files' })
})
