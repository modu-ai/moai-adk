---
id: SPEC-20261119-card-reclaim-guard
title: "card reclaim takeover gates — live holder, dirty worktree, three reclaims"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-card-reclaim
---

# card reclaim 인계 안전 게이트 — 더러운 worktree·검증 불가·세 번 한도

## Intent

카드 C-43 기준 4의 기계적 절반: "만료 + 미커밋/미push → 회수 안 됨(사유 포함). 네 번째 회수 거절".
[`SPEC-20261119-card-reclaim`] 의 revision 1 은 `card reclaim` 을 끝난 lease 면 조건 없이 넘겼고
(src/ledger/reservation.rs:1054 "Nothing checks that the old holder is really gone"), 그 SPEC 의
Non-goals 가 이 조각을 후속으로 이름뒀어요. 이 조각은 명령 층에 **worktree 증거 게이트**를 더하고
원장 층에 **3회 한도**를 더해요. 게이트는 인계(takeover)에만 적용하고, 보유자 자신의 끝난 lease
재예약은 갱신으로 보고 건너뛰어요(revision 1 AC-2 유지).

"살아 있는 프로세스" 게이트는 이 조각 밖이에요(아래 Non-goals): 등록부 항목이 호스트 세션 id 를
기록하지 않아서(src/lane_registry.rs 의 register json!: schema_version·name·instance·lease_until·
registered_at·mode 뿐, 2026-11-19 측정) 카드 보유자(세션 UUID)를 생존 기록과 이어줄 키가 없어요.
옵션은 카드 C-43 의 결정 노트에 기록했어요.

[`SPEC-20261119-card-reclaim`]: .moai/specs/SPEC-20261119-card-reclaim/spec.md

## Non-goals

- 살아 있는 lease 의 회수(거절은 revision 1 REQ-2 가 이미 해요).
- **보유자 생존 게이트** — 세션 id ↔ 레인 이름의 기계적 연결이 없어 후속 조각(카드 C-43 결정 노트의
  옵션: 등록부가 세션 id 를 기록 / 카드가 레인 이름을 기록).
- 옛 세션이 시작한 외부 작업(push·PR·댓글)의 원격 대조 — 후속 조각.
- `expired_candidate`·`held_for_review` 상태 만들기(복구 모듈 쪽, C-109 영역).
- `src/ops/card.rs`(C-94 예약)·`src/ledger/recover.rs`(C-109 대기) 편집.

## Requirements

- REQ-1: When the reclaimed reservation names a worktree that exists, the command shall refuse while that worktree holds uncommitted changes or unpushed commits, naming both counts.
- REQ-2 (REQ-1): IF the recorded worktree path does not exist, THEN the command shall go on to the ledger.
- REQ-3 (REQ-1): IF the worktree's git reads fail, THEN the command shall refuse as unverifiable and change nothing.
- REQ-4: When the reclaimer is the old holder itself, the command shall skip the worktree gate · failure: none — a self-reclaim renews its own reservation instead of taking one over, and the ledger's own rules still bind it.
- REQ-5: IF a card already carries three `card.reclaim` events, THEN the ledger's `reclaim` shall refuse the fourth in the same transaction and change nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 기록된 worktree 가 존재하고 미커밋 변경이나 미push 커밋을 가지면 개수를 적어 `conflict` 거절하고 원장의 모든 행이 그대로예요; 폴더가 없으면 원장까지 진행하고, git 읽기 실패는 검증 불가 거절이어요.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_gates_`
- AC-2 (REQ-4): 보유자 자신의 끝난 lease 재예약은 worktree 가 더러워도 게이트를 건너뛰어 갱신돼요.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_gates_self_reclaim`
- AC-3 (REQ-5): 같은 카드의 회수 세 번까지는 되찾아지고(각각 옛 세대 `reclaimed`, 새 세대 `held`), 네 번째는 원장이 같은 트랜잭션에서 거절해요 — 명령 줄에서도 `conflict` 로, 행은 그대로예요.
  verify: `cargo test --locked --offline --test ledger_reservation reclaim_limit_`; `cargo test --locked --offline --test card_reclaim_cli reclaim_limit_`
- AC-4 (REQ-1…REQ-5): 이웃 검사.
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

1. RED: `tests/card_reclaim_cli.rs` — `tests/common/git.rs` 로 만든 worktree(미커밋·미push·깨끗함·없음·git 못
   읽음), 자기 재예약 건너뛰기, 세 번의 회수 뒤 네 번째 거절. `tests/ledger_reservation.rs` — 원장 한도.
2. GREEN: `src/card/dispatch.rs` `reclaim` — 원장을 읽기 전용으로 살짝 열어 현재 보유자·worktree 를 읽고,
   자기 재예약이 아니면 worktree 게이트(`git status --porcelain` 개수, `rev-list --count HEAD --not
   --remotes`)를 지나요. `src/ledger/reservation.rs` `reclaim` — 같은 트랜잭션에서 `card.reclaim` 사건을
   계수해 네 번째를 거절해요.
3. 이웃 검사(AC-4)와 heavy-slot.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | worktree 증거는 `status --porcelain --untracked-files=normal` 개수와 `rev-list --count HEAD --not --remotes`(fetch 없음) — `lane_reset` 과 같은 의미 | 한 가지 낱말 한 가지 판정; 기존 검사자와 같은 의미 | 다른 git 질의로 |
| D-2 | worktree 경로가 없으면 진행(잃을 작업이 없음), git 읽기 실패만 검증 불가 거절 | "증거 없음 ≠ 검증 실패"; 회수 대상이 사라진 폴더인 흔한 모양을 막지 않기 | 부재도 거절로 |
| D-3 | 게이트는 인계에만; 자기 끝난 lease 재예약은 건너뜀 | revision 1 AC-2 의 self-reclaim 유지; 06-A D1(자기 만료는 인계 아님) | 자기 재예약도 게이트에 넣기 |
| D-4 | 3회 한도는 원장 트랜잭션의 `card.reclaim` 사건 계수; 자기 재예약도 셈 | 판정을 원장 한 곳에; 사건이 유일한 역사 | 자기 재예약을 다른 kind 로 분리 |
| D-5 | 생존 게이트는 후속: 등록부 항목이 호스트 세션 id 를 기록하지 않아 보유자와 이어줄 키가 없음(2026-11-19 측정) | 옵션(등록부가 세션 id 기록 / 카드가 레인 이름 기록)은 카드 C-43 결정 노트가 이름둠 | 등록부가 세션 id 를 기록하게 되면 후속 조각이 넣음 |

## Risks

- 미push 개수는 원격 조회 없이 로컬로 셈 — 원격에만 있는 작업은 못 봐요(후속).
- worktree 경로는 예약이 기록한 문자열이므로 심링크 뒤 다른 폴더일 수 있어요 — 게이트는 읽기만 해요
  (status·rev-list), 쓰지 않으니 심링크 따라가기의 쓰기 위험이 없어요.
- worktree 게이트는 예약 기록과 실제 폴더 사이의 누수만 막아요: 보유자가 게이트 뒤에서 계속 일하면
  원장은 못 막아요(생존 게이트의 후속 근거).
