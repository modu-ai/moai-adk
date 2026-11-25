---
id: SPEC-20261125-card-ready-pass-landing-path
title: "PASS 판정이 기록된 ready 카드도 리더가 착지 기록까지 이어 가요"
status: draft
revision: 1
created: 2026-11-25
branch: spec/card-ready-pass-landing-path-20261125
---

# PASS 판정이 기록된 ready 카드도 리더가 착지 기록까지 이어 가요

## Intent

mo.ai.kr 리더 덧붙임 8 O(이 카드의 출처)는 레인을 재시작하려고 카드를 `ready` 로 돌린 뒤, PASS 판정이 이미 기록된 그 카드를 착지 기록으로 이어 갈 수 없었다고 말해요. 이 저장소에서 같은 길을 임시 HOME 과 임시 리포지터리로 다시 돌려 봤어요(설치된 빌드 46 으로 돌렸고, 이 트리의 빌드가 아니에요). 카드를 `review` 에서 PASS 판정을 받게 하고, `reopen` 과 `release` 로 `ready` 로 돌린 뒤 세 명령을 불렀어요:

| 명령 | 읽은 결과 |
| --- | --- |
| `card landed <card> --head <H> --by leader` | 거절: "card conflict in state: <card> is `ready`; only a `review` card can be landed" |
| `card assign <card> --auditor <id> --by leader` | 거절: "<card> is ready, and an auditor is assigned to a card in review; nothing was written" |
| `card land <card> --head <H> --by leader` | 상태 검사는 통과해요(`finish` 계열이 `ready`, `active`, `review` 를 받아요). 이 픽스처의 head 에 `Moai-Session` 트레일러가 없어 `authors_unknown` 으로 거절됐고, 그것은 이 SPEC 이 만지지 않는 판정 검사예요 |

코드로도 확인했어요: `card landed` 의 상태 검사는 `src/ledger/workflow.rs` 의 `land` 함수(`from != "review"`), `card assign` 의 것은 `src/card/roles.rs` 의 `assign` 이에요. 판정 기록(`card verdict`, `record_verdict`)은 카드 상태를 보지 않아서 `ready` 카드에도 PASS 가 남아요. 그래서 막히는 곳은 "PASS 가 있는 `ready` 카드를 기록하는" 두 명령뿐이에요. 이 SPEC 은 그 두 곳을 좁게 열고, 판정과 head 검사는 약하게 하지 않아요.

선택지 (D-1):

| 선택지 | 내용 | 따르는 일 | 고름 |
| --- | --- | --- | --- |
| A. 두 기록 명령이 PASS 있는 `ready` 카드를 받아요 | `card landed` 와 `card assign` 이 `ready` 를 받되, 카드에 그 head 의 PASS 판정이 있을 때만이에요. 판정이 없거나 FAIL 이 최신이면 거절해요 | `src/ledger/workflow.rs` 와 `src/card/roles.rs` 만 바뀌어요. 새 상태 전이가 없어요 | **추천 기본값** |
| B. 리더 전용 `ready → review` 되돌리기 명령 | 새 명령(또는 `card reopen --by leader`)이 `ready` 카드를 `review` 로 돌려 기존 `landed`, `assign` 이 그대로 받게 해요 | `confirm_leader`(리더 release 작업이 head `78784f8b5f` 에서 `pub(crate)` 로 여는 것)와 `src/ops/card.rs` 의 새 팔이 필요해서, review 중인 그 작업과 같은 파일을 만져요. 새 상태 전이는 원장 상태기계를 넓혀요 | 후속 카드 후보. 그 작업이 착지한 뒤에 다시 읽어요 |
| C. 코드 없이 거절 문구만 | `landed`, `assign` 의 `ready` 거절이 `card land` 를 가리키게 해요 | 거절 문구 두 줄 | 부족해요: 착지가 git 쪽에서 이미 일어난 카드의 기록이 여전히 막혀요 |

## Non-goals

- `card land` 의 판정 검사(`check::judge`: `authors_unknown`, `self_verdict`, `lander_review`, `receipt_missing`)를 바꾸기. 그 검사는 git 을 읽는 착지 관문이고 그대로예요.
- 선택지 B 의 새 상태 전이.
- `ready` 외의 상태(`draft`, `active`, `done`, `hold`)가 `landed`, `assign` 을 받게 하기.
- `card release`, `card reopen`, 예약과 reclaim 코드(리더 release 작업의 범위).
- 판정 기록 쪽(`card verdict`)의 어떤 변경.

## Requirements

