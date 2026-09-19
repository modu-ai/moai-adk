# progress.md — SPEC-DOCTOR-PLUGIN-DIGEST-PREFILTER-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-19
tier: S
artifacts: spec.md, plan.md, progress.md
card: t963

## §E.2 Run-phase Evidence

Tree: `WT-doctor-home-scan` @ `4a97dd2bf`, `develop...HEAD` = `0 0`.
Implementation by manager-develop (cycle_type=tdd); every figure below was
re-observed by the lane itself, not carried over from the agent's report.

### AC matrix

| AC | Verdict | Evidence |
|---|---|---|
| AC-DPP-001 | PASS | `TestPluginDigestCandidatesAllDistinctYieldsNone` |
| AC-DPP-002 | PASS | `TestPluginDigestCandidatesSharedPairSurvives` |
| AC-DPP-003 | PASS | `TestPluginDigestCandidatesMixedIsSortedAndDeterministic` |
| AC-DPP-004 | PASS | `TestPluginHashDoesNotEquateSameSizeDifferentContent`, file byte-unchanged |
| AC-DPP-005 | PASS | no wall-clock assertion in the added diff; report states both terms |

### Targeted run (lane's own re-execution, `.moai/reports/t963/targeted-run.log`)

    go test ./internal/cli -run 'TestPluginDigestCandidates|TestPluginHashDoesNotEquateSameSizeDifferentContent' -count=1 -v -timeout 30m

    === RUN   TestPluginHashDoesNotEquateSameSizeDifferentContent
    --- PASS: TestPluginHashDoesNotEquateSameSizeDifferentContent (0.00s)
    === RUN   TestPluginDigestCandidatesAllDistinctYieldsNone
    --- PASS: TestPluginDigestCandidatesAllDistinctYieldsNone (0.00s)
    === RUN   TestPluginDigestCandidatesSharedPairSurvives
    --- PASS: TestPluginDigestCandidatesSharedPairSurvives (0.00s)
    === RUN   TestPluginDigestCandidatesMixedIsSortedAndDeterministic
    --- PASS: TestPluginDigestCandidatesMixedIsSortedAndDeterministic (0.00s)
    PASS
    ok  github.com/modu-ai/moai-adk/internal/cli 0.636s

Four `=== RUN` and four matching `--- PASS` — the filter matched a non-empty set,
so exit 0 is not an empty result set reading as success.

### Non-vacuity (mutation check, run by manager-develop)

With `pluginDigestCandidates` forced to `return nil`: AC-DPP-002, AC-DPP-003 and
AC-DPP-004 all turn red; AC-DPP-001 stays green. That split is the point —
the negative control alone is satisfied by `return nil` and is therefore not
evidence on its own. Mutation reverted before the final run above.

### Toolchain

- `go vet ./internal/cli` → empty output, exit 0
- `gofmt -l internal/cli/doctor_disk.go internal/cli/doctor_disk_test.go` → empty
  output (judged by emptiness; `gofmt -l` prints findings and still exits 0)
- `git diff -U0 internal/cli/doctor_disk_test.go | grep '^-[^-]'` → empty, so the
  AC-DPP-004 guard test carries zero deleted lines

### Effect (motivating evidence, NOT an assertion — `.moai/reports/t963/toggle-diff.log`)

Same-tree toggle on `TestRunDoctor_WithFix` against the real `~/.moai` (14G):

| Arm | Fix | CPU idle | Test time |
|---|---|---|---|
| A1 | yes | 18.62% | 13.55s |
| B  | no  | 73.88% | 15.32s |
| A2 | yes | 54.97% |  8.72s |
| B2 | no  | 57.59% | 16.05s |

Matched pair A2/B2 → **7.33s removed**, coherent with the separately measured
5.73s plugins content read plus sha256 compute. Cross-check: the most-contended
fix arm (A1) still beats the least-contended control arm (B), so the effect is
not explained by machine load.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-19
ac_pass: 5/5
production_diff: internal/cli/doctor_disk.go +30/-5
test_diff: internal/cli/doctor_disk_test.go +61/-0
evidence_dir: .moai/reports/t963/

### Gaps — explicitly not observed

- Full `internal/cli` package verdict. Only targeted runs were executed, per the
  repo's standing rule (a full local suite measures the machine, not the code).
  CI on the pushed head is the full-suite judge. `go vet ./internal/cli` did
  compile the whole package, so nothing in it fails to build.
- Cross-platform build. No `GOOS=windows` build was attempted. The change adds a
  map keyed by a comparable struct plus `sort.Strings` — no platform-dependent
  code — but that is reasoning, not a measurement.
- Runtime `moai doctor` invocation. Behavior was exercised through the test
  binary only.
- The zero-candidate outcome on this machine (five profiles, all distinct
  `(Size, Files)`) is a data state, not correctness evidence. Correctness rests
  on AC-DPP-001..004.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
