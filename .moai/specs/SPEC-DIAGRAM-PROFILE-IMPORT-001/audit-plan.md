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
