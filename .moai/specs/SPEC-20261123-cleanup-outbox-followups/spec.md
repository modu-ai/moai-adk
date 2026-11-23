---
id: SPEC-20261123-cleanup-outbox-followups
title: "The cleanup executor refuses to finish an attempt whose lease ended, leaves a step line for a dead attempt, records the drained ref queue, and runs from spec close --card"
status: draft
revision: 1
created: 2026-11-23
parent: SPEC-20261123-cleanup-outbox
---

# The cleanup executor refuses to finish an attempt whose lease ended, leaves a step line for a dead attempt, records the drained ref queue, and runs from spec close --card

## Intent

`SPEC-20261123-cleanup-outbox` 의 독립 감사(카드 노트 7393)가 막지 않는 항목 넷을 남겼고, 리더가 후속 카드로 넘겼어요 (LEADER 결정: N2·N4·N5·N6). 넷 모두 정리 실행기가 삭제를 하는 길의 기록과 경계라서 한 SPEC 에 모았어요.

- N2: 시도의 리스가 끝났거나 더 새 시도가 시작됐는데도 `finish_attempt` 가 step 과 result 를 써요. 늦게 끝난 옛 시도가 새 시도의 기록과 결과 자리를 차지할 수 있어요.
- N4: `kill -9` 로 죽은 시도는 `cleanup.started` 만 남기고 step 줄을 남기지 않아요. 다음 시도가 이전 시도가 있었다는 사실을 한 줄로도 알아보지 못해요.
- N5: `moai spec close <id> --card` 는 `card done` 을 불러서 정리를 이미 돌리지만, 부모 SPEC 의 트리거 목록(done·finish·land)에도 테스트에도 없어요.
- N6: `reference-transaction` 훅이 쌓은 큐를 정리가 비워도 outbox step 으로 남지 않아서, 어느 ref 줄이 소비됐는지 원장으로 읽을 수 없어요.

### 이미 있는 것

- `src/ledger/cleanup_outbox.rs` 의 `claim` 과 `finish_attempt`: 요청을 리스로 잡고 step·result 를 한 트랜잭션에 써요. 부모 SPEC D-4 가 이벤트 키와 리스를 정해요.
- `src/card/cleanup.rs` 의 실행기와 `src/worktree_cleanup.rs` 의 큐 읽기(`take_queue`): 답의 `queue` 키가 `gone` 과 `still_there` 를 줘요.
- `src/card/work.rs` 의 `done` 은 닫은 뒤에 `cleanup::run_pending` 을 부르고, `src/spec_close.rs` 는 `--card` 에서 그 `done` 을 불러요.

## Non-goals

- 리스 길이, 재시도 명령, 타이머. 부모 SPEC 의 갭으로 남아요.
- 죽은 시도가 지운 것을 되돌리거나 다시 쓰는 일. 다음 시도는 늘 처음부터 다시 읽어요 (부모 REQ-6).
- 큐 파일 형식과 훅 스크립트(`src/card/guard.rs`)의 변경. 이 SPEC 은 큐를 비운 결과를 기록만 해요.
- `.claude/worktrees/` 보존 범위 확대, 원격 호출, 새 명령, `moai help` 변경.

## Requirements

