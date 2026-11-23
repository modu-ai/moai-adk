---
id: SPEC-20261124-desktop-install-agents
title: "moai update puts the flow rule directly under rules/ and the preview names each Claude folder it writes and says when Desktop would not read it"
status: approved
revision: 1
created: 2026-11-24
parent: SPEC-20261122-claude-md-flow-rule
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7937 (2026-11-23T21:55:56Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 7934 에 답함): 판 1 승인(커밋 87c2c637375f), D-1 rules/moai-flow.md, D-2 옛 flow.md 는 그대로, D-3 새 옵션 없이 env -u 안내; 알림 부분(REQ-5 에서 REQ-9)은 src/update.rs 를 쥔 C-318 이 놓은 뒤 SCOPE-REQUEST; 레인은 그 노트를 인용만 해요"
  digest: "sha256:a19038c1f5c7de7fcca492753ce845cb37e2dc554c18a19e3ab12a3a2344782c"
---

# moai update puts the flow rule directly under rules/ and the preview names each Claude folder it writes and says when Desktop would not read it

## Intent

카드 C-435 의 첫 조각이에요. 리더 결정(카드 노트 7929, 2026-11-23T21:49:04Z, 사용자 위임 2026-11-23T13:57Z)이 범위를 정했어요: 에이전트 자리는 건드리지 않고(실제 claude 로 확인한 결과를 받은 뒤 두 번째 조각), 아래 둘만 해요.

1. 흐름 규칙을 `rules/moai/flow.md`(하위 폴더)에서 `rules/moai-flow.md`(`rules/` 바로 아래)로 옮겨요. 공식 문서는 프로젝트 rules 만 재귀로 읽는다고 적고 사용자 rules(`~/.claude/rules/`)의 재귀는 적지 않아서(노트 7928: 검증 안 됨) 하위 폴더 파일은 세션에 안 실릴 수 있어요. 라우팅 파일은 이미 `rules/moai.md` 로 바로 아래예요.
2. `moai update` 미리보기가 Claude 폴더에 쓰는 층마다 그 폴더를 이름으로 말하고, `CLAUDE_CONFIG_DIR` 이 `~/.claude` 와 다르면 "데스크탑은 로그인 셸이 export 하고 데스크탑을 다시 켠 경우에만 그 폴더를 읽고, 아니면 `~/.claude` 를 읽어요" 를 알리고 `~/.claude` 로 쓰는 길을 알려요. 이 미리보기는 아무것도 쓰지 않아요.

### What exists (tip `b77d1efe75`, 읽은 시각 2026-11-24)

- 라우팅 층은 `~/.claude` 에만 써요(`src/home_routing.rs:315-336`). `CLAUDE_CONFIG_DIR` 이 가리키는 다른 폴더는 파일을 쓰지 않는 줄 하나(`linked` 또는 `not_linked`)만 받아요(`src/home_routing/rules.rs:394-466`). 파일 이름은 `RULES_FILE = "rules/moai.md"`, `FLOW_FILE = "rules/moai/flow.md"`(`src/home_routing.rs:49-50`).
- 되돌리기는 영수증 항목의 `kind` 이름으로 파일을 골라요: `"flow" => FLOW_FILE`(`src/home_routing/rules.rs:739`). 그래서 경로만 바꾸면 빌드 22 이전 영수증의 `flow` 항목이 새 경로를 가리키게 돼요.
- 스킬, 에이전트, 스타일은 `host_home` 이 푼 폴더(`CLAUDE_CONFIG_DIR` 이 이 계정 것으로 셀 때 그 폴더, 아니면 `~/.claude`)에 써요(`src/host_home.rs`, `src/agent_install.rs:36`).
- `UserLayerPlan.notices` 는 읽기 전용이고 모든 다이제스트 밖이에요(`src/update.rs:1722-1726`); JSON 은 `"notices"`(`src/update.rs:2218`)로, 터미널 화면은 `src/update.rs:2111` 이 낱말 `notice` 로 보여요.
- 카드 본문이 말한 "rules 와 훅이 데스크탑이 안 보는 곳에 쓰인다" 는 라우팅 층에는 해당하지 않아요(위 첫 항목). 훅과 설정의 자리는 이 SPEC 이 다루지 않아요.

## Interview

