---
id: SPEC-20261126-claude-md-fresh-merge
title: "Go-generation update: new managed CLAUDE.md, then merge the project lines the user picks"
status: draft
revision: 2
created: 2026-11-26
---

# Go-generation update: new managed CLAUDE.md, then merge the project lines the user picks

## Intent

Go 세대에서 올라온 프로젝트는 업데이트 뒤에도 옛 지침이 모든 세션에 실려요. 사용자의 방향(2026-11-26, 원장 노트 10297)은 "과거 내용을 말고 신규로 설정 후 병합"이에요. 이 SPEC 은 `project-update-plan` 에 `claude-md` 단계를 만들어요. 사용자가 "바꾸기"를 고르면 옛 지침 파일은 지우지 않고 백업으로 옮기고, 새 `CLAUDE.md` 를 프로필에서 새로 만들고, 옛 글 가운데 이 프로젝트만의 줄을 미리보기로 보여 사용자가 고른 곳에만 합쳐요. 새 `CLAUDE.md` 는 `@AGENTS.md` 를 싣지 않아요.

실제 프로젝트(mo.ai.kr)를 2026-11-26 에 읽기만 해서 확인한 것이에요. 아무것도 쓰지 않았어요.

- `CLAUDE.md`(15,746 바이트)는 Go 매니페스트(`.moai-backup/manifest.json`)의 `deployed_hash` 와 바이트까지 같아요. 사용자 줄이 없는 Go 템플릿이고, `@AGENTS.md`, `@AGENTS.local.md`, `.moai/config/sections/` 의 yaml 두 개를 실어요.
- `AGENTS.md`(80,608 바이트)는 기록과 달라요. 그 경로의 git 기록 가운데 커밋 `9938a58e`(2026-10-24)의 blob `4618e1ee` 의 sha256 이 `deployed_hash` 와 같아요. 그래서 원본을 읽기만으로 되찾아 사용자 줄만 가릴 수 있어요.
- `AGENTS.local.md`(54,748 바이트)와 `CLAUDE.local.md`(54,875 바이트)는 매니페스트에 없고 git 이 무시하며, 줄 차이 도구가 낸 다른 줄이 2개뿐이에요. Claude Code 는 `CLAUDE.local.md` 를 스스로 읽고 `CLAUDE.md` 는 `AGENTS.local.md` 를 또 실으니, 거의 같은 글이 세션마다 두 번 실려요.
- `.moai/docs/local-rules/` 는 git 이 추적하고, `go-residue` 조사는 `docs` 폴더를 건너뛰어요.

코드에서 읽은 것(base `74b27c0723`): `src/agents_local_merge/picks.rs:170` 은 `CLAUDE.md` 가 `@AGENTS.md` 를 실으면 "이미 끝났다"로 답해요. `skills/moai/workflows/update.md:40` 은 `claude-md` 단계를 말하지만 `src/` 에는 그 단계가 없어요.

## Interview

세 질문의 답은 사용자가 리더 대화에서 직접 고른 것이에요(2026-11-26, 원장 노트 10325). 따옴표 안은 고른 선택지의 글 그대로예요.

- Q1: "바꾸기"를 고르면 Go 증명이 없는 파일(`CLAUDE.local.md`, `AGENTS.local.md`, `.moai/docs/local-rules/`)도 백업으로 옮겨도 되나요? 원래 작업 기준은 "바이트로 증명된 Go 파일만 후보이고 사용자 데이터는 그대로"였어요. → A: "'바꾸기' 고를 때만 옮김 (Recommended)". 미리보기를 보고 "바꾸기"를 고른 때에만 옮기고, "그대로"는 어떤 지침 파일도 바꾸지 않아요(D-3).
- Q2: `local-rules` 파일 가운데 일부를 제자리에 남기는 선택이 필요한가요? → A: "필요 없음 (Recommended)". "바꾸기"는 전부 옮기고, 필요한 줄은 골라 합치거나 되돌리기로 전부 되돌려요(D-5).
- Q3: 아무 선택도 하지 않은 프로젝트 줄은 어디에 두나요? → A: "백업에만 둠 (Recommended)". 합칠 곳은 줄 묶음마다 사용자가 골라요(D-6).

## Non-goals

