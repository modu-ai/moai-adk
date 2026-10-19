# SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001 — Acceptance Criteria

## A. Conventions

### A.1 Pin, patterns, and pathspecs

Document-level pin: tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea` (short form `74ad197380`). It binds every RED-now cell, control, and evidence entry in §C that carries no pin of its own. The commands in §B judge the sweep tip, so they read `HEAD` of the card branch. Where a command compares two trees it names the pin `74ad197380` as its left end, never `main`: a moving ref advances under an absence claim (verification-completeness §4). If the card branch is rebased or absorbs `main`, the left end is re-measured and re-pinned in the same pass (plan.md M4); until then `74ad197380` is the base.

The patterns and pathspecs below are defined once and used by name. `U` is the union of the umbrella's families F1 and F2 (the card's contract for `mcp-server.md`). `UW` is `U` with `[[:space:]]+` in place of each single space; it is the mutant guard, because a stale citation respelled with two spaces passes `U` and still tells the reader to run an absent verb. `E03` and `E03W` are the same pair for the umbrella's family F3.

```
U    = (^|[^/])moai (cg|gpt|audit|codemaps|model|codex (job|task|setup)|todo (rename|pick)|spec list)([^[:alnum:]-]|$)
UW   = (^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model|codex[[:space:]]+(job|task|setup)|todo[[:space:]]+(rename|pick)|spec[[:space:]]+list)([^[:alnum:]-]|$)
E03  = (^|[^/])moai (run|sync) SPEC-ID
E03W = (^|[^/])moai[[:space:]]+(run|sync)[[:space:]]+SPEC-ID
SL   = "/moai (run|sync)([ <"]|$)
CG   = (^|[^/])moai cg([^[:alnum:]-]|$)
M5   = ^\| `mcp__moai__(audit_cache|codex_task|codex_job_status|codex_job_result|codex_job_cancel)` .*\| —[^|]*\|$
M2   = ^\| `mcp__moai__(spec_progress|codex_setup)` .*\| (—|`moai spec status --list`|`moai codex status`)[^|]*\|$
ROW  = ^\| `mcp__moai__
MCP  = 'docs-site/content/*/guides/mcp-server.md'
WT   = 'docs-site/content/*/worktree/_index.md'
PC   = 'docs-site/content/*/cost-optimization/prompt-caching.md'
```

`M5` and `M2` are the cell-form patterns of AC-PRC-001-003: the five rows with no registered candidate must end in the no-CLI marker `—`; the two rows with a partial candidate end in the marker or in the one candidate named for them. Under a Q13 verdict that names a different registered form, `M2` is amended by one alternative in the same revision as the verdict.

### A.2 Command form

- Every command runs from the card worktree root. `git` stands at the head of a plain command, never inside `$(...)` or `<(...)`, because the worktree-session guard refuses those forms. The RED-now cells in §C were executed as `git -C <worktree root> <the same arguments>`; the arguments are identical to the AC commands.
- A cited command is a single invocation. Its exit code is a separate field. The Bash tool does not display the exit status of a command that prints nothing, so each cell in §C read the status by running the same command followed by `; echo "exit=$?"`; that wrapper is the measurement harness and is not part of the cited command. The outcomes stay distinct: `0` means a match (or a difference), `1` means none, any other value means git failed and is never read as "no match".
- `git grep -c` is used instead of `git grep -q` so stdout carries the swept count; an absence is the empty stdout with exit 1.
- Commands live in fenced blocks, never in table cells, so an alternation `|` is a regex alternation.

### A.3 Classification

AC-PRC-001-001 to AC-PRC-001-004 are release-blocking and flip from red to green. AC-PRC-001-005 and AC-PRC-001-006 are preservation criteria: green at the pin by construction, so each carries a control, a command run on a known input that turns the same form red. Branch-dependent statuses are stated per branch in each criterion and summarized here.

| Decision row | Branch | Effect on the criteria |
|--------------|--------|------------------------|
| Q13, `mcp-server.md` cells | any permitted replacement (`—`, or a named candidate for its own row) | AC-001 and AC-003 hold identically; the edit is one removed and one added line per cell in every branch |
| Q13, Mermaid labels | R: replace by the slash form | AC-002 commands 1 to 3 and AC-004 (`2	2`) as stated under R |
| Q13, Mermaid labels | D: remove the two lines | AC-002 commands 1 to 3 and AC-004 (`0	2`) as stated under D |
| Q17 | default: four locales in one change | all criteria are written for it |
| Q17 | alternate: ko now, others later | not planned: AC-001 to AC-004 read four locales and would stay red; the SPEC returns to the plan phase |
| Q-SWD-1 | default: retain the four lines | AC-005 holds |
| Q-SWD-1 | alternate: treat them as stale | not planned: AC-005 is not applicable; the SPEC returns to the plan phase |
| Q-SWD-2 | default: leave the adjacent labels | AC-004 bound keeps them untouched |
| Q-SWD-2 | alternate: extend to the 20 labels | not planned: AC-004 would be red; the SPEC returns to the plan phase |

