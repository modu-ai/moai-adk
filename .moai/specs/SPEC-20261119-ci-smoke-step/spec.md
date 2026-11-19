---
id: SPEC-20261119-ci-smoke-step
title: "The CI check job runs the smoke harness for install and update after the check set"
status: draft
revision: 1
created: 2026-11-19
---

# The CI check job runs the smoke harness for install and update after the check set

## Intent

카드 C-101의 세 번째 기준 중 CI 쪽이에요: "CI가 install과 update 영역을 PR에서 돌린다". harness(`tools/smoke/run.sh`, SPEC-20261119-smoke-harness)는 병합됐고, VERIFY 글이 harness를 첫 점검으로 부르는 문단은 SPEC-20261119-smoke-verify-step이 더해요. 빠진 것은 CI가 같은 harness를 돌리는 단계예요. 레인이 손으로만 돌리면 돌렸다는 말이 증거가 못 돼요.

이 SPEC은 `.github/workflows/ci.yml`의 `check` 잡에 `Check set` 단계 바로 뒤에 한 단계를 더해요. 그 단계는 `check` 잡이 방금 빌드한 디버그 바이너리로 `install`과 `update` 영역을 돌리고, 영역마다 마지막 줄을 잡 요약에 적고, 하나라도 실패하면 잡을 실패시켜요. `CI_SCOPE`가 `full`일 때만 돌아요(문서만 바뀐 실행은 가벼운 CI라 건너뛰어요).

harness는 임시 HOME에서만 돌고 실제 HOME을 읽지 않아요. 영역을 "건드린 PR"만 골라 돌리는 대신 `full` 범위의 모든 실행에서 돌려요: harness가 영역당 몇 초라서 경로 판정 규칙을 새로 만드는 비용(과 틀릴 위험)이 더 커요(D-1).

## Non-goals

- `launcher`·`project` 영역을 CI에서 돌리기(카드는 install·update만 말해요; launcher는 pty가 필요해요).
- harness 코드(`tools/smoke/**`)와 `tools/local_check.sh`의 단계 목록 바꾸기(그 목록은 `tests/local_check_test.py`가 고정해요).
- Windows(`windows-build` 잡)에서 돌리기.

## Requirements

- REQ-1: The `check` job of `ci.yml` shall run `tools/smoke/run.sh` for the areas `install` and `update` in a step placed after the `Check set` step.
- REQ-2: The smoke step shall run only when `CI_SCOPE` is `full`.
- REQ-3: The smoke step shall run the debug binary of this job (`${CARGO_TARGET_DIR:-target}/debug/moai-cli`, named with `--bin`) after `cargo build --offline --locked --bin moai-cli` has made sure it exists.
- REQ-4: The smoke step shall append one line per area to the job summary that carries the exit code and the last line of the harness output.
- REQ-5 (REQ-1): IF an area ends with a non-zero exit, THEN the smoke step shall still run the other area and then fail the job.
- REQ-6: The `ci.yml` test shall read the workflow text and fail with the name of each part that is missing: the step, its place after `Check set`, the `full` condition, the build command, `--bin`, an area, the summary line, or the failing exit.
- REQ-7 (REQ-6): IF the workflow text lacks the step, THEN the `ci.yml` test shall fail with `smoke step missing` and no other claim.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): `ci.yml`에 단계가 있고, 시험이 그 모양(자리·조건·빌드·`--bin`·영역 둘·요약·실패 종료)을 읽어요. 상자(fixture) 글에서 부분을 하나씩 빼면 시험이 그 부분의 이름으로 실패해요.
  verify: `python3 -m unittest tests.ci_smoke_step_test`
- AC-2 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): `ci.yml`의 실제 글이 그 모양을 갖고, 단계의 스크립트가 맥과 Linux에서 실제로 돌아요(Result에 실행 줄). 실제 CI 러너에서의 실행은 병합 전에는 읽을 수 없어서 `not_run`으로 적어요.
  verify: `python3 -m unittest tests.ci_smoke_step_test.RealWorkflowTest`
