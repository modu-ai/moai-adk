---
id: SPEC-20261125-generation-empty-request
title: "generation-plan removes an owned project agent or skill projection file that the request no longer names, so an empty request leaves no needs_decision"
status: draft
revision: 1
created: 2026-11-25
branch: spec/generation-empty-request-20261125
---

# generation-plan removes an owned project agent or skill projection file that the request no longer names, so an empty request leaves no needs_decision

## Intent

SPEC-20261125-generation-request-record 재감사(head 10d2c688f7)가 남긴 길이에요. `generation-apply` 로 프로젝트 agent(role)와 skill 을 만든 뒤 **빈 요청**(role·skill·projection 이 하나도 없는 요청)을 적용하면, 계획은 skill 파일만 `remove` 로 내요(프로젝트 skill 은 "더 이상 요청되지 않음" 규칙이 있어요, #2706 P2b). role 파일과 projection 파일은 그런 규칙이 없어서 파일도 매니페스트의 소유 항목(`owned_files`)도 남아요. 그런데 빈 요청은 기록 키 `generation_request` 를 지워요(REQ-2 of SPEC-20261125-generation-request-record). 그 결과 소유 항목은 남고 기록은 없는 매니페스트가 되고, 다음 `project-update-plan` 의 generation 단계는 `needs_decision`(`not_rebuilt` 에 그 role 파일)에 영구히 머물러요. 앞선 SPEC 이 닫으려던 "needs_decision never clears" 가 이 길로 남아 있었어요.

세 갈래를 읽고 골랐어요 (D-1):

- 삭제: 요청에 없는 소유 파일은 프로젝트 skill 과 똑같이 `remove`(손대지 않은 파일) 또는 `conflict`(손댄 파일)로 계획해요. 미리보기에 `remove` 줄이 보이고 `generation-apply` 를 해야만 지워지고, 지우기 전 바이트는 영수증에 남아 `apply-recover` 로 되돌려요.
- 소유 해제: 파일은 두고 소유 항목만 매니페스트에서 빼요. 상태는 풀리지만 요청이 없애려던 파일이 아무도 관리하지 않는 채 남아요.
- 안내: 파일을 두고 상태도 두되 답에 푸는 방법을 적어요. 사용자가 손으로 지우거나 `generation-apply` 를 되풀이해야 해요.

삭제를 골랐어요. 프로젝트 skill 이 이미 같은 생명 주기를 쓰고 있어서(요청이 전체 원하는 상태이고, 빠진 소유 파일은 지워져요) role 과 projection 만 다르게 두면 같은 요청이 파일 종류에 따라 다르게 풀려요.

## Non-goals

- 기록이 없는 옛 매니페스트의 요청을 파일에서 복원하기(SPEC-20261125-generation-request-record D-4 가 그대로예요).
- 소유 항목이 아닌 파일(사용자의 `.claude/agents/*.md` 등)을 읽거나 지우기: 계획은 소유 항목에서만 시작해요.
- 하네스 전용 요청(`harness_only`)에서 소유 파일을 지우기: 그 요청은 파일에 대해 아무 말도 하지 않아요.
- `agents-apply`·`github-apply` 등 다른 명령의 help 예시: 카드는 `generation-*` 둘만 말해요.
- 기록 모양 검사 바꾸기(모양은 온전하지만 소유 파일보다 작은 기록은 D-4 에 판단만 적어요).

## Requirements

- REQ-1: When an owned native-role or skill-projection file is not named by the request, `generation-plan` shall plan `remove` for it while the file's bytes equal their recorded baseline. · failure: see REQ-2 and REQ-3.
- REQ-2 (REQ-1): IF an owned file that the request no longer names was edited after generation, THEN `generation-plan` shall plan `conflict` and keep the file.
- REQ-3 (REQ-1): IF the recorded path of such an entry is not a path its kind names, or a symbolic link is on that path, THEN `generation-plan` shall plan `conflict` and remove nothing through it.
- REQ-4: When `generation-plan` plans a removal, it shall write nothing, so that every owned file keeps its bytes until `generation-apply`. · failure: none — a test reads the files after the plan.
- REQ-5: When `generation-apply` applies such a `remove`, it shall keep the file's bytes in the receipt and drop its entry from `owned_files`. · failure: none — a test recovers the apply and reads the file back.
- REQ-6: When `generation-apply` applies an empty request to a project that owns a project agent, a project skill and a skill projection, the next `project-update-plan` shall read the generation step `unchanged`. · failure: none — a test plans after the apply.
- REQ-7: While a request is for the harness alone, `generation-plan` shall plan no removal of an owned file. · failure: none — a test plans a harness-only request over a project with agents.
- REQ-8: The `--help` examples of `generation-plan` and `generation-apply` shall show the input each command reads: the request for the plan and the plan for the apply. · failure: none — a test reads the help text.
- REQ-9: While the request still names a skill projection whose source is not license-reviewed, `generation-plan` shall plan no removal of that projection's file. · failure: none — a test plans the same request after the source loses its review.

## Acceptance

- AC-1 (REQ-1, REQ-4): in a temporary project with one project agent and one project skill generated, the plan of an empty request lists a `remove` row for the agent file and one for the skill file, and both files and the manifest keep their bytes after the plan · verify: `cargo test --locked --offline --test generation_request_record an_empty_request_previews_the_removal_of_the_agent_and_skill_files`
- AC-2 (REQ-5, REQ-6): after the apply of that empty request both files are gone, the manifest owns no generated file and holds no `generation_request`, the next `project-update-plan` reads the generation step `unchanged` (not `needs_decision`), and `apply-recover` puts the agent file back with its bytes · verify: `cargo test --locked --offline --test generation_request_record an_empty_request_clears_the_decision_and_the_removal_is_recoverable`
- AC-3 (REQ-2, REQ-3): an agent file edited after generation, and an owned entry whose recorded path is another path, each plan `conflict` with the file kept, and `generation-apply` of that plan stops with nothing applied · verify: `cargo test --locked --offline --test generation_request_record an_edited_or_misplaced_owned_agent_is_a_conflict_and_is_kept`
- AC-4 (REQ-1): a skill projection that the request no longer names is planned `remove` and, applied, leaves the manifest with no owned file · verify: `cargo test --locked --offline --test generation_request_record an_unrequested_skill_projection_is_removed_like_a_project_skill`
- AC-5 (REQ-7): a harness-only request over a project with agents and skills plans no `remove` · verify: `cargo test --locked --offline --test generation_request_record a_harness_only_request_plans_no_removal`
- AC-6 (REQ-8): `moai generation-plan --help` shows `< request.json` and `moai generation-apply --help` shows `< plan.json`, and no example of the apply shows `< request.json` · verify: `cargo test --locked --offline --test generation_request_record the_help_examples_name_the_input_each_command_reads`
- AC-7 (REQ-1…REQ-9): the neighbour tests of generation stay green · verify: `cargo test --locked --offline --test generation_request_record`
  verify: `cargo test --locked --offline --test project_update_plan_cli`
  verify: `cargo test --locked --offline --test generation_manifest_single_write --test generation_apply_result_cli --test agents_apply_linkage --test generation_receipt_order --test generation_own_reference --test project_skill_links --test boundary_d3_skill_revert --test generation_native_role_link --test project_agents_nested --test project_skills_generation`
  verify: `cargo test --locked --offline --lib record_tests`
- AC-8 (REQ-3): an owned agent whose folder is a symbolic link to a folder outside the project is planned `conflict` with `symbolic link` in its detail, `generation-apply` of that plan stops, and the file outside stays · verify: `cargo test --locked --offline --test generation_request_record an_owned_agent_behind_a_link_is_a_conflict_and_nothing_outside_is_removed`
- AC-9 (REQ-1): an owned agent recorded at the earlier flat path `.claude/agents/<id>.md` is planned `remove` when the request no longer names it · verify: `cargo test --locked --offline --test generation_request_record an_unrequested_role_at_its_earlier_flat_path_is_removed`
- AC-10 (REQ-9): a projection that the request still names but whose source lost its license review is not planned and not removed, and the plan warns `not license-reviewed` · verify: `cargo test --locked --offline --test generation_request_record a_projection_the_request_names_but_cannot_plan_is_not_removed`

## Scope

- write: `.moai/specs/SPEC-20261125-generation-empty-request/spec.md`, `tests/generation_request_record.rs`, `src/generation.rs`
- shared (line lease): `src/help.rs` (the two example lines of `generation-plan` and `generation-apply`)
- read-only: `src/project_update.rs`, `src/agents.rs`, `src/ops/generation_apply.rs`, `src/ops/generation_plan.rs`, `skills/moai/workflows/project.md`
- forbidden: the real HOME, the installed moai binary, other projects, the manifest schema's required fields, card numbers in files

## Plan

1. RED first: the tests of AC-1…AC-6 in `tests/generation_request_record.rs` run the built binary in a temporary project; on the base they fail for the right reason (no `remove` row for the agent file; `needs_decision` after the empty apply; the apply help names `request.json`).
2. `src/generation.rs`: the loop that plans the project skills no longer named (`#2706 P2b`) takes the three kinds `generation:project-skill:`, `generation:native-role:` and `generation:skill-projection:`. What the request names is built per kind; the recorded path is trusted only when it is a path the kind names (a native role of Claude Code: `.claude/agents/moai/<id>.md` or the earlier flat `.claude/agents/<id>.md`; a projection: `.claude/skills/<id>/SKILL.md`). The rest of the loop (link check, edited → conflict, `remove` otherwise) is the one that runs today.
3. `src/help.rs`: the apply examples read `< plan.json`, the way `project-apply` already does.
4. The comment of `tests/generation_request_record.rs` that says the key never outlives its files is rewritten to say what holds: a request with nothing left drops the record and plans the removal of the files it named.
5. Guard mutants for each guard, the neighbour tests, then one bare `check affected` at READY.

## Risks

- A request that leaves out an agent by mistake now plans its removal. The removal is a plan line the user applies, never a silent delete; the file's bytes go to the receipt and `apply-recover` restores it; an edited file is a conflict that stops the apply.
- An older manifest records a native role at the flat path `.claude/agents/<id>.md`. Without the flat path in the trusted list its entry would turn into a conflict for a retired role; the flat path is accepted for Claude Code roles for that reason.
- A projection whose source is not license-reviewed is "not planned" and warns; it stays requested, so its owned file is not removed because of the missing review.
- A hand-edited record that holds fewer agents than the manifest owns plans the removal of the others (D-4); the removal is previewed and needs the user's apply.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A native-role or skill-projection file the request no longer names is planned like a project skill: `remove` when its bytes equal the baseline, `conflict` when edited or reached through a link | lane reading of `src/generation.rs` (the project-skill loop, `#2706 P2b`) and of SPEC-20261125-generation-request-record (REQ-2: an empty request drops the record, so a retained owned entry can never be rebuilt); the card asks that a growing deletion show a `remove` line and wait for the user's apply, which a plan row does | plan nothing for the two kinds again (the loop takes the project-skill kind only) |
| D-2 | The scope of the fix is the plan, not the project-update reader | the stuck state exists because the manifest keeps owned entries without a record; after the apply there are none, so `unrebuildable_paths` is empty and `src/project_update.rs` needs no change | add a release path to `project_update` |
| D-3 | The apply examples of `generation-*` read `< plan.json` | `generation-apply` reads `GenerationPlan` through `plan_input` (`src/ops/generation_apply.rs`), the way `project-apply` does, and `project-plan` already prints `< plan.json` for its apply | restore `< request.json` |
| D-4 | No new check against a record that is whole in shape but smaller than the owned files (hand edit) | the record is read only when it has exactly the three lists and at least one entry (SPEC-20261125-generation-request-record REQ-11, D-6); a smaller record would plan the removal of the other owned files, and those removals are plan rows the user applies, with the bytes kept in the receipt; a check would need a digest of what the apply wrote, which is a new manifest field and a schema change beyond this card | add a digest of the record to the manifest and read it in `recorded_request_lists` |

## Result

`spec/generation-empty-request-20261125` 가지에서 구현했어요. 기준은 앞선 SPEC(`SPEC-20261125-generation-request-record`)의 감사 통과 head `10d2c688f7` 이고, 이 SPEC 은 승인 전 `draft` 로 두어요.

- Run 0 (RED, 코드를 고치기 전): `cargo nextest run --locked --offline --test generation_request_record --no-fail-fast` — 13 tests run, 8 passed, 5 failed. 다섯 모두 맞는 이유로 실패했어요: 빈 요청의 계획에 agent·projection 파일의 `remove` 줄이 없고, 적용한 뒤에도 agent 파일이 남고, 적용 help 가 `request.json` 이에요. 출력은 `target/agent-3/red1.out`.
- Run 1: 같은 명령, 코드를 고친 뒤 — 12 passed, 1 failed (`the_help_examples_name_the_input_each_command_reads` 가 `Usage:` 줄까지 예시로 읽었어요). 시험을 ` < ` 가 있는 예시 줄로 좁혔어요.
- Run 2: 이웃 20개 시험 대상과 lib 전체를 한 슬롯으로 — 2063 tests run, 2063 passed (`target/agent-3/neighbours1.out`).
- Run 3: 링크·평평한 경로·검토를 잃은 projection 시험 3개를 더한 뒤 같은 파일 — 16 passed (`target/agent-3/green3.out`).
- 검증 명령 13개를 이 SPEC 에 적힌 그대로(`cargo test --locked --offline …`) 한 슬롯 호출로 돌렸어요. 모두 exit 0 이고 출력은 `target/agent-3/verify_ac1.out` 이에요 (`target/agent-3/verify_ac1.py` 가 SPEC 에서 명령을 읽어요).
- AC-1: met — `an_empty_request_previews_the_removal_of_the_agent_and_skill_files`: 빈 요청의 계획이 agent 파일과 skill 파일에 `remove` 줄을 내고, 계획 뒤에도 두 파일과 매니페스트의 바이트가 그대로예요.
- AC-2: met — `an_empty_request_clears_the_decision_and_the_removal_is_recoverable`: 적용 뒤 두 파일이 없고 소유 항목과 기록 키가 없으며, 다음 `project-update-plan` 의 generation 단계가 `unchanged` 예요. `apply-recover` 가 두 파일을 바이트 그대로 되돌려요.
- AC-3: met — `an_edited_or_misplaced_owned_agent_is_a_conflict_and_is_kept`: 손댄 agent 파일은 `conflict`(detail 에 `edited`)이고 적용은 아무것도 하지 않은 채 멈추며, 경로가 다른 소유 항목은 `conflict` 예요. 요청이 아직 이름 붙인 role 의 항목은 판단하지 않아요.
- AC-4: met — `an_unrequested_skill_projection_is_removed_like_a_project_skill`.
- AC-5: met — `a_harness_only_request_plans_no_removal`.
- AC-6: met — `the_help_examples_name_the_input_each_command_reads`.
- AC-7: met — `generation_request_record` 16, `project_update_plan_cli` 52, 이웃 대상 10개와 lib `record_tests` 4 모두 통과.
- AC-8: met — `an_owned_agent_behind_a_link_is_a_conflict_and_nothing_outside_is_removed`: 폴더가 프로젝트 밖으로 걸린 링크여도 계획은 `conflict`(`symbolic link`), 적용은 멈추고 밖의 파일은 남아요.
- AC-9: met — `an_unrequested_role_at_its_earlier_flat_path_is_removed`.
- AC-10: met — `a_projection_the_request_names_but_cannot_plan_is_not_removed`.
- 가드 뮤턴트(한 번에 한 줄씩, 한 슬롯 호출, 출력 `target/agent-3/mutants1.out`, 스크립트 `target/agent-3/mutants1.py`): 10개 중 10개 killed(exit 100, 끝난 뒤 트리 깨끗): M1 agent 종류를 지움, M2 projection 종류를 지움, M3 요청의 projection 을 빼먹음, M4 요청의 role 을 빼먹음, M5 평평한 role 경로를 신뢰 목록에서 뺌, M6 기록된 경로를 그대로 믿음, M7 경로의 링크를 거절하지 않음, M8 손댄 파일을 지움, M9 harness-only 에서도 지움, M10 지운 파일이 `owned_files` 에 남음. help 예시 뮤턴트 1개(`target/agent-3/mutants_help1.py`, 공유 파일이라 줄 임대 아래서, 출력 `target/agent-3/mutants_help1.out`): killed.
- 위험 경로: 삭제를 늘리는 변경이라 시험이 빌드된 `moai` 바이너리를 버리는 프로젝트·버리는 HOME 에서 실제로 돌려요 — 빈 요청 적용과 `apply-recover`, 손댄 파일, 프로젝트 밖으로 걸린 링크, 평평한 옛 경로. 파일은 `generation-apply` 가 있어야만 지워지고, 지우기 전 바이트는 영수증에 남아요.
- 항목 4 판단(D-4): 모양은 온전하지만 소유 파일보다 작은 손편집 기록은 막는 검사를 더하지 않아요. 다른 소유 파일의 삭제는 미리보기에 `remove` 줄로 보이고 사용자가 적용해야 하며 영수증으로 되돌려요.
- 본 것, 이 SPEC 밖: `agents-apply` 와 `github-apply` 의 help 예시도 `< request.json` 이에요(`src/help.rs` 332-337줄). 두 명령도 계획을 읽는지 확인한 뒤 따로 고칠 후속이에요.
- 하지 않은 것: 전체 `cargo test`(리더 몫), 넓은 `check affected`(READY 에서 한 번 돌려 노트에 인용), Windows·Linux.
