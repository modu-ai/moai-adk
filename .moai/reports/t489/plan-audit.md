# SPEC Review Report: SPEC-ERRCHECK-CATALOG-HASH-001

Card: t489 / lane-15 · Iterations: 1/3 FAIL (0.75) → 2/3 PASS (0.94) → 3/3 **PASS** (1.00) — final verdict; iteration-1/2 records retained verbatim beneath the final section.

- **Iteration 3 (FINAL)**: Verdict **PASS** · Score 1.00 (Clarity 1.0 / Completeness 1.0 / Testability 1.0 / Traceability 1.0) · Tier M (0.80) cleared · zero findings · skip-eligible (PASS + score ≥ threshold + no plan-artifact edits after this verdict).
- **Iteration 2**: Verdict PASS · Score 0.94 · D1-D5 resolved; two new minor findings (N1, N2) — both RESOLVED in iteration 3.
- **Iteration 1 (historical)**: Verdict FAIL · Score 0.75 · MP-3 FAIL (`lifecycle: spec-first`) + 4 further findings (D2-D5).

Reasoning context from the SPEC author was not provided; M1 Context Isolation holds. All premises below were re-verified against this worktree's actual files in this run (catalog_tree_hash.go, .golangci.yml, internal/spec/lint.go, corpus greps) — no premise was taken on faith.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**: REQ-001..REQ-006 sequential, no gaps, no duplicates, consistent zero-padding (spec.md:L42, L47, L54, L59, L66, L73).
- **[PASS] MP-2 GEARS format compliance** (judged against the requirement layer, spec.md §2 only): REQ-001 Ubiquitous ("The `internal/template` package shall produce...", L44-45); REQ-002 Event-driven ("**When** the catalog tree hash writes... shall explicitly discard...", L49-52); REQ-003 Unwanted ("shall not propagate, wrap, or branch", L56-57); REQ-004 Event-driven ("**When** the fix commit lands... shall have been committed", L61-64); REQ-005 Ubiquitous ("shall be... no new test shall be added", L68-69); REQ-006 Ubiquitous ("shall land... shall carry", L75-76). All six match canonical GEARS patterns; no informal-language entries. The Given-When-Then entries live in acceptance.md as AC-*, which is the correct verification layer — not penalized here.
- **[FAIL] MP-3 YAML frontmatter validity**: `lifecycle: spec-first` (spec.md:L12) is outside the canonical enum `spec-anchored|spec-lite|exploratory` (SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields). The Go lint cannot catch it — `FrontmatterSchemaRule.Check` (internal/spec/lint.go:989-1008) validates field **presence** only, never enum values — so this passes mechanically and silently; the doc-level schema is the binding check. Corpus context: 671 SPECs carry `spec-anchored`, only 4 carry `spec-first` — drift, not convention. The other 11 fields are valid: `id` SPEC-ERRCHECK-CATALOG-HASH-001 matches the enforced multi-segment pattern `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (lint.go:952), `phase: "v3.2.0 target"` is a release label (not a prohibited stage name), dates are ISO, priority P2 is in-enum, tags is a comma-separated string.
- **[N/A] MP-4 Section 22 language neutrality**: single-language (Go) defect-repair SPEC — auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation**: grep of `SPEC-([A-Z][A-Z0-9]+-)+[0-9]+` across spec.md/plan.md/acceptance.md returns only the SPEC's own ID (4 occurrences) — zero external SPEC references, nothing to reconcile, no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline**: zero `syscall` occurrences in any SPEC artifact → auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate**: `grep -rn '\[NEEDS CLARIFICATION'` on plan.md → 0 matches; research.md does not exist (plan.md exists and was checked).

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.0 | 1.0 band | Every REQ has a single interpretation. REQ-002 pins the exact expression — verified byte-identical to `internal/template/catalog_tree_hash.go:60` (`fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)`), with `h := sha256.New()` confirmed at :58. No pronoun ambiguity; discard rationale (hash.Hash never-error contract) stated at spec.md:L35-38. |
| Completeness | 0.50 | 0.50 band | Body is complete — HISTORY (L18), §1 Overview (WHY/WHAT), §2 Requirements, §3 AC summary, four `### Out of Scope — <topic>` H3s each with specific bullets (L88-110). Sole driver of the band: one frontmatter field carries an invalid enum value (the "missing/invalid one-to-two fields" grade). Fix is a one-word edit. |
| Testability | 1.0 | 1.0 band | All five ACs binary. acceptance.md:L9-11 explicitly judges on **match count, never exit code**; AC-MUTANT is an anti-vacuous guard with a closed revert cycle (count 1 → re-apply → re-confirm GREEN, acceptance.md:L30-31 + §D.5); AC-ORDERING uses an ancestry predicate, not a message claim; §D.5 demands verbatim command output. No weasel words found in any AC. |
| Traceability | 0.50 | 0.50 band | REQ-001→AC-GREEN/AC-MUTANT, REQ-004→AC-ORDERING, REQ-005→AC-SCOPE, REQ-006→AC-EVIDENCE all bind. REQ-002 and REQ-003 have **no discriminating AC**: a `//nolint:errcheck` mutant at line 60 passes all five ACs while violating both (see D2) — "multiple REQs lack ACs" per the anchor. |

