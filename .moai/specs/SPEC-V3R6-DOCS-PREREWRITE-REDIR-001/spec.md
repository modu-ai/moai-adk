---
id: SPEC-V3R6-DOCS-PREREWRITE-REDIR-001
title: "Docs Pre-Rewrite Cleanup — redirect repair (child 2 of 11 under SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, card t1718)"
version: "0.2.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "docs-site/vercel.json, docs-site/content/*/changelog"
lifecycle: spec-anchored
tags: "docs, docs-site, redirects, vercel, i18n, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 — Redirect repair (child 2 of 11)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1718 (Class C), child REDIR of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (card t1655). Every count and line number was measured on tree commit 0ddf5ca48b (tree a2c0d16b90c0e49371b6efec3084e6e270781bd7, clean); the umbrella's lines 88/93/98, 113/118, and 163/168 are unchanged between 8816de4dae and 0ddf5ca48b. Per the card's precondition, the moai-mx pair decision row (Q23) and its acceptance criterion (AC-PRC-011, extended to the pair) were recorded first. Status `draft`. |
| 2026-10-19 | 0.2.0 | manager-spec | r2: leader verdicts applied (board record `d-20261018T170818Z-a207`, decided by the leader under operator delegation, with Jev 0.88 and 0.92). Q23 = option (c): remove only the rule at line 163, keep line 168. Q21 = option (b): keep the two external changelog rules and amend REQ-PRC-010 so `https://github.com/modu-ai/moai-adk/releases` is an admitted destination. Q16: delete the four changelog index pages, no longer conditional. Q15 and Q17 stay `DEFAULT-APPLIED`. The five-file scope is final: `docs-site/vercel.json` and the four `changelog/_index.md` pages; the footer file is not changed. The decision-dependent end states collapse to the chosen ones. The check script gained an optional third argument that names the admitted external destination (acceptance.md E-R0, new SHA-256), and the controls were re-run. REQ-REDIR-001 to -003 and AC-REDIR-001 to -003 are restated for the chosen end state. Status stays `draft`. |

## Overview (WHY)

The deployed docs site (adk.mo.ai.kr, Vercel) carries 39 redirect rules in `docs-site/vercel.json`. Seven of them are the subject of this child, because they either point at nothing that exists or hide pages that do exist. The umbrella states two requirements for them: REQ-PRC-010 (every redirect destination resolves) and REQ-PRC-011 (no redirect source matches an existing page). Both are red on the pre-change tree.

Measured state on tree 0ddf5ca48b (full evidence in acceptance.md §C):

- **The three `/agency` rules** (sources at `docs-site/vercel.json` lines 88, 93, 98) send `/:locale/agency[/...]` and `/agency/...` to a `design` section that exists in no locale: 9 unresolved destination lines (E-R4). On the deployed site `/ko/agency` answers 308 to `/ko/design/`, then 308 to `/ko/design`, then 404 (E-R8, external).
- **The two changelog rules** (lines 113 and 118) send `/:locale/changelog` and `/changelog` to `https://github.com/modu-ai/moai-adk/releases`: 5 external lines (E-R4). The locale-aware source (line 113) also matches the four `changelog/_index.md` pages, so those pages are never served: 4 shadow lines (E-R5). On the deployed site `/ko/changelog` answers 307 to the GitHub URL (E-R8, external).
- **The moai-mx pair** (lines 163 and 168). The locale-aware source (line 163) matches the four existing `utility-commands/moai-mx.md` pages: 4 shadow lines (E-R5). The root-level source (line 168) matches no content page, because every page lives under a locale prefix. On the deployed site `/ko/utility-commands/moai-mx` answers 308 to the section index although the page file exists (E-R8, external). Thirteen files link to the shadowed URL (17 lines, E-R9).
- **The site footer** (`docs-site/layouts/partials/site-footer.html` line 9) links to `/changelog/`. That path is served by the root-level rule at line 118 (inference from the content tree and E-R8; acceptance.md G10). The chosen end state keeps that rule, so the footer file does not change.

### Recorded verdicts (board `d-20261018T170818Z-a207`)

- Q23 = option (c): remove only the rule at line 163 (it shadows the four pages); keep line 168 (it shadows nothing).
- Q21 = option (b): keep the rules at lines 113 and 118; REQ-PRC-010 is amended to admit the GitHub Releases URL.
- Q16: with the rule at line 113 kept, the four changelog index pages are shadowed, so they are deleted.
- Q15 (Default: remove the three `/agency` rules) and Q17 (Default: the four pages in one change) stay `DEFAULT-APPLIED`.

Expected end state: 35 redirect rules, 608 content pages, `shadow` and `resolve` (with the admitted URL) both exit 0 (scratch-copy proof in acceptance.md E-R14).

## Scope (WHAT)

Files this child changes (the card's five):

1. `docs-site/vercel.json` — remove the rule at line 163 (Q23) and the three `/agency` rules at lines 88, 93, 98 (Q15 Default); keep every other rule.
2. `docs-site/content/ko/changelog/_index.md`, 3. `.../en/changelog/_index.md`, 4. `.../ja/changelog/_index.md`, 5. `.../zh/changelog/_index.md` — deleted together (Q16, Q17).

Not changed: the footer file, the four `utility-commands/moai-mx.md` pages (the fix is in `vercel.json`), any other docs page or layout, `hugo.toml`, any workflow, rule, template, or Go source. `docs-site` is not mirrored into `internal/template/templates/` (E-R13), so no Template-First step applies.

### Locale scope (Q17 and the four-locale obligation)

The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` (lines 61 and 65) and digested in `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md` §2. For this child it is met by construction: `vercel.json` is locale-independent, and the four pages are deleted in one change. On a scratch copy, deleting only the en page makes `scripts/docs-i18n-check.sh` exit 1 ("missing translation"), but deleting only the ko page makes it exit 0 with three "orphan translation" warnings (E-R10b); the check must therefore also read `Warnings: 0`. No prose is translated.

## Requirements (GEARS)

REQ-PRC-010 and REQ-PRC-011 keep the umbrella's identifiers and base text (umbrella spec.md, revision 0.3.1). The clauses under each base text are this child's scope under the recorded verdicts.

- **REQ-PRC-011** (Ubiquitous; base text from the umbrella; the moai-mx rule is recorded first under Q23): The redirect sources in `docs-site/vercel.json` **shall** match zero existing content pages.
  - Under Q23 option (c), the redirect configuration **shall** contain no rule whose source is `/:locale(ko|en|ja|zh)/utility-commands/moai-mx`, each of the four `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-mx.md` pages **shall** remain published, and the root-level rule with source `/utility-commands/moai-mx` **shall** remain (it matches no content page).
  - Under Q16 (with Q17), the four `docs-site/content/{ko,en,ja,zh}/changelog/_index.md` pages **shall** be absent in the same change, because the locale-aware rule for `changelog` stays.

- **REQ-PRC-010** (Ubiquitous; base text from the umbrella, amended by Q21 option (b)): Every redirect destination in `docs-site/vercel.json` **shall** resolve to an existing content page of the same locale, or the redirect **shall** be removed, except the admitted external destination below.
  - **Where** the destination is exactly `https://github.com/modu-ai/moai-adk/releases` and the source is `/:locale(ko|en|ja|zh)/changelog` or `/changelog` (Q21 option (b)), the destination **shall** be admitted as an allowed redirect destination; the redirect configuration **shall** contain no other external destination.
  - Under Q15 Default, the redirect configuration **shall** contain no rule whose source names `agency`.

- **REQ-REDIR-001** (Ubiquitous; new): After the change, every internal reference to a URL whose serving the change alters **shall** resolve. In particular the footer link `/changelog/` **shall** keep resolving through the kept changelog rules, and the redirect rules for `/changelog` and `/:locale/changelog` **shall** remain after the changelog pages are deleted, so that the public changelog URL serves something.

- **REQ-REDIR-002** (Ubiquitous; new, scope confinement): The change **shall** modify only the five files named in § Scope and this SPEC's own records; within `docs-site/vercel.json` it **shall** remove only the four rules the recorded verdicts name (line 163 and lines 88, 93, 98).

- **REQ-REDIR-003** (Ubiquitous; new, locale parity): The four changelog index pages **shall** be deleted together, and `scripts/docs-i18n-check.sh` in its default strict mode **shall** exit 0 and report zero errors and zero warnings after the change.

## Out of Scope

Each item below is named so that a later change cannot claim it silently.

### Out of Scope — other redirect rules and the footer

- The 35 redirect rules that remain after the change (the kept changelog rules at lines 113 and 118, the root-level moai-mx rule at line 168, and the 32 rules outside this child such as `quality-commands/*`, `getting-started/*`, `kanban-*`, `book`, and `moai-rank`), the one rewrite, and the two header blocks.
- Adding new redirect rules.
- `docs-site/layouts/partials/site-footer.html` and the `footer_changelog` strings in `docs-site/i18n/*.yaml`.

### Out of Scope — page content and translation

- The prose of the four `moai-mx.md` pages and of every other docs page; stale-citation fixes belong to the SWEEP children.
- Translating any text into en, ja, or zh.

### Out of Scope — CI, link checking, and deployment

- The Hugo-build and link-check CI jobs and the i18n job's blocking mode belong to the GATE child (SPEC-V3R6-DOCS-PREREWRITE-GATE-001). This child's checks run locally; it adds no workflow.
- Vercel project settings, branch protection, and any deployment action. The live observations in acceptance.md (E-R8) are external and never recorded as a pass.

### Out of Scope — Go code and templates

- Any change under `internal/` or `cmd/`, and any change to `internal/template/templates/`.

## Dependencies and Parallel Work

- No `depends_on`. The umbrella's child table lists REDIR with no after-relations.
- Relation to GATE (t1717): GATE's link-check job (REQ-PRC-014) will inspect internal links once it lands; the footer link `/changelog/` keeps resolving through the kept rule, so REDIR does not add a failure for it. The two children share no file.
- File overlap: this author did not re-run the in-flight branch overlap measurement (the command was refused as a compound git form, acceptance.md G5). The lane's observation, relayed by the leader, is that none of twelve in-flight branches touches the card's five files; the command is `git diff --name-only 0ddf5ca48b...<branch> -- docs-site/vercel.json docs-site/content/*/changelog` printing nothing per branch. Under the chosen end state the footer file and the moai-mx pages are not touched, so the five-file observation covers the whole change.
- Open decisions that block the run phase: none. Q23, Q21, and Q16 carry the a207 verdicts; Q15 and Q17 carry `DEFAULT-APPLIED`. See decision-index.md.
