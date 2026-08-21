# Plan-Phase Audit — SPEC-DIAGRAM-PROFILE-IMPORT-001 (card t167)

Audit artifact (in-SPEC copy). Canonical review-stream report: `.moai/reports/plan-audit/SPEC-DIAGRAM-PROFILE-IMPORT-001-review-1.md`.
Auditor: plan-auditor (iteration 1 of Tier M ceiling 2). M1 Context Isolation active — only spec.md / plan.md / acceptance.md / progress.md read; every ground-truth claim re-measured against the worktree (HEAD `03d77e4c3`, branch `WT-diagram-profile-import`). No audit_multi (worktree-blind, per dispatch).

**Verdict: FAIL — Overall score 0.84 (Tier M threshold 0.80).** Two blocking findings (D1 evidence integrity, D2 internal consistency); both are small, precisely-scoped fixes. All seven must-pass criteria PASS. The architecture, decisions, and traceability are sound — this FAIL is a fix-and-return, not a redesign.

---

## 1. Must-pass results

| MP | Result | Evidence |
|----|--------|----------|
| MP-1 REQ consistency | **PASS** | REQ-1..REQ-15 sequential, no gaps/duplicates (spec.md §3 L118-221; token census `grep -o "REQ-[0-9]*" \| sort \| uniq -c` → exactly 15 distinct ids, counts 1-2 each = definition + cross-ref). 15 ≤ 16 Tier M REQ ceiling. |
| MP-2 GEARS compliance | **PASS** | Requirement layer only (REQ entries, spec.md §3). All 15 shall-based: Ubiquitous REQ-1/3/6/7/8/11/12/13/15; When REQ-4/5/9/10; Where REQ-2 (dual Where clauses, static-config gate), REQ-14 (Where+Ubiquitous compound — PASS-equivalent). No informal REQs; no Given/When/Then in the requirement layer. acceptance.md ACs are the verification layer (Given-When-Then by design) — graded under §4 below, not penalized here. |
| MP-3 Frontmatter | **PASS** | All 12 canonical fields present, correct types (spec.md L2-15): quoted semver `"0.1.0"`, `status: draft`, ISO dates, `priority: P1`, `phase: "v3.1.3 target"` (valid release-target label; not a prohibited lifecycle-stage whole value), path-like module, `lifecycle: spec-anchored`, comma-string tags. No snake_case aliases. Optional extras valid (`era: V3R6`, `tier: M`, `depends_on:` list). |
| MP-4 Language neutrality | **PASS (N/A enumeration)** | No language-specific tool names anywhere in the SPEC; the deliverable is language-neutral skill prose for the distributed template. REQ-15 + AC-TPL-003 bind the template-neutrality classes (C1-C8). |
| MP-5 D7 cross-SPEC | **PASS** | Only external reference: SPEC-SVG-QUALITY-ABSORB-001 (6 refs). `.moai/specs/SPEC-SVG-QUALITY-ABSORB-001/spec.md` exists; `status: completed` (frontmatter L5, verified). Not retired/superseded/archived → no reconciliation owed; no missing-SPEC refs; no BLOCKING finding. |
| MP-6 D8 cross-platform | **PASS** | `grep -c "syscall" spec.md` → 0 (auto-pass). |
| MP-7 Clarification gate | **PASS** | `grep -c "NEEDS CLARIFICATION" plan.md` → 0 (exit 1). research.md does not exist (Tier M artifact set = spec/plan/acceptance) — N/A for that file. |

Structure (Group 2): SC-1..SC-6 all PASS — HISTORY §6; Problem/WHY §1; Scope/WHAT §2; REQUIREMENTS §3 (15 entries); ACs in acceptance.md §D (Tier M placement); **four** `### Out of Scope — <topic>` H3 sub-headings each with concrete `-` bullets (extractor scripts / B-10 icons / t166 verifier / default-flow).

## 2. Ground-truth verification (spec §4 claims, re-measured)

