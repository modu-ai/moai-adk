---
id: SPEC-20261122-worktree-place
title: "New worktrees live in the project's .claude/worktrees and the home worktree folder is retired"
status: draft
revision: 1
created: 2026-11-22
parent: null
---

# New worktrees live in the project's .claude/worktrees and the home worktree folder is retired

## Intent

사용자 결정(2026-11-22): 홈 워크트리 폴더는 더 이상 쓰지 않고, 새 워크트리는 프로젝트 자체 폴더 안의 Claude Code 기본 자리 `.claude/worktrees/<name>/` 으로 통일해요. 빌드는 각 워크트리 안의 도구 기본 폴더를 쓰고, 별도 빌드 홈을 새 워크트리에 고르지 않아요. 이는 새 워크트리를 홈 moai 폴더 아래 두던 규칙(홈 워크디렉터 설계 D1–D8)을 대체해요. 기존 워크트리는 있는 자리에서 브랜치가 끝날 때까지 그대로 동작하고, 이 카드는 그것을 옮기거나 지우지 않아요. 삭제는 언제나 사용자 픽이 필요해요.

이 슬라이스(조각 1)는 SPEC 만 써요. 코드는 뒤 조각이 써요.

## Non-goals

- 기존 워크트리의 이동이나 삭제. 홈·Go·과거 프로젝트 폴더의 워크트리는 분류기에서 계속 인식되고, 브랜치가 끝날 때까지 그대로 동작해요.
- 워크트리 삭제·정리의 사용자 픽 절차 변경: guarded 규칙은 그대로예요.
- 도구 폴더의 Python·셸 파일 정리: 다른 카드가 Rust 이관과 함께 처리해요.
- 통합 브랜치 이름과 리펙 가드: 다른 결정 노트가 소유해요.

## Requirements

- REQ-1: Every moai path that creates a new worktree shall create it inside the project's `.claude/worktrees/<name>/` folder.
- REQ-2: A worktree creation or a worktree path check shall refuse a new-worktree path outside that folder and name the new place in the refusal.
- REQ-3: The folder name of a new worktree shall stay the name `home_workdir::worktree_name` builds today, the safe `<project>-<id>` form.
- REQ-4: The tool shall stop choosing a separate build home for new worktrees, whose builds then use the tool's default build folder inside that worktree.
- REQ-5: Repository-wide scans and the test selection shall skip every path under `.claude/worktrees/`.
- REQ-6: The Claude Code settings moai ships shall set the worktree base to the local head, so a new worktree starts at the local integration branch.
- REQ-7: The worktree classifier shall keep recognising worktrees in the old places — the home folder, the Go generation folder and the earlier project state folder — and keep every command working on them.
- REQ-8: Instruction texts and shipped documents shall name only the new place for new worktrees and stop ordering a new worktree under the home worktree folder.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a new worktree made by every moai path lands in the project's `.claude/worktrees/<name>/` and a path elsewhere is refused with the new place named · verify: a temp-project test of the worktree command, the bundle build worktree and the launcher creation
- AC-2 (REQ-4, REQ-5): a new worktree holds no separate build home and builds use the default folder inside it; the repository scans and the affected-test selection contain no `.claude/worktrees/` path · verify: a scan fixture test and a worktree test
- AC-3 (REQ-6): the shipped settings carry the local-head worktree base · verify: the settings plan test
- AC-4 (REQ-7, REQ-8): a worktree in an old place keeps working in the classifier and the commands, and no instruction text orders the home folder for new worktrees · verify: a repository text scan and a classifier test

## Scope

The survey read the worktree-path lines of every source file that mentions worktrees and grouped them by the role this SPEC gives them.

- The path keystone (the functions the others call): `src/home_workdir.rs` (worktree_dir, worktrees_root, worktree_name, location_of, location_in, the reserved names), `src/lib.rs` (the WORKTREES_DIR constant of the earlier state folder).
- Creators of new worktrees: `src/worktree.rs` (the worktree command: the create path and the `git worktree add` spawn), `src/bundle/build.rs` (the bundle build worktree and its folder), `src/claude.rs` (the launcher's worktree settings and their validation), `src/main.rs` (the command dispatch).
- Checkers and guards: `src/hooks.rs` (the hook path check against the home worktrees root), `src/hook_guards.rs` (the worktree path guard), `src/worktree_include.rs` (the worktree include list), `src/audit_clone.rs` (the probe advice text that names the home folder).
- Move, clean-up and report: `src/worktree_gc.rs`, `src/worktree_move.rs`, `src/worktree_report.rs`, `src/ops/worktree_gc_plan.rs`, `src/ops/worktree_gc_apply.rs`, `src/ops/worktree_move_plan.rs`, `src/ops/worktree_move_apply.rs`, `src/ops/worktree_report.rs`, `src/optimize.rs` (the leftover worktree judge).
- Folder and slot naming: `src/heavy_slot.rs` (the slot folder), `src/home_layout.rs` (the home layout inventory row), `src/home_state.rs`, `src/home_install.rs`, `src/go_state.rs` (the Go folder recognition), `src/lane_env.rs` (the lane build environment).
- Instruction texts that name a worktree place: `src/help.rs`, `skills/moai/workflows/mode-agent.md`, the AGENTS.md managed section template in `src/agents.rs`, `src/audit_clone.rs`.
- Statusline and summary reads: `src/statusline/state.rs`, `src/statusline/counts.rs`, `src/ui_summary.rs`.
- Mentions to re-word with their own slices: `src/lane_registry.rs`, `src/trailers.rs`, `src/claim.rs`, `src/claim_link.rs`, `src/lane_loop.rs`, `src/lane_reset.rs`, `src/lane_state.rs`, `src/lane_status.rs`, `src/queue_query.rs`, `src/baseline.rs`, `src/versioning.rs`, `src/migrate.rs`, `src/hygiene.rs`, `src/ops/hygiene_apply.rs`, `src/ops/hygiene_plan.rs`, `src/ops/lane_safe_to_reset.rs`, `src/ops/ledger_recover_plan.rs`, `src/ledger_recover.rs`, `src/ledger/recover.rs`, `src/card/landed.rs`, `src/hook_engine.rs`, `src/scope.rs`.

## Plan

1. This slice: the SPEC only, linked to the card.
2. Next slices: the keystone change (the create path and the classifier), then the creators, then the checks and guards, then the texts — each with its own red test first and the check affected once per slice.

## Risks

- `.claude/worktrees/` sits inside the repository: every scan, test selector and packaging list must skip it, or the tool starts reading worktrees as project content.
- The classifier vocabulary (home, go, project, claude, other) is used by the guard, the clean-up and the move: a new place changes the default, not the words, so old places keep their names.
- The home folder retire touches texts in the skill and the managed templates; a missed text keeps ordering the old place.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The one place for new worktrees is `.claude/worktrees/<name>/` inside the project, and the home worktree folder stops taking new ones. | user decision 2026-11-22, the worktree-place card | a later user decision |
| D-2 | The classifier keeps recognising the old places and nothing moves or deletes them. | the same decision, the guarded-deletion rule | a later user decision |

## Result

(status draft — written when the slice lands)
