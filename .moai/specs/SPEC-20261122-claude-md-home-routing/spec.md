---
id: SPEC-20261122-claude-md-home-routing
title: "The /moai routing text moves from the user's CLAUDE.md to rules/moai.md in the Claude config folder"
status: draft
revision: 2
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Home routing: rules/moai.md

## Intent

사용자 전역 `CLAUDE.md` 안의 `moai-routing` 블록(`src/home_routing.rs`)을 설정 폴더의 `rules/moai.md` 로 옮겨요. 설정 폴더는 `host_home` 이 정해요(`CLAUDE_CONFIG_DIR`, 없으면 `~/.claude`). 새 파일은 moai 혼자 쓰는 파일이라 마커 대신 홈 스타일 파일처럼 "마지막에 쓴 바이트와 같을 때만 덮어쓰기"로 판정해요(`src/home_style.rs:313-324`). 옛 블록은 사용자 `CLAUDE.md` 에 남아 있으므로 지우기는 따로 고르는 선택이고 기본이 아니에요: Orca 프로필은 `~/.claude/CLAUDE.md` 를 가져다 쓰고 `rules/` 폴더가 없어서, 블록을 지우면 그 세션들의 안내가 조용히 사라져요(결정 D-8 of `SPEC-20261122-claude-md-layout`). Claude Code 가 프로필 폴더의 `rules/` 를 읽는지는 아직 확인 전이에요. 이 SPEC 은 Codex 쪽 줄(`home_routing.rs` 의 codex 항목)을 건드리지 않아요.

개정 2(2026-11-22, 코드 대조 뒤, 레이아웃 SPEC 개정 2 의 전역 설치 결정 반영): 이 SPEC 은 라우팅 파일 `rules/moai.md` 를 다루고, 같은 층이 쓰는 흐름 규칙 파일 `rules/moai/flow.md` 의 내용·한도·충돌 규칙은 자식 SPEC `SPEC-20261122-claude-md-flow-rule` 이 맡아요(요구 12개 한도 때문에 나눴어요). 둘 다 설정 폴더에 하나씩만 놓아요. 미리보기에는 기본 설정 폴더가 미리 골라진 채로 두 파일이 나오고, `moai update` 의 확인 토큰(사용자의 선택)이 있어야만 적용돼요. 추가 설정 폴더와 옛 블록 지우기는 따로 고르는 선택이에요. 시험 대상은 한 파일 `tests/claude_md_home_cli.rs` 하나예요(레이아웃 SPEC 의 `flow_` 와 같은 파일). 링크이거나 UTF-8 이 아닌 대상 파일은 충돌로 보고하고(레이아웃 REQ-12), 막힘(`blocked`)은 변수 자체가 거부될 때만이에요(REQ-2). 영수증은 항목마다 종류(`kind`)와 설정 폴더를 적어서 한 호스트에 파일이 둘이어도 되돌리기가 섞이지 않아요. 기존 영수증은 `kind` 가 없으면 옛 블록으로 읽어요. 이 SPEC 은 `src/update.rs` 를 고치지 않아요: `moai update` 는 이미 `home_routing::plan_for` 와 `apply_for` 로 라우팅 층을 계획하고 적용하므로, 호스트의 계획 항목이 한 파일 계획으로 바뀌면 코드 수정 없이 이어져요. 선택(흐름 규칙 설치, 설정 폴더, 옛 블록 지우기)을 `moai update` 의 확인 토큰으로 묶는 일은 `update.rs` 가 비는 뒤의 다른 변경이에요.

## Requirements

