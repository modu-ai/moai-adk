# Research: where a host reads agents and skills below a folder of ours

Official documents only. Read times are UTC, 2026-11-20, from `date -u` in the same step as each fetch (a `moai-research` agent read them; the quotes are as it returned them). `S` = https://code.claude.com/docs/en/sub-agents.md, `K` = https://code.claude.com/docs/en/skills.md, `C` = https://code.claude.com/docs/en/commands.md, `P` = https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview.md.

## Claude Code

| Question | Answer | Quote | Source, read time |
| --- | --- | --- | --- |
| Are subfolders of `.claude/agents/` read? | yes, recursively | "Claude Code scans `.claude/agents/` and `~/.claude/agents/` recursively, so you can organize definitions into subfolders such as `agents/review/` or `agents/research/`." | S, 20:28:36Z |
| What names an agent? | the `name` field only; the folder is not part of it | "The subdirectory path doesn't affect how a subagent is identified or invoked, because identity comes only from the `name` frontmatter field." | S, 20:28:36Z |
| Two agents with the same `name` under one `.claude/agents/` (subfolders included)? | only one loads, no documented precedence | "if two files under the same `.claude/agents/` directory, including its subfolders, declare the same name, Claude Code loads only one of them, chosen by filesystem read order rather than a documented precedence." | S, 20:28:36Z |
| Same `name` in different places? | the higher-priority location wins | "When multiple subagents share the same name, Claude Code uses the one from the higher-priority location." (managed settings, `--agents`, `.claude/agents/`, `~/.claude/agents/`, plugin) | S, 20:28:40Z |
| Folder part of the name for project agents? | no; only plugin agents | "a subfolder inside a plugin's `agents/` directory becomes part of the scoped identifier: a file at `agents/review/security.md` in plugin `my-plugin` registers as `my-plugin:review:security`." (the page contrasts this with project and user scopes) | S, 20:28:36Z |
| Skill folder: must `name` equal the folder? | no | "`name` \| No \| Command name shown in the `/` menu. Defaults to the directory name." and "The directory name also invokes the skill." | K, 20:28:41Z |
| Skill documented layout | one level only | "Project \| `.claude/skills/<skill-name>/SKILL.md`"; the page never says "recursively" for skills | K, 20:28:46Z |
| A grouping folder `.claude/skills/<group>/<id>/SKILL.md`? | the documents are silent | (none) | K, 20:28:46Z |
| Skill name rules | Claude Code page: no length or case rule, refuses zero-width and control characters, `<`, `>`; folders `synced` and `anthropic-skills` do not load. The platform page: at most 64 characters, lowercase letters, digits and hyphens, no reserved words `anthropic`, `claude` | "`name`: Maximum 64 characters / Must contain only lowercase letters, numbers, and hyphens / Cannot contain XML tags / Cannot contain reserved words: `anthropic`, `claude`" | P, 20:28:57Z; K, 20:28:46Z |
| Does `/agents` mark which agents a project made? | no marker in the documents | "Print a reminder to ask Claude to create or manage subagents, or to edit `.claude/agents/` or `~/.claude/agents/` directly." | C, 20:28:52Z |

What this decides for Claude Code: agents may live in `.claude/agents/moai/` (read, named by `name`, so the file keeps `name: prj-<name>`); skills stay one level deep, so the namespace is a folder-name prefix `moai-` (`.claude/skills/moai-<id>/`), not a grouping folder. Two files with the same `name` anywhere under `.claude/agents/` are a hazard (one is picked by read order), so a move never leaves a copy of ours at the old path next to the new one, and a file of the user's with the same `name` is reported, not resolved.

## Codex CLI

Sources: `SUB` = https://developers.openai.com/codex/subagents.md, `SKL` = https://developers.openai.com/codex/skills.md, `SRC` = `https://raw.githubusercontent.com/openai/codex/a1afa122738382adabe2dd1ccfc510a70ed381f0/codex-rs/…` (main at 2026-11-20 19:57Z; the same files were read at tag `rust-v0.171.0`). Read 2026-11-20 20:28:51Z to 20:31Z. The documents are silent on subfolders; the source is not, and it is the ground for the Codex rows below.

| Question | Answer | Quote or source line | Source, read time |
| --- | --- | --- | --- |
| Are subfolders of `.codex/agents/` read? | yes in the source, silent in the documents | `agent-roles/src/discovery.rs:12-36`: `let mut dirs = vec![dir.clone()]; while let Some(dir) = dirs.pop() { … if entry.is_directory { dirs.push(path); continue; } if entry.is_file && path.as_path().extension().is_some_and(\|extension\| extension == "toml") { files.push(path); } }`, then `files.sort();`. The document says only: "add standalone TOML files under `~/.codex/agents/` … or `.codex/agents/`". | SRC 20:29:11Z; SUB 20:28:55Z |
| What identifies a Codex agent? | the `name` key, not the file name | "Codex identifies the custom agent by its `name` field. Matching the filename to the agent name is the simplest convention, but the `name` field is the source of truth." A file found in a folder without a `name` is skipped with "agent role file at {} must define a non-empty `name`". | SUB 20:28:55Z; SRC `agent_role_config.rs:73-84` 20:29:11Z |
| Same agent name twice? | the first file in path order wins, the other is skipped with a warning | `discover_agent_roles_in_dir`: `if roles.contains_key(&role_name) { push_agent_role_warning(… "duplicate agent role name `{role_name}` discovered in {}" …); continue; }`; files are sorted by path. A higher config layer (the project) overrides a lower one (the user). | SRC `loader.rs` 20:29:15Z |
| Is the project layer's config folder `.codex`? | yes (confirmed by the researcher in `config/src/state.rs:233`) | the project layer reads `config_folder.join("agents")` | SRC 20:30Z |
| Project skills folder | `.agents/skills` | "For repositories, Codex scans `.agents/skills` in every directory from your current working directory up to the repository root." | SKL 20:28:52Z |
| Nested skills folder `.agents/skills/<group>/<id>/SKILL.md`? | yes in the source (depth 6), silent in the documents | `ext/skills/src/loader/mod.rs:26`: `pub(super) const MAX_SCAN_DEPTH: usize = 6;`; `loader/discovery.rs:149`: `entry.path.basename().as_deref() == Some(SKILLS_FILENAME)`; a root that is not a plugin scans `Recursive` (`loader/host.rs:108`) | SRC 20:29:47Z, 20:31Z |
| Skill name rules | at most 64 characters (`parser.rs`: `const MAX_NAME_LEN: usize = 64;`); no character set check found; the folder name is only the fallback when `name` is empty; `description` must be non-empty | `let name = parsed.name.as_deref().map(sanitize_single_line).filter(\|value\| !value.is_empty()).unwrap_or_else(default_name);` | SRC 20:31Z |
| Two skills with one name? | both stay | "If two skills share the same `name`, Codex doesn't merge them; both can appear in skill selectors." | SKL 20:28:52Z |
| A namespace or prefix convention for project agents or skills? | none; the only namespace is a plugin's (`sample:search`) | `loader/namespace.rs`: "A plugin namespace is the plugin name from the nearest valid plugin manifest above a skill path." | SRC 20:31Z |

The researcher wrote one temporary copy of a source file into the scratchpad of its own session and deleted it; no project file was written.

What this decides for Codex CLI: an agent file in `.codex/agents/moai/<role>.toml` is read (source), keeps `name = "<role>"`, and a same-name file loses by path order with a warning, so a move still leaves no copy beside the old path. Skills take the one-level `moai-<id>` folder, which the document lists, instead of the nested folder the source also reads.
