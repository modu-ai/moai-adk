# decision-index.md — SPEC-PLUGIN-MARKETPLACE-001

`interview.decision_gate: on` and `interview.recommendation_mode: pull`. Each row below is a decision surfaced
while assembling card t1435 that no operator answer in the interview settled. Rows state what is unresolved and
why; none carries a preferred answer. Options for every row are enumerated in `spec.md` §5 (Q-n is OD-n there), and
`spec.md` §5's marker table lists, for each row, the requirement clauses that carry its default and what changes if the verdict
differs. `Operator verdict` is empty at authoring and is read by the Kickoff gate. Label vocabulary: DECIDED /
POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER. No row is DECIDED or POLICY-COVERED, because no committed
artifact in the authority register answers any of them as written.

Revision 0.2.0 (after plan-audit iteration 1): Q1 no longer says the listing cost cannot be measured offline — it was, and the figure is
now in the row; Q4 records the GitHub state read in this revision; Q9 to Q13 are new (plan-audit D2 and the measurements behind it).

Revision 0.3.0 (after plan-audit iteration 2): Q14 is new (plan-audit N4 and the reading of `moai init` it required). Q3 now says 11 agents
at the default tier and 12 only if every tier is admitted (N6). No other row changed.

Revision 0.4.0 (final delta after plan-audit iteration 3): the fourteen verdict lines are filled by the batch acceptance recorded in the next
section; no row text, label, authority anchor or option changed.

## 운영자 일괄 수용 10-03

On 2026-10-03 the operator accepted, in one batch, the "default if unanswered" value of every open decision OD-1 to OD-14 (Q1 to Q14
below). **The source is the leader's relay, by a cross-session message; it is not an answer typed in this session**, and the authoring
lane received the leader's statement, not the operator's own words. Each verdict line below therefore reads
`accepted default (operator batch 10-03, via leader relay)` followed by the accepted value, copied from that row's default in the
`spec.md` §5 table. The batch changes no default, so no clause of the §5 marker table moves. This entry adds no recommendation and
changes no option text. The relay is not a committed artifact, so it is not an authority anchor: no row becomes DECIDED or
POLICY-COVERED, and every label and "Why unresolved" text stays as authored. The Definition of Done's "operator verdict or the line
`default adopted`" condition is met by these lines for all fourteen rows.

## Q1: Does `moai init` install the plugin by default while it still deploys the whole scaffold?

- Label: FOUNDER
- Authority anchor: none — the only source is the card text in the live queue, which is not a committed artifact
- Why unresolved: the card text says init installs automatically, and it was written before the init shrink (card t1438) was sequenced behind it. Until t1438 lands both copies of every skill and command exist in a user's session. The listing cost of the plugin side was measured offline in this revision: `claude plugin details` printed a projected always-on estimate of about 3,653 tokens for the core-only payload (24 skills and 17 commands) and about 6,483 for an all-tier payload, both labelled estimates by the tool. What scratch-home commands cannot measure is shadowing or collision of the plugin's `moai:<command>` names with the scaffold's own, and the scaffold copy's own listing cost; those need an authenticated session. Whether the operator accepts that interim state, or wants the step opt-in until t1438 lands, is a product call the evidence does not make.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) on by default now, as the card text says

## Q2: Which components that already exist in the scaffold ship in the v1 plugin: hook registrations, the MCP entry, both, or neither?

- Label: EVIDENCE-NEEDED
- Authority anchor: none — the measurements that would decide it do not exist
- Why unresolved: card t1434 measured one hook registration (R07, R08) and one MCP server (R06) in a plugin, never the same registration in a plugin and in project settings at once, so double firing is unmeasured. For the MCP entry the code already records that an explicit decline of the project entry is honored absolutely, and a plugin MCP server is suppressed only when a project server with the same command exists; `--llm gpt` declines the project entry and `--llm both` forces it (`init.go:997-1003`). The effect on a declined or non-interactive init was inferred from those facts, not observed. An authenticated session is needed to measure both.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) MCP entry only, no hooks

## Q3: Do the template agents ship in the v1 plugin (11 at the default tier, 12 if every tier is admitted)?