- Q1: 에이전트도 이 조각에서 옮기나요? -> A: 아니요(노트 7929): 공식 문서가 `agents/` 를 재귀로 읽는다고 적어서 R6 의 "안나온다" 원인이 하위 폴더라고 아직 말할 수 없어요. 두 번째 조각이에요.
- Q2: 새 파일 이름은? -> A: `rules/moai-flow.md`(D-1). `moai-` 접두어가 사용자 rules 와 구분해요.
- Q3: 빌드 22 이전이 쓴 `rules/moai/flow.md` 는? -> A: 그대로 두고 미리보기가 이름을 말해요(D-2). 지우기는 사용자 선택이에요.
- Q4: `~/.claude` 를 고르는 방법은? -> A: 이 조각은 새 옵션 없이 미리보기가 정확한 명령(`env -u CLAUDE_CONFIG_DIR moai update`)을 알려요(D-3). 새 옵션이 필요하면 리더가 정해요.

## Non-goals

- 에이전트 자리(`agents/moai/` 와 `agents/` 바로 아래)와 그 옮김 미리보기: 두 번째 조각.
- 훅과 `settings.json` 의 자리, Cowork, 사용자가 만든 `rules/` 링크: 이 SPEC 이 바꾸지 않아요.
- 실제 `~/.claude`, `~/.moai` 에 쓰기: 모든 시험은 임시 HOME 에서만 해요.
- `rules/moai/flow.md` 의 삭제 또는 이동 쓰기: 이 조각의 apply 는 그 파일을 읽지도 쓰지도 않아요.
- `--config-dir` 이 만드는 "읽기만 하는 줄" 의 동작 변경.

## Requirements

- REQ-1: When `moai update` previews the Claude routing layer, the plan shall name the flow rule file `rules/moai-flow.md` directly under the default Claude folder. · failure: none — a folder or file that cannot be read is named as blocked or conflict as before this SPEC.
- REQ-2: When the routing apply writes the flow rule, it shall write `rules/moai-flow.md` and create no `rules/moai/` folder. · failure: none — a write that fails leaves the receipt pending and the earlier files as they were, as before this SPEC.
- REQ-3: IF an earlier build left `rules/moai/flow.md` in the folder, THEN the plan shall name that file in the flow row's detail as left in place. · failure: none — a duplicate of the same text is the outcome until the user deletes the old file.
- REQ-4: When a receipt entry of kind `flow` from an earlier build is reverted, the revert shall use the path `rules/moai/flow.md`. · failure: none — an entry of the new kind reverts `rules/moai-flow.md` only.
- REQ-5: When `moai update` previews, the preview shall carry one notice of kind `claude_folders` that names the Claude folder each of the parts skills, agents, style and rules writes. · failure: none — a part left out of the preview is not named.
- REQ-6: When `CLAUDE_CONFIG_DIR` counts for this account and names a folder other than the default Claude folder, the notice shall say that Desktop reads that folder only after a login shell export and a Desktop restart and otherwise reads the default folder. · failure: none — the notice is text only.
- REQ-7: When the notice warns about another folder, it shall name `env -u CLAUDE_CONFIG_DIR moai update` as the run that targets the default Claude folder. · failure: none — nothing is written without the user's own run.
- REQ-8: While the notice is built, the preview shall write nothing and leave `plan_digest` and every part digest unchanged. · failure: none — the notice is read-only text.
- REQ-9: When `CLAUDE_CONFIG_DIR` is unset or does not count for this account or names the default Claude folder, the notice shall name the one folder and carry no warning. · failure: none — the notice is text only.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary HOME the routing plan names `rules/moai-flow.md`, the apply writes that file with the packaged text, and no `rules/moai/` folder exists afterwards · verify: `cargo test --locked --offline --test desktop_folder_rules_cli flow_rule_is_written_directly_under_rules`
- AC-2 (REQ-3): in a temporary HOME that holds an earlier `rules/moai/flow.md`, the plan names it as left in place, the apply leaves its bytes and mtime as they were, and a second plan is `unchanged` · verify: `cargo test --locked --offline --test desktop_folder_rules_cli earlier_flow_file_is_left_in_place`
- AC-3 (REQ-4): a receipt written by the earlier layout (entry kind `flow`, path `rules/moai/flow.md`) still reverts that path and never touches `rules/moai-flow.md`, and a receipt of the new apply reverts `rules/moai-flow.md` while it holds the bytes written · verify: `cargo test --locked --offline --test desktop_folder_rules_cli revert_follows_the_kind`
- AC-4 (REQ-5, REQ-6, REQ-7, REQ-9): the preview of `moai update` in a temporary HOME names one folder per part with no warning when the variable is unset, and with `CLAUDE_CONFIG_DIR` set to another absolute folder under that account names both folders and carries the warning text and the `env -u` command · verify: `cargo test --locked --offline --test desktop_folder_rules_cli preview_names_each_claude_folder`
- AC-5 (REQ-8): the notice leaves `plan_digest` equal to the one computed without it, and the preview run leaves the temporary HOME byte-identical · verify: `cargo test --locked --offline --test desktop_folder_rules_cli notice_is_read_only`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4): the tests that pinned the earlier path (`tests/claude_md_home_cli.rs`, `tests/home_routing_cli.rs`, `tests/host_home_cli.rs`, `tests/update_user_layer_cli.rs`, `tests/revert_twice_cli.rs`, `tests/e2e_temp_home_cli.rs`) name the new path and pass · verify: `moai check affected <base>..<head>` (bare, once, at READY)

