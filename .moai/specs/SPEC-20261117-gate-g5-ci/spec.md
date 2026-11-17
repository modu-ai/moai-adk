---
id: SPEC-20261117-gate-g5-ci
title: 관문 G5가 PR head의 GitHub ci 성공을 읽어요 (C-84)
status: draft
revision: 1
created: 2026-11-17
parent: null
---

## Intent

사용자 결정(2026-11-17, 선택 "CI 성공 + PASS")은 병합 조건을 바꿨어요: PR head의 GitHub `ci` 성공이 필요하고,
audit-tier 1·2는 그 실행을 인용한 작성자 아닌 PASS도 필요해요. tier 0은 FAIL이 게시되지 않았을 때 `ci`
성공만으로 병합하고, 게시된 FAIL은 모든 tier에서 병합을 막아요. `local-check:` 줄은 선택 증거이고 `ci`를
대신하지 못해요. 그런데 `moai gate-check`의 G5(`src/gate.rs`)는 아직 PR 댓글의 `local-check: … result=pass`
줄을 요구해서, 그 줄이 없는 PR은 모두 멈춰요. 이 SPEC은 G5를 결정대로 바꾸고 두 모드 파일의 "C-84 전까지"
문장을 정리해요.

## Non-goals

- 새 네트워크 경로: 수집기는 이미 `commits/{head}/check-runs`를 읽어요.
- `tools/local_check.sh` 자체와 그 줄의 형식(선택 증거로 남고, 해석 규칙은 그대로예요).
- 병합 규칙의 다른 규칙(G1–G4, G6–G12)과 tier 판정.
- AGENTS.md, `src/optimize.rs`, `tools/build_folder.py`, `install/**`, `ci.yml`(다른 작업이 고쳐요).

## Requirements

- REQ-1: When `moai gate-check` judges a pull request, G5 shall pass only when a `ci` or `rc-ci` check run on the exact PR head concluded `success`.
- REQ-2 (REQ-1): IF the head has no `ci` run, or its `ci` ran only on an older head or names no head, THEN G5 shall stop the merge with a reason that says so.
- REQ-3 (REQ-1): IF a watched run on the head (`ci`, `rc-ci`, their jobs, `windows-build` other than `skipped`) failed, was cancelled or timed out, or is still running, THEN G5 shall stop the merge and name that run.
- REQ-4: The gate shall treat a `local-check:` line as optional evidence that never decides G5.
- REQ-5: When the collector reads the head's check runs, it shall record for each run the commit GitHub ran it on, or the requested head when GitHub gives none. · failure: none — 응답을 읽지 못하면 수집 전체가 지금처럼 오류로 끝나요.
- REQ-6: When a pull request is audit-tier 0, the gate shall stop it (G3) while any FAIL verdict on any head has no later PASS by the same auditor.
- REQ-7 (REQ-6): IF the authors or the merge base of a tier-0 pull request are unknown, THEN the gate shall stop as before (G9, tier).
- REQ-8: When `moai gate-merge` merges a tier-0 pull request without a `local-check:` line, it shall cite the head's `ci` success in its status and record.
- REQ-9 (REQ-8): IF the merge status or record cannot be written, THEN `gate-merge` shall report it as before and merge nothing more.
- REQ-10: The mode files shall say that G5 reads the exact head's `ci` and that the `local-check:` line is optional, with mode-agent.md no larger than before.

## Acceptance

- AC-1 (REQ-1, REQ-2): 정확한 head의 `ci` 성공은 통과(N), 한 커밋 전 head의 성공·head가 없는 실행·`ci` 없음은 G5에서 멈춰요(N+1).
  verify: `cargo test --locked --offline --test gate_check_cli -- g5_passes_on_an_exact_head_ci_success_without_a_local_check_line g5_needs_the_ci_success_on_this_exact_head`
- AC-2 (REQ-3): failure·cancelled·timed_out·실행 중·작업 실패·windows-build 실패는 G5에서 멈추고, windows-build `skipped`와 상관없는 검사는 통과해요.
  verify: `cargo test --locked --offline --test gate_check_cli -- g5_stops_on_a_failed_cancelled_running_or_missing_ci g5_check_runs_without_the_aggregate_or_with_a_bad_job_stop_it`
- AC-3 (REQ-4): 어떤 형태·순서의 `local-check:` 줄도 G5를 정하지 않아요.
  verify: `cargo test --locked --offline --test gate_check_cli -- a_local_check_line_never_decides_g5 g5_no_local_check_line_decides_it`
- AC-4 (REQ-5): 수집한 실행마다 GitHub의 `head_sha`(없으면 요청한 head)가 남고, 이전 커밋의 성공은 G5에서 멈춰요.
  verify: `cargo test --locked --offline --test gate_collect collected_check_runs_carry_their_head_and_g5_reads_only_this_head`
