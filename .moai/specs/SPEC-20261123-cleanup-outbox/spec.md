---
id: SPEC-20261123-cleanup-outbox
title: "A closed card queues its worktree and branch cleanup in the ledger in the same transaction, and an idempotent executor removes them after a fresh re-check"
status: draft
revision: 1
created: 2026-11-23
parent: SPEC-20261123-worktree-slots
---

# A closed card queues its worktree and branch cleanup in the ledger in the same transaction, and an idempotent executor removes them after a fresh re-check

## Intent

카드가 끝난 뒤에도 그 카드의 워크트리와 가지가 로컬에 남아요. 2026-11-23 에 이 저장소를 읽은 값은 연결된 워크트리 347개(상한 25), 로컬 브랜치 1842개예요. `card land` 만 워크트리를 정리하고, `card done` 과 `card finish` 는 정리를 부르지 않으며, 가지는 어느 경로도 지우지 않아요 (계획서 `hooks-rules-final-plan-2026-11-23.md` §0-2).

이 SPEC 은 사용자가 확정한 지시(카드 노트 7299)를 C-363 의 다음 조각으로 구현해요. 정리 요청은 완료와 같은 SQLite 트랜잭션에 원장 행(outbox)으로 남기고, 실행기가 그 행을 멱등으로 처리해요. 실행기는 지우기 직전에 모든 조건을 다시 읽고, 워크트리를 강제 없이 지운 뒤에만 가지를 `git branch -d` 로 지워요. 정리는 Claude 의 Stop·SessionEnd 훅이 아니라 원장과 명령이 맡아요. 훅은 작업 성공을 보증하지 않기 때문이에요.

### 이미 있는 것

- `src/worktree_cleanup.rs` 의 `run`: 닫힌 카드의 워크트리를 정리 계획과 그 digest 의 apply 로 지워요 (`card land` 가 불러요). 안전 기준, `.claude/worktrees/` 보존, 고정 슬롯은 `src/worktree_gc.rs` 가 정해요.
- `src/branch_gc.rs` 의 `plan` 과 `delete_one`: 합쳐지고 쓰이지 않는 로컬 브랜치만 `git branch -d` 로 지워요 (SPEC-20261123-branch-gc).
- 원장 `event` 표: 추가만 되고 `event_key` 가 유일해요. 스키마 버전은 엄격 비교(`found == SCHEMA_VERSION`)라서 새 표를 더하면 모든 기존 원장에 마이그레이션이 필요해요. 그래서 outbox 행은 `event` 행이에요.

## Non-goals

- 기존 워크트리와 브랜치를 일괄로 지우는 일. 이미 쌓인 것은 `worktree-gc-plan`·`branch-gc-plan` 과 사용자 선택의 몫이에요.
- `.claude/worktrees/` 아래 워크트리를 지우도록 넓히는 일. 정리 계획의 `preserved_location` 은 그대로고, 거기 있는 워크트리와 그 가지는 남기고 이유를 보고해요 (SPEC-20261123-worktree-slots D-8 의 사용자 선택).
- Claude Stop·SessionEnd 훅과 `reference-transaction` 훅을 정리 실행의 시작점으로 쓰는 일. 훅 설치도 이 SPEC 의 일이 아니에요.
- 원격 브랜치와 원격 추적 ref 를 지우거나 읽는 일, `git branch -D`, `--force`.
- 새 명령과 `moai help` 변경. 실행기는 `card done`·`card finish`·`card land` 가 불러요.
- 감사 판정(`card verdict`)에서 정리를 시작하는 일. 판정은 카드를 닫지 않아서, 그때는 원장이 `done` 이 아니에요.
- 스쿼시 병합이나 리베이스 병합으로 들어간 가지를 지우는 일. 가지 끝이 통합 가지의 조상이 아니면 남기고 보고해요.

## Requirements

