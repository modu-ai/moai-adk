---
id: SPEC-20261125-generation-empty-request
title: "요청에서 빠진 소유 agent·skill 파일을 계획이 지울 파일로 올려서, 빈 요청이 needs_decision 을 남기지 않아요"
status: draft
revision: 2
created: 2026-11-25
branch: spec/generation-empty-request-20261125
---

# 요청에서 빠진 소유 agent·skill 파일을 계획이 지울 파일로 올려서, 빈 요청이 needs_decision 을 남기지 않아요

## Intent

SPEC-20261125-generation-request-record 재감사(head 10d2c688f7)가 남긴 길이에요. `generation-apply` 로 프로젝트 agent(role)와 skill 을 만든 뒤 **빈 요청**(role·skill·projection 이 하나도 없는 요청)을 적용하면, 계획은 skill 파일만 `remove` 로 내요(프로젝트 skill 은 "더 이상 요청되지 않음" 규칙이 있어요, #2706 P2b). role 파일과 projection 파일은 그런 규칙이 없어서 파일도 매니페스트의 소유 항목(`owned_files`)도 남아요. 그런데 빈 요청은 기록 키 `generation_request` 를 지워요(앞선 SPEC 의 REQ-2). 그 결과 소유 항목은 남고 기록은 없는 매니페스트가 되고, 다음 `project-update-plan` 의 generation 단계는 `needs_decision`(`not_rebuilt` 에 그 role 파일)에 영구히 머물러요. 앞선 SPEC 이 닫으려던 "needs_decision never clears" 가 이 길로 남아 있었어요.

세 갈래를 읽고 골랐어요 (D-1):

- 삭제: 요청에 없는 소유 파일은 프로젝트 skill 과 똑같이 `remove`(손대지 않은 파일) 또는 `conflict`(손댄 파일)로 계획해요. 미리보기에 `remove` 줄이 보이고 `generation-apply` 를 해야만 지워지고, 지우기 전 바이트는 영수증에 남아 `apply-recover` 로 되돌려요.
- 소유 해제: 파일은 두고 소유 항목만 매니페스트에서 빼요. 상태는 풀리지만 요청이 없애려던 파일이 아무도 관리하지 않는 채 남아요.
- 안내: 파일을 두고 상태도 두되 답에 푸는 방법을 적어요. 사용자가 손으로 지우거나 `generation-apply` 를 되풀이해야 해요.

삭제를 골랐어요. 프로젝트 skill 이 이미 같은 생명 주기를 쓰고 있어서, role 과 projection 만 다르게 두면 같은 요청이 파일 종류에 따라 다르게 풀려요.

삭제를 넓히면 두 가지를 같이 막아야 해요 (개정 2, 독립 감사가 재현했어요):

- 지울 경로를 만드는 재료가 추적되는 매니페스트 안의 글이에요. 항목의 kind 안의 id 와 path 를 함께 바꾸면(`../` 포함) 프로젝트의 `README.md`·`CLAUDE.md`, 프로젝트 밖의 `.md` 파일이 `remove` 로 계획되고 적용하면 지워졌어요. 이 구멍은 이 변경 전부터 프로젝트 skill 에 `SKILL.md` 라는 이름의 파일에 한해 있었고, 이 변경이 임의의 `.md` 와 프로젝트 밖으로 넓혔어요. id 는 한 경로 조각이어야 하고 지울 경로는 그 종류의 폴더 안이어야 해요(REQ-3, 세 종류 모두).
- 첫 실행(`moai cc|glm`)이 빈 요청을 미리보기 없이 적용해요. 소유 파일이 있는데 설정이 없는 프로젝트(예: harness-only 로만 만든 뒤 설정을 잃은 프로젝트)에서는 그 적용이 소유 파일을 지울 수 있어요. 미리보기가 없는 길에서는 삭제를 적용하지 않아요(REQ-10).

## Non-goals

- 기록이 없는 옛 매니페스트의 요청을 파일에서 복원하기(앞선 SPEC 의 D-4 가 그대로예요).
- 소유 항목이 아닌 파일(사용자의 `.claude/agents/*.md` 등)을 읽거나 지우기: 계획은 소유 항목에서만 시작해요.
- 하네스 전용 요청(`harness-only`)에서 소유 파일을 지우기: 그 요청은 파일에 대해 아무 말도 하지 않아요.
- `agents-apply`·`github-apply` 등 다른 명령의 help 예시: 카드는 `generation-*` 둘만 말해요.
- 기록 모양 검사 바꾸기(모양은 온전하지만 소유 파일보다 작은 기록은 D-4 에 판단만 적어요).
- `apply-recover` 가 되돌린 파일의 권한 비트 복원(Risks 에 적어요).

## Requirements

- REQ-1: When an owned native-role or skill-projection file is not named by the request, `generation-plan` shall plan `remove` for it while the file's bytes equal their recorded baseline. · failure: REQ-2 와 REQ-3 이에요.
- REQ-2 (REQ-1): IF an owned file that the request no longer names was edited after generation, THEN `generation-plan` shall plan `conflict` and keep the file.
- REQ-3 (REQ-1): IF the id in the kind of an owned entry is not one plain path piece, or its recorded path is not one the kind builds from that id below the kind's folder, or a symbolic link is on that path, THEN `generation-plan` shall plan `conflict` and remove nothing through it.
- REQ-4: When `generation-plan` plans a removal, it shall write nothing, so that every owned file keeps its bytes until `generation-apply`. · failure: none — 테스트가 계획 뒤에 파일을 읽어요.
- REQ-5: When `generation-apply` applies such a `remove`, it shall keep the file's bytes in the receipt and drop its entry from `owned_files`. · failure: none — 테스트가 적용을 되돌려 파일을 읽어요.
- REQ-6: When `generation-apply` applies an empty request to a project that owns a project agent, a project skill and a skill projection, the next `project-update-plan` shall read the generation step `unchanged`. · failure: none — 테스트가 적용 뒤에 계획을 읽어요.
- REQ-7: While a request is for the harness alone, `generation-plan` shall plan no removal of an owned file. · failure: none — 테스트가 agent 가 있는 프로젝트에 harness-only 요청을 계획해요.
- REQ-8: The `--help` examples of `generation-plan` and `generation-apply` shall show the input each command reads: the request for the plan and the plan for the apply. · failure: none — 테스트가 help 글을 읽어요.
- REQ-9: While the request still names a skill projection whose source is not license-reviewed, `generation-plan` shall plan no removal of that projection's file. · failure: none — 테스트가 소스가 검토를 잃은 뒤 같은 요청을 계획해요.
- REQ-10: When a first launch plans its empty request and the plan holds a `remove` line, the launcher shall apply nothing of that plan and say that `CLAUDE.md` was not generated. · failure: none — 파일은 그대로 두고, 다음 `moai update` 의 미리보기가 그 삭제를 보여 줘요.

## Acceptance

- AC-1 (REQ-1, REQ-4): 에이전트 하나와 스킬 하나를 만든 임시 프로젝트에서 빈 요청의 계획이 agent 파일과 skill 파일에 `remove` 줄을 내고, 계획 뒤에도 두 파일과 매니페스트의 바이트가 그대로예요 · verify: `cargo test --locked --offline --test generation_request_record an_empty_request_previews_the_removal_of_the_agent_and_skill_files`
- AC-2 (REQ-5, REQ-6): 그 빈 요청을 적용하면 두 파일이 없고, 매니페스트가 생성 파일을 소유하지 않고 `generation_request` 도 없으며, 다음 `project-update-plan` 의 generation 단계가 `needs_decision` 이 아니라 `unchanged` 이고, `apply-recover` 가 agent 파일을 바이트 그대로 되돌려요 · verify: `cargo test --locked --offline --test generation_request_record an_empty_request_clears_the_decision_and_the_removal_is_recoverable`
- AC-3 (REQ-2, REQ-3): 생성 뒤 손댄 agent 파일과 기록된 경로가 다른 소유 항목이 각각 `conflict` 이고 파일은 그대로이며, 그 계획의 `generation-apply` 는 아무것도 적용하지 않고 멈춰요 · verify: `cargo test --locked --offline --test generation_request_record an_edited_or_misplaced_owned_agent_is_a_conflict_and_is_kept`
- AC-4 (REQ-1): 요청에서 빠진 skill projection 은 `remove` 로 계획되고, 적용하면 매니페스트에 소유 파일이 없어요 · verify: `cargo test --locked --offline --test generation_request_record an_unrequested_skill_projection_is_removed_like_a_project_skill`
- AC-5 (REQ-7): agent 와 skill 이 있는 프로젝트에 harness-only 요청을 계획하면 `remove` 가 없어요 · verify: `cargo test --locked --offline --test generation_request_record a_harness_only_request_plans_no_removal`
- AC-6 (REQ-8): `moai generation-plan --help` 는 `< request.json`, `moai generation-apply --help` 는 `< plan.json` 을 보이고, 적용 예시 어디에도 `< request.json` 이 없어요 · verify: `cargo test --locked --offline --test generation_request_record the_help_examples_name_the_input_each_command_reads`
- AC-7 (REQ-1…REQ-10): generation 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test generation_request_record`
  verify: `cargo test --locked --offline --test project_update_plan_cli`
  verify: `cargo test --locked --offline --test generation_manifest_single_write --test generation_apply_result_cli --test agents_apply_linkage --test generation_receipt_order --test generation_own_reference --test project_skill_links --test boundary_d3_skill_revert --test generation_native_role_link --test project_agents_nested --test project_skills_generation --test launcher_auto_init`
  verify: `cargo test --locked --offline --lib record_tests`
- AC-8 (REQ-3): 폴더가 프로젝트 밖 폴더로 걸린 링크인 소유 agent 는 detail 에 `symbolic link` 가 있는 `conflict` 이고, 그 계획의 `generation-apply` 는 멈추고, 밖의 파일은 그대로예요 · verify: `cargo test --locked --offline --test generation_request_record an_owned_agent_behind_a_link_is_a_conflict_and_nothing_outside_is_removed`
- AC-9 (REQ-1): 옛 평평한 경로 `.claude/agents/<id>.md` 로 기록된 소유 agent 는 요청에서 빠지면 `remove` 로 계획돼요 · verify: `cargo test --locked --offline --test generation_request_record an_unrequested_role_at_its_earlier_flat_path_is_removed`
- AC-10 (REQ-9): 요청이 아직 이름 붙였지만 소스가 라이선스 검토를 잃은 projection 은 계획에 오르지 않고 지워지지도 않으며, 계획이 `not license-reviewed` 를 경고해요 · verify: `cargo test --locked --offline --test generation_request_record a_projection_the_request_names_but_cannot_plan_is_not_removed`
- AC-11 (REQ-3): kind 의 id 에 `../` 나 구분자가 든 소유 항목(프로젝트 안 `README.md`, 프로젝트 밖 `.md`, project-skill 의 `SKILL.md`, 구분자가 있되 경로는 폴더 안인 항목)이 모두 `conflict` 이고, 적용은 멈추고, 파일이 그대로이며, 업데이트 미리보기가 삭제를 내놓지 않아요 · verify: `cargo test --locked --offline --test generation_request_record an_owned_entry_whose_id_leaves_its_folder_is_a_conflict_and_the_file_stays`
- AC-12 (REQ-3): id 검사와 폴더 검사 함수가 한 경로 조각, 폴더 안 경로, 종류가 이름 붙이는 경로를 각각 가려내요 · verify: `cargo test --locked --offline --lib owned_path_tests`
- AC-13 (REQ-10): harness-only 로 만들어 소유 agent·skill 이 있고 설정과 지시 파일이 없는 프로젝트에서 첫 실행이 `CLAUDE.md not generated` 를 말하고 소유 파일을 그대로 둬요 · verify: `cargo test --locked --offline --test generation_request_record a_first_launch_never_applies_the_removal_of_an_owned_file`

## Scope

- write: `.moai/specs/SPEC-20261125-generation-empty-request/spec.md`, `tests/generation_request_record.rs`, `src/generation.rs`, `src/launcher_init.rs`
- shared (줄 임대): `src/help.rs` (`generation-plan`·`generation-apply` 예시 두 줄)
- read-only: `src/project_update.rs`, `src/agents.rs`, `src/apply.rs`, `src/ops/generation_apply.rs`, `src/ops/generation_plan.rs`
- forbidden: 진짜 HOME, 설치된 moai 바이너리, 다른 프로젝트, 매니페스트 스키마의 필수 필드, 파일 안의 카드 번호

## Plan

1. RED 먼저: AC-1…AC-6 의 테스트가 임시 프로젝트에서 빌드된 바이너리를 돌려요. 기준에서는 맞는 이유로 실패해요(agent 파일의 `remove` 줄이 없음, 빈 요청 적용 뒤 `needs_decision`, 적용 help 가 `request.json`).
2. `src/generation.rs`: 더 이상 요청되지 않은 프로젝트 skill 을 계획하는 반복문(#2706 P2b)이 세 종류 `generation:project-skill:`·`generation:native-role:`·`generation:skill-projection:` 를 받아요. 요청이 이름 붙인 것은 종류별로 만들고, 기록된 경로는 `kind_names_path` 가 믿을 때만 써요: id 는 한 경로 조각(`one_path_piece`), 경로는 그 id 로 종류가 만드는 경로 중 하나(Claude Code role 은 `.claude/agents/moai/<id>.md` 또는 옛 평평한 `.claude/agents/<id>.md`, projection·skill 은 `.claude/skills/...`), 그리고 `.`·`..` 를 접은 결과가 종류의 폴더 안(`stays_below`)이에요. 나머지(링크 검사, 손댄 파일은 conflict, 그 밖은 `remove`)는 지금 도는 코드 그대로예요.
3. `src/launcher_init.rs`: 첫 실행의 빈 요청 계획에 `remove` 줄이 있으면 아무것도 적용하지 않고 `CLAUDE.md not generated` 를 말해요(충돌 때와 같은 방식).
4. `src/help.rs`: 적용 예시가 `project-apply` 처럼 `< plan.json` 을 읽어요.
5. 테스트 파일의 틀린 주석을 사실대로 고쳐요.
6. 지킴이마다 뮤턴트, 이웃 테스트, READY 에서 `check affected` 한 번.

## Risks

- 실수로 agent 를 빼먹은 요청이 이제 그 agent 의 삭제를 계획해요. 삭제는 사용자가 적용하는 계획 줄이지 조용한 삭제가 아니에요. 바이트는 영수증에 남아 `apply-recover` 가 되돌리고, 손댄 파일은 적용을 멈추는 `conflict` 예요.
- 옛 매니페스트는 native role 을 평평한 경로 `.claude/agents/<id>.md` 로 기록해요. 그 경로를 믿는 목록에서 빼면 사라진 role 의 항목이 conflict 로 바뀌어서, Claude Code role 에는 평평한 경로를 받아요.
- 소스가 라이선스 검토를 잃은 projection 은 "계획에 오르지 않고" 경고해요. 요청에는 남아 있으니 소유 파일은 검토가 없다는 이유로 지워지지 않아요.
- 손으로 줄인 기록(모양은 온전하지만 소유 파일보다 작음)은 나머지 소유 파일의 삭제를 계획해요(D-4). 삭제는 미리보기에 보이고 사용자의 적용이 있어야 해요.
- 첫 실행 경로는 미리보기가 없어서 삭제가 있는 계획을 적용하지 않아요(D-6). 그 프로젝트의 `CLAUDE.md` 는 첫 실행에서 만들어지지 않고, 사용자가 `moai update` 로 삭제 미리보기를 본 뒤 정해요.
- `apply-recover` 는 되돌린 파일의 권한 비트를 되돌리지 않아요(0o640 이던 파일이 기본 모드로 돌아와요). 영수증 스키마가 바이트의 digest 만 담고(`deny_unknown_fields`) 모든 apply 엔진이 같이 쓰는 인터페이스라, 권한을 담는 일은 이 카드의 범위 밖이에요. 데이터가 아니라 모드만 달라져요.
- 매니페스트는 추적되는 파일이라 받아 온 커밋이 항목을 심을 수 있어요. 그래서 kind 의 id 와 path 는 둘 다 믿지 않고(D-5), 링크 검사와 적용 때의 해시 비교(`stale_plan`)가 그 뒤를 받쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 요청에서 빠진 native-role·skill-projection 파일은 프로젝트 skill 처럼 계획해요: 바이트가 기준과 같으면 `remove`, 손댔거나 링크 뒤면 `conflict` | `src/generation.rs` 의 프로젝트 skill 반복문(#2706 P2b)과 앞선 SPEC 의 REQ-2(빈 요청은 기록을 지워서, 남은 소유 항목은 다시 만들 수 없어요)를 읽은 결과예요. 카드의 "삭제가 늘면 `remove` 줄을 보이고 적용 전에는 그대로"는 계획 줄이 해요 | 두 종류를 다시 계획하지 않아요(반복문이 project-skill 종류만 받음) |
| D-2 | 고치는 곳은 계획이고 project-update 읽기가 아니에요 | 막힌 상태는 매니페스트가 기록 없이 소유 항목을 가져서 생겨요. 적용 뒤에는 소유 항목이 없어 `unrebuildable_paths` 가 비고 `src/project_update.rs` 는 바꿀 것이 없어요 | `project_update` 에 놓아 주는 길을 더해요 |
| D-3 | `generation-*` 의 적용 예시는 `< plan.json` 을 읽어요 | `generation-apply` 가 `plan_input` 으로 `GenerationPlan` 을 읽어요(`src/ops/generation_apply.rs`). `project-apply` 도 같고 `project-plan` 은 이미 적용 예시에 `< plan.json` 을 써요 | `< request.json` 으로 되돌려요 |
| D-4 | 모양은 온전하지만 소유 파일보다 작은 손편집 기록을 막는 검사는 더하지 않아요 | 기록은 세 목록이 모두 있고 항목이 하나 이상일 때만 읽어요(앞선 SPEC 의 REQ-11, D-6). 더 작은 기록은 나머지 소유 파일의 삭제를 계획하는데, 그 삭제는 사용자가 적용하는 계획 줄이고 바이트는 영수증에 남아요. 검사를 하려면 적용이 쓴 내용의 digest 가 매니페스트의 새 필드로 필요하고 그것은 이 카드를 넘는 스키마 변경이에요 | 매니페스트에 기록의 digest 를 담고 `recorded_request_lists` 가 읽어요 |
| D-5 | 지울 경로는 kind 의 id 와 path 를 둘 다 믿지 않고 만들어요: id 는 한 경로 조각, 경로는 그 id 로 종류가 만드는 경로 중 하나이고 종류의 폴더 안이에요. 세 종류 모두 같은 검사를 써요 | 독립 감사가 id 에 `../` 를 넣고 path 를 맞춘 항목으로 `README.md`·`CLAUDE.md`·프로젝트 밖 `.md` 가 `remove` 로 계획되고 지워짐을 재현했어요. 매니페스트는 추적 파일이라 받아 온 커밋이 항목을 심을 수 있고 kind 와 path 는 같은 항목에 있어서 함께 바꿀 수 있어요. 만든 경로와의 비교만으로 폴더 안임이 따라오지만, 접은 결과의 검사(`stays_below`)를 따로 두어 한 겹이 빠져도 밖으로 못 나가요 | 세 종류의 `named_paths` 에서 id 검사를 빼요 |
| D-6 | 첫 실행의 빈 요청 계획에 `remove` 줄이 있으면 아무것도 적용하지 않아요 | `src/launcher_init.rs` 의 `agents` 는 미리보기 없이 `generation_apply` 를 불러요. 소유 agent·skill 이 있고 설정·지시 파일이 없는 프로젝트(harness-only 로 만든 뒤 설정을 잃은 경우)에서 이 길이 소유 파일을 지울 수 있음을 테스트가 재현했어요. 계획의 줄을 빼면 digest 가 달라져 적용이 거절되므로 줄만 거를 수는 없고, 충돌 때처럼 적용을 건너뛰어요. 지울지는 다음 `moai update` 의 미리보기에서 사용자가 정해요 | 그 분기를 빼요 |

## Result

`spec/generation-empty-request-20261125` 브랜치에서 구현했어요. 기준은 앞선 SPEC(`SPEC-20261125-generation-request-record`)의 감사 통과 head `10d2c688f7` 이고, 이 SPEC 은 승인 전 `draft` 로 두어요.

### 개정 1 (head 8eed93d693)

- Run 0 (RED, 코드를 고치기 전): `cargo nextest run --locked --offline --test generation_request_record --no-fail-fast` — 13 tests run, 8 passed, 5 failed. 다섯 모두 맞는 이유로 실패했어요: 빈 요청의 계획에 agent·projection 파일의 `remove` 줄이 없고, 적용한 뒤에도 agent 파일이 남고, 적용 help 가 `request.json` 이에요. 출력은 `target/agent-3/red1.out`.
- Run 1: 같은 명령, 코드를 고친 뒤 — 12 passed, 1 failed(`the_help_examples_name_the_input_each_command_reads` 가 `Usage:` 줄까지 예시로 읽었어요). 테스트를 ` < ` 가 있는 예시 줄로 좁혔어요.
- Run 2: 이웃 20개 테스트 대상과 lib 전체를 한 슬롯으로 — 2063 tests run, 2063 passed(`target/agent-3/neighbours1.out`).
- Run 3: 링크·평평한 경로·검토를 잃은 projection 테스트 3개를 더한 뒤 같은 파일 — 16 passed(`target/agent-3/green3.out`).
- 검증 명령 13개를 이 SPEC 에 적힌 그대로 한 슬롯 호출로 돌렸어요. 모두 exit 0, 출력은 `target/agent-3/verify_ac1.out` 이에요.
- AC-1…AC-10: met(개정 1 head 에서 각 테스트가 통과).
- 가드 뮤턴트 10개 모두 killed(출력 `target/agent-3/mutants1.out`), help 예시 뮤턴트 1개 killed(`target/agent-3/mutants_help1.out`).
- 항목 4 판단(D-4): 모양은 온전하지만 소유 파일보다 작은 손편집 기록은 막는 검사를 더하지 않아요.
- 본 것, 이 SPEC 밖: `agents-apply` 와 `github-apply` 의 help 예시도 `< request.json` 이에요(`src/help.rs` 332-337줄). 두 명령도 계획을 읽는지 확인한 뒤 따로 고칠 후속이에요.
- 개정 1 에서 고친 뒤 더한 테스트 넷(링크, 평평한 경로, 검토를 잃은 projection, harness-only)은 RED 가 없고 뮤턴트로만 받쳐져요.

### 개정 2 (독립 감사 FAIL 의 수리: B1 경로 밖 삭제, B2 첫 실행의 미리보기 없는 삭제)

- Run 4 (RED, 고치기 전): `cargo nextest run --locked --offline --test generation_request_record --no-fail-fast` — 18 tests run, 16 passed, 2 failed, 맞는 이유로 실패했어요: 항목의 id 에 `../` 를 넣으면 `README.md` 가 `remove` 로 계획되고, 첫 실행이 `CLAUDE.md` 를 만들며 소유 파일을 지워요(`target/agent-3/red5.out`). 첫 실행 테스트는 처음에 앞 단계(`project settings plan is conflict`, 다음 `generation manifest belongs to another schema or project`)에서 멈췄어요. 도달 조건은 harness-only 로 만들어 `CLAUDE.md` 소유가 없고, 설정과 기준선을 잃고, 프로젝트 id 가 폴더 이름에서 만든 id 와 같은 프로젝트예요.
- Run 5: 같은 파일과 새 단위 테스트 — 25 tests run, 25 passed(`target/agent-3/green4.out`).
- Run 6: 이웃 21개 테스트 대상(첫 실행 `launcher_auto_init` 포함)과 lib 전체를 한 슬롯으로 — 2092 tests run, 2092 passed(`target/agent-3/neighbours2.out`). clippy `-D warnings` exit 0(`target/agent-3/clippy2.out`).
- 검증 명령 16개를 이 SPEC 에 적힌 그대로 한 슬롯 호출로 돌렸어요. 모두 exit 0, 출력은 `target/agent-3/verify_ac2.out` 이에요.
- AC-1…AC-13: met(개정 2 head 에서 각 테스트가 통과).
- 수리 가드 뮤턴트 8개 모두 killed(exit 100, 끝난 뒤 트리 깨끗, 출력 `target/agent-3/mutants2.out`): N1 id 검사를 뺌, N2 폴더 검사를 뺌, N3 폴더 자신을 안으로 셈, N4 native-role id 검사를 뺌, N5 project-skill id 검사를 뺌, N6 skill-projection id 검사를 뺌, N7 id 와 폴더 검사를 모두 뺌(감사의 세 모양), N8 첫 실행이 삭제 계획을 적용함.
- `kind_names_path` 가 `stays_below` 를 부르는 줄만 빼는 뮤턴트는 만들지 않았어요: id 검사가 있으면 만든 경로와의 비교가 이미 폴더 안임을 보장해서 그 한 줄을 뺀 코드는 같은 결과를 내요(동치). 폴더 검사는 함수 자체를 단위 테스트와 N2·N3 이 받쳐요.
- 위험 경로: 삭제 경로 검사를 바꾸는 변경이라 테스트가 빌드된 `moai` 바이너리를 버리는 프로젝트·버리는 HOME 에서 실제로 돌려요. `../` 가 든 id 로 프로젝트 안 `README.md`, 프로젝트 밖의 `.md`, project-skill 의 `SKILL.md` 를 가리키는 항목을 심어 계획이 `conflict` 이고 적용이 멈추며 파일이 그대로임을 읽어요. 첫 실행은 진짜 `moai cc` 를 버리는 프로젝트에서 돌려요.
- 작은 것: `apply-recover` 의 파일 권한 비트는 되돌리지 않고 Risks 에 적었어요. 한국어 낱말(브랜치·테스트)과 영어 문장을 고쳤어요. REQ 줄은 다른 SPEC 처럼 GEARS 영어 문장과 한국어 실패 설명으로 두었어요.
- 하지 않은 것: 전체 `cargo test`(리더 몫), 넓은 `check affected`(READY 에서 한 번), Windows·Linux.
