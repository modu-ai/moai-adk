---
id: SPEC-20261121-mixed-profile-read
title: "In a mixed .moai-cli/.moai project the profile reader opens the newer profile, and card done names the profile it read instead of hiding a read problem"
status: draft
revision: 1
created: 2026-11-21
---

# 섞인 .moai-cli/.moai 저장소에서 프로필을 제대로 읽고, 못 읽으면 이유를 말한다

## Interview

- Q1: 무엇이 문제였나요? → A: 리더가 2026-11-21 에 빌드 13(4.2647.13, tip 041536fe07)으로 착지한 카드를 닫다가 발견했어요. 이 저장소에는 `.moai-cli/project.yaml`(work_flow 없음, 옛 파일)과 `.moai/project.yaml`(work_flow m-flow, 통합 브랜치 v2026.11.0)이 함께 있어요. 프로필 읽기는 상태 폴더를 `state_root::dir` 로 정하고, `.moai/` 에 유효한 표지가 없으면 옛 `.moai-cli` 를 골라요. 그래서 `card done` 의 M-Flow 착지 검사(`mflow_landing`)가 work_flow 없는 프로필을 읽고 조용히 `None` 을 돌려, `card landed` 로 착지를 기록한 카드도 "no merged pull request is linked" 로 거절돼요. 그 때문에 착지 카드 5장이 review 에 멈춰 있어요.
- Q2: 어느 쪽을 읽어야 하나요? → A: 두 프로필이 모두 있는 섞인 배치에서는 새 배치(`.moai/`)의 프로필을 읽어요(리더 제안, 가역). 어느 한쪽만 있거나 배치가 이미 `rust` 이면 지금처럼 그 배치의 폴더를 읽어요. 쓰기는 이 SPEC 이 바꾸지 않아요(한 곳만 쓴다는 규칙 그대로).
- Q3: 읽기에 실패하면요? → A: 지금은 `Err(_) => return Ok(None)` 이 읽기 오류를 모두 삼켜요. 프로필이 스텁이어도(카드 명령이 id 만 쓴 프로필) PR 로 닫는 카드는 계속 닫혀야 하므로 규칙은 그대로 두고, 착지 기록이 있는 카드를 PR 규칙으로 거절할 때 어떤 프로필을 읽었고 왜(읽지 못했거나 m-flow 가 아니라서) 착지 기록을 쓰지 않았는지 거절 문장 끝에 붙여요.
- Q4: 멈춘 카드 5장을 닫는 건요? → A: 이 SPEC 의 코드가 병합되고 설치된 CLI 가 새 빌드가 된 뒤의 운영 작업이라 SPEC 범위 밖이에요. 카드 본문이 "그다음 닫기"를 요청하므로 병합 뒤에 리더 지시로 해요.

## Intent

`card done` 은 M-Flow 프로필을 읽고 착지 head 가 통합 브랜치 tip 의 조상인지 확인해 닫아요. 프로필을 엉뚱한 파일에서 읽거나 읽기 오류를 삼키면 같은 카드가 영원히 안 닫히고 이유도 안 보여요. 이 SPEC 은 (1) 읽기 전용 프로필 위치 규칙을 섞인 배치에서 새 배치로 정하고, (2) `card done` 이 프로필 읽기 문제를 경로와 이유와 함께 말하게 해요. 프로필을 읽는 모든 명령(`spec take`, `integration-base`, `work-flow-plan` 등)이 같은 위치 규칙을 얻어요.

## Non-goals

- 프로필·상태 쓰기 위치 변경, 배치 판정(`classify`) 변경, 옛 `.moai-cli/project.yaml` 삭제나 이동. 정리는 사용자가 정해요.
- 멈춘 카드 5장을 닫는 운영 작업(Q4).
- `card done` 의 다른 거절 문장, `card landed` 의 규칙, 병합된 PR 규칙 자체.

## Requirements

