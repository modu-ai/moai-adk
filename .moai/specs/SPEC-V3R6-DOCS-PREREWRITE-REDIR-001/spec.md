---
id: SPEC-V3R6-DOCS-PREREWRITE-REDIR-001
title: "Docs Pre-Rewrite Cleanup — redirect repair (child 2 of 11 under SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, card t1718)"
version: "0.1.0"
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
| 2026-10-19 | 0.1.0 | manager-spec | Initial plan-phase authoring for card t1718 (Class C), child REDIR of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (card t1655). Every count and line number below was measured on tree commit 0ddf5ca48b (tree a2c0d16b90c0e49371b6efec3084e6e270781bd7, clean) in this planning pass; the umbrella's lines 88/93/98, 113/118, and 163/168 are unchanged between 8816de4dae and 0ddf5ca48b. Per the card's precondition, the moai-mx pair decision row (Q23) and its acceptance criterion (AC-PRC-011, extended to the pair) are recorded first, in decision-index.md and acceptance.md. The decision rows Q15, Q16, Q21, Q23, and the cross-cutting Q17 are authored in full in decision-index.md. Status stays `draft`. |

## Overview (WHY)

The deployed docs site (adk.mo.ai.kr, Vercel) carries 39 redirect rules in `docs-site/vercel.json`. Seven of them are the subject of this child, because they either point at nothing that exists or hide pages that do exist. The umbrella states two requirements for them: REQ-PRC-010 (every redirect destination resolves) and REQ-PRC-011 (no redirect source matches an existing page). Both are red on the current tree.

Measured state on tree 0ddf5ca48b (full evidence in acceptance.md §C):

- **The three `/agency` rules** (sources at `docs-site/vercel.json` lines 88, 93, 98) send `/:locale/agency[/...]` and `/agency/...` to a `design` section that exists in no locale. The redirect-check instrument prints 9 unresolved destination lines for them (E-R4). On the deployed site, `/ko/agency` answers 308 to `/ko/design/`, which answers 308 to `/ko/design`, which answers 404 (E-R8, external).
- **The two changelog rules** (sources at lines 113 and 118) send `/:locale/changelog` and `/changelog` to `https://github.com/modu-ai/moai-adk/releases`, an external URL, which REQ-PRC-010 does not admit (5 unresolved lines, E-R4). The locale-aware source (line 113) also matches the four `changelog/_index.md` pages, so those pages are never served: 4 shadow lines (E-R5). On the deployed site `/ko/changelog` answers 307 to the GitHub URL (E-R8, external).
- **The moai-mx pair** (sources at lines 163 and 168) send `/:locale/utility-commands/moai-mx` to `/:locale/utility-commands` and `/utility-commands/moai-mx` to `/ko/utility-commands`. The locale-aware source (line 163) matches the four existing `utility-commands/moai-mx.md` pages: 4 shadow lines (E-R5). The root-level source (line 168) matches no content page, because every page lives under a locale prefix (`defaultContentLanguageInSubdir = true`). On the deployed site `/ko/utility-commands/moai-mx` answers 308 to `/ko/utility-commands`, although the page file exists in the tree (E-R8, external). Thirteen files link to the shadowed URL (17 matching lines, E-R9).
- **One internal link depends on a rule this child may remove.** The site footer (`docs-site/layouts/partials/site-footer.html` line 9) links to `/changelog/`. That path is served only by the root-level rule at line 118: every content page lives under a locale prefix, so no page sits behind a root-level `/changelog` (E-R9; an inference from the content tree and E-R8, not a measurement with the rule removed, acceptance.md G10). Removing that rule without changing the footer would, by that inference, leave a dead footer link on every page of every locale.

The card (t1718) scopes five files: `docs-site/vercel.json` and the four `docs-site/content/{ko,en,ja,zh}/changelog/_index.md` pages. The finding above shows one decision combination for which the card's list of five files is the wrong list: under Q21 option (a) the footer file must change, and the four pages must stay (Q16 Alternate). That is recorded in decision-index.md Q21 and in REQ-REDIR-001 below, and it is not decided here.

## Scope (WHAT)

Files this child may change (five, per the card):

1. `docs-site/vercel.json` — the seven rules named above, and no other rule.
2. `docs-site/content/ko/changelog/_index.md`, 3. `.../en/changelog/_index.md`, 4. `.../ja/changelog/_index.md`, 5. `.../zh/changelog/_index.md` — deleted together or not at all, and only under the Q16 Default.

Conditional extra file: `docs-site/layouts/partials/site-footer.html`, only when REQ-REDIR-001 requires it (Q21 option (a), which also keeps the four pages). Changing it is a scope call for the leader (see decision-index.md Q21); the card does not authorize it.

Not changed: the four `utility-commands/moai-mx.md` pages (the fix for the shadowing is in `vercel.json`, not in the pages), any other docs page, any layout other than the conditional footer, `hugo.toml`, any workflow, rule, template, or Go source. `docs-site` is not mirrored into `internal/template/templates/` (0 matching tracked paths, E-R13), so no Template-First step applies.

### Locale scope (Q17 and the four-locale obligation)

