---
id: SPEC-20261123-progress-board-flow-icons
title: "The progress board is the one progress display, with one icon per step of the flow"
status: draft
revision: 1
created: 2026-11-23
---

# The progress board is the one progress display, with one icon per step of the flow

## Intent

User order of 2026-11-23: progress is shown with the MoAI progress board alone, never with the host's task list, and the board's icons follow the steps of the NeXT-Flow so the state of one piece of work reads at a glance. The leader wrote the new `## Progress board` section and already installed it in the four global files (rendered style file sha256 starts `5f6b5fd1d93f3b0a`; text of the section: `~/.moai/reports/moai-adk/leader-checkpoint-2026-11-22/style/output-style.md`). This slice ships the same text as the package source, keeps the render shipped until now in the earlier-renders list so a home that holds it is updated and never a conflict, and pins the new section.

## Requirements

- REQ-1: The package style source `output-style.md` of the home references folder shall equal the leader's text byte for byte, so its render has a sha256 that starts `5f6b5fd1d93f3b0a`.
- REQ-2: The earlier-renders list of `src/home_style.rs` shall hold the render shipped until now, with its text kept as a fixture, so a home that holds it plans as an update and not as a conflict.
- REQ-3: The progress board section shall name the eleven icons in flow order, the line order with what needs the user first and the failed last, and the turn-end rule that closes with the board.
- REQ-4: The progress board section shall tell a session never to use the host's task-list tool for progress.
- REQ-5: The source manifest shall carry the sha256 of the new style source.
- REQ-6: The other shipped skill texts shall name no task list or task tool for showing progress.

## Acceptance

- AC-1 (REQ-1): the render of the source has sha256 `5f6b5fd1d93f3b0a4553ba74df113bba99bb018ca819120e1a25926fe242d006`, and a home holding that render plans as unchanged or adopted, never conflict · verify: `cargo test --locked --offline --test home_style_ownership_cli`
- AC-2 (REQ-2): the earlier-renders list holds the render `80e189d2…` in the order of the fixtures, the fixture `render-5.md` has that sha256, and a home holding it plans an update · verify: `cargo test --locked --offline --test home_style_ownership_cli`
- AC-3 (REQ-3, REQ-4): the section holds the sample board with the eleven icons in order, the flow order line, the line order sentence and the turn-end sentence, and names no task tool · verify: `cargo test --locked --offline --test home_style_ownership_cli the_progress_board_is_the_one_progress_display_with_flow_icons`
- AC-4 (REQ-5): the manifest sha256 of that `output-style.md` equals the file's · verify: `cargo test --locked --offline --test home_style_cli`
- AC-5 (REQ-6): none of the listed skill and workflow files contains `task list`, `task-list`, `TaskCreate` or `TaskUpdate`, and the style names `task-list` once, to withdraw it · verify: `cargo test --locked --offline --test home_style_ownership_cli no_other_shipped_skill_text_points_progress_to_the_task_list`

## Scope

- write: this SPEC folder, `skills/moai/references/home` (its `output-style.md`), `skills/moai/references/source-manifest.json`
- write: `src/home_style.rs`, `tests/fixtures/home-style/render-5.md`, `tests/home_style_ownership_cli.rs`, `tests/home_style_cli.rs`
- read-only: everything else. The flow rule file `rules/moai/flow.md` is not in this repository and no shipped text here tells a session to use the task list; `templates/claude/output-styles/moai.md` carries the old icon list and is another card's.

## Plan

1. RED: the ownership test names five renders, the new current render and the new board section; it fails on the old list and the old pointer.
2. GREEN: install the leader's text as the source, append the outgoing render to `EARLIER_RENDERS` with its fixture, update the manifest sha256.
3. Check the affected test targets; one receipt run at READY.