- REQ-1: Where both `.moai/project.yaml` and `.moai-cli/project.yaml` are regular files and the state layout is not `rust`, the profile reader shall read `.moai/project.yaml`.
- REQ-2: Where only one of the two profile files exists or the state layout is `rust`, the profile reader shall read the profile file of the layout's own state folder · failure: none
- REQ-3 (REQ-1): IF `.moai/project.yaml` is chosen in a mixed layout and cannot be parsed or validated, THEN the reader shall return that error without falling back to `.moai-cli/project.yaml`.
- REQ-4: IF the project profile exists but cannot be read and the card has a recorded landing, THEN the merged-pull-request refusal of `card done` shall name the project-relative profile path and the reason.
- REQ-5: Where the readable profile names a flow other than m-flow and the card has no recorded landing, `card done` shall keep the plain merged-pull-request refusal without a profile hint · failure: none
- REQ-6: When `card done` refuses a card with a `card.landed` event for lack of a merged pull request, the refusal shall name the project-relative profile path it read and why its work_flow was not used · failure: none
- REQ-7: When `card done` runs in a mixed layout whose `.moai/project.yaml` names the m-flow work_flow with an integration branch and the newest landed head of the card is an ancestor of that branch's tip, the command shall close the card on the landing · failure: none

## Acceptance

- AC-1 (REQ-1…REQ-3): unit tests with temporary project folders: both profiles → the reader returns the `.moai/` profile (a `work_flow` only the new file has is seen); only the earlier file → that file; only the new file with a `rust` layout marker → that file; a new profile that is invalid → its error and no fallback; a symlinked `.moai/project.yaml` is not chosen.
  verify: `cargo test --offline --locked --lib mixed_layout`
- AC-2 (REQ-7): temp-repo test: `.moai-cli/project.yaml` without work_flow and `.moai/project.yaml` with m-flow beside a card in review with a recorded landing inside the integration branch → `card done` closes and the `card.done` event names the head and the tip.
  verify: `cargo test --offline --locked --test card_done_mflow_cli`
- AC-3 (REQ-4): temp-repo test: a profile with a valid project id whose work_flow names a flow that does not exist, with a landing on record → `card done` keeps the "no merged pull request is linked" refusal and adds that a local landing is recorded, the profile path (`.moai-cli/project.yaml`) and that it could not be read; the card stays in review with no `card.done` event; with a newer `.moai/project.yaml` that cannot be read the newer path is named and the earlier m-flow profile is no fallback.
  verify: `cargo test --offline --locked --test card_done_mflow_cli`
