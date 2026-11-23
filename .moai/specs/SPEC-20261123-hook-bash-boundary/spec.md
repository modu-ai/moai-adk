---
id: SPEC-20261123-hook-bash-boundary
title: "moai hook bash-boundary: PreToolUse Bash boundary check"
status: draft
revision: 1
created: 2026-11-23
---

# moai hook bash-boundary: PreToolUse Bash boundary check

## Intent

새 MoAI 훅 설계의 단계 1-c, H2 예요 (최종 계획서 `~/.moai/reports/moai-adk/hooks-rules-final-plan-2026-11-23.md` §4 H2, §7 1-c, §8). 단계 1-a 가 만든 진입점 `moai hook bash-boundary` 의 처리기를 채워요: Claude Code 가 Bash 를 부르기 전에 명령 문자열을 읽고, 네 가지 경계를 넘으면 `permissionDecision: "deny"` 와 이유를 내고, 아니면 아무것도 출력하지 않아요 (`ask` 는 쓰지 않아요, 사건 #8).

네 가지 검사: ① 세션 체크아웃과 임시 폴더 밖의 삭제 ② 주 체크아웃·남의 워크트리의 git 상태 변경 ③ 통합·보호 브랜치 ref 를 직접 바꾸기 ④ 커밋 훅 건너뛰기. 판정 원칙은 하나예요: **읽을 수 없는 값은 거절해요.** 이 SPEC 은 변수 값을 따라가며 추적하지 않아요 (옛 처리기가 이 추적을 따라가다 감사에서 거듭 FAIL 했어요). 값이 대입·치환·펼침으로 정해져서 명령 문자열만으로 읽을 수 없으면, 그 값이 판정 자리에 있을 때 거절해요.

이 조각은 처리기와 테스트만 해요. 옛 처리기(`src/hook_engine.rs`, `src/hook_guards.rs`, `src/claim_verify.rs`)의 판정 코드는 다시 쓰지 않아요 (사용자 지시 msg_3dc06e21df4c). 그 처리기의 사건 기록(C-327, C-240 카드 노트)은 *어떤 부류가 있었는지* 알려 주는 증거로만 읽었고, 코드와 픽스처는 가져오지 않았어요. 우회 사례의 목록은 이 SPEC 에도 카드에도 적지 않아요: 허용·거절 표와 우회 모양은 `tests/hook_bash_boundary_cli.rs` 의 픽스처 안에만 있어요.

### 이미 있는 것

- `src/hook_entry.rs` 는 `moai hook bash-boundary` 를 이미 받아요 (SPEC-20261123-hook-entry-skeleton): 표준 입력 한 번 읽기, 비MoAI 폴더는 출력 0 · exit 0, 입력 오류는 영수증 한 줄. 이 조각은 그 `bash_boundary` 처리기 자리에 분기 한 줄만 더해요 (C-397 도 같은 파일을 고치니 판정 코드는 새 모듈에 둬요).
- 훅 입력은 PreToolUse Bash 호출에 `tool_name`, `tool_input.command`, `cwd`, `permission_mode`, `session_id`, `scratchpad_dir` 등을 담아요 (공식 문서, Sources).
- 통합 브랜치 이름은 `.moai/project.yaml` 의 `work_flow.integration_branch` 예요 (이 저장소는 `NeXT`).
- 통합 브랜치의 최종 강제는 git `reference-transaction` 가드(`src/card/guard.rs`, 설치기는 C-363)가 해요. 이 훅은 그 앞의 빠른 거절이고, 훅이 없거나 실패해도 불변조건은 깨지지 않아야 해요 (계획서 §1 원칙 1).

## Non-goals

- 훅 설치, settings 항목, `onFailure`, 타임아웃 등록 (단계 1-e). 이 조각은 settings 를 읽지도 쓰지도 않아요.
- H3 쓰기 경계 (Edit/Write/NotebookEdit), H1 맥락, PowerShell (Windows 확인 뒤).
- Bash 의 `mv`, `cp`, `chmod`, 리다이렉션으로 경계 밖 파일을 바꾸는 일. 이 조각의 ①은 삭제만 봐요 (후속 후보).
- 태그를 옮기거나 지우는 일. ③은 브랜치 ref 만 봐요.
- 셸의 완전한 해석. 이 처리기는 명령 문자열을 구문 분석하되 변수·함수·별칭·셸 설정을 실행하지 않고, 읽을 수 없으면 거절해요.
- Windows 에서의 실행 확인. 검증은 macOS 에서만 해요.

## Requirements

<!-- This SPEC adds a check that blocks a tool call: a violation is a deny, never an ask; the final enforcement of the integration branch stays in the Rust reference-transaction guard. -->
- REQ-1: When `moai hook bash-boundary` runs in a MoAI project with a `Bash` call whose `tool_input.command` is a string, the bash boundary handler shall judge that command against the checks of REQ-4 to REQ-7 and the unreadable-value rule of REQ-8 · failure: none — a call whose `tool_name` is present and is not `Bash`, or whose command is missing or is not a string, gets no decision.
- REQ-2: When the command breaks a check, the bash boundary handler shall print exactly one JSON object `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":<text>}}` on stdout and exit 0, and a command that breaks none prints nothing on stdout and stderr and exits 0 · failure: none — the handler never prints `ask` or `allow`, and silence approves nothing.
- REQ-3: The bash boundary handler shall name the session checkout as the checkout that holds the hook input `cwd` (the nearest ancestor with a `.git` entry, read from the file system without starting a process) and judge the later commands of a command string against the working folder that a literal `cd`, `pushd` or `popd` of that string set.
- REQ-4: IF a command deletes (`rm`, `rmdir`, `unlink`, `shred`, `find` in a deleting form, `git clean`, or another command that runs one of them) and a target resolves outside the session checkout and the temporary folders, or is the checkout or a temporary root itself, or is inside a `.git` folder, or holds another checkout, THEN the bash boundary handler shall deny the command with the class `delete-outside`.
- REQ-5: IF a git command changes the state of a checkout other than the session checkout, or, in a session whose checkout is the primary checkout, runs a branch-moving or history-rewriting subcommand (the set in D-5), THEN the bash boundary handler shall deny the command with the class `git-state`.
- REQ-6: IF a git command creates, moves, renames, deletes or force-updates a ref of a protected branch (D-4), or advances the branch that `HEAD` of the target checkout names when that branch is protected, THEN the bash boundary handler shall deny the command with the class `protected-ref`.
- REQ-7: IF a git command skips or replaces the commit hooks (a no-verify option in any spelling git accepts, a hooks-path setting by option, config write or environment, or a prefix of such an option), THEN the bash boundary handler shall deny the command with the class `hook-skip`.
- REQ-8: IF a value that a check needs is unreadable (a path, a branch name, a git subcommand, a working folder or a command word determined by an assignment, expansion, substitution, glob root or alias), or a command exceeds 8 levels, 2048 commands or 256 KiB, or its syntax cannot be split, THEN the bash boundary handler shall deny the command with the class `unreadable`; a value that no check needs, such as the text of a commit message or a file body written by here-document, is data and is not judged.
- REQ-9: The bash boundary handler shall judge a command run through another command or string (wrapper programs, shell `-c` strings, here-documents fed to a shell, command and process substitution, subshells, groups and compound-command bodies) as if it stood alone, to a depth of 8.
- REQ-10: When the reason text is written, the bash boundary handler shall begin it with `moai-boundary/<class>:`, name the violated rule and one allowed way, keep it under 400 bytes, and carry no byte of the command · failure: none — a reason that cannot be built still denies with the class word alone.
- REQ-11: The bash boundary handler shall make no network call, start no process, read no environment variable beyond what the entry reads (`CLAUDE_PROJECT_DIR`, `HOME`), write no file and not depend on `permission_mode` for its decision.
- REQ-12: When the measuring test runs, it shall print p50 and p90 milliseconds of the real binary for an allowed and a denied command with the build profile and the machine load · failure: none — the numbers are reported and a load-dependent bound is not a test failure (D-10).

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8): the table of allowed and denied cases is fixed as tests, at least 12 allowed and 12 denied cases for each of the four checks and for the unreadable class, each running the real binary in a temporary project and HOME and comparing the class and the decision · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac1_/)'`
- AC-2 (REQ-2): outside a MoAI project the hook prints nothing on either stream and exits 0 for every case of the table, and in a MoAI project a denied case prints one JSON object with exactly the three fields of REQ-2, never `ask` or `allow`, and an allowed case prints nothing · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac2_/)'`
- AC-3 (REQ-1): input with another `tool_name`, a missing, numeric, null or array `command`, and an empty string command each exit 0 with nothing printed · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac3_/)'`
- AC-4 (REQ-3, REQ-4): deletion targets inside the checkout, inside a temporary root (not the root) and inside `scratchpad_dir` are allowed, while the checkout itself, `.git`, a sibling checkout, a folder holding a registered worktree, a link that leads out, a path reached through `..` and `~` are denied, and a literal `cd` before a relative target changes the judged folder · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac4_/)'`
- AC-5 (REQ-5): read-only git subcommands are allowed in any checkout, state-changing subcommands are denied toward another checkout (by `-C`, `cd`, `--git-dir`, `--work-tree`), allowed inside a linked worktree session, and the D-5 set is denied in a primary-checkout session · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac5_/)'`
- AC-6 (REQ-6): ref changes of `NeXT`, `main` and `master` by every route of the table are denied, the same on an unprotected branch are allowed, and a commit on a checkout whose `HEAD` names the integration branch is denied while the same commit on another branch is allowed · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac6_/)'`
- AC-7 (REQ-7): every hook-skipping spelling of the table is denied for each hook-running subcommand, and a commit message or an argument that only contains the same letters is allowed · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac7_/)'`
- AC-8 (REQ-8, REQ-9): an unreadable value at a decision place is denied, the same unreadable value in a data place (message text, quoted here-document body) is allowed, nesting through each wrapper kind of the table is judged like the bare command, and 9 levels, 2049 commands and 256 KiB plus one byte are denied while the limits themselves are judged by content · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac8_/)'`
- AC-9 (REQ-10): every denial reason starts with `moai-boundary/` and one of the five class words, is under 400 bytes and contains no byte of the command (a marker word in the command never appears in the output) · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac9_/)'`
- AC-10 (REQ-11): `src/hook_bash_boundary.rs` names no process-start, network, environment-read or file-write word, no file is created or changed in the project or HOME by any case, the answer is the same for every `permission_mode` value, and the diff touches no settings file, `src/hook_guards.rs`, `src/hook_engine.rs` or `src/claim_verify.rs` · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac10_/)'`
- AC-11 (REQ-12): the measuring test runs the real binary 200 times for an allowed and a denied command in a temporary project, prints p50 and p90 in milliseconds with the build profile and the machine load, fails only when p90 is over 10 times the targets, and a command with a trigger word is never answered by the pre-scan alone (D-11) · verify: `cargo nextest run --test hook_bash_boundary_cli -E 'test(/^ac11_/)'`
- AC-12 (REQ-2, REQ-11): the entry skeleton tests still pass, so the other registered names and the non-MoAI answer are unchanged · verify: `cargo nextest run --test hook_entry_cli --test hook_latency_cli --test hook_dispatch_cli --test hook_legacy_cli`
- AC-13 (REQ-1…REQ-12): the SPEC itself passes the spec check · verify: `moai spec check SPEC-20261123-hook-bash-boundary`

