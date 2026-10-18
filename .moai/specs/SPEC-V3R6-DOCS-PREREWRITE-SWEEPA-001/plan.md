# SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001 — Implementation Plan

## A. Measured inventory

### A.1 Pin

- Tree commit: `2bcce5d9dc4c5f57bb98bcfa553e0061fa373681`, branch `WT-id-spec-v3r6-2`, working tree clean. Local `main` resolves to the same commit at plan time (acceptance.md E-SWA-0). This is the document-level pin: every tree-attributed cell here and in acceptance.md §C names this commit unless it carries its own.
- The twelve files are reached by three pathspecs, written once in acceptance.md §A. `git ls-files` expands them to exactly 12 paths (E-SWA-1).

### A.2 E-01 hit inventory (12 matches in 12 files, one per file)

Per-line inventory. The `Context` column is what the run phase must read before editing; it is not a proposed edit.

| # | Page | Locale | Line | Class | Matched form | Context |
|---|------|--------|------|-------|--------------|---------|
| 1 | bas-navigator | ko | 75 | C | `moai codemaps` | Last command of the fenced bash block (lines 67 to 76). Line 74 is the step-3 comment `# 3) 그래프 다시 짜기`; line 65 is the prose `토큰을 한 번 찍어 보고 그래프를 다시 짜 봅니다.` |
| 2 | bas-navigator | en | 75 | C | `moai codemaps` | Same block. Line 74 `# 3) Rebuild the graph`; line 65 `Let's stamp a token once and rebuild the graph.` |
| 3 | bas-navigator | ja | 75 | C | `moai codemaps` | Same block. Line 74 `# 3) グラフを組み直す`; line 65 `トークンを一度打って、グラフを組み直してみます。` |
| 4 | bas-navigator | zh | 75 | C | `moai codemaps` | Same block. Line 74 `# 3) 重画图`; line 65 `先打一枚令牌试试，再重画图。` |
| 5 | factory-mode | ko | 56 | N | `moai cg`, `moai gpt` | Last row of the "Refused forms" table; line 57 is blank. The row says `moai cg` is retired (preview with `moai migrate cg`) and `moai gpt` does not exist, GPT models run through `moai codex`. |
| 6 | factory-mode | en | 56 | N | `moai cg`, `moai gpt` | Same row, en wording: `` `moai cg` is retired and exits with a migration notice (preview with `moai migrate cg`). `moai gpt` does not exist; GPT models run through `moai codex` ``. |
| 7 | factory-mode | ja | 56 | N | `moai cg`, `moai gpt` | Same row. |
| 8 | factory-mode | zh | 56 | N | `moai cg`, `moai gpt` | Same row. |
| 9 | profile-matrix | ko | 88 | N | `moai model profile` | Clause inside a sentence: the 39-cell table, the resolver, and the `moai model profile` accessor "have receded", while the measured model-choice standards remain. |
| 10 | profile-matrix | en | 95 | N | `moai model profile` | Line 95 ends `...the `moai model profile` accessor are`; line 96 begins `gone, but the **model-choice standards** the matrix established from`. The match is on line 95 only. |
| 11 | profile-matrix | ja | 53 | N | `moai model profile` | Same clause. |
| 12 | profile-matrix | zh | 53 | N | `moai model profile` | Same clause. |

Totals: per page 4, 4, 4; per locale 3, 3, 3, 3; Class C 4, Class N 8. Line numbers are coordinates at the pin and move (R1); a hit is identified by its content.

**Edit units.** The hit is the trigger, the edit unit is what may change. This table is the single statement of the allowed edit; spec.md Scope and REQ-PRC-001-004 point here, and the per-file numstat bounds in acceptance.md AC-PRC-001-005 are derived from it.

| Page | Edit unit per locale (lines at the pin) | Removed lines | Added lines |
|------|------------------------------------------|---------------|-------------|
| bas-navigator | The step-3 comment (line 74) and the `moai codemaps` command (line 75), as a pair: both deleted, or both replaced. May also remove the blank line 73 before the step and reword the one sentence at line 65 that announces the rebuild. No other line. | 2 to 4 | 0 to 3 |
| factory-mode | The one "Refused forms" table row (line 56): deleted, or reworded in place. | exactly 1 | 0 or 1 |
| profile-matrix | The sentence that holds the match (en 95 to 98, ko 88 to 90, ja 53, zh 53): reworded, or the clause deleted. | 1 to 4 | 0 to 4 |

For bas-navigator the line-74 and line-75 pair must leave together. A step comment without its command, or a command without its comment, is the dangling remainder that AC-PRC-001-005 catches. The profile-matrix and factory-mode rows apply only under Q-SWA-1 option (a); under option (b) those eight files do not change.

### A.3 CLI surface check of each cited form

Probes use the installed build (A.4). Source corroboration reads the tree at the pin.

| Form | Probe | Observed | Source at the pin |
|------|-------|----------|-------------------|
| `moai codemaps` | `moai codemaps --help` | exit 1, `Unknown command "codemaps" for "moai".` | no top-level `Use: "codemaps"` in `internal/` or `cmd/` |
| `moai gpt` | `moai gpt --help` | exit 1, `Unknown command "gpt" for "moai"`, suggestions glm, gate, gtd, pr | none |
| `moai cg` | `moai cg --help` | exit 1, `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration` | `internal/cli/cg.go:13`, hidden retired launcher |
| `moai migrate cg` | `moai migrate --help` | listed: `cg  Preview migration of legacy CG teammate roles` | `internal/cli/migrate_cg.go:22` |
| `moai model` | `moai model --help` | exit 1, `Unknown command "model" for "moai"`, suggestion codex | none |
| `moai navigator-sync` | `moai navigator-sync --help` | prints help; absent from `moai --help` | `internal/cli/navigator_sync.go:32-34`, `Hidden: true`, writes `nav-graph.json` |

