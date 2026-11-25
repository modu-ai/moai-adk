---
id: SPEC-20261125-update-next-say-ask-chain
title: "The project update plan says per step what comes next in one easy sentence and asks one question for every runnable step, with the apply order and the places to preview again"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9245): \"승인, D-4~D-9 기본값 (Recommended)\"; 리더가 묻기 직전과 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:9cfdb6a8ee7b11fefd2b81f004c60313c1240a953397d581a7aa409d5f24234e"
---

# The project update plan says per step what comes next in one easy sentence and asks one question for every runnable step, with the apply order and the places to preview again

## Intent

`/moai update` 경로의 세션은 `project-update-plan` 이 내놓은 단계 목록을 필드 순서대로 읽고 단계마다 따로 물어요(최악 약 24번). 다음에 할 일(적용·묻기·기다리기·알리기)은 `state`, `apply`, `preview.choices`, `reason` 문장을 맞춰 읽어야 알 수 있고, 다시 미리보기할 자리는 `re-run after …` 이유 문장에만 있어요. `reason` 은 이슈 번호와 내부 명령 이름을 담고 있어서 사용자에게 그대로 읽어 줄 수 없어요. 적용할 것이 없어도 묻는 규칙이 글에 남아 있고, 선택 단계를 거절하는 길(`decline`)은 단계 안쪽 `preview` 에만 있어요. 사용자는 2026-11-25 에 (가) 실행할 수 있는 모든 단계와 뒤따를 단계를 한 화면에 보이고 한 번 묻기, (나) 적용할 것이 없으면 묻지 않고 끝내기를 골랐어요(리더의 기록, 사용자 선택 그대로). 이 SPEC 은 그 결정에 필요한 사실만 `project-update-plan` 출력에 더해요: 단계마다 `next`(apply·ask·wait·report·none), 쉬운 한 문장 `say`, 선택 단계의 `decline`, 맨 위에 실행할 단계와 뒤따를 단계를 담은 물음 하나 `ask`, 적용 순서와 다시 미리보기할 자리 `chain`, 물을 것이 없을 때 끝을 알리는 `finish`. 파일을 쓰는 동작, 기존 필드, 확인 토큰, 계획 digest 는 바꾸지 않아요. 지침 글(`update.md`)은 이 출력을 읽는 다음 조각이에요(update 경로 ⑦).

## Interview

- Q1: 한 번의 물음은 무엇을 덮나요? → A: 사용자가 골랐어요(리더 기록 2026-11-25): 실행할 수 있는 모든 단계와 뒤따를 단계를 한 화면에 보이고 한 번 물어요. 예전 규칙 "보지 않은 단계는 앞의 yes 가 덮지 않는다"를 대신해요.
- Q2: 적용할 것이 없을 때는 물을까요? → A: 사용자가 골랐어요: 묻지 않고 끝내요.
- Q3: 남는 공간 보기 단계(옛 OPTIMIZE)는? → A: 사용자가 골랐어요: update 지침에서 빼요. 이 SPEC 은 글을 바꾸지 않아서 다음 조각의 몫이에요.
- Q4: 백업 위치는? → A: 사용자가 골랐어요: 백업만 사용자 홈의 MoAI 상태 폴더로 돌려요. 이 SPEC 의 일이 아니에요(update 경로 ⑧).
- Q5: 보호 대상 단계(설정·훅·상태 이동·보고서 이동)도 한 번의 yes 에 실을까요? → A: 아직 묻지 않았어요. 기본: 싣되 `guarded` 표시로 물음이 이름을 밝히게 해요(D-5).
- Q6: 고를 선택지가 있는 단계는 yes 에 실릴까요? → A: 아직 묻지 않았어요. 기본: 추천 표시가 있는 선택지는 실리고, 없는 선택지는 `picks` 로 따로 물어요(D-6).
- Q7: 명령이 있는 `needs_decision` 단계는 yes 에 실릴까요? → A: 아직 묻지 않았어요. 기본: 실려요(`next` 가 `apply`, D-7).
- Q8: 새 문장은 어떤 언어인가요? → A: 아직 묻지 않았어요. 기본: 영어로 내고 사용자 언어로 옮기는 일은 세션이 해요(D-8).
- Q9: 적용할 것도 물을 것도 없고 막힌 단계만 남으면? → A: 아직 묻지 않았어요. 기본: 묻지 않고 끝내며 `finish.open` 으로 알려요(D-9).

