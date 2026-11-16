# AGENTS.md maintenance and host loading

`AGENTS.md` keeps the rules every agent follows. The parts below moved here from it (Issue #3096 S3) so the rules stay short; the text is the same as before the move, apart from the heading.

## How this file evolves: command forms

1. Capture the lesson or decision first in its ledger: user decision in the decisions log and the affected Issue; agent correction in the project's memory store.
2. Edit the authored body in small diffs, keep the whole file under the Codex 32 KiB instruction budget, then regenerate the managed section: `moai agents-plan <root>` with the typed request on stdin, review the plan, then `moai agents-apply <root> --plan-digest <d> --permission-reference <ref>`. Exactly one marker pair must remain.
3. Verify markers, `git diff --check`, no secrets or personal absolute paths; record the regeneration (date, plan digest, reason) in the decisions log.
4. Lessons that should become product defaults are added to the MoAI templates through their own Issue, never copied blindly.

## Host loading guidance

Official sources checked 2026-11-08; recheck when host behavior changes. [OpenAI AGENTS.md guide](https://learn.chatgpt.com/docs/agent-configuration/agents-md): Codex chains global guidance and project-root-to-cwd files; nearer scope wins; combined budget defaults to 32 KiB. [Anthropic memory guide](https://code.claude.com/docs/en/memory#agentsmd): direct AGENTS.md loading needs v2.1.277+ and applies only when no CLAUDE.md variant exists; `/memory` verifies loading. AGENTS.override.md and AGENTS.local.md are not Claude equivalents. [Anthropic best practices](https://code.claude.com/docs/en/best-practices): specific, short durable instructions with linked procedures and runnable checks. Unsupported sessions report loading as not_run; an explicit request to read this file is session context, not proof of discovery. Measured 2026-11-13 (Issue #2667, comment 6848071468): Claude Code 2.1.323 and Codex CLI 0.167.2 read only the `AGENTS.md` inside the session's own working tree, never the primary checkout's from a linked worktree, so this file is tracked (user decision 2026-11-13, replacing the per-machine rule of 2026-11-08).
