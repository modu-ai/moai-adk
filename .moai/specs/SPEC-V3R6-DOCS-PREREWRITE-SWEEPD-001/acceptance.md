# SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001 — Acceptance Criteria

## A. Conventions

### A.1 Pin, patterns, and pathspecs

Document-level pin: tree commit `26834dd17d3eb9597cde0366a5d4f01130417172` (short form `26834dd17d`). It binds every RED-now cell, control, and evidence entry in §C that carries no pin of its own. Revision 0.1.0 was pinned to `74ad197380a0711cdb9816d4e37c94d29af730ea`; the docs-site trees of the two commits are identical (E-SWD-0), and every cell that reads docs-site content was re-run on 26834dd17d for revision 0.2.0. The commands in §B judge the sweep tip, so they read `HEAD` of the card branch. Where a command compares two trees it names the pin `26834dd17d` as its left end, never `main`: a moving ref advances under an absence claim (verification-completeness §4). If the card branch is rebased or absorbs `main`, the left end is re-measured and re-pinned in the same pass (plan.md M4); until then `26834dd17d` is the base.

The patterns and pathspecs below are defined once and used by name. `U` is the union of the umbrella's families F1 and F2 (the card's contract for `mcp-server.md`). `UW` is `U` with `[[:space:]]+` in place of each single space; it is the mutant guard, because a stale citation respelled with two spaces passes `U` and still tells the reader to run an absent verb. `E03` and `E03W` are the same pair for the umbrella's family F3.

