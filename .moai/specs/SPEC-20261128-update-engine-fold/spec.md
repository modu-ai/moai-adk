---
id: SPEC-20261128-update-engine-fold
title: "The extra-skill part of moai update uses the update engine's receipt and known-good code instead of its own copy"
status: draft
revision: 1
created: 2026-11-28
---

# The extra-skill part of moai update uses the update engine's receipt and known-good code instead of its own copy

## Intent

`moai update` 의 추가 스킬 파트(`src/update/extra_skill.rs`, 1134줄)는 스킬 설치 엔진(`src/update.rs`)의 영수증 해석, known-good 복사, 영수증 경로와 저장, 상태 폴더 아래 링크 거절을 줄 단위로 다시 써 두었어요. 두 본문은 폴더 이름(`hosts/<스킬>/claude` 대 `hosts/moai/<호스트>`)과 메시지 세 줄만 달라요. 같은 규칙을 두 곳에서 고쳐야 하니 한쪽만 고치면 두 파트가 어긋나요(2026-11-25 사용자 지시: `moai update` 전체 로직을 분석해 개선).

이 SPEC 은 그 복사를 한 본문으로 접어요. 동작은 한 바이트도 바꾸지 않아요: 메시지 글, 영수증과 known-good 경로, 확인 순서, 공개 함수의 서명이 그대로예요. 바꾸는 코드가 `src/update` 아래이고 홈 상태 폴더에 쓰는 엔진이라 tier 2 바닥이어서(`src/gate.rs` 의 `TIER2_PREFIXES`) 사용자가 승인한 SPEC 이 먼저예요(리더 결정 노트 11061).

사실(2026-11-28, NeXT `af867cf151`, 읽은 사람은 `moai-plan` 에이전트이고 레인이 `resolve_receipt` ↔ `resolve_host_receipt` 한 쌍은 두 본문을 통째로 다시 읽어 확인했어요; 나머지 쌍은 구현 때 다시 읽어요):

- `extra_skill.rs` `resolve_receipt`(575-719)와 `update.rs` `resolve_host_receipt`(1310-1468)는 같은 논리예요. 다른 곳은 (a) 영수증 경로 함수, (b) `r.host == HOST` 대 `r.host == host`, (c) 읽을 수 없을 때 글의 "the {skill} receipt" 대 "the host receipt", (d) 다른 설치를 가리킬 때 글의 같은 차이, (e) 이른 설치의 pending 영수증 글(`pending_reason(..)` 대 고정 글)이에요.
- `known_good_copies`(extra 723-765, update 812-889), `receipt_path`/`save_receipt`(extra 550-569, update 729-795), `refuse_links_below_state`(extra 524-546)와 같은 링크 거절 반복문이 `update.rs` 에 두 번 인라인으로 있어요(합쳐 세 곳).
- `preflight`(extra 393-470, update 1106-1175)는 같은 뼈대지만 메시지 다섯 곳과 검사 둘이 달라서 일부러 다른 코드예요. `apply_part`(extra 855-1067)와 `update.rs` 적용 반복문은 되돌리기와 pending 영수증 처리가 달라 일부러 달라요. `layer_summary`/`plan_summary`(`update.rs`)는 `preview_rows`(`update_tux.rs`)와 겹치지 않아요: 사람이 아닌 출력은 요약 글자열이, 사람 출력은 JSON 으로 그리는 `preview_rows` 가 쓰여서 지우면 질문 글이 바뀌어요.
- 이 일의 카드 글에 적힌 "약 −900줄"은 동작을 바꾸지 않는 접기로는 닿지 않아요. 에이전트의 추정은 순 −270줄쯤이고 이 SPEC 은 줄 수를 목표로 삼지 않고 재서 적기만 해요.

## Non-goals

- `apply_part`, `preflight`, `layer_summary`, `plan_summary`, `src/update_tux.rs` 는 바꾸지 않아요.
- 메시지 글, 영수증·known-good 경로, JSON 모양, 종료 코드를 바꾸지 않아요.
- 줄 수 목표는 없어요. 재서 적어요.
- `src/update.rs` 는 다른 카드(검사 호출을 넣는 중)가 쓰는 동안 건드리지 않아요: 이 SPEC 의 구현은 그 카드가 랜딩한 뒤에 시작해요.

## Requirements

