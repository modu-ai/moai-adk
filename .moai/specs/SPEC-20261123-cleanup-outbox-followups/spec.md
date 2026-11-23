---
id: SPEC-20261123-cleanup-outbox-followups
title: "The cleanup executor refuses to finish an attempt whose lease ended, leaves a step line for a dead attempt, records the drained ref queue, and runs from spec close --card"
status: approved
revision: 1
created: 2026-11-23
parent: SPEC-20261123-cleanup-outbox
approval: {by: user, at: 2026-11-23, basis: "user's in-session selection 2026-11-23T13:12:18Z, '그대로 승인 (Recommended)', as the leader recorded it in note 7524 of the card that implemented this SPEC (decision note, ref note:401:17743939b692ead2b71de9dc5d76fb6c); revision 1 as audited by AUDIT #401 (note 7545), not re-approved by the lane", digest: "sha256:5fd4a2aa2fba612eef64f1a11c7945ac5e35f5b2079b2b9bc93112adfc59d64b"}
---

# The cleanup executor refuses to finish an attempt whose lease ended, leaves a step line for a dead attempt, records the drained ref queue, and runs from spec close --card

## Interview

- none: 리더 결정이 요구사항을 정했어요 — 감사 노트 7393 의 N2·N4·N5·N6 을 후속으로 넘긴 LEADER 결정(2026-11-23)이 범위를 정했고, 별도 사용자 인터뷰는 없었어요. 사용자는 완성된 SPEC 을 '그대로 승인'으로 골랐어요(노트 7524).

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
| D-5 | The `lease_lost` answer of a run that got as far as the worktree and branch steps carries what that run removed and deleted (`worktrees`, `branches`, `queue`, `cleanup`, as a finished run's answer does), so `cleanup::wrote` and `card land`'s `writes` see it. The ledger still gets no line from that run (REQ-1, REQ-2). The text of REQ-3 is unchanged; the answer only says more. | AUDIT #401 (note 7545) N-1: the answer had `worktrees` and `branches` as `[]` after a run that did delete them. | Fill the lists with `[]` again at the end of `run`'s `one`. |
| D-6 | The `abandoned` line a claim leaves says the ledger saw no end of the attempt and names both causes, a run that died and a run whose lease ended before it finished, because a `lease_lost` run writes nothing either. Status `abandoned` and reason `lease_ended` are unchanged. | AUDIT #401 (note 7545) N-3. | Restore the earlier detail text. |

## Result

C-363 감사 노트 7393 의 비차단 N2·N4·N5·N6 을 후속 조각으로 구현했어요. 이 SPEC 의 승인은 사용자 몫이고, 아래 숫자는 임시 HOME 과 임시 저장소에서 잰 값이에요(실제 원장·체크아웃·워크트리·브랜치는 읽지도 쓰지도 않았어요).

- AC-1 met: `fu1_` 2 tests passed: 리스가 끝난 시도, 더 새 시도로 밀린 시도, 존재하지 않는 시도, 다른 holder, 이미 release 한 시도는 step 도 result 도 못 쓰고(`Finish::LeaseEnded`), 가장 새 살아 있는 시도만 써요. 실행기는 `lease_lost` 와 시도 번호를 담은 `error` 로 답하고 다음 실행이 `settled` 로 끝내요.
- AC-2 met: `fu2_` 2 tests passed: 죽은 시도(중단 seam, 그리고 브랜치 삭제 중 프로세스 그룹에 실제 `kill -9`) 다음 시도가 `abandoned` step 한 줄(kind `attempt`, reason `lease_ended`)을 한 번 남기고 요청을 `settled` 로 끝내요. 답에 `abandoned_attempt` 가 있어요.
- AC-3 met: `fu3_` 1 test passed: 큐에 있던 ref 마다 `queue` step 이 한 줄이고(`gone`·`still_there`), 큐가 없는 실행은 한 줄도 남기지 않아요.
- AC-4 met: `fu4_` 1 test passed: `spec close --card` 가 닫은 카드의 깨끗하고 합쳐진 워크트리와 브랜치를 지우고 답의 `card.cleanup_outbox` 에 요청이 `settled` 로 나와요. 코드는 바뀌지 않았고(D-4) 테스트가 트리거를 고정해요.

Other runs: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli --test spec_close_cli --test card_finish_cli` 85 of 85 passed; `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0; `cargo fmt --check` and `git diff --check` clean. RED: `fu1_`(2)·`fu2_`·`fu3_` 4 tests failed on the stubbed head before the change; `fu4_` passed from the start because it only pins existing behavior.

Guard mutants (committed head, each mutated line restored afterwards, all red with nextest exit 100): 리스 검사, 가장 새 시도 검사, holder 검사, release 검사, abandoned 줄, queue step, `lease_lost` 답, `card done` 의 실행기 호출(`src/card/work.rs`) — 여덟 개가 모두 죽었어요.

Gaps: `kill -9` 가 남기는 git 락(`refs/heads/<branch>.lock`, `packed-refs.lock`)은 실행기가 지우지 않아요. 다음 시도는 `failed` 로 요청을 보류하고, 락을 치운 뒤의 실행이 끝내요(테스트가 이 순서를 고정해요). 큐 줄이 임시 사유로 큐에 되돌아가도 step 은 `gone` 으로 남아요(D-3).

### 후속 조각: 감사 #401 의 비차단 N-1·N-2·N-3 (D-5, D-6)

- AC-1 met (N-1, D-5): `fu1_` 4 tests passed(새 3개 포함). 리스가 끝난 실행의 `lease_lost` 답이 이번 실행이 실제로 지운 워크트리(`removed`)와 브랜치(`deleted`)를 싣고 `cleanup::wrote` 가 true 예요. 지운 것이 없는 실행은 `removed` 가 없어서 `wrote` 가 false 예요. 원장에는 여전히 step·result 를 쓰지 않아요(claim 한 줄뿐).
- AC-2 met (N-3, D-6): `fu2_` 4 tests passed(새 1개 포함). `lease_lost` 로 끝난 시도 다음의 `abandoned` 줄이 "기록된 끝이 없고, 실행이 죽었거나 끝나기 전에 리스가 끝났다"고 말해요. 상태 `abandoned`·이유 `lease_ended` 는 그대로예요.
- N-2: 이 SPEC 의 front matter 가 `approved` 이고 `approval.basis` 가 사용자 선택(2026-11-23T13:12:18Z, '그대로 승인')을 리더가 옮겨 적은 C-401 노트 7524 를 인용해요. 승인은 새로 만들지 않았고, 요구사항·수용 기준·범위는 바뀌지 않아서 digest 가 같아요(감사 #401 이 본 리비전 1).
- 위험 경로(실제 바이너리, 임시 HOME·임시 저장소): `fu1_a_real_executor_whose_lease_ends_under_it_answers_what_it_deleted` 가 `moai card finish` 를 브랜치 삭제 도중 훅으로 멈추고, 픽스처 원장에서만 리스를 끝낸 뒤 풀어줘요. 답은 `status: lease_lost`, `worktrees[0].status: removed`, `branches[0].status: deleted`, `attempt: 1`; 원장에는 `cleanup.requested` 1·`cleanup.started` 1·`cleanup.step` 0·`cleanup.result` 0 줄이 남아요.
- Other runs: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli --test card_finish_cli --test spec_close_cli` 89 of 89 passed; `cargo fmt --check` clean.
- Guard mutants (committed head `897c1d5ed0`, each restored afterwards, `git status` clean, nextest exit 100): `lease_lost` 답에 목록을 싣는 줄 → 건너뜀(3 tests red), 브랜치 목록 → `[]`(2 red), `abandoned` 줄 문구 → 옛 문구(1 red), `wrote` → 항상 true(2 red).
