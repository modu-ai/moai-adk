---
id: SPEC-20261119-smoke-verify-step
title: "The VERIFY workflow names the smoke harness as the first check before commit"
status: draft
revision: 1
created: 2026-11-19
---

# The VERIFY workflow names the smoke harness as the first check before commit

## Intent

카드 C-101의 두 번째 기준이에요: "VERIFY 워크플로가 harness를 그것이 다루는 영역의 첫 점검으로 이름 짓는다". harness(`tools/smoke/run.sh`, SPEC-20261119-smoke-harness)는 PR #3226으로 병합됐고, 실제 HOME을 읽지 않는다는 보호 시험도 거기에 있어요. 빠진 것은 `skills/moai/workflows/verify.md`가 그 harness를 말하는 한 문단이에요. 글에 없으면 레인과 감사자가 harness를 모르고 지나쳐요.

이 SPEC은 `verify.md`에 한 문단을 더하고, 그 문장을 시험으로 고정하고, 소스 manifest를 맞춰요. CI가 PR에서 harness를 돌리는 일(카드의 세 번째 기준)은 `.github/workflows/**`를 건드려서 별도 SPEC이에요.

## Non-goals

- CI 워크플로(`.github/workflows/**`)와 harness 코드(`tools/smoke/**`) 바꾸기.
- VERIFY의 다른 단계·표 바꾸기: 한 문단을 더할 뿐이에요.

## Requirements

- REQ-1: The VERIFY workflow shall name `tools/smoke/run.sh <area>` as the check to run first, before the first commit of a change that touches install, update, `/moai project` or the launcher.
- REQ-2: The VERIFY workflow shall name the areas the harness accepts: `install`, `update`, `project`, `launcher` and `all`.
- REQ-3: The VERIFY workflow shall take the last line of the harness output, `smoke: <area> pass|fail steps=<n>`, as the evidence of the check.
- REQ-4 (REQ-3): IF a smoke run was not made, THEN the VERIFY workflow shall mark the check `not_run` and never `passed`.
- REQ-5: The source manifest shall match the edited workflow file.
- REQ-6: The smoke-harness test shall require each sentence of REQ-1 to REQ-4 in the VERIFY workflow file.
- REQ-7 (REQ-6): IF a sentence of REQ-1 to REQ-4 is missing from the VERIFY workflow file, THEN the smoke-harness test shall fail and name the missing sentence.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-6, REQ-7): `verify.md`에 문단이 있고, 문장을 하나 지우면 시험이 빠진 문장을 이름으로 대며 실패해요. 영역 이름은 harness가 받는 이름(`tools/smoke/smoke.py`의 `AREAS`와 `all`)과 같아요.
  verify: `python3 -m unittest tests.smoke_harness_test.VerifyStepTest`
- AC-2 (REQ-5): `source-manifest.json`이 `verify.md`의 현재 글과 맞아요.
  verify: `cargo test --locked --offline --test product_cli skill_package`
- AC-3: 이웃 검사.
  verify: `python3 -m unittest tests.smoke_harness_test`; `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test workflow_size_budget`; `cargo test --locked --offline --test home_isolation_scan`

## Scope

- write: `.moai/specs/SPEC-20261119-smoke-verify-step/spec.md`, `skills/moai/workflows/verify.md`, `skills/moai/references/source-manifest.json`, `tests/smoke_harness_test.py`
- forbidden: 실제 HOME 쓰기, 위 목록 밖의 `skills/moai/**`, `.github/workflows/**`, `tools/smoke/**`, `src/**`

## Plan

1. 시험을 먼저 더해요(RED): 문단이 아직 없어서 실패해야 해요.
2. `verify.md`에 한 문단을 더해요(Procedure 목록 바로 뒤).
3. `tools/source_manifest.py write`로 manifest를 맞춰요.
4. 시험이 통과(GREEN)하고, 문장을 하나씩 지우는 변이가 시험에 잡히는지 읽어요.
5. 이웃 검사, fmt.

## Risks

- 문단이 길어져 워크플로 크기 한도를 넘어요 → `workflow_size_budget`을 이웃 검사에 넣었어요.
- harness의 영역 이름이 바뀌면 글이 낡아요 → 시험이 `AREAS`를 읽어 글의 이름과 맞춰요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 한 문단을 Procedure 목록 뒤에 더하고 번호 목록은 건드리지 않는다 | 번호를 바꾸면 다른 글이 "단계 n"으로 가리키는 곳이 깨져요 | 문단을 지우면 돼요 |
| D-2 | 시험은 `tests/smoke_harness_test.py`에 둔다 | 이 카드의 쓰기 범위에 있고, harness의 시험과 한 곳에서 돌아요 | 시험 클래스를 지우면 돼요 |

## Result

- AC-1: met — `python3 -m unittest tests.smoke_harness_test.VerifyStepTest`: 시험 먼저(RED) 문단이 없을 때 2개 모두 실패(`FAILED (failures=2)`, `verify.md lacks the smoke paragraph`) → 문단을 더한 뒤(GREEN) `Ran 2 tests … OK`. 문단은 `skills/moai/workflows/verify.md`의 Procedure 6번 뒤에 있고, 네 문장(REQ-1~4)과 영역 이름 다섯(`install`·`update`·`project`·`launcher`·`all`)을 시험이 읽어요. 영역 이름은 `tools/smoke/smoke.py`의 `AREAS`에서 읽어 글과 맞춰요.
- AC-2: met — `python3 tools/source_manifest.py write` → `changed: workflows/verify.md`, 이어서 `check` → `{"conflicted": false, "changed": [], "dropped": [], "ok": true}`; `cargo test --locked --offline --test product_cli` 79 passed(`skill_package` 포함).
- AC-3: met — `python3 -m unittest tests.smoke_harness_test tests.source_manifest_test tests.check_pr_self_audit_test tests.ci_scope_test` 152 OK(harness 시험은 실제로 빌드한 `moai-cli`를 써요); `cargo test --locked --offline`로 `ac_command_lint` 5, `workflow_size_budget` 4, `home_isolation_scan` 18, `spec_cli` 17, `sdd_guidance` 6, `test_isolation_rules` 3, `lane_rules_2879` 5, `lane_rules_2836` 2, `agent_defs_package` 19 passed; `cargo fmt --check` 0.

변이(문단의 문장을 하나씩 바꾸고 `VerifyStepTest`): M1 "run the smoke harness first" → 다른 말 → KILLED(두 시험 다 `AssertionError: 'Before the first commit…' not found`); M2 "never reads the real HOME" 삭제 → KILLED; M3 "Take its last line … as the evidence" 삭제 → KILLED; M4 "`not_run`, never `passed`" → "skipped" → KILLED; M5 영역 이름 `launcher` 빠짐 → KILLED(`'`launcher`' not found`). 5/5 KILLED, 변이마다 `git checkout --` 뒤 작업 트리 clean.

그 밖의 검사: `cargo test --locked --offline --lib` 1453 passed; `cargo clippy --locked --offline --all-targets -- -D warnings` 0; `spec check` errors 0 warnings 0(Rust 코드는 안 바뀌었지만 레인 규칙에 따라 돌렸어요).

검사하지 않은 것: 문단의 지시가 실제 레인과 감사자의 행동을 바꾸는지는 이 SPEC의 시험이 말하지 못해요(글 규칙만 고정). C-101의 세 번째 기준(CI)은 별도 SPEC이에요.