- REQ-1: When the update plans the Claude routing layer, the planner shall target the file `rules/moai.md` inside the folder that `host_home` resolves for Claude Code.
- REQ-2 (REQ-1): IF `host_home` refuses the config folder because the variable is empty or relative, THEN the planner shall block the routing layer and name the variable without its value.
- REQ-3 (REQ-1): IF `rules/moai.md` exists and its bytes differ from the bytes moai last wrote, or the file, the `rules` folder or the config folder is a link or the file is not valid UTF-8, THEN the planner shall report a conflict and keep the file.
- REQ-4: The renderer shall keep the routing text within 12 lines and omit the `paths` frontmatter key.
- REQ-5: When the earlier routing block sits in the `CLAUDE.md` of the config folder, the planner shall offer its removal only as a separate pick whose text says that sessions of a profile that imports this `CLAUDE.md` may lose the routing text. · failure: none — nothing is removed without the pick
- REQ-6: When the removal pick is given and the block bytes equal the recorded baseline, the updater shall remove exactly the block.
- REQ-7 (REQ-6): IF the block bytes differ from the recorded baseline, THEN the updater shall leave the block in place and name the reason.
- REQ-8: The updater shall write no other text into the `CLAUDE.md` of the user's config folder.
- REQ-9: When the routing layer is reverted, the revert command shall restore `rules/moai.md` and a removed block to their earlier bytes. · failure: none — the changed-file case follows the receipt rule of the project picks SPEC
- REQ-10 (REQ-9): IF a restored file changed after the apply, THEN the revert command shall leave that file untouched and report it by name.
- REQ-11: Where the user picks more than one config folder, the planner shall plan the routing file and, where picked, the flow rule once for each picked folder and keep the last-written bytes of each folder in its own state file.
- REQ-12 (REQ-9): IF a receipt written before this revision has no entry kind, THEN the revert command shall read it as the entry of the old block.

## Acceptance

- AC-1 (REQ-1…REQ-4): the routing file lands in the resolved config folder with and without `CLAUDE_CONFIG_DIR`, an edited file and a link are conflicts, a refused variable blocks the layer, and the text keeps its limits.
  verify: `cargo test --locked --offline --test host_home_cli`
  verify: `cargo test --locked --offline --test claude_md_home_cli target_`
- AC-2 (REQ-5…REQ-8): the old block is removed only on the pick and only when unedited, and no other text reaches the user's `CLAUDE.md`.
  verify: `cargo test --locked --offline --test claude_md_home_cli removal_`
- AC-3 (REQ-9, REQ-10, REQ-12): the revert restores the files and a removed block and skips a changed one by name, and a receipt without an entry kind still reverts.
  verify: `cargo test --locked --offline --test claude_md_home_cli revert_`
  verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-5 (REQ-11): each picked config folder gets its own routing file and its own state file, and the other folders stay untouched.
  verify: `cargo test --locked --offline --test claude_md_home_cli folders_`
