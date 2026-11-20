---
id: SPEC-20261121-worktreeinclude-personal-rules
title: "MoAI gives every project the same .worktreeinclude block for its personal files"
status: draft
revision: 2
created: 2026-11-21
---

# MoAI gives every project the same .worktreeinclude block for its personal files

## Intent

호스트가 새로 만드는 워크트리(`claude --worktree`, `moai cc -w`, `isolation: worktree` 서브에이전트)는 git 이 추적하는 파일만 가지고 시작해요. 그래서 개인 규칙 `AGENTS.local.md` 나 `.claude/settings.local.json` 처럼 git 에 올리지 않는 개인 파일은 새 워크트리에 없고, 새 워크트리의 에이전트는 개인 규칙 없이 일하고 자기 개선 내용(관리 블록)도 못 읽어요. Claude Code 는 프로젝트 루트의 `.worktreeinclude` 가 고르는 "git 이 무시하는 파일"을 새 워크트리에 복사해 주는데, MoAI 제품(Rust)은 이 파일을 지금까지 한 번도 쓰지 않았어요(읽어서 비밀 경고만 해요). 사용자 지시(2026-11-20, 리더 노트로 전달): "`.worktreeinclude` 를 추가해서 고정적으로 모든 사용자에게 동일하게 제공". 이 SPEC 은 사용자가 `worktree-include` 를 고르면 계획(plan)에 `.worktreeinclude` 의 관리 블록 하나를 넣어요. 블록의 줄은 모든 프로젝트에서 같은 고정 목록 하나(MoAI 가 개인 파일로 다루는 경로)이고, 패키지 자산 템플릿 한 곳에만 있어요. 미리보기에서 줄이 보이고, 사용자가 블록 밖에 쓴 줄은 그대로이며, 적용은 영수증(receipt)에 남아 되돌릴 수 있어요. git 이 무시하지 않는 파일은 호스트가 복사하지 않으므로 블록에 있어도 무해해요(공식 문서, D-3). 모든 사용자 프로젝트에 같은 말이 통해요(이 저장소 전용 도구 없음).

## Non-goals

- `.worktreeinclude` 의 사용자 줄을 고치거나 지우기: 하지 않아요. 관리 블록 바깥은 한 바이트도 안 바꿔요.
- 비밀 파일(`.env`, 키, 인증서): 블록에 절대 넣지 않아요. 템플릿의 모든 줄이 `src/worktree_include.rs` 의 비밀 검사를 통과하는지 시험으로 고정해요. 그 검사 코드는 바꾸지 않아요.
- 프로젝트마다 다른 목록, 사용자가 목록을 고르기: 하지 않아요. 고정 목록 하나예요(사용자 지시).
- 세션이 이 선택지를 `/moai project` 와 `moai update` 의 선택지 목록에 내놓는 부분과 두 워크플로 글: 이 조각에는 없어요. 그 파일들을 다른 카드가 쥐고 있어서 후속 조각이에요(Scope 참고).
- 머신 출력(적용 영수증 폴더, 상태 폴더 `/*`): 이번 목록에 넣지 않아요. 새 워크트리가 복사할 개인 파일이 아니라 도구가 만든 출력이에요(D-4).

## Requirements

- REQ-1: When a request picks `worktree-include`, the generation plan shall carry one managed block in `.worktreeinclude` whose lines are the personal-file list read from the packaged template, with those lines in the row's preview. · failure: none
- REQ-2: The managed block shall hold the same lines for every project, whatever the request's `local_file` is and whether the personal files exist. · failure: none
- REQ-3: The packaged template shall list no secret file, so that every line of it passes the secret scan of `.worktreeinclude`. · failure: none
- REQ-4: The packaged template's file list shall equal the list of personal paths MoAI's code names. · failure: none
- REQ-5: When the plan is applied, the apply shall write only the managed block and keep every byte of the file outside it. · failure: none
- REQ-6 (REQ-5): IF the same request is planned again after the apply, THEN the plan shall show the block as `unchanged`.
- REQ-7: When a request carries no `worktree-include` pick, the plan shall carry no `.worktreeinclude` row. · failure: none
- REQ-8: When a request carries no picks at all, the plan digest shall be the one it was before this pick existed. · failure: none
- REQ-9: When the apply changes `.worktreeinclude`, the receipt shall name it and keep the file as it was before the apply, so a revert restores it. · failure: none
- REQ-10: The request schema and the pick list shall name `worktree-include` and refuse any other unknown id as before. · failure: none

## Acceptance