## Non-goals

- `workflows/update.md` 와 스킬 글의 문장(update 경로 ⑦), 남는 공간 단계를 지침에서 빼는 일, 모델·되돌리기 글을 참조로 옮기는 일.
- 백업 위치와 HOME 표식(update 경로 ⑧).
- 거절 기록과 `project-update-decline` 명령: `SPEC-20261125-update-declined-steps` 의 몫이에요. 이 SPEC 은 그 단계의 `preview.decline` 을 복사해서 보여 줄 뿐 다시 정의하지 않아요.
- 단계의 `state`, `reason`, `preview`, `plan_digest`, `apply` 값과 단계 순서, 확인 토큰, 무엇을 적용하는지. `all_unchanged` 를 계산하는 식도 그대로예요.
- 셸의 `moai update`(사용자 계층) 화면, 새 명령, 도움말 문구.

## Requirements

- REQ-1: The project update plan shall carry the keys `next`, `say` and `decline` in every step, next to the keys a step carried before this SPEC.
- REQ-2: The project update plan shall keep every field it carried before this SPEC with the value it had.
- REQ-3: The project update plan shall leave the project folder and the HOME folder byte-identical.
- REQ-4: The `next` of a step shall be exactly one of `apply`, `ask`, `wait`, `report` and `none`.
- REQ-5: When a step reads `unchanged`, `skip` or `optional`, its `next` shall be `none`. · failure: none — the state is read from the step as it is printed.
- REQ-6: When a step reads `blocked` and waits behind other steps, its `next` shall be `wait`. · failure: none — a step that waits prints no command and no choices.
- REQ-7: When a step reads `ready`, `needs_decision` or `conflict` and prints choices or a command that carries `<warning-confirmation>` or `--confirm-replace`, its `next` shall be `ask`. · failure: none — the choices and the command are read from the step as it is printed.
- REQ-8: When a step prints a command that carries neither of those words and prints no choices, its `next` shall be `apply`. · failure: none — the command is read from the step as it is printed.
- REQ-9: When no rule above names the `next` of a step that reads `ready`, `needs_decision`, `conflict` or `blocked`, its `next` shall be `report`. · failure: none — the step prints no command that could run.
- REQ-10: Every sentence the plan adds for a person shall be one line of plain words that names no step id, state word, issue or card number, SPEC id, digest or command; the sentences are the `say` of a step, the `question` and the option `say` of `ask`, and the `say` of `finish`.
- REQ-11: When the preview of a step carries a `decline` object, the `decline` of the step shall equal that object. · failure: none — the object is the one the declined-steps update prints, copied as it is.
- REQ-12: When the preview of a step carries no `decline` object, the `decline` of the step shall be null. · failure: none — a step that cannot be declined has nothing to copy.
- REQ-13: When a step has `next` `apply` or `ask`, the plan shall carry `ask` as one object with a `question`, the lists `steps`, `follows` and `picks`, and the `options`. · failure: none — a plan without such a step carries `ask` as null (REQ-21).
- REQ-14: The `options` of `ask` shall be `apply`, `choose` and `leave` in that order with `apply` marked recommended, or `choose` and `leave` with `choose` marked recommended while `steps` is empty.
- REQ-15: The `steps` of `ask` shall list in printed order every step with `next` `apply` and every step with `next` `ask` that has a recommended choice and needs no confirmation word, each with its `say`, its `guarded` flag and the id of its recommended choice or null.
- REQ-16: The `follows` of `ask` shall list every step with `next` `wait`, each with its `say`, its `guarded` flag and the ids of the runnable steps it waits behind.
- REQ-17: The `picks` of `ask` shall list every step with `next` `ask` that is not in `steps`, each with its `say` and its `guarded` flag.
- REQ-18: The `guarded` flag of an entry shall be true for the steps `go-settings`, `go-state`, `reports`, `claude-settings`, `hooks` and `claude-settings-merge` and false for every other step.
- REQ-19: The `apply_order` of `chain` shall list the ids of the `steps` of `ask` in the same order.
- REQ-20: The `preview_again` of `chain` shall hold one entry for every runnable step that another step waits behind, with that step as `after` and the steps waiting behind it as `then`.
- REQ-21: While no step has `next` `apply` or `ask`, the plan shall carry `ask` as null and `finish` with a `say` sentence that tells the update ends without a question and an `open` list of the steps with `next` `report` or `wait`.
- REQ-22: While a step has `next` `apply` or `ask`, the plan shall carry `finish` as null.
- REQ-23: While `all_unchanged` is true, the `say` of `finish` shall tell that nothing is left to apply.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): 임시 프로젝트 다섯 곳(빈 프로젝트, 프로필이 있는 프로젝트, 백업 단계가 열린 Go 프로젝트, `AGENTS.local.md` 가 설치를 기다리는 프로젝트, 막힌 단계 하나)의 계획에서 모든 단계가 `next`·`say`·`decline` 을 갖고 `next` 가 다섯 낱말 중 하나이고, 예전 키(`step`·`state`·`reason`·`preview`·`plan_digest`·`apply`)와 최상위 키가 그대로 있으며 계획 전후 트리가 같아요 · verify: `cargo test --locked --offline --test project_update_plan_cli next_say_ac1_`
- AC-2 (REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): `moai_cli::project_update::guide::next_of` 를 상태·명령 유무·확인 낱말·선택지·기다림의 모든 조합으로 불러 표가 맞고, 이름 붙인 단계가 맞아요: 최신·건너뜀 → `none`, 준비된 generation → `apply`, 백업을 기다리는 설치 단계 → `wait`, 프로필 없음의 설치 단계 → `report`, 선택지가 있는 설정 합치기 충돌 → `ask`, 명령이 없는 go-state 보류 → `report`, 거절한 단계와 `AGENTS.local.md` 가 선택을 기다리는 건너뜀 → `none` · verify: `cargo test --locked --offline --test project_update_plan_cli next_say_ac2_`
- AC-3 (REQ-10): AC-1 의 모든 계획에서 모든 `say`, `ask.question`, 옵션 문장, `finish.say` 가 한 줄이고 200자 이하이고 마침표로 끝나고 단계 id·상태 낱말·`#숫자`·`SPEC-`·`digest`·`argv`·`<`·역따옴표를 담지 않아요 · verify: `cargo test --locked --offline --test project_update_plan_cli next_say_ac3_`
- AC-4 (REQ-11, REQ-12): 합치기를 거절하기 전의 `decline` 은 그 단계 `preview.decline` 과 같고, 거절한 뒤의 단계는 `next` 가 `none` 이고 `decline.ask_again_argv` 가 있고, 거절할 수 없는 단계의 `decline` 은 null 이에요 · verify: `cargo test --locked --offline --test project_update_plan_cli next_say_ac4_`
- AC-5 (REQ-13, REQ-14, REQ-15, REQ-16, REQ-17, REQ-18): 설치가 끝난 프로젝트에 `AGENTS.local.md` 만 더하면 `ask.steps` 에 agents-picks 가 `pick` 으로 `move` 와 함께 들어가요. generation 이 준비된 프로젝트에서는 generation 이 `steps` 에, 기다리는 합치기 단계들이 `follows` 에 `after` 로 `generation` 과 함께 들어가요. 설정 합치기 충돌은 `picks` 에 들어가고 `steps` 가 비면 옵션이 `choose`·`leave` 예요. `guarded` 는 REQ-18 의 여섯 단계에서만 참이에요 · verify: `cargo test --locked --offline --test project_update_plan_cli next_say_ac5_`
- AC-6 (REQ-6, REQ-19, REQ-20): `apply_order` 가 `ask.steps` 의 id 와 같고, 기다림 이유 여섯 가지(백업·go-settings·reports·프로필·settings·install)를 만드는 기존 시험의 프로젝트마다 기다리는 단계가 그 이유가 말하는 실행 단계의 `then` 에 있어요. 명령이 없는 go-state 뒤에서 기다리는 단계는 `follows` 에 있고 `after` 가 비어 있어요 · verify: `cargo test --locked --offline --test project_update_plan_cli next_say_ac6_`
- AC-7 (REQ-21, REQ-22, REQ-23): 모두 최신인 프로젝트는 `ask` 가 null 이고 `finish.say` 가 남은 것이 없다고 말하고 `finish.open` 이 비고 모든 단계가 `none` 이에요. 빈 프로젝트는 `ask` 가 null 이고 `finish.open` 에 설치 단계 다섯이 있고 `all_unchanged` 는 false 예요. 실행할 단계가 있는 계획은 `finish` 가 null 이에요 · verify: `cargo test --locked --offline --test project_update_plan_cli next_say_ac7_`
- AC-8 (REQ-19, REQ-20, REQ-21): `chain` 을 따라 `apply_order` 를 찍힌 명령 그대로 차례로 실행하고(추천 선택지는 그 선택지의 명령) 다시 미리보고 `ask` 가 null 이 될 때까지 되풀이하면, 설치를 기다리는 프로젝트와 설치가 끝나고 `AGENTS.md` 를 가져오는 프로젝트 둘 다 끝에서 `finish` 가 있고 `all_unchanged` 가 true 예요. 다시 미리본 계획에서 `then` 의 단계는 더 이상 `wait` 가 아니에요 · verify: `cargo test --locked --offline --test project_update_plan_cli next_say_ac8_`
- AC-9 (REQ-2): 이웃 시험이 그대로 통과해요 · verify: `cargo test --locked --offline --test project_update_plan_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-update-next-say-ask-chain/spec.md`
- write: `src/project_update.rs`, `src/project_update/guide.rs`
- write: `tests/project_update_plan_cli.rs`
- write: `docs/design-review/appendix/operations/project_update_plan.md`
- read-only: `src/project_update/declined.rs`, `src/agents_local_merge.rs`, `src/claude_settings_merge.rs`, `src/project_hosts.rs`, `src/ops/project_update_plan.rs`
- forbidden: `skills/moai/workflows/update.md` 과 다른 스킬 글, `src/ops/project_update_plan.rs`(도움말 골든이 바뀌어요), 실제 HOME, 설치된 moai, 다른 프로젝트, `src/update*`, `src/home*`, `src/hook*`, `.claude/`, 파일을 쓰는 새 코드

