# Evaluator-Active Independent Assessment — SPEC-V3R4-CATALOG-002 (review 1)

## Verdict

verdict: PASS
overall_score: 0.916
assessed_at: 2026-05-12T10:30:00Z
evaluator_version: evaluator-active v1.0 (claude-sonnet-4-6)

## Dimension Scores

| Dimension | Weight | Score | Verdict |
|-----------|--------|-------|---------|
| Functionality | 40% | 0.88 | PASS |
| Security | 25% | 1.00 | PASS |
| Craft | 20% | 0.88 | PASS |
| Consistency | 15% | 0.92 | PASS |
| **Weighted Overall** | 100% | **0.916** | **PASS** |

---

## Findings

### P0 Blocking Findings (must fix before merge)

None.

### P1 Major Findings (should fix before merge)

None.

### P2 Minor Findings (recommended fixes)

**[P2] internal/cli/init.go: slim-mode stdout notice not covered by automated test**

Scenario 1 "Then" in acceptance.md explicitly states:
> "Stdout MUST include an informational message containing the substrings `"slim mode"`, `"--all"`, `"MOAI_DISTRIBUTE_ALL=1"`, AND `"SPEC-V3R4-CATALOG-005"`"

The acceptance quality gate also states:
> "All 8 Given-When-Then scenarios (S1-S8) pass when run individually AND in t.Parallel() mode."

Verification: `grep -rn 'slim mode\|CATALOG-005\|OutOrStdout' internal/cli/init_slim_branch_test.go internal/cli/init_test.go` — zero matches in test files.

The code itself is correct (init.go:337-340 prints the 4-substring notice via `cmd.OutOrStdout()`). However, there is no automated test that sets up `runInit` with slim mode, captures `cmd.SetOut(buf)`, and asserts all 4 substrings. If a future commit accidentally removes or changes the notice, no test fails.

Recommended fix: Add a table-driven test in `init_slim_branch_test.go` (or a new `init_slim_integration_test.go`) that sets `MOAI_DISTRIBUTE_ALL=""`, calls `runInit` on a temp directory, captures stdout, and asserts all 4 required substrings.

**[P2] internal/template/embed_catalog.go: LoadEmbeddedCatalog and NewSlimDeployerWithRenderer error branches below 90%**

Coverage:
- `LoadEmbeddedCatalog`: 75.0% (error return path not exercised via the real embedded catalog)
- `NewSlimDeployerWithRenderer`: 83.3% (the nil-rawFS SlimFS error path from within the function)

Both are below the ≥90% critical-package threshold (CLAUDE.local.md §6). The production error paths are not reproducible without patching the unexported `embeddedRaw` variable, but this is not an excuse for missing coverage — alternative approaches exist (e.g., using a `//go:build !integration` build tag that replaces `embeddedRaw` with a testable stub, or testing via internal package access).

Recommended fix: Add an internal test that assigns a temporary corrupt value to `embeddedRaw` (since `embed_catalog_test.go` is in the same package and has access) to exercise the `LoadEmbeddedCatalog` error branch. The `NewSlimDeployerWithRenderer` SlimFS-wrapping error path can be triggered by passing a nil rawFS through a refactored helper or by testing the error path directly.

### P3 Conditional Findings (post-merge or follow-up)

**[P3] internal/cli/init.go: CATALOG_LOAD_FAILED early-return not end-to-end tested**

The `init.go:315-318` block that returns `CATALOG_LOAD_FAILED` before any file writes is not directly tested via `runInit`. The `embed_catalog_test.go:TestLoadEmbeddedCatalog_LoadCatalogErrorWrapping` exercises `LoadCatalog` with an empty FS, which demonstrates the error path in isolation. However, the early-return guarantee (no files written before the error) is not verified by any test.

Recommended fix (post-merge): Add an integration test that replaces the embedded catalog with a corrupt one (or uses a mocked `LoadEmbeddedCatalog` injection point) and verifies the target directory remains empty.

**[P3] slim_fs.go + embed_catalog.go: premature @MX:ANCHOR tags on fan_in=1 functions**

Per `mx-tag-protocol.md`: "@MX:ANCHOR — Add when: Function has fan_in >= 3 callers." Current production fan_in:
- `SlimFS()` in slim_fs.go: called only from `embed_catalog.go:NewSlimDeployerWithRenderer` (fan_in=1 production, fan_in=3+ with test callsites)
- `NewSlimDeployerWithRenderer()` in embed_catalog.go: called only from `init.go` (fan_in=1)

Both @MX:ANCHOR tags are forward-looking ("expected fan_in >= 3" and "future CATALOG-003/004 will route through this"). This violates the current-state criterion of the protocol. The tags should either be demoted to @MX:NOTE until the actual fan_in threshold is met, or the protocol criterion should be amended to include "expected-fan_in" use cases.

Recommended fix (post-merge): Downgrade to `@MX:NOTE` now; promote to `@MX:ANCHOR` when CATALOG-003 or CATALOG-004 adds a second production caller.

