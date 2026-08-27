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
**when** it is read, **then** it contains all three predicate tests of `spec.md` §D.1 (substitution,
falsification source, re-measurement expectation), all three remediation branches of §D.2, and all
four grounded instances of §D.3.

**Decided by:** `grep -c` for the three test names, the three branch labels `R1`/`R2`/`R3`, and the
four instance identifiers, in the doctrine file.
**Fails when:** the doctrine ships with the warning but without the predicate — the card's [HARD]
prohibition, and its dominant failure mode.
**Mutation that must turn it red:** delete the "falsification source" test from the doctrine; the
grep count drops and the criterion fails.

### AC-MRG-008 — the finding message names all three branches (MUST)

**Given** any emitted `MovingRefUnpinned` finding,
**when** its `Message` is asserted on in `lint_movingref_test.go`,
**then** it names pinning, freezing at pre-flight, and declaring the exemption — and does not
present pinning as the sole remedy.

**Fails when:** the message reads "pin the SHA" alone. This is the mechanism by which the dominant
failure mode reaches a reader: most people act on the message and never open the doctrine.
**Mutation that must turn it red:** shorten the message to name only pinning; the string assertion
fails.

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
**when** `.moai/reports/t342/corpus-triage.md` is read and
`git status --short -- .moai/specs` is run,
**then** the report classifies every finding as anchor / subject / already-frozen with a one-line
reason, **and** no SPEC artifact outside `SPEC-MOVING-REF-GUARD-001/` is modified.

**Fails when:** the run-phase bulk-pins the corpus — this card enacting the failure mode it exists
to prevent.
**Mutation that must turn it red:** pin a single occurrence in any other SPEC directory; the
`git status` check reports it and the criterion fails.

### AC-MRG-011 — the detection limits are stated (MUST)

**Given** the M1 doctrine section,
**when** it is read, **then** it states all five limits L1-L5 of `spec.md` §F, and no acceptance
criterion in this SPEC asserts coverage of any of them.

**Decided by:** `grep -c 'L[1-5] —'` over the doctrine file returning 5, and a manual read of §C
confirming no criterion claims L1 coverage.
**Fails when:** L1 (refs expressed without an `origin/` token) is dropped — it is the limit most
likely to be quietly omitted, because stating it weakens the apparent value of the deliverable.
**Mutation that must turn it red:** delete the L1 paragraph; the grep count drops to 4.

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

## §D. Definition of Done

- Criteria AC-MRG-001 through AC-MRG-005 and AC-MRG-007 through AC-MRG-012 PASS; AC-MRG-006 PASSes
  or is withdrawn with the withdrawal recorded in `spec.md` HISTORY.
- Every criterion's stated mutation was actually planted, observed to turn the criterion red, and
  reverted — recorded in `progress.md` §E.2 with the verbatim failing output. A criterion asserted
  to be falsifiable without the mutation having been run is not evidence (`verification-claim-integrity.md` §1).
- `go test ./internal/spec/...` green; `go vet ./internal/spec/...` rc 0.
- `go build ./...` rc 0.
- This SPEC's own PRESERVE evidence cites the frozen `BASELINE_SHA`, not a moving ref.
- Open questions Q1-Q3 (`spec.md` §H) each carry a recorded disposition: resolved, deferred with a
  reason, or escalated to the operator.
