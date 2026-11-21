---
id: SPEC-20261121-claude-agents-md-import
title: "CLAUDE.md imports AGENTS.md and AGENTS.md tells every agent to read AGENTS.local.md"
status: draft
revision: 1
created: 2026-11-21
---

# CLAUDE.md imports AGENTS.md and AGENTS.md tells every agent to read AGENTS.local.md

## Intent

MoAI 가 만드는 규칙 파일은 `AGENTS.md`(팀 규칙)와 `AGENTS.local.md`(개인 규칙)예요. 그런데 두 호스트가 이 파일을 읽는 길이 달라서 틈이 있어요. Claude Code 는 `AGENTS.md` 를 `CLAUDE.md` 나 `CLAUDE.local.md` 가 없을 때만 직접 읽고, 개인 파일 `AGENTS.local.md` 는 읽지 않아요(공식 문서, D-3). Codex 는 `AGENTS.md` 를 읽지만 파일을 불러오는 문법이 없어서 `AGENTS.local.md` 를 모르고요. 사용자 지시(2026-11-21): "CLAUDE.md 안에 @AGENTS.md 를 추가해서 사용하도록 하자", 그리고 개인 규칙은 "AGENTS.local.md 파일을 읽으라고 AGENTS.md 지시문 추가". 이 SPEC 은 생성 계획(미리보기 → 선택 → 영수증)에 두 가지를 더해요. (1) 사용자가 고르면 `CLAUDE.md` 에 관리 블록 하나, 내용은 `@AGENTS.md` 한 줄(파일이 없으면 이 블록만 든 파일을 만들어요). (2) `AGENTS.md` 관리 블록에, 개인 파일을 쓰기로 한 프로젝트(`local_file` 이 `ignored` 나 `tracked`)에서는 "세션을 시작할 때 `AGENTS.local.md` 가 있으면 읽고 따르라"는 지시문과 `@AGENTS.local.md` 한 줄을 넣어요. Claude 는 그 줄을 직접 불러오고, Codex 는 지시문대로 도구로 파일을 열어요. 모든 사용자 프로젝트에 같은 말이 통해요(이 저장소 전용 도구 없음).

## Non-goals

- `AGENTS.local.md` 를 만드는 일(`/moai project` 가 개인 규칙 파일 만들기를 선택으로 내놓는 일)과 새 워크트리로 가져가는 일(`.worktreeinclude`): 각각 다른 SPEC 이 맡아요. 이 SPEC 은 그 파일이 있을 때 읽게 하는 말만 더해요.
- `ja`·`zh` 문구: 지금 생성기가 아는 언어는 `en`·`ko` 둘이에요. 나머지 언어 문구는 언어 확장 SPEC 이 맡아요.
- 사용자가 `CLAUDE.md` 나 `AGENTS.md` 에 쓴 줄 고치기: 하지 않아요. 관리 블록 바깥은 한 바이트도 안 바꿔요.
- 선택지를 `/moai project`·`moai update` 의 선택지 목록에 내놓는 부분과 두 워크플로 글: 이 조각에 없어요(Scope 참고, 다른 카드가 쥔 파일).
- 파일이 없는 `@path` 가 호스트에서 어떻게 되는지 확인하기: 공식 문서에 없어서 모른다고 두고, 그래서 개인 파일을 쓰지 않는 프로젝트에는 줄을 넣지 않아요(D-2).

## Requirements

