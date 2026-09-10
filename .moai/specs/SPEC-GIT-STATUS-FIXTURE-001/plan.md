# plan.md — SPEC-GIT-STATUS-FIXTURE-001 (card t474)

Branch: `WT-git-status-fixture` · Base: `25a3212a9` (origin/develop tip) · Tier S · 2-3 line fixture fix. Resist inflation.

## §A Context

Develop CI is dual-job red on two `internal/core/git` tests (`TestStatusAheadBehindFromHeader`, `TestStatusBranchHeaderShapes`); both are green on dev machines. The pre-plan probe (`.moai/reports/t474/repro-probe.md`) established the single root cause: unpinned `git init -q --bare` at `status_branch_test.go:99` and `:235` leaves the bare remote's HEAD symref at the ambient `init.defaultBranch` value — `main` where Xcode's system gitconfig pins it (local green), `master` on CI (git compiled default) → unborn-HEAD clones → wrong-branch pushes that exit 0 → the exact three CI failure signatures. The identity hypothesis was measured and eliminated (gitFixture injects `GIT_AUTHOR_*`/`GIT_COMMITTER_*`, `status_branch_test.go:22-27`). The repair precedent is in-repo: `helpers_test.go:42` (`init --bare -b main`).

## §B Known Issues

- Local RED is unachievable in the lane (probe L7: worktree-scoped config forcing measured ineffective; system/global config and env-var forms rejected for lane safety). Verification leans on grep-able pin assertions + full-package local sweep + the merged-tree CI verdict.
- If GitHub's runner image ever pins `init.defaultBranch=main` system-wide, an unfixed fixture would go silently green-but-config-dependent — the pin removes that dependence (durable repair, probe Residual-risk).
- `status_branch_test.go:315` (`TestNewRepositoryErrorTaxonomy`) carries the same unpinned shape but is benign today — nothing clones from that bare repo; HEAD value cannot affect its assertions (probe L6).

## §C Pre-flight (run-phase entry checks)

1. `git rev-parse --short HEAD` = `25a3212a9` on `WT-git-status-fixture`; `git status --porcelain` clean of foreign modifications.
2. Read the kickoff decision on the optional `:315` pin (§F M1 gate) — accept or drop, recorded explicitly.
3. Confirm the baseline: `grep -n '"--bare"' internal/core/git/status_branch_test.go` shows `:99` and `:235` unpinned (the RED-now state, AC-GSF-001).

## §D Constraints

- Diff is additive only: insert `"-b", "main",` into the two (or three, if accepted) `gitFixture` bare-init calls. No other edits to the file.
- No commits, no pushes from the lane — the lane orchestrator commits after review; develop push is the lead's batch (repo GitFlow discipline).
- No CI workflow files, no runner config, no production paths.
- Lane-local verification only (`go test ./internal/core/git/...`, `gofmt -l .`); no full-suite local run (load discipline).

## §E Self-Verification (run-phase exit checks — map to spec.md §3 ACs)

1. AC-GSF-001: `grep -n '"init", "-q", "--bare", "-b", "main"' internal/core/git/status_branch_test.go` matches `:99` and `:235` (and `:315` iff accepted). Record verbatim output. Mutant probe: temporarily revert one pin → grep must show the miss → restore. A grep that always passes regardless of the pin is vacuous and does not count.
2. AC-GSF-002: `go test ./internal/core/git/ -count=1` → `ok` line, plus a visible non-zero test count (e.g. `-v 2>&1 | grep -c '^--- PASS:'` with the number recorded).
3. AC-GSF-004: `git diff --name-only 25a3212a9...HEAD` → exactly `internal/core/git/status_branch_test.go`.
4. AC-GSF-005: `gofmt -l .` → 0 rows (empty output, exit 0).
5. AC-GSF-003: EXTERNAL — report as pending-external in the run-phase record; the deciding verdict is the merged-tree CI after the lead's push. The card does not close before that verdict.

## §F Milestones (priority-ordered; kickoff-decision first)

- **M1 — Kickoff gate: decide the optional `:315` pin** (the only decision with reviewer discretion; everything else is mechanical). Surface to the Implementation Kickoff Approval: accept (consistency: all bare inits in the file pinned, +1 line) or drop (minimal diff, `:315` byte-unchanged; benign per probe L6). Record the decision in the run-phase progress record. Either answer keeps AC-GSF-006 satisfiable.
- **M2 — The repair + local verification.** Insert `"-b", "main",` at `:99` and `:235` (and `:315` iff M1 accepted). Run §E checks 1-4 in one batch. Evidence to the progress record.
- **M3 — External CI confirmation (lead-owned).** Merge to develop per GitFlow; the lead's batch push triggers the deciding CI run. Green on `Test (ubuntu-latest)` + `Race Test` closes AC-GSF-003 and the card. Red → re-open with the CI log as new evidence (do not claim local equivalence).

## §G Anti-Patterns

- Bundling the `:315` pin silently instead of gating it (violates AC-GSF-006).
- "Tests pass locally" offered as fix evidence — locally they never failed (probe L0); the discriminating environment is CI.
- Expanding into a fixture-helper refactor — scope discipline; the 2-line diff is the whole repair.
- Touching CI workflow config as a workaround — leaves the defect latent.

## §H Cross-References

- spec.md (this directory) — requirements + ACs (Tier S inline).
- `.moai/reports/t474/repro-probe.md` — measured causal chain (L0-L7), the evidence base for every AC's RED-now cell.
- `internal/core/git/helpers_test.go:42` — in-repo pinned-pattern precedent.
- Repairs the develop CI red relayed in the lead's t474 dispatch (3 failure lines, 2 tests).
