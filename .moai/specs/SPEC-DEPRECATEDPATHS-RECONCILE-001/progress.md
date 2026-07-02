---
id: SPEC-DEPRECATEDPATHS-RECONCILE-001
title: "Reconcile DeprecatedPaths — progress tracking"
version: "1.0.0"
status: completed
tier: S
era: V3R6
---

# Progress — SPEC-DEPRECATEDPATHS-RECONCILE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-07-02
- plan_status: audit-ready
- tier: S (2 files edited: `internal/defs/dirs.go` + `internal/defs/dirs_test.go`; + doc-comment sites; no template)
- recommended_direction: **A (un-deprecate)** — remove `design.yaml` + `db.yaml` from `DeprecatedPaths`
- artifacts: spec.md (§1 context + §2 GEARS REQ-DPR-001..005 + §3 inline AC-DPR-001..009 + §4 Out of Scope + §5 cross-refs), plan.md (M1 RED→GREEN, M2 reconcile+verify), progress.md
- investigation summary: origin SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 REQ-VVCR-011/012 deliberately marked both files for removal on premises TRUE at authoring (design-system retirement; `_TBD_` placeholder), which became STALE — `db.yaml` re-established live by SPEC-DB-SYNC-001 (real `migration_patterns`, read by `loadMigrationPatterns`), `design.yaml` retained + reclassified live by SPEC-DEAD-CONFIG-001 §D. Ground-truth-shift reconciliation, not repudiation.
- count reconciliation: total 43→41, Category B 31→29 (Category A 9 + Category C 3 unchanged)
- no-overlap check: SPEC-DEAD-CONFIG-001 (completed) scoped `runtime.yaml` only and explicitly left `DeprecatedPaths` intact + parked `design.yaml` "Do NOT touch" — no duplication with this SPEC
- run-phase decision point: §A.4 count-derivation reconciliation (Option A.2 recommended — redirect `@MX:ANCHOR` to this SPEC, preserve completed origin body immutability)

## §E.2 Run-phase Evidence

Milestones executed: M1 (RED→GREEN lockstep) + M2 (reconcile + verify). cycle_type=tdd. 3 files edited: `internal/defs/dirs.go`, `internal/defs/dirs_test.go`, `internal/cli/v2_detection.go` (all committed by orchestrator — race-safe path-limited commit). No template, no loader, no `make build`.

### AC PASS/FAIL matrix (observed evidence)

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-DPR-001 | PASS | `grep -c 'sections/design.yaml\|sections/db.yaml' internal/defs/dirs.go` | `0` (neither path appears as a DeprecatedPaths entry) |
| AC-DPR-002 | PASS | `go test -run TestDeprecatedPathsTotalCount ./internal/defs/` | `ok` — `want = 41` asserted, `len(DeprecatedPaths) == 41` |
| AC-DPR-003 | PASS | `go test -run TestDeprecatedPathsCategorySplit ./internal/defs/` | `ok` — `wantCategoryB = 29` asserted |
| AC-DPR-004 | PASS | `go test -run TestDeprecatedPathsCategoryBExpectedEntries ./internal/defs/` | `ok` — 2 config-yaml paths dropped from enumeration slice (count 29) |
| AC-DPR-005 | PASS | per-path `grep -c` on `internal/defs/dirs.go` | `.moai/db`→1, `project/brand`→1, `gate.yaml`→1, `github-actions.yaml`→1, `memo.yaml`→1 (all 5 retained) |
| AC-DPR-006 | PASS | `git status --porcelain internal/template/` + loaders | `template_changes=0`; `loader_changes=0` (loader.go / loader_design.go / hook.go untouched). Only `internal/defs/{dirs.go,dirs_test.go}` + `internal/cli/v2_detection.go` modified |
| AC-DPR-007 | PASS | `go test ./internal/defs/ ./internal/cli/` | `ok internal/defs 0.4s`; `ok internal/cli 10.9s` — self-adjusting `len(defs.DeprecatedPaths)` consumers (`update_e2e_test.go:100`, `update_cleanup_test.go:145`) auto-adjusted to 41 and stayed green |
| AC-DPR-008 | PASS | `grep -c 'SPEC-DEPRECATEDPATHS-RECONCILE-001' internal/defs/dirs.go internal/defs/dirs_test.go` | `dirs.go:3`, `dirs_test.go:8` — `@MX:ANCHOR`/`@MX:REASON` in BOTH files cite this reconcile SPEC as the 41-entry authority (Option A.2), origin §A.4 kept as historical 43-entry derivation. Completed origin SPEC body NOT edited |
| AC-DPR-009 | PASS | `GOOS=windows GOARCH=amd64 go build ./internal/defs/ ./internal/cli/` | `windows_touched_exit=0`; native `go build ./...` also exit 0 |

