---
id: SPEC-20261120-reports-move-follow-ups
title: "Reports move follow-ups: one error kind for a blocked apply, a temporary working folder for the relative-HOME test, and a case-insensitive .git"
status: draft
revision: 1
created: 2026-11-20
---

# Reports move follow-ups

## Intent

보고서 폴더를 프로젝트 밖으로 옮기는 변경(SPEC-20261119-update-moves-reports)의 감사가 남긴 작은 셋이에요. 모두 그 변경을 로컬에 넣은 뒤 알려진 일이고, 막는 힘을 바꾸지 않고 말과 시험을 바로잡아요.

1. 막힌 적용의 오류 종류: 같은 "막힘"인데 링크는 `invalid_input`, `.git` 과 HOME 문제는 `internal` 로 나와요(`src/contract.rs` 가 오류 문장 속 낱말로 종류를 정하고, 문장에 표에 있는 낱말이 없으면 `internal` 이 돼요). 사용자가 폴더나 HOME 을 바꾸면 풀리는 일이라 모두 `invalid_input` 이에요. 문장 앞에 `reports move blocked: ` 를 붙이고 표에 그 접두어 한 줄을 더해요. 사람이 읽는 이유 문장은 그대로예요.
2. 상대 HOME 시험의 작업 폴더: `a_home_that_is_relative_or_inside_the_project_blocks_the_move` 는 상대 경로 `relative-home` 을 테스트 프로세스의 작업 폴더(저장소 루트)에서 풀어요. 가드가 깨지면 저장소 안에 폴더가 생겨요. 상대 HOME 은 CLI 를 임시 작업 폴더에서 돌려 시험하고, 그 폴더에 `relative-home` 이 생기지 않았음을 확인해요.
3. `.git` 이름 비교: 하위 폴더의 자기 저장소 검사가 `name == ".git"` 로 대소문자를 구분해요. 대소문자를 구분하지 않는 APFS 에서 `.GIT` 은 같은 폴더예요. 이름을 대소문자 없이 비교해요(어느 파일시스템에서든).

## Non-goals

- 막는 조건을 늘리거나 줄이지 않아요(링크·중첩 저장소·HOME·한도 판단은 그대로).
- 다른 작업의 오류 종류 표를 바꾸지 않아요(접두어 한 줄만 더해요).
- 사람이 읽는 막힘 이유 문장(`reason`)은 바꾸지 않아요.

## Requirements

- REQ-1: The reports move apply operation shall report a blocked plan as the error kind `invalid_input` for every cause.
- REQ-2: The error classification table shall map the message prefix `reports move blocked: ` to `invalid_input`.
- REQ-3: The test of a relative HOME shall run the CLI in a temporary working folder and fail when a folder named like that HOME appears there.
- REQ-4: The reports move plan shall treat a folder named `.git` in any letter case below the report folder as a repository of its own and block.
- REQ-5 (REQ-4): IF a folder below the report folder is named `.GIT`, THEN the plan shall be blocked with no file to move.

## Acceptance

- AC-1 (REQ-1, REQ-2): 링크·중첩 저장소·프로젝트 안 HOME·상대 HOME 네 원인의 막힌 적용이 CLI 에서 모두 `error.kind` `invalid_input` 을 내고 파일은 그대로예요.
  verify: `cargo test --locked --offline --test reports_move_cli every_blocked_apply_reports_invalid_input_through_the_cli`
- AC-2 (REQ-2): 분류 표가 접두어 `reports move blocked: ` 를 `invalid_input` 으로 돌려줘요.
  verify: `cargo test --locked --offline --lib contract::tests::a_reports_move_blocked_message_is_invalid_input`
- AC-3 (REQ-3): 상대 HOME 이 임시 작업 폴더에서 막히고 그 폴더에 `relative-home` 이 생기지 않아요.
  verify: `cargo test --locked --offline --test reports_move_cli every_blocked_apply_reports_invalid_input_through_the_cli`
- AC-4 (REQ-4, REQ-5): `.GIT` 폴더가 보고서 폴더 아래에 있으면 계획이 막히고 아무것도 옮기지 않아요.
  verify: `cargo test --locked --offline --test reports_move_cli a_git_folder_in_any_letter_case_below_the_report_folder_blocks_the_move`

## Scope

- write: `.moai/specs/SPEC-20261120-reports-move-follow-ups/spec.md`, `src/reports_move.rs`(적용 오류 문장과 `.git` 비교), `src/contract.rs`(표에 한 줄과 단위 시험), `tests/reports_move_cli.rs`(시험 둘)
- read-only: `src/ops/reports_move_apply.rs`, `src/ops/reports_move_plan.rs`, `tests/common/home.rs`
- forbidden: 막는 조건의 추가·삭제, 다른 작업의 분류 규칙, 이유 문장의 바꿈

## Plan