- REQ-1: When the leader runs `card landed` on a `ready` card with a head that a verdict of the card names exactly, and the newest verdict of every reviewer on that head is a PASS, the ledger shall record one `card.landed` event that names the head, the by and `from_state` `ready`. · failure: none — 거절 경로는 REQ-2 부터 REQ-5 가 말해요.
- REQ-2: IF a `ready` card has no verdict that names the given head, THEN `card landed` shall refuse with `conflict`, name the missing PASS, and write nothing. · failure: 원장이 그대로예요.
- REQ-3: IF the newest verdict of any reviewer on the given head is a FAIL, THEN `card landed` on a `ready` card shall refuse with `conflict` and write nothing. · failure: 원장이 그대로예요.
- REQ-4: IF `card landed` runs on a `ready` card with a `--by` other than `leader`, THEN it shall refuse with `conflict` and write nothing. · failure: `review` 카드의 `--by` 는 지금처럼 자유 텍스트예요.
- REQ-5: IF a reviewer of the PASS held a reservation of the card in any generation, THEN `card landed` on a `ready` card shall refuse with `conflict` and write nothing. · failure: 감사자가 저자가 아니라는 규칙(`self_verdict` 의 예약 쪽)이에요.
- REQ-6: When the leader runs `card assign` on a `ready` card with an auditor identity that is the reviewer of a PASS verdict of the card, and the identity never held the card, the ledger shall write one `card.audit_assigned` event. · failure: none — 거절 경로는 REQ-7 이 말해요.
- REQ-7: IF a `ready` card has no PASS verdict whose reviewer is the given auditor, THEN `card assign` shall refuse with `conflict` and write nothing. · failure: 지금의 저자 검사와 신원 모양 검사는 그대로예요.
- REQ-8: When `card landed` or `card assign` runs on a card in `review`, the command shall behave as it does today. · failure: none — 이웃 테스트가 그대로 통과해요.
- REQ-9: IF `card landed` or `card assign` runs on a card that is neither `ready` nor `review`, THEN it shall refuse with `conflict` naming `ready` and `review` as the accepted states. · failure: 원장이 그대로예요.
- REQ-10: When a `ready` card has a recorded `card.landed` head inside the integration branch, `card next` shall stop offering it. · failure: none — 이미 있는 `landed_head` 증명이 하는 일이에요.
- REQ-11: When a `ready` card has a recorded `card.landed` head inside the integration branch, `card finish` shall close it on the recorded head. · failure: none — 이미 있는 `landed_head` 증명이 하는 일이에요.

## Acceptance

