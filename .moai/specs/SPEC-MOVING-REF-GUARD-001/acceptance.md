# SPEC-MOVING-REF-GUARD-001 — Acceptance Criteria

## §A. Falsifiability contract

[HARD] Every criterion below names two things: **the command that decides it**, and **the input that
makes it fail**. A criterion with no stated falsifying input is not accepted into this SPEC.

The reason is specific to this card rather than general hygiene. The deliverable is a guard. A guard
whose acceptance criterion cannot fail is indistinguishable from a guard that is switched off — the
criterion passes, the report reads green, and nothing is being detected. This lane has produced
vacuous controls that each passed until a mutant was planted, so the mutation is stated up front
rather than discovered by an auditor.

Fixture SPEC directories live under `internal/spec/testdata/movingref/`. Each is a minimal,
schema-valid SPEC directory; the line under test is the only thing that varies between them.

**[HARD] Fixture era precondition.** Every fixture SPEC directory MUST classify as era **V3R6** —
either by carrying `era: V3R6` in frontmatter, or by carrying a `progress.md` with `§E.2` and `§E.4`
markers and a `sync_commit_sha` value. This is not cosmetic. `internal/spec/lint.go` demotes
findings on a grandfather-era or terminal-status SPEC (`isGrandfatheredSpecDir` → `applyEraDemotion`,
which sets `Advisory = true` on every warning), and `Report.HasErrors` escalates under `--strict`
only for a warning that is **not** advisory. A "minimal, schema-valid" fixture with no `progress.md`
classifies V2.x under era heuristic H-1, and one whose `progress.md` lacks the markers classifies
V3R2-R4 under H-2 — both demoted, both silently un-escalatable.

**The precondition survives the v0.8.0 amendment, with its reason restated.** It was written to
keep AC-MRG-005's `--strict` half satisfiable; that half is now retracted (`spec.md` §D.7), and the
precondition matters for the opposite reason. `MovingRefUnpinned` is now advisory **at emission**,
so a fixture that is *also* era-demoted would be advisory for two independent reasons and could not
distinguish them: the rule's own `Advisory: true` would be unobservable behind the era path, and a
build that dropped it would still pass. Holding every fixture at V3R6 is what makes AC-MRG-005's
first mutation — drop `Advisory: true` — able to redden the criterion at all.

## §B. Baseline anchor

All criteria are decided against a frozen pre-flight literal recorded in `plan.md` §C — step 4
(`BASELINE_SHA`, the mainline point this work is measured from) or step 5 (`MERGE_BASELINE_SHA`,
the branch point beyond which only this branch's own commits lie) — never against `origin/develop`
directly. A criterion here that named a moving ref would make this SPEC an instance of its own
defect.

**Two literals, because one anchor cannot answer both questions.** `BASELINE_SHA` answers "what has
this work changed since it began", which is the right range for evidence about the deliverable.
`MERGE_BASELINE_SHA` answers "what has *this branch* added beyond mainline", which is the only
range in which a PRESERVE claim about other SPECs' directories is decidable — a range starting
before the divergence point necessarily contains mainline's own commits, and reds on them (see
AC-MRG-010). Both are recorded values with a stated re-recording obligation, never computed at
read time.

## §C. Criteria

### AC-MRG-001 — the detector fires on the true-positive shape (MUST)

**Given** a fixture SPEC whose `spec.md` carries the row
`` | AC-X | `git diff --name-only origin/main -- internal/` | empty (unchanged) | ``,
**when** `go test ./internal/spec/ -run TestMovingRef_FiresOnUnpinnedAnchor` runs,
**then** exactly one `MovingRefUnpinned` finding is reported, at that file and line.

**[HARD] The fixture lives in `spec.md`, not in a sibling artifact, and the placement is
load-bearing.** `SPECDoc.Body` carries `spec.md` alone, so a row placed there survives
AC-MRG-009's body-only mutant — which is the only arrangement under which that mutant can take
AC-MRG-009 red *while this criterion stays green*, the separation AC-MRG-009 asserts. The fixture
was originally placed in `acceptance.md`, which made that separation unobservable; the placement
was corrected at v0.7.0 and the separation then **measured** rather than asserted.

