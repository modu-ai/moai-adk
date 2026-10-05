# Acceptance Criteria — SPEC-HEADING-REQ-COLLECT-001

Every criterion below names the command whose output decides it. Where the
deciding command is the linter, the measuring instrument is a build made FROM the
run-phase tree and invoked BY PATH — written `<build>/moai` below. A PATH-resolved
`moai` satisfies no criterion here: a stale installed build reports a clean pass
byte-identically to a fresh one (`verification-claim-integrity.md` §2.2).

`<fix>` abbreviates `.moai/reports/t894/repro/fixtures`. `<base>` is the literal
run-phase base SHA, recorded in `progress.md` before the first source edit — an
anchor at which a measurement is taken, not a moving ref.

## §A. Collection and extraction

- AC-HRC-001 (maps REQ-HRC-001, 002, 003): Given fixture H at `<fix>/SPEC-FIXH-001/`, whose `spec.md` defines `REQ-FIXH-001` as a `### REQ-FIXH-001 — <title>` heading followed by a blank line and a `SHALL`-bearing paragraph, and whose sibling `acceptance.md` maps nothing, When `<build>/moai spec lint <fix>/SPEC-FIXH-001` runs, Then stdout carries a `CoverageIncomplete` line naming `REQ-FIXH-001` and at least one non-empty-sweep witness line from an unrelated rule, and exit is 0.

  - RED-now (tree `<base>`, same command): no line names `REQ-FIXH-001` at all — the heading is uncollected, so no rule visits it. The witness line is present in both runs and is what distinguishes "collected and clean" from "never swept".
  - Green path: flipped by M1.

- AC-HRC-002 (maps REQ-HRC-004): Given the body

  ```
  ### REQ-FIXH-002 — Event-driven (When) — a section title

  **When** the trigger fires, the system SHALL respond.
  ```

  When the collector is exercised directly by the M1 unit test, Then the entry for `REQ-FIXH-002` carries `Text` equal to the paragraph (the `**When** …` sentence) and NOT equal to the heading's trailing text (`Event-driven (When) — a section title`). Both directions are asserted: an equality against the paragraph AND an inequality against the title. This is the variant-A/B decision at unit level — the two variants collect the same entry, so an assertion on collection alone cannot decide this criterion.

- AC-HRC-003 (maps REQ-HRC-005): Given a heading-form definition with no non-empty line before the next heading, When the collector is exercised by the M1 unit test, Then the entry is still produced and its `Text` is the heading's own trailing text, non-empty. An entry with empty `Text` fails this criterion.

- AC-HRC-004 (maps REQ-HRC-006): Given a body in which `### REQ-FIXH-003 — <title>` is followed immediately by `## <next section>` and only then by a `SHALL`-bearing paragraph, When the collector is exercised by the M1 unit test, Then the entry's `Text` is the fallback of AC-HRC-003 and does NOT contain any substring of the paragraph belonging to the following section. The boundary is asserted against a heading of a DIFFERENT level than the one that produced the entry, because a search bounded only by `###` crosses a `##` and passes a same-level-only test.

## §B. Provenance and severity

- AC-HRC-005 (maps REQ-HRC-007): Given fixture I at `<fix>/SPEC-FIXI-001/`, whose heading-form requirement is written so that it would trigger an error-severity code (a malformed-modality statement), When `<build>/moai spec lint <fix>/SPEC-FIXI-001` runs, Then the finding is reported at `warning` severity, the run's summary reads `0 error(s)`, and exit is 0. The exit status and the error count are both asserted: a finding reported with a `WARNING` prefix while the run exits non-zero would pass a text-only check and still mean the change started gating.

- AC-HRC-006 (maps REQ-HRC-008): Given the shipped M1 code, When the M1 test that flips every `Source` value on a fixed entry set and re-derives the severity distribution runs, Then the distribution is byte-identical before and after the flip, and

  ```
  grep -rn 'Source' internal/spec/lint.go | grep -n 'reqFindingSeverity'
  ```

  prints nothing. The flip test is the load-bearing half — the grep sees only one spelling of the prohibition, and is a tripwire, not enforcement.

## §C. No-regression, activation, and reconciliation