## Plan

1. 구현은 `SPEC-20261125-update-declined-steps` 가 착지한 NeXT(`59ac3344cd` 이후) 위에서 해요. 이 SPEC 의 결정이 답을 받은 뒤 카드 범위를 위 write 경로로 넓히고 시작해요.
2. 시험 먼저(RED): `tests/project_update_plan_cli.rs` 에 AC-1~AC-8 의 시험을 `next_say_acN_` 이름으로 쓰고 실패 이유를 기록해요. 새 시험의 상태 폴더 경로는 `state_rel` 로 써요(리터럴을 늘리면 `allow.tsv` 의 72 를 넘어요).
3. `src/project_update.rs`:
   - `Step` 에 `behind: &'static [&'static str]`(기다리는 단계 id)을 더해요. `Step::waiting` 이 `AFTER_*` 상수 바로 옆의 표 하나로 채우고, 직접 `Step {…}` 로 만드는 기다림(설정 합치기의 install 대기)만 손으로 채워요.
   - `pub mod guide;` 한 줄을 더하고 `Step::to_json` 이 `next`·`say`·`decline` 을 더해요. `plan()` 이 `ask`·`chain`·`finish` 를 더해요.
   - `all_unchanged` 식과 `Step::waiting("generation"`, `or_blocked("generation"` 꼴, `name: "generation"` 리터럴 두 개는 그대로 둬요(`project_harness_label` 이 읽어요). `AFTER_*` 상수는 옮기거나 이름을 바꾸지 않아요(`project_preview_order` 가 읽어요).