## Scope

- write: `.moai/specs/SPEC-20261123-hook-bash-boundary/**`, `src/hook_bash_boundary.rs` (new), `src/hook_entry.rs` (the `bash_boundary` branch, one line), `src/lib.rs` (one `pub mod` line), `tests/hook_bash_boundary_cli.rs` (new), `tests/areas.toml` (only when a new test-file word needs a line; `hook` is already listed).
- read-only: `src/hook_entry.rs` (rest), `src/hook_latency.rs`, `src/card/guard.rs`, `src/ledger/location.rs`, `.moai/project.yaml`, the card notes of C-327 and C-240 as incident evidence.
- forbidden: any settings file (`.claude/**` of the home and of the project), the installed hooks of this clone, `src/hook_guards.rs`, `src/hook_engine.rs`, `src/claim_verify.rs`, `src/main.rs`, `src/help.rs`, the real HOME and its `.moai/projects` folder, and any network call. Tests use temporary HOMEs, folders and repositories only, and a temporary HOME is set on the child process only.

## Plan

1. (this slice) SPEC, status draft; the user's approval is not written by a lane. Until it is approved nothing below starts.
2. Tests first (RED): `tests/hook_bash_boundary_cli.rs` with the case table and the fixtures (temporary repository, linked worktrees made by `git worktree add`, link-out and sibling-checkout shapes) against a stub that answers no decision.
3. `src/hook_bash_boundary.rs`: bounded word splitter and command walker with a depth, count and size budget, the fast pre-scan (REQ-13), the checkout resolver (REQ-4), the four checks, the reason writer; `src/lib.rs` and `src/hook_entry.rs` get one line each.
4. Guard mutants: remove each guard in turn and show the test that fails (`guard-mutant:` lines in READY).
5. Dangerous-path run in a throwaway: the real binary against a throwaway project, a plain folder at a registered path, a link out and a deep nesting.
6. Measurement run: p50 and p90 of the real binary for an allowed and a denied command, release and debug label, load noted (D-10).
7. Targeted tests with nextest `--test` names, `heavy-slot --max 4`, one bare `check affected` at READY.