- AC-3: 이웃 검사.
  verify: `python3 -m unittest tests.github_assets_test`; `python3 -m unittest tests.ci_scope_test`; `python3 -m unittest tests.local_check_test`

## Scope

- write: `.moai/specs/SPEC-20261119-ci-smoke-step/spec.md`, `.github/workflows/ci.yml`, `tests/ci_smoke_step_test.py`
- forbidden: 실제 HOME 쓰기, `tools/smoke/**`, `tools/local_check.sh`, `tests/local_check_test.py`, `tests/github_assets_test.py`(C-97 예약), 위 목록 밖의 `.github/workflows/**`

## Plan

1. SPEC을 쓰고 시험을 먼저 더해요(RED): 단계가 아직 없어서 `smoke step missing`으로 실패해야 해요.
2. harness가 Linux에서 되는지 VM에서 읽어요(`install`·`update`).
3. `ci.yml`에 단계를 더해요(GREEN).
4. 상자 변이(단계 부분 빼기)와 이웃 검사를 돌려요.
5. `.github/workflows/**`를 건드려서 tier 2: 카드에 `감사 요청 tier 2 head=<sha>`를 적어 리더 감사를 받아요.

## Risks

- `cargo nextest run`이 `moai-cli` 바이너리를 `target/debug`에 만들지 않으면 harness가 종료 2로 끝나요 → 단계가 `cargo build --offline --locked --bin moai-cli`를 먼저 돌려 바이너리를 보장해요(이미 있으면 몇 초).
- harness가 CI의 Linux나 macOS에서 다르게 동작해요 → 로컬 Linux VM에서 두 영역을 읽어요(AC-2); macOS는 로컬 맥에서 같은 harness를 돌린 결과(SPEC-20261119-smoke-harness)가 있어요.
- 단계가 `check` 잡을 느리게 해요 → 영역당 몇 초(`all`이 약 10초)라 허용해요.
- `ci.yml`의 다른 시험(`tests/github_assets_test.py`)이 단계 목록을 이름으로 찾아요 → 새 단계는 이름으로 추가하고 기존 이름·순서를 바꾸지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | install·update를 `full` 범위의 모든 실행에서 돌린다(경로로 고르지 않는다) | 영역당 몇 초라서 경로 판정 규칙을 새로 만들 이유가 없고, 카드 문구(영역을 건드린 PR에서)를 포함해요 | 단계를 지우면 돼요 |
| D-2 | 단계는 `ci.yml`에 둔다(`tools/local_check.sh`가 아니라) | local_check의 단계 목록은 `tests/local_check_test.py`의 시험 36개(`Ran 36 tests`)가 읽고 그중 `CI_LINUX_STEPS` 한 줄이 목록을 고정하고, 그 목록은 레인의 로컬 점검과 CI가 같아야 해서 바꾸면 영향이 커요 | `ci.yml`의 단계를 지우면 돼요 |
| D-3 | 바이너리를 `cargo build --bin moai-cli`로 보장한다 | nextest가 바이너리를 `target/debug`에 남기는지는 이 변경이 의존할 일이 아니에요 | 줄을 지우면 돼요 |

## Result

