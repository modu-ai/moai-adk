# SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 — Implementation Plan (revision 0.1.0)

Sections are ordered by decision reversibility: the decisions most likely to change come first (which public URLs are served, which rules go), and the mechanical steps come last.

## A. Open decisions that shape the plan

The verdict state of each row is in decision-index.md. The plan below holds for every recorded answer; where an answer changes the work, the plan names the branch.

| Row | State | What it changes in the work |
|-----|-------|-----------------------------|
| Q23 | product-level, empty | which of the rules at lines 163 and 168 go (M1); under option (b) the requirement is amended instead and no rule at lines 163 and 168 is removed |
| Q21 | product-level, empty | whether the rules at lines 113 and 118 go (M1), and whether the footer file changes (M1); decides Q16's coherence |
| Q15 | DEFAULT-APPLIED (remove) | whether the three `agency` rules go (M1) |
| Q16 | DEFAULT-APPLIED (delete pages), conditional on Q21 | whether the four changelog pages are deleted (M2) |
| Q17 | DEFAULT-APPLIED (same change, four locales) | the four pages go together (M2) |

## B. Milestones (priority order; no time estimates)

### M0 — Decision gate and baseline (Priority High; no file edits)

- Read the verdicts of Q23 and Q21 from decision-index.md. A `DEFAULT-APPLIED` line does not satisfy either (both are product-level). If either is empty, stop and return a blocker report to the orchestrator; do not edit.
- Under Q21 (a): confirm with the orchestrator that the footer file is in scope, or that the dead footer link is accepted in writing. The card does not name that file.
- Record BASE (the commit the run starts from) in progress.md §E.2, with `git rev-parse HEAD` and `git status --short` output.
- Extract the instrument from acceptance.md E-R0 to a scratch directory outside the repository, confirm its SHA-256, and re-run the controls of E-R6 and E-R7 to see their recorded outputs.
- Re-run E-R4 and E-R5 on the BASE tree (the RED-now cells) and record the outputs; the values may differ from the 0ddf5ca48b values if main moved.
- Record the baselines AC-REDIR-003 needs: `bash scripts/docs-i18n-check.sh` (expected `Errors: 0`, `Warnings: 0`) and the Hugo build into a scratch destination (expected exit 0, no `WARN` or `ERROR` line). Delete the `.hugo_build.lock` the build leaves in `docs-site` (it is gitignored) if the run is in a shared tree.
- Re-run E-R8 (external) and record the UTC time; it is an observation, not a pass.
- Re-measure the in-flight branch overlap for `docs-site/vercel.json`, `docs-site/content/*/changelog`, `docs-site/content/*/utility-commands/moai-mx.md`, and `docs-site/layouts/partials/site-footer.html`; run it as plain single git commands, one branch per invocation, if the worktree guard refuses a compound form.

### M1 — Redirect rule edits in `docs-site/vercel.json` (Priority High; the public URLs change here)

Order inside the edit, matching the card's precondition: the moai-mx pair first, then the changelog rules, then the `agency` rules. Each edit removes whole rule objects and keeps the file valid JSON with its 2-space indentation; the instrument exits 2 on unparsable JSON, so validity is part of AC-PRC-010 and AC-PRC-011.

- Q23 (a): remove the rule objects whose sources are at the former lines 163 and 168. Q23 (c): remove only the object at the former line 163. Q23 (b): remove neither; record the amended REQ-PRC-011 in spec.md under a new HISTORY row (owner: manager-spec, by orchestrator re-delegation).
- Q21 (a): remove the rule objects at the former lines 113 and 118; under this answer the Q16 Alternate applies (pages stay), and the footer file changes in the same commit to a target that resolves (the target is chosen by the run phase from the built output; the footer text and the four `footer_changelog` strings are not edited). Q21 (b): remove neither; record the amended REQ-PRC-010 (admitted external destination: `https://github.com/modu-ai/moai-adk/releases`) in spec.md under a new HISTORY row.
- Q15 Default: remove the rule objects at the former lines 88, 93, and 98. Q15 Alternate: change their destinations to the existing page the operator names.
- Stage by explicit pathspec (`git add docs-site/vercel.json` and, under Q21 (a), `docs-site/layouts/partials/site-footer.html`); never `git add -A` or `git add .`.

### M2 — Delete the four changelog index pages (Priority Medium; reversible by reverting one commit)

- Starts only after the Q21 verdict is read at M0, and only under the Q16 Default with Q21 not (a).
- Delete `docs-site/content/{ko,en,ja,zh}/changelog/_index.md` in one commit (all four or none; REQ-REDIR-003). Do not touch the four `moai-mx.md` pages.
- After the commit, the `changelog` directories are empty and disappear from the tree; the section URLs then rest on the redirect rules alone.

### M3 — Closing measurement (Priority High)

