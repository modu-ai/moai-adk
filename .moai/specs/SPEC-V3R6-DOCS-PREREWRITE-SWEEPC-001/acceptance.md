# SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001 — Acceptance Criteria

## A. Conventions

### A.1 Pin, patterns, and pathspecs

Document-level pin: tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea` (HEAD of the card branch and local `main` at plan time, E-SWC-0). It binds every RED-now cell, control, and evidence entry in §C that carries no pin of its own. The commands in §B judge the sweep tip, so they read `HEAD` of the card branch and `main...HEAD`; at plan time `HEAD` is the pin.

The patterns and pathspecs below are defined once and used by name. `W01` is defined in spec.md Scope (the E-01 pattern with `[[:space:]]+`). `S` and `T-eq` are defined in spec.md Scope.

```
W01   = (^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)
P-CLI = 'docs-site/content/*/getting-started/cli.md'
P-WIN = 'docs-site/content/*/getting-started/windows-guide.md'
P-WHAT = 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
P-ALL = P-CLI P-WIN P-WHAT
```

### A.2 Command form

- Every command runs from the card worktree root. `git` stands at the head of a plain command, never inside `$(...)` or `<(...)`, because the worktree-session guard refuses those forms.
- A command that prints nothing is read by its exit status, recorded as its own field in §C (the Bash tool does not show an exit status for a silent command, so it was read with a measurement wrapper outside the cell and the entry says so). `git grep` exits 1 on no match and 0 on a match; `git diff --name-only` and `--numstat` exit 0 whether or not they print. Exit 128 or any other value means git failed and is never read as "no match".
- Commands live in fenced blocks, never in table cells, so a `|` is a regex alternation.
- The left end `main` in `main...HEAD` is deliberately not pinned: the claims are about this card's own contribution, so the left end is the merge-base with the absorbed local `main`, recomputed at read time. It cannot be used after the card has merged into `main`.

## B. Acceptance criteria

### B.0 Applicability and classification (stated once; nothing else restates it)

Which criteria apply depends on the open row Q-SWC-1 (decision-index.md, Gate). Q13 and Q17 carry the operator verdict `d-20261019T015908Z-220a` and are inputs to the criteria below, not branches.

| AC | Requirement | Under Q-SWC-1 (a), edit | Under Q-SWC-1 (b), retain | RED-now cell | Flips at |
|----|-------------|-------------------------|---------------------------|--------------|----------|
| AC-PRC-001-001 | REQ-PRC-001-001 | release-blocking; red now | not applicable | E-SWC-12 | M2 |
| AC-PRC-001-002 | REQ-PRC-001-005 (within the Edit unit) | release-blocking; red now | not applicable | E-SWC-13 | M2 |
| AC-PRC-001-003 | REQ-PRC-001-004 | release-blocking; red now | not applicable | E-SWC-14 | M2 |
| AC-PRC-001-004 | REQ-PRC-001-003 | release-blocking; red now | not applicable | E-SWC-15 | M2 |
| AC-PRC-001-005 | REQ-PRC-001-002 | not applicable | regression-guard (true at the pin; control E-SWC-16a) | none can be red | M2 |
| AC-PRC-001-006 | REQ-PRC-001-005 (outside the twelve) | regression-guard (true at the pin; control E-SWC-16b) | regression-guard (same) | none can be red | M2 |

A regression-guard is true before any work, so its adoption rests on its control (verification-completeness §2): a historical input that turns the criterion's own command form red. A guard is not recorded as a pass of the work.

### AC-PRC-001-001 — no line of the twelve files matches W01 (REQ-PRC-001-001)

- **Given** the card branch tip,
- **When** command 1 and command 2 run,
- **Then** command 1 prints nothing and exits 1, and command 2 prints 12 paths (the sweep did not delete a file).

```
git grep -c -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/getting-started/cli.md' 'docs-site/content/*/getting-started/windows-guide.md' 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
git ls-files -- 'docs-site/content/*/getting-started/cli.md' 'docs-site/content/*/getting-started/windows-guide.md' 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
```

RED now (E-SWC-12): command 1 prints 12 lines and exits 0, because the twelve notice lines are still there. It is red for the stated reason that no edit exists; it flips at M2. The criterion is an end-state check: fewer matches caused by another card's landing (the README child) also satisfy it, which is intended. A locale left unedited keeps a match, so a partial-locale edit fails here.

### AC-PRC-001-002 — each edit stays inside its Edit unit (REQ-PRC-001-005)

- **Given** the card branch,
- **When** command 1 runs and its printed diff (command 2) is read,
- **Then** command 1 prints one line per file of S, each with removed column 1 to 2 and added column 0 (deleted line) or 1 to 2 (replaced line); and every removed line printed by command 2 is either blank or a W01 match.

```
git diff --numstat main...HEAD -- 'docs-site/content/*/getting-started/cli.md' 'docs-site/content/*/getting-started/windows-guide.md' 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
git diff -U0 main...HEAD -- 'docs-site/content/*/getting-started/cli.md' 'docs-site/content/*/getting-started/windows-guide.md' 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
```

The bounds are derived from the Edit unit (spec.md Scope): the matched line, plus one blank line on deletion. Command 2 is a reviewer read of at most a few dozen printed lines; the numeric bounds are the machine check.

RED now (E-SWC-13): command 1 prints nothing and exits 0, because no edit exists, so no line of the expected set is present. It flips at M2. Mutants: blanking or truncating a page prints a removed column far above 2 (control E-SWC-16c); changing a second, unrelated line inside the two-line allowance passes the numbers and is caught only by the reader of command 2 (Residual risk).

### AC-PRC-001-003 — the changed files are exactly S, in every locale (REQ-PRC-001-004)

- **Given** the card branch,
- **When** the two commands run,
- **Then** the path list printed by command 1 is the same set as the paths printed by command 2 (ref prefix removed), and that set is non-empty.

```
git diff --name-only main...HEAD -- 'docs-site/content/*/getting-started/cli.md' 'docs-site/content/*/getting-started/windows-guide.md' 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
git grep -l -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" main -- 'docs-site/content/*/getting-started/cli.md' 'docs-site/content/*/getting-started/windows-guide.md' 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
```

This is the Q17 criterion: a page edited in some locales only prints a shorter list than S and fails; an edit to a file outside S prints a longer one. If the README child has made a Korean file compliant, that file is not in S and is not required.

RED now (E-SWC-14): command 1 prints nothing and exits 0; command 2 prints 12 paths and exits 0 (E-SWC-7). The sets differ by the whole of S, because no edit exists. It flips at M2. Control E-SWC-16d: the same command 1 form on commit 609ac419eb prints three paths of a page that has four locale copies, so a partial-locale change shows as a short list.

### AC-PRC-001-004 — each line is replaced or deleted as T-eq says (REQ-PRC-001-003)

- **Given** the card branch and the T-eq result per line in plan.md §A.2,
- **When** the command runs and its list is compared with the numstat of AC-PRC-001-002,
- **Then** the printed list equals the set of files of S whose T-eq result is replaceable, and the added column of AC-PRC-001-002 is 1 to 2 exactly on those files and 0 on every other file of S.

```
git diff --name-only -G"moai migrate cg" main...HEAD -- 'docs-site/content/*/getting-started/cli.md' 'docs-site/content/*/getting-started/windows-guide.md' 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
```

`-G` lists a file only when a changed line of its diff contains the string; none of the twelve removed lines contains it (E-SWC-2), so a listed file has an added line that cites R. At the pin all twelve results are replaceable, so the expected list is S itself. If open item OI-1 turns the results to deletable, only the result column of plan.md §A.2 changes and this criterion's expected list becomes empty.

RED now (E-SWC-15): the command prints nothing and exits 0, because no edit exists, while the expected list is the twelve paths. It flips at M2. Mutants: deleting a replaceable line leaves its file off the list; a replacement that keeps the W01 form fails AC-PRC-001-001; a replacement that does not cite R leaves its file off the list.

### AC-PRC-001-005 — this card's diff touches none of the twelve files (REQ-PRC-001-002)

- **Given** the card branch under a retained verdict,
- **When** the command runs,
- **Then** it prints nothing.

```
git diff --name-only main...HEAD -- 'docs-site/content/*/getting-started/cli.md' 'docs-site/content/*/getting-started/windows-guide.md' 'docs-site/content/*/core-concepts/what-is-moai-adk.md'
```

True at the pin (E-SWC-16a, first entry: empty, exit 0), so it is a regression-guard. Control E-SWC-16a, second entry: the same form on commit e4fb8c3ca3 prints four paths. A change landed on `main` by another card and absorbed here is not this card's contribution and does not move this command.

### AC-PRC-001-006 — nothing outside the twelve files changes (REQ-PRC-001-005)

- **Given** the card branch,
- **When** the command runs,
- **Then** it prints nothing.

```
git diff --name-only main...HEAD -- . ':!docs-site/content/*/getting-started/cli.md' ':!docs-site/content/*/getting-started/windows-guide.md' ':!docs-site/content/*/core-concepts/what-is-moai-adk.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001' ':!.moai/reports/t1723'
```

True at the pin, so a regression-guard. Control E-SWC-16b: the same form on commit c024772b98 prints paths outside the twelve.

## C. Evidence ledger

Each entry gives the command, the exit code as its own field, and the stdout verbatim in a fenced block unless marked. Entries read from a command that prints nothing say how the exit status was read.

### E-SWC-0 — pin and tool provenance

- `git rev-parse HEAD` — exit 0 — `74ad197380a0711cdb9816d4e37c94d29af730ea`
- `git rev-parse main` — exit 0 — `74ad197380a0711cdb9816d4e37c94d29af730ea`
- `git status --short` — exit 0 — no output (clean at entry)
- `moai version` (version line) — `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z
- `git merge-base --is-ancestor 7705b365e2 HEAD` — exit 0
- `git rev-list --count 7705b365e2..HEAD` — exit 0 — `68`

