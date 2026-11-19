---
id: SPEC-20261119-switch-script-v2
title: "Name-switch script v2 preview: spec, statusline and session go to Rust (C-33 S2a)"
status: draft
revision: 1
created: 2026-11-19
---

# Name-switch script v2 preview: spec, statusline and session go to Rust (C-33 S2a)

## Intent

카드 C-33 (p2, 사용자 결정 11-15·11-19)의 조각 S2a. 지금 `~/.moai-cli/path/moai`·`~/go/bin/moai`(같은 1747 B 전환 스크립트, `cmp` 로 같음)는 낱말 63개를 Go(`moai.old`)로 보내요. S1 조사(C-33 note 2777)가 찾은 새는 곳: 스킬·도구·레인이 `moai spec …` 를 치면 Rust 에 `spec` 이 있는데도 Go `spec` 이 받고(F1), Orca 프로필 3곳의 상태줄 스크립트가 `moai statusline` 을 Go 로 보내요. 이 SPEC 은 **v2 스크립트의 미리보기**만 만들어요 — 실제 HOME 에는 아무것도 쓰지 않고, 파일은 이 SPEC 폴더 안 `preview/` 에 있어요. 적용은 S4(guarded)에서 리더가 미리보기·되돌리기를 보여 드리고 사용자가 고른 뒤에만 해요.

v2 가 v1 과 다른 점은 세 가지뿐이에요:
1. `statusline` → Rust (Rust 에 있음).
2. `spec` → **동작 단위로 나눠요**: Rust `spec` 의 동작(`new check list show take close help`)은 Rust, Go 에만 있는 동작(`status drift view lint audit archive ceiling`)은 Go. 이유: 제 컴퓨터의 Go 세대 프로젝트 16곳이 `moai spec lint`(85곳)·`moai spec audit`(5곳)을 지금도 부르는데 Rust `spec` 에는 `lint`·`audit` 가 없어서(`unrecognized subcommand`, 종료 2) 통째로 Rust 로 보내면 그 프로젝트들이 깨져요.
3. `session` → **동작 단위로 나눠요**: Go 에만 있는 동작(`register heartbeat deregister list purge current doctor`)은 Go, 나머지(Rust `session` 은 옛 Claude 시작 명령 `moai session --name <NAME>`)는 Rust. 이유(감사 M1): Go 세대 템플릿 규칙 문서가 에이전트에게 `moai session list --json` 과 `moai session current` 를 치라고 해요(17개 프로젝트의 `.claude/rules`). Rust `session` 에는 그 동작이 없어요(종료 2). 처음 판은 `session` 을 통째로 Rust 로 보냈는데, 호출자 수를 `.claude/rules` 를 빼고 세어서 0 이라고 잘못 적었어요.

## Non-goals

- 실제 HOME 의 `~/.moai-cli/path/moai`·`~/go/bin/moai` 교체, Go 진입점 비활성화, 설치·update 가 놓는 새 링크 — S3·S4(guarded). 이 SPEC 은 읽기 전용 미리보기예요.
- `hook`(다른 프로젝트 5곳의 `.codex/hooks.json` 이 Go 로 부름)과 호출자가 없는 나머지 낱말의 경로 — v1 그대로 Go 예요. 호출자 이전은 그 프로젝트들의 일이에요.
- 저장소 안 `moai-cli` → `moai` 글자 바꾸기(조각 S2b) — `moai spec` 같은 낱말은 이 미리보기가 적용되기 전에는 바꾸지 않아요.
- `.moai-cli/` 폴더 이름, crate 이름.

## Requirements