4. `src/project_update/guide.rs`(파일을 읽지 않는 순수 함수): `pub fn next_of(state, 명령 argv, 선택지 유무, 기다림 유무) -> &'static str`, 단계 이름×{하는 일, 현재 상태, 막힘}과 거절 문장의 `say` 표, `overview` 가 `ask`·`chain`·`finish` 를 만들어요. 문장에 카드·이슈 번호와 개발 도구 이름을 쓰지 않아요(`cli_text_citations` 가 이 파일을 훑어요).
5. 문서 조각 `project_update_plan.md` 에 새 키를 한 단락 더해요(문서 단계).
6. `cargo fmt --all`, AC-9 의 이웃 시험(`project_update_hooks_cli`, `project_update_drops_cli`, `agents_choices_step_cli`, `claude_md_picks_cli`, `generation_request_record`, `project_harness_label`, `project_preview_order`, `cli_text_citations`, `path_root_direct_sites`), `moai check affected --card <base>..<head>` 한 번. READY 에 `audit-tier: 1` 과 그 까닭(경로가 tier 2 접두사에 걸리지 않고 실행 때 아무것도 쓰지 않음)을 적어요.

예시(설명용, 한 프로젝트의 일부): generation 만 실행할 수 있고 합치기 둘이 기다리는 경우.

    "ask": {"question": "Apply these changes to your project now?",
      "options": [{"id": "apply", "recommended": true, "say": "Apply every change listed, then check again for the steps that come after."},
                  {"id": "choose", "say": "Go through the steps one by one."},
                  {"id": "leave", "say": "Leave the project as it is."}],
      "steps": [{"step": "generation", "say": "…", "guarded": false, "pick": null}],
      "follows": [{"step": "agents-picks", "say": "…", "guarded": false, "after": ["generation"]},
                  {"step": "agents-local-merge", "say": "…", "guarded": false, "after": ["generation"]}],
      "picks": []},
    "chain": {"apply_order": ["generation"],
      "preview_again": [{"after": "generation", "then": ["agents-picks", "agents-local-merge"]}]},
    "finish": null

