---
id: SPEC-20261120-audit-warm-start
title: "An audit worktree starts with a copy-on-write clone of the lane's build folder instead of a cold build"
status: draft
revision: 1
created: 2026-11-20
---

# An audit worktree starts with a copy-on-write clone of the lane's build folder instead of a cold build

## Intent

사용자 결정(2026-11-20): 독립 감사가 30~60분 걸리는 가장 큰 이유는 감사자마다 저장소를 받아 자기 `CARGO_TARGET_DIR` 에서 작업공간 전체를 처음부터 빌드하기 때문이에요. 감사자의 빌드 폴더를 레인의 빌드 폴더에서 **복사 없는 복제**(APFS `clonefile`, Linux btrfs·xfs 의 reflink)로 시작해요. 복제가 안 되는 곳에서는 바이트 복사를 하지 않고 차가운 빌드로 가요.

이 컴퓨터에서 먼저 쟀어요(2026-11-20, APFS, 레인 폴더 4.4 GiB): 복제 8초에 디스크 증가 없음, 다른 worktree 경로에서 같은 커밋을 `cargo test --no-run` 하면 86개 crate 중 작업공간 crate 하나만 다시 컴파일하고 의존성 85개는 그대로 써요(경로가 바뀌면 작업공간 crate 의 id 가 바뀌어서 그 하나는 다시 빌드돼요). 시간 비교는 아래 Result 에 있어요.

이 SPEC 은 도구 한 가지만 해요: `tools/build_folder.py seed --from <레인 폴더 이름>`. 감사 에이전트 지침·감사 워크플로·감사 clone 계획에 이 도구를 쓰라고 적는 일은 그 글을 쥔 다른 변경이 끝난 뒤에 넣어요(Non-goals).

## Non-goals

- 지침 글(`agents/audit.md`, `skills/moai/workflows/audit.md`, `verify.md`, `skills/moai/references/internal/build-folder.md`)과 `moai audit-clone-plan` 의 명령 목록에 `seed` 를 더하는 일: 그 파일을 쥔 변경이 끝난 뒤에 따로 넣어요.
- 새 `moai audit-workspace` 낱말: 이 도구는 이미 있는 `tools/build_folder.py` 의 하위 명령이고, 바이너리에 새 낱말을 넣지 않아요.
- 빌드 폴더 하나를 여러 worktree 가 함께 쓰는 일(잠금 때문에 빌드가 한 줄로 서요): 폴더는 계속 worktree 마다 하나예요.
- 복제를 지원하지 않는 곳에서 바이트 복사로 대신하는 일.

## Requirements

- REQ-1: When `build_folder.py seed --from <name>` runs in a worktree whose own build folder holds no cargo output, the tool shall clone the cargo output of the folder `<HOME>/.moai/build/<name>` into it by copy-on-write and print `seeded: true` with the method used. · failure: none
- REQ-2: The tool shall leave the source folder unchanged, cloning neither its owner record nor any file that is not cargo output. · failure: none
- REQ-3 (REQ-1): IF the platform or the filesystem cannot clone, or `BUILD_FOLDER_NO_COW=1` is set, THEN the tool shall print `seeded: false` with the reason, exit 0 and copy no bytes.
- REQ-4 (REQ-1): IF the cargo lock of the source folder is held, THEN the tool shall print `seeded: false` with the reason `source_building` and clone nothing.
- REQ-5 (REQ-1): IF the worktree's own build folder already holds cargo output, THEN the tool shall print `seeded: false` with the reason `already_built` and change nothing.
- REQ-6: IF the source is not a real folder directly under `<HOME>/.moai/build/`, is a link, or is the worktree's own folder, THEN the tool shall refuse with exit 1 and write nothing. · failure: none
- REQ-7: When the folder is seeded, the tool shall write the worktree's owner record as `env` does. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-7): a seeded folder holds the source's cargo output, equal in content, and the source and its record are unchanged.
  verify: `python3 -m unittest tests.build_folder_test.SeedTest`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): every non-seeding case prints `seeded: false` or refuses, and copies nothing.
  verify: `python3 -m unittest tests.build_folder_test`
- AC-3 (REQ-1): a seeded worktree builds an unchanged crate without recompiling its dependencies; the before and after numbers are in Result.
  manual: seed a folder from a built lane folder, run `cargo test --no-run` in a worktree at another path and count the `Compiling` lines against a cold build; write both counts and the CPU times in Result.

## Scope

- write: `.moai/specs/SPEC-20261120-audit-warm-start/**`, `tools/build_folder.py`, `tests/build_folder_test.py`
- read-only: `src/optimize.rs` (the owner record rule), `src/worktree_gc.rs` (what counts as cargo output)
- forbidden: any write in the real HOME by a test, a byte copy of a build folder, a write into the source folder, a new card, Issue or pull request number in the text this change adds

## Plan

1. SPEC, then RED tests (`SeedTest` in `tests/build_folder_test.py`).
2. `tools/build_folder.py`: `seed` (macOS `clonefile` through `ctypes`, Linux `cp --reflink=always`, no other platform), the lock probe, the refusals.
3. Checks, mutants, a read-only audit by another agent. Skill text follows later.

## Risks

- A clone taken while the lane builds could hold a half-written artifact: the tool takes a shared lock on the source's cargo lock and does not clone when the lane holds it.
- After a clone, cargo rebuilds the workspace crate (its id depends on the path) and reuses the dependencies; a source folder built at a very different commit saves less. The tool does not judge that.
- `clonefile` is not atomic across the whole set of entries: a failure part-way leaves the cloned entries, so the tool removes the entries it made before it reports `seeded: false`.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A clone that cannot be made is not an error: exit 0, `seeded: false` | The auditor then builds cold, as before; a failed seed must not stop an audit | Make it exit 1 |
| D-2 | Clone only cargo output, never the record or evidence | Evidence stays with its worktree; the record names the owner | none needed |
| D-3 | No `--force-copy` option | A byte copy of tens of GB is what this change avoids | Add the option |

## Result

(written after the checks)
