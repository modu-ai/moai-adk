---
id: SPEC-20261121-audit-fixed-place
title: "An audit reuses one fixed audit worktree and its build folder instead of a fresh clone, so a nearby head rebuilds only what changed"
status: draft
revision: 1
created: 2026-11-21
---

# An audit reuses one fixed audit worktree and its build folder instead of a fresh clone, so a nearby head rebuilds only what changed

## Intent

사용자 지시(2026-11-21): 무거운 테스트 로직을 전수 조사해 고쳐요(보고서 `test-load-redesign-2026-11-21.md` S7). 독립 감사는 매번 새 경로에서 받은 저장소를 자기 빌드 폴더에서 빌드해요. 이미 있는 `tools/build_folder.py seed` 는 의존성 85개를 재사용하게 해 주지만, 작업공간 crate(`moai-cli`)는 빌드 기록에 소스 경로가 들어 있어서 경로가 다르면 **늘 다시 빌드**돼요(lib 한 번이 약 85 CPU-초, 테스트 crate 도 전부).

경로가 같으면 그 문제가 없어요. 감사자마다 고정 경로의 감사 워크트리와 고정 빌드 폴더를 두고, 감사할 때마다 그 워크트리를 새 head 로 **옮겨요**(지우고 다시 받지 않아요). git 은 달라진 파일만 다시 쓰므로 안 바뀐 파일의 수정 시각이 그대로라서, 두 번째 감사는 바뀐 crate 만 다시 컴파일해요.

이 SPEC 은 도구 한 가지만 해요: `tools/build_folder.py audit-place --name <감사 자리 이름> --head <40자 SHA>`. 감사 에이전트 지침·감사 워크플로·`moai audit-clone-plan` 명령 목록에 이 도구를 쓰라고 적는 일은 그 글을 쥔 변경과 겹치지 않을 때 다음 조각으로 해요(Non-goals).

## Non-goals

- 지침 글(`agents/audit.md`, `skills/moai/workflows/audit.md`, `skills/moai/references/internal/build-folder.md`)과 `moai audit-clone-plan` 의 명령 목록: 다음 조각이에요. 이 조각은 도구와 그 측정까지예요.
- 감사 자리를 여러 감사가 동시에 쓰는 일: 한 자리는 한 번에 한 감사예요(빌드 잠금이 잡혀 있으면 거절해요). 감사자가 둘이면 자리 이름도 둘이에요.
- `CARGO_HOME` 을 이 도구가 정하는 일: 레인과 감사자의 `CARGO_HOME` 은 호스트 설정이에요. 도구는 쓰인 값을 기록하고 바뀌면 알려 줄 뿐이에요.
- 빌드 산출물을 지우는 일: 그건 `clean` 의 일이고, 이 도구는 폴더도 산출물도 지우지 않아요.

## Requirements

