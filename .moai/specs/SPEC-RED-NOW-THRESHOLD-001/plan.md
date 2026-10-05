# SPEC-RED-NOW-THRESHOLD-001 — Implementation Plan

Tier M. Authored on `WT-red-now-threshold@a6bbbf82b` (card t343).

## §A Context

`verification-completeness.md` §2 requires a RED-now cell to be *"the criterion observed red on
the pre-implementation tree, pinned to the tree it was measured on"*. `plan-auditor.md` contains
no criterion that reads a RED cell. This plan adds one, in three layers (spec.md §B), and reuses
the t338 sentinel-extraction seam rather than building a new checker.

Pre-measurement evidence, with its own Gaps section, is at
`.moai/reports/t343/red-now-premeasurement.md`. That file's Gaps record — carried forward here so
it is not lost — is that no corpus sweep was run for other prose-only RED cells across the 706
SPEC directories, the `0.94` audit figure was quoted rather than re-derived, and lane-5's
measurement was taken from the dispatch rather than opened from this tree.

## §B Known Issues — the card's own framing is wrong, and this plan contradicts it

The card is titled **"임계를 조정한다"** (adjust the threshold). That framing is refuted by
measurement 1, and this plan does not adopt it.

```
$ grep -c "RED\|two-cell\|verification-completeness\|mutant" .claude/agents/moai/plan-auditor.md
0
```

Measured on `a6bbbf82b`, against a 482-line file. `plan-auditor.md` carries **no criterion that
reads a RED cell at all**. Therefore `SPEC-TODO-SQLITE-001`'s 0.94 pass did not come from a
threshold set too low — it came from the RED cell being outside the scoring surface entirely.
There is no dial to turn.

The corrective is a **new must-pass criterion (MP-8)**, not a numeric adjustment. This is stated
here rather than silently written into spec.md as though it had always been the card's framing:
the card said one thing, the measurement says another, and the record should show which was
which.

*Discrepancy against the dispatch, recorded rather than smoothed over:* the dispatch reported this
grep as "matches only line 4". Re-run on this tree, case-sensitively as written, it matches
**zero** lines; case-insensitively it matches only incidental substrings (`color: red`,
`required`, `Required format`). The conclusion — zero coverage — is strengthened, not weakened.

Two further dispatch figures were re-measured and differ:

| Dispatch claim | Re-measured on `a6bbbf82b` |
|---|---|
| grep matches line 4 | 0 matches (case-sensitive) |
| "four cells, identical sentence" (`Red via missing test.`) | **7** cells (`grep -c` on t306 acceptance.md) |

Neither correction changes any requirement.

## §C Pre-flight

- [ ] `.claude/rules/moai/development/verification-completeness.md` and its template mirror are
      byte-identical at start (`diff -q` exit 0 — verified on `a6bbbf82b`).
- [ ] `.claude/agents/moai/plan-auditor.md` and its template mirror **differ** at start
      (expected neutralization; `diff -q` exit 1 — verified). Mirror edits are therefore
      hand-applied, never `cp`.
- [ ] `internal/spec/ac_count_clause_test.go` is present and its sentinel-extraction helper is
      readable (the seam being reused).

## §D Constraints

- **No lexical discriminator, anywhere.** No clause may key on tense, mood, or a word list
  (REQ-RNT-002). Card t342 measured such discriminators unsound in both directions, with 7 false
  positives classifying live criteria as ghosts on vocabulary. The tense/mood cut is additionally
  *insufficient* on its own evidence: it catches AC-TOSQ-013 and AC-TOSQ-014 (future-sense) and
  **misses AC-TOSQ-001**, the only cell measured factually false, which is written in the
  assertive present and reads like an observation report.
- **MP-8 is score-independent** (REQ-RNT-005). It is not a Group 1-8 rubric dimension.
- **Template neutrality** on the mirror (REQ-RNT-011): no SPEC ID, card id, or internal date.
- **Nothing ships from the Go test** (REQ-RNT-012): fixtures under `internal/spec/testdata/`,
  outside the distributed template tree.
- No time estimates. Priority labels only.

## §E Self-Verification