The installed build is a strict ancestor of the tree HEAD, 68 commits behind. No Go build was made; CLI claims are attributed to "installed build 7705b365e2 against tree 74ad197380".

### E-SWC-1 — the three pathspecs expand to twelve paths

Command: `git ls-files -- <P-ALL>`. Exit 0. Output, 12 paths:

```
docs-site/content/en/core-concepts/what-is-moai-adk.md
docs-site/content/en/getting-started/cli.md
docs-site/content/en/getting-started/windows-guide.md
docs-site/content/ja/core-concepts/what-is-moai-adk.md
docs-site/content/ja/getting-started/cli.md
docs-site/content/ja/getting-started/windows-guide.md
docs-site/content/ko/core-concepts/what-is-moai-adk.md
docs-site/content/ko/getting-started/cli.md
docs-site/content/ko/getting-started/windows-guide.md
docs-site/content/zh/core-concepts/what-is-moai-adk.md
docs-site/content/zh/getting-started/cli.md
docs-site/content/zh/getting-started/windows-guide.md
```

### E-SWC-2 — the twelve hit lines (verbatim), umbrella pattern E-01 over the twelve files

Command: `git grep -n -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 74ad197380a0711cdb9816d4e37c94d29af730ea -- <P-ALL>`. Exit code: 0. Output (the line numbers are the Line column of plan.md §A.2):