- AC-HRC-007 (maps REQ-HRC-003): Given the whole `.moai/specs` corpus, When the sorted `(file, line, REQ-ID, Source)` list of collected entries is produced with the `<base>` build and again with the M1 build, Then every entry present in the BEFORE list is present in the AFTER list with identical `Line`, `Text`, `Widened` and `Source`, and the AFTER list differs from the BEFORE list only by ADDED heading-source entries. A removed or mutated pre-existing entry fails this criterion.

- AC-HRC-008 (maps REQ-HRC-009, 004): Given the whole `.moai/specs` corpus, When `<build>/moai spec lint` is run over it with the `<base>` build and again with the M1 build and the per-code counts are taken for all six consuming codes (`ModalityMalformed`, `ModalityUnjudged`, `LegacyEARSKeyword`, `InvalidREQID`, `DuplicateREQID`, `CoverageIncomplete`), Then the evidence report states both counts side by side per code with the command and each build's tree SHA, AND the two discriminators below hold:

  1. **Variant discriminator.** `ModalityUnjudged`-delta ÷ newly-collected-entry-count is **below 0.50**. Measured at `9dcbc3dbe` the ratio is ~0.99 under variant A and ~0.14 under variant B (`spec.md` §A.3); the threshold sits between them, far from both, so it survives corpus movement. A ratio at or above 0.50 means the title was fed to the judge and fails this criterion.
  2. **Activation.** The total delta across the six codes is **greater than zero**, and no code's delta is negative. A zero total means nothing was activated; a negative delta means an existing finding was silenced, which no part of this SPEC authorises.

  A delta asserted without the re-run fails this criterion. A non-zero negative delta is a finding to report and adjudicate, never a number to absorb.

- AC-HRC-009 (maps REQ-HRC-010): Given the probe `internal/spec/heading_req_volume_probe_test.go` run at the SAME tree SHA as the M1 build, When its variant-B projected per-code counts are compared against the actual per-code deltas measured in AC-HRC-008, Then the evidence report states, per code, the projected value, the actual value, and their difference — and names a cause for every non-zero difference. The report also states the corpus census (documents swept, files carrying heading-form REQs, entries newly collected, files gaining an entry) **re-measured at that SHA**, not quoted from `spec.md` §A.3. Citing §A.3's `1634 / 218 / 1029 / 124` as an implementation-time measurement fails this criterion: those are attributed to `9dcbc3dbe` and the corpus moves.

- AC-HRC-010 (maps REQ-HRC-010): Given a mutant produced on a scratch copy by removing **only** the heading source from `parseREQsWithProvenance`'s merge — every other file, including the heading collector and its pattern, left byte-unchanged — When the mutant is built (`go build` exit 0, so the probe is not satisfied by a compile failure) and `<build-mutant>/moai spec lint <fix>/SPEC-FIXH-001` runs, Then the `CoverageIncomplete` line naming `REQ-FIXH-001` DISAPPEARS while the witness line remains. The evidence report quotes both the mutant's stdout and the non-mutant stdout from AC-HRC-001; a probe recorded as RED without both outputs does not discharge this criterion. The mutant is reverted and never committed.

- AC-HRC-011 (maps REQ-HRC-011): Given the shipped M1 code, When

  ```
  grep -n 'four' internal/spec/lint_req_widen.go internal/spec/lint_req_table.go
  ```

  runs, Then no surviving line claims `doc.REQs` feeds four findings or that four rules consume it, and both comment blocks name the six codes explicitly. The corrected text is read, not inferred from the grep's silence: an unrelated occurrence of the word is permitted and a zero-hit result is not by itself the pass condition.

## §D. Quality gate

- AC-HRC-GATE-001: `go test ./internal/spec/...` exits 0, with the new tests present and named in the output, and the pre-existing tests unmodified. `go test -tags heading_req_probe ./internal/spec/ -run TestHeadingREQVolume` also builds and exits 0.
- AC-HRC-GATE-002: `go vet ./internal/spec/...` exits 0 and prints nothing.
- AC-HRC-GATE-003: `golangci-lint run` over the changed files exits 0 and prints nothing new against the pre-existing baseline, which is measured in the same run-phase and quoted beside it.

## §E. Definition of Done

All of §A-§D pass; the M2 evidence report exists under `.moai/reports/t894/`,
exported to the primary checkout before it is cited, carrying the fixture outputs
with their witness lines, the six-code before/after table with both discriminators
computed, the probe-vs-linter reconciliation with every discrepancy named, and the
mutant's RED and non-RED stdout.
