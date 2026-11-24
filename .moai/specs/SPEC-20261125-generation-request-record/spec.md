---
id: SPEC-20261125-generation-request-record
title: "generation-apply records the project agent and skill request, so an apply with every file unchanged clears needs_decision"
status: draft
revision: 1
created: 2026-11-25
branch: spec/generation-request-record-20261125
---

# generation-apply records the project agent and skill request, so an apply with every file unchanged clears needs_decision

## Intent

mo.ai.kr 실측(2026-11-19, 2026-11-21): `/moai project` 가 `generation-apply` 로 프로젝트 agent 와 skill 을 만든 뒤 `project-update-plan` 의 generation 단계가 `needs_decision` ("request is not recorded") 에 영구히 머물러요. 처음 보고된 "`generation-apply` 가 `generation-plan` 이 받은 stdin 을 `invalid typed JSON input` 으로 거절한다"는 사용 실수였어요(help 대로 plan 의 `result` 를 stdin 으로 주면 apply 는 `ok`, 상태 `applied`). 진짜 결함은 둘이에요.

1. `record_generation`(`src/generation.rs`)은 `owned_files` 베이스라인과 `workflow_id`·`interview_revision`·`profile_sha256` 만 쓰고, 파일을 만든 요청(`native_roles`, `skill_projections`, `project_skills`)은 어디에도 적지 않아요. 그래서 `project-update` 의 generation 단계는 `owned_files` 에 다시 만들 수 없는 섹션(`generation:native-role`·`project-skill`·`skill-projection`)이 있으면 요청을 알 수 없다고 보고 `needs_decision` 을 내요.
2. 모든 파일이 이미 같은 내용(unchanged)인 plan 을 apply 하면 만들거나 합친 파일이 없어서, 매니페스트가 바뀌지 않는 한 아무것도 기록되지 않아요. 기록이 없으니 `needs_decision` 이 풀리지 않아요.

이 SPEC 은 적용할 때 그 요청을 매니페스트 쓰기 한 번에 함께 기록하고(파일이 모두 unchanged 여도), 다음 `project-update-plan` 이 기록을 읽어 `unchanged`(프로필이 바뀌었으면 다시 만드는 계획)로 읽게 해요. 기록 모양은 요청 자신의 타입을 그대로 써서, 프로젝트 harness 제안이 `NativeRole`·`ProjectSkill` 에 필드를 더하면 같은 기록이 그 필드를 같이 실어요.

## Non-goals

- 프로젝트 harness 를 웹 검색으로 제안하기: 제안이 쓸 기록 자리만 만들어요.
- 기록이 없는 옛 매니페스트의 요청을 파일에서 거꾸로 복원하기: 추측으로 진행하지 않아요. 사용자가 같은 요청으로 `generation-apply` 를 한 번 더 하면 그때 기록돼요(REQ-4).
- `generation-plan` 의 입력과 출력 모양 바꾸기, `generation-apply` 가 받는 stdin 의 모양 바꾸기, 호스트별 파일 이름 바꾸기.
- 스킬 글(`skills/**`) 바꾸기.
- `needs_decision` 의 다른 원인(placeholder 프로필 등) 바꾸기.

## Requirements