```
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/en/core-concepts/what-is-moai-adk.md:439:`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session.
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/en/getting-started/cli.md:336:`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session. [CG retirement and migration](/en/multi-llm/cg-mode/)
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/en/getting-started/windows-guide.md:108:`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session.
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/ja/core-concepts/what-is-moai-adk.md:442:`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/ja/getting-started/cli.md:336:`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。 [CG の廃止と設定の移行](/ja/multi-llm/cg-mode/)
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/ja/getting-started/windows-guide.md:125:`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/ko/core-concepts/what-is-moai-adk.md:441:`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다.
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/ko/getting-started/cli.md:338:`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다. [CG 폐기와 설정 이전](/ko/multi-llm/cg-mode/)
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/ko/getting-started/windows-guide.md:108:`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다.
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/zh/core-concepts/what-is-moai-adk.md:440:`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/zh/getting-started/cli.md:336:`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。 [CG 停用与配置迁移](/zh/multi-llm/cg-mode/)
74ad197380a0711cdb9816d4e37c94d29af730ea:docs-site/content/zh/getting-started/windows-guide.md:113:`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。
```

Reading: 12 matches in 12 files, one per file. Each says `moai cg` has been retired; none instructs the reader to run it (Class N). No line contains the string `migrate`.

### E-SWC-3 — W01 over the twelve files (per-file counts)