## B. Acceptance criteria

| AC | Requirement | Depends on | Class | RED-now | Why red now | Green path |
|----|-------------|------------|-------|---------|-------------|------------|
| AC-PRC-001-001 | REQ-PRC-001-001, REQ-PRC-001-004 | Q13, Q17 | release-blocking | E-SWD-2, E-SWD-3 | each of the four `mcp-server.md` files carries 7 lines that match `U` and `UW` | M3 |
| AC-PRC-001-002 | REQ-PRC-001-002, REQ-PRC-001-004 | Q13, Q17 | release-blocking | E-SWD-4 (command 1, 2); E-SWD-5 (command 3 under R, command 4 under D) | each of the four `worktree/_index.md` files carries 2 lines that match `E03`; command 3 prints nothing and command 4 prints 2 per file | M2 |
| AC-PRC-001-003 | REQ-PRC-001-001, REQ-PRC-001-005 | Q13, Q17 | release-blocking | E-SWD-6 | no edit exists, so the numstat is empty and neither cell-form pattern matches the seven rows | M3 |
| AC-PRC-001-004 | REQ-PRC-001-002, REQ-PRC-001-005 | Q13, Q17 | release-blocking | E-SWD-7 | no edit exists, so the numstat is empty | M2 |
| AC-PRC-001-005 | REQ-PRC-001-003 | Q-SWD-1 (default branch) | preservation | control E-SWD-8 | green at the pin; red only when a retirement line is deleted or a `prompt-caching.md` file is touched | M5 |
| AC-PRC-001-006 | REQ-PRC-001-005 | none | preservation | control E-SWD-9 | green at the pin; red only when a path outside the swept files changes | M5 |

### AC-PRC-001-001 — the `mcp-server.md` stale forms are gone

AC-PRC-001-001 maps REQ-PRC-001-001 and, for the four-path swept set, REQ-PRC-001-004.

- **Given** the sweep tip is `HEAD` of the card branch and Q13 and Q17 carry verdicts,
- **When** these three commands run,
- **Then** the first two print nothing and exit 1, and the third prints exactly four paths and exits 0.

```
git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model|codex (job|task|setup)|todo (rename|pick)|spec list)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model|codex[[:space:]]+(job|task|setup)|todo[[:space:]]+(rename|pick)|spec[[:space:]]+list)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git ls-files -- 'docs-site/content/*/guides/mcp-server.md'
```

RED-now at the pin 74ad197380 (E-SWD-2, E-SWD-3): commands 1 and 2 each print four lines ending `:7` and exit 0. The third command is the swept-set control: a pass over an empty set asserts nothing. This criterion does not stop a cell from being blanked or respelled to a non-registered form; AC-PRC-001-003 does.

### AC-PRC-001-002 — the `worktree/_index.md` labels are gone

AC-PRC-001-002 maps REQ-PRC-001-002 and, for the four-path swept set, REQ-PRC-001-004.

- **Given** the sweep tip and the Q13 branch for the Mermaid labels (R: replaced, or D: removed),
- **When** these four commands run,
- **Then** commands 1 and 2 print nothing and exit 1 under both branches; command 3 prints four lines ending `:2` and exits 0 under R, and prints nothing and exits 1 under D; command 4 prints four lines ending `:2` and exits 0 under R, and prints nothing and exits 1 under D.

```
git grep -c -E "(^|[^/])moai (run|sync) SPEC-ID" HEAD -- 'docs-site/content/*/worktree/_index.md'
git grep -c -E "(^|[^/])moai[[:space:]]+(run|sync)[[:space:]]+SPEC-ID" HEAD -- 'docs-site/content/*/worktree/_index.md'
git grep -c -E "\"/moai (run|sync)([ <\"]|$)" HEAD -- 'docs-site/content/*/worktree/_index.md'
git grep -c -w -E "B[34]" HEAD -- 'docs-site/content/*/worktree/_index.md'
```

