---
id: SPEC-20261127-update-old-manifest-way
title: "update gives an old generation manifest without generation_request a way forward and stops it hiding the managed CLAUDE.md block"
status: draft
revision: 1
created: 2026-11-27
branch: spec/update-old-manifest-way-20261127
---

# update gives an old generation manifest without generation_request a way forward and stops it hiding the managed CLAUDE.md block

## Intent

사용자 지시(2026-11-26) "/moai update ... 전수 조사해라" 의 조사 결과예요. 실제 프로젝트(mo.ai.kr)를 읽기만 해서 확인했어요. `.moai/manifests/generation.json` 이 프로젝트 agent 와 skill(`prj-test-runner`, `prj-reviewer`)을 `owned_files` 에 적고 있는데 `generation_request` 키가 없어요(`has_request` 가 false). `SPEC-20261125-generation-request-record` 는 새로 하는 apply 에만 요청을 기록해서, 그 전에 만든 매니페스트는 요청을 몰라요.

코드에서 읽은 것(base `2e97d012e5`): `src/project_update.rs` 의 `generation_step` 은 `unrebuildable_paths` 가 비어 있지 않으면 계획을 하나도 세우지 않고 `needs_decision`("regenerate them through /moai project")을 명령 없이 돌려줘요. 그 이른 반환이 같은 생성의 나머지, 곧 관리 `CLAUDE.md`/`AGENTS.md` 블록과 workflow 파일 줄도 계획하지 못하게 해요. 그래서 이 프로젝트는 업데이트를 돌릴 때마다 같은 선택 요청에서 멈추고, 선택지로 가는 명령이 없어요. 그 SPEC 은 이 상태를 일부러 그대로 뒀어요(그 SPEC 의 REQ-7, D-4: 파일에서 요청을 거꾸로 만들면 사용자가 손본 파일이나 목적 한 줄을 추측해야 해요).

이 SPEC 은 추측하지 않고 길을 하나 더 열어요. 요청을 모르는 프로젝트 agent·skill 파일과 그 매니페스트 줄은 있는 그대로 두고(지우지도 다시 쓰지도 않아요) 생성의 나머지만 계획해서, 사용자가 `keep-project-files` 를 고르면 그 계획을 적용해요. 다른 길(`/moai project` 로 다시 만들기)은 그대로 남아요. 이 단계가 관리 `CLAUDE.md` 블록 계획을 막지 않게 되는 것이 기준 2 예요.

## Interview

- none: 사용자 답이 아직 없어요. 리더가 사용자에게 물어 카드 노트로 남겨요. 이 SPEC 이 가정하는 물음과 기본값은 아래와 같아요.

물음 1: 요청 기록이 없는 옛 매니페스트에서 프로젝트 agent·skill 을 어떻게 할까요? 기본값 A: "그대로 두고 나머지만 진행 (Recommended)"(이 파일들은 지우지도 다시 쓰지도 않고 매니페스트 줄도 그대로예요). B: "`/moai project` 로 다시 만들기만"(지금과 같아요: 명령이 없고 멈춰요). C: "파일에서 요청을 거꾸로 만들기"(목적 한 줄과 사용자가 손본 곳을 추측해야 해서 이 SPEC 에서 빠져요).

물음 2: A 를 고르면 이 상태가 영구히 남나요? 기본값: 아니요, 다음에 `/moai project` 가 같은 요청으로 apply 하면 그때 기록돼요(`SPEC-20261125-generation-request-record` REQ-4). 그 전까지는 A 를 골라도 계획이 `unchanged` 로 읽혀서 업데이트가 더 멈추지 않아요.

## Non-goals

- 파일에서 요청을 거꾸로 복원하기(`native_roles`, `project_skills`, `skill_projections` 의 목적·필드 추측). `SPEC-20261125-generation-request-record` D-4 가 막은 일이에요.
- `agents-picks` 단계(`rebuilt_request`)가 이 상태에서 `skip` 인 것을 바꾸기. 관리 `CLAUDE.md` 의 `@AGENTS.md` 병합은 `SPEC-20261126-claude-md-fresh-merge` 가 맡고, 그 SPEC 의 `claude-md` 단계는 이 상태와 무관하게 계산돼요.
- `SPEC-20261125-generation-request-record` 의 기록 읽기·쓰기 규칙 바꾸기. 기록이 있는 매니페스트의 동작은 그대로예요.
- 프로젝트 agent·skill 의 내용이나 이름 규칙 바꾸기, 호스트별 파일 이름 바꾸기.
- 스킬 글(`skills/**`) 바꾸기. 도움말과 기존 `update.md` 의 말은 이미 선택지(`preview.choices`)를 읽도록 쓰여 있어서 이 SPEC 이 다시 쓰지 않아요(READY 전에 `update.md` 가 `choices` 를 어떻게 읽는지 한 번 더 읽어 확인해요).

## Requirements

