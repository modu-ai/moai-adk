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
- REQ-4: While the place holds changes to tracked files or untracked files that git does not ignore, the tool shall reset the tracked files and remove those untracked files, print how many it took back, and leave the files git ignores alone. · failure: none
- REQ-5: IF a cargo run holds the cargo lock of the place's build folder, THEN the tool shall refuse with exit 1 and the reason `place_busy` and change nothing.
- REQ-6: IF `<sha>` is not 40 lowercase hexadecimal digits or names no commit of this repository, or `<n>` is not one safe folder name, THEN the tool shall refuse with exit 1 and write nothing.
- REQ-7: IF the place path exists but is a link, is not a folder, or is not a worktree of this repository, THEN the tool shall refuse with exit 1 and change nothing.
- REQ-8: When the tool runs, it shall record `CARGO_HOME` (the environment value, else `<HOME>/.cargo`) in the build folder's `.moai-audit-place.json`. · failure: none
- REQ-11: IF a run sees a `CARGO_HOME` other than the recorded one, THEN the tool shall print `cargo_home_changed: true` with the earlier value as `cargo_home_was`.
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

(written after the checks)
