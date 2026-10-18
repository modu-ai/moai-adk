# SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001 — Implementation Plan

## A. Measured inventory

### A.1 Pin

- Tree commit: `2bcce5d9dc4c5f57bb98bcfa553e0061fa373681`, branch `WT-tokenomics-overview-mult`, working tree clean. Local `main` resolves to the same commit at plan time (acceptance.md E-SWB-0). This is the document-level pin: every tree-attributed cell here and in acceptance.md §C names this commit unless it carries its own.
- The twelve files are reached by three pathspecs, written once in acceptance.md §A. `git ls-files` expands them to exactly 12 paths (E-SWB-1).

### A.2 E01 hit inventory (12 matches in 12 files, one per file; 0 Class C, 12 Class N)

Per-line inventory and the edit unit, in one table so that the inventory and the allowed edit are the same object. The `Unit` column is the exact range a disposition may change; the hit line is inside it. `Base lines` is the file's line count at the pin (E-SWB-7).

| # | Page | Locale | Hit line | Unit (lines) | Class | Matched form | Base lines |
|---|------|--------|----------|--------------|-------|--------------|-----------|
| 1 | tokenomics-overview | en | 106 | 106 | N | `moai cg` | 124 |
| 2 | tokenomics-overview | ja | 106 | 106 | N | `moai cg` | 124 |
| 3 | tokenomics-overview | ko | 112 | 112 | N | `moai cg` | 131 |
| 4 | tokenomics-overview | zh | 106 | 106 | N | `moai cg` | 124 |
| 5 | multi-llm/_index | en | 172 | 172 | N | `moai cg` | 204 |
| 6 | multi-llm/_index | ja | 165 | 165 | N | `moai cg` | 197 |
| 7 | multi-llm/_index | ko | 164 | 164 | N | `moai cg` | 196 |
| 8 | multi-llm/_index | zh | 153 | 153 | N | `moai cg` | 183 |
| 9 | multi-llm/model-policy | en | 190 | 187 to 190 | N | `moai model profile` | 259 |
| 10 | multi-llm/model-policy | ja | 125 | 125 | N | `moai model profile` | 163 |
| 11 | multi-llm/model-policy | ko | 182 | 180 to 183 | N | `moai model profile` | 246 |
| 12 | multi-llm/model-policy | zh | 124 | 124 | N | `moai model profile` | 162 |

Content of each hit (verbatim lines are in E-SWB-5):

- Rows 1 to 8: the paragraph "`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session." in each locale, ending with the link to the CG retirement page. One line per locale. Same sentence as the retained notice at `multi-llm/cg-mode.md` line 8 (umbrella E-01a).
- Rows 9 to 12: the first paragraph of the "What happened to per-agent assignment (history)" section (zh and ja: one line; en and ko: four hard-wrapped lines). It says that through v3.1 a 39-cell profile matrix was resolved per spawn and could be inspected with the `moai model profile` command; the next paragraph says the apparatus was retired.

Totals: per page 4, 4, 4; per locale 3, 3, 3, 3; Class C 0, Class N 12. The four locales carry the same logical citation on every page; only the line numbers differ (the pages are not line-aligned), so there is no count drift between locales. A hit is identified by content (R1).

**Edit-unit bounds (superseded option (a) path, kept for the record).** No criterion measures these under the ruling.

| Page | Locale | Removed lines | Added lines |
|------|--------|---------------|-------------|
| tokenomics-overview, multi-llm/_index | all four | exactly 1 | 0 or 1 |
| multi-llm/model-policy | en, ko | 1 to 4 | 0 to 4 |
| multi-llm/model-policy | ja, zh | exactly 1 | 0 or 1 |

The ruling is (b): the twelve files do not change, and the bounds above are the superseded option (a) path, kept for the record.

### A.3 CLI surface check of every `moai` form the twelve pages cite

Probes use the installed build (A.4). Source corroboration reads the tree at the pin. Each form is listed once; the four locales carry the same set of forms (E-SWB-6).

