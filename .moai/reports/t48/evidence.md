# t48 Evidence — Release PR CI duplication cleanup (PR #1564 follow-up)

Branch: `WT-t48` (on top of `origin/release/v3.1.1`, merge commit `09516ff0c`)
Files touched: `.github/workflows/ci.yml`, `.github/workflows/release-pr-multi-os.yml`
Commits: `25f59ff49` (item 1), `20e7f04e3` (item 2)

## (a) Measurement gate — required-contexts, observed verbatim

Command:

```bash
gh api repos/modu-ai/moai-adk/branches/main/protection --jq '.required_status_checks.contexts'
```

Observed output (2026-08-17, re-run at final state after both edits):

```
["Test (ubuntu-latest)","Lint","Build (linux/amd64)","Analyze (Go) (go)","Release PR Multi-OS Gate"]
```

Asymmetry check demanded by the card:

- `Race Test` — **ABSENT** from required contexts (advisory). Matches card premise.
- `Release PR Multi-OS Gate` — **PRESENT** (required). Matches card premise.

The card's premise was confirmed by measurement before any job was disabled, so
the guard rail against a repeat of the #1564 opposite mistake (matrix ubuntu
removal) held. No contradiction; proceeded.

## (b) Diff summary + rationale for the exact condition expressions

### Item (1) — `.github/workflows/ci.yml` `test-race` job

```diff
   test-race:
     name: Race Test
     needs: detect
-    if: needs.detect.outputs.go_code == 'true'
+    if: needs.detect.outputs.go_code == 'true' && !startsWith(github.head_ref, 'release/')
```

Expression rationale:

- `startsWith(github.head_ref, 'release/')` is the repo's existing idiom —
  `.github/workflows/release-pr-multi-os.yml` line 35 already uses exactly
  `startsWith(github.head_ref, 'release/')` (positive form) to detect release
  PRs. The negation `!startsWith(...)` is the exact complement, so both files
  classify the same set of PRs as "release".
- Placement of `!` mid-scalar: the `if:` value begins with
  `needs.detect.outputs.go_code == 'true' &&`, so the plain YAML scalar never
  starts with `!` (which would be parsed as a YAML tag). Verified by python3
  YAML parse — the value round-trips as one plain scalar (see (c)).
- `github.head_ref` is set only on `pull_request` events; on `push` to main it
  is empty, so `!startsWith('', 'release/')` is true and the job still runs on
  push-to-main (behavior preserved — that path keeps Race Test as its only
  race run).

Comment block above the job extended to document the skip and the push-event
case.

### Item (2) — `.github/workflows/release-pr-multi-os.yml` paths-filter

Folded dorny/paths-filter (same pinned SHA as ci.yml:
`dorny/paths-filter@ceb8a2b8f2d89434be7ff52d3de7ec3738c5cc9d # v4.0.3`) into
the existing `detect-release` job instead of adding a separate detect job:

```yaml
    outputs:
      proceed: ${{ steps.gate.outputs.proceed }}
      go_code: ${{ steps.f.outputs.go_code }}
```

```diff
   full-matrix-test:
     needs: detect-release
-    if: needs.detect-release.outputs.proceed == 'true'
+    if: needs.detect-release.outputs.proceed == 'true' && (github.event_name == 'workflow_dispatch' || needs.detect-release.outputs.go_code != 'false')
```

Expression rationale:

- `go_code != 'false'` (not `== 'true'`) is deliberate: the matrix is skipped
  ONLY on a positively observed docs-only diff. If the filter step cannot
  evaluate (empty output), the comparison is true and the matrix RUNS — the
  failure direction is fail-safe for a required gate. Reinforced by
  `continue-on-error: true` on the paths-filter step so an abnormal filter
  failure cannot fail `detect-release` and silently skip the matrix through
  need-propagation.
- `github.event_name == 'workflow_dispatch'` bypass preserves the manual
  dispatch's run-anytime semantics (today dispatch always runs the matrix;
  `detect-release` itself ORs the same condition). Verified from
  dorny/paths-filter v4.0.3 source that the action does not error on
  workflow_dispatch (falls through to git diff vs default branch), but the
  bypass makes the filter result irrelevant on that path regardless.
