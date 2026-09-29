---
description: Compatibility alias for the canonical GTD task-management entry point
argument-hint: "[\"<description>\"|list|next|done <n>|hold <n>|unhold <n>]"
allowed-tools: Skill
---

Dispatch the moai workflow `gtd` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `gtd` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `gtd` subcommand with the same arguments

The queue carries four states: `queued`, `picked`, `dropped`, `hold`.
`gtd hold <n>` parks a queued card out of every machine selector's reach —
operator (or lead) only, text untouched, no marker written; `gtd unhold <n>`
returns it to `queued`. A pick attempt on a held card is refused with
`unhold` named as the recovery verb.