- REQ-1: When a request picks `claude-agents-import` and the profile targets the `claude` host, the generation plan shall carry one managed block in CLAUDE.md whose only line is `@AGENTS.md`, with that line in the row's preview. · failure: none
- REQ-2: When CLAUDE.md does not exist, the plan shall create it holding only that managed block. · failure: none
- REQ-3: When CLAUDE.md exists, the apply shall append the managed block and keep every earlier byte of the file. · failure: none
- REQ-4: IF a line outside the managed block already imports AGENTS.md, THEN the plan shall carry no CLAUDE.md row. · failure: none
- REQ-5: IF the same request is planned again after the apply, THEN the plan shall show the CLAUDE.md block as `unchanged`. · failure: none
- REQ-6: When a request carries no `claude-agents-import` pick or the profile does not target the `claude` host, the plan shall carry no CLAUDE.md row. · failure: none
- REQ-7: When `local_file` is `ignored` or `tracked`, the AGENTS.md managed block shall carry the instruction to read AGENTS.local.md at session start as the user's personal rules, with the team safety rules winning on conflict, in `en` and in `ko`. · failure: none
- REQ-8: When `local_file` is `ignored` or `tracked`, the AGENTS.md managed block shall carry the line `@AGENTS.local.md` on a line of its own outside any code span or fence. · failure: none
- REQ-9: When `local_file` is `none`, the AGENTS.md managed block shall be the text it was before this change. · failure: none
- REQ-10: The managed blocks this change renders shall hold no `@` inside a code span, so the host imports exactly the lines meant to be imported. · failure: none
- REQ-11: When the projected AGENTS.md passes 200 lines, the plan shall warn that it passes the length Claude Code advises. · failure: none
- REQ-12: The request schema and the pick list shall name `claude-agents-import` and refuse any other unknown id as before. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): 선택지가 있고 `CLAUDE.md` 가 없으면 `CLAUDE.md` 행이 `create` 이고 미리보기에 `@AGENTS.md` 가 있고 계획은 아무것도 쓰지 않아요 · verify: `cargo test --locked --offline --test agents_md_import_cli the_pick_plans_a_claude_md_block_with_only_the_import_line`
- AC-2 (REQ-3, REQ-5): 사용자 줄(CRLF 와 마지막 줄바꿈 없음 포함)이 있는 `CLAUDE.md` 에 적용하면 사용자 바이트가 앞에 그대로 있고 블록은 한 번만 있고, 다시 계획하면 `unchanged` 예요 · verify: `cargo test --locked --offline --test agents_md_import_cli applying_appends_after_the_user_text_and_a_second_plan_is_unchanged`
- AC-3 (REQ-4): 블록 바깥 줄이 이미 `@AGENTS.md` 를 불러오면 `CLAUDE.md` 행이 없어요 · verify: `cargo test --locked --offline --test agents_md_import_cli a_claude_md_that_already_imports_agents_md_gets_no_block`
- AC-4 (REQ-6): 선택지가 없거나 `claude` 호스트가 아니면 `CLAUDE.md` 행이 없어요 · verify: `cargo test --locked --offline --test agents_md_import_cli without_the_pick_or_the_claude_host_there_is_no_claude_md_row`
- AC-5 (REQ-7, REQ-8): `ignored`·`tracked` 로 계획하면 `AGENTS.md` 블록에 `en`·`ko` 지시문과 `@AGENTS.local.md` 줄이 있어요 · verify: `cargo test --locked --offline --test agents_md_import_cli agents_md_tells_every_agent_to_read_the_local_file_and_imports_it`
- AC-6 (REQ-9): `none` 으로 계획한 블록은 `ignored` 블록에서 이 변경이 더한 두 줄을 뺀 것과 같아요 · verify: `cargo test --locked --offline --test agents_md_import_cli agents_md_is_unchanged_when_no_local_file_is_chosen`
- AC-7 (REQ-10): 두 파일의 블록을 코드 스팬과 펜스를 걷어 낸 뒤에 읽으면 `@` 로 시작하는 줄이 `@AGENTS.md` 와 `@AGENTS.local.md` 뿐이고, 코드 스팬 안에는 `@` 가 없어요 · verify: `cargo test --locked --offline --test agents_md_import_cli the_rendered_blocks_have_no_at_sign_inside_a_code_span`
- AC-8 (REQ-11): 사용자 글이 길어서 합친 `AGENTS.md` 가 200줄을 넘으면 계획 경고가 있고, 짧으면 없어요(199·200·201 줄 경계) · verify: `cargo test --locked --offline --test agents_md_import_cli the_plan_warns_when_agents_md_passes_two_hundred_lines`
- AC-9 (REQ-12): 새 선택지는 계획에 받아들여지고 모르는 id 는 거절돼요. 스키마의 선택지 목록과 코드의 목록이 같아요 · verify: `cargo test --locked --offline --test agents_md_import_cli the_pick_is_a_known_id_and_an_unknown_one_is_refused`

## Scope

- write: 이 SPEC, `tests/agents_md_import_cli.rs`(새), `src/agents.rs`(계획 행·적용·`render_agents` 지시문·길이 경고), `schemas/protocol/v1/agents-request.schema.json`
- 기다려요(2026-11-21 예약 읽음): `src/agents.rs` 는 다른 카드가 쥐고 있어서 구현 조각은 그 카드가 돌려줄 때까지 못 시작해요. 선택지 목록과 워크플로 글(`src/project_update.rs`·`src/update.rs`·`src/ops/project_update_plan.rs`·`skills/moai/workflows/project.md`·`update.md`)도 다른 카드가 쥐고 있어서 후속 조각이에요.
- read-only: `tests/common/`, `src/worktree_include.rs`
- forbidden: `CLAUDE.md`·`AGENTS.md` 의 사용자 줄을 바꾸는 코드, 파일 안의 카드 번호, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것

## Plan

