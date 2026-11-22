---
id: SPEC-20261122-claude-md-project-picks
title: "moai update offers move, import and leave for a project's AGENTS.md and AGENTS.local.md"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Project picks: moving a project's AGENTS files

## Intent

`moai update` 가 프로젝트에 이미 있는 `AGENTS.md`·`AGENTS.local.md` 를 어떻게 새 배치로 옮길지 사용자가 고르게 해요(`SPEC-20261122-claude-md-layout` 의 결정 D-2…D-5). 선택은 셋이고 이 순서로 보여 줘요: `move`(권장, 글을 `CLAUDE.md`·`CLAUDE.local.md` 로 복사해 새 배치 한 곳으로 모아요 — 사용자의 2026-11-22 방향은 moai 프로젝트가 `AGENTS.md` 를 쓰지 않는다는 거예요), `import`(새 `CLAUDE.md` 에 블록과 `@AGENTS.md` 한 줄), `leave`(아무것도 쓰지 않고 경고). 어느 선택에서도 원본은 지우지도 바꾸지도 않아요. `import` 에서 `AGENTS.local.md` 는 아무 일도 없어요(Claude Code 가 원래 읽지 않던 파일이에요). 쓰기 규칙(영수증, 링크·UTF-8 거부, baseline, 미리보기 키, 토큰)은 `SPEC-20261122-claude-md-project-layer` 가 맡아요. 선택마다 명령 하나와 계획 다이제스트 하나를 찍는 `project-update-plan` 의 `choices` 모양을 써요.

## Requirements

- REQ-1: When the project root holds `AGENTS.md` or `AGENTS.local.md` and no pick was given, the update preview shall list the picks `move`, `import` and `leave` in that order, mark `move` as recommended, give one command line each and write nothing.
- REQ-2 (REQ-1): IF the given pick is none of `move`, `import` and `leave`, THEN the updater shall refuse the pick and write nothing.
- REQ-3: When the pick is `move`, the updater shall copy the text of `AGENTS.md` outside its managed block into `CLAUDE.md` below the new managed block and the text the user wrote in `AGENTS.local.md` into `CLAUDE.local.md`. · failure: none — a source that holds no user text yields no copy and no error
- REQ-4: When the pick is `import`, the updater shall write `CLAUDE.md` with the managed block and the single line `@AGENTS.md`. · failure: none — the existing-file case is REQ-6
- REQ-5: When the pick is `leave`, the updater shall write neither `CLAUDE.md` nor `CLAUDE.local.md` and report that Claude Code keeps reading `AGENTS.md`. · failure: none — nothing is written
- REQ-6: When `CLAUDE.md` or `CLAUDE.local.md` already exists, the updater shall add the chosen text on a new line after the last byte of that file and keep every earlier byte. · failure: none — the conflict cases are REQ-5 of the layout SPEC and REQ-3 of the project layer SPEC
- REQ-7: The updater shall never delete or change `AGENTS.md` or `AGENTS.local.md`.
- REQ-8: When the project holds `.claude/CLAUDE.md`, `CLAUDE.md` or `CLAUDE.local.md` that is a regular file without an `@AGENTS.md` import and the project holds `AGENTS.md`, the update preview shall warn that Claude Code reads only the CLAUDE files. · failure: none — a warning only
- REQ-9: When the updater writes `CLAUDE.local.md` and no rule of the project ignores it, the updater shall add the ignore rule in the same apply. · failure: none — the rule is written by the ignore block writer under its own conflict rule
- REQ-10: When `CLAUDE.md` already holds the line `@AGENTS.md` or the text of `AGENTS.md`, the update preview shall offer no pick for it. · failure: none — an already settled project has nothing to offer
- REQ-11: When the preview lists the pick `move`, the preview shall state that `AGENTS.md` and `AGENTS.local.md` stay unchanged and that the user may delete them later. · failure: none — a preview writes nothing

## Acceptance

- AC-1 (REQ-1…REQ-7, REQ-11): in a temporary project with `AGENTS.md` and `AGENTS.local.md` the update preview lists the three picks in the order `move` (recommended), `import`, `leave`, tells under `move` that the originals stay and may be deleted later, and writes nothing; each pick writes exactly its files; both originals stay byte-identical; a user `CLAUDE.md` keeps every earlier byte and the added text starts on a new line even when the file does not end with one. This is the first criterion of the card (its restore half is AC-1 of the project layer SPEC).
  verify: `cargo test --locked --offline --test claude_md_picks_cli picks_`
- AC-2 (REQ-8): the shadow warning appears for a regular `CLAUDE.md`, `.claude/CLAUDE.md` or `CLAUDE.local.md` without the import and not for one that imports `AGENTS.md` or one that is a link.
  verify: `cargo test --locked --offline --test claude_md_picks_cli shadow_`
- AC-3 (REQ-9, REQ-10): the ignore rule for `CLAUDE.local.md` is added in the same apply, and a project that already holds the import or the copied text is offered no pick.
  verify: `cargo test --locked --offline --test claude_md_picks_cli settled_`
- AC-4: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/project_update.rs`, `src/agents.rs`, `src/agents_local_merge.rs`, `src/ops/agents_*.rs`, `tests/claude_md_picks_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-project-picks/**`
- read-only: `src/update.rs`, `src/go_state.rs`, `src/apply.rs`
- forbidden: `AGENTS.md`, `AGENTS.local.md` of this repository, the real HOME, the network, `.claude/` of this checkout

## Plan

1. A pure planner for the picks: reads the project root and returns rows for each pick (path, action, earlier bytes); the three picks are three plans with their own digest.
2. Print them as `choices` like `project-update-plan` does: one command and one digest per pick.
3. Remove `claude_shadow` and the pick `claude-import` from the writer; keep the revert path for old receipts.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A moai-generated `AGENTS.local.md` is not copied. | The family decision D-4. | Copy it too. |
| D-2 | Each pick is one printed command with one plan digest. | The `choices` shape of `project-update-plan` (`src/project_update.rs:771-816`, which prints an `argv` and the plan on stdin). | Ask for the pick in the terminal. |

## Result

(Filled when the work is done.)
