---
id: SPEC-20261120-reclaim-liveness
title: "Reclaim liveness gate: an ended lease whose holder process still runs is not reclaimed"
status: draft
revision: 1
created: 2026-11-20
---

# Reclaim liveness gate: an ended lease whose holder process still runs is not reclaimed

## Intent

C-43 기준 4 의 남은 문장(C-43 노트 3453 의 조각 D): 임대가 끝났어도 보유자 프로세스가 살아 있으면 회수하지 않아요. 리더 결정(2026-11-19T20:5xZ): (1) ㅂ — 세션 등록부 항목이 있는 보유자만 게이트를 걸고 항목이 없으면 오늘 동작을 유지해요. (2) 게이트의 거절문이 사용자 결정 경로(C-117 조각 2)를 안내해요. (3) 생존 탐침은 이미 있는 `handoff::probe_liveness`(살아 있음·죽음·판정 불가) 하나를 재사용해요.

연결 고리(결정 노트 2836 의 옵션 가): 보유자는 예약의 holder 세션 UUID 이고, `session-register` 항목에 그 세션 id 를 기록해 둘을 잇습니다. 항목은 등록한 프로세스의 pid 와 그 시작 시각(`ps` 의 lstart, 같은 탐침으로 읽음)도 기록해요 — 이것이 생존의 닻이에요. 등록은 짧은 명령 프로세스가 하므로 오늘의 세션에서 닻은 곧 죽은 프로세스를 가리켜요(게이트는 통과); 세션을 싸는 긴 프로세스가 닻을 남기는 날부터 게이트가 지켜줘요. ㅂ 선택이 이 과도기를 안전하게 만들어요.

## Non-goals

- 항목이 없는 보유자에 대한 게이트(ㅁ 규칙) — `card take` 의 등록 요건과 함께 별개 조각이에요.
- 살아 있는 임대의 사용자 결정 인수(C-117 조각 2) — 이 SPEC 은 거절문에서 그 경로를 안내하기만 해요.
- 박동(heartbeat) 쓰기 명령이나 임대 단축(C-117 노트 2930 의 B·D2) — 기각된 선택지예요.

## Requirements

- REQ-1: The session registry entry of `moai session-register` shall record the host session id (as `session-current` resolves it, when the host gives one), the registering process id and that process's start time, under schema version 2.
- REQ-2: A renewal by the same instance shall keep the session id, process id and start time of the first registration, because the renewing process is not the session's process.
- REQ-3: The session registry shall answer the liveness anchor (the process id and the start time) of the entry that names a given session id, reading entries of both schema versions, skipping entries it cannot read, and answering none when no entry names the session.
- REQ-4: The command `card reclaim` shall refuse, with a reason that names the holder process and points to asking the holder or a recorded user decision, when the holder's lease has ended and the holder's entry records a process that is still running.
- REQ-5: The command `card reclaim` shall refuse as unknown when the holder's entry exists but liveness cannot be determined, while a holder with no entry is answered exactly as today.
- REQ-6: The command `card reclaim` shall let a holder whose recorded process is dead, or whose process id was reused by another process, be reclaimed exactly as today.
- REQ-7: The liveness decision of the gate shall come from the one shared probe `handoff::probe_liveness`, with no second liveness code path.

## Acceptance

- AC-1 (REQ-1, REQ-2): 등록 항목 파일이 세션 id·pid·시작 시각을 기록하고, 같은 instance 의 갱신이 첫 닻을 유지해요.
  verify: `cargo test --locked --offline --test card_reclaim_liveness_cli registry_anchor`
- AC-2 (REQ-3): 세션 id 로 닻을 찾아요 — 등록된 id 는 닻, 다른 id·옛 스키마 항목은 없음.
  verify: `cargo test --locked --offline --test card_reclaim_liveness_cli anchor_lookup`
- AC-3 (REQ-4): 임대가 끝난 보유자의 닻 프로세스가 살아 있으면 회수가 거부되고 원장이 그대로예요(살아 있는 자식 프로세스로 시험).
  verify: `cargo test --locked --offline --test card_reclaim_liveness_cli alive_holder`
- AC-4 (REQ-5): 항목은 있지만 판정이 안 서면 unknown 거부, 항목이 없으면 오늘 동작.
  verify: `cargo test --locked --offline --test card_reclaim_liveness_cli unknown_and_absent`
- AC-5 (REQ-6, REQ-7): 죽은 닻·pid 재사용은 오늘처럼 회수되고, 게이트가 탐침 한 곳만 씀은 코드 읽기로 확인.
  verify: `cargo test --locked --offline --test card_reclaim_liveness_cli dead_holder`

## Scope

- write: `src/lane_registry.rs`, `src/card/dispatch.rs`, `tests/card_reclaim_liveness_cli.rs`,
  `.moai/specs/SPEC-20261120-reclaim-liveness/spec.md`
- read-only: `src/handoff.rs`(probe_liveness 호출만), `src/session_current.rs`, `tests/common/**`,
  `tests/card_reclaim_cli.rs`, `src/ledger/reservation.rs`
- forbidden: 실제 HOME과 원장, 네트워크, 실제 원격 저장소, 두 번째 생존 탐침 코드

## Plan

1. 시험 먼저(RED): `tests/card_reclaim_liveness_cli.rs` — 등록 닻 기록·유지(AC-1), 닻 조회(AC-2); 게이트 셋(AC-3·4·5)은 2b 에서 같은 파일에 더해요.
2. `src/lane_registry.rs`: 항목에 세션 id·pid·process_start 기록(스키마 2, 읽기는 1·2 둘 다), 같은 instance 갱신은 첫 닻 유지, `anchor_of_session` 조회 함수(2a).
3. `src/card/dispatch.rs`: `takeover_gates` 에 생존 게이트 — 닻이 있으면 probe: 살아 있음 → 거절(사유·C-117 경로 안내), 판정 불가 → unknown 거부, 죽음·재사용 → 통과; 항목 없음 → 오늘 동작(2b, C-132 예약이 풀리는 뒤).
4. 문서: 이 SPEC 의 Result 에 조각 나뉨과 각 시험 결과를 적어요.

## Risks

- 등록은 짧은 명령 프로세스가 하므로 오늘의 닻은 곧 죽어요 — 게이트가 실제로 지키는 날은 세션을 싸는 긴 프로세스가 닻을 남길 때예요(ㅂ 선택으로 지금은 동작 불변).
- 같은 세션 id 로 여러 이름이 등록되면 조회가 한 항목을 골라요 — 살아 있는 임대를 우선하고, 없으면 첫 항목이에요.
- `ps` 의 lstart 형식은 플랫폼별이에요(CI 는 linux-x86_64, 로컬은 macOS — 둘 다 같은 옵션으로 답해요).

## Result

- 조각 나뉨: 2a(등록부 닻 기록·유지·조회 + AC-1·AC-2 시험)를 이 가지에서 먼저; 2b(dispatch.rs 게이트 + 살아 있는 자식 프로세스 시험 AC-3·4·5)는 C-132 의 src/card/dispatch.rs 예약이 풀린 뒤 같은 카드에 이어요(노트 3550).
