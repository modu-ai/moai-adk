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
- REQ-5 (REQ-1): IF the relocated folder or a folder below it on the way to a target is a link, THEN the step shall block that host without following the link.
- REQ-6: The Codex skill copy shall stay in the `.agents/skills` folder under HOME.
- REQ-7: The user-scope guards shall refuse writing into a folder that `CLAUDE_CONFIG_DIR` or `CODEX_HOME` names.
- REQ-8: The Go-generation cleanup and legacy scans shall keep their HOME-relative roots.
- REQ-9: The repository SPECs shall pass `moai spec check --all`.

## Acceptance

- AC-1 (REQ-1, REQ-3): 두 변수를 절대 경로로 주면 라우팅 계획·적용·되돌리기, 에이전트, 출력 스타일, Claude 스킬 복사, Codex 진단이 옮겨진 폴더를 쓰고 HOME 아래 `.claude`·`.codex`에는 아무것도 생기지 않아요.
  verify: `cargo test --locked --offline --test host_home_cli -- routing_writes_and_reverts_in_the_relocated_folders agents_and_the_output_style_follow_the_variables the_skill_copy_follows_claude_config_dir the_codex_doctor_reads_codex_home`
- AC-2 (REQ-2): 빈 값·상대 경로는 그 도구만 막고(`blocked`) 사유가 변수 이름을 말하며 값은 되풀이하지 않아요; `codex-doctor`는 종료 1로 거절해요.
  verify: `cargo test --locked --offline --test host_home_cli an_empty_or_relative_variable_blocks_its_host_by_name`;
  `cargo test --locked --offline --lib host_home::`
- AC-3 (REQ-4): 계정 홈이 아닌 홈(임시 HOME, 다른 계정 홈)은 변수가 있어도 그 아래 폴더를 써요.
  verify: `cargo test --locked --offline --test host_home_cli the_variables_count_only_for_the_account_home`
- AC-4 (REQ-5): 링크인 `CLAUDE_CONFIG_DIR`은 라우팅·에이전트·스킬 복사에서 막히고 링크 너머에 아무것도 쓰지 않아요.
  verify: `cargo test --locked --offline --test host_home_cli a_linked_config_folder_is_refused_not_followed`
- AC-5 (REQ-6): `CODEX_HOME`이 있어도 Codex 스킬 복사 위치는 HOME 아래 `.agents/skills/moai`예요.
  verify: `cargo test --locked --offline --test host_home_cli the_skill_copy_follows_claude_config_dir`
- AC-6 (REQ-7): 프로젝트 `.claude`가 `CLAUDE_CONFIG_DIR`이거나 `.codex`가 `CODEX_HOME`이면 거절하는 기존 시험이 그대로 통과해요.
  verify: `cargo test --locked --offline --test claude_settings_cli --test codex_config_cli`
- AC-7 (REQ-8): Go 정리·옛 스킬 검사 시험이 그대로 통과하고 그 뿌리 표는 바뀌지 않아요.
  verify: `cargo test --locked --offline --test update_cleanup_cli`;
  `git diff 25881165be -- src/skills_legacy.rs src/baseline.rs` → 출력 없음
- AC-8 (REQ-9): 저장소의 SPEC이 모두 통과해요.
  verify: `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`
- AC-9 (REQ-1…REQ-9): 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --lib`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --test home_routing_cli --test home_style_cli --test agents_home_cli --test codex_adapter --test codex_config --test recovery_cli --test home_isolation_scan --test git_env_scrub --test product_cli`;
  `cargo test --locked --offline --test update_cli --test update_backup_replace --test update_command_cli --test update_end_to_end_cli --test update_home_layer_cli --test update_hooks_cli --test update_keys_cli --test update_known_good --test update_known_good_replace --test update_moved_receipts --test update_route --test update_tux_golden --test update_unmanaged_failclosed --test update_user_layer_cli --test update_version_folders`;
  `cargo fmt --check`; `rustfmt --check --edition 2021 src/ops/*.rs`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests.skill_references_test`

## Scope

- write: `src/host_home.rs`, `src/lib.rs`, `src/home_routing.rs`, `src/agent_install.rs`, `src/home_style.rs`,
  `src/update.rs`, `src/recovery.rs`, `src/update_cleanup.rs`, `src/claude.rs`, `src/hosts/codex_config.rs`,
  `src/ops/codex_doctor.rs`, `tests/host_home_cli.rs`, `tools/repo_scans.json`,
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

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 도구 폴더는 `src/host_home.rs` 하나가 정함: Claude `CLAUDE_CONFIG_DIR`, Codex `CODEX_HOME`, 없으면 HOME 아래; OMP는 변수 없음 | 카드 C-79, 공식 문서(Intent) | 모듈 삭제, 호출하는 곳을 HOME 고정으로 |
| D2 | 변수는 계정 홈(비밀번호 데이터베이스, 시험은 `MOAI_ACCOUNT_HOME`)일 때만 셈; 다른 홈은 그 아래 폴더 | 시험 30여 개가 임시 HOME에 실제 환경을 물려받아, `$HOME`만 비교하면 실제 설정 폴더에 쓸 수 있음(이 세션에 `CLAUDE_CONFIG_DIR`이 설정돼 있어 확인) | `$HOME`과 비교(그 전에 시험 환경을 모두 비워야 함) |
| D3 | 빈 값·상대 경로는 거절, 메시지는 변수 이름만(값 없음); `~/x`도 상대로 봄 | 리더 지시, 다른 폴더 기준으로 읽힐 위험 | 상대 경로를 작업 폴더 기준으로 풀기 |
| D4 | 변수 경로는 적힌 그대로(정규화 안 함); 그 폴더 자체가 링크면 `~/.claude`처럼 막음 | 리더 지시(링크 처리 바꾸지 않기) | `canonicalize` 뒤 사용 |
| D5 | 계획에서 폴더를 정할 수 없으면 그 도구만 `blocked`(표시 경로는 HOME 아래), 적용·되돌리기는 오류 | 다른 도구 계속 진행 | 계획 전체 실패 |
| D6 | Codex 스킬은 `~/.agents/skills` 그대로 | Codex 문서·기존 `hosts/codex.rs` | `CODEX_HOME/skills` |
| D7 | 사용자 범위 가드(`claude.rs`, `codex_config.rs`)는 계정 홈과 상관없이 변수가 가리키는 폴더를 막고, 빈 값·상대 경로면 거절 | 가드는 넓게 막는 쪽이 안전 | 예전처럼 상대 값을 작업 폴더 기준으로 |
| D8 | `src/update.rs`는 스킬 폴더를 찾는 줄(`skills_base`·`skills_parent`·`projection_root`)과 그 호출 네 곳, `preflight`의 링크 검사 시작점만 바꿈 | 리더: 피할 수 없을 때만, 보고 | 그 줄들 되돌리기 |
| D9 | Go 세대 정리(`update_cleanup` 뿌리 표)·`skills_legacy`·`baseline`은 HOME 기준 유지; `projection_extra`는 HOME 밖이면 건너뜀 | Go 도구의 고정 설치 경로 | 각 표를 리졸버로 |

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 빌드 폴더는 이 worktree의 `target/`,
`CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는 한 번에 하나(heavy slot 안).