```
U    = (^|[^/])moai (cg|gpt|audit|codemaps|model|codex (job|task|setup)|todo (rename|pick)|spec list)([^[:alnum:]-]|$)
UW   = (^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model|codex[[:space:]]+(job|task|setup)|todo[[:space:]]+(rename|pick)|spec[[:space:]]+list)([^[:alnum:]-]|$)
E03  = (^|[^/])moai (run|sync) SPEC-ID
E03W = (^|[^/])moai[[:space:]]+(run|sync)[[:space:]]+SPEC-ID
SL   = "/moai (run|sync)([ <"]|$)
CG   = (^|[^/])moai cg([^[:alnum:]-]|$)
M    = ^\| `mcp__moai__(spec_progress|audit_cache|codex_task|codex_setup|codex_job_status|codex_job_result|codex_job_cancel)` \| [^|]+ \| [^|]+ \| — \|$
MID  = M with one tool id in place of the alternation (the per-row form, one command per id)
MC   = ^\| `mcp__moai__(glm_task|audit_multi)` \| [^|]+ \| [^|]+ \| —[^|]*\|$
ROW  = ^\| `mcp__moai__
B3L  = ^[[:space:]]+B2 --> B3\["/moai run SPEC-ID"\]$
B4L  = ^[[:space:]]+B3 --> B4\["/moai sync SPEC-ID"\]$
MCP  = 'docs-site/content/*/guides/mcp-server.md'
WT   = 'docs-site/content/*/worktree/_index.md'
PC   = 'docs-site/content/*/cost-optimization/prompt-caching.md'
```

`M` is the exact cell form of AC-PRC-001-003. A line matches only when it is a table row of one of the seven tools, its second and third cells are non-empty and hold no pipe, and its fourth cell is the marker `—` alone between single spaces, with the line ending at the closing pipe. Nothing can follow the marker: a cell that ends `— ; run ...`, a blank cell, a registered command, and a typo all fail it. `M` says nothing about the text of the second and third cells; the literal-prefix commands of AC-PRC-001-003 hold those. `MC` is the control: the same row prefix followed by the marker and a free tail, which matches the two rows that already carry the marker with a parenthetical. Only the bare marker is permitted as the replacement (decision-index.md, Q13; plan.md §A.5); a different replacement form, or a candidate for one row, needs a plan-phase revision, not an amendment of `M`.

### A.2 Command form

- Every command runs from the card worktree root. `git` stands at the head of a plain command, never inside `$(...)` or `<(...)`, because the worktree-session guard refuses those forms. The RED-now cells in §C were executed as `git -C <worktree root> <the same arguments>`, with one stated difference: a cell reads the pin `26834dd17d` where the AC command reads `HEAD`, and a command that compares two trees names the pin as its left end in both. The cell's stdout prefix is therefore `26834dd17d:` where the AC command prints `HEAD:`; the counts are the same while HEAD equals the pin. The literal-prefix commands of AC-PRC-001-003 are the exception: their cell is the AC command itself, run with `HEAD`, which equalled the pin at measurement (E-SWD-0).
- A cited command is a single invocation. Its exit code is a separate field. The Bash tool does not display the exit status of a command that prints nothing, so each cell in §C read the status by running the same command followed by `; echo "exit=$?"`; that wrapper is the measurement harness and is not part of the cited command. The outcomes stay distinct: `0` means a match (or a difference), `1` means none, any other value means git failed and is never read as "no match".
- `git grep -c` is used instead of `git grep -q` so stdout carries the swept count; an absence is the empty stdout with exit 1.
- Commands live in fenced blocks, never in table cells, so an alternation `|` is a regex alternation.

### A.3 Classification

AC-PRC-001-001 to AC-PRC-001-004 are release-blocking and flip from red to green. AC-PRC-001-005 and AC-PRC-001-006 are preservation criteria: green at the pin by construction, so each carries a control, a command run on a known input that turns the same form red. AC-PRC-001-003 and AC-PRC-001-002 each hold a preservation part inside the release-blocking criterion (the row count, the literal-prefix commands, the row-position command, and the `B[34]` count are green at the pin and must stay green); the criterion as a whole is red at the pin and each preservation part has a mutant that turns it red (E-SWD-15). Each decision row has one planned branch, stated in the table; the alternates are not planned.

| Decision row | Planned branch | Effect on the criteria |
|--------------|----------------|------------------------|
| Q13, `mcp-server.md` cells | the bare marker `—` alone, row kept (operator verdict 220a, leader reading 4604) | AC-001 and AC-003 are written for it; the edit is one removed and one added line per cell |
| Q13, `mcp-server.md` cells | alternate: delete the table row, or a replacement other than the marker | not planned: AC-003 keeps all 41 rows and admits only the marker; the SPEC returns to the plan phase |
| Q13, Mermaid labels | R: replace by the slash forms (the slash commands exist; the leader confirmed R in the revision brief) | AC-002 and AC-004 (`2	2`) are written for it |
| Q13, Mermaid labels | alternate D: remove the two lines | not planned: AC-002 commands 3 to 6 would stay red; the SPEC returns to the plan phase |
| Q17 | four locales in one change (operator verdict 220a) | all criteria are written for it |
| Q17 | alternate: ko now, others later | not planned: AC-001 to AC-004 read four locales and would stay red; the SPEC returns to the plan phase |
| Q-SWD-1 | retain the four lines (leader ruling 4604) | AC-005 holds |
| Q-SWD-1 | alternate: treat them as stale | not planned: AC-005 is not applicable; the SPEC returns to the plan phase |
| Q-SWD-2 | leave the adjacent labels (leader ruling 4604) | AC-004 bound keeps them untouched |
| Q-SWD-2 | alternate: extend to the 20 labels | not planned: AC-004 would be red; the SPEC returns to the plan phase |

## B. Acceptance criteria

| AC | Requirement | Depends on | Class | RED-now | Why red now | Green path |
|----|-------------|------------|-------|---------|-------------|------------|
| AC-PRC-001-001 | REQ-PRC-001-001, REQ-PRC-001-004 | Q13, Q17 | release-blocking | E-SWD-2, E-SWD-3 | each of the four `mcp-server.md` files carries 7 lines that match `U` and `UW` | M3 |
| AC-PRC-001-002 | REQ-PRC-001-002, REQ-PRC-001-004 | Q13, Q17 | release-blocking | E-SWD-4 (commands 1, 2); E-SWD-5 (commands 3, 5, 6 red; command 4 preserved) | each of the four `worktree/_index.md` files carries 2 lines that match `E03`; no file carries the slash forms on B3 and B4, so commands 3, 5, and 6 print nothing | M2 |
| AC-PRC-001-003 | REQ-PRC-001-001, REQ-PRC-001-004, REQ-PRC-001-005 | Q13, Q17 | release-blocking | E-SWD-6, E-SWD-6a, E-SWD-6b, E-SWD-6c | no edit exists, so the numstat is empty and the exact-form patterns match none of the seven rows; the row count, the literal prefixes, and the row positions are green and preserved | M3 |
| AC-PRC-001-004 | REQ-PRC-001-002, REQ-PRC-001-005 | Q13, Q17 | release-blocking | E-SWD-7 | no edit exists, so the numstat is empty | M2 |
| AC-PRC-001-005 | REQ-PRC-001-003 | Q-SWD-1 (retain) | preservation | control E-SWD-8 | green at the pin; red only when a retirement line is deleted or a `prompt-caching.md` file is touched | M5 |
| AC-PRC-001-006 | REQ-PRC-001-005 | none | preservation | control E-SWD-9 | green at the pin; red only when a path outside the swept files changes | M5 |

### AC-PRC-001-001 — the `mcp-server.md` stale forms are gone

AC-PRC-001-001 maps REQ-PRC-001-001 and, for the four-path swept set, REQ-PRC-001-004.

- **Given** the sweep tip is `HEAD` of the card branch and Q13 and Q17 carry their recorded answers,
- **When** these three commands run,
- **Then** the first two print nothing and exit 1, and the third prints exactly four paths and exits 0.

```
git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model|codex (job|task|setup)|todo (rename|pick)|spec list)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model|codex[[:space:]]+(job|task|setup)|todo[[:space:]]+(rename|pick)|spec[[:space:]]+list)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git ls-files -- 'docs-site/content/*/guides/mcp-server.md'
```

RED-now at the pin 26834dd17d (E-SWD-2, E-SWD-3): commands 1 and 2 each print four lines ending `:7` and exit 0. The third command is the swept-set control: a pass over an empty set asserts nothing. This criterion does not stop a cell from being blanked, respelled to a non-registered form, or followed by text that is not a stale form; AC-PRC-001-003 does.

### AC-PRC-001-002 — the `worktree/_index.md` labels are gone and carry the right commands

AC-PRC-001-002 maps REQ-PRC-001-002 and, for the four-path swept set, REQ-PRC-001-004. Branch R (replace in place by the slash forms) is the planned branch; the alternate D (remove the two lines) is not planned and returns the SPEC to the plan phase.

- **Given** the sweep tip and branch R,
- **When** these six commands run,
- **Then** commands 1 and 2 print nothing and exit 1; commands 3 and 4 print four lines ending `:2` and exit 0; commands 5 and 6 print four lines ending `:1` and exit 0.

```
git grep -c -E "(^|[^/])moai (run|sync) SPEC-ID" HEAD -- 'docs-site/content/*/worktree/_index.md'
git grep -c -E "(^|[^/])moai[[:space:]]+(run|sync)[[:space:]]+SPEC-ID" HEAD -- 'docs-site/content/*/worktree/_index.md'
git grep -c -E "\"/moai (run|sync)([ <\"]|$)" HEAD -- 'docs-site/content/*/worktree/_index.md'
git grep -c -w -E "B[34]" HEAD -- 'docs-site/content/*/worktree/_index.md'
git grep -c -E "^[[:space:]]+B2 --> B3\[\"/moai run SPEC-ID\"\]$" HEAD -- 'docs-site/content/*/worktree/_index.md'
git grep -c -E "^[[:space:]]+B3 --> B4\[\"/moai sync SPEC-ID\"\]$" HEAD -- 'docs-site/content/*/worktree/_index.md'
```

Status at the pin, per command. Commands 1 and 2 (E-SWD-4): four lines `:2`, exit 0, red. Command 3 (E-SWD-5): empty, exit 1, red (the slash-form labels do not exist yet). Command 4 (E-SWD-5): four lines `:2`, exit 0, green and preserved: the references survive; it is the dangling-reference check, and a half-edited pair that drops a `B3` or `B4` line fails it. Commands 5 and 6 (E-SWD-5): empty, exit 1, red. The flip is M2 for commands 1, 2, 3, 5, and 6.

Commands 5 and 6 are the label check: command 3 and command 4 count slash-form labels and `B3`/`B4` nodes but cannot tell which command sits on which node. Command 5 requires B3 to carry `/moai run SPEC-ID` and command 6 requires B4 to carry `/moai sync SPEC-ID`, each as the whole line, so a swap, a duplicate, and trailing text after the closing bracket all fail them. Mutants, each observed on a scratch fixture of the ko page (E-SWD-15): B3 and B4 swapped passes commands 1 to 4 and prints nothing for commands 5 and 6; both labels set to `/moai run` passes commands 1 to 4 and prints nothing for command 6; trailing text after the B3 label passes commands 1 to 4 and prints nothing for command 5.

### AC-PRC-001-003 — the seven `mcp-server.md` cells change to the marker, the rest of the row stays, each row keeps its place, and the locales agree

AC-PRC-001-003 maps REQ-PRC-001-001, REQ-PRC-001-004, and REQ-PRC-001-005. The edit unit and its bounds are the table in plan.md §A.3. This criterion measures four things separately, because the line bound alone cannot tell which cell of a row changed or where a row sits: the fourth cell of each row (exact form), the first three cells of each row (literal prefix), the line at which each row sits (position), and the agreement of the four locales (per-row compare).

- **Given** the sweep tip and the recorded answers for Q13,
- **When** the commands of the three blocks below run,
- **Then** the cell-form block prints: command 1 four numstat lines, one per locale, each `7` added and `7` removed; command 2 four lines ending `:7`. The per-row block prints, for each of its seven commands, four lines ending `:1` (one per locale). The rest-of-row block prints: command 10 four lines ending `:41`; commands 11 to 14 one line each, ending `:7`, for ko, en, ja, and zh in that order. The position block prints: command 15 twenty-eight lines, exit 0, the seven rows at lines 130, 163, and 182 to 186 in each of en, ja, ko, and zh (listed below the command).

Cell form (commands 1 and 2):

```
git diff --numstat 26834dd17d HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__(spec_progress|audit_cache|codex_task|codex_setup|codex_job_status|codex_job_result|codex_job_cancel)\` \| [^|]+ \| [^|]+ \| — \|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
```

Per-row compare across locales (commands 3 to 9), one command per tool id; each reads the four locales at once, so a row whose cell differs in one locale prints three lines instead of four:

```
git grep -c -E "^\| \`mcp__moai__spec_progress\` \| [^|]+ \| [^|]+ \| — \|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__audit_cache\` \| [^|]+ \| [^|]+ \| — \|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__codex_task\` \| [^|]+ \| [^|]+ \| — \|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__codex_setup\` \| [^|]+ \| [^|]+ \| — \|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__codex_job_status\` \| [^|]+ \| [^|]+ \| — \|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__codex_job_result\` \| [^|]+ \| [^|]+ \| — \|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -E "^\| \`mcp__moai__codex_job_cancel\` \| [^|]+ \| [^|]+ \| — \|$" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
```

Rest of the row and of the page (command 10, then commands 11 to 14 per locale; the patterns are the first three cells of each of the seven rows as they stand at the pin, E-SWD-6b):

```
git grep -c -E "^\| \`mcp__moai__" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
git grep -c -F -e "| \`mcp__moai__spec_progress\` | SPEC 문서 목록 + frontmatter 조회 | manager-spec, manager-docs | " -e "| \`mcp__moai__audit_cache\` | plan-audit PASS 캐시 (compute_hash / lookup / store, 프로세스 간 공유) | sync-auditor | " -e "| \`mcp__moai__codex_task\` | 코딩/조사 작업을 codex에 위임 (동기 또는 백그라운드) | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_setup\` | 로컬 codex 설치 탐지 (LookPath + 버전 + 인증) | super-advisor | " -e "| \`mcp__moai__codex_job_status\` | 백그라운드 codex 작업 상태/기록 읽기 | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_job_result\` | 백그라운드 codex 작업 출력 읽기 | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_job_cancel\` | 실행 중인 백그라운드 codex 작업 중단 | manager-develop, super-advisor | " HEAD -- 'docs-site/content/ko/guides/mcp-server.md'
git grep -c -F -e "| \`mcp__moai__spec_progress\` | List SPEC documents + frontmatter | manager-spec, manager-docs | " -e "| \`mcp__moai__audit_cache\` | plan-audit PASS cache (compute_hash / lookup / store, process-shared) | sync-auditor | " -e "| \`mcp__moai__codex_task\` | Delegate a coding/investigation task to codex (sync or background) | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_setup\` | Probe local codex install (LookPath + version + auth) | super-advisor | " -e "| \`mcp__moai__codex_job_status\` | Read a background codex job's status/record | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_job_result\` | Read a background codex job's output | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_job_cancel\` | Stop a running background codex job | manager-develop, super-advisor | " HEAD -- 'docs-site/content/en/guides/mcp-server.md'
git grep -c -F -e "| \`mcp__moai__spec_progress\` | SPEC 文書リスト + frontmatter 照会 | manager-spec, manager-docs | " -e "| \`mcp__moai__audit_cache\` | plan-audit PASS キャッシュ（compute_hash / lookup / store、プロセス間共有） | sync-auditor | " -e "| \`mcp__moai__codex_task\` | コーディング/調査ジョブを codex に委譲（同期またはバックグラウンド） | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_setup\` | ローカル codex インストール検出（LookPath + バージョン + 認証） | super-advisor | " -e "| \`mcp__moai__codex_job_status\` | バックグラウンド codex ジョブ状態/記録読み取り | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_job_result\` | バックグラウンド codex ジョブ出力読み取り | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_job_cancel\` | 実行中のバックグラウンド codex ジョブ中断 | manager-develop, super-advisor | " HEAD -- 'docs-site/content/ja/guides/mcp-server.md'
git grep -c -F -e "| \`mcp__moai__spec_progress\` | SPEC 文档列表 + frontmatter 查询 | manager-spec, manager-docs | " -e "| \`mcp__moai__audit_cache\` | plan-audit PASS 缓存（compute_hash / lookup / store，进程间共享） | sync-auditor | " -e "| \`mcp__moai__codex_task\` | 将编码/调查任务委托给 codex（同步或后台） | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_setup\` | 探测本地 codex 安装（LookPath + 版本 + 认证） | super-advisor | " -e "| \`mcp__moai__codex_job_status\` | 读取后台 codex 任务的状态/记录 | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_job_result\` | 读取后台 codex 任务的输出 | manager-develop, super-advisor | " -e "| \`mcp__moai__codex_job_cancel\` | 中断正在运行的后台 codex 任务 | manager-develop, super-advisor | " HEAD -- 'docs-site/content/zh/guides/mcp-server.md'
```

Position of the seven rows (command 15). The dot in the pattern stands for the backtick that closes and opens the row id; it keeps the command runnable under the worktree guard, which refuses a backtick. The pattern stops before the fourth cell, so its output is the same at the pin and at the sweep tip when the edit is line for line:

```
git grep -n -o -E "^\| .mcp__moai__(spec_progress|audit_cache|codex_task|codex_setup|codex_job_status|codex_job_result|codex_job_cancel). \|" HEAD -- 'docs-site/content/*/guides/mcp-server.md'
```

Expected stdout, exit 0: 28 lines, locales in the order en, ja, ko, zh (git sorts by path), and in each locale these seven in this order, each as ``HEAD:docs-site/content/<locale>/guides/mcp-server.md:<line>:| `mcp__moai__<id>` |``: line 130 `spec_progress`, 163 `audit_cache`, 182 `codex_task`, 183 `codex_setup`, 184 `codex_job_status`, 185 `codex_job_result`, 186 `codex_job_cancel`. Any other line number, any other id at a line, or any missing line is a failure; the exit code is 0 whenever any row matches and is not the verdict.

RED-now at the pin: command 1 prints nothing, exit 0 (E-SWD-6); command 2 and commands 3 to 9 print nothing, exit 1 (E-SWD-6, E-SWD-6a). Commands 10 to 15 are green at the pin and are the preservation part of this criterion, a regression guard rather than a RED-now cell: command 10 prints four lines `:41` (the page still carries 41 tool-row lines; the count does not see where a row sits), commands 11 to 14 print `:7` each (E-SWD-6, E-SWD-6b), and command 15 prints the 28 lines above (E-SWD-6c). Command 15 is green at the pin by construction, because the pin is where the line numbers were measured; the edit changes only the fourth cell of each row, which the pattern does not read, so it stays green through M3 and no milestone flips it. It is proved by its mutants (E-SWD-15), not by a red arrival. Why the criterion as a whole is red: no edit has been made, so no cell holds the marker yet. The preservation part exists because the line bound of command 1 is a count of lines and does not see which cell of a row changed, and neither the line bound, the row count, `M`, nor the literal prefixes sees where a row sits.

Mutants this closes, each observed on a scratch fixture (E-SWD-15): a cell followed by text after the marker (`— ; run ...`) leaves 6 of 7 rows in command 2 and passes AC-PRC-001-001; blanking a cell, a typo such as `moai spec lst`, a registered command in a cell (the stale-row candidates `moai codex status` and `moai spec status --list` included), and a stale form left in one cell each leave 6 of 7 rows; overwriting the purpose or agents cell of a row while setting its CLI cell correctly passes commands 1 and 2 (the line bound is still `7	7`) and prints `6` in the matching prefix command; blanking the page or deleting a row breaks command 1 or command 10; editing three locales instead of four leaves a file with no numstat line and a per-row command with three lines; one locale carrying a different disposition for a row prints three lines for that row. Because `M` admits one cell form, four per-row lines of `:1` and four per-file lines of `:7` mean that all 28 cells have the same disposition. The reviewer's read of the edited lines is no longer the only guard for the pairing of a replacement to a row: no replacement other than the marker passes.

Mutants that move a row, each observed on a scratch fixture (E-SWD-15): a row deleted from the table and appended after the last line of the page, two adjacent `codex_*` rows swapped, two `codex_*` rows far apart swapped, and a row moved up or down by one line each pass commands 1 to 14 (the line bound stays `7 7`, `M` stays 7, every per-row command prints `:1`, the literal prefixes stay 7, the row count stays 41) and each prints different lines from command 15. What the commands now measure about position: with command 1 reading `7 7` per file against the pin, AC-PRC-001-001 showing that no stale row remains (so the seven removed lines are the seven stale rows), command 2 and commands 3 to 9 showing that exactly seven lines hold the marker form, and command 15 showing that those seven rows sit at lines 130, 163, and 182 to 186, the seven added lines are the seven rows at their pinned lines and every other line of each file is the line the pin has at the same position. Within a file this leaves no row, target or not, that can have moved: moving a non-target row adds one removed and one added line, which command 1 reads as `8 8` (observed on the non-target pair swap, E-SWD-15). The one remainder is stated in G12.

### AC-PRC-001-004 — the Mermaid labels change within their edit unit

AC-PRC-001-004 maps REQ-PRC-001-002 and REQ-PRC-001-005.

- **Given** the sweep tip and branch R,
- **When** this command runs,
- **Then** it prints four numstat lines, one per locale, each `2` added and `2` removed.

```
git diff --numstat 26834dd17d HEAD -- 'docs-site/content/*/worktree/_index.md'
```

RED-now at the pin (E-SWD-7): empty stdout, exit 0, because no edit has been made. The bound also keeps the five adjacent labels per locale (plan.md §A.2) untouched: an edit to any of them adds a changed line and the row reads `3	3` or higher. Removing the two lines instead (`0	2`) is the alternate D; it is not planned and also fails AC-PRC-001-002.