- Label: FOUNDER
- Authority anchor: none — the cited documentation is external and no committed artifact decides the question
- Why unresolved: the Claude Code documentation (`https://code.claude.com/docs/en/sub-agents`, "Choose the subagent scope", re-fetched in this revision) states that plugin subagents ignore `hooks`, `mcpServers` and `permissionMode`. All twelve template agents carry `permissionMode`, two of them (`super-advisor`, `sync-auditor`) as `plan`, which is what keeps them read-only, and four carry `hooks` or `mcpServers`. Shipping them as written changes their permission posture in the plugin copy. The runtime behavior of a plugin agent next to a project agent of the same role was not observed. Accepting or refusing a changed safety posture is the operator's call.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (c) ship no agents in v1

## Q4: Where does the plugin payload live in the repository, and how does the marketplace entry pin it?

- Label: EVIDENCE-NEEDED
- Authority anchor: none — no measurement of how a GitHub-hosted marketplace refreshes a relative-path entry exists
- Why unresolved: a sha pin written inside the repository cannot name its own commit; a tag ref can be written before the tag exists; a committed generated tree carries several hundred files while a release-time build changes the release harness. The default branch of `modu-ai/moai-adk` is `main` and it carries neither `.claude-plugin` nor `plugins/` today (read in this revision, both contents reads returned 404), and `main` advances only through release PRs. How Claude resolves and refreshes a relative-path plugin entry from a marketplace cloned from GitHub was not observed, because no scratch-home command can reach GitHub without a network call.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) committed generated tree `plugins/moai/`, entry `source: "./plugins/moai"`, no ref pin (tracks the default branch)

## Q5: What is the opt-out surface for the install step?

- Label: FOUNDER
- Authority anchor: none — no committed setting names it
- Why unresolved: the card text does not name an opt-out. The three install scripts run before any project exists, so a project config key cannot cover them; an environment variable and a flag can. Whether one switch covers both tools or each tool has its own, and whether a user-scope key is wanted, is not decided by any committed artifact.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) `--no-plugin` on `moai init` plus env `MOAI_SKIP_PLUGIN_INSTALL=1` honored by init, the verb, all three scripts and the automated callers of REQ-016

## Q6: Do the install scripts call a `moai` verb after the binary lands, or inline the two tool commands?

- Label: FOUNDER
- Authority anchor: none — the card text describes the scripts running the tool commands without saying how
- Why unresolved: the card names the commands, not the mechanism. Calling a verb gives one Go implementation with an injected-runner test seam and adds a public CLI verb, but the scripts download the latest release, so a script that calls a verb the release lacks fails until the release ships and must be written to tolerate that; inlining follows the card literally and needs three implementations plus two byte-identical docs-site copies, with no unit seam and no version skew. Which cost the operator prefers is a product call.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) call a `moai` verb after the binary lands (proposed name `moai plugin install`; one Go implementation, injected-runner tests, REQ-019)

## Q7: What severity does a plugin and binary version mismatch carry in `moai doctor`?

- Label: FOUNDER
- Authority anchor: none — no configuration or constitution clause sets doctor severities for a new check
- Why unresolved: the card says a check compares the two versions without naming a severity. The existing binary-lag check is annotated as never promoting to fail (`internal/cli/doctor.go`, a code comment, not a committed setting), which is context and not authority for a different check. Warn, fail and info are all consistent with the card text. A maintainer's own build stamps the binary from `make build` (`v3.2.0-rc.N` or the nearest tag) while the plugin carries the SSOT fallback `v3.1.3`, so at warn a maintainer machine warns.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) warn

## Q8: Which catalog tiers does the `moai` plugin carry?

- Label: FOUNDER
- Authority anchor: none — `internal/template/catalog.yaml` defines tiers and `moai init` defaults to the core tier, but neither decides what a plugin named "moai core plugin" carries
- Why unresolved: "core" in the card can mean the catalog `core` tier or simply the main plugin. The catalog holds 24 core and 13 optional-pack skills and 12 agents (11 core, 1 harness-generated). Including optional packs adds 13 skills a default init does not deploy; the offline estimates are about 3,653 always-on tokens for the core-only payload and about 6,483 for all tiers.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) core tier only

## Q9: Does the `--llm` harness decide which tool or tools the init step acts on?

- Label: FOUNDER
- Authority anchor: none — the closest committed decision, `SPEC-INIT-HARNESS-PROMPT-001` (completed) REQ-IHP-009, makes the harness decide the MCP-entry provisioning call; it answers that question and not this one, so it is context and not an anchor
- Why unresolved: the card text says init runs the commands "when `claude` exists" and "Codex likewise", and the install scripts have no harness at all. `moai init`'s flag help says a default init deploys the Claude tree and `--llm gpt` deploys no `.claude/` tree, and init's other tool-specific surfaces already follow the harness. Without a gate a default init writes the Codex profile whenever `codex` is on PATH and `--llm gpt` writes the Claude profile. Whether a person who chose one harness expects the other tool's profile to be written is a product call.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) gated as the precedents are: `claude` selects Claude only, `gpt` Codex only, `both` both; the verb and the scripts act on every tool found

