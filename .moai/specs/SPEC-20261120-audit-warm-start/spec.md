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

이 컴퓨터에서 먼저 쟀어요(2026-11-20, APFS, 레인 폴더 4.4 GiB): 복제 8초에 디스크 증가 없음, 다른 worktree 경로에서 같은 커밋을 `cargo test --no-run` 하면 86개 crate 중 작업공간 crate 하나만 다시 컴파일하고 의존성 85개는 그대로 써요(경로가 바뀌면 작업공간 crate 의 id 가 바뀌어서 그 하나는 다시 빌드돼요). 단 이 재사용은 원본과 `CARGO_HOME` 경로가 같을 때만 돼요(독립 감사가 다른 `CARGO_HOME` 에서 86개를 모두 다시 컴파일하는 것을 재현했어요). 시간 비교는 아래 Result 에 있어요.

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
- REQ-6: IF the source is not a real folder directly under `<HOME>/.moai/build/`, is a link, or is the worktree's own folder (in any letter case on a case-insensitive disk), THEN the tool shall refuse with exit 1 and write nothing. · failure: none
- REQ-7: When the folder is seeded, the tool shall write the worktree's owner record as `env` does. · failure: none
- REQ-8 (REQ-1): IF the worktree's folder already holds a path the clone would take, THEN the tool shall print `seeded: false` with the reason `target_in_the_way`, clone nothing and change nothing there.
- REQ-9 (REQ-1): IF the source holds no cargo output, THEN the tool shall print `seeded: false` with the reason `source_empty`.
- REQ-10 (REQ-1): IF a clone fails part-way, THEN the tool shall remove only the paths it made in this run and print `seeded: false` with the reason `unsupported` when the platform cannot clone and `clone_failed` otherwise.
- REQ-11 (REQ-1): IF the folder `<HOME>/.moai/build` does not exist, THEN the tool shall print `seeded: false` with the reason `source_missing` and exit 0.
- REQ-12 (REQ-1): While the tool clones, it shall hold a shared lock on the cargo locks of the source, so that a lane that starts to build waits.
- REQ-13: When `clean` removes the cargo output of a clone that kept the read-only modes of its source, the tool shall make the folders writable for their owner and remove them. · failure: none
- REQ-14: When the tool seeds a folder, it shall say in its answer that dependencies are reused only when CARGO_HOME is the same path as the lane's. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-7): a seeded folder holds the source's cargo output, equal in content, and the source and its record are unchanged.
  verify: `python3 -m unittest tests.build_folder_test.SeedTest`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): every non-seeding case prints `seeded: false` or refuses, and copies nothing.
  verify: `python3 -m unittest tests.build_folder_test`
- AC-4 (REQ-6, REQ-8, REQ-9, REQ-11, REQ-13, REQ-14): a path in the way, an empty source, the own folder in another letter case, a missing build root, a read-only clone and the CARGO_HOME note behave as written.
  verify: `python3 -m unittest tests.build_folder_test.SeedTest`
- AC-5 (REQ-10, REQ-12): a failure part-way removes only what the run made and tells "cannot clone" from "failed"; the source is locked while it is cloned.
  verify: `python3 -m unittest tests.build_folder_test.SeedRollbackTest`
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

- A clone taken while the lane builds could hold a half-written artifact: the tool does not clone when the lane holds the source's cargo lock, and it keeps a shared lock on it while it clones, so a lane that starts meanwhile waits.
- After a clone, cargo rebuilds the workspace crate (its id depends on the path) and reuses the dependencies only when `CARGO_HOME` is the same path as the lane's: the build info records the registry source paths, and a different path made every dependency "dirty" in the audit of this change (86 crates compiled, as cold). A source folder built at a very different commit saves less. The tool does not judge either; its answer carries the CARGO_HOME note.
- `clonefile` is not atomic across the whole set of entries: a failure part-way leaves the cloned entries, so the tool removes the entries it made in this run before it reports `seeded: false`. It checks first that no path of the clone is taken, so everything it removes is something it made.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A clone that cannot be made is not an error: exit 0, `seeded: false` | The auditor then builds cold, as before; a failed seed must not stop an audit | Make it exit 1 |
| D-2 | Clone only cargo output, never the record or evidence | Evidence stays with its worktree; the record names the owner | none needed |
| D-3 | No `--force-copy` option | A byte copy of tens of GB is what this change avoids | Add the option |

## Result

`python3 -I tools/build_folder.py seed --from <name>` 가 생겼어요. 레인의 빌드 폴더(`~/.moai/build/<name>`)에서 cargo 산출물만(`debug`·`CACHEDIR.TAG`·`.rustc_info.json`·`tmp`) 복사 없이 복제하고(macOS `clonefile`, Linux `cp --reflink=always`), 소유 기록은 이 worktree 의 것으로 새로 쓰고, 원본은 읽기만 해요. 복제할 수 없으면 바이트 복사 없이 `seeded: false` 와 이유(`disabled`, `source_missing`, `already_built`, `source_building`, `unsupported`, `clone_failed`)를 내고 exit 0 이라서 감사자는 차가운 빌드로 이어가요. 링크·폴더 이름 하나가 아닌 값·자기 폴더·링크된 빌드 루트는 exit 1 로 거절해요.

