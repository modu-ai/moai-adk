# SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 — Implementation Plan (revision 0.3.0)

r3: plan-audit iter1 FAIL 0.72, D1 to D4 fixed (the check script now compares locales and scopes the admitted URL to its two sources; the scope guard is a numstat; REQ-REDIR-003 maps to a criterion that fails now). r2: leader verdicts `d-20261018T170818Z-a207` applied. Sections are ordered by decision reversibility: the rule edits (public URLs) come before the page deletions, and the mechanical measurement comes last.

## A. Recorded verdicts that shape the work

| Row | State | What it means for the work |
|-----|-------|----------------------------|
| Q23 | (c), board a207 | remove only the rule at line 163 (M1); keep line 168 |
| Q21 | (b), board a207 | keep the rules at lines 113 and 118; REQ-PRC-010 amended to admit the GitHub Releases URL; the footer file does not change |
| Q16 | delete, board a207 | delete the four `changelog/_index.md` pages (M2); no longer conditional |
| Q15 | DEFAULT-APPLIED (remove) | remove the three `/agency` rules (M1) |
| Q17 | DEFAULT-APPLIED | the four pages go in one change (M2) |

Resulting change set: `docs-site/vercel.json` loses four rule objects (the former lines 88, 93, 98, 163) and four files are deleted. Expected end state: `swept: redirects=35 pages=608`. The leader's stated change set (line 163 plus the four pages) alone turns `shadow` green and leaves `resolve` at `unresolved: 14`; the three `/agency` rules (Q15 Default) are what take it to 5 admitted external lines (acceptance.md E-R14).

## B. Milestones (priority order; no time estimates)

### M0 — Baseline (Priority High; no file edits)

- Record BASE (the commit the run starts from) in progress.md §E.2, with `git rev-parse HEAD` and `git status --short`.
- Extract the check script from acceptance.md E-R0 to a scratch directory outside the repository, confirm its SHA-256, and re-run the controls of E-R6, E-R7, and E-R14 to see their recorded outputs.
- Re-run E-R4 and E-R5 on the BASE tree and record the outputs (the values can differ from the 0ddf5ca48b values if main moved).
- Record the baselines AC-REDIR-003 needs: `bash scripts/docs-i18n-check.sh` (expected `Errors: 0`, `Warnings: 0`) and the Hugo build into a scratch destination (expected exit 0, no `WARN` or `ERROR` line). Delete the gitignored `.hugo_build.lock` the build leaves in `docs-site` if the run is in a shared tree.
- Re-run E-R8 (external) and record the UTC time; it is an observation, not a pass.
- Re-measure the in-flight branch overlap for `docs-site/vercel.json` and `docs-site/content/*/changelog`, as plain single git commands (one branch per invocation) if the worktree guard refuses a compound form.

### M1 — Redirect rule edits in `docs-site/vercel.json` (Priority High; the public URLs change here)

- Remove the rule object whose source is `/:locale(ko|en|ja|zh)/utility-commands/moai-mx` (former line 163). Keep the object with source `/utility-commands/moai-mx` (former line 168).
- Keep the objects at the former lines 113 and 118 untouched.
- Remove the three objects at the former lines 88, 93, and 98 (Q15 Default).
- Record the amended REQ-PRC-010 wording (already in spec.md r2); no further spec edit is needed for Q21.
- Keep the file valid JSON with its 2-space indentation; the check script exits 2 on unparsable JSON, so validity is part of AC-PRC-010 and AC-PRC-011.
- After the edit, `git diff --numstat <BASE> -- docs-site/vercel.json` must read `0 20` (AC-REDIR-002 D2); any other figure means more or less than the four named blocks changed.
- Stage by explicit pathspec (`git add docs-site/vercel.json`); never `git add -A` or `git add .`.

### M2 — Delete the four changelog index pages (Priority Medium; reversible by reverting one commit)

- Delete `docs-site/content/{ko,en,ja,zh}/changelog/_index.md` in one commit (all four; REQ-REDIR-003). Do not touch the `moai-mx.md` pages or the footer file.
- After the commit the `changelog` directories are gone; `/changelog` and `/<locale>/changelog` are served by the kept rules alone.

### M3 — Closing measurement (Priority High)