Command: the AC-PRC-001-001 command 1 (`git grep -c -E <W01> HEAD -- <P-ALL>`). Exit code: 0. Output: twelve lines `HEAD:<path>:1`, one per path of E-SWC-1 (shown verbatim in E-SWC-12). W01 adds no match beyond E-01 on these files. Summary count for E-01: 12 lines in 12 files (E-SWC-2).

### E-SWC-4 — the other umbrella families over the twelve files

Command: `git grep -c -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)|(^|[^/])moai (run|sync) SPEC-ID|/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 74ad197380a0711cdb9816d4e37c94d29af730ea -- <P-ALL>`. Exit code: 1. Output: none. The families E-02, E-03, and E-04 match nothing in the twelve files.

### E-SWC-5 — line counts of the twelve files

Command: `git grep -c "" 74ad197380a0711cdb9816d4e37c94d29af730ea -- <P-ALL>`. Exit code: 0. Output (path suffix, count): en what-is 609, en cli 488, en windows 158, ja what-is 612, ja cli 485, ja windows 144, ko what-is 611, ko cli 490, ko windows 157, zh what-is 610, zh cli 488, zh windows 132. A record of the pin, not a criterion.

### E-SWC-6 — CLI surface, provenance, and source corroboration

Installed build 7705b365e2 against tree 74ad197380.

```
moai cg --help       -> exit 1   moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration
moai migrate --help  -> lists    cg [--flags]   Preview migration of legacy CG teammate roles
moai instance --help -> Unknown command "instance" for "moai"
moai inventory --help, moai web --help -> usage printed
```

Source at the pin: `git grep -n -E "Use: +\"(cg|gpt|codemaps|model)\"" 74ad197380a0711cdb9816d4e37c94d29af730ea -- internal cmd` prints `internal/cli/cg.go:13` (`Use: "cg"`, `Short: "Retired CG launcher"`, `Hidden: true`) and `internal/cli/migrate_cg.go:22` (`Use: "cg"`, `Short: "Preview migration of legacy CG teammate roles"`). `git grep -n -F "migrate cg" 74ad197380a0711cdb9816d4e37c94d29af730ea -- internal/cli/cg.go internal/cli/migrate_cg.go` prints `internal/cli/cg.go:9` with `errors.New("moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration")`. These are the two conditions of T-eq.

### E-SWC-7 — the sweep-start set S at the pin

Command: `git grep -l -E <W01> main -- <P-ALL>`. Exit code: 0. Output: the twelve paths of E-SWC-1 (`wc -l` of the same command printed 12, in a measurement wrapper outside the cell). S is all twelve at the pin.

### E-SWC-8 — overlap with the README child (Korean pages)