측정(2026-11-20, 이 Mac, APFS, 레인 폴더 `moai-adk-audit-copies` 약 4.4 GiB, 같은 커밋 `b578e11043` 을 다른 worktree 경로에서 `cargo test --offline --locked --test audit_copy_guard --no-run`):
- 복제: `cp -c -R` 8초, 복제 전후 `df` 빈 공간이 줄지 않았어요(오히려 30116 KB 늘었고, 다른 프로세스의 영향이라 "디스크 증가 없음"까지만 말해요). 명령은 `warm-probe.sh` 의 `df`/`cp -c -R`.
- 컴파일한 crate: 따뜻한 시작 1개(`moai-cli`) / 차가운 시작 86개. 두 라운드 모두 같았어요.
- CPU 시간(user+sys, 기계 부하 55~70 이라 벽시계 시간은 믿지 않아요): 1라운드 따뜻 84.17+3.42=87.6초 / 차가움 114.58+8.83=123.4초, 2라운드 따뜻 85.17+3.52=88.7초 / 차가움 123.76+9.38=133.1초. 따뜻한 쪽이 CPU 를 약 29~33% 덜 써요. 벽시계: 56.2 / 60.0 초, 58.4 / 83.3 초.
- 왜 1개가 남나: 소스 파일 날짜를 전부 2020 년으로 돌려도 `moai-cli` 1개는 다시 빌드돼요(소스를 읽는 `env!("CARGO_MANIFEST_DIR")` 가 빌드 기록에 레인의 경로로 남아서, 경로가 다른 worktree 에서는 값이 달라요). 의존성 85개는 worktree 경로와 상관없이 그대로 재사용돼요 — 단 원본과 `CARGO_HOME` 경로가 같을 때만이에요: 첫 독립 감사가 새 `CARGO_HOME`(`/tmp/moai-audit-cargo-warm`)에서 같은 방식으로 `seed` 한 뒤 86개를 모두 다시 컴파일했고(`CARGO_LOG=cargo::core::compiler::fingerprint=info` 에서 모든 의존성이 `PathToSourceChanged`, 원본의 빌드 정보에 `/tmp/moai-rc-cargo-agent-1/registry/src/...` 가 박혀 있었어요), 같은 `CARGO_HOME` 에서는 0개였어요. 위 86 → 1 은 같은 `CARGO_HOME`(`/tmp/moai-rc-cargo-agent-1`)에서 쟀어요. 그래서 절약은 "같은 `CARGO_HOME` 일 때 의존성 crate 전부"이고 작업공간 crate 하나는 늘 다시 빌드돼요. 측정에 쓴 커밋은 `b578e11043`(다른 가지 `spec/20261120-audit-no-repo-copies`, `audit_copy_guard` 시험이 있는 head)이고, 이 SPEC 의 base 에는 그 시험 대상이 없어요.
- `seed` 를 실제로 쓴 증거: 이 SPEC 의 READY 묶음이 `seed --from moai-adk-audit-copies` 로 폴더를 시작해(`seeded: true`, `method: clonefile`) clippy 부터 `--lib` 1636 개까지 돌렸어요.

- AC-1: met — `python3 -m unittest tests.build_folder_test.SeedTest` 11개 통과(복제 내용 같음, 원본·기록 불변, 증거 파일 안 복제, 소유 기록은 이 worktree 것). 변이 S3·S4·S5·S6 이 이 시험에 걸려요.
- AC-2: met — `python3 -m unittest tests.build_folder_test` 27개 통과(`moai-cli` 가 필요한 RemoveTest 는 빌드한 바이너리를 `target/debug` 에 복제해 두고 돌렸어요). 켜짐·꺼짐·진행 중·이미 빌드·원본 없음·링크·자기 폴더·링크된 루트가 모두 시험돼요. 변이 S1(진행 중인 레인을 그대로 복제)이 걸려요.
- AC-3: met — 위 측정: 따뜻한 시작 crate 1개 / 차가운 시작 86개, CPU 약 29~33% 절약(같은 `CARGO_HOME` 일 때. 다른 `CARGO_HOME` 이면 86개 모두 다시 컴파일돼요).
- AC-4: met — `python3 -m unittest tests.build_folder_test.SeedTest` 19개 통과(이미 있던 파일·폴더·링크는 그대로이고 `target_in_the_way`, 빈 원본은 `source_empty`, 대소문자만 다른 자기 폴더는 exit 1, 빌드 루트가 없으면 `source_missing` exit 0, 읽기 전용 원본의 복제본도 `clean` 이 지워요, 답에 CARGO_HOME 조건이 있어요).
- AC-5: met — `python3 -m unittest tests.build_folder_test.SeedRollbackTest` 3개 통과(둘째 항목 실패 때 첫 항목만 되돌리고 원본 불변, `ENOTSUP`·`EXDEV` 는 `unsupported`, 복제하는 동안 원본 잠금이 `held` 이고 끝나면 `free`).