- REQ-1: When the v2 script is run with `statusline` as its first word, or with `session` and a second word that is not one of `register`, `heartbeat`, `deregister`, `list`, `purge`, `current` or `doctor`, the script shall hand the call, with every argument, to the Rust build. · failure: none
- REQ-2: When the v2 script is run with `spec` and a second word that is one of `new`, `check`, `list`, `show`, `take`, `close`, `help`, `-h` or `--help`, or with `spec` alone, the script shall hand the call, with every argument, to the Rust build. · failure: none
- REQ-3: When the v2 script is run with `spec` and a second word that is one of `status`, `drift`, `view`, `lint`, `audit`, `archive` or `ceiling`, or with `session` and a second word that is one of `register`, `heartbeat`, `deregister`, `list`, `purge`, `current` or `doctor`, the script shall hand the call, with every argument, to the Go build. · failure: none
- REQ-4: The v2 script shall send to the Go build exactly the words that v1 sent there, less `spec`, `statusline` and `session`, which have their own requirements. · failure: none
- REQ-5: The v2 script shall send the launcher words `cc`, `glm`, `codex`, `omp`, `-l`, `-a`, `--leader` and `--agent`, any word it does not list, and an empty command line to the Rust build, as v1 does. · failure: none
- REQ-6: When the v2 script hands `mcp-server` to the Go build, the script shall set `MOAI_MCP_ONESHOT=1` for that call and for no other word. · failure: none
- REQ-7 (REQ-3): IF the Go build is not found when a word that needs it is run, THEN the script shall print one line that names the word and the path searched and exit with status 1. · failure: none
- REQ-8: When the script hands a call over, the script shall pass the arguments unchanged and give the called build's exit status as its own. · failure: none
- REQ-9: The SPEC shall list each word v2 changes with the callers that the change reaches, and each word it leaves on the Go build with its callers or the statement that none was found, each count tied to the command that produced it. · failure: none
- REQ-10: The SPEC shall hold an apply line and an undo line that, run in a temporary HOME, save v1 beside each installed script, restore a file byte-identical to v1, refuse to overwrite a saved copy or to run without a stamp, and report any failure through their exit status. · failure: none
- REQ-11: While the script is only a preview, the work shall write no file in the real HOME: the preview files live under this SPEC folder, the tests run only in temporary folders with fake builds, and applying the script is left to a separate user-approved step. · failure: none
- REQ-12: When the script hands a word to the Rust build, the script shall run it without looking for the Go build. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): `statusline`, `session` and the Rust actions of `spec` reach the fake Rust build; the Go-only actions of `spec` reach the fake Go build; arguments arrive unchanged. v1 sends the same three words to Go (the baseline that shows F1).
  verify: `cargo test --locked --offline --test switch_script_v2 v2_sends_statusline_and_the_session_launcher_form_to_rust`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_splits_spec_by_action`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_keeps_the_go_session_actions_on_go`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_routes_by_the_first_two_words_only`
  verify: `cargo test --locked --offline --test switch_script_v2 v1_still_sends_the_three_words_to_go`
- AC-2 (REQ-4, REQ-5): the Go word list of v2 equals the list of v1 minus the three words, read from the two preview files; launcher, unknown and empty command lines reach Rust.
  verify: `cargo test --locked --offline --test switch_script_v2 v2_changes_exactly_three_words_of_v1`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_keeps_launcher_and_unknown_words_on_rust`
- AC-3 (REQ-6, REQ-7, REQ-8): the one-shot export belongs to `mcp-server` only; a missing Go build fails only the words that need it; arguments and exit status pass through.
  verify: `cargo test --locked --offline --test switch_script_v2 v2_exports_the_mcp_oneshot_flag_for_mcp_server_only`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_without_a_go_build_fails_only_for_go_words`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_hands_over_arguments_and_exit_status`
- AC-4 (REQ-9, REQ-10): the SPEC names every changed word, the Go-only `spec` and `session` actions and the apply and undo lines; the undo line restores v1 byte for byte in a temporary HOME, a second apply with the same stamp refuses and keeps the saved v1, and a failed apply or undo (empty stamp, missing script, nothing saved) fails through the exit status.
  verify: `cargo test --locked --offline --test switch_script_v2 the_spec_names_every_word_it_changes_and_the_undo_line`
  verify: `cargo test --locked --offline --test switch_script_v2 the_undo_line_restores_v1_byte_for_byte`
  verify: `cargo test --locked --offline --test switch_script_v2 a_second_apply_with_the_same_stamp_refuses_and_keeps_the_saved_v1`
  verify: `cargo test --locked --offline --test switch_script_v2 apply_and_undo_report_failure_through_the_exit_status`
- AC-5 (REQ-11): the diff of this branch against its base touches only this SPEC folder and `tests/switch_script_v2.rs`.
  verify: `git diff --name-only origin/v2026.11.0...HEAD`
