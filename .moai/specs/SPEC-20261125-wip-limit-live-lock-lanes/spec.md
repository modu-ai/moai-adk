---
id: SPEC-20261125-wip-limit-live-lock-lanes
title: "The work-in-progress limit counts a lane with a held live lock, and a registry that cannot be read no longer turns it off"
status: draft
revision: 1
created: 2026-11-25
---

# The work-in-progress limit counts a lane with a held live lock, and a registry that cannot be read no longer turns it off

## Interview

- Q1: 상한의 "살아 있는 레인 수"는 무엇으로 세나요? 레인의 등록 임대(4시간)가 끝나면 세션은 멀쩡히 일하는데도 수에서 빠져 take 가 막혀요. → A: (추천) 세션 실행기가 쥔 `<이름>.live` 잠금이 잡혀 있거나 등록 임대가 남아 있으면 살아 있는 레인이에요(`session-register --lane-check` 와 같은 기준).
- Q2: 레지스트리를 못 읽으면 상한은 어떻게 하나요? 지금은 조용히 꺼져요(`unwrap_or(0)`). → A: (추천) 폴더를 통째로 못 읽으면 막고 이유를 말하고, 항목 하나만 못 읽으면 그 이름만 빼고 목록에 적어요. 레지스트리 폴더가 아예 없으면(등록한 적 없음) 지금처럼 꺼져 있어요.
- Q3: 상한 숫자(레인 수 × 2)를 바꾸나요? → A: 아니요. 세는 대상만 바뀌고 `wip_limit` 배수와 "0 이면 끔"은 그대로예요.

## Intent

출처: mo.ai.kr 리더 덧붙임 8 M 실측(2026-11-25 23:45~11-26 00:50 KST, 이 SPEC 의 카드 본문). 살아 있는 레인 셋(`session-register --lane-check` 가 셋 다 `lane`, basis `live_lock`)인데 `card take` 가 `the work-in-progress limit is reached: 3 cards are in work or in review and the limit is 2 (2 per lane, 1 lane(s))` 로 거절됐어요. 두 레인의 등록 임대가 끝나 `session-register --show` 가 `free` 였고, 레인이 다시 등록하자 바로 take 가 됐어요.

읽은 코드(`src/card/wip_limit.rs`, 2026-11-25): `live_lanes` 는 `lane_registry::session_entries` 의 항목 중 `lease_until > now` 이고 `mode` 가 `leader` 가 아닌 것을 세요. 세션 실행기가 쥔 `.live` 잠금은 보지 않아요. 같은 함수가 `session_entries` 의 오류(항목 하나가 깨졌거나, 폴더가 링크이거나, 읽기 권한이 없을 때)를 `unwrap_or(0)` 으로 받아 레인 0, 곧 상한 꺼짐으로 바꿔요. 깨진 항목 하나가 상한을 소리 없이 끄는 길이에요.

이 SPEC 이 하는 일: (1) 살아 있는 레인을 `lane-check` 와 같은 기준으로 세요(잠금이 잡혔거나 임대가 남은 이름, 리더 제외, 이름마다 한 번). (2) 읽을 수 없는 레지스트리를 조용히 꺼짐으로 두지 않아요.

## Non-goals

- 상한 배수(`work_flow.wip_limit`), "0 이면 끔", "살아 있는 레인이 없으면 끔"(앞선 SPEC 의 D-3)을 바꾸지 않아요.
- 카드 쪽 세기(임대가 살아 있는 `active`·`review` 카드만 세요, SPEC-20261125-wip-limit-live-leases)와 PASS 가 난 review 카드의 처리(별도 카드)는 건드리지 않아요.
- 등록 항목의 열쇠, 레인 등록 단계, `-a` 이름 고르기(SPEC-20261125-session-registry-session-key)는 바꾸지 않아요.
- `card take`·`card review` 가 레인의 등록 임대를 갱신하게 하지 않아요(D-1 의 B).
- 레지스트리 항목이나 잠금 파일을 만들거나 지우지 않아요. 읽기만 해요.

## Requirements

