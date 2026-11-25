---
id: SPEC-20261125-card-take-records-worktree
title: "card take 와 card reclaim 이 예약에 워크트리를 항상 기록해서, 재시작된 레인이 card adopt 로 자기 카드를 이어받아요"
status: draft
revision: 2
created: 2026-11-25
branch: spec/card-take-records-worktree-20261125
---

# card take 와 card reclaim 이 예약에 워크트리를 항상 기록해서, 재시작된 레인이 card adopt 로 자기 카드를 이어받아요

## Intent

mo.ai.kr 리더 세션이 보낸 개선 요청(카드 본문, 그 저장소에서의 실측)은 예약을 이어받거나 풀 수 없는 네 가지 길을 말해요. 이 저장소에서는 아직 재현하지 않은 것이라, 읽은 코드로 각 길을 확인했어요.

- (a) 세션 재시작 뒤 `session-register --continue` 다음 `card adopt` 가 "the registry entry records another worktree than the running one" 로 거절돼요. `src/card/adopt.rs` 의 증명 7번이 예약의 워크트리, 실행 폴더의 워크트리, 등록 항목의 워크트리가 하나일 것을 요구해요. 등록 항목은 처음 등록한 폴더를 기록해요. 레인이 다른 폴더에서 다시 등록하면 어긋나는 것이 맞고, 그 판단은 등록 쪽(`session-register`)의 몫이라 이 SPEC 이 만지지 않아요.
- (b) `card take` 에 `--worktree` 가 없으면 예약에 워크트리가 없어요(`src/card/dispatch.rs` 의 `TakeRequest.worktree` 는 `Option` 이고 그대로 저장돼요). 그러면 `card adopt` 는 "the reservation records no worktree" 로 거절되고, `card reclaim` 은 옛 예약의 값을 그대로 이어서(`src/ledger/reservation.rs` 의 `reclaim_with`) 새 예약도 워크트리가 없어요. 한 번 없으면 영영 없어요. 이 SPEC 의 첫 조각이 이것을 닫아요.
- (c) 끝난 세션의 예약이 리스 만료까지 감사가 끝난 head 의 착지를 막고, 리더가 풀 명령이 없어요.
- (d) 평생 한도 3회(`reclaim_with` 의 `reclaims >= 3`)에 걸리고 adopt 도 안 되는 카드는 아무도 예약을 쥘 수 없어요.

네 길마다 선택지를 읽었어요 (D-1):

| 길 | 선택지 | 고름 |
| --- | --- | --- |
| (a) adopt 가 받아들일 워크트리 조건 | 등록 항목의 워크트리를 예약의 워크트리로 갈아 끼우는 것을 허용 / 등록 항목이 없을 때 예약의 워크트리만 믿기 / 지금 그대로 | 첫 조각이 아니에요. 증명 7번은 "같은 레인, 같은 트리" 의 유일한 증거라 약하게 하지 않아요. (b) 가 닫히면 등록과 예약이 같은 폴더를 가리키는 평소 길은 통과해요 |
| (b) take·reclaim 이 워크트리를 항상 기록 | 실행 폴더의 git 워크트리를 기록 / `--worktree` 를 필수로 만들어 거절 / 지금 그대로 | **첫 조각**: 실행 폴더의 워크트리를 기록해요. 필수로 만들면 기존 사용법(`--worktree` 없는 take)이 깨져요 |
| (c) 끊어진 예약을 리더가 사유와 함께 풀기 | `card reclaim` 에 리더 사유로 리스 확인을 건너뛰는 길 / 새 명령 `card release --by leader` / 지금 그대로 | 후속 카드. 소유 검사의 약화가 걸려 있어 따로 SPEC 이 필요해요 |
| (d) 한도에 걸린 카드의 출구 | 리더 사유로 한 번 더 / 교체 명령 / 한도 올리기 | 후속 카드, (c) 와 같은 SPEC 에서 다뤄요 |

고른 조각은 (b) 하나예요. 소유 검사는 어느 곳도 약해지지 않아요: 살아 있는 세션의 예약은 여전히 못 가져가고(`LeaseLive`), adopt 의 일곱 증명과 reclaim 의 평생 한도는 그대로예요.