- REQ-1: The extra-skill part shall resolve the receipt of its skill by calling the update engine's generic receipt resolver with the receipt folder, the label and the pending-earlier text of the skill.
- REQ-2: The update engine's generic receipt resolver shall return the same receipt, source folder, or blocked text as `resolve_host_receipt` does for the host receipts, for every input.
- REQ-3: The extra-skill part shall produce the same blocked text as before for an unreadable receipt, a receipt that names another installation or host, a pending earlier receipt and two folders finishing in the same second.
- REQ-4: The extra-skill part shall copy the files it overwrites to the same known-good folder, with the same bytes and the same failure text, by calling the update engine's generic known-good copy.
- REQ-5: The refusal of a symlink between the state root and a receipt or known-good folder shall be one function that every caller in `src/update.rs` and `src/update/extra_skill.rs` calls.
- REQ-6: The extra-skill part shall derive its file plan (three-way compare, disk-only files, unmanaged paths) from the update engine's file-plan function, in the same order as before.
- REQ-7: The public items of the extra-skill part and the functions that code outside the two files calls shall keep their names and signatures.
- REQ-8: The tests that exist before the change shall pass without any edit, apart from the three rows of the write-scan table (`tests/home_write_scan.rs`) that name the moved functions.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): The pins of the receipt messages and of the version folders pass unchanged, and the pins added in slice 0 pass before and after · verify: `cargo nextest run --locked --offline --test update_version_folders --test update_moved_receipts --test update_user_layer_cli`
- AC-2 (REQ-4): The known-good pins of the update engine and of the extra-skill part pass before and after · verify: `cargo nextest run --locked --offline --test update_known_good --test update_user_layer_cli`
- AC-3 (REQ-5, REQ-6): The remaining update targets pass unchanged · verify: `cargo nextest run --locked --offline --test update_run_receipt_cli --test host_home_cli --test home_write_scan --test path_root_direct_sites --test codex_home_update_removed`
- AC-4 (REQ-7): The crate builds with every test target and the public items keep their signatures · verify: `cargo check --locked --offline --lib --bins --tests`
- AC-5 (REQ-8): `git diff --numstat <base>..<head> -- tests/` shows no deleted line in a test that exists before, except the rows of the write-scan table that name the moved functions; the READY note lists each such row · verify: `git diff --numstat af867cf1515382cf98e3091f4bfca3b411f12626..HEAD -- tests/`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8): The READY note carries the line counts of `src/update.rs`, `src/update/extra_skill.rs` and `src/update_tux.rs` before (4709, 1134, 2840 at `af867cf151`) and after, and the card-set check of the head passes · verify: `moai check affected --card af867cf1515382cf98e3091f4bfca3b411f12626..HEAD`

## Scope

- write: `src/update/extra_skill.rs`
- write: `src/update.rs`
- write: `tests/home_write_scan.rs` (rows only, by line lease)
- write: `tests/update_user_layer_cli.rs` (slice 0, tests added only)
- write: `.moai/specs/SPEC-20261128-update-engine-fold/**`
- read-only: `src/update_tux.rs`, `src/main.rs`, `src/gate.rs`, `tests/update_version_folders.rs`, `tests/update_known_good.rs`, `tests/update_moved_receipts.rs`
- forbidden: `src/update/switch.rs` and `src/update_models.rs` (another card writes them now), `src/update_tux.rs`, any generated skill file

## Plan

1. Slice 0 (no product code, no approval needed): add the pins that are missing for the extra-skill part to `tests/update_user_layer_cli.rs`: the known-good copy with its old bytes and its failure text, the unreadable earlier receipt text and the "names another installation or host" text. Record the line counts before.
2. Wait for the card that writes `src/update.rs` and `tests/home_write_scan.rs` to land, and for the user's approval of this SPEC. The leader asks the user; the lane records no approval.
3. Slice B1: in `src/update.rs` split `host_receipt_path`, `known_good_copies`, `resolve_host_receipt` and `save_host_receipt` into `_in` functions that take the folder, the label and the pending text, keep the old names as wrappers, move `refuse_links_below_state` in and call it from the two inline loops; in `src/update/extra_skill.rs` delete the copies and keep thin wrappers of the same names; fix the three rows of the write-scan table. Expected change from the plan agent's reading: `extra_skill.rs` about −210, `update.rs` about +28 (estimate, not measured).
4. Slice B2: extract the file plan (three-way compare, disk-only files, unmanaged paths) into one function in `src/update.rs` that both parts call. Expected about −85 net (estimate).
5. Each slice: `cargo fmt --all`, the card-set check bare at the head, `moai bundle check` (floor 2 is expected, so the review is the full audit), one guard mutant per new pin from slice 0.

