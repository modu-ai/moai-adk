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
