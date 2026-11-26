---
id: SPEC-20261126-go-local-settings-reference
title: "A Go-generation project held by its local settings reference gets one pick-gated way to move on"
status: draft
revision: 3
created: 2026-11-26
---

# A Go-generation project held by its local settings reference gets one pick-gated way to move on

## Intent

Go 세대 프로젝트의 `.claude/settings.local.json` 이 `.moai/` 를 가리키면 Go 폴더가 옮겨지지 않고, `go-state-plan` 과 `project-update-plan` 이 "직접 고치라" 고만 해요. 그런데 `project.md` 2단계는 세션이 그 설정을 고치지 말라고 해요. SPEC-20261126-no-state-folder-writes(아직 착지 전)가 들어오면 `project-apply` 도 멈추므로 프로필을 만들 길이 없어져요(감사 노트 9916). 이 SPEC 은 사용자의 선택 하나로 넘어가는 길을 만들어요: 원본을 백업하고, 로컬 설정의 프로젝트 `.moai/` 참조를 Go 백업 폴더 이름으로 바꿔 쓴 뒤 폴더를 옮겨요. 사용자가 설정해 둔 것은 하나도 없어지지 않아요.

base `41c16e60ac` 의 동작을 일회용 프로젝트(임시 HOME, `env -i`)에서 읽은 것이에요. 쓴 바이너리는 `6bb4f8a031` 의 빌드이고 `src/go_state.rs`·`src/ops/project_apply.rs`·`src/project_update.rs`·`src/go_settings_clean.rs` 가 base 와 같아요:

- `go-state-plan`: `layout` `folder_stays`, `.moai` 는 `leave`("settings reference .moai/").
- `project-update-plan`: `go-state` `blocked`("edit or remove these keys yourself"), `profile` `needs_decision`.
- `go-state-apply`: `ok`, `unchanged`.
- base 의 `project-apply` 는 `.moai-cli/project.yaml` 을 만들어요. 완전한 막다른 길은 위 SPEC 의 멈춤이 더해질 때 생겨요.
- 새로 찾은 것: 로컬 설정에 HOME 경로(`<HOME>/.moai/bin` 을 담은 `env.PATH`)만 있어도 폴더가 묶여요. Go 도구는 세션 `PATH` 를 이 파일의 `env.PATH` 에 써요(읽기 전용 참고 `moai-adk-go` `internal/template/settings.go`). 이 SPEC 은 그 판정을 바꾸지 않아요(D-5).

## Interview

- none: 사용자에게 직접 묻지 않았어요. 선택은 위임에 따른 리더 결정(원장 노트 9939, 10074)과 위임 승인 노트에 있고, HOME 경로 판정(D-5)만 사용자에게 물을 일로 남았어요

## Non-goals

- HOME 의 `.moai` 를 가리키는 값이 폴더를 묶는 판정: 지금 그대로 두고 `docs/later.md` 한 줄로 남겨요(D-5).

- 공유 `.claude/settings.json` 의 사용자 값이 폴더를 묶는 경우(`go-settings` 단계의 `kept`).
- worktree, 중첩 체크아웃처럼 설정이 아닌 이유로 폴더가 남는 경우.
- 프로필 생성 자체(`project-plan`, `project-apply`)와 launcher 의 멈춤 줄.
- Go 폴더가 남은 채 프로필을 만드는 길(D-2).

## Requirements

