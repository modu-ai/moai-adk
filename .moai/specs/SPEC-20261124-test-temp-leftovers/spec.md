---
id: SPEC-20261124-test-temp-leftovers
title: "테스트가 시스템 임시 폴더에 남기는 래퍼 파일을 막아요"
status: approved
revision: 1
created: 2026-11-24
approval: {by: leader, at: 2026-11-23, basis: "leader decision note 7890 of card C-432 (2026-11-23T20:57:14Z, delegation of the user 2026-11-23T13:57Z, the leader's judgement and not a selection by the user) answering request note 7889 with D-1 option (a)", digest: "sha256:fc9b8e7afed524cc190282c4b109d30bca1564d6c882fc4cff01c98d74f54196"}
---

# 테스트가 시스템 임시 폴더에 남기는 래퍼 파일을 막아요

## Interview

- none: 리더 배차가 요구사항을 전달했어요 — 카드 본문(테스트가 남기는 이름을 성공·실패 모두에서 지우고, 임시 폴더에 그 이름이 없음을 확인하는 테스트)과 리더 배차 지시가 범위를 정했고, 별도 사용자 인터뷰는 없었어요.

## Intent

시스템 임시 폴더(`$TMPDIR`)에 `c105-plan-*` 폴더 339개와 빈 `moai-test-home-ref.*` 파일이 쌓였어요. 2026-11-24 에 리더가 지웠고, 새로 쌓이지 않게 만드는 일이 이 SPEC 이에요. 생산자를 이 실행에서 읽고 재 보았어요(NeXT `6054432304`).

- `c105-plan-*`: `tests/profile_plan_cli.rs:46`, `:346` 이 `temp_dir().join(format!("c105-plan-…{pid}…"))` 를 만들고 지우지 않았어요. 이미 `SPEC-20261124-test-temp-names`(커밋 `70fc6cb75e`)가 두 곳을 `tempfile::tempdir()` 로 바꾸고 `tests/test_temp_names_scan.rs` 로 새 pid 이름을 막았어요. 지금 시스템 임시 폴더의 `c105-plan-*` 는 0개예요. 이 SPEC 은 그쪽을 다시 고치지 않아요.
- `moai-test-home-ref.*` 파일과 `moai-test-home.*` 폴더: `tools/test_home_guard.sh:39`, `:30` 이 `mktemp` 로 만들고 `:40` 의 EXIT trap 이 지워요. nextest 래퍼라서 nextest 가 도는 모든 테스트가 하나씩 거쳐요. 비공개 `TMPDIR` 에서 직접 돌려 보면, 정상 종료와 테스트 실패(rc 3)는 아무것도 남기지 않고 래퍼에 SIGKILL 을 보내면 파일 하나와 폴더 하나가 남아요.
- 두 번째 생산자: `tests/test_home_guard.rs:566` 의 `nextest()` 가 중첩 `cargo nextest` 를 `TMPDIR` 를 바꾸지 않고 돌려요(여섯 테스트가 불러요). 진짜 래퍼가 시스템 임시 폴더에 이름을 만들고, 바깥 테스트가 중간에 죽으면 거기에 남아요.

확인하지 못한 것: 리더가 지운 339개가 어느 종료 경로에서 나왔는지는 보지 못했어요(지금 시스템 임시 폴더에는 낡은 `ref` 가 0개예요). 이 SPEC 은 잡을 수 있는 종료 경로와 중첩 nextest 를 막고, SIGKILL 은 `D-1` 로 리더의 선택에 맡겨요.

## Non-goals

- 시스템 임시 폴더에 이미 있는 것을 지우지 않아요(`wfplan-*` 418개는 수리 전에 생긴 것이고, 지우는 일은 운영자의 몫이에요).
- 래퍼가 SIGKILL 을 맞았을 때의 찌꺼기: 셸 trap 이 잡을 수 없는 신호라서 이 SPEC 의 요구가 아니에요(`D-1`).
- `c105-plan-*` 와 `wfplan-*` 생산자를 다시 고치지 않아요(`SPEC-20261124-test-temp-names` 가 닫았어요).
- `src/` 안 단위 테스트의 임시 폴더(`src/review.rs:538` 등): 그 SPEC 의 후속 목록에 있고 그 파일을 가진 카드의 일이에요.
- 래퍼의 HOME 감시 규칙(무엇을 건드린 것으로 보는지)을 바꾸지 않아요.

## Requirements

- REQ-1: When the wrapper ends by a normal exit, a failing test, or a HUP, INT or TERM signal, the wrapper shall leave neither a `moai-test-home.*` folder nor a `moai-test-home-ref.*` file in its `TMPDIR` · failure: none — a SIGKILL cannot be caught by a shell and is outside this REQ (`D-1`).
- REQ-2: When a test of `tests/test_home_guard.rs` runs `cargo nextest` on a scratch crate, the test shall set `TMPDIR` to a folder inside that scratch crate's `TempDir` · failure: none — the folder is removed when the test ends, panic or not.
- REQ-3: When a nextest run of a scratch crate ends, the run helper shall fail the test if the private `TMPDIR` still holds a name that starts with `moai-test-home` · failure: none — the test names the leftover and fails.

## Acceptance

- AC-1 (REQ-1): the wrapper run with a private `TMPDIR` under the probes exit 0, exit 7, `kill -TERM`, `kill -INT` and `kill -HUP` of its own process leaves that folder empty; the signal probes end with exit code 143 · verify: `cargo test --locked --offline --test test_home_guard a_wrapper_that_ends_leaves_nothing_in_its_tmpdir`
- AC-2 (REQ-2): a scratch-crate test that reads `TMPDIR` inside the nextest run sees the private folder, and the six nextest tests of the file pass · verify: `cargo test --locked --offline --test test_home_guard nextest`
- AC-3 (REQ-3): the run helper names the leftover and fails when the wrapper leaves a `moai-test-home*` name, shown by the guard mutant that drops the `ref` removal · verify: `cargo test --locked --offline --test test_home_guard`
- AC-4 (REQ-1, REQ-2): after the whole file ran, the system temp folder holds no `moai-test-home.*` or `moai-test-home-ref.*` made by this run · verify: not a test (other lanes run the wrapper at the same time); a hand run of the file with a private `TMPDIR` and the entry count of that folder, recorded in Result

## Scope

- write: `.moai/specs/SPEC-20261124-test-temp-leftovers/spec.md`, `tests/test_home_guard.rs`
- read-only: `tools/test_home_guard.sh` (the guard mutants change it for one run and `git checkout` puts it back; the diff holds no change of it), `.config/nextest.toml`, `tests/profile_plan_cli.rs`
- forbidden: the existing contents of the system temp folder, `tools/repo_scans.json`, `src/`

## Plan

1. 이 SPEC 과 리더 승인 요청 노트(`D-1` 선택 포함). 승인 전에는 코드를 쓰지 않아요.
2. RED: `a_wrapper_that_ends_leaves_nothing_in_its_tmpdir`(다섯 경우), 중첩 nextest 의 비공개 `TMPDIR` 를 보는 plant(`plant_tmpdir_is_private`), 도우미의 `moai-test-home*` 검사를 먼저 쓰고 빨간 것을 봐요(`nextest()` 가 아직 `TMPDIR` 를 안 바꾸므로 plant 가 빨개요).
3. 구현: `nextest()` 가 `TempDir` 안의 `private-tmp` 를 만들어 `TMPDIR` 로 주고, 끝난 뒤 `moai-test-home*` 가 남았는지 검사해요.
4. guard-mutant 셋(아래 Risks), 한 번 되돌린 뒤 `moai check affected` 맨몸으로 한 번, READY.

## Risks

- 신호 경우는 프로브가 `kill -s <신호> $PPID` 로 래퍼 셸에 직접 신호를 보내요. 셸은 자식(프로브)이 끝난 뒤에 trap 을 돌리므로 결과가 정해져 있어요. 이것이 다른 셸에서도 같은지는 이 저장소의 `/bin/sh`(macOS)에서만 봤어요. Linux 결과는 주장하지 않아요.
- 비공개 `TMPDIR` 는 `TempDir` 안이라 `cargo`/`rustc` 가 거기에 쓰는 임시 파일도 같이 지워져요. 중첩 실행이 시스템 임시 폴더를 쓰던 경로가 있었다면 달라질 수 있어서 여섯 테스트를 그대로 통과시켜 확인해요.
- 래퍼가 SIGKILL 을 맞으면 `moai-test-home.*` 폴더 하나와 `moai-test-home-ref.*` 파일 하나가 남아요(이 SPEC 은 고치지 않아요, `D-1` (a)). 리더가 지운 339개의 원인이 SIGKILL 인지는 모르니, 다시 쌓이는 게 보이면 (b) 를 후속 카드로 해요.
- guard-mutant(각각 한 번 적용하고 반대 편집으로 되돌려요, 되돌린 뒤 `git status` 가 깨끗해야 해요):
  - M1: `tools/test_home_guard.sh:40` 의 `rm -f "$ref"` 를 뺀다 → `a_wrapper_that_ends_leaves_nothing_in_its_tmpdir` 와 `nextest()` 도우미가 잡아야 해요.
  - M2: `tools/test_home_guard.sh:24` 의 `trap 'exit 143' HUP INT TERM` 를 뺀다 → 신호 경우가 잡아야 해요.
  - M3: `nextest()` 의 `TMPDIR` 설정을 뺀다 → `plant_tmpdir_is_private` 가 잡아야 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 래퍼의 SIGKILL 찌꺼기는 이번에 고치지 않고 `ref` 를 폴더 안으로 옮기는 설계 변경도 하지 않아요(기본값). 대안 (b): `ref` 를 `moai-test-home.*` 폴더 안으로 옮겨 SIGKILL 때 남는 것을 폴더 하나로 줄여요. 대안 (c): 다음 실행이 낡은 이름을 지워요 | 셸은 SIGKILL 을 잡을 수 없어요(측정: 비공개 `TMPDIR` 에서 SIGKILL 하면 파일 하나와 폴더 하나가 남음). (b) 는 `find -newer` 비교의 기준이 폴더 자신의 수정 시각과 겹쳐 HOME 감시의 판정이 바뀌어요. (c) 는 다른 실행의 살아 있는 파일을 지울 위험이 있어요. 두 대안 모두 감시 도구의 설계 변경이라 리더의 승인이 필요해요 | 승인 노트에 (b) 나 (c) 를 고르면 후속 SPEC 으로 나눠요 |