### AC-PRC-001-005 — the four retirement statements stay

AC-PRC-001-005 maps REQ-PRC-001-003 and holds on the retain ruling of Q-SWD-1 (board record 4604). On a later verdict for the alternate it is not applicable and the SPEC returns to the plan phase.

- **Given** the sweep tip and the retain ruling,
- **When** these two commands run,
- **Then** command 1 prints four lines ending `:1` and exits 0, and command 2 prints nothing and exits 0.

```
git grep -c -E "(^|[^/])moai cg([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
git diff --numstat 26834dd17d HEAD -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
```

Green at the pin (E-SWD-8). Controls: command 2 over the range of commit c5620da649 prints four numstat rows (`2 0`, `1 1`, `2 1`, `1 1`), so the form does print when these files change; command 1 run on a scratch copy of the en page with its sentence deleted prints nothing and exits 1 (E-SWD-15). Command 2 compares with the pin, so it reads red if another card's edit to these files is absorbed into the branch; that case is the re-pin of plan.md M4, not a failure of this criterion.

### AC-PRC-001-006 — nothing outside the swept files changes

AC-PRC-001-006 maps REQ-PRC-001-005.

- **Given** the sweep tip,
- **When** this command runs,
- **Then** it prints nothing and exits 0.

```
git diff --name-only 26834dd17d HEAD -- . ':!docs-site/content/*/guides/mcp-server.md' ':!docs-site/content/*/worktree/_index.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001' ':!.moai/reports/t1724'
```

Green at the pin (E-SWD-9). The four `prompt-caching.md` files are deliberately not excluded, so a change to them prints. Control: the same exclusion form over the range of commit 0de4cfabe2 prints one path, `.moai/specs/SPEC-ASIDE-BROWSER-001/progress.md`, so the form prints when something outside the list changes.

## C. Evidence ledger

Each entry gives the command, the observed output (verbatim unless marked), the exit code as its own field, and the tree it read. Entries that pin no tree of their own read the document-level pin 26834dd17d. The entries E-SWD-10, E-SWD-11, E-SWD-14, and E-SWD-16 carry their own pin 74ad197380 and were not re-run for revision 0.2.0. The entry E-SWD-6c (revision 0.2.1) carries its own pin e11f08f6a7b07e8f878b3e7d2a456a60aacdfd97, the HEAD it was measured on, whose docs-site tree equals that of 26834dd17d.

### E-SWD-0 — pin and tool provenance

```
git rev-parse HEAD                                 -> 26834dd17d3eb9597cde0366a5d4f01130417172   exit 0
git diff --stat 74ad197380 26834dd17d -- docs-site -> (no stdout)   exit 0
git diff --name-only 74ad197380 26834dd17d         -> .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001/{acceptance,decision-index,plan,progress,spec}.md   exit 0
SPEC id check  ID="SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL   -> PASS
```

The docs-site tree of 26834dd17d equals that of 74ad197380, and `internal/` and `cmd/` are the same tree in both commits (only this SPEC's five files differ). The installed-build facts of revision 0.1.0 stand and keep their pin 74ad197380: `moai version` printed `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z; `git merge-base --is-ancestor 7705b365e2 74ad197380` printed nothing, exit 0; `git rev-list --count 74ad197380..7705b365e2` printed 0 and `git rev-list --count 7705b365e2..74ad197380` printed 68. The installed build is a strict ancestor of the tree, 68 commits behind. No Go build was made; CLI claims in this SPEC are attributed to "installed build 7705b365e2 against tree 74ad197380" (plan.md §A.6).

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

Command: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model|codex (job|task|setup)|todo (rename|pick)|spec list)([^[:alnum:]-]|$)" 26834dd17d -- 'docs-site/content/*/guides/mcp-server.md'`. Pin 26834dd17d. Exit code: 0. Stdout:

```
26834dd17d:docs-site/content/en/guides/mcp-server.md:7
26834dd17d:docs-site/content/ja/guides/mcp-server.md:7
26834dd17d:docs-site/content/ko/guides/mcp-server.md:7
26834dd17d:docs-site/content/zh/guides/mcp-server.md:7
```

Why red: each of the four files carries seven lines that match `U`, the seven rows of plan.md §A.2. A green path exists because the sweep edits exactly those cells (M3).

### E-SWD-3 — RED-now, pattern UW over `mcp-server.md` (AC-PRC-001-001, command 2)

Command: the same as E-SWD-2 with `UW` in place of `U`. Pin 26834dd17d. Exit code: 0. Stdout: the same four lines, each ending `:7`. The whitespace guard adds no match today and closes only the respelling route.

### E-SWD-4 — RED-now, patterns E03 and E03W over `worktree/_index.md` (AC-PRC-001-002, commands 1 and 2)

Command 1: `git grep -c -E "(^|[^/])moai (run|sync) SPEC-ID" 26834dd17d -- 'docs-site/content/*/worktree/_index.md'`. Pin 26834dd17d. Exit code: 0. Stdout:

```
26834dd17d:docs-site/content/en/worktree/_index.md:2
26834dd17d:docs-site/content/ja/worktree/_index.md:2
26834dd17d:docs-site/content/ko/worktree/_index.md:2
26834dd17d:docs-site/content/zh/worktree/_index.md:2
```

Command 2 (`E03W`): `git grep -c -E "(^|[^/])moai[[:space:]]+(run|sync)[[:space:]]+SPEC-ID" 26834dd17d -- 'docs-site/content/*/worktree/_index.md'`. Exit code 0, stdout the same four lines. Why red: each file carries the two Mermaid label lines B3 and B4 of plan.md §A.2. A green path exists because the sweep replaces exactly those lines (M2).

### E-SWD-5 — AC-PRC-001-002, commands 3 to 6

Command 3: `git grep -c -E "\"/moai (run|sync)([ <\"]|$)" 26834dd17d -- 'docs-site/content/*/worktree/_index.md'`. Pin 26834dd17d. Exit code: 1. Stdout: empty. Reading: red (the slash-form labels do not exist yet).

Command 4: `git grep -c -w -E "B[34]" 26834dd17d -- 'docs-site/content/*/worktree/_index.md'`. Pin 26834dd17d. Exit code: 0. Stdout: four lines, `26834dd17d:docs-site/content/{en,ja,ko,zh}/worktree/_index.md:2`. Reading: green and preserved (the references must survive).

Command 5: `git grep -c -E "^[[:space:]]+B2 --> B3\[\"/moai run SPEC-ID\"\]$" 26834dd17d -- 'docs-site/content/*/worktree/_index.md'`. Pin 26834dd17d. Exit code: 1. Stdout: empty. Reading: red; the B3 line holds `moai run SPEC-ID` without the slash.

Command 6: `git grep -c -E "^[[:space:]]+B3 --> B4\[\"/moai sync SPEC-ID\"\]$" 26834dd17d -- 'docs-site/content/*/worktree/_index.md'`. Pin 26834dd17d. Exit code: 1. Stdout: empty. Reading: red; the B4 line holds `moai sync SPEC-ID` without the slash.

Why red and where it flips: the sweep replaces the B3 and B4 lines in all four locales (M2); the green output is four lines ending `:1` for commands 5 and 6 and four lines ending `:2` for command 3. The same commands were run on scratch fixtures of the ko page (E-SWD-15): the replaced fixture prints `2`, `2`, `1`, `1` for commands 3 to 6.

### E-SWD-6 — RED-now, AC-PRC-001-003 (cell form, row count, controls)

```
git diff --numstat 26834dd17d HEAD -- 'docs-site/content/*/guides/mcp-server.md'      -> (no stdout)   exit 0   (HEAD == 26834dd17d at measurement)
git grep -c -E "<M>" 26834dd17d -- 'docs-site/content/*/guides/mcp-server.md'         -> (no stdout)   exit 1
git grep -c -E "^\| \`mcp__moai__" 26834dd17d -- 'docs-site/content/*/guides/mcp-server.md'   -> 26834dd17d:docs-site/content/{en,ja,ko,zh}/guides/mcp-server.md:41   exit 0
```

