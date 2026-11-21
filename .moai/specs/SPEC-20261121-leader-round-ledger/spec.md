---
id: SPEC-20261121-leader-round-ledger
title: "leader-round summarises per-lane next actions from the card ledger and flags a dispatch that forbids its own work"
status: draft
revision: 2
created: 2026-11-21
parent: null
---

# leader-round summarises per-lane next actions from the card ledger and flags a dispatch that forbids its own work

## Intent

codex 리더 무중단 보고서(codex-leader-nonstop-2026-11-21.md, 02:59-03:07Z 실측): Orca 자동화는 5분마다 살아 있지만 판단하지 않는 감독자를 깨우고, 판단 리더는 13초 요약 뒤 턴을 끝내며, 26개 응답 없는 inbox 요청과 14라운드 0 done 이었어요. 보고서 P6(P1): leader-round 이 카드 원장을 읽어 레인별 다음 행동을 요약해요 — inbox 항목은 answer, 만료 임대는 reclaim 제안(결정 L7), 보유 세션 없음은 stranded 보고, 남은 컨텍스트 15% 이하+카드 경계는 clear_redispatch. P7: 배차 문장이 스스로의 일을 금지하면(읽기 전용인데 코드 카드 등) 한 줄 경고(#3005 F4·N3 계열). leader-round 의 Reads only 계약은 그대로고요 — 읽기는 card 명령의 원장 읽기 경로를 재사용해요. C-271(loop 도구의 원장 읽기)과 같은 영역 조율은 카드 노트로 했어요.

## Non-goals

- leader-round 의 쓰기 동작 추가: Reads only 계약 유지 — reclaim·answer 의 실행은 리더가 한다.
- GitHub 스냅샷 기존 부분(close·handoff·gate·end_condition)의 변경.
- 카드 원장 스키마 변경: 읽기만.

## Requirements

- REQ-1: leader-round shall read the project card ledger read-only and add a cards section to its answer whose items name, per card, one next action: answer for a card whose newest note is an unanswered decision request, reclaim for a held card whose lease ended before now, and stranded for a held card whose holder session no lane in the snapshot reports. · failure: none — a ledger read failure is one not_run item and the rest of the answer stays.
- REQ-2: The cards section shall carry a clear_redispatch item for a lane whose snapshot shows the context at fifteen percent or less and a card boundary marker. · failure: none — the lane's own end condition stays in its existing section.
- REQ-3: leader-round shall add a one-line warning to its answer when the round's dispatch text contains a prohibition word and the dispatched card's kind changes the repository. · failure: none — no prohibition word means no warning.
- REQ-4: The cards section shall not write anything: the ledger opens read-only and no command in this path writes a file, a note or a reservation. · failure: none — the read-only contract of leader-round stays.
- REQ-5 (slice 1): The round answer shall carry a lanes section built from the lane registry joined to the ledger: for each registered lane whose entry names a session, the cards that session holds with their state, the seconds of lease left, the newest note kind of each held card and whether a held card is recorded landed while the card is not closed. · failure: none — a registry or ledger read failure is one not_run item and the rest of the answer stays.
- REQ-6 (slice 1): The lanes section shall mark a lane idle when its session holds no card and no newest note kind names a leader action. · failure: none — a lane the registry does not list is not in the section.
- REQ-7 (slice 1): The lane join shall read the registry entry's session field and the ledger reservation's holder, and a session id that changed after a clear simply matches the new holder or holds nothing, without a stale-session report. · failure: none — the read-only contract stays.

## Acceptance

- AC-1 (REQ-1, REQ-4): a card fixture ledger with an unanswered decision card, an expired-lease card and a held card of a session no lane reports yields the three actions answer, reclaim and stranded in one cards section, with the ledger file bytes unchanged.
  verify: `cargo test --locked --offline --test leader_round_cli` — the card fixture scenes
- AC-2 (REQ-2): a lane observation at fifteen percent context with a card boundary marker yields the clear_redispatch item for that lane.
  verify: `cargo test --locked --offline --test leader_round_cli` — the redispatch scene
- AC-3 (REQ-3): a dispatch text with read-only wording over a fix card yields exactly one warning line, and the same text over a docs card yields none.
  verify: `cargo test --locked --offline --test leader_round_cli` — the warning scenes
- AC-4: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-leader-round-ledger/**`, `src/leader_round.rs`, `src/ops/leader_round.rs`, `tests/leader_round_cli.rs`, `src/lane_ledger.rs`, `tests/lane_ledger_cli.rs`
- read-only: `src/card/command.rs`, `src/ledger/**`, `src/lane_registry.rs`
- forbidden: leader-round 의 쓰기 동작, GitHub 스냅샷 부분 변경, 원장 스키마 변경, 카드 번호의 출력 사용

## Plan

1. RED: 카드 픽스처 원장(응답 없는 결정 카드·만료 임대 카드·홀더 없는 카드)을 tests/leader_round_cli.rs 에 올려 세 행동과 무기록을 먼저 요구해요.
2. GREEN: 원장 읽기 헬퍼(읽기 전용) + round 답의 cards 섹션 + P7 경고 단위.
3. READY 전체 검사 → 감사는 리더에게 → READY.

### Slice 1 (2026-11-21 리더 디스패치, 레인별 원장 요약)

- RED: 임시 HOME + 임시 원장 다섯 장면을 tests/lane_ledger_cli.rs 에 올려요 — MERGE-READY 카드를 쥔 레인, SWITCH + 반납한 카드를 가진 레인, landed 인데 아직 안 닫힌 카드를 쥔 레인, 아무것도 안 쥔 레이idle, /clear 뒤 세션 id 가 바뀐 레인.
- GREEN: src/lane_ledger.rs 의 읽기 전용 요약(원장 읽기 경로 재사용) — registry entry.session ↔ reservation.holder 조인.
- 짝 줄: src/leader_round.rs 와 tests/leader_round_cli.rs 는 Codex 제거 카드(감사 PASS, 미착지, head 확인 — 그 head 는 leader_round 를 아직 만지지 않음)가 쥐어요; PAIR-REQUEST 를 그 카드에 남겼고 배선과 그 장면은 해제 뒤 옮겨요. check affected 의 base 는 머지 첫 부모.

## Risks

- 원장 읽기는 프로젝트 원장 파일이 있을 때만 의미가 있어요: 없는 프로젝트는 cards 섹션이 not_run 한 줄이고 답의 나머지는 그대로예요.
- 보유 세션 유무는 스냅샷의 레인 관측과 비교하는 근사예요: 세션이 관측에 없어도 곧 돌아올 수 있으므로 stranded 는 보고일 뿐 명령이 아니에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 카드 읽기는 card 명령의 원장 읽기 경로를 재사용하고 leader-round 는 아무것도 쓰지 않는다 — Reads only 계약과 원장 스키마 무변경을 지킨다. | src/leader_round.rs 문서, 노트 4866 | 카드 열림 |

## Result

(status draft — written when the slice lands)