- AC-1 (REQ-1): 선택지가 있으면 `.worktreeinclude` 행이 `create`(파일 없음)이고 미리보기에 템플릿의 두 경로가 `+` 줄로 있고 계획은 아무것도 쓰지 않아요 · verify: `cargo test --locked --offline --test agents_worktreeinclude_cli the_pick_plans_one_managed_block_with_the_template_lines`
- AC-2 (REQ-2): `local_file` 이 `ignored`·`tracked`·`none` 인 세 프로젝트에 적용하면 블록의 줄이 모두 같아요 · verify: `cargo test --locked --offline --test agents_worktreeinclude_cli the_block_is_the_same_whatever_the_local_file_setting`
- AC-3 (REQ-3, REQ-4): 템플릿의 비밀 검사 결과가 비어 있고, 템플릿의 경로 목록이 코드의 개인 경로 목록(`agents::PERSONAL_PATHS`)과 같아요 · verify: `cargo test --locked --offline --test agents_worktreeinclude_cli the_template_lists_exactly_the_personal_paths_and_no_secret`
- AC-4 (REQ-5, REQ-6): 사용자 줄(CRLF 와 마지막 줄바꿈 없음 포함)이 있는 파일에 적용하면 사용자 바이트가 앞에 그대로 있고 블록은 한 번만 있고, 다시 계획하면 `unchanged` 예요 · verify: `cargo test --locked --offline --test agents_worktreeinclude_cli applying_keeps_every_user_line_and_a_second_plan_is_unchanged`
- AC-5 (REQ-7): 선택지 없이 계획하면 `.worktreeinclude` 행이 없어요 · verify: `cargo test --locked --offline --test agents_worktreeinclude_cli without_the_pick_nothing_is_planned`
- AC-6 (REQ-9): 적용 결과가 `.worktreeinclude` 를 이름으로 대고 적용 전 파일을 영수증 쪽에 보관해요 · verify: `cargo test --locked --offline --test agents_worktreeinclude_cli the_apply_receipt_keeps_the_file_as_it_was_and_names_it`
- AC-7 (REQ-10): 새 선택지는 계획에 받아들여지고 모르는 id 는 거절돼요. 스키마의 선택지 목록과 코드의 목록이 같아요 · verify: `cargo test --locked --offline --test agents_worktreeinclude_cli the_pick_is_a_known_id_and_an_unknown_one_is_refused`
- AC-8 (REQ-8): 선택지가 하나도 없는 요청의 계획 digest 는 이 선택지가 생기기 전과 같아요(기존 시험이 고정) · verify: `cargo test --locked --offline --test agents_conflict_choices_cli picks_absent_or_empty_plan_digest_and_apply_as_before`
- AC-9 (REQ-1): 적용된 파일이 `src/worktree_include.rs` 의 비밀 경고를 만들지 않아요 · verify: `cargo test --locked --offline --test agents_worktreeinclude_cli the_applied_file_raises_no_secret_warning`

## Scope

- write: 이 SPEC, `src/agents.rs`, `src/claude.rs`(개인 경로 상수를 `pub(crate)` 로 한 줄), `schemas/protocol/v1/agents-request.schema.json`, `templates/worktreeinclude/personal.txt`(새), `tests/agents_worktreeinclude_cli.rs`(새)
- 쥔 카드가 풀 때까지 기다려요(2026-11-21 예약 읽음): `src/project_update.rs`·`src/update.rs`·`src/ops/project_update_plan.rs`(다른 카드가 쥐어요: `/moai project`·`moai update` 의 선택지 목록과 update 화면), `skills/moai/workflows/project.md`·`update.md`(다른 카드가 쥐어요: 워크플로 글), `skills/moai/references/source-manifest.json`(여러 카드가 쥐어요: 새 자산이 패키지에 들어가는지는 병합 직전에 한 번에 확인)
- read-only: `src/worktree_include.rs`, `tests/common/`
- forbidden: `.worktreeinclude` 의 사용자 줄을 바꾸는 코드, 비밀 파일 이름을 목록에 넣는 것, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것

## Plan

1. 조각 1(이번): 계획 쪽. 템플릿 자산 하나, `PICKS` 에 `worktree-include`, 관리 블록 구역 하나(`.worktreeinclude`, 해시 주석 표지), 계획 행, 적용, 스키마, 시험(RED 먼저). 선택지를 사용자에게 내놓는 목록과 워크플로 글은 후속.
2. 조각 2(후속, 쥔 카드가 풀면): `generation_choices` 가 `/moai project`(새 프로젝트)와 `moai update`(기존 프로젝트) 두 흐름에서 `worktree-include` 선택지를 내놓고(미리보기 → 선택 → 영수증 → 되돌리기), 두 워크플로 글이 한 줄로 말해요. 소스 목록에 자산을 올려요.

## Risks

