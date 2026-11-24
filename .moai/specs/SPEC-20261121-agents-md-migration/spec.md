---
id: SPEC-20261121-agents-md-migration
title: "An existing AGENTS.md moves into CLAUDE.md with preview, backup and rollback"
status: draft
revision: 1
created: 2026-11-21
parent: SPEC-20261121-claude-md-canon
---

# An existing AGENTS.md moves into CLAUDE.md with preview, backup and rollback

## Intent

The team guidance canon is `CLAUDE.md` (parent SPEC). Many projects already have an `AGENTS.md` that people wrote and that Claude Code reads today. The memory page says Claude reads `AGENTS.md` only while no `CLAUDE.md`, `.claude/CLAUDE.md` or `CLAUDE.local.md` exists (lines 355, 360, 385 of the copy read on 2026-11-21). The first `CLAUDE.md` or `CLAUDE.local.md` therefore removes `AGENTS.md` from Claude's context, and the text in it has to be carried over, with the user's yes, before or together with that first file.

This SPEC adds a second kind to the merge engine of `src/agents_local_merge.rs` (the first kind moves `AGENTS.local.md` into `CLAUDE.local.md`): the text a person wrote in `AGENTS.md` goes into `CLAUDE.md`, once, between two marker lines, with a preview, a receipt that keeps the earlier bytes, and a revert. `AGENTS.md` is never changed or deleted; other tools and the history still read it.

What counts as the person's text: everything in `AGENTS.md` outside MoAI's own managed section (the block between `<!-- moai:managed:begin id=moai-workflow -->` and its end marker) and outside any block this engine wrote earlier. A project whose `AGENTS.md` has no managed section is entirely the person's text. MoAI's own section is not copied: the generation of the parent SPEC writes the managed section into `CLAUDE.md` itself.

## Non-goals

- Changing or deleting `AGENTS.md`.
- Writing the managed section into `CLAUDE.md` (the generation slice of the parent SPEC).
- Wiring new `moai` commands: `src/help.rs` and the help goldens are held elsewhere. The functions are tested as a library API now; the commands follow when those files are free.
- A `CLAUDE.md` that already imports `AGENTS.md` with an `@AGENTS.md` line: it is left as it is (REQ-6).
- Telling a hand edit inside MoAI's managed section from MoAI's own text: the section is not copied at all, and the preview names that it is left out.

## Requirements

- REQ-1: When the migration is previewed, the plan shall show the text of `AGENTS.md` that would go into `CLAUDE.md`, its position there, and that `CLAUDE.md` is a committed team file · failure: none — a preview writes nothing.
- REQ-2: When the migration is applied after a confirmed preview, the apply shall create `CLAUDE.md` when it is missing or add one marked block to it, and keep a receipt with the before bytes and a copy of the source · failure: the receipt cannot be written, so nothing is written to the target.
- REQ-3: When the migration runs again over an unchanged source, the plan shall be `unchanged` and write nothing · failure: none — an unchanged plan needs no confirmation.
- REQ-4: When the migration is reverted from its receipt, the revert shall restore the earlier bytes of `CLAUDE.md` or remove the file the apply created, only while the file still holds the bytes the apply wrote.
- REQ-5: Where `AGENTS.md` holds no MoAI managed section, the plan shall take the whole file as the person's text · failure: none — an `AGENTS.md` with only MoAI's section, or with MoAI's section and the one `# title` line MoAI writes above it, has no text to move and the plan says so.
- REQ-6: Where `CLAUDE.md` already imports `AGENTS.md` with an `@AGENTS.md` line, the plan shall be `unchanged` and say that the text is already read through the import · failure: none — the import is the user's choice and stays.
- REQ-7 (REQ-2): IF the target is a link or cannot be read, THEN the migration shall block the plan and write nothing.
- REQ-8 (REQ-2): IF the plan differs from the confirmed preview at apply time, THEN the apply shall write nothing and report a stale plan.
- REQ-9 (REQ-4): IF the file changed since the apply, THEN the revert shall write nothing and say so.
- REQ-10 (REQ-1): IF a line of the text is shaped like a credential, THEN the migration shall block the plan without printing the value.
- REQ-11 (REQ-1): IF the text holds a MoAI marker line, THEN the migration shall block the plan, because the copy would be read as a second managed block.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5, REQ-6): 임시 프로젝트에서 `AGENTS.md` 의 사람이 쓴 글(MoAI 관리 구역과 이전에 옮긴 블록은 뺀)이 `CLAUDE.md` 로 가는 미리보기가 글·위치·커밋되는 팀 파일임을 보여 주고, 관리 구역이 없으면 파일 전체가 글이며, 다시 하면 `unchanged`, `@AGENTS.md` 를 가진 `CLAUDE.md` 는 `unchanged` 예요 · verify: `cargo test --offline --locked --test agents_migrate_cli`
- AC-2 (REQ-2, REQ-4, REQ-7, REQ-8, REQ-9): 적용은 없던 `CLAUDE.md` 를 만들거나 기존 것 끝에 블록 하나를 덧붙이고 영수증에 적용 전 바이트와 원본 사본을 남기며, 되돌리기는 만든 파일을 지우고 뒤에 고친 파일은 건드리지 않아요 · verify: `cargo test --offline --locked --test agents_migrate_cli`
- AC-3 (REQ-10, REQ-11): 비밀처럼 생긴 줄이나 MoAI 표시 줄이 든 글은 계획이 막히고 값은 출력되지 않아요 · verify: `cargo test --offline --locked --test agents_migrate_cli`
- AC-4 (REQ-2): `AGENTS.md` 는 적용 뒤에도 한 바이트도 바뀌지 않아요 · verify: `cargo test --offline --locked --test agents_migrate_cli`