- AC-1 (REQ-1): PASS 판정이 기록된 `ready` 카드에 `card landed --by leader` 가 성공하고 `card.landed` 이벤트가 head, by, `from_state: ready` 를 담아요 · verify: `cargo test --locked --offline --test card card_ready_pass_landing_cli::ac1_`
- AC-2 (REQ-2, REQ-3): 판정이 없는 `ready` 카드, 다른 head 의 PASS 만 있는 카드, 최신이 FAIL 인 카드는 거절되고 원장이 그대로예요 · verify: `cargo test --locked --offline --test card card_ready_pass_landing_cli::ac2_`
- AC-3 (REQ-4, REQ-5): `--by leader` 가 아니거나 PASS 감사자가 카드의 예약자였으면 거절되고 원장이 그대로예요 · verify: `cargo test --locked --offline --test card card_ready_pass_landing_cli::ac3_`
- AC-4 (REQ-6, REQ-7): PASS 의 감사자를 `ready` 카드에 `card assign` 하면 `card.audit_assigned` 가 써지고 `card show` 의 `roles.auditor` 가 그 신원이에요. PASS 가 없거나 PASS 감사자가 아닌 신원이면 거절돼요 · verify: `cargo test --locked --offline --test card card_ready_pass_landing_cli::ac4_`
- AC-5 (REQ-8, REQ-9): `review` 카드의 `landed` 와 `assign` 은 그대로 성공하고, `draft`, `active`, `done` 카드는 `ready` 와 `review` 를 말하며 거절돼요 · verify: `cargo test --locked --offline --test card card_ready_pass_landing_cli::ac5_`
- AC-6 (REQ-10, REQ-11): `ready` 카드의 착지 기록 뒤 head 가 통합 브랜치 안에 있으면 `card next` 가 그 카드를 건너뛰고 `card finish` 가 기록된 head 로 닫아요 · verify: `cargo test --locked --offline --test card card_ready_pass_landing_cli::ac6_`
- AC-7 (REQ-8): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test card card_roles_cli`
  verify: `cargo test --locked --offline --test card card_done_mflow_cli`
  verify: `cargo test --locked --offline --test card card_land_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-card-ready-pass-landing-path/spec.md`, `src/ledger/workflow.rs`, `src/card/roles.rs`, `src/card/work.rs`, `tests/card/card_ready_pass_landing_cli.rs`, `tests/card/main.rs`
- shared (줄 임대): `src/ops/card.rs` (`Landed` 와 `Assign` 의 도움말 본문 두 문단. 첫 줄 요약은 그대로)
- read-only: `src/card/land.rs`, `src/card/land/check.rs`, `src/card/landed.rs`, `src/card/dispatch.rs`, `src/ledger/reservation.rs`
- forbidden: 진짜 HOME, 설치된 moai 바이너리, 다른 프로젝트, 원장 스키마, 파일 안의 카드 번호, 리더 release 작업의 쓰기 범위 코드

## Plan

1. RED 먼저: AC-1 부터 AC-6 을 임시 HOME 과 임시 리포지터리의 빌드된 바이너리로 돌려요. 기준에서 `only a review card can be landed` 와 `an auditor is assigned to a card in review` 로 실패해요. 거절 케이스(AC-2, AC-3, AC-5 의 `draft`, `active`, `done`)는 처음부터 초록이라, 지킴이 테스트마다 뮤턴트(지운 줄)를 READY 에 적어요.
2. `src/ledger/workflow.rs`: `land` 가 `ready` 를 받되 새 읽기 함수 `pass_on_head`(그 head 에 판정이 있고, 검토자별 최신이 모두 PASS 이고, 검토자가 예약자가 아니에요)가 통과할 때만이에요. 새 거절은 `WorkflowError` 의 기존 `WrongState` 말투 옆에 `conflict` 로 분류돼요. 이벤트 payload 에 `from_state` 를 더해요.
3. `src/card/roles.rs`: `assign` 이 `ready` 를 받되 `named_auditor_passed`(카드의 PASS 판정 중 하나의 검토자가 그 신원이에요)일 때만이에요. 저자 검사와 신원 모양 검사는 앞에 그대로 있어요.
4. `src/card/work.rs`: `landed` 가 `--by leader` 를 `ready` 길에서 요구해요(`by` 값을 `workflow::land` 에 이미 넘겨요).
5. `src/ops/card.rs`: 도움말 두 본문 문단(줄 임대).
6. 지킴이마다 뮤턴트, 이웃 테스트, READY 에서 `check affected --card` 한 번.

## Risks

- `ready` 길의 PASS 읽기는 원장만 읽어요: head 의 저자를 git 트레일러로 가려내는 `authors_unknown` 과 티어 계산은 `card land` 의 일이고 여기서 하지 않아요. 이 길이 하는 일은 기록이라 통합 브랜치를 움직이지 않고, 기록된 head 가 통합 브랜치 안에 있어야 `card next` 가 건너뛰고 `card finish` 가 닫아요(REQ-10). 약해진 점은 없어요: `review` 카드의 `card landed` 는 판정을 아예 읽지 않았고, `ready` 길은 그보다 엄격해요.
- `card assign` 이 `ready` 카드에서 감사자를 못 박으면 이후의 판정 쓰기가 그 신원으로 제한돼요(`verdict_writer_refusal`). REQ-6 이 "PASS 의 검토자 본인"만 받게 해서, 그 카드가 다시 review 로 갔을 때 같은 감사자가 이어서 판정하는 것 말고 다른 길로 감사자를 바꾸지 못해요. 감사자를 바꿔야 하면 지금처럼 `review` 에서 `assign` 해요.
- 리더 release 작업(head `78784f8b5f`)이 `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `src/ops/card.rs` 의 `Release` 팔, `tests/card/main.rs` 를 바꿔요. 이 SPEC 은 `tests/card/main.rs` 의 `mod` 한 줄(다른 카드도 이 파일을 잡고 있어요)과 `src/ops/card.rs` 의 도움말 문단만 겹쳐서, 먼저 착지한 쪽이 다른 쪽에서 합쳐지는 줄 충돌 한 곳으로 끝나요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 선택지 A: `card landed` 와 `card assign` 이 PASS 있는 `ready` 카드를 받아요 | 레인이 추천한 기본값(리더가 위임에 따라 기록해요). 새 상태 전이 없이 카드가 말한 막힘 두 곳만 열고, 리더 release 작업과 겹치지 않아요 | 두 함수의 `ready` 분기를 지워요(`from != "review"` 로 복원) |
| D-2 | `card assign` 의 `ready` 길은 "PASS 의 검토자 본인"만 받아요 | 레인이 추천한 기본값. 나중에 감사자를 PASS 한 사람과 다른 신원으로 못 박는 길을 막아요 | `named_auditor_passed` 를 "PASS 가 하나라도 있으면"으로 느슨하게 하는 것이고, 약해지는 쪽이라 새 결정이 필요해요 |
| D-3 | `card landed` 의 `ready` 길은 `--by leader` 를 요구해요 | 레인이 추천한 기본값. 카드 본문이 "리더 전용"을 바랐어요. `review` 길의 자유 `--by` 는 그대로예요 | `work.rs` 의 한 검사를 지워요 |
| D-4 | `ready` 길의 판정 읽기는 원장만 읽고 git 을 읽지 않아요 | 레인이 추천한 기본값. 착지 관문(`card land` 의 `check::judge`)은 그대로 두고 기록 명령에는 원장의 증거만 요구해요 | 읽기 함수에 git 입력을 더해요(별도 SPEC) |

