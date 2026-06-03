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
| M1 — era exemption + terminal authority (drift.go) | done | bd7b6bcc4 | DetectDrift ③+④; D3 record-emission contract |
| M2 — stale sync→completed rule correction (transitions.go) | done | 4b301e0ab | 4-phase model: sync = implemented |
| M3 — combined-scope secondary prefix-grep fallback (drift.go) | done | c64b050f7 | 3-gate LSGF-001-safe; live CCSYNC CLAUDEMD+TOOLCAT resolved |
| M4 — close-subject full-ID doctrine amendment + mirror | done | (this commit) | lifecycle-sync-gate.md + spec-frontmatter-schema.md + template mirror (§25 C2 generalized) |
| M5 — verification gate + genuine-⑤ residual classification | pending | — | strict drift decrease + residual handoff |

## §E.2 Run-phase Evidence

| AC | Severity | Status | Verification | Actual Output |
|----|----------|--------|--------------|---------------|
| AC-DLC-001 | MUST-PASS | PASS | TestDetectDrift_CombinedScopeFallback + TestCombinedScopeCloseMatches + live jq | synthetic: BOTH FOO+BAR resolve completed; live: CLAUDEMD+TOOLCAT Drifted=false |
| AC-DLC-002 | MUST-PASS | PASS | TestClassifyPRTitle_StaleSyncRuleCorrected + live | bare sync/docs(sync)→implemented; close-infix→completed; STATUSLINE-STDINFIELDS absent from drift |
| AC-DLC-003 | MUST-PASS | PASS | TestDetectDrift_TerminalStateAuthoritative | superseded/archived/rejected → Drifted=false (record PRESERVED) |
| AC-DLC-004 | MUST-PASS | PASS | TestDetectDrift_GrandfatherEraExempt + V3R6StillDetected | V2.x/V3R2-R4 exempt; V3R6 still detected; AP-3 guard passes |
| AC-DLC-005 | MUST-PASS (gate) | PASS | drift_chore_skip_test + drift_test backfill | all chore-skip + backfill-skip regression green |
| AC-DLC-006 | MUST-PASS (gate) | PASS | TestGetGitImpliedStatus_SPECIDWordBoundary (5 subcases) | word-boundary protection unchanged; HARNESS001 NOT planned |
| AC-DLC-007 | MUST-PASS (gate) | PASS | main-HEAD binary vs my binary audit --json diff (same corpus) | byte-identical; era.go/audit.go call NONE of modified funcs |
| AC-DLC-008 | MUST-PASS | PASS | TestDetectDrift_CombinedScopeCollisionGuard + era V3R2/V3R3 exemption | grandfathered sibling not newly flagged |
| AC-DLC-009 | SHOULD-PASS | PASS | drift --count before/after | 54 → 7 (strict decrease) |
| AC-DLC-010 | MUST-PASS (gate) | PASS | go test ./internal/spec/... + go vet + golangci-lint | green; lint 0 issues |
| AC-DLC-011 | MUST-PASS | PASS | TestCloseSubjectDoctrineAmendment (portable Go regexp, D-NEW-2) | both doctrine files co-located prohibition; template mirror §25-generalized; mirror-parity + leak tests green |
| AC-DLC-012 | MUST-PASS | PASS | TestDetectDrift_CombinedScopeCollisionGuard + TestCombinedScopeCloseMatches (OTHER/FOOBAR/EXTRA) | non-sibling partial-prefix not mapped; D-NEW-1 FOO vs FOOBAR guard |

## PRESERVE Verification

- `internal/spec/era.go` — byte-unchanged (verified `git status --porcelain` clean): PASS
- `internal/spec/audit.go` — byte-unchanged (verified): PASS

## Baselines (captured pre-fix)

- `moai spec drift --count` = 54 (drifted records)
- `moai spec audit --json`: grandfathered=268, modern_era_clean=24, drift_findings=291
- Named exemplars (pre-fix): SPEC-CCSYNC-CLAUDEMD-001 [completed↔implemented Drifted],
  SPEC-CCSYNC-TOOLCAT-001 [completed↔implemented Drifted],
  SPEC-CCSYNC-DYNWF-001 [in-progress↔implemented Drifted — collision case].