<!-- This SPEC changes which sessions the work-in-progress gate counts and what the gate does when the registry cannot be read; a session whose launcher lock or registration lease is alive is counted exactly as a lane is today or more, and an unreadable registry no longer switches the gate off, so the gate is not weakened. -->
- REQ-1: The work-in-progress limit shall count as a live lane every session name whose `<name>.live` lock is held or whose registry entry has a lease that has not ended, each name once. · failure: none — a name with neither proof is no live lane.
- REQ-2: The work-in-progress limit shall leave out of the live lanes every name whose registry entry has the mode `leader` and the project's leader name `<project name>-leader`, whatever the state of its lock or lease. · failure: none — a leader never raises the limit.
- REQ-3: The work-in-progress limit shall count a name whose lock is held and that has no readable registry entry only when it starts with `agent-` or `<project name>-agent-`. · failure: none — a lock alone does not say what the session is, so an unknown name stays out.
- REQ-4: IF the registry folder exists and cannot be read as a whole (a link, no permission, a non-folder), THEN `card take` and `card add` shall refuse and name the reason. · failure: the refusal changes nothing in the ledger; `work_flow.wip_limit` 0 still turns the limit off.
- REQ-5: IF one registry entry or lock cannot be read, THEN the work-in-progress limit shall leave that name out of the live lanes and list it as unreadable in a refusal. · failure: none — one broken file never blocks the rest and never turns the limit off.
- REQ-6: The `wip` object of `card next` shall carry `registry_error` (the reason of REQ-4 or null) and `unreadable` (the names of REQ-5). · failure: none — `card next` reads only, and an old reader that ignores the new keys keeps the old ones.
- REQ-7: IF `registry_error` is not null, THEN `card next` shall answer that the limit is reached and name no card. · failure: none — `card take` refuses in the same case, so the two answers agree.
- REQ-8: The limit shall stay off when `work_flow.wip_limit` is 0, when the registry folder does not exist, or when no live lane is found. · failure: none — the same off-switches as before.

## Acceptance

- AC-1 (REQ-1): in a temporary project whose three lane entries have ended leases while their `.live` locks are held, `card take` succeeds up to six cards in work and the seventh refuses with `3 lane(s)` · verify: `cargo test --locked --offline --test card card_wip_limit::lock_ac1_`
- AC-2 (REQ-1): a lane with no lock and an ended lease stays out, and a lock that is let go takes its lane out again (guard against counting too much) · verify: `cargo test --locked --offline --test card card_wip_limit::lock_ac2_`
- AC-3 (REQ-2, REQ-3): a leader entry with a held lock and an ended lease, a held lock named `<project name>-leader` with no entry, and a held lock with an unknown name and no entry all stay out of the lane count · verify: `cargo test --locked --offline --test card card_wip_limit::lock_ac3_`
- AC-4 (REQ-4, REQ-6, REQ-7): with the sessions folder replaced by a link, `card take` and `card add` refuse naming the registry, `card next` answers `wip.registry_error` and `at_limit` true, and `wip_limit` 0 takes · verify: `cargo test --locked --offline --test card card_wip_limit::lock_ac4_`
- AC-5 (REQ-5, REQ-6): one malformed entry beside two live lanes leaves the lane count at two and is listed as unreadable in the refusal and in `card next` · verify: `cargo test --locked --offline --test card card_wip_limit::lock_ac5_`
- AC-6 (REQ-8): `wip_limit` 0, a project with no registry folder and a registry with no live lane stay off · verify: `cargo test --locked --offline --test card card_wip_limit::lock_ac6_`

## Decisions

각 결정은 되돌릴 수 있어요. 추천이 맨 앞(기본)이에요. 이 SPEC 은 게이트(작업 수 상한)를 다루는 tier 2 라서 사용자 승인이 필요하고, 리더가 위임에 따라 추천안으로 기록해요.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | (A, 기본·권장) 살아 있는 레인은 `.live` 잠금이 잡혔거나 등록 임대가 남은 이름이에요 — `session-register --lane-check` 와 같은 기준이라 리더가 메시지를 보내기 전에 보는 답과 상한이 어긋나지 않아요. 읽기 전용이에요. (B) `card take`·`card review` 가 그 레인의 등록 임대를 갱신하게 해요: 카드 경로가 레지스트리에 쓰게 되고, 갱신이 세션이 살아 있다는 증거가 되지 못해서 고르지 않아요. (C) 레저의 보유자(holder)로 세요: 카드가 없는 레인이 안 세어져 상한이 거꾸로 막혀요. | 카드 본문의 바라는 동작 첫째("살아 있는 레인(live lock 이 잡힌 이름)으로 세기")와 실측 | `live_lanes` 를 임대 기준으로 되돌려요 |
| D-2 | (A, 기본·권장) 항목이 없고 잠금만 잡힌 이름은 `agent-` 나 `<프로젝트>-agent-` 로 시작할 때만 세요. 잠금은 세션이 무엇인지 말해 주지 않아서, 이름 모양을 모르는 잠금(`lane-0`, 수동 이름)이 상한을 올리지 않게 해요. (B) 잠금만 있는 이름은 모두 세요: 리더·보조 세션이 레인으로 세어져 상한이 느슨해져요. (C) 잠금만 있는 이름은 세지 않아요: 항목이 지워졌거나 만료 정리된 레인이 다시 빠져 실측 문제가 남아요. | 상한을 느슨하게 하지 않는 쪽, `-a` 의 이름 규칙(`session.rs` `pick_agent_name`) | 이름 규칙을 지워요 |
| D-3 | (A, 기본·권장) 레지스트리 폴더를 통째로 못 읽으면(링크, 권한 없음, 폴더 아님) `card take`·`card add` 가 이유를 말하고 막아요. 항목·잠금 하나만 못 읽으면 그 이름만 빼고 `unreadable` 로 적어요. 폴더가 없으면 지금처럼 꺼져 있어요. (B) 조용히 꺼요(지금): 깨진 파일 하나나 링크 하나가 게이트를 끄는 길이라 고르지 않아요. (C) 못 읽으면 끄되 답에 경고만 붙여요: 경고는 읽는 사람이 없으면 게이트가 없는 것과 같아요. (D) 깨진 항목 하나에도 전체를 막아요: 오래된 깨진 파일 하나가 모든 take 를 막아 이 카드가 고치려는 문제를 새로 만들어요. | 카드 기준 2("조용히 상한을 끄지 않음"), CLAUDE.md 의 "게이트는 약하게 하지 않는다" | `registry_error` 줄을 지우면 지금 동작으로 돌아가요 |
| D-4 | (A, 기본·권장) 폴더를 못 읽을 때 `card next` 는 읽기만 하고 카드를 고르지 않아요: `wip.registry_error` 를 채우고 `at_limit` 을 참으로 해서 `next_is` 가 `limit` 이에요. (B) 카드를 계속 골라 줘요: take 가 어차피 막혀서 레인이 헛걸음해요. | `card take` 와 `card next` 가 같은 답을 해야 해요 | `at_limit` 규칙을 지워요 |
| D-5 | (A, 기본·권장) 이름마다 한 번 세는 규칙은 이 SPEC 의 이름 기준이고, 같은 세션이 두 이름으로 등록된 중복(세션 id 열쇠)은 SPEC-20261125-session-registry-session-key 가 착지하면 그 위에서 합쳐요. (B) 이 카드에서 세션 id 중복까지 구현: 그 SPEC 과 같은 줄을 두 번 고쳐 충돌해요. | SPEC-20261125-session-registry-session-key 가 같은 함수(`live_lanes`)를 고치고 있어요(커밋 6c90ae5690) | 없음(구현 순서만의 결정) |