변이 6개(`mutate-warm.py`, `mutate-warm-b.py`): S1·S3·S4·S5·S6 은 시험이 잡았고, S2(소유 기록도 복제)는 같은 동작이에요 — `is_cargo_entry` 가 기록 파일을 애초에 cargo 산출물로 보지 않아서(확인: `is_cargo_entry(record)` = False) `e.name != RECORD` 는 방어용 중복 조건이에요. 그래서 같은 뜻의 S6(증거 파일까지 복제)로 시험했고 잡혔어요.

READY 검사(`warm-ready-batch.sh`, 앞 커밋 `f49c718469` 기준): `git diff --check`·`cargo fmt --check`·`rustfmt --edition 2021 --check src/ops`·`cargo clippy --all-targets -D warnings`·`cargo build --bins` 종료 0, `--lib` 1636 통과, `--bins` 56 통과, `build_folder` 를 언급하는 시험 대상(`worktree_move`·`lane_env_cli`·`lane_rules_2879`·`agent_defs_package`·`cli_text_citations`·`optimize_plan_cli`·`worktree_cli`·`spec_cli`) 통과, `tools/repo_scans.json` 검사 41개 명령, macOS 시험 파일 10개, 파이썬 `build_folder_test`·`agents_local_build_folder_test`·`heavy_check_slot_test` 통과. 하나 실패했어요: `tests.repo_scans_test` — 내 시험 도우미가 `rglob` 을 써서 "저장소를 훑는 시험" 규칙에 걸렸고, `os.scandir` 로 바꾼 뒤(`cc6105738f`) `repo_scans_test` 25개와 `build_folder_test` 27개를 다시 돌려 통과했어요. 그 뒤 바뀐 건 시험 도우미 하나뿐이라 나머지 묶음은 다시 돌리지 않았어요.

감사 수리(독립 감사가 head `f89488a470` 에서 FAIL): ① 되돌리기가 이미 있던 항목까지 지웠어요(미리 있던 `tmp` 파일, `debug` 증거 폴더, `debug` 링크가 사라짐) → 이제 복제 전에 모든 목적지가 비어 있는지 보고 하나라도 차 있으면 하나도 복제하지 않고 `target_in_the_way`(REQ-8)이며, 되돌리기는 이 실행이 만든 경로만 지워요(REQ-10). ② 복제할 게 없는데 `seeded: true` 였어요(cargo 산출물이 없는 원본, 대소문자만 다른 자기 폴더) → `source_empty`(REQ-9), 자기 폴더는 `os.path.samefile` 로 거절(REQ-6). ③ 실패 경로 시험이 없었어요 → `SeedRollbackTest` 3개(함수를 바꿔 끼운 주입), `SeedTest` 8개 추가. ④ "86 → 1" 은 `CARGO_HOME` 이 같을 때만이라는 조건을 Intro·Risks·Result 에 적었고 답에도 `note` 로 넣었어요(REQ-14). 비차단도 고쳤어요: 복제하는 동안 원본 잠금을 쥐어요(N1, REQ-12), `.moai/build` 가 없으면 오류가 아니라 `source_missing`(N2, REQ-11), Linux `cp` 의 "not supported"·"cross-device" 메시지를 `unsupported` 로 읽어요(N3), 읽기 전용 원본의 복제본도 `clean` 이 지워요(N4, REQ-13), 원본 폴더를 읽다 실패하면 트레이스백 대신 `clone_failed`(N5). 변이 5개(`mutate-warm-r2.py`, 수리 커밋 `2908ce60fb`)를 모두 시험이 잡았어요: 되돌리기 제거(감사의 M7), 이유 고정(감사의 M11), 차 있는 경로 검사 제거, 빈 원본 검사 제거, 대소문자 자기 폴더 검사 제거. 수리 뒤 `python3 -m unittest tests.build_folder_test` 38개, `tests.repo_scans_test` 25개 통과(앞 묶음의 나머지 검사는 이 수리가 `tools/build_folder.py` 와 그 시험·SPEC 뿐이라 다시 돌리지 않았어요; Rust 소스는 바뀐 게 없어요). Linux `reflink` 경로는 여전히 이 Mac 에서 돌리지 못했어요.

아직 안 한 것: 감사 에이전트 지침·감사 워크플로·`internal/build-folder.md` 에 `seed` 를 쓰라는 문장과 `moai audit-clone-plan` 명령 목록에 `seed` 넣기(그 파일을 쥔 변경이 끝난 뒤), Linux `reflink` 길은 이 Mac 에서 시험하지 못했어요(코드로만 읽음, Linux 시험 없음), 독립 감사는 아직 안 했어요. 레인이 빌드하는 중인지 보는 잠금 검사는 시험에서는 `flock` 으로 만든 잠금으로 확인했고, 실제 cargo 로도 한 번 쟀어요(8초 걸리는 `build.rs` 를 가진 임시 크레이트를 `cargo build` 하는 3초째에 `source_is_building` = True, cargo 가 끝난 뒤 False). 시험 자동화에는 넣지 않았어요.