## Q10: Which profile does the install step act on when the launcher's recorded profile and the environment differ?

- Label: FOUNDER
- Authority anchor: none — no committed artifact decides which profile an init-time plugin install belongs to
- Why unresolved: `moai init` run from a plain shell acts on `CLAUDE_CONFIG_DIR` or `~/.claude`, while a session started by `moai cc` reads the profile the launcher recorded for the project, and a project has a record only after a prior launch. The plugin may therefore land in a profile the person's sessions never read. Acting on the inherited environment, on the recorded profile, or refusing on a mismatch each has a cost the evidence does not rank.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) the inherited environment unchanged, and the config home printed

## Q11: Does the payload carry one language, several, or no commands?

- Label: FOUNDER
- Authority anchor: none — no committed artifact decides the language of a plugin payload
- Why unresolved: the plugin is one committed tree, while the scaffold's commands are rendered per `conversation_language`. The only locale-dependent text in the payload is the `description` and `argument-hint` line of 15 commands (18 locale conditionals, measured); a Korean, Japanese or Chinese user would see English there for the plugin copy while the scaffold copies stay localized until the init shrink. One plugin per locale multiplies the committed payload; shipping no commands avoids the question and the command-name collision at the cost of leaving commands in the scaffold longer.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) English only, one plugin

## Q12: At which scope does the install step install the plugin?

- Label: FOUNDER
- Authority anchor: none — no committed artifact decides installation scope
- Why unresolved: Claude's install and marketplace-add commands default to user scope and also accept project and local scope; Codex's `plugin add` has no scope option; the install scripts have no project. User scope installs once per profile for every project, project scope declares the plugin in the project's committed settings, and local scope in its ignored settings. The files project and local scope write and the registry shape they produce were not observed. Which scope a person expects from `moai init` is a product call.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) the tool default, user scope, no scope argument

## Q13: How does `moai doctor` read the Codex plugin state?

- Label: FOUNDER
- Authority anchor: none — no committed artifact decides how a diagnostic may obtain a tool's state
- Why unresolved: the Codex installed state lives in `config.toml` registration, and the cache directory outlives a removed registration, so the answer cannot come from the cache directory. Running `codex plugin list --json` answers correctly and took 0.02 s, but it starts the CLI, which created a `tmp/arg0` directory inside an empty Codex home; reading the `config.toml` stanza starts nothing but depends on a file layout observed at one codex version; skipping Codex leaves half the check out. Whether a diagnostic may start a tool that writes into the tool's home, and whether it may depend on a file layout rather than the tool's own report, is a product call.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) run `codex plugin list --json`, bounded by `config.DefaultPluginVersionProbeTimeout`

## Q14: How do the harness cases that run the real binary's `moai init` stay off the real home?

- Label: FOUNDER
- Authority anchor: none — the nearest text, `.claude/rules/moai/workflow/kanban-dispatch-mechanics.md` (a rule file, outside the authority register), says a command moved into a script file is not a workaround for the worktree guard and that a verification which cannot be expressed in one guarded invocation is to be reduced; it is context for this question, not an answer to it
- Why unresolved: `moai init` writes `$HOME/.claude/settings.json` on every run whatever `CLAUDE_CONFIG_DIR` says (read in this revision: `internal/cli/init.go:881-892`, `internal/core/project/autonomy_bundle.go:71-83`), and `HOME` is the only seam that moves it; `MOAI_HOME` moves the `~/.moai` root only. The worktree guard refuses a `HOME=` prefix, and the iteration-1 harness got round that by setting `HOME` inside a script file, which the audit named as the thing the doctrine says is not a workaround. Without a `HOME` override, a harness case that runs the real binary's `init` writes the real user settings file. Dropping those nine cases leaves init's call into the step covered by Go tests with an injected runner and leaves its real process start unexercised (G-8); keeping them needs either an exception to the doctrine or a CI-only run. Which cost the dispatcher accepts is not decided by anything committed.
- Operator verdict: accepted default (operator batch 10-03, via leader relay) — (a) no harness case runs the real binary's `init`: the init flow is asserted by Go tests with an injected runner (their `TestMain` already sandboxes the home), the real exec path by the verb and installer cases, and the gap is recorded (G-8)
