# SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 — Implementation Plan (umbrella, revision 0.2.0)

## A. Pin, provenance, and method

### A.1 Measurement pin

- Tree commit: `f35103480f03ad5a8418652bb03223dd67efa3dd`. Tree object: `ca8fc594f5f23c4691cf41c876a4b12191cab47d`. Both were re-read with `git rev-parse` in revision 0.1.1.
- The document-level pin binds every tree-attributed cell in acceptance.md §C, per verification-completeness §2.1. Each tree-reading command also names the commit (`git grep … f35103480f`), so the pin is in the cell.
- The SPEC directory is untracked. The tracked tree was clean at every measurement (`git status --short` shows only that directory).
- External observations (the GitHub branch protection API and the public release list) are not tree properties. They carry the observation window in acceptance.md §A and are labelled `external`.

### A.2 Tool provenance (verification-claim-integrity §2.2)

- The verb-surface probes (E-06, E-07) use a binary built at f35103480f. It was built with the `go build` recipe line that `make -n build` prints for the `build` target, so the target's LDFLAGS stamp the commit. Its version output carries the BuildID `moai_cp/20261007_195927-847-gf35103480f`, which contains the tree commit.
- Deviation, stated in the record: the `build` target's prerequisites (three read-only drift checks, `templ generate`, and `gen-catalog-hashes --all`) were not run, and the output was written to the session scratchpad rather than `bin/moai`. The prerequisites write generated files and run targeted test suites, and this SPEC's write scope is its own directory. The tree stayed clean, as §A.1 states.
- The installed `moai` (v3.2.0-rc.31, commit ec193e4777, a strict ancestor of f35103480f) is not cited anywhere in this SPEC.
- Hugo `v0.160.1+extended+withdeploy` is installed locally and matches `HUGO_VERSION` `0.160.1` in `docs-site/vercel.json`. The Vercel install step `hugo mod get -u` was not run, so the local build is not a full CI replica (gap G3).

### A.3 Method limits

- Each evidence cell in acceptance.md §C is one read-only invocation: no `;`, `&&`, `|`, subshell, redirection, or `echo $?` wrapper. Its exit code is a separate field. Where the Bash tool does not display a zero-match exit status, the status was confirmed with a measurement wrapper outside the cell, and the ledger says so.
- Exploration pipelines (`sort | uniq`, `awk` sums) produced candidate lists and derived totals. A derived total is labelled `derived` and is never the RED observation itself. No pipeline output is cited as evidence.
- The stale-reference families are a sweep, not the producer. The producer (M1) must reproduce the family counts on the same tree before the RED cells can be called a finding of the producer.
- Prose noise is excluded by scope: tokens such as "moai is", "moai command", "moai available", "moai folder", "moai managed", "moai entry", "moai sessions", "moai runs", "instance already", and "plugin carries" match a broad regex but are not commands. The producer scans code contexts only (inline code and fenced blocks) plus the slash surface, and derives the CLI verb set from the tree build, not from a hard-coded list.

## B. Child map and closing measurement (umbrella, revision 0.2.0)

Revision 0.2.0 splits the work into eleven child SPECs (decision Q20 = (c), `d-20261014T055241Z-1bc6`). The former milestones M0 to M5 are owned by the children in the table below, and the children land independently, in any order. M6 stays in this umbrella as the closing measurement. A child starts only when each decision row it depends on carries an operator verdict, attributed to the operator with a UTC timestamp. A `Planner default applied:` value is not a verdict.

### B.1 Child map (former milestones)

