# t343 — pre-work measurements (RED-now threshold)

tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t343`
branch: `WT-red-now-threshold`
HEAD: `a6bbbf82b` (== `origin/develop` at time of measurement)
card: t343

All figures below were measured in THIS tree at THIS HEAD. Nothing is carried over.

## M1 — plan-auditor has zero coverage of the RED-now cell

Claim: the plan-phase audit contains no criterion that inspects an acceptance
criterion's RED-now cell.

Command + output:

```
$ grep -c "RED\|two-cell\|verification-completeness\|mutant" \
    .claude/agents/moai/plan-auditor.md
0
$ echo "exit=$?"
exit=1
$ grep -c "" .claude/agents/moai/plan-auditor.md
482
```

Zero matching lines in 482. Must-pass criteria are MP-1..MP-7; Group 4
(Acceptance Criteria Quality) is AC-1..AC-5 (GWT form, binary-testability,
weasel words, REQ traceability both directions). None reads the RED cell.

**Correction (this file's first revision).** An earlier draft of this section
reported the same four-token grep as "matches only line 4 (`pre-implementation`
in the frontmatter description)". That attribution was wrong: line 4 matched
only because the invocation actually run also carried `pre-implementation` as a
fifth alternation branch. The four-token pattern above matches nothing, exit 1.
manager-spec caught the mis-citation and it was re-measured here before being
accepted. The conclusion is unchanged and its evidence is stronger than first
stated. A case-insensitive run of the same pattern hits only `color: red`,
`required`, and `Required format` — no criterion text.

Consequence: the observed 0.94 pass is not a threshold set too low. The
criterion does not exist. The corrective is a new must-pass item, not a numeric
adjustment. This contradicts the card's own title ("임계를 조정") and the
contradiction is recorded rather than smoothed over.

## M2 — the live instance (t306) carries a pinned SHA and no output

File: `.moai/specs/SPEC-TODO-SQLITE-001/acceptance.md`

Header (lines 3-4) claims RED `today (against WT-todo-sqlite@d29b8942e)` — a
tree SHA IS pinned. No RED cell in the table carries a command output.

Verbatim RED cells:

| AC | RED-now cell |
|---|---|
| AC-TOSQ-001 | `Test name does not exist → suite failure ("no tests to run" surfaces red).` |
| AC-TOSQ-002 | `Characterization in same suite; red via missing test.` |
| AC-TOSQ-003 / 004 / 005 / 007 / 008 / 017 / 018 | `Red via missing test.` |
| AC-TOSQ-013 | `Current tree red in the future sense — driver+pragmas don't exist yet` |
| AC-TOSQ-014 | `Absent new code measures trivially below bar once landed-without-tests` |

The pin is present as decoration: it names a tree but attributes no observation
to it. The combination "pin present, output absent" is what L1 targets.

Census, re-measured:

```
$ grep -c "Red via missing test" .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md
7
```

**Correction (first revision).** An earlier draft counted four cells carrying
that identical sentence. Seven carry it; AC-TOSQ-004, -007, -008 and -018 were
omitted. manager-spec caught the undercount and it was re-measured here.

**Correction (second revision) — the family is wider than the exact phrase.**
Counting only the literal string undercounts the pattern. The cells whose RED
rests on a test or a verb being absent:

```
$ awk -F'|' '/missing test|does not exist|no tests to run/ {gsub(/ /,"",$2); print $2}' \
    .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md
AC-TOSQ-001  AC-TOSQ-002  AC-TOSQ-003  AC-TOSQ-004  AC-TOSQ-005
AC-TOSQ-007  AC-TOSQ-008  AC-TOSQ-011  AC-TOSQ-017  AC-TOSQ-018
```

Ten cells, of which **nine** (001, 002, 003, 004, 005, 007, 008, 017, 018) rest
on the same false premise that an absent *test* turns the suite red — false in
both directions available to them, since a package suite lacking the test is
green and a `-run` selector missing it is also green (M3).

