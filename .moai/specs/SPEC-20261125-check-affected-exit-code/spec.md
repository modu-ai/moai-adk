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

## Result

구현은 `src/main.rs`(`check.affected` 답의 `result` 가 PASS 가 아니면 종료 코드 1), `src/ops/check.rs`(도움말과 `--keep-going` 플래그), `src/check_affected.rs`(`run_with(.., keep_going)`; `run` 은 시그니처 그대로 `false` 로 호출해요. 앞 단계가 실패하면 뒤 단계를 멈추는 규칙, 답의 `skipped` 목록, `fork_point`·`with_base_from`·`first_failed`), `src/card/submit.rs`(거절 문구의 낡은 한 문장)에 있어요. 기본값은 첫 실패 단계 뒤의 첫 단계부터 시작하지 않는 것이고(값싼 단계는 모두 먼저 돌아요), 얻지 못한 heavy slot 의 종료 코드는 모든 heavy 단계에 그대로 가요. `RunOptions` 와 `SummaryParts` 는 그대로라서 요약 단어의 순서가 유지되고 `base_from=<12 hex>` 는 마지막 `result=` 바로 앞에 들어가요. 커밋은 `6924903638`(SPEC), `8018546536`(실패하는 테스트), `a85478034d`(종료 코드), `3f2a0536ea`(멈춤, `--keep-going`, 분기점 base), `ef068085da`(옛 테스트 보정과 기본 멈춤 테스트)예요.

- AC-1 met: `tests/check_affected_exit_code.rs` 의 `ac1_a_failed_step_exits_non_zero_with_result_fail`, `ac1_a_failed_declared_check_exits_non_zero_with_result_fail`, `ac1_nothing_declared_exits_non_zero_with_result_not_run`, `ac1_a_passing_run_exits_zero`, `ac1_list_exits_zero` 가 통과해요(실패는 종료 코드 1 과 `result=FAIL`, 선언 없음은 `result=not_run`, 통과와 `--list` 는 0).
- AC-2 met: `ac2_a_failed_clippy_step_skips_the_test_step`, `ac2_keep_going_runs_the_test_step_after_a_failed_clippy_step`, `ac2_a_run_without_a_failure_skips_nothing`, `ac2_a_refused_slot_still_gives_its_code_to_every_heavy_step` 가 통과하고, `tests/check_affected_non_rust.rs` 의 `a_failed_declared_check_stops_the_next_check_by_default` 와 `a_failed_declared_check_fails_the_run`(`--keep-going`)도 통과해요.
- AC-3 met: `ac3_a_base_ahead_of_the_lane_is_replaced_by_the_fork_point`, `ac3_an_ancestor_base_writes_the_line_as_before`, `ac3_unrelated_histories_are_refused_as_before` 가 통과해요(분기점이 base 로 바뀌는 것, 조상 base 의 줄이 전과 같은 것, 분기점이 없는 기록의 거절).
- 함께 돌린 것(슬롯 실행, 모두 exit 0): `check_affected_exit_code` 12개(`ac1_` 5, `ac2_` 4, `ac3_` 3), `check_affected_cli` 19개, `check_affected_non_rust` 7개, `state_folder_literal_scan` 10개, `--lib check_affected` 63개, 같은 작업 중 앞서 `check_affected_one_slot` 7, `check_affected_tables_cli` 5, `check_affected_scans_cli` 5, `check_affected_touch` 4, `check_affected_tmpdir` 8, `check_affected_card_areas` 9, `check_affected_registry_add` 5, `check_affected_modes_cli` 13, `check_affected_card_select` 8, `check_pipeline_config` 3개가 통과했어요. `cargo clippy --locked --offline --all-targets -- -D warnings` 도 exit 0 이었어요.
- 가드 뮤턴트: none. 이 변경은 tier 2 바닥 목록의 경로를 건드리지 않아서 요구되지 않고, 뮤턴트는 돌리지 않았어요.
- RED: 테스트 커밋(`8018546536`)에서 `cargo test --offline --locked --test check_affected_exit_code` 가 7개 실패, 5개 통과였어요. 실패한 7개는 `ac1_` 종료 코드 테스트 3개, `ac2_` 2개, `ac3_` 2개예요. 분기점이 없는 기록(unrelated histories) 테스트는 이후 "전과 같이 거절"로 문구를 바꿔서 옛 코드에서도 통과하고, 가드로 남겼어요.
- 범위: 카드의 쓰기 범위를 한 번 넓혀(장부 이벤트에 이유를 남김) `tests/check_affected_cli.rs`, `tests/check_affected_non_rust.rs` 와 `tests/fixtures/state_folder_literals/allow.tsv` 한 줄을 더했어요. 새 멈춤 규칙이 옛 테스트 둘을 틀리게 만들었고(둘은 `--keep-going` 을 받게 고침), 새 테스트 파일이 상태 폴더 이름을 두 번 쓰기 때문이에요. 위 Scope 절의 목록은 그대로 두었어요. 범위 밖 수정은 없어요.
- 남은 틈: `src/card/submit.rs` 의 `read_facts` 는 아직 주어진 base 를 받아요(submit 은 병합 기준점 자체를 넘기므로 깨지지 않아요).
- Not run: 전체 `cargo test`, `--bin moai`·`--bin moai-cli` 테스트, `check affected` 를 부르는 다른 테스트 파일(`tests/card/card_submit_cli.rs`, `card_ready_receipt_cli.rs`, `card_land_batch_cli.rs`, `heavy_slot_fifo_cli.rs`, `product_cli.rs`, `project_interview_minimal.rs`), `tools/self_check.sh`, Windows·Linux. `moai check affected --card` 영수증은 이 글 뒤 READY 에서 레인이 돌려요.
