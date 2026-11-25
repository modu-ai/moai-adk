---
id: SPEC-20261125-update-same-bytes-no-conflict
title: "A flat project agent file equal to the new render is moved, and one conflict does not stop the rest of the update"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9097): \"승인, 세 가지 모두 (Recommended)\"; 리더가 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:8cd73aebe6b81f1200ced70359a09cb4da0008b02d959ff2114ce3a08b8970c6"
---

# A flat project agent file equal to the new render is moved, and one conflict does not stop the rest of the update

## Intent

`/moai update` 의 generation 단계(update 경로 ③)는 두 군데에서 쓸데없이 멈춰요. 첫째, 예전 버전이 `.claude/agents/<id>.md` 에 평평하게 쓴 프로젝트 에이전트 파일이 지금 만들 `.claude/agents/moai/<id>.md` 의 새 글과 바이트까지 같아도 `conflict` 로 잡혀요. 판정이 기록해 둔 기준 해시(baseline)와만 비교하기 때문이에요(`src/generation.rs` 의 평평한 역할 파일 절). 실제로 `prj-local-check.md` 가 그렇게 잡혔어요. 사용자가 고친 게 없는 파일을 충돌로 보여 주면 사용자는 아무 잘못이 없는데 막혀요. 둘째, 계획에 충돌 행이 하나라도 있으면 `generation-apply` 가 전체를 `nothing applied` 로 거절해요. 충돌을 빼고 나머지를 적용하는 선택(`skip-conflicts`)이 있기는 하지만 `CLAUDE.md` 의 표식이 없어진 한 가지 충돌에만 주어지고, 그것도 agents 엔진이 가진 파일에만 듣고 generation 이 가진 파일(워크플로 글, 역할 파일, 스킬 투영, 프로젝트 스킬)의 충돌은 건너뛸 방법이 없어요. 이 SPEC 은 (1) 새 글과 같은 평평한 파일은 충돌이 아니라 옮김(`remove`)으로 계획하고, (2) 어떤 충돌이든 `skip-conflicts` 선택을 주고, (3) generation 이 가진 충돌 행도 그 선택으로 `skip` 이 되게 해요. 사용자가 고친 파일은 지금처럼 그대로 남아요.

## Interview

- Q1: 지우는 범위가 넓어져요. 평평한 옛 에이전트 파일이 새 글과 같으면 사용자에게 묻지 않고 옮기(지우)고 그 바이트는 백업에 남겨도 될까요? → A: 아직 묻지 않았어요. 기본: 예. 백업은 지금 `remove` 가 이미 하는 것과 같은 자리(`save_before_bytes`)에 남고, 바이트가 같아서 잃는 내용이 없어요(D-1).
- Q2: 충돌이 있을 때 "충돌 파일만 빼고 나머지 적용"을 `CLAUDE.md` 한 가지가 아니라 어떤 충돌에도 선택지로 줄까요? → A: 아직 묻지 않았어요. 기본: 예. 선택지는 지금처럼 사용자가 고른 뒤에만 실행돼요(D-2).
- Q3: 사용자가 고친 평평한 에이전트 파일을 건너뛸 때 그 역할의 새 파일(`.claude/agents/moai/<id>.md`)도 같이 건너뛸까요? → A: 아직 묻지 않았어요. 기본: 예. 같은 이름의 에이전트 두 개가 한 프로젝트에 남는 일을 막아요(D-3).

## Non-goals

- 사용자가 고친 파일을 덮어쓰거나 합치는 일: 고친 파일은 `conflict` 로 남고 `skip-conflicts` 를 골라도 바이트를 건드리지 않아요.
- `AGENTS.md` 합치기 단계(`agents-picks`)와 GitHub 템플릿·hooks·settings 단계의 충돌 처리: 그 단계들의 거절 기록과 이유 문장은 별도 조각(update 경로 ⑤)이에요.
- `/moai update` 지침(`update.md`)의 문장, 물음 횟수, 단계 순서(update 경로 ⑥·⑦).
- 백업 위치와 HOME 표식(update 경로 ⑧).
- `harness-only` 요청이 쓰는 건너뛰기: 그 동작은 그대로예요.

## Requirements