Reading: the Class C instruction does not resolve. The three Class N statements (`moai cg` retired, `moai gpt` absent, the `moai model` accessor gone) match what the build does.

### A.4 Tool provenance

The installed `moai` reports `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z. Commit 7705b365e2 is a strict ancestor of the tree HEAD (`merge-base --is-ancestor` exit 0; `rev-list --count 7705b365e2..2bcce5d9dc` prints 48). The probes were therefore not made by a build of this tree, and none was built (no Go build, by the lane's machine rule). The mitigation is a source read at the pin, not a substitute: of the 48 commits, 15 files under `internal/cli` and `cmd` differ, and the `Use:` declarations at the pin show no top-level registration of `codemaps`, `gpt`, or `model` (E-SWA-6). The claim cited is "installed build 7705b365e2 against tree 2bcce5d9dc".

### A.5 Per-form evidence for the Q13 choice

- **Class C, `moai codemaps` (step 3, "rebuild the graph").** The step regenerates `nav-graph.json`. That file's producer is the hidden verb `moai navigator-sync` (A.3). The registered slash command `/moai codemaps` resolves, but `.claude/commands/moai/codemaps.md` describes it (its `description:` line, translated from Korean) as scanning the codebase and generating architecture docs in `codemaps/`, and neither that file nor `.claude/skills/moai/workflows/codemaps.md` contains `nav-graph` (E-SWA-6). It shares a name with the cited form and does not do the cited job, so it is not an equivalent. A replace option would therefore document a hidden verb, which Q13 leaves to the operator to name.
- **Class N lines.** A retired or absent verb has no current equivalent; `moai migrate cg` is already named inside the factory row. Whether these lines are edited at all is Q-SWA-1; how they are edited, if so, is Q13.

## B. Milestones (ordered by likelihood of change; mechanical steps last)

- **M1 (Priority High): close the decision gate.** Done when Q13, Q17, and Q-SWA-1 each carry an operator verdict; the gate set is stated in decision-index.md (Gate) and not here. No tracked file outside this SPEC directory changes before M1 closes.
- **M2 (Priority High): Class C edit.** Edit the four `bas-navigator.md` files within their edit unit (§A.2), one disposition in all four locales. Read the sentence at line 65 and the lines around the pair in each locale, and keep the page coherent after the step is gone or replaced (R2).
- **M3 (Priority Medium): Class N edit, or none.** If the Q-SWA-1 verdict is option (a), edit the `factory-mode.md` row and the `profile-matrix.md` sentence in all four locales within their edit units, reworded or deleted as the verdict states. Rewording in ko, ja, and zh follows the native-idiom rule inline; these are clause-level edits, not a heavy artifact. Q13 does not govern these lines. If the verdict is option (b), no file changes in this milestone and acceptance.md lists the eight lines.
- **M4 (Priority High before landing): sequencing re-verification.** Re-run the ancestry probes of acceptance.md E-SWA-8 against current `main`. If commit 84f9387535 has landed, re-run AC-PRC-001-005 on the merged tree. If card t1726 has landed, re-read the surrounding lines of `factory-mode.md` by content.
- **M5 (Priority Medium): verification.** Run the AC commands of acceptance.md §B from the card worktree root and carry the deciding output into the verdict file (`.moai/reports/t1721/verdict.md`), not into scratch.

## C. Risks

- **R1, moving coordinates.** Among the 25 most recent commits touching these three pages, 18 changed `factory-mode.md`, 5 changed `profile-matrix.md`, none changed `bas-navigator.md`, and 3 were merges listing no file (E-SWA-8). Line 56 can move before the run phase starts. Every check keys on content.
- **R2, prose adjacency in bas-navigator.** Line 65 announces the rebuild in all four locales. Removing step 3 leaves that sentence promising something the block no longer shows. The edit unit (§A.2) therefore includes that one sentence, so the edit may reword it; it may add no new prose beyond the mechanical need. The dangling comment is machine-checked (AC-PRC-001-005); the sentence is a reviewer read.
- **R3, information lost by deleting a Class N line.** Deleting the factory row removes the only sentence on the en page (the other locales were not checked) that tells the reader that GPT models run through `moai codex`; rewording keeps it, together with the `moai migrate cg` pointer. The choice between the two is an explicit option of Q-SWA-1.
- **R4, the umbrella's closing check.** If Q-SWA-1 is answered "retained", the umbrella's closing check (E-01 exits 1 with empty stdout) and the GATE checker must admit the same eight lines. That is an amendment outside this SPEC.
- **R5, installed-build lag.** The CLI probes rest on a build 48 commits behind the tree (A.4).
- **R6, merge with card t1863.** The two ranges are disjoint with one unchanged blank line between them (E-SWA-8). No trial merge was made, so a clean merge is expected, not observed.

## D. Tests named, not run

- No Go test exists or is added; no Go build or test was run in this plan phase. The checks are the commands in acceptance.md §B. Each RED-now observation and each control was run once against the pin (acceptance.md §C). No green-path run exists yet, because nothing has been edited.

## E. Open questions

The decision rows Q13, Q17, and Q-SWA-1 live in decision-index.md and are not restated here. No other question is open.
