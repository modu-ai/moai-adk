---
id: SPEC-20261126-wip-limit-all-unreadable
title: "wip 한도: 레지스트리 항목이 모두 읽히지 않으면 한도를 끄지 않고 take 를 막아요"
status: approved
revision: 1
created: 2026-11-26
approval:
  by: leader
  at: 2026-11-26
  basis: "리더 DELEGATED-APPROVAL 노트 9537 (사용자 위임 2026-11-25 ~14:50Z, 리더 판단이고 사용자 직접 선택은 아님, 노트 9536 에 답함: 개정 1 승인, 커밋 b204a60971c972e606e46354f31f8042e526a42b, 추천 기본값 그대로); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:75a8353883570ea2a5bc4162fb5b857d2cfd7fd0fe71a284c97e92e9fcb959cb"
---

# wip 한도: 레지스트리 항목이 모두 읽히지 않으면 한도를 끄지 않고 take 를 막아요

## Intent

작업 한도(`work_flow.wip_limit`)는 살아 있는 레인 수에 곱해요. 레지스트리 폴더 안의 항목이 하나 이상 있는데 하나도 읽히지 않으면(깨진 JSON, 읽을 수 없는 잠금) 지금은 살아 있는 레인이 0 으로 세어져서 한도가 꺼지고, `card take` 와 `card add` 가 아무 말 없이 통과해요. 레지스트리를 읽지 못하는 상태가 한도를 푸는 쪽으로 틀리는 것이에요. 폴더 전체를 읽지 못할 때(SPEC-20261125-wip-limit-live-lock-lanes REQ-4)는 이미 막는 쪽이라서, 같은 쪽으로 맞춰요. 이것은 관문(tier 2)이라 SPEC 과 승인이 있어야 구현해요.

## Interview

- Q1: 읽을 수 있는 항목이 하나라도 있으면 깨진 항목이 있어도 한도가 켜져 있나요? → A: 예. 그 경우는 지금처럼 이름만 알리고 세지 않아요(Non-goals, REQ-2).
- Q2: 새 필드와 새 거절 문구를 만드나요? → A: 아니요. 기존 `registry_error` 길과 문구를 다시 써요(D-1).
- Q3: 관문을 조이는 변경인데 누가 승인하나요? → A: 사용자 위임에 따른 리더 승인이에요(D-3).

## Non-goals

- 항목이 정말 하나도 없는 레지스트리, 읽을 수 있는 항목이 하나라도 있는 레지스트리의 지금 동작은 바꾸지 않아요.
- 읽을 수 있는 항목이 있을 때 깨진 항목 하나를 이름만 알리고 세지 않는 지금 규칙(REQ-5 of 위 SPEC)은 그대로예요.
- 레지스트리 파일의 복구나 삭제는 하지 않아요.
- `wip_limit` 0 으로 한도를 끄는 길은 그대로 열려 있어요.

## Requirements

- REQ-1: When the session registry holds at least one item and none of its items can be read, the wip limit check shall answer that the limit cannot be checked, naming each item that could not be read, so that `card take` and `card add` refuse and `card next` names no card.
- REQ-2 (REQ-1): IF the registry holds no item, or holds at least one item that can be read (an entry, or a `.live` lock that is held), THEN the wip limit check shall behave as it did before this change.
- REQ-3 (REQ-1): IF `work_flow.wip_limit` is 0, THEN the wip limit check shall stay off whatever the registry holds.

## Acceptance

- AC-1 (REQ-1): a registry whose only item is a malformed entry makes `card take` and `card add` refuse with the registry-cannot-be-read text that names that entry, and `card next` answers `registry_error` as a string, `at_limit` true, `unreadable` with the name and no card · verify: `cargo test --locked --offline --test card card_wip_limit::all_unreadable_ac1_`
- AC-2 (REQ-2): an empty registry folder, a registry of one readable entry whose lease ended beside one malformed entry, and a held lock beside one malformed entry each keep their earlier answer, and the existing wip limit tests are unchanged and pass · verify: `cargo test --locked --offline --test card card_wip_limit::`
- AC-3 (REQ-3): with `wip_limit` 0 a registry whose only item is malformed takes and adds without a refusal · verify: `cargo test --locked --offline --test card card_wip_limit::all_unreadable_ac3_`
- AC-4: light checks · verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `moai spec check`; `moai trailer-check`

## Scope

- write: `src/card/wip_limit.rs`, `tests/card/card_wip_limit.rs`, `.moai/specs/SPEC-20261126-wip-limit-all-unreadable/**`
- read-only: `src/lane_registry.rs`, `src/card/work.rs`, `src/session.rs`
- forbidden: `src/lane_registry.rs` (the scan keeps its shape), the registry files of a real project

## Plan

1. RED: in `tests/card/card_wip_limit.rs` add `all_unreadable_ac1_…` (AC-1), `all_unreadable_ac2_…` (AC-2: empty folder, ended lease plus malformed entry, held lock plus malformed entry) and `all_unreadable_ac3_…` (AC-3); AC-1 fails on this tree because the limit is off. The test binary is `card` (`tests/card/main.rs`), the module `card_wip_limit`.
2. GREEN: in `live_lanes` of `src/card/wip_limit.rs`, after the scan, when `scan.entries` and `scan.locked` are both empty and `scan.unreadable` is not, answer `registry_error: Some(..)` that names every unreadable item (and keep `unreadable` filled); `at_limit`, `refusal`, `card next` and `card/work.rs` already treat `registry_error` as the limit being reached, so they need no change.
3. Guard mutant: drop the all-unreadable condition and see the AC-1 test fail (exit 100).
4. Record the Result with the commands and exit codes.

## Risks

- A leftover broken file in the registry folder now stops every take until it is fixed: the refusal names the item and the way out (`work_flow.wip_limit` 0), so the cause is not hidden.
- The scan does not record a `.live` lock file that nobody holds, so a registry of only released locks beside one broken entry also counts as all unreadable. This is the stricter side and is kept (D-2).
- `src/card/wip_limit.rs` is also in the write scope of another active card (read of the ledger for `card next`); this change stays inside `live_lanes`, so a merge conflict is unlikely and would show at landing.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Reuse `registry_error` for the all-unreadable case instead of a new field and a new refusal text. | One path: `at_limit`, the refusal, `card next` and `card/work.rs` already stop on it, and the wording "the session registry cannot be read" is true here too. | Give the case its own field and text. |
| D-2 | A registry is "all unreadable" when it has no readable entry, no held lock and at least one unreadable item; a released lock file is not an item. | The scan records only these three kinds; the stricter reading fails closed and needs no change to `lane_registry`. | Count released locks as readable items by extending the scan. |
| D-3 | The card is a tier 2 gate change, so the SPEC needs the leader's delegated approval before the code is written. | `CLAUDE.md` Flow step 3. | none |

## Result

구현했어요(2026-11-26, 브랜치 `spec/wip-limit-all-unreadable-20261126`). 승인 front matter 는 `3493b0ba27`, verify 명령 수정은 `96b1bc5bac`, red 시험은 `605eb629d7`, 코드는 `7d46809fdf`(`src/card/wip_limit.rs` 한 파일)예요. 기반은 통합 브랜치 `NeXT` 의 `e946ae8102` 와 합쳐진 세션 레지스트리 세션 키 헤드 `76b6b54dc8`(머지 `b509058658`)이에요. red 시험은 코드를 바꾸기 전에 먼저 돌렸어요. 독립 감사는 이 레인이 하지 않아요.

- 바뀐 것: `live_lanes` 에서 `scan.entries` 와 `scan.locked` 가 모두 비었고 `scan.unreadable` 만 비지 않았을 때 `registry_error: Some("the sessions folder holds N item(s) and none can be read: <names>")` 를 돌려줘요. `live` 는 0 이고 `unreadable` 은 채워 둬요. `Reading::at_limit`, `refusal`, `json`, `read`, `src/card/work.rs`, `src/lane_registry.rs` 는 건드리지 않았어요(D-1, D-2).
- 동작: 이 변경 전에는 모든 항목이 읽히지 않는 레지스트리가 살아 있는 레인 0 으로 세어져서 한도가 꺼지고 `card take` 가 아무 말 없이 통과했어요. 이제는 읽히지 않는 항목의 이름을 밝히며 거절해요. 항목이 없는 레지스트리, 읽을 수 있는 항목이 있는 레지스트리, `wip_limit` 0 은 전과 같아요.
- 시험: `all_unreadable_ac1_…`(깨진 항목 하나, 그다음 둘: `card take` 와 `card add` 가 항목을 이름으로 밝히며 거절하고, `card next` 는 `registry_error` 를 문자열로, `at_limit` 을 true 로, `unreadable` 을 채워서 답하며 다음 카드는 없어요), `all_unreadable_ac2_…`(빈 폴더, 끝난 lease 와 깨진 항목, 살아 있는 레인과 깨진 항목: 각각 전의 답을 지켜요), `all_unreadable_ac3_…`(`wip_limit` 0 은 꺼진 채예요)를 `tests/card/card_wip_limit.rs` 끝에 더했어요. 기존 wip 한도 시험은 하나도 고치지 않았고 모두 통과해요.
- 어긋난 점: SPEC 의 verify 명령이 처음에는 없는 시험 바이너리를 가리켰어요. `96b1bc5bac` 이 `card` 바이너리로 고쳤고, verify 명령은 승인 digest 에 들지 않아서 digest 는 그대로예요.

