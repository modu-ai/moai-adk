---
id: SPEC-20261122-claude-md-project-picks
title: "moai update offers move and leave for a project's AGENTS.md and AGENTS.local.md, and removing the originals as its own pick"
status: draft
revision: 4
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Project picks: moving a project's AGENTS files

## Intent

`moai update` 가 프로젝트에 이미 있는 `AGENTS.md`·`AGENTS.local.md` 를 어떻게 새 배치로 옮길지 사용자가 고르게 해요(`SPEC-20261122-claude-md-layout` 의 결정 D-2…D-5). 쓰기 규칙(영수증, 링크·UTF-8 거부, baseline, 미리보기 키, 토큰)은 `SPEC-20261122-claude-md-project-layer` 가 맡아요. 선택마다 명령 하나와 계획 다이제스트 하나를 찍는 `project-update-plan` 의 `choices` 모양을 써요.

개정 2(2026-11-22, 코드 대조 뒤): 선택은 `project-update-plan` 의 새 단계 `agents-picks` 가 `choices` 로 찍고, 적용은 기존 `agents-merge-*` 명령의 `--pick` 인자 하나가 맡아요(결정 D-3, D-4). 한 선택은 블록 → `AGENTS.md` 글 → `AGENTS.local.md` 글 → ignore 줄을 한 영수증으로 묶는 합성 하나이고, 되돌리기는 거꾸로 풀어요. `update.rs` 의 `project_files` 층은 이 SPEC 이 아니라 `SPEC-20261122-claude-md-project-layer` 의 두 번째 조각이에요. 이전 조각에서 `claude_shadow` 와 선택 `claude-import` 는 이미 쓰기기에서 빠졌어요.

개정 4(2026-11-22, 사용자 결정 'AGENTS.md 제거해라. CLAUDE.md CLAUDE.local.md 만 사용한다', 범위는 제품과 이 저장소): moai 는 `CLAUDE.md` 와 `CLAUDE.local.md` 만 써요. 선택은 둘이고 이 순서로 보여 줘요: `move`(권장, `AGENTS.md` 의 글을 `CLAUDE.md` 로, `AGENTS.local.md` 의 글을 `CLAUDE.local.md` 로 복사)와 `leave`(moai 는 그 파일을 건드리지도 읽지도 않고, CLAUDE 파일도 쓰지 않아요). `import` 선택(`@AGENTS.md` 한 줄)은 없어지고, 주면 모르는 선택으로 거절해요. `move` 가 끝난 뒤 복사가 확인되면, 원본을 지우는 일은 따로 고르는 선택 `remove-originals` 예요. 그 선택 없이는 사용자 파일을 지우지 않고, 지운 바이트는 영수증에 남아 되돌릴 수 있어요.

## Requirements

