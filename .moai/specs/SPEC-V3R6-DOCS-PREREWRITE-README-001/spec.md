---
id: SPEC-V3R6-DOCS-PREREWRITE-README-001
title: "Docs Pre-Rewrite README — stale citations, Q8 badge rollback to v3.1.2, and the CG retirement paragraphs in the four README files (card t1720)"
version: "0.1.0"
status: draft
created: 2026-10-26
updated: 2026-10-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "README.md, README.ko.md, README.ja.md, README.zh.md, docs-site/hugo.toml, docs-site/content/ko/core-concepts/what-is-moai-adk.md"
lifecycle: spec-anchored
tags: "docs, readme, i18n, badge, cg-retirement, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, SPEC-V3R6-DOCS-PREREWRITE-GATE-001, SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001, SPEC-V3R6-DOCS-PREREWRITE-FLAG-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-README-001 — README child of the docs pre-rewrite umbrella

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-26 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1720 (Class C, docs-only), child README of SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (umbrella split, operator decisions `d-20261014T055241Z-1bc6` and `d-20261014T060405Z-0fcd`). Ownership imported from the pre-revision umbrella at commit `bdb7b764` by path and SHA (umbrella spec.md child table; acceptance.md cells E-01 to E-04, E-08, E-08b, E-20a, E-20b, E-20c; decision-index.md rows Q1, Q8, Q13, Q17). Every RED-now figure was re-measured on THIS tree, commit `cb30e05b3e` (full `cb30e05b3e5f64774b8dabd87e80c24332f4551d`, tree object `9c08830bdec23b81334840307ee745e468500174`); no umbrella figure is carried over. Q8 is resolved for this card by the leader lens (dispatch to lane-14, 2026-10-26) plus fresh measurement; it is recorded as a leader-lens resolution, not an operator verdict. Status `draft`. |

## 1. Overview (WHY)

This child owns the README part of the umbrella's stale-citation requirement (REQ-PRC-001), the badge requirement (REQ-PRC-003), and the CG retirement passage as it appears in README.ko.md and the one Korean page missing its link (REQ-PRC-004). Measured state on `cb30e05b3e` (full evidence in acceptance.md §C):

- The umbrella's E-01 family (CLI forms of the absent verbs `cg`, `gpt`, `audit`, `codemaps`, `model`), scoped to the four README files, matches **4 lines in each of the 4 README files** (16 lines; R-01, exit 0). The E-02, E-03, and E-04 families, scoped the same way, match **nothing** (R-02, R-03, R-04: no output, exit 1, wrapper-confirmed). The README part of AC-PRC-001 is therefore red through one family only.
- All four README Release badges (line 24 of each file) name **v3.2.0** (R-20a), and `docs-site/hugo.toml` lines 55–56 carry `version = "v3.2.0"` and `releaseDate = "2026-10-23"` (R-20b). The public GitHub release list (external, observed 2026-10-26T10:31:30Z, R-20c) marks **v3.1.2 as Latest** (published 2026-08-21T06:50:41Z), holds drafts v3.1.3 and v3.2.1, and carries **no v3.2.0 release object** — the "3.2 rollback" state. Local tags list v3.2.0, v3.1.2, v3.1.1, v3.1.0 and no v3.1.3 (R-20d).
- The CG retirement sentence `설정 이전 안내와 함께 종료합니다` appears on 9 Korean docs-site pages (R-08, one copy each) and **twice in README.ko.md** (R-08r). Seven of the eight non-canonical ko pages already link to the canonical page `docs-site/content/ko/multi-llm/cg-mode.md`; `core-concepts/what-is-moai-adk.md` is the one without a link (R-08b). The en, ja, and zh counterparts of what-is-moai-adk.md also carry no cg-mode link (R-08b2, exit 1).

## 2. Ownership and edit surface

From the umbrella child table (umbrella spec.md at `bdb7b764`): this child owns REQ-PRC-001 (README part — the four README files ONLY, not the docs-site pages), REQ-PRC-003, REQ-PRC-004; AC-PRC-001 (README partial criterion), AC-PRC-003, AC-PRC-004. Decision rows owned: Q1, Q8, Q13, Q17.

