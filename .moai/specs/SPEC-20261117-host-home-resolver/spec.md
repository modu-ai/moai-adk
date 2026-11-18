---
id: SPEC-20261117-host-home-resolver
title: 도구 폴더 찾기 하나 — CLAUDE_CONFIG_DIR와 CODEX_HOME을 따름 (C-79)
status: draft
revision: 1
created: 2026-11-17
---

## Intent

Claude Code는 `CLAUDE_CONFIG_DIR`, Codex는 `CODEX_HOME`으로 사용자 폴더를 옮길 수 있어요. 그런데 MoAI의
라우팅(`home_routing.rs`)·에이전트(`agent_install.rs`)·출력 스타일·스킬 복사는 `~/.claude`·`~/.codex`를 그대로
썼고, `codex-doctor`만 `CODEX_HOME`을 읽었어요. 그래서 폴더를 옮긴 사용자에게는 MoAI가 로드되지 않고, 진단과
설치가 서로 다른 폴더를 봤어요(전체 조사 2026-11-17 C3). 이 SPEC은 도구 폴더를 정하는 곳을 하나(`src/host_home.rs`)로
만들고 모든 곳이 그것을 쓰게 해요.

공식 문서(2026-11-17T09:33Z에 읽음):
- https://code.claude.com/docs/en/claude-directory.md — "If you set `CLAUDE_CONFIG_DIR`, every `~/.claude` path on this
  page lives under that directory instead." (`CLAUDE.md`, `settings.json`, `agents/`, `skills/`, `output-styles/` 포함)
- https://learn.chatgpt.com/docs/agent-configuration/agents-md — 전역 AGENTS 파일 폴더는 "defaults to `~/.codex`,
  unless you set `CODEX_HOME`".
- Codex 사용자 스킬은 `~/.agents/skills`(build-skills 페이지 요약, 기존 `src/hosts/codex.rs`와 같음)이라
  `CODEX_HOME`이 옮기지 않아요.

## Non-goals

- OMP 폴더 옮기기(문서에 변수가 없음), Kiro·Cursor.
- Go 세대가 남긴 것(`update_cleanup`의 Go 뿌리, `skills_legacy`, `baseline`)의 위치 바꾸기: Go 도구는 HOME 기준 고정
  경로에 설치했어요(`moai-adk-go/internal/config/userroots.go` `userRootFixedDirs`).
- 옮겨진 폴더 안의 남은 파일(`projection_extra`) 정리. 정리 항목은 HOME 기준 경로라 HOME 밖은 다루지 않아요(후속).
- `~`로 시작하는 값 풀기: 셸이 풀어 주지 않은 `~/x`는 상대 경로로 보고 거절해요.

## Requirements

- REQ-1: When a step needs a host's user folder for the account's own home, the resolver shall return `CLAUDE_CONFIG_DIR` for Claude Code and `CODEX_HOME` for Codex when that variable holds an absolute path, and the `.claude` or `.codex` folder under HOME when it is unset.
- REQ-2 (REQ-1): IF the variable is set but empty or holds a relative path, THEN the resolver shall refuse with a message that names the variable and not its value.
- REQ-3: Each MoAI step that reads or writes a host folder for the routing section or the home agents or the output style or the Claude skill copy or the Codex doctor or the Claude user settings shall take that folder from the one resolver.
- REQ-4: Where the home is not the account's own home, the resolver shall return the folders under that home.
- REQ-5 (REQ-1): IF the relocated folder or a folder below it on the way to a target is a link that does not resolve to the same folder under the account home's default host folder, THEN the step shall block that host without following the link and name the link in the reason.
- REQ-6: The Codex skill copy shall stay in the `.agents/skills` folder under HOME.
- REQ-7: The user-scope guards of the Claude settings and the Codex config and the hook writer shall refuse writing into a folder that `CLAUDE_CONFIG_DIR` or `CODEX_HOME` names.
- REQ-8: The Go-generation cleanup and legacy scans shall keep their HOME-relative roots.
- REQ-9: The repository SPECs shall pass `moai spec check --all`.
- REQ-10 (REQ-1): IF a relocation variable holds a path that is not UTF-8, THEN the skill-copy preview and the Codex doctor shall leave out or refuse only that host without a panic.

## Acceptance