## Risks

- The tier-2 floor paths: a fold that changes one message byte or one path breaks `moai update` for a user; the pins of both sides and the full audit are the guard. A message that no test pins is a gap: slice 0 closes the three that the plan agent found, and an unpinned one found later is pinned before it moves.
- `tests/home_write_scan.rs` and `src/update.rs` are written by another card now; starting early would conflict. The slice waits for that landing.
- The `// @MX:DEBT` tags above `known_good_copies` move with the generic body.
- `resolve_host_receipt`, `host_receipt_path`, `load_host_receipt` and `known_good_copies` have callers outside the two files (`recovery.rs`, `update_cleanup.rs`); the wrappers keep their signatures.
- A fold that merges `receipt_path` directory chains changes the `home_root::path` counts of the write-scan rows; this SPEC fixes those rows in the same commit.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `layer_summary`, `plan_summary` and `update_tux.rs` stay unchanged | The two are not duplicates of `preview_rows` (read in `main.rs`: the plain summary is the question text for a non-human run); lane proposal from the plan agent's reading, confirmed by the lane for this pair | Open a new card |
| D-2 | `preflight` is not folded | Five message sites and two checks differ on purpose; a fold needs per-message pins for both skills first and saves about 35 lines | Open a new card after the pins exist |
| D-3 | `apply_part` is not folded | The extra part undoes its own writes and records `recovered`; the engine keeps a `pending` receipt and returns `Err` | Open a new card |
| D-4 | The card's "about −900 lines" is replaced by measured line counts | The folds that keep behaviour reach about −270 by estimate; leader decision note 11061 | None needed |
| D-5 | Slice B waits for the landing of the card that writes `src/update.rs` and for the user's approval | `src/update.rs` and `tests/home_write_scan.rs` are written by that card; `src/update` is a tier-2 floor path (leader decision note 11061) | Record a leader decision that lifts the wait (the user still approves the SPEC) |

## Result

조각 0(승인 전, 제품 코드 변경 없음)의 구현 커밋은 `b6fb694572`(브랜치 `refactor/update-fold-duplicates-20261128`, 기준 NeXT `af867cf151`)이에요. 이 SPEC 은 아직 승인 전(draft)이고 조각 B 는 시작하지 않았어요. 아래 AC 는 조각 B 가 끝나야 판정할 수 있어서 모두 not_run 이에요.

- 조각 0: `tests/update_user_layer_cli.rs` 에 특성 테스트 넷(`extra_skill_update_keeps_the_replaced_bytes_as_known_good_copies`, `extra_skill_known_good_copy_failure_stops_the_part_before_it_starts`, `extra_skill_unreadable_earlier_receipt_blocks_the_part_and_says_nothing_was_adopted`, `extra_skill_receipt_naming_another_installation_or_host_blocks_the_part`)를 더했어요. 현재 코드에서 4개 모두 통과하고(`-E 'test(extra_skill_)'` → 4 passed) 파일 전체는 29개가 통과했어요(구현 에이전트 보고). 기존 테스트 줄은 하나도 지우거나 바꾸지 않았어요(`git diff --numstat` 175 추가 0 삭제 + 형식 정리 뒤 176 추가).
- 뮤턴트(테스트마다 되돌린 뒤 `git status --short` 가 비었어요): known-good 폴더 경로에서 `.join(skill)` 을 뺌 → 앞 두 테스트가 실패, "읽을 수 없음" 글의 끝을 바꿈 → 셋째 테스트가 실패, "다른 설치나 호스트" 글의 끝을 바꿈 → 넷째 테스트가 실패(각각 exit 100).
- 줄 수 전(기준 `af867cf151`, `wc -l`): `src/update.rs` 4709, `src/update/extra_skill.rs` 1134, `src/update_tux.rs` 2840. 조각 0 은 이 세 파일을 바꾸지 않았어요.
- AC-1: not_run (조각 B 전). AC-2: not_run. AC-3: not_run. AC-4: not_run. AC-5: not_run. AC-6: not_run.
- 건너뛴 것: 조각 B1·B2(승인과, `src/update.rs` 를 쓰는 다른 카드의 조각 2 랜딩 뒤), 전체 `cargo test --locked`(리더 몫), Windows·Linux.