Supplementary: `golangci-lint run ./internal/defs/... ./internal/cli/...` → `0 issues.` (no NEW lint findings). RED evidence captured pre-GREEN: the 3 count tests failed at `43→41` / `31→29` before the `dirs.go` slice edit, confirming the tests exercise the slice.

### REQ-DPR-004 §A.4 reconciliation — Option A.2 applied

The `@MX:ANCHOR`/`@MX:REASON` in `internal/defs/dirs.go` and `internal/defs/dirs_test.go` were redirected to state that the 41-entry count is governed by **SPEC-DEPRECATEDPATHS-RECONCILE-001**, citing origin SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 §A.4 as the historical 43-entry derivation. The completed origin SPEC body was left immutable (zero completed-SPEC-body edit), honoring the plan's recommended Option A.2.

### Follow-up (deferred sibling reconcile — plan-auditor debt D1/D2, recorded not dropped)

The plan-auditor flagged that this SPEC's harm-#2 fix (false-v2 detection via `probeDeprecatedPathSignal`) is **necessary but not sufficient**. The same mis-deprecation condition (shipped-in-template AND still listed in `DeprecatedPaths`) also holds for the sibling entries below, which this SPEC deliberately did NOT touch to preserve its Tier S scope:

- `.moai/project/brand` — the `.moai/project/brand/` directory is shipped in the v3 template (`internal/template/templates/.moai/project/brand/`) yet still a Category B `DeprecatedPaths` entry.
- `.claude/rules/moai/design/constitution.md` — parent dir `.claude/rules/moai/design` is a Category B `DeprecatedPaths` entry, but `design/constitution.md` is referenced as a live FROZEN design-system asset (CLAUDE.md §9).

Because these siblings remain in `DeprecatedPaths`, `probeDeprecatedPathSignal` will STILL return positive on a freshly-initialized v3 project that ships them — so removing only design.yaml + db.yaml does not fully eliminate false-v2 detection. This debt is recorded here (not silently dropped). A separate reconcile SPEC should evaluate whether `.moai/project/brand` and `.claude/rules/moai/design` are (a) genuinely shipped-and-live (→ un-deprecate, same Direction A as this SPEC) or (b) shipped-in-error (→ remove from template, keep deprecated). That evaluation requires verifying downstream Go/skill consumption of brand + design assets and is out of this SPEC's narrow scope.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-02
run_commit_sha: 0cdf18e07
run_status: implemented
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 0   # PRESERVE list respected: template + loaders + Category A/C + .moai/db + project/brand + gate/github-actions/memo entries all untouched
l44_pre_commit_fetch: orchestrator-owned (this agent does NOT commit/push per delegation)
l44_post_push_fetch: orchestrator-owned
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues on touched packages
cross_platform_build:
  native: pass          # go build ./... exit 0
  windows: pass         # GOOS=windows GOARCH=amd64 go build ./internal/defs/ ./internal/cli/ exit 0
total_run_phase_files: 3   # internal/defs/dirs.go, internal/defs/dirs_test.go, internal/cli/v2_detection.go
m1_to_mN_commit_strategy: single race-safe path-limited commit by orchestrator (agent edit + test only; DO NOT commit per delegation)
```

## §E.4 Sync-phase Audit-Ready Signal

- Tier S 통합 close (run+sync 단일 커밋, small SPEC consolidated lifecycle).
- run-phase는 격리 워크트리(`agent-a44b6e4497c5ed39a`, base==main HEAD `7ca661c33`)에서 실행 → orchestrator가 4파일 diff를 main으로 reconcile(base==main 확인, 충돌 없음) 후 재검증(`go test ./internal/defs/ ./internal/cli/` PASS) + path-limited 커밋.
- spec.md/progress.md frontmatter: `status: draft → completed`, `era: V3R6`.
- CHANGELOG.md `[Unreleased] Fixed` 항목 추가.
- 9 AC 전부 GREEN (§E.3). follow-up sibling debt(brand/ + design/constitution.md) 기록 유지(§E.2) — 별도 reconcile SPEC 소관.

### sync_commit_sha

sync_commit_sha: 0cdf18e07
run_commit_sha: 0cdf18e07
