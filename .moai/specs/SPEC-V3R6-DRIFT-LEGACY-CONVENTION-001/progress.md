# Progress — SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001

> Run-phase progress tracking for the `moai spec drift` comprehensive false-positive
> elimination + era/grandfather/terminal alignment + close-subject doctrine SPEC.
> Tier M, cycle_type=tdd (RED-GREEN-REFACTOR). Status transitioned draft→in-progress
> on the M1 commit per the Status Transition Ownership Matrix (manager-develop owns
> `draft → in-progress`).

## §E — Phase 0.95 Mode Selection

- **Input parameters**: tier = M; scope = ~4 files (drift.go, transitions.go, 2-3 new
  _test.go, 2 doctrine files + 1 template mirror); domain = Go source + doctrine markdown
  + template mirror; file language mix = mostly Go + some markdown; concurrency benefit =
  LOW (coding-heavy, single-package internal/spec, sequential milestone dependency);
  Agent Teams prereqs = not met (single-package coding task).
- **Decision**: `sub-agent` (Mode 5 — sequential sub-agent per milestone).
- **Justification**: This is coding-heavy work in a single package (`internal/spec`) with
  a strict milestone dependency chain (M1/M2 independent → M3 depends on classifier seam
  stability → M4 doctrine independent → M5 verification gate). Per Finding A4
  (coding-task parallelism caveat), sequential sub-agent is the correct default for coding
  tasks; parallel/agent-team modes add coordination overhead without benefit for a
  single-package edit.

## Milestone Progress

| Milestone | Status | Commit | Notes |
|-----------|--------|--------|-------|
| M1 — era exemption + terminal authority (drift.go) | in-progress | (this commit) | DetectDrift ③+④; D3 record-emission contract |
| M2 — stale sync→completed rule correction (transitions.go) | pending | — | 4-phase model: sync = implemented |
| M3 — combined-scope secondary prefix-grep fallback (drift.go) | pending | — | 3-gate LSGF-001-safe |
| M4 — close-subject full-ID doctrine amendment + mirror | pending | — | lifecycle-sync-gate.md + spec-frontmatter-schema.md |
| M5 — verification gate + genuine-⑤ residual classification | pending | — | strict drift decrease + residual handoff |

## §E.2 Run-phase Evidence

| AC | Severity | Status | Verification | Actual Output |
|----|----------|--------|--------------|---------------|
| AC-DLC-001 | MUST-PASS | pending | drift_combined_scope_test + live jq | — |
| AC-DLC-002 | MUST-PASS | pending | transitions_test + live | — |
| AC-DLC-003 | MUST-PASS | PASS | TestDetectDrift_TerminalStateAuthoritative | superseded/archived/rejected → Drifted=false (record PRESERVED) |
| AC-DLC-004 | MUST-PASS | PASS | TestDetectDrift_GrandfatherEraExempt + V3R6StillDetected | V2.x/V3R2-R4 exempt; V3R6 still detected; AP-3 guard passes |
| AC-DLC-005 | MUST-PASS (gate) | pending | drift_chore_skip_test | — |
| AC-DLC-006 | MUST-PASS (gate) | pending | drift_specid_grep_test | — |
| AC-DLC-007 | MUST-PASS (gate) | pending | audit --json diff vs /tmp/audit-baseline.json | — |
| AC-DLC-008 | MUST-PASS | pending | grandfathered-sibling no-false-positive subtest | — |
| AC-DLC-009 | SHOULD-PASS | pending | drift --count before/after | baseline 54 |
| AC-DLC-010 | MUST-PASS (gate) | pending | go test ./internal/spec/... + coverage | — |
| AC-DLC-011 | MUST-PASS | pending | doctrine co-located prohibition oracle (Go regexp) | — |
| AC-DLC-012 | MUST-PASS | pending | drift_combined_scope_test collision guard | — |

## PRESERVE Verification

- `internal/spec/era.go` — byte-unchanged (verified `git status --porcelain` clean): PASS
- `internal/spec/audit.go` — byte-unchanged (verified): PASS

## Baselines (captured pre-fix)

- `moai spec drift --count` = 54 (drifted records)
- `moai spec audit --json`: grandfathered=268, modern_era_clean=24, drift_findings=291
- Named exemplars (pre-fix): SPEC-CCSYNC-CLAUDEMD-001 [completed↔implemented Drifted],
  SPEC-CCSYNC-TOOLCAT-001 [completed↔implemented Drifted],
  SPEC-CCSYNC-DYNWF-001 [in-progress↔implemented Drifted — collision case].
