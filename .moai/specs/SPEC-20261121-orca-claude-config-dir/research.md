# Research: an Orca profile's moved Claude settings folder

Answers one question: does `moai update` reach the folder an Orca profile gives Claude Code, so that `/moai` is there in that session?

## Measured

Every command ran in a clean environment with a temporary HOME that is also the account home (`MOAI_ACCOUNT_HOME`); the real HOME was neither read nor written. The commands are `update-plan` and `update-apply` with a package folder, the two that `moai update` runs (D-1 of the SPEC). The tests are in `tests/orca_claude_config_dir.rs`.

| Case | Observed | Test |
| --- | --- | --- |
| `CLAUDE_CONFIG_DIR` names an absolute folder, preview | the plan names the profile's folder for Claude Code and leaves the Codex copy in `.agents/skills/moai` of the home | `profile_plan_targets_the_moved_folder` |
| the same, apply | `SKILL.md` with the name `moai` lands in the profile's folder and nothing is created in the home's `.claude` folder | `profile_apply_puts_the_skill_where_the_session_reads_it` |
| the same, second preview | nothing left to apply for Claude Code | `profile_second_plan_has_nothing_to_apply` |
| control: variable not set | the skill goes to `.claude/skills/moai` of the home and the profile folder stays untouched | `without_the_variable_the_skill_goes_to_home_claude` |

## Sources (adopted)

- https://code.claude.com/docs/en/claude-directory.md: the page the host-folder resolver cites for `CLAUDE_CONFIG_DIR` moving Claude Code's settings folder. Adopted for: Claude Code reads the moved folder.
- https://orca.dev/docs/cli/automations: `--reuse-session` continues in the previous live automation terminal; the page has no concurrency, lock, skip-if-running or dedupe field (its only skip is `--precheck`). Adopted for the automations paragraph of the adapter.
- https://orca.dev/docs/settings, https://orca.dev/docs/agents/supported, https://orca.dev/docs/cli/skills: each Claude account keeps its own config folder; none names `CLAUDE_CONFIG_DIR`, and none says where a skill install lands per account. Read times are in the adapter text.

## Not verified

- No live Orca profile session was run. That Claude Code in such a session reads the moved folder rests on the Claude Code page above.
- That Orca sets `CLAUDE_CONFIG_DIR` is unverified: Orca's pages never name it. It is the report behind the Issue that tracked the gap.
- What Orca's scheduled automation runs do across a restart, a sleeping machine, a manual run that overlaps a running one, or the missed-run grace window: the page is silent.
- Whether a managed account's session sees skills in the default Claude home: the skills page is silent, so `/moai` is checked in the session itself after SETUP or UPDATE.

## Overlap

The older drafts on the Orca messaging rules change other parts of `skills/moai/references/adapters/orca.md`; this change touches only the two table rows and the paragraphs named in the SPEC (D-3).
