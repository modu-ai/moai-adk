# t581 — lane verdict (t570 M-4: fromConfigPath seam-bypass mutant)

card: t581 (Class B, Tier S)
worktree: .claude/worktrees/t581
branch: WT-seam-bypass-mutant
base: c8203fbf3f3443feec351085e6a7d6910dbdba5a (local develop at dispatch)
commit under verdict: a2e7b9172 (guard test + reproduction evidence)
measured by: lane-5 orchestrator, directly — no implementing agent was used on this card

## Claim

1. **The defect was live on this tree.** With the seam bypass the card names injected —
   `strings.ReplaceAll(e.Path, "/", "\\")` in place of
   `fromConfigPath(e.Path, configPathSeparator)` at the `codexPathAbsolute` arm of
   `internal/cli/doctor_codex.go` — all three t570 guards passed. They pin the separator to a
   backslash and compare against a backslash literal, and under that pin the seam call and the
   bypass produce the same bytes. They pinned the converted VALUE, not that the seam was USED.
2. **The added guard is discriminating.** A slash-pinned run separates the two, because the seam
   is the identity there while the bypass still converts. The new identity guard fails under
   the bypass and passes on clean code.
3. **The guard's code path is reachable.** A separate reachability test passes in both states,
   so the identity guard's failure under the bypass is a value failure, never an unreached arm.

## Evidence

Both runs are recorded in full in `.moai/reports/t581/reproduction.md` (committed in
a2e7b9172): the exact commands, verbatim output, the load condition, and the slot census.
Summarised here without re-quoting the path literals.

| tree | reachability guard | identity guard | three t570 guards | run exit |
|---|---|---|---|---|
| seam bypass injected | PASS | **FAIL** — names the value difference | PASS (the defect) | `TEST_EXIT=1` |
| clean | PASS | **PASS** | PASS | `TEST_EXIT=0` |

- Both runs used one `-run` selection naming all five tests, one compile each, output redirected
  to file and the exit code read with no pipe. In both, the file was read and every test was
  confirmed by name — a selection matching nothing would also exit 0.
- Mutant revert verified: `git diff -- internal/cli/doctor_codex.go` printed nothing.
- Clean-tree run precondition, re-verified in the same invocation:
  `git diff --quiet -- internal/cli/doctor_codex.go` exited 0, HEAD `c8203fbf3`.
- Slot discipline: both runs were granted by the lead as single scoped `internal/cli` compiles.
  Before the clean-tree run, the lead's census showed exactly one `go` process on
  `internal/cli` (lane-6's full suite) and no other scoped compile.
- Concurrency condition recorded for both: lane-6's full `internal/cli` suite running alongside;
  load averages 11.02 / 11.05 / 12.22 at the first run, 9.27 / 11.15 / 10.69 at the second.
  Both are value measurements — which assertion fires on which bytes — so contention cannot
  flip the verdict.
- Before commit: `gofmt -l` on the new test file printed nothing; staging was by explicit
  pathspec only.

## Baseline-attribution

Every figure above was measured by the lane in this session, on this worktree, at HEAD
c8203fbf3 — the mutant run with exactly the one shown production-source mutation applied, the
clean-tree run with none. Nothing is carried over from t570's `m4-missed.log`, which records the
same shape on that card's tree.

## Gaps — not observed

- **Absorb and re-measure are not yet done.** Per the lead's revised rule, the lane does not run
  the full `internal/cli` suite: after absorbing local develop inside the integration window,
  the lane re-runs the same five tests with the same `-run` selection on the merged tree, then
  merges. The full suite, `go vet`, and `golangci-lint` run once at the lead's batch close on
  the develop tip.
- The new guard has not been exercised on any platform other than darwin. The reachability
  result rests on `filepath.IsAbs`, which is host-dependent.
- The four known-red `internal/cli` baseline tests the lead identified by name were never
  exercised — the `-run` selection does not match them, and the full suite is now the lead's.

## Residual-risk

The new guard pins the seam's identity **behaviour** under a slash separator; it does not pin
that `fromConfigPath` is **called**. A bypass that reproduces the seam's own separator branching
would pass both pins. Closing that needs a call-observing seam rather than a value comparison.
The lead issued this as card t608, whose first job is to inject that bypass and observe whether
both guards actually pass — the lane identified it by reading and did not reproduce it.

## Sync decision

**No CHANGELOG entry.** The card adds one test file and changes no production source, so there
is no behaviour change for a user to see. The lead may override.
