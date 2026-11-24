---
id: SPEC-20261125-card-take-records-worktree
title: "card take 와 card reclaim 이 예약에 워크트리를 항상 기록해서, 재시작된 레인이 card adopt 로 자기 카드를 이어받아요"
status: draft
revision: 1
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

## Non-goals

- (a), (c), (d): 후속 카드로 따로 SPEC 을 써요(Plan 끝에 적어요).
- adopt 의 증명, reclaim 의 리스·한도 검사, early reclaim 의 증명을 바꾸기.
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

## Scope

- write: `.moai/specs/SPEC-20261125-card-take-records-worktree/spec.md`, `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `tests/card/card_take_worktree_cli.rs`, `tests/card/main.rs`
- shared (줄 임대): `src/ops/card.rs` (take 와 reclaim 의 두 팔에서 실행 폴더를 넘기는 줄)
- read-only: `src/card/adopt.rs`, `src/card/early_reclaim.rs`, `src/lane_registry.rs`
- forbidden: 진짜 HOME, 설치된 moai 바이너리, 다른 프로젝트, 원장 스키마, 파일 안의 카드 번호

## Plan

1. RED 먼저: AC-1…AC-5 를 임시 HOME 과 임시 git 리포지터리(링크된 워크트리 둘)에서 빌드된 바이너리로 돌려요. 기준에서는 맞는 이유로 실패해요(워크트리가 비어 있음, adopt 가 `the reservation records no worktree` 로 거절).
2. `src/ledger/reservation.rs`: `reclaim_with` 가 `worktree_if_missing` 를 받아 옛 예약에 워크트리가 없을 때만 그 값을 새 예약과 이벤트에 써요. 기존 세 함수는 `None` 을 넘겨 그대로예요. 새 공개 함수 `reclaim_recording` 이 모두 받아요.
3. `src/card/dispatch.rs`: `TakeRequest` 와 `ReclaimRequest` 에 `running_dir` 를 더해요. take 는 `--worktree` 가 없으면 `lane_registry::worktree_of(running_dir)` 를 쓰고, reclaim 은 `reclaim_recording` 한 호출로 합쳐요(세 갈래 호출은 인자만 다르고 같은 일을 해요).
4. `src/ops/card.rs`: 두 팔이 `std::env::current_dir()` 를 넘겨요(adopt 의 `running_dir` 와 같은 방식).
5. 지킴이마다 뮤턴트, 이웃 테스트, READY 에서 `check affected` 한 번.
6. 후속 카드(리더가 열어요): (c)+(d) 리더가 사유와 함께 끊어진 예약을 풀고 한도에 걸린 카드의 출구를 내는 SPEC, (a) adopt 가 받아들일 워크트리 조건의 SPEC.

## Risks

- 실행 폴더는 레인의 워크트리일 수도, 공유 체크아웃일 수도 있어요. 공유 체크아웃에서 `--worktree` 없이 take 하면 그 경로가 기록돼요. 사실대로의 기록이고, `worktree_gate` 는 기록된 워크트리가 지워졌거나 더러울 때 takeover 를 막는 쪽이라 공유 체크아웃의 변경이 레인의 변경으로 읽힐 수 있어요. 레인은 `--worktree` 를 명시하는 것이 규칙이고(CLAUDE.md 의 "Take" 절), 이 변경은 규칙을 지키지 않은 호출에 대한 안전한 기본값이에요.
- reclaim 이 채우는 값은 reclaimer 의 폴더예요. 옛 예약이 워크트리 없이 끝났을 때, 새 holder 가 일하는 곳이 기록돼요. takeover 의 워크트리 게이트(`dispatch.rs`)는 채우기 전의 옛 값(없음)으로 돌아서 이 reclaim 에서는 이전과 똑같이 워크트리를 읽지 않아요. 다음 takeover 부터 읽어요.
- 이벤트 payload 에 키가 하나 늘어요(`worktree_filled`, 채울 때만). 읽는 쪽은 기존 키만 읽으므로 깨지지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 첫 조각은 (b) take·reclaim 이 워크트리를 항상 기록 | 카드 본문의 (b) 와 (d) 가 이것 없이는 풀리지 않아요(워크트리 없는 예약은 adopt 가 거절하고 reclaim 이 그 없음을 이어요). 소유 검사를 건드리지 않는 가장 작은 조각이고, 리더의 배차 노트가 이 조각을 권장했어요 | 기록을 빼요(두 필드) |
| D-2 | 기록하는 값은 실행 폴더의 git 워크트리(`lane_registry::worktree_of`) | adopt 가 실행 폴더의 같은 함수로 증명 7번을 읽어요. 같은 함수로 쓰면 평소 길(등록·예약·실행이 같은 폴더)에서 증명이 맞아요 | 다른 함수를 써요 |
| D-3 | `--worktree` 를 필수로 만들지 않아요 | 지금의 사용법이 깨지고, 폴더가 git 워크트리가 아닌 호출은 기록 없이 경고만 받아요 | 필수로 해요 |
| D-4 | reclaim 의 채움은 옛 예약에 워크트리가 없을 때만이고, 있는 값은 덮지 않아요 | 새 holder 는 기록된 워크트리에서 이어 가요(앞선 SPEC 의 규칙). 덮으면 소유 증거를 바꾸는 일이 돼요 | 채움을 빼요 |
| D-5 | 세 갈래 reclaim 호출을 `reclaim_recording` 한 호출로 합쳐요 | 세 갈래는 `early` 와 `kept` 의 유무만 다르고 `reclaim_with` 가 두 값을 모두 `Option` 으로 받아요. 합치면 채움 인자를 한 곳에서만 넘겨요 | 세 호출로 되돌려요 |

## Result

(구현 뒤에 적어요.)