- AC-1: met — `all_unreadable_ac1_a_registry_with_only_unreadable_items_stops_the_take` 가 red(`605eb629d7`)에서는 `card take` 가 거절 1 대신 exit 0 으로 통과해 실패했고, green(`7d46809fdf`)에서는 통과해요. 가드 뮤턴트 (1) 이 이 시험을 실패시켜요.
- AC-2: met — `all_unreadable_ac2_a_readable_item_or_no_item_keeps_the_earlier_answer` 가 세 경우를 고정하고, 기존 wip 한도 시험이 고치지 않은 채 같은 실행에서 통과해요. AC-2 시험은 old 트리에서도 통과했어요. 바뀌면 안 되는 동작을 고정하는 시험이기 때문이에요.
- AC-3: met — `all_unreadable_ac3_wip_limit_zero_stays_off` 가 통과해요. AC-3 시험도 old 트리에서 통과했어요(위와 같은 이유예요).
- AC-4: met — 가벼운 검사 다섯 개를 직접 돌렸어요: `git diff --check` exit 0, `cargo fmt --all -- --check` exit 0, `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0(`target/c550/clippy.log`), `moai spec check` errors 0·warnings 0, `moai trailer-check b509058658..HEAD`(이 카드의 커밋만) failing 0. 주의: 아직 착지하지 않은 앞선 카드의 합친 커밋을 포함한 `e946ae8102..HEAD` 범위는 그쪽 커밋의 `Moai-Verified` 가 규칙의 시작점이 되어 앞선 네 커밋이 `missing_verified` 로 걸려요. 그 카드가 착지하면 범위가 줄어서 풀려요.

증거(헤드 `7d46809fdf` 의 작업 트리):

- red, 커밋 `605eb629d7`: `cargo nextest run --locked --offline --no-fail-fast --test card -E 'test(/card_wip_limit::/)'` exit 100, 24개 실행, 23 통과·1 실패(`all_unreadable_ac1_a_registry_with_only_unreadable_items_stops_the_take`; `card take` 가 거절 1 대신 exit 0 으로 통과했어요) (`target/c550/red.log`).
- green, 커밋 `7d46809fdf`: `cargo nextest run --locked --offline --no-fail-fast --test card -E 'test(/card_wip_limit::/) | test(/wip_limit/) | test(/card_next/)'` 28개 실행, 28 통과, 0 실패(`target/c550/green.log`).
- clippy: `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0 (`target/c550/clippy.log`).
- 가드 뮤턴트 (1): `live_lanes` 의 `!unreadable.is_empty()` 를 `unreadable.is_empty()` 로 뒤집고 `cargo nextest run --locked --offline --no-fail-fast --test card -E 'test(/card_wip_limit::/)'` 를 돌려 exit 100, 24개 실행, 2개 실패(`all_unreadable_ac1_…`, `all_unreadable_ac2_…`) (`target/c550/mutant1.log`).
- 가드 뮤턴트 (2): `scan.entries.is_empty() &&` 조건을 빼고 같은 명령을 돌려 exit 100, 24개 실행, 2개 실패(`all_unreadable_ac2_…`, `lock_ac5_one_broken_entry_is_left_out_and_listed`) (`target/c550/mutant2.log`). 두 뮤턴트 모두 되돌린 뒤 `git status --short` 는 아무것도 찍지 않았어요.
- 위험한 길의 실행: 임시 HOME 과 프로젝트를 `target/c550/` 아래에 만들고, 헤드 `7d46809fdf` 로 빌드한 실제 `target/debug/moai-cli` 로 `target/c550/dry.sh` 를 돌렸어요(exit 0, 출력은 `target/c550/dry.out`). 등록된 레인이 하나일 때 한도는 2 이고 `card take` 가 통과해요. 레지스트리의 그 항목 하나를 `{not json` 으로 덮어 모든 항목이 읽히지 않게 하자 `card next` 는 `lanes` 0, `at_limit` true, 항목 이름을 밝힌 `registry_error` 를 보이고, `card take` 는 `card conflict in wip_limit: card take refused: the session registry cannot be read, so the work-in-progress limit cannot be checked: the sessions folder holds 1 item(s) and none can be read: <항목 이름>; fix the registry folder, or set work_flow.wip_limit to 0 …` 로 거절돼요. `wip_limit` 0 에서는 같은 `card take` 가 성공해요.

not_run:

- `check affected --card` 영수증은 이 Result 가 든 헤드를 이름 짓는 READY 노트가 담아요.
- 실제 프로젝트 레지스트리에 대한 실행, 다른 운영체제, 독립 감사는 돌리지 않았어요.
