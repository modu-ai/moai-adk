# SPEC-AC-ANCHOR-SCOPE-001 — Acceptance Criteria

frontmatter-less sibling artifact. All scenarios assume the frozen denominator
(`.moai/reports/t747/probe/filelist.txt`, 860 files) and in-run derivation (REQ-ACAS-005/006).

## §A Scope of Verification

Two headline numbers (narrow-miss, empty-anchor) with before/after columns; one control number
(no-regression); one bound check (prose layer). Corpus-level ACs are verified by the committed
two-column probe (M1); unit-level ACs by table-driven parser tests.

## §D AC Matrix

| AC | Requirement | Severity | Verification surface |
|----|-------------|----------|----------------------|
| AC-747-001 | REQ-ACAS-001 | Must | Probe narrow-miss column + per-file disposition record |
| AC-747-002 | REQ-ACAS-002 | Must | Probe empty-anchor column + per-file disposition record |
| AC-747-003 | REQ-ACAS-003 | Must | Probe control column (byte-identical parse results) |
| AC-747-004 | REQ-ACAS-004 | Must | Newly-included declaration attribution audit |
| AC-747-005 | REQ-ACAS-005 | Must | Probe output shape (two-column, single-run) |
| AC-747-006 | REQ-ACAS-006 | Must | Frozen-before-image integrity + in-run re-derivation |
| AC-747-007 | REQ-ACAS-007 | Should | Sibling path untouched (no diff in `lint_coverage_sibling.go` behavior) |

### AC-747-001 — Narrow axis: 14 → 0 repaired-or-justified

**Given** the frozen 860-file corpus and the committed two-column probe
**When** the probe runs post-repair in a single in-run derivation
**Then** for every file listed in `probe/anchor-missed.txt`, either (a) the file's parse yields
an anchor whose region includes ≥1 of its AC-declaration lines, or (b) the run evidence carries
a written per-file disposition showing its declarations are prose-shaped, not AC — and the
unjustified residual count is 0.

### AC-747-002 — Loose axis: 9 → 0 repaired-or-justified

**Given** the frozen corpus and the committed probe
**When** the probe runs post-repair
**Then** for every file listed in `probe/empty-anchor.txt`, the anchored region includes ≥1
AC-declaration line (no file remains anchored at an empty section), or the run evidence carries
a written per-file disposition — unjustified residual count 0.

### AC-747-003 — No-regression control

**Given** the control set = declaration-bearing files NOT in `anchor-missed.txt` ∪
`empty-anchor.txt`, derived in-run
**When** pre-repair and post-repair parse results are compared in the same run, same tree
**Then** every control file's in-section declaration set is byte-identical, or each delta
appears on an explicit justified-delta list in run evidence with a one-line reason per entry.

### AC-747-004 — Prose bound

**Given** the out-section prose layer (in=10/out=1 shape; SPEC-CLAUDEMD-DIET-V2-001,
SPEC-DB-SYNC-HARDEN-001 among the specimens)
**When** the post-repair in-section declaration set is diffed against the pre-repair set
**Then** every newly in-section declaration line lies inside a region covered by a defect-file
repair (AC-747-001/002 dispositions) or an AC-747-003 justified-delta entry — no new inclusion
originates from the prose layer.

### AC-747-005 — Probe output shape

**Given** the committed probe
**When** it runs with `T747_PROBE_OUT=<abs> go test ./internal/spec/ -run TestT747AnchorScope -v -count=1`
**Then** its output reports narrow-miss and empty-anchor counts as before/after columns derived
from a single run, and the denominator matches the frozen filelist (860).

### AC-747-006 — Frozen baseline discipline

**Given** the plan-phase frozen artifacts under `.moai/reports/t747/probe/` and the t528 probe
`internal/spec/zz_t528_anchor_probe_test.go`
**When** run-phase completes
**Then** no frozen before-image file is modified (git diff empty on those paths), `declRe` is
byte-frozen, the t528 probe still passes (frozen-anchor PRESERVE), and every corpus figure in
run evidence is attributed to an in-run command output.

### AC-747-007 — Sibling path untouched

**Given** REQ-ACAS-007 and `lint_coverage_sibling.go`'s `ExtractRequirementMappings` acceptance.md path
**When** the repair lands
**Then** the sibling rule's behavior is unchanged (existing sibling tests pass without
modification) and no commit in the SPEC touches that file.

## §D.1 Edge Cases

- A file with BOTH defect shapes (no vocabulary heading AND an empty-first fallback candidate)
  — repair order must be well-defined; covered by unit fixtures.
- A vocabulary heading whose section holds only prose AC-id mentions (not colon-form) — must
  NOT become an anchor via the new criteria (prose bound, AC-747-004).
- `acNegativeSectionMarkers` headings ("Out of Scope" etc.) that mention acceptance criteria —
  must continue to never anchor.
- A file whose only declarations sit under a level-3 heading inside a level-2 non-vocabulary
  section — the region-selection level semantics are part of M2 design and get a fixture.

## §D.2 Quality Gates

- `go test ./internal/spec/...` green (changed-package scope; full-suite verdict by CI)
- `golangci-lint run` clean; gofmt clean
- Package coverage ≥85% on `internal/spec` for the new anchor logic
- Conventional commits; card id t747 + SPEC ID traceable in commit messages

## §D.3 Definition of Done

- AC-747-001 through AC-747-006 all Must-severity PASS with cited in-run evidence under
  `.moai/state/verify/` or `progress.md` §E.2
- No-regression control closed (byte-identical or justified delta list)
- t528 baseline PRESERVE verified (AC-747-006)
- Doc comments updated (M5) and lint suite green

## §D.4 Traceability

REQ-ACAS-001→AC-747-001 · REQ-ACAS-002→AC-747-002 · REQ-ACAS-003→AC-747-003 ·
REQ-ACAS-004→AC-747-004 · REQ-ACAS-005→AC-747-005 · REQ-ACAS-006→AC-747-006 ·
REQ-ACAS-007→AC-747-007