| D-2 | 새 시험은 `tests/test_home_guard.rs` 에 넣고 새 테스트 파일을 만들지 않아요 | 새 파일은 `tools/repo_scans.json` 의 표 항목을 요구할 수 있고 그 파일은 다른 카드가 쥐고 있어요. 테스트 파일 하나를 늘리면 빌드도 늘어요 | 필요하면 새 파일로 나눠요 |
| D-3 | 시스템 임시 폴더에 이미 있는 이름은 지우지 않아요 | 운영자의 폴더예요. 이미 리더가 지웠어요 | 해당 없음 |

## Result

판 1, 통합 팁 `6054432304` 에서 딴 브랜치예요(팁이 안 움직여서 머지가 없었어요). 승인 근거는 리더 결정 노트 7890(2026-11-23T20:57:14Z)이고 `D-1` (a) 예요. 승인 뒤 `## Interview` 구역을 하나 넣었고 요구사항의 digest 는 그대로예요(`spec show` 로 확인). 숫자는 이 결과를 쓰려고 이 워크트리에서 돌린 실행에서 나온 거예요.

- AC-1 met (REQ-1): `a_wrapper_that_ends_leaves_nothing_in_its_tmpdir` 가 exit 0, exit 7, `kill -TERM`, `kill -INT`, `kill -HUP` 다섯 경우에 비공개 `TMPDIR` 가 비어 있고 신호 경우는 143 으로 끝남을 확인해요. `cargo test --locked --offline --test test_home_guard` 41개 통과(exit 0).
- AC-2 met (REQ-2): scratch crate 의 plant `plant_tmpdir_is_private` 가 중첩 nextest 안에서 `TMPDIR` 가 `…/private-tmp` 로 끝남을 봐요. 구현 전에는 FAIL(RED, 같은 명령 exit 101, `nextest_applies_the_repository_configuration_to_a_scratch_crate` 실패), 구현 뒤에는 PASS.
- AC-3 met (REQ-3): `nextest()` 도우미가 끝난 뒤 비공개 `TMPDIR` 에서 `moai-test-home` 로 시작하는 이름을 찾아 실패해요. M1 이 이 검사와 래퍼 테스트로 잡혔어요.
- AC-4 met (REQ-1, REQ-2): 파일 전체를 비공개 `TMPDIR` 로 돌린 뒤(41개 통과, exit 0) `find <그 폴더> -mindepth 1` 이 0개였어요(테스트가 아닌 손 확인).
- guard-mutant(각각 한 번 적용하고 `git checkout` 으로 되돌렸고, 되돌린 뒤 `git status` 는 깨끗했어요): M1 `tools/test_home_guard.sh` 의 trap 에서 `rm -f "$ref"` 제거 → `a_wrapper_that_ends_leaves_nothing_in_its_tmpdir` 와 중첩 nextest 도우미 검사가 잡음(여섯 테스트 실패, cargo exit 101). M2 `trap 'exit 143' HUP INT TERM` 제거 → `a_wrapper_that_ends_leaves_nothing_in_its_tmpdir` 만 실패(40 통과 1 실패, exit 101). M3 `nextest()` 의 `TMPDIR` 설정 제거 → `nextest_applies_the_repository_configuration_to_a_scratch_crate` 실패(`TMPDIR is "/var/folders/…/T/"`, exit 101).
- 고치지 않은 것(`D-1` (a)): 래퍼가 SIGKILL 을 맞았을 때 남는 `moai-test-home.*` 폴더와 `moai-test-home-ref.*` 파일. 리더가 지운 339개의 원인이 이 경로인지는 못 봤고, 다시 쌓이는 게 보이면 (b) 를 후속 카드로 해요.
- `c105-plan-*` 의 만든 곳은 `tests/profile_plan_cli.rs:46`, `:346`(통합 팁 `466b59fc62` 기준)이고 `SPEC-20261124-test-temp-names`(`70fc6cb75e`)가 이미 고쳤어요. 이 SPEC 이 고친 것이 아니에요.
- Not run: 전체 `cargo test`, Linux·Windows 에서의 신호 경우, 시스템 임시 폴더 전체를 보는 테스트(다른 레인의 실행과 겹쳐서 테스트로 만들 수 없어요).
- `moai check affected` 영수증은 카드의 READY 노트에 있어요.
- closure_status: open until the audit.