Edit surface (six files plus this SPEC directory):

| File | Edit |
|------|------|
| `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` | line 24 badge (Q8); the four E-01-flagged lines of each file (Q13/Q17): line 58, the GPT-gateway callout (ko 291 / en 296 / ja 290 / zh 289), and the two CG retirement paragraphs (ko 386+591 / en 387+592 / ja 385+590 / zh 384+589) (Q1/Q17) |
| `docs-site/hugo.toml` | lines 55–56: `version = "v3.1.2"`, `releaseDate = "2026-08-21"`, in the SAME commit as the badge change (REQ-PRC-003) |
| `docs-site/content/ko/core-concepts/what-is-moai-adk.md` | the missing cg-mode link, per Q1 (disposition of its own sentence copy is child decision QR-1) |

The other seven Korean pages that already link are NOT this child's surface (leader-lens scope, dispatch to lane-14, 2026-10-26); their sentence copies remain after this child lands, and the umbrella's AC-PRC-004 full closure (E-08 listing exactly one file) is not achieved by this child alone.

## 3. Requirements (GEARS)

ID scheme: `REQ-PRC-*` items keep their umbrella IDs and texts (scoped here to this child's files); `REQ-README-*` items are child-local refinements, the prefix the traceability verb reads as separate requirements.

- **REQ-PRC-001** (Ubiquitous; README part, umbrella text kept): The four README files **shall** cite only `moai` CLI verbs, sub-verbs, and `/moai` slash commands that resolve on the tree. A CLI verb or sub-verb resolves when the tree's own build lists it in the COMMANDS block of its parent's `--help` output. A slash command resolves when `.claude/commands/moai/` registers it. The two umbrella exemptions (the four `docs-site/content/<locale>/multi-llm/cg-mode.md` retirement notices and the four `moai-todo.md` example pages) name no README file, so no README line is exempt under them; a README narrative line is exempt only under the umbrella's Q24 verdict, which exempts narrative examples that carry an explicit example marker — no README line carries one at this plan's measurement (R-01 lists every matching line; none has a marker).
- **REQ-PRC-003** (Ubiquitous, conditional on decision Q8): **Where** decision Q8 records a release tag, the Release badge in `README.md`, `README.ko.md`, `README.ja.md`, and `README.zh.md`, and the `version` and `releaseDate` lines of `docs-site/hugo.toml`, **shall** name that tag, and the two `hugo.toml` lines **shall** change in the same commit. **Where** decision Q8 records removal of the badge, the Release badge **shall** be absent from those four README files. (Q8 resolution for this card: tag = `v3.1.2`, `releaseDate` = `2026-08-21`; see § 5.)
- **REQ-PRC-004** (Ubiquitous, conditional on decision Q1): The Korean docs-site **shall** carry the CG retirement passage on exactly one page, the page that decision Q1 records (planner default: `docs-site/content/ko/multi-llm/cg-mode.md`), and every other page that carries the passage **shall** link to that page instead of repeating it. This child discharges the requirement on its own surface: the two README.ko.md copies and the ko `what-is-moai-adk.md` page; the seven other ko pages are outside this child's surface (§ 2).
- **REQ-README-001** (Event-driven): **When** a line of the four README files belongs to the flagged set — a line the scoped E-01 pattern (acceptance.md R-01) matches and that carries no example marker — the sweep **shall** replace the flagged citation with a current equivalent where one is registered on the tree, and otherwise delete it (Q13), and **shall not** leave any flagged line unchanged (Q24: unmarked lines are fixed).
- **REQ-README-002** (Ubiquitous): Where a flagged README line is a CG retirement paragraph, the sweep **shall** replace that paragraph with a link sentence pointing at that locale's `multi-llm/cg-mode` page (Q1: ko links to the canonical Korean page; en/ja/zh link to their own locale's cg-mode page, which carries the retirement notice in that locale — umbrella E-01a at `bdb7b764`).
- **REQ-README-003** (Ubiquitous): Every locale (ko, en, ja, zh) **shall** receive the same fix in the same change — badge line, flagged lines, and retirement paragraphs change in all four README files in one commit (Q17).
- **REQ-README-004** (Ubiquitous): The sweep **shall** change no file outside the edit surface of § 2 (this SPEC directory and `.moai/reports/t1720/` excepted), **shall** leave every non-flagged line of the four README files byte-identical, and **shall** create no release, no tag, and no push to a protected branch (Q8: no new release or tag is permitted).