Status at the pin, per command and branch. Commands 1 and 2 (E-SWD-4): four lines `:2`, exit 0, so red under both branches. Command 3 (E-SWD-5): empty, exit 1; that is red under R (R expects `:2`) and green under D (D expects empty), so under D it is a preservation check. Command 4 (E-SWD-5): four lines `:2`, exit 0; that is green under R (R expects `:2`, the references survive) and red under D (D expects empty). The flip is M2 and it is the same milestone under both branches. Command 4 is the dangling-reference check: a half-removed pair leaves one `B3` or `B4` line in a file and fails under both branches.

### AC-PRC-001-003 — the seven `mcp-server.md` cells change, and only they

AC-PRC-001-003 maps REQ-PRC-001-001 and REQ-PRC-001-005. The edit unit and its bound are the table in plan.md §A.3; this criterion measures them for all four files, so a page cannot be emptied, a cell cannot be blanked, and a replacement outside the permitted set cannot pass.

- **Given** the sweep tip and the Q13 verdicts for the seven rows,
- **When** these four commands run,
- **Then** command 1 prints four numstat lines, one per locale, each `7` added and `7` removed; command 2 prints four lines ending `:5`; command 3 prints four lines ending `:2`; command 4 prints four lines ending `:41`.

```
git diff --numstat 74ad197380 HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__(audit_cache|codex_task|codex_job_status|codex_job_result|codex_job_cancel)\` .*\| —[^|]*\|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__(spec_progress|codex_setup)\` .*\| (—|\`moai spec status --list\`|\`moai codex status\`)[^|]*\|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
```

RED-now at the pin (E-SWD-6): command 1 prints nothing, exit 0; commands 2 and 3 print nothing, exit 1; command 4 prints four lines `:41`, exit 0, so it is green at the pin and is the preservation part of this criterion (the tool table keeps all its rows). Why red: no edit has been made. Mutants this closes, each observed on a scratch fixture (E-SWD-15): blanking a cell leaves 4 of 5 rows in command 2; a typo such as `moai spec lst` leaves 1 of 2 rows in command 3 and passes AC-PRC-001-001; blanking the page or deleting a row breaks command 1 or command 4; editing three locales instead of four leaves a file with no numstat line. A replacement that is a registered form but the wrong one for its row (for example `moai codex status` on the `audit_cache` row) passes every command here; the reviewer reads the seven cells against plan.md §A.5.

### AC-PRC-001-004 — the Mermaid labels change within their edit unit

AC-PRC-001-004 maps REQ-PRC-001-002 and REQ-PRC-001-005.

- **Given** the sweep tip and the Q13 branch for the Mermaid labels,
- **When** this command runs,
- **Then** under R it prints four numstat lines, one per locale, each `2` added and `2` removed; under D it prints four lines, each `0` added and `2` removed.

```
git diff --numstat 74ad197380 HEAD -- 'docs-site/content/*/worktree/_index.md'
```

RED-now at the pin (E-SWD-7): empty stdout, exit 0, because no edit has been made; red under both branches. The bound also keeps the five adjacent labels per locale (plan.md §A.2) untouched: an edit to any of them adds a changed line and the row reads `3	3` or higher.

### AC-PRC-001-005 — the four retirement statements stay

AC-PRC-001-005 maps REQ-PRC-001-003 and holds on the default branch of Q-SWD-1 (retain). On the alternate branch it is not applicable and the SPEC returns to the plan phase.

- **Given** the sweep tip and Q-SWD-1 carries the retain verdict,
- **When** these two commands run,
- **Then** command 1 prints four lines ending `:1` and exits 0, and command 2 prints nothing and exits 0.

```
git grep -c -E "(^|[^/])moai cg([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
git diff --numstat 74ad197380 HEAD -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
```

Green at the pin (E-SWD-8). Controls: command 2 over the range of commit c5620da649 prints four numstat rows (`2 0`, `1 1`, `2 1`, `1 1`), so the form does print when these files change; command 1 run on a scratch copy of the en page with its sentence deleted prints nothing and exits 1 (E-SWD-15). Command 2 compares with the pin, so it reads red if another card's edit to these files is absorbed into the branch; that case is the re-pin of plan.md M4, not a failure of this criterion.

### AC-PRC-001-006 — nothing outside the swept files changes

AC-PRC-001-006 maps REQ-PRC-001-005.

- **Given** the sweep tip,
- **When** this command runs,
- **Then** it prints nothing and exits 0.

```
git diff --name-only 74ad197380 HEAD -- . ':!docs-site/content/*/guides/mcp-server.md' ':!docs-site/content/*/worktree/_index.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001' ':!.moai/reports/t1724'
```

Green at the pin (E-SWD-9). The four `prompt-caching.md` files are deliberately not excluded, so a change to them prints. Control: the same exclusion form over the range of commit 0de4cfabe2 prints one path, `.moai/specs/SPEC-ASIDE-BROWSER-001/progress.md`, so the form prints when something outside the list changes.

