# SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001 — Implementation Plan

## A. Measured inventory

### A.1 Pin

- Tree commit: `26834dd17d3eb9597cde0366a5d4f01130417172`, branch `WT-prompt-caching-guides`, working tree clean. This is the document-level pin of revision 0.2.0: every tree-attributed cell here and in acceptance.md §C names this commit unless it carries its own pin. Revision 0.1.0 was pinned to `74ad197380a0711cdb9816d4e37c94d29af730ea`; the docs-site trees of the two commits are identical (`git diff --stat 74ad197380 26834dd17d -- docs-site` prints nothing, exit 0, acceptance.md E-SWD-0), and every cell that reads docs-site content was re-run on 26834dd17d in this revision. The entries that read the installed CLI, the Go source, and the commit history (E-SWD-10, E-SWD-11, E-SWD-14, E-SWD-16) keep the pin 74ad197380 and are not re-run; between the two commits only this SPEC's own five files changed, so `internal/` and `cmd/` are the same tree.
- The twelve files are reached by three pathspecs, written once in acceptance.md §A. `git ls-files` expands them to exactly 12 paths (E-SWD-1).

### A.2 Hit inventory (40 matching lines in 12 files, measured on 26834dd17d)

Per-line inventory. The `Context` column is what the run phase must read before editing; it is not a proposed edit. Line numbers are coordinates at the pin and move (R1); a hit is identified by its content.

**`prompt-caching.md` — 4 lines, all Class N, E-01 (`moai cg`).**

| # | Locale | Line | Context |
|---|--------|------|---------|
| 1 | ko | 243 | `moai cg`는 폐기되었습니다 ... `moai cc`의 별칭이 아닙니다. Ends with a link to `/ko/multi-llm/cg-mode/`. |
| 2 | en | 268 | `` `moai cg` has been retired. ... It is not an alias for `moai cc`. `` Ends with a link to `/en/multi-llm/cg-mode/`. |
| 3 | ja | 171 | Same sentence, ja wording; link to `/ja/multi-llm/cg-mode/`. |
| 4 | zh | 171 | Same sentence, zh wording; link to `/zh/multi-llm/cg-mode/`. |

The whole file carries no other `moai` token: `git grep -n "moai"` over the four files prints exactly these four lines (E-SWD-12). E-02 and E-03 match nothing in them, and no slash-form family (E-04) matches.

**`mcp-server.md` — 7 lines per locale, 28 in all, all Class C.** The line numbers are identical in the four locales. The matched form sits in the last column, "CLI equivalent", of the tool-catalog tables.

| Row (tool) | Line | Matched form | Pattern |
|------------|------|--------------|---------|
| `mcp__moai__spec_progress` | 130 | `moai spec list` | E-02 |
| `mcp__moai__audit_cache` | 163 | `moai audit cache` | E-01 |
| `mcp__moai__codex_task` | 182 | `moai codex task` | E-02 |
| `mcp__moai__codex_setup` | 183 | `moai codex setup` | E-02 |
| `mcp__moai__codex_job_status` | 184 | `moai codex job status` | E-02 |
| `mcp__moai__codex_job_result` | 185 | `moai codex job result` | E-02 |
| `mcp__moai__codex_job_cancel` | 186 | `moai codex job cancel` | E-02 |

Totals per locale: 1 E-01 line and 6 E-02 lines. Over the four locales: 4 E-01 lines and 24 E-02 lines, which equals the umbrella's own count of 24 CLI lines for this page (umbrella acceptance.md E-02, measured at its pin 8816de4dae).

**`worktree/_index.md` — 2 lines per locale, 8 in all, all Class C, E-03.**

| Locale | Lines | Content |
|--------|-------|---------|
| en | 91 to 92 | `B2 --> B3["moai run SPEC-ID"]`, `B3 --> B4["moai sync SPEC-ID"]` |
| ja | 87 to 88 | same two lines |
| ko | 119 to 120 | same two lines |
| zh | 96 to 97 | same two lines |