**Fails when:** the rule is not registered in `l.rules`, or the pattern misses the two-dot form.
**Mutation that must turn it red:** change the fixture's `origin/main` to `origin/mainx` — the count
must drop to 0, proving the moving-ref token is what drives the finding and not some incidental
substring of the row.

### AC-MRG-002 — a pinned claim is not flagged (MUST)

**Given** the AC-MRG-001 fixture **retaining** its `origin/main` token and carrying the anchor's
resolved 40-character hexadecimal SHA on the same line,
**when** the linter runs over it,
**then** zero `MovingRefUnpinned` findings are reported.

**[HARD] The moving-ref token is retained deliberately, and the earlier wording — "`origin/main`
**replaced by** a SHA" — was corrected at v0.7.0 because it specified a vacuous fixture.** Read
literally, replacement deletes the token, so the line fails conjuncts 1+2 and is never *exempted*
at all; this criterion's own mutation could then not turn it red, which is exactly the shape §A
exists to prevent. Retaining the ref and recording the SHA beside it is the shape filter 4 of
`spec.md` §B.3 actually removes, and the shape REQ-MRG-008 is worded for ("While a **line's moving
ref** carries a resolved SHA pin"). The implemented fixture already had the retaining form; the
correction brings the criterion's text to what was built and measured, not the reverse.

**Fails when:** the SHA-pin exclusion is absent — the rule then flags every occurrence, including
already-correct ones, which is the shape that trains readers to ignore it.
**Mutation that must turn it red:** delete the hex-SHA exclusion branch from the rule; the finding
must reappear. Observed red under the retaining fixture at v0.7.0 (`progress.md` §E.2 records the
M3 run; the v0.7.0 re-observation is in the commit body).

### AC-MRG-003 — the marker suppresses, but only with a reason (MUST)

**Given** three fixtures differing only in their marker: (a) no marker, (b)
`<!-- moving-ref-ok: the command string is the subject, not an anchor -->`, (c)
`<!-- moving-ref-ok: -->`,
**when** the linter runs over each,
**then** (a) reports `MovingRefUnpinned`, (b) reports zero findings, and (c) reports a finding
naming the marker as incomplete.

**Fails when:** the marker is honoured without requiring a reason — which makes silencing cheaper
than fixing and inverts the whole incentive.
**Mutation that must turn it red:** remove the non-empty-reason check; fixture (c) then reports
zero and the criterion fails.
**Fixture (b) is modelled on the real subject-class line** `AC-COORD-016`
(`SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md:296`, `spec.md` §B.3), so the suppression path is
exercised against an actual corpus case rather than an invented one.

### AC-MRG-004 — the three-dot form is not exempt (MUST)

**Given** a fixture using `git diff --stat origin/main...HEAD -- internal/` to decide "unchanged",
**when** the linter runs over it,
**then** a `MovingRefUnpinned` finding is reported.

**Fails when:** an implementer adds a `...` exemption on the belief that merge-base is stable under
upstream advance. It is not: `spec.md` §B.2 measured the identical wrong answer
(`3 files changed, 288 insertions(+)`) from both forms in this tree.
**Mutation that must turn it red:** add a `strings.Contains(line, "...")` early-return to the rule;
this criterion must fail immediately.

### AC-MRG-005 — the finding is advisory, and it is still reported (MUST)

*(Amended at v0.8.0. The prior form asserted `--strict` exits non-zero; that clause is retracted
with REQ-MRG-009 — see `spec.md` §D.7.)*

**Given** a fixture corpus whose only findings are `MovingRefUnpinned`,
**when** `moai spec lint` runs, **then** it exits 0 **and reports a non-zero count of
`MovingRefUnpinned` findings**; **and when** `moai spec lint --strict` runs, **then** it likewise
exits 0 **and reports the same non-zero count**.

**[HARD] Both halves are required, and they falsify in opposite directions.** Exit 0 alone is
satisfied trivially by a rule that emits nothing at all, which is precisely the confusion this
criterion must rule out: an advisory guard and a switched-off guard produce the same exit code and
are separated only by whether the finding is still in the report.

- **Half 1 — does not gate.** *Fails when:* the finding is non-advisory, so `Report.HasErrors`
  escalates it under `--strict` (`internal/spec/lint.go`: `r.Strict && f.Severity ==
  SeverityWarning && !f.Advisory`), or it is emitted at `SeverityError` and reds even the
  non-strict run. *Falsifying input:* the fixture corpus above, run under `--strict`, against a
  build in which the rule does not set `Advisory: true` — the run exits non-zero.
  **Mutation that must turn it red:** drop `Advisory: true` from the emitted finding; the
  `--strict` exit becomes non-zero while the reported count is unchanged.
- **Half 2 — still reported.** *Fails when:* the rule is disabled, unregistered, or narrowed to
  match nothing, in which case both invocations exit 0 with **zero** findings and half 1 passes
  vacuously. *Falsifying input:* the same fixture corpus against a build with the rule removed from
  the `l.rules` registration slice — both exits are 0 and the count is 0.
  **Mutation that must turn it red:** remove the rule from registration (or return `nil` from its
  `Check`); the reported count drops to 0 under both invocations while both exits stay 0.

**Decided by:** the exit code of each invocation, and the count of `MovingRefUnpinned` entries in
the same run's report, read together per invocation. A run whose count is not read has decided
half 1 only.

### AC-MRG-006 — the divergence-figure variant fires (SHOULD)

**Given** a fixture `progress.md` line citing
`` `git rev-list --count --left-right origin/main...HEAD` → 0 0 `` with no SHA and no date,
**when** the linter runs, **then** a `MovingRefUnpinned` finding is reported.

**Fails when:** only diff-shaped claims are matched and the divergence form is missed.
**Mutation that must turn it red:** append a resolved SHA to the fixture line; the finding must
disappear — proving the rule keys on the missing pin rather than on the `rev-list` verb alone.
**SHOULD-tier** per `spec.md` §H Q2: this is the least certain requirement and may be withdrawn at
run-phase without disturbing the rest.

### AC-MRG-007 — the predicate ships with the warning (MUST)

**Given** the M1 doctrine section,
**when** it is read, **then** it contains all four predicate tests of `spec.md` §D.1 (substitution,
falsification source, re-measurement expectation, read-time action), all four remediation branches
of §D.2, all five grounded instances of §D.3, and the §D.1 statement that classification and remedy
selection are two separate steps.

**Decided by:** `grep -c` for the four test names, the four branch labels `R1`/`R2`/`R3`/`R4`, and
the five instance identifiers, in the doctrine file.
**Fails when:** the doctrine ships with the warning but without the predicate — the card's [HARD]
prohibition, and its dominant failure mode.
**Mutation that must turn it red:** delete the "falsification source" test from the doctrine; the
grep count drops and the criterion fails.
**Second mutation, for the addendum:** delete Test 4 (read-time action). The S1/S2 split becomes
undecidable, so instances 3 and 5 can no longer be routed to R4 — the criterion must fail on the
test-name count.

### AC-MRG-008 — the finding message names all four branches (MUST)

**Given** any emitted `MovingRefUnpinned` finding,
**when** its `Message` is asserted on in `lint_movingref_test.go`,
**then** it names pinning, freezing at pre-flight, declaring the exemption, and stating the
measuring command — and does not present pinning as the sole remedy.

**Fails when:** the message reads "pin the SHA" alone. This is the mechanism by which the dominant
failure mode reaches a reader: most people act on the message and never open the doctrine.
**Mutation that must turn it red:** shorten the message to name only pinning; the string assertion
fails.
**Second mutation, for the addendum:** drop only R4 from the message, leaving three branches. The
assertion must still fail — R4 is the branch a reader cannot reach by intuition, and the one the
lead's own dispatch failure demonstrates is needed (`spec.md` §B.5).

### AC-MRG-009 — the rule reads sibling artifacts (MUST)

**Given** a fixture SPEC whose `spec.md` is clean but whose `progress.md` carries the flagged shape,
**when** the linter runs over the SPEC directory,
**then** the finding is reported against `progress.md`.

**Fails when:** the rule inspects only `SPECDoc.Body`, which carries `spec.md` alone — most real
occurrences live in `progress.md` and `acceptance.md` (`spec.md` §B.3), so a body-only rule would
miss the majority of the corpus while appearing to work.
**Mutation that must turn it red:** restrict the rule to `doc.Body`; this criterion fails while
AC-MRG-001 still passes — which is exactly why it is a separate criterion.

**The separation is measured, not asserted, and it required a fixture correction to become
observable.** At M3 the body-only mutant took **seven** criteria red at once, AC-MRG-001 among
them, because AC-MRG-001's fixture row was specified in `acceptance.md` while `SPECDoc.Body`
carries `spec.md` alone — the clause above was then unsatisfiable as written. Rather than delete
the clause, the fixture was re-placed into `spec.md` at v0.7.0 (AC-MRG-001, above) and the mutant
re-run. Observed: AC-MRG-009 red, **AC-MRG-001 green**, alongside AC-MRG-002, -005, -008 and -014.
AC-MRG-003, -004 and -006 remain collateral-red — their fixtures also live in sibling artifacts,
which is the same property this criterion names rather than a separate signal. The clause is
therefore kept because it now holds, not because it was left standing.

### AC-MRG-010 — corpus triage classifies without editing (MUST)

**Given** M4 complete,
**when** `.moai/reports/t342/corpus-triage.md` is read and **both** of the following are run —

```
git diff --name-only "$MERGE_BASELINE_SHA"..HEAD -- .moai/specs | grep -v SPEC-MOVING-REF-GUARD-001
git status --short -- .moai/specs
```

**then** the report classifies every finding as anchor / subject / already-frozen with a one-line
reason, **and both commands return empty**.

**[HARD] The first command's anchor was corrected at v0.7.0, from `$BASELINE_SHA` to
`$MERGE_BASELINE_SHA` — the second frozen literal recorded at pre-flight (`plan.md` §C step 5).**
`$BASELINE_SHA` names a point on mainline *before* this branch diverged, so its range spans
mainline as well as this branch, and the command returned **four paths** at every measurement:
`SPEC-GRAPH-FRESHNESS-CADENCE-001/{spec,progress}.md` and
`SPEC-SYNC-STRATEGY-KEY-001/{spec,progress}.md`, from commits `f9c827217`, `28f16d030`,
`bc66c30b7` (card t322) and `ed68889e3` (card t303) — all landed on mainline, absorbed by this
branch's merge commits, none authored here (`merge-base --is-ancestor` rc 0 on all four,
`progress.md` §E.2 M4 Claim 4).