커밋: 기준 `25881165be` 위 `52dfc86a5d`(시험 7개만, RED) → `7f87d69038`(리졸버·호출하는 곳·문서·이 SPEC; 시험 파일은
`cargo fmt`만) → `48add8e1ad`(HOME 밖 링크 경우 추가) → 이 Result(다음 커밋). RED: `host_home_cli` exit 101, 5 failed
(경로가 여전히 HOME 아래 `.claude`·`.codex`, 빈 값·상대 경로·링크가 막히지 않음); 계정 홈 고정 시험과 doctor의
`CODEX_HOME` 읽기는 지금 동작이라 처음부터 통과. GREEN 7 passed.

mutation 5개 모두 KILLED: M1 계정 홈 규칙 빼기, M2 상대 경로 받기, M3 에이전트가 도구 폴더 자체 링크 검사 빼기,
M4 스킬 복사의 링크 검사를 HOME부터(처음엔 살아남음 — 시험의 옮긴 폴더가 모두 임시 HOME 안이었음 → `48add8e1ad`로
HOME 밖 링크 경우를 더한 뒤 KILLED), M5 doctor가 리졸버를 안 씀.

실제 홈 안전: 이 세션에는 `CLAUDE_CONFIG_DIR`이 설정돼 있어요. 시험 전후로 그 폴더의 `CLAUDE.md`·`settings.json`·
`agents`·`skills/moai`·`output-styles`와 HOME 아래 `.codex/AGENTS.md`·`.codex/agents`의 수정 시각·크기를 비교했고
(읽기만), 모든 실행(mutation 포함) 뒤에도 같았어요.

- AC-1: met — `48add8e1ad`에서 `host_home_cli` exit 0, 7 passed
- AC-2: met — 같은 실행, `--lib host_home::` exit 0, 6 passed; M2·M5 KILLED
- AC-3: met — 같은 실행; M1 KILLED
- AC-4: met — 같은 실행(HOME 안·밖 링크); M3·M4 KILLED
- AC-5: met — 같은 실행
- AC-6: met — `7f87d69038`에서 `claude_settings_cli` 22·`codex_config_cli` 10 passed
- AC-7: met — `update_cleanup_cli` 11 passed; `git diff 25881165be -- src/skills_legacy.rs src/baseline.rs` 출력 없음
- AC-8: met — `spec_cli` exit 0, 17 passed; 이 가지에서 빌드한 `moai spec check --all --project .` exit 0, SPEC 6개, 오류 0·경고 0
- AC-9: met — `7f87d69038`에서 `--lib` 1374, `--bin moai` 27, 시험 대상 28개 한 번에 exit 0(`host_home_cli` 7,
  `agents_home_cli` 17, `claude_settings_cli` 22, `codex_adapter` 20, `codex_config` 10, `codex_config_cli` 10,
  `git_env_scrub` 3, `home_isolation_scan` 16, `home_routing_cli` 11, `home_style_cli` 16, `recovery_cli` 7, `product_cli` 75,
  `update_*` 16개 모두 통과); `ac_command_lint` 5; `cargo fmt --check` 0; `rustfmt --check --edition 2021 src/ops/*.rs` 0;
  clippy `-D warnings` 0; Python 46개 OK(처음 1개 실패: 새 시험이 꾸러미를 해시해 저장소를 읽음 → `tools/repo_scans.json`의
  `skill-package-agents` 항목에 추가)

실행 안 함: 전체 `cargo test`(바뀐 곳과 이웃 시험만), Linux·Windows(비밀번호 데이터베이스가 없는 Windows는 `$HOME`과
비교 — 시험 없음), 실제 `CLAUDE_CONFIG_DIR`로 `moai update` 실행(실제 홈 쓰기라 하지 않음).
