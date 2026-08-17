# Plan — SPEC-V3R6-MOAI-CLEAN-HOME-001

Tier M · Route B (PR) · cycle_type tdd · depends_on SPEC-V3R6-MOAI-HOME-PATHS-001 (run AFTER its run phase — paths.MoaiHome() is the home root)

## 1. Milestones

| M | Scope | Owner | Verification |
|---|-------|-------|--------------|
| M1 | `internal/cli/doctor_disk.go` `checkHomeDisk` + registration; duplicate-cluster heuristic; WARN threshold via defaults.go | manager-develop | `go test ./internal/cli/ -run TestCheckHomeDisk -count=1` |
| M2 | `clean --home`: allowlist category scanners (debug/releases/logs/backups-removed), dry-run report, `--force` deletion, `state.home_retention_days` loader + defaults | manager-develop | `go test ./internal/cli/ -run TestCleanHome -count=1` |
| M3 | Carve-out predicate + guard test; releases keep current+N; project-scope clean regression test | manager-develop | `go test ./internal/cli/ -run TestCleanHomeCarveOut -count=1` |
| M4 | Docs (`clean --help` text, doctor output sample), full verification batch, PR | manager-develop → manager-git | acceptance.md AC matrix |

## 2. Key Files

- `internal/cli/clean.go` — extend `newCleanCmd` with `--home`; keep existing project-scope path byte-compatible.
- `internal/cli/doctor.go` — register `checkHomeDisk` in the check slices of `runGroupedChecksObserved` (doctor.go:160-215; `runDiagnosticChecks` at :226 is only the backward-compat flattener — audit D5 correction).
- `internal/cli/doctor_disk.go` (new), `internal/cli/clean_home.go` (new) — implementation split.
- `internal/config/defaults.go` + state loader — three new constants + one key.
- Home root resolution: `paths.MoaiHome()` — the dependency SPEC-V3R6-MOAI-HOME-PATHS-001 is COMPLETED and merged into this release lane, so `MoaiHome()` (paths.go:68) and `UserConfigSectionsDir()` are live in this tree; sub-paths reuse the existing accessors (audit bonus confirmation).

## 3. PRESERVE List (do not touch)

All SPEC-V3R6-MOAI-HOME-PATHS-001 artifacts · the session WIP files: .mcp.json · internal/cli/todo.go · internal/cli/todo_test.go · internal/cli/update.go · internal/config/defaults.go **(only the three new constants append; surgical Edit only)** · .moai/config/sections/cache.yaml · internal/cli/memory.go · internal/cli/memory_test.go · internal/cli/migrate_profiles.go · internal/cli/migrate_profiles_test.go · internal/hook/memo/taxonomy/linkage.go · linkage_test.go · .moai/reports/**. (The plan-time "uncommitted hunks survive" premise is void — those files are tracked + clean in this release-lane tree.)

## 4. Resolved Decisions

- **[RESOLVED 2026-08-17: duplicate-cluster strictness — audit iteration 1 D1 remediation]** The duplicate-cluster predicate is the lightweight heuristic: same category name + byte-equal total size + equal entry count → reported as a cluster. No content hashing. False positives are accepted by design — the check is advisory/report-only (REQ-MCH-001/006), so an occasional false cluster costs one report line, while content hashing would cost real I/O over the largest trees. (User-approved via AskUserQuestion, 2026-08-17.)
- **[RESOLVED 2026-08-17: carve-out depth semantics — D2]** `isCarvedOut` matches carve-out names against ANY path segment, recursively, at every depth (root and per-profile under `claude-profiles/<p>/`), and the carve-out WINS inside allowlisted containers: a `credentials*`-named file inside an aged `backups/removed-*` directory is never deleted even under `--force`. (User-approved.)
- **[RESOLVED 2026-08-17: retention config source — D3]** `state.home_retention_days` is read from the HOME tier — `~/.moai/config/sections/state.yaml` via `paths.UserConfigSectionsDir()` — NOT from a project-tier state.yaml (cwd-dependent reading would let two projects clean one home with different retentions). (User-approved.)

## 5. Risks

- **Deletion safety**: home-level `--force` is the highest-blast-radius change in this SPEC — dry-run default + carve-out predicate test + allowlist-only scanning are the three guards; no `os.RemoveAll` outside scanned allowlisted paths.
- **HOME-fixture tests**: non-parallel `t.Setenv("HOME", t.TempDir())` discipline (CLAUDE.local.md §6/§13).
- **Threshold mis-tuning**: 500MB seed may WARN-noise on developer machines; threshold is configurable and the check is advisory-only.