- AC-5 (REQ-6, REQ-7): tier 0은 어느 head의 FAIL이든 같은 감사자의 나중 PASS 전까지 G3에서 멈춰요.
  verify: `cargo test --locked --offline --test gate_check_cli -- a_posted_fail_stops_a_tier_0_merge_until_the_same_auditor_passes tier0_still_meets_every_other_rule`
- AC-6 (REQ-8, REQ-9): `local-check:` 줄 없는 tier-0 병합이 `ci` 성공을 인용해요.
  verify: `cargo test --locked --offline --test gate_merge`
- AC-7 (REQ-10): 두 모드 파일의 문장과 고정 시험.
  verify: `cargo test --locked --offline --test mode_cards_53`
- AC-8 (REQ-1…REQ-10): 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --test gate_loop`; `cargo test --locked --offline --test gate_stacked_2859`;
  `cargo test --locked --offline --test queue_query_cli`; `cargo test --locked --offline --test mode_workflow_steps`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo test --locked --offline --test git_env_scrub`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --lib gate`; `cargo fmt --check`; `rustfmt --check --edition 2021 src/ops/*.rs`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests.repo_scans_test tests.ci_scope_test tests.source_manifest_test tests.check_pr_self_audit_test`

## Scope

- write: `src/gate.rs`, `src/gate_collect.rs`, `src/gate_merge.rs`, `tests/gate_check_cli.rs`, `tests/gate_collect.rs`,
  `tests/gate_merge.rs`, `tests/gate_loop.rs`, `tests/mode_cards_53.rs`, `skills/moai/workflows/mode-agent.md`,
  `skills/moai/workflows/mode-leader.md`, `skills/moai/workflows/audit.md`, `skills/moai/references/source-manifest.json`,
  `.moai/specs/SPEC-20261117-gate-g5-ci/spec.md`
- read-only: `src/ops/gate_check.rs`, `src/queue_query.rs`, `tools/local_check.sh`
- forbidden: `AGENTS.md`, `src/optimize.rs`, `tools/build_folder.py`, `install/**`, `.github/workflows/ci.yml`, `src/headver.rs`,
  `src/update.rs`, `src/agents.rs`, `src/card*`, `src/ledger*`, 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 시험 먼저(RED): `gate_check_cli`에 새 시험 5개.
2. `src/gate.rs`: `CheckRun.head_sha`, G5를 정확한 head의 `ci`로, tier-0 FAIL 규칙(`g3_tier0`); `src/gate_collect.rs`가 `head_sha`를 채움; `src/gate_merge.rs`는 `local-check:` 줄을 선택으로.
3. 옛 규칙을 고정하던 시험을 새 규칙으로 옮김(이름과 문서에 "was …"를 남김).
4. 두 모드 파일과 audit.md 문장, 고정 시험, manifest, 이웃 검사, mutation 5개 이하.

## Risks

- 손으로 쓴 snapshot 파일에 `head_sha`가 없으면 G5가 멈춰요("names no head") → 수집기는 늘 채우고, 이유가 그대로 보여요.
- 실행 중인 `windows-build`도 G5를 멈춰요(병합 뒤에 도는 경우는 `skipped`로 와야 해요) → 멈춘 이유가 실행 이름을 말해요.
- tier 0의 옛 head FAIL이 이제 병합을 막아요 → 같은 감사자의 나중 PASS로 풀려요(리더 결정의 뜻대로).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 쓰기 범위를 카드 범위(`src/gate.rs`, `src/gate_collect.rs`, `tests/**`) 밖 `src/gate_merge.rs`·두 모드 파일·`audit.md`·manifest로 넓힘 | gate-merge가 `local-check:` 줄이 없으면 거절했고, 모드 파일 문장은 리더 지시 | 해당 파일 되돌리기 |
| D2 | `head_sha`가 없는 실행은 G5에서 "head를 알 수 없음"으로 멈춤; 수집기는 GitHub 값이 없을 때 요청한 head로 채움 | 정확한 head 요구(카드 AC) | 없는 값을 head로 보기 |
| D3 | head의 감시 실행 중 하나라도 실행 중이면 G5가 멈춤(집계가 성공이어도) | 구현 판단(덜 끝난 CI로 병합하지 않기) | 집계 성공이면 통과 |
| D4 | tier 0: 어느 head의 FAIL이든 같은 감사자의 나중 PASS 없이는 G3에서 멈춤 | 리더 결정 2026-11-17("게시된 FAIL은 모든 tier에서 병합을 막음") | 현재 head의 FAIL만 보기 |
| D5 | tier-0 병합의 상태 링크는 `local-check:` 줄이 없으면 PR의 checks 페이지 | 구현 판단 | 줄이 없으면 병합 거절 |
| D6 | 옛 `local-check` G5 시험은 지우지 않고 "G5가 무시함"으로 옮김; 줄 해석은 `src/gate.rs` 단위 시험이 고정 | 고정 시험을 통과용으로 지우지 않기 | 시험 이름·단언 되돌리기 |

## Result

(채우기 전 — 구현 에이전트가 검사 뒤에 채워요.)
