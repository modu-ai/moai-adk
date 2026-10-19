# SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 — Implementation Plan (revision 0.1.1)

Methodology: documentation edits verified by the RED-now and green cells of acceptance.md §C (no code, so no test-first cycle; the grep cells are the red state and the green check). Milestones are ordered by decision reversibility: the page, navigation, and baseline deletions (the user-visible change) come first in the section order, and the mechanical row edits after. Execution order is the dependency order in § B. No time estimates are given.

## A. Context

- **Card and branch.** Factory card t1727 (Class C, docs only), branch `WT-workflow-cmd-index`, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1727`, base `74ad197380a0711cdb9816d4e37c94d29af730ea` (local `main`). Child 11 of 11 of SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001.
- **Artifacts.** spec.md, plan.md, acceptance.md (Tier M), plus decision-index.md and progress.md. Measurements are the ledger cells of acceptance.md §C.
- **Rulings applied (revision 0.1.1).** Operator verdicts of record `d-20261019T015908Z-220a` (Q13 replace where an equivalent exists, delete where none exists; Q17 all four locales in the same change) and `d-20261019T021508Z-258e` (Q14: delete the `moai-design` page in all four locales, remove the `main.yaml` entry and baseline line 51, delete the `_index.md` design rows; relayed by the dispatching coordinator, acceptance.md G-10). Leader ruling `d-20261019T020706Z-1c58` (decided by the leader, session `d1aa9eb6`, under operator delegation, ladder path 6): O-1 = (a), O-2 = include `main.yaml:204-208`, O-3 = update the parity baseline line, O-4 = correct the "All subcommands (16)" count to the actual row count. The leader ruling is not an operator verdict. No decision row is open.
- **Change set.** Fourteen paths: ten files edited (four `_index.md`, four `_meta.yaml`, `main.yaml`, `.locale-parity-baseline`) and four pages deleted (spec.md § Scope).

### A.1 Who edits what in the run phase (ruling O-1 = (a))

The lane (an orchestrator for its card) does the spawning; this SPEC names the writers and does not spawn anyone. Read from the agent definitions in `.claude/agents/harness/`, and from the ruling:

| Paths | Single writer | Evidence |
|-------|---------------|----------|
| the four `docs-site/content/<L>/workflow-commands/_meta.yaml` files; `docs-site/data/menu/main.yaml` | `hns-oss-docs-structure-curator-specialist` | Its description: "Single writer on shared config: per-locale content/<locale>/_meta.yaml section order, data/menu/main.yaml 4-locale name maps + icons, ...". Role items 1 and 2 name both surfaces; item 1 requires all four `_meta.yaml` files in the same change. Boundaries: its writes are confined to `_meta.yaml`, `main.yaml`, `menu.html`, `vercel.json`; it never edits page bodies, and never runs `git commit`, `git push`, or `gh pr`. The content-author definition hands both surfaces to it ("do not edit shared config yourself (single-writer rule)"). Ruling O-1: manager-develop does not edit this config |
| ko `_index.md` (rows and count statements); ko `moai-design.md` (delete); the `.locale-parity-baseline` line | `hns-oss-docs-content-author-specialist` | Its Role: canonical-locale (ko) docs-site pages; hands derived locales to the translator and navigation to the curator. **The baseline file is not on the curator's write list and no agent definition names a writer for it** (the repository search of C-15 finds the file name only in the verify recipe and CHANGELOG.md); assigning the one-line deletion to the content author, the specialist that deletes the page, is a plan choice, not an ownership finding |
| en, ja, zh `_index.md` and `moai-design.md` (one writer per locale) | `hns-oss-docs-locale-translator-specialist`, one instance per derived locale | Its Role: derives en, ja, zh in the same change, writes its one assigned locale only (parallel siblings own the others); chain ko, then en, then ja and zh |

Rules this table fixes:

1. **One writer per file.** No file in the change set has two writers; each `_index.md` is edited once per locale (rows and count statements in the same pass).
2. **Order.** Canonical locale first, then en, then ja and zh in parallel (`.moai/docs/docs-site-i18n-rules.md` §17.3; the translator definition). The curator is not part of that chain: it runs in one pass over all four `_meta.yaml` files and `main.yaml`.
3. **manager-develop** does not edit `_meta.yaml` or `main.yaml`. It keeps what the status-transition matrix gives it: the `draft → in-progress` transition on the first run-phase commit and the progress.md §E.2 and §E.3 evidence. Whether the lane also spawns it for the page edits, or uses the harness specialists for those, is the lane's decision; the ruling assigns page bodies to the content author and the translator.
4. **The harness specialists do not commit or push.** The lane stages by explicit pathspec and commits; the push stays with the operator.
5. **Not a grant.** The lane's standing spawn authority (`factory-dispatch.md` § Lane spawn authority) names the matrix specialists and the chain's auditors; the harness specialists are not named there. This plan records the ruling and does not itself grant a spawn.
6. **Deleting a page.** Neither the content-author nor the translator definition names deletion; a canonical deletion with no derived deletion is a locale-parity failure in the translator's own words, so each specialist deletes the page file of its own locale (gap G-3).

- **Exit gate of the specialists.** The curator's gate is `cd docs-site && hugo --minify --gc` warning-free; the verify recipe (`hns-oss-docs-verify`) is the exit gate every specialist runs. Neither was run in this planning pass (acceptance.md G-2).

## B. Milestones (dependency order)

### M1 — Priority High: pre-flight and the Q13 determination

1. Confirm HEAD is `74ad197380` or a descendant that leaves acceptance.md cells C-1 to C-14 unchanged (re-run C-1, C-2, C-4, C-8, C-11, C-12, C-13, C-14; compare); if any differs, stop and report the drift.
2. Re-run C-7 and C-9 to confirm the determination behind the Q13 verdict: no current equivalent exists for the design family. If a registered command or CLI verb for the design phase has appeared, the verdict says replace, and the replacement is that command; stop and report so the requirement text (REQ-SWF-002) is applied with the equivalent named.
3. Cite the board records `d-20261019T015908Z-220a` (Q13, Q17), `d-20261019T021508Z-258e` (Q14), and `d-20261019T020706Z-1c58` (O-1 to O-4) in `progress.md` §E.2.

### M2 — Priority High: the page, navigation, and baseline deletions (REQ-SWF-004, -009, -010; AC-SWF-005, -008, -009)

Order inside M2:

1. Content author: delete the ko page and baseline line 51 (delete only; add no line).
2. Translators: delete the en page, then the ja and zh pages in parallel.
3. Curator, one pass: remove the `moai-design` key and its `title` from the four `_meta.yaml` files, and delete `main.yaml` lines 203 to 208 as one block (the `- name:` line, the four name-map lines, and the `ref` line). Never delete the name-map lines and leave the `ref`.
4. Run C-5 (expect no output), C-13 (expect no output), C-14 (expect no output), R-5, R-8 (no output), R-9 (`6`), R-10 (`0	1	docs-site/.locale-parity-baseline`).

### M3 — Priority High: `_index.md` design rows and count statements in four locales (REQ-PRC-001, REQ-SWF-002, -003, -008)

Edit order follows the i18n chain (ko canonical, then en, then ja and zh in parallel). Edit by content, not by the line numbers measured on `74ad197380`; the first deletion shifts the lines after it.

1. ko `_index.md` (content author): delete the summary-table row (`조건부`, line 42) and the subcommand-table row (line 66). Do not touch the `/moai goal` summary row (line 41): it resolves. Correct the heading (`전체 서브커맨드 (16개)`, line 53) and the sentence (line 55) to the row count of the section after the deletion (15).
2. en, ja, zh `_index.md` (translator, per locale): delete the single design row (line 62 in each) and correct the heading and sentence (lines 50 and 52 in each) to the same number.
3. Do not replace a deleted row by a sentence that names the command (REQ-SWF-002). Do not delete any other row. Do not add a row for `/moai gtd` or `/moai todo` (O-6).
4. Run C-4 (expect en 20, ja 20, ko 21, zh 20), C-12 (expect `15 15 15 15`), C-11 (expect `:2` for each of the four files), and the `_index.md` lines of C-1 and C-2 (expect none).

The count is 15: the design row of the "All subcommands" table is deleted. If M3 adds any row to that section, the number is the new row count and AC-SWF-007 is met by agreement, not by the literal 15.

M2 and M3 may land as two commits or one; all four locales of each land together (Q17; R-3).

### M4 — Priority Medium: verification and the preservation proof

1. Re-run every cell of acceptance.md §C that the closing state uses, on the closing commit, and record command, stdout, and exit code in `progress.md` §E.2.
2. R-4 (changed set): only the fourteen paths of spec.md § Scope appear (four as deletions); none of the six other workflow-commands pages, `docs-site/vercel.json`, or `docs-site/layouts/`.
3. `git status --short` prints nothing; `moai spec lint --strict` on this SPEC directory stays clean.
4. Optional, by the specialists: the verify recipe and `cd docs-site && hugo --minify --gc` warning-free.

## C. Dependencies between pieces

| Piece | Files | Depends on |
|-------|-------|-----------|
| page deletion, baseline line | 4 + 1 | nothing open (Q14 decided) |
| `_meta.yaml` keys, `main.yaml` entry | 4 + 1 | nothing open; same pass as the page deletion so no link to a missing page exists in any committed state |
| `_index.md` rows and count statements | 4 | nothing open |
| Other citations in the twelve files | 0 | none exist: every unresolved citation in the twelve files is the design family (C-2 equals C-1). Stated as a measured fact, not a gap |

## D. PRESERVE list and its check

- Byte-unchanged (REQ-SWF-007): the six other `workflow-commands` pages in each locale (`moai-goal`, `moai-harness`, `moai-plan`, `moai-project`, `moai-run`, `moai-sync`), every other page of `docs-site/content/`, `docs-site/vercel.json`, `docs-site/layouts/`, `docs-site/static/`, the four READMEs, `.moai/config/sections/design.yaml`, and everything under `.claude/`, `internal/`, `cmd/`, `scripts/`. `docs-site/.locale-parity-baseline` changes by exactly one deleted line.
- Check: acceptance.md R-4 on the closing commit against `74ad197380` (a pathspec of `docs-site` and the four READMEs prints only the Scope paths; the pathspec `.claude .moai/docs scripts internal cmd` prints nothing), plus the numstat of AC-SWF-009. This is a guard criterion (AC-SWF-004) with named mutants.
- The ko `/moai goal` summary row (line 41), the `new_items` list in each `_meta.yaml`, and every other `main.yaml` entry stay as they are.

## E. Constraints

- Touch only the paths of spec.md § Scope. No page body outside the deleted pages, no URL, no Mermaid change. No edit under `docs-site/layouts/`.
- No emoji in body text; keep the icon shortcode in the ko `/moai goal` row untouched.
- Four locales in one change (Q17 verdict); ko first, then en, then ja and zh, per `.moai/docs/docs-site-i18n-rules.md` §17.3.
- No icon is added or changed in `main.yaml`; the curator's icon-to-SVG coupling is not engaged.
- No push, no pull request, no branch-protection change; the operator batches the push. Stage by explicit pathspec in the card tree.

## F. Risks

- **Public URLs.** `/<locale>/workflow-commands/moai-design` stop serving. No redirect is in this change (the redirect rules are REDIR's, Q15 relates to `/design`); inbound links 404 until REDIR or another card adds one.
- **Stale sidebar data.** `menu.html` hides a sidebar entry whose page is missing (lines 31 and 77), but `docs-site/layouts/_default/baseof.html` lines 58 to 86 builds the previous/next chain from every non-external `main.yaml` sub entry with no page-exists guard, so a left-behind entry would give `/moai sync` a "next" link and `/moai harness` a "previous" link to a missing page. Read from the template source, not rendered (G-4). REQ-SWF-009 deletes the entry as one block; after it, `/moai sync` is followed by `/moai harness`.
- **Half-deleted entry.** Deleting the four name-map lines and leaving `ref: /workflow-commands/moai-design` attaches a second `ref` key to the preceding `/moai sync` entry, or leaves a null-name entry. AC-SWF-008 names this as a mutant.
- **Count statement.** The number is the row count (15), not the registered count (17); "All" stays inexact because `/moai gtd` and `/moai todo` have pages but no row (O-6). The decision and its alternative are in AC-SWF-007.
- **Baseline reader.** The ruling's premise that CI reads the baseline is not what the tree shows: only the oss-docs verify recipe reads it (C-15, C-16). The deletion is required by the ruling regardless.
- **Spawn authority.** The harness specialists are not in the lane's standing spawn list (A.1 rule 5).
- **Later measurement.** The GATE checker and the parent's M6 will measure these files; this child does not wait for them and does not tune its cells to their output.

## G. Self-verification deliverables (reported verbatim with command, output, exit code, HEAD SHA)

1. C-1, C-2, C-3, C-4, C-5, C-11, C-12, C-13, C-14 on the closing commit.
2. R-4 output and `git status --short`.
3. `moai spec lint --strict` on this directory.
4. A blocker report, never a question to the user, for any decision this plan did not settle.