`B3` and `B4` occur nowhere else in each file: `git grep -c -w -E "B[34]"` prints 2 per file (E-SWD-5), so the two lines hold every reference to them. Line `B1 --> B2[...]` (ko 118) stays unchanged.

**Adjacent labels, outside every family pattern (Q-SWD-2, left unchanged by the leader ruling of board record 4604).** Five Mermaid labels per locale name an unregistered form but match none of E-01 to E-04, because they carry no `SPEC-ID`: `A1["moai plan<br/>..."]` (en 85, ja 81, ko 113, zh 90), `T1A["moai plan"]` (en 272, ja 266, ko 311, zh 275), `T2C` and `T3C` `["moai run<br/>..."]` (en 280 and 286, ja 274 and 280, ko 319 and 325, zh 283 and 289), and `T4C["moai sync<br/>..."]` (en 292, ja 286, ko 331, zh 295). `git grep -c -E "\"moai (plan|run|sync)(<br/>|\")"` prints 5 per file (E-SWD-13). `moai run` and `moai sync` are `Unknown command`; `moai plan` is a registered verb that is a plan-phase utility group, not the feature-description entry the label implies. These 20 lines are listed so the sweep does not touch them silently.

Totals: Class C 36 lines (28 + 8), Class N 4 lines. Per file: 1 line in each `prompt-caching.md`, 7 in each `mcp-server.md`, 2 in each `worktree/_index.md`.

### A.3 Edit units and bounds

The hit is the trigger, the edit unit is what may change. This table is the single statement of the allowed edit; spec.md Scope and REQ-PRC-001-005 point here, and the bounds in acceptance.md (AC-PRC-001-003, AC-PRC-001-004) are derived from it.

| Page | Edit unit per locale (lines at the pin) | Removed lines | Added lines |
|------|------------------------------------------|---------------|-------------|
| `mcp-server.md` | The "CLI equivalent" cell of the seven rows of §A.2, each edited in place on its own line, to the bare marker `—`. The first three cells of each of the seven rows (tool id, purpose, agents) and every other row stay byte-identical. | exactly 7 | exactly 7 |
| `worktree/_index.md` | The two lines that hold `B3` and `B4`, replaced in place by the registered slash forms: `B3["/moai run SPEC-ID"]` and `B4["/moai sync SPEC-ID"]`, in that order. | exactly 2 | exactly 2 |
| `prompt-caching.md` | none (retain ruling of Q-SWD-1) | 0 | 0 |

The numstat form of git prints added before removed, so the expected rows read `7	7` for `mcp-server.md` and `2	2` for `worktree/_index.md`. The numstat bound is on lines, not cells and not characters: a cell edited in place is one removed and one added line, whatever its new length, and overwriting a different cell of the same row gives the same `7	7` (observed on a scratch fixture, acceptance.md E-SWD-15). The cell-level claim "the row's other cells stay" is therefore measured by its own commands: the first three cells of each row by the per-locale literal-prefix check, and the fourth cell by the exact-form check (AC-PRC-001-003). The slash forms are checked by label, so a swap of the two does not pass (AC-PRC-001-002).

### A.4 CLI surface check of each cited form

Probes use the installed build (A.6). Source corroboration reads the tree at the pin. Two of the probes below need a caution that the first lane observation missed: `moai codex task --help`, `moai codex setup --help`, and `moai codex job status --help` all exit 0 and print the launcher help of `moai codex`, and `moai spec list --help` exits 0 and prints the help of `moai spec`. That is not evidence the verb exists; the commands table printed by the parent help and the source `Use:` declarations are what decide.