That is a **wrong-reason red** in the `verification-completeness.md` §2 sense: red on arrival, red
regardless of what this work does, and — worse — unable to distinguish the mutation this criterion
names (pin a line in another SPEC directory *and commit it*, on this branch) from a routine
mainline merge, since both produce the same four-path output. A decider that cannot tell its own
mutation from ordinary background motion decides nothing.

`$MERGE_BASELINE_SHA` measures only what this branch adds beyond mainline, and returns empty.
**It remains a *recorded* value, not a computed one** — that distinction is the whole point of
`plan.md` §C's freeze discipline, and a merge-base recomputed at read time would reintroduce the
moving anchor this SPEC exists to prohibit. Its own residual is stated rather than hidden: the
literal goes stale the moment this branch absorbs mainline again, so **any further absorption
obliges re-recording it in `plan.md` §C step 5**, dated, before this criterion is decided again.
That is limit L6 acting on this SPEC's own evidence, handled by R2 + R4 — the command is the
criterion, the literal is a dated reference.

`git status --short -- .moai/specs` is unchanged and stays as the second surface: it is what
catches an *uncommitted* edit, which no committed-diff range can see.

**Fails when:** the run-phase bulk-pins the corpus — this card enacting the failure mode it exists
to prevent.
**Mutation that must turn it red:** pin a single occurrence in another SPEC directory **and commit
it**. The committed form is the one that matters: `git status` alone reads the working tree only, so
a committed edit — the exact shape plan.md's [HARD] M4 clause forbids — leaves `git status` clean and
the criterion passing green. The two-surface decider is taken from `SPEC-DESIGN-MOAIWEBV2-002`
AC-MWA-007a, which pairs the same two commands for the same reason.