- Run the done checks of AC-PRC-011, AC-PRC-010, AC-REDIR-001, AC-REDIR-002, and AC-REDIR-003 against the closing commit, as separate read-only invocations in one parallel batch where independent. Write each command, its verbatim stdout, its exit code, and the commit SHA into progress.md §E.2 (manager-develop's section).
- Re-run the controls once more with the extracted script so the verdict is read from a verified instrument.
- Regression-guard criteria are recorded as observations and not counted as passes.

## C. Technical approach and constraints

- The change is data editing in a JSON file and four file deletions; no code, template, or workflow. `docs-site` is not mirrored into `internal/template/templates/` (E-R13), so there is no Template-First step.
- The checks use Python 3 (`python3 -I`), `git`, `bash`, and `hugo`; none needs a Go toolchain (`hugo.toml` declares no module, E-R13). The lane's resource rule bars `go test`, `go build`, `go vet`, lint, and `make`; none is part of this plan.
- Locale rules: the changelog deletions land in all four locales in one commit; no prose is translated; ko is canonical and the chain ko -> en -> ja/zh is satisfied by deleting all four at once (`hns-oss-docs-i18n-rules` §1 and §2). Page removal carries no redirect obligation of its own (`hns-oss-docs-i18n-rules` §8 covers moves and renames); the changelog URL stays served by the kept rules.
- Commits: Conventional Commits with the card id `t1718` in the message, staged by explicit pathspec, one commit per milestone that edits files.

## D. Green path per acceptance criterion

- AC-PRC-011: the four moai-mx lines flip at M1; the four changelog lines flip at M2. Proof on a scratch copy: acceptance.md E-R14 (`shadowed: 0`, exit 0).
- AC-PRC-010: the nine `/agency` lines flip at M1; the five external lines become admitted by the amended REQ-PRC-010 (the check is run with the admitted URL and its two admitted sources as arguments). Proof on a scratch copy: E-R14.
- AC-REDIR-001, -002, -003: regression-guards, observed at M0 (baselines) and M3.

## E. Risks

- **Public URL change.** Removing a rule changes what a public URL serves. The verdicts for Q21, Q23, and Q16 are recorded; Q15 rests on live observations (E-R8) of a deployment whose commit is unknown (G1).
- **Admitted-URL argument.** The `resolve` check passes only when run with the admitted URL; run without it, five external lines remain by design. A reader who runs it without the argument sees `unresolved: 5` and exit 1; the done check names the argumentss.
- **Footer dependency.** The footer link depends on the kept rule at line 118 (G10). A later change that removes that rule would orphan it; AC-REDIR-001 D1 guards the rule's presence while the footer link exists.
- **Vercel precedence is inferred.** The claim that a redirect hides the built page rests on E-R8 for one deployment (G1, G8).
- **Check blind spots.** Case-variant sources, `has` or `missing` conditions, `url:` front matter, static files, and Hugo-generated URLs are not seen (G3). The mutant probes in acceptance.md § B name the variants that other commands catch.
- **Rebase drift.** If main moves, the line numbers (88, 93, 98, 113, 118, 163, 168) and the counts (39 redirects, 612 pages) can change. M0 re-measures them.

## F. Tests named, not run

- `redirect_destinations_resolve_all_locales`: realized as the check script's `resolve` mode with the fixtures of E-R6 and the scratch copies of E-R14. Not run as a repository test; this child adds no test file.
- `redirect_sources_shadow_no_page`: realized as the `shadow` mode with the same fixtures and copies.
- The Hugo build is named in AC-REDIR-003 and not run in this planning pass (G2); the i18n script was run (E-R10).

No Go test was run, and no test files are added by this child.

## G. File overlap with other cards

- The lane's observation, relayed by the leader: none of twelve in-flight branches touches `docs-site/vercel.json` or `docs-site/content/*/changelog`; the command per branch is `git diff --name-only 0ddf5ca48b...<branch> -- docs-site/vercel.json docs-site/content/*/changelog`, printing nothing. This author did not re-run it (acceptance.md G5). Under the chosen end state no other path changes, so this covers the whole change.
- Within the umbrella: GATE edits workflow files and the cited-verbs producer; README and the SWEEP children edit pages and README files. The four `moai-mx.md` pages may be edited by a SWEEP child for stale citations; REDIR does not edit them, and AC-PRC-011 D2 and D3 still apply to them after both land.

## H. Cross-references

- Umbrella: `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (spec.md § Child SPECs, decision-index.md stubs, acceptance.md E-15a to E-15e).
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- Verification discipline: `.claude/rules/moai/development/verification-completeness.md` §1 to §4; `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 and §3.1.
- Docs i18n rules: `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md`, `.claude/skills/hns-oss-docs-verify/SKILL.md`.
