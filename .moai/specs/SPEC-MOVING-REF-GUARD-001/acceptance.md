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
V3R2-R4 under H-2 — both demoted, both silently un-escalatable. AC-MRG-005's `--strict` half would
then fail for a reason having nothing to do with the rule under test.

## §B. Baseline anchor

All criteria are decided against the frozen pre-flight baseline recorded in `plan.md` §C step 4
(`BASELINE_SHA`), never against `origin/develop` directly. A criterion here that named a moving ref
would make this SPEC an instance of its own defect.

## §C. Criteria

### AC-MRG-001 — the detector fires on the true-positive shape (MUST)

**Given** a fixture SPEC whose `acceptance.md` carries the row
`` | AC-X | `git diff --name-only origin/main -- internal/` | empty (unchanged) | ``,
**when** `go test ./internal/spec/ -run TestMovingRef_FiresOnUnpinnedAnchor` runs,
**then** exactly one `MovingRefUnpinned` finding is reported, at that file and line.

**Fails when:** the rule is not registered in `l.rules`, or the pattern misses the two-dot form.
**Mutation that must turn it red:** change the fixture's `origin/main` to `origin/mainx` — the count
must drop to 0, proving the moving-ref token is what drives the finding and not some incidental
substring of the row.

### AC-MRG-002 — a pinned claim is not flagged (MUST)

**Given** the AC-MRG-001 fixture with `origin/main` replaced by a 40-character hexadecimal SHA,
**when** the linter runs over it,
**then** zero `MovingRefUnpinned` findings are reported.

**Fails when:** the SHA-pin exclusion is absent — the rule then flags every occurrence, including
already-correct ones, which is the shape that trains readers to ignore it.
**Mutation that must turn it red:** delete the hex-SHA exclusion branch from the rule; the finding
must reappear.

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

### AC-MRG-005 — severity is warning, and the exit code follows (MUST)

**Given** a fixture corpus whose only findings are `MovingRefUnpinned`,
**when** `moai spec lint` runs, **then** it exits 0; **and when** `moai spec lint --strict` runs,
**then** it exits non-zero.

**Fails when:** the finding is emitted at `SeverityError`, which reds the 42 existing corpus
candidates (`spec.md` §B.3) on first run.
**Mutation that must turn it red:** change the emitted severity to `SeverityError`; the non-strict
exit becomes non-zero.

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

### AC-MRG-010 — corpus triage classifies without editing (MUST)

**Given** M4 complete,
**when** `.moai/reports/t342/corpus-triage.md` is read and **both** of the following are run —

```
git diff --name-only "$BASELINE_SHA"..HEAD -- .moai/specs | grep -v SPEC-MOVING-REF-GUARD-001
git status --short -- .moai/specs
```

**then** the report classifies every finding as anchor / subject / already-frozen with a one-line
reason, **and both commands return empty**.

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
**when** it is read, **then** it states all six limits L1-L6 of `spec.md` §F, and no acceptance
criterion in this SPEC asserts coverage of any of them.

**Decided by:** `grep -c 'L[1-6] —'` over the doctrine file returning 6, and a manual read of §C
confirming no criterion claims L1 or L6 coverage.
**Fails when:** L1 (refs expressed without an `origin/` token) is dropped — it is the limit most
likely to be quietly omitted, because stating it weakens the apparent value of the deliverable.
**Mutation that must turn it red:** delete the L1 paragraph; the grep count drops to 5.
**Second mutation, for the addendum:** delete L6 (a rotted reference value is indistinguishable
from a live one). L6 is the limit the R4 exemption *creates*, so omitting it would let the SPEC
introduce a blind spot in the same delivery that claims to enumerate them — the count drops to 5
and the criterion fails.

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

### AC-MRG-013 — the R4 form is not flagged (MUST)

**Given** a fixture line in the R4 form —
`base: measure at entry with git fetch origin develop (dispatch-time reference value: 44095ddc2)` —
**when** the linter runs over it,
**then** zero `MovingRefUnpinned` findings are reported.

**Fails when:** the R4 exclusion is absent, and the guard flags the very form the doctrine
recommends. That failure is worse than a plain false positive: it teaches readers to avoid the
correct remedy, which is the card's dominant failure mode arriving by a different road.
**Mutation that must turn it red:** remove the R4-form exclusion from the rule; the finding
reappears.

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

- Criteria AC-MRG-001 through AC-MRG-005 and AC-MRG-007 through AC-MRG-014 PASS; AC-MRG-006 PASSes
  or is withdrawn with the withdrawal recorded in `spec.md` HISTORY.
- **Both** of AC-MRG-013's counter-mutations were run: CM-1 observed to keep AC-MRG-001 red under the
  positional bypass, and CM-2 observed to go red under a command-token-keyed exclusion while
  AC-MRG-001, -006 and CM-1 stayed green. An R4 exclusion accepted on AC-MRG-013 alone, or on CM-1
  alone, is **not accepted** — CM-1 protects AC-MRG-001's shape and nothing else.
- AC-MRG-014's mutation was run and the all-matching mutant observed to be caught. A suite that
  passes the claim-marker-dropped mutant is not accepted regardless of its other results.
- Every criterion's stated mutation was actually planted, observed to turn the criterion red, and
  reverted — recorded in `progress.md` §E.2 with the verbatim failing output. A criterion asserted
  to be falsifiable without the mutation having been run is not evidence (`verification-claim-integrity.md` §1).
- `go test ./internal/spec/...` green; `go vet ./internal/spec/...` rc 0.
- `go build ./...` rc 0.
- This SPEC's own PRESERVE evidence cites the frozen `BASELINE_SHA`, not a moving ref.
- Open questions **Q0-Q4** (`spec.md` §H) each carry a recorded disposition: resolved, deferred with
  a reason, or escalated to the operator. Q0 is the one M3 cannot be implemented without answering,
  so its omission from this list was itself a gap.