Each milestone closes by citing a command and its output in `progress.md` §E.2, and by naming
which acceptance criterion the evidence discharges.

## §F Milestones — ordered by decision-reversibility

Leading with the decisions most likely to change under review; mechanical steps last.

### M1 — The RED-cell content contract (Priority: High) — most reversible decision

Edit `verification-completeness.md` §2 to add the four-element RED-cell obligation (command,
verbatim stdout, exit code, tree SHA — see decision (a) below for why the fourth element exists)
and the carrier definition (table cell or fenced evidence-ledger entry)
(REQ-RNT-001), the structural-not-lexical constraint (REQ-RNT-002), and the undecidable
disposition (REQ-RNT-003).

This is the milestone most likely to be revised, because it fixes the wording every later layer
reads. Iteration 1 surfaced two wording decisions as open; **both are now settled** (auditor
ruling, adopted in full — see §I). They are recorded here rather than deleted, so the record shows
what was proposed and what replaced it.

**(a) What counts as "the command" — SETTLED: read-only, single shell invocation.** No pipes, no
redirection, no `&&`, no `;` chaining, no subshells. The plan originally proposed a permissive
FORM on the reasoning that the *disposition* is permissive. The auditor separated the two, and the
separation is right: MP-8 **executes** the string, so a permissive disposition (an unrunnable
citation is a demotion, not an error — unchanged) does not license a loose form. Consequences:

- The two iteration-1 command defects become structurally unrepeatable rather than individually
  patched — a pipe cannot appear in a conforming command at all.
- A side-effecting command never reaches the auditor's hands (composes with the D6 read-only
  constraint, REQ-RNT-013).
- **One consequence the ruling implies but did not state**: `; echo $?` is now illegal, so the
  exit code can no longer ride inside the command. The RED cell therefore carries **four**
  elements, not three — command, verbatim stdout, **exit code**, tree SHA. REQ-RNT-001 was written
  with the fourth element added.

**(b) Document-level tree pin — SETTLED: allowed, with a stronger condition than proposed.** The
plan's condition was "the cell does not itself contradict it". That excludes nothing: a silent
cell contradicts nothing, and silence is the common case. The adopted rule:

- The document-level pin **explicitly binds every cell carrying no pin of its own** (stated
  inheritance, not implied).
- The pin is a **SHA, never a branch name** (rule §4 — a moving ref falsifies itself between
  measurement and reading).
- A cell-level pin **wins** wherever present.

This SPEC's own `acceptance.md` already behaves more strictly than it proposes — header pin plus
per-cell restatement — so the requirement is written to match the practice rather than below it.
The single counter-example was this plan's own `spec.md` REQ-RNT-011 (D8), which pinned line
numbers with no SHA into a requirement about a file M4 edits; it has been corrected.

### M2 — MP-8 in plan-auditor (Priority: High)

Add MP-8 to §M5 Must-Pass Firewall, the Group 4 checklist item, and the `## Must-Pass Results`
report row (REQ-RNT-004, -005, -006, -007), following the MP-5/MP-6/MP-7 shape verbatim,
including the MP-4 `N/A`-with-stated-reason precedent.

Wrap the extractable form contract in a `# MOAI-REDNOW-BEGIN` / `# MOAI-REDNOW-END` sentinel pair
so M3 can extract rather than restate it.

### M3 — The repository-local form test (Priority: Medium)

`internal/spec/red_now_cell_test.go`, on the `ac_count_clause_test.go` pattern (REQ-RNT-008):
sentinel extraction, exactly-one-non-empty-span asserted before any comparison, fixtures under
`internal/spec/testdata/red_now/`.

Both-direction fixtures (REQ-RNT-009): a `violating/` fixture whose RED cell is prose-only, and a
`legitimate/` fixture whose RED cell carries command + stdout + exit code + SHA. Confirming only
the pass direction is indistinguishable from not having raised anything.

Three additions from iteration 1, all reusing the same span-extraction seam rather than adding a
mechanism:

- **Class- AND carrier-independent command form check (D4 + N1).** Wherever a command is carried —
  table cell, fenced evidence-ledger entry, or any other fenced block — check it against the
  REQ-RNT-001 form regardless of the criterion's class. Release-blocking-only scoping let two
  broken commands through in iteration 1; cell-only scoping then let **every** command out of
  scope in iteration 2, because the D1 corrective had moved them all into the ledger. Both
  qualifiers are required; either alone is an escape hatch. A third fixture
  `testdata/red_now/ledger/` carries its commands only in a ledger, and the relocation mutant M-5
  must be observed failing.
- **Span-scoped clause predicates (D5).** The five green paths that were `grep -c <token> <file>`
  ≥1 plus an unjudgeable prose conjunct become named subtests asserting inside an extracted
  section span (MP-8 span, `### Group 4` span, `## Must-Pass Results` span). A token pasted
  elsewhere in the file no longer satisfies them.
- **Liveness anchors (D7, REQ-RNT-014).** Assert both the MP-8 sentinel span and the
  `- [PASS/FAIL/N/A] MP-8` report-template row, so MP-8's disappearance turns CI red instead of
  silently reverting to an MP-1..MP-7 PASS.

### M4 — Mirror + build (Priority: Medium) — mechanical

Hand-apply the M1 and M2 edits to the two template mirrors, run `make build`, confirm neutrality
(REQ-RNT-010, -011).

## §G Anti-Patterns

- Writing a prose RED cell **in this SPEC's own acceptance.md**. Immediate self-refutation.
- Using a `-run <TestName>` selector against a not-yet-existing test as a RED proof. It exits 0
  (spec.md §A.2). Use a file-absence or grep-count baseline instead.
- `cp`-ing `plan-auditor.md` over its mirror. The two legitimately differ.
- Absorbing t341's scope (see §H).
- Putting a shell command inside a GFM table cell. Iteration 1's two defective commands both came
  from escaped pipes in table cells; `acceptance.md` §D.0 now holds every command in a fenced
  ledger and cells cite a ledger id.
- Scoping the L1 form check to release-blocking criteria. That is the demotion escape hatch (D4).
- A green path of the shape `grep -c <token> <file>` ≥1 **AND** <prose the command cannot judge>.
  The prose half is not verified by anything; move it into a span-scoped test predicate.

## §H Cross-References — coordination notes

### t341 (lane-5) — one shared sample, two distinct axes

**The sample is shared and this is on record deliberately.** Measurement 3 of this SPEC —
`go test -run <nonexistent>` exiting 0 with `ok` — is precisely card **t341**'s subject, held by
**lane-5**, which recorded its own instance of the same behaviour as `ok ... 0.424s [no tests to
run]` at `.moai/reports/t350/discovery.md`.

**The sample is a family of nine, not one cell** (spec.md §A.2 census): AC-TOSQ-001, -002, -003,
-004, -005, -007, -008, -017, -018 all rest on the premise that an absent test turns the suite red.
AC-TOSQ-011 is excluded — its RED rests on an absent CLI verb, a different mechanism, unmeasured
and asserted neither defective nor sound.

**Two findings came from lane-5 and are credited here.** (1) The corrected refuting command
(`-run TestMigrationParityDoesNotExistXYZ`) — this card's first two iterations cited an existing
test and were refuted by lane-5 asking whether the named test existed (spec.md §A.2.1, §A.2.2).
(2) The `internal/hook/evidence_writer.go` `deriveFromOutputText` seam, where `hasPrecisePass`
matches `ok  \t` and returns before inspecting any count, so `ok … [no tests to run]` is written
to the evidence ledger as an observed PASS. That seam is **t341's surface**; this card constrains
only its own L2 verdict rule (REQ-RNT-015) and does not touch the file. lane-5 is recording the
reciprocal citation in its own `spec.md`.

The two cards use the one sample from opposite sides:

| Card | Reads the sample as | Builds |
|---|---|---|
| t341 (lane-5) | a zero-match selector produces **green** | a mechanical warning for the vacuous selector |
| t343 (this) | a cell asserting that selector produces **red** passed as release-blocking | a check on the RED cell's truth |