| Claim | Result | Measured evidence |
|---|---|---|
| t166 absent: check-svg.mjs has no geometry checks | **TRUE** | `wc -l` → 609 ✓; `grep -c "paint-order\|attachment\|mask.*gap"` → 0 ✓; diagnostic codes SVG001-003/010-011/020-021/030-031/040/**050**/060-064 — spec's enumeration omits SVG050 (D3, minor) |
| authoring.md §2.5 carries the six connector rules with REQ-11's numbers | **TRUE** | `### 2.5 The six mandatory connector rules` at line 152 ✓; C1 `r = 8` + floor `r = 6` (L159-162), C2 gap 6-10 (L166-171), C3 `≥ 12` + `rHop = 5` (L180-184), C4 `L·k/(N+1)` + `>= 12` (L193-202), C5 dashed transit (L215), C6 mask-vs-later-node + paint order (L226-242) |
| catalog hashes only root SKILL.md | **TRUE** | `internal/template/catalog_hash_norm.go` L26-27 verbatim: "For skill directories, only the root SKILL.md or skill.md is hashed. Sub-files (workflows/*.md, references/*) are NOT included"; catalog.yaml carries both skills' hash entries (L101-110) |
| design-dna has no profile mechanism (group-a gap) | **TRUE in substance, FALSE as measured** | Mechanism markers `grep -ri "marker\|slug\|snapshot\|persist"` → SKILL.md:0, effects:0, dna-schema:1 — the persistence mechanism is genuinely absent. But the stated `grep -c "profile"` → "0 matches" is **false**: 6 lines match (SKILL.md:69,147,185; dna-schema.md:4,115; effects-implementation.md:102), all in the instance sense ("a DNA profile"), not persistence. → **D1** |
| svg-infographic has no importer (group-b gap) | **TRUE in substance, FALSE as measured** | `grep -ri "drawio"` whole skill dir → 0 ✓. But the stated `grep -ci "drawio\|import"` → "0 matches" is **false**: `-i` matches "important" (SKILL.md:353, authoring.md:352). → **D1** |
| Upstream citations | **TRUE** | survey.md L36 (profiles.md row: slug grammar, marker-first, schema check + backfill), L38-39 (import rows), L90 (ADR 0006 incl. rejected alternatives: in-install storage, token-merge, central path index), L98-99 (extract-first, untrusted, never carry coordinates/colors/fonts, fidelity ledger); L50-51 ground the extractor sizes (drawio_extract.py 31,364 B; mermaid_extract.py 47,650 B ≈ the "78 KB" total); absorption-verdict.md L9 classifies B-8→design-dna, B-9→opt-in importers, B-10→icon normalization "if needed, with THIRD_PARTY notice" (matches the deferral framing); UPSTREAM-LICENSE present |
| t165 deferral | **TRUE** | Sibling §2 `### Out of Scope — deferred to sibling cards` at L83; status `completed` |
| 12-glyph icon set (B-10 rationale) | **TRUE** | authoring.md L385 "Twelve single-path glyphs on a 24x24 grid" + L404 `stroke="currentColor"` |
| Sibling phase-label anomaly (spec §5 note) | **CONFIRMED** | t165 frontmatter `phase: "v3.2"` while its artifacts landed in origin/release/v3.1.3 — exactly as spec §5 reports for the orchestrator |

## 3. Dimension findings (dispatch-requested)

**(1) REQ/AC quality.** 15/15 REQs GEARS-conformant and atomic; every AC is mechanically testable without Go code (grep counts, diff exit codes, `make build` exit, file reads); weasel-word grep over acceptance.md → 0; 15 AC ≤ 16 ceiling; §D.2 traceability table verified complete in both directions (no orphan ACs, no uncovered REQs). Minor: AC-IMP-006(b) "present unchanged in substance" and AC-TPL-003's date classification require minor interpretation (scored in Testability).

**(2) Ground truth.** All load-bearing facts verified (table above) — including the two false evidence lines (D1) and the SVG050 omission (D3). The substantive gap claims survive: both mechanisms are genuinely absent from the tree.

