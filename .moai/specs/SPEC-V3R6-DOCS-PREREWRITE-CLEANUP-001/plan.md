# SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 — Implementation Plan (revision 0.1.1)

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

## B. Milestones (ordered by decision reversibility)

Decisions that change public or shared state come first, because a later decision can reverse them only at a higher cost. Mechanical edits come last. Every decision row in decision-index.md is open (its `Operator verdict:` is empty), so no milestone starts until the rows it depends on carry verdicts.

### M0 — Decision gate (prerequisite for M1 through M6)

- Each open row that a milestone depends on must carry a non-empty `Operator verdict:`, attributed to the operator (not to the planner or a runner), with a UTC timestamp. A `Planner default applied:` value is not a verdict.
- The gate table. Each milestone names the rows it depends on:

| Milestone | Rows that must carry operator verdicts before the milestone starts |
|-----------|-------------------------------------------------------------------|
| M1 (producer, inventory) | Q5 (translate and migrate classification); Q18 (whether a flag clause is counted) |
| M2 (CI wiring) | Q9 (PR blocking), Q10 (Hugo version), Q11 (link-check tool) |
| M3 (redirects) | Q15 (`/agency` rules), Q16 (changelog pages), Q21 (external changelog rules) |
| M4 (content) | Q1 (CG page), Q8 (badge; REQ-PRC-003), Q13 (stale citations), Q14 (moai-design page), Q17 (four-locale scope) |
| M5 (residue and assets) | Q2 (images), Q3 (OG budget), Q4 (Nextra scripts), Q6 (docs-site README), Q7 (llms.txt), Q19 (maintainer path) |
| M6 (closing) | Q12 and Q22 (operator-held external change); every row above, re-measured at closing |
| The Kickoff of this SPEC (all milestones) | Q20 (tier) |

- Green path for M0: every row in the table carries an operator verdict. M1 and M3 are gated behind M0 explicitly; M2, M4, and M5 are gated behind M0 for their listed rows.

### M1 — Producer and inventory (no content change yet)

- Build the docs cited-verbs producer under `scripts/` or as a workflow step. Its metric and format are fixed here (verdict defect D14):
  - Scanned paths: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, and `docs-site/content/**/*.md`.
  - Scanned contexts: inline code spans and fenced code blocks only. Prose outside code is not scanned (§A.3).
  - Resolution inputs, read from the tree at run time: (i) the verb and sub-verb set, from the COMMANDS block of `moai --help` and of each top-level verb's `--help` in the tree build. A sub-verb resolves only when the COMMANDS block lists it, because an absent sub-verb prints its parent's help and exits 0 (observation O8; E-07b). (ii) The slash set, from the file names under `.claude/commands/moai/`. (iii) The flag set, only if decision Q18 adds a flag clause; otherwise flags are not counted.
  - Exemptions, explicit in the producer's configuration, each with its reason: the four `multi-llm/cg-mode.md` pages (the retirement notice, REQ-PRC-001) and the four `utility-commands/moai-todo.md` pages (natural-language input and documented error examples). Any further exemption, such as the narrative examples under E-04 (N2), needs its own decision and reason.
  - Metric: `N`, the number of unresolved citations, counted once per (path, line, cited form).
  - Output (stdout, fixed format): one line per unresolved citation, `<path>:<line>:<cited form>`, then one final line `cited-verbs: <N>`.
  - Exit status: 0 when `N` is 0; 1 when `N` is above 0; 2 when a producer input cannot be read. Exit 2 is fail-closed: an unreadable input never produces a silent 0.
- Observe it red on a fixture that cites `moai codex job` and `/moai design`, before it is adopted (verification-completeness §1.1). Observe it exit 0 on the cleaned tree at the M6 closing run.
- Classify `docs-site/scripts/translate.mjs` and `docs-site/scripts/migrate-frontmatter.sh` (Q5). Their callers outside their own folders are none (E-12).
- Reproduce the family counts on the tree as an independent check of the sweep: E-01 (60 CLI lines in 48 files, notice excluded), E-02 (24 lines in 4 files), E-03 (8 lines), and E-04 (45 lines in 23 files, slash family).
- Named test, not run: a shell fixture test `docs_cited_verbs_test` with three cases. (a) A fixture README citing `moai codex job` must fail. (b) A fixture citing `moai spec status` must pass. (c) A fixture citing `moai run SPEC-ID` inside a Mermaid label must fail. Each case records its exit code.

### M2 — CI wiring (blocking gates)

