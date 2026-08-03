# Progress — SPEC-UPDATE-YAML-PRESERVE-001

## §E.1 Plan-phase Audit-Ready Signal

- **Artifacts**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` created 2026-07-31 by `manager-spec`.
- **Tier**: M (justified in `plan.md` §A).
- **SPEC ID regex check**: executed as Bash, verbatim output `PASS` for `SPEC-UPDATE-YAML-PRESERVE-001` against `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`.
- **ID uniqueness**: `ls .moai/specs/ | grep -c "SPEC-UPDATE-YAML-PRESERVE-001"` → `0` prior occurrences.
- **Baseline verification**: every file:line reference in `spec.md` §A was re-read this session and confirmed against the current tree (table in `plan.md` §A). The round-trip loss was reproduced by executing a scratch test against `internal/template/templates/.moai/config/sections/cache.yaml` with the pinned `gopkg.in/yaml.v3 v3.0.1` — observed `comments in=16 out=0`, keys alphabetized, `"1h"` → `1h`, 2-space → 4-space indent.
- **Edge-case survey**: executed over the 30 shipped section templates — 0 anchors, 0 merge keys, 0 multi-document; 8 files carry block sequences (so REQ-UYP-014 is a live decision, the other three are defensive contracts). 1 `.tmpl` (`quality.yaml.tmpl`) is unparseable by the map decoder due to unquoted `{{...}}` placeholders — see `plan.md` Decision D6 and `spec.md` §A blast radius.
- **Open clarifications**: none. Decision D5 (`SaveTemplateDefaults` base derivation) is explicitly DEFERRED with rationale in `plan.md` §E, per the SPEC's fourth stated requirement.
- **Plan-audit iter-1 (FAIL 0.71) revision (2026-08-03)**: D1–D6 + SHOULD-FIX D9–D13 applied. `version: "0.2.0"`, `updated: 2026-08-03`. Re-audit scoped to the 6 blocking items. Coverage baseline captured at plan time: **88.9%** (acceptance.md AC-UYP-020). Misnamed subtest at `update_yaml_test.go:591-603` was rewritten by an interim commit on `main` (the destructive `t.Errorf("expected user_added to be dropped…")` line is gone, subtest renamed to `"user added key not in base preserved"`); M4 therefore becomes a partial no-op + add the missing stderr-advisory sibling assertion (REQ-UYP-007). See `plan.md` M4 pre-flight note.
- **Status**: `draft` — awaiting re-audit and Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

Milestones M1–M7 executed sequentially (cycle_type=tdd) against worktree HEAD `ee1a0766d`. All verbatim commands run from the worktree root.

### Pre-flight (Section C)

```
$ grep 'gopkg.in/yaml.v3' go.mod
gopkg.in/yaml.v3 v3.0.1
$ go test ./internal/cli/... 2>&1 | tail -1
ok  	github.com/modu-ai/moai-adk/internal/cli	166.274s
$ go test -cover ./internal/cli/update/backup/
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	(cached)	coverage: 88.9% of statements
$ grep -rn 'expected user_added to be dropped' --include='*_test.go' .
(0 matches — destructive line already removed by interim main commit; M4 rename is a no-op, only the stderr-advisory sibling is added)
```

### M1 — node_merge.go primitives

Exit gate: `go test ./internal/cli/update/backup/ -run 'TestMergeMappingNode|TestNodeValuesEqual|TestEncodeNode|TestNodeMerge|TestDecodeDoc' -count=1` → all 14 tests PASS.

### M2 exit gate (build + vet)

```
$ go build ./...                          → exit 0
$ go vet ./internal/cli/...               → exit 0
```

### M3 — golden preservation test (AC-UYP-022 falsifiability)

RED (reverted map round-trip impl, via `git stash push -u` including untracked node_merge.go; intermediate-state assertion: merge.go reverted to HEAD map-form, node_merge.go absent):
```
$ go test ./internal/cli/update/backup/ -run TestPreserveGolden -count=1
# github.com/modu-ai/moai-adk/internal/cli/update/backup [...test]
internal/cli/update/backup/backup_error_test.go:101:17: undefined: deepMerge3WayTo
internal/cli/update/backup/backup_test.go:427:18: assignment mismatch: 2 variables but DeepMerge3Way returns 1 value
internal/cli/update/backup/backup_test.go:427:32: cannot use newN (variable of type *yaml.Node) as map[string]any value in argument to DeepMerge3Way
... too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup [build failed]
```
RED is a build failure (non-zero exit): the reverted impl removes the node-typed API the golden test is built on — structural evidence that the reverted impl lacks the preservation mechanism. `git stash pop` restored the new impl.

GREEN (node-tree impl):
```
$ go test ./internal/cli/update/backup/ -run TestPreserveGolden -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.242s
```

Gaps: the RED manifests as a build failure, not an assertion failure — the reverted impl removes the node-typed API the golden test depends on, so the test cannot compile, let alone assert. A build failure is a non-zero exit (satisfies AC-UYP-022's "FAIL (non-zero exit)"), but it does not directly demonstrate "comments are lost"; that direct demonstration is provided by AC-UYP-021 (end-to-end restore preserves comments on cache.yaml) and the green golden test (comment-line count preserved for all 30 templates under the node impl).

### AC binary PASS/FAIL matrix

| AC | Status | Verbatim command | Observed |
|----|--------|------------------|----------|
| AC-UYP-001 | PASS | `go test ./internal/cli/update/backup/ -run TestPreserveGolden_Comments -count=1 -v` | `--- PASS: TestPreserveGolden_Comments` |
| AC-UYP-002 | PASS | `go test ... -run TestPreserveGolden_KeyOrder` | `--- PASS: TestPreserveGolden_KeyOrder` |
| AC-UYP-003 | PASS | `go test ... -run TestPreserveGolden_Quoting` | `--- PASS: TestPreserveGolden_Quoting` |
| AC-UYP-004 | PASS | `grep -n 'SetIndent(2)' internal/cli/update/backup/node_merge.go` + `TestPreserveGolden_Indent` | `45: enc.SetIndent(2)`; `--- PASS: TestPreserveGolden_Indent` |
| AC-UYP-005 | PASS | `go test ... -run TestPreserveGolden_PropertySet` | `--- PASS: TestPreserveGolden_PropertySet`; byte-stable=5 byte-differ=25 (differ expected for reflow templates); cache.yaml asserted byte-identical |
| AC-UYP-006 | PASS | `go test ./internal/cli/ -run 'TestMergeYAML3Way/user_added_key_not_in_new_template_is_preserved'` | subtest PASS |
| AC-UYP-007 | PASS | `go test ... -run TestMergeYAML3Way_ReportsRetainedKey` | `--- PASS: TestMergeYAML3Way_ReportsRetainedKey` |
| AC-UYP-008 | PASS | `go test ... -run TestMerge3WayNotMoreDestructiveThan2Way` | `--- PASS` (3 subtests) |
| AC-UYP-009 | PASS | pre-existing decision tests (TestDeepMerge3Way, TestMergeYAML3Way) | all PASS |
| AC-UYP-010 | PASS | system-field tests (TestDeepMerge3Way "system field", TestMergeYAML3Way "system field version/template_version") | PASS |
| AC-UYP-011 | PASS | invalid-YAML error tests (TestMergeYAML3Way_InvalidNewData/OldData/BaseData) | PASS |
| AC-UYP-012 | PASS | `go test ... -run TestNodeMerge_AliasNotExpanded` | `--- PASS` |
| AC-UYP-013 | PASS | `go test ... -run TestNodeMerge_MergeKeyNotResolved` | `--- PASS` |
| AC-UYP-014 | PASS | `go test ... -run TestNodeMerge_SequenceReplaced` | `--- PASS` |
| AC-UYP-015 | PASS | `go test ... -run TestMergeYAML3Way_MultiDocumentErrors` | `--- PASS` |
| AC-UYP-016 | PASS | `go test ... -run TestNodeValuesEqual_NullVsEmptyMap` | `--- PASS` |
| AC-UYP-017 | PASS | `ls internal/template/templates/.moai/config/sections/ \| grep -cE '\.yaml(\.tmpl)?$'` = 30; `go test ... -run TestPreserveGolden` subtests = 30 | 30 == 30 |
| AC-UYP-018 | PASS | `grep -rn 'expected user_added to be dropped' --include='*_test.go' .` | 0 matches → PASS |
| AC-UYP-019 | PASS | `grep -rn 'SPEC-UPDATE-YAML-PRESERVE\|REQ-UYP-' internal/template/templates/` | 0 matches → PASS; `git status --porcelain internal/template/templates/` empty |
| AC-UYP-020 | PASS | `go test -cover ./internal/cli/update/backup/` | `coverage: 89.6% of statements` (≥ 88.9% baseline) |
| AC-UYP-021 | PASS | `go test ... -run TestUpdateEndToEnd_PreservesCustomizedSection` | `--- PASS` (cache.yaml + workflow.yaml subtests) |
| AC-UYP-022 | PASS | falsifiability RED-then-GREEN (captured above) | RED build-fail (non-zero); GREEN ok (zero) |
| AC-UYP-023 | PASS | `go test ... -run TestMergeYAML3Way_QualityTemplateParses` | `--- PASS`; map-decoder baseline errors (discriminator live) |

### E2 cross-platform build

```
$ go build ./...                            → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0
```

### E3 coverage

```
$ go test -cover ./internal/cli/update/backup/
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	coverage: 89.6% of statements
```

### E4 subagent-boundary grep

```
$ grep -rn 'AskUserQuestion' internal/cli/update/backup/ | grep -v _test.go
(0 matches → PASS)
```

### E5 lint

```
$ golangci-lint run ./internal/cli/update/backup/ ./internal/cli/
0 issues.
```

### E8 RED failure output (TDD falsifiability)

Captured verbatim in the M3 falsifiability block above (AC-UYP-022 step 3). The RED observation against the reverted map round-trip implementation is a build failure: the reverted `merge.go` removes the node-typed `DeepMerge3Way`/`deepMerge3WayTo` API the golden test and converted call sites depend on, so the test cannot compile. This is honest RED evidence — the reverted impl is structurally incapable of supporting the node-based preservation contract.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-03
run_commit_sha: pending-backfill-M-final
run_status: audit-ready
ac_pass_count: 23
ac_fail_count: 0
preserve_list_post_run_count: 3   # restore.go call sites (MergeYAML3Way/MergeYAMLDeep wrappers unchanged), ValuesEqual retained, internal/template/templates/ unmodified
l44_pre_commit_fetch: true
l44_post_push_fetch: pending
new_warnings_or_lints_introduced: none
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 8   # merge.go, node_merge.go (new), node_merge_test.go (new), preserve_golden_test.go (new), uyp_ac_test.go (new), node_merge_edge_test.go (new); revised: backup_test.go, backup_error_test.go, merge_useradd_test.go, update_yaml_test.go, update_test.go
m1_to_mN_commit_strategy: single-conventional-commit-per-logical-unit referencing #1243
```

Run-phase MUST ACs (001–023) all PASS. Coverage 89.6% ≥ 88.9% baseline. Lint clean. Cross-platform build clean. Template neutrality preserved. D5 (`SaveTemplateDefaults` base derivation) deferred to `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` per plan.md §E — not fixed inline (B10 scope discipline).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Decision: `sub-agent` (Mode 5)** — logged before first run-phase `Agent()` spawn per orchestration-mode-selection.md §D.

### Input parameters

| Parameter | Value |
|---|---|
| tier | M (spec frontmatter; 3 artifacts spec/plan/acceptance) |
| scope (file count) | 6-9 files (merge.go rewrite, node_merge.go new, 4-5 test files, preserve_golden_test.go new) |
| domain count | 1 (Go backend — yaml merge internal mechanism, single package `internal/cli/update/backup`) |
| file language mix | 100% Go (`.go` production + `_test.go`) |
| concurrency benefit | LOW — coding-heavy work (representation refactor + test contract), not research |

### Mode evaluation

| Mode | Selected? | Rationale |
|---|---|---|
| 1 trivial | no | 6-9 files, 300-1000 LOC, test contract to build — not a single-line fix |
| 2 background | no | write-capable implementation work, not read-only async analysis |
| 3 agent-team | no | RETIRED (Agent Teams static layer) |
| 4 parallel | no | single domain + coding-heavy → violates Anthropic coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research") |
| 5 sub-agent | **yes** | coding-heavy single-domain work; sequential milestone delegation (M1→M7); Mode 5 is the safe default for coding work per the decision tree |
| 6 workflow | no | not ≥30 files, not a single uniform mechanical transform — node-tree merge has inter-step data dependencies (M1 primitives → M2 rewire → M3 contract depends on M1/M2) |

