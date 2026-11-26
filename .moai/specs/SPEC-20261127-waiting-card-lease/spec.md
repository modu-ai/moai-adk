---
id: SPEC-20261127-waiting-card-lease
title: "An ended lease names card renew as the way back, and a waiting lane renews its card"
status: draft
revision: 2
created: 2026-11-27
---

# An ended lease names card renew as the way back, and a waiting lane renews its card

## Interview

- Q1: 보유자가 끝난 임대를 `card renew` 로 되살릴 수 있다는 것을 알렸어요. 범위를 어떻게 할까요? → A: 리더 메시지(2026-11-27, 세션 간 메시지): "1. card renew --help 의 틀린 문장을 실제 동작에 맞게 고쳐요. 2. 레인 규칙에 한 줄: 답·착지를 기다리기 전에 card renew --lease-secs 86400, 끝났으면 reclaim 대신 renew. 3. 위의 1·2 로 충분하면 원장 동작은 바꾸지 않아요. SPEC 도 그 크기로 줄여요. 기준 2(측정 앞뒤)는 그대로예요."
- Q2: 카드 기준 1 을 어떻게 판단할까요? → A: 아직 답이 없어요. Decisions D-1(추천 A)이에요.

## Intent

보유 레인이 리더 답이나 다른 카드의 착지를 기다리는 동안 예약의 임대(lease)가 끝나면 `card check` 는 "the lease ended" 로만 거절해요. 레인은 `card reclaim` 을 써서 카드마다 3회인 한도를 하나씩 써요.

실제 동작은 달라요. 끝난 임대도 예약이 `held` 이고 같은 generation 이면 보유자가 `card renew` 로 되살려요(`src/ledger/reservation.rs` 의 `renew` 는 살아 있는 임대를 요구하지 않아요, 커밋 a82f01c016; 시험 `an_ended_lease_is_refused_by_the_fence_and_handoff_and_accepted_by_renew_and_release`, 2026-11-27 에 돌려 1 passed). 그런데 `card renew --help` 는 "Needs a lease that has not ended", 운영 문서는 "살아 있는 lease 필요" 라고 틀리게 적었고, 거절 문장은 빠져나갈 길을 적지 않아요.

장부를 읽어서 잰 "앞" 값(2026-11-27 측정, reclaim 이 일어난 UTC 날짜별; "blocked"는 옛 임대가 끝난 때의 최신 노트가 `LANE-BLOCKED`, "review"는 reclaim 때 카드 상태, 둘은 겹치지 않아요):

| 날짜 | reclaim | review | blocked |
| --- | --- | --- | --- |
| 11-19 | 9 | 4 | 0 |
| 11-20 | 22 | 3 | 1 |
| 11-21 | 25 | 3 | 0 |
| 11-22 | 23 | 5 | 0 |
| 11-23 | 13 | 0 | 1 |
| 11-24 | 22 | 1 | 0 |
| 11-25 | 16 | 7 | 2 |
| 11-26 | 8 | 0 | 5 |

review 카드의 임대는 이미 24시간으로 늘어나요(`SPEC-20261125-review-lease-hold`). 남은 것은 `active` 카드를 든 채 기다리는 때예요.

## Non-goals

- 원장 동작(쓰기, 거절 코드, 임대 규칙, sweep, reclaim, 3회 한도)은 바꾸지 않아요.
- 노트를 적을 때 임대를 자동으로 늘리거나, 끝나기 전에 경고를 넣지 않아요.

## Requirements

- REQ-1: The help of `card renew` shall say that the holder renews its own held reservation whether or not the lease has ended, at the same generation, until another session reclaims it or a sweep returns the card. · failure: none — help text.
- REQ-2: The operations reference and the leader rules shall state the rule of REQ-1 wherever they speak of renewing or taking over an ended lease. · failure: none — document text.
- REQ-3: The lane rules and the WORK route shall tell a lane to renew its held card with `--lease-secs 86400` before it waits for an answer or for another card's landing, and to renew rather than reclaim its own ended lease. · failure: none — document text.
- REQ-4 (D-1 A): When a fence refuses a holder because its own lease has ended, the refusal shall name `moai card renew <card> --generation <generation>` as the way to keep the reservation. · failure: none — only the message changes; the code and the ledger writes stay.
- REQ-5: The change shall leave every ledger write, refusal code and lease rule as it is. · failure: none — nothing else changes.
- REQ-6: The SPEC folder shall hold the read-only query that counts reclaim events per day for cards in review or blocked, with the values measured before the change. · failure: none — a query file and a table.

## Acceptance

