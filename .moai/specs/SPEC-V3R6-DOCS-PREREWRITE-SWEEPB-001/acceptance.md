# SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001 — Acceptance Criteria

## A. Conventions

### A.1 Pin, patterns, and pathspecs

Document-level pin: tree commit `2bcce5d9dc4c5f57bb98bcfa553e0061fa373681`. It binds every RED-now cell, control, and evidence entry in §C that carries no pin of its own. The commands in §B judge the sweep tip, so they read `HEAD` of the card branch, not the pin.

The patterns and pathspecs below are defined once and used by name. `E01` is the umbrella's family F1 pattern. `W01` is `E01` with `[[:space:]]+` in place of the single space; it is the mutant guard, because a stale citation respelled with two spaces passes `E01` and still tells the reader to run an absent verb. `P-ALL` is the three pathspecs together and reaches exactly 12 files.

```
E01 = (^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)
W01 = (^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)
P-TOK = 'docs-site/content/*/advanced/tokenomics-overview.md'
P-IDX = 'docs-site/content/*/multi-llm/_index.md'
P-POL = 'docs-site/content/*/multi-llm/model-policy.md'
P-ALL = P-TOK P-IDX P-POL
```

### A.2 Command form

- Every command runs from the card worktree root. `git` stands at the head of a plain pipeline or of a `&&` / `||` chain, never inside `$(...)` or `<(...)`, because the worktree-session guard refuses those forms.
- A quiet check is followed by `; echo "exit=$?"`, because the Bash tool does not display the exit status of a command that prints nothing. `exit=0` means a match (or a difference), `exit=1` means none, and `exit=128` or any other value means git failed. A failure is never read as "no match". A count or a path list is read directly. A pipeline that ends in a filter is paired with a swept-set control (`git ls-files` printing 12 paths) so that a git error cannot read as a pass.
- Commands live in fenced blocks, never in table cells, so an alternation `|` is a regex alternation.
- The left end `main` in `main...HEAD` is deliberately not pinned: the claim is about this card's own contribution, so the left end is the merge-base with the absorbed local `main`, recomputed at read time. It cannot be used after the card has merged into `main`.

### A.3 Classification

Q-SWB-1 is decided (b) on board record `d-20261018T201957Z-c26e` (decision-index.md, Gate), so the sweep edits no page and every criterion below is true at the pin. A criterion that is true before any work is vacuous unless it carries a control (verification-completeness §2); each one does, so all five are **regression-guards**. None is release-blocking: the card closes as rationale recorded plus zero page changes. Each control is a command run on a known historical input that turns the same form red. The old option (a) path (edit the twelve units) is not applicable and has no criterion; its text survives only in the Q-SWB-1 options for the record.

## B. Acceptance criteria

| AC | Requirement | Class | State at the pin | Control | Milestone |
|----|-------------|-------|------------------|---------|-----------|
| AC-PRC-001-001 | REQ-PRC-001-001 | regression-guard | true (no output) | E-SWB-10a | M2 |
| AC-PRC-001-002 | REQ-PRC-001-002 | regression-guard | true (12 lines, `exit=0`) | E-SWB-10b | M2 |
| AC-PRC-001-003 | REQ-PRC-001-003 | regression-guard | true (4 lines per page) | E-SWB-10c | M2 |
| AC-PRC-001-004 | REQ-PRC-001-004 | regression-guard | true (no output) | E-SWB-10d | M2 |
| AC-PRC-001-005 | REQ-PRC-001-002 | regression-guard | true (zero diff, 12 base counts) | E-SWB-10e, E-SWB-7 | M2 |

### AC-PRC-001-001 — no line tells the reader to run an unresolved form

AC-PRC-001-001 maps REQ-PRC-001-001.

- **Given** the tip is `HEAD` of the card branch,
- **When** these two commands run,
- **Then** the first prints nothing and the second prints exactly 12 paths.