## Scope

- write: `.moai/specs/SPEC-20261124-desktop-install-agents/spec.md`
- write: `src/home_routing.rs`
- write: `src/home_routing/rules.rs`
- write: `src/update.rs`
- write: `src/ops/home_routing_plan.rs`
- write: `tests/desktop_folder_rules_cli.rs`
- write: `tests/claude_md_home_cli.rs`
- write: `tests/home_routing_cli.rs`
- write: `tests/host_home_cli.rs`
- write: `tests/update_user_layer_cli.rs`
- write: `tests/revert_twice_cli.rs`
- write: `tests/e2e_temp_home_cli.rs`
- read-only: `src/host_home.rs`
- read-only: `src/agent_install.rs`
- read-only: `skills/moai/references/source-manifest.json`
- forbidden: `skills/moai/SKILL.md`
- forbidden: `CLAUDE.md`

## Plan

1. Write failing tests in a temporary HOME for AC-1 to AC-5 (the new file `tests/desktop_folder_rules_cli.rs`); run them red.
2. Move `FLOW_FILE` to `rules/moai-flow.md`, give the new write its own receipt kind, keep the earlier kind reverting the earlier path (REQ-1, 2, 4).
3. Add the left-in-place detail for an earlier `rules/moai/flow.md` (REQ-3).
4. Build the `claude_folders` notice from the resolved folders and add it to `UserLayerPlan.notices` outside every digest (REQ-5 to 8).
5. Update the six tests that pinned the earlier path; guard mutants, one `check affected` bare at READY.

## Risks

- A user's Claude may still not read `~/.claude/rules/` files below the top level or at all in a profile whose `rules/` is not a link to `~/.claude/rules`: the existing `not_linked` row says so; this SPEC does not change it. Noticed by the AC-4 text and the operator run of card criterion 3.
- A receipt written by build 22 names kind `flow`: a revert that maps it to the new path would undo the wrong file. Pinned by AC-3.
- `rules/moai/flow.md` and `rules/moai-flow.md` hold the same text until the user deletes the old one; if Claude loads both the rule appears twice. The text is short (at most 80 lines, 5 KiB); noticed by the operator run.
- `tests/fixtures/home-style/render-5.md` and `skills/moai/references/source-manifest.json` still name the earlier path in prose; not changed here (out of write scope), reported in the READY gaps.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The flow rule file is `rules/moai-flow.md` | 리더 결정 노트 7929 (rules 를 `rules/` 바로 아래 `moai-*.md` 로), 조사 노트 7928 (사용자 rules 재귀 검증 안 됨); 레인이 이름을 `moai-flow.md` 로 골랐고 승인은 리더 몫 | `FLOW_FILE` 을 되돌리고 새 kind 의 영수증은 새 경로로 되돌려요 |
| D-2 | The earlier `rules/moai/flow.md` stays; the preview names it | 리더 결정 노트 7926·7929 의 "지우기는 사용자 선택" 을 rules 에 같은 뜻으로 읽은 레인 해석; 대안: `--remove-old-flow` 선택지를 `remove_old_block` 처럼 더한다(기본: 안 더함) | 해당 없음(쓰지 않아요) |
| D-3 | The `~/.claude` target is offered as the command `env -u CLAUDE_CONFIG_DIR moai update`, no new option | 레인 제안, 리더 승인 대기; 대안: `--claude-folder default` 같은 옵션(모든 층이 폴더를 받아야 해서 더 큰 변경, 기본: 안 함) | 안내 문장만 지우면 돼요 |
| D-4 | The notice lives in `UserLayerPlan.notices`, outside every digest | 기존 `notices` 의 뜻(읽기 전용, 다이제스트 밖, `src/update.rs:1722`) | 필드 하나를 지우면 돼요 |
| D-5 | Agents are not touched in this slice | 리더 결정 노트 7929 | 해당 없음 |

## Result

Not yet implemented. Approval is the leader's note, not this lane's.
