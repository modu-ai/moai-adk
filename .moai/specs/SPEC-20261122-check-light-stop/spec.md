---
id: SPEC-20261122-check-light-stop
title: "check affected stops before the heavy steps when a cheap step fails, and the build settings stop wasting disk and time"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# check affected stops before the heavy steps when a cheap step fails, and the build settings stop wasting disk and time

## Intent

`moai check affected <base>..<head>` 는 `git diff --check` 나 `cargo fmt --check` 처럼 몇 초면 끝나는 단계가 실패해도 clippy, 빌드, 시험을 끝까지 돌려요. 최근 점검 로그 세 번에서 이렇게 날린 시간이 1.3시간이에요(리더의 빌드·시험 조사, 2026-11-21). 사용자가 2026-11-21 에 고른 결정은 "가벼운 점검이 실패하면 무거운 단계 앞에서 멈춘다" 예요. 이것은 `SPEC-20261120-check-affected` 의 REQ-8("never stop at the first failure")과 REQ-11, AC-4 가운데 가벼운 단계에 해당하는 부분을 뒤집어요. 무거운 단계끼리는 지금처럼 하나가 실패해도 나머지가 돌아요.

"가벼운 단계" 는 계획에서 첫 무거운 단계보다 앞에 오는 단계예요: `git diff --check`, `cargo fmt --all -- --check`, 바뀐 `src/ops/*.rs` 의 `rustfmt --check`, 바뀐 셸 파일의 `bash -n`. 가벼운 단계는 서로 멈추지 않고 모두 돌아요. 몇 초면 끝나고, 사람이 가벼운 실패를 한 번에 다 보게 되니까요.

같은 조사가 짚은 작은 낭비 셋도 이 SPEC 이 바로잡아요. (1) 저장소 설정에 incremental 빌드를 끄는 줄이 없어서 워크트리 빌드 폴더 55개 가운데 52개가 incremental 자료를 들고 있고 각각 약 6.5 GiB 예요(지금은 레인 환경 변수가 있을 때만 꺼져요). (2) `tools/self_check.sh` 의 기본 base 가 515 커밋 뒤처진 `origin/v2026.11.0` 이에요. (3) 한 번의 점검이 `moai-cli` 바이너리를 따로 한 번 더 지어요: 시험 단계가 이미 지은 같은 바이너리를 `build_cli` 단계가 다시 짓는 호출이에요.

## Non-goals

- 시험이 통과했다고 판정하거나 PASS 를 선언하지 않아요: 독립 감사와 CI 는 그대로예요.
- 무거운 단계끼리의 규칙은 바꾸지 않아요: clippy 가 실패해도 시험은 돌아요.
- 느린 lib 시험 두 개와 다섯 개의 늘 고르는 느린 시험 대상은 이 SPEC 이 다루지 않아요(조사의 다른 항목이에요).
- 두 번째 바이너리(`moai`)를 빌드에서 빼지 않아요: 통합 시험이 두 바이너리를 모두 지어야 하는 cargo 규칙이에요.
- 일반 프로젝트용 로컬 검사(`project-check`)의 "모든 단계를 돌린다" 는 그대로예요: 그것은 선언된 목록을 순서대로 보여 주는 명령이고, 이 저장소의 점검이 아니에요.

## Requirements

- REQ-1: When a cheap step exits with a non-zero status, the runner shall run no clippy step, no build step, no test step and no Python step · failure: none — the unrun steps have no log and the result is FAIL.
- REQ-2: When a cheap step exits with a non-zero status, the runner shall still run every other cheap step · failure: none — they take seconds and the report then lists every cheap failure at once.
- REQ-3: When the runner stops before the heavy steps, the summary line shall name the first failed cheap step as `stopped_at=<step>` right before `result=FAIL` · failure: none — a step that did not run has no word in the line.
- REQ-4: When every cheap step exits with 0, the runner shall run every later step even after one of them fails · failure: none — the heavy steps keep the rule of the earlier SPEC.
- REQ-5: When a selected Rust test target makes cargo build the `moai-cli` binary, the planner shall plan no `build_cli` step · failure: none — the Python step reads the binary that the test step built.
- REQ-6: When only Python modules are selected, the planner shall plan one `build_cli` step that builds the `moai-cli` binary alone · failure: none — that step is the only build of the binary in the run.
- REQ-7: The repository cargo config shall turn incremental compilation off · failure: none — an environment variable `CARGO_INCREMENTAL` set by a caller still decides for that run.
- REQ-8: When `SELF_CHECK_BASE` is unset, the self check shall use the local branch `v2026.11.0` as its base · failure: none — an explicit `SELF_CHECK_BASE` still decides.
- REQ-9 (REQ-8): IF the base does not resolve, THEN the self check shall report the diff check and the SPEC check as `not_run` and never as a pass.
- REQ-10: The operation page of `check` shall describe the stop before the heavy steps and the `stopped_at` word · failure: none — a test reads the page.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 가짜 도구에서 `fmt` 가 실패하면 clippy·시험·Python 호출 흔적이 없고, 다른 가벼운 단계(`diff_check`, `ops_rustfmt`)는 돌고, 요약 줄에서 `stopped_at=fmt` 가 `result=FAIL` 바로 앞에 있고, 안 돈 단계의 낱말이 줄에 없어요 · verify: `cargo test --locked --offline --test check_affected_cli a_failed_cheap_step_stops_before_the_heavy_steps`
- AC-2 (REQ-1): 진짜 `git diff --check` 실패(끝 공백), 시작할 수 없는 `rustfmt`(127), 진짜 `bash -n` 실패 각각이 같은 멈춤을 만들고 `stopped_at` 이 그 단계를 가리켜요 · verify: `cargo test --locked --offline --test check_affected_cli every_kind_of_cheap_failure_stops_the_run`
- AC-3 (REQ-4): 가벼운 단계가 모두 0 이면 clippy 가 실패해도 이후 단계가 모두 돌고 `stopped_at` 이 줄에 없어요 · verify: `cargo test --locked --offline --test check_affected_cli a_failed_step_does_not_stop_the_rest_and_the_summary_says_fail`
- AC-4 (REQ-5, REQ-6): Rust 대상과 Python 모듈이 함께 골라지면 계획에 `build_cli` 가 없고, Python 만 골라지면 `--bin moai-cli` 하나만 짓는 `build_cli` 가 하나 있어요 · verify: `cargo test --locked --offline --lib check_affected::tests::the_binary_is_built_by_one_step_at_most`
- AC-5 (REQ-7): 저장소의 `.cargo/config.toml` 을 TOML 로 읽으면 `build.incremental` 이 `false` 예요 · verify: `cargo test --locked --offline --test check_pipeline_config the_cargo_config_turns_incremental_off`
- AC-6 (REQ-8): 로컬 가지 `v2026.11.0` 만 있고 origin 이 없는 임시 저장소에서 `SELF_CHECK_BASE` 없이 돌린 자체 점검이 `diff-check` 를 실제로 돌려요(`not_run` 아님) · verify: `cargo test --locked --offline --test check_pipeline_config the_self_check_base_is_the_local_integration_branch`
- AC-7 (REQ-10): `check` 의 연산 페이지가 `stopped_at` 과 멈춤 규칙을 적어요 · verify: `cargo test --locked --offline --test check_pipeline_config the_operation_page_describes_the_stop`

