# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-REDIR-001

Revision 0.1.0 (2026-10-19). This child owns the full rows Q23, Q21, Q15, and Q16, and carries a copy of the cross-cutting row Q17. The rows moved here from the umbrella (SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, decision-index.md, revision 0.3.1, where each is a stub pointing at this child). The source text of the moved rows is at commit bdb7b764 of the umbrella's decision-index.md. `interview.decision_gate` is `on` (`.moai/config/sections/interview.yaml`), so this file is part of the artifact set and every row routes with one of the four labels.

The Q23 row is recorded first. The card names that as its before-start condition: audit finding B04 says the redirect rules at `docs-site/vercel.json` lines 163 and 168 hide the four existing `moai-mx.md` pages. The row, and its acceptance criterion (AC-PRC-011 in acceptance.md, which also comes first there), are written ahead of the other rows. The ordering is positional within these files; it is not a commit-order witness (the lane owns commits).

Labels: DECIDED and POLICY-COVERED need an authority anchor in the committed tree (product.md, a completed SPEC's HISTORY or Amendments row, `.moai/config/sections/*.yaml`, or the constitution). None of the rows below has one, so every row is FOUNDER. `.moai/docs/docs-site-i18n-rules.md` and `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md` are committed documents outside the authority register; the rows cite them as facts, not as authority. No row carries a recommendation. The `Default:` lines on Q15, Q16, and Q17 are applications of the published Default rule, not recommendations.

Evidence references (E-R<n>) point to acceptance.md §C, measured at tree commit 0ddf5ca48b.

### Q23: Are the two moai-mx redirect rules removed, or is REQ-PRC-011 amended to exempt the pair?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: `docs-site/vercel.json` line 163 redirects `/:locale(ko|en|ja|zh)/utility-commands/moai-mx` to `/:locale/utility-commands`, and line 168 redirects `/utility-commands/moai-mx` to `/ko/utility-commands` (both permanent; E-R2). The four pages `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-mx.md` exist, are not drafts, and are listed in each locale's `_meta.yaml` (E-R3, E-R9). The instrument prints 4 shadow lines for line 163 and none for line 168, because line 168 is a root-level URL and no content page lives at a root-level URL (E-R5). On the deployed site (E-R8, external, 2026-10-18T16:48Z to 16:54Z) the four locale URLs answer 308 to the section index and the root-level URL answers 308 to `/ko/utility-commands`; the page file exists in the tree, so the observed redirect is hiding it at least for that deployment. Seventeen lines in thirteen files link to the shadowed URL, among them the `moai mx` references in `advanced/mx-scanner-internals.md`, `cli-reference/graph.md`, and `workflow-commands/_index.md` in each locale, and `docs-site/data/menu/main.yaml` line 284 (E-R9). History: the rule was added on 2026-07-13 in e7e8115a21, which also deleted the page; the page was added back on 2026-07-14 in 7feec50a22 with the rule left in place (E-R11). The audit (B04) and the leader's ruling `d-20261014T060405Z-0fcd` ask for this row and an acceptance extension; REQ-PRC-011 says no redirect source may match an existing page.
Requirement: REQ-PRC-011
Class: product-level (option (a) and option (c) change what the public moai-mx URLs serve; option (b) amends a requirement)
Child: REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001), decision row Q23
Options: (a) remove both rules (lines 163 and 168); (b) amend REQ-PRC-011 by operator decision to exempt the pair, naming the exempt source by exact string; (c) remove only the locale-aware rule at line 163, which is the one that matches pages, and keep the root-level rule at line 168 (option (c) is added in this revision: the umbrella listed (a) and (b), and the instrument shows that line 168 matches no page). No Default is ranked.
Operator verdict:

### Q21: How are the two external changelog redirect rules handled?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: `docs-site/vercel.json` line 113 (`/:locale(ko|en|ja|zh)/changelog`) and line 118 (`/changelog`) send the reader to `https://github.com/modu-ai/moai-adk/releases` (temporary redirects; E-R2). REQ-PRC-010 requires every redirect destination to resolve to a same-locale content page, or the redirect to be removed; the instrument prints 5 unresolved lines for these two rules (E-R4). REQ-PRC-010 is not amended by this revision. The options change what the public changelog URLs serve: (a) remove the two rules, after which `/ko/changelog` serves the page that exists there (if Q16 keeps the pages) and `/changelog` serves nothing; (b) amend REQ-PRC-010 to admit operator-approved external destinations, with a separate check. Two facts bear on the choice. First, the root-level rule at line 118 is the only thing that keeps the footer link alive: `docs-site/layouts/partials/site-footer.html` line 9 links to `/changelog/`, which resolves today as 308 to `/changelog` and then 307 to the GitHub URL (E-R8, E-R9); the pages live under locale prefixes, so a root-level `/changelog` has no page behind it (an inference from the content tree and E-R8; acceptance.md G10). Under option (a) the footer link is dead on every page unless `site-footer.html` changes, and that file is not one of the card's five (a scope call outside the card). Second, Q21 and Q16 are coupled (see Interaction).
Requirement: REQ-PRC-010
Class: product-level (the public changelog URLs change what they serve under either option)
Child: REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001), decision row Q21; decided inside this child (operator decision `d-20261014T114834Z-e039`)
Options: (a) remove the two rules; (b) amend REQ-PRC-010 to admit the GitHub Releases destination by exact string, with a separate check. No Default is ranked.
Interaction (Q21 x Q16): the four combinations are (b) with Q16 Default (rules kept, pages deleted): `/changelog` and `/ko/changelog` still redirect to GitHub, no page is shadowed, five files change. (a) with Q16 Alternate (rules removed, pages kept): the four pages are served at `/ko/changelog` and siblings, `/changelog` serves nothing, the footer link needs a change (the change touches `vercel.json` and `site-footer.html`, no page is deleted). (a) with Q16 Default (rules removed and pages deleted): the public changelog URL serves nothing and the footer link is dead; REQ-REDIR-001 forbids this combination. (b) with Q16 Alternate is not coherent: the Alternate removes the rules that (b) keeps.
Operator verdict:

### Q15: Are the three redirect rules for `/agency` removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the three rules (`docs-site/vercel.json` lines 87 to 101; sources at lines 88, 93, and 98) take `/:locale/agency/:path*`, `/:locale/agency`, and `/agency/:path*` to `/:locale/design[/:path*]` and `/design/:path*`, a section that exists in no locale (E-R4: 9 unresolved lines; E-R3c: no `design` path in any locale). The umbrella reopened this row because its first ranking criterion (preserves current behavior) rested on an unmeasured inference that the source paths already return not-found. That premise is now measured: on the deployed site `/ko/agency` answers 308 to `/ko/design/`, which answers 308 to `/ko/design`, which answers 404; `/agency` answers 308 to `/design/`, and `/design` answers 404 (E-R8, external, 2026-10-18T16:48Z to 16:54Z). The reader's end state is a not-found page before and after removal. No internal link to `/agency` or a `/design` section was found in `docs-site` (E-R9). The deployed commit is not known to equal 0ddf5ca48b (gap G1).
Class: implementation-level. Justification: the rules route to a destination that resolves to no page in any locale (E-R4, E-R8), so removing them loses no reader-visible destination. If a later measurement shows a reader-reachable destination for any of the three sources, this row is reclassified product-level (fail-closed).
Child: REDIR
Default: remove the three rules (rule: preserves current behavior; the not-found end state is unchanged)
Alternate: point the rules at an existing page chosen by the operator
Operator verdict: DEFAULT-APPLIED 2026-10-18T17:01:34Z claude-sonnet-5-5 manager-spec (plan close; premise measured in E-R8)

### Q16: Are the four shadowed changelog index pages removed? (reopened in the umbrella)
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the umbrella reopened this row because the planner's earlier application of the Default rested on an unmeasured reachability inference (gap G5 there). "Reopened" means: the earlier planner application was withdrawn until a live request or consumption check measures the current behavior. This revision measures it: the locale-aware changelog rule at line 113 matches the four `changelog/_index.md` pages (E-R5), and on the deployed site `/ko/changelog`, `/en/changelog`, `/ja/changelog`, and `/zh/changelog` each answer 307 to GitHub (E-R8, external), so the pages are not served. The card text says "delete the four index pages according to Q16": that is the Default below, not an operator or leader verdict; no committed record answers Q16. The en, ja, and zh pages say in their own text that the address redirects to GitHub Releases (E-R12); the ko page describes release notes at length and links to GitHub Releases without saying so. The four files are 736 to 2,487 bytes. Q16 is coupled to Q21 (see Q21 Interaction): the Default is coherent with Q21 option (b) and is forbidden by REQ-REDIR-001 together with Q21 option (a).
Class: implementation-level. Justification: with the redirect rules kept, the four pages are unreachable, so deleting them changes no reader-visible behavior (E-R8). If a measurement shows any page is reader-reachable, this row is reclassified product-level (fail-closed). This classification holds only while Q21 is not answered (a).
Child: REDIR
Default: remove the four pages (rule: preserves current behavior, because the redirect keeps serving the public URL)
Alternate: remove the redirect rules and keep the pages
Condition: the applied Default stands only while Q21 is not answered (a). An answer of (a) to Q21 re-opens this row, and the run phase deletes no page before the Q21 verdict is read (plan.md M0).
Operator verdict: DEFAULT-APPLIED 2026-10-18T17:01:34Z claude-sonnet-5-5 manager-spec (plan close; premise measured in E-R5 and E-R8; Condition applies)

### Q17: Do the same edits land in en, ja, and zh in the same change? (cross-cutting; REDIR copy)
Label: FOUNDER
Authority anchor: none in the authority register. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 and digested in `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md` §2; both are committed documents outside the authority register, so they cannot support a DECIDED or POLICY-COVERED label.
Why unresolved: for REDIR the question has one live case: the four changelog pages. `vercel.json` is locale-independent, and the card's file list already names all four locales. On a scratch copy of the tree, deleting only the ko page leaves `scripts/docs-i18n-check.sh` at exit 0 with three "orphan translation" warnings, and deleting only the en page makes it exit 1 with "missing translation" (E-R10b). Whether the obligation binds this child's deletions is the operator's call, since the document is not an authority.
Class: implementation-level
Child: cross-cutting (REDIR, README, SWEEPA to SWEEPF, FLAG); REDIR copy
Default: delete the four pages in the same change (rule: preserves current behavior; the strict i18n check keeps reporting zero errors and zero warnings)
Alternate: delete the ko page first and the other three later
Operator verdict: DEFAULT-APPLIED 2026-10-18T17:01:34Z claude-sonnet-5-5 manager-spec (plan close; applies only if Q16 deletes the pages)

## Rows that block the run phase

| Row | Class | Verdict state | Effect |
|-----|-------|---------------|--------|
| Q23 | product-level | empty | M1 does not start the moai-mx edit until the operator answers |
| Q21 | product-level | empty | M1 does not start the changelog-rule edit until the operator answers; the answer also decides whether the footer file (outside the card's five) changes |
| Q15 | implementation-level | DEFAULT-APPLIED | M1 may remove the three `/agency` rules |
| Q16 | implementation-level | DEFAULT-APPLIED, conditional on Q21 | M2 starts only after the Q21 verdict is read |
| Q17 | implementation-level | DEFAULT-APPLIED | none for REDIR beyond the four-page rule |
