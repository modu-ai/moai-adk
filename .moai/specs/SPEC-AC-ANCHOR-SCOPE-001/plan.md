# SPEC-AC-ANCHOR-SCOPE-001 — Implementation Plan

frontmatter-less sibling artifact (per schema SSOT § Artifact Statelessness).
SPEC: `.moai/specs/SPEC-AC-ANCHOR-SCOPE-001/spec.md` · card t747 · branch `WT-ac-anchor-scope`

## §A Context

- Defect surface: `findACSectionStart` (`internal/spec/parser.go:123-138`) and its vocabulary
  gate `isACSectionHeading` / `acSectionVocabulary` (`parser.go:69-116`).
- Two measured defect axes on the live corpus (evidence:
  `.moai/reports/t747/anchor-scope-measurement.md`): 14 narrow-miss files (declarations, no
  anchor) and 9 empty-anchor files (fallback anchors the first empty vocabulary section; all 9
  via the fallback).
- Guard to preserve: `lint_coverage_sibling.go:29-32` — the double scoping exists because
  spec.md is a mixed document; prose must not be read as AC.
- Methodology precedent: t528's committed two-column probe
  (`internal/spec/zz_t528_anchor_probe_test.go`), frozen denominator, in-run re-derivation.
- Development mode: per `.moai/config/sections/quality.yaml` (DDD/TDD selected at run-phase
  entry); the probe promotion in M1 provides the RED before-column for either cycle type.

## §B Known Issues

| # | Issue | Evidence |
|---|-------|----------|
| B1 | 14 files: vocabulary + level≥2 gate misses their AC sections entirely | `probe/anchor-missed.txt` |
| B2 | 9 files: multiple vocabulary headings, first empty, fallback anchors it; declarations under later headings | `probe/empty-anchor.txt` |
| B3 | 165 out-section declarations include a correctly-excluded prose layer (in=10/out=1 shape) — any anchor widening risks absorbing it | measurement doc §「나머지 out-section 165의 성격」 |
| B4 | Probe-level scan (break at any `##`) vs `extractACLines` (anchor-level break) divergence — recorded, not resolved here | measurement doc §범위 메모 |

## §C Pre-flight and Design Space

Pre-flight (run-phase entry):

- [ ] Re-run the probe on the pre-change tree in-run (`T747_PROBE_OUT=…`) and confirm the
      before-column reproduces 14 / 9 / 1240 against the frozen denominator — never reuse the
      plan-phase numbers as the run baseline (t528 discipline).