- `.claude/CLAUDE.md`, `.claude/rules` 와 그 밖의 `.claude/` 파일은 옮기지 않아요. `.claude/CLAUDE.md` 가 있으면 미리보기가 경고만 해요.
- `go-state` 가 이미 백업의 `root-files/` 로 옮긴 지침 파일에서 줄을 되찾아 합치는 일. 이 단계는 프로젝트 루트에 있는 파일만 읽어요.
- 생성 매니페스트에 요청이 없어 생성 단계가 `needs_decision` 에 머무는 문제 자체를 고치는 일(별도 작업이 맡아요). 이 단계는 그 상태와 무관하게 계산돼요.
- `go-residue` 가 "아직 쓰는 파일"을 가리는 판정(별도 작업).
- `.moai/config/sections/` 의 yaml 을 옮기는 일.
- 고른 줄을 요약하거나 고쳐 쓰는 일. 줄은 글자 그대로 옮겨요.

## Requirements

- REQ-1: `project-update-plan` shall list a `claude-md` step, computed whatever state the generation step reads, that reads `needs_decision` and offers `replace` (recommended) and `leave` while the root `CLAUDE.md` is a Go-generation instruction file as D-2 defines it, `skip` once the user declined it with the existing `project-update-decline` record (the only file `leave` writes), and `unchanged` otherwise. · failure: none — the preview writes nothing; while the `go-settings` step still has a command or the `go-state` step still has a move to make (also while that move waits behind the `go-settings` or `reports` apply) the step reads `blocked`, no other command of a backup step blocks it, and while the profile is missing or the first-launch placeholder it reads `needs_decision` with no apply command and says why
- REQ-2 (REQ-1): While the root `CLAUDE.md` is a Go-generation instruction file as D-2 defines it, every other update writer shall keep away from it: in `project-update-plan` the `agents-picks` and `agents-local-merge` steps read `skip` and the generation step offers no `append-managed`, and the direct `agents-merge-plan` and `agents-merge-apply` commands offer and apply no pick and no merge, in each case naming the `claude-md` step and never counting an `@AGENTS.md` line of that file as a finished merge. · failure: none — a refused direct apply writes nothing
- REQ-3: The `claude-md-plan` preview shall list, in this order, `CLAUDE.md`, `AGENTS.md`, `AGENTS.local.md`, `CLAUDE.local.md` and every file under `.moai/docs/local-rules/` that exists, each with its sha256, line count, git tracking and the comparison used, then print for the `--keep` picks it was given one plan digest and one apply command, and write nothing. · failure: a source that is a link, lies below a link folder, is not UTF-8 text or is larger than 1 MiB is named and no apply command is printed
- REQ-4: The preview shall count as the project lines of a source no line when its bytes equal the sha256 a found Go manifest records for that path, only the lines absent from the version of that path in git history whose sha256 equals that record when such a version is found, and every line otherwise, in each case leaving out the lines of moai's `moai-workflow` block, the Go curator marker lines and a personal file that is exactly a first-creation text of moai. · failure: none — no git repository, a failed git read or no matching version within the search limit counts every line, and the preview names the comparison it used
- REQ-5: The preview shall group the project lines into hunks of one heading section each with the id `<path>:<first line>`, show a hunk equal to an earlier one as "same as <id>" without its text, replace each line that holds a credential-shaped value with a count, mark a hunk that names a path this step moves, and warn when a resulting file would exceed 200 lines. · failure: none — preview only
- REQ-6: When `claude-md-apply` replaces, the updater shall write a new `CLAUDE.md` made of the managed block rendered from the profile (its language items, else the request language of the recorded generation request, else `en`) followed by the hunks picked with `--keep <id>=shared` in source order, leaving out every line that imports a file this step moves and every Go curator marker line. · failure: none — a refusal is REQ-9 and a failed write is REQ-10
- REQ-7: When at least one hunk is picked with `--keep <id>=personal`, the updater shall write `CLAUDE.local.md` made of moai's first-creation personal text followed by those hunks, with the same lines left out, and add an ignore rule in the same apply when no rule ignores the file. · failure: none — without a personal pick no `CLAUDE.local.md` is written; a refusal is REQ-9 and a failed write is REQ-10
- REQ-8: The apply shall save a pending receipt, move every listed source into an owner-only folder under `.moai/backups/` at the same relative path, then write the new files, delete nothing, and answer with the receipt id and the notice that settings apply at once while running sessions and lanes read changed instructions only after a restart. · failure: REQ-10
- REQ-9: IF the profile is missing or the placeholder, a source is refused by REQ-3, a source or the profile changed since the preview, a `--keep` id is unknown, a hunk picked `shared` holds a credential-shaped value, or git would not ignore the backup folder, THEN `claude-md-apply` shall name the reason and write nothing. · failure: none — refused before the first write
- REQ-10 (REQ-8): IF a move or a write fails after another one succeeded, THEN the apply shall restore the earlier files from the same receipt and name every file it could not restore. · failure: the receipt reads `partial`, the originals stay in the backup and `claude-md-revert` undoes the rest
- REQ-11: When `claude-md-revert` reverts that receipt, the updater shall put every moved file back with its bytes and mode, removing a file this step wrote only while its bytes equal what was written and keeping any other file in place, named as a conflict. · failure: none — the original of a conflict stays in the backup and the revert can run again
- REQ-12: The `update.md` workflow shall read the `claude-md` step as the commands print it: ask `replace` or `leave`, show the hunks and warnings, ask the destination of each hunk the user wants to keep, run `claude-md-plan` with those picks before the printed apply, record `leave` with the decline command, say that running sessions need a restart, and list the moved tracked paths in the closing commit offer. · failure: none — text only

