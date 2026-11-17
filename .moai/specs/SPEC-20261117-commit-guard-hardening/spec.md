---
id: SPEC-20261117-commit-guard-hardening
title: commit guard의 lane 등록부 조회가 링크로 꺼지지 않게
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-commit-guard-users
---

## Intent

C-80(PR #3167) 감사의 후속 두 가지예요. `issue/<n>` commit guard는 그 clone에 살아 있는 lane 세션이 있을 때만
적용돼요(`claim_verify::lane_clone`). 그 조회는 링크로 된 `sessions` 폴더만 거절해서, git common dir 아래 clone
상태 폴더(`.moai-cli`)가 빈 등록부를 가리키는 링크면 "lane 없음"으로 읽고 guard가 꺼졌어요(N1, 감사 probe 7).
또 시험 도우미 `tests/common::register_lane`은 실제 환경 그대로 `session-register`를 돌렸어요(N4).

## Non-goals

- guard 규칙 자체(언제 lane으로 보는지), 등록부 형식, `session-register` 동작 바꾸기.
- 다른 `lane_registry` 사용처(launcher, statusline)의 링크 처리 바꾸기.

## Requirements

- REQ-1: When the commit guard reads the session registry of a clone, it shall check every path component below the git common dir with `lstat` before it reads the folder.
- REQ-2 (REQ-1): IF any of those components is a symbolic link, THEN the guard shall treat the clone as a lane clone and keep refusing an unclaimed `issue/<n>` commit.
- REQ-3: The test helper that registers a lane shall run `session-register` with a temporary HOME. · failure: none — a test fixture; the home-isolation scan pins it.

## Acceptance

- AC-1 (REQ-1, REQ-2): `.git/.moai-cli`가 빈 등록부(빈 폴더, 또는 빈 `sessions`가 든 폴더)를 가리키는 링크면 두 guard
  (`hook-dispatch --event tool-preflight`, `claim-verify . --hook`) 모두 거절하고, 같은 모양의 진짜 폴더면 둘 다 허용해요.
  verify: `cargo test --locked --offline --test commit_guard_users`
- AC-2 (REQ-3): `tests/common/mod.rs`가 HOME 없이 binary를 돌리는 목록에서 빠지고, 그 scan이 통과해요.
  verify: `cargo test --locked --offline --test home_isolation_scan`
- AC-3 (REQ-1…REQ-3): 이웃 검사가 통과해요.
  verify: `cargo test --locked --offline --test claim_verify_cli`; `cargo test --locked --offline --test hook_dispatch_cli`;
  `cargo test --locked --offline --test session_registry_cli`; `cargo test --locked --offline --test git_env_scrub`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --lib claim_verify`;
  `cargo test --locked --offline --lib lane_registry`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`; `python3 -m unittest tests.repo_scans_test tests.ci_scope_test`

## Scope

- write: `src/lane_registry.rs`, `src/claim_verify.rs`, `tests/commit_guard_users.rs`, `tests/common/mod.rs`,
  `tests/fixtures/home_isolation/binary_without_home.txt`, `.moai/specs/SPEC-20261117-commit-guard-hardening/spec.md`
- read-only: `src/clone_root.rs`, `src/identity.rs`
- forbidden: 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 시험 먼저(RED): 링크된 clone 상태 폴더 두 모양에서 guard가 꺼지는지 봐요.
2. `lane_registry::registry_dir_unlinked`(common dir 아래 구성요소마다 `lstat`)와 `lane_clone`이 그것을 쓰게.
3. `register_lane`에 임시 HOME, home-isolation 기준 목록에서 한 줄 빼기.
4. 시험·형식·clippy·Python, mutation 3개.

## Risks

- git common dir 자체(`.git`)가 링크인 경우는 검사하지 않아요 → common dir은 git이 알려 준 경로이고, 링크된 `.git`은
  정상 설정일 수 있어서 거절하면 일반 사용자의 commit까지 막혀요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | PR #3167이 아직 열려 있어 그 head(`5bb154ffb7`)에서 시작 | 리더 지시(병합 전이면 그 가지에서) | 병합 뒤 tip을 병합 |
| D2 | 검사 범위는 common dir 아래 구성요소(상태 폴더, `sessions`) | 리더 지시("from the common dir down"), Risks의 `.git` 링크 | common dir도 검사 |
| D3 | 새 함수 `registry_dir_unlinked`를 두고 guard만 씀 | 다른 사용처 동작을 바꾸지 않기 | `registry_dir` 자체를 바꾸기 |

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 상태는 `draft` 그대로예요(승인은 사용자 몫).
기준 `5bb154ffb7`(PR #3167 head; 10:13Z에 아직 OPEN). 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`,
`CARGO_BUILD_JOBS=4`, 무거운 작업 슬롯 안에서 cargo 한 번에 하나. RED: 새 시험 `a_linked_clone_state_folder_keeps_the_guard_on`이
바뀌기 전 코드에서 `(false, Some(0))`(guard 꺼짐)으로 실패 → GREEN.

- AC-1: met — `commit_guard_users` 7 passed. mutation M1(링크 검사 없이 읽기 = N1 버그) KILLED, M2(마지막 구성요소만 검사)
  KILLED, M3(링크면 lane 아님으로 봄) KILLED
- AC-2: met — `home_isolation_scan` 16 passed(기준 목록에서 `tests/common/mod.rs` 한 줄 뺌)
- AC-3: met — `claim_verify_cli` 39, `hook_dispatch_cli` 56, `session_registry_cli` 22, `git_env_scrub` 3, `product_cli` 75,
  `hook_mode_cli` 5, `hook_status_cli` 10, `hooks_cli` 22, `spec_cli` 17, `--lib claim_verify lane_registry` 12,
  `--bin moai` 27; fmt 0, `rustfmt src/ops` 0, clippy 0; Python `repo_scans`·`ci_scope` 44개 OK; `moai spec check --all` 0

실행 안 함: 전체 `cargo test`, Linux·Windows.
