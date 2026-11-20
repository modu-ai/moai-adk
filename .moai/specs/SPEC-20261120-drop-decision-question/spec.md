---
id: SPEC-20261120-drop-decision-question
title: "Drop the decision model question from the project interview"
status: approved
revision: 1
created: 2026-11-20
approval: {by: user, at: 2026-11-21, basis: "AskUserQuestion selection: 승인 (Recommended)", digest: "sha256:92ee5e290a412041c7b703e4291ff3cda8eea16dd8fddd21d665414f7ee96253"}
---

# Drop the decision model question from the project interview

## Intent

사용자 결정(2026-11-21 KST, `/moai project` 인터뷰 중): "결정 모델 질문은 질문 리스트에서 삭제, 결정 모델은 더 이상 사용하지 않는다." 지금 PROJECT 인터뷰는 질문 은행(`interview.md`)의 `decision_model` 행과 `project.md` 3단계의 "Decision model" 항목 때문에 매번 이 질문을 해요. 이 SPEC 은 질문과 그 안내를 스킬에서 지우고, 그 질문을 고정하던 테스트를 "질문이 없다"로 바꿔요.

## Requirements

- REQ-1: When the PROJECT route runs its interview, the question bank in `skills/moai/references/internal/interview.md` shall hold no `decision_model` row and no paragraph that describes one · failure: none — a documentation change; nothing runs.
- REQ-2: When the PROJECT route reads `skills/moai/workflows/project.md`, the workflow shall name no decision model, neither as a step 3 item nor in the final report's settings table · failure: none — a documentation change; nothing runs.
- REQ-3: When a profile from before this change carries a `decisions` entry with id `decision_model`, `moai project-plan` shall accept that profile · failure: none — no code path changes; the test pins the existing behaviour.
- REQ-4: When a packaged skill file changes, `skills/moai/references/source-manifest.json` shall carry its new sha256 · failure: none — the package test fails on a stale digest, which is the guard.

## Acceptance

- AC-1 (REQ-1, REQ-2): 질문 은행 행 목록에 `decision_model` 이 없고, `project.md` 에 `Decision model` 항목과 "decision model" 표 문구가 없음. verify: `cargo test --locked --offline --test project_decide_model --test agent_defs_package`
- AC-2 (REQ-3): `decision_model` 결정 값(`off`·`typesafe`·`openai`)을 가진 예전 프로필이 `project-plan` 에서 `ok: true`. verify: `cargo test --locked --offline --test project_decide_model`
- AC-3 (REQ-4): 패키지 매니페스트 sha256 이 두 스킬 파일과 맞음. verify: `cargo test --locked --offline --test product_cli skill_package`
- AC-4: 저장소 가벼운 검사. verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `skills/moai/references/internal/interview.md`, `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`, `tests/project_decide_model.rs`, `tests/agent_defs_package.rs`, `.moai/specs/SPEC-20261120-drop-decision-question/spec.md`
- read-only: `src/decide/`(`moai decide` 기능 코드 — 이번에 지우지 않음), `.moai/specs/SPEC-20261118-project-decide-model/spec.md`, `.moai/specs/SPEC-20261118-decide-control/spec.md`
- forbidden: 키 보관소·실제 HOME 접근, 카드 번호를 SPEC id·브랜치·커밋에 사용

## Interview

- Q1: 결정 모델 질문(AI 결정 모델 Jev·OpenAI)을 인터뷰에 계속 둘까요? → A: 질문 리스트에서 삭제, 결정 모델은 더 이상 사용하지 않음 (사용자, 2026-11-21 `/moai project` 중)
- Q2: 범위를 질문 제거로 할까요, `src/decide` 기능 코드 제거까지 넓힐까요? → A: 승인 (Recommended) — 질문 제거만, `src/decide` 유지

## Plan

1. 테스트 먼저: `tests/project_decide_model.rs` 를 "질문 은행에 `decision_model` 행이 없다 / `project.md` 에 Decision model 항목이 없다"로 바꾸고, `tests/agent_defs_package.rs` 기대 목록에서 `decision_model` 을 빼서 지금 스킬로는 실패하는지 확인.
2. `interview.md` 의 행과 설명 문단, `project.md` 3단계 항목과 최종 보고 표 문구를 지움.
3. `source-manifest.json` 의 두 sha256 갱신, AC 검사 실행.
4. `moai decide` 기능 코드(`src/decide/`) 제거는 후속 과제로 남김.

## Result

- AC-1: met — head 853778b69c 에서 `cargo test --locked --offline --test project_decide_model --test agent_defs_package` exit 0 (19 + 4 passed); 고치기 전 스킬 텍스트에서는 같은 명령이 exit 101 (새 테스트 2개와 `interview_question_bank_covers_the_project_axes` 실패); `grep -rn -i -e decision_model -e 'decision model' -e Jev skills/` 출력 없음 (exit 1)
- AC-2: met — 같은 `project_decide_model` 실행에서 예전 프로필 호환 테스트(`off`·`typesafe`·`openai` 가 `project-plan` 에서 `ok: true`) 통과
- AC-3: met — `cargo test --locked --offline --test product_cli skill_package` exit 0 (`source_manifest_covers_every_package_file` 포함); `python3 -I tools/source_manifest.py check` exit 0
- AC-4: met — `git diff --check` exit 0; `cargo fmt --check` exit 0; `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0
