---
id: SPEC-20261122-claude-md-home-routing
title: "The /moai routing text moves from the user's CLAUDE.md to rules/moai.md in the Claude config folder"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Home routing: rules/moai.md

## Intent

사용자 전역 `CLAUDE.md` 안의 `moai-routing` 블록(`src/home_routing.rs`)을 설정 폴더의 `rules/moai.md` 로 옮겨요. 설정 폴더는 `host_home` 이 정해요(`CLAUDE_CONFIG_DIR`, 없으면 `~/.claude`). 새 파일은 moai 혼자 쓰는 파일이라 마커 대신 홈 스타일 파일처럼 "마지막에 쓴 바이트와 같을 때만 덮어쓰기"로 판정해요(`src/home_style.rs:313-324`). 옛 블록은 사용자 `CLAUDE.md` 에 남아 있으므로 지우기는 따로 고르는 선택이고 기본이 아니에요: Orca 프로필은 `~/.claude/CLAUDE.md` 를 가져다 쓰고 `rules/` 폴더가 없어서, 블록을 지우면 그 세션들의 안내가 조용히 사라져요(결정 D-8 of `SPEC-20261122-claude-md-layout`). Claude Code 가 프로필 폴더의 `rules/` 를 읽는지는 아직 확인 전이에요. 이 SPEC 은 Codex 쪽 줄(`home_routing.rs` 의 codex 항목)을 건드리지 않아요.

## Requirements

- REQ-1: When the update plans the Claude routing layer, the planner shall target the file `rules/moai.md` inside the folder that `host_home` resolves for Claude Code.
- REQ-2 (REQ-1): IF `host_home` refuses the config folder because the variable is empty, relative or a link, THEN the planner shall block the routing layer and name the variable without its value.
- REQ-3 (REQ-1): IF `rules/moai.md` exists and its bytes differ from the bytes moai last wrote, THEN the planner shall report a conflict and keep the file.
- REQ-4: The renderer shall keep the routing text within 12 lines and omit the `paths` frontmatter key.
- REQ-5: When the earlier routing block sits in the `CLAUDE.md` of the config folder, the planner shall offer its removal only as a separate pick whose text says that sessions of a profile that imports this `CLAUDE.md` may lose the routing text. · failure: none — nothing is removed without the pick
- REQ-6: When the removal pick is given and the block bytes equal the recorded baseline, the updater shall remove exactly the block.
- REQ-7 (REQ-6): IF the block bytes differ from the recorded baseline, THEN the updater shall leave the block in place and name the reason.
- REQ-8: The updater shall write no other text into the `CLAUDE.md` of the user's config folder.
- REQ-9: When the routing layer is reverted, the revert command shall restore `rules/moai.md` and a removed block to their earlier bytes. · failure: none — the changed-file case follows the receipt rule of the project picks SPEC
- REQ-10 (REQ-9): IF a restored file changed after the apply, THEN the revert command shall leave that file untouched and report it by name.

## Acceptance

- AC-1 (REQ-1…REQ-4): the routing file lands in the resolved config folder with and without `CLAUDE_CONFIG_DIR`, an edited file is a conflict, a refused variable blocks the layer, and the text keeps its limits.
  verify: `cargo test --locked --offline --test host_home_cli`
  verify: `cargo test --locked --offline --test claude_md_routing_cli target_`
- AC-2 (REQ-5…REQ-8): the old block is removed only on the pick and only when unedited, and no other text reaches the user's `CLAUDE.md`.
  verify: `cargo test --locked --offline --test claude_md_routing_cli removal_`
- AC-3 (REQ-9, REQ-10): the revert restores both files and skips a changed one by name.
  verify: `cargo test --locked --offline --test claude_md_routing_cli revert_`
  verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-4: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/home_routing.rs`, `src/ops/home_routing_plan.rs`, `skills/moai/references/**` (the routing text file), `tests/claude_md_routing_cli.rs`, `tests/home_routing_cli.rs`, `tests/host_home_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-home-routing/**`
- read-only: `src/host_home.rs`, `src/home_style.rs`
- forbidden: the real HOME and the real config folder, the Codex entries of `src/home_routing.rs`, the network

## Plan

1. Give `home_routing` a second target kind, a whole file under `<config dir>/rules/`, judged like the style file; check that `host_home::place` follows the same link policy as the style file (the research found `home_routing` uses `resolve_named` and `home_style` uses `place`).
2. Add the removal pick for the old block, with the baseline taken from the routing state file.
3. Receipt: both files restore from before-copies; `home_routing.rs:527-547` already skips a file that changed since the apply.
4. Re-pin `home_routing_cli` and `update_user_layer_cli` (they pin `~/.claude/CLAUDE.md` as the routing file).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The folder is created when it is absent, under the resolved config folder only. | `rules/` is a plain folder Claude Code reads (claude-directory.md). | Refuse when absent. |
| D-2 | A `rules` link is refused unless it points inside the default folder, like the other config folders. | `SPEC-20261117-host-home-resolver` D10. | Allow any link. |

## Result

(Filled when the work is done.)