(`<M>` stands for the pattern `M` of §A.1, run in the exact spelling of AC-PRC-001-003, command 2.) Why red: the first command is empty because nothing is edited, and the exact-form pattern matches none of the seven target rows because their fourth cells still hold the stale forms. The seven target rows exist: `git grep -c -E "^\| \`mcp__moai__(spec_progress|audit_cache|codex_task|codex_setup|codex_job_status|codex_job_result|codex_job_cancel)\` " 26834dd17d -- 'docs-site/content/*/guides/mcp-server.md'` prints `:7` for each locale, exit 0, the same count as E-SWD-2, so each `U` hit is one of the seven rows.

Controls, both on the repository at the pin. The pattern shape on the rows that already carry the marker with a tail: `git grep -c -E "^\| \`mcp__moai__(glm_task|audit_multi)\` \| [^|]+ \| [^|]+ \| —[^|]*\|$" 26834dd17d -- 'docs-site/content/*/guides/mcp-server.md'` (`MC`) prints `:2` for each locale, exit 0. The bare marker as the whole fourth cell, over every tool row: `git grep -c -E "^\| \`mcp__moai__[a-z_]+\` \| [^|]+ \| [^|]+ \| — \|$" 26834dd17d -- 'docs-site/content/*/guides/mcp-server.md'` prints `:18` for each locale, exit 0, so the exact form `M` is the page's own convention and the engine matches it where it occurs.

Control for the numstat bound: `git diff --numstat f32c831de5~1 f32c831de5 -- 'docs-site/content/*/guides/mcp-server.md'` prints `1	1` for each of the four paths, exit 0, below the `7	7` bound.

### E-SWD-6a — RED-now, AC-PRC-001-003 (per-row compare, commands 3 to 9)

Each of the seven commands `git grep -c -E "^\| \`mcp__moai__<id>\` \| [^|]+ \| [^|]+ \| — \|$" 26834dd17d -- 'docs-site/content/*/guides/mcp-server.md'`, with `<id>` in turn `spec_progress`, `audit_cache`, `codex_task`, `codex_setup`, `codex_job_status`, `codex_job_result`, and `codex_job_cancel`. Pin 26834dd17d. Exit code: 1 for each of the seven. Stdout: empty for each. Why red: the fourth cell of the row still holds the stale form in all four locales. The green output is four lines ending `:1`, one per locale, for each id; a row whose cell differs in one locale prints three lines (E-SWD-15).

### E-SWD-6b — AC-PRC-001-003 (commands 11 to 14, the literal prefixes of the seven rows)

Commands: the four `git grep -c -F -e ...` lines of AC-PRC-001-003, one per locale, run with `HEAD`, which equalled `26834dd17d` at measurement (E-SWD-0). Exit code: 0 for each. Stdout:

```
HEAD:docs-site/content/ko/guides/mcp-server.md:7
HEAD:docs-site/content/en/guides/mcp-server.md:7
HEAD:docs-site/content/ja/guides/mcp-server.md:7
HEAD:docs-site/content/zh/guides/mcp-server.md:7
```

Reading: green and preserved. Each pattern is the first three cells of one of the seven rows with their separators, copied from the pin; the count is the number of lines that match any of the seven, so it is 7 only when all seven rows still carry their tool id, purpose, and agents cells byte-identical. The patterns read the pin's text, so they are valid at the pin and at the sweep tip and at no point depend on a moving ref.

### E-SWD-6c — AC-PRC-001-003 (command 15, the position of the seven rows)

Revision 0.2.1. Pin e11f08f6a7b07e8f878b3e7d2a456a60aacdfd97 (HEAD at measurement; `git rev-parse HEAD` printed it, exit 0). The docs-site tree equals both earlier pins: `git diff --stat 26834dd17d HEAD -- docs-site` and `git diff --stat 74ad197380 HEAD -- docs-site` each printed nothing, exit 0.

Command: the one in AC-PRC-001-003 under "Position of the seven rows", run unmodified (it carries no backtick, so cell and run are the same text). Exit code: 0. Stdout, 28 lines:

```
HEAD:docs-site/content/en/guides/mcp-server.md:130:| `mcp__moai__spec_progress` |
HEAD:docs-site/content/en/guides/mcp-server.md:163:| `mcp__moai__audit_cache` |
HEAD:docs-site/content/en/guides/mcp-server.md:182:| `mcp__moai__codex_task` |
HEAD:docs-site/content/en/guides/mcp-server.md:183:| `mcp__moai__codex_setup` |
HEAD:docs-site/content/en/guides/mcp-server.md:184:| `mcp__moai__codex_job_status` |
HEAD:docs-site/content/en/guides/mcp-server.md:185:| `mcp__moai__codex_job_result` |
HEAD:docs-site/content/en/guides/mcp-server.md:186:| `mcp__moai__codex_job_cancel` |
HEAD:docs-site/content/ja/guides/mcp-server.md:130:| `mcp__moai__spec_progress` |
HEAD:docs-site/content/ja/guides/mcp-server.md:163:| `mcp__moai__audit_cache` |
HEAD:docs-site/content/ja/guides/mcp-server.md:182:| `mcp__moai__codex_task` |
HEAD:docs-site/content/ja/guides/mcp-server.md:183:| `mcp__moai__codex_setup` |
HEAD:docs-site/content/ja/guides/mcp-server.md:184:| `mcp__moai__codex_job_status` |
HEAD:docs-site/content/ja/guides/mcp-server.md:185:| `mcp__moai__codex_job_result` |
HEAD:docs-site/content/ja/guides/mcp-server.md:186:| `mcp__moai__codex_job_cancel` |
HEAD:docs-site/content/ko/guides/mcp-server.md:130:| `mcp__moai__spec_progress` |
HEAD:docs-site/content/ko/guides/mcp-server.md:163:| `mcp__moai__audit_cache` |
HEAD:docs-site/content/ko/guides/mcp-server.md:182:| `mcp__moai__codex_task` |
HEAD:docs-site/content/ko/guides/mcp-server.md:183:| `mcp__moai__codex_setup` |
HEAD:docs-site/content/ko/guides/mcp-server.md:184:| `mcp__moai__codex_job_status` |
HEAD:docs-site/content/ko/guides/mcp-server.md:185:| `mcp__moai__codex_job_result` |
HEAD:docs-site/content/ko/guides/mcp-server.md:186:| `mcp__moai__codex_job_cancel` |
HEAD:docs-site/content/zh/guides/mcp-server.md:130:| `mcp__moai__spec_progress` |
HEAD:docs-site/content/zh/guides/mcp-server.md:163:| `mcp__moai__audit_cache` |
HEAD:docs-site/content/zh/guides/mcp-server.md:182:| `mcp__moai__codex_task` |
HEAD:docs-site/content/zh/guides/mcp-server.md:183:| `mcp__moai__codex_setup` |
HEAD:docs-site/content/zh/guides/mcp-server.md:184:| `mcp__moai__codex_job_status` |
HEAD:docs-site/content/zh/guides/mcp-server.md:185:| `mcp__moai__codex_job_result` |
HEAD:docs-site/content/zh/guides/mcp-server.md:186:| `mcp__moai__codex_job_cancel` |
```

The exit code was read by running the command with its output redirected to a file and the status echoed in a wrapper script (`exit=0`, 28 lines counted); the wrapper is the measurement harness and is not part of the cited command.

Why this is not red now, and what it is: this command asserts the positions the rows already hold, so it is green at the pin by construction. It is a preservation (regression-guard) part of AC-PRC-001-003, like commands 10 to 14, not a RED-now cell; its red is the output differing from the 28 lines above, observed on the mutants of E-SWD-15 and not on the repository. Green path: M3 edits the fourth cell of each row on its own line and the pattern stops before that cell, so the output is the same 28 lines at the sweep tip, and M5 re-runs it. No milestone flips it.