개정 2 (2026-11-26)는 남은 조각 (c)와 (d)를 닫아요. 개정 1 이 착지한 뒤(머지 `1637e28a7b`)의 코드를 다시 읽고, 원장(`ledger.db`, 읽기 전용)으로 지난 26시간의 `card.reclaim` 13건을 셌어요.

- 13건 중 조기 증명(`early_proof`)을 쓴 것은 0건이에요. 모두 임대가 끝난 뒤의 reclaim 이었어요. 끝난 세션의 예약은 임대가 끝날 때까지 기다린 뒤 `reclaim` 으로만 풀렸어요.
- 그 reclaim 은 새 holder 를 만들고 평생 한도 3회를 써요. 지금 한도에 닿은 카드는 4장이에요(원장에서 `card.reclaim` 이벤트가 세 번 센 카드의 수). `reclaim` 은 한 번 더 받지 않으므로 그 카드는 `card reclaim` 으로는 아무도 쥘 수 없어요.
- `review` 카드는 `card release` 가 거절하고 `card sweep` 도 건너뛰어요. 착지를 git 으로 증명한 `review` 카드의 닫기(`card finish --by leader`)는 다른 세션의 살아 있는 임대에도 이미 통과해요(`src/ledger/workflow.rs` 의 `proven_review`). 막히는 것은 닫기가 아니라 이어 가기예요: 착지 뒤에 기준이 열려 있는 카드(`card land --keep-open`)나 한도에 닿은 카드를 아무도 쥘 수 없어요. 기준 표시·노트·reopen 은 holder 만 하기 때문이에요.
- 리더가 reclaim 하지 않고 "이 예약을 풀어요" 라고 사유와 함께 말할 길이 없어요. `card withdraw`·`card fold` 는 카드를 `hold` 로 보내서 일이 사라져요.

그래서 개정 2 는 새 길 하나를 더해요. 확인된 리더 세션이 `card release --by leader --reason <text>` 로 다른 세션의 `held` 예약을 `released` 로 끝내요. 임대가 끝났으면 그대로, 임대가 살아 있으면 조기 증명(`launcher-exited`, 워크트리 깨끗함, 프로세스 없음)이 모두 맞을 때만이에요. 이 길은 reclaim 이 아니라서 한도를 쓰지 않고, 리더가 카드를 넘겨받지도 않아요. 카드는 `ready` 로 돌아가요(`review` 카드는 허용된 전이 `review→active→ready` 를 한 트랜잭션에서 지나요). 아무 레인이나 `card take` 로 쥐면 되고 take 는 한도와 무관해요. 살아 있는 세션의 예약은 증명이 없으므로 풀리지 않고, 판정 PASS 가 기록된 `review` 카드는 풀지 않아요(착지가 먼저예요).

## Non-goals

- (a): adopt 가 받아들일 워크트리 조건은 후속 카드예요. (b)가 닫히면 평소 길은 통과하고, 등록 항목의 워크트리 판단은 `session-register` 의 몫이에요(D-9).
- adopt 의 증명, reclaim 의 리스·한도 검사, early reclaim 의 증명을 바꾸기. 개정 2 의 리더 해제는 조기 증명 함수를 그대로 부르고 약하게 하지 않아요.
- 한도 3회를 올리거나 한도를 건너뛰는 reclaim 옵션(D-8).
- FAIL 뒤의 수리 길: FAIL 뒤 첫 reclaim 은 한도에 세지 않아요(`SPEC-20261125-review-lease-hold` REQ-7). 이 SPEC 은 그것을 바꾸지 않아요.
- 임대가 살아 있는 세션의 예약을 증명 없이 푸는 길, 리더 말 한마디로 임대 확인을 건너뛰는 길.
- 이미 워크트리 없이 만들어진 예약을 고치기: 그 예약은 다음 `card reclaim` 의 채움으로만 풀려요(REQ-3).
- `session-register` 가 기록하는 폴더를 바꾸기.

## Requirements

