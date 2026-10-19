# SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001 — Implementation Plan

## A. Measured inventory

### A.1 Pin

- Tree commit: `74ad197380a0711cdb9816d4e37c94d29af730ea`, branch `WT-cli-windows-guide`, working tree clean. Local `main` resolves to the same commit at plan time (acceptance.md E-SWC-0). This is the document-level pin: every tree-attributed cell here and in acceptance.md §C names this commit unless it carries its own.
- The twelve files are reached by three pathspecs, written once in acceptance.md §A.1. `git ls-files` expands them to exactly 12 paths (E-SWC-1).

### A.2 W01 hit inventory (12 matches in 12 files, one per file) and the T-eq result per line

Line numbers are coordinates at the pin and move (R1); a hit is identified by its content. The `Context` column is what the run phase must read before editing; it is not a proposed edit. All 12 lines are Class N: each is the sentence "`moai cg` has been retired ..." in its locale (verbatim in acceptance.md E-SWC-2). T-eq is defined in spec.md Scope; its basis is in A.3. Result for every row: **replaceable**, with replacement R = `moai migrate cg`. The result column is the single place the per-line outcome is recorded; the acceptance criteria point here.

| # | Page | Locale | Line | Class | Form | T-eq | Context |
|---|------|--------|------|-------|------|------|---------|
| 1 | getting-started/cli | ko | 338 | N | `moai cg` | replaceable | Paragraph after the `moai cc` / `moai glm` usage block; the same line ends with the CG-page link. Same sentence as README REQ-PRC-004's Korean target (E-SWC-8). |
| 2 | getting-started/cli | en | 336 | N | `moai cg` | replaceable | Same position; the line ends with the link `[CG retirement and migration](/en/multi-llm/cg-mode/)`. |
| 3 | getting-started/cli | ja | 336 | N | `moai cg` | replaceable | Same position, same link at line end. |
| 4 | getting-started/cli | zh | 336 | N | `moai cg` | replaceable | Same position, same link at line end. |
| 5 | getting-started/windows-guide | ko | 108 | N | `moai cg` | replaceable | First paragraph of the section "CG retirement and migration"; the link sits on a separate later line. Same sentence as README's Korean target. |
| 6 | getting-started/windows-guide | en | 108 | N | `moai cg` | replaceable | Heading at line 106; line 110 is the `claude-glm` paragraph; line 112 is the link line. |
| 7 | getting-started/windows-guide | ja | 125 | N | `moai cg` | replaceable | Same section. |
| 8 | getting-started/windows-guide | zh | 113 | N | `moai cg` | replaceable | Same section. |
| 9 | core-concepts/what-is-moai-adk | ko | 441 | N | `moai cg` | replaceable | First paragraph under the subheading "CG retirement and migration". Same sentence as README's Korean target. |
| 10 | core-concepts/what-is-moai-adk | en | 439 | N | `moai cg` | replaceable | Subheading at line 437. Line 435 already says "use `moai migrate cg` to preview explicit migration choices"; line 441 begins "This writes `llm.team_mode: claude` ...". |
| 11 | core-concepts/what-is-moai-adk | ja | 442 | N | `moai cg` | replaceable | Same section. |
| 12 | core-concepts/what-is-moai-adk | zh | 440 | N | `moai cg` | replaceable | Same section. |

Totals: per page 4, 4, 4; per locale 3, 3, 3, 3; Class C 0, Class N 12. Only the English windows-guide, cli, and what-is-moai-adk neighborhoods and the Korean hit lines were read in full; the other locales were identified by content (acceptance.md G6).

**Edit unit.** Defined in spec.md Scope: the matched line, and one adjacent blank line when the line is deleted. For cli.md the matched line also carries the link to the CG page, so a replacement keeps the link and a deletion would lose it; at the pin every line is replaceable, so no deletion is expected, and the loss is a risk only if OI-1 turns every line deletable (R3). The numeric bounds derived from this definition are written once, in acceptance.md AC-PRC-001-002.

### A.3 Basis of the T-eq result

Probes use the installed build (A.4); source corroboration reads the tree at the pin (E-SWC-6).

| Condition of T-eq | Probe | Observed |
|-------------------|-------|----------|
| (1) the retirement diagnostic of `moai cg` names R | `moai cg --help` | exit 1, `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration`; source `internal/cli/cg.go:9` carries the same text |
| (2) R is registered | `moai migrate --help` | lists `cg [--flags]  Preview migration of legacy CG teammate roles`; source `internal/cli/migrate_cg.go:22`, `Use: "cg"` |