- REQ-1: When `build_folder.py audit-place --name <n> --head <sha>` runs and the place `<HOME>/.moai/worktrees/<project>/audit-<n>` does not exist, the tool shall create it as a detached worktree of this repository at `<sha>`, make its build folder with the owner record as `env` does, and print `created: true`. · failure: none
- REQ-2: When the place exists and is a worktree of this repository, the tool shall move it to `<sha>` without removing its folder, so that files which did not change keep their modification times, and print `created: false` with the previous head and the new head. · failure: none
- REQ-3: When the tool has created or moved the place, it shall print the tree hash of `<sha>` and check that the tree of the place's HEAD equals it and that `git status --porcelain` lists nothing, else refuse with exit 1. · failure: none
- REQ-4: While the place holds changes to tracked files or untracked files that git does not ignore, the tool shall reset the tracked files and remove those untracked files, nested repositories included, print how many it took back, and leave the files git ignores alone. · failure: none
- REQ-5: IF a cargo run holds the cargo lock of the place's build folder, THEN the tool shall refuse with exit 1 and the reason `place_busy` and change nothing.
- REQ-6: IF `<sha>` is not 40 lowercase hexadecimal digits or names no commit of this repository, or `<n>` is not one safe folder name, THEN the tool shall refuse with exit 1 and write nothing.
- REQ-7: IF the place path exists but is a link, is not a folder, is not a worktree of this repository, has its git common folder in another repository, has a branch checked out, or has no owner record of this tool in its build folder, THEN the tool shall refuse with exit 1 and change nothing.
- REQ-8: When the tool runs, it shall record `CARGO_HOME` (the environment value, else `<HOME>/.cargo`) in the build folder's `.moai-audit-place.json`. · failure: none
- REQ-11: IF a run sees a `CARGO_HOME` other than the recorded one, THEN the tool shall print `cargo_home_changed: true` with the earlier value as `cargo_home_was`.
- REQ-12: IF the new head, or an edited or untracked `.gitignore` in the place, changes the ignore rules while the place holds files that git ignores, THEN the tool shall refuse with exit 1 and the reason `ignore_rules_changed` and change nothing.
- REQ-13: IF a step after the tool's own `worktree add` made the place refuses, THEN the tool shall remove that place and never a place it did not make.
- REQ-14: When the place's folder is gone but git still lists it as a worktree, the tool shall take over that one registration with `worktree add --force` and leave every other registration alone. · failure: none
- REQ-15: IF the build folder of the place holds a record of another worktree or a link is on its path, THEN the tool shall refuse with exit 1 before it makes or moves anything.
- REQ-9: The tool shall print one JSON line that names the `place`, the build `folder`, the `head`, the `tree`, `created`, the `previous` head, the number of files that differ between the previous and the new head (`moved_files`), `taken_back` and `cargo_home`. · failure: none
- REQ-10: The tool shall change nothing in the checkout it runs in and delete neither the place nor the cargo output of its build folder. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-9): a first call makes the place and its folder, and the answer names them with the tree hash of the head.
  verify: `python3 -m unittest tests.build_folder_test.AuditPlaceTest`
- AC-2 (REQ-2, REQ-3, REQ-4): a second call at another head moves the same folder, keeps the modification time of a file that did not change, and takes back a changed tracked file and a stray untracked file but not an ignored one.
  verify: `python3 -m unittest tests.build_folder_test.AuditPlaceTest`
- AC-3 (REQ-5, REQ-6, REQ-7, REQ-8, REQ-10): the refusals, the cargo lock, the `CARGO_HOME` record.
  verify: `python3 -m unittest tests.build_folder_test`
- AC-4 (REQ-2): in the fixed place a second audit at a nearby head recompiles only the crates whose files changed, and the place is reused, not recloned; the counts are in Result.
  manual: run `cargo test --no-run` in the place at one head, move it to a head that changes one test file, run it again and count the `Compiling` lines; compare with a fresh worktree at another path.

## Scope

- write: `.moai/specs/SPEC-20261121-audit-fixed-place/**`, `tools/build_folder.py`, `tests/build_folder_test.py`
- read-only: `tools/build_folder.py` `seed` and `env` (reused), `src/audit_clone.rs`, `agents/audit.md`, `skills/moai/workflows/audit.md`
- forbidden: any write in the real HOME by a test, a recursive delete of the place or of a build folder, a card, Issue or pull request number in the text this change adds

## Plan

1. SPEC, then RED tests (`AuditPlaceTest` in `tests/build_folder_test.py`).
2. `tools/build_folder.py`: `audit-place` (worktree add or detached move, tree-hash check, cleanup of untracked files, cargo-lock probe, `CARGO_HOME` record).
3. Measure on a real build in a slot: Compiling lines in the first and the second audit at a nearby head, in the fixed place and in a fresh path.
4. Checks, mutants, a read-only audit by another agent. Guidance text follows as the next piece.

## Risks

- A move that rewrites files git thinks changed gives them a new modification time and so rebuilds their crate; that is the point. A move that leaves a stale file with an old modification time would not rebuild a changed file, so the tool checks the tree hash and a clean status after every move, never trusting the move alone.
- `git clean` removes untracked files that git does not ignore. The place is the auditor's own worktree and evidence goes to ignored folders (build-folder rule 6); a stray file there would otherwise become a test crate of the audited tree.
- The place is shared by successive audits of one auditor: two audits at once would move the tree under a build. The cargo-lock probe stops the move while a build holds the lock; it cannot see a build that has not taken the lock yet, so the name stays one per auditor.
- Reuse of the dependencies needs the same `CARGO_HOME` path as the earlier run (`seed` note); the tool records the value and reports a change, it does not choose one.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A move (`git checkout --detach --force`) instead of delete and clone | Only changed files get a new modification time, which is what lets cargo skip the rest | Remove the place and call again |
| D-2 | Refuse (exit 1) when the place is busy, not an answer like `seed` | An audit that cannot have its tree must not go on silently on another head | Make it an answer with exit 0 |
| D-3 | The place lives under `<HOME>/.moai/worktrees/<project>/` | The project's rule for new worktrees; `optimize` and `worktree-report` already read that folder | Change one path function |