1. SPEC 을 커밋해요(이 문서). `spec check` 를 돌려요.
2. RED: `every_blocked_apply_reports_invalid_input_through_the_cli`, `a_git_folder_in_any_letter_case_below_the_report_folder_blocks_the_move`, `contract::tests::a_reports_move_blocked_message_is_invalid_input` 를 먼저 써서 실패하는 것을 봐요. 기존 `a_home_that_is_relative_or_inside_the_project_blocks_the_move` 는 프로젝트 안 HOME 만 남기고 상대 HOME 은 새 시험이 임시 작업 폴더에서 다뤄요.
3. GREEN: `apply` 의 막힘 문장을 `reports move blocked: …; nothing moved` 로, 표에 접두어 한 줄, `.git` 비교를 `eq_ignore_ascii_case` 로 바꿔요.
4. 확인: `reports_move_cli` 전체, `--lib contract`, fmt, clippy `--all-targets`, `moai check affected`, 변이(표의 줄 지우기, 비교를 `==` 로 되돌리기, 문장 접두어 빼기).

## Risks

- 오류 문장에 `blocked` 가 그대로 있어서 기존 시험(`err.contains("blocked")`)과 `project_update` 의 문장 읽기가 그대로 통과해야 해요: `moai check affected` 가 둘 다 돌려요.
- 분류 표는 `contains` 로 찾아서 접두어가 다른 문장에 들어 있으면 종류가 바뀌어요: 접두어가 길고(`reports move blocked: `) 이 작업의 문장에만 있어요.
- 대소문자 무시 비교는 소문자 `.git` 아닌 이름(`.Git`, `.GIT`)도 막아서, 그런 이름의 일반 폴더가 보고서 아래에 있으면 막혀요: 막는 쪽으로 틀리는 거라 안전한 쪽이에요.
- 상대 HOME 시험이 CLI 프로세스를 쓰므로 시험 시간이 조금 늘어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 막힌 적용은 모두 `invalid_input` 으로 한다 | 사용자가 폴더나 HOME 을 바꾸면 풀리는 일이고, 링크 막힘이 이미 그 종류예요(감사 후속 제안) | 표의 한 줄 지우기 |
| D-2 | 접두어 `reports move blocked: ` 를 쓰고 이유 문장은 건드리지 않는다 | 표가 낱말로 찾으니 이 작업만의 긴 접두어가 다른 문장을 건드리지 않아요 | 접두어 빼기 |
| D-3 | `.git` 은 어느 파일시스템에서든 대소문자 없이 비교한다 | 같은 코드가 모든 곳에서 같게 도는 편이 시험하기 쉽고, 틀려도 막는 쪽이에요 | `==` 로 되돌리기 |

## Result

구현 끝(2026-11-20, 가지 `spec/20261120-reports-move-follow-ups`). 먼저 시험을 쓰고 실패하는 것을 봤어요(`every_blocked_apply…` 는 `internal` 로, `.GIT` 은 `ready` 로 실패, 단위 시험은 `Internal` 로 실패). 그다음 고쳤고, 변이 넷(`target/c/mut222.sum`)을 모두 시험이 잡았으며 각 변이 뒤 파일은 그대로 되돌렸어요.

- AC-1: met — `cargo test --locked --offline --test reports_move_cli every_blocked_apply_reports_invalid_input_through_the_cli` 통과(링크·중첩 저장소·프로젝트 안 HOME·상대 HOME 네 원인이 CLI 에서 `error.kind` `invalid_input`, 문장에 `blocked` 가 남고, 파일과 실제 HOME 은 그대로). 변이 "분류 표의 접두어 줄 지우기", "적용 문장에서 접두어 빼기", "표의 종류를 `Blocked` 로"에 이 시험이 실패했어요.
- AC-2: met — `cargo test --locked --offline --lib contract::tests::a_reports_move_blocked_message_is_invalid_input` 통과. 변이 "표의 줄 지우기"와 "종류를 `Blocked` 로"에 실패했어요.
- AC-3: met — 같은 시험이 상대 HOME(`relative-home`)을 임시 작업 폴더에서 돌리고 그 폴더에 `relative-home` 이 없음을 확인해요(통과). 옛 시험 `a_home_that_is_relative_or_inside_the_project_blocks_the_move` 는 저장소 루트에서 풀리던 상대 HOME 사례를 빼고 프로젝트 안 HOME 만 남겼어요(이전 SPEC 의 검증 줄이 그 이름으로 돌려서 이름은 그대로 두고 주석에 적었어요).
- AC-4: met — `cargo test --locked --offline --test reports_move_cli a_git_folder_in_any_letter_case_below_the_report_folder_blocks_the_move` 통과(`.GIT`·`.Git` 모두 막히고 파일과 HOME 그대로). 변이 "`eq_ignore_ascii_case` 를 `==` 로"에 실패했어요.
- 확인: `reports_move_cli` 23개와 `--lib contract` 11개 통과. 전체 `moai check affected` 결과는 병합 준비 노트에 줄째로 적어요(이 SPEC 이 닫힌 뒤의 병합 커밋 위에서 돌려요).
