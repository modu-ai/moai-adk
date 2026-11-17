---
id: SPEC-20261117-ci-superseded-neutral
title: "병합 뒤 ci가 밀려난 macOS 대기 작업 때문에 빨갛게 되지 않게"
status: draft
revision: 1
created: 2026-11-17
---

## Intent

통합 가지에 push가 오면 `ci.yml`의 macOS 대상은 실행 사이에 concurrency 그룹 하나(`ci-push-<target>`)를 나눠 써서, 새
병합이 아직 기다리는 macOS 작업을 대신해요(#2554 D8 (a)). 밀려난 작업은 `cancelled`로 끝나고 matrix 결과가
`cancelled`가 되어, 실패한 시험이 없는데도 `ci` 집계가 빨갛게 됐어요. 2026-11-17 전수 조사 I6에서 push 실행 80개 중
40개가 이것뿐이었어요(예: 실행 43728008190, 작업 132861325600 `check: cancelled`). CI 재설계 D6("밀려난 취소는 중립")에
따라 그런 실행의 `ci`는 건너뜀(중립)으로 보이고, 진짜 실패·시간 초과·수동 취소는 그대로 빨강이어야 해요.

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: When a push run's matrix result is `cancelled` and every target passed except macOS targets replaced while waiting by their own push concurrency group, the `ci` aggregate shall be skipped instead of failing.
- REQ-2 (REQ-1): IF any target failed, timed out, was cancelled for another reason, or was cancelled after it started, THEN the `ci` aggregate shall fail.
- REQ-3 (REQ-1): IF the jobs of the run cannot be read, THEN the `ci` aggregate shall fail.
- REQ-4: The pull request behaviour of the `ci` check shall stay unchanged.
- REQ-5: The superseded decision shall live in a repository script that the workflow only calls and that tests check with recorded job shapes.
- REQ-6: Every action the change adds shall be pinned to a commit SHA.
- REQ-7: The changed workflow shall pass actionlint.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-5): 기록한 실행 43728008190(작업과 annotation)은 `superseded`이고, 그 변형(대상 실패, 시간
  초과·수동·다른 그룹 문장, 시작한 작업, linux 대상, push 아닌 이벤트, `cancelled` 아닌 결과, 읽지 못한 입력)은 모두 `fail`이에요.
  verify: `python3 -m unittest tests/ci_superseded_test.py`
- AC-2 (REQ-1, REQ-4, REQ-6, REQ-7): `ci-verdict`는 push이고 matrix 결과가 `cancelled`일 때만 돌고 읽기 권한만 가져요. `ci`는
  `ci-verdict`가 `superseded`를 쓸 때만 건너뛰어요. 이름 `ci`와 PR 동작은 그대로예요.
  verify: `python3 -m unittest tests/github_assets_test.py`; `actionlint .github/workflows/ci.yml`
- AC-3 (REQ-1…REQ-7): ci.yml을 읽는 다른 시험과 바뀐 곳에 닿는 검사가 통과해요.
  verify: `python3 -m unittest tests/local_check_test.py tests/ci_install_ripgrep_test.py tests/ci_scope_test.py tests/repo_scans_test.py`;
  `cargo test --locked --offline --test gate_check_cli`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `moai spec check --all --project .`

## Scope

- write: `.github/workflows/ci.yml`, `tools/ci_superseded.py`, `tests/ci_superseded_test.py`,
  `tests/fixtures/ci_superseded/run-43728008190.json`, `tests/github_assets_test.py`,
  `.moai/specs/SPEC-20261117-ci-superseded-neutral/spec.md`
- read-only: `src/flow.rs`, `src/gate_collect.rs`, `.github/workflows/candidate.yml`, `tools/candidate_summary.py`
- forbidden: `.github/workflows/audit-gate.yml`(다른 가지), 저장소 규칙 세트·필수 검사 설정, 네트워크 쓰기

## Plan

