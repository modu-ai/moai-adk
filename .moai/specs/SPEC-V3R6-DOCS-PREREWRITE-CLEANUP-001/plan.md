# SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 — Implementation Plan (index, revision 0.3.0)

## A. Pin, provenance, and method

Revisions 0.1.0 to 0.2.1 measured at f35103480f. Revision 0.2.2 re-measured the RED-now cells at 8816de4dae (the card branch after absorbing main 763193861b) and recorded their outputs verbatim. Revision 0.3.0 adds one cell, E-16e, at 8816de4dae, and changes no other cell value. The supporting, verification, and protocol cells stay at f35103480f and are labelled; the drift table in acceptance section A lists the four items that differ at 8816de4dae. The whole-tree cells E-09b2 and E-28 exclude this SPEC's own directory by definition (leader ruling d-20261014T072226Z-9b67).

### A.1 Measurement pin

- Tree commit: `f35103480f03ad5a8418652bb03223dd67efa3dd`. Tree object: `ca8fc594f5f23c4691cf41c876a4b12191cab47d`. Both were re-read with `git rev-parse` in revision 0.1.1.
- The document-level pin binds every tree-attributed cell in acceptance.md §C, per verification-completeness §2.1. Each tree-reading command also names the commit (`git grep … f35103480f`), so the pin is in the cell.
- The SPEC directory is untracked (as at f35103480f, revision 0.1.x). It has been tracked since d2932a0555. The tracked tree was clean at every measurement (`git status --short` shows only that directory).
- External observations (the GitHub branch protection API and the public release list) are not tree properties. They carry the observation window in acceptance.md §A and are labelled `external`.

### A.2 Tool provenance (verification-claim-integrity §2.2)

- The verb-surface probes (E-06, E-07) use a binary built at f35103480f. It was built with the `go build` recipe line that `make -n build` prints for the `build` target, so the target's LDFLAGS stamp the commit. Its version output carries the BuildID `moai_cp/20261007_195927-847-gf35103480f`, which contains the tree commit.
- Deviation, stated in the record: the `build` target's prerequisites (three read-only drift checks, `templ generate`, and `gen-catalog-hashes --all`) were not run, and the output was written to the session scratchpad rather than `bin/moai`. The prerequisites write generated files and run targeted test suites, and this SPEC's write scope is its own directory. The tree stayed clean, as §A.1 states.
- The installed `moai` (v3.2.0-rc.31, commit ec193e4777, a strict ancestor of f35103480f) is not cited anywhere in this SPEC.
- Hugo `v0.160.1+extended+withdeploy` is installed locally and matches `HUGO_VERSION` `0.160.1` in `docs-site/vercel.json`. The Vercel install step `hugo mod get -u` was not run, so the local build is not a full CI replica (gap G3).

### A.3 Method limits

- Each evidence cell in acceptance.md §C is one read-only invocation: no `;`, `&&`, `|`, subshell, redirection, or `echo $?` wrapper. Its exit code is a separate field. Where the Bash tool does not display a zero-match exit status, the status was confirmed with a measurement wrapper outside the cell, and the ledger says so.
- Exploration pipelines (`sort | uniq`, `awk` sums) produced candidate lists and derived totals. A derived total is labelled `derived` and is never the RED observation itself. No pipeline output is cited as evidence.
- The stale-reference families are a sweep, not the producer. The producer, owned by the GATE child (SPEC-V3R6-DOCS-PREREWRITE-GATE-001), must reproduce the family counts on the same tree before the RED cells can be called a finding of the producer.
- Prose noise is excluded by scope: tokens such as "moai is", "moai command", "moai available", "moai folder", "moai managed", "moai entry", "moai sessions", "moai runs", "instance already", and "plugin carries" match a broad regex but are not commands. The producer scans code contexts only (inline code and fenced blocks) plus the slash surface, and derives the CLI verb set from the tree build, not from a hard-coded list.

## B. Index map and closing measurement (revision 0.3.0)

Revision 0.2.0 split the work into eleven child SPECs (decision Q20 = (c), `d-20261014T055241Z-1bc6`). Revision 0.3.0 makes this file an index (operator decision `d-20261014T114834Z-e039`, route 2). The children own the milestone content. This file keeps M6 and the import rule. A child starts only when each decision row it depends on carries an operator verdict, attributed to the operator with a UTC timestamp. A `Planner default applied:` value is not a verdict. Each child carries its own plan audit.

### B.1 Child map

The child map (REQ, AC, decision rows, file count, and after-relations) is in spec.md § Child SPECs. This file does not repeat it, so the two cannot drift apart.

### B.2 M6 — closing measurement (the one step kept here)

M6 has no content of its own. It re-runs the children's committed criteria and checks the repository verdict.

- After every child has committed its own spec.md, acceptance.md, and plan.md, and each child's plan audit has run in that child, re-run each child's committed acceptance criteria at the child's commit SHA. Record each child's SHA in `progress.md` §E.1, and the verbatim stdout and exit code of each re-run in the §C form of acceptance.md (the command, the output, the exit code, and the commit SHA).
- The GATE child's citation checker must print `cited-verbs: <N>` as its final line, and exit 0 when N is 0. Its contract: one line per unresolved citation (`<path>:<line>:<cited form>`); exit 1 when N is above 0; exit 2 when an input cannot be read, so an unreadable input never produces a silent 0. The closing record quotes that final line and its exit code.
- Regression-guard criteria (AC-PRC-003 in README and AC-PRC-015 in this index) are re-run as observations. Their external results are recorded with UTC times and are not counted as passes.
- The repository verdict is complete only when: every release-blocking criterion imported from a child passes at that child's commit SHA; the cited-verbs metric is 0 with exit 0; and the three CI jobs exist and are blocking in the workflow files. REQ-PRC-015's external observation (operator-held, Q22) is reported apart from the repository verdict, as a regression-guard, and is not counted in it. A missing external observation is recorded as a gap, not as a pass.