- REQ-1: When an earlier flat role file has bytes equal to the rendered bytes of its nested copy, the generation plan shall plan that file as `remove` and not as `conflict`.
- REQ-2 (REQ-1): IF an earlier flat role file differs from both its recorded baseline and the rendered bytes of its nested copy, THEN the generation plan shall keep it as a `conflict` with its current bytes unchanged.
- REQ-3 (REQ-1): When the generation apply removes a flat role file, the apply shall first save its bytes in the request's backup folder. · failure: none — a backup that cannot be saved stops the apply as `partial apply` does today.
- REQ-4: Where the request picks `skip-conflicts`, the generation plan shall plan every conflicting file it owns as `skip` and name the conflict in the row's detail.
- REQ-5 (REQ-4): When a flat role file is planned as `skip`, the generation plan shall plan the nested copy of the same role as `skip`. · failure: none — the nested copy is only left unwritten.
- REQ-6 (REQ-4): When the generation apply meets a `skip` row, the apply shall leave the file's bytes and its recorded baseline as they are. · failure: none — a skipped row writes nothing that could fail.
- REQ-7: When the generation step of the update plan has any conflict row, the step shall list `skip-conflicts` in its choices with the digest and the apply command of that choice.
- REQ-8 (REQ-7): The generation step shall list `append-managed` in its choices only for a missing managed block in `CLAUDE.md`.
- REQ-9 (REQ-7): The choice `skip-conflicts` shall name in its effect the paths it keeps as they are.
- REQ-10 (REQ-7): IF the plan made with a choice's picks still has a conflict row, THEN the step shall leave that choice out of its choices.
- REQ-11: The generation plan shall plan a file that is not in conflict as before this SPEC. · failure: none — only the rows named above change.
- REQ-12: A generation request without picks shall keep the digest it had before this SPEC. · failure: none — the picks field is absent from such a request.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 임시 프로젝트에서 옛 평평한 파일이 새 글과 같으면 계획이 `remove` 이고 적용하면 파일이 사라지고 바이트가 백업에 남아요. 사용자가 고친 평평한 파일은 `conflict` 이고 적용 전후 바이트가 같아요 · verify: `cargo test --locked --offline --test project_agents_nested ac1_`
- AC-2 (REQ-4, REQ-5, REQ-6): `skip-conflicts` 를 고른 요청에서 고친 평평한 파일, 고친 워크플로 글, 고친 프로젝트 스킬 파일의 행이 `skip` 이 되고 역할의 새 파일도 `skip` 이며, 적용하면 충돌이 없는 나머지 파일은 써지고 건너뛴 파일의 바이트와 기록된 기준 해시는 그대로예요 · verify: `cargo test --locked --offline --test project_agents_nested ac2_`
- AC-3 (REQ-7, REQ-8, REQ-9, REQ-10): `project-update-plan` 의 generation 단계가 어떤 충돌에도(고친 `CLAUDE.md` 블록, 고친 워크플로 글, 고친 평평한 파일) `skip-conflicts` 선택을 주고, 그 명령을 실제로 실행하면 충돌 파일은 그대로이고 나머지가 적용돼요. `append-managed` 는 표식이 없어진 `CLAUDE.md` 에만 주어져요. 선택지 `effect` 에 지킬 경로가 있어요 · verify: `cargo test --locked --offline --test agents_choices_step_cli ac3_`
- AC-4 (REQ-11, REQ-12): 충돌이 없는 계획, picks 가 없는 요청의 digest, 기존 선택지 시험이 그대로 통과해요 · verify: `cargo test --locked --offline --test agents_choices_step_cli`
  verify: `cargo test --locked --offline --test project_agents_nested`
  verify: `cargo test --locked --offline --test generation_native_role_link`
  verify: `cargo test --locked --offline --test generation_request_record`
  verify: `cargo test --locked --offline --test project_skills_generation`

## Scope

- write: `.moai/specs/SPEC-20261125-update-same-bytes-no-conflict/spec.md`
- write: `src/generation.rs`, `src/project_update.rs`
- write: `tests/project_agents_nested.rs`, `tests/agents_choices_step_cli.rs`
- read-only: `src/agents.rs`, `src/update.rs`, `src/ops/generation_apply.rs`, `src/ops/generation_plan.rs`
- forbidden: 실제 HOME, 설치된 moai, 다른 프로젝트, 새 최상위 테스트 파일(표 검사), 사용자가 고친 파일을 쓰는 코드, 카드 번호를 이 파일 밖에 쓰는 일

## Plan

