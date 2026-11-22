# CLAUDE.md maintenance and host loading

`CLAUDE.md` keeps the rules every agent follows in this repository. It holds what `AGENTS.md` and the shared rules of `AGENTS.local.md` held; both were removed by the user's decision of 2026-11-22 ('AGENTS.md 제거해라. CLAUDE.md CLAUDE.local.md 만 사용한다', `SPEC-20261122-claude-md-layout` Plan 5), and notes of one user or machine go to the git-ignored `CLAUDE.local.md`. The parts below moved here from `AGENTS.md` (Issue #3096 S3) so the rules stay short; this file keeps its old name.

## How this file evolves: command forms

1. Capture the lesson or decision first in its ledger: user decision in the decisions log and the affected Issue; agent correction in the project's memory store.
2. Edit the authored body in small diffs, keep the whole file under its 44 KiB budget (`tests/repo_claude_md.rs`), then regenerate the managed section: `moai agents-plan <root>` with the typed request on stdin, review the plan, then `moai agents-apply <root> --plan-digest <d> --permission-reference <ref>`. Exactly one marker pair must remain.
3. Verify markers, `git diff --check`, no secrets or personal absolute paths; record the regeneration (date, plan digest, reason) in the decisions log.
4. Lessons that should become product defaults are added to the MoAI templates through their own Issue, never copied blindly.

## Host loading guidance

Official sources checked 2026-11-08; recheck when host behavior changes. [OpenAI AGENTS.md guide](https://learn.chatgpt.com/docs/agent-configuration/agents-md): Codex chains global guidance and project-root-to-cwd files; nearer scope wins; combined budget defaults to 32 KiB. [Anthropic memory guide](https://code.claude.com/docs/en/memory#agentsmd): direct AGENTS.md loading needs v2.1.277+ and applies only when no CLAUDE.md variant exists; `/memory` verifies loading. AGENTS.override.md and AGENTS.local.md are not Claude equivalents. [Anthropic best practices](https://code.claude.com/docs/en/best-practices): specific, short durable instructions with linked procedures and runnable checks. Unsupported sessions report loading as not_run; an explicit request to read this file is session context, not proof of discovery. Measured 2026-11-13 (Issue #2667, comment 6848071468): Claude Code 2.1.323 and Codex CLI 0.167.2 read only the `AGENTS.md` inside the session's own working tree, never the primary checkout's from a linked worktree, so that file was tracked (user decision 2026-11-13, replacing the per-machine rule of 2026-11-08). Since the user's decision of 2026-11-22 the tracked file is `CLAUDE.md`: Claude Code loads the `CLAUDE.md` of the session's own working tree, and a linked worktree loads the primary checkout's `CLAUDE.local.md` only when it sits inside the primary checkout folder (measured in slice 0 of `SPEC-20261122-claude-md-layout`).

## Merge driver for the skill source manifest

`skills/moai/references/source-manifest.json` holds a sha256 line per skill file, so two branches that edit one skill file conflict on that line even when the file itself merges cleanly (61 of the 62 tip merges in this history whose only conflict was the manifest, card C-74). `.gitattributes` names the merge driver `moai-manifest` for it. Register the driver once per clone; it writes only that repository's git folder, which its worktrees share, and is safe to run again:

    python3 -I tools/merge_driver_manifest.py --install
    python3 -I tools/merge_driver_manifest.py --check   # exit 0 when registered and current

`--install` copies the driver and `tools/source_manifest.py` into `<git common dir>/moai-manifest-driver/` and registers that copy by its absolute path, so merging into or rebasing a branch never runs the branch's own version of the script. When `tools/` changes, `--check` reports `stale` until `--install` runs again. Undo with `git config --local --unset merge.moai-manifest.driver` and by deleting that folder. Unregistered, git merges the manifest as text, as before.

The driver resolves the manifest when every skill file merges cleanly and no `origin` changed on both sides; otherwise the manifest stays conflicted with a message, and after the skill files are resolved `python3 -I tools/source_manifest.py write` rebuilds it. It also leaves the conflict when a folder disappears on one side while the other side adds or changes files in it, when a skill file has its own `merge` attribute, and when the git config it sees sets `diff.algorithm` to anything but histogram or turns on `merge.renormalize`. Options given as `git merge -X…` never reach a merge driver, so the driver cannot see them; `python3 -I tools/source_manifest.py check` (run in CI by `tests/source_manifest_test.py`) is the safety net and still fails on any manifest that does not match the files.

GitHub's server runs no merge driver. A pull request whose manifest lines conflict with the tip still shows a conflict on GitHub until someone merges the tip into the branch locally; the driver helps those local tip merges only.