| Form | Probe | Observed | Source at the pin |
|------|-------|----------|-------------------|
| `moai cg` | `moai cg --help` | exit 1, `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration` | `internal/cli/cg.go:13`, `Use: "cg"`, `Short: "Retired CG launcher"`, `Hidden: true` (E-SWD-11) |
| `moai cc` | `moai --help` | listed, `cc [-p profile] [-f \| -l]` | launcher |
| `moai audit` (so `moai audit cache`) | `moai audit --help` | exit 1, `Unknown command "audit" for "moai".`; `moai audit cache --help` the same | no top-level `Use: "audit"`; `audit` is a sub-verb of `spec` (`spec_audit.go:64`) and of `codex` (`codex_audit_launch.go:880`); no `Use: "cache"` in `internal/` or `cmd/` |
| `moai spec list` | `moai spec --help` | commands: status, drift, view, lint, close, audit, archive, ceiling. No `list` | `Use:` declarations of `spec*.go`: archive, audit, ceiling, close, drift, lint, status, view |
| `moai spec status --list` | `moai spec status --help` | flag `--list  List all SPECs and their status` | `spec_status.go:29` |
| `moai codex task`, `setup`, `job *` | `moai codex --help` | commands: `audit <role>`, `role-audit <role>`; launch forms `cli`, `status`, `app`. No task, setup, or job | `codex_audit_launch.go:880` (`audit <role>`), `:907` (`role-audit <role>`); the only `Use: "setup..."` declarations are `glm.go:125` and `profile_setup.go:199` |
| `moai codex status` | `moai codex --help` | `moai codex status  print the readiness readout (starts nothing)`: binary, CODEX_HOME, auth provider, project wiring | launch-form alias |
| `moai run`, `moai sync` | `moai run --help`, `moai sync --help` | exit 1, `Unknown command "run"` and `Unknown command "sync"` | no top-level `Use:` |
| `moai plan` | `moai plan --help` | registered; the help says the CLI verb is "distinct from the `/moai plan` slash-command skill" and only renders the plan-phase HTML report | `plan` command group |
| `/moai run`, `/moai sync` | `git ls-files` | `.claude/commands/moai/run.md` and `sync.md` are tracked | slash commands |

Registered and kept (all probed): `moai mcp-server`, `moai init`, `moai mcp add|remove|list` (and the flags `--command`, `--args`, `--scope` that the aside row uses), `moai spec audit`, `moai verify check`, `moai verify codex-review` (and its `--project-root`), `moai goal arm|status`, `moai session list`, `moai graph build`, `moai glm tools enable`, `moai glm`, `moai todo --auto`. The Class C set above is therefore complete for `mcp-server.md` against the installed build; the page carries no other unregistered `moai` form.

Reading: the 36 Class C instructions do not resolve. The four Class N statements match what the build does.

### A.5 Per-form evidence for Q13 (36 Class C lines) and the disposition applied

Q13 is applied per form. The table lists, for each form, the registered command considered, how close it is, and the disposition that results. Closeness is a reading of the two help texts, not a run of both commands. Under the leader reading of the operator ruling 220a (decision-index.md, Q13: the cell becomes the marker and the row stays), a registered command that overlaps in part is not "the current equivalent", so it is not a permitted replacement.

| Form | Registered command considered | Reading | Disposition applied |
|------|-------------------------------|---------|---------------------|
| `moai spec list` (spec_progress, "List SPEC documents + frontmatter") | `moai spec status --list` | lists all SPECs with their status; it does not print frontmatter beyond the status, so the overlap is partial and it is not an equivalent | cell becomes `—` alone |
| `moai audit cache` (audit_cache, "plan-audit PASS cache") | none found | no `cache` verb anywhere | `—` alone |
| `moai codex task` (codex_task, "delegate a coding task to codex") | none found | `moai codex audit <role>` runs a read-only role, which is a different job | `—` alone |
| `moai codex setup` (codex_setup, "probe local codex install") | `moai codex status` | the readout reports the codex binary, CODEX_HOME, and auth provider, which overlaps the "LookPath + version + auth" probe in part; it is not an equivalent | `—` alone |
| `moai codex job status\|result\|cancel` | none found | no `job` verb | `—` alone |
| `B3["moai run SPEC-ID"]`, `B4["moai sync SPEC-ID"]` | `/moai run SPEC-ID`, `/moai sync SPEC-ID` | registered slash commands (current equivalents); the page body already uses them (en 133 to 134: `> /moai run SPEC-AUTH-001`, `> /moai sync SPEC-AUTH-001`) | replaced in place, B3 with run and B4 with sync |
| the table row of a stale cell | n/a | the other reading of "delete the line" | not planned: every row stays (REQ-PRC-001-001); a ruling to delete the rows returns the SPEC to the plan phase |