Command: `git grep -n -F "설정 이전 안내와 함께 종료합니다" 74ad197380a0711cdb9816d4e37c94d29af730ea -- docs-site/content/ko/getting-started/cli.md docs-site/content/ko/getting-started/windows-guide.md docs-site/content/ko/core-concepts/what-is-moai-adk.md`. Exit code: 0. Output (first 110 characters of each line): hits at `ko/core-concepts/what-is-moai-adk.md:441`, `ko/getting-started/cli.md:338`, and `ko/getting-started/windows-guide.md:108`, the same three lines as the Korean rows of E-SWC-2. The umbrella's E-08 (`git grep -c -F` of this sentence over `docs-site/content/ko`) lists these three files among the nine Korean pages that carry the sentence. The card's "different lines" claim is therefore not true for these three Korean lines.

### E-SWC-9 — other SPECs on disk

Command: `grep -rl -E "getting-started/cli|getting-started/windows-guide|core-concepts/what-is-moai-adk" <the GATE, REDIR, SWEEPA, and SWEEPB SPEC directories under the sibling worktrees>`. Exit code: 0. Output: two files of the GATE spec (`decision-index.md` and `acceptance.md`); their matching lines concern the link-check job's relative links on the `getting-started/cli` pages and not an E-01 line (read at lines 93, 468, 470). The REDIR, SWEEPA, and SWEEPB specs name none of the twelve files.

### E-SWC-10 — churn

Command: `git log --oneline -n 8 74ad197380a0711cdb9816d4e37c94d29af730ea -- <P-ALL>`. Exit code: 0. Output: df7402f4c4, e4fb8c3ca3, c024772b98, 7d3b2be5be, 609ac419eb, 9a53efd246, a9d9779d92, 21d2afc090.

### E-SWC-11 — decision board

Command: `moai decision read` (saved to the session scratchpad, 1870 lines in the first read). Exit code: 0. Record `d-20261019T015908Z-220a` is present (`decided_by=operator (AskUserQuestion 10-19)`, `scope=standing`, `kind=ruling`); its body is quoted in decision-index.md and names t1723 SWEEPC. A search of the first read for `t1723`, `SWEEPC`, and `Q-SWC` found no match (that read predates the record; the record was read again in a second read). Records `d-20261018T201957Z-c26e` (scope `card:t1722`) and `d-20261018T202324Z-cb47` (scope `card:t1721`) were read; neither is scoped to this card.

### E-SWC-12 — RED-now cell for AC-PRC-001-001

- Command: `git grep -c -E <W01> HEAD -- <P-ALL>` (the AC-PRC-001-001 command 1)
- Exit code: 0. Tree: HEAD = `74ad197380a0711cdb9816d4e37c94d29af730ea`.
- Stdout (verbatim; 12 lines):

```
HEAD:docs-site/content/en/core-concepts/what-is-moai-adk.md:1
HEAD:docs-site/content/en/getting-started/cli.md:1
HEAD:docs-site/content/en/getting-started/windows-guide.md:1
HEAD:docs-site/content/ja/core-concepts/what-is-moai-adk.md:1
HEAD:docs-site/content/ja/getting-started/cli.md:1
HEAD:docs-site/content/ja/getting-started/windows-guide.md:1
HEAD:docs-site/content/ko/core-concepts/what-is-moai-adk.md:1
HEAD:docs-site/content/ko/getting-started/cli.md:1
HEAD:docs-site/content/ko/getting-started/windows-guide.md:1
HEAD:docs-site/content/zh/core-concepts/what-is-moai-adk.md:1
HEAD:docs-site/content/zh/getting-started/cli.md:1
HEAD:docs-site/content/zh/getting-started/windows-guide.md:1
```

- Why red: the twelve notice lines have not been edited. Green path: M2; the output becomes empty with exit 1.

### E-SWC-13 — RED-now cell for AC-PRC-001-002

- Command: `git diff --numstat main...HEAD -- <P-ALL>`
- Exit code: 0 (read with a wrapper outside the cell). Tree: HEAD = `74ad197380a0711cdb9816d4e37c94d29af730ea`.
- Stdout: empty.
- Why red: no edit exists, so none of the twelve expected numstat lines is present. Green path: M2; the output becomes twelve lines `<added> <removed> <path>` within the bounds of AC-PRC-001-002.

