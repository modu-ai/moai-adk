---
id: SPEC-20261122-claude-md-project-picks
title: "moai update offers move, import and leave for a project's AGENTS.md and AGENTS.local.md"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Project layer: moving a project's AGENTS files

## Intent

`moai update` 가 프로젝트에 이미 있는 `AGENTS.md`·`AGENTS.local.md` 를 어떻게 새 배치로 옮길지 사용자가 고르게 해요(`SPEC-20261122-claude-md-layout` 의 결정 D-2…D-5, D-10). 선택은 셋이에요: `import`(권장, 새 `CLAUDE.md` 에 블록과 `@AGENTS.md` 한 줄), `move`(글을 `CLAUDE.md`·`CLAUDE.local.md` 로 복사), `leave`(아무것도 쓰지 않고 경고). 어느 선택에서도 원본은 지우지도 바꾸지도 않고, 영수증이 새 파일을 되돌려요. 이 레이어는 `git_hooks` 처럼 자기 확인 토큰을 가지고, 할 일이 있을 때만 미리보기에 `project_files` 키로 나와요(고정 키 13개는 그대로예요). 선택마다 명령 하나와 계획 다이제스트 하나를 찍는 `project-update-plan` 의 `choices` 모양을 그대로 써요.

## Requirements

- REQ-1: When the project root holds `AGENTS.md` or `AGENTS.local.md` and no pick was given, the update preview shall list the picks `import`, `move` and `leave` with one command line each and write nothing.
- REQ-2 (REQ-1): IF the given pick is none of `import`, `move` and `leave`, THEN the updater shall refuse the pick and write nothing.
- REQ-3: When the pick is `move`, the updater shall copy the text of `AGENTS.md` outside its managed block into `CLAUDE.md` below the new managed block and the text the user wrote in `AGENTS.local.md` into `CLAUDE.local.md`. · failure: none — a source that holds no user text yields no copy and no error
- REQ-4: When the pick is `import`, the updater shall write `CLAUDE.md` with the managed block and the single line `@AGENTS.md`. · failure: none — the existing-file case is REQ-6
- REQ-5: When the pick is `leave`, the updater shall write neither `CLAUDE.md` nor `CLAUDE.local.md` and report that Claude Code keeps reading `AGENTS.md`. · failure: none — nothing is written
- REQ-6: When `CLAUDE.md` or `CLAUDE.local.md` already exists, the updater shall add the chosen text after the last byte of that file and keep every earlier byte. · failure: none — the conflict case of an edited block is REQ-5 of the layout SPEC
- REQ-7: The updater shall never delete or change `AGENTS.md` or `AGENTS.local.md`.
- REQ-8: When the updater writes a file, the receipt shall record the earlier bytes or the absence of that file so that the revert command restores them.
- REQ-9 (REQ-8): IF a written file changed after the apply, THEN the revert command shall leave that file untouched and report it by name.
- REQ-10: When the project holds a `CLAUDE.md` without an `@AGENTS.md` import and an `AGENTS.md`, the update preview shall warn that Claude Code reads only the CLAUDE files. · failure: none — a warning only
- REQ-11: The layer `project_files` shall appear in the update preview only when it has something to do and in the revert order before every other layer.
- REQ-12: The confirm token of the layer shall be bound to one pick so that a token of another pick or of a changed project is refused.

## Acceptance

- AC-1 (REQ-1…REQ-9): in a temporary project with `AGENTS.md` and `AGENTS.local.md` the update preview lists the three picks and writes nothing; each pick writes exactly its files; both originals stay byte-identical; the revert command restores the project to its first bytes; a user `CLAUDE.md` keeps every earlier byte; a file changed after the apply is skipped by the revert. This is the first criterion of the card.
  verify: `cargo test --locked --offline --test claude_md_picks_cli picks_`
- AC-2 (REQ-10): the shadow warning appears for a `CLAUDE.md` without the import and not for one that imports `AGENTS.md`.
  verify: `cargo test --locked --offline --test claude_md_picks_cli shadow_`
- AC-3 (REQ-11, REQ-12): the preview keeps its 13 fixed keys without work, adds `project_files` only with work, refuses a token of another pick, and reverts the layer first.
  verify: `cargo test --locked --offline --test update_user_layer_cli`
  verify: `cargo test --locked --offline --test claude_md_picks_cli layer_`
- AC-4: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/update.rs`, `src/project_update.rs`, `src/agents.rs`, `src/agents_local_merge.rs`, `src/ops/agents_*.rs`, `tests/claude_md_picks_cli.rs`, `tests/update_user_layer_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-project-picks/**`
- read-only: `src/go_state.rs`, `src/apply.rs`
- forbidden: `AGENTS.md`, `AGENTS.local.md` of this repository, the real HOME, the network, `.claude/` of this checkout

## Plan

1. A pure planner for the project layer: reads the project root and returns rows for each pick (path, action, earlier bytes); the three picks are three plans with their own digest.
2. The apply and revert pair uses the receipt shape of `agents_local_merge` (before copies, after hashes) and the `publish_exclusive` and `publish_replace` writers of `src/agents.rs`.
3. Wire the layer into `update.rs` after `git_hooks`: own token, conditional key, first in the revert order, an entry in `layer_asks`.
4. Remove `claude_shadow` and the pick `claude-import` from the writer; keep the revert path for old receipts.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A moai-generated `AGENTS.local.md` is not copied. | The family decision D-4. | Copy it too. |
| D-2 | The pick id travels as a named value of the printed command, one command per pick. | The `choices` shape of `project-update-plan` (`src/project_update.rs:771-816`). | Ask for the pick in the terminal. |

## Result

(Filled when the work is done.)
