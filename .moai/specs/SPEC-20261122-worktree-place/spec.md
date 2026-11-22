---
id: SPEC-20261122-worktree-place
title: "New worktrees live in the project's .claude/worktrees and the home worktree folder is retired"
status: draft
revision: 4
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
- REQ-7: The Claude Code project settings moai ships shall carry `worktree.baseRef` set to `"head"`, so a subagent worktree, an audit one included, and a reopened worktree start from the current HEAD of the session instead of the default branch of the remote, while the worktree a lane works in is made by moai from the integration branch (REQ-12).
- REQ-8: The worktree classifier shall keep recognising worktrees in the old places — the home folder, the Go generation folder and the earlier project state folder — and keep every command working on them.
- REQ-9: Instruction texts and shipped documents shall name only the new place for new worktrees and stop ordering a new worktree under the home worktree folder.
- REQ-10: When moai creates a worktree inside the project's `.claude/worktrees/` folder, the creation shall leave that folder listed in the repository's local exclude file unless git already ignores it, so that the status of the main tree does not change.
- REQ-11 (REQ-10): IF the repository's local exclude file cannot be written, THEN the creation shall still succeed and name the error in its answer.
- REQ-12: Moai shall create the worktree of a lane itself from the main checkout, as `git worktree add .claude/worktrees/<name> -b <lane branch> <integration branch>`, so that the session opens that existing worktree with `claude --worktree <name>` and no branch name is ever put in `worktree.baseRef`.
- REQ-13 (REQ-1): IF `.claude`, `.claude/worktrees` or the folder of the new worktree is a symbolic link, THEN the creation shall refuse, name the linked path and change nothing.
- REQ-14: The tidy command of moai shall remove a worktree of a landed lane with `git worktree remove`, after `git worktree unlock` when the lock is Claude Code's from a session that has exited, and only on the user's pick.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a new worktree made by every moai path lands in the project's `.claude/worktrees/<name>/` and a path elsewhere is refused with the new place named · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_creation_` and `worktree_place_refusal_`
- AC-2 (REQ-4): a worktree that Claude Code or Orca made is accepted with any folder name and its branch comes from the card take or the submit · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_host_made_`
- AC-3 (REQ-5): a new worktree holds no separate build home and builds use the default folder inside it · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_build_folder_`
- AC-4 (REQ-6): the scans and the affected-test selection contain no `.claude/worktrees/` path · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_scans_`
- AC-5 (REQ-7): the shipped project settings carry `worktree.baseRef` as `"head"` · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_settings_base_`
- AC-6 (REQ-8): a worktree in an old place keeps working in the classifier and the commands · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_old_places_`
- AC-7 (REQ-9): no instruction text orders the home folder for new worktrees · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_texts_`
- AC-9 (REQ-12): the worktree a moai lane creation makes is `.claude/worktrees/<name>/` on a new lane branch started at the integration base, and the plan names the name that `claude --worktree` is given · verify: `cargo test --offline --locked --test worktree_cli a_new_worktree_goes_under_the_project_place`
- AC-10 (REQ-13): a linked `.claude`, a linked `.claude/worktrees` or a linked worktree folder is refused with the path named and nothing is changed · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_links_`, `cargo test --offline --locked --test worktree_cli worktree_place_links_` and `cargo test --offline --locked --test bundle_build_cli worktree_place_links_`
- AC-11 (REQ-14): the tidy command removes a landed lane worktree only on the user's pick, unlocking a Claude-held lock first · verify: `cargo test --offline --locked --test worktree_place_cli worktree_place_tidy_`
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
2. Slice 2 (revision 4): the nine rules of the official worktrees page enter as Research rows R-1..R-10 and decisions D-4..D-6; REQ-13 (linked places refused) is coded first; REQ-7 settings, REQ-14 tidy and the REQ-6 scans follow; Measurement M-1 stays open for the leader or the user.
3. Slice 3 (stacked on slice 2, branch read as the same SPEC through a repair suffix): REQ-6 for the tree digest, with the selection pinned; the walks of `scope.rs` and `atlas_tag.rs` stay open as a leader decision.
4. Later slices: the keystone change (the create path and the classifier), then the creators, then the checks and guards, then the texts — each with its own red test first and the check affected once per slice.

## Risks

- `.claude/worktrees/` sits inside the repository: every scan, test selector and packaging list must skip it, or the tool starts reading worktrees as project content.
- 측정 사실(리더, 2026-11-22): 루트 Cargo.toml 에 workspace 절이 없고, 이 저장소의 기존 `.claude/worktrees/<name>/` 안에서 cargo metadata 는 exit 0 — 겹친 워크트리도 자체적으로 빌드돼요. 스캔 빠뜨림 위험 선은 그대로 유지해요.
- Two creations that run at the same moment may both see the folder as not ignored and both append the exclude line; a repeated line is harmless to git and the next creation adds nothing.
- The classifier vocabulary (home, go, project, claude, other) is used by the guard, the clean-up and the move: a new place changes the default, not the words, so old places keep their names.
- The home folder retire touches texts in the skill and the managed templates; a missed text keeps ordering the old place.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The one place for new worktrees is `.claude/worktrees/<name>/` inside the project, and the home worktree folder stops taking new ones. | user decision 2026-11-22, the worktree-place card | a later user decision |
| D-2 | The classifier keeps recognising the old places and nothing moves or deletes them. | the same decision, the guarded-deletion rule | a later user decision |
| D-3 | The main tree's status stays clean through the repository's local exclude file (`<common dir>/info/exclude`), not through `.gitignore` | A tracked `.gitignore` is the user's file and a worktree-only change must not edit it; the exclude file is local, never committed and shared by every linked worktree; a bundle build scene (a temporary repository with no ignore rule) failed on an untracked `.claude/` entry | write the line into `.gitignore` after the user's pick |
| D-4 | A lane worktree is made by moai (`git worktree add` on the integration branch) and opened by name; `worktree.baseRef` is never a branch and the project settings carry `"head"`. | R-1, R-2, R-3, R-4, leader decision 2026-11-22 | a later user decision |
| D-5 | Cleanup of landed lane worktrees is moai's tidy command, on the user's pick; the Claude sweep keeps a worktree made with `git worktree add`. | R-5 | a later user decision |
| D-6 | `.claude`, `.claude/worktrees` and the worktree folder must be real folders; a link is refused before anything is changed. | R-8 | a later user decision |

### Research

The leader read the page whole at 01:19Z; the lane fetched the same page again (45,871 bytes, https://code.claude.com/docs/en/worktrees.md) and checked each row below against the fetched text. The quotes are the page's words.

| Id | Page says | Used for |
| --- | --- | --- |
| R-1 | "You can't set `worktree.baseRef` to a branch name. To start a worktree from a specific existing branch, create it with git directly." and `"fresh"` (default) branches "from the repository's default branch on the remote, usually `main`" | REQ-12: moai makes the lane worktree itself from the integration branch |
| R-2 | "If a worktree with the name you pass to `--worktree` already exists, Claude Code opens it instead of creating a new one." | REQ-12: the session opens the made worktree by its folder name |
| R-3 | "Subagent worktrees use the same base branch as `--worktree`, so they branch from your repository's default branch unless `worktree.baseRef` is set to `\"head\"`." and "Inside a worktree, `\"head\"` resolves to that worktree's `HEAD`, not the main checkout's." | REQ-7: the project settings carry `"head"`, audits branch from the current head |
| R-4 | A worktree is reopened at its old tip when "`worktree.baseRef` is `\"head\"`." | REQ-7: a reopened lane worktree is never reset to `main` |
| R-5 | The sweep keeps a worktree when "You created the worktree yourself with `git worktree add`"; "To clean up a worktree that the sweep keeps, run `git worktree remove` ... If git refuses because the worktree is locked, run `git worktree unlock` on it first."; a `-p` run leaves its lock until a later sweep | REQ-14: cleanup is moai's, on the user's pick |
| R-6 | "`${CLAUDE_PROJECT_DIR}` stays put: it still points at the project root where the session started" | The hook slice reads the worktree from the hook input working directory, never from this variable (not a requirement of this slice) |
| R-7 | An isolated session is blocked from a Bash command "that redirects git into the main checkout ... through `git -C`, `--git-dir`, a `GIT_DIR` or `GIT_WORK_TREE` variable, or a `cd` into the main checkout", and when it "can't verify from the command text that any git the command runs stays inside the worktree" | Measurement M-1 below; a refusal is never worked around with `git -C` |
| R-8 | "Claude Code refuses to create a worktree when `.claude`, `.claude/worktrees`, or the worktree directory itself is a symlink" and the tip "Add `.claude/worktrees/` to your `.gitignore` so worktree contents don't appear as untracked files in your main checkout." | REQ-13: moai refuses the same links; D-3 keeps the exclude file as the way moai ignores the place |
| R-9 | A subagent in a worktree in the default location under `.claude/worktrees/` "also doesn't" load that worktree's instruction files; a "Yes, and don't ask again" approval "saves the rule to the main checkout's `.claude/settings.local.json`" and "applies in the main checkout and in every other worktree" | Audit instructions go in the agent definition, not the worktree; an approval is shared by every lane (no requirement here) |
| R-10 | Claude's own worktree branch is named `worktree-<name>`; a moai lane branch follows the SPEC rule and carries no card number | Naming only |