- REQ-1: When the project root holds `AGENTS.md` or `AGENTS.local.md` and no pick was given, the update preview shall list the picks `move` and `leave` in that order, mark `move` as recommended, give one command line each and write nothing. · failure: none — the retired pick `import` is never listed
- REQ-2 (REQ-1): IF the given pick is none of `move`, `leave` and `remove-originals`, THEN the updater shall refuse the pick and write nothing. · failure: `import` is refused as an unknown pick with the same refusal
- REQ-3: When the pick is `move`, the updater shall copy the text of `AGENTS.md` outside its managed block into `CLAUDE.md` below the new managed block and the text the user wrote in `AGENTS.local.md` into `CLAUDE.local.md`. · failure: none — a source that holds no user text yields no copy and no error; the order of the writes is the block, then the text of `AGENTS.md`, then the text of `AGENTS.local.md`
- REQ-4: When the pick is `remove-originals` and the copied text of each original is found in its CLAUDE file, the updater shall delete those originals and keep their bytes in the receipt. · failure: none — an original whose text is not found in its CLAUDE file, or that changed since the preview, is refused and nothing is deleted
- REQ-5: When the pick is `leave`, the updater shall write neither `CLAUDE.md` nor `CLAUDE.local.md` and report that Claude Code keeps reading `AGENTS.md`. · failure: none — nothing is written
- REQ-6: When `CLAUDE.md` or `CLAUDE.local.md` already exists, the updater shall add the chosen text on a new line after the last byte of that file and keep every earlier byte. · failure: none — the conflict cases are REQ-5 of the layout SPEC and REQ-3 of the project layer SPEC
- REQ-7: The updater shall write neither `AGENTS.md` nor an `@AGENTS.md` line and leave `AGENTS.md` and `AGENTS.local.md` untouched outside the pick `remove-originals`.
- REQ-8: When the project holds `.claude/CLAUDE.md`, `CLAUDE.md` or `CLAUDE.local.md` that is a regular file without an `@AGENTS.md` import and the project holds `AGENTS.md`, the update preview shall warn that Claude Code reads only the CLAUDE files. · failure: none — a warning only; the warning is absent once the text of `AGENTS.md` is already in `CLAUDE.md` (REQ-10)
- REQ-9: When the updater writes `CLAUDE.local.md` and no rule of the project ignores it, the updater shall add the ignore rule in the same apply. · failure: none — the rule is one line written through the receipt of the same pick
- REQ-10: When `CLAUDE.md` already holds the line `@AGENTS.md` or the text of `AGENTS.md`, the update preview shall offer no `move` for it and offer `remove-originals` only when the text of each original is found in its CLAUDE file. · failure: none — an already settled project has nothing else to offer
- REQ-11: When the preview lists the pick `move`, the preview shall state that `AGENTS.md` and `AGENTS.local.md` stay unchanged unless the separate pick `remove-originals` is chosen after the move. · failure: none — a preview writes nothing
- REQ-12 (REQ-3): IF a part of the pick fails after another part wrote, THEN the updater shall undo the parts written before it through the same receipt and name any file it could not undo. · failure: the receipt says `partial` and names what is left; `agents-merge-revert` undoes the rest, which for `remove-originals` restores the deleted bytes

## Acceptance

- AC-1 (REQ-1…REQ-3, REQ-5…REQ-7, REQ-11): in a temporary project with `AGENTS.md` and `AGENTS.local.md` the update preview lists the two picks in the order `move` (recommended), `leave`, tells under `move` that the originals stay unless `remove-originals` is chosen, and writes nothing; `import` is refused as an unknown pick; each pick writes exactly its files; both originals stay byte-identical; no pick writes `AGENTS.md` or an `@AGENTS.md` line; a user `CLAUDE.md` keeps every earlier byte and the added text starts on a new line even when the file does not end with one.
  verify: `cargo test --locked --offline --test claude_md_picks_cli picks_`
- AC-2 (REQ-8): the shadow warning appears for a regular `CLAUDE.md`, `.claude/CLAUDE.md` or `CLAUDE.local.md` without the import and not for one that imports `AGENTS.md` or one that is a link.
  verify: `cargo test --locked --offline --test claude_md_picks_cli shadow_`
- AC-3 (REQ-9, REQ-10): the ignore rule for `CLAUDE.local.md` is added in the same apply, a project that already holds the import or the copied text is offered no `move`, and a project whose originals are already copied is offered `remove-originals`.
  verify: `cargo test --locked --offline --test claude_md_picks_cli settled_`
- AC-5 (REQ-12): a failure after a part wrote rolls that part back through the same receipt, and a part that cannot be undone is named.
  verify: `cargo test --locked --offline --lib agents_local_merge::picks::tests`
- AC-6 (REQ-1): when only `AGENTS.local.md` exists and `move` is refused, the step reports a conflict and offers only `leave`.
  verify: `cargo test --locked --offline --test claude_md_picks_cli refuse_only_agents_local_md`
- AC-7 (REQ-4, REQ-12): after `move`, the pick `remove-originals` deletes both originals and keeps their bytes in the receipt, its revert restores them byte-identical, and an original whose text is not in its CLAUDE file or changed since the preview is refused with nothing deleted.
  verify: `cargo test --locked --offline --test claude_md_picks_cli remove_`