1. 승인 전에는 코드를 바꾸지 않아요. 승인 뒤 `card scope` 로 위 write 경로를 예약해요(다른 카드가 같은 경로를 쥐고 있으면 그 카드가 끝난 뒤에 넓혀요).
2. 시험 먼저: `project_agents_nested.rs` 에 AC-1·AC-2, `agents_choices_step_cli.rs` 에 AC-3 의 RED 시험을 쓰고 실패 이유를 기록해요. 이미 있는 `choices_an_edited_block_conflict_offers_no_append_or_skip` 과 `choices_a_choice_whose_own_plan_still_conflicts_is_left_out` 은 새 동작(어떤 충돌에도 건너뛰기를 줌)을 반대로 고정하고 있어서 이 SPEC 의 REQ-7·REQ-10 에 맞게 다시 쓰고, 어떻게 바꿨는지 READY 에 적어요.
3. `src/generation.rs`: 평평한 역할 파일 절에서 현재 바이트의 digest 가 새 렌더의 digest 와 같으면 `remove`(REQ-1)로 계획하고, 그렇지 않고 기준과 다르면 `conflict` 로 두어요(REQ-2). `remove_owned` 는 이미 백업을 먼저 저장해요(REQ-3, 시험으로 고정).
4. `src/generation.rs`: 요청의 `picks` 에 `skip-conflicts` 가 있으면 generation 이 가진 모든 `conflict` 행(역할 파일 링크·평평한 파일·스킬 투영·프로젝트 스킬)을 `skip` 으로 바꿔요(REQ-4). 평평한 파일이 `skip` 이면 같은 역할의 새 파일 행도 `skip`(REQ-5). 적용 루프는 `skip` 행을 `unchanged` 처럼 건너뛰고(REQ-6), 기준 해시 기록(`record_generation`)도 건너뛴 행을 건드리지 않아요.
5. `src/project_update.rs`: `generation_choices` 가 충돌 행이 하나라도 있으면 `skip-conflicts` 를, 표식 없는 `CLAUDE.md` 일 때만 `append-managed` 를 주고(REQ-7·REQ-8), `effect` 에 지킬 경로를 적어요(REQ-9). 선택지의 자기 계획에 충돌이 남으면 그 선택지는 빼는 기존 규칙은 그대로예요(REQ-10).
6. AC-4 의 이웃 시험을 돌리고 Result 에 기준마다 한 줄을 적고 안 돌린 것을 밝혀요. `docs/` 나 스킬 글은 이 SPEC 의 범위가 아니라 바꾸지 않아요.

## Risks

- 지우는 범위가 넓어져요(tier 2): 새 글과 바이트가 같은 평평한 파일만 지우고 백업이 먼저 저장돼요. 바이트 한 개라도 다르면 지금처럼 `conflict` 예요.
- `skip-conflicts` 가 generation 파일까지 건너뛰면 `generation` 단계가 `skip` 행이 남은 채 "적용할 것 있음" 으로 보일 수 있어요. 다음 계획에서도 같은 충돌이 다시 나오고 같은 선택지가 다시 주어지는 것은 `CLAUDE.md` 의 기존 동작과 같아요. 거절을 기억해서 다시 안 묻는 일은 update 경로 ⑤ 의 몫이에요.
- 기존 시험 둘이 반대 동작을 고정하고 있어요. 다시 쓴 시험이 같은 사실(고친 파일은 바이트가 그대로)을 계속 지키는지 읽어야 해요.
- 두 이름의 에이전트가 남지 않게 새 파일 행도 건너뛰지만, 평평한 파일이 이미 있는 프로젝트에서 새 파일이 이미 있는 경우(둘 다 있음)는 건드리지 않아요.
- `SPEC-20261119-agents-choices-step` REQ-6 은 "표식 없는 `CLAUDE.md` 에만 선택지를 준다" 고 했어요. 이 SPEC 의 REQ-7·REQ-8 이 그 문장을 대신해요(Decisions D-2).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 글과 바이트가 같은 평평한 옛 에이전트 파일은 충돌이 아니라 `remove` 로 계획해요(REQ-1). 대안: 지금처럼 충돌로 두고 사용자가 지우게 함. | 바이트가 같아서 잃는 내용이 없고 백업이 먼저 저장돼요. 사용자 지시 2026-11-25 "/moai update 제대로 단계별로 구현이 안된다". 승인 전. | REQ-1 의 분기를 지우면 이전 동작이에요. |
| D-2 | `skip-conflicts` 를 어떤 충돌에도 주고 `append-managed` 는 표식 없는 `CLAUDE.md` 에만 둬요(REQ-7·REQ-8). `SPEC-20261119-agents-choices-step` 의 REQ-6·REQ-12 를 바꿔요. | 카드 기준 3. 선택지는 사용자가 고른 뒤에만 실행돼요. 승인 전. | `generation_choices` 의 조건을 이전 것으로 되돌려요. |
| D-3 | 고친 평평한 파일을 건너뛰면 그 역할의 새 파일도 건너뛰어요(REQ-5). 대안: 새 파일은 써서 같은 이름의 에이전트 둘이 남게 함. | `SPEC-20261121-base-agents-user-folder` REQ-14 가 "새 파일을 쓰지 않음" 이라고 했어요. 승인 전. | REQ-5 의 규칙을 지워요. |
| D-4 | 이 SPEC 은 generation 단계만 다뤄요. 다른 단계의 충돌은 update 경로 ⑤ 에 둬요. | 분석 노트의 조각 순서(③ 다음 ⑤). | 해당 없음. |