```
git grep -n -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/tokenomics-overview.md' 'docs-site/content/*/multi-llm/_index.md' 'docs-site/content/*/multi-llm/model-policy.md' | rg -v -F -e "llm.team_mode: cg" -e "moai model profile"
git ls-files -- 'docs-site/content/*/advanced/tokenomics-overview.md' 'docs-site/content/*/multi-llm/_index.md' 'docs-site/content/*/multi-llm/model-policy.md'
```

The first command prints every W01 match line that carries neither content key of the Class N set (the `llm.team_mode: cg` token that ends every notice, and `moai model profile`); a surviving line is a Class C candidate or an unknown form. The second is the swept-set control. Control (E-SWB-10a): the same filter over `bas-navigator.md`, a page with a real Class C line, prints that line.

### AC-PRC-001-002 — the twelve retained lines are present and the files are unchanged

AC-PRC-001-002 maps REQ-PRC-001-002.

- **Given** the tip,
- **When** these two commands run,
- **Then** the first prints exactly 12 lines, each ending `:1`, and the second prints `exit=0` (no difference from `main`).

```
git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/tokenomics-overview.md' 'docs-site/content/*/multi-llm/_index.md' 'docs-site/content/*/multi-llm/model-policy.md'
git diff --quiet main...HEAD -- 'docs-site/content/*/advanced/tokenomics-overview.md' 'docs-site/content/*/multi-llm/_index.md' 'docs-site/content/*/multi-llm/model-policy.md'; echo "exit=$?"
```

True at the pin (E-SWB-2). Control (E-SWB-10b): the second command prints `exit=1` over a commit that does change these pages. The first command is the swept-count control: a count other than 12 lines means a retained line was removed or a new match added.

### AC-PRC-001-003 — the four locales keep the same citation on every page

AC-PRC-001-003 maps REQ-PRC-001-003.

- **Given** the tip,
- **When** the three commands run, one per page,
- **Then** each prints exactly four lines, one per locale, each ending `:1`.

```
git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/tokenomics-overview.md'
git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/multi-llm/_index.md'
git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/multi-llm/model-policy.md'
```

Control (E-SWB-10c): commit 6581cdccbb changed only the en and ko copies of `model-policy.md`, so the diff-path form prints two paths over that commit, showing the locale-set form can detect a partial-locale change.

### AC-PRC-001-004 — nothing outside the twelve files changes

AC-PRC-001-004 maps REQ-PRC-001-004.

- **Given** the tip,
- **When** this command runs,
- **Then** it prints nothing.

```
git diff --name-only main...HEAD -- . ':!docs-site/content/*/advanced/tokenomics-overview.md' ':!docs-site/content/*/multi-llm/_index.md' ':!docs-site/content/*/multi-llm/model-policy.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001' ':!.moai/reports/t1722'
```

Control (E-SWB-10d): the same exclusion form over the range of commit c5620da649 prints 13 paths.

### AC-PRC-001-005 — the twelve files are byte-identical to the pin

AC-PRC-001-005 maps REQ-PRC-001-002 (byte-unchanged clause). It closes the emptied-page mutant: the line counts and the numstat catch a blanked or truncated page that still contains its E01 line.

- **Given** the tip,
- **When** these two commands run,
- **Then** the first prints nothing and the second prints the 12 base counts of plan.md §A.2 (E-SWB-7).

```
git diff --numstat main...HEAD -- 'docs-site/content/*/advanced/tokenomics-overview.md' 'docs-site/content/*/multi-llm/_index.md' 'docs-site/content/*/multi-llm/model-policy.md'
git grep -c "" HEAD -- 'docs-site/content/*/advanced/tokenomics-overview.md' 'docs-site/content/*/multi-llm/_index.md' 'docs-site/content/*/multi-llm/model-policy.md'
```

Control (E-SWB-10e): over the range of commit 4716869503 the first command prints `162 54`, `156 41`, `141 39` for `_index.md`, so a page edit shows as non-empty output. If a page moves on `main` before the run, the base counts are recomputed from the run-start tree and recorded in the verdict.

## C. Evidence ledger