- i18n job (REQ-PRC-012), the exact edit sites (verdict defect D4). In `.github/workflows/docs-i18n-check.yml`:
  - (a) The pull-request branch of the "Determine strictness" step (lines 76–77): change `echo "strict=false"` on line 77 to `echo "strict=true"`, and update the comment on line 76. The push branch (line 74) and the workflow_dispatch branch (lines 68–69) stay unchanged. With this edit the run step at line 89 executes `DOCS_I18N_STRICT=1` for pull requests.
  - (b) Line 85: change `continue-on-error: ${{ steps.mode.outputs.strict != 'true' }}` to `continue-on-error: ${{ github.event_name != 'pull_request' && steps.mode.outputs.strict != 'true' }}`. For pull_request events the expression is false by construction. For push and workflow_dispatch its value is unchanged.
  - (c) Update the header comment (lines 3–16, "ADVISORY ONLY — NOT BLOCKING") in the same change.
  - (d) Observe: a pull request with an injected drift (for example a ko page without its en counterpart) turns the job red, and the closing-tree run is green. AC-PRC-012 closes only after the red is observed on a pull request run.
  - This revision (0.1.1) does not edit the workflow file. The edit is performed at M2 dispatch.
- Cited-verbs step (REQ-PRC-002): add a pull-request step (or job) that runs the M1 producer. Its path filter matches the scanned paths (README files and `docs-site/content/**`). Observe: an injected stale token on a pull request fails the run. The step becomes a required status context only through Q12 and Q22.
- Hugo job (REQ-PRC-013): run `hugo --minify --gc` in `docs-site` with the version that Q10 records (planner default: the `HUGO_VERSION` in `docs-site/vercel.json`). Observe it red on an injected template error (for example an unclosed shortcode), then green on the tree.
- Link-check job (REQ-PRC-014): add the job once Q11 carries a verdict. Observe it red on an injected broken internal link, then green on the tree.
- Named tests, not run: `hugo_build_gate_red_on_broken_shortcode` (a fixture page with `{{< unclosed` must fail the Hugo job) and `link_check_red_on_broken_internal_link` (a fixture link to `/ko/does-not-exist/` must fail).

### M3 — Redirects and shadowing (user-visible URLs; gated behind M0 on Q15, Q16, Q21)

- Remove the three `/agency` redirect rules per Q15 (`docs-site/vercel.json` lines 87–101; sources at lines 88, 93, and 98; destinations at lines 89, 94, and 99). Planner default: remove.
- Remove the four changelog index pages (`docs-site/content/{ko,en,ja,zh}/changelog/_index.md`) per Q16, only after the redirect sources at lines 113 and 118 are confirmed to remain. Under Q16's alternate, remove the redirect rules instead. Planner default: remove the pages.
- The two external changelog rules (lines 113–121; destination `https://github.com/modu-ai/moai-adk/releases`) remain until Q21 has a verdict. Under Q21(a) they are removed. Under Q21(b) REQ-PRC-010 is amended by the operator's decision, not by this SPEC's revision, and the closing check admits the approved external destinations.
- Check each remaining destination in all four locales at closing, one invocation per destination path. The redirect object count is 39 (derived from the pinned blob with `python3 -I`: redirects 39, rewrites 1, headers 2).

### M4 — Content: stale references, CG passage, badges (gated behind M0 on Q1, Q8, Q13, Q14, Q17)

- Stale references (REQ-PRC-001): apply Q13 per form (planner default: delete the stale citation). Where Q13 selects replacement, the Mermaid labels in `worktree/_index.md` and `worktree/faq.md` (E-03) become the registered slash forms `/moai run SPEC-ID` and `/moai sync SPEC-ID`. Under the default they are removed from those diagrams instead.
- The moai-design page: apply Q14. The moai-codemaps page is not stale under REQ-PRC-001: its citations are the registered slash command (N4).
- CG passage (REQ-PRC-004): keep the page that Q1 names (planner default: `docs-site/content/ko/multi-llm/cg-mode.md`) as the single copy. On each of the eight other Korean pages that carry the sentence (E-08), replace the retirement sentence with one link sentence to that page. Seven of the eight already link to the CG page (E-08b); `what-is-moai-adk.md` needs a new link.
- Badge (REQ-PRC-003): apply the Q8 verdict to the four README badges and the two `hugo.toml` lines in one commit.
- Locale parity (Q17): apply each edit to en, ja, and zh in the same change only if the Q17 verdict says so. Otherwise record the locale scope in the closing measurement.

### M5 — Residue and assets (mechanical; gated behind M0 on Q2, Q3, Q4, Q6, Q7, Q19)