- REQ-1: When `generation-apply` publishes the manifest, it shall record the request's `native_roles`, `skill_projections` and `project_skills` under one `generation_request` key of the same publish. · failure: none — a test reads the manifest back after an apply.
- REQ-2: When the request has no native role, no skill projection and no project skill, `generation-apply` shall leave the manifest without a `generation_request` key. · failure: none — a test reads the manifest after an apply of an empty request.
- REQ-3: The generation step of `project-update-plan` shall rebuild its request from the recorded `generation_request`. · failure: none — a test plans after an apply and reads the step.
- REQ-4: When every managed file is already current, `generation-apply` shall still record the request and leave every file byte-identical. · failure: none — a test removes the record from a settled project, applies the same request and compares the files.
- REQ-5: When every managed file is current, the generation step shall read `unchanged` for a project whose manifest records its agents and skills. · failure: none — a test applies one project agent and one project skill and reads `unchanged`.
- REQ-6: When the profile changed what a recorded project skill is built from, the generation step shall read `ready` and print the apply of the rebuilt request. · failure: none — a test changes the recorded test command and reads the step.
- REQ-7: While the manifest owns a native-role, project-skill or skill-projection file and holds no `generation_request`, the generation step shall read `needs_decision` with the `not_rebuilt` list and no printed command. · failure: none — the existing test `unrebuildable_not_removed` keeps passing.
- REQ-8: A later `generation-apply` shall replace the record whole with its own request. · failure: none — a test applies a second request with one role fewer and reads the record.
- REQ-9: An apply of the agents portion alone shall carry the recorded `generation_request` over unchanged. · failure: none — a test applies the profile or agents step after a generation apply and reads the record.
- REQ-10: The typed manifest of `agents-apply` shall accept an optional `generation_request` and keep it when it republishes the manifest. · failure: none — a test applies a project agent, which fails as an amended manifest outside the schema without the field.

## Acceptance

- AC-1 (REQ-1, REQ-2): 적용 뒤 `generation.json` 의 `generation_request` 가 요청의 세 목록과 같은 JSON 이고, 빈 요청이면 키가 없어요.
  verify: `cargo test --locked --offline --test generation_request_record the_applied_request_is_recorded_as_the_request_itself`
- AC-2 (REQ-3, REQ-5): 프로젝트 agent 하나와 skill 하나를 `generation-apply` 한 뒤 다음 계획의 generation 단계가 `unchanged` 이고 모든 단계가 `unchanged` 예요.
  verify: `cargo test --locked --offline --test generation_request_record a_generated_agent_and_skill_read_unchanged_on_the_next_plan`
- AC-3 (REQ-6): 프로필의 시험 명령을 바꾸고 `project-apply` 한 뒤 generation 단계가 `ready` 이고 명령이 출력되며, 그 명령을 돌리면 다음 계획이 `unchanged` 예요.
  verify: `cargo test --locked --offline --test generation_request_record a_profile_change_plans_the_rebuild_of_a_recorded_skill`
- AC-4 (REQ-8, REQ-9): 두 번째 적용이 기록을 통째로 바꾸고, agents 쪽만의 적용은 기록을 그대로 둬요.
  verify: `cargo test --locked --offline --test generation_request_record a_later_apply_replaces_the_record_and_an_agents_apply_keeps_it`
- AC-5 (REQ-4): 기록이 없는 매니페스트(`needs_decision`)에 같은 요청을 다시 apply 하면 파일은 그대로이고 요청이 기록되며 다음 계획의 generation 단계가 `unchanged` 예요.
  verify: `cargo test --locked --offline --test generation_request_record an_apply_with_every_file_unchanged_still_records_the_request_and_clears_the_decision`
- AC-6 (REQ-7): 기록이 없는 매니페스트는 그대로 `needs_decision` 이에요.
  verify: `cargo test --locked --offline --test project_update_plan_cli unrebuildable_not_removed`
- AC-7 (REQ-10): 위 시험들이 실제 `generation-apply` 로 프로젝트 agent 를 만들어 통과해요(스키마가 새 키를 받아야 적용이 끝나요).
  verify: `cargo test --locked --offline --test generation_request_record`
- AC-8 (REQ-1…REQ-10): 이웃 검사.
  verify: `cargo test --locked --offline --test project_update_plan_cli`
  verify: `cargo test --locked --offline --test generation_manifest_single_write --test generation_apply_result_cli --test agents_apply_linkage --test generation_receipt_order --test generation_own_reference --test project_skill_links --test boundary_d3_skill_revert`
  verify: `cargo test --locked --offline --lib record_tests`

## Scope