The axes stay distinct; t341's scope is not absorbed here. The sharing is recorded because a
later repair of `SPEC-TODO-SQLITE-001` AC-TOSQ-001 would rewrite that cell and silently destroy
the other card's evidence unless the dependency is written down.

*Verification note:* `.moai/reports/t350/discovery.md` does **not** resolve from this worktree —

```
$ ls .moai/reports/t350/discovery.md
ls: .moai/reports/t350/discovery.md: No such file or directory
```

— consistent with the dispatch's own Gaps note that lane-5's line was cited from the dispatch and
not opened from this tree. The path is recorded as a cross-worktree reference, and the `0.424s`
figure is **not** re-measured here.

### t331 — precedent, now citable by path, still not a dependency

`SPEC-TODO-LANDING-STATE-001` §C supplies the release-blocking / regression-guard disposition this
SPEC lifts into the audit layer. It **landed** with the `develop` merge and now resolves at
`.moai/specs/SPEC-TODO-LANDING-STATE-001/` (`status: completed`) on `15453140a`; the earlier
description of it as unlanded and worktree-only was true when written and is now stale.

The §C sentence was **re-read from the landed copy** rather than trusted from the pre-landing
worktree read, and is byte-identical (`acceptance.md:93-95`). AC-TLS-002 and AC-TLS-007 are the two
worked demotions.

What changed is what a reader can verify, not what this card depends on: no requirement, criterion,
or milestone here is conditioned on that file existing. Citation by path is a reader convenience.

### t338 — reused seam

`internal/spec/ac_count_clause_test.go` supplies the sentinel-extraction pattern (spec.md §C).

### t342 — the reason for REQ-RNT-002

Measured natural-language lexical discriminators unsound in both directions (7 false positives).
This is why no clause here keys on tense, mood, or a word list.


## §I Iteration-1 audit record

plan-auditor iteration 1: **FAIL, 0.75** against the Tier M threshold 0.80 (Clarity 0.75 /
Completeness 0.75 / Testability 0.50 / Traceability 1.00). MP-1..MP-7 all PASS; MP-8 self-applied
also passed 10/10 on re-execution. The FAIL came from the rubric and eight blocking defects, not
from a must-pass violation. Full report: `.moai/reports/t343/plan-audit-iter1.md`.

The most valuable finding was that MP-8 passing this SPEC was **true but empty**: both defective
commands sat in regression-guard cells, outside MP-8's declared release-blocking scope. That is
D4, and it is why L1 is now class-independent.

| Defect | Severity | Closed by |
|---|---|---|
| D1 escaped-pipe vacuous command | critical | `acceptance.md` §D.0 ledger E-02 + §D.0.1 divergence probe |
| D2 unrunnable `\| wc -l` | major | ledger E-11 (pipe-free); REQ-RNT-001 defines verbatim = raw file bytes |
| D3 AC-RNT-011 dishonest "being preserved" | minor | reworded to match AC-RNT-012 |
| D4 demotion as escape hatch | major | REQ-RNT-008 + spec.md §B — L1 binds every cell carrying a command |
| D5 token-insertion mutant on five criteria | major | span-scoped subtests; `acceptance.md` §D.2 mutant M-3 |
| D6 MP-8 trust boundary undefined | major | spec.md §B trust-boundary paragraph rewritten; REQ-RNT-013 |
| D7 continued-firing missing axis 2 | major | REQ-RNT-014 + `acceptance.md` §D.3 two-axis split |
| D8 unpinned line numbers in REQ-RNT-011 | minor | line numbers moved to AC-RNT-011 with the SHA pin |
| D9 `module: spec-audit` not path-like | minor | `module: internal/spec` |
| D10 open decisions without markers | minor | both settled in §F M1 (a)/(b) — markers not needed |

Scope change: 12 → 14 REQ, 13 → 15 AC (12 release-blocking, 3 regression-guard — AC-RNT-002,
AC-RNT-011, AC-RNT-012 all stay regression-guard; none was promoted to inflate the gate). No criterion was weakened or removed to clear the threshold; every score-bearing
change adds a mechanical predicate where prose stood.


## §J Iteration-2 audit record