## Result

`python3 -I tools/build_folder.py audit-place --name <n> --head <40자 SHA>` 가 생겼어요. 첫 호출은 `<HOME>/.moai/worktrees/<project>/audit-<n>` 을 그 head 의 분리된 워크트리로 만들고 같은 이름의 빌드 폴더(`<project>-audit-<n>`)와 소유 기록을 만들어요. 다음 호출은 같은 워크트리를 새 head 로 옮겨요(`git checkout --detach --force`, 지우고 다시 받지 않아요). 옮긴 뒤에는 트리 해시가 head 의 것과 같고 `git status` 가 비었는지 확인하고, 편집된 추적 파일과 git 이 무시하지 않는 추적 안 된 파일은 되돌리거나 지우며(몇 개인지 `taken_back` 에 적어요), 무시되는 파일(증거 폴더 `target/` 같은 것)은 그대로 둬요. cargo 가 빌드 폴더의 잠금을 쥐고 있으면 `place_busy` 로 거절해요. 쓰인 `CARGO_HOME` 을 빌드 폴더의 `.moai-audit-place.json` 에 적고, 다음 호출에서 값이 다르면 `cargo_home_changed` 와 이전 값을 알려요.

측정(2026-11-21, 이 Mac, APFS, `CARGO_HOME=/tmp/moai-rc-cargo-agent-1`, 대상 `--lib --test card_inbox_cli --test card_show_notes_cli` 의 `cargo test --no-run`). 머리 A = 지역 통합 가지 끝 `b4681b3a6f`, 머리 B = A 위에 시험 파일 `tests/card_inbox_cli.rs` 에 주석 한 줄만 더한 가지 없는 임시 커밋 `9b996deb7b`. 감사 자리 `audit-measure278` 의 빌드 폴더는 첫 호출 뒤 레인 폴더에서 `seed` 로 시작했어요(`seeded: true`, `clonefile`).
- 첫 감사(자리에서 A): 37.57초(벽시계), user 117.44 + sys 5.09 = 122.5 CPU초. 작업공간 crate 가 처음 빌드돼요(경로가 처음이라).
- 옮김 A→B(`moved_files: 1`) 뒤 둘째 감사: 0.68초, user 0.57 + sys 0.25 = 0.82 CPU초.
- 같은 B 를 한 번 더(`moved_files: 0`): 0.12초, `Compiling` 줄 0개.
- 대조(새 경로 `measure-fresh` 의 B, 빌드 폴더는 자리의 폴더에서 `seed`): 43.71초, user 123.92 + sys 4.80 = 128.7 CPU초.
- `cargo test -v` 로 rustc 실행을 세었어요(`measure-place2.sh`): 자리를 A→B→A→B 로 옮길 때마다 rustc 는 정확히 1번(`--crate-name card_inbox_cli`), 나머지 85개 단위는 `Fresh` 였어요. `Compiling` 줄은 패키지 하나(`moai-cli`)만 세어서 crate 수를 가르지 못해요.
- 바뀐 것: 둘째 감사는 새 경로보다 벽시계로 약 64배(43.71 / 0.68), CPU 로 약 157배(128.7 / 0.82) 덜 들었어요. 이 숫자는 바뀐 파일이 시험 파일 하나일 때예요. `src/` 파일이 바뀌면 lib 와 그것에 기대는 모든 시험 crate 가 다시 빌드되어 새 경로와 같아져요(재지 않았어요).
- 재지 않은 것: 기계 부하를 맞춘 반복(각 1회만 쟀어요), Linux, 같은 head 의 감사가 아니라 `src/` 가 바뀐 두 head 사이의 비용, 전체 시험 바이너리(`--no-run` 대상은 위 3개).

