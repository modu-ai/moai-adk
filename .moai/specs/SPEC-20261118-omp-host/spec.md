---
id: SPEC-20261118-omp-host
title: "OMP as a full host - host-keyed survey table and the decision that replaces #2378 (slice A)"
status: draft
revision: 1
created: 2026-11-18
parent: null
---

# OMP as a full host - host-keyed survey table and the decision that replaces #2378 (slice A)

## Intent

카드 C-62(사용자 지시 2026-11-17, `~/MoAI/moai-cowork`에서 `moai omp -l`이 `unrecognized subcommand omp`로 끝남): OMP를 전체 host로
지원해요. 카드의 범위는 셋이에요: (1) `moai omp`, `moai omp -l|-a` 론처, (2) 코드에서 host로 갈라지는 모든 자리를 claude/codex/omp
열로 조사한 표와 각 빈 곳을 고치거나 카드로 남기기, (3) 도움말에 omp 이름.

이 SPEC은 조각 A예요. 코드를 바꾸지 않고, 카드의 seq 3(조사표 + 코드에 고정하는 시험)과 seq 4(#2378을 대체하는 결정 기록)를 해요.
론처 구현(seq 1·2)과 실제 OMP 세션 관찰(seq 5)은 조각 B예요. 조각 B는 `src/main.rs`, `src/session.rs`, `src/help.rs`,
`src/launcher_init.rs`, `src/launcher_terminal.rs`, `tests/launcher*`를 건드려야 하는데 카드 쓰기 범위 밖이고 일부는 카드 C-47의
범위라서, 조각 A와 따로 두었어요(카드 노트 `blocked-agent-30-scope`).

tip eea59f4ff1에서 읽은 사실: OMP는 이미 에이전트 정의(`src/agent_defs.rs::HOSTS`), hook(`src/hooks.rs::HOSTS`),
프로필(`src/profile.rs::KNOWN_HOSTS`), host 폴더(`src/host_home.rs`), 프로젝트 host 단계(`omp-hooks`·`omp-config`·`omp-agents`)에서
지원돼요. 빠진 곳은 론처 쪽(하위 명령, 도움말, 첫 실행 설정, 세션 등록, glm 키 길)과 설치 쪽(`moai update`·프로젝트 skill·출력 스타일)과
감사 독립성 표, 레인 의무 줄의 철자(코드는 `/moai`, OMP는 `/skill:moai`만 받음)예요.

## Non-goals

- `moai omp`, `-l`, `-a`, `--print-command`와 그 키 길(조각 B, 카드 C-62 seq 1·2).
- 실제 OMP 세션이 skill과 에이전트를 읽는지 보는 관찰(조각 B, seq 5).
- `src/update.rs`·`src/session.rs`(카드 C-47 범위)와 `src/main.rs`의 변경.
- 표의 `gap` 줄 중 리더가 카드로 만들어야 하는 것의 카드 만들기(레인은 카드를 더하지 않아요).
- 질문 도구(OMP `ask`)는 카드 C-52가 맡아요.

## Requirements

- REQ-1: The host support document `docs/host-support-matrix.md` shall carry a survey table with one row for each host-keyed place in the code and one status each for claude, codex and omp.
- REQ-2: Each status in the survey table shall be `yes`, `gap` or `n/a`.
- REQ-3: Each row of the survey table shall name a source file and a symbol in it.
- REQ-4: The source file and the symbol that a row names shall exist in the repository.
- REQ-5: Where a row names a host list in the code, the hosts marked `yes` in that row shall equal the hosts that list contains.
- REQ-6: IF a row marks omp as `gap`, THEN the row shall name the slice or the card that closes it.
- REQ-7: The host support document shall record that the user instruction of 2026-11-17 replaces the 2026-11-10 decision of Issue #2378.
- REQ-8: The host support document shall state the 2026-11-10 decision that OMP is unsupported only together with the statement that the decision is replaced.
- REQ-9: The survey row of the lane-duty line shall mark omp as `yes` only if the prefix that `src/lane_duty.rs` pairs with omp is the spelling that `skills/moai/workflows/mode-agent.md` names for OMP.
- REQ-10: The change shall carry its own SPEC at `.moai/specs/SPEC-20261118-omp-host/spec.md`, tracked by git in the same commit.

## Acceptance

- AC-1 (REQ-1, REQ-2): 표에 17줄(시험은 17줄 이상을 요구)이 있고 열이 `feature`, `claude`, `codex`, `omp`, `evidence`, `gap`이며 모든 상태가 어휘 안이에요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_the_survey_has_the_columns_and_rows OmpHostSurveyTest.test_every_status_is_in_the_vocabulary`
- AC-2 (REQ-3, REQ-4, REQ-5): 근거 파일·기호가 있고, `(list)` 줄 7개(프로필, 에이전트 정의, hook, 세션 등록, `moai update`, 출력 스타일, 감사 독립성)가
  코드의 host 목록과 같아요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_the_evidence_file_and_symbol_exist OmpHostSurveyTest.test_a_list_row_agrees_with_the_hosts_in_the_code`
- AC-3 (REQ-6): omp `gap` 줄 10개가 모두 `slice B`나 `leader card`로 닫는 곳을 적었고, `gap`이 아닌 줄에는 닫는 곳 문장이 없어요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_an_omp_gap_names_the_slice_or_card_that_closes_it OmpHostSurveyTest.test_a_row_without_a_gap_has_no_gap_text`
- AC-4a (REQ-9): 레인 의무 줄의 omp 상태는 코드의 접두사(`/moai`)와 모드 워크플로가 OMP에 적은 철자(`/skill:moai`)가 같을 때만 `yes`예요. 지금은 다르니 `gap`이고, 코드가 고쳐지면 시험이 줄을 `yes`로 바꾸라고 요구해요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_the_lane_duty_row_follows_the_omp_entry_spelling`
- AC-4 (REQ-7, REQ-8): 옛 문구 `OMP — 지원하지 않음(2026-11-10 결정`은 문서에 정확히 한 번(`tests/launcher_three_hosts.rs`가 읽어요) 있고, 그 문장이 `대체`, `2026-11-17`, `#2378`을 함께 담아요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_the_document_replaces_the_2378_unsupported_line`
- AC-5 (REQ-10): SPEC이 git에 추적되고 SPEC 검사기가 받아들여요.
  verify: `git ls-files .moai/specs/SPEC-20261118-omp-host`; `moai-cli spec check SPEC-20261118-omp-host` (이 트리의 빌드: `cargo build --bin moai-cli`)
- AC-6 (REQ-1…REQ-10): 이웃 검사가 통과해요(기존 host 지원표 시험, 문서 범위 분류, 검사 명령 검사).
  verify: `python3 -I tests/host_matrix_test.py`; `python3 -I tests/ci_scope_test.py`; `cargo test --offline --locked --test ac_command_lint`

## Scope

- write: `docs/host-support-matrix.md`, `tests/omp_host_survey_test.py`, `.moai/specs/SPEC-20261118-omp-host/spec.md`
- read-only: `src/**` (조사표의 근거), `skills/moai/**`
- forbidden: 실제 HOME, 네트워크, host 실행

## Plan

1. 시험 먼저: `tests/omp_host_survey_test.py`(RED: 표가 없어 7개 모두 오류).
2. `docs/host-support-matrix.md`에 조사표 절을 더하고 론처 문단의 #2378 문장을 바꿔요.
3. 변이로 시험이 틀린 표를 잡는지 확인해요(표 6개, 레인 의무 줄 규칙 3개, 옛 문구 규칙 2개).
4. 조각 B(별도 SPEC, 쓰기 범위 확대 뒤): `moai omp [-l|-a]`와 `--print-command`는 먼저 OMP 18.x의 명령줄(모델, 첫 프롬프트, 권한)을 공식
   문서로 읽고, glm 키가 OMP에 닿는 길(환경 변수)을 `src/key_vault.rs`·`src/session.rs`의 기존 glm 길과 같은 규칙(argv와 출력에
   키 없음)으로 정한 뒤, 실제 OMP 18.18.7(이 기계에 설치됨)에서 skill과 에이전트가 로드되는지 관찰해요. 조각 B는 `tests/launcher_three_hosts.rs`("론처는 정확히 세 host"를 #2378 결정으로 고정)도 고쳐야 하고, `src/lane_duty.rs`와 `tests/lane_duty_line.rs`의 omp 철자도 `/skill:moai`로 맞춰야 해요(감사 #3203 B1·N4).

## Risks

- 표는 코드를 읽어 적은 것이고 host를 띄워 확인하지 않았어요(`not_run`). 시험은 근거 파일·기호의 존재와 host 목록 일치만 고정하고,
  `yes`가 실제로 동작한다는 뜻은 아니에요.
- `(list)`가 아닌 줄(`Codex(HostLaunch)`, `read_here`, `default_profile`, `host_folder`, `Omp`, `omp-agents`, `OMP_CONFIG_REASON`,
  `shadows every other user context file`, 그리고 레인 의무 줄)의 상태는 사람이 코드를 읽어 적은 값이라 시험이 새 host가 생겨도 알아채지 못해요. 근거 기호 확인은 "파일에 그 글자가 있다"는 검사라, 이름만 남기고 정의를 지워도(주석, 호출부) 통과해요(감사 #3203 N3).
- 기존 Rust 시험 `tests/launcher_three_hosts.rs::the_support_matrix_records_the_removal_and_launcher_docs_list_three_hosts`가 이 문서에서 `OMP — 지원하지 않음(2026-11-10 결정`을 정확히 한 번 찾아요. 첫 head는 그 문구를 지워 PR CI의 `ci / linux-x86_64`가 이 시험 1건으로 붉어졌어요(감사도 놓침: 감사는 "조각 B가 고칠 시험"으로만 읽음). 그래서 문구를 역사 기록으로 남기고(D-6), 조각 B가 그 시험을 고칠 때 지워요.
- 목록에 host가 있다는 것은 그 자리가 host에서 동작한다는 뜻이 아니에요(감사 #3203 B1: 레인 의무 줄은 omp를 갖고 있지만 철자가 틀림). 목록 줄은 이름만 대조하니, 이름이 같아도 틀릴 수 있는 줄은 `(list)`를 빼고 따로 시험해요.
- 표의 `n/a`(OMP의 AGENTS 라우팅, 상태줄)는 "하지 않기로 한 것"이고 코드의 이유 문장이 근거예요(`home_routing.rs`, `project_hosts.rs`).
  그 결정이 바뀌면 줄을 고쳐야 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | OMP는 전체 host다. 2026-11-10 결정(Issue #2378: OMP 어댑터 #2362 제거, OMP 지원 안 함)을 사용자 지시 2026-11-17(카드 C-62)이 **대체**한다 | 카드 C-62 본문(사용자 지시 2026-11-17을 옮겨 적은 것): `error: unrecognized subcommand omp … moai cc|codex|glm -l|-a` 출력과 "The user asks for an OMP launcher and for everything OMP-related to be supported." | 지시가 바뀌면 `docs/host-support-matrix.md`의 해당 문단과 표의 omp 열을 되돌리고 새 결정을 기록 |
| D-2 | 조각 A는 코드를 바꾸지 않고 문서와 시험만 낸다. 론처 구현은 쓰기 범위가 넓어진 뒤 조각 B | 카드 쓰기 범위에 론처 파일이 없고 `src/session.rs`는 카드 C-47 범위 | 범위가 넓어지면 조각 B를 바로 시작 |
| D-3 | 조사표의 근거는 `파일::기호`이고 host 목록 줄은 코드의 배열을 읽어 같은지 본다 | 카드 seq 3 "시험이 코드에 고정" | 목록 줄 표기를 지우고 존재 확인만 남김 |
| D-4 | 표의 빈 곳은 조각 B가 닫는 것(론처 쪽 5줄)과 리더가 카드로 만들 것(설치 쪽 3줄, 감사 독립성 1줄, 레인 의무 줄 1줄)으로 나눈다. 레인은 카드를 더하지 않는다 | 에이전트 모드 규칙: 카드 추가는 리더 | 리더가 카드 번호를 알려 주면 `gap` 열을 `card C-n`으로 바꿈 |
| D-5 | 결정 기록은 이 SPEC의 D-1과 `docs/host-support-matrix.md`의 문단이다. `docs/decisions-*.md`는 쓰기 범위 밖이라 리더가 옮겨 적는다 | 쓰기 범위 | 리더가 결정 로그에 한 줄을 더함 |
| D-6 | 옛 문구를 문서에 역사 기록으로 한 번 남기고 같은 문장에 "대체"를 적는다. 기존 Rust 시험을 고치지 않는다 | `tests/launcher_three_hosts.rs`는 `tests/launcher*`라 카드 C-47 범위이고 쓰기 범위 밖; PR CI 실패로 확인(run 44251319022: 4896개 중 1 실패) | 조각 B가 그 시험을 고치면서 문구와 이 줄을 지움 |

## Result

- AC-1: met — `python3 -I tests/omp_host_survey_test.py` 8개 OK; 표 17줄, 열 6개(`rows 17`은 `survey_rows()`를 읽는 스크립트로 센 값). RED(표 추가 전): 7개 모두 `setUp`의 `ValueError: substring not found`로 오류(절 없음). 그래서 `test_the_document_replaces_the_2378_unsupported_line`이 자기 이유로 붉어지는 것은 변이 M5로 따로 보였다(감사 #3203 N5).
- AC-2: met — `(list)` 줄 7개가 코드 배열과 같음; 변이 M1(세션 등록 줄을 omp `yes`로)·M2(프로필 줄 codex를 `n/a`로)·M4(기호 이름 바꿈)·M6(감사 독립성 줄 omp `yes`)이 모두 FAIL.
- AC-3: met — omp `gap` 10줄 = `slice B` 5 + `leader card` 5(스크립트로 센 값); 변이 M3(`slice B (C-62)` 문장을 `-`로)이 FAIL.
- AC-4: met — 시험이 옛 문구가 정확히 한 번이고 같은 문장에 `대체`·`2026-11-17`·`#2378`이 있는지 봄. RED: 문구를 지운 첫 수리 head에서 새 규칙 시험이 `0 != 1`로 실패(그전에 PR CI가 `launcher_three_hosts` 1건으로 붉음, run 44251319022). 변이: `대체` 삭제(FAIL), 문구 두 번(FAIL). 기존 Rust 시험 `cargo test --offline --locked --test launcher_three_hosts` 4 passed(수리 전 로컬 재현: 3 passed 1 failed).
- AC-4a: met — 감사 #3203 B1의 수리. RED: 새 시험 `test_the_lane_duty_row_follows_the_omp_entry_spelling`이 수리 전 문서에서 `'yes' != 'gap'`으로 실패(코드 `/moai`, 워크플로 `/skill:moai`). 복사본(`target/mut`)에서 변이 셋: 코드를 `/skill:moai`로 고치고 문서는 `gap` 그대로(FAIL), 문서를 `yes`로 되돌림(FAIL 2건), 워크플로 철자를 `/moai`로(FAIL).
- AC-5: met — `moai-cli spec check SPEC-20261118-omp-host` errors 0 warnings 0.
- AC-6: met — `host_matrix_test` 5개, `ci_scope_test` 19개, `cargo test --offline --locked --test ac_command_lint` 5 passed.
- 돌리지 않은 것: host 실행(표는 코드를 읽어 적음), 조각 B(론처, 키 길, 실제 OMP 세션), 전체 `cargo test`.
- 감사 #3203 첫 head cca27a912a는 FAIL(B1). 수리: 레인 의무 줄 omp를 `gap`으로(리더 카드 필요), 새 시험, 하한을 17줄·7목록으로 올림(N1), 프로젝트 host 단계 줄 이름과 `n/a` 정의(N2), 건너뛴 skill 줄 문구(N6), SPEC Risks·Plan에 N3·N4. 변이 6개와 수리 변이는 문서를 백업에서 복원해 `cmp`로 같음을 확인했다. 처음 변이를 돌릴 때 백업 폴더가 없어 변이가 문서에 그대로 남았다가 `git checkout`과 같은 편집 스크립트로 복구한 일이 있었다.
- PR CI 실패(첫 수리 head 8a5a0a9149의 `ci / linux-x86_64`, run 44251319022): `launcher_three_hosts`의 `the_support_matrix_records_the_removal_and_launcher_docs_list_three_hosts` 1건(4896개 중), 내가 지운 옛 문구를 읽는 시험이었다. Python 쪽 이웃 검사(`host_matrix_test`, `ci_scope_test`)는 이 시험을 보지 못했고 감사도 놓쳤다. 이 문서를 읽는 시험을 `grep -rln host-support-matrix tests/ src/ tools/`로 다시 찾아 `launcher_three_hosts.rs`, `goal_block_test.py`, `host_matrix_test.py`를 확인했고 모두 통과.
