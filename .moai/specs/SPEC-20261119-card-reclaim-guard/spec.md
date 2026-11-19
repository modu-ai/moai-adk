---
id: SPEC-20261119-card-reclaim-guard
title: "card reclaim takeover gates — live holder, dirty worktree, three reclaims"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-card-reclaim
---

# card reclaim 인계 안전 게이트 — 살아 있는 보유자·더러운 worktree·세 번 한도

## Intent

카드 C-43 기준 4: "만료 + 살아 있는 프로세스, 만료 + 미커밋/미push, 확인 불가 → 모두 회수 안 됨(사유
포함). 네 번째 회수 거절". [`SPEC-20261119-card-reclaim`] 의 revision 1 은 `card reclaim` 을
끝난 lease 면 조건 없이 넘겼고(src/ledger/reservation.rs:1054 "Nothing checks that the old
holder is really gone"), 그 SPEC 의 Non-goals 가 이 조각을 후속으로 이름둬요. 이 조각은 명령 층에
**인계 게이트**를 더하고 원장 층에 **3회 한도**를 더해요. 게이트는 인계(takeover)에만 적용하고,
보유자 자신의 끝난 lease 재예약은 갱신으로 보고 건너뛰어요(revision 1 AC-2 유지). 등록 항목이
아예 없는 보유자는 증거가 없을 뿐 검증 실패가 아니므로 진행해요(결정 D-2) — 등록부 이전 세대의
막힌 카드를 구하는 게 이 명령의 존재 이유예요.

[`SPEC-20261119-card-reclaim`]: .moai/specs/SPEC-20261119-card-reclaim/spec.md

## Non-goals

- 살아 있는 lease 의 회수(거절은 revision 1 REQ-2 가 이미 해요).
- 옛 세션이 시작한 외부 작업(push·PR·댓글)의 원격 대조 — 후속 조각(카드 본문 요구 4의 남은 절반).
- `expired_candidate`·`held_for_review` 상태 만들기(복구 모듈 쪽, C-109 영역).
- pid 기반 `probe_liveness`(src/handoff.rs) 연결 — journal 의 pid 기록은 구세대 Issue claim 이라
  카드 보유자와 연결이 약해요(결정 D-1).
- `src/ops/card.rs`(C-94 예약)·`src/ledger/recover.rs`(C-109 대기) 편집.

## Requirements

- REQ-1: When `card reclaim` targets another session's reservation, the command shall refuse while the old holder's session-registry entry holds a live lease, naming the holder, the entry name and the lease end.
- REQ-2 (REQ-1): IF the old holder has no session-registry entry, THEN the command shall go on to the worktree gate.
- REQ-3 (REQ-1): IF the holder's registry liveness cannot be told — an unreadable entry or an indeterminate registry answer — THEN the command shall refuse as unverifiable and change nothing.
- REQ-4: When the reclaimed reservation names a worktree that exists, the command shall refuse while that worktree holds uncommitted changes or unpushed commits, naming both counts.
- REQ-5 (REQ-4): IF the recorded worktree path does not exist, THEN the command shall go on to the ledger.
- REQ-6 (REQ-4): IF the worktree's git reads fail, THEN the command shall refuse as unverifiable and change nothing.
- REQ-7: When the reclaimer is the old holder itself, the command shall skip the registry and worktree gates · failure: none — a self-reclaim renews its own reservation instead of taking one over, and the ledger's own rules still bind it.
- REQ-8: IF a card already carries three `card.reclaim` events, THEN the ledger's `reclaim` shall refuse the fourth in the same transaction and change nothing.

## Acceptance

- AC-1 (REQ-1, REQ-7): 다른 세션의 끝난 예약을 되찾을 때 보유자의 세션 등록 리스가 살아 있으면 `conflict` 로 거절(사유에 보유자·항목 이름·리스 끝), 원장의 모든 행이 그대로예요; 보유자 자신의 끝난 lease 재예약은 게이트를 건너뛰어요.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_gates_`
- AC-2 (REQ-2, REQ-3): 보유자의 등록 항목이 없으면 되찾기가 진행되고, 항목을 읽을 수 없거나 등록부가 생존을 판정할 수 없으면 검증 불가 `conflict` 거절이며 원장은 그대로예요.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_gates_`
- AC-3 (REQ-4, REQ-5, REQ-6): 기록된 worktree 가 존재하고 미커밋 변경이나 미push 커밋을 가지면 개수를 적어 거절해요; 폴더가 없으면 원장까지 진행하고, git 읽기 실패는 검증 불가 거절이어요.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_gates_`
- AC-4 (REQ-8): 같은 카드의 회수 세 번까지는 되찾아지고(각각 옛 세대 `reclaimed`, 새 세대 `held`), 네 번째는 원장이 거절해요 — 명령 줄에서도 `conflict` 로, 행은 그대로예요.
  verify: `cargo test --locked --offline --test ledger_reservation reclaim_limit_`; `cargo test --locked --offline --test card_reclaim_cli reclaim_limit_`
- AC-5 (REQ-1…REQ-8): 이웃 검사.
  verify: `cargo test --locked --offline --test card_reclaim_cli`; `cargo test --locked --offline --test card_work_cli`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test ledger_reservation`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test git_env_scrub`; `cargo test --locked --offline --test ac_command_lint`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/card/dispatch.rs`, `src/card/command.rs`, `src/card/work.rs`, `src/ledger/reservation.rs`,
  `tests/card_reclaim_cli.rs`, `tests/card_dispatch_cli.rs`, `tests/ledger_reservation.rs`,
  `.moai/specs/SPEC-20261119-card-reclaim-guard/spec.md`
- read-only: `src/lane_registry.rs`(공개 API `registry_dir`·`session_live`), `src/handoff.rs`, `src/lane_reset.rs`,
  `src/ledger/schema.rs`, `src/ledger/workflow.rs`, `src/contract.rs`, `tests/common/**`
- forbidden: 실제 HOME과 원장, 네트워크, 실제 원격 저장소, 스키마 바꾸기, `src/ops/card.rs`(C-94) 편집,
  `src/ledger/recover.rs`(C-109) 편집, 실제 프로세스 종료

## Plan

1. RED: `tests/card_reclaim_cli.rs` — `session-register` 로 만든 등록 항목(살아 있는 리스/끝난 리스/없음/깨진 파일),
   `tests/common/git.rs` 로 만든 worktree(미커밋·미push·깨끗함·없음), 세 번의 회수 뒤 네 번째 거절.
   `tests/ledger_reservation.rs` — 원장 한도(세 번 되찾음, 네 번째 거절, 행 그대로).
2. GREEN: `src/card/dispatch.rs` `reclaim` — 원장을 읽기 전용으로 살짝 열어 현재 보유자·worktree 를 읽고,
   자기 재예약이 아니면 등록부 게이트(`registry_dir` 나열 + `.session` 대조 + `session_live`)와
   worktree 게이트(`git status --porcelain` 개수, `rev-list --count HEAD --not --remotes`)를 지나요.
   `src/ledger/reservation.rs` `reclaim` — 같은 트랜잭션에서 `card.reclaim` 사건을 계수해 네 번째를 거절해요.
3. 이웃 검사(AC-5)와 heavy-slot.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 생존 증거는 세션 등록부의 리스(`lane_registry::session_live`); pid probe 는 안 씀 | journal pid 는 구세대 Issue claim 기록이라 카드 보유자와 연결이 약해요 | 등록부가 pid 를 기록하면 probe 전환 |
| D-2 | 등록 항목 부재는 진행, 읽기 실패·판정 불가만 거절 | 등록부 이전 세대의 막힌 카드를 구하는 게 이 명령의 존재 이유; "증거 없음 ≠ 검증 실패" | 부재도 거절로 |
| D-3 | 게이트는 인계에만; 자기 끝난 lease 재예약은 건너뜀 | revision 1 AC-2 의 self-reclaim 유지; 06-A D1(자기 만료는 인계 아님) | 자기 재예약도 게이트에 넣기 |
| D-4 | 3회 한도는 원장 트랜잭션의 `card.reclaim` 사건 계수; 자기 재예약도 셈 | 판정을 원장 한 곳에; 사건이 유일한 역사 | 자기 재예약을 다른 kind 로 분리 |
| D-5 | 등록부 조회는 공개 API(`registry_dir`·`session_live`) + 얇은 항목-세션 대조 | 공유 인터페이스는 소유자만 고침(범위 밖) | `lane_registry` 가 `live_by_session` 을 제공하면 중복 제거 |

## Risks

- 등록부 리스는 "최근까지 살아 있었다"의 증거이지 지금 살아 있음의 증거는 아니에요: 리스가 끝난 뒤 죽은 세션은
  통과하고, 리스가 남은 채 죽은 세션은 리스가 끝날 때까지 막혀요(안전 쪽, 최대 4시간).
- 미push 개수는 원격 조회 없이 로컬 `rev-list --count HEAD --not --remotes` 로 셈(fetch 없음, `lane_reset` 과
  같은 의미) — 원격에만 있는 작업은 못 봐요.
- 등록부 파일 규칙(심링크 거절·크기 상한)의 얇은 사본이 `dispatch` 쪽 조회에 필요해요 —
  `lane_registry` 가 세션 id 조회 리더를 제공하면 없어져요(후속 한 줄).
