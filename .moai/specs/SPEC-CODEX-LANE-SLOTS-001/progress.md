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
| V9 | ④ Owner probe is per-run-row | `moai factory runs` = `OpenFactory` + `ClassifyRuns` (`internal/cli/factory_handoff_recover.go:79-101`); the probe is `platformProcessFingerprint(pid)` — per-pid, three build-tagged variants: `process_fingerprint_unix.go:11` (`!windows && !darwin`, `ps` subprocess), `process_fingerprint_darwin.go:11` (darwin, `unix.SysctlKinfoProc` sysctl — no subprocess), `process_fingerprint_windows.go:10` (windows); seam `OwnerClassifier` at `internal/homestate/factory_run_retire.go:44`. (Iteration-1 correction: the earlier citation named only the unix `ps` variant, which is build-excluded on this darwin tree — the measured 3.0s user / 1.86s sys cost cannot be attributed to `ps` alone.) |

SPEC ID self-check: `ID="SPEC-CODEX-LANE-SLOTS-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`. Uniqueness: no `SPEC-CODEX-LANE-SLOTS-*` in `.moai/specs/` (catalog listing observed, 1002 entries).

## Plan-audit Iterations

- **Iteration 1 — FAIL 0.75 vs Tier M 0.80** (report: `.moai/reports/t1378/plan-audit.md`). Fixes applied 2026-09-30 on tree `850aefca9`, no commits: D1 (AC-009/AC-010 added for REQ-007/008/012), D2 (RED-now evidence ledger in acceptance.md — RED-1/RED-2/RED-3/RED-4 executed on the pre-implementation tree at `850aefca9`, scratch probe deleted after capture; regression-guard classification for AC-002/003/004/005/008), D3 (POSIX/Windows exec-split clause at both `syscall.Exec` mentions), D4 (V9, plan §B.4/M3/§E.4 corrected to the three platform variants), D5 (AC-006 package pinned to `internal/cli`), D6 (REQ-012 trigger made operator-configurable with the default in `internal/config/defaults.go`), D9 (`related_specs` dropped from frontmatter; §G body cross-references are the carrier).
- **Iteration 2 — FAIL narrow 0.94, threshold cleared, MP-8 RED re-executability** (report: `.moai/reports/t1378/plan-audit-iter2.md`). Fixes applied 2026-09-30 on tree `4a7a2b0a4`, no commits: D1 (probe committed as `redprobe_t1378_main.go` in this directory with `//go:build ignore` — `go run <file>` file-mode verified by execution; RED-1/RED-2 re-executed at `4a7a2b0a4`, outputs byte-identical to the `850aefca9` originals; RED-4 de-piped to an explicit file list and re-executed; probe documented synthetic-root-only), D2 (Command lines restored on AC-002/003/004/005), D3 (plan §F M2 AC list now includes AC-009), D4/D5 (all `-run` patterns in acceptance.md anchored with `^...$` exact-name regexes to clear the unanchored-run lint warnings).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-01
plan_audit_verdict: PASS (score 1.00, Tier M threshold 0.80; trajectory 0.75 → 0.94 → 1.00 over 3 iterations, no STOP signal)
plan_audit_report: .moai/reports/t1378/plan-audit-iter3.md (full stream: plan-audit.md, plan-audit-iter2.md, plan-audit-iter3.md)
plan_audit_sha: bfdaf564cf8b32755921123780a1485a92bfc980
plan_artifact_hash: 314f8659a13eb14b74678fd343975a8bfdfebea43fca758ac2c9d5694dc21e
recorded_by: lane orchestrator (verdict landed after manager-spec's final fix turn; audit-ready signal derived from the iteration-3 verdict)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

### Kickoff record (operator-held gate)

- The card dispatch (leader → lane, 2026-09-30) marked the plan→run Kickoff for OPERATOR DIRECT ANSWER — the keep-set operator form, not the autonomous transition.
- Operator answered via AskUserQuestion in the lane on 2026-10-01: **착수 (proceed to run phase)** — the recommended option; alternatives offered were SPEC 재검토 (revise) and 중단 (abort).
- Gate evidence at ask time: plan-audit iter3 PASS 1.00 (≥ 0.80 Tier M), RED citations 4/4 re-executed and reproduced at `bfdaf564c`, tree-sourced spec lint 0 errors / 0 warnings, artifact hash `314f8659a13eb14b74678fd343975a8bfdfebea43fca758ac2c9d5694dc21e` fixed since the verdict.

### Input parameters

- tier: M · scope: ~9-10 files (internal/kanban claim+tests, internal/cli factory/codex_factory/launcher, internal/homestate store+probe variants, internal/config defaults) · domain count: 4 · file language mix: 100% Go · concurrency benefit: LOW (coding-heavy, dependency-ordered milestones).

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-file semantic change, not a typo/one-liner |
| fanout | no | coding-heavy work (Anthropic coding-task parallelism caveat); milestones are dependency-ordered |
| sweep | no | not a mechanical uniform transform; Kickoff just passed but the transform rule is per-milestone semantic |
| agent-team | no | not operator-requested; experimental surface stays unselected |
| serial | **yes** | one manager-develop per milestone, M1→M2→M3→M4 |

Decision: serial

Justification: the four milestones are strictly dependency-ordered (M1 capacity record feeds M2 claim semantics; M3 instrumentation is independent but same-tree; M4 parity verifies the whole), the work is coding-heavy Go across three packages, and a single manager-develop with the Section A-E delegation template carries the full context cheaper than any fan-out. Boundary case: scope estimate ~9-10 files sits at the fanout threshold ±1 — the tie-breaker resolves to the simpler mode (serial), consistent with the coding-heavy override.