Moving-coordinate classification (verification-claim-integrity-detail.md, Moving-ref predicate). Test 1: replacing the coordinates by the commit they were measured on keeps the sentence "at e11f08f6a7 the rows sit at these lines" true for a later reader, so the class is ANCHOR. Test 3: re-run next week on the same base with no work in between, the same output is expected. Remedy R1: the literal line numbers, with the tree and date they were measured on (e11f08f6a7, 2026-10-19). They are valid only while the card branch's base is the one measured. If another card shifts these pages above line 130 and the branch absorbs it, command 15 reads red for a correct edit; that is the M4 re-pin trigger (plan.md M4, R1), where the coordinates are re-measured on the new base, not a failure of the sweep.

### E-SWD-7 — RED-now, AC-PRC-001-004

Command: `git diff --numstat 26834dd17d HEAD -- 'docs-site/content/*/worktree/_index.md'`. Pin 26834dd17d (HEAD equals it at measurement). Exit code: 0. Stdout: empty. Why red: no edit has been made. Control for the bound: `git diff --numstat ce79ef7caf~1 ce79ef7caf -- 'docs-site/content/*/worktree/_index.md'` exit 0, prints

```
5	35	docs-site/content/en/worktree/_index.md
6	36	docs-site/content/ja/worktree/_index.md
7	38	docs-site/content/ko/worktree/_index.md
6	35	docs-site/content/zh/worktree/_index.md
```

which is above the `2`-removed bound.

### E-SWD-8 — AC-PRC-001-005 at the pin, and its controls

```
git grep -c -E "(^|[^/])moai cg([^[:alnum:]-]|$)" 26834dd17d -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
  -> 26834dd17d:docs-site/content/{en,ja,ko,zh}/cost-optimization/prompt-caching.md:1   exit 0
git diff --numstat 26834dd17d HEAD -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
  -> (no stdout)   exit 0
git diff --numstat c5620da649~1 c5620da649 -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
  -> 2 0 en, 1 1 ja, 2 1 ko, 1 1 zh   exit 0
git diff --quiet c5620da649~1 c5620da649 -- 'docs-site/content/*/cost-optimization/prompt-caching.md'
  -> (no stdout)   exit 1
```

### E-SWD-9 — AC-PRC-001-006 at the pin, and its control

Command: `git diff --name-only 26834dd17d HEAD -- . ':!docs-site/content/*/guides/mcp-server.md' ':!docs-site/content/*/worktree/_index.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001' ':!.moai/reports/t1724'`. Pin 26834dd17d (HEAD equals it at measurement). Exit code: 0. Stdout: empty. Control: the same form with `0de4cfabe2~1` and `0de4cfabe2` as the two ends prints `.moai/specs/SPEC-ASIDE-BROWSER-001/progress.md`, exit 0.

E-SWD-9a, the no-CLI marker convention: `git grep -n -E "— \(" 26834dd17d -- 'docs-site/content/en/guides/mcp-server.md'` exit 0 prints, in the en copy, line 159 (`audit_multi`, `— (MCP-only convergence entry point)`), 204 (`glm_task`, `— (no such CLI)`), 226 (`jev_ask`, `— (MCP-only)`), and 234 to 238 (`factory_msg_*`, `— (MCP-only)`): the marker followed by a parenthetical phrase in 8 rows. The bare marker as the whole fourth cell is in 18 rows per locale (E-SWD-6).

### E-SWD-10 — CLI surface, installed build 7705b365e2 against tree 74ad197380

E-SWD-10 and E-SWD-11 are corroboration for the classification in plan.md §A.4, not RED-now cells: their commands and outputs are abridged to the deciding lines, and no release-blocking criterion rests on them alone. They were observed in revision 0.1.0 and were not re-run; `internal/` and `cmd/` are the same tree at 26834dd17d (E-SWD-0).

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

Observed in revision 0.1.0 at the pin 74ad197380; not re-run.

```
git grep -n -E 'Use:...("|`)' 74ad197380 -- internal/cli/spec*.go   -> spec.go:13 "spec"; spec_archive.go:35 "archive"; spec_audit.go:64 "audit"; spec_ceiling.go:35 "ceiling <SPEC-ID> ..."; spec_close.go:71 "close SPEC-ID"; spec_drift.go:80 "drift"; spec_lint.go:35 "lint ..."; spec_status.go:29 "status <SPEC-ID> <new-status>"; spec_view.go:20 "view <SPEC-ID>"   (no "list")
git grep -n -E 'Use:...' 74ad197380 -- internal/cli/codex_audit_launch.go   -> :880 "audit <role>"; :907 "role-audit <role>"
git grep -n -E 'Use:\s*("|`)(cache|task|job|setup)' 74ad197380 -- internal cmd   -> internal/cli/glm.go:125 "setup [api-key]"; internal/cli/profile_setup.go:199 "setup [name]"; internal/cli/init_update_profile_path_test.go:278 (a test fixture)   (no cache, task, or job)
git grep -n -E 'Use:|Hidden:|Short:' 74ad197380 -- internal/cli/cg.go   -> :13 var cgCmd = &cobra.Command{Use: "cg", Short: "Retired CG launcher", Hidden: true, ...}
git diff --name-only 7705b365e2 74ad197380 -- internal/cli cmd   -> 24 files; none is a spec*.go file or cg.go; includes codex_audit_launch.go, codex_audit_leg.go, mcp_claude.go, mcp_codex.go, mcp_convergence.go, mcp_factory_card.go, factory_card.go, factory_card_pr.go, integration.go, integration_candidate.go, ci_verdict.go and test files
git ls-files -- .claude/commands/moai/run.md .claude/commands/moai/sync.md .claude/commands/moai/plan.md   -> all three paths
```

### E-SWD-12 — what else the pages say, and the silent families

`git grep -n -E "moai" 26834dd17d -- 'docs-site/content/*/cost-optimization/prompt-caching.md'` exit 0 prints exactly four lines, en 268, ja 171, ko 243, zh 171, each the retirement sentence (en: "`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. ..."). `git grep -c -E "/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 26834dd17d -- <the three pathspecs>` prints nothing, exit 1; the control `git grep -c -E "/moai[: ](plan|run|sync|design|issue)([^[:alnum:]-]|$)" 26834dd17d -- 'docs-site/content/*/worktree/_index.md'` prints `:7` for each locale, exit 0, so the slash-form engine fires where such forms occur. The union `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model|codex (job|task|setup)|todo (rename|pick)|spec list|(run|sync) SPEC-ID)([^[:alnum:]-]|$)" 26834dd17d -- <the three pathspecs>` exit 0 prints twelve lines: `:1` for each `prompt-caching.md`, `:7` for each `mcp-server.md`, `:2` for each `worktree/_index.md`; 4 + 28 + 8 = 40 lines. Per-family counts at the pin: E-01 over the twelve files, exit 0, prints `:1` for each `prompt-caching.md` and each `mcp-server.md`; E-02 (`(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)([^[:alnum:]-]|$)`), exit 0, prints `:6` for each `mcp-server.md`; E-03 prints `:2` for each `worktree/_index.md` (E-SWD-4).

### E-SWD-13 — the adjacent labels

Command: `git grep -c -E "\"moai (plan|run|sync)(<br/>|\")" 26834dd17d -- 'docs-site/content/*/cost-optimization/prompt-caching.md' 'docs-site/content/*/guides/mcp-server.md' 'docs-site/content/*/worktree/_index.md'`. Exit 0. Stdout: `:5` for each of the four `worktree/_index.md` files and nothing for the other eight. The five per locale are listed with their lines in plan.md §A.2.

### E-SWD-14 — parity baseline