- AC-1: met — `python3 -m unittest tests.build_folder_test.AuditPlaceTest` 24개 통과(첫 호출이 자리와 폴더와 기록을 만들고 트리 해시를 답해요).
- AC-2: met — 같은 시험 묶음: 옮긴 뒤 바뀌지 않은 파일의 수정 시각과 inode 가 그대로이고, 고친 추적 파일과 추적 안 된 파일은 되돌려지고(`taken_back` 2), 무시된 파일은 남아요.
- AC-3: met — `python3 -m unittest tests.build_folder_test` 66개 통과(둘째 수리 뒤. 처음 판은 53개, 첫 수리 뒤는 62개였어요. 잘못된 head·이름, 링크·파일·워크트리가 아닌 폴더·클론·다른 워크트리를 가리키는 링크 거절, 잡힌 cargo 잠금 `place_busy`, `CARGO_HOME` 기록과 변경 알림, 호출한 체크아웃과 cargo 산출물이 그대로임). 변이 16개 중 14개를 시험이 잡았어요(`mut-place.py`); 살아남은 P9(40자 형식 검사 제거)와 P17(옮긴 뒤 HEAD·트리 재확인 제거)은 같은 일을 하는 다른 검사(`rev-parse` 결과가 입력과 같은지, 옮긴 뒤 `git status` 가 비었는지)가 이미 있어서 겉으로 보이는 동작이 같아요(등가).
- AC-4: met — 위 측정: 둘째 감사 rustc 1번(바뀐 시험 crate 하나), 자리는 지우고 다시 받지 않고 재사용.

감사 수리(독립 감사가 head `400035542e` 에서 PASS-WITH-FINDINGS, 막는 결함 없음, 비차단 N1~N7): 지적마다 수리 전·후의 재현은 같은 시험을 옛 도구(`400035542e` 의 `tools/build_folder.py`)와 수리한 도구에 돌린 결과예요 — 새 시험 9개가 모두 옛 도구에서 실패(감사자 재실행: 8 FAIL, 1 ERROR)하고 수리한 도구에서 통과해요(`python3 -m unittest tests.build_folder_test.AuditPlaceTest`, 하위 `.gitignore` 시험 1개는 그 뒤에 더했어요).
- N1(무시 규칙이 바뀌면 증거 삭제): `ignore_rules_changed` 로 거절하고 아무것도 바꾸지 않아요(REQ-12). 시험 `test_a_head_that_changes_a_gitignore_is_refused_while_ignored_files_are_there`, `test_a_changed_gitignore_in_a_subfolder_counts_too`. 무시된 파일이 없으면 같은 이동은 허용돼요.
- N2(다른 저장소를 가리키는 `.git` 파일): 자리의 git 공통 폴더가 이 저장소의 것인지 확인해 다르면 거절해요(REQ-7). 시험 `test_a_place_whose_git_file_points_at_another_repository_is_refused`.
- N3(거절한 뒤 자리가 남음): 빌드 폴더의 다른 워크트리 기록·링크는 만들기 전에 보고(REQ-15), 만든 뒤의 단계가 거절하면 이 실행이 만든 자리를 지워요(REQ-13). 시험 `test_a_refusal_after_the_worktree_was_made_takes_it_back`, `test_a_linked_build_root_refuses_before_a_place_is_made`.
- N4(추적 안 된 중첩 저장소로 막힘): `git clean -ffd` 로 중첩 저장소도 되돌려요(REQ-4). 시험 `test_an_untracked_nested_repository_is_taken_back_too`.
- N5(브랜치가 있거나 이 도구가 만들지 않은 자리): 분리된 HEAD 이고 빌드 폴더의 `.moai-audit-place.json` 에 이 자리가 적혀 있을 때만 옮겨요(REQ-7). 시험 `test_only_a_detached_place_made_by_audit_place_is_moved`, `test_the_owner_marker_names_the_place`.
- N6(손으로 지운 자리의 등록이 남음): `git worktree add --force` 로 그 자리의 등록 하나만 넘겨받고 다시 만들어요(REQ-14). 시험 `test_a_place_folder_deleted_by_hand_is_made_again`. (처음 수리는 `git worktree prune` 이었는데 감사 N9 가 다른 워크트리의 등록까지 잊는다고 짚어 바꿨어요.)
- N7(증거 파일 없음, 큰 소리 없는 한계): 측정 명령은 이 Result 에 적었고(`cargo test -v --offline --locked --no-run --lib --test card_inbox_cli --test card_show_notes_cli`, 자리를 A→B→A→B 로 옮기며 rustc 실행 수와 `Fresh` 줄 수를 셈), 변이 목록은 아래에 적었어요. 잠금은 cargo 가 컴파일하는 동안만 잡혀서, 잠금이 풀린 뒤 시험이 자리의 파일을 읽는 동안의 이동은 못 막아요(Risks 에 있음).
변이 24개(수리한 코드 기준: 이동 `--force`, `clean -ff`, `clean` 없음, 되돌린 수, 이동 없음, 잠금 무시, 등록 확인, 링크 확인, 이름 앞 `-`, 소유 기록, `CARGO_HOME` 알림, `moved_files`, `created`, `chdir`, `status`, 무시 규칙 변경 거절(최상위·하위), 공통 폴더, 만든 자리 되돌리기, 브랜치, 소유 기록 요구, `prune`, 빌드 폴더 기록 먼저 보기, 무시된 파일 감지) 중 21개를 시험이 잡았고, 살아남은 셋은 같은 일을 하는 다른 검사가 막아서 겉으로 보이는 동작이 같아요: 등록 확인 제거(공통 폴더·소유 기록·브랜치 검사가 같은 경우를 거절), 링크 확인 제거(소유 기록이 링크 이름의 폴더에 없어서 거절), 만들기 전 빌드 폴더 기록 확인 제거(만든 뒤 거절하면 자리를 지워서 결과가 같아요). 첫 판의 변이 중 살아남은 둘(40자 형식 검사 제거, 이동 뒤 HEAD·트리 재확인 제거)도 같은 이유의 등가예요.

