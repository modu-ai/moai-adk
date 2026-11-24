# Research notes: the harness a project step may propose

Read at 2026-11-20T13:48Z (clock read with `date -u` in the same step). Every fact below comes from the excerpts that a web search returned for that query, not from reading a whole page; where an excerpt did not cover a point, it says so.

## What an agent file and a skill file may hold (the formats a proposal must stay inside)

- Claude Code subagent file (<https://code.claude.com/docs/en/sub-agents>, search excerpt): YAML frontmatter between `---` markers, the system prompt as Markdown after it. Only `name` and `description` are required. `name` is at most 256 characters and has no colon. `description` tells Claude when to delegate. `tools` is optional (a comma-separated string or a list); left out, the subagent inherits every tool, and a list in which no entry resolves usually fails to launch. `model` is optional (`sonnet`, `opus`, `haiku`, a full id or `inherit`). Other fields exist (`disallowedTools`, `permissionMode`, `maxTurns`, `skills`, `mcpServers`, `hooks`, `memory`); a field Claude Code does not know is ignored without an error. The excerpt's own example is `tools: Read, Glob, Grep` with `model: sonnet`.
- Claude Code skill (<https://code.claude.com/docs/en/skills>, search excerpt): every frontmatter field is optional and `description` is recommended; with no `name`, the folder name is the command name. The combined description and `when_to_use` text is cut at 1,536 characters in the skill listing, so the use case goes first. Unknown fields are ignored.
- Skill authoring (<https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices>, search excerpt): descriptions in the third person with trigger terms; the body under 500 lines with detail in separate files one level deep; names that say the work (a verb ending in -ing) and not `helper` or `utils`. The platform's own rule that `name` is required (64 characters, description 1,024) is for uploaded skills, not for Claude Code.
- Not read: the Codex CLI agent and skill formats. `generation-plan` writes them as it does today (a Markdown agent file for Claude Code, a TOML one for Codex CLI), and this research does not check that against the Codex CLI documents.

What follows for a proposal: each proposed agent needs a name that starts with `prj-`, a `description` that says when to use it, a short list of tools (least needed; an empty or misspelled list can stop the agent from launching), and a scope; each proposed skill needs a name and a description that leads with its trigger.

## Dry run of the proposal step on two fixture projects

The step was followed by hand for two invented projects. The detected facts are what `project-detect` and the project's own files would give; the sources are the pages the searches returned (excerpts only).

### Fixture A: a Rust command-line tool

Detected: one binary crate, `clap` for arguments, `cargo test`, an integration-test folder, a CI workflow that runs `cargo test`.

Search for the stack: "Command Line Applications in Rust testing exit codes clap".

| Proposal | Kind | Why | Sources (read 2026-11-20T13:48Z) |
| --- | --- | --- | --- |
| `prj-cli-test-runner` | agent, tools: Read, Bash, scope: `src/`, `tests/` | the project has a CLI whose contract is arguments, output and exit codes; a runner that exercises the binary through its arguments and checks output and exit status keeps tests off the internals | <https://rust-cli.github.io/book/tutorial/testing.html> (named by the search; the excerpt about `assert_cmd` came from a secondary page, so this one is a pointer, not read); <https://rust.code-maven.com/testing/test-command-line-application.html> (excerpt) |
| `prj-exit-code-reviewer` | agent, tools: Read, Grep, scope: `src/` | an error should be printed once, to stderr, and map to a stable exit code; a reviewer that looks for that in each change catches the drift | <https://iankohdes.github.io/posts/0007-truth-or-consequences> (excerpt on `std::process::exit` and exit codes) |
| `prj-add-subcommand` | skill, trigger: adding or changing a `clap` subcommand | a subcommand needs the derive struct, a handler that returns `Result`, and an integration test of its output and exit code | the same testing pages as above |

### Fixture B: a TypeScript web application

Detected: `package.json`, `tsconfig.json`, React with Vite, Vitest for unit tests, a CI workflow that runs the type check and the tests.

Search for the stack: "TypeScript strict tsconfig typescript-eslint Vitest Playwright".

| Proposal | Kind | Why | Sources (read 2026-11-20T13:48Z) |
| --- | --- | --- | --- |
| `prj-type-strictness-reviewer` | agent, tools: Read, Grep, scope: `src/`, `tsconfig.json` | `strict` is the most important compiler option and is usually joined by `noUnusedLocals`, `noUnusedParameters` and `noImplicitReturns`; a reviewer that checks a change against the project's own settings keeps them from eroding | <https://www.totaltypescript.com/tsconfig-cheat-sheet> (listed; excerpt from a community post); <https://typescript-eslint.io/users/configs> (excerpt: `strict-type-checked` only when a nontrivial share of the developers are proficient) |
| `prj-unit-test-runner` | agent, tools: Read, Bash, scope: `src/`, `vitest.config.ts` | the project already uses Vitest; a runner that runs the include pattern and the optional type check inside tests answers a change in one pass | <https://dev.to/gabrielanhaia/real-world-typescript-project-setup-for-2026-5an> (excerpt on `vitest.config.ts` and `include`) |
| `prj-lint-config-check` | skill, trigger: changing ESLint or `tsconfig` settings | the search results mix the old `.eslintrc` form with the flat config, so a change must first check which format the installed ESLint reads | <https://typescript-eslint.io/users/configs> (excerpt) |

No end-to-end test agent is proposed for fixture B: the Playwright search returned nothing, and a proposal without a source is not made; the step would say so.

### The two lists differ

Fixture A proposes a CLI test runner, an exit-code reviewer and a subcommand skill; fixture B proposes a strictness reviewer, a unit-test runner and a lint-config skill. No name is shared, and each has at least one source line with its URL and read time. Neither list has a count in the instruction that produced it; both happen to hold three, which is the answer of two small projects, not a rule.

## What this does not show

- The excerpts are search summaries; a reviewer who needs a claim confirmed opens the page.
- The proposals were made by following the written step by hand, not by a second agent that did not know the SPEC.
- The Codex CLI formats and the web search tool's availability in every host were not checked.