- AC-4: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/project_update.rs`, `src/agents.rs`, `src/agents_local_merge.rs`, `src/agents_local_merge/picks.rs`, `src/ops/agents_merge_*.rs`, `tests/claude_md_picks_cli.rs`, `tests/project_update_plan_cli.rs`, `skills/moai/workflows/update.md`, `.moai/specs/SPEC-20261122-claude-md-project-picks/**`
- read-only: `src/update.rs`, `src/go_state.rs`, `src/apply.rs`
- forbidden: the real HOME, the network, `.claude/` of this checkout; this repository's own `AGENTS.md` and `AGENTS.local.md` belong to the repository slice of the layout SPEC (Plan 5), not to this one

## Plan

1. A pure planner for the picks: reads the project root and returns rows for each pick (path, action, earlier bytes); each pick is a plan with its own digest.
2. Print them as `choices` like `project-update-plan` does: one command and one digest per pick.
3. (Done in the first slice.) `claude_shadow` and the pick `claude-import` left the writer; an old receipt's rows revert through the generic apply recovery (no code reads the pick id).
4. `agents_local_merge/picks.rs` holds the composite: `plan` reads the project and returns the offer, the warnings and the choices; `apply` checks the digest of the pre-state, then runs the writer pick, the text merges and the ignore line, each from its own fresh plan, under one composite receipt; `revert` undoes in reverse order and names a file that changed after the apply while the others still revert.
5. (Revision 4.) `Kind::Import` and the pick `import` are removed from the picks; an old receipt that holds an import row still reverts through the generic recovery.
6. `agents-merge-plan`, `agents-merge-apply` and `agents-merge-revert` take `--pick <move|leave|remove-originals>`; without it they keep today's behaviour for the personal file.
7. `project_update.rs` gets the step `agents-picks` in the merge phase; the standalone personal-file step waits while a pick is pending.
8. (Revision 4.) `remove-originals` is its own composite: the plan checks that the copied text of each original is in its CLAUDE file and digests the originals' bytes; the apply deletes them with their bytes in the receipt; the revert writes them back.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A moai-generated `AGENTS.local.md` is not copied. | The family decision D-4. | Copy it too. |
| D-2 | Each pick is one printed command with one plan digest. | The `choices` shape of `project-update-plan` (`src/project_update.rs:754-816`, which prints an `argv` and the plan on stdin). | Ask for the pick in the terminal. |
| D-3 | A pick is one composite with one digest of the pre-state and one receipt. | Each engine call digests the bytes the call before it changes, so separate printed commands would go stale one after another. | Print the parts as steps and re-run the preview between them. |
| D-4 | The apply surface is the argument `--pick` of the existing `agents-merge-*` commands. | It stays inside the files the card holds and needs no new help text. | New commands `agents-pick-*`. |
| D-5 | (Retired in revision 4.) The import line was a `Kind::Import` of the merge engine. | The user's decision of 2026-11-22 removes the pick `import`. | Bring the pick back. |
| D-6 | `leave` is not remembered; the preview offers the picks again on the next update. | A project that keeps `AGENTS.md` stays visible instead of hiding a choice in a state file. | Record the pick in the generation manifest. |
| D-7 | The picks are `move` and `leave`; `import` is refused as unknown; moai never writes `AGENTS.md` or an `@AGENTS.md` line. | The user's decision of 2026-11-22 on the card ('AGENTS.md 제거해라. CLAUDE.md CLAUDE.local.md 만 사용한다'). | Bring `import` back as a third pick. |
| D-8 | Removing the originals is the separate pick `remove-originals`, offered only once the copied text is found, never the default and never part of `move`. | The same decision: no user file is deleted without that pick, and only after the move is verified. | Make removal part of `move`. |
| D-9 | `remove-originals` keeps the deleted bytes in its receipt so `agents-merge-revert` restores them. | A deletion the user can undo through the same receipt surface as every other pick. | Delete without a stored copy. |

## Result

(Filled when the work is done.)