### B.3 Import rule (finding B02)

- This index imports child criteria by reference, never by copied text. A reference names the child SPEC path, the criterion ID, and, at M6, the child's commit SHA. The text stays in the child.
- A child can change its own criteria without an edit to this index. This index changes only when a reference changes: a criterion is added, removed, or renumbered in a child. The index edit then names the child's new criterion ID and SHA.
- A criterion that replaces a row of acceptance.md §D is named by child SPEC ID and criterion ID. A row whose child has not committed its spec.md is shown as not yet imported; it is not re-run and not counted.
- M6 fails on a reference whose criterion no longer exists in the child. It does not skip the row.

## C. Green path per acceptance criterion

Each green path is in the owning child, and acceptance.md §A.1 names the owner of each criterion. This file keeps two entries, so that no green path is restated here:

- AC-PRC-001 closes at M6, through the partial criteria of its children and the closing check in acceptance.md §B.
- AC-PRC-015 is operator-held (Q12, Q22) and is a regression-guard.

## D. Risks

- **Public URL change.** Removing llms.txt, the redirect rules, or the changelog pages changes served URLs. Mitigation: Q7, Q15, Q16, and Q21 are product-level or reopened, and the owning child blocks until the operator decides. The planner's defaults for Q15 and Q16 rest on an inference from the content tree, because no live request was measured (gap G5).
- **Planner defaults are provisional.** A child started on a planner default without the operator's verdict is rework. Each child's decision gate prevents it.
- **Required checks with path filters.** The i18n workflow runs only when `docs-site/content/**` or `scripts/docs-i18n-check.sh` changes. The cited-verbs step needs an equivalent path filter. A required check that a path filter skips can leave an unrelated pull request waiting. The fix belongs to the GATE child and is observed before REQ-PRC-015 is closed.
- **Job-name drift.** A required-check context must match the job name exactly. The contexts named in E-21 are the current list. Renaming a job without updating the protection breaks the merge gate.
- **Producer false positives.** The broad regex matches prose (§A.3). Producer scope is code contexts only. Without that, the cited-verbs figure would become a prose count.
- **CI build divergence.** The local Hugo build did not run `hugo mod get -u`. A CI job may fail where the local build passes. The GATE child's first pull-request run is the observation.
- **Stage-0 overlap is not yet checked.** The path set in spec.md §Scope may overlap a stage-0 card. The check runs at dispatch: list each stage-0 card's changed paths and confirm an empty intersection with the path set in spec.md.
- **Index drift.** A child criterion that changes without an index edit is invisible here until M6. M6 is the check (B.3).

## E. Test named, not run

- `docs_cited_verbs_test`, `hugo_build_gate_red_on_broken_shortcode`, and `link_check_red_on_broken_internal_link`: named in the GATE child, with their fixture cases. The fixture definitions of revision 0.1.1 (commit d2932a0555, plan.md: the producer and CI wiring sections) are the GATE child's to restate.
- `redirect_destinations_resolve_all_locales`: named in the REDIR child. For every redirect destination in `docs-site/vercel.json`, a fixture check that the destination maps to a content file in each locale named by the rule. The 0.1.1 definition is at commit d2932a0555, plan.md: the redirect section.

No test was run in this revision. The full suite is not run in this SPEC.

## F. Open questions

The decision rows in decision-index.md carry the questions (Q1 to Q24). A row that moves to a child is a stub in decision-index.md, and the child carries the full row. Four points are not decisions and are listed here:

- Body language. `language.yaml` sets `documentation: ko`. Recent docs SPECs mix English structure with Korean narrative, and this SPEC follows that mix: English structure and GEARS text, Korean card labels where they help. The operator may ask for a different convention.
- The page measure for "CG 문단 8쪽→1쪽". The repository defines no page count. This SPEC uses the count of pages that carry the exact retirement sentence (9 files, 8 outside the canonical page) as the proxy. The proxy is labelled as a proxy in acceptance.md (G7).
- The owner of Q22. The proposal is GATE, pending leader confirmation (decision-index.md, Q22; acceptance.md G22).
- The owner of Q7 and Q21 is RESID and REDIR, as child-cards.md assigns them. Their decisions are taken inside those children (decision `d-20261014T114834Z-e039`).

## G. Observations deferred to other cards or rows

- `docs-site/content/en/core-concepts/ddd.md` contains the word "nextra" once. It is content, not a script, and is not classified here (G10).
- E-04 (slash family) also counts narrative and error examples: `docs-site/content/en/getting-started/faq.md` line 24 (`/moai init` as a misconception), `docs-site/content/en/core-concepts/verification-claim-integrity.md` line 84 (a retired `/moai brain` narrative), and the invocation examples at `docs-site/content/en/utility-commands/moai-feedback.md` lines 32–34 (unregistered `/moai fb`, `/moai bug`, `/moai issue`). The first two are governed by Q24 (N2, operator verdict `d-20261014T060405Z-0fcd`): only narrative examples that carry an explicit example marker are exempt, so the SWEEPE child checks each of the two lines for a marker when it starts, and lines without a marker are fixed. The third is a stale citation.
- The `cc` help does not name `--factory-run` or `--clear-policy` (E-25a). This closes gap G6 for `cc`.
