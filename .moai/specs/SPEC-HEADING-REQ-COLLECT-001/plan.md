# Implementation Plan — SPEC-HEADING-REQ-COLLECT-001

Card t894. Tier M. Ordered by decision-reversibility: the text-extraction decision
and its bounding rule lead, the provenance wiring follows, and the mechanical
measurement and comment corrections come last.

## §A. Context

One collection gap on the definition-collector path, plus the extraction decision
collecting it forces. The measurements live in `spec.md` §A and are NOT restated
here; the instrument is `internal/spec/heading_req_volume_probe_test.go` at tree
`9dcbc3dbe`.

Prior cards whose decisions this card inherits and must not reverse:

- **t518** established the merge-by-line architecture and the `Widened` / `Source`
  separation, including the [HARD] prohibition that `Source` never decides
  severity. This card adds a third source under that architecture.
- **t385** established the list-form separator lexicon (`(?:—|:)` plus optional
  classifier). The heading pattern mirrors it rather than re-deriving it.
- **t801** (`SPEC-SIBLING-MAPS-SHORTHAND-001`) owns the sibling coverage extractor.
  No file is touched by both cards.

## §B. The decision this card makes: what becomes `REQEntry.Text` (highest reversibility)

Variant **B** — the first non-empty paragraph below the heading — is adopted. The
rejection of variant A is a measurement, recorded in `spec.md` §A.3, not a
preference: A fires `ModalityUnjudged` on 99.2% of collected entries and finds 28
fewer `ModalityMalformed` defects.

Three properties of the extractor are the SPEC's and are not negotiable at
run-phase; the exact implementation is the run phase's to write.

1. **Bounded above by its own heading.** The search starts on the line after the
   heading that produced the entry.
2. **Bounded below by the next heading of any level, or EOF** (REQ-HRC-006). An
   unbounded search reaches into the following section and attributes a
   neighbouring requirement's sentence to this one — a manufactured judgment,
   which is worse than the silence being repaired.
3. **Fallback is the heading's own trailing text, never the empty string**
   (REQ-HRC-005). An empty `Text` would be judged `ModalityUnjudged` for a reason
   that has nothing to do with the author's writing, re-manufacturing exactly the
   constant-signal failure variant A produces.

"Paragraph" means the run of consecutive non-empty lines beginning at the first
non-empty line inside those bounds. Whether it is joined with spaces or newlines
is a run-phase detail; what is fixed is that the modality judge receives a
statement rather than a title.

### The merge shape

`mergeREQsByLine` is two-way. A third source is merged by composition —
`merge(merge(list, table), heading)` — which is sound because each input is
line-sorted and the helper's output is line-sorted. A three-way rewrite of the
helper is permitted but not required; if it is rewritten, AC-HRC-007's
no-regression evidence covers it.

## §C. Pre-flight

- Re-run the probe from the run-phase HEAD and record the corpus census FROM THAT
  RUN: `go test -tags heading_req_probe ./internal/spec/ -run TestHeadingREQVolume -v`.
  The `spec.md` §A.3 numbers are attributed to `9dcbc3dbe` and are a reference, not
  a baseline for this run.
- Build the measuring instrument FROM the run-phase tree and invoke it BY PATH:
  `go build -o <scratch>/moai ./cmd/moai`. A PATH-resolved `moai` satisfies no
  criterion here — a stale installed build reports a clean pass byte-identically
  to a fresh one.
- Record the whole-corpus BEFORE per-code counts with that build, at the run-phase
  base SHA, before any source edit.
- Re-read `git rev-parse --short HEAD` immediately before any commit; never reuse
  a value read earlier in the turn.

## §D. Milestones

### M1 — the heading collector and its text extractor

Files:

- NEW `internal/spec/lint_req_heading.go` — the heading pattern, the collector, and
  the bounded paragraph extractor. Its header comment states the variant-A/B
  decision with the measured ratio, why the search is bounded by the next heading,
  and why the fallback is the heading text rather than the empty string.
- `internal/spec/lint_req_widen.go` — add the third source to
  `parseREQsWithProvenance`'s merge, and correct the stale "four findings" comment
  to name the six (REQ-HRC-011).
- `internal/spec/lint.go` — the new `REQSource` value plus its `String()` case.
  Nothing else in this file changes; the three consuming loops are untouched.
- `internal/spec/lint_req_table.go` — stale "four rules" comment corrected
  (REQ-HRC-011). No behavioural change.
- NEW `internal/spec/lint_req_heading_test.go` — the positive case, the
  title-vs-paragraph assertion (AC-HRC-002), the no-paragraph fallback, the
  next-heading boundary control, an entry-level `Widened` assertion, and the
  `Source`-flip severity-distribution test mirroring the existing
  `TestTableCollection_SourceDoesNotDecideSeverity`.