| Form | Probe | Observed | Resolves | Page treatment |
|------|-------|----------|----------|----------------|
| `moai cg` | `moai cg --help` | exit 1, `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration` | retired | stated as retired (Class N) |
| `moai migrate cg` | `moai migrate --help` | listed: `cg [--flags]  Preview migration of legacy CG teammate roles` | yes | not cited on these three pages |
| `moai model` / `moai model profile` | `moai model --help` | exit 1, `Unknown command "model" for "moai"`, suggestion codex | no | stated as retired history (Class N) |
| `moai cc` | `moai cc --help` | usage printed | yes | cited as usable |
| `moai glm`, `moai glm setup` | `moai glm --help` | usage printed; `setup [api-key]` listed | yes | cited as usable |
| `moai profile setup` | `moai profile setup --help` | usage printed | yes | cited as usable |
| `moai spec audit` | `moai spec audit --help` | usage printed | yes | cited as usable (`tokenomics-overview.md`) |
| `moai init --model-policy`, `--profile`, `--high`, `--medium-alias`, `--low` | `moai init --help` | all five listed as "Deprecated, no effect" | yes (deprecated stubs) | page says deprecated stubs; accurate |

Reading: no cited form that fails to resolve is presented as usable. The twelve E01 lines are statements that the form is retired or historical, and each statement matches what the build does.

### A.4 Tool provenance

The installed `moai` reports `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z. Commit 7705b365e2 is a strict ancestor of the tree HEAD (`merge-base --is-ancestor` exit 0; `rev-list --count 7705b365e2..2bcce5d9dc` prints 48; E-SWB-0). The probes were therefore not made by a build of this tree, and none was built (no Go build, by the lane's machine rule). The mitigation is a source read at the pin, not a substitute: 15 files under `internal/cli` and `cmd` differ between the two commits, none of them `cg.go` or `migrate_cg.go`, and the `Use:` declarations at the pin show `cg` (hidden, "Retired CG launcher") and `migrate cg`, and no top-level `gpt`, `codemaps`, or `model` (E-SWB-6). The claim cited is "installed build 7705b365e2 against tree 2bcce5d9dc".

## B. Milestones

Q-SWB-1 is decided (b) on board record `d-20261018T201957Z-c26e` (decision-index.md, Gate). No milestone edits a page. The former edit milestones (Class C edit, Class N edit) are not applicable and are removed.

- **M1 (Priority High): re-run the inventory.** On the run-start tree re-run the E01 inventory of §A.2. Any E01 match not in §A.2, by content, is a blocker report (return, do not edit). If a page moved on `main`, recompute the Base lines column from content.
- **M2 (Priority High): regression-guard verification.** Run the AC commands of acceptance.md §B from the card worktree root: the twelve files byte-identical to the pin, the twelve retained lines present in all four locales, the swept-count control. Carry the deciding output into `.moai/reports/t1722/verdict.md`, not into scratch.
- **M3 (Priority Medium): sequencing re-check.** Re-run the overlap probes of acceptance.md E-SWB-8 on current `main` (SWEEPA's file list, any new child spec naming one of the twelve files). Under (b) a finding is recorded, not acted on.

## C. Risks

- **R1, moving coordinates.** The twelve pages are edited occasionally (acceptance.md E-SWB-9: the last 8 commits touching them are model-policy, tokenomics, and index edits). Line numbers can move before the run phase. Every check keys on content; the Unit column is recomputed at M1.
- **R4, the umbrella's closing check.** The umbrella's `cited-verbs: 0` check and GATE's exemption entry C-7 (which names only the `cg-mode.md` and `moai-todo.md` pages) would report these twelve lines. That is outside this card and was recorded as a follow-up of the umbrella closing card.
- **R5, installed-build lag.** The CLI probes rest on a build 48 commits behind the tree (A.4).
- **R6, same-line overlap with the README child.** Korean `tokenomics-overview.md:112` and Korean `multi-llm/_index.md:164` are lines the README child's REQ-PRC-004 targets. It does not conflict under (b), because this SPEC edits nothing.

## D. Tests named, not run

- No Go test exists or is added; no Go build or test was run in this plan phase. The checks are the commands in acceptance.md §B. Each RED-now observation and each control was run once against the pin or a named historical commit (acceptance.md §C). No green-path run exists, because nothing has been edited.

## E. Open questions

The decision rows Q13, Q17, and Q-SWB-1 live in decision-index.md and are not restated here. No other question is open.
