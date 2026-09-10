---
id: SPEC-UPDATE-MERGE-CONFLICT-BLIND-001
title: "Acceptance criteria — update merge conflict blindness"
version: "0.1.0"
created: 2026-09-10
---

# Acceptance — SPEC-UPDATE-MERGE-CONFLICT-BLIND-001

Every criterion below is binary and names the command that decides it. Test
identifiers (`TestSharedKeyControlCells`, `TestConflictSurfaceReachability`,
`TestPreservedSecurityKeySignal`) are **contracts fixed by this SPEC**: the
run-phase may not rename them without amending this file, because the deciding
commands select on them.

`PKG` below is the package the reproduction lands in — `./internal/merge/...` or
`./internal/cli/update/merge/...`, chosen in M1 and recorded in `progress.md`
§E.2. A criterion whose command reports `no tests to run` or `[no test files]`
is **FAIL**, not PASS: a selector that matches nothing prints a green line.

## §D Acceptance matrix

### M1 — the reproduction

**AC-UMC-001 — isolation**
Given the reproduction harness,
When `grep -n 't\.TempDir()' <test file>` is run and the same file is scanned for
absolute paths into a real checkout with
`grep -nE '/Users/|\.claude/settings\.json"|/moai/moai-adk-go' <test file>`,
Then the first command reports at least one match and the second reports zero
matches.
Decides: `REQ-UMC-001`. FAIL on any hit in the second command.

**AC-UMC-002 — no real `moai update` invocation**
Given the reproduction harness,
When `grep -nE 'exec\.Command|RunUpdate|runUpdate' <test file>` is run,
Then it reports zero matches, or every match is shown to target a directory the
harness itself created in the same test body.
Decides: `REQ-UMC-002`.

**AC-UMC-003 — cell (i): shared key the user did not touch**
Given a fixture whose `current` and `updated` documents share a key that the
user's side left at the template's previous value while the template changed it,
When `go test $PKG -run 'TestSharedKeyControlCells/untouched_shared' -v -count=1`
is run,
Then the test passes and its output records the value the merge wrote for that
key, together with the prediction it was compared against
(`plan.md` §F M1 table).
Decides: `REQ-UMC-003`, cell (i).

**AC-UMC-004 — cell (ii): shared key the user emptied to `[]`**
Given the same fixture, in which the user's side holds a literal `[]` for a key
the template ships non-empty,
When `go test $PKG -run 'TestSharedKeyControlCells/emptied_shared' -v -count=1`
is run,
Then the test passes and its output records whether the merge wrote `[]` or the
template's non-empty value.
Decides: `REQ-UMC-003`, cell (ii). This is the `permissions.ask` shape.

**AC-UMC-005 — cell (iii): shared key the user changed to another value**
Given the same fixture, in which the user's side holds a non-empty value
differing from the template's,
When `go test $PKG -run 'TestSharedKeyControlCells/changed_shared' -v -count=1`
is run,
Then the test passes and its output records which of the two values the merge
wrote.
Decides: `REQ-UMC-003`, cell (iii).

**AC-UMC-006 — cell (iv): the counter-cell, a key the user's file omits**
Given the same fixture, in which the user's side omits a key the template
carries,
When `go test $PKG -run 'TestSharedKeyControlCells/omitted' -v -count=1` is run,
Then the test passes and its output records that the template's value landed.
Decides: `REQ-UMC-005` cell (iv), `REQ-UMC-006`.
**This criterion is the discriminator.** AC-UMC-003/004/005 all predict "the
user's side stands", so a harness that merely copied the user's file would pass
all three. If AC-UMC-006 reports the template's value **not** landing while the
other three pass, the harness is measuring nothing and every result is void.

**AC-UMC-007 — all four cells in one run, against one fixture**
Given the four sub-tests above,
When `go test $PKG -run 'TestSharedKeyControlCells' -v -count=1` is run once,
Then the output names exactly the four sub-tests, each with a `--- PASS` or
`--- FAIL` line, and no cell is absent from the run.
Decides: `REQ-UMC-003`, `REQ-UMC-007`. An absent sub-test is a gap, not a pass.

**AC-UMC-008 — the conflict surface is measured, not inferred**
Given a fixture whose shared key carries genuinely divergent values on the
`current` and `updated` sides,
When `go test $PKG -run 'TestConflictSurfaceReachability' -v -count=1` is run,
Then the test passes and its output records the observed
`MergeResult.HasConflict` value and `len(MergeResult.Conflicts)` for that key.
Decides: `REQ-UMC-004`. The assertion must read those two fields from the merge's
own return value; deriving the base and comparing it to the value it was derived
from asserts nothing.

### M2 — instrument honesty

**AC-UMC-009 — no unconflicted report on an undetermined divergence**
Given a shared key whose divergence the merge cannot classify,
When `go test $PKG -run 'TestConflictSurfaceReachability' -v -count=1` is run
after the M2 change,
Then the test passes and asserts that the merge does not report the outcome as an
unconflicted merge.
Decides: `REQ-UMC-008`.

