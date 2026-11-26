---
id: SPEC-20261126-no-state-folder-writes
title: "A project with no earlier state folder and a Go, mixed, interrupted or refused .moai gets no state written by project-apply or a first launch"
status: draft
revision: 2
created: 2026-11-26
---

# A project with no earlier state folder and a Go, mixed, interrupted or refused .moai gets no state written by project-apply or a first launch

## Intent

옛 상태 폴더 `.moai-cli` 가 없고 resolver(`state_root::dir`)가 `.moai` 를 답하지 않는 프로젝트(지금은 판정 `go`·`mixed`·`interrupted`·`refuse`)에서 `project-apply` 와 `moai cc|glm` 의 첫 초기화가 `.moai-cli` 를 새로 만들어요. resolver 가 그 판정에서 `.moai-cli` 를 답하기 때문이에요(SPEC-20261121-path-root-moai-default REQ-4). 리더 결정 (a)(2026-11-26, 원장 노트 9717): resolver 와 그 SPEC 은 그대로 두고, 쓰는 쪽이 아무것도 만들지 않고 멈추며 다음 명령을 한 줄로 알려요.

## Non-goals

- resolver(`state_root::dir`)의 답과 승인된 SPEC-20261121-path-root-moai-default 는 바꾸지 않아요.
- `.moai-cli` 가 있는 프로젝트의 읽기·쓰기는 바꾸지 않아요.
- `project-apply` 와 첫 초기화 밖의 쓰기 명령(forge 증거, skill 되돌리기, 다른 영수증)은 이 SPEC 이 다루지 않아요.
- 파일을 옮기거나 지우지 않고, Go 파일은 읽기만 해요.

## Requirements

- REQ-1: When `project-apply` is called for a project that has no `.moai-cli` and for which the state resolver does not answer `.moai` (today a go, mixed, interrupted or refuse verdict), the command shall refuse before it reads the plan or writes anything, with one line that names the verdict, its reason and the next step. · failure: none — the refusal itself writes nothing, so there is no partial state to clean up
- REQ-2: When a first `cc` or `glm` launch finds such a project, the launcher shall start the host and print that one line (in `--dry-run` as well) without creating a git repository, an `.init-pending` marker, a profile or CLAUDE.md. · failure: none — the note is built from values already read; the launch goes on without any write
- REQ-3: When a project has `.moai-cli` in any form, or the state resolver answers `.moai` for it, the two writers shall behave as they did before this change. · failure: none — the new condition is false for those trees

## Acceptance

- AC-1 (REQ-1): 임시 프로젝트 네 가지 — interrupted(`.moai` 없이 `.moai-backup/manifest.json`), go(Go 표식 `.moai/`), mixed(Go 표식과 옛 이름 표식 `.moai/moai-cli.json`), refuse(`.moai` 가 링크) — 에서 `project-apply` 가 0 이 아닌 코드와 `invalid_input` 으로 끝나고, 메시지에 판정과 다음 걸음이 있으며, 프로젝트 트리 스냅샷(경로·링크 대상·바이트)이 전과 같아요 · verify: `cargo test --locked --offline --test project_apply_no_old_folder_cli apply_`
- AC-2 (REQ-2): 같은 프로젝트에서 stub host 를 둔 `moai cc` 가 host 를 시작하고, 알림 한 줄을 내며, 트리 스냅샷이 같고(`.init-pending` 은 git 공통 폴더에도 없음), git 저장소가 아닌 interrupted·go 폴더에는 `.git` 도 생기지 않아요; `--dry-run` 도 같은 한 줄이에요 · verify: `cargo test --locked --offline --test project_apply_no_old_folder_cli launch_`
- AC-3 (REQ-3): 대조 — 새 프로젝트는 `.moai/project.yaml` 을 만들고, `.moai-cli` 가 있는 Go 프로젝트는 지금처럼 `.moai-cli/project.yaml` 에 써요 · verify: `cargo test --locked --offline --test project_apply_no_old_folder_cli control_`
- AC-4 (REQ-1, REQ-2, REQ-3): Go 표식 옆에 유효한 `.moai/moai.json` 이 있는 프로젝트에서 두 쓰는 쪽은 resolver 의 답을 따라요 — `.moai` 를 답하면 멈추지 않고 `.moai/project.yaml` 에 쓰고, 답하지 않으면 멈추며, 어느 쪽이든 `.moai-cli` 는 생기지 않아요 · verify: `cargo test --locked --offline --test project_apply_no_old_folder_cli follow_`

## Scope

- write: `.moai/specs/SPEC-20261126-no-state-folder-writes/`, `src/launcher_init.rs`(판정 함수 하나와 첫 초기화의 멈춤), `src/ops/project_apply.rs`, `tests/project_apply_no_old_folder_cli.rs`(새)
- read-only: `src/state_root.rs`(다른 카드가 같은 파일을 고치는 중이라 쓰지 않아요), `src/profile.rs`, `src/project_update.rs`, `.moai/specs/SPEC-20261121-path-root-moai-default/`
- forbidden: `state_root::dir` 의 판정 변경, Go 파일 쓰기·이동·삭제, 설치된 HOME 접근

## Plan

1. 승인(위임) 뒤 red 시험: AC-1·AC-2 가 지금 코드에서 실패하는 것(`.moai-cli` 생성)을 먼저 확인해요.
2. `launcher_init` 에 "`.moai-cli` 가 없고 resolver 가 `.moai` 를 답하지 않는" 판정과 한 줄 안내를 만드는 공개 함수를 더해요(`state_root` 의 공개 함수만 읽어요).
3. `project-apply` 입구(`src/ops/project_apply.rs`)와 첫 초기화(`src/launcher_init.rs` 의 두 갈래, `git init`·`settings()` 앞)에서 그 함수를 불러요.
4. guard-mutant(멈춤 조건 제거 시 AC-1·AC-2 실패), `check affected`, READY.

## Risks

- tier 2(쓰기 경로): 멈춤이 너무 넓으면 정상 프로젝트의 apply·초기화가 막혀요. AC-3 대조와 `tests/launcher_auto_init.rs` 가 막아요.
- `.moai` 가 없는 interrupted 모양에는 `go-state-revert-plan` 이 받을 백업 번호(`.moai/backups/<UTC>`)가 없어요. 그래서 안내는 그 경우 사용자가 직접 정한다고 말해요.

## Decisions

- D-1: 멈춤 조건은 "`.moai-cli` 가 어떤 형태로도 없고 `state_root::dir` 이 `.moai` 가 아니다" 하나예요. 판정 목록을 조건에 두지 않아서 resolver 가 바뀌면(예: 되돌아온 Go 표식을 `.moai` 로 답하는 변경) 그대로 따라가요. 판정 이름은 안내 문장의 다음 걸음을 고르는 데만 써요.
- D-2: 첫 초기화는 `git init` 과 CLAUDE.md 까지 통째로 건너뛰어요(결정 9717 "아무것도 만들지 않고"). host 는 시작해요.
- D-3: 안내는 다른 CLI 거절·launcher 알림처럼 영어 한 줄이에요. 화면 번역은 더하지 않아요.
- D-4: `project-apply` 는 plan 입력을 읽기 전에 거절해요. 그래서 낡은 plan 이어도 같은 안내를 받아요.
