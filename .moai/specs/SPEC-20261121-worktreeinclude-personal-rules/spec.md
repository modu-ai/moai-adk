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

- write: 이 SPEC, `src/agents.rs`, `src/claude.rs`(개인 경로 상수를 `pub(crate)` 로 한 줄), `schemas/protocol/v1/agents-request.schema.json`, `templates/worktreeinclude/personal.txt`(새), `tests/agents_worktreeinclude_cli.rs`(새), `tests/worktree_include_guard.rs`(D-5), `tests/agents_choices_step_cli.rs`(선택지 목록을 고정한 기존 시험에 새 id 한 줄)
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
| D-5 | 보호 시험 `tests/worktree_include_guard.rs` 의 "소스 중 `.worktreeinclude` 를 이름으로 대는 파일은 `worktree_include.rs` 뿐" 규칙에 `src/agents.rs` 한 파일만 더해요. 느슨하게 만드는 대신 새 시험 `the_packaged_template_lists_no_secret_file` 을 같은 파일에 더해 템플릿의 모든 줄이 비밀 검사를 통과하는지 고정해요. 배포 파일 검사(`.worktreeinclude` 라는 이름의 파일이 패키지에 없다)와 `include_str!` 검사는 그대로예요 | 이 규칙은 Go 세대가 `.env` 를 모든 워크트리에 복사하던 일을 막으려던 것이고(#2670), 사용자 지시(2026-11-20)로 개인 파일만 나열하는 고정 블록을 제공하기로 했어요. 비밀 파일 금지라는 규칙의 목적은 새 시험이 지켜요 | 허용 목록에서 `agents.rs` 를 빼고 선택지를 지우기 |

## Result

이 조각(조각 1: 계획·적용·영수증·스키마·템플릿)의 근거예요. 코드 head `a192464282`. 카드는 부분 READY 로 열어 두어요: 사용자에게 선택지를 내놓는 일(`/moai project`·`moai update` 의 선택지 목록과 두 워크플로 글)은 이 조각에 없고 후속(조각 2)이에요. 그 파일들(`src/project_update.rs`·`src/update.rs`·`src/ops/project_update_plan.rs`·`skills/moai/workflows/project.md`·`update.md`)을 다른 카드가 쥐고 있어요. 그래서 이 조각만으로는 요청에 `picks` 로 직접 넣을 때만 블록이 계획돼요. 소스 목록(`source-manifest.json`)에 새 자산 올리기도 병합 직전에 한 번에 해요.

Codex CLI 로 만든 워크트리가 `.worktreeinclude` 를 지키는지는 공식 문서가 말하지 않아요(D-3). 이 조각은 그렇다고 주장하지 않아요.

TDD: 시험을 먼저 써서 구현 전에 8개 중 7개가 실패하는 것을 확인했어요(`without_the_pick_nothing_is_planned` 만 통과: 구현 전에는 아무것도 계획하지 않으니까요). 구현 뒤 8개 모두 통과했어요.

### 변형 시험(커밋한 뒤, 한 번에 하나씩 바꾸고 `git checkout` 으로 되돌림)

| 변형 | 바꾼 것 | 잡은 시험 |
| --- | --- | --- |
| M1 | 선택이 없어도 항상 계획 | `without_the_pick_nothing_is_planned` |
| M2 | 코드의 개인 경로 목록에서 한 경로를 다른 것으로 바꿈 | `the_template_lists_exactly_the_personal_paths_and_no_secret` |
| M3 | 적용이 사용자 파일을 무시하고 블록만 씀 | `applying_keeps_every_user_line_and_a_second_plan_is_unchanged` |

세 변형 모두 시험이 실패해서 잡혔어요(생존 0). 변형 뒤 `git status` 로 `src/agents.rs` 가 깨끗한 것을 확인했어요.

### 기존 시험 세 곳을 고친 이유

- `tests/agents_choices_step_cli.rs`: 선택지 id 목록을 고정한 시험에 `worktree-include` 한 줄을 더했어요(새 id 가 스키마에 있는지 고정하는 시험이에요).
- `tests/worktree_include_guard.rs`: D-5 에 적었어요. `src/agents.rs` 한 파일만 허용하고 새 시험 `the_packaged_template_lists_no_secret_file` 을 더했어요.
- `tests/agents_worktreeinclude_cli.rs`: 영수증 시험이 폴더를 걷지 않도록 영수증 경로를 계산해서 읽어요(`repo_scans` 표에 올리지 않으려고요).

### Acceptance 결과

- AC-1: met — `--test agents_worktreeinclude_cli the_pick_plans_one_managed_block_with_the_template_lines`(`create`, 미리보기에 템플릿 두 경로, 계획은 파일을 안 씀) 통과
- AC-2: met — `--test agents_worktreeinclude_cli the_block_is_the_same_whatever_the_local_file_setting`(`ignored`·`tracked`·`none` 세 프로젝트의 블록이 같음) 통과
- AC-3: met — `--test agents_worktreeinclude_cli the_template_lists_exactly_the_personal_paths_and_no_secret` 와 `--test worktree_include_guard the_packaged_template_lists_no_secret_file` 통과
- AC-4: met — `--test agents_worktreeinclude_cli applying_keeps_every_user_line_and_a_second_plan_is_unchanged`(CRLF·마지막 줄바꿈 없음, 다시 계획하면 `unchanged`) 통과
- AC-5: met — `--test agents_worktreeinclude_cli without_the_pick_nothing_is_planned` 통과
- AC-6: met — `--test agents_worktreeinclude_cli the_apply_receipt_keeps_the_file_as_it_was_and_names_it`(영수증 쪽에 적용 전 파일) 통과
- AC-7: met — `--test agents_worktreeinclude_cli the_pick_is_a_known_id_and_an_unknown_one_is_refused` 와 `--test agents_choices_step_cli schema_describes_picks_claude_shadow_and_the_skip_action` 통과
- AC-8: met — `--test agents_conflict_choices_cli picks_absent_or_empty_plan_digest_and_apply_as_before` 통과(영역 시험 `agents_*` 전체 실행에 포함)
- AC-9: met — `--test agents_worktreeinclude_cli the_applied_file_raises_no_secret_warning` 통과

### 검사 결과(필터 없이, 마지막 줄만 다시 실행한 것 아님)

`git diff --check`, `cargo fmt --check`, `rustfmt --edition 2021 --check src/ops/*.rs`, `cargo clippy --all-targets -- -D warnings`, `--lib`, `--bins`, 영역 시험(`agents_*`·`project_*`·`worktree_include_guard`·`bundle_user_assets_cli`, 246개 통과), 저장소 전체 훑기 시험(`git_env_scrub`·`home_isolation_scan`·`windows_unix_gate_scan`·`product_cli`), macOS 전용 시험, `tools/repo_scans.json` 의 검사 41개(실패 0), 파이썬 시험 148개, `trailer-check 33a5846038..HEAD` 모두 종료 코드 0 이에요. 앞서 한 번 돌린 검사에서는 4곳(위 세 곳과 `repo_scans` 표)이 실패했고, 고친 뒤 처음부터 다시 돌려 모두 통과했어요.

돌리지 않은 것: 조각 2(선택지 목록·워크플로 글)가 없어서 `/moai project`·`moai update` 의 실제 화면 흐름은 확인하지 않았어요. 소스 목록에 자산을 올리는 일과 그 패키지 검사도 병합 직전으로 미뤘어요.
