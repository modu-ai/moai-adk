---
id: SPEC-20261124-worktree-entry-no-prompt
title: "A lane makes its worktree with an absolute path and enters it by path, so the host asks no permission"
status: draft
revision: 1
created: 2026-11-24
---

# A lane makes its worktree with an absolute path and enters it by path, so the host asks no permission

## Interview

- none: 별도 사용자 인터뷰는 없었어요. 요구사항은 카드 C-467 의 본문과 두 인수 기준, 리더 배차(2026-11-24: 공식 문서 → 임시 프로젝트 실측 → work 스킬·next-flow 수정, 권한 설정은 바꾸지 않아요)가 정했어요.

## Intent

Twice on 2026-11-24 a lane stopped at the Claude Code question `permission-root relocation to "<path>" — a model-supplied worktree outside .claude/worktrees/` when it called `EnterWorktree` on a path that is inside `.claude/worktrees/`: agent-6 (C-464, 10:45:37Z) and agent-4 (C-461, 11:05:34Z). A lane's permission question belongs to the user, so each lane rested until it was answered (agent-6 for 25 minutes).

The lane transcripts (read 2026-11-24) show the same sequence both times. The session stood in an earlier card's worktree and ran `git worktree add -b <branch> .claude/worktrees/<name> NeXT` with a **relative** path. Git resolved it against the current folder, so the worktree landed at `<earlier worktree>/.claude/worktrees/<name>`. The lane then called `EnterWorktree` with `<primary checkout>/.claude/worktrees/<name>`, a path that did not exist; after the user answered the question the call failed with `ENOENT … realpath`. The lane found the nested worktree with `git worktree list`, moved it with `git worktree move` and entered it, with no question the second time.

The official documentation says that paths under `.claude/worktrees/` and new-worktree creation do not prompt and that a path outside it does (Appendix B). A measurement in throwaway projects on Claude Code 2.1.331 (Appendix A) shows the missing piece: the question also appears for a path under `.claude/worktrees/` **that does not exist yet**, from any starting folder. Entering an existing worktree by path never asks, and neither does creating one by name. Creating one by name is not usable here: it branches from the remote default branch, not from `NeXT`, and `worktree.baseRef` cannot name a branch (documentation, Appendix B).

`workflows/work.md` ends with "enter the worktree the leader made with `EnterWorktree` and its `path`, or leave the current one first before you make a new one", and says nothing on how to make it. This SPEC writes the method down once: leave the card worktree, make the worktree with an absolute path, read it back from git, enter it by that same path.

## Non-goals

- No change to any permission rule or settings file (`permissions`, `worktree.baseRef`, a hook). Widening permissions is the user's decision.
- No new `moai` command; one that runs the steps in one call is offered as a follow-up in D-4.
- No change to the worktree place, the branch naming or the removal rules.
- No claim about versions other than 2.1.331 (measured) or about the interactive screen text, which was not reproduced here.

## Requirements

- REQ-1: When a lane starts a card in a new worktree, `workflows/work.md` shall tell the lane to create it from outside every card worktree with `git worktree add -b <branch> <absolute path under the primary checkout's .claude/worktrees> <integration branch>`.
- REQ-2 (REQ-1): `workflows/work.md` shall tell the lane never to give `git worktree add` a relative path, and to read the primary checkout's path from the first entry of `git worktree list --porcelain`.
- REQ-3 (REQ-1): `workflows/work.md` shall tell the lane to read `git worktree list --porcelain` for that exact path after the creation and to call `EnterWorktree` with that same absolute `path` only after it is listed.
- REQ-4 (REQ-1): IF the path is not listed or the folder does not exist, THEN `workflows/work.md` shall tell the lane not to call `EnterWorktree` on it, because the host asks the user about a path that does not exist, and to fix the creation first.
- REQ-5: `workflows/work.md` shall tell the lane not to make a card worktree with `EnterWorktree` and a `name`, because that branches from the remote default branch and not from the integration branch.
- REQ-6 (REQ-1): `workflows/work.md` shall tell a lane that stands in a worktree it entered with `EnterWorktree` to leave it with `ExitWorktree(keep)` first, because the host refuses a `git worktree add` into the shared checkout from there.
- REQ-7 (REQ-1): The method shall appear outside `work.md` only as a pointer to it, in the flow reference, the Claude Code row of the git safety reference and the worktree entry section of the lane rules.
- REQ-8: The flow reference shall stay within its 8192-byte budget.
- REQ-9: The skill manifest shall carry the new sha256 of every changed skill file.

