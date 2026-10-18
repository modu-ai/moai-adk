# t1865 — sync hold (rule C, board 9e76)

Card t1865, SPEC-UPDATE-OWNERSHIP-REPAIR-001, branch WT-update-ownership-repair. The sync audit failed four times; the correction plus delta audit allowed by rule C after the third round has been used. No further correction is made by the lane. The tree is preserved as is.

## State at the hold

- Code: the run phase is complete (M0-M10, 21 AC-OWR, 12 binding conditions dispositioned in progress.md E.3). Final run code commit `dc32a94c9e`. Sync commit `e99a990296dc8d84b48891a9d916b8379b9f4f3b` (spec.md `completed`, CHANGELOG, E.4), backfill `1b5332af57`, later E.4/CHANGELOG corrections up to `c1772719da`.
- Evidence HEAD at the last audit: `aba63b148319b25ec44080c89071a2b99462a8bd`.
- Sync audit files, all FAIL because the required codex backend failed: `sync-audit-r1.md`, `-r2.md`, `-r3.md`, and `sync-audit.md` (round 4).
- Scoped tests: run by the lane (changed-test-file names, 79 top-level PASS at `28b2882688`, repeats green) and by the round-4 auditor under the admission rule (load, free memory and slot values recorded in its report); package-wide `internal/cli` test, coverage and `-race` were not run (carried to CI and the landing remeasure).

## Rounds

| Round | Audited HEAD | Receipt | Findings |
|---|---|---|---|
| 1 | `dfd7c40fb2` | `rcpt-7770ccadc2f78594b050e777` | F1 BLOCKING (unregistered symlink read outside the project), F2 SHOULD-FIX (same-second prior-manifest overwrite), F3 SHOULD-FIX (FIFO at the user-root counterpart hangs the verdict) |
| 2 | `38c8ca5376` | `rcpt-aeb48f1d567c915a06f7b90f` | N1 SHOULD-FIX (temp-name filter dropped user files under a prefix-named directory), N2 MINOR/P2 (sibling without run marker out of retention) |
| 3 | `1d9b06a737` | `rcpt-665b06f8d214d9a97633b90c` | F-A SHOULD-FIX (link swapped between boundary check and read), F-B SHOULD-FIX (marker write failing after the file exists leaks the sibling) |
| 4 | `aba63b1483` | `rcpt-bc15a518bfe20c055b8f4bda` | N-A SHOULD-FIX (`os.OpenRoot(root)` opens the project root by path after the check; a root-path swap read outside bytes once in 2079 attempts), N-B MINOR (`os.RemoveAll(sibling)` also removes foreign content), N-C optional (absolute in-project link swapped in is kept as symlink-escape) |

## What the trend says (lane's reading, evidence above)

- After round 1 no finding is BLOCKING and none is in the core ownership logic (REQ-OWR / AC-OWR); every later finding was introduced by, or is a narrower variant of, the previous round's repair.
- The attacker model narrowed each round: write access inside the project (round 1) -> a link swapped between two calls (round 3) -> renaming entries in the parent directory of the project root while `moai update` runs (round 4). Removal is disabled in production (`removalEnabled = false`); the bytes read in the affected paths are hashed for provenance and never used to remove anything.
- The required codex backend fails on any P2; every round it found a new edge in the newest small helper, so another correction round is likely to meet another narrower edge.

## Options for the leader / operator (the lane does not choose)

A. **One more correction round plus one audit under an operator exception** (the ceiling is exhausted, so this is operator-held): N-A open the `os.Root` once on the verified root and reuse it for the read, plus a root-swap row in the race table; N-B remove the marker file then a non-recursive `os.Remove` of the sibling; N-C one line in E.4. Roughly 15 lines of code plus tests.
B. **Shrink the card's guarded surface**: drop the sibling-directory scheme (a same-second collision fails the prior-manifest claim closed, so the run removes nothing) and read an unregistered file only when no path component is a symlink. This removes the N-B class and narrows N-A but changes the pinned in-project symlink baseline (an in-project symlink would be kept unread instead of judged by content) and needs a SPEC wording review.
C. **Accept N-A/N-B/N-C as dispositioned residual risk**: the attacker model is outside the update tool's trust boundary and removal is disabled. This needs the operator to waive the required codex gate for this finding class; the lane cannot do that and does not recommend it without an explicit decision.

Lane recommendation: A, because the remaining fixes are small and local, no AC is failing, and the findings no longer touch the ownership logic. It is a recommendation, not a decision: the exception is the operator's.

## Preserved

Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, branch `WT-update-ownership-repair`, unpushed. All audit and card-review reports are committed under `.moai/reports/t1865/`. Nothing is deleted; the rejection record count is 0.