Each entry gives the command, the observed output (verbatim unless marked), and the tree it read. Where the Bash tool shows no error status on a command that prints output, the exit code is 0 (grep and git return non-zero on no match and the tool would display it); where it printed nothing, the entry says `exit=N` from a measurement wrapper of §A.2, run outside the cell.

### E-SWB-0 — pin and tool provenance

```
git rev-parse HEAD                    -> 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681
git rev-parse main                    -> 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681
git status --short                    -> (no output)
moai version (version line)           -> moai_cp/20261007_195927   moai_cp/20261007_195927-1888-g7705b365e2   built 2026-10-18T10:45:08Z
git merge-base --is-ancestor 7705b365e2 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 ; echo "exit=$?"   -> exit=0
git rev-list --count 7705b365e2..2bcce5d9dc4c5f57bb98bcfa553e0061fa373681   -> 48
```

The installed build is a strict ancestor of the tree HEAD, 48 commits behind. No Go build was made; CLI claims are attributed to "installed build 7705b365e2 against tree 2bcce5d9dc" (plan.md §A.4).

### E-SWB-1 — the three pathspecs expand to twelve paths

Command: `git ls-files -- 'docs-site/content/*/advanced/tokenomics-overview.md' 'docs-site/content/*/multi-llm/_index.md' 'docs-site/content/*/multi-llm/model-policy.md'`. Exit 0. Output, 12 paths:

```
docs-site/content/en/advanced/tokenomics-overview.md
docs-site/content/en/multi-llm/_index.md
docs-site/content/en/multi-llm/model-policy.md
docs-site/content/ja/advanced/tokenomics-overview.md
docs-site/content/ja/multi-llm/_index.md
docs-site/content/ja/multi-llm/model-policy.md
docs-site/content/ko/advanced/tokenomics-overview.md
docs-site/content/ko/multi-llm/_index.md
docs-site/content/ko/multi-llm/model-policy.md
docs-site/content/zh/advanced/tokenomics-overview.md
docs-site/content/zh/multi-llm/_index.md
docs-site/content/zh/multi-llm/model-policy.md
```

### E-SWB-2 — baseline (the retained set), pattern E01 over the twelve files