plan-auditor iteration 2: score **0.800** (Tier M threshold 0.80 — the score axis clears),
MP-1..MP-7 all PASS, MP-8 self-applied reproduced 12/12. **Verdict FAIL** on one critical blocking
defect. Closure table: 7 closed (D1, D2, D3, D6, D8, D9, D10), 3 partially closed (D4, D5, D7),
**zero false completion claims**, zero monotonicity violations. Report:
`.moai/reports/t343/plan-audit-iter2.md`.

The head finding, N1, is the sharpest result of the whole audit sequence: **two correctives that
were each individually right cancelled each other.** D1 moved commands out of table cells into the
ledger; D4 scoped L1 to commands carried in cells. Composed, they left zero commands in scope, and
§D.0 made the evasion the recommended pattern. Neither review round could have caught it alone —
it exists only in the composition.

| Defect | Severity | Closed by |
|---|---|---|
| N1 correctives cancel; scope predicate reaches nothing | critical | REQ-RNT-001 admits the **ledger as a named carrier**; REQ-RNT-008 predicate made carrier-independent; AC-RNT-008 gains `TestCommandScopeIsCarrierIndependent` + `testdata/red_now/ledger/`; mutant **M-5** recorded; `acceptance.md` §D.5 restated truthfully |
| N2 three/four element mismatch | major | REQ-RNT-002 and this file's M1 body both say four |
| N3 AC-RNT-001 overclaim | major | Weakened to what holds; M-3 residual recorded in §D.2 |
| N5 expensive-command hole | major | REQ-RNT-013 gains a timeout disposition reusing the auditor's existing Bash bound — no new regime |
| N6 broken cross-reference | minor | Boundary stated where it holds (§D.3), with the reason no `spec.md` exclusion carries it |
| N7 self-matching count command | major | Counts moved to ledger E-12/E-13/E-14, anchored `^| \*\*AC-RNT-` so no prose line satisfies them |

Scope unchanged at 14 REQ / 15 AC (12 release-blocking, 3 regression-guard). No criterion was
weakened, removed, or reclassified. N3's change weakens a *claim*, not a criterion — the green path
keeps every predicate it had and stops asserting an immunity it does not have.

**Iteration cap.** `harness.yaml:77` sets the Tier M ceiling at 2 and it is reached. Whether a
third audit runs is the operator's decision via the lead. This revision is defect closure against
a delivered audit, not a bid for a further grade.


## §K Correction round (post-iteration-2, coordinator-relayed)

Four corrections arrived after the iteration-2 audit and are folded into the same revision.

| Item | What was wrong | What changed |
|---|---|---|
| C1 | The §A.2 refuting command named an **existing** test (`grep -rn "func TestMigrationParity"` → 2 hits), so it proved nothing. It survived two audits, both of which re-executed it. | Evidence replaced with `-run TestMigrationParityDoesNotExistXYZ` (→ `ok … [no tests to run]`, exit 0). **Conclusion unchanged, evidence replaced** — recorded in new §A.2.1 rather than swapped silently. |
| C2 | §A.2 framed the defect as one cell. | Reframed to the measured census: **nine** criteria on one false premise, AC-TOSQ-011 explicitly excluded and asserted neither defective nor sound. |
| C3 | L2's verdict could be routed through a seam that reads `ok  \t` as a pass before inspecting any count. | **REQ-RNT-015** + **AC-RNT-015**: L2 keys on executed-test count, never on an `ok` token. Seam cited as the reason; `evidence_writer.go` untouched (t341's surface). |
| C4 | §H credited neither the nine-cell family nor lane-5's two findings. | §H extended with both, plus the reciprocal-citation note. |
| C5 | The failure chain itself was unrecorded. | **§A.2.2** records it as the mechanism argument: four actors, rule known and being applied, defect survived every human-judgment layer — so the corrective must be mechanical, not a doctrine clause. Bounded honestly: **MP-8 would not have caught this incident either** (mutant M-6). |

Scope after the correction round: **15 REQ / 16 AC** (13 release-blocking, 3 regression-guard).
The AC count sits exactly at the Tier M ceiling of 16 — a further criterion requires tiering up or
splitting, not relaxing the budget.
