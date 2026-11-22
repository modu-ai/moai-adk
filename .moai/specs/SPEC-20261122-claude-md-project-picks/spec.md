---
id: SPEC-20261122-claude-md-project-picks
title: "moai update offers move, import and leave for a project's AGENTS.md and AGENTS.local.md"
status: draft
revision: 2
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Project picks: moving a project's AGENTS files

## Intent

`moai update` 가 프로젝트에 이미 있는 `AGENTS.md`·`AGENTS.local.md` 를 어떻게 새 배치로 옮길지 사용자가 고르게 해요(`SPEC-20261122-claude-md-layout` 의 결정 D-2…D-5). 선택은 셋이고 이 순서로 보여 줘요: `move`(권장, 글을 `CLAUDE.md`·`CLAUDE.local.md` 로 복사해 새 배치 한 곳으로 모아요 — 사용자의 2026-11-22 방향은 moai 프로젝트가 `AGENTS.md` 를 쓰지 않는다는 거예요), `import`(새 `CLAUDE.md` 에 블록과 `@AGENTS.md` 한 줄), `leave`(아무것도 쓰지 않고 경고). 어느 선택에서도 원본은 지우지도 바꾸지도 않아요. `import` 에서 `AGENTS.local.md` 는 아무 일도 없어요(Claude Code 가 원래 읽지 않던 파일이에요). 쓰기 규칙(영수증, 링크·UTF-8 거부, baseline, 미리보기 키, 토큰)은 `SPEC-20261122-claude-md-project-layer` 가 맡아요. 선택마다 명령 하나와 계획 다이제스트 하나를 찍는 `project-update-plan` 의 `choices` 모양을 써요.

개정 2(2026-11-22, 코드 대조 뒤): 선택은 `project-update-plan` 의 새 단계 `agents-picks` 가 `choices` 로 찍고, 적용은 기존 `agents-merge-*` 명령의 `--pick` 인자 하나가 맡아요(결정 D-3, D-4). 한 선택은 블록 → `AGENTS.md` 글 → `AGENTS.local.md` 글 → ignore 줄을 한 영수증으로 묶는 합성 하나이고, 되돌리기는 거꾸로 풀어요. `update.rs` 의 `project_files` 층은 이 SPEC 이 아니라 `SPEC-20261122-claude-md-project-layer` 의 두 번째 조각이에요. 이전 조각에서 `claude_shadow` 와 선택 `claude-import` 는 이미 쓰기기에서 빠졌어요.

## Requirements

- REQ-1: When the project root holds `AGENTS.md` or `AGENTS.local.md` and no pick was given, the update preview shall list the picks `move`, `import` and `leave` in that order, mark `move` as recommended, give one command line each and write nothing. · failure: none — a project that holds only `AGENTS.local.md` is offered `move` and `leave`, because `import` would name a file that is not there
- REQ-2 (REQ-1): IF the given pick is none of `move`, `import` and `leave`, THEN the updater shall refuse the pick and write nothing.
- REQ-3: When the pick is `move`, the updater shall copy the text of `AGENTS.md` outside its managed block into `CLAUDE.md` below the new managed block and the text the user wrote in `AGENTS.local.md` into `CLAUDE.local.md`. · failure: none — a source that holds no user text yields no copy and no error; the order of the writes is the block, then the text of `AGENTS.md`, then the text of `AGENTS.local.md`
- REQ-4: When the pick is `import`, the updater shall write `CLAUDE.md` with the managed block and the single line `@AGENTS.md`. · failure: none — the existing-file case is REQ-6
- REQ-5: When the pick is `leave`, the updater shall write neither `CLAUDE.md` nor `CLAUDE.local.md` and report that Claude Code keeps reading `AGENTS.md`. · failure: none — nothing is written
- REQ-6: When `CLAUDE.md` or `CLAUDE.local.md` already exists, the updater shall add the chosen text on a new line after the last byte of that file and keep every earlier byte. · failure: none — the conflict cases are REQ-5 of the layout SPEC and REQ-3 of the project layer SPEC
- REQ-7: The updater shall never delete or change `AGENTS.md` or `AGENTS.local.md`.
- REQ-8: When the project holds `.claude/CLAUDE.md`, `CLAUDE.md` or `CLAUDE.local.md` that is a regular file without an `@AGENTS.md` import and the project holds `AGENTS.md`, the update preview shall warn that Claude Code reads only the CLAUDE files. · failure: none — a warning only; the warning is absent once the text of `AGENTS.md` is already in `CLAUDE.md` (REQ-10)
- REQ-9: When the updater writes `CLAUDE.local.md` and no rule of the project ignores it, the updater shall add the ignore rule in the same apply. · failure: none — the rule is one line written through the receipt of the same pick
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

- write: `src/project_update.rs`, `src/agents.rs`, `src/agents_local_merge.rs`, `src/agents_local_merge/picks.rs`, `src/ops/agents_merge_*.rs`, `tests/claude_md_picks_cli.rs`, `tests/project_update_plan_cli.rs`, `skills/moai/workflows/update.md`, `.moai/specs/SPEC-20261122-claude-md-project-picks/**`
- read-only: `src/update.rs`, `src/go_state.rs`, `src/apply.rs`
- forbidden: `AGENTS.md`, `AGENTS.local.md` of this repository, the real HOME, the network, `.claude/` of this checkout

## Plan

1. A pure planner for the picks: reads the project root and returns rows for each pick (path, action, earlier bytes); the three picks are three plans with their own digest.
2. Print them as `choices` like `project-update-plan` does: one command and one digest per pick.
3. (Done in the first slice.) `claude_shadow` and the pick `claude-import` left the writer; the revert path of old receipts stays.
4. `agents_local_merge/picks.rs` holds the composite: `plan` reads the project and returns the offer, the warnings and the three choices; `apply` checks the digest of the pre-state, then runs the writer pick, the text merges and the ignore line, each from its own fresh plan, under one composite receipt; `revert` undoes in reverse order and names a file that changed after the apply while the others still revert.
5. `Kind::Import` in the merge engine writes the line `@AGENTS.md` between two marker lines, so the audited revert and the idempotence check apply to it.
6. `agents-merge-plan`, `agents-merge-apply` and `agents-merge-revert` take `--pick <move|import|leave>`; without it they keep today's behaviour for the personal file.
7. `project_update.rs` gets the step `agents-picks` in the merge phase; the standalone personal-file step waits while a pick is pending.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A moai-generated `AGENTS.local.md` is not copied. | The family decision D-4. | Copy it too. |
| D-2 | Each pick is one printed command with one plan digest. | The `choices` shape of `project-update-plan` (`src/project_update.rs:754-816`, which prints an `argv` and the plan on stdin). | Ask for the pick in the terminal. |
| D-3 | A pick is one composite with one digest of the pre-state and one receipt. | Each engine call digests the bytes the call before it changes, so separate printed commands would go stale one after another. | Print the parts as steps and re-run the preview between them. |
| D-4 | The apply surface is the argument `--pick` of the existing `agents-merge-*` commands. | It stays inside the files the card holds and needs no new help text. | New commands `agents-pick-*`. |
| D-5 | The import line is a `Kind::Import` of the merge engine, between two marker lines. | The audited revert and idempotence code then apply to it. | A plain unmarked line. |
| D-6 | `leave` is not remembered; the preview offers the picks again on the next update. | A project that keeps `AGENTS.md` stays visible instead of hiding a choice in a state file. | Record the pick in the generation manifest. |

## Result

(Filled when the work is done.)