| Former milestone | Owning child (card) | Decision rows | File count |
|------------------|---------------------|---------------|------------|
| M0 (decision gate) | each child, for the rows it depends on (its `Child:` line in decision-index.md) | the decision rows of every other row in this table | — |
| M1 (producer, inventory) | GATE (t1717) | Q5 (input), Q18 (input) | 9 (shared with M2) |
| M2 (CI wiring) | GATE (t1717) | Q9, Q10, Q11; Q12 and Q22 (parent, operator-held) if the cited-verbs step becomes a required context | 9 |
| M3 (redirects and shadowing) | REDIR (t1718) | Q15, Q16, Q21, Q23 | 5 |
| M4 (content: stale references) | README (t1720) for the four README files; SWEEPA (t1721), SWEEPB (t1722), SWEEPC (t1723), SWEEPD (t1724), SWEEPE (t1725), and SWEEPF (t1727) for the page files; FLAG (t1726) for the flag clause | Q13, Q17 (all); Q14 (SWEEPF); Q24 (SWEEPE); Q18 (FLAG) | README 13; SWEEPA 12; SWEEPB 12; SWEEPC 12; SWEEPD 12; SWEEPE 10; SWEEPF 12; FLAG 8 |
| M4 (CG passage and badge) | README (t1720) | Q1, Q8, Q17 | 13 (the same README child files) |
| M5 (residue and assets) | RESID (t1719) | Q2, Q3, Q4, Q6, Q7, Q19 | 14 |

- While Q17 keeps the four-locale scope, the four locale copies of a page sit in the same child. If Q17 answers no, the grouping can be split by locale.

### B.2 M6 — Umbrella closing measurement (no new content; the same set as the DoD)

- After every child has landed, re-run the release-blocking set listed in acceptance §D on the closing tree. This is the same set the DoD names. Record the closing commit SHA, the verbatim stdout, and the exit codes in the §C form.
- The GATE child's citation checker must print `cited-verbs: <N>` as its final line, and exit 0 when N is 0. Its contract: one line per unresolved citation (`<path>:<line>:<cited form>`); exit 1 when N is above 0; exit 2 when an input cannot be read, so an unreadable input never produces a silent 0. The closing record quotes that final line and its exit code.
- Regression-guard cells (AC-PRC-003, AC-PRC-015) are re-run as observations. Their external results are recorded with UTC times and are not counted as passes.
- The completion verdict is complete only when: every §D gating criterion passes at its gate; the cited-verbs metric is 0 with exit 0; the three CI jobs exist and are blocking in the workflow files; and REQ-PRC-015 is observed on the external branch protection API (operator-held, Q22).

## C. Green path per acceptance criterion

Each AC names its former milestone in acceptance.md §B and its owner child in acceptance.md §A.1. The short map:

- AC-PRC-001 (stale references): owner README (t1720) and SWEEPA to SWEEPF (t1721 to t1725, t1727), through partial criteria; the parent closes it at M6. Former milestone M4 with Q13 and Q14; the M1 producer (GATE, t1717) is the independent check.
- AC-PRC-002 (producer): owner GATE (t1717). M1 (red on fixtures), M2 (red on a pull request run), M6 (exit 0 on the closing tree).
- AC-PRC-003 (badge): owner README (t1720). Regression-guard. M4 with Q8; the external observation is recorded, not passed.
- AC-PRC-004 (CG passage): owner README (t1720). M4 with Q1.
- AC-PRC-005 (unreferenced images): owner RESID (t1719). M5 with Q2.
- AC-PRC-006 (OG budget): owner RESID (t1719). M5 with Q3. The number is recorded only after the Q3 verdict.
- AC-PRC-007 (Nextra residue): owner RESID (t1719). M5 with Q4. The translate and migrate scripts are not in this criterion (Q5 is a GATE input to M1).
- AC-PRC-008 (old README): owner RESID (t1719). M5 with Q6.
- AC-PRC-009 (llms.txt): owner RESID (t1719). M5 with Q7 (operator-held).
- AC-PRC-010 (redirect destinations): owner REDIR (t1718). M3 with Q15 and Q21.
- AC-PRC-011 (redirect shadowing): owner REDIR (t1718). M3 with Q16, extended to the moai-mx pair under Q23.
- AC-PRC-012 (i18n job blocking): owner GATE (t1717). M2 with Q9. Closes only after the pull-request red is observed. The GATE child corrects the RED cell (audit defect B01); that correction is not made in this revision.
- AC-PRC-013 (Hugo job): owner GATE (t1717). M2 with Q10.
- AC-PRC-014 (link-check job): owner GATE (t1717). M2 with Q11.
- AC-PRC-015 (required checks): owner parent, operator-held (Q12, Q22). Regression-guard.