Command: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- 'docs-site/content/*/advanced/tokenomics-overview.md' 'docs-site/content/*/multi-llm/_index.md' 'docs-site/content/*/multi-llm/model-policy.md'`. Exit code: 0. Output:

```
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/en/advanced/tokenomics-overview.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/en/multi-llm/_index.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/en/multi-llm/model-policy.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ja/advanced/tokenomics-overview.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ja/multi-llm/_index.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ja/multi-llm/model-policy.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ko/advanced/tokenomics-overview.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ko/multi-llm/_index.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ko/multi-llm/model-policy.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/zh/advanced/tokenomics-overview.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/zh/multi-llm/_index.md:1
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/zh/multi-llm/model-policy.md:1
```

Reading: each of the twelve files carries one line that matches E01. Under the ruling (b) these twelve lines are the retained set and no criterion is red at the pin. All twelve are Class N (E-SWB-5, E-SWB-6).

### E-SWB-3 — pattern W01 over the twelve files

Command: the same as E-SWB-2 with W01 in place of E01. Exit code: 0. Output: the same twelve lines, line for line (observed in this run). The mutant guard adds no match today and closes only the respelling route.

### E-SWB-4 — quiet verdict word, and the other umbrella families

Command A: `git grep -q -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- <P-ALL>; echo "exit=$?"` — Output: `exit=0` (a match), observed in this run.

Command B: `git grep -c -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)|(^|[^/])moai (run|sync) SPEC-ID|/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- <P-ALL>; echo "exit=$?"`. Output: `exit=1` (no match). Slash forms: `git grep -n -o -E "/moai[: ][a-z][a-z0-9-]*" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- <P-ALL>` printed nothing; the three pages cite no slash command. The umbrella's families E-02, E-03, and E-04 therefore match nothing here.

### E-SWB-5 — the twelve hit lines (verbatim)

Command: `git grep -n -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- <P-ALL>`. Exit code: 0. Output, line numbers are the Hit column of plan.md §A.2:

```
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/en/advanced/tokenomics-overview.md:106:`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session. [CG retirement and migration](/en/multi-llm/cg-mode/)
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/en/multi-llm/_index.md:172:`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session. [CG retirement and migration](/en/multi-llm/cg-mode/)
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/en/multi-llm/model-policy.md:190:with the `moai model profile` command.
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ja/advanced/tokenomics-overview.md:106:`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。 [CG の廃止と設定の移行](/ja/multi-llm/cg-mode/)
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ja/multi-llm/_index.md:165:`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。 [CG の廃止と設定の移行](/ja/multi-llm/cg-mode/)
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ja/multi-llm/model-policy.md:125:v3.1 まで MoAI-ADK は**プロファイルマトリクス**でエージェント 1 個 1 個に `{model, effort}` を割り当てていました。エージェント 13 個 × プロファイル 3 個 = 39 セルの表がアクティブプロファイルの 1 列を選び、リゾルバがその値を spawn 時点に注入し、`moai model profile` コマンドで解決値を覗き見できました。
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ko/advanced/tokenomics-overview.md:112:`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다. [CG 폐기와 설정 이전](/ko/multi-llm/cg-mode/)
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ko/multi-llm/_index.md:164:`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다. [CG 폐기와 설정 이전](/ko/multi-llm/cg-mode/)
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/ko/multi-llm/model-policy.md:182:골랐고, 리졸버가 그 값을 부름 시점에 주입했으며, `moai model profile` 명령으로
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/zh/advanced/tokenomics-overview.md:106:`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。 [CG 停用与配置迁移](/zh/multi-llm/cg-mode/)
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/zh/multi-llm/_index.md:153:`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。 [CG 停用与配置迁移](/zh/multi-llm/cg-mode/)
2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/zh/multi-llm/model-policy.md:124:到 v3.1 为止，MoAI-ADK 用**配置矩阵**给每个智能体分配 `{model, effort}`。13 个智能体 × 3 个配置 = 39 格的表格由活动配置选出一列，解析器在调用时注入该值，还能用 `moai model profile` 命令查看解析结果。
```

Why Class N: the notices say "has been retired" in every locale; the model-policy lines sit in the section headed "What happened to per-agent assignment (history)" (en, with the Korean, Japanese, and Chinese equivalents) in the past tense ("through v3.1 ... could be inspected"), and the following paragraph says the apparatus was retired. The paragraph ranges of the en and ko units (187 to 190, 180 to 183) were read in this pass.

### E-SWB-6 — CLI surface, provenance, and source corroboration

Installed build 7705b365e2 against tree 2bcce5d9dc. Each probe's stderr or stdout is the whole output.

```
moai cg --help        -> exit 1   moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration
moai model --help     -> exit 1   Unknown command "model" for "moai"   (suggestion: codex)
moai migrate --help   -> lists    cg [--flags]   Preview migration of legacy CG teammate roles
moai spec audit --help      -> usage printed (usage: moai spec audit [--flags])
moai profile setup --help   -> usage printed (usage: moai profile setup [name] [--flags])
moai glm setup --help       -> usage printed (usage: moai glm setup [api-key] [--flags])
moai glm --help             -> usage printed; commands setup, status, tools
moai cc --help              -> usage printed
moai init --help            -> lists --model-policy, --profile, --high, --medium-alias, --low, each "Deprecated, no effect"
```

Source at the pin: `git grep -n -E "Use: +\"(cg|gpt|codemaps|model)\"" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- internal cmd` prints two lines: `internal/cli/cg.go:13` (`Use: "cg"`, `Short: "Retired CG launcher"`, `Hidden: true`) and `internal/cli/migrate_cg.go:22` (`Use: "cg"`, `Short: "Preview migration of legacy CG teammate roles"`). No line declares top-level `gpt`, `codemaps`, or `model`. `git diff --name-only 7705b365e2 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- internal/cli cmd` lists 15 files (audit_delta_persist_test.go, codex_audit_launch.go, codex_audit_leg.go, codex_audit_leg_test.go, codex_child_orca_env_test.go, factory_card_pr_gitenv_test.go, factory_mcp_verb_parity_test.go, integration.go, integration_common_root_test.go, integration_session_sidecar_test.go, mcp_build_identity_test.go, mcp_claude.go, mcp_codex.go, mcp_convergence.go, mcp_factory_card.go), none of them `cg.go` or `migrate_cg.go`. Form sets: `git grep -n -o` over each page listed the same distinct forms in all four locales (tokenomics: `spec audit`, `cg`, `cc`; index: `profile setup`, `cc`, `glm`, `cg`, `glm setup`; model-policy: `profile setup`, `init`, `glm`, `model profile`), with equal counts per locale.

### E-SWB-7 — line counts of the twelve files

Command: `git grep -c "" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- <P-ALL>`. Exit code: 0. Output (path suffix, count): en tokenomics 124, en index 204, en model-policy 259, ja tokenomics 124, ja index 197, ja model-policy 163, ko tokenomics 131, ko index 196, ko model-policy 246, zh tokenomics 124, zh index 183, zh model-policy 162. These are the Base lines column of plan.md §A.2.

### E-SWB-8 — overlap and sequencing facts

Command A: `git grep -n -F "설정 이전 안내와 함께 종료합니다" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- docs-site/content/ko/advanced/tokenomics-overview.md docs-site/content/ko/multi-llm/_index.md`. Exit code: 0. Output: the Korean notice lines `docs-site/content/ko/advanced/tokenomics-overview.md:112` and `docs-site/content/ko/multi-llm/_index.md:164`, the same two lines as E-SWB-5. The same command over `docs-site/content/ko` with `-c` prints nine files (this child's two, the other six pages, and the canonical `cg-mode.md`), matching the umbrella's E-08. The README child's REQ-PRC-004 therefore targets the two Korean lines that are this child's units.

Command B: `grep -rl -E "tokenomics-overview|multi-llm/_index|multi-llm/model-policy" /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1721/.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001; echo "exit=$?"`. Output: `exit=1` (no file names any of the twelve). Same command over the GATE spec (worktree t1717) and the REDIR spec (worktree t1718): `exit=1`. SWEEPA's twelve files are `advanced/{bas-navigator,factory-mode,profile-matrix}.md` in four locales (its spec.md Scope), disjoint from these twelve by path. The README, SWEEPC to SWEEPF, and FLAG specs do not exist on disk (`ls -d` over `.moai/worktrees/*/.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-*` listed only the umbrella, GATE, REDIR, and SWEEPA directories).

GATE exemption: `grep -n -i -E "exempt|retire|notice|marker" <t1717 GATE spec.md>` shows requirement C-7: exemptions are the four `multi-llm/cg-mode.md` pages and the four `moai-todo.md` pages by name, and narrative examples only through the Q24 marker defined by SWEEPE. None of the twelve files is named.

### E-SWB-9 — churn

Command: `git log --oneline -n 8 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- <P-ALL>`. Exit code: 0. Output: 6581cdccbb, a9859a2199, c5620da649, be5d56b8c5, 1f287c5791, 4716869503, 211defda4f, d97290ede2.

### E-SWB-10 — controls

- **E-SWB-10a (AC-PRC-001-001).** `git grep -n -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681 -- docs-site/content/en/advanced/bas-navigator.md | rg -v -F -e "llm.team_mode: cg" -e "moai model profile"` prints `2bcce5d9dc4c5f57bb98bcfa553e0061fa373681:docs-site/content/en/advanced/bas-navigator.md:75:moai codemaps`. The same filter with W01 over `<P-ALL>` prints nothing (`exit=1` from `rg`).
- **E-SWB-10b (AC-PRC-001-002 b).** `git diff --quiet 6581cdccbb~1 6581cdccbb -- <P-ALL>; echo "exit=$?"` prints `exit=1`.
- **E-SWB-10c (AC-PRC-001-003).** `git diff --name-only 6581cdccbb~1 6581cdccbb` prints `docs-site/content/en/multi-llm/model-policy.md` and `docs-site/content/ko/multi-llm/model-policy.md` (two paths, neither none nor four).
- **E-SWB-10d (AC-PRC-001-004).** `git diff --name-only c5620da649~1 c5620da649 -- . ':!<P-TOK>' ':!<P-IDX>' ':!<P-POL>' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001' ':!.moai/reports/t1722'` prints 13 paths (`docs-site/content/en/advanced/token-budget.md`, `.../en/claude-code/foundations/commands.md`, `.../en/cost-optimization/prompt-caching.md`, and ten more in ja, ko, zh).
- **E-SWB-10e (AC-PRC-001-005).** `git diff --numstat 4716869503~1 4716869503 -- 'docs-site/content/*/multi-llm/_index.md'` prints `162 54` (en), `156 41` (ja), `141 39` (zh), so a page edit shows as non-empty output.

