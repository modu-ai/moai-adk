---
id: SPEC-20261119-project-preview-order
title: "The PROJECT route walks the project-update-plan preview in its printed order and never asks about a waiting step"
status: draft
revision: 1
created: 2026-11-19
---

# The PROJECT route walks the project-update-plan preview in its printed order and never asks about a waiting step

## Intent

카드 C-59 의 기준 1 이에요: "project.md 가 미리보기 단계 순서를 따르고, 앞 단계에 막힌 단계는 묻지 않음". 사용자가 mo.ai.kr 에서 돌린 `/moai project`(2026-11-16)에서 라우트가 아직 `blocked` 로 읽히는 단계를 먼저 물었어요.
tip(05cbe0ac67)에서 읽은 사실이에요. `moai project-update-plan` 은 세 단계(`backup`, `install`, `merge`)를 정해진 순서로 찍고, 앞 단계가 쓰기를 남겼으면 뒤 단계를 `blocked` + `re-run after …` 로 읽어요(`src/project_update.rs` 의 `AFTER_BACKUP`, `AFTER_PROFILE`, `AFTER_SETTINGS`, `AFTER_GO_SETTINGS`, `AFTER_INSTALL`). UPDATE 라우트(`workflows/update.md` 단계 2–4)는 이 규칙을 전부 말해요. PROJECT 라우트(`workflows/project.md`)는 단계 2(backup 단계의 `go-settings` → `go-state`)에서만 말하고, 단계 5 는 `generation` 단계를 "쓴다"고만 하고 읽힌 상태를 가리지 않아요. 마지막 선택(Completion Check)은 `blocked` 를 제안하지 않는다고만 하고 순서는 말하지 않아요.
이 SPEC 은 `project.md` 의 글과 그 글이 인용하는 대기 이유가 코드의 상수와 같다는 시험만 다뤄요. 새 Rust 명령은 없어요.

## Non-goals

- 기준 2(MCP plan/apply): 리더 결정(2026-11-19T02:22:17Z)으로 이 카드에서 하지 않아요(가드된 호스트 전역 설정 변경). 기준 3(worktree 이동 선택)은 다음 조각이에요.
- `project-update-plan` 의 코드, 단계 이름, 대기 이유의 문구 바꾸기: 읽기만 해요.
- `update.md` 바꾸기: 이미 이 규칙을 말해요.

## Requirements

- REQ-1: When the PROJECT workflow runs `moai project-update-plan` in step 5, the workflow shall ask about the `generation` step only if that step reads `ready` in that run.
- REQ-2: The PROJECT workflow shall present the steps its Completion Check offers in the printed order of the plan: phase `backup` first, then `install`, then `merge`, each phase in the printed order of its steps.
- REQ-3 (REQ-1, REQ-2): IF a step reads `blocked` or `re-run after …` in the run the user was shown, THEN the PROJECT workflow shall ask nothing about it, apply the earlier step that the reason names, run `moai project-update-plan` again, and ask about it only from that new run.
- REQ-4: The PROJECT workflow shall name the waiting reasons it quotes exactly as `src/project_update.rs` prints them. · failure: none — the text is checked against the constants by a test, and a wrong quote fails that test.

## Acceptance

- AC-1 (REQ-1, REQ-3): `project.md` 단계 5 가 `generation` 단계를 `ready` 로 읽힐 때만 묻고, `blocked`/`re-run after …` 면 묻지 않으며 앞 단계의 apply 와 다시 미리보기와 새 실행에서만 묻기, 앞선 yes 가 보지 못한 단계를 덮지 않는다는 것을 말해요.
  verify: `cargo test --locked --offline --test project_preview_order step_5_asks_about_generation_only_when_it_reads_ready`
- AC-2 (REQ-2, REQ-3): `project.md` 의 Completion Check 가 제안 목록을 phase 순서 `backup` → `install` → `merge` 와 각 phase 안의 찍힌 순서로 말하고, 기다리는 단계는 묻지 않고 막는 단계의 이유로 보고해요.
  verify: `cargo test --locked --offline --test project_preview_order completion_check_offers_steps_in_the_printed_phase_order`
- AC-3 (REQ-4): 글이 인용하는 대기 이유 문구가 `src/project_update.rs` 의 상수와 글자까지 같고, phase 이름 `backup`, `install`, `merge` 도 같은 파일의 `phase(...)` 호출 순서와 같아요.
  verify: `cargo test --locked --offline --test project_preview_order quoted_waiting_reasons_are_the_ones_the_code_prints`
