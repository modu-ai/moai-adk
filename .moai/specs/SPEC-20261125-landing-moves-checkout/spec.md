---
id: SPEC-20261125-landing-moves-checkout
title: "card land 가 체크아웃된 깨끗한 통합 브랜치를 작업 트리와 함께 옮기고, 착지하면 카드가 닫히고, 작업 수에 상한이 생겨요"
status: draft
revision: 1
created: 2026-11-25
branch: spec/landing-moves-checkout-20261125
---

# card land 가 체크아웃된 깨끗한 통합 브랜치를 작업 트리와 함께 옮기고, 착지하면 카드가 닫히고, 작업 수에 상한이 생겨요

## Intent

카드 완료 흐름 재설계의 두 번째 조각이에요(SPEC-20261124-leader-brief). 승인 근거는 사용자 결정 2026-11-24T21:52:39Z(리더 대화, 선택지로 받은 답) 3·4번이라 승인을 기다리지 않아요. 이 SPEC 도 승인을 기록하지 않고 READY 에 digest 를 적어요.

- 착지 경로: 통합 브랜치(`NeXT`)가 주 체크아웃에 체크아웃돼 있으면 `card land` 가 거절해서(`src/card/land.rs`, "a move under it would leave that worktree behind the reference") 리더가 손으로 옮겨 왔어요(133번). 깨끗하고 HEAD 가 옛 tip 이면 브랜치와 작업 트리를 함께 앞으로 옮겨요.
- 착지하면 닫힘: 읽어 보니 `--keep-open` 이 없는 착지는 이미 카드를 닫아요. 열린 기준은 PASS 판정의 증거로 `met` 이 되고(`SPEC-20261124-land-closes-card`), 그 규칙은 그대로 둬요. 배차 노트의 "남은 기준을 withdrawn" 은 따르지 않아요: PASS 가 덮은 기준을 withdrawn 으로 적으면 감사가 덮은 사실이 사라지고, 착지 때 닫히지 않고 남는 기준은 없어요(D-2). 이 SPEC 은 그 사실을 `card land` 와 `card land-batch` 양쪽 테스트로 고정해요.
- 작업 수 상한: 작업 중(`active`)과 검사 대기(`review`) 카드가 많으면 새 카드를 잡는 대신 검사와 착지를 도와야 해요.

## Non-goals

- `card drift --restore` 의 체크아웃 거절: 되돌림은 그대로 거절해요.
- 열린 채로 착지된 기존 카드를 닫기: 리더가 `card finish` 로 처리해요.
- 레인 수를 세는 새 레지스트리: 기존 세션 레지스트리를 읽기만 해요.

## Requirements

- REQ-1: When the integration branch is checked out in exactly one worktree whose tracked files are unchanged and whose HEAD is the tip that `card land` read, `card land` shall move the branch and that worktree to the head together. · failure: none — 옮긴 뒤 트리의 HEAD 와 파일이 head 와 같아요.
- REQ-2: IF the worktree that has the integration branch checked out holds a change to a tracked file, THEN `card land` shall refuse, name that worktree and write nothing. · failure: none — 원장과 브랜치와 트리가 그대로예요.
- REQ-3: IF the integration branch is checked out in more than one worktree or the HEAD of that worktree is not the tip that `card land` read, THEN `card land` shall refuse and write nothing. · failure: none — 이유를 말하고 아무것도 쓰지 않아요.
- REQ-4: While the worktree that has the integration branch checked out holds only ignored files, `card land` shall land. · failure: none — 무시된 파일은 막지 않아요.
- REQ-5: When `card land-batch` lands a head, it shall use the same move, so that a checked-out integration branch is moved with its worktree. · failure: none — 테스트가 묶음 착지에서 트리를 읽어요.
- REQ-6: When `card land` or `card land-batch` lands a head without `--keep-open`, the card shall be `done` with no open criterion. · failure: none — 테스트가 착지 뒤 카드 상태를 읽어요.
- REQ-7: IF the cards in state `active` or `review` number at least the limit, THEN `card take` shall refuse and name the cards that wait for a check or a landing. · failure: none — 원장이 그대로예요. 한도는 살아 있는 등록된 레인 수(최소 1)에 `work_flow.wip_limit`(기본 2)를 곱한 수예요.
- REQ-8: IF the cards in state `active` or `review` number at least the limit, THEN `card add` shall refuse unless it is given `--by user`. · failure: none — 사용자의 add 는 막지 않아요.
- REQ-9: When `work_flow.wip_limit` is `0`, the limit shall be off. · failure: none — 같은 키의 다른 숫자는 곱이에요.
- REQ-10: When `card next` finds the limit reached, it shall name the cards in `review` that carry no audit verdict and that this session did not write, in place of a card to take. · failure: none — 감사할 카드가 없으면 그 사실을 말해요.

## Acceptance

