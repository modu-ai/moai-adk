---
id: SPEC-20261117-spec-before-code
title: "SPEC 작업 가지에서 SPEC보다 먼저 코드를 쓰지 못하게 하는 hook 규칙"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

2026-11-17 사용자 지시 "권고 말고 강제"(결정 기록 "권고 말고 강제" 행, 카드 C-86)는 꼭 지킬 규칙을 지침 문장이 아니라
hook·거부 규칙으로 강제하라고 해요. 그 규칙 가운데 "작업 가지에 SPEC 없이 코드 쓰기 금지"를 첫 조각으로 넣어요. MoAI의 SPEC
작업 가지는 `spec/<yyyymmdd>-<slug>`이고 그 SPEC은 `.moai/specs/SPEC-<yyyymmdd>-<slug>/spec.md`예요(SPEC id 결정). 지금은
SPEC을 나중에 써도 아무것도 막지 않아요. 이 조각은 tool-preflight hook(Claude Code와 Codex가 부르는 `hook-dispatch`)의
기본 거부 규칙 A5로, SPEC이 생기기 전에는 그 가지에서 `.moai/specs/` 밖의 파일 쓰기를 거부해요.

## Non-goals

- 승인 상태(`status: approved`) 요구, GREEN 동안 시험 파일 잠금, AC verify 기록 없는 완료 거부, 작성 세션 판정 거부(C-86의
  다음 조각들).
- Bash 리디렉션으로 쓰는 파일, hook을 부르지 않는 호스트의 쓰기.
- `spec/` 이외 가지(이슈 가지 `issue/…`, 카드 가지)의 SPEC 요구.

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `IF …, THEN …`은 원치 않는 동작이에요.

- REQ-1: When a file-writing tool targets a path inside the work tree but outside `.moai/specs/` on a branch `spec/<yyyymmdd>-<slug>` whose SPEC file is missing, tool-preflight shall refuse the call with a fixed reason that names the SPEC path to write first.
- REQ-2 (REQ-1): IF the SPEC file exists but its frontmatter names another id or it is not a regular file, THEN tool-preflight shall refuse the call as if the SPEC were missing.
- REQ-3 (REQ-1): IF a path reaches outside `.moai/specs/` through `..` or a symbolic link, THEN tool-preflight shall judge it where it lands.
- REQ-4: Tool-preflight shall always allow a write under `.moai/specs/`.
- REQ-5: When the branch's SPEC file names the branch's id, tool-preflight shall allow the write. · failure: none
- REQ-6: Tool-preflight shall leave writes on other branches, writes outside the work tree and Bash commands to the other rules.
- REQ-7: The rule shall read the branch of a linked worktree from its `.git` pointer file without starting a program.
- REQ-8: The rule shall cover the Write, Edit, MultiEdit and NotebookEdit tools of Claude Code and the `apply_patch` tool of Codex.

실패 경로: REQ-1 → REQ-2, REQ-3(괄호로 연결). REQ-5 → 없음(줄 끝 `failure: none`).

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-5): SPEC이 없으면 `src/`·`tests/`·문서·루트 파일 쓰기가 거부되고 이유 문장에 입력 경로가 없어요.
  `.moai/specs/` 쓰기는 통과하고, SPEC을 쓰면 코드 쓰기가 통과해요.
  verify: `cargo test --locked --offline --test spec_before_code_hook a_code_write_on_a_spec_branch_waits_for_the_branch_spec`
- AC-2 (REQ-2): 다른 id, frontmatter 밖의 `id:` 줄, `spec.md`라는 폴더, 빈 SPEC 폴더는 SPEC으로 치지 않아요.
  verify: `cargo test --locked --offline --test spec_before_code_hook a_spec_of_another_id_or_a_bare_folder_does_not_count`
- AC-3 (REQ-3): `.moai/specs/../../src/…`와 `.moai/specs/` 안의 `src`로 가는 링크가 거부돼요.
  verify: `cargo test --locked --offline --test spec_before_code_hook a_path_that_leaves_the_specs_folder_is_judged_where_it_lands`
- AC-4 (REQ-6): `main`, `issue/…`, 날짜가 아닌·짧은·대문자·끝 하이픈·하위 경로 `spec/` 가지, 작업 트리 밖 쓰기, Bash는 통과해요.
  verify: `cargo test --locked --offline --test spec_before_code_hook other_branches_outside_writes_and_bash_are_left_alone`
- AC-5 (REQ-7, REQ-8): 연결된 worktree의 가지를 읽고, Edit·MultiEdit·Codex `apply_patch`도 같게 거부·통과해요.
  verify: `cargo test --locked --offline --test spec_before_code_hook linked_worktrees_and_every_write_tool_are_covered`