## Defects Found

D1. MP3-LIFECYCLE-ENUM — spec.md:L12 — `lifecycle: spec-first` is outside the canonical enum (`spec-anchored|spec-lite|exploratory`, spec-frontmatter-schema.md § Canonical 12 Required Fields); the mechanical lint checks presence only (internal/spec/lint.go:996-1007) so it passes silently — Severity: critical — Class: blocking — Required fix: change line 12 to `lifecycle: spec-anchored`.

D2. AC-NOLINT-MUTANT-PASSES — acceptance.md:L42-49 (AC-SCOPE) + spec.md:L47-52, L54-57 (REQ-002/REQ-003) — No AC discriminates the sanctioned `_ =` discard from a prohibited alternative: a `//nolint:errcheck` comment at catalog_tree_hash.go:60 passes AC-GREEN (count 0), AC-MUTANT (revert → count 1), AC-ORDERING, AC-SCOPE (whose predicates test only "no new test file and no `.golangci.yml` change"), and AC-EVIDENCE — while violating REQ-002 (explicit discard + contract comment) and REQ-003, both of which the SPEC's own Out of Scope bans (spec.md:L101). By the repo's own mutant-probe doctrine (.claude/rules/moai/development/verification-completeness.md §2), a writable mutant means the criterion set is too shallow for these two REQs — Severity: major — Class: blocking — Required fix: extend AC-SCOPE's Then-clause (or add one AC-FORM): the branch diff at catalog_tree_hash.go must show the bare `fmt.Fprintf(...)` call prefixed with `_ =` plus a comment citing the hash.Hash never-error contract, and the diff must add no `//nolint` directive anywhere.

D3. TIER-UNDECLARED-MISSTATED — progress.md:L5 ("Tier S compact scope") vs the actual artifact set (spec.md + plan.md + acceptance.md = the Tier M set; Tier S is 2 files with AC inline in spec.md §3, per spec-workflow.md § SPEC Complexity Tier), with no `tier:` field in frontmatter (absence → Tier L treatment: threshold 0.85, retry ceiling 3) — Severity: minor — Class: blocking (a stated classification its own shape contradicts) — Required fix: add `tier: M` to spec.md frontmatter and correct progress.md §E.1 to "Tier M"; alternatively merge the AC matrix into spec.md §3, drop acceptance.md, and declare `tier: S`. The former is cheaper and matches the existing artifact set.

D4. AC-MUTANT-BLOB-RESTORE-SYNTAX — acceptance.md:L26-27 — the parenthetical alternative "or the recorded pre-fix SHA" offers the blob SHA (`735b7d7b...`) as a `git checkout <ref> -- <path>` source; `git checkout` takes a tree-ish, so the blob form fails at run time and sends the implementer into unplanned recovery — Severity: minor — Class: optional — Required fix: restrict the revert mechanism to the tree-ish form (`git checkout 615d18c1f -- internal/template/catalog_tree_hash.go`), or state the blob alternative as `git show <blob> > internal/template/catalog_tree_hash.go`.

