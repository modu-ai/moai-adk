---
id: SPEC-20261119-update-moves-reports
title: "The project update moves the earlier untracked report files to the user-level reports folder"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-reports-outside-repo
---

# The project update moves the earlier untracked report files to the user-level reports folder

## Intent

사용자 지시(2026-11-19, 카드 C-126 결정 노트 3273): `/moai update` 가 프로젝트의 옛 보고서 폴더(`.moai-cli/reports`, 그리고 조사에서 나오는 `.moai/reports`)에 남은 파일을 새 자리 `~/.moai/reports/<project>/` 로 옮겨요. 조건: update 의 기존 미리보기(plan) → 적용(apply) 흐름 안에서 하고 미리보기에 옮길 파일 수와 목적지를 보여 줘요; 추적되지 않는 파일만 옮기고 git 이 추적하는 파일은 "추적 중, 건드리지 않음" 으로 수만 보여 줘요; 목적지에 같은 이름이 있으면 덮어쓰지 않고 건너뛰며 수를 알려요; 두 번 돌려도 같은 결과(멱등)예요; 폴더 구조를 유지하고 목적지 권한은 본인만이에요; 시험은 임시 HOME + 임시 프로젝트에서 옮김·추적 파일 유지·이름 충돌·재실행 네 경우예요.

이 조각은 실제 HOME 에 쓰는 길이라 tier 2 감사예요. 목적지는 부모 SPEC 의 `home_workdir::reports_dir` 가 줘요(새 해석기를 만들지 않아요).

## Non-goals

- 추적 파일 옮기기·지우기: 삭제는 사용자 선택이에요(부모 SPEC 의 결정 노트).
- `.gitignore` 고치기, 옛 폴더 지우기: 옮긴 뒤 비는 폴더도 그대로 둬요.
- Go 세대 `~/.moai/reports` 의 옛 파일 옮기기.

## Requirements

- REQ-1: The command `reports-move-plan` shall list the files that git does not track under `.moai-cli/reports` and `.moai/reports` of the project with their destinations under the HOME reports folder.
- REQ-2: The command `reports-move-plan` shall count the files that git tracks and name them as tracked and left alone.
- REQ-3: The command `reports-move-apply` shall move each planned file to the same relative path under the HOME reports folder and keep the folder structure.
- REQ-4: The command `reports-move-apply` shall skip a file whose destination name already exists and count it as skipped without overwriting it.
- REQ-5: The command `reports-move-apply` shall give the same result when it runs a second time by moving nothing and counting the files already in place.
- REQ-6: The command `reports-move-apply` shall make the folders it creates readable by the owner only and clear the group and other permission bits of the files it moves.
- REQ-7: The command `reports-move-apply` shall refuse a stale plan digest or a missing permission reference and write nothing.
- REQ-8: The command `project-update-plan` shall show one step `reports` whose preview carries the counts to move and to skip and the tracked count and the destination.
- REQ-9: The step `reports` shall offer its apply command only with the plan digest of the same preview.
- REQ-10: The command `reports-move-plan` shall report the state `blocked` and plan no move when the project is not a git repository or git cannot tell what is tracked.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 HOME·임시 프로젝트에서 `reports-move-plan` 이 추적되지 않은 파일 수와 목적지와 "추적 중, 건드리지 않음" 수를 보여 줘요.
  verify: `cargo test --locked --offline --test reports_move_cli plan_counts_untracked_and_tracked_files`
- AC-2 (REQ-3): 적용하면 추적되지 않은 파일이 폴더 구조 그대로 새 자리로 옮겨지고 추적 파일은 제자리에 남아요.
  verify: `cargo test --locked --offline --test reports_move_cli apply_moves_untracked_and_keeps_tracked`
- AC-3 (REQ-4): 목적지에 같은 이름이 있으면 덮어쓰지 않고 건너뛰며 그 수를 알려요(목적지 바이트가 그대로예요).
  verify: `cargo test --locked --offline --test reports_move_cli apply_skips_a_name_that_exists`
- AC-4 (REQ-5): 두 번째 적용은 옮김 0·이미 있음을 세고 어떤 파일도 바꾸지 않아요.
  verify: `cargo test --locked --offline --test reports_move_cli apply_twice_gives_the_same_result`
- AC-5 (REQ-6): 새로 만든 폴더는 0700 이고 옮긴 파일에서 그룹·다른 사람 권한 비트가 지워져요(Unix).
  verify: `cargo test --locked --offline --test reports_move_cli apply_makes_the_destination_owner_only`
- AC-6 (REQ-7): 어긋난 digest 와 빠진 permission reference 는 거절되고 아무 파일도 옮겨지지 않아요.
  verify: `cargo test --locked --offline --test reports_move_cli apply_refuses_a_stale_digest_or_missing_permission`
