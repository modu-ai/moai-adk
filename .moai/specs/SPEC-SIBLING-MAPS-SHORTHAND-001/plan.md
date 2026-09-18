# Implementation Plan — SPEC-SIBLING-MAPS-SHORTHAND-001

Card t801. Tier M. Ordered by decision-reversibility: the one design decision
this SPEC carries — what textual unit bounds the expansion — leads; the
mechanical measurement work comes last.

## §A. Context

One collection gap on the sibling coverage path, measured at tree HEAD
`881aa4bb8` with a path-invoked build. See `spec.md` §A for the measurements
themselves — they are not restated here.

Two prior cards fixed their own slice and left decisions this card must honour:

- t518 established the merge-by-line definition-collector architecture and the
  `Widened` / `Source` separation. Nothing in this card touches it.
- t561 decided `ExtractRequirementMappings` is IMMUTABLE and added a sibling-only
  extractor instead. Its `cellREQIDs` already implements the exact numeric-tail
  expansion this card needs, scoped to one table cell.

The sibling card **t894** carries the heading-form definition axis. It edits the
definition collectors; this card edits the sibling coverage extractor union. No
file is touched by both.

## §B. The decision this card makes: the textual unit (highest reversibility)

`cellREQIDs` is safe because it is handed ONE TABLE CELL. `fullREQIDPattern`
matches every `REQ-…` token in whatever string it receives, so reuse on the
`maps` path needs a unit supplied from outside — and the obvious candidates are
both wrong:

- **The line** is wrong. `- AC-X-001 (maps REQ-X-001) — verifies REQ-X-009, 010 are unreachable.` would count `REQ-X-009` and an expanded `REQ-X-010` as covered, suppressing two genuine findings. That is a manufactured silence: the exact defect class this card exists to close.
- **`reqSectionPattern`'s capture** is wrong. Its truncation at the first non-`REQ-` element IS the defect being repaired, and REQ-SMS-003 forbids editing it.

The decision: the unit is the capture of a **sibling-local widened `maps`
locator**, shaped like

```
(?i)maps[ \t]+(REQ-[A-Z0-9-]+(?:[ \t]*,[ \t]*(?:REQ-[A-Z0-9-]+|[0-9]+))*)
```

The exact regex is the run phase's to write; the properties below are the SPEC's
and are not negotiable at run-phase:

1. **Anchored at `maps`.** Tokens before `maps` on the line are outside the unit.
2. **A contiguous comma-separated run.** The capture ends at the first element that is neither a full `REQ-…` id nor a bare numeric tail, so trailing prose (`— verifies REQ-X-009, 010 …`) is outside it.
3. **Horizontal whitespace only between elements.** `\s` would admit `\n` and let a section ending in a comma absorb the next list item. `[ \t]` bounds the unit to one line (REQ-SMS-005).
4. **Enumeration reads the capture alone.** It never receives the line. This is what makes reuse of `cellREQIDs` sound rather than merely convenient.

Property 3 also makes the reuse of `numericTailPattern` (`^\s*,\s*([0-9]+)\b`,
whose `\s` DOES admit `\n`) safe without editing it: the capture handed to the
enumerator contains no newline, so the enumerator cannot cross one. The
line-bounding is enforced once, at the locator, and inherited.

Each `maps` section is its own unit; a tail in one section never takes a prefix
from a previous one (REQ-SMS-005).

## §C. Pre-flight

- Confirm the fixture set is reachable: `.moai/reports/t801/repro/fixtures/SPEC-FIX{A,B}-001`.
- Rebuild the measuring instrument from the run-phase HEAD and invoke it BY PATH.
- Record the before numbers from the run-phase base, not from memory: `2018` `CoverageIncomplete`, `3 error(s), 3151 warning(s)` at `881aa4bb8` (`.moai/reports/t801/repro/corpus-before-counts.txt`).
- Re-read `git rev-parse --short HEAD` immediately before any commit; do not reuse a value read earlier in the turn.

## §D. Milestones

### M1 — the locator, the expander, and its unit

Files:

- NEW `internal/spec/lint_coverage_sibling_maps.go` — the widened `maps` locator and the expander. The header comment states the four unit properties above, why the line and `reqSectionPattern` were both rejected as the unit, and the declined forms mirrored from `spec.md` §D (REQ-SMS-006).
- `internal/spec/lint_coverage_sibling_table.go` — if REQ-SMS-002 is satisfied by extraction rather than by calling `cellREQIDs` directly, the shared helper is lifted here or into the new file and MUST be named `expandNumericTails` (the name AC-SMS-010 greps for); either way exactly one numeric-tail rule survives (C2), and the new `maps`-path file must *name* the shared symbol — `cellREQIDs` when called directly, `expandNumericTails` when extracted. `numericTailPattern` is not edited.
- `internal/spec/lint_coverage_sibling.go` — add the third extractor to `siblingAcceptanceCoveredREQIDs`'s union alongside `ExtractRequirementMappings` and `siblingTableREQIDs`, with a comment naming the inline-path residual and citing the measured 0-instance census.
- `internal/spec/ears.go` — UNCHANGED, verified by diff against the pinned base, not by intention (AC-SMS-006).
- NEW `internal/spec/lint_coverage_sibling_maps_test.go` — the positive case, the no-preceding-id negative control, the over-reach control, the line-boundary control, a case asserting a full-id-only `maps` list produces exactly the ids it produces today (the 1099-occurrence no-regression direction), and `TestSiblingMapsExpansionSharesTheTableRule` — the shared-rule test AC-SMS-010 names by exact string, exercising the table path and the `maps` path on the same input through the same helper.

New fixtures under `.moai/reports/t801/repro/fixtures/`: `SPEC-FIXE-001`
(no preceding id), `SPEC-FIXF-001` (over-reach), `SPEC-FIXG-001` (line boundary),
each a `spec.md` + `acceptance.md` pair shaped like fixtures A and B.

### M2 — evidence

Files: `.moai/reports/t801/` only — no source change.

- The four fixture runs, stdout quoted verbatim with exit codes, each showing its `MissingExclusions` witness line.
- Fixture B before/after, quoted on both sides.
- The mutant probe: the one-line deletion from the union, `go build` exit 0, fixture A's stdout with the false warning back, plus the non-mutant stdout beside it.
- Whole-corpus before/after `CoverageIncomplete` counts with the command and the build's tree SHA on both sides, and the per-directory attribution if the after count is not `2018`.

## §E. Self-verification

- `go test ./internal/spec/...` — the affected package only. The full-suite verdict is CI's, per the repository's test-execution rule.
- `go vet ./internal/spec/...`
- `golangci-lint run` scoped to the changed files, with the pre-existing baseline measured in the same run-phase and quoted beside it.
- The measuring binary is rebuilt from the run-phase HEAD and cited with the tree SHA that produced it.

## §F. Risks

| Risk | Consequence | Mitigation |
|---|---|---|
| The unit is scoped to the line rather than the capture | Genuine `CoverageIncomplete` findings silenced — this card's own defect class | AC-SMS-003 over-reach positive control, which FAILS on a zero-finding result |
| The locator's separator admits `\n` | A `maps` section ending in a comma absorbs the next list item | `[ \t]` separator (REQ-SMS-005); AC-SMS-004 line-boundary control |
| A second expansion rule is written instead of reusing `cellREQIDs` | Two rules drift; the table and list paths diverge silently | REQ-SMS-002; **AC-SMS-010's positive call-site assertion** (a hand-rolled duplicate fails it by construction); AC-SMS-005's two greps as regression tripwires only — measured to miss five of six duplication spellings |
| The widened locator changes the id set on the 1099 existing full-id `maps` sections | A silent corpus-wide coverage shift | AC-SMS-009's measured before/after; the full-id no-regression unit test in M1 |
| `ears.go` edited for convenience | Reverses t561's decision and reaches the inline path | AC-SMS-006 pinned diff |
| The mutant probe passes vacuously (compile failure read as RED) | The probe asserts nothing | AC-SMS-008 requires `go build` exit 0 and both stdouts quoted |

## §G. Anti-patterns

- Reusing `reqSectionPattern`'s capture as the unit. Its truncation is the defect.
- Modifying `ExtractRequirementMappings` — reverses t561's decision and reaches the inline path.
- Declaring this card unnecessary because its corpus delta is zero. The zero is why it is cheap, not why it is skippable: fixture A proves the mechanism is live.
- Reporting the corpus after number without re-measuring the before against the same instrument.
- Reading a fixture run that printed nothing as a pass. A zero from an empty sweep asserts nothing; the `MissingExclusions` witness line is what makes the zero mean something.

## §H. Cross-references

- `internal/spec/lint_coverage_sibling.go` — `siblingAcceptanceCoveredREQIDs`, the union
- `internal/spec/lint_coverage_sibling_table.go` — `cellREQIDs`, `numericTailPattern`, `fullREQIDPattern`
- `internal/spec/ears.go` — `ExtractRequirementMappings`, `reqSectionPattern` (immutable)
- `internal/spec/lint.go` — `CoverageRule`, `collectAllREQIDs`
- Cards t518, t561, t524 — lineage; card t894 — the sibling heading-collection axis