- [ ] Confirm `internal/spec/zz_t528_anchor_probe_test.go` still passes (t528 anchor behavior
      is a PRESERVE surface: the repair must not regress the 216 frozen-anchor acceptances of
      t528's baseline).
- [ ] Confirm no file under `.moai/reports/t528/**` is touched.
- [ ] Freeze the pre-repair anchor behavior as a baseline column re-derived in-run
      (zz_t528 pattern), per file, over the control set (129 decl-bearing files outside the
      defect lists) — this frozen per-file baseline is the comparison instrument for
      AC-747-003.

Design space weighed (decision record — the run phase implements, this plan commits direction):

**(i) Narrow axis — anchor criteria beyond the fixed vocabulary.** Options:
  (a) Declaration-presence-based region selection: a heading (level ≥2) whose immediate section
      holds ≥1 discriminator-B declaration names an AC region, with the vocabulary heading as
      the common specific case. Risk: a prose paragraph that mentions an AC id inside an
      arbitrary section could anchor prose — bound it by requiring the AC-shaped list form
      (bullet + `AC-…:` colon), i.e. reuse the existing line-level scoping as the region
      qualifier, so the section-naming gate is widened, not removed.
  (b) Widen `acSectionVocabulary` only. Rejected as the primary fix: it re-creates the same
      miss class for the next un-listed heading (the 14 files exist precisely because the
      vocabulary is an open set), though adding measured phrases is allowed as a complement.
  Direction: (a) primary, with the mixed-document prose bound of REQ-ACAS-004 as the acceptance
  gate. Negative markers (`acNegativeSectionMarkers`, `parser.go:83`) continue to veto.

**(ii) Loose axis — fallback ordering.** Options:
  (a) Prefer a later non-empty region over an empty first: when all vocabulary sections are
      empty, fall back to the first declaration-bearing region (or the first NON-empty
      vocabulary section when one exists — the loop already tries this; the defect is only the
      terminal `return first`). Minimal change: make the terminal fallback declaration-aware
      instead of first-vocabulary-aware.
  (b) Return -1 (no anchor) when all vocabulary sections are empty. Rejected: it degrades the 9
      defect files from wrong-anchor to no-anchor without recovering their declarations.
  Direction: (a).

**(iii) Sibling path.** `lint_coverage_sibling.go`'s `ExtractRequirementMappings` full-text path
is out of scope (see spec.md §D). No change, no test modification there.

## §D Constraints

- No file under `internal/` other than the parser + tests changes without a blocker report.
- `declRe` (discriminator B) stays byte-frozen — it is the corpus-comparability contract with
  t528 and the t747 before-column.
- Frozen before-images are never overwritten; the committed probe writes only its own output.
- Corpus numbers are derived in a single run; no cross-run/cross-tree figure reuse
  (verification-claim-integrity §2 baseline attribution).

## §E Self-Verification

Run-phase evidence lands in `progress.md` §E.2/§E.3 (manager-develop). Plan-phase commitment:
every headline AC carries a command and an expected before/after pair (see acceptance.md §D).

## §F Milestones

Ordered by decision-reversibility: the two design decisions lead (M2, M3); the measurement
harness (M1) precedes them only because both consume its in-run before-column; mechanical steps
(docs) close.

| ID | Priority | Milestone | Deliverable |
|----|----------|-----------|-------------|
| M1 | High | Probe promotion + RED baseline | Probe committed in-tree as a two-column corpus test (t528 pattern, denominator frozen via `filelist.txt` artifact); in-run before-column recorded (14 / 9 / 1240 / 165); no-regression control set derived in-run and frozen as per-file baseline data — the comparison instrument for AC-747-003 |
| M2 | High | Narrow-axis anchor criteria | Declaration-presence-based region selection with vocabulary headings as the common case; negative markers preserved; unit fixtures from the narrow-miss shapes (SPEC-AC-COLLECTOR-ANCHOR-001, SPEC-CC297-001, SPEC-STATUS-AUTO-001) |
| M3 | High | Loose-axis fallback ordering | Declaration-aware terminal fallback (design §C(ii)(a)); unit fixtures from `probe/empty-anchor.txt` shapes; all 9 files anchor a declaration-bearing region |
| M4 | Medium | No-regression + prose-bound verification | Control set byte-identical (or justified-delta list); prose bound AC-747-004 verified; t528 frozen-anchor PRESERVE check green |
| M5 | Low | Docs + lint close | `acSectionVocabulary`/`findACSectionStart` doc comments updated to describe the new anchor contract; `go test ./internal/spec/`, `golangci-lint run`, gofmt green |

## §G Anti-Patterns

- Do NOT absorb the out-section prose layer to make the headline numbers move (the in=10/out=1
  shape is the guard working, not a miss).
- Do NOT edit `declRe` or the frozen before-images to reconcile a mismatch — re-derive in-run
  and explain the delta instead.
- Do NOT remove the section-naming gate entirely ("parse AC anywhere") — that is the
  mixed-document failure the sibling guard documents.
- Do NOT hardcode the 23 file paths as the fix; the criteria must be general (the frozen lists
  are verification fixtures, not production inputs).

## §H Cross-References

- Evidence: `.moai/reports/t747/anchor-scope-measurement.md`, `.moai/reports/t747/probe/*`
- Baseline: `.moai/reports/t528/probe/merged-tree-remeasure.md`, `internal/spec/zz_t528_anchor_probe_test.go`
- Guard rationale: `internal/spec/lint_coverage_sibling.go:29-32`
- Sibling SPEC: SPEC-AC-COLLECTOR-ANCHOR-001 (itself a narrow-miss subject — its own inline AC
  section becomes parseable as a consequence of this repair)