**(3) Decision quality.**
- **`.design-dna/` project-root storage — SOUND, code-certified.** `ManagedCleanTargets` (`internal/cli/update/deploy/deploy.go:56-88`) cleans exactly seven `.claude/*` roots (settings.json, commands/moai, agents/moai, `skills/moai*` glob, rules/moai, output-styles/moai, hooks/moai) plus `.moai/config`; project-root `.design-dna/` matches none → survives `moai update`. plan §G's anti-pattern is code-accurate (the skill dir IS wiped by the glob). The deviation from upstream's home-global `~/.diagram-design/profiles/` store is conscious and recorded with rejected alternatives (plan §F M1).
- **Procedural-not-script — SOUND.** The 78 KB of upstream Python justifies ORDER and discipline only; REQ-7/REQ-8 restate survey L98-99's extract-first/untrusted/never-carry rules; no vendoring → REQ-13's no-THIRD_PARTY reasoning is internally consistent with B-10's conditional notice.
- **B-10 icon deferral — SOUND** (natively-present 24×24 currentColor set verified; "future card" path left open).
- **t166 contract-against-rules-not-code — SOUND.** Geometry checks verified absent on this branch; REQ-11 binds to authoring.md §2.5 numbers (all verified present), never to t166 code numbers; AC-VERIFY-001 gates run-phase alignment with an honest gap path.

**(4) t166 interface (non-overlap).** **SUPPORTED by structure.** svg-infographic SKILL.md carries `## Linting the source` (L229) and `## Bundled references` (L285) as distinct sections — t167's delta (table rows) and t166's expected delta (Linting section) do not collide at section level; the scripts table already lists `scripts/fixtures/` so t166 has no reason to touch the tables. Residual merge risk is honestly fenced (spec §5, plan §B1 re-read at M3).

**(5) Migration contract.** REQ-10 = strict same-change replacement, no coexistence path — the correct extension of the verified one-home rule (SKILL.md L65-67 already sanctions replace-and-delete-in-same-change). The routing tension is resolved **for the routing table** (REQ-12) but **not** for the categorical Step 0 sentence — see **D2**. Edge-case gap: REQ-10 assumes the caller owns/removes the source (D5).

**(6) Scope fence.** Clean: t166's `scripts/*` fenced in two places (spec §2 Out of Scope; plan §D NEVER-touch), icon vendoring fenced with rationale, description/when_to_use byte-stability required (AC-TPL-004), 16-language neutrality via REQ-15/AC-TPL-003, groups (a)/(b) separable with no shared file except the section-disjoint SKILL.md delta.

## 4. Defects