- Images (REQ-PRC-005; Q2). Under deletion, remove the five files named in E-09a. Verify with `git ls-tree` (zero entries) and a reference grep (zero matches, E-09b). Under the reference alternative, add the reference and remove nothing.
- OG image (REQ-PRC-006; Q3). Re-encode to the Q3 budget, and measure with `git cat-file -s`. The four README references (E-10) still resolve. No budget is set in this SPEC.
- Nextra scripts (REQ-PRC-007; Q4). Under deletion, delete `docs-site/scripts/fix-mdx-formatting.js` and `scripts/convert-nextra-to-hextra/main.go` after the M1 caller check is repeated on the working tree. Correct the negative existence statements about `docs-i18n-check.sh`: `.claude/skills/hns-oss-docs-verify/SKILL.md` lines 9 and 23, and `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md` lines 115–116 (the sentence is split across the two lines). Correct the Hextra heading and sentence at `.moai/docs/docs-site-i18n-rules.md` lines 35 and 37. The done checks are in acceptance §B, AC-PRC-007.
- docs-site README (REQ-PRC-008; Q6). Delete, or replace with a README that describes the current docs-site.
- llms.txt (REQ-PRC-009; Q7; product-level, operator-held). Regenerate (which needs the generator that Q7's verdict names) or remove. In either branch, remove the EARS line (line 8; E-14b).
- Maintainer path (Q19). Remove the absolute path on `docs-site/scripts/README_TRANSLATE.md` line 9 per Q19. Its edit has no REQ or AC yet (O7, open).

### M6 — Closing measurement (no new content; the same set as the DoD)

- Re-run the release-blocking set listed in acceptance §D on the closing tree. This is the same set the DoD names. Record the closing commit SHA, the verbatim stdout, and the exit codes in the §C form.
- Regression-guard cells (AC-PRC-003, AC-PRC-015) are re-run as observations. Their external results are recorded with UTC times and are not counted as passes.
- The completion verdict is complete only when: every §D gating criterion passes at its gate; the cited-verbs metric is 0 with exit 0; the three CI jobs exist and are blocking in the workflow files; and REQ-PRC-015 is observed on the external branch protection API (operator-held, Q22).

## C. Green path per acceptance criterion

Each AC names its milestone in acceptance.md §B. The short map:

- AC-PRC-001 (stale references): M4 with Q13 and Q14; the M1 producer is the independent check.
- AC-PRC-002 (producer): M1 (red on fixtures), M2 (red on a pull request run), M6 (exit 0 on the closing tree).
- AC-PRC-003 (badge): regression-guard. M4 with Q8; the external observation is recorded, not passed.
- AC-PRC-004 (CG passage): M4 with Q1.
- AC-PRC-005 (unreferenced images): M5 with Q2.
- AC-PRC-006 (OG budget): M5 with Q3. The number is recorded only after the Q3 verdict.
- AC-PRC-007 (Nextra residue): M5 with Q4. The translate and migrate scripts are not in this criterion (Q5 is an M1 input).
- AC-PRC-008 (old README): M5 with Q6.
- AC-PRC-009 (llms.txt): M5 with Q7 (operator-held).
- AC-PRC-010 (redirect destinations): M3 with Q15 and Q21.
- AC-PRC-011 (redirect shadowing): M3 with Q16.
- AC-PRC-012 (i18n job blocking): M2 with Q9. Closes only after the pull-request red is observed.
- AC-PRC-013 (Hugo job): M2 with Q10.
- AC-PRC-014 (link-check job): M2 with Q11.
- AC-PRC-015 (required checks): regression-guard. Operator-held (Q12, Q22).

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

The decision rows in decision-index.md carry the questions (Q1 to Q22). Two points are not decisions and are listed here:

- Body language. `language.yaml` sets `documentation: ko`. Recent docs SPECs mix English structure with Korean narrative, and this SPEC follows that mix: English structure and GEARS text, Korean card labels where they help. The operator may ask for a different convention.
- The page measure for "CG 문단 8쪽→1쪽". The repository defines no page count. This SPEC uses the count of pages that carry the exact retirement sentence (9 files, 8 outside the canonical page) as the proxy. The proxy is labelled as a proxy in acceptance.md (G7).

## G. Observations deferred to other cards or rows

- `docs-site/content/en/core-concepts/ddd.md` contains the word "nextra" once. It is content, not a script, and is not classified here (G10).
- E-04 (slash family) also counts narrative and error examples: `docs-site/content/en/getting-started/faq.md` line 24 (`/moai init` as a misconception), `docs-site/content/en/core-concepts/verification-claim-integrity.md` line 84 (a retired `/moai brain` narrative), and the invocation examples at `docs-site/content/en/utility-commands/moai-feedback.md` lines 32–34 (unregistered `/moai fb`, `/moai bug`, `/moai issue`). The first two need an exemption decision before the producer counts them (N2, open). The third is a stale citation.
- The `cc` help does not name `--factory-run` or `--clear-policy` (E-25a). This closes gap G6 for `cc`.