## Acceptance

- AC-1 (REQ-1, REQ-2): `work.md` has the create command with `-b <branch>`, an absolute path and the integration branch, and the rule against a relative path, in one paragraph · verify: `cargo test --locked --offline --test next_flow_skill_scan worktree_entry_work_creates_by_absolute_path`
- AC-2 (REQ-3, REQ-4): `work.md` names the `git worktree list --porcelain` read-back before `EnterWorktree`, and the ban on entering a path that is not listed, with its reason · verify: `cargo test --locked --offline --test next_flow_skill_scan worktree_entry_work_reads_back_before_entering`
- AC-3 (REQ-5, REQ-6): `work.md` names the ban on `name`, with its reason, and the `ExitWorktree(keep)`-first step · verify: `cargo test --locked --offline --test next_flow_skill_scan worktree_entry_work_bans_name_and_exits_first`
- AC-4 (REQ-7): `next-flow.md`, `git-safety.md` and `lane-rules.md` each point to the method, name the absolute path, and none spells the create command again · verify: `cargo test --locked --offline --test next_flow_skill_scan worktree_entry_other_files_point_to_work`
- AC-5 (REQ-8, REQ-9): the flow reference is at most 8192 bytes and the manifest hashes match the files · verify: `cargo test --locked --offline --test next_flow_skill_scan refs_flow_reference_holds_each_stated_item` and `cargo test --locked --offline --test skill_structure`
- AC-6: the Result holds the throwaway run of the final text: a headless session that follows only the new `work.md` paragraph from a worktree it entered makes and enters a new worktree with no permission denial, and the same session with a relative path is the contrast case · verify: the Result lines of this SPEC (`moai spec check SPEC-20261124-worktree-entry-no-prompt --project .`)
- AC-7: Appendix A lists the 14 measured cases with their outcome, and the Result names each case not reproduced · verify: `moai spec check SPEC-20261124-worktree-entry-no-prompt --project .`

## Scope

- write: `.moai/specs/SPEC-20261124-worktree-entry-no-prompt/**`
- write: `skills/moai/workflows/work.md`, `skills/moai/references/flows/next-flow.md`, `skills/moai/references/internal/git-safety.md`, `docs/lane-rules.md`
- write: `skills/moai/references/source-manifest.json` (the sha256 of the changed skill files only)
- write: `tests/next_flow_skill_scan.rs` (new `worktree_entry_*` tests appended at the end)
- read-only: `CLAUDE.md` (names `EnterWorktree`; its wording stays true, see D-3)
- forbidden: every settings file, `src/`, other tests

## Plan

1. RED: four `worktree_entry_*` scan tests in `tests/next_flow_skill_scan.rs`; they fail on the current text.
2. GREEN: one paragraph in `work.md` replaces its last sentence; one clause in `next-flow.md`; the Claude Code row of `git-safety.md`; the first bullet of `lane-rules.md`; the manifest hashes by the project's own command.
3. Guard mutants: remove the absolute-path rule, the read-back rule and the `name` ban from a copy of `work.md`; each must fail a named test.
4. Throwaway run (AC-6): the final paragraph given to a headless session in a temporary project, with and without the relative path.
5. Result: each AC as met, not met or not_run, with the commands and exit codes.

## Risks

