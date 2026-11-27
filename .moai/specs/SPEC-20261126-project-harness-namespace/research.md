# Research: where Claude Code reads agents and skills below a folder of ours

Official documents only. Read times are UTC, 2026-11-25, from `date -u` in the same step as each fetch (a `moai-research` agent read the raw Markdown with `curl` and grepped it, so the quotes are verbatim from the page bytes; line numbers are those of the raw `.md` as served at that read). `S` = https://code.claude.com/docs/en/sub-agents, `K` = https://code.claude.com/docs/en/skills, `P` = https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview. This document replaces the first reading of 2026-11-20, which also covered Codex CLI; Codex is withdrawn from the product (`SPEC-20261121-remove-codex-support`).

## Claude Code

| Question | Answer | Quote | Source, read time |
| --- | --- | --- | --- |
| Are subfolders of `.claude/agents/` read? | yes, recursively | "Claude Code scans `.claude/agents/` and `~/.claude/agents/` recursively, so you can organize definitions into subfolders such as `agents/review/` or `agents/research/`." (S line 173) | S, 21:55:57Z |
| What names an agent? | the `name` field only; the folder is not part of it | "The subdirectory path doesn't affect how a subagent is identified or invoked, because identity comes only from the `name` frontmatter field." (S line 173) | S, 21:55:57Z |
| Two agents with one `name` under one `.claude/agents/`? | only one loads | "If two files under the same `.claude/agents/` directory declare the same name, Claude Code loads only one of them." (S line 177) | S, 21:55:57Z |
| Same `name` in different places? | the higher-priority location wins | "When multiple subagents share the same name, Claude Code uses the one from the higher-priority location." (S line 155) | S, 21:55:57Z |
| Folder part of the name for project agents? | no; only plugin agents | "Unlike project and user scopes, a subfolder inside a plugin's `agents/` directory becomes part of the scoped identifier" (S line 179) | S, 21:55:57Z |
| Project skill layout | one level: a folder named for the skill | "`.claude/skills/<skill-name>/SKILL.md` \| Sessions in this repository." (K line 164) | K, 21:55:50Z |
| A grouping folder `.claude/skills/<group>/<id>/SKILL.md`? | the documents are silent; not confirmed | (none: the page never says such a folder is scanned or ignored) | K, 21:55:56Z |
| Must a skill `name` equal its folder? | no; it defaults to the directory name | "`name` \| No \| Command name shown in the `/` menu. Defaults to the directory name." (K line 419) and "The directory name also invokes the skill." (K line 461) | K, 21:55:56Z |
| Skill name rules | the Claude Code page lists characters it refuses (zero-width, control, `<`, `>`) and gives no length; the platform page gives the limits | "`name`: Maximum 64 characters" and "Must contain only lowercase letters, numbers, and hyphens" and "Cannot contain reserved words: `anthropic`, `claude`" (P lines 243-248) | K, 21:55:56Z; P, 21:56:04Z |
| Same skill name in two places? | personal over project; a plugin skill is namespaced | "Enterprise over personal, and personal over project." (K line 215) | K, 21:55:56Z |

What this decides: agents may live in `.claude/agents/moai/` (read, named by `name`, so the file keeps `name: prj-<name>`); skills stay one level deep, so the namespace is the folder-name prefix `moai-` (`.claude/skills/moai-<id>/`), not a grouping folder, because the document is silent on one. The folder name is also the skill's `name` (`moai-<id>`), which keeps a skill of the user's with the plain id from sharing a name with ours; the limit of 64 characters from the platform page bounds `moai-` plus the id. Two agent files with one `name` anywhere under `.claude/agents/` are a hazard (one is picked), so a move never leaves a copy of ours at the old path next to the new one.

## Not read

- A live test that Claude Code discovers `.claude/skills/<group>/<skill>/SKILL.md`: not run (read-only research), so the prefix layout is the documented one.
- Whether Claude Code enforces the platform page's name limits is not verified; the product bounds the folder name at 64 characters because the only documented limit is that one.
