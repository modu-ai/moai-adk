# SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 — Implementation Plan (revision 0.1.0)

Methodology: documentation edits verified by the RED-now and green cells of acceptance.md §C (no code, so no test-first cycle; the grep cells are the red state and the green check). Milestones are ordered by decision reversibility: the decision-dependent work that is most likely to change is stated first in the section order of this plan, and the mechanical row deletions that are already decided come later. Execution order is the dependency order in § B. No time estimates are given.

## A. Context

- **Card and branch.** Factory card t1727 (Class C, docs only), branch `WT-workflow-cmd-index`, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1727`, base `74ad197380a0711cdb9816d4e37c94d29af730ea` (local `main`). Child 11 of 11 of SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001.
- **Artifacts.** spec.md, plan.md, acceptance.md (Tier M), plus decision-index.md and progress.md. Measurements are the ledger cells of acceptance.md §C.
- **Decision rows.** Q13 and Q17 carry operator verdicts (board record `d-20261019T015908Z-220a`, 2026-10-19): replace where a current equivalent exists, delete where none exists; all four locales in the same change. They are inputs, not choices. Q14 is open.
- **Who edits `docs-site/content/<locale>/workflow-commands/_meta.yaml` (card instruction: confirm the editing actor first).** Read from the agent definitions in `.claude/agents/harness/`:
  - `hns-oss-docs-structure-curator-specialist.md`, description: "Single writer on shared config: per-locale content/<locale>/_meta.yaml section order, data/menu/main.yaml 4-locale name maps + icons, matching SVG cases in layouts/partials/menu.html, and vercel.json redirects". Its Role item 1 names `docs-site/content/<locale>/_meta.yaml` for all four locales and says a change lands in all four in the same change. Its Boundaries confine its writes to `_meta.yaml`, `main.yaml`, `menu.html`, and `vercel.json`, forbid editing page body content, and forbid `git commit`, `git push`, and `gh pr` (publishing is orchestrator-gated).
  - `hns-oss-docs-content-author-specialist.md` says: "Navigation/structure (`_meta.yaml`, `data/menu/main.yaml`, …) → the `structure-curator` specialist … do not edit shared config yourself (single-writer rule)". It edits canonical-locale (ko) page bodies.
  - `hns-oss-docs-locale-translator-specialist.md` derives en, ja, and zh page content from ko in the same change.
  - **Conclusion for this plan:** the four `_meta.yaml` edits (the `moai-design` key and its `title: "/moai design"`) belong to the structure curator. The `_index.md` row deletions and the page work (ko first, then en, then ja and zh in parallel) belong to the content author and the locale translator, whose definitions cover page bodies. Citing: the curator definition's description and Role item 1, the content author's Role section (handoff to the curator), and the curator's Boundaries.
  - **Open item O-1 (two readings, not guessed).** The curator is a spawnable harness specialist (`tools: Read, Write, Edit, Grep, Glob, Bash, Skill`), but the repository's run-phase owner for edits is `manager-develop`, and the status-transition matrix names no harness specialist. Reading 1: the lane (an orchestrator for its card) spawns the curator for the four `_meta.yaml` edits and the content-author and translator specialists for the page edits, and `manager-develop` is not used for these files. Reading 2: `manager-develop` makes all twelve edits as the run-phase owner and honors the curator's quality bar (four locales together, `hugo --minify --gc` warning-free) without spawning it. The definitions state who is the single writer; they do not state how the run-phase delegation reaches it. The leader rules before the Kickoff. Either reading changes who performs M3 step 1 only; no requirement changes.
- **Hugo build gate of the curator.** The curator's exit gate is `cd docs-site && hugo --minify --gc` warning-free. This plan lists it as a verification the curator or the lane may run at M4; it was not run in this planning pass (acceptance.md G-2).

## B. Milestones (dependency order; the milestone that waits is marked)

### M1 — Priority High: pre-flight and the Q13 determination (proceeds now)

1. Confirm HEAD is `74ad197380` or a descendant that leaves acceptance.md cells C-1 to C-9 unchanged (re-run C-1, C-2, C-4, C-8; compare); if any differs, stop and report the drift.
2. Re-run C-7 and C-9 to confirm the determination behind the Q13 verdict: no current equivalent exists for the design family. If a registered command or CLI verb for the design phase has appeared, the verdict says replace, and the replacement is that command; stop and report so the requirement text (REQ-SWF-002) is applied with the equivalent named.
3. Record the leader's ruling on O-1 (editing actor), O-2 (`main.yaml`), O-3 (parity baseline), and O-4 (count sentences) in `progress.md` §E.2. None blocks M2.

### M2 — Priority High: `_index.md` design rows in four locales (proceeds without Q14; REQ-PRC-001, REQ-SWF-002, REQ-SWF-003)

Edit order follows the i18n chain (ko canonical, then en, then ja and zh in parallel) and all four files land in one commit.

1. ko `_index.md`: delete the summary-table row (line 42, `조건부`) and the subcommand-table row (line 66). Do not touch the `/moai goal` summary row (line 41): it resolves.
2. en, ja, zh `_index.md`: delete the single design row (line 62 in each).
3. Do not replace a deleted row by a sentence that names the command (REQ-SWF-002). Do not delete any other row.
4. Run C-4 (expect en 20, ja 20, ko 21, zh 20) and the `_index.md` lines of C-1 and C-2 (expect none). This brings the locale asymmetries level as follows: the two extra Korean design lines and the single design line elsewhere all go, leaving only the `/moai goal` summary row as a documented, untouched, non-stale difference (+1 in ko, unchanged by this work).

### M3 — Priority High: the `moai-design` page and the `_meta.yaml` entries (waits for Q14)

Starts only when Q14 carries an operator verdict. Which pieces run is set by the answer:

- Q14 = remove (REQ-SWF-004): the structure curator removes the `moai-design` key and title from the four `_meta.yaml` files (O-1); the page files are deleted in the same commit; run AC-SWF-005 (C-5 expect no line under `docs-site/content`, R-5 parse check). The leader rules O-2 and O-3 first, because `main.yaml:208` and `.locale-parity-baseline:51` then name a missing page.
- Q14 = rewrite (REQ-SWF-005): the content author rewrites the ko page, the translator derives en then ja and zh to one structure (R-6); the curator retitles the four `_meta.yaml` entries. Run AC-SWF-006 Then 1 and 3 and C-1, C-2.
- Q14 = marker (REQ-SWF-006): record the marker token in `progress.md` §E.2 (coordinate with SWEEPE's Q24 token and GATE's exemption list, G-6); mark each remaining design line in the four pages; add the opening statement; the curator retitles the `_meta.yaml` entries to a non-slash title (`_meta.yaml` carries no marker). Run AC-SWF-006 Then 2 and 3 and R-1.

### M4 — Priority Medium: verification and the preservation proof (waits for M2, and for M3 where Q14 applies)

1. Re-run every cell of acceptance.md §C that the answered branch uses, on the closing commit, and record command, stdout, and exit code in `progress.md` §E.2.
2. R-4 (changed set): only the paths of spec.md § Scope appear; none of `main.yaml`, `.locale-parity-baseline`, or the six other workflow-commands pages (REQ-SWF-007). A Q14 = remove run shows four deleted paths.
3. `git status --short` prints nothing; `moai spec lint --strict` on this SPEC directory stays clean.
4. Optional, by the curator: `cd docs-site && hugo --minify --gc` warning-free.

## C. What proceeds without Q14, and what waits

| Piece | Files | Q14 needed? | Why |
|-------|-------|-------------|-----|
| `_index.md` design rows (Q13 verdict: delete; Q17: four locales) | 4 | no | no equivalent exists (C-7 to C-9); the rows are deleted whatever happens to the page |
| Q13 equivalence determination, pre-flight, ruling on O-1 to O-4 | none | no | M1 |
| Other citations in the twelve files (Q13/Q17 work on non-design citations) | 0 files | no | there are none: every unresolved citation in the twelve files is the design family (C-2 equals C-1), so the "other citations" part of the verdict has nothing to act on here. Stated as a measured fact, not a gap |
| `_meta.yaml` entry and title (design) | 4 | yes | key removed under (a); retitled under (b) and (c); a kept page needs a title |
| `moai-design.md` (four locales) | 4 | yes | deleted under (a), rewritten under (b), marked under (c) |
| REQ-SWF-004, -005, -006, AC-SWF-005, AC-SWF-006 | spec | yes | one of the three applies |

## D. PRESERVE list and its check

- Byte-unchanged (REQ-SWF-007): the six other `workflow-commands` pages in each locale (`moai-goal`, `moai-harness`, `moai-plan`, `moai-project`, `moai-run`, `moai-sync`), every other page of `docs-site/content/`, `docs-site/data/menu/main.yaml`, `docs-site/.locale-parity-baseline`, `docs-site/vercel.json`, the four READMEs, and everything under `.claude/`, `internal/`, `cmd/`, `scripts/`. Unless the leader rules O-2 or O-3 otherwise.
- Check: acceptance.md R-4 on the closing commit against `74ad197380` (a pathspec of `docs-site` and the four READMEs prints only the Scope paths; the pathspec `.claude .moai/docs scripts internal cmd` prints nothing). This is a guard criterion (AC-SWF-004) with named mutants.
- The ko `/moai goal` summary row (line 41) and the `new_items` list in each `_meta.yaml` stay as they are.

## E. Constraints

- Touch only the paths of spec.md § Scope. No page body outside the three kinds, no URL, no Mermaid change (none is added, so the TD-only rule is not engaged).
- No emoji in body text; keep the icon shortcode in the ko `/moai goal` row untouched.
- Four locales in one commit (Q17 verdict); ko first, then en, then ja and zh, per `.moai/docs/docs-site-i18n-rules.md` §17.3.
- No push, no pull request, no branch-protection change; the operator batches the push. Stage by explicit pathspec in the card tree.

## F. Risks

- **Orphaned page.** Under Q14 = rewrite or marker the `_index.md` rows are gone by the Q13 verdict, so the page has no index link; the sidebar data file still reaches it. A non-slash pointer is part of the Q14 answer.
- **Stale sidebar data.** `main.yaml` lines 204 to 208 keep the `/moai design` name map unless O-2 adds the file; under Q14 = remove, the menu template hides the entry for a missing page (read from the template, not run).
- **Count sentence.** "16" beside 15 table rows (O-4) is a visible inconsistency after M2.
- **Marker interface.** Q14 = marker adds a token that GATE's checker must exempt; two independently defined tokens would make the parent's closing run fail (G-6).
- **Later measurement.** The GATE checker and the parent's M6 will measure these files; this child does not wait for them and does not tune its cells to their output.

## G. Self-verification deliverables (reported verbatim with command, output, exit code, HEAD SHA)

1. C-1, C-2, C-3, C-4, and the branch cells (C-5 or C-6) on the closing commit.
2. R-4 output and `git status --short`.
3. `moai spec lint --strict` on this directory.
4. A blocker report, never a question to the user, for any decision this plan did not settle.