Observed in revision 0.1.0 at the pin 74ad197380; not re-run. Command: `bash scripts/docs-i18n-check.sh` at that pin; exit 0; the output was redirected to a scratch file and read back. Tail:

```
=== Summary ===
Errors:   0
Warnings: 0
OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.
```

File counts printed: ko 153, en 153, ja 153, zh 153. The check passes at the pin, so on its own it cannot go red for this change; it is a confirmatory check at M5, not a criterion. The docs-site tree of 26834dd17d equals the tree it ran on (E-SWD-0).

### E-SWD-15 — mutant and green fixtures (scratch copies, not repository trees)

Fixtures were made with `sed` from the en and ko `mcp-server.md`, the ko `worktree/_index.md`, and the en `prompt-caching.md`, copied from the working tree at 26834dd17d into the session scratch directory, and read with `grep -c -E` and `grep -c -F -f` (the same POSIX engine as `git grep -E`; the pattern `M` printed the same counts on the repository by `git grep` in E-SWD-6 and on the pin copy by `grep`: 0). They are inputs constructed to be green or to be a mutant; they do not stand for the sweep. The numbers are the counts the AC commands print (per file for the `mcp-server.md` fixtures, `removed/added` from `diff` for the line bound). "Old" is the cell-form patterns of revision 0.1.0 (`M5` and `M2`), run on the same fixture to show which mutants they accepted.

| Fixture | Construction | Result |
|---------|--------------|--------|
| green (mcp, en) | the seven fourth cells replaced by the bare marker `—` | `U` 0, `UW` 0, `M` 7, each `MID` 1, literal prefixes 7, `ROW` 41, line bound 7 removed and 7 added |
| green (mcp, ko) | the same on ko | `U` 0, `M` 7, each `MID` 1 |
| D1 trailing text | green, the `audit_cache` cell set to ``— ; run `moai spec lst` `` | `U` 0 and `UW` 0 (passes AC-PRC-001-001), old `M5` 5 (passes the old criterion), new `M` 6 (red), literal prefixes 7 |
| D2 purpose overwritten | green, the `codex_task` purpose cell set to `OVERWRITTEN PURPOSE` | `U` 0, `M` 7, old `M5` 5, old `M2` 2, line bound 7 and 7 (all pass), literal prefixes 6 (red) |
| D2 agents overwritten | green, the `audit_cache` agents cell set to `OVERWRITTEN AGENTS` | `U` 0, `M` 7, old `M5` 5, line bound 7 and 7 (pass), literal prefixes 6 (red) |
| D3 candidate for another row | green, the `spec_progress` cell set to ``` `moai codex status` ``` | `U` 0, old `M2` 2 (accepted by the old criterion), new `M` 6 (red), `MID` for `spec_progress` 0 |
| D3 candidate for its own row | green, the `spec_progress` cell set to ``` `moai spec status --list` ``` | `U` 0, old `M2` 2 (accepted), new `M` 6 (red) |
| D3 cross-locale, candidate in one locale | the D3 fixture as the en page and the green fixture as the ko page | en `M` 6 and `MID` for `spec_progress` 0, ko `M` 7: the four-line output of command 2 holds `:6` and `:7`, and the per-row command prints three lines |
| D3 cross-locale, stale row in one locale | green ko, the `audit_cache` cell set to ``` `moai audit cache` ``` | ko `M` 6, `MID` for `audit_cache` 0 (three lines), `U` 1 |
| blank cell | green, the `audit_cache` cell emptied | `U` 0, `M` 6 |
| typo | green, the `spec_progress` cell set to ``` `moai spec lst` ``` | `U` 0 (passes AC-PRC-001-001), `M` 6 |
| worktree R (green) | the ko labels rewritten to `"/moai run SPEC-ID"` and `"/moai sync SPEC-ID"` | `E03` 0, `E03W` 0, `SL` 2, `B[34]` 2, `B3L` 1, `B4L` 1 |
| D4 swap | the ko labels B3 `"/moai sync SPEC-ID"` and B4 `"/moai run SPEC-ID"` | `E03` 0, `SL` 2, `B[34]` 2 (commands 1 to 4 pass), `B3L` 0, `B4L` 0 (red) |
| D4 both run | the ko labels B3 and B4 both `"/moai run SPEC-ID"` | `E03` 0, `SL` 2, `B[34]` 2 (commands 1 to 4 pass), `B3L` 1, `B4L` 0 (red) |
| D4 trailing text | worktree R with `%% x` after the B3 label | `SL` 2 and `B[34]` 2 (pass), `B3L` 0 (red), `B4L` 1 |
| worktree removed (D) | the two ko lines deleted | `E03` 0, `SL` 0, `B[34]` 0, `B3L` 0, `B4L` 0: red under R, which is the alternate D that is not planned |
| retire line deleted | the en `prompt-caching.md` with its `moai cg` sentence deleted | `CG` prints nothing, exit 1 (the pin copy prints 1) |
| green, position (mcp, en and ko) | the green fixture of the first row, read by command 15 | stdout equals the pristine copy's output (7 lines per locale), exit 0: green |
| relocation (mcp, en and ko) | green, the `spec_progress` row deleted from line 130 and appended after the last line (line 320) | passes every other command: `U` 0, `M` 7, each `MID` 1, literal prefixes 7 (en), `ROW` 41, line bound 7 removed and 7 added. Command 15 red: `audit_cache` at 162, `codex_task` to `codex_job_cancel` at 181 to 185, `spec_progress` at 320 |
| reorder, adjacent (mcp, en and ko) | green, lines 182 and 183 swapped (`codex_task` and `codex_setup`) | passes every other command, same figures as the relocation row. Command 15 red: `codex_setup` at 182, `codex_task` at 183 |
| reorder, far (mcp, en) | green, lines 182 and 186 swapped (`codex_task` and `codex_job_cancel`) | passes every other command. Command 15 red: `codex_job_cancel` at 182, `codex_task` at 186 |
| shift down by one (mcp, en and ko) | green, lines 130 and 131 swapped (`spec_progress` moves below `spec_audit`) | passes every other command. Command 15 red: `spec_progress` at 131 |
| shift up by one (mcp, en) | green, lines 162 and 163 swapped (`audit_cache` moves above its neighbour) | passes every other command. Command 15 red: `audit_cache` at 162 |
| non-target pair swap (mcp, en) | green, lines 131 and 132 swapped (`spec_audit` and `spec_drift`) | command 15 prints the expected output (green); the line bound reads 8 removed and 8 added (red): a moved non-target row is caught by command 1, not by position |

Position fixtures, revision 0.2.1. The en and ko `mcp-server.md` were copied with `cp` from the working tree at e11f08f6a7 (clean, equal to HEAD; E-SWD-6c). A script run with `python3 -I` replaced the last cell of lines 130, 163, and 182 to 186 by the bare marker to make the green fixture, then built each mutant from it by deleting, appending, or swapping whole lines at the stated line numbers. Each fixture was read with `grep -c -E` for `U`, `M`, the seven `MID`, and `ROW`; `grep -c -F -e` with the seven en prefixes of command 12 (en only); `diff` against the pristine copy, counting removed and added lines, as the stand-in for the numstat of command 1; and `grep -n -o -E` with the pattern of command 15. Command 15's exit code was 0 on every fixture, because rows matched on each, so its verdict is the stdout compared with the pristine copy's. "Every other command" in the mutant rows means commands 1 to 14 of this revision. The relocation, the adjacent reorder, and the shift down were built on en and ko; the far reorder, the shift up, and the non-target swap on en only. The fixtures were read with `grep` and `diff`, not `git grep` and `git diff --numstat` (the same limit as G10).

### E-SWD-16 — sequencing and churn

