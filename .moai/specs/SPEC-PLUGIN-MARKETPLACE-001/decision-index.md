# decision-index.md — SPEC-PLUGIN-MARKETPLACE-001

`interview.decision_gate: on` and `interview.recommendation_mode: pull`. Each row below is a decision surfaced
while assembling card t1435 that no operator answer in the interview settled. Rows state what is unresolved and
why; none carries a preferred answer. Options for every row are enumerated in `spec.md` §5 (Q-n is OD-n there).
`Operator verdict` is empty at authoring and is read by the Kickoff gate. Label vocabulary: DECIDED /
POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER. No row is DECIDED or POLICY-COVERED, because no committed
artifact in the authority register answers any of them as written.

## Q1: Does `moai init` install the plugin by default while it still deploys the whole scaffold?

- Label: FOUNDER
- Authority anchor: none — the only source is the card text in the live queue, which is not a committed artifact
- Why unresolved: the card text says init installs automatically, and it was written before the init shrink (card t1438) was sequenced behind it. Until t1438 lands both copies of every skill and command exist in a user's session; the effect of two copies (shadowing, collision, listing cost) was not measured by card t1434 and cannot be measured by scratch-home commands. Whether the operator accepts that interim state, or wants the step opt-in until t1438, is a product call the evidence does not make.
- Operator verdict:

## Q2: Which components that already exist in the scaffold ship in the v1 plugin: hook registrations, the MCP entry, both, or neither?

- Label: EVIDENCE-NEEDED
- Authority anchor: none — the measurements that would decide it do not exist
- Why unresolved: card t1434 measured one hook registration (R07, R08) and one MCP server (R06) in a plugin, never the same registration in a plugin and in project settings at once, so double firing is unmeasured. For the MCP entry the code already records that an explicit decline of the project entry is honored absolutely, and a plugin MCP server is suppressed only when a project server with the same command exists; the effect on a declined or non-interactive init was inferred from those two facts, not observed. An authenticated session is needed to measure both.
- Operator verdict:

## Q3: Do the twelve template agents ship in the v1 plugin?

- Label: FOUNDER
- Authority anchor: none — the cited documentation is external and no committed artifact decides the question
- Why unresolved: the Claude Code documentation (`https://code.claude.com/docs/en/sub-agents`, "Choose the subagent scope") states that plugin subagents ignore `hooks`, `mcpServers` and `permissionMode`. All twelve template agents carry `permissionMode`, two of them (`super-advisor`, `sync-auditor`) as `plan`, which is what keeps them read-only, and four carry `hooks` or `mcpServers`. Shipping them as written changes their permission posture in the plugin copy. The runtime behavior of a plugin agent next to a project agent of the same role was not observed. Accepting or refusing a changed safety posture is the operator's call.
- Operator verdict:

## Q4: Where does the plugin payload live in the repository, and how does the marketplace entry pin it?

- Label: EVIDENCE-NEEDED
- Authority anchor: none — no measurement of how a GitHub-hosted marketplace refreshes a relative-path entry exists
- Why unresolved: a sha pin written inside the repository cannot name its own commit; a tag ref can be written before the tag exists; a committed generated tree carries several hundred files while a release-time build changes the release harness. How Claude resolves and refreshes a relative-path plugin entry from a marketplace cloned from GitHub was not observed, because the repository has no marketplace manifest on `main` yet and no scratch-home command can reach GitHub without a network call.
- Operator verdict:

## Q5: What is the opt-out surface for the install step?

- Label: FOUNDER
- Authority anchor: none — no committed setting names it
- Why unresolved: the card text does not name an opt-out. The three install scripts run before any project exists, so a project config key cannot cover them; an environment variable and a flag can. Whether one switch covers both tools or each tool has its own, and whether a user-scope key is wanted, is not decided by any committed artifact.
- Operator verdict:

## Q6: Do the install scripts call a `moai` verb after the binary lands, or inline the two tool commands?

- Label: FOUNDER
- Authority anchor: none — the card text describes the scripts running the tool commands without saying how
- Why unresolved: the card names the commands, not the mechanism. Calling a verb gives one Go implementation with an injected-runner test seam and adds a public CLI verb; inlining follows the card literally and needs three implementations plus two byte-identical docs-site copies, with no unit seam. Which cost the operator prefers is a product call.
- Operator verdict:

## Q7: What severity does a plugin and binary version mismatch carry in `moai doctor`?

- Label: FOUNDER
- Authority anchor: none — no configuration or constitution clause sets doctor severities for a new check
- Why unresolved: the card says a check compares the two versions without naming a severity. The existing binary-lag check is annotated as never promoting to fail (`internal/cli/doctor.go`, a code comment, not a committed setting), which is context and not authority for a different check. Warn, fail and info are all consistent with the card text.
- Operator verdict:

## Q8: Which catalog tiers does the `moai` plugin carry?

- Label: FOUNDER
- Authority anchor: none — `internal/template/catalog.yaml` defines tiers and `moai init` defaults to the core tier, but neither decides what a plugin named "moai core plugin" carries
- Why unresolved: "core" in the card can mean the catalog `core` tier or simply the main plugin. The catalog holds 35 core, 13 optional-pack and 1 harness-generated entries by tier grep. Including optional packs adds skills a default init does not deploy.
- Operator verdict:
