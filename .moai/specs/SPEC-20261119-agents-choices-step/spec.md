---
id: SPEC-20261119-agents-choices-step
title: "project update: the generation step offers the append, skip and CLAUDE.md import choices"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-agents-conflict-choices
---

# project update: the generation step offers the append, skip and CLAUDE.md import choices

## Intent

카드 C-57 조각 2b예요. 2a(SPEC-20261119-agents-conflict-choices)가 `agents-plan`/`agents-apply` 에 `picks`(append-managed, skip-conflicts, claude-import)를 더했지만, 사용자가 실제로 보는 `project-update-plan` 의 `generation` 단계는 충돌이면 "관리 표식이 없다"는 말만 하고 길을 주지 않아요. 이 SPEC 은 그 단계의 충돌 상태에 `preview.choices` 를 붙여서, 사용자가 고른 뒤 세션이 그대로 실행할 명령을 주게 해요.

선택은 `generation` 요청의 `picks` 로 흘러요. `generation-plan` 이 그 picks 로 계획하고 `generation-apply` 가 계획에 든 picks 로 다시 계획·적용해요. 선택이 없는 요청은 이전과 같이 계획·digest·적용돼요(`picks` 는 비면 JSON 에서 빠져요). 단계의 `state` 와 `apply` 는 그대로예요: 충돌이면 여전히 `conflict` 에 `apply` 가 없고, 선택지마다 따로 명령이 있어요.

## Non-goals

- `update_tux` 가 `skip` 행을 세는 방식과 선택지를 화면에 그리는 일: `src/update_tux.rs` 가 C-33 예약이에요. 이 SPEC 은 JSON 계약까지만 해요.
- Codex 파일 충돌(`codex-config`, `codex-hooks`)의 선택지: 카드 C-57 의 별개 부분이에요.
- 워크플로 글(`skills/moai/workflows/project.md`)이 선택지를 묻는 문단: C-57 기준 5, skills manifest 예약 뒤에 해요.
- `picks` 를 CLI 플래그로 받는 것: `src/main.rs`·`src/help.rs` 를 다른 카드가 쥐어요(2a 의 D-1 과 같아요).

## Requirements

- REQ-1: The generation request shall carry an optional `picks` list that is left out of the JSON when it is empty.
- REQ-2: The `generation-plan` operation shall plan the agents portion with the picks of its request.
- REQ-3: The `generation-apply` operation shall re-plan and apply the agents portion with the picks of the plan it is given.
- REQ-4 (REQ-2): IF a pick id is not one the agents tool knows, THEN the `generation-plan` operation shall refuse the request as invalid input and write nothing.
- REQ-5: The generation plan shall carry the agents plan's `claude_shadow`, left out of the JSON when absent and left out of the plan digest.
- REQ-6: When the `generation` step of `project-update-plan` has a conflict on `AGENTS.md` for a missing managed block, the step shall list `preview.choices` with `append-managed` first and marked recommended and `skip-conflicts` second. · failure: none
- REQ-7: Each choice shall carry the picks it applies, the digest of the plan made with those picks and an `apply` command with that plan on stdin, built the way the step's own `apply` is.
- REQ-8: When the generation plan has a `claude_shadow`, the step shall also list the choice that adds the import line, which is `claude-import` alone while `AGENTS.md` has no conflict and `append-managed+claude-import` while it has one. · failure: none
- REQ-9: The step shall carry no `choices` when `AGENTS.md` has no missing-block conflict and the plan has no `claude_shadow`.
- REQ-10: The step shall keep its `state` and its own `apply` as they were without the choices.
- REQ-11: The agents request and plan schemas shall describe `picks`, `claude_shadow` and the `skip` action.
- REQ-12 (REQ-7): IF the plan made with a choice's picks still has a conflict row, THEN the step shall leave that choice out of `preview.choices`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): `generation-plan` 에 picks 를 주면 AGENTS.md 행이 `append`/`skip` 으로 계획되고, 계획의 digest 가 선택 없는 계획과 다르며, `generation-apply` 가 그 계획의 picks 로 적용해요. 알 수 없는 id 는 `invalid_input` 이고 아무것도 쓰지 않아요. 선택 없는 요청의 JSON 에 `picks` 가 없어요.
  verify: `cargo test --locked --offline --test agents_choices_step_cli generation_`
- AC-2 (REQ-6, REQ-7, REQ-10, REQ-12): 표식 없는 AGENTS.md 의 `project-update-plan` 이 `generation` 을 `conflict`·`apply` 없음으로 두고 선택지 둘을 주며, 각 선택지의 명령을 실제로 실행하면 AGENTS.md 의 앞 바이트를 지키고 표식 한 쌍을 더하거나(append) 그대로 두고(skip) 다음 계획이 `unchanged` 예요. 다른 파일의 충돌이 남아 적용할 수 없는 선택지는 목록에서 빠져요.
  verify: `cargo test --locked --offline --test agents_choices_step_cli choices_`
- AC-3 (REQ-8, REQ-9): CLAUDE.md 가 있고 AGENTS.md 를 안 부르면 import 선택지가 있고, 실행하면 한 줄만 더해요. 충돌도 그림자도 없으면 `choices` 가 없어요.
  verify: `cargo test --locked --offline --test agents_choices_step_cli claude_`
- AC-4 (REQ-11): 스키마가 `picks` 와 `claude_shadow` 와 `skip` 을 말하고, 실제 `agents-plan` 출력의 키가 스키마의 properties 안에 있어요.
  verify: `cargo test --locked --offline --test agents_choices_step_cli schema_`
- AC-5: 이웃 검사.
  verify: `cargo test --locked --offline --test project_update_plan_cli`; `cargo test --locked --offline --test agents_conflict_choices_cli`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test spec_cli`

## Scope

- write: `.moai/specs/SPEC-20261119-agents-choices-step/spec.md`, `src/generation.rs`, `src/project_update.rs`, `src/launcher_init.rs`, `src/agents.rs`, `schemas/protocol/v1/agents-request.schema.json`, `schemas/protocol/v1/agents-plan.schema.json`, `tests/agents_choices_step_cli.rs`
- forbidden: 실제 HOME 쓰기, `src/update_tux.rs`(C-33 예약), `src/main.rs`·`src/help.rs`, 위 목록 밖의 파일

## Risks

- 선택이 없는 `project-update-plan` 출력이 바뀌면 기존 시험·골든이 깨져요 → `choices` 는 충돌·그림자가 있을 때만 키로 나타나게 하고(REQ-9), `project_update_plan_cli` 59개와 `update_tux_golden` 이 그대로 통과하는지 이웃 검사로 읽어요.
- 선택지마다 한 번씩 다시 계획하므로 단계가 느려져요(충돌일 때만, 최대 세 번) → 충돌이 없으면 다시 계획하지 않아요.
- `generation-apply` 가 계획의 picks 를 쓰는데 요청 digest 가 picks 를 포함하지 않으면 다른 선택의 계획을 적용할 수 있어요 → `plan_digest` 가 요청 전체를 묶고 있어서 picks 가 들어요(AC-1 이 선택 없는 계획 digest 와 다름을 봐요).
- `update_tux` 가 `skip` 행을 "막힘"으로 세요(`tally` 의 기본 갈래) → 이 SPEC 은 화면을 건드리지 않고 Result 에 적어 C-33 뒤 후속으로 넘겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 선택지마다 따로 계획하고 digest·명령을 단다 | 명령은 digest 에 묶이는데 digest 는 picks 를 포함한 요청에 묶여요. 선택 뒤에 다시 계획할 필요가 없어요 | `choices` 를 지우면 돼요 |
| D-2 | CLAUDE.md import 는 AGENTS.md 충돌이 있을 때 `append-managed+claude-import` 한 선택지로 합친다 | 충돌이 남은 채 import 만 고르면 적용이 거절돼요(충돌 행이 남아서) | 따로 두 선택지로 나누면 돼요 |
| D-3 | `update_tux` 는 이 SPEC 에서 건드리지 않는다 | C-33 이 예약 중이에요 | C-33 뒤 후속 카드로 이어요 |

## Plan

1. SPEC 을 쓰고 시험을 먼저 더해요(RED): `generation-plan` 이 `picks` 를 모르는 필드로 거절하고, 단계에 `choices` 가 없어서 실패해야 해요.
2. `GenerationRequest.picks`·`GenerationPlan.claude_shadow` 를 더하고 `generation_plan`·`generation_apply` 가 넘겨요. `launcher_init` 요청은 picks 없이 둬요.
3. `project_update` 의 `generation_step` 이 충돌·그림자일 때 선택지별 계획·digest·명령을 `preview.choices` 로 줘요.
4. 스키마를 고쳐요.
5. 변이와 이웃 검사를 돌려요.
