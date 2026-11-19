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
2. `session` → Rust (Rust 에 있음, 호출자는 설명 글뿐).
3. `spec` → **동작 단위로 나눠요**: Rust `spec` 의 동작(`new check list show take close help`)은 Rust, Go 에만 있는 동작(`status drift view lint audit archive ceiling`)은 Go. 이유: 제 컴퓨터의 Go 세대 프로젝트 16곳이 `moai spec lint`(85곳)·`moai spec audit`(5곳)을 지금도 부르는데 Rust `spec` 에는 `lint`·`audit` 가 없어서(`unrecognized subcommand`, 종료 2) 통째로 Rust 로 보내면 그 프로젝트들이 깨져요.

## Non-goals

- 실제 HOME 의 `~/.moai-cli/path/moai`·`~/go/bin/moai` 교체, Go 진입점 비활성화, 설치·update 가 놓는 새 링크 — S3·S4(guarded). 이 SPEC 은 읽기 전용 미리보기예요.
- `hook`(다른 프로젝트 5곳의 `.codex/hooks.json` 이 Go 로 부름)과 호출자가 없는 나머지 낱말의 경로 — v1 그대로 Go 예요. 호출자 이전은 그 프로젝트들의 일이에요.
- 저장소 안 `moai-cli` → `moai` 글자 바꾸기(조각 S2b) — `moai spec` 같은 낱말은 이 미리보기가 적용되기 전에는 바꾸지 않아요.
- `.moai-cli/` 폴더 이름, crate 이름.

## Requirements

- REQ-1: When the v2 script is run with `statusline` or `session` as its first word, the script shall hand the call, with every argument, to the Rust build. · failure: none
- REQ-2: When the v2 script is run with `spec` and a second word that is one of `new`, `check`, `list`, `show`, `take`, `close`, `help`, `-h` or `--help`, or with `spec` alone, the script shall hand the call, with every argument, to the Rust build. · failure: none
- REQ-3: When the v2 script is run with `spec` and a second word that is one of `status`, `drift`, `view`, `lint`, `audit`, `archive` or `ceiling`, the script shall hand the call, with every argument, to the Go build. · failure: none
- REQ-4: The v2 script shall send to the Go build exactly the words that v1 sent there, less `spec`, `statusline` and `session`, which have their own requirements. · failure: none
- REQ-5: The v2 script shall send the launcher words `cc`, `glm`, `codex`, `omp`, `-l`, `-a`, `--leader` and `--agent`, any word it does not list, and an empty command line to the Rust build, as v1 does. · failure: none
- REQ-6: When the v2 script hands `mcp-server` to the Go build, the script shall set `MOAI_MCP_ONESHOT=1` for that call and for no other word. · failure: none
- REQ-7 (REQ-3): IF the Go build is not found when a word that needs it is run, THEN the script shall print one line that names the word and the path searched and exit with status 1. · failure: none
- REQ-8: When the script hands a call over, the script shall pass the arguments unchanged and give the called build's exit status as its own. · failure: none
- REQ-9: The SPEC shall list each word v2 changes with the callers that the change reaches, and each word it leaves on the Go build with its callers or the statement that none was found, each count tied to the command that produced it. · failure: none
- REQ-10: The SPEC shall hold one command line that, run in a temporary HOME, restores a file byte-identical to v1 from the copy the apply step saves. · failure: none
- REQ-11: While the script is only a preview, the work shall write no file in the real HOME: the preview files live under this SPEC folder, the tests run only in temporary folders with fake builds, and applying the script is left to a separate user-approved step. · failure: none
- REQ-12: When the script hands a word to the Rust build, the script shall run it without looking for the Go build. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): `statusline`, `session` and the Rust actions of `spec` reach the fake Rust build; the Go-only actions of `spec` reach the fake Go build; arguments arrive unchanged. v1 sends the same three words to Go (the baseline that shows F1).
  verify: `cargo test --locked --offline --test switch_script_v2 v2_sends_statusline_and_session_to_rust`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_splits_spec_by_action`
  verify: `cargo test --locked --offline --test switch_script_v2 v1_still_sends_the_three_words_to_go`
- AC-2 (REQ-4, REQ-5): the Go word list of v2 equals the list of v1 minus the three words, read from the two preview files; launcher, unknown and empty command lines reach Rust.
  verify: `cargo test --locked --offline --test switch_script_v2 v2_changes_exactly_three_words_of_v1`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_keeps_launcher_and_unknown_words_on_rust`