- AC-1: met — `python3 -m unittest tests.ci_smoke_step_test`: 시험 먼저(RED, d6ae822556) `FAILED (failures=2)`: 실제 워크플로 시험 둘이 `Lists differ: ['smoke step missing'] != []` 와 단계 이름 목록에 `Smoke harness (install, update)` 없음으로 실패하고, 모양 시험 7개는 통과했어요. 단계를 더한 뒤(GREEN, 98351d78f9) `Ran 9 tests … OK`. 시험은 단계의 자리(`Check set` 뒤, `check` 잡 안)·`full` 범위 조건·빌드가 하니스 앞·`--bin`·영역 둘·요약 줄·실패 종료를 각각 이름으로 읽고, 상자 글에서 부분을 하나씩 빼면 그 이름으로 실패해요.
- AC-2: met — 맥(darwin arm64)에서 `ci.yml` 단계의 `run` 본문을 yaml 로 꺼내 `bash -eo pipefail` 로 그대로 실행(`GITHUB_STEP_SUMMARY` 는 임시 파일): `smoke: install pass steps=4 time=1.6s`, `smoke: update pass steps=4 time=9.0s`, 종료 0, 요약 두 줄(`- smoke install: exit 0 · …`). 바이너리 없는 경로(빌드 줄을 빼고 `bin=/nonexistent/moai-cli`): 두 영역이 모두 돌아 각각 종료 2(`smoke: binary not found: …`), 단계 종료 1, 요약 두 줄. Linux: colima 프로필 `etxtbsy`(Linux aarch64, git 2.43.0, Python 3.12.3) VM 에서 d6ae822556 의 `git archive` 를 풀어 `cargo build --locked --bin moai-cli`(3m36s) 뒤 `tools/smoke/run.sh install --bin target/debug/moai-cli` → `smoke: install pass steps=4 time=2.2s`, `update` → `smoke: update pass steps=4 time=13.8s`, 둘 다 종료 0. `python3 -m unittest tests.ci_smoke_step_test.RealWorkflowTest` 2개 OK. 실제 CI 러너(linux-x86_64, macos-arm64)에서의 실행은 병합 전에는 읽을 수 없어서 `not_run`: 묶음 PR 의 rc-ci 가 확인하고, VM 은 aarch64 라 x86_64 는 읽지 못했어요. `cargo nextest run` 이 바이너리를 `target/debug` 에 남기는지는 읽지 않았고, 단계가 `cargo build --bin moai-cli` 로 직접 보장해요(D-3).
- AC-3: met — 이웃 검사(작업 트리에서 `CARGO_TARGET_DIR` 를 이 작업 트리의 빌드 폴더로 줌): `python3 -m unittest tests.github_assets_test` `Ran 39 tests … OK`; `tests.ci_scope_test` `Ran 40 … OK`; `tests.local_check_test` `Ran 36 … OK`; `tests.repo_scans_test` `Ran 25 … OK`; `tests.ci_install_ripgrep_test` `Ran 20 … OK`; `cargo fmt --check` 출력 없음. clippy 와 Rust 시험은 이 변경이 Rust 코드를 바꾸지 않아 돌리지 않았어요(`not_run`).

변이(커밋 뒤 한 곳씩 바꾸고 `python3 -m unittest tests.ci_smoke_step_test`, 변이마다 `git checkout --` 뒤 작업 트리 clean): 워크플로 줄 — M1 `if: ${{ env.CI_SCOPE == 'full' }}` 삭제 → KILLED(`smoke step lacks the full-scope condition`); M2 `for area in install update` → `install` 만 → KILLED(`… lacks the areas install and update`); M3 `|| code=$?` 삭제 → KILLED(`… lacks the failing exit`); M4 빌드 줄 삭제 → KILLED(`… lacks the build of moai-cli before the harness`); M5 `|| failed=1` → `|| true` → KILLED(`… lacks the failing exit`); M6 `--bin "$bin"` 삭제 → KILLED(`smoke step lacks --bin`); M7 `>> "$GITHUB_STEP_SUMMARY"` → 다른 파일 → KILLED(`… lacks the summary line`). 시험 코드 — M8 빌드 순서 검사 약화 → KILLED(`test_the_build_after_the_harness_is_refused`); M9 `Check set` 뒤 검사 약화 → KILLED(`test_the_step_before_check_set_is_refused`). 9/9 KILLED.

tier 2: `.github/workflows/ci.yml` 를 바꿔서 카드 노트 `감사 요청 tier 2 head=<sha>` 로 리더 감사를 받아요.