Other first-verb forms on the three English pages (a candidate list from a counting pipeline over `git grep -o`, not evidence): `moai cc`, `moai glm`, `moai profile`, `moai update`, `moai init`, `moai hook`, `moai doctor`, `moai version`, `moai inventory`, `moai worktree`, `moai status`, `moai web`, `moai session`, `moai handoff`, `moai goal`, `moai spec`, `moai harness`, `moai migrate cg`. Each of `inventory`, `web`, `status`, `migrate` appears in the installed `moai --help` list. The form `moai instance` (one match in the counting list) printed `Unknown command "instance"` when probed; the umbrella plan §A.3 names "instance already" among its prose-noise tokens, but this pass did not read that line, so it is neither classified nor claimed as prose (acceptance.md G5).

### A.4 Tool provenance

The installed `moai` reports `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z. Commit 7705b365e2 is a strict ancestor of the tree HEAD (`merge-base --is-ancestor` exit 0; `rev-list --count 7705b365e2..HEAD` prints 68). The probes were therefore not made by a build of this tree, and none was built (no Go build in this pass). The mitigation is a source read at the pin (E-SWC-6), not a substitute. The claim cited is "installed build 7705b365e2 against tree 74ad197380".

## B. Milestones (ordered by likelihood of change; mechanical steps last)

- **M1 (Priority High): close the decision gate.** Done when Q-SWC-1 carries an operator verdict; the state of the rows is stated in decision-index.md (Gate) and not here. If OI-1 is resolved against the T-eq reading, A.2's result column is re-recorded in this milestone. No tracked file outside this SPEC directory changes before M1 closes.
- **M2 (Priority High): the edit, or none.** Under Q-SWC-1 = (a): recompute S at read time, then edit the Edit unit of every file in S in all four locales in one change, replacing each line (T-eq true) with wording that drops the W01 form, keeps the information that the launcher is retired, and cites `moai migrate cg`, or deleting it (T-eq false). Rewording in ko, ja, and zh follows the native-idiom rule inline; these are one-line edits, not a heavy artifact. Read the lines around each unit and keep the page coherent (R2). Under (b): no file changes, and acceptance.md lists the twelve lines.
- **M3 (Priority High before landing): sequencing re-verification.** Re-read S on current `main`; if the README child has landed, confirm which Korean files left S and that no obligation of this SPEC targets them. Re-run AC-PRC-001-003 on the merged tree.
- **M4 (Priority Medium): verification.** Run the AC commands of acceptance.md §B from the card worktree root and carry the deciding output into the verdict file (`.moai/reports/t1723/verdict.md`), not into scratch.

## C. Risks

- **R1, moving coordinates.** The 8 most recent commits touching these three pages are listed in E-SWC-10; one of them (e4fb8c3ca3, card t1441) adds 11 lines to `getting-started/cli.md` in all four locales. Which commits touched the notice lines was not examined. Line numbers can move before the run phase starts. Every check keys on content.
- **R2, prose adjacency.** Under a replacement the surrounding text stays; under a deletion the heading in `windows-guide.md` and `what-is-moai-adk.md` is followed directly by the next paragraph, and in the English `what-is-moai-adk.md` that paragraph begins "This writes ...". Page coherence in four locales is a reviewer read; no criterion measures it.
- **R3, information lost by deletion.** A deletion of a `cli.md` line removes the only link to the CG page on that line. At the pin T-eq makes every line replaceable, so this arises only if OI-1 turns lines deletable.
- **R4, the umbrella's closing check.** If Q-SWC-1 is answered "retained", the umbrella's closing check (E-01 exits 1 with empty stdout) and the GATE checker (exemption C-7 names only `cg-mode.md` and `moai-todo.md`) must admit the same twelve lines. That is an amendment outside this SPEC.
- **R5, installed-build lag.** The CLI probes rest on a build 68 commits behind the tree (A.4).
- **R6, merge with the README child.** Both cards may edit the same three Korean lines (spec.md Dependencies). No trial merge was made, so a clean merge is not observed; S handles the case where README lands first.

## D. Tests named, not run

- No Go test exists or is added; no Go build or test was run in this plan phase. The checks are the commands in acceptance.md §B. Each RED-now observation and each control was run once against the pin (acceptance.md §C). No green-path run exists yet, because nothing has been edited.

## E. Open questions

The decision rows Q13, Q17, and Q-SWC-1 and the open item OI-1 live in decision-index.md and are not restated here. No other question is open.