1. 조각 1(이번): 이 SPEC 과 RED 시험을 먼저 올려요(구현 전에 실패하는 것을 확인). `src/agents.rs` 가 풀리면 계획 행, 적용, `render_agents` 지시문, 길이 경고, 스키마를 구현해요.
2. 조각 2(후속, 쥔 카드가 풀면): `/moai project`·`moai update` 가 `claude-agents-import` 선택지를 내놓고, 이 저장소 자신의 `CLAUDE.md`·`AGENTS.md` 도 `moai update` 로 같은 블록을 받아요.

## Risks

- `AGENTS.md` 블록 문구를 바꾸면 이미 만든 프로젝트의 다음 계획이 `merge` 로 나와요. 개인 파일을 쓰는 프로젝트만 바뀌고 `none` 은 그대로라서(REQ-9) 범위가 좁아요. 사용자가 블록을 고쳤다면 기존 규칙대로 충돌로 보여요.
- `CLAUDE.md` 가 있는 프로젝트에서는 Claude Code 가 `AGENTS.md` 를 직접 읽지 않아요(공식 문서, D-3). 이 SPEC 의 블록이 그 틈을 막아요. 블록을 지우면 다시 틈이 생기므로 미리보기에서 이유를 한 줄로 보여요.
- 파일이 없는 `@AGENTS.local.md` 가 호스트에서 어떻게 되는지 공식 문서에 없어요(D-2). 그래서 개인 파일을 쓰기로 하지 않은 프로젝트에는 줄을 넣지 않아요. 개인 파일을 쓰기로 했다가 나중에 지운 프로젝트의 동작은 확인하지 않았어요.
- `CLAUDE.md` 의 관리 블록 표지는 HTML 주석이에요. 호스트가 그 주석을 다르게 읽는다는 근거는 조사에서 나오지 않았어요(D-3).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 선택(opt-in)이에요: 새 선택지 id `claude-agents-import` 를 고를 때만 `CLAUDE.md` 블록을 계획해요. 기존 `claude-import`(그림자 CLAUDE.md 에 한 줄 덧붙이기)는 그대로 두고, 새 선택지를 고르면 그 한 줄 행은 계획하지 않아요(같은 줄이 두 번 들어가지 않게) | 카드 본문("preview + pick"). 기존 선택지의 뜻을 바꾸면 기존 시험·사용자 흐름이 달라져요 | 기존 선택지에 합치기 |
| D-2 | `AGENTS.md` 블록의 개인 규칙 두 줄은 `local_file` 이 `ignored`·`tracked` 일 때만 넣어요. `none` 의 블록은 이전과 같아요 | 파일이 없을 때 `@path` 의 동작을 공식 문서가 말하지 않아요(2026-11-21 읽음). 모르는 동작을 모든 프로젝트에 퍼뜨리지 않고, 기존 프로젝트의 블록 다이제스트를 바꾸지 않으려는 거예요 | 항상 넣기 |
| D-3 | 근거는 공식 문서 직접 읽음(읽은 시각 2026-11-21T01:07Z, https://code.claude.com/docs/en/memory): "CLAUDE.md files can import additional files using `@path/to/import` syntax", "Import parsing skips Markdown code spans and fenced code blocks", "Inside each `AGENTS.md`: `@path` imports are expanded", "Not read: `AGENTS.local.md`, `AGENTS.override.md`", "An `AGENTS.md` and a `CLAUDE.md` or `CLAUDE.local.md` in your working directory or above it → Your `CLAUDE.md` files only", "Keeping the import never makes Claude read `AGENTS.md` twice". Codex 쪽은 카드 본문이 인용한 것(리더가 2026-11-21T00:2xZ 에 읽음)을 쓰고, 이 조각에서 따로 다시 읽지는 않았어요: 폴더마다 `AGENTS.override.md` → `AGENTS.md`, 불러오기 문법 없음, 합친 크기 상한 `project_doc_max_bytes` 32 KiB | 사용자 지시와 리더 카드 | 문서가 바뀌면 문구만 고쳐요 |
| D-4 | `CLAUDE.md` 블록은 HTML 주석 표지(`AGENTS.md` 와 같은 방식)이고 구역 이름은 `moai-claude-import` 예요. 블록 안에는 `@AGENTS.md` 한 줄만 둬요 | `markers`·`splice`·`record_owned` 를 그대로 써요 | 구역 이름 변경 |
| D-5 | 길이 경고는 줄 수(200줄)로 재요. Claude Code 문서의 권고가 "200줄 아래"라서요. Codex 의 32 KiB 경고는 이미 있어요 | 공식 문서 "target under 200 lines per CLAUDE.md file"(같은 읽음) | 임계값 변경 |

## Result

(아직 구현 전이에요.)
