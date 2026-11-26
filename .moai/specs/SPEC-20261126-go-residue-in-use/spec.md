---
id: SPEC-20261126-go-residue-in-use
title: "go-residue keeps the Go files something still reads and names the readers"
status: draft
revision: 1
created: 2026-11-26
---

# go-residue keeps the Go files something still reads and names the readers

## Intent

`/moai update` 의 `go-residue` 단계는 Go 매니페스트와 바이트가 같은 파일을 `go_same` 으로 분류해서 보관(archive) 명령에 넣어요. 지금 읽는 곳이 있는지는 보지 않아요(`src/go_residue.rs:350-368`, `src/go_residue_apply.rs:55-75`). 실제 프로젝트 `mo.ai.kr` 을 2026-11-26 에 읽기만 해서 확인한 것은 이래요.

- `CLAUDE.md` 가 `@.moai/config/sections/user.yaml`, `@.moai/config/sections/language.yaml` 을 가져와요(115, 116행).
- `.claude/settings.local.json` 의 `outputStyle` 이 `MoAI-Easy` 이고 파일은 `.claude/output-styles/moai/moai-easy.md` 예요.
- 두 위치 모두 Go 매니페스트(`.moai-backup/manifest.json`)에 `template_managed` 항목이 있는 폴더예요.

이 SPEC 은 읽는 곳이 있는 파일을 보관 후보에서 빼고, 미리보기에 읽는 곳을 보여 줘요. 조사(survey)가 보지 않던 `.moai/docs/local-rules` 와 사용자가 만든 `.claude` 파일도 보여 주기만 해요. 옮기는 일은 지금보다 줄기만 하고 늘지 않아요.

근거:

- `https://code.claude.com/docs/en/memory.md`, 2026-11-26T11:23Z 에 읽음, sha256 `7e52db857de441b381a3ee7ce1182a4267f6bbddee08000574e61e54fc335eac`. 가져온 규칙: `@path` 가져오기, 상대 경로는 가져오는 파일 기준, 최대 네 단계, 코드 구간과 코드 블록은 건너뜀, 공백은 `\` 로 이스케이프, 따옴표로 감싸면 가져오지 않음, 4 MiB 를 넘는 파일은 건너뜀, 작업 폴더 밖 가져오기는 사용자 승인이 필요함. 적용: REQ-2, REQ-3.
- `mo.ai.kr` 을 2026-11-26 에 읽기만 했어요(아무것도 쓰지 않았어요): 위의 파일과 설정의 읽는 곳 키, `.moai-backup/manifest.json` 의 경로 목록. REQ-3 의 조건은 하나도 걸리지 않아 보였어요: `CLAUDE.md`, `AGENTS.md`, `AGENTS.local.md`, `CLAUDE.local.md` 에 밖 가져오기 0건, 링크 아님, `.claude/rules` 없음, 설정의 상태줄과 훅은 설치된 `moai` 실행 파일 호출이에요. 구현으로 확인한 것은 아니에요.
- 제품의 훅과 claim 코드(`src/hook*.rs`, `src/claim*.rs`)를 `config/sections`, `.yaml`, ast-grep, `.moai/workflows` 글자로 찾았을 때 읽는 Go 시대 파일은 없었고 `.moai/project.yaml`(이번 세대 파일, 조사에서 빠짐)만 나왔어요. 상태줄 목록은 `src/statusline/state.rs:432-480`, `:822`, `:874` 에서 옮겼어요.

읽는 사람: 승인하는 사용자와 리더, 구현 레인, 감사자.

## Interview

- Q1: 무엇을 원하나요? → A: "/moai update가 제대로 잘 안되는거 같은데? 과거 go 세대 제거 및 업데이트가 되어야 하는데 왜 안되는지 전수 조사해라" (사용자 지시 2026-11-26, 리더가 작업 카드 본문에 옮겨 적었어요).
- Q2: 아래 Decisions 의 "승인 때 골라요" 항목(D-1, D-2, D-6, D-7) → A: 아직 답이 없어요. 승인 노트의 답을 따르고, 답이 없으면 구현하지 않아요.

## Non-goals

- 지시 글이 파일 이름을 문장으로만 부르는 것(`CLAUDE.md` 나 규칙, 스킬 글 속의 이름)은 읽는 곳으로 세지 않아요(D-1 이 바뀌면 이 줄도 바뀌어요).
- 프로젝트 밖의 설정과 지시 파일(`~/.claude/**`)은 열지 않아요. 하위 폴더의 `CLAUDE.md`, 플러그인과 스킬의 frontmatter 훅, 코드가 계산해서 만드는 경로도 찾지 않아요. 이 경우들은 놓칠 수 있고, 옮긴 파일은 `go-residue-revert` 로 되돌려요.
- `go-state` 단계가 `CLAUDE.md` 가져오기를 지키는 일은 하지 않아요(`src/go_state.rs` 는 바꾸지 않아요).
- `skills/moai/workflows/update.md` 의 분류 설명은 바꾸지 않아요(D-10).
- 탐지 SPEC(`SPEC-20261121-go-residue-detect`)의 REQ-9 문장과 `src/project_update.rs:2072-2074` 주석 중 SPEC 쪽은 고치지 않아요. 주석만 코드에 맞춰요.

## Requirements

<!-- 이 SPEC 은 옮기는 파일을 줄이는 쪽으로만 바꿔요. 읽는 곳을 확인하지 못하면 옮기지 않아요. -->
- REQ-1: When the go-residue survey finds a `go_same` or `go_changed` file that a reader of REQ-2 names, the survey shall list it as class `go_in_use` with the suggested action `keep` and a `used_by` list that holds, for each reader, its project path and how it names the file (`import`, `settings`, `script` or `statusline`), and no setting value.
- REQ-2: The survey shall treat these project files as readers and count every file one of them names as read: `CLAUDE.md`, `.claude/CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `AGENTS.local.md`, `.claude/rules/**/*.md` and the files they import with `@path` (relative to the importing file, at most four hops, code spans and fenced blocks skipped), where an import is the name; `.claude/settings.json` and `.claude/settings.local.json`, where any string outside `permissions` is a name and `outputStyle` names the style file whose file name or frontmatter `name` matches it; and the project scripts those strings name, followed at most four hops. The survey also counts as read the status line's own files `statusline.yaml`, `llm.yaml` and `system.yaml` under `.moai/config/sections/` and `.moai/statusline.local.yaml`. · failure: REQ-3
- REQ-3 (REQ-1, REQ-2): IF a reader of REQ-2 that exists cannot be read (a read error, a link on its path, more than 4 MiB, settings that are not JSON, text that is not UTF-8, a secret-shaped name), the listing of `.claude/rules` is cut at its limit, an `@path` import points outside the project, or a settings string names a script outside the project (a path passed to `bash`, `sh`, `zsh`, `node`, `python3`, `ruby`, `perl`, `deno` or `bun`, or a path ending in `.sh`, `.bash`, `.zsh`, `.js`, `.mjs`, `.cjs`, `.ts`, `.py`, `.rb` or `.ps1`; an installed program such as `$HOME/.moai/bin/moai` is none of these and is not followed), THEN the survey shall suggest `keep` for every `go_same` and `go_changed` file and list each such reader with its reason under `unread_readers`, which leaves the step without an apply command.
- REQ-4: The `go-residue-apply` command shall move only files that its own fresh survey lists neither as `go_in_use` nor as kept under REQ-3, and refuse a `--pick` of any other file with a message that names a reader. · failure: none — 거절은 아무것도 옮기지 않고, 미리보기 뒤에 참조가 생기면 다이제스트(digest)가 달라져 기존의 stale 거절이 나요
- REQ-5: The survey shall list each file under `.moai/docs/local-rules/`, `.claude/commands/harness/` and each `.claude/workflows/hns-*` file with the suggested action `keep`, outside every apply bundle, and mark the preview when that listing was cut at its limit. · failure: none — 보여 주기만 해요
- REQ-6: The go-residue step shall read `optional` while a `go_same` or `go_changed` file is left to move and `unchanged` otherwise, with a reason that gives the number of files in use and of unread readers and with no `needs_decision` state. · failure: none — 상태를 정하는 규칙이에요
- REQ-7 (REQ-6): The SPEC `SPEC-20261125-go-residue-apply` shall state in its REQ-10 and AC-1, as revision 2, the state rule of REQ-6 and that a referenced output style stays in place. · failure: none — 문서 요구예요

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 프로젝트에서 `@` 가져오기(두 단계 포함), `outputStyle`, 설정의 상태줄 스크립트, 훅 스크립트가 읽는 파일, 상태줄의 yaml 이 각각 `go_in_use` 와 `keep` 이고 `used_by` 가 읽는 곳의 경로와 종류를 대며, 코드 블록 속의 `@` 와 `permissions` 의 글자는 읽는 곳이 아니고, 읽는 곳이 없는 증명된 파일은 `go_same` 으로 남아요 · verify: `cargo test --locked --test project_update_go_residue_cli a_go_file_still_read_is_in_use_and_names_its_reader`
- AC-2 (REQ-3): 깨진 JSON 설정, 링크인 지시 파일, UTF-8 이 아닌 스크립트, 프로젝트 밖 `@` 가져오기, 설정이 `bash` 로 부르는 프로젝트 밖 스크립트 각각에서 모든 `go_same`/`go_changed` 가 `keep` 이고 `unread_readers` 가 그 파일과 까닭을 대며 적용 명령이 없고 트리는 그대로예요. 읽는 곳 파일이 아예 없는 것은 막지 않아요 · verify: `cargo test --locked --test project_update_go_residue_cli a_reader_that_cannot_be_read_keeps_every_go_file`
- AC-3 (REQ-4): 쓰는 중인 파일은 증명 묶음 적용에 들어가지 않고 `--pick` 은 읽는 곳을 대며 거절하고, 미리보기 뒤에 참조를 더하면 인쇄된 명령이 stale 로 거절되고 아무것도 옮기지 않아요 · verify: `cargo test --locked --test project_update_go_residue_cli an_in_use_file_is_never_moved_and_a_new_reference_makes_the_plan_stale`
- AC-4 (REQ-5): `.moai/docs/local-rules` 파일(매니페스트가 이름 붙인 것 포함), `.claude/commands/harness` 파일, `.claude/workflows/hns-` 파일이 `keep` 으로 보이고 적용 전후 바이트가 같으며 `.claude/worktrees` 아래는 보지 않아요 · verify: `cargo test --locked --test project_update_go_residue_cli the_survey_lists_local_rules_and_harness_files_and_moves_none`
- AC-5 (REQ-6): 후보가 남으면 `optional`, 쓰는 중이거나 못 읽는 것만 남으면 `unchanged`, 어느 프레임에서도 `needs_decision` 이 없고 이유 글에 두 수가 있어요 · verify: `cargo test --locked --test project_update_go_residue_cli the_step_is_optional_with_a_candidate_and_never_needs_decision`
- AC-6 (REQ-1, REQ-2, REQ-4): `mo.ai.kr` 모양의 임시 프로젝트(`.moai-backup/manifest.json` 만, `CLAUDE.md` 의 네 가져오기, `settings.local.json` 의 `outputStyle: MoAI-Easy`, 설치 실행 파일의 `statusline` 명령과 `moai` 훅 명령)에서 계획에 `unread_readers` 가 없고, 인쇄된 명령이 읽는 곳이 없는 증명된 파일만 옮기며 가져온 yaml 과 출력 스타일은 그 자리에 있어요 · verify: `cargo test --locked --test project_update_go_residue_cli the_mo_ai_kr_shape_moves_only_unread_proven_files`
- AC-7 (REQ-7): `moai spec show SPEC-20261125-go-residue-apply` 가 판 2 를 읽고 REQ-10, AC-1 이 REQ-6 과 같은 말을 해요 · manual: 감사자가 두 문장을 REQ-6 과 대조해요
- AC-8 (REQ-1, REQ-4): 갱신 흐름 전체를 도는 테스트가 통과하고, 설정이 이름을 부르는 훅 스크립트는 `go-residue` 가 도는 시점에 옮겨지지 않아요 · verify: `cargo test --locked --test update_end_to_end_cli`

## Scope

- write: `src/go_residue.rs`, `src/go_residue/readers.rs`(새 파일), `src/go_residue_apply.rs`, `src/project_update.rs`(`go_residue_step` 과 그 주석만), `tests/project_update_go_residue_cli.rs`, `tests/update_end_to_end_cli.rs`, `.moai/specs/SPEC-20261125-go-residue-apply/**`, `.moai/specs/SPEC-20261126-go-residue-in-use/**`
- read-only: `src/go_state.rs`(`read_bounded_text` 를 불러 써요), `src/statusline/state.rs`
- forbidden: `src/go_state.rs` 와 `src/statusline/**` 의 변경, 실제 프로젝트(`mo.ai.kr` 포함)와 실제 HOME 에 쓰기, 파일 삭제

## Plan

1. 이 SPEC 과 적용 SPEC 판 2 를 승인받아요.
2. `tests/update_end_to_end_cli.rs` 를 바꾸는 `SPEC-20261126-go-local-settings-reference` 가 착지했는지 확인하고, 착지했으면 NeXT 끝을 합친 뒤 시작해요. `src/go_residue.rs` 와 같은 테스트를 가진 `SPEC-20261126-claude-md-fresh-merge` 의 담당과는 함수와 줄 범위를 pair 노트로 나눠요.
3. 빨간 테스트(AC-1 에서 AC-6)를 먼저 써요. 프레임은 지금 테스트 파일의 `survey_frame` 을 이어 써요. 그 프레임의 `moai-easy.md` 는 설정이 이름을 부르므로 지금 테스트 `applying_the_proven_bundle_moves_only_the_go_same_files` 의 옮긴 수는 3 에서 2 가 돼요.
4. `src/go_residue/readers.rs` 에 읽는 곳 모으기를 두고, `survey` 가 후보에 적용해요. 적용은 이미 새로 조사하므로(`src/go_residue_apply.rs:186-191`) 거절 문구만 더해요.
5. 보여 주기 전용 목록(REQ-5)과 단계의 이유 글, 주석을 고쳐요.
6. 위험 경로는 임시 프로젝트와 임시 HOME 에서 진짜 실행 파일을 돌리는 Rust 테스트로 확인하고, 가드마다 `guard-mutant:` 줄을 READY 에 적어요.

## Risks

- 지금 `update_end_to_end_cli` 는 단계 순서에 `go-residue` 를 두고(`tests/update_end_to_end_cli.rs:316-326`), 그 프레임의 설정은 Go 훅 스크립트 두 개를 아직 이름으로 불러요(`:44`). 코드로 읽은 예상(아직 안 돌렸어요): 규칙 파일은 앞의 `go-state` 가 옮기므로 이 변경 뒤에는 `go-residue` 에 후보가 없어 순서가 `go-settings`, `go-state`, `generation`, `hooks` 가 돼요. 빨간 테스트에서 확인하고, 예상과 다르면 이 줄을 고쳐요.
- Go 훅 스크립트는 `hooks` 단계(`go-hooks-apply`)가 설정을 바꾼 뒤에야 안 쓰는 파일이 돼요. 갱신 흐름 테스트의 반복은 `all_unchanged` 에서 끝나므로(`:196-210`) 같은 반복에서는 옮기지 않고, 다음 `/moai update` 에서 `optional` 후보로 보여요. Go 흔적이 한 번에 다 없어지지는 않아요.
- 읽는 곳 규칙과 제품의 실제 읽기가 어긋날 수 있어요(상태줄 yaml 목록은 `src/statusline/state.rs:432-480`, `:822`, `:874` 를 옮겨 적어요). 새 읽기가 생기면 목록도 고쳐야 해요.
- 문장으로만 부르는 파일은 옮겨져요: `mo.ai.kr` 에서 `quality.yaml`(`CLAUDE.md:69`, `AGENTS.md:328`), `harness.yaml`(`AGENTS.md:328`), `.moai/workflows/project/`(`AGENTS.md:396`)가 그래요. 코드나 `skills/` 에서 이 파일을 읽는 곳은 계획 조사에서 찾지 못했어요(확인하지 못한 것으로 남겨요). 옮김은 되돌릴 수 있어요.
- `SPEC-20261126-claude-md-fresh-merge`(`CLAUDE.md` 새로 합치기)가 가져오기를 없애면 그 yaml 은 다시 후보가 돼요. 의도한 동작이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 (승인 때 골라요) | 읽는 곳은 `@` 가져오기, 설정, 설정이 이름 부른 스크립트, 상태줄 파일이에요(기본값 A). 대안 B: 지시 글과 `.claude` 의 스킬·명령 글 속 경로 이름도 세요 | 카드 완료 기준 1 이 세 가지를 대요. B 는 남는 Go 파일이 훨씬 많아져요. 카드 본문이 든 `harness.yaml`·`quality.yaml`·ast-grep 규칙·`.moai/workflows` 는 A 에서 보관 후보로 남아요 | 규칙을 더하면 후보가 줄어요 |
| D-2 (승인 때 골라요) | 프로젝트 밖을 가리키는 가져오기나 스크립트가 있으면 모두 `keep`, 적용 명령 없음(기본값 A). 대안 B: 이름만 알리고 나머지 읽는 곳으로 분류해요. 설치된 `moai` 실행 파일을 부르는 명령은 어느 쪽에서도 막지 않아요 | 못 읽은 곳이 있으면 안 옮겨요. A 는 `@~/…` 를 쓰는 사용자의 Go 보관을 모두 멈춰요. `mo.ai.kr` 의 `CLAUDE.md` 에는 밖 가져오기가 없었어요 | 규칙 한 곳 |
| D-3 | `go_changed` 도 읽는 곳이 있으면 `go_in_use` 예요 | 고른 파일도 옮기면 읽는 곳이 깨져요 | 규칙 한 곳 |
| D-4 | 읽는 곳 파일이 보관 후보여도 읽는 곳으로 세요 | 옮기기 전에는 읽혀요 | 규칙 한 곳 |
| D-5 | `.claude` 의 새 목록은 `commands/harness/**` 와 `workflows/hns-*` 뿐이고 보여 주기만 해요 | 카드 본문의 두 모양. 매니페스트가 이름 붙인 `.claude` 파일은 지금도 조사해요 | 모양을 더해요 |
| D-6 (승인 때 골라요) | 완료 기준 3 은 SPEC 을 코드에 맞춰요: 적용 SPEC 의 REQ-10, AC-1 을 판 2 로 고쳐요(기본값 A). 대안 B: 코드가 남은 파일이 있을 때 `needs_decision` 을 내고 거절 기록을 새로 만들어요 | B 는 `all_unchanged` 를 붙잡던 문제를 다시 열고 `SPEC-20261126-claude-md-fresh-merge`, `SPEC-20261122-project-ci-design` 이 바꾸는 파일을 건드려요. A 는 위임 승인 노트 9548 이 가리키는 적용 SPEC 판 1 다이제스트를 바꾸므로 판 2 승인이 따로 필요해요 | 판 1 로 되돌려요 |
| D-7 (승인 때 골라요) | tier 2 로 다뤄요(기본값) | 사용자 파일을 옮기는 경로이고 적용 SPEC 이 tier 2 를 선언했어요. 바닥 경로 목록으로는 tier 1 이에요(`src/go_residue*` 는 목록에 없어요) | 리더가 tier 1 로 정해요 |
| D-8 | `go_state::read_bounded_text` 는 불러 쓰고, 출력 스타일 이름 맞추기는 `readers.rs` 에 따로 둬요 | `src/go_state.rs` 는 다른 작업이 바꾸고 있는 tier 2 경로라 바꾸지 않아요. 두 곳이 어긋날 수 있어요 | 나중에 한 곳으로 합쳐요 |
| D-9 | 테스트는 지금 파일 `tests/project_update_go_residue_cli.rs` 를 늘려요 | 카드 범위의 `tests/go_residue_cli.rs` 는 없는 파일이에요 | 파일을 나눠요 |
| D-10 | `update.md` 는 바꾸지 않아요 | 완료 기준에 없고 `SPEC-20261126-claude-md-fresh-merge` 가 같은 파일을 바꿔요. 새 분류 `go_in_use` 설명은 후속이에요 | 후속에서 더해요 |

## Result

`status: draft`예요. 승인은 이 파일에 기록하지 않았어요(리더 노트 10375 가 사용자 승인을 전해요: D-1=A, D-2=A, D-6=A, D-7 tier 2). 빨간 테스트 커밋은 `386086932e` 예요.

한 일: `src/go_residue/readers.rs`(새 파일)가 읽는 곳을 모아요: 지시 파일과 `@` 가져오기, `.claude/rules/**/*.md`, 두 설정 파일의 글자와 `outputStyle`, 설정이 이름 부른 스크립트, 상태줄 yaml. `src/go_residue.rs` 의 `survey` 가 `go_same`/`go_changed` 후보에 적용해 `go_in_use` 와 `used_by`(경로와 종류만, 값 없음)를 붙이고, 못 읽은 곳이 있으면 모든 후보를 `keep` 으로 두며 `unread_readers` 에 파일과 까닭을 대요. 보여 주기 전용 목록(REQ-5)은 `keep` 이에요. `src/go_residue_apply.rs` 는 `--pick` 이 쓰는 중이거나 못 읽는 곳 때문에 남긴 파일이면 읽는 곳을 대고 거절해요. `go_residue_step` 의 이유 글에 두 수를 더하고 주석을 코드에 맞췄어요.

기존 테스트를 고친 곳: `tests/project_update_go_residue_cli.rs` 는 출력 스타일 파일이 쓰는 중이라 옮긴 수가 3 에서 2 가 되고, 한 테스트에는 진짜 세 번째 증명 파일을 만들어 줬어요(`tdd reopen` 한 번). `tests/update_end_to_end_cli.rs` 는 단계 순서에서 `go-residue` 가 빠지고(Risks 의 예상 그대로) 되돌리기 순서 테스트의 명령 수가 7 에서 6 이 돼요.

고른 것(SPEC 이 정하지 않은 곳): 출력 스타일은 대소문자 없이 파일 이름, frontmatter `name`, `dir/name` 끝 경로로 맞춰요. 폴더 안 규칙 파일의 `@` 가져오기는 그 파일 기준과 프로젝트 루트 기준을 모두 세요(과보존). 읽는 파일은 1000개, `.claude/rules` 목록은 1000개와 깊이 8 에서 끊고, 끊으면 못 읽은 곳이에요. 보여 주기 전용 목록은 1000개까지예요.

- AC-1 met: `a_go_file_still_read_is_in_use_and_names_its_reader` 가 통과했어요.
- AC-2 met: `a_reader_that_cannot_be_read_keeps_every_go_file` 이 통과했어요.
- AC-3 met: `an_in_use_file_is_never_moved_and_a_new_reference_makes_the_plan_stale` 가 통과했어요.
- AC-4 met: `the_survey_lists_local_rules_and_harness_files_and_moves_none` 이 통과했어요.
- AC-5 met: `the_step_is_optional_with_a_candidate_and_never_needs_decision` 이 통과했어요.
- AC-6 met: `the_mo_ai_kr_shape_moves_only_unread_proven_files` 가 통과했어요.
- AC-7 met: `moai spec show SPEC-20261125-go-residue-apply` 가 `revision: 2` 를 읽고, REQ-10 과 AC-1 이 REQ-6 과 같은 말을 해요(AC-1 은 출력 스타일이 제자리에 남는 것도 말해요). 리뷰어가 두 문장을 REQ-6 과 대조해 주세요(manual).
- AC-8 met: `update_end_to_end_cli` 9개가 통과했어요.
