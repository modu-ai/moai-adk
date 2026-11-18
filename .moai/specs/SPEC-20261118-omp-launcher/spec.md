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

- write: `src/main.rs`, `src/session.rs`, `src/help.rs`, `src/launcher_init.rs`, `src/launcher_terminal.rs`, `src/lane_duty.rs`, `docs/host-support-matrix.md`, `tests/omp_launcher.rs`(새 파일), `tests/launcher_three_hosts.rs`(삭제), `tests/lane_duty_line.rs`, `tests/launcher_auto_init.rs`, `tests/omp_host_survey_test.py`, `.moai/specs/SPEC-20261118-omp-launcher/spec.md`
- 시험이 고쳐야 하면 같은 범위 안의 `tests/launcher_*.rs`(카드 범위에 이름이 있는 파일만)
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

(구현 뒤에 AC마다 한 줄씩 적어요.)
