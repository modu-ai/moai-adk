---
id: SPEC-20261125-slow-test-top2
title: "느린 테스트 순위 1·2위를 빠르게"
status: draft
revision: 1
created: 2026-11-25
---

# 느린 테스트 순위 1·2위를 빠르게

## Intent

`SPEC-20261125-slow-test-top` 의 순위 표(24시간, 로그 71개)에서 1위 `go_state_cli::a_receipt_above_256_kib_reverts_completely` 와 2위 `worktree_gc::tests::the_recheck_before_a_removal_refuses_files_and_an_install_folder` 를 같은 단언으로 빠르게 해요. 원인은 카드의 후보일 뿐이라서 먼저 `SPEC-20261124-test-cost-profile-git-path` 의 방법(번갈아 5회, 중앙값, 부하 기록)으로 재고, 잰 원인만 고쳐요.

## Requirements

- REQ-1: While the Result holds the cost of a ranked test, each row shall give the median of five alternating runs of the baseline binary and the repaired binary, with the load of each run.
- REQ-2: While `the_recheck_before_a_removal_refuses_files_and_an_install_folder` judges files and install folders and no process, its re-checks shall read an empty process table that belongs to the thread running the test.
- REQ-3: While one test thread holds an empty process table, a check on another thread of the same process shall read the table of its own thread, the real one unless that thread set one.
- REQ-4: While `go-state-apply` or `go-state-revert` runs in a debug build with `MOAI_TEST_NO_FULL_SYNC` set, the receipt save shall skip the file and the directory sync.
- REQ-5: While a release build runs `go-state-apply` or `go-state-revert` with `MOAI_TEST_NO_FULL_SYNC` set, the receipt save shall keep both syncs.
- REQ-6: While `a_receipt_above_256_kib_reverts_completely` starts the commands that save the receipt, the test shall give them `MOAI_TEST_NO_FULL_SYNC`.
- REQ-7: While a test file or a unit test in the write scope is changed, it shall keep every assertion it had.

## Acceptance

- AC-1 (REQ-1): the Result gives the cost split of both tests as tables with the data they came from · verify: `moai spec check SPEC-20261125-slow-test-top2 --project .`
- AC-2 (REQ-2, REQ-3, REQ-7): the `worktree_gc` unit tests pass with the same names, and the guard tests of the thread table pass · verify: `cargo test --locked --offline --lib worktree_gc::tests`
- AC-3 (REQ-4, REQ-5, REQ-6, REQ-7): `go_state_cli` passes with the same test names and the 256 KiB test runs in seconds · verify: `cargo test --locked --offline --test go_state_cli`
- AC-4 (REQ-7): no assertion line was removed from the changed test files and the unit-test module · verify: `test -z "$(git diff -U0 NeXT...HEAD -- tests/go_state_cli.rs src/worktree_gc.rs | grep -E '^-[^-].*assert')"`
- AC-5 (REQ-2, REQ-3, REQ-5, REQ-6): each new guard is killed by a mutant that removes the guarded line (the Result lists each `guard-mutant:` line) · verify: `test "$(grep -c '^- guard-mutant:' .moai/specs/SPEC-20261125-slow-test-top2/spec.md)" -ge 2`

## Scope

- write: `.moai/specs/SPEC-20261125-slow-test-top2/**`
- write: `tests/go_state_cli.rs`, `src/worktree_gc.rs`
- write (line lease, after the leader's answer to the scope request on the card): `src/apply.rs`, the two sync calls of `stage` and `sync_parent`
- read-only: everything else; a further path waits for a SCOPE-REQUEST note on the card

## Plan

1. 두 테스트의 원인을 번갈아 5회(1위는 단계별로 시간을 찍은 사본으로)로 잰다.
2. 2위: 단위 테스트 모듈에 스레드별 빈 프로세스 표를 두고 그 테스트만 쓴다.
3. 1위: 리더의 결정(노트)에 따라 디버그 전용 스위치로 영수증 저장의 동기화를 건너뛴다.
4. 같은 단언을 둔 채 가드는 뮤턴트로 확인하고, 번갈아 5회로 전후를 비교한다.

## Result

(SYNC 에서 채워요.)