- The measurement is one Claude Code version (2.1.331) and a headless session; the lanes run 2.1.319 per `CLAUDE.local.md`. The question is a host behavior and may change; the scan tests pin our text, not the host. Mitigation: the Result records the version, and the method asks nothing the host could refuse (create, list, enter an existing path).
- The cause "a path that does not exist counts as outside" is read from outcomes (cases I and I2 against B), not from the host's source.
- `next-flow.md` has 141 bytes of headroom; the clause is written to fit, and the budget test fails otherwise.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The method is text in the four files above, with the steps spelled once in `work.md` | recurrence on 2026-11-24 (C-464, C-461); the lanes read `work.md` at every take | revert the commit |
| D-2 | A card worktree is never made with `EnterWorktree` and a `name` | measured and documented: it branches from the remote default branch (Appendix A F, Appendix B) | none needed |
| D-3 | `CLAUDE.md` and every settings file stay untouched | the leader's instruction; permissions are the user's decision | a later card may name `CLAUDE.md` in its scope |
| D-4 | Open: a `moai` command that runs create, read-back and path output in one call is not part of this SPEC | the second recurrence rule asks for a source change that stops the class; text is the first step | a follow-up card, if the leader picks it |

### Appendix A — measured cases

Claude Code 2.1.331, `claude -p --model haiku --permission-mode default`, throwaway git projects under the lane scratch folder, 2026-11-24 (11:3xZ to 11:5xZ). "asked" is the tool result `Enter the worktree at "<path>"? This moves the session's working directory …` plus a `permission_denials` entry; a headless session cannot answer it.

| Case | Starting folder | Call | Result |
| --- | --- | --- | --- |
| A | project root | `EnterWorktree(name)` | created and entered, no question |
| F | worktree outside `.claude/worktrees/` | `EnterWorktree(name)` | created under the project's `.claude/worktrees/`, no question |
| B | project root | `EnterWorktree(path)`, existing worktree under `.claude/worktrees/` | entered, no question |
| C | another worktree under `.claude/worktrees/` | `EnterWorktree(path)`, sibling | entered, no question |
| E | worktree outside `.claude/worktrees/` | `EnterWorktree(path)`, worktree under `.claude/worktrees/` | entered, no question |
| G | project root | `EnterWorktree(path)` twice in one session | both entered, no question |
| G2 | project root | enter, `ExitWorktree(keep)`, enter another | all entered, no question |
| K | project root | `EnterWorktree(path)`, worktree nested in another's `.claude/worktrees/` | entered, no question |
| D | project root | `EnterWorktree(path)`, existing worktree outside `.claude/worktrees/` | **asked** |
| I | project root | `EnterWorktree(path)`, path under `.claude/worktrees/` that does not exist | **asked** |
| I2 | worktree under `.claude/worktrees/` | same missing path | **asked** |
| J | project root | `EnterWorktree(path)`, existing plain folder under `.claude/worktrees/` | refused: `is not a linked worktree`, no question |
| M | project root | `git worktree add -b <b> <absolute path> NeXT`, then `EnterWorktree(path)` | created and entered, no question |
| N | worktree entered by `EnterWorktree` | `git worktree add` with an absolute path in the shared checkout | refused by the host guard: `gives worktree add a path in the shared checkout` |

### Appendix B — sources

Read 2026-11-24T11:37Z with `curl`, saved under the lane scratch folder: `https://code.claude.com/docs/en/worktrees.md` (sha256 `3912fa949021d32e…`, 45 829 bytes) and `https://code.claude.com/docs/en/tools-reference.md` (sha256 `acfb2f5d6090825c…`, 98 122 bytes). Quotes: "New-worktree creation and paths under `.claude/worktrees/` don't prompt" and "a `path` outside `.claude/worktrees/` prompts for your approval before entering" (tools-reference, EnterWorktree tool behavior); "Claude Code asks for your approval first … An `EnterWorktree` permission rule or choosing "don't ask again" doesn't suppress this prompt; only `bypassPermissions` mode skips it" (worktrees, Ask Claude to create a worktree); "New worktrees branch from the repository's default branch … You can't set `worktree.baseRef` to a branch name. To start a worktree from a specific existing branch, create it with git directly" (worktrees, Choose the base branch). Lane transcripts: the sessions of agent-6 (`ended-lease-sweep`) and agent-4 (`installed-version-by-name`) under the Claude projects folder.