- AC-4 (REQ-1…REQ-4): 이웃 검사.
  verify: `cargo test --locked --offline --test project_flow_closing`; `cargo test --locked --offline --test update_route`; `cargo test --locked --offline --test project_update_plan_cli`; `cargo test --locked --offline --test agent_defs_package`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test doc_references`; `python3 -m unittest tests.source_manifest_test tests.skill_references_test`

## Scope

- write: `.moai/specs/SPEC-20261119-project-preview-order/spec.md`, `tests/project_preview_order.rs`, `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/project_update.rs`, `skills/moai/workflows/update.md`, `tests/update_route.rs`, `tests/project_flow_closing.rs`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, push 이외의 원격 쓰기, MCP 설정 파일

## Plan

1. 시험 먼저: `tests/project_preview_order.rs` 의 시험 셋을 쓰고, 글 시험 둘이 실패하는 것(RED)을 읽어요. 상수 시험은 이미 통과해야 해요(글이 아직 인용하지 않으면 인용 목록이 비어 있어도 되는지 시험이 정해요).
2. `project.md` 단계 5 와 Completion Check 에 문장을 더해 GREEN 으로 만들고, 매니페스트 해시는 마지막 커밋에서만 써요(예약 경로를 최소로).
3. 이웃 검사(AC-4)와 READY 전 검사 전부를 돌려요.

## Risks

- 글 시험은 문장을 고정해서 나중에 문구를 고치면 시험도 같이 고쳐야 해요 → 시험은 규칙을 이루는 문장(`ready` 일 때만 묻기, 기다리는 단계는 묻지 않기, phase 순서)과 인용한 대기 이유만 고정해요. 문장 전체가 아니에요.
- 대기 이유 문구가 코드에서 바뀌면 글의 인용이 틀려져요 → AC-3 이 상수와 글자까지 맞춰서 바로 실패해요.
- `project.md` 는 매니페스트의 해시가 걸린 파일이라 다른 레인의 예약과 겹칠 수 있어요 → 매니페스트 경로는 마지막 커밋 하나에만 예약하고 바로 풀어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기준 2(MCP plan/apply)는 이 SPEC 에서 하지 않아요 | 리더 결정 2026-11-19T02:22:17Z(note 2521 답)과 note 3437: MCP 설정 편집은 `guarded` 호스트 전역 변경이고 `project.md` 는 "MoAI never creates or edits an MCP configuration" 를 지켜요; 기준 2 는 이 이유로 철회했어요 | 사용자가 고르면 별도 카드로 올려요 |
| D-2 | install·merge 단계를 PROJECT 에서 따로 한 단계씩 묻는 새 절차를 만들지 않고, 단계 5 와 Completion Check 에 한 줄씩 더해 UPDATE 라우트(`update.md` 2–4)와 같은 규칙을 가리켜요 | UPDATE 가 이미 전부 말하고 `update_route` 시험이 지켜요; 같은 규칙을 두 번 쓰면 어긋나요 | 문장 두 곳을 지워요 |

## Result

- AC-1: met — `cargo test --locked --offline --test project_preview_order step_5_asks_about_generation_only_when_it_reads_ready` passes (RED 먼저: 3 중 3 failed, exit 101, 글에 문장이 없어서). 단계 5 가 `ready` 일 때만 묻기, `blocked`·`re-run after …` 는 묻지 않기, 앞 단계 apply 후 다시 미리보기, 앞선 yes 가 보지 못한 단계를 덮지 않는다는 것을 말해요.
- AC-2: met — 같은 시험 파일 `completion_check_offers_steps_in_the_printed_phase_order` passes. Completion Check 가 `backup` → `install` → `merge` 와 각 phase 안의 찍힌 순서, 기다리는 단계의 이유 보고를 말하고, `src/project_update.rs` 의 `phase("backup"`·`phase("install"`·`phase("merge"` 순서와 같아요.
- AC-3: met — 같은 시험 파일 `quoted_waiting_reasons_are_the_ones_the_code_prints` passes. 글이 인용한 `re-run after …` 이유 다섯 개가 상수 `AFTER_BACKUP`·`AFTER_PROFILE`·`AFTER_SETTINGS`·`AFTER_GO_SETTINGS`·`AFTER_INSTALL` 과 같고, `generation`·`hooks`·`go-state` 의 짝이 코드의 `Step::waiting(...)` 줄과 같아요.
- Mutation (committed c4dbdba9e7 first): 변이 5 개를 `project.md` 에 하나씩 넣고(적용됐는지 old 문구가 정확히 1번인지 확인) 컴파일된 시험 바이너리를 돌린 뒤 `git restore` 했어요: M1 `ready` 규칙을 "also when blocked" 로, M2 phase 순서를 merge 앞으로, M3 인용을 `re-run after the go-settings step` 으로, M4 "earlier yes" 절 삭제, M5 `hooks` 를 profile 이유와 짝지음. 5 개 모두 exit 101 로 죽었고(M1·M4 는 단계 5 시험, M2 는 Completion Check 시험, M3·M5 는 인용 시험), 복원 뒤 파일이 커밋본과 같아요.
- 이웃: `project_flow_closing` 10, `project_preview_order` 3, `project_update_plan_cli` 59, `update_route` 9 passed(exit 0).
- Not run here: 전체 `cargo test`; 실제 `/moai project` 세션이 이 글을 읽는 것(시험은 문구와 코드의 인용이 맞는지만 고정해요).