D5. AC-GREEN-COUNT-SCOPE-NARROWER-THAN-REQ — acceptance.md:L20-21 vs spec.md:L44-45 — AC-GREEN's count predicate is file-scoped ("lines matching `catalog_tree_hash.go.*errcheck`") while REQ-001 is package-scoped; under the measured baseline they are extensionally equal, but the criterion would stay green if a future errcheck finding appeared in another package file — the verification is narrower than the requirement it certifies (verification-completeness.md §3) — Severity: minor — Class: optional — Required fix: bind AC-GREEN additionally (or instead) to the package-wide errcheck summary line `* errcheck: 0`.

## Verified Positives (premises checked, none invented)

- REQ-002's prescribed expression is byte-identical to the actual flagged line; `sha256.New()` is at :58 as claimed (catalog_tree_hash.go:L58, L60).
- `.golangci.yml` citations are accurate: `check-blank: false` at :35 within the :31-35 errcheck settings block; version-skew rationale at :1-16 exactly as the SPEC and red-baseline.md describe.
- Baseline-first ordering (verification-claim-integrity §2.3) is correctly the milestone's leading step (plan.md §F M1 step 1) and is witnessed by the commit graph via an ancestry predicate, not a message claim (AC-ORDERING).
- The linter-only verification scope with match-count-not-exit-code judging is stated in three places and is internally consistent (acceptance.md §A, plan §D, REQ-005).
- No cross-SPEC references, no `syscall`, no clarification markers, no REQ numbering defects.

## Recommendation

FAIL — route fixes directly from the defect list (this is iteration 1 of a Tier-L-ceiling 3; confirming re-audit may be scoped to the D1-D5 delta). Minimal path to PASS:

1. spec.md:L12 — `lifecycle: spec-first` → `lifecycle: spec-anchored` (D1).
2. acceptance.md — extend AC-SCOPE (or add AC-FORM) with the diff-shape predicates: `_ =` prefix + contract comment present at the flagged line, zero `//nolint` additions (D2).
3. spec.md frontmatter — add `tier: M`; progress.md:L5 — "Tier S" → "Tier M" (D3).
4. Optional (D4, D5) — fix the blob-restore wording and widen AC-GREEN's count predicate to the package-wide `* errcheck:` summary.

No commit was made by this audit; the lane session owns all commits with explicit pathspecs.

---
---

# Iteration 2 (2026-09-06) — Re-audit: PASS