### E-SWC-14 — RED-now cell for AC-PRC-001-003

- Command: `git diff --name-only main...HEAD -- <P-ALL>` (command 1), with command 2 = E-SWC-7
- Exit code: 0 (command 1, read with a wrapper outside the cell); 0 (command 2). Tree: HEAD = `74ad197380a0711cdb9816d4e37c94d29af730ea`.
- Stdout: command 1 empty; command 2 the twelve paths of E-SWC-7.
- Why red: the two sets differ by all of S because no edit exists. Green path: M2; command 1 prints the same paths as command 2.

### E-SWC-15 — RED-now cell for AC-PRC-001-004

- Command: `git diff --name-only -G"moai migrate cg" main...HEAD -- <P-ALL>`
- Exit code: 0 (read with a wrapper outside the cell). Tree: HEAD = `74ad197380a0711cdb9816d4e37c94d29af730ea`.
- Stdout: empty.
- Why red: the expected list is the twelve replaceable files (plan.md §A.2) and no edit exists. Green path: M2; the output becomes the twelve paths.

### E-SWC-16 — controls (each runs its own command form on a named historical input)

- **E-SWC-16a (AC-PRC-001-005).** `git diff --name-only main...HEAD -- <P-ALL>` at the pin: empty, exit 0 (true at the pin). Control: `git diff --name-only e4fb8c3ca3~1...e4fb8c3ca3 -- <P-ALL>` prints `docs-site/content/{en,ja,ko,zh}/getting-started/cli.md` (four paths in the numstat of the same commit, `11 0` each), exit 0.
- **E-SWC-16b (AC-PRC-001-006).** `git diff --name-only c024772b98~1...c024772b98 -- . ':!<P-CLI>' ':!<P-WIN>' ':!<P-WHAT>'` prints paths; its first lines are `.moai/specs/SPEC-MANAGER-TODO-001/progress.md` and `docs-site/content/en/advanced/agent-guide.md` (output cut to five lines for the control), exit 0.
- **E-SWC-16c (AC-PRC-001-002).** `git diff --numstat 609ac419eb~1...609ac419eb -- <P-ALL>` prints `1 1` for the ja, ko, and zh `what-is-moai-adk.md` copies, exit 0; the form therefore prints one line per changed file. A page-blanking commit was not searched for (G4); the removed-column bound is checked by construction.
- **E-SWC-16d (AC-PRC-001-003).** `git diff --name-only 609ac419eb~1...609ac419eb -- <P-ALL>` prints the ja, ko, and zh `what-is-moai-adk.md` paths and not the en path, exit 0: three of four locales.
- **E-SWC-16e (W01 discrimination, AC-PRC-001-001).** `grep -n -E <W01>` over four fixture lines (printed by `printf`: a double-space `moai  cg`, a tab `moai<TAB>cg`, `use \`moai migrate cg\` now`, `Run \`moai cg\` now`) printed lines 1, 2, and 4 and not line 3, exit 0. A respelling is caught; the replacement that cites `moai migrate cg` is not matched.
- **E-SWC-16f (AC-PRC-001-001 green form).** `git grep -c -E <W01> HEAD -- docs-site/content/en/getting-started/faq.md` prints nothing, exit 1: the green form on a file without a match.

## D. Gaps, mutant probe, and residual risk

**Gaps (not observed in this pass):**

