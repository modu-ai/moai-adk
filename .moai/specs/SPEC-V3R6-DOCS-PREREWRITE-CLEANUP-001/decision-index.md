# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001

Revision 0.2.1 (2026-10-14). Each row states one decision the operator does not settle in the interview. No row carries a recommendation.

Revision 0.2.1 corrects revision 0.2.0 under leader ruling `d-20261014T070143Z-75ac`: the G21 gap sentence in acceptance.md, the E-04 count (22 files, 45 matching lines), and the Q2 reference scope (E-09b and E-09b2). No label, class, default, or verdict changed.

Revision 0.2.0 changes five things. It follows the umbrella split (decision Q20 = (c), `d-20261014T055241Z-1bc6`) and two later operator records (`d-20261014T060405Z-0fcd` and `d-20261014T061812Z-f5ca`):

1. Every row carries a `Child:` line that names the child SPEC owning the decision. Q12 and Q22 are parent rows (operator-held). Q20 is the umbrella's own tier row.
2. Q20 carries its operator verdict: (c), split into child SPECs with tier M kept.
3. Q23 is new. It covers the moai-mx redirect pair in `docs-site/vercel.json` (lines 163 and 168), raised by the leader's B04 ruling and owned by the REDIR child.
4. Q24 is new. It records the operator's verdict on N2: only narrative examples that carry an explicit example marker are exempt from the stale-citation check.
5. Every other `Operator verdict:` stays empty. Every `Planner default applied:` line is kept unchanged; those values are provisional, not verdicts.

Labels: DECIDED and POLICY-COVERED need a committed anchor in the authority register (committed artifacts only). None of the rows below has one, so every row routes to FOUNDER or EVIDENCE-NEEDED. `.moai/docs/docs-site-i18n-rules.md` is a committed document outside the authority register; the rows cite it as a fact, not as an authority.

Evidence references (E-nn) point to acceptance.md §C, measured at tree f35103480f.

### Q1: Which single Korean page keeps the CG retirement passage?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: nine Korean pages carry the exact sentence (E-08). Choosing the surviving copy is an editorial decision. REQ-PRC-004 is conditional on this row. Seven of the eight other pages already link to the CG page (E-08b); `what-is-moai-adk.md` does not.
Class: implementation-level
Child: README
Default: keep `docs-site/content/ko/multi-llm/cg-mode.md` (rule: preserves current behavior)
Alternate: a new dedicated migration page
Planner default applied: keep `docs-site/content/ko/multi-llm/cg-mode.md` (provisional, 0.1.1)
Operator verdict:

### Q2: Do the five unreferenced images go, or do they gain a reference?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: no live reference from documentation, the docs site, templates, or code exists in the scoped paths (E-09b: no output, exit 1). The whole-tree check E-09b2 records one tracked file, `.moai/reports/historical/gpt-final-design-20260914-b40ff9f4/inventory.json`, which inventories past state and is not a live reference, and the Hugo source root `docs-site/` does not contain them. The deletion removes no displayed content. Whether a future page needs them is not decided. REQ-PRC-005 is conditional on this row.
Class: implementation-level
Child: RESID
Default: delete (rule: preserves current behavior)
Alternate: keep and add a reference
Planner default applied: delete (provisional, 0.1.1)
Operator verdict:

### Q3: What byte budget does the README OG image meet after re-encoding?
Label: EVIDENCE-NEEDED
Authority anchor: n/a
Why unresolved: the current file is 7,837,648 bytes (E-10). No committed rule fixes a budget, so REQ-PRC-006 is conditional on this row and this SPEC sets no number. The needed evidence: re-encodes of the same artwork at two or three budgets, each checked for legibility at README width.
Class: implementation-level
Child: RESID
Operator verdict:

### Q4: Are the two Nextra-named scripts deleted?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: `docs-site/scripts/fix-mdx-formatting.js` (the token `nextra` on 2 lines) and `scripts/convert-nextra-to-hextra/main.go` (5 lines) carry the Nextra name (E-11a); the first is Nextra-named by content, not by file name. The caller search found no workflow, Makefile, lefthook, Hugo, Vercel, or script caller outside their own folders (E-12). REQ-PRC-007 is conditional on this row. Deleting them is still a removal the operator should see.
Class: implementation-level
Child: RESID
Default: delete both (rule: preserves current behavior; no invocation path exists in the tree, E-12)
Alternate: keep both and rename them to describe their history
Planner default applied: delete both (provisional, 0.1.1)
Operator verdict:

### Q5: Are `translate.mjs` and `migrate-frontmatter.sh` kept?
Label: EVIDENCE-NEEDED
Authority anchor: n/a
Why unresolved: `translate.mjs` is the documented Korean-to-other-locales pipeline (its own README, E-12), and `migrate-frontmatter.sh` is a Hextra-to-geekdoc migration (E-12). Whether maintainers still run either is not recorded anywhere in the tree. The row gates M1's classification only; it does not gate the Nextra criterion (AC-PRC-007). The needed evidence: a maintainer's statement of use, or a dated run record.
Class: implementation-level
Child: GATE (M1 input only)
Operator verdict:

### Q6: Does the Python-era `docs-site/README.md` go or get replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the file is 42,512 bytes (E-13) and describes a previous generation of the docs toolchain. Replacement would need content that the rewrite will define. REQ-PRC-008 names both branches and is not pre-decided.
Class: implementation-level
Child: RESID
Default: delete (rule: smaller user-visible surface)
Alternate: replace with a short README that describes the current docs-site
Planner default applied: delete (provisional, 0.1.1)
Operator verdict:

### Q7: Is `docs-site/static/llms.txt` removed or regenerated?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the file is served at the public docs URL, last changed 2026-04-20T06:15:43Z (E-14a), and carries one EARS-style line (E-14b). Either choice changes a public artifact. The regenerate branch needs a generator, and the tree contains none that writes that path (verdict section 5, item 8; not re-measured in 0.1.1). A generator outside the SPEC's path set would be a separate deliverable. The operator's verdict must name the generator or choose removal.
Class: product-level
Child: RESID
Default: none ranked here. The published rule would keep the file, but the regenerate branch has no generator, so the operator decides.
Alternate: remove the file
Operator verdict:

### Q8: Which release does the README badge and `hugo.toml` name until v3.2.0 GA?
Label: FOUNDER
Authority anchor: none in the committed tree (no badge policy in the committed docs or rules, E-22)
Why unresolved: the public Latest is v3.1.2, and v3.1.3 is a Draft (E-20c, external). The four badges and `hugo.toml` show v3.1.3 (E-20a, E-20b). REQ-PRC-003 is conditional on this row. The options are: name the public Latest v3.1.2; leave v3.1.3 (a Draft on the public list); or remove the release badge until GA. Each option states a different shipped version to readers. No Default is ranked: the published rule would keep v3.1.3, a Draft, and the operator must weigh that.
Class: product-level
Child: README
Operator verdict:

### Q9: Does the i18n parity job block pull requests from this change onward?
Label: FOUNDER
Authority anchor: none in the committed tree (SPEC-DOCS-SITE-001 is `implemented` and silent on strict mode, E-23)
Why unresolved: the strict run passes on this tree with 0 errors and 0 warnings (E-16d). The workflow's comment names the Phase 2 flip trigger as "0 ERROR drifts on main" (`.github/workflows/docs-i18n-check.yml:13–14`). This tree is the pull-request tree and meets that count; `main` itself was not measured. A workflow comment is not a register authority. Blocking pull requests changes what every contributor waits on. No Default is ranked: the published rule would keep the advisory behavior, which the card does not want.
Class: implementation-level
Child: GATE
Operator verdict:

### Q10: Which Hugo version does the new CI build job use?
Label: FOUNDER
Authority anchor: `docs-site/vercel.json` `HUGO_VERSION` (committed; E-17b)
Why unresolved: a job that builds with another version would not test the build that ships. REQ-PRC-013 is conditional on this row. The choice is between the version the deploy uses and a newer one.
Class: implementation-level
Child: GATE
Default: `0.160.1` (rule: preserves current behavior; the deployed build uses this version, and the local binary matches it, E-17a)
Alternate: a newer Hugo version chosen by the operator
Planner default applied: `0.160.1` (provisional, 0.1.1)
Operator verdict:

### Q11: Which link-check tool runs in the new job?
Label: EVIDENCE-NEEDED
Authority anchor: n/a
Why unresolved: no link checker is referenced in the workflows (E-18a), and none of the four probed tools is installed locally: lychee, linkinator, htmltest, and markdown-link-check (E-18b). The candidates have not been run against the built site. The needed evidence: each candidate run on the Hugo output, with one injected broken internal link observed as red and the tree observed as green.
Class: implementation-level
Child: GATE
Operator verdict:

### Q12: Are the three new jobs added to the branch protection's required checks?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the current required contexts are five, and none of the three jobs is among them (E-21, external). Changing branch protection is an operator-held action on an external shared system (keep-set). The operator form is needed, and the change cannot be observed from the tree. If REQ-PRC-002's cited-verbs step becomes a required context, it is a fourth candidate for the same answer.
Class: implementation-level
Child: parent (operator-held)
Operator verdict:

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: 20 distinct stale forms are cited (acceptance §B; E-01 to E-04). Some have a current equivalent: `mcp-server.md` pairs each absent CLI form with a registered MCP tool name (for example `moai codex job status` with `mcp__moai__codex_job_status`, E-02). Some have none. Replacing a citation changes what a reader is told to run. The default and the alternate apply per form.
Class: implementation-level
Child: README and SWEEPA to SWEEPF
Default: delete the stale citation (rule: smaller user-visible surface)
Alternate: replace the citation with the current equivalent, where the operator names one for each form
Planner default applied: delete the stale citation (provisional, 0.1.1)
Operator verdict:

### Q14: Is the `moai-design` page kept, rewritten, or removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: correction (0.1.1). The earlier premise, that the `moai-codemaps` page cites the absent CLI verb, is wrong: that page's 80 matches are the registered slash command `/moai codemaps` (`.claude/commands/moai/codemaps.md` exists; E-01s, E-26b), so the page's citations resolve and it is not stale under REQ-PRC-001. The open decision is the `moai-design` page: `/moai design` is not user-invocable (E-26a) and is not a registered slash command (E-26b), and the page cites it on 4 to 5 lines per locale (E-04). The options are removal, a rewrite around the verb that replaces it, or an internal-only marker.
Class: product-level
Child: SWEEPF
Operator verdict:

### Q15: Are the three redirect rules for `/agency` and `/design` removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the three rules (`docs-site/vercel.json` lines 88–100) take `/:locale/agency/:path*`, `/:locale/agency`, and `/agency/:path*` to `/:locale/design` or `/design/:path*`, a section that exists in no locale (E-15b, E-15d-design). REQ-PRC-010 is not amended by this row. The source paths return not-found today only by inference from the content tree. No live request was measured (gap G5).
Reopened (0.1.1): the first ranking criterion (preserves current behavior) rests on that unmeasured inference, so the planner's application below is provisional until a live request or a consumption check measures the current behavior.
Class: implementation-level. Justification: the change edits routing configuration whose destinations resolve to no page in any locale, so no reader-visible destination is lost. If the reachability premise fails on verification, this row is reclassified product-level (fail-closed).
Child: REDIR
Default: remove the three rules (rule: preserves current behavior, if the premise holds)
Alternate: point the rules at an existing page chosen by the operator
Planner default applied: remove the three rules (provisional, 0.1.1; reopened)
Operator verdict:

### Q16: Are the four shadowed changelog index pages removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the redirect rules at `docs-site/vercel.json` lines 113–116 and 118–121 point the changelog URLs at GitHub Releases (temporary redirects), and the four `changelog/_index.md` pages sit under the same URLs (E-15c, E-15d-changelog). Whether the pages are reachable is inferred from the redirect precedence (G5). Removing the pages with the rules kept changes no reader-visible behavior only if that inference holds.
Reopened (0.1.1): the premise is unmeasured, so the planner's application below is provisional.
Class: implementation-level. Justification: if the reachability inference holds, removing unreachable pages changes no reader-visible behavior. If verification shows that any of the four pages is reader-reachable, this row is reclassified product-level (fail-closed).
Child: REDIR
Default: remove the four pages (rule: preserves current behavior, if the premise holds)
Alternate: remove the redirect rules and keep the pages
Planner default applied: remove the four pages (provisional, 0.1.1; reopened)
Operator verdict:

### Q17: Do the same edits land in en, ja, and zh in the same change?
Label: FOUNDER
Authority anchor: none in the authority register. Correction (0.1.1): the four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 ("동일 PR 내에서 en/zh/ja 모두 반영"; "ko/en/zh/ja 4개 locale 해당 파일 모두 수정", E-24). That document is outside the authority register (manager-spec.md), so it cannot support a DECIDED or POLICY-COVERED label. The row stays FOUNDER.
Why unresolved: the card scopes the cleanup to the Korean pages where it counts the copies, and the en, ja, and zh copies were measured only as family counts (E-01 to E-04), not per page. The rule document's obligation is stated for content changes in general. Whether it binds this SPEC's edits is the operator's call, since the document is not an authority.
Class: implementation-level
Child: cross-cutting (REDIR, README, SWEEPA to SWEEPF, FLAG)
Operator verdict:

### Q18: Is each `--factory-run` and `--clear-policy` citation kept, with the flag added to help, or removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: both flags are defined and parsed in the Go source (E-05b: 101 lines in 24 files) and cited on 36 lines across the READMEs and the factory and launcher pages (E-05). The help text of `factory`, `goal`, `codex`, and `cc` does not name them (E-25a; `cc` was not measured in 0.1.0). Adding help text changes the command surface a reader sees. Removing the citations removes documented behavior of a real flag. The flags are outside AC-PRC-001 (verdict D3). If the operator wants a flag clause, it needs its own REQ with its own disposition.
Class: product-level
Child: FLAG (GATE input)
Operator verdict:

### Q19: Is the maintainer absolute path in `docs-site/scripts/README_TRANSLATE.md` removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: line 9 names `/Users/goos/MoAI/moai-docs/scripts/translate.mjs`, a maintainer path in a tracked file (E-12). Removing it changes no behavior of any command. The row is open because the path is outside the seven card items, and its edit has no REQ or AC yet (verdict O7, open).
Class: implementation-level
Child: RESID
Default: remove the absolute path and keep the relative command (rule: preserves current behavior; the command is unchanged)
Alternate: leave the line as it is and record it under the card that owns translation tooling
Planner default applied: remove the absolute path (provisional, 0.1.1)
Operator verdict:

### Q20: Is this SPEC's tier M, or L with the Tier L artifact set?
Label: FOUNDER
Authority anchor: none in the committed tree that decides this SPEC's tier. The tier table is `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.
Why unresolved: the frontmatter says `tier: M` (three artifacts, PASS threshold 0.80). The planned edit set spans more than the 15 files that Tier M allows. The stale families alone touch 48 files for CLI forms (E-01) and 22 files for slash forms (E-04; 45 matching lines), and the image, script, workflow, redirect, and skill edits add to that. Tier L requires `design.md` and `research.md` and a 0.85 threshold. The operator decides between the two options. This revision keeps `tier: M`, authors neither file, and does not decide the question.
Options: (a) keep Tier M, reduce the scope to 15 files or fewer, and state the reasoning; (b) move to Tier L, author `design.md` and `research.md`, and re-audit at 0.85; (c) split into child SPECs, each tier M, with the umbrella keeping tier M.
Class: implementation-level. This row changes the artifact set and the audit gate, not shipped behavior. No Default is ranked.
Child: parent (umbrella tier)
Operator verdict: (c) split into child SPECs, tier M kept — operator, d-20261014T055241Z-1bc6 (2026-10-14T05:52:41Z)

