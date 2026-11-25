---
id: SPEC-20261124-desktop-install-agents
title: "moai update puts the flow rule directly under rules/ and the preview names each Claude folder it writes and says when Desktop would not read it"
status: approved
revision: 1
created: 2026-11-24
parent: SPEC-20261122-claude-md-flow-rule
branch: spec/desktop-install-agents-r2-20261124
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
- `tests/fixtures/home-style/render-5.md` still names the earlier path in prose; not changed here (out of write scope), reported in the READY gaps. The one line of `skills/moai/references/source-manifest.json` that named it was corrected in the repair after the first audit (under a line lease; see the Result).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The flow rule file is `rules/moai-flow.md` | 리더 결정 노트 7929 (rules 를 `rules/` 바로 아래 `moai-*.md` 로), 조사 노트 7928 (사용자 rules 재귀 검증 안 됨); 레인이 이름을 `moai-flow.md` 로 골랐고 승인은 리더 몫 | `FLOW_FILE` 을 되돌리고 새 kind 의 영수증은 새 경로로 되돌려요 |
| D-2 | The earlier `rules/moai/flow.md` stays; the preview names it | 리더 결정 노트 7926·7929 의 "지우기는 사용자 선택" 을 rules 에 같은 뜻으로 읽은 레인 해석; 대안: `--remove-old-flow` 선택지를 `remove_old_block` 처럼 더한다(기본: 안 더함) | 해당 없음(쓰지 않아요) |
| D-3 | The `~/.claude` target is offered as the command `env -u CLAUDE_CONFIG_DIR moai update`, no new option | 레인 제안, 리더 승인 대기; 대안: `--claude-folder default` 같은 옵션(모든 층이 폴더를 받아야 해서 더 큰 변경, 기본: 안 함) | 안내 문장만 지우면 돼요 |
| D-4 | The notice lives in `UserLayerPlan.notices`, outside every digest | 기존 `notices` 의 뜻(읽기 전용, 다이제스트 밖, `src/update.rs:1722`) | 필드 하나를 지우면 돼요 |
| D-5 | Agents are not touched in this slice | 리더 결정 노트 7929 | 해당 없음 |

## Result

첫 조각(rules 자리 바꾸기와 되돌리기 kind, REQ-1 에서 REQ-4)을 구현했어요. 판 1 의 승인 근거는 카드 원장 노트 7937 의 `LEADER 결정 (2026-11-23T21:55:56Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 되돌릴 수 있음, 분류 none)` 이고, 이 레인은 그 노트를 머리말에 인용만 했어요. 알림 조각(REQ-5 에서 REQ-9, AC-4, AC-5)은 `src/update.rs` 를 다른 카드가 쥐고 있어 이 조각에 넣지 않았어요(노트 7937). 에이전트 자리는 건드리지 않았어요(노트 7929).

### 실행 기록 (2026-11-24, 이 워크트리, 임시 HOME 만)

- RED: `moai heavy-slot-acquire --name desktop-install-agents --max 5 -- cargo test --locked --offline --test desktop_folder_rules_cli` (`RUSTC_WRAPPER=sccache`): exit 101, 3 failed (계획이 옛 경로 `rules/moai/flow.md` 를 말함, 옛 파일이 있으면 새 파일 계획이 `conflict`, 새 파일이 안 만들어짐).
- GREEN: 같은 명령 exit 0, 3 passed.
- 옛 경로를 못박은 시험 갱신 뒤: `claude_md_home_cli` 48, `home_routing_cli` 11, `host_home_cli` 12, `revert_twice_cli` 2, `e2e_temp_home_cli` 10 모두 passed. `update_user_layer_cli` 는 두 줄(296, 1293)을 바꾸기 전에 20 passed, 2 failed 였고 범위를 넓힌 뒤(노트 7955 의 리더 답) 그 두 줄을 새 경로로 맞췄어요.
- `check affected` 본 실행의 영수증은 READY 노트에 있어요.

### 뮤턴트 (이 워크트리에서 손으로 한 줄 변이, `cargo test --test desktop_folder_rules_cli`, 커밋 뒤 `git checkout` 으로 복구)

- `revert_follows_the_kind`: `src/home_routing/rules.rs` 의 `"flow" => LEGACY_FLOW_FILE` 를 `FLOW_FILE` 로: exit 101 (죽음).
- `revert_follows_the_kind`: 새 쓰기의 kind `"moai_flow"` 를 `"flow"` 로: exit 101 (죽음).
- `earlier_flow_file_is_left_in_place`: 옛 파일 존재 검사 `is_ok()` 를 `is_err()` 로: exit 101 (죽음, 미리보기가 옛 파일 이름을 말하지 않음).
- `flow_rule_is_written_directly_under_rules`: `FLOW_FILE` 을 옛 경로로: exit 101 (죽음, 세 시험 모두).

### 인수 기준