- REQ-1: When `card take` is given no `--worktree`, it shall record the git worktree of the folder it runs in as the worktree of the new reservation. · failure: none — 실행 폴더가 git 워크트리가 아니면 워크트리를 기록하지 않고, 답의 `warnings` 가 그 사실을 말해요.
- REQ-2: When `card take` is given `--worktree`, it shall record that path as given. · failure: none — 지금과 같은 동작이에요.
- REQ-3: When `card reclaim` takes over a reservation that records no worktree, it shall record the git worktree of the folder it runs in on the new reservation. · failure: none — 실행 폴더가 git 워크트리가 아니면 비어 있는 채로 두고, 이벤트는 채움이 없었다고 말해요.
- REQ-4: When `card reclaim` takes over a reservation that records a worktree, it shall keep that worktree on the new reservation. · failure: none — 지금과 같은 동작이에요.
- REQ-5: When `card reclaim` writes its event, the `card.reclaim` event shall carry `worktree_filled: true` exactly when the reclaim filled a missing worktree. · failure: none — 테스트가 이벤트 payload 를 읽어요.
- REQ-6: When a lane takes a card without `--worktree` from its own worktree and its session is later replaced by `/clear`, `card adopt` from the same worktree shall move the reservation to the new session at the next generation. · failure: none — adopt 의 증명은 바뀌지 않아요.
- REQ-7: IF the holder session is alive, meaning its lease has not ended and no proof of an early reclaim exists, THEN `card reclaim` shall refuse with `lease_live`. · failure: none — 워크트리를 채우는 일이 살아 있는 예약을 가져가는 길이 되지 않아요.
- REQ-8: When the confirmed leader session runs `card release` with `--by leader`, a `--reason` and the generation of a `held` reservation of another session whose lease has ended, the ledger shall end that reservation as `released`, write one `card.release` event that names the old holder, generation, lease end, `--by` and `--reason`, and leave the reclaim count of the card as it was. · failure: none — 아래 REQ-10~REQ-13 이 거절을 말해요.
- REQ-9: When the lease of that reservation has not ended and every proof of the early reclaim holds for its holder (launcher exited, worktree gone or clean, no process in it), the ledger shall end it the same way and record the proof as `early_proof` in the `card.release` event. · failure: none — 증명 함수는 `card reclaim` 이 쓰는 것 그대로예요.
- REQ-10: IF the caller is not the session of the live registered leader entry, THEN `card release --by leader` shall refuse with `not_leader` and write nothing. · failure: none — `--by leader` 라는 낱말만으로는 증명이 되지 않아요.
- REQ-11: IF the lease has not ended and a proof is missing, THEN `card release --by leader` shall refuse with `lease_live`, name the first missing proof, and write nothing. · failure: none — 살아 있는 세션의 예약은 풀리지 않아요.
- REQ-12: When a card in `active` or `review` has its reservation released this way, the ledger shall return the card to `ready` through the allowed transitions and keep its notes, criteria, links, branch and worktree record. · failure: none — 스키마와 상태 전이표는 바뀌지 않아요.
- REQ-13: IF the card is in `review` and its newest verdict is `pass`, THEN `card release --by leader` shall refuse with `verdict_pass` and write nothing. · failure: none — 감사가 끝난 head 는 풀지 않고 착지해요(`card land`).
- REQ-14: When a card that has three counted reclaims is released this way, `card take` shall take it as it takes any ready card. · failure: none — 출구는 해제 뒤의 take 예요.
- REQ-16: IF a card has three counted reclaims, THEN `card reclaim` shall refuse a fourth with `reclaim_limit` as before. · failure: none — 한도는 그대로예요.
- REQ-15: The leader and agent workflow texts of the skill shall name the leader release, its refusals, the limit exit and the rule that the reservation of a live session is never taken. · failure: none — 테스트가 두 파일의 낱말을 읽어요.

## Acceptance

