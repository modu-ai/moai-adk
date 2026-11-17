---
id: SPEC-20261117-spec-before-code-followups
title: SPEC-먼저 hook의 감사 후속(상대 gitdir, 쓰는 파일의 작업 폴더, --date 안내)
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-spec-before-code
---

## Intent

C-86 slice 1(PR #3173, `9bb0f746d9`) 감사의 후속 세 가지예요. tool-preflight의 A5 규칙은 `spec/<yyyymmdd>-<slug>`
가지에서 그 가지의 SPEC이 생기기 전 코드 쓰기를 막아요.

- N1: linked worktree의 `.git` 파일이 상대 `gitdir:`를 쓰면(git `worktree.useRelativePaths`) 그 경로는 작업 폴더 기준으로
  읽어야 해요. 지금 코드는 맞지만, 프로세스 폴더 기준으로 읽는 mutant(M4)를 잡는 시험이 없었어요.
- N2: 작업 폴더를 세션 폴더로 정해서, 주 checkout이 SPEC 없는 spec 가지일 때 그 안에 들어 있는 다른 worktree
  (`.claude/worktrees/<x>/`, 다른 가지)로의 쓰기도 막혔어요. 쓰는 파일 위의 가장 가까운 `.git`으로 정해야 해요.
- N4: 거절 문장이 가지 날짜가 오늘이 아닐 때의 방법(`moai spec new --date YYYY-MM-DD`)도 알려야 해요. 문장은 고정
  문자열 그대로예요(경로·가지를 되풀이하지 않음).

## Non-goals

- 셸 리다이렉션으로 쓰기 막기(A4 빈틈), 다른 guard 규칙 바꾸기.

## Requirements

- REQ-1: When a write is judged, the guard shall take the work tree of each written path from the nearest `.git` folder or file above that path. · failure: none — a path in no work tree is not this rule's.
- REQ-2: When a linked worktree's `.git` file names a relative git folder, the guard shall read it against that work tree. · failure: none — an unreadable pointer reads as no branch, as before.
- REQ-3: The refusal text shall name `moai spec new --date YYYY-MM-DD` and echo no path or branch. · failure: none — a fixed string.

## Acceptance

- AC-1 (REQ-1): 주 checkout(SPEC 없는 spec 가지)과 그 안의 worktree(`work` 가지), 그리고 반대(주 checkout `main`, 안의
  worktree가 SPEC 없는 spec 가지)에서, 어느 폴더에서 쓰든 쓰는 파일의 가지로 막거나 허용해요.
  verify: `cargo test --locked --offline --test spec_before_code_hook the_written_file_decides`
- AC-2 (REQ-2): 상대 `gitdir:`를 쓰는 worktree에서, 세션이 하위 폴더에 있어도 SPEC 없는 코드 쓰기가 막히고 SPEC 뒤에 허용돼요.
  verify: `cargo test --locked --offline --test spec_before_code_hook a_relative_gitdir`
- AC-3 (REQ-3): 거절 문장에 `--date` 안내가 있고 입력의 경로·가지 이름이 없어요.
  verify: `cargo test --locked --offline --test spec_before_code_hook a_code_write_on_a_spec_branch`
- AC-4 (REQ-1…REQ-3): 이웃 검사가 통과해요.
  verify: `cargo test --locked --offline --lib hook_guards`; `cargo test --locked --offline --test hook_dispatch_cli`;
  `cargo test --locked --offline --test hooks_cli`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`; `python3 -m unittest tests.ci_scope_test tests.repo_scans_test`

## Scope

- write: `src/hook_guards.rs`, `tests/spec_before_code_hook.rs`, `.moai/specs/SPEC-20261117-spec-before-code-followups/spec.md`
- read-only: `src/hook_config.rs`, `src/ops/spec.rs`
- forbidden: 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 시험 먼저(RED): N4 문장, N2 두 방향, N1 상대 gitdir(하위 폴더에서).
2. `spec_before_code`가 쓰는 경로마다 `work_tree_top(경로)`로 작업 폴더와 가지를 정하게; 거절 문장에 `--date`.
3. 시험·형식·clippy·Python, mutation 3개(M4 포함).

## Risks

- 이제 세션 폴더 밖의 다른 저장소로 쓰는 경우도 그 저장소가 SPEC 없는 spec 가지면 막혀요 → 쓰는 파일의 가지가 기준이라는
  규칙(N2)과 같은 방향이에요. 저장소 밖 경로는 여전히 이 규칙과 상관없어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 작업 폴더는 쓰는 경로마다 가장 가까운 `.git`으로 | 리더 지시(감사 N2) | 세션 폴더 기준으로 되돌리기 |
| D2 | 거절 문장에 `--date` 안내를 고정 문자열로 더함 | 리더 지시(감사 N4) | 문장 되돌리기 |

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 상태는 `draft` 그대로예요(승인은 사용자 몫).
기준 `9bb0f746d9`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, 무거운 작업 슬롯
안에서 cargo 한 번에 하나. RED(10:52Z): N4 문장 시험과 N2 두 방향 시험이 바뀌기 전 코드에서 실패(안에 든 worktree로의
쓰기가 막힘). N1 시험은 바뀌기 전에도 통과해요 — 코드는 맞았고, 시험은 mutant M4를 잡으려고 더했어요.

- AC-1: met — `the_written_file_decides_the_work_tree_in_both_directions`. mutation M5(세션 폴더 기준, N2 버그) KILLED
- AC-2: met — `a_relative_gitdir_is_read_from_the_work_tree`. mutation M4(상대 gitdir를 프로세스 폴더 기준으로) KILLED
- AC-3: met — `a_code_write_on_a_spec_branch_waits_for_the_branch_spec`. mutation M6(`.moai/specs/` 아래도 막음) KILLED
- AC-4: met — `spec_before_code_hook` 7, `hook_dispatch_cli` 56, `hooks_cli` 22, `home_isolation_scan` 16, `product_cli` 75,
  `spec_cli` 17, `--lib hook_guards` 17, `--bin moai` 27; fmt 0, clippy 0; Python `ci_scope`·`repo_scans` 44개 OK;
  `moai spec check --all` 0

실행 안 함: 전체 `cargo test`, Linux·Windows, 실제 host 세션.