- AC-6 (REQ-1…REQ-8): 기존 hook 시험과 바뀐 곳에 닿는 검사가 통과해요.
  verify: `cargo test --locked --offline --lib hook_guards::`; `cargo test --locked --offline --test hook_dispatch_cli`;
  `cargo test --locked --offline --test hooks_cli`; `cargo test --locked --offline --test hook_mode_cli`;
  `cargo test --locked --offline --test hook_status_cli`; `cargo test --locked --offline --test hook_legacy_cli`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/hook_guards.rs`, `tests/spec_before_code_hook.rs`, `.moai/specs/SPEC-20261117-spec-before-code/spec.md`
- read-only: `src/hook_engine.rs`, `src/hook_config.rs`, `src/spec_doc.rs`, `tests/hook_dispatch_cli.rs`
- forbidden: `src/hook_engine.rs`(다른 가지 commit-guard-users가 바꿈), 사용자 HOME 설정, 호스트 설정 파일

## Plan

1. RED: `tests/spec_before_code_hook.rs` 5개(실제 바이너리, 임시 저장소·HOME). 쓰기가 허용되어 실패하는지 확인.
2. GREEN: `hook_guards::write_guard` 끝에 A5 `spec_before_code`: 가지 → SPEC id, worktree `.git` 파일 읽기, 경로 해석
   (`hook_config::canonical`), SPEC frontmatter의 `id:` 확인.
3. 기존 hook 시험, 검사, 커밋, mutation.
4. 다음 조각(이 SPEC 밖): GREEN 동안 시험 파일 잠금, AC verify 기록 없는 완료 거부, 작성 세션 판정 거부(관문, SDD-6),
   승인 상태 요구(아래 D-3).

## Risks

- SDD 이름을 쓰지만 SPEC을 나중에 쓰던 흐름은 막혀요(의도). 빠져나가는 길은 SPEC 파일을 먼저 쓰는 것이고, 그 쓰기는 늘 허용돼요.
- Bash 리디렉션(`>`, `tee`)이나 hook을 부르지 않는 호스트(OMP의 쓰기 도구 등)는 막지 못해요 → 이 SPEC에 적고, 관문(병합 전
  SPEC 확인)이 마지막 그물이에요.
- 설치된 `moai`가 이 판보다 오래되면 규칙이 없어요 → 다음 설치(빌드) 뒤에 적용돼요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기본으로 켜진 거부 규칙(A5)이고 hook-config의 opt-in이 아니다 | 사용자 지시 "강제"; SPEC 쓰기가 늘 열려 있어 막힌 채로 남지 않음 | `write_guard`의 마지막 줄을 `None`으로 |
| D-2 | `hook_engine.rs`가 아니라 `write_guard` 안에 넣는다 | 다른 가지(commit-guard-users)가 `hook_engine.rs`를 바꿈; 같은 기본 규칙 묶음 | 엔진에서 따로 부르게 옮김 |
| D-3 | SPEC의 `status`는 보지 않는다(draft도 통과) | 지금 SPEC은 draft로 병합되고 승인은 사용자가 모아서 고름(리더 메모 09:1xZ); 승인 상태를 요구할지는 사용자 결정 | 다음 조각에서 `status: approved` 요구 |
| D-4 | 이유 문장은 고정 문자열이고 경로·가지 이름을 싣지 않는다 | A1–A4와 같은 규칙(입력 반복 금지) | — |

## Result

구현 에이전트가 2026-11-17에 채웠어요. SYNC와 AUDIT가 다시 확인해요. 기준 `cbf1c7d959`(origin/v2026.11.0), 변경 커밋
`7a6c7e2abb`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는
`moai heavy-slot-acquire` 슬롯 안에서 한 번에 하나.

- AC-1…AC-5: met — RED(구현 전): exit 101, 4개 FAILED(거부가 와야 할 곳이 허용, "stdout is one JSON object: EOF"),
  `other_branches…`만 통과. GREEN: exit 0, 5 passed.
- AC-6: met — `--lib hook_guards::` 0(17); `hook_dispatch_cli` 0(56); `hooks_cli` 0(22); `hook_mode_cli` 0(5);
  `hook_status_cli` 0(10); `hook_legacy_cli` 0(10); `home_isolation_scan` 0(16); `product_cli` 0(75); `--bin moai` 0(27);
  `ac_command_lint` 0(5); `git_env_scrub` 0(3); `cargo fmt --check` 0(처음 1, 시험 파일 줄바꿈 세 곳 뒤 0); clippy 0;
  Python `ci_scope`·`repo_scans`·`skill_references` OK; `moai spec check --all` 0(처음 `section_missing`: Risks가 있으면
  Non-goals도 필요, 절을 더한 뒤 0); `git diff --check` 0.
- mutation 3개(커밋 뒤, 각각 `git restore` 뒤 tree 깨끗): 아무 `id:` 줄이나 인정 → 1 FAIL(AC-2); `..`·링크를 풀지 않고
  글자로 판단 → 1 FAIL(AC-3); worktree `.git` 파일을 읽지 않음 → 1 FAIL(AC-5). 되돌린 뒤 5 passed.

실행 안 함: 전체 `cargo test`, 실제 Claude Code·Codex 세션에서의 hook 실행(설치된 `moai`가 이 판이 아님), Linux·Windows.