**Traces to:** REQ-MRG-011.

### AC-MRG-011 — the detection limits are stated (MUST)

**Given** the M1 doctrine section,
**when** it is read, **then** it states all seven limits L1-L7 of `spec.md` §F, and no acceptance
criterion in this SPEC asserts coverage of any of them.

**Decided by:** `grep -c 'L[1-7] —'` over the doctrine file returning 7, plus
`grep -c 'the ANCHOR branch of the predicate is unvalidated' <doctrine>` returning 1 for L7
specifically, and a manual read of §C confirming no criterion claims L1 or L6 coverage.

**[HARD] The second half was `grep -c 'ANCHOR' <doctrine>` ≥ 1 until v0.7.0, and it guarded
nothing.** `ANCHOR` is one of the predicate's two **class names**, so it appears throughout the
tests, the tie-break, the remedy table and the instances. Measured against the shipped doctrine
with L7 deleted, the count half correctly drops 7 → 6 while `grep -c 'ANCHOR'` returns **9** — the
`≥ 1` check is satisfied by any doctrine that publishes the predicate at all, with or without L7.
The replacement keys on L7's **distinguishing phrase** rather than on the class name. It is
deliberately *not* keyed on `L7 —`, which would only re-measure what the count half already
measures and would give the criterion two halves reading the same axis.

