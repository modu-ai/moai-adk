---
id: SPEC-20261120-reclaim-dead-holder
title: "A reservation is reclaimed before its lease ends when its holder session is proven gone"
status: draft
revision: 1
created: 2026-11-20
---

# A reservation is reclaimed before its lease ends when its holder session is proven gone

## Intent

한 세션의 터미널이 사라졌는데도 그 세션의 예약이 임대(lease) 4시간 동안 남아서, 같은 파일을 쓸 다른 세션이 약 10시간 기다린 일이 있었어요. 지금 `card reclaim` 은 임대가 끝난 예약만 받아요. 이 SPEC 은 사용자 결정(2026-11-20, 리더 대화 선택 "허용")을 그대로 구현해요: 임대가 남아 있어도, 아래 세 증거가 모두 있을 때만 reclaim 을 받아요.

- 홀더 세션의 생존 판정이 `launcher-exited` 로 답해요(런처 생존 판정 조회, 읽기 전용).
- 예약에 적힌 워크트리가 사라졌거나, 커밋 안 한 변경과 푸시 안 한 커밋이 없어요.
- 그 워크트리를 작업 폴더로 쓰는 프로세스가 없어요.

증거가 하나라도 없거나 읽을 수 없으면 지금처럼 거절해요. 임대 기본 4시간, reclaim 평생 3회, 원장 스키마 v1 은 그대로예요. 증거는 reclaim 사건의 사유와 내용에 남겨요.

## Non-goals

- 생존 판정 자체와 레지스트리 기록(런처 생존 카드의 일).
- 임대 기본값, reclaim 한도, 원장 스키마 변경(사용자 결정 필요).
- 자기 세션의 예약 갱신(자기 reclaim)의 동작 변경.
- 임대가 이미 끝난 예약의 reclaim 규칙 변경.
- 실행 워크플로 문서(`mode-agent`·`mode-leader`)의 문장: 그 파일은 다른 카드가 쥐고 있어요.

## Requirements

- REQ-1: The `card reclaim` command shall take over a held reservation whose lease has not ended when the liveness lookup of its holder session answers launcher-exited and its recorded worktree is gone or clean and no process has that worktree as its working folder.
- REQ-2: The `card reclaim` command shall refuse a reservation with a live lease and name the first missing proof when any of the three proofs is missing.
- REQ-3: The worktree proof shall count a recorded path that is not absolute or cannot be read as unproven and refuse a worktree with uncommitted changes or unpushed commits.
- REQ-4: The process proof shall read the working folders of the running processes without writing anything and count a platform or a tool that cannot answer as unproven.
- REQ-5: The reclaim event of an early reclaim shall carry the proof in its reason and in its own payload field.
- REQ-6: The ledger transaction of an early reclaim shall refuse the reclaim when the current holder is not the holder the proof was made for.
- REQ-7: The reclaim limit of three and the lease default of four hours shall stay as they are for an early reclaim.
- REQ-8: The early reclaim shall leave the ledger tables and the schema version as they are.
- REQ-9: A reservation whose lease has ended shall be reclaimed as before and need no proof.
- REQ-10: The proof shall read the registry and the ledger and the worktree and write none of them.
- REQ-11: The `card reclaim` command shall refuse before any write a reason that is too long to carry the proof.

## Acceptance

- AC-1 (REQ-1, REQ-5): with a launcher-exited answer and a clean worktree and no process in it, a reclaim of a live lease succeeds and the event reason names the proof.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli early_reclaim_succeeds`
- AC-2 (REQ-2): an alive answer and an unknown answer each leave the live lease refused, with the missing proof named and nothing written.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli refused_without_the_liveness_proof`
- AC-3 (REQ-3): a worktree with an uncommitted change, a worktree with an unpushed commit and a reservation with no recorded worktree each leave the live lease refused.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli refused_without_the_worktree_proof`
- AC-4 (REQ-4): a process whose working folder is inside the worktree leaves the live lease refused and a folder that only shares the name prefix does not.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli process_in_the_worktree`
- AC-5 (REQ-5, REQ-8): the event keeps its kind and carries the proof in a payload field while the ledger tables stay the same.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli event_names_the_proof`
- AC-6 (REQ-6): the ledger refuses an early reclaim whose proof names another holder than the current one.
  verify: `cargo test --locked --offline --lib reclaim_early_holder`
- AC-7 (REQ-7): a card with three reclaims refuses a fourth early reclaim and the new lease is four hours.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli early_reclaim_counts`
- AC-8 (REQ-9, REQ-10): the reclaim of an ended lease works without any registry, and the proof changes no file of the registry or the worktree.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli --test card_reclaim_cli`
- AC-9 (REQ-11): a reason of the longest allowed length is refused with the proof's length named and nothing written.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli reason_too_long`

## Scope

- write: `.moai/specs/SPEC-20261120-reclaim-dead-holder/**`, `src/card/early_reclaim.rs`, `src/card/mod.rs` (one module line), `src/card/dispatch.rs`, `src/ledger/reservation.rs` (a new function beside `reclaim`), `tests/card_reclaim_early_cli.rs`
- read-only: every other file, the real HOME
- depends on: the liveness lookup `lane_registry::liveness_of_session` of the launcher-liveness SPEC; the two heads land together

## Plan

1. SPEC first (this file).
2. RED: the new integration test file and a lib test for the holder check.
3. `early_reclaim.rs`: the three proofs, each read-only, each answering proven or a named reason.
4. `reservation::reclaim_early`: the shared transaction of `reclaim` with a proven holder and without the lease check; `reclaim` calls it with none.
5. `dispatch::reclaim`: on a live lease build the proof, fold it into the reason and call the early function.
6. The unnarrowed READY set, then an independent audit (tier 2: a lease rule).

## Risks

- A proof taken just before the write can age: the holder is proven gone, so only a new session with the same id could change it; the ledger transaction re-checks the holder and the generation.
- `lsof` or `/proc` may be missing or denied: the process proof then answers unproven and the old refusal stays.
- The liveness lookup lives in another SPEC's change; until it is in the tip this change builds on its pushed head.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the early path is automatic: `card reclaim` has no new flag | the user decision names the conditions, not an option | add a flag later |
| D-2 | a reservation with no recorded worktree is unproven | the clean-or-gone proof needs a path | allow it by a later decision |
| D-3 | the proof goes into the reason when it fits and always into a payload field | the user asked for the reason; the payload is machine-readable | drop the payload field |

## Result

(written at SYNC)
