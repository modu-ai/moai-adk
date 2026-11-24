---
id: SPEC-20261121-claude-md-generation
title: "Generation writes CLAUDE.md and CLAUDE.local.md and never AGENTS.md"
status: draft
revision: 2
created: 2026-11-21
parent: SPEC-20261121-claude-md-canon
---

# Generation writes CLAUDE.md and CLAUDE.local.md and never AGENTS.md

## Intent

The team guidance canon is `CLAUDE.md` and the personal file is the git-ignored `CLAUDE.local.md` (parent SPEC). MoAI's generation still writes `AGENTS.md` and `AGENTS.local.md`. This SPEC makes generation write `CLAUDE.md` (the managed section `moai-workflow`, with the same markers) and `CLAUDE.local.md`, and stop writing `AGENTS.md`, `AGENTS.local.md` and the `claude-import` shadow line.

Claude Code reads `AGENTS.md` only while no `CLAUDE.md`, `.claude/CLAUDE.md` or `CLAUDE.local.md` exists in the working directory or above (memory page, copy of 2026-11-21, lines 355, 360 and 385). The first `CLAUDE.md` or `CLAUDE.local.md` over an `AGENTS.md` therefore hides the text in it. Generation shall never create one over an `AGENTS.md` or `.claude/AGENTS.md` that holds text of its own: the migration of the sibling SPEC (`SPEC-20261121-agents-md-migration`) comes first, with a preview, a backup and a rollback.

## Non-goals

- Changing or deleting any `AGENTS.md` (the migration copies, never writes it).
- Help text, goldens, skills and documents that name the old files (the docs agent's slice).
- Hosts other than Claude Code: MoAI writes no `AGENTS.md` for them (decision 2026-11-21, Claude Code and GLM only).
- Moving the text of a Go-generation project's `AGENTS.md` that `go-state-apply` already put into a backup.

## Requirements

- REQ-1: When generation is planned for a project, the plan shall list `CLAUDE.md` with the managed section `moai-workflow` and `CLAUDE.local.md`, and no `AGENTS.md`, `AGENTS.local.md` or `@AGENTS.md` import · failure: none — a plan writes nothing.
- REQ-2: When generation is applied, the apply shall write the managed section into `CLAUDE.md` and create `CLAUDE.local.md` once.
- REQ-3 (REQ-1): IF `AGENTS.md` or `.claude/AGENTS.md` holds text of its own and none of `CLAUDE.md`, `.claude/CLAUDE.md` and `CLAUDE.local.md` exists, THEN the plan shall mark the `CLAUDE.md` and `CLAUDE.local.md` rows as conflicts whose detail says to migrate first.
- REQ-4 (REQ-1): IF `AGENTS.md` holds only MoAI's own section, THEN the plan shall allow the create and warn that `AGENTS.md` stays as it is.
- REQ-5 (REQ-1): IF `CLAUDE.md` is a link, THEN the plan shall show a conflict row that names the link instead of failing the whole plan.
- REQ-6: When a project was generated before this change, the readers shall accept `AGENTS.local.md` and `CLAUDE.local.md` as the personal file in the manifest and in the ignore block · failure: none — an unknown name is read as before.
- REQ-7: Where a request carries the pick `claude-import`, the plan shall accept it and write nothing for it · failure: none — a pick that does nothing is not an error.
- REQ-8: When the ignore block is planned, the block shall list `CLAUDE.local.md` and keep the `AGENTS.local.md` line of an earlier block · failure: none — a plan only proposes the block.
- REQ-9: When the launcher starts in a project that already has `AGENTS.md`, `CLAUDE.md`, `.claude/CLAUDE.md` or `CLAUDE.local.md`, the launcher shall skip initialization · failure: none — skipping writes nothing.
- REQ-10: When `update` runs on a project that has an `AGENTS.md` with text of its own, the update shall offer the migration as a step before generation · failure: none — an offer changes nothing until the user says yes.
- REQ-11 (REQ-10): IF the user declines the migration, THEN the update shall leave the `CLAUDE.md` row as a conflict and write nothing that hides `AGENTS.md`.
- REQ-12 (REQ-2): IF any row of the plan is a conflict, THEN the apply shall refuse and write nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-7, REQ-8): 임시 프로젝트에서 생성 계획·적용이 `CLAUDE.md`(관리 구역 `moai-workflow`)와 `CLAUDE.local.md` 만 쓰고, `AGENTS.md`·`AGENTS.local.md`·`@AGENTS.md` 줄이 없고, `.gitignore` 블록이 `CLAUDE.local.md` 를 적고, `claude-import` 선택은 받되 아무것도 안 써요 · verify: `cargo test --offline --locked --test claude_md_generation_cli`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-9, REQ-11): 글이 있는 `AGENTS.md` 위에는 `CLAUDE.md`·`CLAUDE.local.md` 를 만들지 않고(conflict, 먼저 이관), 관리 구역뿐이면 경고와 함께 만들고, `CLAUDE.md` 가 링크면 conflict 행이고, 런처는 네 이름 중 하나라도 있으면 건너뛰어요 · verify: `cargo test --offline --locked --test claude_md_generation_cli`
- AC-3 (REQ-6, REQ-10): 이전에 생성된 프로젝트(매니페스트가 `AGENTS.local.md`)의 `update` 계획이 결정 되묻기 없이 깨끗하게 나오고, 글이 있는 `AGENTS.md` 에는 이관 단계가 생성 앞에 있어요 · verify: `cargo test --offline --locked --test project_update_plan_cli` and `--test claude_md_generation_cli`