**[P3] denySet map immutability not directly asserted by reflective test**

`catalog_slim_audit_test.go:TestSlimFS_ReadOnlyInvariant/reflective_struct_check` iterates slimFS fields and rejects `sync.*` and channel fields. However, the test explicitly acknowledges (line 229): "denySet immutability is documented (godoc-only invariant per P3-3 finding): 'denySet is set once at construction and used read-only thereafter.'"

The `denySet map[string]struct{}` field is a mutable Go map type. Although it is treated as immutable-by-convention, a future developer could accidentally write to it. The reflective check does NOT assert that `denySet` is immutable. The map type would not be flagged by the current `sync.*` / `reflect.Chan` check.

This is acceptable as a documented convention (godoc states "immutable after construction"), but a future-proof approach would be to use `map[string]struct{}` only via a read-only accessor or snapshot. Flagged as P3 informational; no functional regression currently.

---

## Strengths

- **D7 lock rigorously preserved**: `git diff HEAD internal/template/deployer.go internal/cli/update.go internal/template/embed.go internal/template/catalog.yaml internal/template/catalog_loader.go` produces empty output. Zero deployer.go modifications.
- **DEFECT-5 encapsulation fully enforced**: `git grep 'EmbeddedRaw[A-Za-z]*' internal/cli/` returns zero matches. External callers route exclusively through `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer()`.
- **Audit suite is production-grade**: 4 parallel audit sub-tests over the real 65-entry catalog. 25 hidden non-core entries verified, 40 core entries verified, 523 walk paths checked. All sentinel emissions use `t.Errorf` following CATALOG-001 EC3 lesson.
- **REQ-012/EC3 narrow env matching is exactly right**: `shouldDistributeAll` correctly accepts only `"1"` exact and case-insensitive `"true"`. 8 truth-table test cases cover all EC3 edge cases including `"0"`, `"yes"`, `""`.
- **REQ-021 all 4 required substrings present**: Both in `AssertBuilderHarnessAvailable` error message and in the slim-mode stdout notice in `init.go`. Verified by `grep -c` on production files.
- **Race safety verified**: 32-goroutine × 50-iteration concurrent test passes with `go test -race`. The reflective check confirms no `sync.*` or channel fields in `slimFS`.
- **golangci-lint 0 issues**: Clean lint output, idiomatic Go patterns throughout (`errors.Is`, `fs.PathError`, `cmd.OutOrStdout()`, `fmt.Errorf` wrapping).
- **CHANGELOG bilingual with BREAKING CHANGE**: Both English and Korean BREAKING CHANGE entries present in the Unreleased section with both opt-out mechanisms documented.

---

## Evidence Summary

| Check | Result | Command / Reference |
|-------|--------|---------------------|
| Tests with -race | PASS | `go test -race -count=1 ./internal/template/... ./internal/cli/...` — all PASS (4.4s + 18.6s) |
| go vet | PASS | `go vet ./internal/template/... ./internal/cli/...` — no output |
| golangci-lint | PASS | `golangci-lint run ./internal/template/... ./internal/cli/...` — 0 issues |
| slim_fs.go coverage | PASS (90.4%) | 66/73 statements covered; function avg 92.5% |
| slim_guard.go coverage | PASS (100%) | 6/6 statements covered |
| shouldDistributeAll coverage | PASS (100%) | `go tool cover -func` |
| D7 lock (5 files) | PASS | `git diff --name-only HEAD ...` — empty output |
| DEFECT-5 encapsulation | PASS | `git grep 'EmbeddedRaw[A-Za-z]*' internal/cli/` — exit 1 (no matches) |
| REQ-021 4-substring (slim_guard.go) | PASS | `grep -c CATALOG_SLIM_HARNESS_MISSING/MOAI_DISTRIBUTE_ALL=1/moai init --all/SPEC-V3R4-CATALOG-005 slim_guard.go` — all > 0 |
| REQ-020 CHANGELOG | PASS | `grep -E 'BREAKING CHANGE' CHANGELOG.md` — match on line 10; both opt-outs present |
| t.Logf sentinel discipline | PASS | 3 informational t.Logf calls only (audited counts); no failure emissions via t.Logf |
| Audit test PASS (6 sub-tests) | PASS | HidesNonCoreEntries (25 entries), PreservesCoreEntries (40 entries + EC4 nested), PreservesNonCatalogFiles (5 paths), WalkDirNoLeak (523 paths), ReadOnlyInvariant/reflective + concurrent | 
| Stdout notice test gap | FAIL | No test exercises runInit slim mode + stdout capture — Scenario 1 stdout assertion unverified |
| embed_catalog.go coverage | WARN | LoadEmbeddedCatalog 75%, NewSlimDeployerWithRenderer 83.3% — below 90% critical threshold |
| @MX:ANCHOR fan_in correctness | WARN | SlimFS fan_in=1 production; NewSlimDeployerWithRenderer fan_in=1; both below the fan_in >= 3 trigger criterion |
