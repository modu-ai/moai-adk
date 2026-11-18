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

tip eea59f4ff1에서 읽은 사실: OMP는 이미 에이전트 정의(`src/agent_defs.rs::HOSTS`), hook(`src/hooks.rs::HOSTS`), 레인 의무 줄,
프로필(`src/profile.rs::KNOWN_HOSTS`), host 폴더(`src/host_home.rs`), 프로젝트 host 단계(`omp-hooks`·`omp-config`·`omp-agents`)에서
지원돼요. 빠진 곳은 론처 쪽(하위 명령, 도움말, 첫 실행 설정, 세션 등록, glm 키 길)과 설치 쪽(`moai update`·프로젝트 skill·출력 스타일)과
감사 독립성 표예요.

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
- REQ-8: The host support document shall not state that OMP is unsupported.
- REQ-9: The change shall carry its own SPEC at `.moai/specs/SPEC-20261118-omp-host/spec.md`, tracked by git in the same commit.

## Acceptance

- AC-1 (REQ-1, REQ-2): 표에 17줄이 있고 열이 `feature`, `claude`, `codex`, `omp`, `evidence`, `gap`이며 모든 상태가 어휘 안이에요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_the_survey_has_the_columns_and_rows OmpHostSurveyTest.test_every_status_is_in_the_vocabulary`
- AC-2 (REQ-3, REQ-4, REQ-5): 근거 파일·기호가 있고, `(list)` 줄 8개(프로필, 에이전트 정의, hook, 레인 의무 줄, 세션 등록, `moai update`, 출력 스타일, 감사 독립성)가
  코드의 host 목록과 같아요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_the_evidence_file_and_symbol_exist OmpHostSurveyTest.test_a_list_row_agrees_with_the_hosts_in_the_code`
- AC-3 (REQ-6): omp `gap` 줄 9개가 모두 `slice B`나 `leader card`로 닫는 곳을 적었고, `gap`이 아닌 줄에는 닫는 곳 문장이 없어요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_an_omp_gap_names_the_slice_or_card_that_closes_it OmpHostSurveyTest.test_a_row_without_a_gap_has_no_gap_text`
- AC-4 (REQ-7, REQ-8): 문서에 옛 문장 `OMP — 지원하지 않음`이 없고 `2026-11-17`과 `#2378`이 있어요.
  verify: `python3 -I tests/omp_host_survey_test.py OmpHostSurveyTest.test_the_document_replaces_the_2378_unsupported_line`
- AC-5 (REQ-9): SPEC이 git에 추적되고 SPEC 검사기가 받아들여요.
  verify: `git ls-files .moai/specs/SPEC-20261118-omp-host`; `moai-cli spec check SPEC-20261118-omp-host` (이 트리의 빌드: `cargo build --bin moai-cli`)
- AC-6 (REQ-1…REQ-9): 이웃 검사가 통과해요(기존 host 지원표 시험, 문서 범위 분류, 검사 명령 검사).
  verify: `python3 -I tests/host_matrix_test.py`; `python3 -I tests/ci_scope_test.py`; `cargo test --offline --locked --test ac_command_lint`

## Scope

- write: `docs/host-support-matrix.md`, `tests/omp_host_survey_test.py`, `.moai/specs/SPEC-20261118-omp-host/spec.md`
- read-only: `src/**` (조사표의 근거), `skills/moai/**`
- forbidden: 실제 HOME, 네트워크, host 실행

## Plan

1. 시험 먼저: `tests/omp_host_survey_test.py`(RED: 표가 없어 7개 모두 오류).
2. `docs/host-support-matrix.md`에 조사표 절을 더하고 론처 문단의 #2378 문장을 바꿔요.
3. 변이 6개로 시험이 틀린 표를 잡는지 확인해요.
4. 조각 B(별도 SPEC, 쓰기 범위 확대 뒤): `moai omp [-l|-a]`와 `--print-command`는 먼저 OMP 18.x의 명령줄(모델, 첫 프롬프트, 권한)을 공식
   문서로 읽고, glm 키가 OMP에 닿는 길(환경 변수)을 `src/key_vault.rs`·`src/session.rs`의 기존 glm 길과 같은 규칙(argv와 출력에
   키 없음)으로 정한 뒤, 실제 OMP 18.18.7(이 기계에 설치됨)에서 skill과 에이전트가 로드되는지 관찰해요.

## Risks

- 표는 코드를 읽어 적은 것이고 host를 띄워 확인하지 않았어요(`not_run`). 시험은 근거 파일·기호의 존재와 host 목록 일치만 고정하고,
  `yes`가 실제로 동작한다는 뜻은 아니에요.