### Q21: How are the two external changelog redirect rules handled?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: `docs-site/vercel.json` lines 113–121 send `/:locale(ko|en|ja|zh)/changelog` and `/changelog` to `https://github.com/modu-ai/moai-adk/releases` (E-15c). REQ-PRC-010 requires every redirect destination to resolve to an existing same-locale content page, or the redirect to be removed. These two destinations do not meet it, and REQ-PRC-010 is not amended by this SPEC. The options change what the public changelog URLs serve: (a) remove the two rules, after which the URLs serve the four changelog pages unless Q16 also removes them (then they would serve nothing); (b) amend REQ-PRC-010 to admit operator-approved external destinations, with a separate check.
Class: product-level (the public changelog URLs change what they serve under either option)
Child: REDIR
Operator verdict:

### Q22: Is REQ-PRC-015 moved to an operator-held section, or is the branch-protection exclusion amended?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: REQ-PRC-015 requires the three jobs to be required status checks on `main` (external), while the Out of Scope section excludes changing the branch protection of `main` (spec.md). The requirement and the exclusion conflict, and neither is decided here. Revision 0.2.0 places REQ-PRC-015 under an operator-held heading in spec.md; that placement is not the verdict, and the exclusion stays as written until Q22 is answered. The current required contexts are five (E-21). The options: (a) move the closure out of the requirements into an operator-held section, keeping the exclusion; (b) amend the exclusion to name this change as an operator-held action.
Class: implementation-level (an external setting; keep-set)
Child: parent (operator-held)
Operator verdict:

### Q23: Are the two moai-mx redirect rules removed, or is REQ-PRC-011 amended to exempt the pair?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: `docs-site/vercel.json` lines 163 and 168 redirect `/:locale(ko|en|ja|zh)/utility-commands/moai-mx` to `/:locale/utility-commands`, and `/utility-commands/moai-mx` to `/ko/utility-commands` (both permanent). The four existing `utility-commands/moai-mx.md` pages (ko, en, ja, zh) sit under those source URLs, so the pair shadows them, which REQ-PRC-011 says no redirect source may do. Audit B04 names this pair; AC-PRC-011 did not cover it. The leader's ruling in `d-20261014T060405Z-0fcd` adds this row and an AC-PRC-011 extension, both owned by the REDIR child.
Requirement: REQ-PRC-011
Class: product-level (option (a) changes what the public moai-mx URLs serve; option (b) amends a requirement)
Child: REDIR
Options: (a) remove the two shadowing rules; (b) amend REQ-PRC-011 by operator decision to exempt the pair. No default is ranked.
Operator verdict:

### Q24: Which narrative examples are exempt from the stale-citation check (N2)?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the E-04 slash family also counts narrative and error examples: `docs-site/content/en/getting-started/faq.md` line 24 (`/moai init` described as a misconception) and `docs-site/content/en/core-concepts/verification-claim-integrity.md` line 84 (a retired `/moai brain` narrative). The tree does not decide whether lines of that kind are exempt from the REQ-PRC-001 check (E-04, N2).
Class: implementation-level
Child: SWEEPE
Operator verdict: exempt only narrative examples that carry an explicit example marker; lines without a marker are fixed — operator, d-20261014T060405Z-0fcd (2026-10-14T06:04:05Z)