- AC-4 (REQ-5, REQ-6): temp-repo tests: a readable profile without m-flow plus a recorded landing → the refusal keeps "no merged pull request is linked" and adds that a local landing is recorded, the profile path and why its work_flow was not used (no work_flow, or the flow's name); the same profile without a landing → the plain refusal with no hint.
  verify: `cargo test --offline --locked --test card_done_mflow_cli`
- AC-5 (REQ-1…REQ-7): at most five mutants of the new code, each turning a test red, are restored.
  verify: manual: the mutant table in the Result.

## Scope

- write: `.moai/specs/SPEC-20261121-mixed-profile-read/**`, `src/profile.rs` (the read-path rule and its tests only), `src/card/work.rs` (`mflow_landing` and the refusal), `tests/card_done_mflow_cli.rs`
- read-only: `src/state_root.rs`, `src/card/spec_link.rs`, `src/ledger/workflow.rs`, `.moai/specs/SPEC-20261120-card-done-mflow/spec.md`
- forbidden: profile writes (`apply`, `write`, baseline files), `src/state_root.rs` classification, the real HOME and ledger, the five stuck cards' ledger rows
- 읽은 사실(2026-11-21, tip `041536fe07`): `src/profile.rs:388-413` `state_path` 가 `crate::state_root::path(&root, "")` 로 폴더를 정하고 `read_profile`(`src/profile.rs:1050`)이 `target(root)` 를 읽어요. `src/state_root.rs:566` `dir` 는 `classify(root).generation == "rust"` 일 때만 `.moai` 예요. `src/card/work.rs:492-498` `mflow_landing` 은 `read_profile` 의 어떤 오류도 `Ok(None)` 으로 바꿔요. `profile.rs` 는 다른 카드가 예약 중이라 읽기 경로 한 곳만 짧게 잡아요.

## Plan

1. RED: `tests/card_done_mflow_cli.rs` 에 섞인 배치, 읽을 수 없는 프로필, 프로필 없음, m-flow 아닌 프로필+착지 시험을 먼저 쓰고, `src/profile.rs` 단위 시험(`mixed_layout…`)을 더해요.
2. GREEN: `profile.rs` 에 읽기 경로 규칙(`read_path`)과 `PROFILE_MISSING` 상수를 두고 `read_profile` 이 그 경로를 읽게 해요. `work.rs` 의 `mflow_landing` 은 "프로필 없음"만 `None`, 그 밖의 읽기 오류는 경로·이유를 말하는 거절로 바꾸고, m-flow 가 아닌 경우의 이유를 `done` 의 거절 문장 끝에 붙여요.
3. 변형 5개 이하, 이웃 검사(fmt, ops rustfmt, clippy 전체, `check affected`, 파이썬, `spec check`, `trailer-check`)를 돌려 Result 에 적어요.

## Decisions

각 결정은 (무엇 · 선택지 · 기본값 · 되돌릴 수 있음 · 범주 · 기한 · 결정자) 순서예요. 모두 제안이고 승인은 SPEC 이 병합될 때 사용자·리더가 골라요.

- D1 (제안, 되돌릴 수 있음, 범주 none): 섞인 배치에서는 `.moai/` 프로필을 읽어요. 대안: 두 파일을 말하며 거절. 기본값: `.moai/` 읽기(리더 제안, 이 저장소의 카드가 막히지 않아요). 되돌리기: 규칙 한 곳.
- D2 (제안, 되돌릴 수 있음, 범주 none): 읽기만 바꾸고 쓰기 위치는 그대로예요. 이유: 쓰기까지 옮기면 마이그레이션 의미가 바뀌어요.
- D3 (제안, 되돌릴 수 있음, 범주 none): "프로필 파일이 없음"은 오류가 아니라 정책 없음으로 남겨요(`mflow_landing` 안에서). `card done` 은 그 전에 프로필의 project_id 로 장부를 찾기 때문에 프로필이 없으면 거기서 먼저 멈추고, 이 SPEC 의 시험은 그 경로를 만들 수 없어요(시험 대상 아님).

## Risks

- 섞인 배치에서 두 프로필의 내용이 서로 다르면 `.moai/` 쪽이 이겨요. 이 저장소에서는 의도한 결과지만, 다른 프로젝트에서 새 파일이 옛 파일보다 덜 갱신된 경우가 있으면 같은 문제가 반대로 나타나요. 거절 문장과 시험이 읽은 경로를 말하도록 해서 알아볼 수 있게 해요.
- 프로필을 읽는 다른 명령(`spec take` 등)도 새 위치 규칙을 얻어요. 그 명령들의 시험 중 섞인 배치를 쓰는 것이 있으면 이 SPEC 의 이웃 검사에서 드러나요.
- `src/profile.rs` 를 다른 카드가 예약 중이라 한 곳만 잡게 돼요. 반환이 늦으면 GREEN 이 밀려요.

## Result

구현 결과예요(2026-11-21, 머리 5f8baeb181).

- AC-1: met — `cargo test --offline --locked --lib mixed_layout` exit 0 (4 passed: 두 프로필이 있으면 `.moai/` 를 읽고 `work_flow` 와 `project_id` 를 봄, 한쪽만 있거나 배치가 `rust` 면 그 폴더, 새 프로필이 읽히지 않으면 오류이고 옛 프로필로 대신하지 않음, 링크인 새 프로필은 고르지 않음). 고치기 전에는 이 시험이 없었어요.
- AC-2: met — `cargo test --offline --locked --test card_done_mflow_cli` exit 0 (9 passed)의 `a_mixed_layout_reads_the_newer_profile_and_closes_on_the_landing`: 옛 프로필(work_flow 없음)과 새 프로필(m-flow)이 함께 있어도 착지 기록으로 닫히고 `card.done` 이벤트가 head 를 남겨요. 고치기 전(읽기 규칙만 없을 때)에는 "no merged pull request is linked" 로 거절돼 실패했어요.
- AC-3: met — 같은 실행의 `an_unreadable_profile_is_named_in_the_refusal` 과 `an_unreadable_newer_profile_is_named_and_the_earlier_one_is_no_fallback`: "no merged pull request is linked" 는 그대로이고 "a local landing is recorded", 프로필 경로(`.moai-cli/project.yaml`, 새 프로필이 깨지면 `.moai/project.yaml`)와 "could not be read" 가 붙고, 카드는 review 로 남고 `card.done` 이벤트는 0개예요.
- AC-4: met — 같은 실행의 `a_profile_that_is_not_m_flow_is_named_when_a_landing_is_on_record`(work_flow 없음, github-flow 두 경우)와 `without_a_landing_the_refusal_carries_no_profile_hint`(착지 기록 없으면 힌트 없음), 기존 `a_project_without_m_flow_keeps_the_pr_rule` 통과. 고치기 전에는 힌트 시험 3개가 실패했어요(4개 실패 중 읽기 규칙으로 1개, 거절 문장으로 3개가 초록).
- AC-5: met — 변형 5개를 하나씩 적용해 `cargo test --offline --locked --no-fail-fast --lib mixed_layout` 와 `--test card_done_mflow_cli` 를 돌렸고 모두 빨갛게 바뀐 뒤(컴파일 오류 없음) 파일을 복원했어요(`git diff` 깨끗).

  | 변형 | 바꾼 곳 | 잡은 시험 |
  | --- | --- | --- |
  | 옛 프로필 존재 조건 제거 | `read_dir` | `mixed_layout_keeps_the_layouts_own_folder_when_only_one_profile_exists` |
  | 새 프로필 존재 조건 제거 | `read_dir` | `mixed_layout` 시험 2개와 `card_done_mflow_cli` 시험 여러 개 |
  | 읽기 오류를 모두 삼킴 | `mflow_landing` | `an_unreadable_profile_is_named_in_the_refusal`, `an_unreadable_newer_profile_is_named_and_the_earlier_one_is_no_fallback` |
  | 힌트를 만들지 않음 | `mflow_landing` | 읽을 수 없는 프로필 시험 2개와 m-flow 아닌 프로필 시험 |
  | 착지 기록 없이도 힌트 | `mflow_landing` | `without_a_landing_the_refusal_carries_no_profile_hint` |

- 설계 변경 기록: 처음 초안은 읽을 수 없는 프로필을 거절로 만들었는데, 카드 명령이 id 만 쓴 스텁 프로필(`invalid profile` 로 읽힘)인 프로젝트에서 PR 로 닫는 카드가 모두 막혀 시험(`a_project_without_m_flow_keeps_the_pr_rule`)이 빨개졌어요. 그래서 거절 대신 PR 거절 문장에 이유를 붙이는 쪽으로 정했어요(REQ-4).
- 수리(2026-11-21 감사의 B1–B5·N1–N3): B1 트레일러는 base 위 재커밋으로 교정(여섯 커밋 모두 `Moai-Spec: SPEC-20261121-mixed-profile-read`, `trailer-check` failing 0). B2 프로필 해석 거절이 파일 경로를 앞세운다(`{read_rel}: existing YAML requires preservation review`). B3 m-flow 무 `integration_branch` 힌트의 회귀 시험 추가. B4 혼합 배치의 쓰기가 읽기를 따른다(plan 이 새 프로필에 묶이고 apply 가 그 파일을 고치며 영수증이 실제 파일을 이름한다 — `work_flow` 는 사용자 편집 키라 REQ-03A-03대로 plan 델타에 없으므로 그 읽기는 위의 done 시험들이 핀). B5 AC-1 verify 가 `--lib`. N1 힌트는 랜딩 관련 거절에만 붙고, N2 읽는 파일을 메시지와 존재 검사가 이름한다(archive·github-templates·context·check-tree·identity·generation), N3 두 프로필의 `project_id` 불일치는 양쪽 파일과 id 를 이름해 거절. N5 는 후속 카드 몫으로 기록.
- 수리 검증: `cargo test --locked --offline --lib` exit 0 (1556 passed), `--test card_done_mflow_cli` exit 0 (13 passed — 기존 9 + B2·B3·B4·N3 신규 4). fmt·clippy·`check affected`(머지 첫째 부모 base)·나머지 대상은 READY 노트의 요약 줄 그대로.
- 점검(fmt·clippy·`check affected`·파이썬)은 READY 노트에 요약 줄 그대로 적어요.