- AC-6: neighbours (targeted; the full run is cited from the PR's CI run).
  verify: `cargo test --locked --offline --test switch_script_v2`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261119-switch-script-v2/**`, `tests/switch_script_v2.rs`
- read-only: the installed switch script (copied once as `preview/moai.v1`, `cmp` equal), `src/shim_check.rs`, `tools/claude-statusline.sh`, `docs/2383-statusline-switch-design.md`
- forbidden: any write in the real HOME (the Go build's folder, the MoAI state folders, Orca profile folders) or in another project; starting the Go build outside a temporary HOME; deleting anything

## Plan

1. SPEC (this file) and `preview/moai.v1` (copy of the real script), then RED: `tests/switch_script_v2.rs` against `preview/moai.v2` still equal to v1.
2. GREEN: edit `preview/moai.v2`, add `preview/v1-to-v2.diff` (made by `diff -u`), then fill the Result: affected-caller table, the temporary-HOME probe with the real `moai.old` and the installed Rust build, the apply and undo lines.
3. fmt, clippy, mutants on the script (≤5), trailer-check, self-audit (tier 1), `LANE-READY-INTEGRATION`.

## Risks

- `spec close` exists in both builds. v2 sends it to Rust (this repository's calls want Rust). A Go-generation project that ran `moai spec close <SPEC-AUTH-001>` would get Rust, which only reads Go-generation SPECs and refuses; no command, skill or agent file of the 16 projects measured calls `spec close` (only `lint` and `audit`), but hooks and scripts outside those folders were not scanned.
- A status-line script that was written for the Go render (Orca profile folders, `moai_status_line.sh`) gets the Rust three-line render after the switch. That is the intended change; the output differs.
- A listed Rust action that Go also has with different flags is routed by the first two words only; the script does not read flags. A flag before the action (`moai spec --project . lint`), a different case (`spec Lint`), an empty action and `spec help lint` read as "not a Go-only action" and go to Rust; the test `v2_routes_by_the_first_two_words_only` pins this.
- The Go template's status-line wrapper `.moai/status_line.sh` runs `exec moai statusline` and the project's `settings.json` calls it every 60 seconds (audit M2). Every Go-initialized project that keeps that wrapper therefore changes from the Go render to the Rust three-line render once v2 is applied. That is the leader's intended change (note 2781) but it reaches far more than the three Orca profile scripts — see the counts in the Result.
- The apply line copies `preview/moai.v2` verbatim. Its header names the SPEC and its undo hint tells the reader to put back `moai.before-<stamp>`; the stamp is not recorded anywhere else, so the S4 step must print it and keep it with the undo line.
- If both installed scripts are symlinks to one file, apply replaces the links with regular files and undo restores regular-file copies (bytes equal). The two real scripts are regular files.

## Result

미리보기만 만들었어요. 실제 HOME 의 어떤 파일도 쓰지 않았고(읽기: 설치된 스크립트를 `preview/moai.v1` 로 한 번 복사, `cmp` 같음), 프로브는 임시 HOME 에서만 돌렸어요. 파일: `preview/moai.v1`(1747 B, 설치본 그대로), `preview/moai.v2`, `preview/v1-to-v2.diff`(`diff -u moai.v1 moai.v2`, 37줄, `sh -n moai.v2` 통과).

### 무엇이 바뀌나 (v1 → v2)

| 낱말 | v1 | v2 | 영향 받는 곳 (S1 note 2777, 2026-11-19 측정) |
|---|---|---|---|
| `statusline` | Go | Rust | Orca 프로필 3곳(05fa6229, b784936c, e1232167)의 `moai_status_line.sh`·`statusline.sh` 가 `moai statusline` 을 불러요 → Go 렌더 대신 Rust 3줄 렌더가 나와요(의도한 변화). 감사 M2: Go 템플릿의 `.moai/status_line.sh` 도 `exec moai statusline` 이고 `settings.json` 이 60초마다 불러요 — 제 컴퓨터에 그 파일이 있는 프로젝트 30곳, 그 파일을 statusLine 으로 가리키는 `.claude/settings.json` 28곳(명령은 아래). 이 저장소·mo.ai.kr 의 statusLine 은 `path/moai-cli statusline` 직통이라 영향 없어요 |
| `session` (Rust 형태: `session --name …` 등 Go 동작이 아닌 호출) | Go | Rust | Go `session` 은 다중 세션 등록부(`.moai/state/active-sessions.json`) 도구이고 Rust `session` 은 옛 Claude 시작 명령(`moai session --name <NAME>`)이라 뜻이 달라요. 명령어·스킬·에이전트 파일에서 이 형태를 부르는 곳은 못 찾았어요 |
| `session` (Go 에만 있는 동작) | Go | Go | `register` `heartbeat` `deregister` `list` `purge` `current` `doctor`. 감사 M1: Go 템플릿 규칙 문서(`.claude/rules/moai/core/agent-common-protocol.md`, `…-reference.md`, `workflow/session-handoff.md`)가 에이전트에게 `moai session list --json`·`moai session current` 를 치라고 해요 — 제 컴퓨터 17개 프로젝트의 `.claude/rules` 에 있어요(아래 명령으로 셈). Rust `session list`·`session current` 는 종료 2 라서 Go 로 남겨요 |
| `spec` (Rust 동작) | Go | Rust | `new` `check` `list` `show` `take` `close` `help` `-h` `--help` 와 `moai spec` 단독. 스킬 10줄·도구 2줄·레인 지시문의 `moai spec …` 가 이제 Rust `spec` 을 불러요(F1 해소) |
| `spec` (Go 에만 있는 동작) | Go | Go | `status` `drift` `view` `lint` `audit` `archive` `ceiling`. 제 컴퓨터 Go 세대 프로젝트 16곳이 `moai spec lint`(85곳)·`moai spec audit`(5곳)을 지금도 불러요(`grep -rhoE` 로 `.claude/{commands,skills,agents}`·`.codex` 를 셈) — Rust `spec` 에는 `lint`·`audit` 가 없으니(종료 2) Go 로 남겨요 |
| 나머지 60개 | Go | Go | 그대로. 호출자가 있는 것은 `hook`(다른 프로젝트 5곳의 `.codex/hooks.json`)뿐이고 나머지 59개는 설정·저장소 비-시험 기준 호출자 0 — 표시: **호출자 없음** (`init` 포함) |

### 임시 HOME 프로브 (`MOAI_GO_BIN`=실제 `moai.old`, `MOAI_RUST_BIN`=설치된 4.2647.8, `sh preview/moai.vN <낱말> --help`, 첫 줄)

| 호출 | v1 | v2 |
|---|---|---|
| `statusline --help` | Go "Generate a compact statusline string…" | Rust "Render the Claude Code status line" |
| `session --help` (Go 동작 없음) | Go "Manage the multi-session coordination registry…" | Rust "Start a Claude Code session (legacy form)" |
| `session list --help` | Go "List active sessions (optionally filtered by --filter-spec)" | Go (같음; 실제 Rust `session list` 는 종료 2) |
| `spec check --help` | Go "Manage SPEC documents in .moai/specs/ directory." | Rust "Reads only. Check SPEC files and report problems by code" |
| `spec lint --help` | Go "Accepted arguments (SPEC-SPEC-LINT-ID-ARG-001)…" | Go (같음) |
| `spec audit --help` | Go "Scan .moai/specs/SPEC-*/ directories…" | Go (같음) |
| `hook --help` | Go "Execute Claude Code hook event handlers…" | Go (같음) |
| `card --help`, `--version` | Rust | Rust (같음; `--version` 첫 줄 `moai-cli 4.2647.8 (rust)`) |

종료 코드는 모두 0. 이 프로브는 도움말만 불러요: 실제 `hook` 이벤트·`statusline` 렌더·`spec check` 실행은 돌리지 않았어요.

### 적용과 되돌리기 (S4 만 — 리더가 미리보기와 함께 보여 드리고 사용자가 이 세션에서 고른 뒤에만 해요; 이 SPEC 은 실행하지 않아요)

`STAMP=$(date -u +%Y%m%dT%H%M%SZ)` 와 `PREVIEW=<이 저장소>/.moai/specs/SPEC-20261119-switch-script-v2/preview` 를 먼저 정해요. 적용은 두 곳(`~/.moai-cli/path/moai`, `~/go/bin/moai`)의 옛 스크립트를 `moai.before-$STAMP` 로 남기고 새 것을 같은 폴더에 `moai.new` 로 쓴 뒤 `mv` 로 바꿔요(삭제 없음, `moai.old`·Go 바이너리·다른 `moai.before-*` 는 그대로).

- apply lines (S4 only, after the user picks): `rc=0; [ -n "$STAMP" ] || rc=1; if [ $rc = 0 ]; then for d in "$HOME/.moai-cli/path" "$HOME/go/bin"; do [ ! -e "$d/moai.before-$STAMP" ] && cp -p "$d/moai" "$d/moai.before-$STAMP" && cp "$PREVIEW/moai.v2" "$d/moai.new" && chmod 755 "$d/moai.new" && mv "$d/moai.new" "$d/moai" || { echo "apply stopped in $d" >&2; rc=1; break; }; done; fi; [ $rc = 0 ] && echo "saved the old script as moai.before-$STAMP in both folders"`
- undo line: `rc=0; [ -n "$STAMP" ] || rc=1; if [ $rc = 0 ]; then for d in "$HOME/.moai-cli/path" "$HOME/go/bin"; do mv "$d/moai.before-$STAMP" "$d/moai" || { echo "undo stopped in $d" >&2; rc=1; }; done; fi; [ $rc = 0 ]`

앞의 줄은 STAMP 가 비었거나 저장본(`moai.before-$STAMP`)이 이미 있으면 아무것도 바꾸지 않고 실패해요(감사 M3). 둘 중 한 폴더에서 멈추면 메시지를 내고 실패 종료 코드로 끝내요(감사 M4); 그때 이미 바뀐 폴더는 되돌리기 줄이 되돌려요(되돌리기는 두 폴더를 다 시도하고, 저장본이 없는 폴더는 `undo stopped` 로 알려요). STAMP 는 적용 때 화면에 찍힌 값을 그대로 써야 하고 어디에도 따로 기록되지 않아요(Risks). 모두 `tests/switch_script_v2.rs` 의 `the_undo_line_restores_v1_byte_for_byte`·`a_second_apply_with_the_same_stamp_refuses_and_keeps_the_saved_v1`·`apply_and_undo_report_failure_through_the_exit_status` 가 임시 HOME 에서 실제로 돌려요.

### 호출자 수를 낸 명령 (읽기 전용, `~/MoAI/*` 의 프로젝트 폴더에서, 2026-11-19)

- `moai spec lint` 85곳·`moai spec audit` 5곳: `cd ~/MoAI && grep -rhoE "(^|[^A-Za-z0-9_./-])moai[ ]+spec([ ]+-{0,2}[a-z][a-z0-9-]*)?" */.claude/commands */.claude/skills */.claude/agents */.codex | sed -E 's/^[^m]*//' | sort | uniq -c | sort -rn`; 그 낱말을 가진 프로젝트 16곳은 같은 식에 `-rlE` 를 쓴 결과.
- `.claude/rules` 가 `moai session list|current|…` 를 말하는 프로젝트 17곳: `cd ~/MoAI && grep -rlE 'moai[ ]+session[ ]+(list|current|register|heartbeat|purge|doctor)' */.claude/rules | awk -F/ '{print $1}' | sort -u`.
- `.moai/status_line.sh` 가 있는 프로젝트 30곳: `cd ~/MoAI && ls -d */.moai/status_line.sh | wc -l`; 그 파일을 statusLine 으로 가리키는 `.claude/settings.json` 28곳: `cd ~/MoAI && grep -l 'status_line.sh' */.claude/settings.json | awk -F/ '{print $1}'`.
- Go 에만 있는 동작: `moai.old spec --help`·`moai.old session --help`(임시 HOME) 와 Go 소스의 `Use:` 줄(`internal/cli/spec.go`, `internal/cli/session.go`).

### 검사 (이 레인, 2026-11-19, 전용 빌드 폴더·무거운 작업 칸)

첫 판(head 0ea3db730d):
- RED: 새 시험 파일을 v2 가 v1 과 같은 채로 돌리면 10개 중 7개 실패(3개 통과: v1 기준선·바뀌지 않는 동작). GREEN 10/10. 변이 5개(`lint` 를 Rust 쪽으로, `spec` 기본값을 Go 로, `statusline` 을 Go 목록에 되돌림, one-shot 표시를 모든 낱말에, `spec` 의 Go 없음 종료를 0 으로) 5/5 처음에 잡혔어요. `cargo fmt --check` 처음에 시험 파일 줄바꿈 4곳 지적 → 정리. `cargo test --locked --offline --lib` 1453 통과 0 실패(src 는 이 SPEC 이 안 건드려서 수리 뒤에도 같은 소스예요).
- 첫 감사(c33a-audit, tier 1): 차단 0, MAJOR 5(M1 `session` 호출자, M2 `statusline` 호출자 누락, M3 적용 줄이 같은 STAMP 로 두 번 돌면 백업을 덮어씀, M4 적용·되돌리기 줄이 실패해도 종료 코드 0, M5 시험이 인자 경계를 못 잡음) → 이 판이 수리예요.

수리판(커밋 a9380d9882):
- RED(수리 전 v2 로 새 시험): `session` 을 통째로 Rust 로 보내는 이전 v2 에서 14개 중 3개 실패(`v2_keeps_the_go_session_actions_on_go`, `v2_without_a_go_build_fails_only_for_go_words`, `v2_hands_over_arguments_and_exit_status`). 적용·되돌리기 시험 3개는 새 줄을 읽는 시험이라 이전 v2 로는 RED 를 못 만들어서 변이로 확인했어요(아래).
- GREEN: `cargo test --locked --offline --test switch_script_v2` 14/14 통과.
- 변이 5개(커밋된 파일에서 하나씩): S1 스크립트에서 `session:list|` 삭제, S2 마지막 `exec "$RUST" "$@"` 를 따옴표 없이, S3 `statusline` 을 Go 목록에 되돌림, A1 적용 줄에서 저장본 덮어쓰기 가드 삭제, A2 되돌리기 줄이 항상 성공 — 5/5 처음에 시험이 잡았어요(S2 는 감사 M5 에서 살아남던 변이).
- `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0, `cargo test --locked --offline --test home_isolation_scan` 통과. `cargo test --lib` 는 수리판에서 다시 돌리지 않았어요(수리가 src 를 안 건드렸고 `git diff --name-only 0ea3db730d..HEAD` 에 src 가 없어요).
- 돌리지 않은 것: 실제 HOME 에 적용(S4), 실제 훅 이벤트·`statusline` 렌더 실행, Linux/macOS 이외, 프로젝트 전체 `cargo test`(통합 시험 전량은 묶음 PR 의 CI 가 맡아요). 감사 에이전트는 허용된 명령 모양을 벗어나 설치본을 `cmp` 로 읽고 `moai-cli` 를 실제 HOME 으로 5번(사용법 오류 3, `--help` 2) 불렀어요 — 쓰기는 없을 것으로 보지만 `~/.moai` 항목 수는 전후로 재지 않았어요.

### 리더 결정이 필요한 것 (되돌릴 수 있음, 범주 none)

1. `spec close`는 두 빌드에 다 있어요. v2 는 Rust 로 보내요(이 저장소 호출이 Rust 용). Go 세대 프로젝트의 명령·스킬·에이전트 파일은 `spec close` 를 안 불러요(`lint`·`audit` 만). 기본값: 그대로 Rust.
2. `session` 은 뜻이 다른 두 도구예요(Go: 다중 세션 등록부, Rust: 옛 Claude 시작 명령). 기본값: 동작별로 나눔 — Go 동작 7개는 Go, 나머지는 Rust. 리더 지시("session → Rust")에서 벗어난 점이에요: 17개 프로젝트의 규칙 문서가 `moai session list|current` 를 에이전트에게 치라고 해서 통째로 Rust 로 보내면 깨져요. 지시대로 통째로 Rust 로 보내려면 v2 의 `session:` 줄을 지우면 돼요(그 경우 위 17곳 문서의 명령이 종료 2).
3. `statusline` 은 영향이 넓어요: Go 템플릿 `.moai/status_line.sh` 가 있는 프로젝트 30곳 중 28곳의 `.claude/settings.json` 이 60초마다 `moai statusline` 을 불러서, 적용하면 그 프로젝트들의 상태줄이 Go 렌더에서 Rust 3줄 렌더로 바뀌어요. 리더 지시(note 2781)대로라 기본값은 Rust. 되돌리려면 `statusline` 을 Go 목록에 되돌리면 돼요(S4 이후에도 undo 한 줄로 복구).
4. 적용 전에 확인할 것: Orca 프로필 상태줄 스크립트 3개를 누가 실행하는지 못 찾았어요(S1 한계). 안 쓰는 스크립트라면 그 3곳의 영향은 0 이에요.