1. 실제 실행 기록: `gh api`로 실행 43728008190의 작업 목록과 밀려난 작업의 annotation을 읽어 시험 자료로 둬요(읽기만).
2. 시험 먼저: `tests/ci_superseded_test.py`(기록 실행과 변형 9가지, 명령줄 출력).
3. `tools/ci_superseded.py`: 순수 판정 함수 + `gh api`로 읽는 얇은 부분.
4. `ci.yml`: `ci-verdict` 작업과 `ci`의 `needs`/`if`. `tests/github_assets_test.py`의 고정을 새 모양으로.
5. 검사, 커밋, mutation.

## Decisions

- D1 "밀려남"의 근거는 GitHub가 그 작업의 check run에 남기는 annotation 문장 "Canceling since a higher priority waiting
  request for ci-push-<target> exists"와 정확히 같을 때만이에요(실행 43728008190에서 읽음). 문장이 바뀌면 `fail`이 되어
  빨강으로 남아요(숨기지 않는 쪽).
- D2 그 밖에 작업이 시작하지 않았어야 해요: 단계 0개, runner 이름 없음. macOS 대상만(그 그룹을 나누는 것이 macOS뿐).
- D3 "중립"은 `ci`를 건너뛰는 것으로 나타내요. 작업은 neutral로 끝날 수 없고, 성공으로 보이게 하면 정직하지 않아요.
  건너뜀은 PR에서 필수 검사를 통과로 셀 수 있으므로 `ci-verdict`는 push에서만 돌아요.
- D4 PR 실행에는 `ci-verdict`가 건너뜀으로 하나 더 보여요. `ci`의 판정과 이름은 그대로예요.
- D5 수동 취소는 실행 전체가 취소되어 `ci`가 지금처럼 돌지 않아요(`!cancelled()`). 실행 안에서 그 밖의 이유로 취소된
  대상은 `fail`이에요.
- D6 G8 tip 판정(`src/gate_collect.rs` `d8a_judgment`)은 이미 집계 검사를 보지 않고 밀려난 작업을 빼서 읽으므로 바꾸지
  않았어요.

## Result

구현 에이전트가 2026-11-17에 채웠어요. SYNC와 AUDIT가 다시 확인해요. 기준 `d032d0be1f`(origin/v2026.11.0), 변경 커밋
`334ae7d42a`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, cargo는
`moai heavy-slot-acquire` 슬롯 안에서 한 번에 하나. 실제 GitHub 실행은 이 가지가 push되어야 볼 수 있어 아직 없어요.

- AC-1: met — `tests/ci_superseded_test.py` OK(10). 스크립트가 없을 때의 첫 실행은 import 오류(TOOL_FAILURE)였어요.
  기록 자료는 `gh api`로 읽은 실행 43728008190의 작업 목록과 대조해 id·이름·상태·결론·runner·단계가 같음을 확인했어요.
- AC-2: met — `tests/github_assets_test.py` OK(새 시험 `test_rc_ci_reads_a_superseded_push_run_as_neutral_only_through_the_verdict`
  포함); `actionlint .github/workflows/ci.yml` exit 0(1.7.10).
- AC-3: met — `local_check_test`·`ci_install_ripgrep_test`·`ci_scope_test`·`repo_scans_test`와 위 둘 합쳐 147개 OK;
  `gate_check_cli` 0(45); `product_cli` 0(75); `--bin moai` 0(27); `cargo fmt --check` 0; `moai spec check --all` 0(SPEC 10개,
  errors 0, warnings 0); `git diff --check` 0.
- mutation 3개(커밋 뒤, 각각 `git restore` 뒤 tree 깨끗): "시작 안 함(단계 0)" 조건 삭제 → 1 FAIL; push 조건 삭제 → 4 FAIL;
  `ci-verdict`의 `github.event_name == 'push'` 삭제 → 1 FAIL. 되돌린 뒤 OK.

실행 안 함: 실제 push 실행에서의 확인(병합 뒤 다음 밀려남 때), 전체 `cargo test`, clippy(Rust 변경 없음).
