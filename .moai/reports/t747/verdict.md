# t747 — SPEC-AC-ANCHOR-SCOPE-001 Run-phase Verdict

Card: t747 · Branch: `WT-ac-anchor-scope` (absorb base 188ece2f9) · Date: 2026-09-14
Code-verification baseline: 2960af5d5 (evidence commit follows this verdict)

## Claim

SPEC-AC-ANCHOR-SCOPE-001 run phase complete, M1-M5 landed, AC-747-001..007 all PASS
(7/7). The `findACSectionStart` anchor repair (declaration-named region selection +
declaration-aware terminal fallback) repairs 3 narrow-miss and 6 empty-anchor corpus
files with the 129-file no-regression control set byte-identical on both comparison
shapes, the prose layer untouched, and the t528 frozen-anchor PRESERVE (216) intact.
Residual 14 defect files carry written justified dispositions; unjustified residual 0.

### Headline clarification (sync-audit F3)

"Repairs 3 narrow-miss and 6 empty-anchor files" spans three mechanically distinct
classes, recorded here so no reader takes one for another:

1. **End-to-end criteria recovery (+4 criteria across 2 files)** — the only net
   parse change: SPEC-AC-COLLECTOR-ANCHOR-001 (+3, its `## 1. 배경과 문제` region
   declarations become parsed criteria) and SPEC-V3R6-I18N-VALIDATOR-BUDGET-001
   (+1 via the loose-axis fallback).
2. **Anchor-only selection, zero parseable lines** — SPEC-CC297-001 (19) and
   SPEC-STATUS-AUTO-001 (25) now anchor correctly, but their `AC-1.1` numeric-sub
   ids do not match `acIDPattern` (parser.go:407), so the parse still yields
   empty-with-discarded-error and lint output is identical. The anchor repair is
   real; the criteria recovery belongs to the line-grammar axis (t528 line, out
   of scope per §D).
3. **Strict-shape measurement artifacts** — 5 of the 6 empty-axis "repairs" have
   base==live anchor in `defect-disposition.txt` (the anchor never moved; only the
   dual-shape instrument stopped misreporting them). The strict-shape empty count
   is reported as after(strict-shape)=1240-line comparability only.

## Evidence

All commands run in this worktree; verbatim outputs persisted under
`.moai/state/verify/t747/` and quoted in `progress.md` §E.2.

- Corpus probe, after (HEAD 2960af5d5, single run):
  `T747_PROBE_OUT=…/probe-after2 go test ./internal/spec/ -run TestT747AnchorScope -v -count=1`
  ```
  DENOMINATOR spec.md read = 861
  NARROW axis (decls, no anchor): before=14 after=11 repaired=3
  EMPTY axis (anchored, 0 in-section decls): before=9 after=3 repaired=6
  declarations WHOLE FILE = 1405
  declarations IN-SECTION: before(frozen-strict)=1240 after(parse-region)=1353 after(strict-shape)=1240
  CONTROL set (decl-bearing, non-defect) = 129, strict-shape deltas = 0, parse-shape deltas = 0
  NEWLY-INCLUDED declarations = 48
  ```
  Before-column (pre-repair, commit d5ac26636 tree state) reproduced the frozen
  numbers exactly: before=14 / before=9 / before(frozen-strict)=1240, and both
  `FROZEN-CROSSCHECK` lines: `anchor-missed.txt: MATCH (14 files)`,
  `empty-anchor.txt: MATCH (9 files)`.
- Prose bound: all 48 newly-included declarations attributed to defect files
  (47 narrow-defect, 1 empty-defect); 0 in control files (`newly-included.txt`).
- RED evidence (pre-GREEN, verbatim): `.moai/state/verify/t747/red-m2-unit.txt` —
  7 failing subtests across both defect axes; e.g.
  `findACSectionStart = -1, want 7` (narrow, SPEC-CC297-001 shape) and
  `findACSectionStart = 3, want 5` (loose, empty-anchor shape).
- Full affected package: `go test ./internal/spec/ -count=1` →
  `ok github.com/modu-ai/moai-adk/internal/spec 99.574s`.