## C. Evidence ledger

Each entry gives the command, the observed output (verbatim unless marked), the exit code as its own field, and the tree it read. Entries that pin no tree of their own read the document-level pin 74ad197380.

### E-SWD-0 — pin and tool provenance

```
git rev-parse HEAD                                 -> 74ad197380a0711cdb9816d4e37c94d29af730ea
git rev-parse main                                 -> 74ad197380a0711cdb9816d4e37c94d29af730ea
moai version (version line)                        -> moai_cp/20261007_195927   moai_cp/20261007_195927-1888-g7705b365e2   built 2026-10-18T10:45:08Z
git merge-base --is-ancestor 7705b365e2 74ad197380 -> (no stdout)   exit 0
git rev-list --count 74ad197380..7705b365e2        -> 0
git rev-list --count 7705b365e2..74ad197380        -> 68
SPEC id check  ID="SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL   -> PASS
```

The installed build is a strict ancestor of the tree HEAD, 68 commits behind. No Go build was made; CLI claims in this SPEC are attributed to "installed build 7705b365e2 against tree 74ad197380" (plan.md §A.6).

### E-SWD-1 — the three pathspecs expand to twelve paths

Command: `git ls-files -- 'docs-site/content/*/cost-optimization/prompt-caching.md' 'docs-site/content/*/guides/mcp-server.md' 'docs-site/content/*/worktree/_index.md'`. Exit 0. Output, 12 paths:

```
docs-site/content/en/cost-optimization/prompt-caching.md
docs-site/content/en/guides/mcp-server.md
docs-site/content/en/worktree/_index.md
docs-site/content/ja/cost-optimization/prompt-caching.md
docs-site/content/ja/guides/mcp-server.md
docs-site/content/ja/worktree/_index.md
docs-site/content/ko/cost-optimization/prompt-caching.md
docs-site/content/ko/guides/mcp-server.md
docs-site/content/ko/worktree/_index.md
docs-site/content/zh/cost-optimization/prompt-caching.md
docs-site/content/zh/guides/mcp-server.md
docs-site/content/zh/worktree/_index.md
```

### E-SWD-2 — RED-now, pattern U over `mcp-server.md` (AC-PRC-001-001, command 1)

Command: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model|codex (job|task|setup)|todo (rename|pick)|spec list)([^[:alnum:]-]|$)" 74ad197380 -- 'docs-site/content/*/guides/mcp-server.md'`. Pin 74ad197380. Exit code: 0. Stdout:

```
74ad197380:docs-site/content/en/guides/mcp-server.md:7
74ad197380:docs-site/content/ja/guides/mcp-server.md:7
74ad197380:docs-site/content/ko/guides/mcp-server.md:7
74ad197380:docs-site/content/zh/guides/mcp-server.md:7
```

Why red: each of the four files carries seven lines that match `U`, the seven rows of plan.md §A.2. A green path exists because the sweep edits exactly those cells (M3).

### E-SWD-3 — RED-now, pattern UW over `mcp-server.md` (AC-PRC-001-001, command 2)

Command: the same as E-SWD-2 with `UW` in place of `U`. Pin 74ad197380. Exit code: 0. Stdout: the same four lines, each ending `:7`. The whitespace guard adds no match today and closes only the respelling route.

### E-SWD-4 — RED-now, patterns E03 and E03W over `worktree/_index.md` (AC-PRC-001-002, commands 1 and 2)

Command 1: `git grep -c -E "(^|[^/])moai (run|sync) SPEC-ID" 74ad197380 -- 'docs-site/content/*/worktree/_index.md'`. Pin 74ad197380. Exit code: 0. Stdout:

```
74ad197380:docs-site/content/en/worktree/_index.md:2
74ad197380:docs-site/content/ja/worktree/_index.md:2
74ad197380:docs-site/content/ko/worktree/_index.md:2
74ad197380:docs-site/content/zh/worktree/_index.md:2
```

Command 2 (`E03W`): exit code 0, stdout the same four lines. Why red: each file carries the two Mermaid label lines B3 and B4 of plan.md §A.2. A green path exists because the sweep replaces or removes exactly those lines (M2).

### E-SWD-5 — AC-PRC-001-002, commands 3 and 4

Command 3: `git grep -c -E "\"/moai (run|sync)([ <\"]|$)" 74ad197380 -- 'docs-site/content/*/worktree/_index.md'`. Pin 74ad197380. Exit code: 1. Stdout: empty. Reading: red under branch R (the replaced labels do not exist yet), green and preserved under branch D.