- AC-1 (REQ-1, REQ-3): 두 변수를 절대 경로로 주면 라우팅 계획·적용·되돌리기, 에이전트, 출력 스타일, Claude 스킬 복사, Codex 진단이 옮겨진 폴더를 쓰고 HOME 아래 `.claude`·`.codex`에는 아무것도 생기지 않아요.
  verify: `cargo test --locked --offline --test host_home_cli -- routing_writes_and_reverts_in_the_relocated_folders agents_and_the_output_style_follow_the_variables the_skill_copy_follows_claude_config_dir the_codex_doctor_reads_codex_home`
- AC-2 (REQ-2): 빈 값·상대 경로는 그 도구만 막고(`blocked`) 사유가 변수 이름을 말하며 값은 되풀이하지 않아요; `codex-doctor`는 종료 1로 거절해요.
  verify: `cargo test --locked --offline --test host_home_cli an_empty_or_relative_variable_blocks_its_host_by_name`;
  `cargo test --locked --offline --lib host_home::`
- AC-3 (REQ-4): 계정 홈이 아닌 홈(임시 HOME, 다른 계정 홈)은 변수가 있어도 그 아래 폴더를 써요.
  verify: `cargo test --locked --offline --test host_home_cli the_variables_count_only_for_the_account_home`
- AC-4 (REQ-5): 링크인 `CLAUDE_CONFIG_DIR`은 라우팅·에이전트·스킬 복사·출력 스타일에서 막히고 링크 너머에 아무것도 쓰지 않아요. 옮겨진 폴더 안의 `skills`·`agents` 링크는 계정 홈의 같은 기본 폴더(`.claude/skills`·`.claude/agents`·`.codex/agents`)로 갈 때만 따라가고, 계획과 적용이 그 폴더를 써요; 다른 곳으로 가는 링크나 없는 기본 폴더로 가는 링크는 막히고 사유가 그 링크 경로를 말해요.
  verify: `cargo test --locked --offline --test host_home_cli -- a_linked_config_folder_is_refused_not_followed a_link_to_the_default_folder_is_followed_and_any_other_link_is_blocked the_output_style_refuses_a_linked_config_folder`
- AC-5 (REQ-6): `CODEX_HOME`이 있어도 Codex 스킬 복사 위치는 HOME 아래 `.agents/skills/moai`예요.
  verify: `cargo test --locked --offline --test host_home_cli the_skill_copy_follows_claude_config_dir`
- AC-6 (REQ-7): 프로젝트 `.claude`가 `CLAUDE_CONFIG_DIR`이거나 `.codex`가 `CODEX_HOME`이면(또는 프로젝트가 그 폴더 안이면) Claude 설정·Codex 설정·훅 쓰기가 두 도구 모두 거절하고 아무것도 쓰지 않으며, 빈 값·상대 경로도 변수 이름으로 거절해요.
  verify: `cargo test --locked --offline --test host_home_cli -- hooks_refuse_the_relocated_user_folders the_user_scope_guards_refuse_empty_and_relative_values`;
  `cargo test --locked --offline --test claude_settings_cli --test codex_config_cli --test hooks_cli`
- AC-7 (REQ-8): Go 정리·옛 스킬 검사 시험이 그대로 통과하고 그 뿌리 표는 바뀌지 않아요.
  verify: `cargo test --locked --offline --test update_cleanup_cli`;
  `git diff origin/v2026.11.0...HEAD -- src/skills_legacy.rs src/baseline.rs` → 출력 없음
- AC-8 (REQ-9): 저장소의 SPEC이 모두 통과해요.
  verify: `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`
- AC-10 (REQ-10): UTF-8이 아닌 `CLAUDE_CONFIG_DIR`은 `update-plan`에서 Claude만 막고(Codex는 계획), `codex-doctor`는 UTF-8이 아닌 `CODEX_HOME`을 종료 1로 거절해요; 실행기 경고는 폴더를 못 찾으면 이유를 말해요.
  verify: `cargo test --locked --offline --test host_home_cli a_variable_that_is_not_utf8_blocks_only_its_host`;
  `cargo test --locked --offline --lib launcher_init::`