**AC-TOSQ-011 is deliberately excluded from the nine.** Its RED rests on an
absent *CLI verb* (`moai todo export-json` → non-zero exit), a different
mechanism that may well be true. It was not measured here, so it is neither
counted as defective nor asserted sound.

This makes the instance systemic within one SPEC rather than a single slip:
nine release-blocking criteria carrying one false premise. The MP-8 argument is
correspondingly stronger. Scope re-measured at the lead's request after the lead
independently found the same family.

## M3 — AC-TOSQ-001's RED claim is false, and re-execution overturns it

**Correction (this file's second revision) — the first command cited here was
invalid as evidence, and the conclusion survives on a replacement.**

The original measurement ran `go test ./internal/kanban -run TestMigrationParity`
and read its `ok` / `exit=0` as proof that an absent test name yields green. That
command proves nothing of the sort, because the test exists:

```
$ grep -rn "func TestMigrationParity" internal/kanban/
internal/kanban/backlog_migrate_test.go:70:func TestMigrationParity(t *testing.T) {
internal/kanban/backlog_migrate_test.go:424:func TestMigrationParityCatchesTamperedRecord(t *testing.T) {
```

The `ok` was a real, passing test — not a zero-match selector. The premise of the
refuting command was never verified before the refutation was asserted, which is
the same unobserved-claim defect this card exists to prohibit, committed while
documenting it. The shape is worth stating plainly, because it is this card's own subject:
**in a card whose purpose is to catch unobserved claims in RED cells, the
refutation was itself an unobserved claim.** That is the strongest argument the
SPEC has for a mechanical MP-8 — the defect is not prevented by knowing about
it. The author of this file knew the rule, was actively applying it, and
committed it anyway.

The lead recorded its own share on the same failure: it confirmed the cell's
TEXT against the develop copy and reported that as independent confirmation,
then passed the sample to lane-5 and the operator. What it verified was the
cited fact; what it let through was the inference. Two actors, one chain, and
the check that would have caught it — re-running the command and asking whether
its premise holds — was performed by neither until lane-5 ran it.

The replacement command, run in this tree, uses a name that genuinely does not
exist:

```
$ go test ./internal/kanban -run TestMigrationParityDoesNotExistXYZ -count=1 ; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.395s [no tests to run]
exit=0
```

A `-run` selector matching zero tests exits 0 and prints `ok`, with
`[no tests to run]` as the only distinguishing token. The cell asserted that the
absent test name surfaces red; the tree is green. **The conclusion is unchanged;
its evidence is replaced.** Cite the replacement command, never the original.

Two consequences:

1. An unobserved RED cell is not merely unverified — it can be actively false
   and still be classified release-blocking and pass at 0.94.
2. The discriminator that catches this is re-execution, not reading. The cell is
   written in the assertive present and reads like an observation report.

### M3a — shared sample with t341 (lane-5)

The fact M3 rests on — a zero-match selector exiting green — is the subject of
card t341, which lane-5 holds. lane-5's own instance is recorded as
`ok ... 0.424s [no tests to run]` in `.moai/reports/t350/discovery.md`.

The two cards use one sample from opposite sides:

- t341: a zero-match selector produces green, so build a mechanical warning.
- t343: a cell asserting that a zero-match selector produces red passed as
  release-blocking, so check the RED cell's truth.

The axes are distinct; the sample is shared. Recorded here so that a later
repair of `SPEC-TODO-SQLITE-001` AC-TOSQ-001 cannot silently remove the other
card's evidence without the removal being visible.

## M3b — the evidence ledger records a zero-execution run as a test PASS

Found by lane-5 (t341) and re-read here. `internal/hook/evidence_writer.go`
`deriveFromOutputText` treats `ok  \t` as a precise pass marker and returns
before any count is inspected:

```
$ sed -n '216,222p' internal/hook/evidence_writer.go
	hasPrecisePass := strings.Contains(text, "ok  \t") ||
		strings.Contains(text, "ok \t") ||
		strings.Contains(text, "test result: ok")
	if hasPreciseFail {
		return false, true, true
	}
	if hasPrecisePass {
		return true, false, true
	}
```

`ok ... [no tests to run]` contains `ok  \t`, so a run that executed nothing is
written to the evidence ledger as an observed test PASS and the Stop evidence
gate stays quiet.

**Consequence for the L2 layer proposed by this card.** If the auditor's
re-execution verdict is derived through this seam, a re-run that matches zero
tests is recorded as a pass rather than as a failure to reproduce the RED — the
exact inversion the layer exists to prevent. L2's verdict must therefore key on
**the count of tests actually executed**, not on the `ok` token. Recorded as a
design constraint, not as a defect of this card.

## M4 — the card's stated target shape misses the worst instance

The card names the target shape as "counterfactual / future-sense sentences".
AC-TOSQ-013 and AC-TOSQ-014 match that description. AC-TOSQ-001 — the cell M3
falsified — does not: it is assertive present. A tense/mood discriminator
catches the two harmless cells and misses the false one.

Corroborating measurement from another lane (cited, not re-measured here):
t342 found natural-language lexical discriminators unsound in both directions,
with 7 false positives where live criteria were classified as ghosts on
vocabulary alone.

Conclusion: no clause may key on tense, mood, or a word list.

## M5 — classification precedent already exists (t331, not yet on develop)

`.claude/worktrees/t331/.moai/specs/SPEC-TODO-LANDING-STATE-001/acceptance.md`
§C splits criteria into release-blocking (each carrying an observed RED) and
regression-guard ("must stay green; a failure is a contract break, but there is
no RED to flip"), demoting AC-TLS-002 and AC-TLS-007 rather than inventing a
failure for them. Its text: "No criterion is classified release-blocking on a
RED cell that argues counterfactually or in the future tense; where a cell has
no observation it says so and the criterion is a regression-guard instead."

This is a local convention inside one SPEC. Lifting it into the audit layer is
this card's work. t331 is unlanded at the time of measurement — cited as
precedent, not depended on as a file.

## M6 — a reusable mechanical seam exists (t338)

`internal/spec/ac_count_clause_test.go`: a repository-local Go test that
extracts a command from an agent file between sentinel comments
(`# MOAI-AC-COUNTER-BEGIN` / `-END`), asserts exactly-one-and-non-empty span
before comparing so a zero- or two-match anchor cannot become a vacuous pass,
keeps fixtures under `internal/spec/testdata/ac_count/`, and ships nothing.

## M7 — template mirrors

```
$ diff -q .claude/rules/moai/development/verification-completeness.md \
    internal/template/templates/.claude/rules/moai/development/verification-completeness.md
(no output; exit 0)
$ diff -q .claude/agents/moai/plan-auditor.md \
    internal/template/templates/.claude/agents/moai/plan-auditor.md
Files ... differ (exit 1)
```

Both surfaces are Template-First. The rule pair is byte-identical; the agent
pair differs by expected neutralization.

## Gaps — what was NOT observed

- No measurement of how many OTHER landed SPECs carry prose-only RED cells. One
  instance (t306) was read in full; the corpus was not swept.
- The 0.94 figure for t306 is quoted from the card, not re-derived here.
- t331's §C text was read from that worktree's disk; its landing state on
  develop was not verified beyond its absence from
  `.moai/specs/SPEC-TODO-LANDING-STATE-001` in this tree.
- t342's 7 false positives are cited from the dispatch, not re-measured.
- lane-5's `ok ... 0.424s [no tests to run]` line (M3a) is cited from the lead's
  dispatch; `.moai/reports/t350/discovery.md` was not opened from this tree.

## Residual risk

- M3 used the Go toolchain's `-run` semantics. A different language's test
  runner may exit non-zero on an unmatched selector, so the specific instance
  generalizes as "selector semantics must not be assumed", not as "unmatched
  selectors are always green".
- `(cached)` in the second M3 invocation means the exit code was read from Go's
  test cache. The first invocation ran uncached and printed `ok` with a
  duration, so the green is not a cache artifact.
- M1 is a negative claim over one file. It establishes that `plan-auditor.md`
  carries no such criterion; it does not establish that no other audit surface
  does. `sync-auditor` and `moai spec audit` were not swept for it.