- AC-4: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/home_routing.rs`, `src/home_routing/rules.rs` (new, declared inside `src/home_routing.rs`), `src/agents.rs` (one visibility change), `src/ops/home_routing_plan.rs`, `src/ops/home_routing_apply.rs`, `src/ops/home_routing_revert.rs`, the three operation fragments of the home routing commands in the design-review appendix, `skills/moai/references/**` (the routing text, the flow text and the source manifest), `skills/moai/workflows/update.md`, `tests/claude_md_home_cli.rs`, `tests/home_routing_cli.rs`, `tests/host_home_cli.rs`, `tests/revert_twice_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-home-routing/**`
- read-only: `src/host_home.rs`, `src/home_style.rs`, `src/update.rs`
- forbidden: the real HOME and the real config folder, the Codex entries of `src/home_routing.rs`, the network

## Plan

1. Give `home_routing` a second target kind, a whole file under `<config dir>/rules/`, judged like the style file; check that `host_home::place` follows the same link policy as the style file (the research found `home_routing` uses `resolve_named` and `home_style` uses `place`).
2. Add the removal pick for the old block, with the baseline taken from the routing state file.
3. Receipt: entries carry `kind` and the config folder, before-copies are named by entry, and a file that changed since the apply is skipped by name (`home_routing.rs:527-547` does this for one block per host; two files per host need the new naming).
4. Re-pin `home_routing_cli` and `update_user_layer_cli` (they pin `~/.claude/CLAUDE.md` as the routing file).
5. `home_routing/rules.rs` holds the whole-file three-way plan copied from the style file plan (not edited in `home_style.rs`), the source limit checks and the per-folder state files `home-routing/rules-<12 hex of the folder path>.json`; the old `claude.json` stays as the baseline of the old block.
6. The ops `home_routing_plan`, `home_routing_apply` and `home_routing_revert` take `--install-flow`, `--remove-old-block` and `--config-dir <abs>` (repeatable); the first doc line of each op stays.
7. Texts first: the flow text file beside the routing text file under the skill references, from the content list of REQ-14, checked by tests for limits and contents; the leader reads the wording before it lands.
8. Measure (not product code): whether the rules of a temporary config folder and the Compact Instructions section reach a summarizer request and the first request after a compaction, with a local stand-in API and a temporary HOME; whether a background session keeps the flow text beside its built-in git instructions. The result goes into the report folder; a run against the real API is the user's choice.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The folder is created when it is absent, under the resolved config folder only. | `rules/` is a plain folder Claude Code reads (claude-directory.md). | Refuse when absent. |
| D-2 | A `rules` link is refused unless it points inside the default folder, like the other config folders. | `SPEC-20261117-host-home-resolver` D10. | Allow any link. |
| D-3 | The routing file and the flow rule of the default config folder are shown in the `moai update` preview, preselected, and applied only with the confirm token of the update, which is the user's pick; extra config folders and the removal of the old block are separate picks. | The user's decision of 2026-11-22: moai's generic content is installed in each config folder the user picks, and the confirm of the update is the pick. | Make the flow rule a separate pick. |
| D-4 | The old block is no longer refreshed; its baseline stays for the removal pick, and the plan says the text is duplicated until the block is removed. | A removal needs the last-written bytes to know the block is unedited. | Keep refreshing it. |
| D-5 | Links and non-UTF-8 files are conflicts; only a refused variable blocks the layer. | Layout REQ-12 says conflict; the old code says blocked for links and its tests pin that, so those tests are re-pinned. | Keep blocked. |
| D-6 | The two separate picks (extra config folders, removal of the old block) are options of the ops (`--config-dir`, `--remove-old-block`) in this change; the routing and flow rows already appear in the `routing` part of the update preview and ride the update's own confirm token. | `src/update.rs` is edited by another change; no edit of it is needed for the rows. | Add a confirm token for the two picks to `update.rs` later. |
| D-7 | The safety line sits in the flow file. | The leader's note for the flow rule lists it there. | Put it in the routing file. |
| D-8 | The removal of the old block needs the routing file in place and owned by moai in the same folder. | Removing the block first would leave sessions of a profile without any routing text. | Allow removal without the file. |

## Result

- AC-1 met: `claude_md_home_cli target_` runs 7 tests and `host_home_cli` 12 (the routing file in the resolved folder with and without the config variable; an edited file and a link are conflicts; a refused variable blocks the layer).
- AC-2 met: `claude_md_home_cli removal_` runs 8 tests (the old block goes only on the pick, only when it equals its recorded baseline and only while the routing file is in place; no other text reaches the user's `CLAUDE.md`).
- AC-3 met: `claude_md_home_cli revert_` runs 6 tests and `update_user_layer_cli` 11 (files and the removed block come back, a changed file is skipped by name, a receipt entry without a kind reverts as the old block).
- AC-5 met: `claude_md_home_cli folders_` runs 4 tests (each picked folder has its own routing file and state file; the others stay untouched).
- AC-4 met: `moai check affected` on the slice reports fmt, ops rustfmt and clippy `--all-targets -D warnings` at exit 0.
- Open: the update preview does not yet show the removal of the old block as its own pick; that row needs the update flow and waits for the slice that edits it. The CLI flags `--remove-old-block` and `--config-dir` are the pick for now.
- closure_status: awaiting the independent audit.