- write: `.moai/specs/SPEC-20261125-generation-request-record/spec.md`, `tests/generation_request_record.rs`, `src/generation.rs`, `src/agents.rs`, `src/project_update.rs`
- read-only: `src/profile.rs`, `skills/moai/workflows/update.md`, `skills/moai/workflows/project.md`
- forbidden: 실제 HOME, 설치된 moai 바이너리, 다른 프로젝트, 매니페스트 스키마의 required 필드 추가, 카드 번호의 출력 사용

## Plan

1. 시험 먼저: `tests/generation_request_record.rs` 의 시험 다섯을 쓰고 실패(RED)를 읽어요.
2. `src/agents.rs` 의 타입 있는 `Manifest` 에 선택 필드 `generation_request` 를 더하고(`deny_unknown_fields` 라서 필드가 없으면 매니페스트를 못 읽어요) apply 가 이전 매니페스트에서 그대로 넘기게 해요.
3. `src/generation.rs` 의 `record_generation` 이 세 목록을 같은 매니페스트 쓰기에 적고(빈 요청이면 키를 빼요), 읽는 함수를 하나 둬요. 모든 파일이 unchanged 여도 매니페스트 쓰기가 일어나게 해요.
4. `src/project_update.rs` 의 `generation_request` 가 기록을 읽어 쓰고, 기록이 있으면 `UNREBUILDABLE` 문을 건너뛰어요. 기록이 없는 매니페스트는 그대로예요.
5. 이웃 검사와 READY 전 검사를 돌려요. 스킬 글이 안 바뀌어 매니페스트 해시는 쓰지 않아요.

## Risks

- 타입 있는 `Manifest` 가 모르는 키를 거절해요 → 필드를 더하지 않으면 적용 자체가 실패해요. 시험이 실제 적용으로 이를 잡아요.
- 매니페스트 바이트가 바뀌면 `generation` 계획의 "매니페스트 줄이 움직이는가" 판단이 어긋날 수 있어요 → 같은 함수(`record_generation`)가 쓰기와 예측에 쓰여서 어긋나지 않고, 적용 직후 계획이 `unchanged` 인 것을 AC-2 와 AC-5 가 봐요.
- 기록과 디스크가 어긋나는 경우(사용자가 agent 파일을 지움) → 기록으로 다시 계획하면 그 파일이 `create` 로 계획돼요. 사용자가 일부러 지운 파일을 되살리는 것이라 계획에서 보여 주고 적용은 사용자의 선택을 거쳐요(기존 `ready` 단계의 규칙).
- 기록에 요청 문자열이 남아요(역할 목적 한 줄들) → 이미 만든 파일에 같은 글이 들어 있고 비밀이 아니에요. 요청 검증이 쓰는 한 줄 한도를 적용 전에 거쳐요.
- 기록을 읽을 수 없으면(손으로 고친 매니페스트) → 추측하지 않고 사용자에게 묻는 `needs_decision` 으로 읽어요(fail-closed).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 요청은 프로필이 아니라 `generation.json` 에 적어요 | 프로필은 사용자의 답이 사는 곳이라 요청까지 넣으면 추적되는 `project.yaml` 이 커지고 사용자가 직접 고치기 쉬워요; 요청은 "마지막 generation 적용"에 속하는 연결 정보이고 `workflow_id`·`profile_sha256` 도 같은 곳에 적혀요 | 키를 지우면 옛 동작으로 돌아가요 |
| D-2 | 기록은 요청의 `native_roles`·`skill_projections`·`project_skills` 를 그 타입 그대로 담아요 | 하네스 제안과 기록 모양을 맞춰요; 같은 타입이면 필드 추가가 따로 형식을 안 만들어요 | 타입이 같으니 되돌릴 일이 없어요 |
| D-3 | 비어 있는 요청은 키를 쓰지 않아요 | 기록이 없는 프로젝트의 매니페스트 바이트와 계획 digest 가 그대로예요 | 빈 목록을 쓰도록 바꿔요 |
| D-4 | 기록이 없는 옛 매니페스트의 `needs_decision` 은 그대로 둬요 | 파일에서 요청을 거꾸로 만들면 사용자가 손본 파일이나 목적 한 줄을 추측해야 해요; 같은 요청으로 한 번 더 apply 하면 기록돼요(REQ-4) | 거꾸로 복원하는 별도 조각 |
| D-5 | 기록 키 이름은 `generation_request` | 앞서 만든 구현이 쓴 이름이고 `owned_files` 와 나란히 읽혀요; 처음 시험 초안이 가정한 `request` 는 `deny_unknown_fields` 스키마의 다른 키와 헷갈려요 | 키 이름 바꾸기는 읽기 함수 한 곳이에요 |