Measurement M-1 (open): that `moai card land` is accepted from an isolated lane session. It needs a real `claude --worktree` session, which needs the user's login in the real HOME, so it is not run by the lane; the leader or the user runs it, and a refusal is recorded here, never worked around.

## Result

Slice 1 (the path keystone, the classifier and the status-clean creation). Status draft; only the lines below are claimed.

- AC-1: partly met. A new worktree of the worktree command and of the bundle build lands in the project place (`worktree_place_creation_`, `tests/worktree_cli.rs` `a_new_worktree_goes_under_the_project_place`, `tests/bundle_build_cli.rs`). Not met: the refusal of a path outside the place with the new place named (REQ-2), no `worktree_place_refusal_` test yet.
- AC-2: met for the classifier only. `worktree_place_host_made_` passes (any folder name under the project place is `claude`, a deeper path is `other`); that the branch of a host-made worktree comes from the card take or the submit is not tested here.
- AC-4: met for the tree digest and pinned for the selection. `worktree_place_scans_` passes in `tests/worktree_place_cli.rs`: the digest that RUN records and AUDIT compares (`run::TREE_EXCLUDES`, now with `.claude/worktrees`) does not change when a worktree appears under the place, and a project's own `.claude/settings.json` still counts; a changed path under the place selects no test (the selection reads changed paths and `area` answers none, so this holds by construction). The research read every other repository-wide walk (read 2026-11-22 04:19-04:25 UTC, nothing written): the project-detect, baseline, go-state and worktree-report walks already skip it; the tests walk named folders (`src`, `tests`, `tools`, `docs`, `skills/moai`, `templates`, `.github`) and none contains `.claude`. Not changed here and left to a decision: `scope.rs` `inspect_entry` and `atlas_tag::walk_files` enter the place when a caller names the project root (the first is a fail-closed link guard, so skipping a path there is a policy call, tier 2 guard work), and `early_reclaim.rs:249` was not measured.
- AC-9: partly met after the slice 2 audit. `worktree_place_lane_plan_names_the_session_worktree_name` passes (the plan carries `session_name`, the folder name a session gives `claude --worktree`) and a new worktree lands in `.claude/worktrees/<name>/`; that the lane branch starts at the integration base was read by the auditor (probe A), not by a test of mine.
- AC-10 repair (audit B1): the link check finds the project by what each ancestor of the path resolves to (`home_workdir::place_owner`), so an explicit `--path` spelled through a linked parent is refused as well (`worktree_place_links_an_explicit_path_through_a_linked_parent` and `worktree_place_links_the_owner_is_found_by_what_the_path_resolves_to`); the bundle build uses the same owner instead of counting path parts (audit N1). Repair checks: `worktree_cli` 30, `worktree_place_cli` 14, `bundle_build_cli` 18 passed; 5 mutants killed (lexical compare, no place prefix, apply gate removed, bundle gate removed, session name dropped).
- AC-3, AC-5, AC-7 and AC-11: not yet met (later slices; `claude-settings-apply` already manages `worktree.baseRef` when asked, what the project settings carry by default is still to be read before AC-5).
- AC-10: met. `worktree_place_links_` passes in three targets: the helper refuses a linked `.claude`, a linked `.claude/worktrees` and a linked or dangling worktree folder with the path named and leaves the folder behind the link empty; real or missing folders are accepted; the worktree apply answers a `place_is_a_link` conflict and makes no branch; the bundle build refuses before it makes a ref or a folder.
- AC-6: met for the classifier only. `worktree_place_old_places_` passes (home, go and project places keep their names); that every command keeps working on a worktree in an old place is not tested here.
- AC-8: met after the audit repair. `worktree_place_status_` passes in three targets: the helper writes one line once from the primary checkout, a linked worktree and a subfolder, writes nothing when git already ignores the folder, keeps a last line without a newline whole and refuses a link; the worktree apply and the bundle build write the line after a successful add (a failed add leaves the file alone) and put a failure under `place_exclude_error`, which a test per creator reads while the creation still happened.

Checks run. Slice 1 (head 3696cc647a4f, base 7f1da593d1eb): `moai check affected` PASS with 66 targets; the audit re-check PASS; 5 mutants killed in the repair. Slice 2 (head e4ed1bdacf64, base 3696cc647a4f): `moai check affected` PASS with 66 targets, py 9, every step exit 0; `cargo test --offline --locked --test worktree_place_cli` 13 passed, `worktree_cli` 28, `bundle_build_cli` 18; 6 mutants killed (the three entries of the link check, the apply call, the bundle call, and the audit's M11, the exclude line written before the bundle add); `spec check` 0 errors with the `size_requirements` warning (14 requirements, folded in one SPEC by the leader's order). Not run: the REQ-6 scans, the REQ-7 settings default, the REQ-14 tidy command, the REQ-2 refusal of a path outside the place, a Windows type check; Measurement M-1 stays open.