**Fails when:** L1 (refs expressed without an `origin/` token) is dropped — it is the limit most
likely to be quietly omitted, because stating it weakens the apparent value of the deliverable.
**Mutation that must turn it red:** delete the L1 paragraph; the grep count drops to **6**. (The
criterion said "to 5" until v0.7.0; the limit set grew 5 → 6 at v0.2.0 and 6 → 7 at v0.4.0 and the
mutation text was not swept with it. The criterion still went red throughout, so falsifiability
held — only the stated figure was wrong.)
**Second mutation, for the addendum:** delete L6 (a rotted reference value is indistinguishable
from a live one). L6 is the limit the R4 exemption *creates*, so omitting it would let the SPEC
introduce a blind spot in the same delivery that claims to enumerate them — the count drops to 6
and the criterion fails.
**Third mutation, for L7:** delete L7 (the ANCHOR branch is unvalidated). Before iter-2 this
limitation lived only in `spec.md` §D.3 prose and `grep -n 'ANCHOR' acceptance.md` returned nothing,
so M1 could have published the doctrine without it while passing every criterion — the one
limitation the audit round added was the one thing droppable. The count drops to 6 **and** the
distinguishing-phrase check returns 0; both halves must fail. Observed at v0.7.0: count `7 → 6`,
phrase `1 → 0`, with the retired `grep -c 'ANCHOR'` key holding at **9** throughout — which is the
measurement that condemns the old half rather than an argument about it.

### AC-MRG-012 — template mirror and neutrality (MUST)

**Given** M5 complete,
**when** `make build` runs and the template-neutrality guard executes,
**then** the doctrine exists under `internal/template/templates/.claude/rules/moai/core/`, the build
exits 0, and the neutrality check reports no forbidden content.

**Fails when:** the mirrored copy retains a SPEC ID, an internal date, or a commit SHA from the
local doctrine — all four grounded instances in §D.3 name SPEC IDs and SHAs, so the neutralization
is not cosmetic here.
**Mutation that must turn it red:** copy the local doctrine verbatim into the template; the
neutrality guard reports the SPEC-ID and SHA tokens.

### AC-MRG-013 — the R4 form is not flagged (delivered by follow-up card t353, 2026-09-02)