## Risks

- 이 SPEC 은 `behind` 표 하나에 기대요. 새 `AFTER_*` 상수를 더하면서 표에 줄을 안 더하면 그 단계는 `wait` 가 아니라 `report` 로 읽혀요. AC-6 이 지금 여섯 가지를 고정해요.
- 한 번의 yes 에 보호 대상 단계가 실려요(D-5). 물음이 `guarded` 단계를 이름으로 밝히고 `choose` 가 있어요. 이전에는 단계마다 따로 yes 를 받았어요.
- 뒤따르는 단계의 내용은 다시 미리보기 전에는 몰라요. 이 SPEC 은 사실만 내요. 앞의 yes 가 그 단계를 덮는 조건은 지침 글의 몫이에요(D-4).
- `update.md` 는 다음 조각이 바뀔 때까지 옛 규칙("항상 묻기", 단계마다 yes)을 말해요. 새 키는 옛 글을 깨지 않아요. 글은 키를 모를 뿐이에요.
- `say` 는 영어 한 표예요. 시험은 문장이 아니라 규칙(한 줄, 길이, 금지 낱말)을 고정해요.
- 카드·이슈 번호 검사(`cli_text_citations`)가 `guide.rs` 를 훑어요. `src/project_update.rs` 는 PENDING 목록에 있어서 면제예요.
- 이 SPEC 의 (가)~(나) 결정은 리더 기록으로 전해졌어요. 승인 때 사용자가 다시 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `next` 는 상태·명령·선택지·기다림에서 정해져요: 최신·건너뜀·선택 → `none`, 기다림 → `wait`, 선택지나 확인 낱말 → `ask`, 명령 → `apply`, 나머지 → `report`. | 리더의 목표 모양(2026-11-25). | `Step::to_json` 의 세 키를 지워요. |
| D-2 | 물음은 하나(`ask`)이고 실행 단계(`steps`), 뒤따르는 단계(`follows`), 따로 고를 단계(`picks`)를 담아요. 실행할 것이 없으면 `ask` 는 null 이고 `finish` 가 끝을 알려요. | 사용자 결정 (가)(나), 2026-11-25. | `plan()` 의 세 키를 지워요. |
| D-3 | 단계의 `decline` 은 `preview.decline` 의 복사본이에요. 대안: 명령 낱말만. | `SPEC-20261125-update-declined-steps` REQ-7. | 키를 지워요. |
| D-4 | 뒤따르는 단계는 다시 미리본 계획이 `apply` 로 읽을 때만 앞의 yes 로 가요. 다른 값이면 그 계획이 말하는 대로 묻거나 알려요. (이 조각은 사실만 내요. 문장은 다음 조각.) | 제안, 승인 전. | 해당 없음. |
| D-5 | 보호 표시 단계: `go-settings`, `go-state`, `reports`, `claude-settings`, `hooks`, `claude-settings-merge`. 대안: 모두 아님. | 이전 설계의 보호 항목(프로젝트 상태 이동 포함)과 권한·HOME 규칙. 제안, 승인 전. | 표를 비워요. |
| D-6 | 추천 표시가 있는 선택지(agents-picks `move`, generation `append-managed`)는 yes 에 실려요. 추천이 없는 것(`keep-mine`, `skip`, `skip-conflicts`)은 `picks`. | 제안, 승인 전. | 추천 선택지를 `picks` 로 보내요. |
| D-7 | 명령이 있는 `needs_decision`(go-state 이동 등)은 `apply` 예요. 대안: `ask`. | 사용자 결정 (가): 실행할 수 있는 모든 단계를 한 번에. 제안, 승인 전. | `next_of` 의 조건 하나를 바꿔요. |
| D-8 | 새 문장은 영어로 내요. 옮기는 일은 세션이 해요(`reason` 과 같아요). | 제안, 승인 전. | 해당 없음. |
| D-9 | 막힌 단계만 남아도 `ask` 는 null 이고 `finish.open` 이 알려요. | 사용자 결정 (나)를 넓힌 해석. 제안, 승인 전. | `finish` 의 조건을 `all_unchanged` 로 좁혀요. |