| ID | Location | Finding | Severity | Class | Fix |
|----|----------|---------|----------|-------|-----|
| D1 | spec.md L33-35 (§1a), L228-231 (§4) | Two §4 evidence measurements do not reproduce: `grep -c "profile"` → 6 lines, not 0; `grep -ci "drawio\|import"` → 2 lines ("important" at SKILL.md:353, authoring.md:352), not 0. plan §C pre-flight #3 will print the contradiction at run-phase. | major | **blocking** | Restate both bullets with commands that reproduce: `grep -rni "drawio" → 0` and mechanism markers (`marker\|slug\|snapshot\|persist`) → ~0 as the gap evidence; or corrected counts + a note that residual "profile"/"important" hits are the instance sense / an adjective, not the mechanism. |
| D2 | spec.md §2 L74-76, §5 L263-270; plan.md §B1, §D; target SKILL.md:43-44 | Preserved Step 0 sentence "Nothing here migrates, rewrites, or deprecates an existing mermaid diagram" contradicts the new opt-in migration references landing in the same file. REQ-12's compose-with clause binds the routing **table** only; nothing requires reconciling this categorical prose. Following plan §D's PRESERVE verbatim ships a self-contradicting SKILL.md. | major | **blocking** | Require one explicit reconciliation: extend REQ-12 or plan M2 decision 2 (witnessed by AC-IMP-006 or a new sub-check) so the opt-in table row / adjacent sentence states the importers are the explicit caller-invoked exception to Step 0's default no-migration stance — or scope the sentence ("Nothing here migrates ... by default"). SKILL.md L65-67 already contains the reconciliation seed (replace-and-delete), so this is a one-sentence requirement. |
| D3 | spec.md L233-234 | §4 diagnostic-code enumeration omits SVG050 (actual set: SVG001-003, 010-011, 020-021, 030-031, 040, **050**, 060-064). | minor | blocking (fold into D1's §4 correction) | Add SVG050 while correcting §4. |
| D4 | plan.md §A, §B5 | Line counts off-by-one (design-dna "193" vs `wc -l` 192; svg-infographic "386" vs 385) — trailing-newline counting artifact; §C pre-flight re-measures. | minor | optional | Correct at D1-edit time or leave (pre-flight self-corrects). |
| D5 | spec.md REQ-10; acceptance.md §D.5 | No edge case for a source the caller does not own / cannot remove (e.g., diagram imported from a vendor doc) — REQ-10's "never coexist" is categorical. | minor | optional | One line in §D.5: external source → state one-home cannot be satisfied; decline or keep the copy annotated as derived. |

## 5. Scores

Clarity 0.75 · Completeness 0.75 · Testability 0.85 · Traceability 1.00 → **Overall 0.84** (mean; Tier M threshold 0.80 — met, but verdict FAIL on blocking D1+D2 per the must-fix-first discipline).

## 6. Fix route for iteration 2 (delta re-audit scope)

1. D1+D3: rewrite the two §4 evidence bullets (and the §1 restatements) with reproducible commands; add SVG050.
2. D2: add the explicit Step-0 reconciliation requirement (REQ-12 or plan M2 + an AC witness).
3. Optional: D4 line counts, D5 edge-case line.

No other changes requested — do not touch the REQ architecture, the storage decision, or the t166 interface; they verified clean.

---

# Iteration 2 — Delta Re-audit (FINAL; Tier M ceiling 2/2)

Revision audited: commit `baef137ec` (SPEC v0.2.0). Scope per the iteration-1 fix route: D1+D2+D3 delta + regression sweep + D4/D5 spot-check. Every fix claim re-measured; nothing taken from the delegation message on trust.

**Verdict: PASS — Overall score 0.92** (Tier M threshold 0.80). All 7 must-pass criteria PASS. No blocking defect remains. Score trajectory 0.84 → 0.92 (no regression; no STOP signal).

## Regression check — iteration-1 defects

| ID | Status | Evidence |
|----|--------|----------|
| D1 (false evidence measurements) | **RESOLVED** | §1/§4 now carry three discriminating greps; all three re-run verbatim by the auditor on this tree → 0 matches each (exit 1): `grep -rniE "\.design-dna\|profile[ -]?(marker\|snapshot)\|active profile marker"` over design-dna/; `grep -rn "drawio"` and `grep -rn "import-mermaid\|import-drawio"` over svg-infographic/. The naive forms are documented as non-discriminating with exactly the line numbers the auditor measured (profile → 6 instance-sense lines; `-ci drawio\|import` → 2, both "important": SKILL.md:353, authoring.md:352). plan §C #3 carries the same greps + the do-not-use-naive-forms warning — the run-phase contradiction vector is closed. |
| D2 (Step-0 contradiction) | **RESOLVED** | New REQ-13 (spec §3 L223-234) mandates amending exactly the SKILL.md:43-44 sentence with the caller-invoked exception, seeded from the one-home passage at :65-67; §2 scope table + §5 "confined to one sentence" bullet state it; plan §D gains the explicit "ONE EXCEPTION to PRESERVE (REQ-13)" clause and §F M2 decision 2 carries the amendment instruction — spec and plan are mutually consistent (the feared spec-vs-plan contradiction is absent). New AC-IMP-007 verifies the diff shows exactly one amended sentence outside the bundle table; AC-IMP-006(b) is re-scoped to workflow/routing-table/dials so the inter-AC contradiction is gone. The reconciliation is requirement-witnessed at both spec and AC level. |
| D3 (SVG050 omitted) | **RESOLVED** | §4 L276 now lists "SVG040, SVG050 (parser failure)" — gloss verified against the checker (SVG050 emitted from the structural/parse-error tier; header comment: "error tier covers document structure (SVG001-SVG050)"). |
| D4 (line counts) | **RESOLVED** | plan §A/§B5 now read "192 lines, `wc -l`" / "385" — matches the auditor's measurements. |
| D5 (external-source edge) | **RESOLVED (beyond asked)** | REQ-10 gains three GEARS-formed boundaries (no auto-discovery; replacement requires caller intent; non-owned source → untrusted + ledger records a derivation, not a migration); §D.5 carries the matching edge case. |

## Delta verification summary

- REQ census: REQ-1..REQ-16 sequential, no gaps/duplicates; groups (a) 1-5, (b) 6-13, (c) 14-16. 16 ≤ 16 ceiling (at ceiling, not exceeding). New REQ-13 is shall-based GEARS (Ubiquitous with embedded exception scoping); renumbered REQ-14/15/16 are the old 13/14/15 verbatim.
- AC census: 16 distinct ACs (AC-IMP-007 added); §D.2 traceability fully renumbered — 16/16 both directions, no orphans. §D.1 must-pass list extended to AC-IMP-001..007.
- Frontmatter: v0.2.0, HISTORY 0.2.0 entry, all 12 canonical fields intact. MP-7 re-verified (plan.md NEEDS CLARIFICATION → 0); D8 syscall → 0.
- Diff scope 03d77e4c3..baef137ec: only the 5 SPEC-artifact files — zero template/Go files touched (plan-phase discipline held).
- Lint claim verified by the auditor: `moai spec lint .moai/specs/SPEC-DIAGRAM-PROFILE-IMPORT-001/spec.md` → "No findings — all SPEC documents are valid" (exit 0).
- progress.md §E.1 counts updated to 16 REQ / 16 AC.

## Residuals (non-blocking)

| ID | Location | Finding | Class | Disposition |
|----|----------|---------|-------|-------------|
| R1 | plan.md §A L19, §F M3 (Attribution/catalog/neutrality items), §H L200-203 | Stale cross-cutting REQ pointers after the renumber: five citations still use the old REQ-13/14/15 where the correct numbers are REQ-14 (attribution), REQ-15 (Template-First/catalog/mirrors), REQ-16 (neutrality). acceptance.md — the binding verification surface — is correctly renumbered, and the plan items are self-describing, so misimplementation risk is low. **Routing note:** manager-develop cannot edit plan.md body (Forbidden ownership crossings) — fix via a trivial manager-spec touch-up before kickoff, or accept as documented debt. | minor, SHOULD-FIX | Orchestrator discretion |
| R2 | plan.md §B1 | Wording still says the SKILL.md delta is "confined to the bundled-references table + at most one opt-in sentence near it" — superseded by §D's explicit ONE EXCEPTION clause; harmonize when touching R1. | NOTE | Fold into R1's touch-up |
| R3 | spec.md §5 non-overlap bullet | Parenthetical "t167's delta (bundled-references table)" no longer mentions the Step-0 sentence amendment (fully stated in §2 table + §5's own bullet). The non-overlap claim itself remains true — Step 0 is disjoint from t166's Linting section and scripts. | NOTE | No action required |

## Iteration-2 scores

Clarity 0.90 · Completeness 0.90 · Testability 0.90 · Traceability 1.00 → **Overall 0.92** (mean). Deductions are the R1-R3 residuals; no dimension has a blocking defect.

**Verdict: PASS.** The batch kickoff applies; run-phase M1 may proceed on this SPEC. The R1 touch-up (five stale REQ numbers in plan.md) is recommended before M1 so the implementer's milestone checklist reads the correct requirements, but it does not gate the kickoff.