Command 4: `git grep -c -w -E "B[34]" 74ad197380 -- 'docs-site/content/*/worktree/_index.md'`. Pin 74ad197380. Exit code: 0. Stdout: four lines, `74ad197380:docs-site/content/{en,ja,ko,zh}/worktree/_index.md:2`. Reading: green and preserved under R, red under D (the references must disappear with the lines).

Both commands were also run on scratch fixtures of the ko page, one with the labels replaced and one with them removed (E-SWD-15): the replaced fixture prints `2` for both, the removed fixture prints nothing for both.

### E-SWD-6 — RED-now, AC-PRC-001-003

```
git diff --numstat 74ad197380 HEAD -- 'docs-site/content/*/guides/mcp-server.md'      -> (no stdout)   exit 0   (HEAD == 74ad197380 at measurement)
git grep -c -E "<M5>" 74ad197380 -- 'docs-site/content/*/guides/mcp-server.md'        -> (no stdout)   exit 1
git grep -c -E "<M2>" 74ad197380 -- 'docs-site/content/*/guides/mcp-server.md'        -> (no stdout)   exit 1
git grep -c -E "^\| \`mcp__moai__" 74ad197380 -- 'docs-site/content/*/guides/mcp-server.md'   -> {en,ja,ko,zh}/guides/mcp-server.md:41   exit 0
```

(`<M5>` and `<M2>` stand for the patterns of §A.1, run in the exact spelling of AC-PRC-001-003.) Why red: the first command is empty because nothing is edited, and the two cell-form patterns match no target row because the seven cells still hold the stale forms. The seven target rows exist: `git grep -c -E "^\| \`mcp__moai__(spec_progress|audit_cache|codex_task|codex_setup|codex_job_status|codex_job_result|codex_job_cancel)\` "` prints `:7` for each locale, exit 0, the same count as E-SWD-2, so each `U` hit is one of the seven rows. Positive control for the marker form on the repository: `git grep -c -E "^\| \`mcp__moai__(glm_task|audit_multi)\` .*\| —[^|]*\|$" 74ad197380 -- 'docs-site/content/*/guides/mcp-server.md'` prints `:2` for each locale, exit 0, so the pattern matches the rows that already carry the marker.

Control for the numstat bound: `git diff --numstat f32c831de5~1 f32c831de5 -- 'docs-site/content/*/guides/mcp-server.md'` prints `1	1` for each of the four paths, below the `7	7` bound.

### E-SWD-7 — RED-now, AC-PRC-001-004

Command: `git diff --numstat 74ad197380 HEAD -- 'docs-site/content/*/worktree/_index.md'`. Pin 74ad197380 (HEAD equals it at measurement). Exit code: 0. Stdout: empty. Why red: no edit has been made. Control for the bound: `git diff --numstat ce79ef7caf~1 ce79ef7caf -- 'docs-site/content/*/worktree/_index.md'` prints

```
5	35	docs-site/content/en/worktree/_index.md
6	36	docs-site/content/ja/worktree/_index.md
7	38	docs-site/content/ko/worktree/_index.md
6	35	docs-site/content/zh/worktree/_index.md
```

which is above the `2`-removed bound.

### E-SWD-8 — AC-PRC-001-005 at the pin, and its controls

```
git grep -c -E "(^|[^/])moai cg([^[:alnum:]-]|$)" 74ad197380 -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
  -> {en,ja,ko,zh}/cost-optimization/prompt-caching.md:1   exit 0
git diff --numstat 74ad197380 HEAD -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
  -> (no stdout)   exit 0
git diff --numstat c5620da649~1 c5620da649 -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
  -> 2 0 en, 1 1 ja, 2 1 ko, 1 1 zh   exit 0
git diff --quiet c5620da649~1 c5620da649 -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
  -> (no stdout)   exit 1
```

### E-SWD-9 — AC-PRC-001-006 at the pin, and its control

Command: `git diff --name-only 74ad197380 HEAD -- . ':!docs-site/content/*/guides/mcp-server.md' ':!docs-site/content/*/worktree/_index.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001' ':!.moai/reports/t1724'`. Pin 74ad197380 (HEAD equals it at measurement). Exit code: 0. Stdout: empty. Control: the same form with `0de4cfabe2~1` and `0de4cfabe2` as the two ends prints `.moai/specs/SPEC-ASIDE-BROWSER-001/progress.md`, exit 0.