- **G1: the card's named sources are absent.** `child-cards.md` and `split-proposal.md` of the t1655 tree do not exist: the t1655 tree is gone (`ls` of both paths: No such file or directory; `find` under the worktrees found no `child-cards.md` for t1655), and the only `split-proposal.md` files are under other cards. Nothing was taken from them. The file list, the E-01 pattern, Q13, and Q17 come from the umbrella SPEC committed at bdb7b764 (the same gap SWEEPA recorded as its G1).
- **G2: the README child has no SPEC on disk.** The overlap with its REQ-PRC-004 rests on the umbrella's E-08 and on the text of the three Korean lines (E-SWC-8); its actual edit set is unread, and no trial merge was made.
- **G3: mutants were analysed, not built.** The worktree-isolation guard refuses git commands outside this tree, so no scratch repository was created. The mutant table below rests on the named historical inputs (E-SWC-16) and on reading the command forms.
- **G4: no page-blanking commit** was searched for, so AC-PRC-001-002's removed-column bound has no observed red on a real blanking input.
- **G5: forms outside the four families.** The first-verb list of plan.md §A.3 is a candidate list; sub-verbs and flags were not probed, and the one `moai instance` match was not read.
- **G6: locales read.** The neighbourhood of the hit was read in full only for the English pages and as hit lines for the other locales; the line numbers of ja, ko, and zh outside the hit were not read.
- **G7: decision-board scope.** The board was read in full twice (E-SWC-11); only the first read was searched for `t1723`, `SWEEPC`, and `Q-SWC`. A refused command: a `sed` on the saved board with a variable path was refused by the worktree guard; it was rerun with a literal path, and nothing was inferred in its place.
- **G8: tool provenance.** All CLI probes ran on a build 68 commits behind the tree (plan.md A.4). The source read at the pin corroborates both T-eq conditions.
- **G9: parity gate.** `bash scripts/docs-i18n-check.sh` was not run, and `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 were taken from the umbrella, not re-read.

**Mutant probe** (a mutant satisfies one criterion while violating the requirement only if the right-hand column does not catch it):

| Mutant | Caught by | Basis |
|--------|-----------|-------|
| (a) edit a subset of the locales only | AC-PRC-001-001 (remaining matches) and AC-PRC-001-003 (shorter list) | E-SWC-16d |
| (b) blank or truncate a page | AC-PRC-001-002 (removed column above 2); AC-PRC-001-001 alone passes it | construction; G4 |
| (c) respell `moai  cg` or `moai<TAB>cg` | AC-PRC-001-001 (W01) | E-SWC-16e |
| (d) delete a line whose T-eq result is replaceable | AC-PRC-001-004 (file missing from the list) | reading |
| (e) replace a line with text that keeps the W01 form | AC-PRC-001-001 | E-SWC-16e |
| (f) replace a line with text that does not cite R | AC-PRC-001-004 (file missing from the list) | reading |
| (g) edit a file outside the twelve | AC-PRC-001-006 | E-SWC-16b |
| (h) a change landed on `main` by another card and absorbed | must not fail AC-PRC-001-002 to -006 (merge-base form) | construction |
| (i) edit one more, unrelated line inside the two-line allowance | the reader of AC-PRC-001-002 command 2 only | Residual risk |
| (j) replace the line with text that cites R but drops the information that the launcher is retired | no criterion | Residual risk |

**Residual risk:** a green result proves that no W01 line remains, that every locale in S changed together within its Edit unit, and that each replaced line cites `moai migrate cg`. It does not prove the page still reads coherently in four languages (plan.md R2), that a replacement keeps the retirement information (mutant j), or that the operator accepts T-eq as the reading of "equivalent" (decision-index.md, OI-1).

## E. Definition of Done

- Under Q-SWC-1 = (a): AC-PRC-001-001 to -004 print their expected output at the sweep tip, AC-PRC-001-006 prints nothing, and the deciding lines are carried into `.moai/reports/t1723/verdict.md` with the command and output.
- Under Q-SWC-1 = (b): AC-PRC-001-005 and AC-PRC-001-006 print nothing, and the card closes as rationale recorded plus zero page changes.
- The state of Q13, Q17, and Q-SWC-1 is as decision-index.md (Gate) states; this file adds no second reading.
- The umbrella's closing measurement is not part of this Definition of Done.