- REQ-1: While the generation manifest owns a native-role, project-skill or skill-projection file and holds no `generation_request`, the generation step of `project-update-plan` shall plan every other row of the generation (the managed block files, the workflow files and the manifest line) from the request the profile and the manifest give, with the owned project agent and skill files and their manifest entries left as they are. · failure: none — the preview writes nothing
- REQ-2: While the manifest is in that state and the plan of REQ-1 has a row to write and no conflict, the step shall read `needs_decision`, keep the `not_rebuilt` list and the sentence that names `/moai project` as the other way, print no step-level apply, and offer the choice `keep-project-files` with the one `generation-apply` command and its stdin for that plan. · failure: none — preview only
- REQ-3: While the manifest is in that state and every row of the plan of REQ-1 is `unchanged`, the step shall read `unchanged`, name the kept files under `not_rebuilt` and offer no choice. · failure: none
- REQ-4: While the manifest is in that state and a row of the plan of REQ-1 is a conflict, the step shall read `conflict` and offer the choices a recorded request would offer for that conflict, each planned with the pick of REQ-5 as well. · failure: none — a choice whose own plan still holds a conflict is left out, as before
- REQ-5: When a generation request carries the pick `keep-harness`, the plan shall hold no `remove` row and no conflict row for a manifest-owned native-role, project-skill or skill-projection file that the request does not name. · failure: none — the pick names no file of its own
- REQ-11 (REQ-5): When the plan of a request that carries `keep-harness` is applied, the apply shall leave every manifest-owned native-role, project-skill and skill-projection file the request does not name, and its manifest entry, as it is. · failure: none — a test compares the bytes and the manifest before and after
- REQ-6: When a request carries `keep-harness` and names project agents or skills, the plan shall plan those named files as it does without the pick. · failure: none
- REQ-7: The agents engine shall accept the id `keep-harness` among the `picks` of a request. · failure: an unknown pick keeps its refusal
- REQ-8: When the plan of REQ-2 is applied, the manifest shall gain no `generation_request` key and lose no owned entry. · failure: none — a test reads the manifest back
- REQ-12 (REQ-8): When the plan of REQ-2 has been applied and nothing else moved, the next `project-update-plan` shall read the generation step as `unchanged`. · failure: none — a test applies and plans again
- REQ-9: IF the plan of REQ-1 cannot be built, THEN the step shall read `needs_decision` with the `not_rebuilt` list, the same reason as before and no choice. · failure: none
- REQ-10: While the manifest holds a `generation_request`, or owns no such file, the generation step shall read as it did before this SPEC. · failure: none — the earlier tests keep passing

## Acceptance

모든 테스트는 임시 HOME 과 임시 프로젝트에서 실제 바이너리로 돌려요. 실제 프로젝트에는 쓰지 않아요.