E-SWD-9a, the no-CLI marker convention: `git grep -n -E "— \(" 74ad197380 -- 'docs-site/content/*/guides/mcp-server.md'` prints, in the en copy, line 159 (`audit_multi`, `— (MCP-only convergence entry point)`), 204 (`glm_task`, `— (no such CLI)`), 226 (`jev_ask`, `— (MCP-only)`), and 234 to 238 (`factory_msg_*`, `— (MCP-only)`); the ko copy has the same lines, with 204 reading `— (해당 CLI 없음)`.

### E-SWD-10 — CLI surface, installed build 7705b365e2 against tree 74ad197380

E-SWD-10 and E-SWD-11 are corroboration for the classification in plan.md §A.4, not RED-now cells: their commands and outputs are abridged to the deciding lines, and no release-blocking criterion rests on them alone.

```
moai --help (COMMANDS, LAUNCH, PROJECT, TOOLS)  -> lists cc, glm, codex, init, mcp-server, mcp, spec, verify, goal, session, graph, todo, worktree, plan; lists no audit, run, sync, model, codemaps, gpt
moai cg --help          -> exit 1   moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration
moai audit --help       -> exit 1   Unknown command "audit" for "moai".
moai audit cache --help -> exit 1   Unknown command "audit" for "moai".
moai run --help         -> exit 1   Unknown command "run" for "moai".
moai sync --help        -> exit 1   Unknown command "sync" for "moai"
moai spec --help        -> exit 0   COMMANDS: status, drift, view, lint, close, audit, archive, ceiling
moai spec status --help -> exit 0   flag --list  List all SPECs and their status
moai codex --help       -> exit 0   COMMANDS: audit <role>, role-audit <role>; forms: cli | status | app
moai mcp --help         -> exit 0   COMMANDS: add, remove, list
moai verify --help      -> exit 0   COMMANDS: record, check, run, audit-plan, codex-review, sync-gate
moai goal --help        -> exit 0   verbs: arm, status, clear, render
moai session --help     -> exit 0   COMMANDS: register, heartbeat, deregister, list, purge, current, doctor
moai graph --help       -> exit 0   COMMANDS: build, query, check, stamp codemaps, report
moai glm tools --help   -> exit 0   enable [vision|websearch|webreader|all], disable [...]
moai mcp add --help     -> exit 0   flags: --args, --command, --env, --headers, --scope, --type, --url
moai todo --help        -> exit 0   flag --auto listed
moai plan --help        -> exit 0   "Plan-phase CLI utilities"; "distinct from the '/moai plan' slash-command skill"
```

Caution recorded with the evidence: `moai spec list --help`, `moai codex task --help`, `moai codex setup --help`, and `moai codex job status --help` each exit 0 and print the parent help, so those probes establish nothing about the verb; existence is read from the parent's command list and from E-SWD-11. The verbs themselves were not executed, because `moai codex task` would launch Codex.

### E-SWD-11 — source corroboration and provenance diff

```
git grep -n -E 'Use:...("|`)' 74ad197380 -- internal/cli/spec*.go   -> spec.go:13 "spec"; spec_archive.go:35 "archive"; spec_audit.go:64 "audit"; spec_ceiling.go:35 "ceiling <SPEC-ID> ..."; spec_close.go:71 "close SPEC-ID"; spec_drift.go:80 "drift"; spec_lint.go:35 "lint ..."; spec_status.go:29 "status <SPEC-ID> <new-status>"; spec_view.go:20 "view <SPEC-ID>"   (no "list")
git grep -n -E 'Use:...' 74ad197380 -- internal/cli/codex_audit_launch.go   -> :880 "audit <role>"; :907 "role-audit <role>"
git grep -n -E 'Use:\s*("|`)(cache|task|job|setup)' 74ad197380 -- internal cmd   -> internal/cli/glm.go:125 "setup [api-key]"; internal/cli/profile_setup.go:199 "setup [name]"; internal/cli/init_update_profile_path_test.go:278 (a test fixture)   (no cache, task, or job)
git grep -n -E 'Use:|Hidden:|Short:' 74ad197380 -- internal/cli/cg.go   -> :13 var cgCmd = &cobra.Command{Use: "cg", Short: "Retired CG launcher", Hidden: true, ...}
git diff --name-only 7705b365e2 74ad197380 -- internal/cli cmd   -> 24 files; none is a spec*.go file or cg.go; includes codex_audit_launch.go, codex_audit_leg.go, mcp_claude.go, mcp_codex.go, mcp_convergence.go, mcp_factory_card.go, factory_card.go, factory_card_pr.go, integration.go, integration_candidate.go, ci_verdict.go and test files
git ls-files -- .claude/commands/moai/run.md .claude/commands/moai/sync.md .claude/commands/moai/plan.md   -> all three paths
```

### E-SWD-12 — what else the pages say, and the silent families

`git grep -n -E "moai" 74ad197380 -- 'docs-site/content/*/cost-optimization/prompt-caching.md'` prints exactly four lines, en 268, ja 171, ko 243, zh 171, each the retirement sentence ("`moai cg` has been retired ... not an alias for `moai cc`"), exit 0. `git grep -c -E "/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)"` over the twelve files prints nothing, exit 1; the control `git grep -c -E "/moai[: ](plan|run|sync|design|issue)([^[:alnum:]-]|$)" 74ad197380 -- 'docs-site/content/*/worktree/_index.md'` prints `:7` for each locale, exit 0, so the slash-form engine fires where such forms occur. The union `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model|codex (job|task|setup)|todo (rename|pick)|spec list|(run|sync) SPEC-ID)([^[:alnum:]-]|$)" 74ad197380 -- <the three pathspecs>` prints twelve lines: `:1` for each `prompt-caching.md`, `:7` for each `mcp-server.md`, `:2` for each `worktree/_index.md`; 4 + 28 + 8 = 40 lines; exit 0. Per-family counts at the pin: E-01 over the twelve files prints `:1` for each `prompt-caching.md` and each `mcp-server.md`; E-02 prints `:6` for each `mcp-server.md`; E-03 prints `:2` for each `worktree/_index.md`.