## Scope

- write: `src/agents.rs`, `src/generation.rs`, `src/launcher_init.rs`, `src/project_update.rs`, `src/agents_local_merge.rs`, `tests/claude_md_generation_cli.rs`, this SPEC folder; the held files are written only after their holders agree on a line range in a pair note
- read-only: `src/apply.rs`, `src/go_state.rs`, `src/claude.rs`
- forbidden: the user's real home and projects, any `AGENTS.md` of a project (never written), any release or deploy action

## Plan

1. B1 — a public judgement in `src/agents_local_merge.rs`: does `AGENTS.md` hold text of its own, only MoAI's section, nothing, or is it blocked (a link, not UTF-8, a credential-shaped line). RED in `tests/agents_migrate_cli.rs`. Writes no held file and changes no output.
2. B2 — readers accept both names: `ignore_has_rule` split in two, `created_once`, `adopt_unchanged` in `src/agents.rs`, and `local_file()` in `src/project_update.rs`. No generated byte changes.
3. B3 — writers flip in `src/agents.rs`, `src/generation.rs`, `src/launcher_init.rs`, with the guard of REQ-3 to REQ-5 and the no-op pick of REQ-7. B3 never lands without B2.
4. B4 — `update` puts the migration step before generation (`src/project_update.rs`). B3 and B4 land together or B4 right after.
5. The test and golden rename wave belongs to B3 and B4 (`product_cli.rs`, `gitignore_block_cli.rs`, `launcher_auto_init.rs`, `generation_*`, fixtures); held tests follow their holders.

## Risks

- A `CLAUDE.md` that is a link to `AGENTS.md` works today because the shadow check skips links; after B3 it would fail the plan with a bare error unless REQ-5 gives it a row.
- Dropping the `AGENTS.local.md` ignore line would un-ignore a private personal file in a project that has one (REQ-8 keeps the line).
- `render_local` bytes are what recognises a generated legacy `AGENTS.local.md`; they must not change.
- `generation_choices` in `src/project_update.rs` and the held choices tests break silently if B3 lands without B4.
- Profiles that still list the `codex` or `omp` hosts read `AGENTS.md`, which MoAI no longer writes; the plan warns about it.
- Many goldens and fixtures change together.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The ignore block keeps the `AGENTS.local.md` line that an earlier block had, and adds none when no earlier block had it (reversible, category none, decided by the leader, note 5638) | Dropping it could un-ignore a private file | Drop the line |
| D-2 | A legacy project that has `AGENTS.local.md` and no `CLAUDE.local.md` gets no new `CLAUDE.local.md` from generation; the personal-file migration creates it (reversible, category none, decided by the leader, note 5638) | Either name counts as "created once" | Create the template |
| D-3 | `update` migrates first, then generation appends the managed section (reversible, category none, decided by the leader, note 5638) | There is never a window where `AGENTS.md` is hidden | Reverse the order |
| D-4 | The 32 KiB Codex warning is dropped (reversible, category none, decided by the leader, note 5638) | Codex support was withdrawn on 2026-11-21 | Restore the text |
| D-5 | No `AGENTS.md` is generated for non-Claude hosts; the plan warns (reversible, category none, decided by the leader, note 5638) | Claude Code and GLM only | Generate it for those hosts |
| D-6 | A link as `CLAUDE.md` is a conflict row (reversible, category none, decided by the leader, note 5638) | The migration does not replace links | Offer to replace it |

## Result

pending