<!-- This SPEC adds no check that blocks a write that is allowed today. -->
- REQ-1: When `card done`, `card finish` or `card land` closes a card, the ledger shall write one `cleanup.requested` event that names the worktrees and branches the card's reservations named, in the same transaction that closes the card.
- REQ-2 (REQ-1): IF the transaction that closes a card is rolled back or refused, THEN the ledger shall hold no `cleanup.requested` event for that close.
- REQ-3: When a landing records its head and leaves the card open, the ledger shall write no `cleanup.requested` event · failure: none — nothing is written, so there is no step that can fail.
- REQ-4: When the executor takes a request that has no `cleanup.result` event, it shall write a `cleanup.started` event with a lease inside one `BEGIN IMMEDIATE` transaction.
- REQ-5 (REQ-4): IF another `cleanup.started` lease of the request is still live, THEN the executor shall leave the request to that claim.
- REQ-6: When the executor is about to remove a worktree or delete a branch, it shall read again that the card is `done` in the ledger, that the head the request names is an ancestor of the integration tip, that no active reservation or live claim names the target, and that the cleanup plan and the branch plan list the target as removable.
- REQ-7: The executor shall remove a worktree only through the cleanup plan and its apply on the plan's own digest without `--force`, and delete a branch only after every worktree of the card is removed or no longer registered and only with `git branch -d` for a branch the branch plan lists as `delete`.
- REQ-8 (REQ-6, REQ-7): IF a worktree of the card is kept or the branch plan lists the branch as `keep`, THEN the executor shall keep the branch and report the reasons, a branch whose tip is not an ancestor of the integration tip included.
- REQ-9: The executor shall touch no ref and no worktree that a reservation of the card did not name, no preserved branch, no ref another worktree has checked out, and make no remote call · failure: none — it names what it will not touch, so a refusal there is the absence of an action.
- REQ-10: When a request ends, the executor shall write one `cleanup.step` event per target and one `cleanup.result` event.
- REQ-11 (REQ-10): IF a target was kept only for a reason that passes by itself or a step failed, THEN the executor shall leave the request without a `cleanup.result` event so that a later run takes it again.
- REQ-12: When `card done`, `card finish` or `card land` ran its close, the command shall run every pending request of the project and answer a `cleanup_outbox` key · failure: none — a problem of the executor is the `error` of that key and leaves the close as it was.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): the close of a card by `card done`, by `card finish` and by `card land` each writes exactly one `cleanup.requested` event whose payload names the card's worktree and branch; a refused done, a refused finish and a landing that leaves the card open write none · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^ac1_/)'`
- AC-2 (REQ-6, REQ-7): a card whose worktree is clean and merged and whose branch is merged is closed; the worktree is gone, the branch is deleted, the answer's `cleanup_outbox` lists both as done and the ledger holds `cleanup.step` rows and one `cleanup.result` · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^ac2_/)'`
- AC-3 (REQ-8, REQ-9): a dirty worktree and its branch, a branch the integration tip does not contain, a preserved branch, a branch another worktree has checked out and a worktree under `.claude/worktrees/` each stay with their reason, and no remote ref changes · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^ac3_/)'`
- AC-4 (REQ-4): a second executor run after a settled request writes no event and changes no ref or worktree · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^ac4_/)'`
- AC-5 (REQ-4, REQ-5, REQ-10): two executors that start together on one request leave exactly one live claim and one `cleanup.result`; a run that stopped after the worktree step leaves the branch and the request, and the next run once the lease ended deletes the branch · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^ac5_/)'`
- AC-6 (REQ-11, REQ-12): a worktree another card still claims keeps the request pending, the next close of any card runs it after that claim ended, and a git failure on the branch delete leaves the request pending until the cause is gone · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^ac6_/)'`
- AC-7 (REQ-6, REQ-9): a branch another worktree checks out after the worktree step and a card reopened after its request are kept and reported, and nothing is removed · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^ac7_/)'`
- AC-8 (REQ-12): a problem of the executor (no integration branch in the profile) is reported as the `error` of `cleanup_outbox` while the card stays closed · verify: `cargo nextest run --locked --offline --test card_cleanup_outbox_cli -E 'test(/^ac8_/)'`

## Scope

- write: `.moai/specs/SPEC-20261123-cleanup-outbox/spec.md`, `src/ledger/cleanup_outbox.rs` (new), `src/ledger/mod.rs` (one `pub mod` line), `src/ledger/workflow.rs` (the request in the two close transactions), `src/card/cleanup.rs` (new), `src/card/mod.rs` (one `pub mod` line), `src/card/work.rs`, `src/card/finish.rs` and `src/card/land.rs` (the call after the close), `src/worktree_cleanup.rs` and `src/branch_gc.rs` (what the executor reuses), `tests/card_cleanup_outbox_cli.rs` (new)
- not in this SPEC: `src/lib.rs` (another card holds it), widening the deletion scope to `.claude/worktrees/`, a verdict trigger, a retry command, installing any hook, removing the worktrees and branches that already exist
- read-only: everything else

## Decisions

- D-1: The outbox row is an `event` row, not a table. A new table needs a schema version bump and a migration of every user's ledger; an event row is append-only and has a unique key, which is what a request and its results need. A request is pending while no `cleanup.result` event names its event id.
- D-2: The request names its targets from the card's reservation rows (`worktree`, `branch`) inside the closing transaction, so what is cleaned is what the ledger says the card owned. A target no reservation of the card named is never touched (REQ-9).
- D-3: The request is written by `workflow::done` and by `workflow::finish_naming` when it closes the card, in their own transaction. A landing-only call returns before the close and writes none.
- D-4: Step and result events carry natural keys (`cleanup.step:<request>:<attempt>:<kind>:<target>`), so a replay of the same write adds nothing. A claim is `cleanup.started` with `attempt` and `lease_until` from the DB clock; the default lease is 300 seconds. A run that ends without a result writes `cleanup.released` for its attempt, so the next run needs no wait; a run that dies leaves its lease to end.
- D-5: The worktree step reuses `worktree_cleanup::run` unchanged, so the safe criteria, the preserved `.claude/worktrees/` folder, the fixed slots and the settled-card rules stay the plan's. In this repository a lane worktree under `.claude/worktrees/` is therefore kept with `preserved_location`, and so is its branch, until the user picks the widening of SPEC-20261123-worktree-slots D-8.
- D-6: The branch step reads `branch_gc::plan` again for the exact integration ref and acts on the row of the one branch; the deletion is `branch_gc`'s own `git branch -d` function. A branch the card held under another card's active reservation, a preserved name and a branch another worktree has checked out are `keep` rows there already.
- D-7: A kept target is final (the request ends with a `cleanup.result` of status `settled_with_kept`) when every reason is lasting: dirty, not merged, preserved, checked out elsewhere. A kept target whose reasons all pass by themselves (the `TEMPORARY` list of `worktree_cleanup`) or a failed step keeps the request pending.
- D-8: The executor runs after the close in `card done`, `card finish` and `card land`, and it takes every pending request, so a request whose first run was interrupted is finished by the next close of any card. There is no timer and no retry command in this slice; the leader may add one later.
- D-9: Test seams are an `Options` value of the library function (the lease seconds, a stop after the worktree step, a call before the branch step), never environment variables, so parallel tests cannot disturb one another.
- D-10: The card verdict does not start a cleanup: a verdict leaves the card in `review`, and the executor refuses a card that is not `done`. The function `worktree_cleanup::run` with `Trigger::Verdict` stays as it is.

## Plan

1. This SPEC, then the two test files and the library stubs, so the tests fail first (RED).
2. `src/ledger/cleanup_outbox.rs`: the request, the pending read, the claim, the step and result writes; the request calls in `workflow::done` and `workflow::finish_naming`.
3. `src/card/cleanup.rs`: the executor (re-check, worktree step, branch step, events); `branch_gc` and `worktree_cleanup` expose what it reuses.
4. The calls in `card done`, `card finish` and `card land`, and the `cleanup_outbox` key of their answers.
5. Light checks, one affected run, then the READY note with head and base.

## Risks

- A request that never ends (a claim another card never releases) is retried by every later close and adds one step row per run. The rows are small and append-only; a cap and a retry command are a follow-up.
- The executor deletes things, so the dangerous path (a dirty tree, a link, a checked-out branch, a lock file) is run in a throwaway repository with the real binary before READY.
- Existing tests that count every `event` row after a close see one more row per closed card; they are found by running the card test targets and repaired in this slice.

## Result

Slice 2-a of the worktree-slots card: the outbox rows, the executor and the three close commands. The SPEC is not approved by this text; the numbers below are from runs on a temporary home and temporary repositories, never the real ledger, checkout, worktrees or branches.

- AC-1 met: `ac1_` (3 tests) passed: `card finish`, `card done` and `card land` each write one `cleanup.requested` event naming the reservation's worktree and branch, in the closing transaction; a refusal and a landing that leaves the card open write none.
- AC-2 met: `ac2_` (1 test) passed: a clean merged worktree and its merged branch are removed (`git worktree remove` without force, then `git branch -d`) and one `cleanup.result` of status `settled` is written.
- AC-3 met: `ac3_` (5 tests) passed: a dirty worktree and its branch, a branch the tip does not contain, a preserved name, a branch checked out elsewhere, a worktree under the preserved `.claude/worktrees/` place and a worktree path that is a link (the folder behind the link keeps its file) all stay and are reported with their reasons.
- AC-4 met: `ac4_` (1 test) passed: a second run after a settled request writes no event and changes nothing.
- AC-5 met: `ac5_` (2 tests) passed: two executors started together leave one claim and one result; a run stopped after the worktree step is finished by the next run after its lease ends, with one result.
- AC-6 met: `ac6_` (2 tests) passed: a worktree another card still claims keeps the request pending until a later close; a git failure on the branch delete (a lock file) leaves it pending.
- AC-7 met: `ac7_` (2 tests) passed: a branch another worktree checks out after the worktree step is kept; a card reopened after its request ends as `skipped`.
- AC-8 met: `ac8_` (1 test) passed: with no integration branch the answer carries `cleanup_outbox.error` and the card stays closed.

Other runs: `cargo nextest run --test card_cleanup_outbox_cli` 17 of 17 passed; `card_` / `ledger_` / `worktree_` / `branch_gc` targets (2686 tests) passed except seven `card_finish_cli` tests that counted every `event` row, repaired here and then 45 of 45 passed; `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0; `cargo fmt --check` and `git diff --check` clean.

The first `moai check affected` run on the first head (`dcbe8d8340`, 3298 tests) failed three tests: `state_folder_literal_scan` and `windows_unix_gate_scan` named lines of this slice's test file (a state folder spelled out, an `os::unix` link outside `cfg(unix)`), and `test_area_layout` named the file's first word. The file is now `tests/card_cleanup_outbox_cli.rs` (first word `card`, already in `tests/areas.toml`), reads the profile folder through `state_root::path` and gates the link test with `#[cfg(unix)]`. `test_area_layout` also names `tests/branch_gc_cli.rs` (first word `branch` is not in the table); that file is not this slice's and the table is not in its write list.

Gaps (not part of this slice): `.claude/worktrees/` lane worktrees and their branches stay preserved (widening is the user's pick); a verdict starts no cleanup; there is no retry command or timer, a pending request runs at the next close; `tests/card_finish_cli.rs` was edited (event counts exclude `cleanup.%` rows) though the card's write list did not name it.
