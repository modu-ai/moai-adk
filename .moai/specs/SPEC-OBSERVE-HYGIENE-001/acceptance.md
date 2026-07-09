---
id: SPEC-OBSERVE-HYGIENE-001
title: "Observation Sink Hygiene — Acceptance Criteria"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "observation-sinks, log-hygiene, pruning, sync-gate, workflow-reflex, acceptance"
---

# SPEC-OBSERVE-HYGIENE-001 — Acceptance Criteria

> Observable, testable assertions derived from spec.md §Requirements (GEARS). Each AC traces to REQs and an audit finding (H4/H5/H6/L7).

## §D AC Matrix

| AC ID | REQ trace | Finding | Severity | Description |
|-------|-----------|---------|----------|-------------|
| AC-OBH-001 | REQ-OBH-001 | H4/H5 | MUST-PASS | `moai spec audit` consumes status-transition-audit.log: fixture with an ownership-mismatched transition site yields an INFO finding; absent/corrupt log yields zero findings and exit 0 (or the documented D1(b) fallback if that direction is decided) |
| AC-OBH-002 | REQ-OBH-002 | H4 | MUST-PASS | SessionEnd path prunes zero-byte trace-*.jsonl and ages out traces past the D2 threshold (tests in t.TempDir()); task-metrics.jsonl disposition recorded (repair / document+age-out / retire) with a root-cause note |
| AC-OBH-003 | REQ-OBH-003 | H4 | MUST-PASS | fact-force-skip.log documented intentionally write-only in gateguard-fact-force.sh header (+ referencing doctrine line) |
| AC-OBH-004 | REQ-OBH-004 | H6 | MUST-PASS | D3 decision recorded and implemented: per the recommended direction, go vet/go build block by default in the sync gate (tests/coverage stay advisory); stdout-JSON exit-0 signaling, --skip-hook escape, and the §4 recovery carve-out text all preserved |
| AC-OBH-005 | REQ-OBH-005 | L7 | MUST-PASS | loop.md snapshot/resume sections carry the best-effort marker (no mechanical writer guarantee); no mechanical persistence code added |
| AC-OBH-006 | REQ-OBH-006 | L7 | MUST-PASS | sunset.yaml + harness.yaml model_upgrade_review carry one-line dormancy annotations (sunset: no runtime hot path; MUR: trigger env unset), live + mirrors, with zero semantic change (config tests bit-identical green) |
| AC-OBH-007 | REQ-OBH-001..006 | — | MUST-PASS | Full suite + cross-platform build green; new Go paths ≥85% coverage; all edited mirrored surfaces template-first; `make build` green; no `.moai/logs/` content committed |

## §D.1 Severity Classification

All 7 ACs are MUST-PASS. AC-OBH-004 carries a **decision-gate precondition**: it is evaluated only against the D3 direction the user confirms at Implementation Kickoff Approval; implementing a blocking default without the recorded decision is itself a FAIL condition. AC-OBH-001 similarly binds to the D1 recorded direction (recommended (a)).

## §D.2 Given-When-Then Scenarios

### AC-OBH-001 — audit log consumer

**Given** a fixture `.moai/logs/status-transition-audit.log` containing a transition line whose writer does not match the Ownership Matrix owner for that transition
**When** `moai spec audit --json` runs against the fixture project
**Then** the output carries an INFO-severity finding naming the mismatch; **and given** the log is absent or unparseable, **then** the audit completes with zero such findings and exit 0.

### AC-OBH-002 — telemetry pruning

**Given** a `t.TempDir()` logs dir with a zero-byte trace file, a fresh non-empty trace, an over-threshold-age non-empty trace, and a task-metrics.jsonl
**When** the SessionEnd hook path executes
**Then** the zero-byte and over-age traces are removed, the fresh trace and (per D2 disposition) task-metrics survive or age out as documented, and no file outside the logs dir is touched.

### AC-OBH-004 — gate promotion (assuming D3 recommended direction)

**Given** a Go project with a deliberate compile error and no `MOAI_SYNC_GATE_BLOCKING` env var
**When** the sync-phase quality gate Stop hook fires on a sync-commit completion turn
**Then** the hook emits the stdout-JSON block decision on exit 0 for the go build failure; **and given** a recovery-signal turn or `--skip-hook`, **then** the documented escape/carve-out paths behave unchanged; **and** test/coverage failures alone do not block.

### AC-OBH-006 — dormancy annotations

**Given** the post-M3 config files
**When** reading sunset.yaml and harness.yaml model_upgrade_review
**Then** each carries a one-line comment naming its dormancy layer and pointer (sunset → no runtime hot path, Go-side notice exists; MUR → consumer exists, CLAUDE_MODEL_PREVIOUS/CLAUDE_MODEL never set), and `go test ./internal/config/...` passes with zero loader behavior diffs.

## §D.3 Edge Cases

- **EC-1**: audit-log line format drift across hook versions → parser tolerates unknown line shapes (skip + count, never error).
- **EC-2**: logs dir absent entirely at SessionEnd → pruning no-ops silently.
- **EC-3**: trace file being written concurrently at prune time (live session) → prune only files not owned by the current session-id; never the active trace.
- **EC-4**: D3 blocking default fires on a non-Go project → language detection resolves elsewhere; no vet/build steps, no block (scope is Go-toolchain-only per REQ-OBH-004's Where gate).
- **EC-5**: sibling SPEC's loop.md hunks land between plan and run → re-anchor the marker insertion by content tokens, not line numbers.

## §D.4 Verification Commands (indicative)

```bash
go test -run TestSpecAudit ./internal/spec/... ./internal/cli/...      # AC-OBH-001
go test -run TestPrune -run TestSessionEnd ./internal/hook/...          # AC-OBH-002 (names indicative)
grep -n "write-only" .claude/hooks/moai/gateguard-fact-force.sh         # AC-OBH-003
grep -n "MOAI_SYNC_GATE_BLOCKING\|blocking" .claude/hooks/moai/sync-phase-quality-gate.sh  # AC-OBH-004
grep -n "best-effort" .claude/skills/moai/workflows/loop.md             # AC-OBH-005
grep -n "dormant\|DORMANT" .moai/config/sections/sunset.yaml .moai/config/sections/harness.yaml  # AC-OBH-006
go test ./internal/config/...                                            # AC-OBH-006 (bit-identical semantics)
git diff --cached --stat | grep -c ".moai/logs/"                        # AC-OBH-007 (expect 0)
make build && go test ./... && GOOS=windows GOARCH=amd64 go build ./... # AC-OBH-007
```

## §D.5 Quality Gate Criteria

- TRUST 5: Tested (fixture-driven consumer + pruning tests in t.TempDir()), Readable (per-sink disposition stated where each sink lives), Unified (lint clean vs baseline), Secured (no log content committed; prune never escapes logs dir), Trackable (Conventional Commits; D1/D2/D3 decisions recorded in progress.md + commit bodies).
- Config safety: comment-only YAML edits proven behavior-identical by the config test suite.

## §D.6 Definition of Done

1. All 7 ACs PASS with verbatim evidence in run-phase §E self-verification.
2. D1/D2/D3 decisions recorded (user-confirmed where required) before their milestones commit.
3. Every one of the 6 sink groups has exactly one recorded disposition — consumer, documented write-only, or pruned/retired — with no sink left ambiguous.
4. Sibling boundary check: zero diffs under `.moai/harness/` and `.moai/state/verify/`.
