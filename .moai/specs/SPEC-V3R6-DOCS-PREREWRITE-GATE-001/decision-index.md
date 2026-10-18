# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-GATE-001

Revision 0.1.0 (2026-10-18). Each row states one decision the operator does not settle in the interview. Rows Q5, Q9, Q10, Q11, Q18, and Q22 come from the parent's decision index (SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, revision 0.3.1) under the numbers the parent gave them. Rows Q25 and Q26 surfaced while measuring in this planning pass; the parent's numbering ends at Q24, so these two numbers are proposed and the leader confirms them. Every observation below was taken on tree `ae34f268319ce91f028ee692f25c1bd476f9e6e5` unless it says otherwise; evidence ids (EG-nn, EP-nn) point to acceptance.md §C, where the command and the verbatim output are recorded.

Labels: DECIDED and POLICY-COVERED need a committed anchor in the authority register (committed artifacts only: `.moai/project/product.md`, prior completed SPECs' HISTORY and `## Amendments` rows, `.moai/config/sections/*.yaml`, the constitution). None of the rows below has one, so every row routes to FOUNDER or EVIDENCE-NEEDED. Workflow comments, `vercel.json`, and `.moai/docs/*` are committed documents outside the register; the rows cite them as facts, not as authority.

No row carries a recommendation. A `Default:` line is the output of the published Default rule (the option that preserves current behavior; else the option whose undo is a single revert of this SPEC's own commits; else the option with the smaller user-visible surface), not a preference. At plan close, implementation-level rows that carry a `Default:` and an empty verdict are filled with `DEFAULT-APPLIED`. Rows that are open block the autonomous Kickoff and go to the operator; the gating column of the summary below says which acceptance criteria each open row holds.

| Row | Label | Owner | Verdict state | Gates |
|-----|-------|-------|---------------|-------|
| Q5 | EVIDENCE-NEEDED | GATE (producer input) | open | none of AC-PRC-002, -012, -013, -014 |
| Q9 | FOUNDER | GATE | open | AC-PRC-012 |
| Q10 | FOUNDER | GATE | DEFAULT-APPLIED | AC-PRC-013 |
| Q11 | EVIDENCE-NEEDED | GATE | open | AC-PRC-014 (and the conditional file `docs-site/.htmltest.yml`) |
| Q18 | FOUNDER | FLAG (input to GATE) | open at FLAG | none of the four criteria |
| Q22 | FOUNDER | GATE (leader-confirmed) | open | none of the four criteria |
| Q25 | FOUNDER | GATE (proposed number) | open | the trigger shape of AC-PRC-002, -013, -014 |
| Q26 | FOUNDER | GATE (proposed number) | DEFAULT-APPLIED | AC-PRC-014 |

### Q5: Are `translate.mjs` and `migrate-frontmatter.sh` kept?
Label: EVIDENCE-NEEDED
Authority anchor: n/a
Why unresolved: `translate.mjs` is the documented Korean-to-other-locales pipeline (its own README, `docs-site/scripts/README_TRANSLATE.md`), and `migrate-frontmatter.sh` is a Hextra-to-geekdoc migration. Whether maintainers still run either is not recorded in the tree. The tracked tree holds no caller outside `docs-site/scripts/` (EP-13: `git grep` over the whole tree, this SPEC's parent directory excluded, printed nothing; the zero-match exit status is not displayed by the tool, gap G-3). The row is an input to the checker's inventory only: the checker resolves `moai` verbs and `/moai` slash forms in code spans (C-1 to C-3) and neither script is such a form, so no output of the checker, and no acceptance criterion of this SPEC, changes with the answer. The needed evidence: a maintainer's statement of use, or a dated run record.
Class: implementation-level
Operator verdict:

### Q9: Does the i18n parity job block pull requests from this change onward?
Label: FOUNDER
Authority anchor: none in the committed tree (SPEC-DOCS-SITE-001 is `implemented` and silent on strict mode; the workflow's own header comment names the flip trigger, and a workflow comment is not a register authority)
Why unresolved: the facts the operator weighs, all measured. The strict run passes on this tree with 0 errors and 0 warnings (EP-01). The header comment's flip trigger is "scripts/docs-i18n-check.sh reports 0 ERROR drifts on main" (`docs-i18n-check.yml` lines 13 to 14); the parent recorded that `main` itself was not measured, and it is measured here: on a tree extracted from `origin/main` at `6dde1c1c1bdc84bf82c1b8266c071c44756fe7c0` (the local remote-tracking ref, not re-fetched in this run, gap G-4) the strict run exits 0 with 0 errors (EP-02). The strict run does go red on a known failing input: with `en/getting-started/faq.md` deleted from that extracted tree it exits 1 with one error, and the warn-only run on the same tree exits 0 (EP-03), which is exactly what the advisory pull-request path does today. The pull-request path writes `strict=false` (lines 74 and 77, EG-05) and the `continue-on-error` expression at line 85 evaluates to true for it (EG-06); a simulation of both gives `event=pull_request strict=false continue-on-error=true` (EG-08). Blocking pull requests changes what every contributor waits on whenever `docs-site/content/**` changes. The card's scope text asks to fix those lines; that text is the work order, recorded here as context, and it is not an operator verdict on this row. No Default is ranked: the published rule would keep the advisory behavior, and applying it would silently decide against the work order, so the row is left to the operator.
Class: implementation-level
Operator verdict:

### Q10: Which Hugo version does the new CI build job use?
Label: FOUNDER
Authority anchor: `docs-site/vercel.json` `HUGO_VERSION` (committed; EP-04). It is a committed document, not a register authority, so the label stays FOUNDER.
Why unresolved: a job that builds with another version would not test the build that ships. The deployed pin is `0.160.1` (`docs-site/vercel.json` line 11) and the local binary is `hugo v0.160.1+extended+withdeploy` (EP-04). Release v0.160.1 exists upstream and its Linux amd64 extended archive answers HTTP 200 with a published SHA-256 (EP-04), so the job can install exactly that version and verify the archive. The job reads the version from `docs-site/vercel.json` rather than repeating the literal, so the default holds by construction and a future pin change cannot drift from the job. Not measured: the Hugo edition Vercel installs (extended or not), and the Vercel install step `hugo mod get -u` (gap G-5).
Class: implementation-level
Default: `0.160.1`, read from `docs-site/vercel.json` (rule: preserves current behavior; the deployed build uses this version, and the local binary matches it)
Alternate: a newer Hugo version chosen by the operator, which also requires the same change in `docs-site/vercel.json`, outside this SPEC's paths
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:30:04Z card-t1717/manager-spec (plan author)

### Q11: Which link-check tool runs in the new job?
Label: EVIDENCE-NEEDED
Authority anchor: n/a
Why unresolved: no link checker is referenced in any workflow (EG-13), and none of the four probed tools was installed locally before this pass (EP-09). The parent asked for each candidate to be run on the Hugo output with one injected broken internal link observed red and the tree observed green. Measured here, on the site built from this tree:
- htmltest v0.17.0 (installable with `go install github.com/wjdp/htmltest@v0.17.0`, built into a scratch directory, EP-06): with its default configuration it reports 73,692 findings on the baseline, so it is red on arrival; with a configuration that suppresses the trailing-slash and empty-alt classes and ignores three documented baseline patterns it exits 0 on the baseline (657 documents) and exits 1, naming the target, on one injected broken link. It needs a configuration file, hence the conditional file `docs-site/.htmltest.yml`.
- linkinator 8.1.0 (installable with `npm install`, EP-08): started from the Korean home page it scanned 46 links, found `/changelog/` returning 404 and exited 1; its recursion did not reach the full page set, so the measurement is partial and no baseline-green or injected-red result is recorded for it.
- lychee: not run. No Rust toolchain is installed and the release binary was not downloaded or executed (gap G-6).
- markdown-link-check: not run. It reads Markdown source; the requirement is about the built site (gap G-6).
The needed evidence that remains: lychee run to the same standard, and linkinator's full-set result, if the operator wants them compared.
Class: implementation-level
Operator verdict:

### Q18: Is each `--factory-run` and `--clear-policy` citation kept, with the flag added to help, or removed?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: decided by the FLAG child (SPEC-V3R6-DOCS-PREREWRITE-FLAG-001, card t1726); this row is the input GATE reads. The cited-verbs check does not count flags (C-6), so no output of the checker changes with the answer. Neither flag is named by any workflow or script (EP-13). If FLAG's answer adds a flag clause to the check, extending the check is outside this SPEC and needs an amendment. The stub verdict line is empty here by design: an answer written in this index is not a verdict for FLAG.
Class: product-level
Operator verdict:

### Q22: Is REQ-PRC-015 moved to an operator-held section, or is the branch-protection exclusion amended?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: REQ-PRC-015 (parent) requires the three jobs of REQ-PRC-012 to REQ-PRC-014 to be required status checks on `main` (external), while the parent's Out of Scope section excludes changing the branch protection of `main`. The requirement and the exclusion conflict, and neither is decided here. Q22 has owner GATE (leader-confirmed, recorded for card:t1655 on 2026-10-14; operator decision `d-20261014T114834Z-e039` requires the answer to be decided inside the owning child). The options: (a) move the closure out of the requirements into an operator-held section, keeping the exclusion; (b) amend the exclusion to name this change as an operator-held action. This SPEC contains no step that depends on the answer: its own Out of Scope keeps the branch-protection exclusion as written (spec.md), and the job names it publishes are fixed in spec.md § Definitions so that either option can be applied afterwards. The five required contexts of `main` do not include the three jobs; that observation is external and is recorded only in the parent ledger (E-21).
Class: implementation-level (an external setting; keep-set)
Operator verdict:

### Q25: Do the three new jobs report a status on every pull request, or only when their paths match?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the existing parity workflow is path-filtered (`docs-i18n-check.yml` lines 22 to 26: `docs-site/content/**`, the script, the workflow file). The repository has a second convention for checks that are required: `ci.yml` always runs and a stub job satisfies the required Test check when a path filter says no Go inputs changed (EP-14). A required check that a path filter skips can leave an unrelated pull request waiting (parent plan §D, which assigns the fix to this child and says it is observed before REQ-PRC-015 closes). The choice has no effect until the jobs become required, which is the parent's operator-held step. The two shapes: (a) path-filtered triggers on each workflow, like the existing parity workflow; (b) triggers on every pull request, with the work steps skipped by an in-job path filter and a status always reported, like `ci.yml`. No Default is ranked: the first rule step has nothing to preserve (the jobs do not exist), the second gives no order (both are undone by one revert), and the third would rank (a), but applying it would leave unaddressed a risk the parent assigns to this child, so the row is left to the operator.
Class: implementation-level
Operator verdict:

### Q26: How does the link-check job treat the links that are already broken on the baseline?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: with the noise classes suppressed, 668 targets are unresolved on the built site from this tree (EP-06, EP-07): 636 are the footer link `/changelog/` (`docs-site/layouts/partials/site-footer.html` line 9), which resolves only through the `/changelog` redirect at `docs-site/vercel.json` line 118; 19 are relative links on the `getting-started/cli` pages (`./update` 6, `./status` 5, `./profile` 5, `./inventory` 3) that, read from a URL without a trailing slash, point at `/<locale>/getting-started/<name>`, which the redirects at `vercel.json` lines 28, 38, 58, and 68 serve; and 13 are relative links that no redirect serves, so by URL resolution rules they are broken for readers today (6 on the `getting-started` section index, `./update` 3 and `./profile` 3, which from `/<locale>/getting-started` resolve to `/<locale>/update` and `/<locale>/profile`; 7 on `cli-reference/*` pages, `./cli` 4, `./init-wizard` 2, `./quickstart` 1). The reading of those 32 is an inference from URL resolution and the redirect table; no live request was made (gap G-16). A checker that reads the built files cannot see redirects. The content and redirect children do not remove these: per the split's file lists no child edits `docs-site/layouts/`, so the footer link stays. Two shapes: (a) the tool configuration carries an explicit, documented ignore list for these patterns, and any other broken internal link turns the job red; (b) the job stays red on arrival until all of the above are fixed elsewhere, which cannot happen inside this SPEC. Residual risk of (a): an ignore pattern matches by the link's text, so a future link with the same text in another page is also ignored, and if REDIR removes the `/changelog` rule (its Q21 option (a)) the footer link becomes a real 404 that the entry hides (gap G-9). Anchors are not checked in the measured configuration: with anchor checking on, the tool reports 20 anchors as missing on the baseline (percent-encoded CJK ids and ASCII ids both appear, for example `/ko/advanced/security-notes/#cwe-345`), and whether they are real is not established (gap G-8).
Class: implementation-level
Default: (a) an explicit ignore list in the tool configuration (rule: the option whose undo is a single revert of this SPEC's own commits; option (b) depends on commits of other SPECs)
Alternate: (b) no ignore list, with the job red until the baseline links are fixed elsewhere
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:30:04Z card-t1717/manager-spec (plan author)
