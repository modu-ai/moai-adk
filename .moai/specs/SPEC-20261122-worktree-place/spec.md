---
id: SPEC-20261122-worktree-place
title: "New worktrees live in the project's .claude/worktrees and the home worktree folder is retired"
status: draft
revision: 3
created: 2026-11-22
parent: null
---

# New worktrees live in the project's .claude/worktrees and the home worktree folder is retired

## Intent

사용자 결정(2026-11-22): 홈 워크트리 폴더는 더 이상 쓰지 않고, 새 워크트리는 프로젝트 자체 폴더 안의 Claude Code 기본 자리 `.claude/worktrees/<name>/` 으로 통일해요. 빌드는 각 워크트리 안의 도구 기본 폴더를 쓰고, 별도 빌드 홈을 새 워크트리에 고르지 않아요. 이는 새 워크트리를 홈 moai 폴더 아래 두던 규칙(홈 워크디렉터 설계 D1–D8)을 대체해요. 기존 워크트리는 있는 자리에서 브랜치가 끝날 때까지 그대로 동작하고, 이 카드는 그것을 옮기거나 지우지 않아요. 삭제는 언제나 사용자 선택이 필요해요.

이 슬라이스(조각 1)는 SPEC 만 써요. 코드는 뒤 조각이 써요.

## Non-goals

- 기존 워크트리의 이동이나 삭제. 홈·Go·과거 프로젝트 폴더의 워크트리는 분류기에서 계속 인식되고, 브랜치가 끝날 때까지 그대로 동작해요.
- 워크트리 삭제·정리의 사용자 선택 절차 변경: guarded 규칙은 그대로예요.
- 도구 폴더의 Python·셸 파일 정리: 다른 카드가 Rust 이관과 함께 처리해요.
- 통합 브랜치 이름과 참조 가드: 다른 결정 노트가 소유해요.

## Requirements

- REQ-1: Every moai path that creates a new worktree shall create it inside the project's `.claude/worktrees/<name>/` folder.
- REQ-2: A worktree creation or a worktree path check shall refuse a new-worktree path outside that folder and name the new place in the refusal.
- REQ-3: The folder name of a worktree that moai creates shall stay the name `home_workdir::worktree_name` builds today, the safe `<project>-<id>` form.
- REQ-4: A worktree that Claude Code or Orca made under the project's `.claude/worktrees/` shall be accepted with whatever folder name it has — the name rule of REQ-3 applying only to worktrees moai creates — with the branch of such a host-made worktree named by the card take or the submit that uses it, not by the folder.
- REQ-5: The tool shall stop choosing a separate build home for new worktrees, whose builds then use the tool's default build folder inside that worktree.
- REQ-6: Repository-wide scans and the test selection shall skip every path under `.claude/worktrees/`.
- REQ-7: The Claude Code settings moai ships shall set the worktree base to the head of the session that makes the worktree, so a host-made worktree starts from that session's local history instead of the remote default branch, while moai's own creation passes the integration branch explicitly and the base read at submit time is what counts.
- REQ-8: The worktree classifier shall keep recognising worktrees in the old places — the home folder, the Go generation folder and the earlier project state folder — and keep every command working on them.
- REQ-9: Instruction texts and shipped documents shall name only the new place for new worktrees and stop ordering a new worktree under the home worktree folder.
- REQ-10: When moai creates a worktree inside the project's `.claude/worktrees/` folder, the creation shall leave that folder listed in the repository's local exclude file unless git already ignores it, so that the status of the main tree does not change.
- REQ-11 (REQ-10): IF the repository's local exclude file cannot be written, THEN the creation shall still succeed and name the error in its answer.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a new worktree made by every moai path lands in the project's `.claude/worktrees/<name>/` and a path elsewhere is refused with the new place named · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_creation_` and `worktree_place_refusal_`
- AC-2 (REQ-4): a worktree that Claude Code or Orca made is accepted with any folder name and its branch comes from the card take or the submit · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_host_made_`
- AC-3 (REQ-5): a new worktree holds no separate build home and builds use the default folder inside it · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_build_folder_`
- AC-4 (REQ-6): the scans and the affected-test selection contain no `.claude/worktrees/` path · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_scans_`
- AC-5 (REQ-7): the shipped settings carry the maker-head worktree base · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_settings_base_`
- AC-6 (REQ-8): a worktree in an old place keeps working in the classifier and the commands · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_old_places_`
- AC-7 (REQ-9): no instruction text orders the home folder for new worktrees · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_texts_`
- AC-8 (REQ-10, REQ-11): a creation leaves the main tree's status as it was in a repository that does not ignore the folder, adds the one exclude line only once, and writes nothing when git already ignores the folder, and an exclude file that cannot be written leaves the creation done with the error named · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_status_`