## Risks

- R-1: A string analysis cannot be complete. Unreadable input is denied, and the integration branch stays guarded by the Rust reference-transaction guard (C-363) and the landing command; this hook is the fast first refusal, not the only one.
- R-2: Over-blocking. A legitimate command with an unreadable word at a decision place is denied (D-2, D-8). The reason names an allowed way (a literal path, a literal command). Relaxing a rule is a decision of this SPEC, never a silent special case.
- R-3: `cwd` follows `cd` between calls (official documentation). A session can move into a sibling checkout in one call and act in the next, where that checkout is then the session checkout. The hook sees only the current `cwd`; the card reservation and the landing check remain the second line (D-3).
- R-4: `onFailure: "block"` (installed by 1-e) blocks every Bash call of a MoAI project when the binary is missing or broken. The first rollback is removing `onFailure` (plan §8); this slice installs nothing.
- R-5: Whether a PreToolUse deny holds in `bypassPermissions`, `dontAsk` and `auto` modes is not verified from the documentation read (Sources). The check is part of the real-load check of 1-e; D-9 makes the handler mode-independent so the handler itself needs no change either way.
- R-6: Latency depends on machine load (80 to 90 observed on 2026-11-23). The numbers are reported with the load; a hard bound would be flaky (D-10).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The judgment lives in a library module `src/hook_bash_boundary.rs` (`pub mod` in `src/lib.rs`), and the entry's `bash_boundary` calls it in one line. | The card scope names `src/lib.rs` and not `src/main.rs`; `src/hook_entry.rs` is a binary-side module, so a library module is testable by name and keeps C-397's merge small. | Move the module into the binary side with a `main.rs` card. |
| D-2 | No value tracking: a word with an expansion, substitution or glob root at a decision place is unreadable and denied; no variable name is allowed by a list. | Plan §4 H2 principle; the older handler's variable tracking is the part audits rejected again and again. Default; options: allow a short list of environment names such as the temporary folder variable. | Add a name list in a later SPEC revision. |
| D-3 | The session checkout is the checkout holding the hook `cwd`; when `CLAUDE_PROJECT_DIR` names a repository, `cwd` must be in the same repository (same git common folder) to count, else the project folder is the session checkout. Temporary roots: `/tmp`, `/private/tmp`, `/var/tmp`, the macOS per-user temporary folder pattern `/private/var/folders/*/*/T`, and the input's `scratchpad_dir`; a root itself is never a target. | Official documentation: `cwd` follows `cd`, `CLAUDE_PROJECT_DIR` stays at the session start; lane worktrees entered after the start lie under the project folder. Gap: R-3. | Change the root list in the constants. |
| D-4 | Protected branches: `work_flow.integration_branch` of `.moai/project.yaml`, plus `main` and `master`. A name match is case-insensitive. | The ruleset of this repository and the card text ("통합·보호 브랜치"); the project file has no protected-branch list today. Default; option: add a list key to the project profile with its own SPEC. | Edit the constant list or add the profile key later. |
| D-5 | In a session whose checkout is the primary checkout, the branch-moving and history-rewriting set is `checkout`, `switch`, `reset`, `rebase`, `stash`, `clean`, `merge`, `pull`, `cherry-pick`, `revert`, `am`; `add`, `commit`, `fetch`, `restore` and `rm` stay allowed there. Toward another checkout every state-changing subcommand is denied. | `CLAUDE.md` ("no branch switch, reset or stash in the shared checkout", leader commits after re-reading HEAD); incident #7. The read-only and state-changing lists are fixed in the case table. | Edit the sets in the module and the table. |
| D-6 | An unknown git subcommand (an alias or an external subcommand) is denied as `unreadable`. | An alias can hide any option; the known porcelain list is in the case table. Default; option: allow unknown. | Add names to the known list. |
| D-7 | The reason text is English with a class prefix and no command bytes. | It is read by the agent, and it must not echo a command that may hold a secret (principle: secrets never enter messages). | Translate the text part; the prefix stays. |
| D-8 | A command word that cannot be read (starts with an expansion or substitution) is denied as `unreadable`. | The word may name any program; plan principle. Default; option: allow it when the last path segment is a literal script that is not a program of the checks. | Add the option in a later revision. |
| D-9 | `permission_mode` is read and never changes the decision: the deny is the same in every mode. | Plan §4 H2 lists it as an input; the hook is the boundary in every mode. R-5. | Make a mode skip a check with its own SPEC. |
| D-10 | The measuring test prints p50 and p90 and fails only above 10 times the targets (200 ms and 400 ms); the READY note reports the measured numbers with profile and load. | Card criterion 4 asks for measured numbers; a hard 20 ms bound fails under load 80 and would train people to ignore the test. | Tighten the bound when a quiet machine is the norm. |
| D-11 | A command with none of the words the checks react to (no `rm`-family, `find`, `git`, `xargs`, shell or wrapper name, `eval`, `source`, `exec`, substitution marker) as a substring answers no decision without splitting. | Latency and the same answer the full walk would give; the pre-scan is a superset of the walk's trigger words, pinned by a test. | Remove the pre-scan; answers do not change. |