> **DELIVERED — adopted and decided by follow-up card t353 (2026-09-02).** The criterion below,
> retained verbatim since v0.7.0, was the specification the follow-up ran against and is now a
> decided criterion of this SPEC. Implementation at commit `0026dc7b9` (branch
> `WT-r4-imperative-exempt`): `internal/spec/lint_movingref.go` keys the exclusion on
> **imperative structure** — both conjuncts, an imperative measuring directive plus a
> syntactically demoted dated reference — and never on a command token, exactly per the [HARD]
> constraint the v0.7.0 deferral recorded. RED observed pre-fix, GREEN after (13 MovingRef
> tests); both counter-mutations run and observed blocking their bypasses (CM-1 positional,
> CM-2 command-token). Corpus lint 115→113 findings, with exactly the two genuinely-R4 lines
> newly exempted (`SPEC-IGNORED-EVIDENCE-CITATION-001/progress.md:480`, self-declared R4;
> `SPEC-SPECLINT-GITBLIND-001/progress.md:233`, Korean-form) and zero over-exemption. Lane
> verdict: `.moai/reports/t353/verdict.md` (commit `d191d28b4`); implementation record:
> `.moai/reports/t353/impl-record.md`.
>
> **Why:** `spec.md` §B.7 measured R4's reachable class as **0 of 42** candidate lines on two
> independent probes, and M4 re-measured it against the rule's own 97 external findings with the
> same result — **external R4 occupancy 0**. With no occupants the exclusion can only
> over-exempt today, never under-exempt, so every error available to it is a D11-shaped bypass —
> and the iter-1 audit measured that bypass concretely: an exclusion keyed on the fetch verb passes
> all thirteen criteria, counter-mutation included, while silencing **76 of 117** real unpinned
> divergence lines. Deferring removes the bypass entirely, because no exclusion exists to key
> wrongly, and replaces Q0's unsolved shape-recognition problem with the already-solved marker one.
>
> **Meanwhile:** early R4-form lines are silenced with the R3 marker
> (`<!-- moving-ref-ok: <reason> -->`), whose mandatory reason records *why* the line is S2. R4
> itself is unaffected — it remains a doctrine remedy and stays named in the finding message
> (REQ-MRG-004 / AC-MRG-008); only its **lint exclusion** is deferred.
>
> **Resume condition (historical):** the condition was measured **MET on 2026-09-02** — live
> external R4-form occupants exist (the two lines named in the delivered-record header above,
> previously silenced via R3 markers), the exclusion acquired something to under-exempt, and the
> option-C trade-off inverted. The operator issued follow-up card t353 and it has delivered; see
> the delivered-record header. The original condition, as recorded at v0.7.0: reconsider when the
> R4 form is actually observed in the corpus. M4 gave this a sharper reading than §B.7 could: the
> rule found **0 external S2 lines** but **2 S2 lines inside this SPEC's own directory**, both
> R4-form — exactly what §B.7 predicted when it wrote that the class is "populated
> prospectively, by the R4 remediations M1's doctrine will produce". The class began to fill from
> this card's own output, and has since filled externally.
>
> No follow-up card is issued from here: card issuance is the operator's act.

**Given** a fixture line in the R4 form, drawn from REQ-MRG-010's residual class —

```
- verify `internal/hook` is unchanged by this work: run `git diff --name-only origin/develop -- internal/hook` at read time (reference reading 2026-08-28: empty)
```

**when** the linter runs over it,
**then** zero `MovingRefUnpinned` findings are reported.