## Scope

- write: `.moai/specs/SPEC-20261125-wip-limit-live-lock-lanes/**`, `src/card/wip_limit.rs`, `src/lane_registry.rs`(새 읽기 함수 하나만 더하고 기존 함수는 고치지 않아요), `tests/card/card_wip_limit.rs`(끝에만 덧붙여요)
- read-only: `src/session.rs`(`live_lock_held`, `lock_named_live`, `leader_session_name`, `agent_prefix`), `src/card/work.rs`, `src/card/dispatch.rs`, `src/card/command.rs`, `src/ledger/schema.rs`, `docs/decisions-2026-11-08.md`(11:27Z 줄)
- forbidden: 상한 배수·끄는 규칙을 느슨하게 하는 변경, `tools/repo_scans.json` 증가(새 시험 파일 없음), 레지스트리 항목·잠금 파일을 만들거나 지우는 코드, 전역 설치본과 실제 HOME, 네트워크, Python

## Plan

1. RED: `tests/card/card_wip_limit.rs` 끝에 `lock_ac1_` … `lock_ac6_` 를 덧붙여요. 임시 HOME 과 임시 git 저장소에서 `lock_named_live` 로 잠금을 쥐고(프로세스 안, 다른 파일 기술자라 `flock` 이 부딪쳐요) 항목의 임대는 `expire` 로 끝내요. 지금 코드에서 AC-1·AC-3·AC-4·AC-5 가 실패해요.
2. `lane_registry.rs` 에 읽기 함수 하나를 더해요: 폴더가 없으면 `None`, 폴더를 통째로 못 읽으면 오류, 아니면 항목(깨진 것은 이름만)과 잠금이 잡힌 이름을 돌려줘요.
3. `wip_limit.rs` 의 `live_lanes` 를 `Lanes { live, unreadable, registry_error }` 로 바꾸고 `Reading`·`refusal`·`json` 에 `registry_error` 와 `unreadable` 을 실어요(REQ-1…REQ-8). GREEN.
4. 가드 뮤턴트(잠금 기준을 빼면 AC-1 이, 리더 제외를 빼면 AC-3 이, `registry_error` 막기를 빼면 AC-4 가 죽어요), `cargo fmt --all`, `moai check affected --card` 영수증, READY.

## Risks

- R-1: 잠금이 잡힌 채 멈춘 세션(실행기는 살았는데 호스트가 죽음)을 레인으로 세요. 그 세션은 임대 기준에서도 이미 4시간 동안 세어지므로 새 위험이 아니에요. 카드 쪽 임대(`card reclaim`)가 죽은 카드를 상한에서 빼요.
- R-2: 이름 규칙(`agent-`, `<프로젝트>-agent-`)을 벗어난 레인 이름은 항목이 없으면 안 세어요. `session-register` 로 등록하면 항목이 생겨서 세어져요(레인 글의 등록 단계).
- R-3: 폴더를 못 읽으면 모든 take 가 막혀요. 이유 문구가 폴더 경로를 말하고, `wip_limit` 0 이 우회로가 아니라 응급 스위치로 남아요(사용자의 프로필 값이에요).