- AC-9 (REQ-1…REQ-10): 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --lib`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --test home_routing_cli --test home_style_cli --test agents_home_cli --test codex_adapter --test codex_config --test recovery_cli --test home_isolation_scan --test git_env_scrub --test product_cli`;
  `cargo test --locked --offline --test update_cli --test update_backup_replace --test update_command_cli --test update_end_to_end_cli --test update_home_layer_cli --test update_hooks_cli --test update_keys_cli --test update_known_good --test update_known_good_replace --test update_moved_receipts --test update_route --test update_tux_golden --test update_unmanaged_failclosed --test update_user_layer_cli --test update_version_folders`;
  `cargo fmt --check`; `rustfmt --check --edition 2021 src/ops/*.rs`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests.skill_references_test`

## Scope

- write: `src/host_home.rs`, `src/lib.rs`, `src/home_routing.rs`, `src/agent_install.rs`, `src/home_style.rs`,
  `src/update.rs`, `src/recovery.rs`, `src/update_cleanup.rs`, `src/claude.rs`, `src/hosts/codex_config.rs`,
  `src/ops/codex_doctor.rs`, `src/hooks.rs`, `src/launcher_init.rs`, `tests/host_home_cli.rs`, `tools/repo_scans.json`,
  `docs/design-review/appendix/operations/home_routing_plan.md`, `docs/design-review/appendix/operations/agents_home_plan.md`,
  `docs/design-review/appendix/operations/home_style_plan.md`, `docs/design-review/appendix/operations/codex_doctor.md`,
  `.moai/specs/SPEC-20261117-host-home-resolver/spec.md`
- read-only: `src/skills_legacy.rs`, `src/baseline.rs`, `src/hosts/codex.rs`, `src/home_root.rs`
- forbidden: 실제 HOME과 실제 `CLAUDE_CONFIG_DIR`·`CODEX_HOME` 폴더 쓰기, `install/**`, `.github/**`, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 공식 문서 확인(위 Intent의 출처), 코드에서 도구 폴더를 만드는 곳 조사.
2. 시험 먼저: `tests/host_home_cli.rs` 7개, RED 기록 뒤 따로 커밋.
3. `src/host_home.rs`(변수 읽기·검사·계정 홈 규칙)와 호출하는 곳 바꾸기.
4. GREEN, 이웃 시험, 형식·clippy·Python, 실제 `CLAUDE_CONFIG_DIR` 폴더가 시험 전후 같은지 확인, mutation 5개 이하.
5. 동작 문서 네 개와 이 SPEC의 Result.

## Risks

- 계정 홈 규칙 때문에 `HOME`을 계정 홈과 다르게 둔 사용자는 변수가 적용되지 않아요(오늘과 같은 동작) → 그 경우
  `MOAI_ACCOUNT_HOME`으로 계정 홈을 정할 수 있어요.
- 적용과 되돌리기 사이에 변수를 바꾸면 되돌리기가 다른 폴더를 봐요 → 그 폴더의 바이트가 영수증과 다르면 건너뛰고
  보고해요(기존 `user_changed_since_apply` 규칙).
- `update_cleanup`은 HOME 밖으로 옮겨진 스킬 폴더의 남은 파일을 정리하지 않아요(Non-goals).
- 비밀번호 데이터베이스가 없는 시스템(Windows 등)에서는 계정 홈이 `key_vault::home`(`HOME`, Windows는 `USERPROFILE`)이에요
  (감사 N2). 그래서 거기서는 임시 `HOME`도 계정 홈이 되어, 실제 `CLAUDE_CONFIG_DIR`·`CODEX_HOME`을 물려받은 채
  임시 `HOME`으로 실행한 시험·명령이 그 실제 폴더를 쓸 수 있어요. 그런 시스템의 시험은 이 변수들을 지운 환경
  (`clean_env`)에서 돌려야 해요. 이 동작을 확인하는 시험은 없어요(이 SPEC의 시험은 Unix).
- C-79 전에 만든 영수증(감사 N4): 라우팅·스타일·에이전트 영수증은 도구와 파일 이름만, 스킬 복사 영수증은 스킬 폴더 아래
  경로만 적고 폴더 자체는 적지 않아요. 되돌리기·복구는 실행할 때 폴더를 다시 정하므로, 그 사이 변수를 설정하면 옛
  영수증이 옮겨진 폴더를 봐요. 그 폴더의 바이트가 영수증과 다르면 건너뛰거나(`user_changed_since_apply`·
  `file_not_plain`) 충돌·없음으로 보고하고, HOME 아래 `.claude`·`.codex`의 옛 파일은 지우지도 되돌리지도 않아요. 변수를
  지우고 실행하면 옛 위치로 되돌릴 수 있어요. 영수증 형식은 바꾸지 않았어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 도구 폴더는 `src/host_home.rs` 하나가 정함: Claude `CLAUDE_CONFIG_DIR`, Codex `CODEX_HOME`, 없으면 HOME 아래; OMP는 변수 없음 | 카드 C-79, 공식 문서(Intent) | 모듈 삭제, 호출하는 곳을 HOME 고정으로 |
| D2 | 변수는 계정 홈(비밀번호 데이터베이스, 시험은 `MOAI_ACCOUNT_HOME`)일 때만 셈; 다른 홈은 그 아래 폴더 | 시험 30여 개가 임시 HOME에 실제 환경을 물려받아, `$HOME`만 비교하면 실제 설정 폴더에 쓸 수 있음(이 세션에 `CLAUDE_CONFIG_DIR`이 설정돼 있어 확인) | `$HOME`과 비교(그 전에 시험 환경을 모두 비워야 함) |
| D3 | 빈 값·상대 경로는 거절, 메시지는 변수 이름만(값 없음); `~/x`도 상대로 봄 | 리더 지시, 다른 폴더 기준으로 읽힐 위험 | 상대 경로를 작업 폴더 기준으로 풀기 |
| D4 | 변수 경로는 적힌 그대로(정규화 안 함); 그 폴더 자체가 링크면 `~/.claude`처럼 막음 | 리더 지시(링크 처리 바꾸지 않기) | `canonicalize` 뒤 사용 |
| D5 | 계획에서 폴더를 정할 수 없으면 그 도구만 `blocked`(표시 경로는 HOME 아래), 적용·되돌리기는 오류 | 다른 도구 계속 진행 | 계획 전체 실패 |
| D6 | Codex 스킬은 `~/.agents/skills` 그대로 | Codex 문서·기존 `hosts/codex.rs` | `CODEX_HOME/skills` |
| D7 | 사용자 범위 가드(`claude.rs`, `codex_config.rs`, 그리고 D11의 `hooks.rs`)는 계정 홈과 상관없이 변수가 가리키는 폴더를 막고, 빈 값·상대 경로면 거절 | 가드는 넓게 막는 쪽이 안전 | 예전처럼 상대 값을 작업 폴더 기준으로 |
| D8 | `src/update.rs`는 스킬 폴더를 찾는 줄(`skills_place`·`skills_parent`·`projection_root`)과 그 호출 네 곳, `preflight`의 링크 검사 시작점과 UTF-8이 아닌 경로의 그 도구 제외(D12)만 바꿈 | 리더: 피할 수 없을 때만, 보고 | 그 줄들 되돌리기 |
| D9 | Go 세대 정리(`update_cleanup` 뿌리 표)·`skills_legacy`·`baseline`은 HOME 기준 유지; `projection_extra`는 HOME 밖이면 건너뜀 | Go 도구의 고정 설치 경로 | 각 표를 리졸버로 |
| D10 | 옮겨진 도구 폴더 안의 폴더 링크(`skills`·`agents`·`output-styles` 등)는 계정 홈의 같은 기본 폴더(예: `<계정 홈>/.claude/skills`)와 실제 경로가 같을 때만 따라가고, 그때 계획·적용·링크 검사는 그 기본 폴더를 써요; 다른 링크·없는 기본 폴더로 가는 링크는 막고 사유가 링크 경로와 할 일을 말해요; 파일 링크(`settings.json` 등)는 그대로 거절 | 사용자 선택 2026-11-17 ~23:2xZ "기본 폴더 링크만 허용"(감사 B2) | 모든 링크를 막기(감사 전 판), 또는 모든 폴더 링크 따라가기 |
| D11 | 훅 쓰기(`hooks::project_root`)도 D7처럼 변수가 가리키는 폴더(프로젝트 자체나 그 `.claude`·`.codex`)를 거절 | 감사 B1 | 그 검사 빼기 |
| D12 | UTF-8이 아닌 변수 값: `update-plan`은 그 도구만 `blocked`, `codex-doctor`는 거절 | 감사 N3, D5와 같은 원칙 | 계획 전체 실패 |

## Result

구현 에이전트가 2026-11-18에 채웠어요(감사 FAIL `75ebca6750` 수리 뒤). 감사가 다시 확인해요. 빌드 폴더는 이 worktree의
`target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는 한 번에 하나(heavy slot 안).

커밋: 첫 구현 `bd74b730c2`…`23be93aa4a`(RED 5 failed → GREEN, mutation M1–M5 KILLED) → 감사 FAIL(B1, B2, N1–N4) →
`5a546d28ac`(tip `21c43c2af7` 병합) → `16eca65bab`(수리 시험 먼저; RED: B1·B2·N3 시험 3개 실패, M5–M7 고정 시험은 처음부터 통과) →
`986bc4898a`(수리) → `ceab95a3cb`(tip `1d49d55062` 병합) → 이 Result. 수리 mutation 8개 모두 KILLED(R1 훅 검사 끄기,
R2 아무 폴더 링크나 따라가기, R3 링크를 아예 안 따라가기, M5 출력 스타일의 도구 폴더 링크 검사 빼기, M6·M7 Claude·Codex
가드가 빈 값·상대 경로 받기, M8 실행기 경고 끄기, R4 UTF-8이 아니면 계획 전체 실패), 매번 트리 깨끗.

최종 검사(`ceab95a3cb`, 02:42:37Z–02:56:50Z, 단계마다 종료 코드 0): 시험 대상 42개 한 번에, `--lib` 1425 passed,
`--bin moai` 27, clippy `-D warnings`, `cargo fmt --check`, `rustfmt --check --edition 2021 src/ops/*.rs`, Python 49개 OK,
`source_manifest.py check`, `moai spec check --all --project .`(SPEC 45개, 오류 0; 경고 4개는 다른 SPEC의 `needs_clarification`).

실제 홈 안전: 이 세션에는 `CLAUDE_CONFIG_DIR`이 설정돼 있어요. 그 폴더의 `CLAUDE.md`·`settings.json`·`agents`·`skills/moai`·
`output-styles`와 HOME 아래 `.codex/AGENTS.md`·`.codex/agents`의 수정 시각·크기를 읽기만으로 각 실행 앞뒤에 비교했고 늘
같았어요. 첫 지문(11-17 09:43Z)과 다른 곳은 `skills/moai`·`.codex/agents` 두 개뿐이고 11-17 10:17:55–56Z의 빌드 5 설치
(`~/.moai-cli`의 `update.apply` 4.2647.5 영수증, 권한 "moai-cli update confirmation")가 바꾼 것이에요; 그 폴더의 `skills`·
`agents`는 HOME 아래 `.claude/skills`·`.claude/agents`로 가는 링크예요.

- AC-1: met — `ceab95a3cb`에서 `host_home_cli` 12 passed(라우팅·에이전트·출력 스타일·스킬 복사·doctor가 옮겨진 폴더 사용)
- AC-2: met — 같은 실행(빈 값·상대 경로는 그 도구만 `blocked`), `--lib` 안의 `host_home::` 6개 통과
- AC-3: met — 같은 실행의 `the_variables_count_only_for_the_account_home`; mutation M1 KILLED
- AC-4: met — 같은 실행(링크인 설정 폴더는 막힘, 기본 폴더로 가는 `skills`·`agents` 링크만 따라가고 계획·적용이 같은 폴더, 다른 링크는 링크 경로를 말하며 막힘, 출력 스타일도 막힘); R2·R3·M5 KILLED
- AC-5: met — 같은 실행(`CODEX_HOME`이 있어도 Codex 스킬은 HOME 아래 `.agents/skills/moai`)
- AC-6: met — 같은 실행의 훅·가드 시험, `hooks_cli` 22·`claude_settings_cli` 18·`codex_config_cli` 10 passed; R1·M6·M7 KILLED
- AC-7: met — `update_cleanup_cli` 13 passed; `git diff origin/v2026.11.0...HEAD -- src/skills_legacy.rs src/baseline.rs` 출력 없음
- AC-8: met — `spec_cli` 17 passed; 이 가지에서 빌드한 `moai spec check --all --project .` exit 0, SPEC 45개, 오류 0
- AC-9: met — 위 최종 검사의 모든 단계 종료 코드 0(시험 대상 42개, `--lib` 1425, `--bin moai` 27, clippy, fmt, rustfmt, Python 49, manifest)
- AC-10: met — 같은 실행의 `a_variable_that_is_not_utf8_blocks_only_its_host`와 `--lib` 안의 `launcher_init::` 6개 통과; R4·M8 KILLED

실행 안 함: 전체 `cargo test`(바뀐 곳과 이웃 시험만), Linux·Windows(비밀번호 데이터베이스가 없는 시스템은 Risks의 N2),
실제 `CLAUDE_CONFIG_DIR`로 `moai update` 실행(실제 홈 쓰기라 하지 않음).