- AC-1 (REQ-1, REQ-2): 요청 없는 옛 매니페스트(프로젝트 skill 하나를 소유)의 프로젝트는 generation 단계가 `needs_decision` 이고 `not_rebuilt` 에 그 파일이 있으며 step 의 apply 는 없고, `choices` 에 `keep-project-files` 가 명령과 stdin 을 갖고 있고, 그 계획에 관리 `CLAUDE.md` 줄이 있고 `remove` 줄이 없어요 · verify: `cargo test --locked --offline --test project_update_generation_cli old_manifest_offers_keep_and_plans_the_managed_block`
- AC-2 (REQ-5, REQ-8): 찍힌 명령을 그대로 실행하면 `CLAUDE.md` 관리 블록이 써지고 프로젝트 skill 파일의 바이트와 매니페스트 줄이 그대로이며 `generation_request` 키가 없어요 · verify: `cargo test --locked --offline --test project_update_generation_cli keep_apply_writes_the_rest_and_keeps_the_project_files`
- AC-3 (REQ-3, REQ-8): 적용 뒤 다시 계획하면 generation 단계가 `unchanged` 이고 `not_rebuilt` 에 그 파일이 있고 선택지가 없어요 · verify: `cargo test --locked --offline --test project_update_generation_cli after_keep_the_step_reads_unchanged`
- AC-4 (REQ-4): 관리 블록이 지워진 `CLAUDE.md` 의 옛 매니페스트 프로젝트는 `conflict` 이고 `append-managed` 선택지의 명령을 돌리면 블록만 덧붙고 프로젝트 파일은 그대로예요 · verify: `cargo test --locked --offline --test project_update_generation_cli a_conflict_keeps_its_choices_and_the_project_files`
- AC-5 (REQ-5, REQ-6, REQ-7): `generation-plan` 에 `picks:["keep-harness"]` 와 새 프로젝트 skill 하나를 주면 새 skill 은 `create`, 요청에 없는 소유 skill 은 줄이 없고, 같은 요청에서 pick 을 빼면 그 소유 skill 이 `remove` 예요 · verify: `cargo test --locked --offline --test project_update_generation_cli the_pick_stops_removals_and_still_plans_the_named_files`
- AC-6 (REQ-9, REQ-10): 기록이 있는 매니페스트와 소유 파일이 없는 매니페스트는 지금과 같고, 옛 매니페스트 시험 `unrebuildable_not_removed` 도 그대로 통과해요 · verify: `cargo test --locked --offline --test project_update_plan_cli unrebuildable_not_removed`
- AC-7 (REQ-1…REQ-10): 이웃 검사.
  verify: `cargo test --locked --offline --test project_update_plan_cli`
  verify: `cargo test --locked --offline --test generation_request_record`
  verify: `cargo test --locked --offline --test project_harness_pick`
  verify: `cargo test --locked --offline --test project_harness_step`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261127-update-old-manifest-way/`, `src/project_update.rs`, `src/generation.rs`(pick 의 소유 파일 제거 건너뛰기만), `src/agents.rs`(`PICKS` 에 id 하나와 그 설명 주석), `tests/project_update_generation_cli.rs`(새)
- read-only: `src/agents_local_merge/picks.rs`, `skills/moai/workflows/update.md`, `skills/moai/workflows/project.md`
- forbidden: 설치된 HOME 과 실제 사용자 설정, mo.ai.kr 을 포함한 실제 프로젝트 쓰기, 사용자 파일 삭제, `generation_request` 를 파일에서 추측해 쓰기, `agents-picks` 단계 바꾸기

## Plan

1. 시험 먼저: `tests/project_update_generation_cli.rs` 의 시험 다섯(AC-1~5)을 쓰고 실패(RED)를 읽어요. 옛 매니페스트 픽스처는 `unrebuildable_not_removed`(`tests/project_update_plan_cli.rs`)의 모양을 따라요.
2. `src/agents.rs`: `PICKS` 에 `keep-harness` 를 더하고 `HARNESS_ONLY` 옆에 이름을 둬요(엔진은 이 pick 으로 아무것도 바꾸지 않아요).
3. `src/generation.rs`: 요청에 `keep-harness` 가 있으면 요청에 없는 소유 native-role·project-skill·skill-projection 줄을 `remove`/`conflict` 로 계획하는 반복을 건너뛰어요(`is_harness_only` 옆에 `is_keep_harness`). 소유 줄은 `record_generation` 이 이미 계획에 없는 항목을 그대로 두므로 적용 쪽 변경은 없어요.
4. `src/project_update.rs`: `generation_step` 의 이른 반환을 두 갈래로 바꿔요. 요청은 `generation_request()` 가 만든 것(목록이 비어 있어요)에 `keep-harness` 를 더해 계획하고, 줄이 모두 `unchanged` 면 `unchanged`, 충돌이 있으면 `generation_choices` 가 내는 선택지에 같은 pick 을 실어 `conflict`, 아니면 `needs_decision` 에 `keep-project-files` 를 내요. 계획을 못 세우면 지금의 문장 그대로예요(REQ-9).
5. 이웃 검사와 READY 전 검사(`check affected --card`)를 돌려요.

## Risks

- 이 pick 은 소유 파일을 지키는 쪽이라 삭제가 늘지 않아요. 그러나 계획 digest 가 pick 을 포함하므로 `generation-apply` 가 같은 pick 으로 다시 계획해요(drift 검사). 시험이 실제 apply 로 이를 봐요.
- `keep-project-files` 를 고른 뒤에도 매니페스트에 요청이 없어서 다음 `/moai project` 전까지 `agents-picks` 는 `skip` 이에요. 이 SPEC 이 바꾸지 않는 부분이고(Non-goals), 관리 `@AGENTS.md` 병합은 `SPEC-20261126-claude-md-fresh-merge` 가 맡아요.
- `src/project_update.rs` 는 다른 카드 셋도 예약했어요(claude-md 단계, 프로젝트 CI, go-residue). 이 SPEC 의 변경은 `generation_step` 의 이른 반환 한 곳과 그 근처의 보조 함수예요. 합칠 때 겹치면 리더가 줄 단위로 맞춰요(PAIR-NOTE).
- 판이 섞이면 → 새 pick 을 낸 계획은 이전 판 바이너리의 `agents-plan` 이 모르는 pick 으로 거절해요(`invalid agents request`). 이 SPEC 의 판으로 올리면 풀려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 요청을 복원하지 않고 소유 파일을 지키는 pick 을 둬요 | `SPEC-20261125-generation-request-record` D-4: 파일에서 요청을 만들면 추측이 돼요; 지키는 쪽은 아무 파일도 지우거나 다시 쓰지 않아요 | pick 과 `generation_step` 갈래를 지워요 |
| D-2 | 선택지 id 는 `keep-project-files`, pick 이름은 `keep-harness` | `harness-only` 와 같이 pick 이름은 요청 파일의 몫이고 사용자에게는 효과를 말하는 id 를 보여 줘요 | 이름 바꾸기는 한 곳이에요 |
| D-3 | 줄이 모두 같으면 `unchanged` 로 읽어요 | 쓸 것이 없으면 물을 것도 없어요; 영구히 `needs_decision` 에 머무는 것이 이 카드의 증상이에요 | 항상 `needs_decision` 으로 읽어요 |
