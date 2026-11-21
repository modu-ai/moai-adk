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
- Q3: 읽기에 실패하면요? → A: 지금은 `Err(_) => return Ok(None)` 이 읽기 오류를 모두 삼켜요. 프로필이 아예 없을 때만 병합된 PR 규칙을 그대로 쓰고, 있는데 못 읽으면 어떤 파일을 왜 못 읽었는지 말하며 거절해요. 프로필은 읽었는데 m-flow 가 아니어서 착지 기록을 못 쓴 경우에도 거절 문장이 읽은 프로필 경로와 이유를 말해요.
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
- REQ-4: IF the project profile exists but cannot be read, THEN `card done` shall refuse, name the project-relative profile path and the reason, change nothing and not apply the merged-pull-request rule.
- REQ-5: Where no profile file exists, `card done` shall keep the merged-pull-request rule · failure: none
- REQ-6: When `card done` refuses a card with a `card.landed` event for lack of a merged pull request, the refusal shall name the project-relative profile path it read and why its work_flow was not used · failure: none
- REQ-7: When `card done` runs in a mixed layout whose `.moai/project.yaml` names the m-flow work_flow with an integration branch and the newest landed head of the card is an ancestor of that branch's tip, the command shall close the card on the landing · failure: none

## Acceptance

- AC-1 (REQ-1…REQ-3): unit tests with temporary project folders: both profiles → the reader returns the `.moai/` profile (a `work_flow` only the new file has is seen); only the earlier file → that file; only the new file with a `rust` layout marker → that file; a new profile that is invalid → its error and no fallback; a symlinked `.moai/project.yaml` is not chosen.
  verify: `cargo test --offline --locked --bin moai-cli mixed_layout`
- AC-2 (REQ-7): temp-repo test: `.moai-cli/project.yaml` without work_flow and `.moai/project.yaml` with m-flow beside a card in review with a recorded landing inside the integration branch → `card done` closes and the `card.done` event names the head and the tip.
  verify: `cargo test --offline --locked --test card_done_mflow_cli`
- AC-3 (REQ-4): temp-repo test: an unreadable profile (invalid content) → `card done` refuses with the profile path (`.moai-cli/project.yaml`) and a reason, not with "no merged pull request", and the card stays in review with no `card.done` event.
  verify: `cargo test --offline --locked --test card_done_mflow_cli`
- AC-4 (REQ-5, REQ-6): temp-repo tests: no profile at all keeps "no merged pull request is linked" as before; a readable profile without m-flow plus a recorded landing → the refusal names the profile path and says its work_flow is not m-flow (or has no integration branch).
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
- D3 (제안, 되돌릴 수 있음, 범주 none): "프로필 없음"은 오류가 아니라 정책 없음이에요. 이유: 프로필이 없는 프로젝트의 동작을 바꾸지 않아요.

## Risks

- 섞인 배치에서 두 프로필의 내용이 서로 다르면 `.moai/` 쪽이 이겨요. 이 저장소에서는 의도한 결과지만, 다른 프로젝트에서 새 파일이 옛 파일보다 덜 갱신된 경우가 있으면 같은 문제가 반대로 나타나요. 거절 문장과 시험이 읽은 경로를 말하도록 해서 알아볼 수 있게 해요.
- 프로필을 읽는 다른 명령(`spec take` 등)도 새 위치 규칙을 얻어요. 그 명령들의 시험 중 섞인 배치를 쓰는 것이 있으면 이 SPEC 의 이웃 검사에서 드러나요.
- `src/profile.rs` 를 다른 카드가 예약 중이라 한 곳만 잡게 돼요. 반환이 늦으면 GREEN 이 밀려요.

## Result

(구현 뒤에 적어요.)