The no-CLI marker `—` alone is the page's own convention for a tool that has no CLI twin: the whole cell of 18 rows per locale holds it (`git grep -c` of the exact bare form over the page prints `:18` for each locale, E-SWD-6), and the rows `audit_multi`, `glm_task`, `jev_ask`, and the `factory_msg_*` family hold it followed by a parenthetical phrase (E-SWD-9a). The bare form is the one applied: it needs no translation, it adds no prose (the SPEC's scope names no new prose), and a check can hold it to an exact cell, which a marker followed by free text does not allow. The cost is that the reader gets the marker without an explanation; the explanation is carried by the neighbouring rows, not by the cell (R3). The acceptance cell-form checks (AC-PRC-001-003) admit exactly this form in all four locales; any other replacement form needs a plan-phase revision.

### A.6 Tool provenance

The installed `moai` reports `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z. Commit 7705b365e2 is a strict ancestor of the tree HEAD (`merge-base --is-ancestor` exit 0; `rev-list --count 74ad197380..7705b365e2` prints 0; `rev-list --count 7705b365e2..74ad197380` prints 68). The probes were therefore not made by a build of this tree, and none was built (no Go build, by the lane's machine rule). The mitigation is a source read at the pin, not a substitute: of the 68 commits, 24 files under `internal/cli` and `cmd` differ (acceptance.md E-SWD-11), none of them a `spec*.go` file or `cg.go`; one, `codex_audit_launch.go`, declares the `audit` and `role-audit` sub-verbs, and the `Use:` lines read from it at the pin match what the installed build printed. The claim cited is "installed build 7705b365e2 against tree 74ad197380"; `internal/` and `cmd/` are unchanged between 74ad197380 and the revision-0.2.0 pin 26834dd17d (A.1).

## B. Milestones (ordered by likelihood of change; mechanical steps last)

- **M1 (Priority High): close the decision gate.** Closed at revision 0.2.0: Q13 and Q17 carry the operator verdict of board record 220a, and Q13's reading, Q-SWD-1, and Q-SWD-2 carry the leader ruling of board record 4604 (decision-index.md, Gate). Nothing is open. No tracked file outside this SPEC directory changes before the plan audit passes and the Kickoff gate is met. A later ruling that picks an alternate of any row returns the SPEC to the plan phase.
- **M2 (Priority High): Mermaid labels.** Edit the two lines in each of the four `worktree/_index.md` files within their edit unit (§A.3): `B3["/moai run SPEC-ID"]` and `B4["/moai sync SPEC-ID"]`, the same in all four locales. Read the lines around B1 to B4 in each locale and keep the diagram coherent (R2). This is the user-visible, most-likely-to-be-revised change, so it comes first.
- **M3 (Priority High): CLI-equivalent cells.** Edit the seven cells in each of the four `mcp-server.md` files within their edit unit (§A.3): the last cell of each row becomes the bare marker `—`, the same in all four locales, with the first three cells of the row left byte-identical. The bare marker needs no translation, so the native-idiom rule has nothing to render here.
- **M4 (Priority High before landing): sequencing re-verification.** Re-run the probes of acceptance.md E-SWD-16 against current `main`. If the README child (t1720) or any card that edits these pages has landed and been absorbed, re-measure and re-pin the base commit used by AC-PRC-001-003 to AC-PRC-001-006 (R6) and the line numbers of AC-PRC-001-003 command 15, and re-read the surrounding lines by content.
- **M5 (Priority Medium): verification.** Run the AC commands of acceptance.md §B from the card worktree root and carry the deciding output into the verdict file (`.moai/reports/t1724/verdict.md`), not into scratch. Run `bash scripts/docs-i18n-check.sh` as the confirmatory parity check (E-SWD-14). The `hns-oss-docs-verify` recipe (warning-free site build) is the check that Mermaid still parses (R2); the plan phase ran no site build.

## C. Risks

- **R1, moving coordinates.** Line numbers on all three pages move. Of the 10 most recent commits touching `mcp-server.md`, 7 list files and each lists all four locales, and 3 are merges listing none; of the 10 most recent touching `worktree/_index.md`, 2 changed the ko copy alone (aa650dd08a, a723c41f63) and 1 changed en, ja, and zh without ko (734ede821f) (E-SWD-16). Every check keys on content or on a row id, except the position check of AC-PRC-001-003 (command 15), which pins the line numbers of the seven `mcp-server.md` rows at the measured tree and is re-measured at M4.
- **R2, Mermaid coherence.** Replacing B3 and B4 keeps the diagram's four nodes. Rendering is not machine-checked. AC-PRC-001-002 checks each of the two lines by label (B3 carries run, B4 carries sync, each line ends at the closing bracket), which catches a swap, a duplicate, and trailing text, but not a syntax slip elsewhere such as an unbalanced quote in another line. The site build in M5 is the check.
- **R3, information lost by `—`.** Replacing a CLI cell by the bare `—` tells the reader there is no CLI twin, which is true for five of the seven rows and only partly true for `spec_progress` and `codex_setup`, where a registered command overlaps in part (A.5). The leader reading of Q13 accepts that loss (decision-index.md, Q13); the cell also carries no explanatory phrase, because the exact-form check admits the marker alone.
- **R4, the umbrella's closing check.** Under the retain ruling of Q-SWD-1 the four `prompt-caching.md` lines stay, and the umbrella's closing check (E-01 exits 1 with empty stdout) and the umbrella's partial criterion for this child (zero E-01 matches on its file list) cannot hold while they stay; the GATE checker must admit the same four lines. That is an amendment outside this SPEC, the same position as the sibling children SWEEPA and SWEEPB, whose retain rulings are recorded as board records cited from their own decision-indexes (not re-read here, acceptance.md G2).
- **R5, installed-build lag.** The CLI probes rest on a build 68 commits behind the tree (A.6).
- **R6, shared pages with other cards.** The README child (t1720) shares ko files with this child per the umbrella; this pass did not scan other branches or worktrees for edits to the three pages (acceptance.md G5). The risk is a base mismatch in the `26834dd17d HEAD` ranges of AC-PRC-001-003 to AC-PRC-001-006 after an absorb, which M4 handles by re-pinning, and a merge conflict if another card edits one of the 36 lines, which a content-keyed re-read at M4 catches.
- **R7, adjacent labels left in place.** After the sweep, the same `worktree/_index.md` pages still show `moai plan`, `moai run`, and `moai sync` as diagram labels (20 lines). A reader of the swept page sees the corrected pair next to uncorrected labels. That is the leader ruling on Q-SWD-2, kept visible here so a reviewer does not read the family-pattern count of zero as "no unregistered form remains on the page".

## D. Tests named, not run

- No Go test exists or is added; no Go build or test was run in this plan phase. The checks are the commands in acceptance.md §B. Each RED-now observation, each control, and each mutant probe was run once against the pin or against a scratch fixture (acceptance.md §C). No green-path run on the repository exists yet, because nothing has been edited; the green outputs were observed on scratch fixtures only (E-SWD-15).

## E. Open questions

The decision rows Q13, Q17, Q-SWD-1, and Q-SWD-2 live in decision-index.md and are not restated here; each carries a recorded answer and none is open. No other question is open.