### Sources

Official hooks documentation, https://code.claude.com/docs/en/hooks, fetched by a research subagent with WebFetch on 2026-11-23 (clock read in this session `2026-11-23T12:54:04Z`, local Claude Code 2.1.330). The page has 293,973 characters; characters 0 to 79,900 came back verbatim, 79,900 to 179,900 only as a model summary, and 179,900 to 293,973 were not read. The quotes below are from the verbatim part unless marked.

- Deny output (section "How a hook resolves"): `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Destructive command blocked by hook"}}`.
- Input of a PreToolUse Bash call: `session_id`, `prompt_id`, `transcript_path`, `cwd`, `scratchpad_dir`, `permission_mode`, `effort`, `hook_event_name`, `tool_name`, `tool_input` (`command`, `description`, `timeout`, `run_in_background`), `tool_use_id`; `permission_mode` values `default`, `plan`, `acceptEdits`, `auto`, `dontAsk`, `bypassPermissions`.
- Exit 0: "Exit code 0 with no output means the hook has no decision to report" and "staying silent doesn't approve it". Exit 2 is a blocking error; a blocking JSON decision's reason replaces stderr.
- `cwd`: "the `cwd` field in the hook's input JSON is the worktree root after Claude enters a worktree, and the new directory after Claude runs `cd`"; `${CLAUDE_PROJECT_DIR}` "stays put".
- `if` filter is "best-effort"; "When Claude Code can't split the Bash input into subcommands, it runs your hook regardless of the pattern"; the page says to use the permission system rather than a hook for a hard allow or deny. This SPEC therefore installs no `if` filter (1-e decides) and treats the hook as the first refusal.
- A PreToolUse hook with the `Bash` matcher also runs on Monitor-tool commands (`allow` and `updatedInput` have no effect there).
- Unverified: whether a PreToolUse deny holds in `bypassPermissions`, `dontAsk` and `auto` modes (not in the verbatim part); the full value set of `permissionDecision` (the summary says `allow`, `deny`, `ask`, `defer`); the part of the page after character 79,900 verbatim. This SPEC uses only `deny`.

## Result

Status: draft, not implemented. The SPEC waits for the user's approval; no test and no code exists for it yet.