## Scope

- write: `src/agents_local_merge.rs`, `tests/agents_migrate_cli.rs`, this SPEC folder; `src/lib.rs` is not needed (the engine already lives in the crate) and the `moai` commands are a follow-up
- read-only: `src/agents.rs`, `src/apply.rs`, `src/go_state.rs`
- forbidden: the user's real home and projects, `AGENTS.md` of any project (never written), `src/project_update.rs` (held), any release or deploy action

## Plan

1. RED: `tests/agents_migrate_cli.rs` first, through a library API (`moai_cli::agents_local_merge::plan_for` and its apply and revert), because no command exists yet.
2. GREEN: a `Kind` for the engine (`Local`: `AGENTS.local.md` into `CLAUDE.local.md`, `Guidance`: `AGENTS.md` into `CLAUDE.md`) that carries the source name, the target name, the marker names and the receipt prefix; the existing public functions keep the first kind.
3. The text of the person: the part of `AGENTS.md` outside MoAI's managed section and outside blocks this engine wrote.
4. Mutants on the new code, then `check affected` on the merge.

## Risks

- One engine with two kinds touches code the first kind uses; the 21 tests of `agents_local_merge_cli` and the module tests run again.
- A hand edit inside MoAI's managed section of `AGENTS.md` is not copied, so a person who wrote text inside that block loses it from `CLAUDE.md`. The preview carries a warning that names the section and the number of lines left out, and `AGENTS.md` itself stays unchanged.
- Without commands a user cannot run the migration from the terminal until the follow-up lands; the library API and the receipt are the same ones the commands will call.
- A project that already has a `CLAUDE.md` gets a block at its end; a person may want it placed differently. The text is theirs to move afterwards, and the revert undoes the whole block.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | MoAI's managed section is left out of the copy (reversible, category none) | The generation writes it into `CLAUDE.md`; copying it would double it | Copy it and strip the markers |
| D-2 | A `CLAUDE.md` with an `@AGENTS.md` line is `unchanged` (reversible, category none) | The import already carries the text; adding it again would double it | Offer to replace the import |
| D-3 | The text goes to the end of `CLAUDE.md` in one marked block (reversible, category none) | Same shape as the personal-file kind; the revert removes it whole | Let the person choose a position |
| D-4 | A credential-shaped line blocks the plan (reversible, category none) | The text goes into a committed team file | Change to a warning |

## Result

What was built: `src/agents_local_merge.rs` gained a public `Kind` (`Local`, `Guidance`) and `plan_for`, `apply_for` and `revert_for`; the earlier `plan`, `apply` and `revert` call them with `Local`. The guidance kind reads `AGENTS.md`, takes out MoAI's managed section and the block an earlier personal-file merge wrote, and moves the rest into `CLAUDE.md` between `moai:merged-agents` markers. A project with only MoAI's section, or that section plus the one `# title` line MoAI writes, has no text to move. A `CLAUDE.md` that imports `AGENTS.md` (the word `@AGENTS.md` or `@./AGENTS.md` outside a code fence and outside backticks) is `unchanged`. The receipt id starts `agents-md-migrate-`, and a receipt id of one kind is unknown to the other. No `moai` command carries the kind yet (see follow-ups).

Evidence (commands run in the SPEC's own worktree, build folder `/Users/goos/.moai/build/moai-adk-c306`):
- RED: `cargo test --offline --locked --test agents_migrate_cli` before any implementation failed to compile on the four missing names (`target/red2.log`).
- GREEN: `--test agents_migrate_cli` 9 passed and `--test agents_local_merge_cli` 21 passed (`target/green1.log`); `--lib agents_local_merge::` 18 passed (`target/green2.log`).
- Mutants (`target/mutants2.py`, `target/mut2.log`), all killed: a join that drops a blank line, a fence that never holds code, guidance receipts with the personal prefix, the import check skipped, the earlier personal block not taken out.

AC-1 met: `the_persons_text_moves_into_a_new_claude_md_once`, `a_file_without_a_managed_section_is_all_the_persons_and_one_with_only_it_is_nothing`, `a_claude_md_that_imports_agents_md_is_left_alone` (`agents_migrate_cli`).
AC-2 met: `an_existing_claude_md_gets_a_block_and_the_revert_restores_it`, `the_revert_removes_a_created_claude_md`, `a_link_blocks_and_a_stale_digest_writes_nothing`.
AC-3 met: `credentials_and_markers_block_the_plan`, `an_earlier_personal_block_is_left_out`.
AC-4 met: `the_persons_text_moves_into_a_new_claude_md_once` compares `AGENTS.md` byte for byte after the apply, and the revert test again after the revert.

Follow-ups outside this change: the `agents-migrate-plan|apply|revert` commands and their help text (`src/help.rs`, the help goldens, `src/ops/`), `/moai update` offering the migration before it creates `CLAUDE.md` (`src/project_update.rs`), and the generation writing MoAI's own section into `CLAUDE.md` (parent SPEC, `src/agents.rs`, `src/generation.rs`).
