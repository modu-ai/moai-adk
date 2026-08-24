# plan.md — SPEC-WORKTREE-REAPER-001

## §A — Context

Card t209, dispatched as "design a worktree reaper". The plan-phase
investigation (`.moai/reports/t209/investigation.md`) established that the
reaper exists, ships in `internal/cli/session_worktree_prmerge.go`, is enabled
in this repository, and runs on every `moai session register` / `moai session
list`. The prior-art survey (`research.md`) then established that there are
**two** sweeps sharing one blind anchor dependency. The work is a repair of
both, plus an extension of the command that already covers the non-`WT-*` tail.

Base: `origin/main` = `cd0cee1b8`, worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209`, branch
`WT-worktree-reaper`.

Revised at v0.2.0 after plan-audit iteration 1 (FAIL 0.55). The changes to this
file: M2's placement moved from `internal/cli` to `internal/session` (§F M2),
M3 changed from "build an inventory" to "extend `clean --stale`" (§F M3), and
the ordering caveat restated (§F preamble).

## §B — Known issues in the current implementation

| # | Location | Issue | Requirement |
|---|---|---|---|
| B1 | `branchMergedForCleanup` (`session_worktree_prmerge.go:189`) | gh non-answer and gh negative are the same value; git fallback unreachable while gh is installed | REQ-WR-001/002 |
| B2 | `ghPRViewStateReal` (:230) | `""` overloaded across three distinct causes | REQ-WR-001 |
| B3 | `prMergeCleanup` anchor guard (:168) | registry-only; names 1 of 5 measured live anchors | REQ-WR-006/009 |
| B4 | `parseWorktreeList` (:82) | discards the `locked` line, so the authoritative signal never reaches the guard | REQ-WR-006 |
| B5 | `cleanStaleWorktrees` (`worktree/clean.go:163`) | consumes the same blind guard, over **all** registered trees | REQ-WR-019 |
| B6 | `newCleanCmd` (`worktree/clean.go:34`) | `--base` defaults to local `main`, diverging from `prMergeCleanup`'s `origin/main` | REQ-WR-022 |
| B7 | `newCleanCmd` flag set | no machine-readable output | REQ-WR-012 |
| B8 | `isProcessAlive` (`session/anchor_pid_*.go`) | bare `bool` cannot express an undetermined probe | REQ-WR-008 |

## §C — Pre-flight

Run once before implementation, in this worktree:

```bash
go build ./... && go vet ./internal/cli/... ./internal/session/...
go test ./internal/cli/ -run '^TestPRMergeCleanup' -count=1
go test ./internal/cli/worktree/ -run '^TestCleanStale' -count=1
git worktree list --porcelain | grep -c '^locked'
```

The last command establishes the lock-line fixture is still shaped as measured.

### [HARD] C.1 — M1 precondition gate (AC-WR-025)

**M1 does not land until this measurement is taken and `design.md` §A.7's fork is
closed.** Run it from the primary checkout, **outside every worktree a session
occupies** — the v1 EC-9 failure established that a measurement taken inside a
live session's tree is not isolated (`design.md` §A.6).

```bash
# per registered worktree, cross-referenced against the M1-unblocked set
git -C <tree> status --porcelain --ignored | grep -c '^!!'
```

Record how many M1-unblocked trees (merged AND porcelain-clean AND unanchored)
carry ignored content, and how many carry ignored content outside regenerable
paths. `design.md` §A.7's table then selects policy P1, P2, or P3.

Why it gates M1 rather than M2: `.moai/state/` is gitignored
(`git check-ignore -v .moai/state/config-cache.json` → `.gitignore:284`) and MoAI
writes into it in every tree a session occupies, so "holds ignored content" may
be true of essentially the whole population. If it is, a preserve-on-ignored
policy cancels M1's unblocking of the ~98 merged trees.

## §D — Constraints

Carried from `spec.md` §E as the implementation contract:

- `auto_cleanup: false` must stay inert.
- Fail-open on the sweep; fail-closed on both guards.
- No bulk deletion; no L2 code paths; the sweep never unlocks and never passes
  `--force`.
- Extend the existing function-variable seams; do not introduce a second
  injection mechanism.
- `t.TempDir()` for isolation; English code, comments, godoc.

## §E — Self-verification

Every milestone is verified by the criteria in `acceptance.md`, each of which
uses the falsifiable form
`go test <pkg> -run '^<Test>$' -v -count=1 2>&1 | grep -c '^--- PASS: <Test>'`
with an expected count of `1` and a recorded pre-implementation observation.
Verification evidence is persisted under `.moai/state/verify/t209/`, not `/tmp`.
No time estimates appear in this plan.

## §F — Milestones

Ordered by decision-reversibility: M1 changes a seam signature (a type-level
decision, hardest to walk back), M2 changes what the guards trust and where the
decision lives (a safety and placement decision), M3 extends an existing command
(additive, easiest to revise).

**Ordering caveat, restated at v0.2.0.** The v0.1.0 caveat claimed M1-before-M2
puts live sessions at risk. That was wrong in both directions
(`design.md` §D): every live anchor here is locked, and `git worktree remove`
without `--force` refuses a locked tree (exit 128), so no live session is
removable today with or without M2; conversely the *unlocked* anchor is a real
shape that M2 does **not** close. What M1-first actually does in this repository
is make ~98 merged trees removable in a single sweep — a bulk act held out of
scope. The constraint is therefore about authorising that removal:

> **Land M1 with `auto_cleanup` temporarily `false`, or land it together with
> the AC-WR-023(b) enumerated expected-removal set**, so the first repaired
> sweep is a deliberate act rather than a side effect of a `moai session list`.

M1 and M2 are otherwise independent. M3 depends on M2 for the anchor column.

### M1 — merge-detection repair

*Priority: High.*

[HARD] **Blocked on §C.1 (AC-WR-025) — do not start until the `design.md` §A.7
fork is closed.** M1 is what makes ~98 merged trees removable in one sweep; if
ignored content proves near-universal, the policy that fork selects determines
whether those removals destroy anything.

1. Change `sessionWorktreeGhPRViewState` to `func(branch string) (string, bool)`;
   `ghPRViewStateReal` returns `("", false)` on non-zero exit or unparseable
   JSON, `(state, true)` otherwise.
2. Rewrite `branchMergedForCleanup` to the `design.md` §A.2 resolution order,
   returning a three-valued result (merged / not merged / undetermined).
3. `prMergeCleanup`: on undetermined, emit a preserve notice naming the path and
   the reason (REQ-WR-003) and continue.
4. Update `swapPRMergeSeams` and every `ghPRState:` closure in
   `session_worktree_prmerge_test.go` / `session_worktree_prmerge_anchor_test.go`
   to the new signature, preserving each existing expectation's meaning.
5. Add `TestPRMergeCleanup_GhNoAnswerConsultsGitFallback`,
   `…_GhOpenSkipsGitFallback`, `…_UndeterminedMergePreserves`,
   `…_ZeroUniqueCommitPreserved` (AC-WR-002, 003, 004, 007).

The fourth new test pins the removal class M1 reaches: a zero-unique-commit
branch is removable when clean and preserved when it carries untracked files.
`design.md` §A.4 carries the accept decision and §A.5 records that audit finding
D2's remedy was resolved against by measurement. **§A.6 now carries a corrected
EC-9**: git does *not* check ignored content at removal time — the ignored-only
tree is removed, exit 0, content destroyed. The v0.2.0 claim of a third backstop
layer is withdrawn; there are two.

Files: `internal/cli/session_worktree_prmerge.go` + its two test files.

### M2 — anchor-guard repair, shared by all three consumers

*Priority: High. Safety-critical.*

1. Add a two-valued liveness probe in `internal/session` —
   `func(pid int) (alive bool, determined bool)` — wrapping the existing
   syscall, with the per-platform mapping in `design.md` §B.5. Leave
   `isProcessAlive` in place for `LiveAnchoredSessions`.
2. Add the exported anchor decision to `internal/session` beside
   `LiveAnchoredSessions`: it takes the tree path and the stored lock reason,
   returns a verdict plus the source that produced it, and implements the
   `design.md` §B.4 fail-closed table. Union with `LiveAnchoredSessions`
   (REQ-WR-009/010).
3. Extend `wtEntry` with the stored lock reason and teach `parseWorktreeList`
   to capture the `locked` line, stripping git's own `locked ` prefix and
   mapping a bare `locked` line to an empty reason (`design.md` §B.3).
4. Wire `prMergeCleanup` to the shared decision; emit a source-naming preserve
   notice (REQ-WR-011). Pre-detect the **refusal class** from the porcelain lock
   line already parsed in step 3, and never attempt a removal known to fail
   (REQ-WR-021). Every preserve notice carries a cause-specific token
   (REQ-WR-023). Do **not** change `worktreeIsDirty` — it is shared with the M4
   session-exit path (`design.md` §B.6a).
5. Wire `cleanStaleWorktrees` (`internal/cli/worktree/clean.go:163`) to the same
   decision, replacing its direct `LiveAnchoredSessions` call, and give it a
   lock-aware keep-reason (REQ-WR-019).
6. Wire the `--merged-only` path (`internal/cli/worktree/clean.go:95`) to the
   same decision — the third consumer, and the one with no dirty guard of its
   own (REQ-WR-019, AC-WR-026).
7. **Ignored-content handling is NOT implemented in M2.** It is governed by
   `design.md` §A.7's open fork and gated on the AC-WR-025 measurement
   (REQ-WR-024). Implementing a probe before that measurement ships policy P1 by
   default into a population where it may cancel M1.
8. Add `TestLockAnchor_LivePidAnchored`, `…_DeadPidNotAnchored`,
   `…_IndeterminateIsAnchored`, `…_DeadLockRemovalIsInert`,
   `TestAnchorDecision_RegistryOnlyPathStillAnchors` (`internal/session`);
   `TestParseWorktreeList_CapturesLockReason`,
   `TestPRMergeCleanup_PorcelainFailureRemovesNothing`,
   `TestPRMergeCleanup_T207SamplePreservedByLock`,
   `TestPRMergeCleanup_RefusalClassNamesCause`,
   `TestPRMergeCleanup_RefusalFallThroughNamesCauseAndContinues` (`internal/cli`);
   `TestCleanStale_LockAnchoredWorktreeKept`,
   `TestCleanMergedOnly_LockAnchoredWorktreeKept` (`internal/cli/worktree`).

Files: `internal/session/anchor.go` (+ a new `anchor_lock.go` for the parser and
decision, keeping them unit-testable), `internal/session/anchor_pid_unix.go`,
`anchor_pid_windows.go`, `internal/cli/session_worktree_prmerge.go`,
`internal/cli/worktree/clean.go`, plus tests in all three packages.

### M3 — non-`WT-*` coverage by extending `worktree clean --stale`

*Priority: Medium. Decision recorded in `design.md` §C (option O3-d).*

1. Add `--json` to `clean`, emitting one object per non-protected tree with
   path, branch, keep-reason, dirty state, merge state, anchor state
   (REQ-WR-012). Human output unchanged.
2. Change the `--base` default from `main` to `origin/main` (REQ-WR-022), and
   note the change where the flag is declared.
3. Confirm the preview/`--yes` gate is untouched (REQ-WR-013/014) — it is the
   shipped mechanism REQ-WR-014 relies on, not something to reinvent.
4. Add `TestCleanStale_JSONEmitsAllTreesWithStates`,
   `…_BaseDefaultsToOriginMain`, `…_NoAskUserQuestion` (AC-WR-017, 019, 020);
   keep `…_PreviewsByDefault` / `…_RemovesWithYes` green (AC-WR-018).
5. Document the manual disposal procedure for the human-created tail in the
   worktree rule surface, pointing at `clean --stale --json` for the inventory.

Files: `internal/cli/worktree/clean.go` + tests,
`.claude/rules/moai/workflow/worktree-integration.md`.

**No new command is added.** A second surface answering the same question as
`clean --stale` was v0.1.0's plan and is explicitly reversed.

## §G — Anti-patterns to avoid

- **Letting git override a determinate gh answer.** git is squash-blind in the
  `prMergeCleanup` fallback; a squash-merged branch it reports unmerged must not
  flip a gh `MERGED`. git fills a hole, never contradicts.
- **Adding a unique-commit predicate alongside `git branch --merged`.** It is
  the same call (`research.md` §C.2). The additional guard that matters is the
  dirty check, which already exists.
- **Replacing the registry source with the lock source.** REQ-WR-010; the two
  are unioned so the separate `RelocateSession` fix composes.
- **Treating "no lock line" as "no session".** It means the lock source has no
  opinion; the registry is still consulted, and the unlocked-anchor residual
  (REQ-WR-020) lives exactly here.
- **Unlocking, or passing `--force` to `git worktree remove`.** Out of scope;
  the refusal class is inert by decision (REQ-WR-021).
- **Attempting a removal git is known to refuse.** A locked tree's condition
  never clears, so an attempt-and-fail design emits a permanent `cleanup failed`
  notice for a correctly-behaving tree.
- **Emitting a preserve notice that does not name its cause.** REQ-WR-023; two
  trees preserved for different reasons must be distinguishable from the output.
- **Implementing an ignored-content probe before AC-WR-025 is measured.** That
  ships policy P1 by default into a population where it may cancel M1
  (`design.md` §A.7).
- **Treating REQ-WR-021's pre-detection set as exhaustive.** It is not — the
  submodule case is a measured member outside it (EC-12).
- **Building a parallel inventory command.** `design.md` §C selected extension.
- **A second test-injection mechanism.** Extend `swapPRMergeSeams` /
  `swapSessionWorktreeSeams`.
- **Unanchored `-run` patterns in verification.** `-run 'NoAskUserQuestion'`
  matches 31 existing tests; `-run <MissingTest>` exits 0. Use the anchored
  `^…$` + `grep -c '^--- PASS: …'` form from `acceptance.md` §0.
- **Running the full test suite locally.** Target `./internal/cli/...`,
  `./internal/cli/worktree/...`, `./internal/session/...`; CI owns the
  full-suite verdict. The `internal/cli` package needs `-timeout 600s`.

## §H — Cross-references

- `spec.md` — 24 requirements, constraints, out of scope
- `design.md` — D1 seam signature, D2 fail-closed table + probe seam + dead-lock policy, D3 the reworked M3 option set, §D the corrected ordering caveat
- `acceptance.md` — 27 criteria, the falsifiable command form, per-criterion pre-implementation baselines
- `research.md` — the two-sweep survey, seam/test conventions, platform probes
- `.moai/reports/t209/investigation.md` — the original measured survey
- `.moai/reports/t209/plan-audit.md` — iteration-1 audit
- SPEC-SESSION-WORKTREE-001 — the SPEC that shipped `prMergeCleanup` (M8)