## Acceptance

모든 테스트는 임시 HOME 과 mo.ai.kr 모양의 임시 git 프로젝트에서 실제 바이너리로 돌려요. 실제 프로젝트에는 쓰지 않아요.

- AC-1 (REQ-1, REQ-2): 요청 없는 옛 생성 매니페스트 때문에 생성 단계가 `needs_decision` 인 픽스처에서도 `claude-md` 단계가 `needs_decision` 으로 `replace` 와 `leave` 를 내고, `agents-picks`·`agents-local-merge` 는 `skip` 이에요. 생성 단계가 파일을 계획하는 픽스처에는 `append-managed` 가 없어요. 직접 실행한 `agents-merge-plan` 은 선택을 내지 않고 `agents-merge-apply --pick move` 는 `claude-md` 단계를 이름 부르며 거절하고 트리가 바이트까지 같아요. 관리 블록이 덧붙은 Go `CLAUDE.md` 에도 단계가 나와요. Go 증명이 없는 `CLAUDE.md` 에는 단계가 `unchanged` 이고 다른 단계는 지금과 같아요. 백업 단계가 남으면 `blocked`, 프로필이 자리표시면 적용 명령 없는 `needs_decision` 이에요 · verify: `cargo test --locked --offline --test project_update_claude_md_cli step_`
- AC-2 (REQ-3, REQ-4): 출처가 정한 순서로 나오고, 기록과 바이트가 같은 `CLAUDE.md` 는 프로젝트 줄이 0개예요. Go 원본을 커밋한 뒤 고쳐 커밋한 `AGENTS.md` 는 고친 줄만, 맞는 blob 이 없으면 전부 나와요. 미리보기 전후 트리가 바이트까지 같아요 · verify: `cargo test --locked --offline --test project_update_claude_md_cli base_`
- AC-3 (REQ-5): 두 로컬 파일의 같은 묶음은 한 번만 글이 나오고, 비밀 모양 줄은 개수만 나오며 그 값은 어느 출력에도 없고, 옮기는 경로를 말하는 묶음에 표시가 있고, 200줄을 넘는 결과에 경고가 있어요 · verify: `cargo test --locked --offline --test project_update_claude_md_cli hunk_`
- AC-4 (REQ-6, REQ-7, REQ-8): 찍힌 명령을 그대로 실행하면 새 `CLAUDE.md` 는 관리 블록과 고른 묶음뿐이고 `@AGENTS.md`·`@AGENTS.local.md`·큐레이터 표지가 없어요. `CLAUDE.local.md` 는 personal 선택이 있을 때만 생기고 git 이 무시해요. 옮긴 파일은 백업에 바이트까지 같고 git 이 무시해요. 답에 재시작 안내가 있어요. 다시 미리보면 요청 없는 옛 매니페스트 픽스처를 포함해 `claude-md` 는 `unchanged` 이고, 요청이 기록돼 생성 단계가 파일을 계획하는 픽스처(옛 `CLAUDE.md` 기준선이 있는 것 포함)에서는 생성 계획의 `CLAUDE.md` 줄도 `unchanged` 예요. `undo_order` 에 이 단계가 있어요 · verify: `cargo test --locked --offline --test project_update_claude_md_cli apply_`
- AC-5 (REQ-9): 자리표시 프로필, 링크 출처, 링크 폴더, UTF-8 아닌 출처, 1 MiB 초과, 미리보기 뒤 바뀐 파일, 모르는 id, 비밀 줄이 있는 묶음의 shared 선택, 무시되지 않는 백업 자리는 각각 이유를 말하고 트리가 바이트까지 같아요 · verify: `cargo test --locked --offline --test project_update_claude_md_cli refuse_`
- AC-6 (REQ-10, REQ-11): 첫 이동 뒤의 실패를 만든 픽스처에서 파일이 모두 돌아와요. revert 는 바이트와 모드를 되살리고, apply 뒤 고친 새 파일은 그대로 두며 이름을 말해요 · verify: `cargo test --locked --offline --test project_update_claude_md_cli receipt_`
- AC-7 (REQ-1): 찍힌 decline 명령은 `.moai/manifests/update-declined.json` 만 쓰고, 단계는 `skip` 이 되며, 출처 하나가 바뀌면 다시 물어요 · verify: `cargo test --locked --offline --test project_update_claude_md_cli leave_`
- AC-8 (REQ-12): `update.md` 의 글이 단계 읽는 법, 두 선택, 목적지, 재시작, 마무리 커밋을 말하고 스킬 검사가 통과해요 · verify: `cargo test --locked --offline --test update_closing_commit`
  verify: `cargo test --locked --offline --test next_flow_skill_scan`
  verify: `cargo test --locked --offline --test source_manifest_facts`
