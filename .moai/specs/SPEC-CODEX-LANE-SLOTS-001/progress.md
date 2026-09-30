# SPEC-CODEX-LANE-SLOTS-001 — Progress

Status: draft (plan-phase artifacts authored 2026-09-30, card t1378).

## Premise Verification (plan-phase, this tree `d194083fb`)

Verdict: **premise HOLDS — root cause confirmed.** Evidence observed in this run, this worktree:

| # | Claim | Evidence (file:line + observed) |
|---|-------|-------------------------------|
| V1 | Bounded auto scan ignores `maxCanonical`; exact error string | `internal/kanban/factory_slots.go:246-256` — `if maxSlots > 0 { if auto { for candidate := 1; candidate <= maxSlots; candidate++ ... if n == 0 { return claim, fmt.Errorf("factory run %s has no free lane slots in 1..%d", runID, maxSlots) } ... } }`. `maxCanonical` computed at lines 220-235 but used only in the unbounded branch (line 262: `n = maxCanonical + 1`). |
| V2 | Codex join passes the 1-slot bound | `internal/cli/codex_factory.go:140-141` — `kanban.ClaimFactoryLaneWithin(root, entry.LaneLabel, entry.LaneRole, os.Getpid(), os.Getenv(config.EnvMoaiKanbanID), config.DefaultFactoryLeaderLanes, factoryProcessAlive)`. `config.DefaultFactoryLeaderLanes = 1` at `internal/config/defaults.go:648`. |
| V3 | Claude join is unbounded (asymmetry) | `internal/cli/factory.go:761` — `kanban.ClaimFactoryLane(root, label, auto, os.Getpid(), runID, factoryProcessAlive)` (maxSlots=0 → grows via `maxCanonical+1`). |
| V4 | The bound arrived with card t1294 in the rc.16→rc.23 window | Batch log `git log a8a9b9376..d194083fb` contains `5ee35b2f5 fix(factory): bound Codex lane claims to run slots (t1294)`. |
| V5 | Legacy machinery present and tested | `FactoryLegacyRunError` `internal/kanban/factory_slots.go:128-136`, fired at 238-240; `IsLegacyFactoryLabel` / `factoryLabelNumber` at `internal/kanban/bootstrap.go:336-356`; tests `role_naming_m1_test.go:76/109/126`, `factory_slots_test.go:75-139`. |
| V6 | ① Claim path has no wait/retry | `claimFactoryLane` (factory_slots.go:167-285) contains no sleep/retry/deadline; failure at line 255 is immediate. Long-launch delay must be pre-claim: `enterCodexFactory` → `enterFactoryLaneRun` (factory.go:492) → `enterSelectedFactoryRun` (factory.go:417) → codex pre-exec init → `syscall.Exec` (`internal/cli/codex_direct_posix.go:53`). |
| V7 | ② No worktree scan in the launch path | Greps `filepath.Walk|os.ReadDir|filepath.Glob` over `internal/cli/codex_launcher.go`, `codex_init.go`, `session*.go`: no hits. |
| V8 | ③ Hook-trust message is a wiring-pass emit | `ReTrustGuidance` at `internal/codexwiring/codexwiring.go:74`; launcher execs and does not gate on hook approval. |
| V9 | ④ Owner probe is per-run-row | `moai factory runs` = `OpenFactory` + `ClassifyRuns` (`internal/cli/factory_handoff_recover.go:79-101`); probe `ps -o lstart= -p <pid>` at `internal/homestate/process_fingerprint_unix.go:12`; seam `OwnerClassifier` at `internal/homestate/factory_run_retire.go:44`. |

SPEC ID self-check: `ID="SPEC-CODEX-LANE-SLOTS-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`. Uniqueness: no `SPEC-CODEX-LANE-SLOTS-*` in `.moai/specs/` (catalog listing observed, 1002 entries).

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-audit>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
