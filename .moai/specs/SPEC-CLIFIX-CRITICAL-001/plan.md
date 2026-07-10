# SPEC-CLIFIX-CRITICAL-001 — Implementation Plan

## §A Context

P0 row of the CLI audit roadmap (`.moai/reports/cli-improvement-audit-20260710.html` §5): eliminate user-data-loss risk. All 8 fixes are local (1 file each). Development mode per quality.yaml; Reproduction-First applies to every defect (REQ-CRIT-001-009).

## §B Known Issues (findings inventory)

| # | File anchor (re-verify before edit) | Defect | Fix direction |
|---|---|---|---|
| a | glm.go:98, launcher.go:241 | closed `SettingsLocal` struct RMW wipes non-struct top-level keys | `map[string]any` round-trip or `mutateSettingsLocal` seam |
| b | team_spawn.go:316-384 | `ClaimTask` opens `O_RDWR`, writes at offset 0 → ledger head overwritten | `O_APPEND\|O_WRONLY` |
| c | harness_mute.go:198-228 | minimal-struct YAML round-trip wipes `workflow.yaml` sibling keys | yaml.v3 Node API (harness.go:363 pattern) |
| d | harness/v4lifecycle.go:257,285 | bare `"harness-"+name` prefix match deletes sibling harness (`release` vs `release-update`) | `prefix+"-"` boundary match, same for EditHarness |
| e | update_cleanup.go:55 | `acquireUpdateLock` has zero production callers → lockless `moai update` | wire into `runUpdate` before destructive steps |
| f | migrate_agency.go:87-95,398,467 | rollback unconditionally deletes recorded paths incl. pre-existing dirs | snapshot pre-existing paths, restore on rollback; delete only newly created |
| g | migrate_agency.go:446-452 | `os.Stat` follows symlinks → `ModeSymlink` check is dead code | `os.Lstat` / reuse `isSymlinkEntry` |
| h | hook.go:773-792,1167-1206 | Stop-hook auto-classify re-appends promotions each session → unbounded `tier-promotions.jsonl` | high-water mark; record only on tier change |

## §C Pre-flight

1. Re-verify all 8 file:line anchors against the live tree (content-token search, not line numbers).
2. Confirm `mutateSettingsLocal` seam exists and its locking/atomicity semantics (it is also the target seam for SPEC-CLIFIX-CONCURRENCY-001).
3. Confirm harness.go:363 Node-API pattern is reusable for harness_mute.go.
4. Baseline: `go test ./internal/cli/... -count=1` green; `golangci-lint run` baseline captured.

## §D Constraints

- Zero behavior change beyond the defect corrections; no CLI flag/surface changes.
- Each fix lands with its reproduction test in the same milestone (RED evidence first).
- No refactoring "while here" — update.go decomposition and dead-code removal belong to SPEC-CLIFIX-HYGIENE-001.
- Cross-platform: fixes b/e/f/g touch file-system semantics — verify on darwin + linux CI at minimum; `O_APPEND` and lock behavior must not assume POSIX-only call sites are compiled on Windows without guards.

## §E Self-Verification

- E1: AC matrix PASS/FAIL against acceptance.md §D (all 9 ACs).
- E2: `go test ./internal/cli/... -count=1` and `go build ./...` verbatim output.
- E3: package coverage for touched packages not lower than baseline.
- E5: `golangci-lint run` no new findings vs baseline.

## §F Milestones (priority order, no time estimates)

- M1 — Reproduction batch: write failing repro tests for defects a-h (9 test funcs incl. lock contention); confirm RED.
- M2 — Config/ledger integrity fixes: a (SettingsLocal), c (workflow.yaml Node API), b (O_APPEND ledger).
- M3 — Destructive-path fixes: d (prefix boundary), e (update lock wiring), f (rollback snapshot/restore), g (Lstat guard).
- M4 — Growth fix + closure: h (high-water mark), full suite + race check on team_spawn tests, §E self-verification, AC matrix.

## §G Anti-Patterns and Risks

- Risk (shared-file overlap across the CLIFIX series — execution order is strictly P0→P1→P2→P3→P4):
  - glm.go/launcher.go also targeted by CONCURRENCY-001 (writer seam) and HYGIENE-001 (env constants) — this SPEC lands first; siblings rebase on its result.
  - hook.go also targeted by CONTRACT-001 (os.Exit) and HYGIENE-001 (timeout constants).
  - team_spawn.go also targeted by CONTRACT-001 (lock-test filename) and LINTER-STALE-001 (claim validation).
  - migrate_agency.go also targeted by CONTRACT-001 (os.Exit at :590).
  - update.go/update_cleanup.go also targeted by HYGIENE-001 (decomposition + dead-code removal) — HYGIENE must not delete the newly wired lock path.
- Anti-pattern: fixing (a) by adding fields to `SettingsLocal` — the struct stays closed against unknown keys; only an open representation satisfies REQ-CRIT-001-001.
- Anti-pattern: fixing (h) by truncating the JSONL file — history must be preserved; only duplicate suppression is in scope.
- Risk: (f) snapshot cost on large pre-existing dirs — snapshot only paths the plan will mutate.

## §H Cross-References

- Findings SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §1, §4 (cross-cutting patterns rows 1/2/5), §5 P0.
- Siblings: SPEC-CLIFIX-CONTRACT-001 (P1), SPEC-CLIFIX-CONCURRENCY-001 (P2), SPEC-CLIFIX-LINTER-STALE-001 (P3), SPEC-CLIFIX-HYGIENE-001 (P4).
- CLAUDE.local.md §2 (settings.local.json separation), §21 (release + release-update coexistence — the live trigger for defect d).