- Folding into `detect-release` (which is already skipped on regular PRs)
  adds zero jobs on regular PRs, unlike a standalone detect job.
- Filter set mirrors ci.yml's `go_code` filter with the workflow self-reference
  swapped (`release-pr-multi-os.yml` instead of `ci.yml`), so edits to this
  workflow always exercise the matrix.
- Measured against a real release PR (#1555, `release/v3.1.0`): 40 files, 26
  `.go` — a real release PR trips the filter and still runs the matrix. Only
  genuinely docs-only release PRs skip it (card's low-frequency assessment).

Comment blocks updated where they described the old behavior (the
full-matrix-test gating description, the ci.yml↔this-workflow race-test
interplay, and the gate's skip note). No other content touched.

## (c) Verification — commands + verbatim output

```
$ actionlint .github/workflows/ci.yml .github/workflows/release-pr-multi-os.yml
actionlint exit: 0
```

```
$ gh api repos/modu-ai/moai-adk/branches/main/protection --jq '.required_status_checks.contexts'
["Test (ubuntu-latest)","Lint","Build (linux/amd64)","Analyze (Go) (go)","Release PR Multi-OS Gate"]
```

```
$ python3 (yaml.safe_load both files; assert test-race.if, full-matrix-test.if/needs, gate.if)
parsed if -> "needs.detect.outputs.go_code == 'true' && !startsWith(github.head_ref, 'release/')"
full-matrix-test if   -> "needs.detect-release.outputs.proceed == 'true' && (github.event_name == 'workflow_dispatch' || needs.detect-release.outputs.go_code != 'false')"
gate if               -> 'always()'

scenario matrix (True = matrix runs):
  False  regular PR (detect-release skipped -> proceed empty)
  True   release PR, go-code diff
  False  release PR, docs-only diff
  True   release PR, filter failed (go_code empty)
  True   manual workflow_dispatch
all 5 scenario assertions PASS
```

Per lane discipline, no local `go test ./...` was run — full-matrix judgment
belongs to CI on the PR head.

## (d) Residual risks

1. **Race coverage on release/* PRs now rides solely on the required gate.**
   ci.yml's Race Test no longer runs there. This is covered:
   `release-pr-multi-os.yml` `full-matrix-test` runs
   `go test -race -timeout 25m ./...` on ubuntu-latest (plus macOS/Windows)
   and reports through `Release PR Multi-OS Gate`, which IS required — a race
   failure there blocks the release. Differences between the two runs: the
   ci.yml job used `-count=1` (no cache) while the gate omits it (may reuse
   cached results within the same run — cross-run cache does not apply to race
   results differently); both install the same pinned ast-grep. The ubuntu leg
   of the required gate is a strict-equivalent race run, so no blocking
   coverage is lost — only the advisory duplicate.
2. **push-to-main race runs unchanged** (empty head_ref keeps the job), so
   post-merge main still gets an advisory race run per go-code push.
3. **Docs-only release PRs skip the 3-OS matrix** (item 2's intent). Residual:
   a release PR whose diff touches only files outside the filter (e.g.
   template content files under `internal/template/templates/**`, docs) would
   skip the matrix, exactly as ci.yml's detect already fast-tracks such PRs —
   a pre-existing accepted tradeoff mirrored here, and measured real release
   PRs (#1555) are .go-heavy so they run it.
4. **First release PR after this change is the live test** of both conditions
   (CI owns final judgment). Watch that "Race Test" reports as skipped (not
   pending-forever — it is not a required context, so no branch-protection
   wait is possible) and that "Release PR Multi-OS Gate" still reports on
   every PR.
5. dorny/paths-filter on workflow_dispatch was verified from source to fall
   through to a git diff (not error), but the dispatch bypass clause makes the
   filter result unused on that path, so this is belt-and-suspenders only.

## (e) Item (2) status

**DONE.** Item (1) was complete and cleanly verified (actionlint + YAML parse
+ required-contexts measurement) before item (2) started, satisfying the
card's conditional. Item (2) mirrors ci.yml's paths-filter idiom (same action,
same pin, same `go_code` filter shape) adapted to this workflow's existing
gating job, with a fail-safe comparison and a dispatch bypass as described in
(b).

## Rework round 1 (review verdict FAIL — F1/F2/F4 addressed)

### F1 [BLOCKING] — required gate could pass with zero verification (FIXED)

Reviewer-confirmed failure chain in my round-1 YAML: `detect-release` gained
failable steps (`actions/checkout@v7` without continue-on-error) → a checkout
failure fails `detect-release` → `full-matrix-test` (no status function in its
`if`) is skipped → `release-pr-gate` (`if: always()`) aggregated only
`full-matrix-test.result == "failure"`, and a skipped dependency reads as
SUCCESS. This contradicted my own comment claiming an abnormal filter can
never silently skip a required-gate verification. The path did not exist
before round 1 (detect-release previously had only the gate-echo step).

Fix applied exactly as prescribed (minimal):

1. `release-pr-gate.needs` is now `[full-matrix-test, detect-release]`.
2. Aggregation extended: `needs.detect-release.result == 'failure'` fails the
   gate with its own `::error` message, checked before the matrix check.
3. Normal-PR path preserved: only a hard `failure` blocks; `skipped`
   (ordinary PR flow, where detect-release's own `if` is false) still passes.
   Gate comment documents the skipped-vs-failed distinction.

```diff
   release-pr-gate:
     name: Release PR Multi-OS Gate
-    needs: [full-matrix-test]
+    needs: [full-matrix-test, detect-release]
     if: always()
       - name: Verify all OS legs passed
         run: |
+          if [ "${{ needs.detect-release.result }}" = "failure" ]; then
+            echo "::error::Release PR detection FAILED before the matrix could run — not treating a skipped matrix as verification"
+            exit 1
+          fi
           if [ "${{ needs.full-matrix-test.result }}" = "failure" ]; then
```

### F2 — filter parity claim (FIXED by adding)

`.github/workflows/codeql.yml` added to the `go_code` filter, so it now
exactly mirrors ci.yml's filter with only the workflow self-reference swapped
(`release-pr-multi-os.yml` for `ci.yml`). The round-1 "mirrors ci.yml's filter"
claim is now true rather than corrected-away.

### F4 — evidence typo (FIXED)

Two occurrences of `Build (linux-amd64)` corrected to `Build (linux/amd64)`
(job-name form, matching the actually observed gh api output).

### Rework verification — commands + verbatim output

```
$ actionlint .github/workflows/ci.yml .github/workflows/release-pr-multi-os.yml
actionlint exit: 0
```

Extended scenario matrix (models need-propagation, `full-matrix-test`'s `if`,
and the gate's two-check aggregation):

```
gate needs -> ['full-matrix-test', 'detect-release'] | gate if -> 'always()'

scenario matrix (F1 cases included):
  matrix=skipped  gate=SUCCESS                            1  normal PR (detect-release skipped)
  matrix=success  gate=SUCCESS                            2  release PR, go-code diff, matrix passes
  matrix=failure  gate=FAILURE (matrix failed)            3  release PR, go-code diff, matrix FAILS
  matrix=skipped  gate=SUCCESS                            4  release PR, docs-only diff
  matrix=success  gate=SUCCESS                            5  release PR, filter unevaluable (go_code empty)
  matrix=skipped  gate=FAILURE (detect-release failed)    6  release PR, detect-release CHECKOUT FAILURE
  matrix=success  gate=SUCCESS                            7  manual workflow_dispatch

all 8 assertions PASS
```

New-case verdicts demanded by the rework instruction:

- (a) detect-release checkout failure → gate **FAILS** (case 6; was SUCCESS
  before the fix).
- (b) detect-release skipped (normal PR) → gate **SUCCESS preserved** (case 1).

All five round-1 scenarios are unchanged (cases 1-5, 7), so no regression in
the round-1 behavior. Note: the round-1 scenario table simulated only
`full-matrix-test`'s `if`; the rework table simulates the full pipeline
through the gate aggregation — the round-1 table alone could not have caught
F1, which is why the reviewer's chain analysis in the YAML was the decisive
evidence.

Gaps: the pipeline model is a simulation of GitHub Actions need-propagation
semantics, not an executed workflow run; live confirmation lands on the first
release PR + the first detect-release infra failure (unobservable on demand).
No local `go test ./...` (lane discipline).