<!-- This SPEC adds no check that blocks a write that is allowed today. -->
- REQ-1: When the executor ends an attempt, the ledger shall write the attempt's steps and its end only while that attempt is the newest one of the request, its lease has not ended and it has not given the claim back.
- REQ-2 (REQ-1): IF the attempt is not live when it ends, THEN the ledger shall write nothing for it.
- REQ-3 (REQ-2): IF the ledger wrote nothing for the ending attempt, THEN the executor shall answer the request as `lease_lost` with an `error` that names the attempt and leave the request pending for the next run.
- REQ-4: When a claim finds that the newest earlier attempt of the request has no release and no result and its lease is over, the ledger shall write one `cleanup.step` event for that attempt with kind `attempt`, status `abandoned` and reason `lease_ended` in the claim's own transaction · failure: none — a claim that cannot write the line is the claim's own `Busy` or error, and no attempt starts.
- REQ-5: When a run takes the reference-transaction queue, the executor shall write one `cleanup.step` event of kind `queue` for each queue ref, with status `gone` for a ref that no longer exists and `still_there` for one that does · failure: none — a queue that could not be taken adds no step, and the run's other steps are written as before.
- REQ-6: When `moai spec close <id> --card` closes a card, the command shall run every pending cleanup request after the card's close and carry the `cleanup_outbox` key in the answer's `card` · failure: none — a problem of the executor is the `error` of that key and leaves the close as it was.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): an attempt whose lease ended before it finished writes no step and no result, and the executor reports it as `lease_lost`; an attempt that is still the newest and live finishes as before · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^fu1_/)'`
- AC-2 (REQ-4): after a run that stopped without a trace, the next attempt leaves one `abandoned` step of kind `attempt` for it, once, and the request still settles · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^fu2_/)'`
- AC-3 (REQ-5): a queue holding a deleted ref and an existing ref leaves one `queue` step per ref with its status, and a run with no queue leaves none · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^fu3_/)'`
- AC-4 (REQ-6): `spec close --card` on a card whose worktree and branch are clean and merged removes both and answers `cleanup_outbox` under `card` · verify: `cargo nextest run --locked --offline --test spec_close_cli -E 'test(/^fu4_/)'`

## Scope

- write: `.moai/specs/SPEC-20261123-cleanup-outbox-followups/spec.md`, `src/ledger/cleanup_outbox.rs` (the live-attempt check and the abandoned line), `src/card/cleanup.rs` (the `lease_lost` answer and the queue steps), `tests/card_cleanup_outbox_cli.rs`, `tests/spec_close_cli.rs` (one test)
- read-only: `src/card/work.rs`, `src/spec_close.rs`, `src/worktree_cleanup.rs`, `src/card/guard.rs`
- forbidden: `CLAUDE.md`, the real ledger, checkout, worktrees and branches (tests run in temporary projects only)

## Plan

1. This SPEC and the tests first (RED): `fu1_`, `fu2_` and `fu3_` in `tests/card_cleanup_outbox_cli.rs`, `fu4_` in `tests/spec_close_cli.rs`.
2. `src/ledger/cleanup_outbox.rs`: `finish_attempt` reads the attempt's own `cleanup.started` row and refuses a dead one; `claim` writes the `abandoned` line.
3. `src/card/cleanup.rs`: the `lease_lost` answer and the queue steps.
4. The dangerous path (a killed executor, a lease that ends under a running attempt) run on a throwaway fixture with the real binary; guard mutants; one `check affected` at READY.

## Risks

- A run whose lease ended mid-way has already removed what it removed, and its steps are dropped. The next attempt reads everything again and records those targets as `absent`, so the ledger still ends with a result; the only loss is the first run's own step rows, which the `lease_lost` answer and the `abandoned` line stand in for.
- The refusal reads the lease against the database clock in whole seconds, as `claim` does; two reads one second apart can disagree at the edge, and the safe side is the refusal.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The attempt's liveness is read from its own `cleanup.started` row (newest attempt, same holder, `lease_until` greater than the database clock, no `cleanup.released`), inside the write transaction. | LEADER 결정 of 2026-11-23 on audit note 7393; the same rule `claim` already uses. | Remove the check in `finish_attempt`. |
| D-2 | The dead attempt's line is written by the next claim, not by the dying run: a killed run cannot write. Its key is `cleanup.step:<request>:<attempt>:attempt:abandoned`, so a replay adds nothing. | A process killed with `kill -9` runs no code; only the next claimer can know. | Stop writing the line in `claim`. |
| D-3 | Queue steps carry the full ref name as target and the cleanup answer's own words (`gone`, `still_there`) as status; a queue line that goes back to the queue for a temporary reason is still recorded as `gone`. | The cleanup answer exposes only those two lists; widening it is not this slice. | Stop adding queue steps in the executor. |
| D-4 | `spec close --card` needs no code change: it calls `card::work::done`, which runs the executor. The slice pins that with a test and names the trigger. | `src/spec_close.rs` calls `crate::card::work::done`. | Remove the test. |

## Result

Not yet implemented.