### E-SWD-13 — the adjacent labels

Command: `git grep -c -E "\"moai (plan|run|sync)(<br/>|\")" 74ad197380 -- 'docs-site/content/*/cost-optimization/prompt-caching.md' 'docs-site/content/*/guides/mcp-server.md' 'docs-site/content/*/worktree/_index.md'`. Exit 0. Stdout: `:5` for each of the four `worktree/_index.md` files and nothing for the other eight. The five per locale are listed with their lines in plan.md §A.2.

### E-SWD-14 — parity baseline

Command: `bash scripts/docs-i18n-check.sh` at the pin; exit 0; the output was redirected to a scratch file and read back. Tail:

```
=== Summary ===
Errors:   0
Warnings: 0
OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.
```

File counts printed: ko 153, en 153, ja 153, zh 153. The check passes at the pin, so on its own it cannot go red for this change; it is a confirmatory check at M5, not a criterion.

### E-SWD-15 — mutant and green fixtures (scratch copies, not repository trees)

Fixtures were made with `sed` from the en `mcp-server.md` and the ko `worktree/_index.md` into the session scratch directory and read with `rg -c`. They are inputs constructed to be green or to be a mutant; they do not stand for the sweep.

| Fixture | Construction | Result |
|---------|--------------|--------|
| green (mcp) | the seven cells replaced by `— (no such CLI)` | `U` prints nothing, exit 1; `M5` prints 5; `M2` prints 2; `ROW` prints 41 |
| blank cell | green, with the `audit_cache` cell emptied | `M5` prints 4 |
| typo | green, with the `spec_progress` cell set to `` `moai spec lst` `` | `U` prints nothing (passes AC-PRC-001-001); `M2` prints 1 |
| worktree R | the ko labels rewritten to `"/moai run SPEC-ID"` and `"/moai sync SPEC-ID"` | `E03` prints nothing; `SL` prints 2; `B[34]` prints 2 |
| worktree D | the two ko lines deleted | `E03` prints nothing; `SL` prints nothing; `B[34]` prints nothing; `diff` against the ko page shows exactly lines 119 and 120 removed |
| retire line deleted | the en `prompt-caching.md` with its `moai cg` sentence deleted | `CG` prints nothing, exit 1 |

### E-SWD-16 — sequencing and churn

`git log -n 10 --format=%h --name-only` over `mcp-server.md`: 7 commits list files and each lists all four locales (f32c831de5, c932aa3582, d94e4763dd, 5629a0d6f8, 0de4cfabe2, 5c006bf9b2, 4a8ab71a79); 3 are merges listing none. Over `worktree/_index.md`: 7 commits list all four locales (ce79ef7caf, 024536e38c, b50c4de719, 5e2c832386, e55b48def9, 47f5bfde91, 5792fc755a), aa650dd08a and a723c41f63 list only ko, and 734ede821f lists en, ja, and zh. Over `prompt-caching.md`, of the 8 most recent commits 6 list files (c5620da649, 4ccb12e2cc, fbd13bbfc4, 81c10109ee, a29db12ec9, ce79ef7caf), each listing all four locales, and 2 are merges listing none. The umbrella records (spec.md lines 122 and 126) that the README child shares ko files with this child on different lines and that this child lands after it; no source in this pass resolves the card text's "card 4" to an id (G5).