- AC-1 (REQ-4, REQ-5): in a temporary ledger, `card check` of an ended own lease is refused with the code it had and a message that names `card renew` with its generation option; `card renew` then succeeds at the same generation, `card check` answers `allowed: true`, and no `card.reclaim` event exists · verify: `cargo test --locked --offline --test card card_ended_lease_renew_cli::ac1_`
- AC-2 (REQ-1): the help of `card renew` says an ended lease is renewed and no longer says that a live lease is needed · verify: `cargo test --locked --offline --test card card_ended_lease_renew_cli::ac2_`
- AC-3 (REQ-2, REQ-3): the WORK route carries the renew sentence (a test pins it), and the lane rules, leader rules and operations reference name `card renew` for a waiting or ended lease · verify: `cargo test --locked --offline --test next_flow_skill_scan work_renews_before_waiting`
  verify: `grep -n "card renew" docs/lane-rules.md docs/leader-rules.md docs/design-review/appendix/operations/card.md`
- AC-4 (REQ-6): the query file in this folder gives the table of the Intent from a read-only export; the values after the change are not part of this AC (card criterion 2 stays open until they are measured) · verify: `moai card export --out target/reclaim-measure`
  verify: `grep -h '^{' target/reclaim-measure/C-*.md > target/reclaim-cards.jsonl`
  verify: `jq -nc -f .moai/specs/SPEC-20261127-waiting-card-lease/reclaim-count.jq --slurpfile ev target/reclaim-measure/events.jsonl --slurpfile cards target/reclaim-cards.jsonl`
- AC-5 (REQ-5): the neighbour tests of reservation and review hold, and the skill manifest test, pass · verify: `cargo test --locked --offline --test card ledger_reservation`
  verify: `cargo test --locked --offline --test card card_review_lease_hold_cli`
  verify: `cargo test --locked --offline --test source_manifest`

## Scope

- write: `.moai/specs/SPEC-20261127-waiting-card-lease/**`
- write: `src/ledger/reservation.rs`
- write: `tests/card/card_ended_lease_renew_cli.rs`, `tests/card/main.rs`, `tests/next_flow_skill_scan.rs`
- write: `docs/lane-rules.md`, `docs/leader-rules.md`, `skills/moai/workflows/work.md`
- write: `src/ops/card.rs`, `docs/design-review/appendix/operations/card.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/ledger/sweep.rs`, `tests/card/ledger_reservation.rs`, `tests/card/card_review_lease_hold_cli.rs`, `tests/source_manifest.rs`
- forbidden: the real HOME, the installed moai binary and skill copies, `CLAUDE.md`, the ledger schema, Python

공유 파일(`src/ops/card.rs`, 운영 문서, source-manifest)은 필요한 동안만 예약 범위에 넣고 커밋 직후 돌려줘요.

## Plan

1. RED: AC-1·AC-2 테스트와 AC-3 의 글 고정 테스트를 먼저 써요.
2. 거절 문장과 도움말, 문서 네 곳을 고치고 `work.md` 의 manifest 해시를 갱신해요.
3. 새 테스트마다 가드 뮤턴트(Edit 로), `moai check affected --card`, READY.
4. 카드 기준 2 의 "뒤" 값은 착지하고 새 빌드를 설치한 뒤에 같은 질의로 재요. 착지는 `--keep-open` 이에요.

## Risks

- R-1: 규칙을 지키지 않고 `renew` 없이 기다리면 임대는 여전히 끝나요. 보유자가 2시간 동안 원장에 아무것도 안 썼으면 리더 라운드의 sweep 이 `active` 카드를 `ready` 로 돌릴 수 있고(`src/ledger/sweep.rs` `SILENCE_SECS`), 그 뒤에는 `renew` 가 안 돼요.
- R-2: 24시간 renew 는 그동안 다른 세션의 reclaim 을 막아요. 지금도 보유자가 할 수 있는 일이에요.

## Decisions

추천이 맨 앞이에요.

- D-1 (카드 기준 1 "review 또는 LANE-BLOCKED 카드가 임대를 조용히 잃지 않아요; 테스트가 유지된 임대나 경고 줄을 보여요"): (A, 추천) 만료 거절 문장에 `card renew` 를 넣고(REQ-4, 문장만), AC-1 테스트가 거절 문장 → 같은 generation renew → reclaim 0 을 보여요. review 쪽은 이미 24시간 연장 테스트가 있어요. 이것을 기준 1 로 인정할지는 리더 판단이에요. 끝나기 전 경고나 자동 연장은 아니에요(R-1). (B) 리더가 기준 1 문구를 이 범위에 맞게 바꿔요. (C) REQ-4 를 빼고 기준 1 을 열어 둔 채 `--keep-open` 착지해요.
- D-2 (측정 도구): (A, 추천) `moai card export` 와 `jq`. (B) sqlite3 읽기 전용(카드 본문과 배정의 말). A 를 추천하는 까닭: sqlite3 는 HOME 아래 원장 파일 경로가 필요한데, 앞서 그 경로를 찾는 명령이 훅에 거절됐고(장부 노트 10049) 답이 없었어요. export 는 원장을 읽기만 해요. export 파일에는 세션 id 가 있어서 `target/` 아래에만 두고 READY 전에 지워요.
