---
id: SPEC-20261118-omp-launcher
title: "OMP launcher - moai omp [-l|-a], the glm key path, help and lane-duty spelling (slice B)"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261118-omp-host
---

# OMP launcher - moai omp [-l|-a], the glm key path, help and lane-duty spelling (slice B)

## Intent

카드 C-62(사용자 지시 2026-11-17: "OMP를 위한 론처를 만들고 OMP 관련한 모든 것을 지원")의 두 번째 조각이에요. 조각 A(`SPEC-20261118-omp-host`, PR #3203)는 host별 조사표와 #2378을 대체하는 결정 행을 넣었고, 조사표에서 `slice B (C-62)`로 표시한 빈자리 다섯 곳이 이 조각의 몫이에요: 론처 하위 명령, 도움말·사용법 줄, 첫 실행의 프로젝트 설정, 세션 등록의 host 목록, glm 키 길. 조각 A가 미룬 두 가지도 함께 닫아요: 레인 의무 줄의 OMP 철자(`/skill:moai`)와, 옛 #2378 문구를 읽던 Rust 시험(`tests/launcher_three_hosts.rs`).

OMP 18.18.7(설치본 `omp --help`, 소스 `@oh-my-pi/pi-coding-agent`, 2026-11-18 읽음)에서 읽은 사실이에요.

- 명령줄: `omp [MESSAGES]`, `--model`, `--thinking <off|minimal|low|medium|high|xhigh|max|auto>`, `--approval-mode <always-ask|always-ask-cached|write|yolo>`, `--auto-approve`. 세션 이름을 정하는 플래그는 없어요.
- 승인: `src/tools/approval.ts`의 모드 목록이 위와 같고, 설정 `tools.approvalMode`의 기본값은 `yolo`예요. 그래서 OMP 세션의 승인 모드는 플래그로 분명히 적어야 론처의 규칙(역할 세션은 전체 권한, `--permissions`는 질문 유지, 역할 없는 세션은 host 설정)을 지켜요.
- 첫 메시지: 위치 인수는 `session.prompt(...)`로 가요(`src/main.ts:849`). `/skill:<이름>` 명령이 풀리는 곳은 대화 입력 쪽(`submitInteractiveInput`, `src/main.ts:489`)에서만 읽었어요. 그래서 첫 메시지로 역할 진입 줄을 보내지 않고, Claude 론처처럼 세션 안에서 사람이 치는 줄을 안내해요.
- z.ai 키: OMP가 z.ai(GLM) 공급자의 키를 읽는 변수는 `ZAI_API_KEY`예요(`src/cli/help-extra.ts:30`, `pi-ai/README.md:960`).

## Non-goals

- 설정 파일에서 모델·노력 읽기(C-47 조각 1이 설정 파일을 만들어요). 이 조각은 `--model`과 `--effort`로 준 값만 OMP로 넘기고, 주지 않으면 OMP 자신의 설정이 정해요.
- 조사표의 `leader card` 다섯 줄(skill 설치 목록, 프로젝트 skill 설치, 출력 스타일, 감사 독립성 표, 레인 의무 줄은 이 조각이 닫아요)은 그대로 `gap`이에요.
- OMP 세션을 띄워 skill과 agent가 로드되는 것을 관찰하기(카드 기준 5). 이 환경에서는 `not_run`으로 적어요(Result).

## Requirements

- REQ-1: The launcher shall accept `omp` as a host word with the options of `cc`, `glm` and `codex`.
- REQ-2: The launcher shall accept `--print-command` as a second spelling of `--dry-run` on every host word.
- REQ-3: Where `--dry-run` is given, the OMP launcher shall print the composed command as JSON without starting the host.
- REQ-4: The OMP command shall start with `omp`.
- REQ-5: Where `--model` is given, the OMP command shall carry `--model` with that value.
- REQ-6: Where `--effort` is `low`, `medium`, `high`, `xhigh` or `max`, the OMP command shall carry `--thinking` with that value.
- REQ-7: Where `--effort` is any other value, the OMP launcher shall pass no `--thinking` and say so in a note.
- REQ-8: While a session has a role and no `--permissions`, the OMP command shall carry `--approval-mode yolo`.
- REQ-9: Where `--permissions` is given, the OMP command shall carry `--approval-mode always-ask-cached`.
- REQ-10: A session without a role shall carry no approval flag.
- REQ-11: The OMP command shall carry no first prompt that enters a role.
- REQ-12: The OMP notes shall name `/skill:moai --mode-leader` and `/skill:moai --mode-agent` as the lines typed inside the session.
- REQ-13: Where `--advisor` is given, the OMP launcher shall pass no advisor and say so in a note.
- REQ-14: Where the model named is a z.ai model, the launcher shall give the glm key to the OMP process as `ZAI_API_KEY`.
- REQ-15: The launcher shall keep the key out of the command, the printed JSON and every note.
- REQ-16: IF the vault cannot give the glm key for a z.ai model, THEN the OMP launcher shall stop with the same `blocked:` line as the glm launcher.
- REQ-17: Where no glm key exists for a z.ai model, the OMP launcher shall go on without `ZAI_API_KEY` and say in a note that OMP uses its own login.
- REQ-18: A `-a` session without `--name` shall take its `agent-N` number from the shared counter.
- REQ-19: A refused OMP launch shall take no agent number.
- REQ-20: Where the first launch in a folder that is not a project uses `omp`, the launcher shall record the host `omp` in the profile.
- REQ-21: The root help and the usage and error lines shall name `omp` beside `cc`, `glm` and `codex`.
- REQ-22: The lane-duty line shall type `/skill:moai` for OMP.
- REQ-23: The host support document shall mark omp `yes` on every row this slice closes.
- REQ-24: The host support document shall not state that OMP is unsupported.
- REQ-25: The change shall carry its own SPEC at `.moai/specs/SPEC-20261118-omp-launcher/spec.md`, tracked by git in the same commit.

## Acceptance

verify 명령의 `omp_launcher`는 새 시험 파일 `tests/omp_launcher.rs`예요(옛 `tests/launcher_three_hosts.rs`를 대신해요).

- AC-1 (REQ-1, REQ-4): `moai omp`, `-l`, `-a`가 `omp`로 시작하는 명령을 만들어요.
  verify: `cargo test --offline --locked --test omp_launcher omp_roles_compose_an_omp_command`
- AC-2 (REQ-2, REQ-3): `--print-command`가 모든 host 단어에서 `--dry-run`과 같고 host를 띄우지 않아요.
  verify: `cargo test --offline --locked --test omp_launcher print_command_is_dry_run_on_every_host omp_dry_run_starts_nothing`
- AC-3 (REQ-5, REQ-6, REQ-7): `--model`과 `--effort`가 `--model`과 `--thinking`으로 가고, 다른 노력 값은 안내 줄만 남겨요.
  verify: `cargo test --offline --locked --test omp_launcher omp_model_and_thinking omp_unknown_effort_is_a_note`
- AC-4 (REQ-8, REQ-9, REQ-10): 역할 세션은 `--approval-mode yolo`, `--permissions`는 `always-ask-cached`, 역할 없는 세션은 승인 플래그가 없어요.
  verify: `cargo test --offline --locked --test omp_launcher omp_approval_modes`
- AC-5 (REQ-11, REQ-12, REQ-13): 첫 프롬프트가 없고 안내 줄이 `/skill:moai --mode-*`를 말하고, 조언자는 넘기지 않고 알려요.
  verify: `cargo test --offline --locked --test omp_launcher omp_has_no_role_prompt_and_names_the_entry_lines omp_advisor_is_a_note`
- AC-6 (REQ-14, REQ-15): z.ai 모델에서만 `ZAI_API_KEY`가 host 프로세스로 가고, 값은 명령·JSON·안내 어디에도 없어요.
  verify: `cargo test --offline --locked --test omp_launcher omp_key_reaches_the_process_for_a_glm_model_only omp_key_never_appears_in_output`
- AC-7 (REQ-16, REQ-17): 금고가 막히면 glm 론처와 같은 `blocked:` 줄로 멈추고, 키가 없으면 안내 줄만 남기고 가요.
  verify: `cargo test --offline --locked --test omp_launcher omp_blocked_vault_stops_like_glm omp_without_a_key_goes_on_with_a_note`
- AC-8 (REQ-18, REQ-19, REQ-20): 번호를 받고, 거절된 시작은 번호를 받지 않고, 새 폴더는 프로필에 `omp`를 적어요.
  verify: `cargo test --offline --locked --test omp_launcher omp_agent_takes_a_number_and_a_refusal_takes_none`; `cargo test --offline --locked --test launcher_auto_init the_first_launch_outside_git_initializes_the_folder_on_every_host`
- AC-9 (REQ-21): 도움말과 사용법과 오류 줄이 omp를 말해요.
  verify: `cargo test --offline --locked --test omp_launcher omp_is_named_in_help_usage_and_errors`
- AC-10 (REQ-22): 레인 의무 줄이 OMP에서 `/skill:moai`를 쳐요.
  verify: `cargo test --offline --locked --test lane_duty_line`; `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_the_lane_duty_row_follows_the_omp_entry_spelling`
- AC-11 (REQ-23, REQ-24): 조사표의 이 조각 줄이 `yes`이고 문서가 OMP 미지원을 말하지 않아요.
  verify: `python3 -I tests/omp_host_survey_test.py`; `python3 -I tests/host_matrix_test.py`; `cargo test --offline --locked --test omp_launcher the_support_matrix_no_longer_says_omp_is_unsupported`
- AC-12 (REQ-1..REQ-24): 이웃 시험이 통과해요.
  verify: `cargo test --offline --locked --test launcher_roleless --test launcher_agent_entry --test launcher_key_vault --test launcher_glm_key_file --test launcher_terminal_output --test launcher_permissions_setting --test launcher_agent_numbering --test lane_duty_line`
- AC-13 (REQ-25): SPEC이 git에 추적되고 검사기가 받아들여요.
  verify: `git ls-files .moai/specs/SPEC-20261118-omp-launcher`; `moai-cli spec check SPEC-20261118-omp-launcher`

## Scope

- write: `src/main.rs`, `src/session.rs`, `src/help.rs`, `src/launcher_init.rs`, `src/launcher_terminal.rs`, `src/lane_duty.rs`, `docs/host-support-matrix.md`, `tests/omp_launcher.rs`(새 파일), `tests/launcher_three_hosts.rs`(삭제), `tests/lane_duty_line.rs`, `tests/launcher_auto_init.rs`, `tests/launcher_agent_entry.rs`, `tests/omp_host_survey_test.py`, `.moai/specs/SPEC-20261118-omp-launcher/spec.md`
- 리더가 더한 범위(C-62 노트 2183): `tests/help_layout.rs`, `tests/output_names_moai.rs`, `tests/fixtures/help/{all,all-60,all-80,all-dumb,root,root-60,root-80,root-dumb}.txt`, `tests/fixtures/protocol/v1/current/01-usage-unknown-subcommand.json`. 골든은 도구(`MOAI_HELP_GOLDEN_WRITE`, `MOAI_UPDATE_FIXTURES`)로만 다시 쓰고 폭 한도는 늘리지 않아요.
- read-only: `skills/moai/**`, `docs/**`(위 문서 제외)
- forbidden: 실제 HOME, 네트워크, 실제 host 실행, 릴리스 태그

## Plan

1. 시험 먼저: `tests/omp_launcher.rs`(RED: `moai omp`는 사용법 오류), `tests/omp_host_survey_test.py`의 이 조각 줄 `yes` 기대와 옛 문구 금지(RED), `tests/lane_duty_line.rs`의 OMP 철자, `tests/launcher_auto_init.rs`의 omp 줄.
2. `src/session.rs`: `HOSTS`에 `omp`, `omp_argv`, `ZAI_API_KEY` 길, 역할 안내 줄, 조언자·노력 안내.
3. `src/main.rs`: `Omp(HostLaunch)`, `--print-command` 별칭, `host_of`, 사용법·오류 문구, `operation`·`role_of`·`launcher_is_human` 갈래.
4. `src/help.rs`, `src/launcher_terminal.rs`, `src/launcher_init.rs`, `src/lane_duty.rs`.
5. 문서 조사표를 `yes`로 바꾸고 옛 문구를 없앤 뒤 `tests/launcher_three_hosts.rs`를 지워요.
6. 이 문서와 바꾼 파일을 읽는 모든 시험을 찾아 함께 돌려요(`grep -rln host-support-matrix tests/ src/ tools/`, 도움말 골든, `product_cli`).

## Risks

- 설치된 OMP 18.18.7은 이 조각의 시험에서 시작하지 않아요(`not_run`). 위 사실은 소스를 읽은 것이고, 첫 메시지의 `/skill:` 처리는 읽지 못해서 안내 줄 쪽을 택했어요(D-3).
- `ZAI_API_KEY`가 z.ai 공급자의 변수 이름이라는 것은 OMP 문서에서 읽었고, 그 변수로 실제 요청이 가는지는 보지 않았어요.
- 이 조각은 skill 설치(`moai update`가 OMP 폴더에 skill을 넣는 일)를 하지 않아서, OMP에 skill이 없으면 세션에서 `/skill:moai`가 풀리지 않아요. 그 줄은 조사표에서 `leader card` gap으로 남아 있어요.
- 도움말 골든과 `product_cli`류 시험이 host 목록 문구를 읽을 수 있어요. 이름 필터로 일부만 돌리면 놓쳐요(조각 A의 교훈), 그래서 Plan 6처럼 읽는 시험을 찾아 돌려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `--print-command`는 `--dry-run`의 별칭이다(모든 host 단어) | 카드 기준 1이 `--print-command`를 말하고, 기존 시험은 `--dry-run`을 써요 | 별칭 한 줄 삭제 |
| D-2 | 역할 세션은 `--approval-mode yolo`, `--permissions`는 `always-ask-cached`, 역할 없는 세션은 플래그 없음 | 사용자 결정 2026-11-10(#2359)의 규칙을 OMP 모드 이름으로 옮김; OMP 기본값이 `yolo`라 분명히 적음 | `omp_argv`의 한 갈래 |
| D-3 | 역할 진입 줄은 첫 메시지로 보내지 않는다 | 위 "첫 메시지" 사실; Claude 론처와 같은 방식 | 확인된 뒤 `codex -a`처럼 첫 프롬프트로 |
| D-4 | `ZAI_API_KEY`는 모델 이름이 z.ai 모델(`zai/` 앞말 또는 `glm` 포함)일 때만 넘긴다 | 비밀을 host에 필요할 때만 주는 최소 노출 | 조건 한 줄 |
| D-5 | 이 조각은 skill 설치와 설정 파일 읽기를 하지 않는다 | 범위(조사표의 leader card 줄, C-47) | 후속 카드 |
| D-6 | 요구가 25개라서 `size_requirements` 경고(12개 초과)가 남는데, 한 SPEC·한 PR로 간다 | `moai omp`는 host 단어, 명령 조립, 키 길, 도움말이 함께 있어야 한 번에 동작하는 한 기능이에요. 나누면 중간 조각에서 `moai omp`가 반쯤 동작해요. 요구마다 한 문장·한 `shall`이고 AC가 짝을 이뤄요 | 다음에 나누려면 키 길(REQ-14~17)이 가장 떼기 쉬워요 |

## Result

- AC-1: met — `omp_launcher omp_roles_compose_an_omp_command` 통과. RED(구현 전): `omp_launcher` 15개 전부 실패(`error: unrecognized subcommand 'omp'`).
- AC-2: met — `print_command_is_dry_run_on_every_host`(cc·glm·codex·omp의 `--print-command`와 `--dry-run` JSON이 같음)와 `omp_dry_run_starts_nothing` 통과.
- AC-3: met — `omp_model_and_thinking`(low·medium·high·xhigh·max)와 `omp_unknown_effort_is_a_note`(ultracode는 안내 줄만) 통과.
- AC-4: met — `omp_approval_modes` 통과: 역할 없는 세션 플래그 없음, `-l|-a` `yolo`, `--permissions` `always-ask-cached`.
- AC-5: met — `omp_has_no_role_prompt_and_names_the_entry_lines`, `omp_advisor_is_a_note` 통과. `launcher_agent_entry`의 철자 시험은 `/skill:moai --mode-*` 두 줄을 허용 목록에 더해 7개 통과.
- AC-6: met — `omp_key_reaches_the_process_for_a_glm_model_only`(stub host가 `ZAI_API_KEY`를 받음: 금고 키, 변수가 금고보다 우선, `zai/…`·`glm` 모델에서만)와 `omp_key_never_appears_in_output`(dry-run·실행·stderr·args 어디에도 값 없음, `env_keys`에 이름만) 통과.
- AC-7: met — `omp_blocked_vault_stops_like_glm`(옛 키 파일이 있으면 glm 론처와 같은 `blocked: legacy_key_only` 줄, host 시작 안 함, 번호 안 받음)와 `omp_without_a_key_goes_on_with_a_note` 통과.
- AC-8: met — `omp_agent_takes_a_number_and_a_refusal_takes_none`과 `launcher_auto_init the_first_launch_outside_git_initializes_the_folder_on_every_host`(omp 줄 더함, 프로필 `hosts: ["omp"]`) 통과. RED: 후자 1개 실패(`unrecognized subcommand 'omp'`).
- AC-9: met — `omp_is_named_in_help_usage_and_errors`, `help_layout` 10개(폭 한도는 늘리지 않음), `help::` 단위 시험 18개 통과. 골든 9개(도움말 8, 프로토콜 1)는 `MOAI_HELP_GOLDEN_WRITE`·`MOAI_UPDATE_FIXTURES`로만 다시 썼고 `git diff --shortstat`은 42 insertions, 34 deletions. omp를 이름 대지 않는 변경 줄은 다시 줄바꿈된 이어지는 줄과 `[aliases: --print-command]` 4줄(D-1)이며 리더에게 보고했어요(카드 노트 `blocked-agent-30-golden-alias`).
- AC-10: met — `lane_duty_line` 8개 통과(RED 1개 실패), `omp_host_survey_test` 9개 중 `test_the_lane_duty_row_follows_the_omp_entry_spelling` 통과.
- AC-11: met — `omp_host_survey_test` 9개 OK(RED: 11건 실패), `host_matrix_test` 5개 OK, `the_support_matrix_no_longer_says_omp_is_unsupported` 통과.
- AC-12: met — `help_layout`, `output_names_moai`, `product_cli`(79), `omp_launcher`, `launcher_agent_entry`, `launcher_roleless`, `launcher_key_vault`, `launcher_glm_key_file`, `launcher_terminal_output`, `launcher_permissions_setting`, `launcher_agent_numbering`, `launcher_auto_init`, `lane_duty_line` 합계 207개 통과; `--lib` 1427개와 `--bins` 28+28개, `ac_command_lint` 5개, `claude_python_adapter_removed` 5개 통과; 파이썬 `ci_scope_test`, `goal_block_test`, `repo_scans_test` OK.
- AC-13: met — `git ls-files .moai/specs/SPEC-20261118-omp-launcher`로 추적, `moai-cli spec check SPEC-20261118-omp-launcher`는 오류 0, 경고 1건(`size_requirements`, D-6).
- 변이 시험: 복사본에서 5개를 하나씩 적용(적용은 `cmp`로 확인, 복원도 `cmp`로 확인)하고 모두 시험에 걸렸어요 — 키를 모든 모델에 보냄, 프로세스에 키를 안 줌, 역할 세션이 질문 모드, 막힌 금고 무시, 레인 의무 줄 OMP를 `/moai`로 되돌림(`killed 5/5`).
- 실제 OMP 18.18.7 관찰: `omp --mode rpc --no-session`을 임시 폴더에서 띄우고 `get_available_commands`를 보내자 `skill:moai`(source `skill`)가 목록에 있었어요(모델 호출 없음). 이는 화면 세션이 아니고 `moai omp -l`이 띄운 것도 아니에요. agent는 `~/.omp/agent/agents/moai-*.md` 파일 5개가 있다는 것만 보았고 OMP가 로드하는 것은 보지 못했어요.
- 돌리지 않은 것: `moai omp -l`로 띄운 OMP 화면 세션(카드 기준 5, 사람이 필요), OMP가 첫 메시지의 `/skill:`을 풀어 주는지, `ZAI_API_KEY`로 실제 z.ai 요청이 가는지, 전체 `cargo test`와 macOS·Windows(PR CI에 맡겨요), `moai update`의 OMP skill 설치(조사표 leader card gap).