둘째 수리(같은 감사자의 델타 감사가 head `b30f15861a` 에서 PASS-WITH-FINDINGS, 막는 결함 없음, 비차단 N8~N12): 새 시험 4개를 먼저 썼고 옛 도구에서 3개가 빨간불(1 ERROR, 2 FAIL), 넷째는 아래 생존 변이 하나를 고정해요.
- N8(정리가 이 실행이 만들지 않은 자리까지 지움): 정리는 이 실행의 `worktree add` 가 끝난 뒤에만 자리를 지워요(`made` 표시, REQ-13). 시험 `test_a_refusal_never_removes_a_place_that_another_run_made`(다른 실행이 먼저 만든 자리를 흉내 내요).
- N9(`prune` 이 남의 등록도 잊음): `prune` 을 빼고 `worktree add --force` 로 그 자리 등록 하나만 넘겨받아요(REQ-14). 시험 `test_forgetting_a_deleted_place_leaves_other_missing_worktrees_registered`.
- N10(고친 `.gitignore`·추적 안 된 `.gitignore` 가 규칙을 바꿈): 무시된 파일이 있을 때 자리의 상태 목록에 `.gitignore` 가 있어도 `ignore_rules_changed` 로 거절해요(REQ-12). 시험 `test_an_edited_or_untracked_gitignore_counts_as_a_rule_change`. 옛 도구가 만든 자리(기록에 `place` 키 없음)는 소유 기록이 없다고 거절되는데, 그 자리는 지우고 다시 만들면 돼요(문서에 남겨요).
- 생존 변이(만들기 전 빌드 폴더 기록 확인 제거)는 이미 있는 자리에서는 같은 결과가 아니었어요. 시험 `test_a_foreign_build_folder_record_refuses_before_an_existing_place_moves` 가 고정해요.
- N11(SPEC 글의 이름에 카드 번호): 감사자·측정 폴더 이름을 번호 없는 말로 바꿨어요. N12(옛 글줄): AC-1 개수와 "9개 중 8개" 줄을 고쳤어요. 이동 뒤에 거절하는 경우의 시험은 N10 시험이 앞단에서 막는 것으로 바뀌었어요.

아직 안 한 것: 감사 에이전트 지침·감사 워크플로·`skills/moai/references/internal/build-folder.md` 에 이 도구를 쓰라는 문장, `moai audit-clone-plan` 명령 목록(다음 조각), 독립 감사, Linux 시험.
