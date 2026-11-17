---
id: SPEC-20261117-spec-before-code-patch-mix
title: "SPEC 먼저 규칙(A5) 후속: SPEC과 코드를 함께 쓰는 한 번의 쓰기, MoAI가 아닌 저장소"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-spec-before-code
---

## Intent

C-86 1조각 후속 감사(PR #3176)의 지적 두 가지를 고쳐요. N1: Codex `apply_patch`나 MultiEdit 한 번이 `.moai/specs/…/spec.md`와
코드 파일을 함께 쓸 때, SPEC 파일이 아직 디스크에 없으면 거부된다는 시험이 없어요(감사 mutant `.any` → `.all`이 살아남음).
N2: 작업 트리 루트에 `.moai/` 폴더가 없는 저장소(MoAI를 쓰지 않는 저장소)의 `spec/…` 가지까지 A5가 막아요.

## Requirements

- REQ-1: IF one tool call writes the branch's SPEC file and a code file while the SPEC file does not exist yet, THEN tool-preflight shall refuse the call.
  · failure: none
- REQ-2: IF the work tree of a written file has no `.moai` folder at its root, THEN tool-preflight shall not apply the SPEC-before-code rule to that file.
  · failure: none

## Acceptance

- AC-1 (REQ-1): SPEC 파일이 없을 때 SPEC과 코드를 함께 쓰는 `apply_patch` 하나와 MultiEdit 묶음 하나는 거부되고, SPEC만 쓰는
  같은 호출은 받아요.
  verify: `cargo test --test spec_before_code_hook one_call_writing_the_spec_and_code_waits_for_the_spec`
- AC-2 (REQ-2): `.moai/` 없는 저장소의 `spec/20261117-demo` 가지에서 코드 쓰기는 받고, `.moai/`를 만들면 거부돼요.
  verify: `cargo test --test spec_before_code_hook a_repository_without_a_moai_folder_is_left_alone`

## Scope

- write: `src/hook_guards.rs`(`spec_before_code`), `tests/spec_before_code_hook.rs`,
  `.moai/specs/SPEC-20261117-spec-before-code-patch-mix/spec.md`
- forbidden: 실제 HOME, `.github/workflows/audit-gate.yml`

## Plan

1. 시험 먼저(RED: N2), N1은 동작을 고정하는 시험.
2. `spec_before_code`에 `.moai` 폴더 확인.
3. 검사, 커밋, mutation 2개(`.any` → `.all`, `.moai` 확인 제거).

## Decisions

- D1 이 가지는 #3176이 아직 병합되지 않아 그 head `a7e3b6d981` 위에 쌓았어요(리더 지시).
- D2 `.moai`는 폴더여야 해요(파일이면 MoAI 저장소로 보지 않음). 기존 시험 저장소는 `.moai/.keep`을 커밋해 MoAI 저장소가 돼요.

## Result

구현 에이전트가 채웠어요. 기준 `a7e3b6d981`(#3176 head), 변경 커밋 `20f95a3d84`.

- RED: 코드 변경 전 `a_repository_without_a_moai_folder_is_left_alone`이 `.moai/` 없는 저장소에서 거부되어 FAIL(EXPECTED_RED).
  `one_call_writing_the_spec_and_code_waits_for_the_spec`은 동작을 고정하는 시험이라 처음부터 통과.
- AC-1: met — mutation `.any` → `.all`(감사 mutant)이 이 시험 하나로만 FAIL.
- AC-2: met — mutation `.is_dir()` → `.exists()`(`.moai` 파일도 인정)가 FAIL. 둘 다 커밋 뒤, `git restore` 뒤 tree 깨끗.
- 함께 돌림: `spec_before_code_hook` 9, `--lib hook_guards::` 17, `hook_dispatch_cli` 56, `hooks_cli` 22, `home_isolation_scan` 16,
  `product_cli` 75, `--bin moai` 27, fmt, `rustfmt --check src/ops/*.rs`, clippy `-D warnings`, `moai spec check --all`
  (19개, errors 0, warnings 0).

실행 안 함: 전체 cargo test, Linux CI.
