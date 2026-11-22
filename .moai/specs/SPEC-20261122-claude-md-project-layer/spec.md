---
id: SPEC-20261122-claude-md-project-layer
title: "The project layer of moai update writes CLAUDE files safely: receipt, refusals, baseline, key and token"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Project layer: how it writes

## Intent

`SPEC-20261122-claude-md-project-picks` 가 무엇을 쓸지 정하면, 이 SPEC 은 어떻게 안전하게 쓰는지를 정해요. 쓰기마다 영수증이 앞선 바이트를 남기고, 적용 뒤에 바뀐 파일은 되돌리기가 건드리지 않고 이름으로 알려요. 프로젝트 안의 `CLAUDE.md`·`CLAUDE.local.md` 나 그 상위 폴더가 심볼릭 링크이거나 UTF-8 이 아니면 쓰지 않아요(`CLAUDE.md -> AGENTS.md` 링크를 통해 사용자의 `AGENTS.md` 를 고치는 일을 막아요; 지금 코드는 `src/apply.rs:80-96` 과 `src/agents.rs:174-177`, `agents.rs:1033` 에서 막고 있어서, 새 계획기도 같은 길을 지나야 해요). 이 레이어가 쓴 블록은 생성 manifest 에 baseline 이 남아야 다음 생성기 실행이 충돌로 읽지 않아요(`src/apply.rs:47-55`, `src/agents.rs:1150`). 미리보기 키 `project_files` 는 할 일이 있을 때만 나오고, 확인 토큰은 선택 하나에 묶여요(`git_hooks` 와 같은 자리, 고정 키 13개 불변).

## Requirements

- REQ-1: When the updater writes a project file, the receipt shall record the earlier bytes or the absence of that file so that the revert command restores them.
- REQ-2 (REQ-1): IF a written file changed after the apply, THEN the revert command shall leave that file untouched and report it by name.
- REQ-3 (REQ-1): IF `CLAUDE.md` or `CLAUDE.local.md`, or a folder above it inside the project, is a symbolic link or the file is not valid UTF-8, THEN the updater shall refuse that file, write nothing to it and name it.
- REQ-4: When the updater writes a managed block, the updater shall record the baseline of that block in the generation manifest so that the next generation plan reads the block as unchanged. · failure: none — a failed manifest write is reported as a partial apply by the existing apply contract
- REQ-5: The layer `project_files` shall appear in the update preview only when it has something to do and in the revert order before every other layer.
- REQ-6: The confirm token of the layer shall be bound to one pick so that a token of another pick or of a changed project is refused.

## Acceptance

- AC-1 (REQ-1…REQ-3): the revert command restores a temporary project to its first bytes after each pick, skips a file that changed after the apply and names it, and the updater refuses a `CLAUDE.md` that is a link to `AGENTS.md`, a link folder and a file that is not UTF-8 without writing anything. This is the restore half of the first criterion of the card.
  verify: `cargo test --locked --offline --test claude_md_picks_cli receipt_`
  verify: `cargo test --locked --offline --test claude_md_picks_cli refuse_`
- AC-2 (REQ-4): after a pick the next generation plan reads the block as unchanged and reports no conflict.
  verify: `cargo test --locked --offline --test claude_md_picks_cli baseline_`
- AC-3 (REQ-5, REQ-6): the preview keeps its 13 fixed keys without work, adds `project_files` only with work, refuses a token of another pick, and reverts the layer first.
  verify: `cargo test --locked --offline --test update_user_layer_cli`
  verify: `cargo test --locked --offline --test claude_md_picks_cli layer_`
- AC-4: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/update.rs`, `src/project_update.rs`, `src/agents.rs`, `tests/claude_md_picks_cli.rs`, `tests/update_user_layer_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-project-layer/**`
- read-only: `src/apply.rs`, `src/go_state.rs`
- forbidden: `AGENTS.md`, `AGENTS.local.md` of this repository, the real HOME, the network, `.claude/` of this checkout

## Plan

1. Write through the existing exclusive and replace writers of `src/agents.rs` so the link and UTF-8 guards apply unchanged.
2. Receipt shape of `agents_local_merge`: before copies and after hashes; the revert checks the after hash first.
3. Wire the layer into `update.rs` after `git_hooks`: own token, conditional key, first in the revert order, an entry in `layer_asks`, a longer `REVERT_ORDER` array.
4. Record the block baseline through the same call the generator uses for its own block.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The layer's token is outside the plan digest. | `git_hooks` does the same and the 13 fixed keys stay. | Move it inside the digest. |

## Result

(Filled when the work is done.)
