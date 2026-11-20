---
id: SPEC-20261120-build-folder-cap
title: "A size cap and least-recently-used cleanup for build folders, and a measured warm start"
status: draft
revision: 1
created: 2026-11-20
---

# A size cap and least-recently-used cleanup for build folders, and a measured warm start

## Intent

레인과 감사가 worktree 마다 자기 빌드 폴더(`~/.moai/build/<project>-<worktree>`)에 Rust 빌드 결과를 쌓고, 끝난 worktree 의 폴더도 남아요. 이 기계에서 164개 폴더가 약 195GiB 이고, `moai optimize` 의 미리보기는 "worktree 가 없어졌거나 병합된" 폴더만 정리 후보로 보여서 상한이 없어요. 이 SPEC 은 (1) 빌드 폴더 전체 크기에 설정할 수 있는 상한을 두고(기본 60GiB), 넘으면 오래 안 쓴 폴더부터(끝난 worktree 의 것 먼저) 정리 후보로 미리보기에 올려요. 지우는 일은 지금처럼 미리보기 → 사용자가 고름 → 적용이고, 상한이 새 적용 경로를 만들지는 않아요. (2) worktree 마다 빌드 폴더 하나를 유지하되(폴더를 공유하면 cargo 의 잠금 때문에 빌드가 한 줄로 서요), 새 worktree 의 첫 빌드를 다른 worktree 의 빌드 폴더를 APFS 복제(`cp -c`)로 시작하면 의존성 산출물을 다시 쓰는지 실측해요. 경로가 바뀌면 cargo 가 다시 빌드할 수 있어서, 켜는 일은 실측에서 재사용이 보일 때만 후속 조각으로 해요.

## Non-goals

- 빌드 폴더를 worktree 끼리 공유: 하지 않아요(잠금으로 빌드가 직렬이 돼요).
- 상한을 넘으면 자동으로 지우기: 하지 않아요. 미리보기에 올리기까지예요.
- 복제 시작을 지금 켜기: 하지 않아요. 실측 결과(Result)를 보고 후속 조각에서 정해요.
- 프로필 파일에 상한 키 두기: 프로필 파일의 범위가 열린 뒤에 이어요. 이번에는 플래그와 홈 설정 파일 한 곳이에요.

## Requirements

- REQ-1: When `optimize-plan` runs, the preview shall report the total size of the build folders under the home build root, the cap in GiB with its source and whether the total is over the cap. · failure: none
- REQ-2 (REQ-1): IF the cap is zero, negative or not a whole number, THEN the preview shall refuse and name the setting.
- REQ-3: When the total is over the cap, the preview shall mark build folders as cap candidates in least-recently-used order, folders of finished worktrees first, until the marked folders bring the total down to the cap. · failure: none
- REQ-4 (REQ-3): IF a build folder belongs to a live worktree and changed less than a day ago, or a process works in it, or it holds a file that is not cargo output, or its owner record is missing or names another repository, THEN the preview shall keep it with every reason.
- REQ-5: While the total is at or under the cap, the preview shall mark no build folder because of the cap. · failure: none
- REQ-6: When picked build folders are applied, the apply shall remove only the cargo output of those folders, each checked against a fresh preview made with the same cap. · failure: none
- REQ-7 (REQ-6): IF a picked build folder is not a candidate of the fresh preview, THEN the apply shall remove nothing at that folder.
- REQ-8: The Result of this SPEC shall record, for a build folder cloned from another worktree's build folder and for an empty one, the time, the CPU time, the units reused and rebuilt and the change in free disk space. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-5): 임시 HOME 의 빌드 폴더 셋에 상한을 크게 주면 미리보기가 합계 바이트·상한·출처 `flag`·초과 `false` 를 말하고 상한 때문에 표시된 폴더가 없어요 · verify: `cargo test --locked --offline --test build_folder_cap the_preview_reports_the_total_against_the_cap`
- AC-2 (REQ-2): 상한 0, 음수, 숫자가 아닌 값은 미리보기가 거절하고 설정 이름(`--build-cap-gib`)을 말해요 · verify: `cargo test --locked --offline --test build_folder_cap a_cap_that_is_not_a_positive_whole_number_is_refused`
- AC-3 (REQ-3): 합계가 상한을 넘으면 끝난 worktree 의 폴더가 먼저, 그다음 오래 안 쓴 순서로 후보가 되고 후보를 더하면 합계가 상한 이하가 되는 데서 멈춰요 · verify: `cargo test --locked --offline --test build_folder_cap over_the_cap_finished_folders_go_first_then_the_least_recently_used`
- AC-4 (REQ-4): 하루가 안 된 살아 있는 worktree 의 폴더, 프로세스가 쓰는 폴더, cargo 산출물이 아닌 파일이 든 폴더, 주인 기록이 없거나 다른 저장소인 폴더는 이유와 함께 유지로 나와요 · verify: `cargo test --locked --offline --test build_folder_cap unsafe_folders_are_kept_with_every_reason`
- AC-5 (REQ-6, REQ-7): 고른 폴더의 cargo 산출물만 지워지고 후보가 아닌 이름은 아무것도 지우지 않아요 · verify: `python3 -I -m unittest tests.build_folder_test`
- AC-6 (REQ-8): 복제한 폴더와 빈 폴더의 빌드 측정(시간·CPU 시간·Fresh/Dirty 수·`df` 변화)과 켤지 여부가 이 SPEC 의 Result 에 있어요 · verify: manual: Result 표를 읽어요(명령과 출력이 Result 에 적혀 있어요)

