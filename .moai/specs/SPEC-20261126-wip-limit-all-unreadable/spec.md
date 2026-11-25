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