- REQ-1: When the local settings file names a path under the project's `.moai/` and so holds the Go folder, the `go-state-plan` preview shall list each such value with the file, its JSON pointer, the `.moai/…` path text alone and its class (a project reference, which is a relative path, a path after the Claude project directory variable in any quoting, or a path after the absolute project root; or any other form, which the pick cannot rewrite), name the pick in the preview and in the leave reason of the `.moai` item, and print no other part of a value. · failure: none — the preview writes nothing; a file it cannot read as a JSON object holds the folder as today and says why
- REQ-2: When a settings value names the `.moai` folder of the user's HOME (HOME written as a tilde, as the HOME variable with or without braces, or as its absolute path) and the project root is not HOME itself, the `go-state-plan` preview shall hold the Go folder as it does today and class that value as one the pick cannot rewrite. · failure: none — nothing is written for such a value
- REQ-3: When `go-state-apply` is given `--rewrite-local-references` with the digest of a preview made with the same pick, the apply shall put the original bytes of the local settings file in the backup, replace with the Go backup folder's name the `.moai` of each occurrence classed as a project reference and no other text, keep every `permissions.deny` and `permissions.ask` entry as it is and add its rewritten twin, write the file atomically with a mode no wider than it had, and move the folder as the preview showed. · failure: REQ-8
- REQ-4: When one reference is not a project reference, or the file is a link, too large, not a JSON object or changed since the plan, or something other than the local settings also holds the folder, the apply with the pick shall refuse and write nothing, naming each value the user has to edit. · failure: none — refused before the first write
- REQ-5: Without the pick the preview and the apply shall leave the local settings file and the folder as they do today. · failure: none — the new write path is not entered
- REQ-6: When `go-state-revert` reverts such an apply, the revert shall restore the original bytes of the local settings file, moving a file that changed after the apply into the replaced-by-revert folder of the backup area first. · failure: none — a stopped revert keeps the current file at its path or in that folder, deletes nothing and can be run again
- REQ-7: The `project.md` step 2 text and the `go-state` reason of `project-update-plan` shall name this pick as the next step, say that a session never edits the file by hand, and promise no `.moai-cli/backups/` folder when nothing moves. · failure: none — text only
- REQ-8 (REQ-3): IF a step of the apply with the pick fails after the first write, THEN the apply shall put the original bytes back into the local settings file, leave the folder or move it back by the existing receipt rules, and name the step. · failure: none — the original bytes stay in the backup and the receipt names what was left
- REQ-9 (REQ-7): Where the local settings file holds the Go folder, the `go-state` step of `project-update-plan` shall print a `go-state-apply` command only in the picked form, which means: if the preview made with the pick moves the folder, the step is that preview, needs a decision, carries the command with `--rewrite-local-references` and that preview's digest, and waits behind the other backup steps as today; if it does not, the step carries no command, keeps the state it has today, and gives as its reason each value the user has to edit or why the folder stays. · failure: none — the plan writes nothing, and the apply checks a printed command again before it writes

## Acceptance

- AC-1 (REQ-1): 임시 프로젝트의 미리보기가 참조마다 파일·pointer·경로 글자를 내고, 같은 값에 있는 토큰 꼴 글자는 어느 출력에도 없으며, 트리가 전후 같아요 · verify: `cargo test --locked --offline --test go_local_reference_cli plan_`
- AC-2 (REQ-2): HOME 경로 참조(틸드, HOME 변수, 절대 경로 꼴)만 있는 프로젝트는 지금처럼 `folder_stays` 이고 미리보기가 바꿀 수 없는 값으로 이름 부르며, 선택을 준 apply 는 거절하고 트리가 바이트까지 같아요 · verify: `cargo test --locked --offline --test go_local_reference_cli home_`
- AC-3 (REQ-3): 프로젝트 참조의 꼴(상대 경로, 프로젝트 폴더 변수의 중괄호·따옴표 꼴, 프로젝트 루트의 절대 경로)을 담은 statusLine·hooks·`permissions.deny` 픽스처에서 선택 하나(`--rewrite-local-references`)로 폴더가 옮겨지고, 원본 바이트가 백업에 있고, 바뀐 참조의 파일이 그 경로에 실제로 있으며, 원래 deny 항목이 남고, 파일 모드가 넓어지지 않아요. 이어서 `project-plan`·`project-apply` 가 `.moai/project.yaml` 을 만들고 `.moai-cli` 는 없어요 · verify: `cargo test --locked --offline --test go_local_reference_cli apply_`
- AC-4 (REQ-4, REQ-5, REQ-8): 알 수 없는 꼴·링크·바뀐 파일·다른 이유의 묶음에서는 거절하고 트리가 바이트까지 같아요. 선택이 없으면 지금과 같아요. 첫 쓰기 뒤에 이동이 실패하게 만든 픽스처에서는 로컬 설정이 원본 바이트로 돌아오고 답이 단계를 말해요 · verify: `cargo test --locked --offline --test go_local_reference_cli refuse_`
- AC-5 (REQ-6): revert 뒤 로컬 설정이 원본과 바이트까지 같고, apply 뒤 고친 파일은 replaced-by-revert 에 남아요 · verify: `cargo test --locked --offline --test go_local_reference_cli revert_`
- AC-6 (REQ-7): `project-update-plan` 의 `go-state` 이유가 선택을 이름 부르고, `project.md` 2단계가 같은 길을 말해요 · verify: `cargo test --locked --offline --test go_local_reference_cli next_step_`
- AC-7 (REQ-9): 다른 옮길 파일이 있든 없든 `project-update-plan` 의 `go-state` 단계는 선택을 넣은 명령만 내고, 찍힌 그대로 실행하면 폴더가 옮겨지며 `.moai-cli` 가 생기지 않아요. HOME 경로 참조가 있으면 명령이 없고, 상태는 지금과 같으며, 이유가 그 pointer 를 이름 불러요 · verify: `cargo test --locked --offline --test go_local_reference_cli update_`

## Scope