## Scope

- write: 이 SPEC, `src/build_cap.rs`(새), `src/optimize.rs`, `src/ops/optimize_plan.rs`, `src/lib.rs`(모듈 한 줄), `tools/build_folder.py`, `tests/build_folder_cap.rs`(새), `tests/build_folder_test.py`, `docs/design-review/appendix/operations/optimize_plan.md`
- read-only: `src/worktree_gc.rs`, `skills/moai/references/internal/build-folder.md`
- forbidden: 자동 삭제, 빌드 폴더 공유, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것

## Plan

1. 조각 1(이번): SPEC, 조사 보고, PLAN 노트, 복제 측정(AC-6).
2. 조각 2(상한과 후보): `src/build_cap.rs` 순수 판정(상한 읽기·정렬·선택)과 `optimize-plan` 의 `--build-cap-gib`, 미리보기의 `build_cap` 객체, RED 시험(AC-1~4).
3. 조각 3(적용): `tools/build_folder.py remove` 가 미리보기의 상한을 새 미리보기에도 넘기도록(AC-5).
4. 조각 4: 측정 결과로 복제 시작을 켤지 정해 후속 SPEC 으로.

## Risks

- 복제본은 `du` 가 다시 세요(실측): 크기 합계는 실제 디스크 사용이 아니라 할당 바이트 합이에요. 회수한 공간은 `df` 변화로만 알 수 있다고 미리보기가 말해요.
- 살아 있는 worktree 의 폴더를 상한 때문에 지우면 다음 빌드가 처음부터예요. 하루 안에 바뀐 폴더와 쓰는 중인 폴더는 지키고, 지우는 것은 사용자가 고른 것뿐이에요.
- `src/optimize.rs` 와 `src/ops/optimize_plan.rs` 는 worktree 개수 상한 조각도 고쳐요. 서로 다른 줄이지만 통합 때 텍스트 병합이 필요할 수 있어요(Risks 로 미리 적어 둬요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기본 상한은 60GiB, 플래그 `--build-cap-gib`, 홈 설정 `hook-config.toml` 의 `[build] cap_gib`(플래그가 우선) | 설계 보고서의 예(60G)와 worktree 개수 상한이 쓰는 같은 홈 설정 파일 | 값 바꾸기 |
| D-2 | 합계는 이 저장소 것이 아닌 폴더까지 포함한 `~/.moai/build` 전체 크기이고, 후보는 이 저장소의 주인 기록이 있는 폴더만이에요 | 상한은 디스크 한 곳의 한도이고, 다른 저장소 폴더는 판정할 근거가 없어요 | 저장소별 상한 |
| D-3 | 정렬은 끝난 worktree(`worktree_gone`·`branch_merged`)가 먼저, 같은 무리 안에서는 마지막 변경이 오래된 순이에요 | 끝난 것을 지우는 위험이 가장 낮아요 | 정렬 바꾸기 |
| D-4 | 복제 시작은 실측 후에만 켜요. 켜는 기준: 복제한 폴더로 빌드했을 때 작업 공간 밖 의존성 단위의 90% 이상이 `Fresh` 이고 CPU 시간이 빈 폴더의 절반 이하 | 경로가 바뀌면 cargo 가 다시 빌드할 수 있다는 점이 미확인이에요 | 기준 바꾸기 |

## Result

(아직 구현 전이에요.)