- AC-9: 바꾼 파일의 이웃 검사
  verify: `cargo test --locked --offline --test claude_md_picks_cli`
  verify: `cargo test --locked --offline --test project_update_go_residue_cli`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261126-claude-md-fresh-merge/`, `src/project_update/`(새 `claude_md.rs`, `declined.rs`), `src/project_update.rs`, `src/agents_local_merge/picks.rs`, `src/agents_local_merge.rs`(직접 명령의 비켜서기), `src/agents.rs`(블록 렌더를 이 단계가 쓰게), `src/go_residue.rs`(Go 매니페스트 찾기를 이 단계가 쓰게), `src/main.rs`, `src/ops/`(`claude-md-plan`·`claude-md-apply`·`claude-md-revert`), `src/lib.rs`·`src/help.rs`·도움말 골든·`tools/repo_scans.json`·`skills/moai/references/source-manifest.json`(line-lease), `docs/design-review/appendix/interfaces.md`, `docs/design-review/appendix/operations/claude_md_*.md`, 이 단계로 글이 틀려지는 `docs/design-review/appendix/operations/{agents_merge_plan,agents_merge_apply,project_update_decline,project_update_plan}.md`·`docs/rust-product-path.md`·`src/ops/project_update_decline.rs` 의 문서 주석, `skills/moai/workflows/update.md`, `tests/project_update_claude_md_cli.rs`(새), `tests/update_closing_commit.rs`, 바뀐 단계 목록을 고정한 단언이 있는 테스트와 `tests/fixtures/update_tux/` 의 미리보기 골든
- read-only: `src/go_state.rs`, `src/go_residue_apply.rs`, `src/generation.rs`, `src/claude.rs`
- forbidden: 설치된 HOME 과 실제 사용자 설정, mo.ai.kr 을 포함한 실제 프로젝트 쓰기, `moai-adk-go` 쓰기, 사용자 파일 삭제, `replace` 선택 없는 지침 파일 쓰기

## Plan

1. 승인 뒤 mo.ai.kr 모양의 픽스처(Go 원본 커밋 → 수정 커밋, 거의 같은 두 로컬 파일, `local-rules`, 백업 속 Go 매니페스트, 요청 없는 생성 매니페스트)와 red 테스트.
2. 판정과 미리보기(`claude-md-plan`): D-2 판정, 출처, 원본 찾기, 묶음.
3. 적용·영수증·되돌리기. `go-residue-apply` 의 영수증 먼저, 이동 나중 순서를 따라요.
4. `project-update-plan` 단계, 다른 단계의 비켜서기, decline, `undo_order`, `update.md`.
5. guard-mutant, 일회용 프로젝트의 위험 경로 실행(임시 HOME 의 Rust 테스트), `check affected`, READY.

## Risks

- tier 2: 사용자 글을 옮기고 새로 써요 → 선택 필수, 계획 digest 결박, 옮기기만, 영수증 먼저, 되돌리기.
- 되돌리기 전에 새 `CLAUDE.md` 를 고치면 그 경로는 충돌로 남아요. 원본은 백업에 있어요.
- `src/project_update.rs` 는 이 SPEC 을 쓸 때 지시 파일 이름 검사 작업의 예약에 들어 있었어요(2026-11-26 읽음). 먼저 READY 를 받은 쪽이 먼저 착지하고 뒤쪽이 합쳐요.
- `local-rules` 를 가리키는 다른 파일(스킬, 설정, CI)이 있으면 옮긴 뒤 죽은 경로가 돼요. 미리보기는 합칠 묶음만 표시하고 저장소 전체는 찾지 않아요. 추적 파일이라 되돌리기와 `git revert` 로 되살려요.
- 원본을 git 기록에서 못 찾으면 사용자 줄이 파일 전체가 돼요. 조용히 버리는 것보다 많이 보여 주는 쪽을 골랐어요(D-4).
- `go-state` 가 루트 지침 파일을 이미 옮긴 프로젝트에는 이 단계가 나오지 않아요(Non-goals). 파일은 그 백업에 있어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 `CLAUDE.md` 는 프로필에서 렌더한 관리 블록과 고른 줄뿐이고 `@AGENTS.md` 를 싣지 않아요. 원래 작업 기준의 "관리 CLAUDE.md(`@AGENTS.md`)"를 대신해요. | 사용자 방향(노트 10297), `SPEC-20261122-claude-md-layout` 개정 3의 사용자 결정(`CLAUDE.md`·`CLAUDE.local.md` 만 써요). | 덧붙이는 방식으로 돌아가요. |
| D-2 | Go 세대 `CLAUDE.md` 는 (a) Go 큐레이터 표지가 있거나, (b) moai 관리 블록이 없고 찾은 Go 매니페스트가 `CLAUDE.md` 를 기록한 일반 파일이에요. 적용한 새 파일은 블록이 있고 표지가 없어서 다시 걸리지 않아요. | 블록이 덧붙은 옛 파일도 옛 글을 계속 싣고, 백업 속 매니페스트는 적용 뒤에도 남아요. | 판정을 매니페스트 해시 일치로 좁혀요. |
| D-3 | `CLAUDE.local.md`, `AGENTS.local.md`, `local-rules` 는 Go 증명이 없어도 `replace` 를 고르면 함께 옮겨요. 선택 없이는 옮기지 않아요. | 사용자 방향과 리더 지시. `go-state` 의 "`CLAUDE.local.md` 는 사용자 것이라 남긴다"와 `go-residue` 의 "사용자 데이터는 남긴다"는 그 두 명령에서는 그대로예요. | 출처 목록에서 빼요. |
| D-4 | 프로젝트 줄의 기준 글은 순서대로: 기록과 같은 현재 바이트(줄 0개) → 그 경로의 git 기록에서 sha256 이 같은 버전 → 없음(전체). 검색은 읽기만 하고 최근 200개 버전까지예요. | 매니페스트는 해시만 기록해요. mo.ai.kr 의 `AGENTS.md` 에서 원본을 찾는 것을 확인했어요. | 항상 전체를 보여 줘요. |
| D-5 | `local-rules` 는 파일을 제자리에 남기는 선택 없이 전부 옮겨요. | 리더 지시 "백업으로 옮겨요"와 작은 범위. | `--stay <path>` 선택을 더해요. |
| D-6 | 고르지 않은 줄은 새 파일에 넣지 않아요. 목적지는 `shared`(`CLAUDE.md`)와 `personal`(`CLAUDE.local.md`) 둘이에요. | 새로 설정하고 고른 것만 합치는 방향. 흐름 규칙이나 `.claude/rules` 는 이 단계가 쓰지 않아요. | 목적지를 더해요. |
| D-7 | 두 옛 요구를 D-2 의 Go 세대 `CLAUDE.md` 에 대해서만 좁혀요. `SPEC-20261122-claude-md-project-picks` REQ-10("`@AGENTS.md` 를 이미 실으면 `move` 없음", `picks.rs` 의 `settled()`)은 그 파일의 import 를 끝난 것으로 세지 않아요. `SPEC-20261124-claude-md-product-next` REQ-2·REQ-3 의 선택 목록(`move`·`leave`·`remove-originals`)은 `claude-md` 단계가 그 파일을 맡는 동안 나오지 않고, 직접 명령도 선택 없이 그 단계를 이름 불러요. 다른 `CLAUDE.md` 에는 두 SPEC 이 그대로예요. | Go 세대 파일의 import 는 바꿀 대상이에요(노트 10309). `move` 가 열리면 옛 `CLAUDE.md` 에 `AGENTS.md` 글(mo.ai.kr 에서 약 80KB)이 덧붙어, 새로 설정하는 방향과 반대가 돼요. | 판정을 되돌려요. |
| D-8 | 비밀 모양 값이 있는 묶음은 추적 파일(`CLAUDE.md`)에 합치지 못해요. `personal` 이나 백업만 가능해요. | 비밀은 저장소에 들어가지 않아요. | 선택을 허용해요. |
| D-9 | `leave` 는 기존 decline 기록을 써요. 출처 파일 하나의 바이트가 바뀌면 다시 물어요. | 선택 단계가 이미 쓰는 방식이에요. | 기록 없이 매번 물어요. |