- 선택지가 계획에 있어도 사용자에게 내놓는 목록이 없으면 이 조각만으로는 쓸 수 없어요(요청에 직접 `picks` 로 넣을 때만 돼요). 그래서 후속 조각이 필요하고 Result 에 그렇게 적어요.
- `.worktreeinclude` 는 `.gitignore` 문법이라 해시 주석 표지(`# moai:begin …`)가 유효한 주석이에요. 호스트가 파일 안의 주석을 다르게 읽는다는 근거는 없어요(조사에서 확인한 만큼만 믿어요, D-3).
- Codex CLI(`--worktree`, `/worktree`)로 만든 워크트리가 `.worktreeinclude` 를 지키는지는 공식 문서에 없어요(D-3). 그 경로에서는 블록이 있어도 파일이 안 올 수 있어요. Result 와 후속 안내 문장에 "문서가 말하지 않음"이라고 적고, 공식 문서가 아닌 GitHub 이슈는 근거로 쓰지 않아요.
- 사용자가 블록 앞에 `!AGENTS.local.md` 같은 부정 줄을 썼다면 블록의 줄이 뒤에 와서 이겨요(gitignore 문법에서 뒤의 줄이 이겨요). 사용자 줄을 바꾸지 않는 것이 이 SPEC 의 약속이라서 그대로 두고, 필요하면 후속에서 미리보기에 알려요.
- 패키지 자산을 소스 목록(`source-manifest.json`)에 올리는 일은 여러 카드가 쥔 파일이라 이 조각에서 하지 않아요. 새 자산이 패키지 검사에 걸리는지는 이 조각의 READY 검사에서 읽어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 선택(opt-in)이에요: `worktree-include` 를 고를 때만 계획해요 | 카드 본문("미리보기 → 선택"), `.worktreeinclude` 는 사용자 파일이에요 | 기본으로 켜기 |
| D-2 | 구역은 `moai-worktree`, 표지는 해시 주석이에요(`.gitignore` 블록과 같은 방식). 템플릿은 `templates/worktreeinclude/personal.txt` 하나이고 생성기가 `include_str!` 로 읽어요(`templates/spec/` 와 같은 방식) | `.worktreeinclude` 는 `.gitignore` 문법이고 `markers`·`splice`·`record_owned` 를 그대로 써요. 사용자 지시: 템플릿은 패키지 자산 하나, 생성기가 읽기 | 구역 이름 변경(사용자 파일에 남은 블록은 표지로 찾아요) |
| D-3 | 같은 `.worktreeinclude` 블록 하나로 두 호스트를 다뤄요. Claude Code 는 문서대로 무시된 파일 중 패턴이 고른 것만 복사하고(프로젝트 루트), Codex 데스크톱 앱도 같은 이름의 파일을 저장소 루트에서 읽어요. Codex CLI 로 만든 관리 워크트리가 이 파일을 지키는지는 공식 문서가 말하지 않아서 이 SPEC 은 그렇다고 주장하지 않아요. 따로 복사 단계는 만들지 않아요 | 공식 문서 직접 읽음(읽은 시각 2026-11-20 22:12~22:13 UTC): "Only files that match a pattern and are also gitignored are copied, so tracked files are never duplicated." (https://code.claude.com/docs/en/worktrees). Codex: "add a `.worktreeinclude` file to the repository root and list the ignored paths or `.gitignore`-style patterns to copy when Codex creates a managed worktree." 와 "This behavior applies to local ChatGPT desktop app managed worktrees, not remote worktrees or Git worktrees you create yourself from the command line." (https://learn.chatgpt.com/docs/environments/git-worktrees.md, `developers.openai.com/codex/app/worktrees` 가 여기로 308 이동). 두 호스트 모두 `AGENTS.local.md` 를 기본으로는 새 워크트리에 안 가져와요(Codex 는 무시된 `AGENTS.override.md` 만 자동 복사) | Codex CLI 지원이 문서로 확인되면 후속에서 안내 문장만 더해요 |
| D-4 | 목록은 `.claude/settings.local.json` 과 `AGENTS.local.md` 둘이에요. 코드의 실제 목록에서 가져왔어요: Claude Code 가 같은 이름의 개인 설정으로 읽는 `claude.rs` 의 `SETTINGS_LOCAL_FILE`, MoAI 가 `.gitignore` 블록에 올리는 개인 규칙 파일. MoAI 의 `.gitignore` 블록에 있는 나머지(영수증 폴더, 상태 폴더)는 도구 출력이라 넣지 않아요. 새 개인 경로가 생기면 상수와 템플릿을 함께 고치고 시험이 둘의 차이를 잡아요 | 리더 노트(사용자 지시): "moai 가 개인 파일로 gitignore 하는 경로를 src/agents.rs·src/claude.rs 의 실제 목록에서 가져와 고정 목록 하나로, 테스트가 두 목록이 같은지 확인" | 목록에 경로 추가 |

## Result

(아직 구현 전이에요.)