## 4. Pinned facts (measured, `cb30e05b3e`)

- Badge line 24 of each README carries `img.shields.io/badge/Release-v3.2.0-blue.svg` (R-20a, verbatim in acceptance.md §C).
- The 16 E-01 README lines are listed verbatim in acceptance.md R-01n; per-locale sites: ko 58/291/386/591, en 58/296/387/592, ja 58/290/385/590, zh 58/289/384/589.
- README.ko.md lines 386 and 591 carry the exact sentence `설정 이전 안내와 함께 종료합니다` (2 copies, R-08r).
- hugo.toml lines 55–56: `version = "v3.2.0"`, `releaseDate = "2026-10-23"` (R-20b).

## 5. Decisions

Rows use the decision-index shape. Umbrella rows Q1, Q8, Q13, Q17 are owned here; their full texts are at `bdb7b764` decision-index.md and are not copied wholesale — the application to this child is recorded.

### Q1: Which single Korean page keeps the CG retirement passage?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: nine Korean pages and README.ko.md (2 copies) carry the sentence (R-08, R-08r). Choosing the surviving copy is editorial. Umbrella planner default applies to this child: keep `docs-site/content/ko/multi-llm/cg-mode.md` (rule: preserves current behavior).
Class: implementation-level
Child: README
Planner default applied: keep `docs-site/content/ko/multi-llm/cg-mode.md` (provisional, umbrella 0.1.1)
Operator verdict:

### Q8: Which release does the README badge and `hugo.toml` name until v3.2.0 GA?
Label: FOUNDER
Authority anchor: none in the committed tree (no badge policy in the committed docs or rules)
Why unresolved (umbrella row): the options — name the public Latest, keep the Draft, or remove the badge — each state a different shipped version to readers, and the umbrella ranked no default.
Class: product-level
Child: README
Resolution for this card: **leader lens, not an operator verdict.** The leader lens (dispatch to lane-14, 2026-10-26) resolved Q8 for card t1720 to **v3.1.2**, the public Latest. Evidence measured 2026-10-26T10:31:30Z (acceptance.md R-20c, R-20d): GitHub Latest = v3.1.2, published 2026-08-21T06:50:41Z; drafts v3.1.3 and v3.2.1 exist; no v3.2.0 release object exists (the "3.2 rollback" state); local tags list v3.2.0, v3.1.2, v3.1.1, v3.1.0 with no v3.1.3 tag. Consequence: REQ-PRC-003 applies with tag `v3.1.2` and `releaseDate = "2026-08-21"` (v3.1.2's measured publish date); the four badges and both hugo.toml lines change in the same commit; **no new release and no new tag is permitted** (REQ-README-004). The umbrella row stays open for the other children and the umbrella itself.
Operator verdict: (none — this is a leader-lens card-scoped resolution; the umbrella row remains open)

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved (umbrella row): replacing a citation changes what a reader is told to run; the default and the alternate apply per form.
Class: implementation-level
Child: README and SWEEPA to SWEEPF (this child applies it to the four README files)
Default: delete the stale citation (rule: smaller user-visible surface)
Alternate: replace with the current equivalent, where the operator names one for each form
Planner default applied: delete; replace only where a registered equivalent is already named on the tree (e.g. the GPT-models sentence already names `moai codex`)
Operator verdict:

### Q17: Do the same edits land in ko, en, ja, and zh in the same change?
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 (umbrella E-24 at `bdb7b764`), a committed document outside the authority register, so it cannot support a DECIDED or POLICY-COVERED label.
Why unresolved (umbrella row): whether that document's obligation binds this SPEC's edits is the operator's call.
Class: implementation-level
Child: cross-cutting (this child applies it to the four README files)
Planner default applied: yes — all four locales change in the same change (REQ-README-003)
Operator verdict:

### QR-1: Does ko what-is-moai-adk.md keep its own sentence copy alongside the new link?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the leader-lens surface names only "add the missing link"; REQ-PRC-004 says a non-canonical page links *instead of repeating*. The link is the named minimum; the sentence copy's disposition is this child's call under Q1.
Class: implementation-level
Default: replace the sentence copy with the link sentence (rule: the Q1 planner default — link instead of repeat)
Alternate: add the link and keep the sentence; the page then both repeats and links, and the umbrella's full E-08 closure is left to a later change
Operator verdict:

### QR-2: Do the en/ja/zh README retirement paragraphs change with the ko ones?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the exact sentence is Korean (R-08r counts ko only), but the en/ja/zh READMEs carry translations of the same two paragraphs and their lines are E-01-flagged (R-01n lists them). Q17's planner application asks that locales carrying the same passage change together.
Class: implementation-level
Default: replace both paragraphs in all four locales with link sentences to that locale's cg-mode page (rule: the Q17 planner default; also the only reading under which the scoped E-01 cell turns green, since the translated paragraphs are flagged lines)
Alternate: edit README.ko.md only; the en/ja/zh paragraphs stay flagged and the README partial criterion cannot close
Operator verdict:

### QR-3: How are the two prose citations per README (line 58 and the GPT-gateway callout) handled?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: Q13's default deletes a stale citation and replaces it only where a registered equivalent is named on the tree. Line 58 of each README cites both `moai cg` and `moai gpt`; the callout (ko 291 / en 296 / ja 290 / zh 289) cites `moai gpt`. Exact rewording is run-phase work; this row fixes the rule.
Class: implementation-level
Default: delete the flagged citation clauses; where the sentence already names a registered equivalent (`moai codex` in line 58; `moai migrate cg` where present), keep the sentence accurate without the flagged token (rule: the Q13 planner default)
Alternate: replace each flagged token with an explicit pointer to the registered equivalent in every affected sentence, wording per locale chosen at run phase
Operator verdict:

## 6. Interface to the GATE and FLAG children

- The GATE child (SPEC-V3R6-DOCS-PREREWRITE-GATE-001) owns the cited-verbs producer; this child's README partial criterion closes on the scoped pattern (acceptance.md R-01), independent of the producer.
- The FLAG child (t1726) owns the `--factory-run` / `--clear-policy` citations (umbrella Q18); those lines are not part of the E-01 family and are outside this child's criteria.
- The SWEEP children (t1721–t1725, t1727) own the docs-site page files. The seven ko pages that already link keep their sentence copies after this child lands; their removal belongs outside this child.

## 7. Out of Scope

### Out of Scope — other children's files

- Every docs-site page file except `docs-site/content/ko/core-concepts/what-is-moai-adk.md` — the SWEEP children own the page part of REQ-PRC-001, and the seven already-linking ko pages are explicitly outside this child's surface (leader-lens scope).
- `.github/workflows/**` (GATE child), the flag citations (FLAG child t1726), `docs-site/vercel.json` (REDIR child), assets and scripts (RESID child).

### Out of Scope — release operations

- Creating a release, publishing a GitHub release, tagging, and any Vercel deployment setting (Q8 resolution: no new release or tag is permitted; REQ-PRC-003 only re-labels the badge and hugo.toml lines).

### Out of Scope — rewrite content and adjacent prose

- The 300-line README rewrite and the docs-site information-architecture rewrite (plan stages 5-1 to 5-3).
- Any README prose line the scoped E-01 pattern does not flag (REQ-README-004 leaves non-flagged lines byte-identical); adjacent factual claims outside the flagged lines are recorded, not edited.

### Out of Scope — non-README version surfaces

- Any version string outside the four badge lines and hugo.toml lines 55–56 (CHANGELOG, other version mentions in README prose, template version constants).