- AC-1 (REQ-1): 통합 브랜치가 주 체크아웃에 체크아웃된 임시 저장소에서 깨끗한 트리의 `card land` 가 브랜치와 트리를 head 로 옮기고, 트리의 파일이 head 의 파일이에요 · verify: `cargo test --locked --offline --test card card_land_cli::checkout_a_clean_checked_out_branch_moves_with_its_tree`
- AC-2 (REQ-2, REQ-3): 추적 파일이 바뀐 트리, HEAD 가 tip 이 아닌 트리, 트리 둘이 같은 브랜치를 가진 경우 모두 거절하고 원장·브랜치·트리가 그대로예요 · verify: `cargo test --locked --offline --test card card_land_cli::checkout_a_dirty_or_moved_or_shared_checkout_refuses_and_writes_nothing`
- AC-3 (REQ-4): 무시된 파일이 있는 깨끗한 트리는 착지가 되고 무시된 파일이 그대로예요 · verify: `cargo test --locked --offline --test card card_land_cli::checkout_an_ignored_file_does_not_stop_the_landing`
- AC-4 (REQ-5): 묶음 착지가 체크아웃된 브랜치를 트리와 함께 옮겨요 · verify: `cargo test --locked --offline --test card card_land_batch_cli::checkout_a_batch_moves_the_checked_out_tree`
- AC-5 (REQ-6): `--keep-open` 없는 착지 뒤 카드가 `done` 이고 열린 기준이 없어요(`land` 와 `land-batch`) · verify: `cargo test --locked --offline --test card card_land_cli::checkout_a_landing_closes_the_card`
- AC-6 (REQ-7, REQ-8, REQ-9): 한도에서 `card take` 와 `card add` 가 거절하고 대기 카드를 말하며, `--by user` 의 add 는 통과하고, `wip_limit` 0 은 한도를 꺼요 · verify: `cargo test --locked --offline --test card card_wip_limit`
- AC-7 (REQ-10): 한도에서 `card next` 가 감사가 필요한 카드를 말해요 · verify: `cargo test --locked --offline --test card card_wip_limit::next_`
- AC-8 (REQ-1…REQ-10): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test card card_land_cli`
  verify: `cargo test --locked --offline --test card card_land_batch_cli`
  verify: `cargo test --locked --offline --lib profile`

## Scope

- write: `.moai/specs/SPEC-20261125-landing-moves-checkout/spec.md`, `src/card/land.rs`, `src/card/landed.rs`, `src/card/dispatch.rs`, `src/card/command.rs`, `src/card/work.rs`, `src/card/wip_limit.rs`, `src/card/mod.rs`, `tests/card/card_land_cli.rs`, `tests/card/card_land_batch_cli.rs`, `tests/card/card_wip_limit.rs`, `tests/card/main.rs`
- shared (줄 임대): `src/profile.rs` (`wip_limit` 키), `src/ops/card.rs` (take 와 add 가 한도 검사에 필요한 값을 넘기는 줄이 있으면)
- read-only: `src/lane_registry.rs`, `src/card/drift.rs`
- forbidden: 진짜 HOME, 설치된 moai 바이너리, 다른 프로젝트, 파일 안의 카드 번호

## Plan

1. RED 먼저: AC-1~AC-3 의 테스트(임시 저장소, 체크아웃된 `integ`)와 AC-6·AC-7 의 한도 테스트를 돌려 맞는 이유로 실패하는 것을 봐요. AC-4·AC-5 는 이미 닫히는 동작이면 고정만 해요.
2. `src/card/landed.rs`: `move_checked_out` 이 트리의 상태를 읽고(`git status --porcelain --untracked-files=no`, HEAD) 깨끗할 때 그 트리에서 `git merge --ff-only <head>` 로 브랜치와 트리를 함께 옮겨요. 아닐 때는 이유를 돌려줘요.
3. `src/card/land.rs`: 체크아웃 거절을 위 검사로 바꾸고(원장을 쓰기 전), 옮기는 단계에서 체크아웃된 경우 `move_checked_out` 을 써요.
4. `src/card/wip_limit.rs`: 한도와 대기 카드 목록, `take`·`add` 가 부르고 `next` 가 감사 후보를 말해요. 프로필 키 `work_flow.wip_limit`.
5. 지킴이마다 뮤턴트(최대 3줄), 이웃 테스트, READY 에서 `check affected --card`.

## Risks

- 트리를 옮기는 `git merge --ff-only` 는 훅(post-merge 등)을 돌릴 수 있어요. 훅의 실패는 이동을 되돌리지 않고 답에 남아요.
- 추적되지 않았고 무시되지도 않은 파일이 head 의 새 파일과 겹치면 git 이 이동을 거절해요. 원장 시작 행은 이미 있고 브랜치는 그대로예요(착지 시작 행이 남는 것은 기존 `after_start_row` 경로와 같아요).
- 한도는 등록된 레인 수에 기대요. 등록이 없으면 레인 수는 1 이라 한도는 2 예요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 체크아웃된 트리는 그 트리에서 `git merge --ff-only` 로 브랜치와 함께 옮겨요 | `update-ref` 로 브랜치만 옮기면 트리가 뒤처져요(기존 거절의 이유). git 이 ref 와 트리를 한 명령으로 옮기고 fast-forward 가 아니면 거절해요 | 거절로 되돌려요 |
| D-2 | 착지 때 열린 기준을 withdrawn 으로 적지 않고 기존 규칙(PASS 증거로 met)을 그대로 둬요 | `src/card/land.rs` 가 `--keep-open` 이 없는 착지의 열린 기준을 PASS 판정 증거로 채우고 카드를 닫아요. withdrawn 은 감사가 덮은 사실을 지워요. 닫히지 않고 남는 기준이 착지 때는 없어요 | 채움을 withdrawn 으로 바꿔요 |
| D-3 | 한도 = `wip_limit`(기본 2) × 살아 있는 레인 수(최소 1), `wip_limit` 0 은 끔 | 사용자 결정 4번(작업 중과 검사 대기를 합쳐 레인 수의 두 배)이에요 | 키를 0 으로 해요 |

## Result

(구현 뒤에 적어요.)
