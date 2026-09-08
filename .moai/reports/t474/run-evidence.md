# Card t474 — Run-phase Verification Evidence

SPEC: SPEC-GIT-STATUS-FIXTURE-001 · Branch: `WT-git-status-fixture`
Baseline tree: pre-repair HEAD `45a089bee` (plan artifacts on top of base `25a3212a9`); repair measurements taken on the working tree of this worktree immediately after the edit, before the run-phase commit.
All commands this run, this worktree, unless noted.

## M1 — Kickoff decision (recorded verbatim from the dispatch)

> The operator ACCEPTED the optional `:315` consistency pin. So the repair touches THREE sites: `:99`, `:235`, AND `:315` — all in `internal/core/git/status_branch_test.go`, and that file only. Rationale recorded: `:315` is harmless today (nothing clones from it) but the same trap remains for the next fixture copied from it.

Implementation Kickoff Approval: APPROVED by the operator (relayed via the lead).

## M2 — Baseline (RED-now, pre-repair)

```
$ git rev-parse --short HEAD && git branch --show-current && git status --porcelain
45a089bee
WT-git-status-fixture
(clean — no output lines)
```

```
$ grep -n '"--bare"' internal/core/git/status_branch_test.go
99:		gitFixture(t, base, "init", "-q", "--bare", remote)
235:	gitFixture(t, base, "init", "-q", "--bare", remote)
315:	gitFixture(t, base, "init", "-q", "--bare", barePath)
```
← all three sites unpinned — the AC-GSF-001 RED-now state (probe L6). The pinned-form grep matches 0 of 3 lines here.

## M2 — The repair

Insert `"-b", "main",` into the three bare inits (additive only; nothing else changed).

## §E verification — each with command + verbatim output

### AC-GSF-001 — Pin present at all three sites (kickoff-accepted scope)

```
$ grep -n '"init", "-q", "--bare", "-b", "main"' internal/core/git/status_branch_test.go
99:		gitFixture(t, base, "init", "-q", "--bare", "-b", "main", remote)
235:	gitFixture(t, base, "init", "-q", "--bare", "-b", "main", remote)
315:	gitFixture(t, base, "init", "-q", "--bare", "-b", "main", barePath)
```
3/3 sites carry the pin, matching the `helpers_test.go:42` shape.

Count form:
```
$ grep -c '"init", "-q", "--bare", "-b", "main"' internal/core/git/status_branch_test.go
3
```

**Mutant probe** (plan §E.1 — revert `:99` pin → grep → restore):

```
--- MUTANT (:99 reverted) ---
$ grep -n '"init", "-q", "--bare", "-b", "main"' internal/core/git/status_branch_test.go
235:		gitFixture(t, base, "init", "-q", "--bare", "-b", "main", remote)
315:	gitFixture(t, base, "init", "-q", "--bare", "-b", "main", barePath)
```
← the miss is visible as 2/3 sites (line 99 absent). Note honestly: `grep -n` still exits 0 here because 2 matches remain — the criterion's asserting form is the **match count == 3**, not the exit code. The count form (`grep -c`) returns 2 for the mutant and 3 post-restore, and exits 1 at 0 matches (the pre-repair baseline). Pin restored immediately after; post-restore count re-measured as 3 (see above).

### AC-GSF-002 — Full package sweep green (regression check, NOT fix proof)

```
$ go test ./internal/core/git/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/core/git	71.171s
exit=0
```

Census form (non-empty swept set):
```
$ go test ./internal/core/git/ -count=1 -v 2>&1 | grep -c '^--- PASS'
107
```

Honest note: the package was green pre-change locally (probe L0) — this run is a **regression check**, not proof of the fix. The deciding check for the defect is AC-GSF-003 (merged-tree CI, external).

### AC-GSF-003 — Config-independence (EXTERNAL — pending)

Not runnable from the lane. The deciding verdict is the CI run on the merged develop tree after the lead's batch push (`Test (ubuntu-latest)` + `Race Test`). Reported as **pending-external**; the card does not close before that verdict.

### AC-GSF-004 — No production-path diff

```
$ git diff --name-only 25a3212a9...HEAD
.moai/reports/t474/repro-probe.md
.moai/specs/SPEC-GIT-STATUS-FIXTURE-001/plan.md
.moai/specs/SPEC-GIT-STATUS-FIXTURE-001/progress.md
.moai/specs/SPEC-GIT-STATUS-FIXTURE-001/spec.md
---unstaged---
internal/core/git/status_branch_test.go
```
The committed range carries only the SPEC `.md` artifacts and the probe record (expected). The only code change is the single test file `internal/core/git/status_branch_test.go` (unstaged at measurement time; it joins the range with the run-phase commit). No non-test `.go` file appears — PASS. (Re-measured post-commit in progress.md §E.3 attribution.)

### AC-GSF-005 — Formatting clean

```
$ gofmt -l .
(no rows — empty output)
exit=0
```

### AC-GSF-006 — Optional `:315` pin decided explicitly

The kickoff **accept** branch was taken (M1 record above): `:315` is pinned (shown in the AC-GSF-001 grep output), not byte-unchanged. Never bundled silently.

## Gaps

- AC-GSF-003 remains pending-external — no CI access from the lane; the merged-tree CI verdict (lead-owned push) is the only discriminating confirmation.
- The pre-repair baseline grep (RED-now for AC-GSF-001) was captured with the `"--bare"` grep form; the pinned-form grep was not run pre-repair as a standalone command — the baseline grep output above establishes the same fact (no `-b main` token on any of the three lines), and the mutant probe demonstrates the pinned form's failure mode post-hoc.
- `gofmt -l .` was run once, post-repair; no pre-repair gofmt baseline was taken (formatting is not the defect; the single-line insertion is gofmt-canonical).

## Residual-risk

If GitHub's runner image ever pins `init.defaultBranch=main` system-wide, an unfixed fixture would go silently green-but-config-dependent; the pin removes the dependence regardless (probe Residual-risk, unchanged by this run). The local sweep cannot discriminate the defect — only the merged-tree CI can.
