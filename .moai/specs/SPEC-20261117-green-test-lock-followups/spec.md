---
id: SPEC-20261117-green-test-lock-followups
title: "GREEN 시험 잠금 후속: 프로젝트 언어로 좁힌 시험 경로, 기록 없는 열린 로그 보고, 다른 단계 기록"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-green-test-lock
---

## Intent

green-test-lock 감사(#3177)의 비차단 지적 세 가지를 고쳐요. N1: `TEST_PATTERNS`는 언어별 표라고 적었지만 모든 언어 행을 한꺼번에
적용해서, 이 Rust 저장소의 `templates/spec/*.md`도 Ruby `spec/` 규칙으로 시험 경로가 돼요. N2: 기록이 Bash 등으로 지워지면
`moai tdd status`가 아무 말도 하지 않아요. N4: `green`·`refactor`가 아닌 단계의 기록이 잠그지 않는다는 시험이 없어요(감사 mutant M2).

## Requirements

- REQ-1: The test-path check of a project shall apply only the pattern rows of the project's languages, the common row and its declared test paths.
- REQ-2: IF a project names no language in its profile or its root manifests, THEN the test-path check shall apply every pattern row.
  · failure: none
- REQ-3: IF the TDD log's last event is `red` or `green` and the record is missing or unreadable, THEN `moai tdd status` shall report it.
  · failure: none
- REQ-4: IF the record's phase is not `green` or `refactor`, THEN tool-preflight shall lock nothing.
  · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): 이 저장소(python·rust)에서 `templates/spec/*.md`·`spec/*_spec.rb`는 시험 경로가 아니고 `tests/`·`*_test.py`·golden은
  시험 경로예요. 프로필 없는 프로젝트는 루트 manifest로 좁히고(Cargo.toml → rust, Gemfile을 더하면 ruby 행도), manifest가 없으면
  모든 행이에요. `run`의 oracle 목록도 같아요.
  verify: `cargo test --test green_test_lock this_repository_uses_only_its_own_languages`
  verify: `cargo test --test green_test_lock root_manifests_narrow_the_rows_and_none_keeps_them_all`
  verify: `cargo test --test run_verify`
- AC-2 (REQ-3): 기록을 지우거나 깨면 `status`의 `problem`이 마지막 사건(`red`·`green`)을 적고, 잠금은 없어요.
  verify: `cargo test --test green_test_lock status_reports_an_open_log_without_a_record`
- AC-3 (REQ-4): 단계가 `done`·`red`·`GREEN`·빈 값인 기록은 잠그지 않고 `status`에 단계가 없어요. `refactor`는 잠가요.
  verify: `cargo test --test green_test_lock a_record_with_another_phase_locks_nothing`

## Scope

- write: `src/tdd.rs`, `src/run.rs`, `src/project_detect.rs`(`manifest_languages`), `tests/green_test_lock.rs`,
  `docs/design-review/appendix/operations/tdd.md`, `.moai/specs/SPEC-20261117-green-test-lock-followups/spec.md`
- read-only: `.moai-cli/project.yaml`(이 저장소의 감지 언어), `tests/run_verify.rs`
- forbidden: 실제 HOME, `.github/workflows/audit-gate.yml`

## Plan

1. `TestPaths`(언어 행 + common + 선언 경로)와 `manifest_languages`; `run`·`tdd red`·A6가 함께 써요.
2. `status`의 `problem`.
3. 시험, 검사, mutation 3개.

## Decisions

- D1 언어는 프로필 `detected.languages`, 없으면 루트 manifest(파일 확인만, 코드 파일을 훑지 않음: hook이 빨라야 함)예요. 프로필
  `tools`는 자유 글이라 쓰지 않아요.
- D2 둘 다 없으면 모든 행을 적용해요(예전 동작). 모르는 프로젝트에서 잠금과 oracle 경고가 약해지지 않게 하려는 거예요. 그래서
  `tests/run_verify.rs`(README만 있는 임시 프로젝트)의 기대값 4개는 그대로예요. 경로만 받는 `oracle_changes`도 모든 행 그대로예요.
- D3 이 가지는 #3177이 아직 병합되지 않아 그 head `4d73feb53f` 위에 쌓았어요(리더 지시). #3177 뒤에 병합해요.