## Scope

- write: `.moai/specs/SPEC-20261122-check-light-stop/**`, `src/check_affected.rs`, `tests/check_affected_cli.rs`, `tests/check_pipeline_config.rs`(새), `tools/self_check.sh`, `.cargo/config.toml`, `docs/design-review/appendix/operations/check.md`
- read-only: `src/ops/check.rs`, `tools/local_check.sh`, `tests/self_check_test.py`, `tests/cargo_env_developer_dir.rs`, `SPEC-20261120-check-affected`
- forbidden: `tools/local_check.sh` 의 기본 base 변경(다른 점검이고 따로 정해요), 새 Python 시험, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME, 다른 레인의 파일

## Plan

1. SPEC 과 빨간 시험: AC-1 에서 AC-3 의 CLI 시험, AC-4 의 단위 시험, AC-5 에서 AC-7 의 새 시험 파일. 코드 전에 모두 실패하는 것을 확인해요.
2. `src/check_affected.rs`: 가벼운 묶음이 끝난 뒤 실패가 있으면 무거운 단계와 Python 을 돌리지 않고 `stopped_at` 을 요약에 넣어요; `steps` 가 시험 대상이 있으면 `build_cli` 를 빼요.
3. `.cargo/config.toml`, `tools/self_check.sh` 의 기본 base 와 머리글 주석, `check.md` 페이지.
4. 영향 시험 한 번, READY.

## Risks

- 요약 줄을 읽는 다른 코드가 없는지 저장소에서 찾았고(`result=PASS`, `check-affected base=` 검색) 연산 페이지와 이 명령 자신의 시험만 읽어요. 그래도 `stopped_at` 는 `result=` 바로 앞의 새 낱말이라서 단계 이름 낱말의 순서와 개수를 읽는 외부 도구가 있다면 거기서 깨질 수 있어요.
- `.cargo/config.toml` 의 `build.incremental = false` 는 이미 만들어진 incremental 폴더를 지우지 않아요. 빌드 폴더 정리는 사용자의 선택이 필요한 별도 일이에요.
- 통합 가지 이름이 `next` 로 바뀔 예정이에요(리더 결정 2026-11-22, 한 번만). 그때 `tools/self_check.sh` 의 기본값 한 곳과 이 SPEC 의 REQ-8 이 같이 바뀌어야 해요.
- `build_cli` 를 빼는 근거는 cargo 문서의 규칙(통합 시험이 선택되면 바이너리 대상이 자동으로 지어진다)이에요. 시험은 계획만 확인하고, 실제 빌드가 한 번인지는 영향 시험 로그에서 `Compiling moai-cli` 줄 수로 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 가벼운 단계는 서로 멈추지 않고 모두 돌고, 그 묶음 뒤에 무거운 단계 앞에서 멈춰요. | 가벼운 단계는 몇 초이고, 실패를 한 번에 다 보여 주면 다시 돌리는 횟수가 줄어요. | 카드 열림 |
| D-2 | 요약에는 안 돈 단계의 낱말을 넣지 않고 `stopped_at` 한 낱말만 더해요. | 안 돈 단계에 종료 코드를 지어내면 0 이나 127 처럼 읽혀요. | 카드 열림 |
| D-3 | `build_cli` 는 Rust 시험 대상이 없을 때만 계획해요. | cargo 가 통합 시험을 위해 같은 바이너리를 같은 폴더에 이미 지어요. | 카드 열림 |
| D-4 | `self_check` 의 기본 base 는 로컬 가지 이름 한 곳에서 정하고, 가지 이름이 바뀔 때 그 한 곳을 고쳐요. | 오래된 origin 가지로 점검하고 있었어요(515 커밋 뒤). | 카드 열림 |

## Result

(status draft — written when the slice lands)