- AC-1: met — `flow_rule_is_written_directly_under_rules` 통과: 계획이 `rules/moai-flow.md` 를 말하고, 적용이 패키지 글 그대로 쓰고, `rules/moai/` 폴더가 생기지 않아요.
- AC-2: met — `earlier_flow_file_is_left_in_place` 통과: 옛 `rules/moai/flow.md` 가 있으면 계획이 그 이름과 `left in place` 를 말하고, 적용 뒤 옛 파일의 바이트와 수정 시각이 그대로이고, 두 번째 계획은 `unchanged` 예요.
- AC-3: met — `revert_follows_the_kind` 통과: 새 적용의 영수증은 `rules/moai-flow.md` 를 되돌리고, 옛 배치가 쓰던 kind `flow` 항목은 `rules/moai/flow.md` 만 되돌리며 새 경로를 건드리지 않아요.
- AC-4: not_run — 알림 조각(REQ-5, REQ-6, REQ-7, REQ-9)은 이 조각에 없어요.
- AC-5: not_run — 알림 조각(REQ-8)은 이 조각에 없어요.
- AC-6: 여섯 파일 중 `update_user_layer_cli` 를 포함해 모두 새 경로를 이름 붙였어요. `check affected` 본 실행의 결과는 READY 노트가 가지고 있어요(같은 head).

### 첫 감사 뒤 수리 (2026-11-24, 같은 가지, 임시 HOME 만)

감사(head `bf14074b7b`)가 FAIL 로 낸 B1·N2·N3 을 고쳤어요. 알림 조각(REQ-5 에서 REQ-9)은 이 수리에 없어요.

- B1: 예전 `rules/moai/flow.md` 안내를 못박는 시험 둘을 `tests/desktop_folder_rules_cli.rs` 에 더했어요. `earlier_flow_note_is_added_after_an_existing_detail` 은 사용자 `rules/moai.md` 때문에 흐름 줄이 `held`(설명이 이미 있음)일 때 안내가 그 설명 뒤에 붙는지, `no_earlier_flow_file_means_no_note` 는 예전 파일이 없으면 두 경우(create, held) 모두 안내가 붙지 않는지 봐요. 같은 파일 5 passed(`moai heavy-slot-acquire … cargo nextest run --locked --test desktop_folder_rules_cli`, exit 0).
- 뮤턴트(`src/home_routing/rules.rs`, 커밋 뒤 손으로 한 줄 변이, `git checkout` 으로 복구): `Some(d) => format!("{d}; {note}")` 를 `Some(d) => d` 로 바꾸면 `earlier_flow_note_is_added_after_an_existing_detail` 이 실패(exit 100). 예전 파일 존재 검사 `is_ok()` 를 `is_ok() || true` 로 바꾸면 `no_earlier_flow_file_means_no_note` 가 실패(exit 100).
- N2: `## Scope` 의 `write:` 줄은 판 1 의 승인 내용이에요. 범위(scope)는 승인 digest 에 들어가서(`src/spec.rs:563` `content_digest`) 줄을 바꾸면 `spec check` 가 `approval_digest` 로 막히고 판 2 와 새 승인이 필요해요. 이 레인은 승인을 기록하지 않아서 줄은 그대로 두고, 판 1 승인 뒤 카드 예약 범위가 넓어진 경로를 여기 적어요(`write:` 줄에 없는 것): `tests/areas.toml`·`tools/repo_scans.json`(노트 7979), `docs/design-review/appendix/operations/home_routing_plan.md`·`home_routing_apply.md`(노트 8005), `skills/moai/references/source-manifest.json` 한 줄(노트 8005, `moai card line-lease` 안에서만; Scope 의 `read-only:` 줄과 다른 쓰기). 이 경로들은 리더 결정으로 범위에 들었고 `tests/update_user_layer_cli.rs` 는 판 1 `write:` 줄에 이미 있어요. 줄을 판 2 로 맞출지는 리더 몫이에요(승인 필요).
- N3: `docs/design-review/appendix/operations/home_routing_plan.md` 와 `home_routing_apply.md` 의 흐름 파일 경로를 `rules/moai-flow.md` 로, 영수증 kind 목록을 `routing`·`moai_flow`·`old_block`(옛 `flow` 는 `rules/moai/flow.md` 만)으로, 계획 문서의 충돌 목록에서 더는 흐름 길이 아닌 `rules/moai` 를 빼고 예전 파일 안내를 한 줄 더했어요. `skills/moai/references/source-manifest.json` 의 한 줄(`rules/moai/flow.md` → `rules/moai-flow.md`)은 `moai card line-lease` 안에서 `sed` 로 바꿨어요.
- 인수 기준은 위와 같아요(AC-1, AC-2, AC-3 met · AC-4, AC-5 not_run · AC-6 는 수리 head 의 `check affected` 영수증이 근거).

### 남은 것

- 알림 조각(REQ-5 에서 REQ-9)은 `src/update.rs` 를 쥔 카드가 놓은 뒤 SCOPE-REQUEST 로 이어요.
- `tests/fixtures/home-style/render-5.md` 의 설명 글이 옛 경로를 말로 적고 있어요(범위 밖, 이 조각에서 바꾸지 않았어요).
- 후속 카드 후보(감사 N1, N4): 예전 영수증을 새 apply 뒤에 되돌리면 공용 `flow_sha256` 칸이 비어 다음 새 판이 `conflict` 로 계획돼요(데이터 손실 없음, 순서대로 되돌리면 맞아요). 되돌리기가 `~/.moai` 의 상태 파일과 영수증 폴더를 남기는 것은 기존 동작이에요.
- 카드 기준 3(임시 HOME 설치 뒤 진짜 claude 세션에서 /moai, 에이전트 다섯, rules 확인)은 운영자 단계라 not_run 이에요.