- AC-3 (REQ-6, REQ-7, REQ-8): the one-shot export belongs to `mcp-server` only; a missing Go build fails only the words that need it; arguments and exit status pass through.
  verify: `cargo test --locked --offline --test switch_script_v2 v2_exports_the_mcp_oneshot_flag_for_mcp_server_only`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_without_a_go_build_fails_only_for_go_words`
  verify: `cargo test --locked --offline --test switch_script_v2 v2_hands_over_arguments_and_exit_status`
- AC-4 (REQ-9, REQ-10): the SPEC names every changed word, the seven Go-only `spec` actions and the undo line; the undo line restores v1 byte for byte in a temporary HOME.
  verify: `cargo test --locked --offline --test switch_script_v2 the_spec_names_every_word_it_changes_and_the_undo_line`
  verify: `cargo test --locked --offline --test switch_script_v2 the_undo_line_restores_v1_byte_for_byte`
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
- A listed Rust action that Go also has with different flags is routed by the first two words only; the script does not read flags.

## Result

미리보기만 만들었어요. 실제 HOME 의 어떤 파일도 쓰지 않았고(읽기: 설치된 스크립트를 `preview/moai.v1` 로 한 번 복사, `cmp` 같음), 프로브는 임시 HOME 에서만 돌렸어요. 파일: `preview/moai.v1`(1747 B, 설치본 그대로), `preview/moai.v2`, `preview/v1-to-v2.diff`(`diff -u moai.v1 moai.v2`, 37줄, `sh -n moai.v2` 통과).

### 무엇이 바뀌나 (v1 → v2)

| 낱말 | v1 | v2 | 영향 받는 곳 (S1 note 2777, 2026-11-19 측정) |
|---|---|---|---|
| `statusline` | Go | Rust | Orca 프로필 3곳(05fa6229, b784936c, e1232167)의 `moai_status_line.sh`·`statusline.sh` 가 `moai statusline` 을 불러요 → Go 렌더 대신 Rust 3줄 렌더가 나와요(의도한 변화). 이 저장소·mo.ai.kr 의 statusLine 은 `path/moai-cli statusline` 직통이라 영향 없어요 |
| `session` | Go | Rust | 명령어·스킬·에이전트 파일에서 `moai session` 을 CLI 로 부르는 곳 0(15곳 글자는 `list active moai sessions` 설명문 하나). 다만 Go `session` 은 다중 세션 등록부(`.moai/state/active-sessions.json`) 도구이고 Rust `session` 은 옛 Claude 시작 명령이라 뜻이 달라요 — 호출자가 0 이라 깨지는 곳은 없지만 리더가 확인할 결정이에요 |
| `spec` (Rust 동작) | Go | Rust | `new` `check` `list` `show` `take` `close` `help` `-h` `--help` 와 `moai spec` 단독. 스킬 10줄·도구 2줄·레인 지시문의 `moai spec …` 가 이제 Rust `spec` 을 불러요(F1 해소) |
| `spec` (Go 에만 있는 동작) | Go | Go | `status` `drift` `view` `lint` `audit` `archive` `ceiling`. 제 컴퓨터 Go 세대 프로젝트 16곳이 `moai spec lint`(85곳)·`moai spec audit`(5곳)을 지금도 불러요(`grep -rhoE` 로 `.claude/{commands,skills,agents}`·`.codex` 를 셈) — Rust `spec` 에는 `lint`·`audit` 가 없으니(종료 2) Go 로 남겨요 |
| 나머지 60개 | Go | Go | 그대로. 호출자가 있는 것은 `hook`(다른 프로젝트 5곳의 `.codex/hooks.json`)뿐이고 나머지 59개는 설정·저장소 비-시험 기준 호출자 0 — 표시: **호출자 없음** (`init` 포함) |

### 임시 HOME 프로브 (`MOAI_GO_BIN`=실제 `moai.old`, `MOAI_RUST_BIN`=설치된 4.2647.8, `sh preview/moai.vN <낱말> --help`, 첫 줄)

| 호출 | v1 | v2 |
|---|---|---|
| `statusline --help` | Go "Generate a compact statusline string…" | Rust "Render the Claude Code status line" |
| `session --help` | Go "Manage the multi-session coordination registry…" | Rust "Start a Claude Code session (legacy form)" |
| `spec check --help` | Go "Manage SPEC documents in .moai/specs/ directory." | Rust "Reads only. Check SPEC files and report problems by code" |
| `spec lint --help` | Go "Accepted arguments (SPEC-SPEC-LINT-ID-ARG-001)…" | Go (같음) |
| `spec audit --help` | Go "Scan .moai/specs/SPEC-*/ directories…" | Go (같음) |
| `hook --help` | Go "Execute Claude Code hook event handlers…" | Go (같음) |
| `card --help`, `--version` | Rust | Rust (같음; `--version` 첫 줄 `moai-cli 4.2647.8 (rust)`) |

종료 코드는 모두 0. 이 프로브는 도움말만 불러요: 실제 `hook` 이벤트·`statusline` 렌더·`spec check` 실행은 돌리지 않았어요.

### 적용과 되돌리기 (S4 만 — 리더가 미리보기와 함께 보여 드리고 사용자가 이 세션에서 고른 뒤에만 해요; 이 SPEC 은 실행하지 않아요)

`STAMP=$(date -u +%Y%m%dT%H%M%SZ)` 와 `PREVIEW=<이 저장소>/.moai/specs/SPEC-20261119-switch-script-v2/preview` 를 먼저 정해요. 적용은 두 곳(`~/.moai-cli/path/moai`, `~/go/bin/moai`)의 옛 스크립트를 `moai.before-$STAMP` 로 남기고 새 것을 같은 폴더에 `moai.new` 로 쓴 뒤 `mv` 로 바꿔요(삭제 없음, `moai.old`·Go 바이너리·다른 `moai.before-*` 는 그대로).

- apply lines (S4 only, after the user picks): `for d in "$HOME/.moai-cli/path" "$HOME/go/bin"; do cp -p "$d/moai" "$d/moai.before-$STAMP" && cp "$PREVIEW/moai.v2" "$d/moai.new" && chmod 755 "$d/moai.new" && mv "$d/moai.new" "$d/moai" || break; done`
- undo line: `for d in "$HOME/.moai-cli/path" "$HOME/go/bin"; do mv "$d/moai.before-$STAMP" "$d/moai" || break; done`

두 줄 모두 `tests/switch_script_v2.rs::the_undo_line_restores_v1_byte_for_byte` 가 임시 HOME 에서 실제로 돌려요(적용 뒤 v2 와 같고, 되돌리면 v1 과 바이트까지 같고, 저장본이 사라져요).

### 검사 (이 레인, 2026-11-19, 전용 빌드 폴더·무거운 작업 칸)

- RED: 새 시험 파일을 v2 가 v1 과 같은 채로 돌리면 10개 중 7개 실패(3개 통과: v1 기준선·바뀌지 않는 동작).
- GREEN: `cargo test --locked --offline --test switch_script_v2` 10/10 통과(`rustfmt` 정리 뒤 다시: 10/10).
- 변이 5개(스크립트 `preview/moai.v2` 에 하나씩, 커밋된 파일에서): M1 `lint` 를 Rust 쪽으로, M2 `spec` 기본값을 Go 로, M3 `statusline` 을 Go 목록에 되돌림, M4 one-shot 표시를 모든 낱말에, M5 `spec` 의 Go 없음 종료를 0 으로 — 5/5 처음에 시험이 잡았어요.
- `cargo fmt --check` 0(처음에 시험 파일 줄바꿈 4곳 지적 → 정리), `cargo clippy --locked --offline --all-targets -- -D warnings` 0, `cargo test --locked --offline --test home_isolation_scan` 통과, `cargo test --locked --offline --lib` 1453 통과 0 실패.
- 돌리지 않은 것: 실제 HOME 에 적용(S4), 실제 훅 이벤트·`statusline` 렌더 실행, Linux/macOS 이외, 프로젝트 전체 `cargo test`(통합 시험 전량은 묶음 PR 의 CI 가 맡아요).

### 리더 결정이 필요한 것 (되돌릴 수 있음, 범주 none)

1. `spec close`는 두 빌드에 다 있어요. v2 는 Rust 로 보내요(이 저장소 호출이 Rust 용). Go 세대 프로젝트의 명령·스킬·에이전트 파일은 `spec close` 를 안 불러요(`lint`·`audit` 만). 기본값: 그대로 Rust.
2. `session` 은 호출자가 0 이지만 뜻이 다른 두 도구예요. 기본값: 리더 지시대로 Rust(v2 에 반영). Go 로 두려면 `session` 을 Go 목록에 되돌리면 돼요.
3. 적용 전에 확인할 것: Orca 프로필 상태줄 스크립트 3개를 누가 실행하는지 못 찾았어요(S1 한계). 안 쓰는 스크립트라면 `statusline` 변화의 영향은 0 이에요.