- `go vet ./internal/spec/` → exit 0. `golangci-lint run ./internal/spec/...` →
  `0 issues.` `gofmt -l internal/spec/` → clean.
- Coverage (per-function, anchor test set): `findACSectionStart 100.0%`,
  `sectionHoldsACDeclaration 100.0%`, `isNegativeSectionHeading 100.0%`,
  `isACSectionHeading 85.7%`.
- t528 PRESERVE: `T528_PROBE_OUT=… go test ./internal/spec/ -run TestT528Anchor -v -count=1`
  → `accepted by FROZEN baseline anchor = 216`.
- Untouched surfaces: `git diff --stat 188ece2f9..HEAD -- internal/spec/lint_coverage_sibling.go
  internal/spec/zz_t528_anchor_probe_test.go .moai/reports/t528/` → 0 lines;
  `git diff --name-only 188ece2f9..HEAD -- .moai/reports/t747/probe/` → 0 lines.
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` → OK.
- Commits (branch WT-ac-anchor-scope, no push): d5ac26636 (M1), ce4446694 (M2 RED),
  01bbb6360 (M2 GREEN), b82f6ecf9 (M3), 2960af5d5 (M4+M5), + evidence commit.

## Baseline-attribution

Every figure above is the verbatim output of a command run in this session, in this
worktree (`.claude/worktrees/t747`), against tree 2960af5d5 (probe runs) or the
commit named with it. The before-column was re-derived in-run from the frozen
baseline anchor copy — never carried from the plan-phase report. Denominator is 861
in-run vs the frozen artifact's 860: the delta is exactly this SPEC's own spec.md
(diff of `filelist.txt` vs frozen), the corpus self-modification the t528 discipline
anticipates. The plan-phase 860/14/9/1240 figures were used only as the expected
values the in-run baseline column had to reproduce — and did.

## Gaps

- The 11 narrow dispositions carry written justifications in two classes (sync-audit
  F2 relabel): (a) 7 genuinely prose-shaped mentions, (b) 4 colon-less AC bullets
  awaiting the line-grammar axis (CODERABBIT-ADOPTION, AGENT-MODEL-ROUTING,
  SKILL-COMPRESS, SKILL-CONSOLIDATE); and 3 empty residuals
  (GLM-EFFORT-MAX-001 metric artifact; OUTOFSCOPE-GUIDANCE-ALIGN-001 and
  _archive/SPEC-DESIGN-CONST-AMEND-001 colon-less declarations) are dispositioned,
  not repaired. Recovering the colon-less classes requires widening the LINE grammar
  (`acIDPattern` numeric-sub / separator set) — the t528 axis, whose frozen
  baselineRe-vs-live contract this SPEC must not disturb. Out of scope per §D.
- `TestT565AnchorVocabularyIsNotWide` was modified (not deleted): its fixture IS the
  narrow-axis defect shape REQ-ACAS-001 repairs. The surviving bound is pinned by a
  companion test (`…_ColonlessProse`). This is the one pre-existing test whose
  assertion changed; full package green with it.
- The strict-shape after-count for the empty axis is not reported as a headline
  (strict-scan cannot see `###`-nested declarations — the recorded B4 divergence);
  after-column headline uses the parse's region. `after(strict-shape)=1240` is
  logged for comparability.
- Not run: full-repo `go test ./...` (lane-local verification policy; CI on
  origin/develop owns the full-suite verdict), `-race` (no goroutines touched).

## Residual-risk

- The region qualifier accepts loose AC ids with a required colon separator; a
  non-vocabulary section holding `- AC-<anything>: …` bullets in prose intent would
  now anchor. Corpus-measured cost: 0 control deltas across 129 files; the prose
  layer's out-section lines (id mentions without id-adjacent colon) do not qualify.
- `pending-backfill-run`: the §E.3 `run_commit_sha` cannot cite its own commit; the
  real SHA lands with the evidence commit (D3 backfill pattern).
- The t565 contract update means any downstream consumer asserting "vocabulary-only
  anchoring" textually may read stale; the behavior contract, not the test name, is
  the SSOT.
- Lane does not push; the verdict is local. CI on `origin/develop` (lead batch push)
  is the integration verdict surface.