- Run the done checks of AC-PRC-011, AC-PRC-010, AC-REDIR-001, AC-REDIR-002, and AC-REDIR-003 against the closing commit, as separate read-only invocations in one parallel batch where they are independent. Write each command, its verbatim stdout, its exit code, and the commit SHA into progress.md §E.2 (manager-develop's section).
- Re-run the instrument controls (E-R6, E-R7) once more with the extracted script, so the verdict is read from a verified instrument.
- Regression-guard criteria are recorded as observations and not counted as passes.

## C. Technical approach and constraints

- The change is data editing in a JSON file and four file deletions; no code, no template, no workflow. `docs-site` is not mirrored into `internal/template/templates/` (E-R13), so there is no Template-First step.
- The checks run with Python 3 (`python3 -I`), `git`, `bash`, and `hugo`; none needs a Go toolchain (`hugo.toml` declares no module, E-R13). The resource rule for the card's lane bars running `go test`, `go build`, `go vet`, lint, and `make`; none of them is part of this plan.
- Locale rules: the changelog deletions land in all four locales in one commit (Q17 Default); no prose is translated; ko is the canonical locale and the derivation chain ko -> en -> ja/zh is satisfied by deleting all four at once (`hns-oss-docs-i18n-rules` §1 and §2).
- Page removals carry no redirect obligation of their own: `hns-oss-docs-i18n-rules` §8 requires a redirect when a page is moved or renamed; under the Q16 Default the changelog URL stays served by the kept redirect rules.
- Commits: Conventional Commits with the card id `t1718` in the message, staged by explicit pathspec, one commit per milestone that edits files.

## D. Green path per acceptance criterion

- AC-PRC-011: the moai-mx lines flip at M1; the changelog lines flip at M2 (Q16 Default) or at M1 (Q16 Alternate).
- AC-PRC-010: the nine `agency` lines flip at M1 (Q15 Default); the five external lines flip at M1 under Q21 (a) and stay as the expected end state under Q21 (b).
- AC-REDIR-001, -002, -003: regression-guards; they are observed at M0 (baselines) and M3.

## E. Risks

- **Public URL change.** Removing a rule changes what a public URL serves. Mitigation: Q21 and Q23 are product-level and block M1 until the operator answers; Q15 and Q16 rest on live observations (E-R8) of a deployment whose commit is unknown (G1), recorded as a residual risk.
- **Q21 and Q16 coupling.** Removing the changelog rules and the changelog pages together leaves the public changelog URL serving nothing and the footer link dead. REQ-REDIR-001 forbids it; M2 does not start before the Q21 verdict is read.
- **Footer link.** The root-level rule at line 118 is the only thing that keeps the footer's `/changelog/` link alive (E-R9a). Under Q21 (a) the footer file changes, and a link-check job (GATE, REQ-PRC-014) would otherwise report it.
- **Vercel precedence is inferred.** The claim that a redirect hides the built page rests on E-R8 for one deployment (G1, G8).
- **Instrument blind spots.** Case-variant sources, `has` or `missing` conditions, `url:` front matter, static files, and Hugo-generated URLs are not seen (G3). The mutant probes in acceptance.md § B name the variants that other commands catch.
- **Rebase drift.** If main moves, the line numbers (88, 93, 98, 113, 118, 163, 168) and the counts (39 redirects, 612 pages) can change. M0 re-measures them; no cell is quoted again without re-measuring.

## F. Tests named, not run

- `redirect_destinations_resolve_all_locales` (the umbrella's name for AC-PRC-010's check): realized as the instrument's `resolve` mode with the fixtures of E-R6. Not run as a repository test; this child adds no test file.
- `redirect_sources_shadow_no_page`: realized as the instrument's `shadow` mode with the fixtures of E-R6 and the scratch copies of E-R7.
- The Hugo build and the i18n script are named in AC-REDIR-003 and are not run in this planning pass (G2 for the build; the i18n script was run, E-R10).

No Go test was run in this revision, and no test files are added by this child.

## G. File overlap with other cards

- The lane's observation, relayed by the leader: none of twelve in-flight branches touches `docs-site/vercel.json` or `docs-site/content/*/changelog`. The command for each branch is `git diff --name-only 0ddf5ca48b...<branch> -- docs-site/vercel.json docs-site/content/*/changelog`, printing nothing. This author did not re-run it (acceptance.md G5).
- Within the umbrella: the child table lists REDIR with no after-relations. GATE (t1717) edits workflow files and the cited-verbs producer; no file is shared. README (t1720) and the SWEEP children edit pages and README files; the four `moai-mx.md` pages are pages the SWEEP children may edit for stale citations, and REDIR does not edit them. If a SWEEP child edits a `moai-mx.md` page, the two changes do not conflict on a line level, but AC-PRC-011 D2 and D3 still apply to the page after both land.

## H. Cross-references

- Umbrella: `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (spec.md § Child SPECs, decision-index.md Q15/Q16/Q21/Q23 stubs, acceptance.md E-15a to E-15e).
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- Verification discipline: `.claude/rules/moai/development/verification-completeness.md` §1 to §4; `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 and §3.1.
- Docs i18n rules: `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md`, `.claude/skills/hns-oss-docs-verify/SKILL.md`.