- AC-7 (REQ-8, REQ-9): `project-update-plan` 의 `reports` 단계가 수와 목적지를 미리 보여 주고 apply 줄은 같은 미리보기의 digest 를 써요.
  verify: `cargo test --locked --offline --test reports_move_cli update_plan_shows_the_reports_step`
- AC-8 (REQ-10): git 저장소가 아닌 프로젝트에서는 `blocked` 이고 아무것도 옮기지 않아요.
  verify: `cargo test --locked --offline --test reports_move_cli plan_is_blocked_without_a_git_repository`

## Scope

- write: `.moai/specs/SPEC-20261119-update-moves-reports/spec.md`, `src/reports_move.rs`(새), `src/lib.rs`(모듈 한 줄), `src/ops/reports_move_plan.rs`·`src/ops/reports_move_apply.rs`(새), `docs/design-review/appendix/operations/reports_move_plan.md`·`reports_move_apply.md`(새), `src/help.rs`, `tests/fixtures/help/*.txt`, `src/project_update.rs`, `skills/moai/workflows/update.md`, `skills/moai/references/source-manifest.json`, `tests/reports_move_cli.rs`(새)
- read-only: `src/home_workdir.rs`(부모 SPEC 이 더한 `reports_dir`), `src/state_root.rs`, `src/plan_input` 관련
- forbidden: 실제 HOME 에 쓰는 시험, 네트워크, 추적 파일 이동·삭제

## Plan

1. 부모 SPEC 의 `reports_dir` 가 있는 가지에서 시작해요(없으면 먼저 부모 조각을 받아요).
2. RED: `tests/reports_move_cli.rs` 의 네 경우(옮김·추적 유지·이름 충돌·재실행)와 권한·digest·단계·git 아님 시험을 먼저 써서 실패를 눈으로 봐요. 새 시험이 git 을 띄우면 `tests/common/git.rs` 도우미로만 해요(환경을 씻어요, 리더 공지).
3. GREEN: `src/reports_move.rs`(plan/apply), 두 연산 파일과 문서 조각, 도움말 표·예시·골든.
4. `project_update.rs` 의 `reports` 단계와 `update.md` 지침 한 줄, 매니페스트 해시.
5. 시험을 돌려요: 위 verify 명령들, fmt, ops rustfmt, clippy `--all-targets`, `project_update` 를 읽는 기존 시험, `--bins`.

## Risks

- 옮김은 실제 HOME 에 쓰는 길이에요: 목적지 이름이 있으면 건너뛰고(덮어쓰기 없음) 추적 파일은 건드리지 않으며 시험은 전부 임시 HOME 에서만 돌려요.
- `project-update-plan` 에 단계를 더하면 단계 목록·골든을 읽는 시험이 달라질 수 있어요: 먼저 그 시험들을 찾아 같이 고쳐요.
- 새 연산 두 개는 도움말 표·예시·문서 조각·골든 네 곳을 함께 건드려요(`operations_doc_matches_code`). 기준 가지에서 이미 실패하는 `--bins` 시험 5개와 섞이지 않게 전후 실패 목록을 따로 비교해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 옮김 대상은 프로젝트의 `.moai-cli/reports` 와 `.moai/reports` 안의 추적되지 않는 파일(무시 규칙이 걸린 파일 포함)이에요. `plan-audit/.gitkeep` 처럼 추적되는 파일은 자동으로 제자리에 남아요 | 사용자 지시와 `.gitignore` 의 `.moai/reports/` 블록 | 대상 폴더 줄이기 |
| D-2 | git 저장소가 아니거나 추적 여부를 알 수 없으면 `blocked` 로 두고 아무것도 옮기지 않는다 | 추적 파일은 그대로라는 지시를 어길 위험이 있는 쪽은 옮기지 않는 쪽이 안전해요 | 저장소 아님을 "모두 추적 안 됨"으로 |
| D-3 | `reports` 단계는 install 단계(phase)에 두고 다른 단계를 기다리게 하지도 기다리지도 않는다 | 옮김은 HOME 에만 쓰고 프로젝트 파일을 바꾸지 않아서 다른 단계의 미리보기를 낡게 하지 않아요 | 다른 단계 뒤로 순서 바꾸기 |
| D-4 | 같은 파일 시스템이면 이름 바꾸기(rename)로 옮기고, 아니면 복사한 뒤 바이트가 같을 때만 원본을 지운다 | 같은 볼륨이면 원자적이고, 아니면 원본을 확인 뒤에만 지워요 | 항상 복사 |
| D-5 | 심볼릭 링크와 폴더가 아닌 것은 옮기지 않고 수만 알린다 | 링크를 따라가면 저장소 밖 파일을 옮길 수 있어요 | 링크 처리 규칙 더하기 |

## Result

(구현 뒤에 채워요.)