**Fixture properties, measured (not asserted) — the fixture MUST be flaggable absent the exclusion.**
The previous fixture (the lead's dispatch line) was vacuous three ways at once, so this replacement's
properties are measured and recorded rather than reasoned about:

| Property | Required | Measured | Command |
|---|---|---|---|
| moving-ref slash token | ≥1 | **1** | `grep -cE 'origin/[a-z]'` |
| hex SHA 7-40 | **0** (else REQ-MRG-008 exempts it) | **0** | `grep -cE '\b[0-9a-f]{7,40}\b'` |
| invariant-claim marker | ≥1 | **1** | filter-3 alternation, `-i` |
| git-command context | ≥1 | **1** | `grep -cE 'git [a-z-]+[^\`]*origin/(main\|develop\|HEAD)'` |
| frozen-baseline variable | **0** | **0** | `grep -cE '\$[A-Z_]*BASELINE[A-Z_]*'` |
| **full REQ-MRG-001 pipeline** | **1** — would be flagged absent R4 | **1** | filters 2-4 chained |

The last row is the one that makes this criterion non-vacuous, and it is the row the old fixture
failed (it measured **0**). The date in the reference is not this line's exemption route: REQ-MRG-001
carries no date conjunct, and the line is not in the divergence class (`grep -c 'rev-list --count
--left-right'` → 0), so **the R4 exclusion is the only thing that can exempt it**.

**[HARD] This criterion is synthetic-only, and that is a limit on what it can establish.** R4's
reachable scope in the live corpus is **empty — 0 of 42 candidate lines, two independent probes**
(`spec.md` §B.7). The fixture above therefore has no live counterpart and cannot acquire one until
the doctrine is in use. The criterion proves the exclusion *works* against a constructed line; it
does **not** demonstrate the exclusion is *needed*. Those are different claims and this criterion
supports only the first. `spec.md` §H Q0 carries the resulting scope decision — option C would defer
this criterion together with REQ-MRG-010.

**Fails when:** the R4 exclusion is absent, and the guard flags the very form the doctrine
recommends. That failure is worse than a plain false positive: it teaches readers to avoid the
correct remedy, which is the card's dominant failure mode arriving by a different road.
**Mutation that must turn it red:** remove the R4-form exclusion from the rule; the finding
reappears. Against the *previous* fixture this mutation produced no finding at all, which is what
made the criterion vacuous — the exemption could have been entirely absent and every criterion would
still have passed. Run this mutation first, before CM-1 and CM-2: it is the only one that proves the
exclusion **exists**, where those two prove it is not too wide.

**Counter-mutation set — TWO fixtures, both required.** One is insufficient, and the audit proved
it rather than argued it.

**CM-1 (positional bypass).** `git diff --name-only origin/main -- internal/ (unchanged)` — mentions
a git command and a parenthesis but states a *result*, not an instruction to measure. MUST still be
flagged. This blocks the exclusion shape "a git verb before the ref and a parenthesized value after
it", which would otherwise exempt AC-MRG-001's fixture too.

**CM-2 (command-token bypass).** A fetch chained to a divergence read, carrying a claim:
`git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → `0 0` (no
divergence). MUST still be flagged.

CM-2 exists because CM-1 protects AC-MRG-001's *shape* and nothing else. An exclusion keyed on the
**fetch verb** passes AC-MRG-001, -006, -013 and CM-1 — no other fixture carries a fetch verb — while
silencing **76 of 117** unpinned divergence lines in the live corpus (`spec.md` §B.6, re-measured at
`43329ec8b`). That is the largest real class of this defect, and it would have shipped green.

**Mutation that must turn CM-2 red:** key the R4 exclusion on any command token — `git fetch`, or
any other verb — instead of on imperative structure. CM-2 must go green-to-red, while AC-MRG-001,
-006 and CM-1 all stay passing. **An R4 exclusion accepted without CM-2 red under this mutation is
not accepted** (see DoD).

**Open:** how imperative structure is *recognized* is `spec.md` §H Q0, unresolved at plan-phase.
What is no longer open is that it MUST NOT key on a command token — a token key is forgeable by
construction. This criterion fixes the behaviour required of whatever signature is chosen; it does
not assert the signature exists yet.

### AC-MRG-014 — negative control on the invariant-claim conjunct (MUST)

**Given** a fixture line naming a moving ref inside a git-command context but carrying **no**
invariant-claim marker — a plain instruction, `git fetch origin main && git rev-list --count
--left-right origin/main...HEAD`, with no assertion about what the result was —
**when** the linter runs over it,
**then** zero `MovingRefUnpinned` findings are reported.

**Fails when:** the rule drops the claim-marker conjunct and flags any unpinned moving ref in a
git-command context. REQ-MRG-001 is a four-way conjunction; three conjuncts have a negative control
(AC-MRG-002 the SHA pin, AC-MRG-003(b) the marker, AC-MRG-013 CM-1/CM-2 the R4 form) and until now
this one had none.
**Mutation that must turn it red:** delete the claim-marker conjunct from the rule. The mutant is
not exotic — it is the *simpler* rule, and it passes AC-MRG-001 (fires), AC-MRG-001's own mutation
(`origin/mainx` → 0), and AC-MRG-002, -004, -005, -006, -008, -009, -013 unchanged. Measured over
`.moai/specs/**` at `43329ec8b`, excluding this SPEC's directory, that mutant yields **495** findings
against the **42** §D.5 sizes its severity argument on — a ~12× over-fire that the entire suite
passes green, and one that defeats §D.5 directly, since bulk suppression is the rational response to
495 warnings.

This is the `verification-completeness.md` §2 mutant-probe defect: an invalid-cases-only suite
passes an all-matching mutant.

## §D. Definition of Done

- Criteria AC-MRG-001 through AC-MRG-005, AC-MRG-007 through AC-MRG-012, and AC-MRG-014 PASS;
  AC-MRG-006 PASSes or is withdrawn with the withdrawal recorded in `spec.md` HISTORY (KEPT at M3).
  **AC-MRG-013 is excluded from this list**: it is deferred with REQ-MRG-010 per `spec.md` §H Q0
  option C, and a deferred criterion is neither a PASS nor a FAIL.
- **AC-MRG-013's counter-mutation obligation is deferred, not discharged.** CM-1 and CM-2 were NOT
  run, because no R4 exclusion was built for them to probe. The obligation travels with the
  requirement: whichever work later implements REQ-MRG-010 inherits it unchanged — **both**
  counter-mutations required, CM-1 observed to keep AC-MRG-001 red under the positional bypass and
  CM-2 observed to go red under a command-token-keyed exclusion while AC-MRG-001, -006 and CM-1
  stay green. An R4 exclusion accepted on AC-MRG-013 alone, or on CM-1 alone, is **not accepted**
  there either — CM-1 protects AC-MRG-001's shape and nothing else.
- AC-MRG-014's mutation was run and the all-matching mutant observed to be caught. A suite that
  passes the claim-marker-dropped mutant is not accepted regardless of its other results.
- Every criterion's stated mutation was actually planted, observed to turn the criterion red, and
  reverted — recorded in `progress.md` §E.2 with the verbatim failing output. A criterion asserted
  to be falsifiable without the mutation having been run is not evidence (`verification-claim-integrity.md` §1).
- `go test ./internal/spec/...` green; `go vet ./internal/spec/...` rc 0.
- `go build ./...` rc 0.
- This SPEC's own PRESERVE evidence cites a frozen literal — `BASELINE_SHA` or, for AC-MRG-010's
  first surface, `MERGE_BASELINE_SHA` (`plan.md` §C steps 4-5) — never a moving ref.
- Open questions **Q0-Q4** (`spec.md` §H) each carry a recorded disposition: resolved, deferred with
  a reason, or escalated to the operator. Q0 is the one M3 cannot be implemented without answering,
  so its omission from this list was itself a gap. **Q0 is answered — option C, by operator
  decision, recorded at `spec.md` §H with its grounds and resume condition; Q1 resolved at M2; Q2
  resolved at M3 (AC-MRG-006 KEPT); Q3 and Q4 stand open by record.**
- **Every criterion corrected at v0.7.0 was re-observed falsifiable after the correction.** A
  criterion whose mutation is re-run only before the correction is indistinguishable from one
  vacuous criterion swapped for another; the re-observation is what separates them. Applies to
  AC-MRG-002, AC-MRG-009 and AC-MRG-011 (mutations re-planted, observed red, reverted) and, for
  AC-MRG-010 and AC-MRG-011's stale figure, to the measurement each correction rests on.