- AC-1 (REQ-1): 리포지터리의 링크된 워크트리에서 `--worktree` 없이 `card take` 하면 예약의 `worktree` 가 그 워크트리의 경로이고, 폴더가 git 워크트리가 아닌 곳에서는 `worktree` 가 비어 있고 `warnings` 가 그것을 말해요 · verify: `cargo test --locked --offline --test card card_take_worktree_cli::ac1_`
- AC-2 (REQ-2): `--worktree` 를 준 `card take` 는 준 경로를 그대로 기록해요 · verify: `cargo test --locked --offline --test card card_take_worktree_cli::ac2_`
- AC-3 (REQ-3, REQ-4, REQ-5): 워크트리가 없는 옛 예약을 리스가 끝난 뒤 `card reclaim` 하면 새 예약이 실행 폴더의 워크트리를 기록하고 이벤트에 `worktree_filled: true` 가 있어요. 워크트리가 있는 예약의 reclaim 은 그 값을 지키고 이벤트에 키가 없어요 · verify: `cargo test --locked --offline --test card card_take_worktree_cli::ac3_`
- AC-4 (REQ-6): 레인이 자기 워크트리에서 `--worktree` 없이 take 하고 `/clear` 시뮬레이션(`session-register --continue`)을 한 뒤 `card adopt` 가 성공하며 새 예약이 같은 워크트리를 기록해요 · verify: `cargo test --locked --offline --test card card_take_worktree_cli::ac4_`
- AC-5 (REQ-7): 살아 있는 리스의 예약은 워크트리가 없어도 `card reclaim` 이 `lease_live` 로 거절되고 원장이 그대로예요 · verify: `cargo test --locked --offline --test card card_take_worktree_cli::ac5_`
- AC-6 (REQ-1…REQ-7): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test card card_adopt_cli`
  verify: `cargo test --locked --offline --test card card_reclaim_cli`
  verify: `cargo test --locked --offline --test card card_reclaim_early_cli`
  verify: `cargo test --locked --offline --test card ledger_reservation`
- AC-7 (REQ-8, REQ-12): 임대가 끝난 다른 세션의 `active` 카드와 `review` 카드를 확인된 리더가 `card release --by leader --reason` 로 풀면 예약은 `released`, 카드는 `ready`, `card.release` 이벤트가 옛 holder·generation·임대 끝·by·reason 을 적고, `card.reclaim` 이벤트 수는 그대로이고 노트·기준이 남아요 · verify: `cargo test --locked --offline --test card card_leader_release_cli::ac7_`
- AC-8 (REQ-9, REQ-11): 임대가 살아 있는 예약은 조기 증명이 모두 맞을 때만 풀리고(이벤트에 `early_proof`), 런처가 살아 있거나 liveness 를 모르거나 워크트리가 더럽거나 프로세스가 있으면 `lease_live` 로 거절하며 첫 증명 이름을 말하고 원장이 그대로예요 · verify: `cargo test --locked --offline --test card card_leader_release_cli::ac8_`
- AC-9 (REQ-10): 리더 항목이 없거나, 항목의 세션이 호출한 세션이 아니거나, 세션 id 가 없으면 `not_leader` 로 거절하고 원장이 그대로예요 · verify: `cargo test --locked --offline --test card card_leader_release_cli::ac9_`
- AC-10 (REQ-13): 판정 PASS 가 기록된 `review` 카드는 `verdict_pass` 로 거절하고 원장이 그대로예요 · verify: `cargo test --locked --offline --test card card_leader_release_cli::ac10_`
- AC-11 (REQ-14, REQ-16): 한도 3회에 닿은 카드는 `card reclaim` 이 `reclaim_limit` 로 거절하고, 리더 해제 뒤에는 `card take` 가 성공해요 · verify: `cargo test --locked --offline --test card card_leader_release_cli::ac11_`
- AC-12 (REQ-15): `skills/moai/workflows/leader.md` 와 `work.md` 가 리더 해제 명령, 거절 코드, 한도 출구를 말해요 · verify: `cargo test --locked --offline --test card card_leader_release_cli::ac12_`
- AC-13 (REQ-8…REQ-15): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli`
  verify: `cargo test --locked --offline --test card card_sweep_cli`
  verify: `cargo test --locked --offline --test card card_take_worktree_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-card-take-records-worktree/spec.md`, `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `src/worktree_gc.rs` (테스트 생성자에 `running_dir: None` 한 줄), `tests/card/card_take_worktree_cli.rs`, `tests/card/card_reclaim_cli.rs`, `tests/card/card_reclaim_early_cli.rs` (두 이웃 테스트의 명령이 cargo 작업 폴더가 아닌 임시 폴더에서 돌게 한 줄씩), `tests/card/main.rs`
- write (개정 2): `src/card/land/batch.rs` (`confirm_leader` 를 `pub(crate)` 로 여는 한 낱말), `tests/card/card_leader_release_cli.rs`, `skills/moai/workflows/leader.md`, `skills/moai/workflows/work.md`
- shared (줄 임대): `src/ops/card.rs` (take 와 reclaim 의 두 팔에서 실행 폴더를 넘기는 줄; 개정 2: `Release` 의 `--by`·`--reason`·도움말과 그 팔), `tests/card/main.rs` (새 테스트 모듈 한 줄)
- read-only: `src/card/adopt.rs`, `src/card/early_reclaim.rs`, `src/lane_registry.rs`
- forbidden: 진짜 HOME, 설치된 moai 바이너리, 다른 프로젝트, 원장 스키마, 파일 안의 카드 번호

## Plan

1. RED 먼저: AC-1…AC-5 를 임시 HOME 과 임시 git 리포지터리(링크된 워크트리 둘)에서 빌드된 바이너리로 돌려요. 기준에서는 맞는 이유로 실패해요(워크트리가 비어 있음, adopt 가 `the reservation records no worktree` 로 거절).
2. `src/ledger/reservation.rs`: `reclaim_with` 가 `worktree_if_missing` 를 받아 옛 예약에 워크트리가 없을 때만 그 값을 새 예약과 이벤트에 써요. 기존 세 함수는 `None` 을 넘겨 그대로예요. 새 공개 함수 `reclaim_recording` 이 모두 받아요.
3. `src/card/dispatch.rs`: `TakeRequest` 와 `ReclaimRequest` 에 `running_dir` 를 더해요. take 는 `--worktree` 가 없으면 `lane_registry::worktree_of(running_dir)` 를 쓰고, reclaim 은 `reclaim_recording` 한 호출로 합쳐요(세 갈래 호출은 인자만 다르고 같은 일을 해요).
4. `src/ops/card.rs`: 두 팔이 `std::env::current_dir()` 를 넘겨요(adopt 의 `running_dir` 와 같은 방식).
5. 지킴이마다 뮤턴트, 이웃 테스트, READY 에서 `check affected` 한 번.
6. 후속 카드(리더가 열어요): (a) adopt 가 받아들일 워크트리 조건의 SPEC. ((c)와 (d)는 개정 2 가 닫아요.)
7. 개정 2, RED 먼저: AC-7…AC-11 을 임시 HOME 과 임시 git 리포지터리에서 빌드된 바이너리로 돌려요(기준에서는 `card release` 가 `--by` 를 모르고 거절해요). 조기 증명 시험의 런처 흔적(`.live`)은 `tests/card/card_reclaim_early_cli.rs` 의 만드는 법을 읽고 새 테스트 파일에서 같은 모양으로 만들어요(그 파일은 고치지 않아요).
8. `src/ledger/reservation.rs`: `release_by_leader` 가 한 트랜잭션에서 현재 예약이 `held` 이고 generation 이 맞는지, 임대 끝, 카드 상태(`done` 거절, `review` 의 최신 판정이 pass 이면 거절)를 읽고 예약을 `released` 로, 카드를 `ready` 로(`review` 는 `active` 를 거쳐요) 바꾸고 `card.release` 이벤트를 써요. reclaim 이벤트는 쓰지 않아서 한도 계산이 그대로예요.
9. `src/card/dispatch.rs`: `release` 에 리더 갈래를 더해요. 순서는 `confirm_leader`(`not_leader`), 임대가 살아 있으면 `early_reclaim::assess` 의 증명(`lease_live` + 첫 증명 이름), 그다음 원장 호출이에요. 증명을 읽은 뒤 holder 가 바뀌면 `holder_changed` 로 거절해요(reclaim 과 같은 규칙).
10. `src/ops/card.rs`: `Release` 에 `--by`(`leader` 만), `--reason`, 도움말 한 문단을 줄 임대로 더해요. `--generation` 은 그대로 필수예요.
11. 뮤턴트(READY 의 guard-mutant 줄): 리더 확인을 뺌 → ac9 실패, 증명 없이 살아 있는 임대를 풂 → ac8 실패, 판정 pass 검사를 뺌 → ac10 실패, `card.reclaim` 이벤트도 씀(한도 소모) → ac7·ac11 실패. 이웃 테스트, `moai check affected --card` 한 번.
12. 스킬 글: `leader.md` 에 "끊어진 예약" 절(언제 `card release --by leader` 를 쓰고 무엇이 거절되는지, 한도에 닿은 카드의 출구는 해제 뒤 take, 살아 있는 세션의 예약은 어떤 길로도 가져가지 않음), `work.md` 의 "When a step stops" 표에 한 줄. 착지한 뒤 리더가 스킬 사본을 `moai update` 로 맞춰요(CLAUDE.local.md 의 사본 규칙).

## Risks

- 실행 폴더는 레인의 워크트리일 수도, 공유 체크아웃일 수도 있어요. 공유 체크아웃에서 `--worktree` 없이 take 하면 그 경로가 기록돼요. 사실대로의 기록이고, `worktree_gate` 는 기록된 워크트리가 지워졌거나 더러울 때 takeover 를 막는 쪽이라 공유 체크아웃의 변경이 레인의 변경으로 읽힐 수 있어요. 레인은 `--worktree` 를 명시하는 것이 규칙이고(CLAUDE.md 의 "Take" 절), 이 변경은 규칙을 지키지 않은 호출에 대한 안전한 기본값이에요.
- reclaim 이 채우는 값은 reclaimer 의 폴더예요. 옛 예약이 워크트리 없이 끝났을 때, 새 holder 가 일하는 곳이 기록돼요. takeover 의 워크트리 게이트(`dispatch.rs`)는 채우기 전의 옛 값(없음)으로 돌아서 이 reclaim 에서는 이전과 똑같이 워크트리를 읽지 않아요. 다음 takeover 부터 읽어요.
- 이벤트 payload 에 키가 하나 늘어요(`worktree_filled`, 채울 때만). 읽는 쪽은 기존 키만 읽으므로 깨지지 않아요.
- (개정 2) 임대가 끝난 예약을 리더가 풀면 그 holder 가 실제로는 살아 있어 일하는 중일 수 있어요(임대를 갱신하지 못한 채 긴 실행을 하는 세션). 이 길은 `card reclaim` 이 이미 허용하는 것(임대가 끝난 예약은 누구나 가져가요)과 같은 소유 기준이고, 그 세션의 다음 쓰기는 옛 generation 이라 거절돼요. 새로 열리는 것은 가져가지 않고 푸는 쪽이라 한도를 쓰지 않는다는 점이에요.
- (개정 2) 임대가 살아 있는 예약은 조기 증명이 전부 맞을 때만 풀려요. 그 증명에는 레지스트리 임대 끝 조건이 들어 있어서, 재시작한 레인의 옛 세션은 그 임대(보통 몇 시간)가 끝나기 전에는 증명되지 않아요. 이 조건을 느슨하게 하는 것은 소유 검사를 약하게 하는 일이라 이 SPEC 이 하지 않아요(D-7).
- (개정 2) `review` 카드를 `ready` 로 돌리면 진행 중이던 감사가 있을 수 있어요. 판정이 이미 PASS 면 풀지 않고, 그 밖에는 리더가 사유에 적고 푸는 판단이에요. 구현 단계에서 `card verdict` 가 `ready` 카드에 어떻게 답하는지 읽고 시험으로 고정해요(읽기 전에는 모르는 부분이에요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 첫 조각은 (b) take·reclaim 이 워크트리를 항상 기록 | 카드 본문의 (b) 와 (d) 가 이것 없이는 풀리지 않아요(워크트리 없는 예약은 adopt 가 거절하고 reclaim 이 그 없음을 이어요). 소유 검사를 건드리지 않는 가장 작은 조각이고, 리더의 배차 노트가 이 조각을 권장했어요 | 기록을 빼요(두 필드) |
| D-2 | 기록하는 값은 실행 폴더의 git 워크트리(`lane_registry::worktree_of`) | adopt 가 실행 폴더의 같은 함수로 증명 7번을 읽어요. 같은 함수로 쓰면 평소 길(등록·예약·실행이 같은 폴더)에서 증명이 맞아요 | 다른 함수를 써요 |
| D-3 | `--worktree` 를 필수로 만들지 않아요 | 지금의 사용법이 깨지고, 폴더가 git 워크트리가 아닌 호출은 기록 없이 경고만 받아요 | 필수로 해요 |
| D-4 | reclaim 의 채움은 옛 예약에 워크트리가 없을 때만이고, 있는 값은 덮지 않아요 | 새 holder 는 기록된 워크트리에서 이어 가요(앞선 SPEC 의 규칙). 덮으면 소유 증거를 바꾸는 일이 돼요 | 채움을 빼요 |
| D-5 | 세 갈래 reclaim 호출을 `reclaim_recording` 한 호출로 합쳐요 | 세 갈래는 `early` 와 `kept` 의 유무만 다르고 `reclaim_with` 가 두 값을 모두 `Option` 으로 받아요. 합치면 채움 인자를 한 곳에서만 넘겨요 | 세 호출로 되돌려요 |
| D-6 | (c) 끊어진 예약을 푸는 길은 `card release --by leader --reason` (기존 `release` 의 리더 갈래). 고르지 않은 선택지: `card reclaim` 에 임대 확인을 건너뛰는 옵션, 새 명령 이름 | 풀기는 가져가기가 아니라서 한도를 쓰지 않아요. `reclaim` 에 건너뛰기를 달면 한도 계산과 소유 검사가 한 함수에서 갈려요. 새 명령 이름은 `release` 와 같은 일이에요. 명령 하나(`release`)가 두 사용자(holder, 리더)를 가지면 `--by leader` 의 유무로 갈라요 | 리더 갈래를 빼요(`--by` 인자) |
| D-7 | 임대가 살아 있는 예약은 조기 증명(`early_reclaim::assess`)이 전부 맞을 때만 풀어요. 레지스트리 임대 끝 조건을 건너뛰는 선택지는 고르지 않아요 | 사용자 결정(2026-11-20): 증명된 holder 만. 지난 26시간 `card.reclaim` 13건 중 조기 증명은 0건이라 이 조건의 비용을 잴 자료가 없어요. 먼저 길을 열고, 임대 끝 조건이 실제 병목이면 측정과 함께 따로 SPEC 으로 다뤄요 | 증명 함수를 안 부르고 임대 끝만 보게 해요 |
| D-8 | (d) 한도에 닿은 카드의 출구는 리더 해제 → `ready` → `card take`. 한도 3회를 올리거나 건너뛰는 옵션은 고르지 않아요 | take 는 한도와 무관해서 새 규칙 없이 출구가 생겨요. 한도는 한 카드를 계속 되살리는 것을 막는 장치라 그대로 둬요 | 한도를 올려요 |
| D-9 | (a) adopt 의 워크트리 조건은 이 SPEC 에서 다루지 않아요 | (b) 뒤에 평소 길은 통과해요. 등록 항목의 워크트리는 `session-register` 가 처음 등록한 폴더라서 어긋남의 원인이 등록 쪽에 있고, 증명 7번은 "같은 레인, 같은 트리"의 유일한 증거예요 | 후속 카드로 열어요 |
| D-10 | 리더 해제의 `review` 카드는 `review→active→ready` 로 돌리고, 판정 PASS 가 있으면 거절해요 | 상태 전이표는 `review→ready` 를 허용하지 않고 스키마는 바꾸지 않아요. PASS 가 있는 카드는 풀 것이 아니라 착지할 것이에요 | `review` 카드는 풀지 않고 거절해요 |

## Result

구현 head `bb44bec790ef8468b13104d7a402733512b8090d` (기준 NeXT `580d47c6ec`), 첫 조각 (b) 만 했어요. (a), (c), (d) 는 후속 카드의 몫이에요.

- AC-1…AC-5 met: 각 verify 줄을 적힌 그대로 돌려 모두 exit 0 (ac1·ac2·ac3·ac4·ac5 각 1개 통과). 고치기 전 RED 는 ac1·ac3·ac4 가 맞는 이유(워크트리 null, adopt 의 `the reservation records no worktree`)로 실패했어요.
- AC-6 met: `card_adopt_cli` 11, `card_reclaim_cli` 27, `card_reclaim_early_cli` 35, `ledger_reservation` 28 통과. 이웃 전체(`--test card`, `--test prop_reservation`, `--lib`) 3310개 모두 통과, clippy `-D warnings` 통과.
- 이웃 실패 3개의 원인(처음 이웃 실행에서 3307/3310): `card_reclaim_cli` 와 `card_reclaim_early_cli` 의 헬퍼가 `card take` 를 cargo 작업 폴더(이 워크트리) 에서 돌렸고, take 가 이제 그 폴더의 워크트리를 기록해서 reclaim 의 `worktree_dirty` 게이트가 이 워크트리의 커밋 안 한 변경 4개를 읽고 거절했어요. 코드의 틀림이 아니라 테스트 환경의 의존이라, 두 헬퍼의 명령이 임시 폴더에서 돌게 한 줄씩 고쳤어요(소유 검사는 그대로).
- 지킴이 뮤턴트 (각각 뒤에 `git checkout` 으로 복원, 트리 깨끗함 확인): take 가 실행 폴더 워크트리를 기록하지 않음 → ac1·ac4 실패 exit 100, reclaim 이 빈 워크트리를 채우지 않음 → ac3 실패 exit 100, reclaim 이 기록된 워크트리를 덮어씀 → ac3 실패 exit 100.
- 위험 경로: 임시 HOME 의 임시 git 리포지터리에서 링크된 워크트리 take(AC-1), git 워크트리가 아닌 일반 폴더 take(AC-1 의 둘째 절, 경고), 살아 있는 리스의 reclaim 거절(AC-5)을 직접 돌렸어요. 하나라도 진짜 HOME 이나 설치된 바이너리를 쓰지 않았어요.
- 안 한 것: `moai check affected --card` 영수증은 READY 노트에 있어요.

개정 2 (2026-11-26): 승인은 원장 노트(리더의 위임 승인)에 있고 승인 블록은 착지 때 리더가 적어요. 구현은 `release_by_leader`(`src/ledger/reservation.rs`), `release_by_leader`(`src/card/dispatch.rs`), `--by`·`--reason`(`src/ops/card.rs`), `confirm_leader` 를 `pub(crate)` 로(`src/card/land/batch.rs`), 스킬 글(`leader.md` 의 두 행, `work.md` 의 한 행)이에요.

- AC-7…AC-11 met: 각 verify 줄을 적힌 그대로 돌려 ac7·ac8·ac9·ac10·ac11 이 통과했어요(고치기 전 RED 는 여섯 개 모두 맞는 이유로 실패: `card release` 가 `--by` 를 몰라 clap 사용법 오류, ac12 는 스킬 글 없음).
- AC-12 met: `card_leader_release_cli::ac12_` 통과(`leader.md` 와 `work.md` 의 낱말). 스킬 글을 읽는 테스트(`leader_route`, `mode_cards_53`, `source_manifest_facts`, `repo_scans_table`, `next_flow_skill_scan`, `skill_structure`, `mode_workflow_steps`, `workflow_size_budget`)도 통과했어요.
- AC-13 met: `card_review_lease_hold_cli`, `card_sweep_cli`, `card_take_worktree_cli` 와 이웃(`card_reclaim_cli`, `card_reclaim_early_cli`, `card_adopt_cli`, `ledger_reservation`, `card_cli`, `card_dispatch_cli`, `card_hold_gate_cli`) 199개가 모두 통과했어요.
- 지킴이 뮤턴트(각각 뒤에 `git checkout` 으로 복원, 트리 깨끗함 확인): 리더 확인을 뺌 → ac9 실패 exit 101, 살아 있는 임대를 증명 없이 풂 → ac8 실패 exit 101, 판정 pass 검사를 뺌 → ac10 실패 exit 101, `card.reclaim` 이벤트도 씀 → ac7·ac11 실패 exit 101.
- 안 한 것: `moai check affected --card` 영수증과 clippy 는 READY 노트에 있어요. 라이브러리 수준의 `HolderChanged` 시험(증명을 읽은 뒤 holder 가 바뀜)은 따로 쓰지 않았어요(not_run).