Scope: delta re-audit per the Retry Loop Contract — the five iteration-1 defects (D1-D5) verified against the current document text (not the lane's claims), plus a regression sweep of the touched regions and the whole SPEC for repair-round plantings. Current verdict: **PASS**, score 0.94, Tier M threshold (0.80) cleared with `tier: M` now declared in frontmatter.

## Prior-Defect Resolution Matrix

| Defect | Status | Evidence (current text) |
|--------|--------|-------------------------|
| D1 lifecycle enum | **RESOLVED** | spec.md:L12 `lifecycle: spec-anchored` — in-enum. All 12 required fields re-validated; optional `tier: M` (L13) is a legal optional field (enum S\|M\|L). |
| D2 nolint mutant passes ACs | **RESOLVED** | acceptance.md:L52-59 — AC-SCOPE gained clause (c): (c)(i) the diff shows the bare call prefixed `_ =` plus the hash.Hash contract comment (explicitly "(satisfying REQ-002)", L56); (c)(ii) the diff adds NO `//nolint` directive (naming REQ-003 and the mutant at L57-59). The mutant is no longer writable: REQ-002 and REQ-003 now have a discriminating AC with explicit citations. |
| D3 tier undeclared/misstated | **RESOLVED** | spec.md:L13 `tier: M`; progress.md §E.1 rewritten to "Tier M compact scope (`tier: M` in frontmatter)" — matches the 3-artifact set + progress skeleton. Threshold becomes Tier M 0.80. |
| D4 blob-restore syntax | **RESOLVED** | acceptance.md:L28-31 — primary form `git checkout 615d18c1f -- internal/template/catalog_tree_hash.go` (tree-ish, valid); alternative `git show 735b7d7b4856088dc7702a56fd450c0800327666 > internal/template/catalog_tree_hash.go` (blob redirect, valid). |
| D5 AC-GREEN count scope | **RESOLVED** | acceptance.md:L20-23 — BOTH scopes bound: file-scoped count 0 AND package summary `* errcheck: 0`; spec.md:L83 summary row updated to match. Now binds package-wide REQ-001 exactly. |

## Regression Check (Iteration 2)

Iteration-1 defects: D1-D5 all RESOLVED (table above) — none unresolved, so the automatic-FAIL rule does not fire.

Whole-SPEC sweep for repair-round plantings:
- Frontmatter: 12 required fields all valid; `tier: M` legal and correctly placed; `updated:` unchanged (same-day edit — correct).
- MP-1/MP-2: REQ-001..006 unchanged, sequential, all GEARS — untouched by the repair.
- MP-5/MP-6/MP-7 re-scanned: still zero external SPEC-ID references, zero `syscall`, zero `[NEEDS CLARIFICATION]`.
- AC structure: §D.1 severity table (5 must-pass ACs), §D.5 gates, and Definition of Done all still coherent with the 5-AC matrix (the repair extended AC-SCOPE rather than adding a 6th AC — the cheaper option I recommended).
- Weasel-word scan on new text: clean; the match-count-not-exit-code discipline (§A) and the AC-MUTANT revert-cycle closure are intact.
- Score regression: iter2 0.94 > iter1 0.75 — no STOP signal.

## New Findings (repair round)

N1. ACSCOPE-NOLINT-PREDICATE-SELF-COLLISION — acceptance.md:L57 (+ pre-existing literals at spec.md:L102, plan.md:L68) — AC-SCOPE(c)(ii) forbids "the diff adds NO `//nolint` directive anywhere on the branch", but the branch diff WILL contain the literal string `//nolint` in the SPEC's own prose (3 files, 4 occurrences) once the plan artifacts and evidence are committed on the branch — which AC-ORDERING and AC-EVIDENCE themselves require. A grep-based closure check for the predicate hits the document carrier: the check as worded either false-positives on a correct implementation or invites an ad-hoc special-case. — Severity: minor — Class: blocking — Required fix: scope the predicate mechanically to Go source, e.g. "the diff of `*.go` files adds no `//nolint` directive (`git diff <base>..HEAD -- '*.go'` contains no `//nolint` match)".

N2. ACSCOPE-REVISION-SWEEP-INCOMPLETE — spec.md:L86, plan.md:L43 — the AC-SCOPE revision added clause (c) but its citing surfaces still carry the old scope: the spec.md §3 summary row reads "verification command set = linter only; no new test file added" (clauses (a)+(b) only) and plan.md §E E4 reads "scope check — no new test files, no `.golangci.yml` diff" — neither mentions the diff-shape/nolint predicate that is now the clause closing REQ-002/REQ-003. A implementer following the E4 checklist literally would skip the newest, load-bearing check at closure. Cross-layer revision sweep (verification-completeness.md §3, [HARD]): a revision does not end in the file it started in. — Severity: minor — Class: blocking — Required fix: one clause each — spec.md:L86 row append "; diff shape = sanctioned `_ =` discard + comment, no `//nolint` (clause c)"; plan.md:L43 E4 append ", diff shape + no `//nolint` in `*.go` (AC-SCOPE c)".

Both findings share one root cause: the AC-SCOPE revision was neither mechanically scoped nor swept to its citing surfaces. Each fix is a one-line edit.

## Verdict Rationale (Iteration 2)

- MP-1 PASS (REQ-001..006 sequential, spec.md:L43-L77); MP-2 PASS (all six REQs GEARS — requirement layer, spec.md §2; ACs correctly Given-When-Then in acceptance.md); MP-3 PASS (frontmatter re-validated field-by-field: 12 required fields valid, `lifecycle: spec-anchored` in-enum, `tier: M` legal optional); MP-4 N/A (single-language Go SPEC); MP-5 PASS (no external SPEC refs); MP-6 PASS (no `syscall`); MP-7 PASS (no clarification markers; research.md still absent).
- Traceability now complete: REQ-001→AC-GREEN(+MUTANT), REQ-002→AC-SCOPE(c)(i), REQ-003→AC-SCOPE(c)(ii), REQ-004→AC-ORDERING, REQ-005→AC-SCOPE(a)(b), REQ-006→AC-EVIDENCE — every REQ has a discriminating AC.
- Score 0.94 ≥ Tier M 0.80. The M5 firewall is fully green; N1/N2 are minor findings routed to the orchestrator and do not overturn the rubric-anchored verdict.

## Routing Note for the Orchestrator

Two paths, both within contract:
1. **Apply N1+N2** (two one-line edits) before kickoff — note this changes the plan-artifact hash, so the Phase 1 Plan Audit Gate will re-execute as iteration 3/3; the delta is trivially confirmable (grep the two edited lines).
2. **Proceed without edits** — the executor-side resolution of N1 (`-- '*.go'` scoping) is the only reasonable reading of "directive", and N2's SSOT (acceptance.md) already governs closure via §D.5/DoD. Document the decision in the run delegation prompt.

No commit was made by this audit; the lane session owns all commits with explicit pathspecs.

---
---

# Iteration 3 (2026-09-06) — Delta Confirmation: PASS (FINAL)

Scope: the trivially-confirmable delta from the Routing Note of iteration 2 — verify N1 and N2 in the documents, regression-scan the touched surfaces. This is the confirming iteration (3/3, ceiling reached). Current verdict: **PASS**, score 1.00, zero findings.

## Iteration-2 Defect Resolution

| Defect | Status | Evidence (current text) |
|--------|--------|-------------------------|
| N1 nolint-predicate self-collision | **RESOLVED** | acceptance.md:L57-63 — AC-SCOPE(c)(ii) now specifies the mechanical closure check `git diff 615d18c1f..HEAD -- '*.go'` and documents the scoping rationale inline ("the SPEC prose itself ... carries the literal string `//nolint` and a whole-branch grep would false-positive on a correct implementation"). The predicate is now precisely executable; the collision with the document carrier is closed. The mutant-hole explanation (L60-63) is retained. |
| N2 revision sweep incomplete | **RESOLVED** | spec.md:L86 — the §3 AC-SCOPE summary row now reads "...diff shape = `_ =` discard (no `//nolint` in `git diff -- '*.go'`)"; plan.md:L43-45 — §E E4 now reads "scope check — no new test files, no `.golangci.yml` diff, and the diff shape matches the sanctioned `_ =` discard form with no `//nolint` directive in `git diff 615d18c1f..HEAD -- '*.go'` (AC-SCOPE)". All three citing surfaces now carry the clause, consistent with the AC SSOT. |

## Regression Check (Iteration 3)

- Frontmatter: unchanged and valid (12 required fields + `tier: M`, `lifecycle: spec-anchored`).
- MP-5/6/7 file-wide re-scan: zero external SPEC-ID references, zero `syscall`, zero `[NEEDS CLARIFICATION]`.
- acceptance.md tail re-read in full this iteration: §D.1 (five must-pass ACs), §D.5 closure gates, and Definition of Done unchanged and coherent with the 5-AC matrix.
- The additional `//nolint` prose mentions introduced by the fixes themselves (plan.md:L44, spec.md:L86) are of the same class as the pre-existing ones and cannot trip the now-Go-scoped predicate — self-consistent by construction.
- Cross-surface command consistency: acceptance.md:L57 and plan.md:L45 name the identical closure command (`git diff 615d18c1f..HEAD -- '*.go'`); spec.md:L86 carries the abbreviated summary form, correctly deferring to acceptance.md as the full matrix.
- Score: no regression (1.00 > 0.94). No new findings.

## Final Verdict Rationale

All must-pass criteria green (MP-1..MP-7); all cumulative defects across three iterations (D1-D5, N1-N2) verified resolved against document text with line citations; every requirement REQ-001..REQ-006 has a discriminating AC; the anti-vacuous mutant is closed; the verification instruments are machine-executable and self-collision-free. Score 1.00 ≥ Tier M 0.80.

Skip-eligibility record: verdict PASS + score 1.00 ≥ 0.80 + no plan-artifact modification after this verdict → the Phase 1 Plan Audit Gate is skip-eligible per spec-workflow.md § Plan Audit Gate skip policy, contingent on the artifact hash computed over {spec.md, plan.md, acceptance.md} remaining unchanged from this point.

No commit was made by this audit; the lane session owns all commits with explicit pathspecs.