The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` (lines 61 and 65) and digested in `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md` §2: a canonical (ko) change lands in en, ja, and zh in the same PR. For this child the obligation is met by construction: `vercel.json` is locale-independent, and the five card files already name all four locales. The one place it can fail is a partial deletion of the changelog pages. `scripts/docs-i18n-check.sh` in its default strict mode exits 0 on the current tree (E-R10). On a scratch copy, deleting only the en page makes it exit 1 with "missing translation", but deleting only the ko page makes it exit 0 with three "orphan translation" warnings (E-R10b). Exit status alone therefore does not catch a ko-only deletion; the check must also read `Warnings: 0`. No prose is translated by this child. Q17 itself is a cross-cutting umbrella row; its REDIR copy is in decision-index.md.

## Requirements (GEARS)

REQ-PRC-010 and REQ-PRC-011 keep the umbrella's identifiers and base text unchanged (umbrella spec.md, revision 0.3.1). The clauses under each base text are this child's verdict-dependent scope; a `Where` clause binds only after the operator records the named answer in decision-index.md.

- **REQ-PRC-011** (Ubiquitous; base text from the umbrella; the moai-mx pair is recorded first under Q23): The redirect sources in `docs-site/vercel.json` **shall** match zero existing content pages.
  - **Where** the operator answers Q23 with option (a), the redirect configuration **shall** contain no rule whose source names `utility-commands/moai-mx`, and each of the four `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-mx.md` pages **shall** remain published.
  - **Where** the operator answers Q23 with option (c), the redirect configuration **shall** contain no rule whose source carries the `:locale` group and names `utility-commands/moai-mx`, and each of the four pages **shall** remain published.
  - **Where** the operator answers Q23 with option (b), the amended REQ-PRC-011 **shall** name the exempt source by exact string, and the redirect configuration **shall** contain no other source that matches an existing content page.
  - **Where** the operator answers Q16 with the Default, the four `docs-site/content/{ko,en,ja,zh}/changelog/_index.md` pages **shall** be absent in the same change.
  - **Where** the operator answers Q16 with the Alternate, the redirect configuration **shall** contain no rule whose source names `changelog`, and the four pages **shall** remain published.

- **REQ-PRC-010** (Ubiquitous; base text from the umbrella): Every redirect destination in `docs-site/vercel.json` **shall** resolve to an existing content page of the same locale, or the redirect **shall** be removed.
  - **Where** the operator answers Q15 with the Default, the redirect configuration **shall** contain no rule whose source names `agency`.
  - **Where** the operator answers Q15 with the Alternate, each of the three `agency` rules **shall** name a destination that resolves to an existing content page of the same locale.
  - **Where** the operator answers Q21 with option (a), the redirect configuration **shall** contain no rule whose destination begins with `http`.
  - **Where** the operator answers Q21 with option (b), the amended REQ-PRC-010 **shall** list each admitted external destination by exact string, and the redirect configuration **shall** contain no external destination outside that list.

- **REQ-REDIR-001** (Ubiquitous; new in this child, raised by the footer finding): After the change, every internal reference to a URL whose serving the change alters **shall** resolve.
  - **When** the change removes the root-level changelog rule (line 118), the footer link target `/changelog/` **shall** be changed in the same change to a target that resolves.
  - The change **shall not** remove both the changelog redirect rules and the changelog index pages: at least one of the two **shall** remain, so that the public changelog URL serves something.

- **REQ-REDIR-002** (Ubiquitous; new, scope confinement): The change **shall** modify only the five files named in § Scope, the conditional footer file when REQ-REDIR-001 requires it, and this SPEC's own records; within `docs-site/vercel.json` it **shall** remove or amend only the rules that a recorded verdict names.

- **REQ-REDIR-003** (Ubiquitous; new, locale parity): The four changelog index pages **shall** be deleted together or not at all, and `scripts/docs-i18n-check.sh` in its default strict mode **shall** exit 0 and report zero errors and zero warnings after the change.

## Out of Scope

Each item below is named so that a later change cannot claim it silently. The seven in-scope rules are the only redirect rules this child touches.

### Out of Scope — other redirect rules

- The other 32 redirect rules in `docs-site/vercel.json` (including the `quality-commands/*`, `getting-started/*`, `kanban-*`, `book`, and `moai-rank` rules) and the one rewrite and two header blocks. They resolve today (E-R4 prints no line for them) and are not changed.
- Adding new redirect rules. If a verdict removes a rule and the operator wants the old URL to land elsewhere, that is a separate decision.

### Out of Scope — page content and translation

- The prose of the four `moai-mx.md` pages and of every other docs page; stale-citation fixes belong to the SWEEP children.
- Translating any text into en, ja, or zh. The `footer_changelog` strings in `docs-site/i18n/*.yaml` are not edited by this child.

### Out of Scope — CI, link checking, and deployment

- The Hugo-build and link-check CI jobs and the i18n job's blocking mode belong to the GATE child (SPEC-V3R6-DOCS-PREREWRITE-GATE-001). This child's checks run locally; it adds no workflow.
- Vercel project settings, branch protection, and any deployment action. The live observations in acceptance.md §C (E-R8) are external and are never recorded as a pass.

### Out of Scope — Go code and templates

- Any change under `internal/` or `cmd/`, and any change to `internal/template/templates/`.

## Dependencies and Parallel Work

- No `depends_on`. The umbrella's child table lists REDIR with no after-relations.
- Relation to GATE (t1717): GATE's link-check job (REQ-PRC-014) will flag a dead internal link such as the footer's `/changelog/` once it lands. REDIR and GATE do not overlap in files. If both land, REDIR's REQ-REDIR-001 is what keeps GATE's link check from going red on the footer.
- File overlap: this author did not re-run the in-flight branch overlap measurement (the command was refused as a compound git form, see acceptance.md §G, G5). The lane's observation, stated by the leader, is that none of twelve in-flight branches touches the card's five files; the command is `git diff --name-only 0ddf5ca48b...<branch> -- docs-site/vercel.json docs-site/content/*/changelog` printing nothing for each branch. That observation covers the five card files, not the conditional footer file or the four moai-mx pages.
- Open decisions that block the run phase: Q21 and Q23 (product-level; no operator verdict recorded). See decision-index.md.