- write: `.moai/specs/SPEC-20261126-go-local-settings-reference/`, `src/go_state.rs`, `src/main.rs`(go-state 두 명령의 플래그), `src/project_update.rs`(D-7), `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`(line-lease), `docs/design-review/appendix/` 의 go-state 설명(있으면), 도움말 골든(바뀌면, line-lease), `tests/go_state_cli.rs`·`tests/go_settings_cli.rs`·`tests/project_update_plan_cli.rs`·`tests/update_end_to_end_cli.rs`(바뀐 이유 문장과 단계를 고정한 단언), `tests/go_local_reference_cli.rs`(새; 다른 `go_*_cli` 처럼 단독 대상이고, 첫 낱말 `go` 는 `tests/areas.toml` 에 이미 있어요. 상태 폴더 이름은 resolver 상수로 써요), `docs/later.md`(D-5 한 줄)
- read-only: `src/go_settings_clean.rs`, `src/state_root.rs`, `src/ops/project_apply.rs`, `src/launcher_init.rs`, `skills/moai/workflows/update.md`
- forbidden: 설치된 HOME 과 실제 사용자 설정, `moai-adk-go` 쓰기, `state_root::dir` 판정 변경, 선택 없는 사용자 설정 쓰기

## Plan

1. 승인 뒤 red 테스트(AC-1·AC-3).
2. 참조 분류(프로젝트 참조 / 그 밖)를 `go_state` 에 더해요. 폴더를 묶는 지금의 `contains(".moai/")` 판정은 그대로 두고, 분류는 미리보기와 선택에만 써요.
3. 선택 플래그를 plan digest 에 넣어요. apply 의 순서: 파일을 다시 읽어 digest 확인 → 폴더 이동(새 `.moai` 와 백업 자리가 생겨요) → 원본 사본(기존 root-files 사본) → 바뀐 파일을 임시 파일 → 검증 → 교체(모드 유지) → 원본 digest 를 영수증에. 쓰기에서 실패하면 이동을 되돌려요(REQ-8).
4. revert 의 "옆으로 치우기" 는 지금 공유 설정만 받아요(`only the shared settings are set aside; stopped`). 로컬 설정까지 넓혀요(REQ-6).
5. 이유 문장과 `project.md` 를 고치고, update 길의 `go-state` 단계가 선택을 넣은 미리보기를 쓰게 해요(AC-7, D-7 의 예약이 풀린 뒤).
6. guard-mutant, 일회용 프로젝트의 위험 경로 실행(`env -i`: 바이트, 모드, 백업, revert), `check affected`, READY.

## Risks