Fixtures for the end-to-end criteria live under `.moai/reports/t894/repro/fixtures/`,
each a `spec.md` (+ `acceptance.md` where coverage is exercised) pair shaped like
card t801's fixtures. Every fixture criterion asserts a non-empty-sweep witness
line alongside its own expectation: a lint run that printed nothing at all could
mean the rule set never ran, and a zero read from an empty sweep asserts nothing.

### M2 — evidence

Files: `.moai/reports/t894/` only — no source change.

- The fixture runs, stdout quoted verbatim with exit codes and witness lines.
- The whole-corpus BEFORE/AFTER per-code table for all six codes, with the command,
  the build's tree SHA, and the run-phase base on both sides.
- The ratio discriminator computed from that table (AC-HRC-008).
- The probe-vs-linter reconciliation at the named SHA, naming every discrepancy
  rather than only the agreements (AC-HRC-009).
- The mutant probe: the one-line removal of the heading source from the merge,
  `go build` exit 0, the fixture's stdout with the findings gone, plus the
  non-mutant stdout beside it. The mutant is reverted and never committed.

Evidence is exported to the PRIMARY checkout before it is cited — `.moai/reports/*`
is gitignored, so an in-worktree copy reaches no clone and the citation would not
resolve at audit time.

## §E. Self-verification

- `go test ./internal/spec/...` — the affected package only. The full-suite verdict
  is CI's, per the repository's test-execution rule.
- `go test -tags heading_req_probe ./internal/spec/ -run TestHeadingREQVolume` — the
  probe still builds and runs after the collector lands.
- `go vet ./internal/spec/...`
- `golangci-lint run` scoped to the changed files, with the pre-existing baseline
  measured in the same run-phase and quoted beside it.

## §F. Risks

| Risk | Consequence | Mitigation |
|---|---|---|
| Variant A ships by accident (the title is used as `Text`) | 1021 constant findings; the rule reads as noise and gets ignored — the card's [HARD] failure | AC-HRC-002 at unit level; AC-HRC-008's ratio discriminator at corpus level. A collection-count check cannot see this: both variants collect the same entries |
| The paragraph search is unbounded | A neighbouring section's sentence is judged as this requirement's — a manufactured judgment | REQ-HRC-006; AC-HRC-004 boundary control |
| The fallback yields empty `Text` | `ModalityUnjudged` fires for a collector reason, re-manufacturing the constant signal | REQ-HRC-005; AC-HRC-003 |
| `Widened` is not set on heading entries | Four error-severity codes arrive at error severity and the change starts gating | REQ-HRC-007; AC-HRC-005 measures exit status and the error count, not just the finding text |
| `Source` is read by a severity decision | Two severity axes; the first rule that forgets one lets an advisory finding gate | REQ-HRC-008; AC-HRC-006's flip test |
| The merge perturbs existing list/table entries | A silent corpus-wide shift attributed to this card | AC-HRC-007's before/after id-set diff |
| Volume pressure produces a staging flag mid-run | The repair reports as done while withholding findings | REQ-HRC-009; AC-HRC-008. Every affected code is already advisory, so there is no gate to protect |
| §A.3's numbers are re-cited as implementation-time measurements | An unattributed claim: the corpus moved | REQ-HRC-010; AC-HRC-009 requires the re-measurement and names the SHA |
| The probe's transcription of the six conditions is unfaithful | Every projection in §A.3 is wrong in an unknown direction | AC-HRC-009 reconciles against the real linter, not against the probe |

## §G. Anti-patterns

- Shipping collection without the extraction decision, on the grounds that
  collection is "the minimal change". An incomplete collection is not a minimal one.
- Adding a flag, cap, or allowlist to soften the 651. Every affected code is already
  advisory; the softening buys nothing and costs the gate.
- Reporting the corpus AFTER number without re-measuring the BEFORE with the same
  instrument at the same base.
- Quoting `spec.md` §A.3 as if it were measured in the run phase.
- Reading a fixture run that printed nothing as a pass.
- Widening the anchor to `##` or `####` because it "seems consistent". Neither has a
  measured population.

## §H. Cross-references

- `internal/spec/lint.go:688` — `doc.REQs = parseREQsWithProvenance(body)`, the live entry point
- `internal/spec/lint.go:774` — `reqFindingSeverity`, the single severity decision point
- `internal/spec/lint.go` `REQEntry` / `REQSource` — the `Widened` / `Source` separation and its [HARD] prohibition
- `internal/spec/lint_req_widen.go` — `parseREQsWide`, `parseREQsWithProvenance`, the stale "four" comment
- `internal/spec/lint_req_table.go` — `parseREQsTable`, `mergeREQsByLine`, the stale "four" comment
- `internal/spec/heading_req_volume_probe_test.go` — the measuring instrument and the reconciliation target
- Cards t518, t385 — collector lineage; card t801 (`SPEC-SIBLING-MAPS-SHORTHAND-001`) — the sibling axis