## Result

Implemented on branch `spec/generation-request-record-20261125`; the SPEC stays `draft` until the leader approves it at READY. The two earlier unlanded drafts of this fix are replaced by this one: the record write and read come from the first (`generation_request` key, `recorded_request_lists`), the apply with every file unchanged from the second; the `src/agents.rs` field that blocked the first is part of this change (REQ-10).

- Run 0 (RED, before any code change): `cargo nextest run --locked --offline --no-fail-fast --test generation_request_record` — 5 tests run, 0 passed, 5 failed, each for the right reason (no `generation_request` in the manifest; the generation step at `needs_decision` with "request is not recorded").
- Run 1: `cargo nextest run --locked --offline --no-fail-fast --test generation_request_record --lib -E 'binary(generation_request_record) or test(record_tests)'` — 8 tests run, 8 passed.
- Run 2: `cargo nextest run --locked --offline --no-fail-fast --test project_update_plan_cli --test generation_apply_result_cli --test agents_apply_linkage --test generation_manifest_single_write` — 58 tests run, 58 passed.
- Run 3: `cargo nextest run --locked --offline --no-fail-fast --test generation_request_record` — 5 tests run, 5 passed (after the added agents-apply assertion).
- Run 4: `cargo nextest run --locked --offline --no-fail-fast --test generation_receipt_order --test generation_own_reference --test project_skill_links --test boundary_d3_skill_revert` — 16 tests run, 16 passed.
- AC-1: met — `the_applied_request_is_recorded_as_the_request_itself`, passed in runs 1 and 3.
- AC-2: met — `a_generated_agent_and_skill_read_unchanged_on_the_next_plan`, passed in runs 1 and 3.
- AC-3: met — `a_profile_change_plans_the_rebuild_of_a_recorded_skill`, passed in runs 1 and 3.
- AC-4: met — `a_later_apply_replaces_the_record_and_an_agents_apply_keeps_it`, passed in runs 1 and 3; it also reads that an agents-only apply of a manifest with no record writes no key, not even a null.
- AC-5: met — `an_apply_with_every_file_unchanged_still_records_the_request_and_clears_the_decision`: a settled project loses its record (the generation step reads `needs_decision`), the same request is applied again, every owned file keeps its bytes, the record is back and the next plan reads `unchanged`. Passed in runs 1 and 3.
- AC-6: met — `unrebuildable_not_removed` passed in run 2.
- AC-7: met — the five tests of `generation_request_record` create the project agent and skill through the real `generation-apply`, which fails as an amended manifest outside the schema without the field; passed in run 3.
- AC-8: met — run 2 (58) and run 4 (16) passed, and the three `record_tests` unit tests passed in run 1.
- Guard mutants, one edit each, all seven run with the filter `binary(generation_request_record) or test(record_tests)` in one slot call, 7 of 7 killed (exit 100, the tree clean afterwards): `record_is_not_written`, `empty_request_keeps_the_key`, `record_skipped_when_every_file_is_unchanged`, `update_ignores_the_record`, `agents_apply_drops_the_record`, `empty_record_is_written_as_null`, `unreadable_record_passes`.
- Not run here: the whole affected check of the range (it runs at READY, and its receipt is quoted in the READY note). The dangerous-path kinds of the repository rules (delete, reset, start a process, build a permission rule) do not apply: the change writes one key of one manifest, and every test runs the built binary in a temporary project with a temporary HOME.
- Not done (Non-goals): restoring the request of an old manifest from its files; a manifest that lost its record is repaired by applying the same request once more.