**AC-UMC-010 — the surface carries no unreachable branch**
Given the M2 change,
When `go test $PKG -count=1` is run with the M2 change reverted in the working
tree (mutant),
Then at least one test that passed with the change **fails** without it.
Decides: `REQ-UMC-009`. A test green on both the repaired and the unrepaired code
establishes nothing; this criterion is what separates a real guard from a
vacuous one.

**AC-UMC-011 — designed resolution preserved**
Given the M2 change,
When `go test $PKG -run 'TestSharedKeyControlCells' -v -count=1` is re-run,
Then every cell's recorded written value is byte-identical to the value recorded
in M1 for the same cell.
Decides: `REQ-UMC-010`. Any change in a written value is a FAIL, whatever else
improved.

### M3 — the signal

**AC-UMC-012 — a signal is emitted on a preserved security-relevant divergence**
Given a fixture in which the merge preserves a user-side value for a key the
template ships with a non-empty security-relevant value,
When `go test $PKG -run 'TestPreservedSecurityKeySignal/emitted' -v -count=1` is
run,
Then the test passes and asserts the signal names the key and both values.
Decides: `REQ-UMC-011`.

**AC-UMC-013 — silence when there is no divergence**
Given a fixture in which no such divergence exists,
When `go test $PKG -run 'TestPreservedSecurityKeySignal/silent' -v -count=1` is
run,
Then the test passes and asserts that no signal was emitted.
Decides: `REQ-UMC-012`.

**AC-UMC-014 — the signal changes no value**
Given the M3 change,
When `go test $PKG -run 'TestSharedKeyControlCells' -v -count=1` is re-run,
Then every cell's written value remains byte-identical to M1's recording.
Decides: `REQ-UMC-013`, and re-asserts `REQ-UMC-010`.

### Cross-cutting

**AC-UMC-015 — no permission value touched in any checkout**
Given the whole SPEC's work,
When `git status --porcelain -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl`
is run in the worktree and in the primary checkout,
Then both report no modification attributable to this SPEC, unless a milestone
explicitly reached `REQ-UMC-014`, in which case both paths appear together.
Decides: `spec.md` §D, `REQ-UMC-013`, `REQ-UMC-014`.

**AC-UMC-016 — template coupling, if reached**
Given a change to `.claude/settings.json`,
When `git diff --name-only` is read and `make build` is run,
Then `internal/template/templates/.claude/settings.json.tmpl` appears in the same
diff and `make build` exits 0.
Decides: `REQ-UMC-014`. **Not applicable** if no settings document changed —
record N/A with the diff as evidence, not PASS.

**AC-UMC-017 — scoped verification**
Given the change set,
When `go test ./internal/merge/... ./internal/cli/update/merge/... -count=1` is
run,
Then it exits 0 and its output names at least one test file per package (a
`[no test files]` line for a package this SPEC changed is a FAIL).
Decides: `plan.md` §D. The full-suite verdict is CI's, not this run's.

## §D.1 Severity

- **MUST-PASS**: AC-UMC-001, 002, 006, 007, 008, 010, 011, 014, 015, 017.
  AC-UMC-006 and AC-UMC-010 are the anti-vacuity criteria; AC-UMC-011 and
  AC-UMC-014 are the designed-behaviour guards.
- **SHOULD-PASS**: AC-UMC-003, 004, 005, 009, 012, 013.
- **CONDITIONAL**: AC-UMC-016 (applies only if a settings document changed).

## §D.2 Traceability

| REQ | AC |
|---|---|
| REQ-UMC-001 | AC-UMC-001 |
| REQ-UMC-002 | AC-UMC-002 |
| REQ-UMC-003 | AC-UMC-003, 004, 005, 007 |
| REQ-UMC-004 | AC-UMC-008 |
| REQ-UMC-005 | AC-UMC-003, 004, 005, 006 |
| REQ-UMC-006 | AC-UMC-006 |
| REQ-UMC-007 | AC-UMC-007 |
| REQ-UMC-008 | AC-UMC-009 |
| REQ-UMC-009 | AC-UMC-010 |
| REQ-UMC-010 | AC-UMC-011, 014 |
| REQ-UMC-011 | AC-UMC-012 |
| REQ-UMC-012 | AC-UMC-013 |
| REQ-UMC-013 | AC-UMC-014, 015 |
| REQ-UMC-014 | AC-UMC-015, 016 |
| REQ-UMC-015 | (documentary — asserted in `spec.md` §B.5; no mechanical AC) |

## §D.3 Definition of Done

- Every MUST-PASS criterion PASS, each with its command and verbatim output cited.
- Every unmeasured cell or criterion recorded in the Gaps section of the phase
  report — an empty Gaps section asserts that nothing was left unobserved, which
  must itself be true.
- The withdrawn framing (`spec.md` §A.2) absent from every artifact:
  `grep -rn 'merge is broken' .moai/specs/SPEC-UPDATE-MERGE-CONFLICT-BLIND-001/`
  reports zero matches outside the HISTORY line that records the withdrawal.

🗿 MoAI