## Result

구현 head `d50f6412ba193dd302942b78482f2790b6a20730` (기준 NeXT `c3e73024ac`), 선택지 A 와 D-2부터 D-4 의 추천 기본값이에요.

- AC-1…AC-6 met: 각 verify 줄을 적힌 그대로 돌려 모두 exit 0 (ac1…ac6 각 1개 통과). 고치기 전 RED 는 6개 모두 실패했어요: ac1·ac6 은 `only a review card can be landed` 로, ac2·ac3·ac4 는 옛 거절 문장에 PASS·leader·reservation 낱말이 없어서 실패했고, ac5 는 `card finish` 가 head 의 독립 PASS 를 요구해서(`no_verdict`) 픽스처를 고쳐야 했어요(고친 픽스처로 다시 돌린 RED 는 따로 읽지 않았어요).
- AC-7 met: `card_roles_cli` 7, `card_done_mflow_cli` 25, `card_land_cli` 70 통과(모두 exit 0).
- 구현 위치: `src/ledger/workflow.rs` 의 `land`(새 `ready_landing_refusal` 읽기 함수와 새 오류 `LandingRefused`), `src/card/roles.rs` 의 `assign`, `src/ops/card.rs` 의 도움말 두 문단(줄 임대). 계획의 `src/card/work.rs` 는 바꾸지 않았어요: `--by leader` 검사를 `workflow::land` 안에 둬서 모든 호출자가 같은 검사를 받아요. 이벤트 payload 의 `from_state` 는 `ready` 길에서만 붙고 `review` 길의 payload 는 그대로예요(AC-5 가 키가 없음을 읽어요).
- 지킴이 뮤턴트 11개(각각 뒤에 `git checkout` 으로 복원, 끝의 `git status --short` 가 비어 있음): 판정 없음 검사 제거(M1)·FAIL 검사 제거(M2)·`--by leader` 검사 제거(M3)·예약자 검토자 검사 제거(M4)·`from_state` 제거(M5)·다른 상태도 받음(M6)·ready 길이 PASS 없이도 받아짐(M7)이 `workflow.rs` 에서, assign 의 PASS 검사 제거(M8)·assign 이 다른 상태도 받음(M9)·FAIL 도 PASS 로 셈(M10)·검토자 신원 무시(M11)가 `roles.rs` 에서 모두 exit 101 로 죽었어요. 베이스라인(변경 없음)은 exit 0 이에요.
- 위험 경로: 임시 HOME 과 임시 git 리포지터리에서 `review` 카드의 PASS 뒤 `reopen`, `release` 로 `ready` 로 돌린 카드에 `card landed`, `card assign`, `card finish`, `card next` 를 직접 돌렸어요(AC-1…AC-6). 진짜 HOME 이나 설치된 바이너리로 쓰기를 하지 않았어요. 처음 재현(설치된 빌드 46, 읽기만)도 임시 HOME 이었어요.
- 안 한 것: `moai check affected --card` 영수증은 READY 노트에 있어요. 설계 부록 `docs/design-review/appendix/operations/card.md` 의 `card landed`, `card assign` 행은 이 SPEC 의 범위 밖이라 고치지 않았어요(후속).
- 기본 에이전트(`moai-plan`, `moai-impl`, `moai-docs`)는 쓰지 않고 이 세션이 계획·구현·문서를 했어요. 변경이 작았고, 감사는 이 세션이 하지 않아요.