- tier 2: 토큰이 들어 있을 수 있는 사용자 파일을 써요 → 선택 필수, digest 결박, 원본 백업(소유자 전용, git 제외는 기존 backups 규칙), 임시 파일 → 검증 → 교체, 값 비출력.
- HOME 의 `.moai` 를 담은 값(Go 도구가 베낀 `env.PATH` 포함)이 있으면 선택으로도 풀리지 않아요 → 미리보기가 그 pointer 를 이름 부르고(AC-2), 판정을 좁힐지는 사용자에게 물어요(D-5).
- 바꿔 쓴 hook 은 백업 폴더의 스크립트를 실행해요. 사용자가 그 폴더를 지우면 참조가 끊겨요 → 미리보기와 `project.md` 가 말해요.
- 다시 직렬화하면 키 순서·들여쓰기가 달라질 수 있어요 → 원본 바이트는 백업에 있고 revert 가 그대로 돌려요.
- update 길에서 옮길 파일이 없던 갈래의 선택은 설치·merge 단계를 막지 않아요(D-8, 지금과 같아요). 선택을 거절해도 그 단계들은 그대로 나와요.
- 옮길 파일이 있는 갈래에서 선택이 폴더를 옮기지 못하면(HOME 경로 값, worktree 등) 명령이 없는 `needs_decision` 이라 설치·merge 단계가 계속 기다려요. 지금은 선택 없는 apply 를 하면 풀렸어요. 사용자가 이유에 적힌 것을 고쳐야 풀려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 참조는 지우지 않고 Go 백업 폴더 이름으로 바꿔 써요. `permissions.deny`·`ask` 는 원본을 남기고 바뀐 사본을 더해요 | 지우면 사용자의 hook·deny 규칙·status line 이 사라져 보호가 약해져요. 설정을 그대로 두고 옮기면 hook 과 status line 이 조용히 끊겨요. 바꿔 쓰기만 지금 동작을 지켜요. 값: 살아 있는 설정이 백업이라는 이름의 폴더에서 Go 스크립트를 실행해서 그 폴더를 치울 수 없고, Go hook 이 이행 뒤에도 계속 돌아요. 카드는 "remove" 를 예로 들었으므로 승인에서 고를 항목이에요. 다른 안: (나) `go_settings_clean` 의 정확한 값 판정을 로컬 파일에도 써서 Go 가 쓰던 키만 지우고 나머지는 사용자가 직접(값이 작고 백업 폴더 의존이 없지만 사용자 값이 남으면 여전히 막혀요), (다) 설정 그대로 두고 폴더만 옮기기 | 플래그와 쓰기 경로를 지워요 |
| D-2 | Go 폴더가 남은 채 프로필을 만드는 길은 택하지 않아요 | `.moai/` 에 쓰면 판정이 `mixed` 가 되고 update 는 그 판정을 "이 경로 밖" 으로 막아요(`src/project_update.rs` `tied`, 읽기). `.moai-cli` 에 쓰는 것은 결정 9717(옛 폴더를 새로 만들지 않기)과 반대예요 | — |
| D-3 | 선택이 있을 때에 한해 앞선 두 규칙을 바꿔요: `go_settings_clean` 의 "settings.local.json 은 쓰지 않음"(#2936 L2a D1 조건 3)은 go-settings 에는 그대로 두고 go-state 의 이 선택에만 예외, `project.md` 2단계의 "설정을 고쳐 넘어가지 말 것" 은 "손으로 고치지 말고 이 선택을 보일 것" 으로. O3(사용자 결정 6877929684: 로컬 deny 규칙이 계속 동작)은 원본 유지 + 사본 추가로 지켜요 | 작업 카드의 본문과 리더 지시(사용자의 선택과 백업) | D-1 과 함께 되돌려요 |
| D-4 | 백업은 이동이 만드는 `.moai/backups/<UTC>/root-files/` 의 기존 사본이에요. 새 명령이나 `.moai-cli/backups/` 를 만들지 않아요 | 이동과 한 번에 일어나서 백업 자리가 늘 있어요 | — |
| D-5 | HOME 의 `.moai` 를 가리키는 값도 지금처럼 폴더를 묶어요. 선택은 그 값을 바꾸지 않고 거절하며 이름 불러요. 판정을 좁히는 일은 `docs/later.md` 한 줄로 남겨요 | 리더 결정(원장 노트 9939): fail-closed 판정을 좁히는 일은 위임으로 정하지 않고 사용자에게 물어요. 알려진 거짓 묶음: Rust launcher 자체가 HOME 의 `.moai/path/moai-cli` 에 있고(`go_settings_clean::launcher_shown`) Go 도구는 상속한 `PATH` 를 `env.PATH` 에 베껴요(HOME `PATH` 값만으로 묶이는 것은 위 재현에서 읽음) | 사용자가 고르면 새 revision |
| D-6 | 기준 1 의 "one user pick" 은 이 막힘을 푸는 선택 하나(바꿔 쓰고 옮기기)로 읽어요. 그 뒤 프로필 계획의 기존 승인은 그대로예요 | `project.md` 4단계의 기존 승인 | — |
| D-7 | `src/project_update.rs` 는 지금 다른 작업(update 이행의 go-residue 후속)이 예약해 두었어요. `go_state` 의 `.moai` 항목 이유는 `go-state` 단계 이유에 그대로 실리므로 선택 이름은 이 파일 없이도 보여요. 그 파일의 "edit or remove these keys yourself" 문장과 같은 `contains(".moai/")` 판정(`moai_dir_references`, 풀리기 전에는 "held by" 문장이 go-state 와 다를 수 있어요)은 예약이 풀린 뒤 이 작업의 마지막 조각으로 고쳐요(리더 결정, 원장 노트 9939). 나머지 코드는 NeXT 에서 바로 시작할 수 있어요(SPEC-20261126-no-state-folder-writes 의 착지를 기다리지 않음: 테스트는 `.moai/project.yaml` 이 생기고 `.moai-cli` 가 없는 것을 봐요) | 원장의 예약을 읽음(2026-11-26) | — |
| D-8 | 로컬 설정이 폴더를 묶으면 update 길의 `go-state` 단계는 선택 없는 apply 명령을 내지 않아요. 선택을 넣은 미리보기가 폴더를 옮길 때만 그 명령을 내고(단계별 yes 가 그 선택), 아니면 명령 없이 상태를 그대로 둬요. 옮길 것이 없던 갈래에서 선택만 낸 단계 뒤에는 설치·merge 단계가 기다리지 않아요 | 리더 결정(원장 노트 10074 = 질문 10073 의 (가), 노트 10109 = 질문 10107 의 (가): 새 막는 기다림은 관찰부터라는 사용자 규칙). 선택 없는 apply 는 사용자에게 두 번 묻고, 옮길 파일이 있으면 백업 자리가 옛 상태 폴더라서(`src/go_state.rs` `layout_and_backup`, 코드에서 읽음) 그 폴더를 새로 만들 수 있어요(결정 9717 과 반대) | 그 갈래를 지워요 |

## Result