### Justification

Mode 5 (sub-agent, sequential) is the documented default for coding-heavy work per the Phase 4 decision tree §B tie-breaker ("Coding-heavy + multi-domain: prefer Mode 5 over Mode 4"). This SPEC is a single-domain representation refactor (`internal/cli/update/backup`) with a strict milestone ordering: M1 node primitives → M2 rewire entry points → M3 preservation test contract → M4 key-policy reversal → M5 assertion-table revision → M6 verification sweep → M7 commit. Each milestone's exit depends on the prior milestone's artifacts (e.g., M3's golden test asserts M1/M2's node-tree output), so parallelism would create coordination overhead with no offsetting benefit. A single `manager-develop` sub-agent drives the milestones sequentially, returning a §E self-verification matrix at completion.

### Pre-spawn state

- Branch: `feat/SPEC-UPDATE-YAML-PRESERVE-001` (rebased onto `7cbe6c51e` origin/main; HEAD `ee1a0766d`)
- Working tree: clean
- Race absorbed: origin/main advance `7cbe6c51e` (SPEC-HOOK-TRACE-FLUSH-001, scope non-overlapping) rebased under this branch's 2 plan-phase commits
- Plan-audit: iter-2 PASS 1.00, skip-eligible (Phase 1 re-execution skipped; Implementation Kickoff Approval obtained)
- Implementation Kickoff Approval: PASSED (user selected "run-phase 진입 (권장)")