## Scope

The survey read the worktree-path lines of every source file that mentions worktrees and grouped them by the role this SPEC gives them.

- The path keystone (the functions the others call): `src/home_workdir.rs` (worktree_dir, worktrees_root, worktree_name, location_of, location_in, the reserved names), `src/lib.rs` (the WORKTREES_DIR constant of the earlier state folder).
- The local exclude helper (one function every creator calls): `src/home_workdir.rs`; it reads `git rev-parse --git-common-dir` and `git check-ignore` and appends to `info/exclude` only when git does not ignore the folder.
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
- 측정 사실(리더, 2026-11-22): 루트 Cargo.toml 에 workspace 절이 없고, 이 저장소의 기존 `.claude/worktrees/<name>/` 안에서 cargo metadata 는 exit 0 — 겹친 워크트리도 자체적으로 빌드돼요. 스캔 빠뜨림 위험 선은 그대로 유지해요.
- The classifier vocabulary (home, go, project, claude, other) is used by the guard, the clean-up and the move: a new place changes the default, not the words, so old places keep their names.
- The home folder retire touches texts in the skill and the managed templates; a missed text keeps ordering the old place.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The one place for new worktrees is `.claude/worktrees/<name>/` inside the project, and the home worktree folder stops taking new ones. | user decision 2026-11-22, the worktree-place card | a later user decision |
| D-2 | The classifier keeps recognising the old places and nothing moves or deletes them. | the same decision, the guarded-deletion rule | a later user decision |
| D-3 | The main tree's status stays clean through the repository's local exclude file (`<common dir>/info/exclude`), not through `.gitignore` | A tracked `.gitignore` is the user's file and a worktree-only change must not edit it; the exclude file is local, never committed and shared by every linked worktree; a bundle build scene (a temporary repository with no ignore rule) failed on an untracked `.claude/` entry | write the line into `.gitignore` after the user's pick |

## Result

Slice 1 (the path keystone, the classifier and the status-clean creation). Status draft; only the lines below are claimed.

- AC-1: partly met. A new worktree of the worktree command and of the bundle build lands in the project place (`worktree_place_creation_`, `tests/worktree_cli.rs` `a_new_worktree_goes_under_the_project_place`, `tests/bundle_build_cli.rs`). Not met: the refusal of a path outside the place with the new place named (REQ-2), no `worktree_place_refusal_` test yet.
- AC-2: met. `worktree_place_host_made_` passes (any folder name under the project place is `claude`, a deeper path is `other`).
- AC-3, AC-4, AC-5, AC-7: not yet met (later slices).
- AC-6: met. `worktree_place_old_places_` passes (home, go and project places keep their names).
- AC-8: met for the helper and the two creators. `worktree_place_status_` 4 tests pass (one exclude line once, nothing written when git already ignores, a linked worktree writes the common folder, an exclude file that is a link is refused and not followed); the worktree apply and the bundle build call it before the add and put a failure in `place_exclude_error`. The answer key itself has no test of its own.

Checks run in `wt-keystone`: `cargo test --offline --locked --test worktree_place_cli` 7 passed, `worktree_cli` 23, `bundle_build_cli` 14, `worktree_branch_cli` 2, `worktree_build_root` 6 (1 ignored), `worktree_gc` 14; `moai check affected` (base 7f1da593d1eb, head b79bf531102a) PASS with 66 targets, every step exit 0; 4 mutants (apply call removed, bundle call removed, an ignored folder written anyway, the git dir instead of the common dir) all killed. Not run: the REQ-6 scans, settings base, guard and launcher paths, the instruction texts, a Windows type check.