## Result

- AC-1: met — `cargo test --locked --offline --test project_update_plan_cli next_say_ac1_` passed (exit 0, 13 passed with the whole `next_say_` filter, log target/impl-green/next_say.log).
- AC-2: met — `... next_say_ac2_` passed in the same run, 74 state/command/confirmation/choices/waiting combinations plus the named steps.
- AC-3: met — `... next_say_ac3_` passed, and the unit test `project_update::guide` (1 passed, target/impl-green/unit.log) checks every sentence of the table.
- AC-4: met — `... next_say_ac4_` passed.
- AC-5: met — `... next_say_ac5_` passed, three tests.
- AC-6: met, with one reading — `... next_say_ac6_` passed. The sentence "after is empty behind go-state" cannot hold literally: while a runnable step exists, `after` names the runnable step the waiting step stands behind (for the backup it is `reports`), and it never names a step that prints no command. The tests assert that reading, which matches REQ-16.
- AC-7: met — `... next_say_ac7_` passed.
- AC-8: met — `... next_say_ac8_` passed; the chain converges in two rounds to `all_unchanged` true.
- AC-9: met — the whole `project_update_plan_cli` target passed (73 passed, target/impl-green/whole2.log) and the nine neighbour targets of the SPEC plan passed (target/impl-green/neighbours.out). `cargo clippy --locked --offline --all-targets -- -D warnings` exited 0 (target/impl-green/clippy2.log).

Deviation from the Plan text: `Step` carries no new `behind` field. The test helper of the declined-step code builds `Step { .. }` literally outside this SPEC's scope, so the table `BEHIND` next to the `AFTER_*` constants and the method `Step::behind` give the same facts without touching that file. When `ask` is null, `chain` prints empty lists instead of null.