### E-SWD-17 — decision-board check

Command: `moai decision read --scope card:t1724 --project-root /Users/goos/MoAI/moai-adk-go`, output saved to a scratch file; exit 0. `rg -c "Q13|Q17|c26e|7f98"` on it prints `1`, and the one line is a record about card t1721 (a direct ledger restamp and the operator's answer), not an answer to Q13, Q17, or the Class N question. `rg -c "id=d-20261018T170912Z-14f9|record"` prints `202`, so the file was read; that second count is a weak liveness check (the word "record" is common), and no known record id for this scope was available as a stronger control. No board record answers Q13 or Q17 for this card.

## D. Gaps and residual risk

**Gaps (not observed in this pass):**

- **G1: the card's named source is absent.** The card cites `.moai/worktrees/t1655/.moai/reports/t1655/child-cards.md`; neither that path nor a hoisted copy exists (the lane's progress record, "Source read attempted and failed"). The E-01, E-02, and E-03 sets were derived from the files and cross-checked against the umbrella's committed counts (24 CLI lines for `mcp-server.md`, 8 Mermaid lines; umbrella acceptance.md E-02 and E-03 at its pin 8816de4dae).
- **G2: the Class N rulings are cited second-hand.** The retain rulings for the sibling children (leader ruling `d-20261018T201957Z-c26e` for SWEEPB Q-SWB-1, and `d-20261018T223420Z-7f98` on the start rule) are read from the SWEEPB decision-index in another worktree, not queried on the board; whether they apply to this card's four lines is not recorded on the board for this scope (E-SWD-17). They are context for the default of Q-SWD-1, not authority.
- **G3: forms outside the families.** `worktree/_index.md` and `prompt-caching.md` cite other `moai` forms (`moai worktree done|sync|remove|clean|recover|snapshot|verify|restore`, `moai cc -w`, `moai glm -w`, with flags such as `--spawn`, `--strategy rebase`, `--merged-only`, `--delete-branch`). Only the top-level listing of `moai --help` was compared with them; their sub-verbs and flags were not probed, and no claim is made about them.
- **G4: no build of the tree.** The CLI probes ran on an installed build 68 commits behind the tree (E-SWD-0). The source read at the pin corroborates the registrations but is not a run of the tree's binary.
- **G5: other cards.** No other branch or worktree was scanned for edits to the three pages, and the card text's "card 4" was not resolved to an id.
- **G6: the i18n rules document.** The umbrella cites `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 for the four-locale rule. This pass did not re-read that document.
- **G7: Mermaid rendering.** No site build was run; the fixtures show the text outcome of a replacement and of a removal, not that the diagram renders.
- **G8: verbs not executed.** `moai codex task`, `setup`, and `job` were judged absent from the listing and the source, not by running them, because running `moai codex task` would launch Codex (E-SWD-10).
- **G9: GATE card state.** The lane's brief says card t1717 is on plan-audit hold. That state was not re-measured, and nothing in this SPEC depends on it.
- **G10: no green run on the repository.** Nothing has been edited, so every green output in this file other than the preservation criteria comes from a scratch fixture (E-SWD-15).

**Residual risk:** a green `U`, `UW`, `E03`, and `E03W` over the swept files proves the matched spellings are gone, not that the page still reads coherently (plan.md R2, R7), that a respelling outside both patterns is absent, or that a replacement that is a registered form is the right form for its row (AC-PRC-001-003 narrows the second set to the marker or the one named candidate, and leaves the pairing of candidate to row to the reviewer). The reviewer reads the edited lines in all four locales.

## E. Definition of Done

- Q13, Q17, Q-SWD-1, and Q-SWD-2 each carry an operator verdict or an explicit leader ruling before the first file edit (the gate set is stated in decision-index.md, Gate).
- AC-PRC-001-001 to AC-PRC-001-004 print their expected output at the sweep tip, in the Q13 branch the verdict chose (R or D for the Mermaid labels), and their results are carried into `.moai/reports/t1724/verdict.md` with the command, the stdout, and the exit code.
- AC-PRC-001-005 and AC-PRC-001-006 print their expected output at the sweep tip. AC-PRC-001-005 applies on the retain branch of Q-SWD-1; on the alternate branch the SPEC has returned to the plan phase and this item does not apply.
- `bash scripts/docs-i18n-check.sh` still prints `Errors:   0` and `Warnings: 0`.
- The umbrella's closing measurement is not part of this Definition of Done; it is the umbrella's M6.