## D. Risks

- **Public URL change.** Removing llms.txt, the redirect rules, or the changelog pages changes served URLs. Mitigation: Q7, Q15, Q16, and Q21 are product-level or reopened, and they block until the operator decides. The planner's defaults for Q15 and Q16 rest on an inference from the content tree, because no live request was measured (gap G5).
- **Planner defaults are provisional.** A milestone started on a planner default without the operator's verdict is rework. M0 prevents it.
- **Required checks with path filters.** The i18n workflow runs only when `docs-site/content/**` or `scripts/docs-i18n-check.sh` changes. The cited-verbs step needs an equivalent path filter. A required check that a path filter skips can leave an unrelated pull request waiting. The fix belongs in M2 and is observed before REQ-PRC-015 is closed.
- **Job-name drift.** A required-check context must match the job name exactly. The contexts named in E-21 are the current list. Renaming a job without updating the protection breaks the merge gate.
- **Producer false positives.** The broad regex matches prose (§A.3). Producer scope is code contexts only. Without that, the cited-verbs figure would become a prose count.
- **CI build divergence.** The local Hugo build did not run `hugo mod get -u`. A CI job may fail where the local build passes. M2's first run on a pull request is the observation.
- **Stage-0 overlap is not yet checked.** The path set in spec.md §Scope may overlap a stage-0 card. The check runs at dispatch: list each stage-0 card's changed paths and confirm an empty intersection with the path set in spec.md.

## E. Test named, not run

- `docs_cited_verbs_test` (M1): three fixture cases, described in M1.
- `hugo_build_gate_red_on_broken_shortcode` and `link_check_red_on_broken_internal_link` (M2): described in M2.
- `redirect_destinations_resolve_all_locales` (M3): for every redirect destination in `docs-site/vercel.json`, a fixture check that the destination maps to a content file in each locale named by the rule. Not run.

No test was run in this revision. The full suite is not run in this SPEC.

## F. Open questions

The decision rows in decision-index.md carry the questions (Q1 to Q24). Two points are not decisions and are listed here:

- Body language. `language.yaml` sets `documentation: ko`. Recent docs SPECs mix English structure with Korean narrative, and this SPEC follows that mix: English structure and GEARS text, Korean card labels where they help. The operator may ask for a different convention.
- The page measure for "CG 문단 8쪽→1쪽". The repository defines no page count. This SPEC uses the count of pages that carry the exact retirement sentence (9 files, 8 outside the canonical page) as the proxy. The proxy is labelled as a proxy in acceptance.md (G7).

## G. Observations deferred to other cards or rows

- `docs-site/content/en/core-concepts/ddd.md` contains the word "nextra" once. It is content, not a script, and is not classified here (G10).
- E-04 (slash family) also counts narrative and error examples: `docs-site/content/en/getting-started/faq.md` line 24 (`/moai init` as a misconception), `docs-site/content/en/core-concepts/verification-claim-integrity.md` line 84 (a retired `/moai brain` narrative), and the invocation examples at `docs-site/content/en/utility-commands/moai-feedback.md` lines 32–34 (unregistered `/moai fb`, `/moai bug`, `/moai issue`). The first two are governed by Q24 (N2, operator verdict `d-20261014T060405Z-0fcd`): only narrative examples that carry an explicit example marker are exempt, so the SWEEPE child checks each of the two lines for a marker when it starts, and lines without a marker are fixed. The third is a stale citation.
- The `cc` help does not name `--factory-run` or `--clear-policy` (E-25a). This closes gap G6 for `cc`.
