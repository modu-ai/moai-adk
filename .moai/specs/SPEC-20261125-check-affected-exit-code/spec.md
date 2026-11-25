---
id: SPEC-20261125-check-affected-exit-code
title: "check affected exits non-zero on FAIL, stops after a failed step, and takes the fork point as the base"
status: draft
revision: 1
created: 2026-11-25
---

# check affected exits non-zero on FAIL, stops after a failed step, and takes the fork point as the base

## Interview

- Q1: mo.ai.kr 리더의 개선 요청 덧붙임 7 의 F·G·H 를 받을까요? → A: "받아라"(사용자 2026-11-25, 카드 본문). 원문은 `~/.moai/reports/mo-ai-kr/leader-20261125/moai-adk-improvement-request-addendum-7.md`.
- Q2: `--keep-going` 의 기본값과 다른 카드와의 겹침은요? → A: 앞 단계 실패 뒤 멈춤이 기본이고 `--keep-going` 은 선택이에요(카드 본문). 관문을 넓히는 쪽이지 줄이는 쪽이 아니에요(리더 2026-11-25, 리더의 디스패치 메시지).

## Intent

`moai check affected` 는 실패(`result=FAIL`)나 검증 못 함(`result=not_run`)에도 종료 코드 0 과 `ok:true` 를 내서 셸, CI 와 카드 기준이 종료 코드로 판정하지 못하고(F), 정적 검사(clippy)가 실패해도 17분짜리 `test` 단계를 끝까지 돌려 heavy slot 을 그만큼 쥐며(G), 통합 브랜치가 레인 브랜치보다 앞서면 `--card <통합>..HEAD` 영수증의 base 가 head 의 조상이 아니라 READY 에서 거절돼 레인이 `git merge-base` 를 손으로 넘겨요(H).

- 바꾸는 것: `result=PASS` 일 때만 종료 코드 0, 앞 단계가 실패하면 뒤 단계를 돌리지 않고 `skipped` 로 적기(`--keep-going` 은 끝까지), base 가 head 의 조상이 아니면 `git merge-base <base> <head>` 를 base 로 쓰고 원래 base 를 요약에 적기.
- 바꾸지 않는 것: 어느 단계도 약해지거나 사라지지 않아요. 통과 조건과 선택 규칙, 요약 줄의 기존 단어와 순서가 그대로예요. 느슨해지는 곳이 없고 FAIL 이던 것이 0 으로 끝나던 구멍이 막혀요.

## Requirements

<!-- This SPEC widens the check command's gate: FAIL now exits non-zero, and no step is removed or weakened. -->
- REQ-1: When a `moai check affected` run ends with `result=FAIL` or `result=not_run`, the command shall exit with a non-zero status and print the same answer as before. · failure: none — the exit code follows the result word, and a `--list` run has no result and exits 0.
- REQ-2: When a `moai check affected` run ends with `result=PASS`, the command shall exit with status 0. · failure: none — a run that is not PASS is REQ-1.
- REQ-3: When a step of a run fails and `--keep-going` was not given, the run shall start no later step, record each unstarted step as skipped in the answer's `skipped` list, and name the failed step in `stopped_at`.
- REQ-7 (REQ-3): IF the heavy slot was not obtained, THEN the run shall give the slot's exit code to every heavy step and start none, as before.
- REQ-4: When `--keep-going` is given, the run shall run every planned step even after a failure. · failure: none — every step runs and the result is `FAIL` as before.
- REQ-5: When the range's base is not an ancestor of its head, the run shall use `git merge-base <base> <head>` as the base and write the given base as `base_from=<12 hex>` in the summary line and as `base_given` in the answer.
- REQ-8 (REQ-5): IF the base and the head have no merge base (unrelated histories), THEN the run shall be refused as it is today and write no `base_from`.
- REQ-6: When the range's base is an ancestor of its head, the run shall write the summary line exactly as before. · failure: none — nothing about an ancestor base changes.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary repository, a run whose step fails exits non-zero with `result=FAIL` in the answer, a run with nothing declared exits non-zero with `result=not_run`, and a passing run exits 0 · verify: `cargo test --offline --locked --test check_affected_exit_code ac1_`
- AC-2 (REQ-3, REQ-4): a failing clippy step skips the `test` step and lists it in `skipped` with `stopped_at=clippy`, and the same run with `--keep-going` runs `test` too · verify: `cargo test --offline --locked --test check_affected_exit_code ac2_`
- AC-3 (REQ-5, REQ-6): with a base that is ahead of the lane (not an ancestor of the head), the summary names the merge base as `base=` and the given base as `base_from=`, and `card note` accepts the receipt; with an ancestor base the line carries no `base_from` · verify: `cargo test --offline --locked --test check_affected_exit_code ac3_`

## Scope

- write: `.moai/specs/SPEC-20261125-check-affected-exit-code/**`, `src/check_affected.rs`, `src/ops/check.rs`, `src/main.rs`, `src/card/submit.rs`, `tests/check_affected_exit_code.rs`

## Plan

1. RED: `tests/check_affected_exit_code.rs` (AC-1..AC-3) fails on the current behavior.
2. F: the exit code of `check.affected` in `src/main.rs`; the help text of `src/ops/check.rs`; the stale sentence of `src/card/submit.rs`.
3. G: `run_with(.., keep_going)` in `src/check_affected.rs` (`run` keeps its signature and calls it with `false`), the `--keep-going` flag.
4. H: the fork point of the base in `run` and `list`, the `base_from` word and the `base_given` field.
5. `cargo fmt --all`, `moai check affected --card` receipt, READY.