Observed in revision 0.1.0 at the pin 74ad197380; not re-run. `git log -n 10 --format=%h --name-only` over `mcp-server.md`: 7 commits list files and each lists all four locales (f32c831de5, c932aa3582, d94e4763dd, 5629a0d6f8, 0de4cfabe2, 5c006bf9b2, 4a8ab71a79); 3 are merges listing none. Over `worktree/_index.md`: 7 commits list all four locales (ce79ef7caf, 024536e38c, b50c4de719, 5e2c832386, e55b48def9, 47f5bfde91, 5792fc755a), aa650dd08a and a723c41f63 list only ko, and 734ede821f lists en, ja, and zh. Over `prompt-caching.md`, of the 8 most recent commits 6 list files (c5620da649, 4ccb12e2cc, fbd13bbfc4, 81c10109ee, a29db12ec9, ce79ef7caf), each listing all four locales, and 2 are merges listing none. The umbrella records (spec.md lines 122 and 126) that the README child shares ko files with this child on different lines and that this child lands after it; no source in this pass resolves the card text's "card 4" to an id (G5).

### E-SWD-17 — decision-board check

Revision 0.2.0, 2026-10-19. Command: `moai decision read --scope standing --project-root <worktree root>`, output saved to a scratch file; exit 0; `grep -c "id=d-20261019T015908Z-220a"` on it prints `1`. Command: `moai decision read --scope card:t1724 --project-root <worktree root>`, output saved to a scratch file; exit 0; `grep -c "id=d-20261019T021509Z-4604"` on it prints `1`. The two records, copied verbatim, are in decision-index.md (Board check). Record 220a (decided_by=operator via AskUserQuestion, kind=ruling, scope=standing) answers Q13 and Q17; record 4604 (decided_by=leader under operator delegation, kind=ruling, scope=card:t1724) adopts the lane reading of Q13 for `mcp-server.md`, retains the four `prompt-caching.md` lines (Q-SWD-1), and leaves the adjacent Mermaid labels (Q-SWD-2). Record 4604 says in its own words that it is the leader reading of operator verdict 220a and not a new operator verdict. It does not name the Mermaid labels B3 and B4; branch R for them rests on the replace half of record 220a (the slash forms exist, plan.md §A.4) and on the leader's confirmation in the revision brief, which is an instruction and not a board record. At the read of revision 0.1.0 no record answered these rows (the one match for `Q13|Q17|c26e|7f98` was a record about card t1721).

## D. Gaps and residual risk

**Gaps (not observed in this pass):**

- **G1: the card's named source is absent.** The card cites `.moai/worktrees/t1655/.moai/reports/t1655/child-cards.md`; neither that path nor a hoisted copy exists (the lane's progress record, "Source read attempted and failed"). The E-01, E-02, and E-03 sets were derived from the files and cross-checked against the umbrella's committed counts (24 CLI lines for `mcp-server.md`, 8 Mermaid lines; umbrella acceptance.md E-02 and E-03 at its pin 8816de4dae).
- **G2: the Class N rulings are cited second-hand.** The retain rulings for the sibling children (leader ruling `d-20261018T201957Z-c26e` for SWEEPB Q-SWB-1, and `d-20261018T223420Z-7f98` on the start rule) are read from the SWEEPB decision-index in another worktree, not queried on the board. Record 4604 on this card names c26e as its precedent; it was read for this card and is the ruling of Q-SWD-1 (E-SWD-17).
- **G3: forms outside the families.** `worktree/_index.md` and `prompt-caching.md` cite other `moai` forms (`moai worktree done|sync|remove|clean|recover|snapshot|verify|restore`, `moai cc -w`, `moai glm -w`, with flags such as `--spawn`, `--strategy rebase`, `--merged-only`, `--delete-branch`). Only the top-level listing of `moai --help` was compared with them; their sub-verbs and flags were not probed, and no claim is made about them.
- **G4: no build of the tree.** The CLI probes ran on an installed build 68 commits behind the tree (E-SWD-0). The source read at the pin corroborates the registrations but is not a run of the tree's binary. E-SWD-10 and E-SWD-11 were not re-run in revision 0.2.0.
- **G5: other cards.** No other branch or worktree was scanned for edits to the three pages, and the card text's "card 4" was not resolved to an id.
- **G6: the i18n rules document.** The umbrella cites `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 for the four-locale rule. This pass did not re-read that document.
- **G7: Mermaid rendering.** No site build was run; the fixtures show the text outcome of a replacement, not that the diagram renders.
- **G8: verbs not executed.** `moai codex task`, `setup`, and `job` were judged absent from the listing and the source, not by running them, because running `moai codex task` would launch Codex (E-SWD-10).
- **G9: GATE card state.** The lane's brief says card t1717 is on plan-audit hold. That state was not re-measured, and nothing in this SPEC depends on it.
- **G10: no green run on the repository.** Nothing has been edited, so every green output in this file other than the preservation criteria and the preservation parts comes from a scratch fixture (E-SWD-15). The fixtures were read with `grep`, not `git grep`; the engine is the same POSIX ERE, and the two agreed on the repository for the pattern `M` (zero matches on the unedited page) and for the control `MC`, but the mutant results are not a `git grep` run on a committed tree.
- **G11: the literal prefixes are copied text.** The four prefix commands carry 28 strings copied from the pin; a copy error would show as a count below 7 at the pin, and each printed 7 (E-SWD-6b). The strings include the purpose and agents cells and so break on any later edit of those cells by another card, which is the intended behavior and also a re-pin trigger (plan.md M4, R1).
- **G12: the mutant set is the set tried, and position is read at a pinned coordinate.** The mutants are the ones the audits of iterations 1 and 2 named plus the blank, typo, removal, relocation, reorder, and shift cases. Before revision 0.2.1 this gap read as a move "elsewhere in the table"; the audit of iteration 2 observed it wider: a row deleted from the table and appended after the last line of the page, or the five `codex_*` rows reordered, passed every command. Command 15 now measures position; the relocation, reorder, and shift mutants print different lines from it (E-SWD-15). What stays open: (a) the expected line numbers are coordinates measured at e11f08f6a7 (E-SWD-6c); if the card branch absorbs a change that shifts these pages above line 130, command 15 reads red for a correct edit and the coordinates are re-measured at M4; (b) the statement that no row of the page can have moved rests on a derivation from commands 1, 2, and 15 and from AC-PRC-001-001, written out in AC-PRC-001-003 under the mutants that move a row, observed on the position mutants above and not tried against every permutation; (c) the dot in command 15 matches any one character where a row id carries a backtick, so a look-alike character in place of the backtick passes command 15 and is caught by `M` and the literal prefixes; (d) the fixtures were read with `grep` and `diff`, not `git grep` and `git diff --numstat`, and no mutant was run on a committed tree; (e) the order of rows as rendered in the browser is not measured (plan.md R2).

**Residual risk:** a green `U`, `UW`, `E03`, `E03W`, `M`, the literal prefixes, `B3L`, and `B4L` over the swept files proves the matched spellings and cells are as specified, not that the page still reads coherently (plan.md R2, R7), that a respelling outside the patterns is absent, or that a cell holding only the bare marker explains itself to a reader (plan.md R3). The reviewer reads the edited lines in all four locales.

## E. Definition of Done

- Q13, Q17, Q-SWD-1, and Q-SWD-2 each carry the operator verdict or the recorded leader ruling before the first file edit (decision-index.md, Gate: satisfied at revision 0.2.0, nothing open). The plan audit has passed and the plan→run Kickoff gate is met before the first file edit; neither is satisfied by this revision.
- AC-PRC-001-001 to AC-PRC-001-004 print their expected output at the sweep tip, in branch R, and their results are carried into `.moai/reports/t1724/verdict.md` with the command, the stdout, and the exit code.
- AC-PRC-001-005 and AC-PRC-001-006 print their expected output at the sweep tip. AC-PRC-001-005 applies on the retain ruling of Q-SWD-1; on a later verdict for the alternate the SPEC has returned to the plan phase and this item does not apply.
- `bash scripts/docs-i18n-check.sh` still prints `Errors:   0` and `Warnings: 0`.
- The umbrella's closing measurement is not part of this Definition of Done; it is the umbrella's M6.
