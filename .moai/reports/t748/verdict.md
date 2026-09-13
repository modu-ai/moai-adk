# verdict.md — card t748 (SPEC-CODEMAPS-FOLD-GUARD-001 run-phase)

## Claim

The fold-preservation guard landed in `internal/graph/codemaps_fold_guard_test.go` (test-only, zero production code changes): it derives the protected unit set from `fold-judgments.txt` with the five t475 floor units pinned (dual lock), fails naming unit+document on any exact-substring token reappearance in the five generator docs (with the `core/git` short form legal), fails on record loss/floor degradation and on missing/unreadable generator docs, is stamp-independent, and flows both the real-tree scan and all injected-fixture scans through the single `runFoldGuard` reader. All three RED failure paths were observed verbatim before restore. SPEC-CODEMAPS-FOLD-GUARD-001 AC-CFG-001..005: 5 PASS / 0 FAIL.

## Evidence

All commands run in `.claude/worktrees/t748` on branch `WT-codemaps-fold-ac`, HEAD `ffc872ae5` at evidence close.

RED proofs (verbatim from `.moai/state/verify/t748/m2_red_probes.log`; transient uncommitted probes, deleted after export):

```
=== RUN   TestZZRedProbeAInjectedToken
    zz_red_probe_test.go:23: guard reports: fold-unit prose leaked into generator documents:
        fold unit "internal/kanban/prlink_landedref.go" appears in modules.md:329
--- FAIL: TestZZRedProbeAInjectedToken (0.00s)
=== RUN   TestZZRedProbeBLostFloorLine
    zz_red_probe_test.go:23: guard reports: fold-judgments record lost floor unit line(s): internal/hook/quality/step_git_env.go
--- FAIL: TestZZRedProbeBLostFloorLine (0.00s)
=== RUN   TestZZRedProbeCMissingDoc
    zz_red_probe_test.go:23: guard reports: generator document missing or unreadable: data-flow.md: open /var/folders/.../codemaps/data-flow.md: no such file or directory
--- FAIL: TestZZRedProbeCMissingDoc (0.00s)
=== RUN   TestZZRedProbeRestore
--- PASS: TestZZRedProbeRestore (0.00s)
```

Restore/GREEN (committed fixture tests, `.moai/state/verify/t748/m2_restored_pass.log`, exit 0):

```
--- PASS: TestCodemapsFoldPreservationGuard (0.00s)
--- PASS: TestCodemapsFoldGuardFixtures (0.01s)
    --- PASS: TestCodemapsFoldGuardFixtures/intact_copy_passes
    --- PASS: TestCodemapsFoldGuardFixtures/short_form_core/git_is_not_a_violation
    --- PASS: TestCodemapsFoldGuardFixtures/injected_fold_token_fails_naming_unit_and_document
    --- PASS: TestCodemapsFoldGuardFixtures/lost_floor_line_fails_naming_the_unit
    --- PASS: TestCodemapsFoldGuardFixtures/missing_generator_document_fails_naming_the_document
    --- PASS: TestCodemapsFoldGuardFixtures/missing_record_fails
--- PASS: TestCodemapsFoldGuardFloorCoverage
ok  	github.com/modu-ai/moai-adk/internal/graph
```

Verification batch:

```
$ go vet ./internal/graph/                                   → VET_EXIT=0
$ go test ./internal/graph/ -count=1                          → ok ... 46.335s (PKG_TEST_EXIT=0)
$ golangci-lint run ./internal/graph/...                      → 0 issues. (LINT_EXIT=0)
$ go build ./...                                              → BUILD_EXIT=0
$ GOOS=windows GOARCH=amd64 go build ./...                    → WIN_BUILD_EXIT=0
$ git status --porcelain -- .moai/project/codemaps/ .moai/reports/t475/ .moai/reports/t747/
                                                              → (empty — tamper-0)
$ git diff --stat HEAD~1..HEAD   (128474b19 = M1)
  .moai/specs/SPEC-CODEMAPS-FOLD-GUARD-001/spec.md |  2 +-
  internal/graph/codemaps_fold_guard_test.go       | 341 ++++++
  2 files changed, 342 insertions(+), 1 deletion(-)
```

Commits on `WT-codemaps-fold-ac`: `128474b19` (M1 guard + spec.md draft→in-progress), `ffc872ae5` (M2 RED proofs + progress.md §E.2/§E.3), plus this verdict/backfill commit.

## Baseline-attribution

- Baseline re-measured at run start on this tree (`f600a1c5d`): guard file absent (ls EXIT=1), fold-token grep across the 5 generator docs = 0 hits (EXIT=1), `grep -c '^fold '` = 5 (EXIT=0), both depends_on `status: completed`, `go test ./internal/graph/` baseline `ok 77.345s`. All values match spec §A.2 (measured at plan time on `146faed9d`).
- RED/GREEN evidence: this run, this tree, outputs above; persisted under `.moai/state/verify/t748/`.
- Full-suite verdict is CI's (lead condition): local scope was the changed package only.

## Gaps

- AC-CFG-004's "no fold rows in `moai graph check` layer output" is covered indirectly (M1 diff scope = zero production files touched; the checker source was never modified) rather than by executing the checker CLI and diffing its layer output.
- The guard's protected set is derived from the record at each run; the record currently carries exactly the 5 floor units (verified in pre-flight). A future record with NEW fold units beyond the floor would be scanned data-driven but that wider-set path was not exercised by a dedicated fixture (the floor-degradation and token-injection fixtures cover the locked paths).
- No coverage percentage was measured for this test file (test-only change; no production statements added, so package coverage is unchanged).

## Residual-risk

- Detection limit accepted by spec REQ-CFG-002: a regeneration that rewrites fold-unit prose without ever naming the unit's path or filename token passes this guard. The incident's regenerator signature carried tokens (t475 §④-b:558), so the observed recurrence class is covered.
- The guard runs wherever `go test ./internal/graph/` runs; if a future change excludes the default test suite or the file, the guard stops firing without a signal (verification-completeness §1.3 continued-firing is owned by CI keeping the default suite intact).

🗿 MoAI