## D. Gaps and residual risk

**Gaps (not observed in this pass):**

- **G1: the README child has no SPEC on disk.** The overlap with its REQ-PRC-004 rests on the umbrella's E-08 cell and on my reading of the same two Korean lines (E-SWB-8); its actual edit set is unread.
- **G2: no build of the tree.** The CLI probes ran on an installed build 48 commits behind the tree (E-SWB-0). The source read at the pin corroborates the registrations but is not a run of the tree's binary.
- **G3: forms outside the four families.** `--help` resolution was probed for every cited verb (plan.md §A.3); sub-flags other than the five deprecated `moai init` flags were not probed, and no claim is made about them.
- **G4: other children.** SWEEPC to SWEEPF and FLAG have no SPEC on disk; whether one of them names these twelve files is unobserved.
- **G5: no trial merge** with the README child or with any other branch.
- **G6: parity gate and i18n rules document.** `bash scripts/docs-i18n-check.sh` was not run, and `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 were taken from the umbrella's E-24, not re-read.
- **G7: decision board.** `moai decision read` was not queried for Q13, Q17, or Q-SWB-1; no claim is made that a board record does or does not answer them.
- **G8: same-disposition across the model-policy locales** has no mechanical check; it is a reviewer read.
- **G9: leased cards.** No queue or factory query was made. The overlap statement rests on the SPEC directories present in the worktrees (E-SWB-8, command B), which cannot show a card that has no SPEC yet.

**Mutant probe.** A lazy implementation (1) blanks or truncates a page: AC-PRC-001-005 numstat and line counts fail; (2) edits en only or ko only: AC-PRC-001-002 (diff) and AC-PRC-001-003 (per-page counts) fail; (3) removes a retained notice: AC-PRC-001-002 and -003 print fewer than 12 lines; (4) respells `moai  cg` with two spaces: W01 in AC-PRC-001-001 catches it; (5) adds a new run instruction for an absent verb: AC-PRC-001-001's filter prints it; (6) appends a content key to a new Class C line to dodge the filter: the byte-unchanged checks of AC-PRC-001-002 and -005 catch it. Not closed: a respelling outside both patterns (a quoted verb name).

**Residual risk:** a green result proves the twelve files are byte-identical to the pin and no new unresolved-verb spelling was added. It does not prove the retained notices are the right call; that is the ruling on board `d-20261018T201957Z-c26e`.

## E. Definition of Done

- The five criteria print their expected output at the tip and the deciding lines are carried into `.moai/reports/t1722/verdict.md` with the command and output.
- The card closes as rationale recorded plus zero page changes; the twelve files are byte-identical to the pin.
- The state of Q13, Q17, and Q-SWB-1 is as decision-index.md (Gate) states; this file adds no second reading.
- The umbrella's closing measurement is not part of this Definition of Done.