- `(list)`가 아닌 줄(`Codex(HostLaunch)`, `read_here`, `default_profile`, `host_folder`, `Omp`, `omp-agents`, `OMP_CONFIG_REASON`,
  `shadows every other user context file`)의 상태는 사람이 코드를 읽어 적은 값이라 시험이 새 host가 생겨도 알아채지 못해요.
- 표의 `n/a`(OMP의 AGENTS 라우팅, 상태줄)는 "하지 않기로 한 것"이고 코드의 이유 문장이 근거예요(`home_routing.rs`, `project_hosts.rs`).
  그 결정이 바뀌면 줄을 고쳐야 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | OMP는 전체 host다. 2026-11-10 결정(Issue #2378: OMP 어댑터 #2362 제거, OMP 지원 안 함)을 사용자 지시 2026-11-17(카드 C-62)이 **대체**한다 | 카드 C-62 본문(사용자 지시 2026-11-17을 옮겨 적은 것): `error: unrecognized subcommand omp … moai cc|codex|glm -l|-a` 출력과 "The user asks for an OMP launcher and for everything OMP-related to be supported." | 지시가 바뀌면 `docs/host-support-matrix.md`의 해당 문단과 표의 omp 열을 되돌리고 새 결정을 기록 |
| D-2 | 조각 A는 코드를 바꾸지 않고 문서와 시험만 낸다. 론처 구현은 쓰기 범위가 넓어진 뒤 조각 B | 카드 쓰기 범위에 론처 파일이 없고 `src/session.rs`는 카드 C-47 범위 | 범위가 넓어지면 조각 B를 바로 시작 |
| D-3 | 조사표의 근거는 `파일::기호`이고 host 목록 줄은 코드의 배열을 읽어 같은지 본다 | 카드 seq 3 "시험이 코드에 고정" | 목록 줄 표기를 지우고 존재 확인만 남김 |
| D-4 | 표의 빈 곳은 조각 B가 닫는 것(론처 쪽 5줄)과 리더가 카드로 만들 것(설치 쪽 3줄, 감사 독립성 1줄)으로 나눈다. 레인은 카드를 더하지 않는다 | 에이전트 모드 규칙: 카드 추가는 리더 | 리더가 카드 번호를 알려 주면 `gap` 열을 `card C-n`으로 바꿈 |
| D-5 | 결정 기록은 이 SPEC의 D-1과 `docs/host-support-matrix.md`의 문단이다. `docs/decisions-*.md`는 쓰기 범위 밖이라 리더가 옮겨 적는다 | 쓰기 범위 | 리더가 결정 로그에 한 줄을 더함 |

## Result

- AC-1: met — `python3 -I tests/omp_host_survey_test.py` 7개 OK; 표 17줄(위 스크립트 `len(rows)`), 열 6개. RED(표 추가 전): 7개 모두 오류(`ValueError: substring not found`, 절 없음).
- AC-2: met — 같은 파일. `(list)` 줄 8개가 코드 배열과 같음; 변이 M1(세션 등록 줄을 omp `yes`로)·M2(프로필 줄 codex를 `n/a`로)·M4(기호 이름 바꿈)·M6(감사 독립성 줄 omp `yes`)이 모두 FAIL.
- AC-3: met — omp `gap` 9줄 = `slice B` 5 + `leader card` 4(스크립트로 센 값); 변이 M3(`slice B (C-62)` 문장을 `-`로)이 FAIL.
- AC-4: met — 변이 M5(옛 문장 `OMP — 지원하지 않음`을 되돌림)가 FAIL; 문서에 `2026-11-17`·`#2378`.
- AC-5: met — `moai-cli spec check SPEC-20261118-omp-host` errors 0 warnings 0(다른 worktree의 tip 근처 빌드 `moai-adk-c89/target/debug/moai-cli`로; 처음엔 REQ-3·REQ-6이 `shall` 두 개라 `req_gears` 오류 2건이어서 REQ를 아홉 개로 쪼갬).
- AC-6: met — `host_matrix_test` 5개, `ci_scope_test` 19개, `cargo test --offline --locked --test ac_command_lint` 5 passed OK.
- 돌리지 않은 것: host 실행(표는 코드를 읽어 적음), 조각 B(론처, 키 길, 실제 OMP 세션), 전체 `cargo test`.
- 변이 6개는 문서를 백업(`target/matrix.bak`)에서 복원했고 `cmp`로 같음을 확인. 처음엔 백업 폴더가 없어 변이 5개가 문서에 그대로 남았다가, `git checkout`으로 기준선에 되돌리고 같은 편집을 스크립트(`target/apply_doc.py`)로 다시 적용했다.
