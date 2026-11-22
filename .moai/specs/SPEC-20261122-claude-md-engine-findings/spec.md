---
id: SPEC-20261122-claude-md-engine-findings
title: "The AGENTS.md merge engine and the first CLAUDE file warn about hidden files, scan only copied text and keep receipts out of git"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Engine findings: four open audit findings of the AGENTS.md merge engine

## Intent

예전 감사가 AGENTS.md 이동 엔진에 남긴 네 가지 지적 중 지금도 열려 있는 것을 닫아요(코드 대조 결과, 2026-11-22). (1) 첫 CLAUDE 파일을 만들면 Claude Code 는 `AGENTS.md` 와 `.claude/AGENTS.md` 를 더는 읽지 않는데, 이 경고가 개인 파일 병합에만 있고 블록 작성기·선택·미리보기에는 없어요(`.claude/AGENTS.md` 만 있는 프로젝트는 아무 경고도 못 받아요). (2) 관리 영역을 빼고 복사할 때 그 둘레의 빈 줄이 한 줄로 합쳐지는데 미리보기가 말해 주지 않아요. (3) 비밀 검사가 복사하지 않는 관리 영역까지 파일 전체를 읽어서, 복사되지 않는 글 때문에 계획이 막힐 수 있어요. (4) 영수증 폴더를 `.gitignore` 에 올리는 줄이 관리 블록에만 있어서, 관리 블록이 아직 없는 프로젝트의 단독 병합은 영수증 폴더를 git 에 노출해요. 하위 폴더의 `AGENTS.md` 는 이 SPEC 의 범위가 아니에요(따로 결정).

## Requirements

- REQ-1: When the writer, a pick or the update preview would create the first CLAUDE file in a project that holds `AGENTS.md` or `.claude/AGENTS.md`, the plan shall warn that Claude Code stops reading each of those files and name the files it would stop reading. · failure: none — the warning needs the same confirmation as today and nothing is written without it
- REQ-2: When the guidance merge removes a managed region from the text it copies, the plan shall tell the user that the blank lines around the removed region are joined into one blank line. · failure: none — the notice is a warning and the copy is unchanged
- REQ-3: The guidance merge shall scan for credential-shaped lines only the text it copies and report line numbers of the whole file. · failure: a credential-shaped line in the copied text blocks the plan as today
- REQ-4: When a merge or a pick writes a receipt folder and no managed `.gitignore` block holds the receipts line, the apply shall add the receipts line to `.gitignore` in the same apply. · failure: none — a folder that is already ignored gets no second line and a project that is not a git repository gets no `.gitignore`

## Acceptance

- AC-1 (REQ-1): a project that holds only `.claude/AGENTS.md`, one that holds only `AGENTS.md` and one that holds both are each warned by the writer, the pick plan and the update preview, naming every file that would stop being read, and a project that already holds a CLAUDE file is not warned.
  verify: `cargo test --locked --offline --test claude_md_engine_cli hidden_`
- AC-2 (REQ-2): the guidance merge plan of an `AGENTS.md` that holds a managed region says that the blank lines around it are joined, and says nothing of the kind when there is no region.
  verify: `cargo test --locked --offline --test claude_md_engine_cli collapse_`
- AC-3 (REQ-3): a credential-shaped line inside the managed region or inside an earlier merged personal block does not block the plan, one in the copied text still blocks it, and the reported line number counts from the first line of the file.
  verify: `cargo test --locked --offline --test claude_md_engine_cli scan_`
- AC-4 (REQ-4): after a standalone merge and after a pick in a git project without a managed `.gitignore` block, `git check-ignore` says the receipts folder is ignored; a project that already ignores it gets no second line, and a project that is not a git repository gets no `.gitignore`.
  verify: `cargo test --locked --offline --test claude_md_engine_cli receipts_`
- AC-5: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/agents.rs`, `src/agents_local_merge.rs`, `src/agents_local_merge/picks.rs`, `src/project_update.rs`, `tests/claude_md_engine_cli.rs`, `tests/agents_local_merge_cli.rs`, `tests/agents_migrate_cli.rs`, `tests/claude_md_picks_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-engine-findings/**`
- read-only: `src/apply.rs`, `src/claude.rs`, `src/generation.rs`
- forbidden: `AGENTS.md`, `AGENTS.local.md` of this repository, the real HOME, the network, `.claude/` of this checkout

## Plan

1. Tests first, in the new file `tests/claude_md_engine_cli.rs`, names starting `hidden_`, `collapse_`, `scan_` and `receipts_`. They run the built binary in a temporary git repository with an empty temporary HOME, like `tests/claude_md_picks_cli.rs`, and walk no repository tree.
2. REQ-1: one helper in `src/agents_local_merge.rs` names the hidden files for a project (root `AGENTS.md`, `.claude/AGENTS.md`) when no CLAUDE file exists; the writer (`src/agents.rs`), the pick plan (`src/agents_local_merge/picks.rs`) and the preview (`src/project_update.rs`) call it. The warning text of the personal-file merge stays as it is.
3. REQ-2: the guidance plan adds one warning line when the removed region leaves text on both sides.
4. REQ-3: the plan scans `text` for every kind and maps each line number back to the whole file; the dead guard `agents_md_state` follows the same rule.
5. REQ-4: the receipts line joins the ignore plan of a merge and of a pick when the managed block is absent; the pin in `tests/agents_local_merge_cli.rs` that expects `.gitignore` to hold exactly one line changes with it.

## Decisions

| Id | Decision | Basis | Undo |
|---|---|---|---|
| D-1 | A project that holds only `.claude/AGENTS.md` gets the warning but no pick. | The picks move and import the root files; moving the second file is a separate feature with the subfolder question. | Offer picks for that file. |
| D-2 | The collapse is told, not prevented. | The copy keeps the text; only blank lines change, and the original file is never touched. | Keep the blank lines. |
| D-3 | The new tests live in one new file with four prefixes. | One verify filter per criterion without touching the older files more than the one pinned line. | Spread the tests. |

## Result

(Filled when the work is done.)
