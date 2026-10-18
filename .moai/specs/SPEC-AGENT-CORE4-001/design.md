---
id: SPEC-AGENT-CORE4-001
title: "Design: neutral agent specs, renders, guards, retirement, hiding"
version: "0.1.0"
created: 2026-10-18
---

# design.md — SPEC-AGENT-CORE4-001

Design direction only; function and type names are chosen by the run phase. Evidence for every
statement of current behavior is in `research.md`.

## §1 Neutral source

**Location.** `internal/template/agentspec/` — an embedded build input beside the existing
`agentemit/agents-codex.yaml` manifest; not distributed to user projects (same status as the
manifest). One file per agent: a small neutral frontmatter block, then the body.

**Frontmatter fields.** `name`, `description`, `capabilities` (list), `sandbox` (`read-only` |
`workspace-write`), `model`, `effort`, `skills` (list of skill names; may be empty), `memory`
(`none` | `project`), `hooks` (optional role hooks, harness-neutral event names), `phase_arg`
(auditor only: allowed values `plan`, `sync`).

**Capability vocabulary.** `read`, `write` (optionally path-scoped: `write:docs`,
`write:.moai/specs/**`, `write:verdict-only`), `shell`, `spawn` (granted to no agent of the resulting roster), `mcp:<group>` (`spec`, `audit`, `verify`, `review-tasks`), `web`, `task-list`.
The Claude renderer maps each token to the tool grants the current definitions use (for example
`write` → `Write, Edit`; `mcp:spec` → the three `spec_*` tools). The Codex renderer reuses the
existing classification in `agents-codex.yaml`; a token with no mapping fails emission naming file
and token, exactly as today's unknown-tool-token failure.

**Bodies.** Written once, harness-neutral and install-root-free: skill references are by name
("load the planning-methodology skill `moai-agent-spec`"), never by path; no `~/.claude/` or
`.claude/skills/` literal. Two comment-delimited conditional blocks, `claude-only` and `codex-only`,
carry the few lines that differ per harness; the renderer for the other harness drops them. A
shared preamble (`_common`) holds the boilerplate the redesign counted at 10.5 KB across agents
(escalation protocol, conditional skill loading, blocker-report format) and is included at render
time into every agent, so it exists once in source (bytes still count in each rendered file).

**Skill bodies.** `skills/<name>.md` beside the specs (≤ 4,000 bytes each): `moai-agent-spec`,
`moai-agent-develop`, `moai-agent-docs`, `moai-agent-audit`. Reference material (the audit rubric,
checklist, shell recipes moved out of the 70 KB plan auditor) lives under
`references/<skill>/…` and is read by path on demand, never preloaded.

**Carried agents.** builder-harness, e2e-tester, manager-design, super-advisor are converted
mechanically: body verbatim except (a) install-root paths replaced by names, (b) references to
retired agents rewritten per §6. Their skills keep today's preloads.

## §2 Render and generation

One command regenerates everything: the existing `make agents-emit` becomes
"render neutral spec → C1, C2, C3, C3p"; `make agents-emit-check` re-renders in memory and fails on
any byte difference in any of the four, prints a one-line summary (`specs=N copies=4 diff=0`), and
never writes (read-only, as today). The existing golden test pattern
(`TestGoldenCommittedArtifactsMatchEmission`) is extended, not replaced. `make build` already runs
`agents-emit-check`, so the embed and the committed copies cannot silently diverge.

Characterization first (M1): for each of today's twelve definitions, the neutral conversion must
re-render to the **C2** bytes exactly (C2 is the distribution truth today). C1's seven drifts
disappear when C1 is regenerated; the dogfood project then reads the same bodies the distribution
ships, which is why bodies must be install-root-free before C1 is regenerated from them (M2).

**Deferred emission.** `TestNoAgentNameIsASubstringOfAnother` in the roster guard fails whenever a
roster name is contained in another, and `auditor` is contained in `plan-auditor` and
`sync-auditor`. The relaxation of that test is not an option, so the `auditor` spec carries
`hold_emission: true` from M3: it renders in memory (size verb, contract and heading tests) but is
written to no copy and sits in no roster literal until M8, where one commit emits it, deletes the
two old auditors, and flips the roster. No other new name is a substring of an existing one.

The Codex render changes in two ways beyond the neutral input: skill bodies of the spec's `skills`
list are appended inside the role instructions between `<name>:begin` / `<name>:end` marker lines,
and `claude-only` blocks are dropped. The emitted `permission_contract` is evaluated for the new
role names; the auditor rows keep `sandbox: UNSUPPORTED` and `write-path-scope: UNSUPPORTED`
(§4).

## §3 Size accounting

`moai agent size [--phase plan|sync|all] [--json]` renders the neutral specs in memory (plus the
skill files) and counts memory at its cap, so the number is reproducible on any clone and does not
depend on which copies are emitted yet (AC-AC4-005 asserts the copies equal those renders):
`total = rendered(agent A) + rendered(agent B) + Σ preloaded skill bytes + Σ memory cap`. The cap
is 8,000 for an agent with `memory: project` and 0 for one with `memory: none`.

Allocation (design values, 1 KB = 1,000 bytes; the AC thresholds are the pair totals, the rows are
the agreed split that makes the totals reachable):

| Component | Plan pair | Sync pair |
|---|---:|---:|
| planner definition | ≤ 14,000 | — |
| documenter definition | — | ≤ 10,000 |
| auditor definition (shared) | ≤ 20,000 | ≤ 20,000 |
| `moai-agent-spec` / `moai-agent-docs` | ≤ 4,000 | ≤ 4,000 |
| `moai-agent-audit` | ≤ 4,000 | ≤ 4,000 |
| planner / documenter memory (cap) | 8,000 | 8,000 |
| auditor memory | 0 | 0 |
| **Pair ceiling** | **56,000** (rows sum to 50,000) | **46,000** (rows sum to 46,000, zero headroom) |

The sync pair has no headroom, which is why it is a regression guard and not the stage criterion.
Baseline: 196,367 (plan), `research.md` §1.

Auditor shrink without loss: the plan auditor's rubric (19 KB), checklist (19 KB) and shell recipes
(7 KB) move verbatim into reference files; AC-AC4-009 checks that every rubric heading of the old
definitions survives in the references.

## §4 Auditor guard

**Where.** A new check in the PreToolUse handler family next to the agent-model and agent-stop
guards, matched on `Write|Edit|MultiEdit|NotebookEdit`. The settings template's PreToolUse
matcher is `Write|Edit|Bash` (`settings.json.tmpl:58`); it is widened by `MultiEdit|NotebookEdit`.
The `MultiEdit` entry at `:140` belongs to PostToolUse and does not cover this.

**Who.** The acting agent is read from the payload's `agent_type`. The set of verdict-only agent
names is generated from the neutral specs (capability `write:verdict-only`) so the guard and the
render cannot disagree. Residual: the identity is the spawn name verbatim (measured note in
`main-checkout-branch-guard.md`); an audit run under another name is not covered, and the retired
names are refused (§6) so the old names cannot carry the role.

**What.** Allowed: a target that, after absolute-ization against the hook's project root,
`Clean`, and symlink resolution of the deepest existing ancestor, lies under
`<project root>/.moai/reports/`. Everything else is denied with
`AUDITOR_WRITE_VIOLATION: <path> is outside .moai/reports/ for a verdict-only agent`. The trailing
separator in the prefix comparison is what rejects `.moai/reportsX/`. Fail-open (allow + an audit
row in `.moai/logs/`) when the path field is absent, the payload does not parse, or the project
root cannot be resolved. Other agents are never touched.

**Not covered, said once.** Bash can write by redirection; this SPEC does not parse shell. The
auditor's Bash is a read-only intent carried by the body and by `permissionMode`, not enforced.

**Codex.** The auditor renders `sandbox_mode = "read-only"`; the emitter records
`write-path-scope: UNSUPPORTED` (measured, 0.156.1) and the verdict file is written by the
launcher tool `codex_role_audit` (`AGENTS.md` capability row). Nothing here claims the sandbox is
applied to a spawned role.

## §5 Memory cap, spawn record, worker, preload probe

**Memory cap.** PreToolUse on `Write|Edit|MultiEdit` whose target is
`.claude/agent-memory/<agent>/MEMORY.md`: Write → the new content length; Edit/MultiEdit → the
current file with the edits applied (fail-open when the application is ambiguous). A result above
8,000 bytes is denied with `MEMORY_CAP_VIOLATION:` and the remedy sentence (shorten entries first;
then `moai memory archive <topic>…`, which moves files to `_archive/` and never deletes). Auditor:
`memory: none` renders no `memory:` field. `moai memory doctor` adds a per-agent size and an
`over_cap` flag to its JSON. Existing oversize files (39,509 / 38,426 / 22,133 / 12,768 on the
author's machine) are machine-local, so migration is an operator-run archive, not a commit; the
doctor lists them.

**Spawn record.** The existing audit row gains `phase`, `effort`, `effort_source`
(`declared` | `frontmatter` | `session`), `prompt_bytes`. Phase derives from the agent (planner
→ `plan`, implementer → `run`, documenter → `sync`) and, for the auditor, from a first prompt line
`phase: plan|sync` (absent → `unknown`; the auditor body returns a blocker report in that case).
Effort: the spawn's `effort` parameter when present, else the rendered agent's frontmatter default.
The prompt body stays out of the log (existing invariant). The attributable-row rule: the
SubagentStop row is written only when it can be tied to a named agent; today's stop rows without a
name (92% of 40,453 rows in the redesign's census) stop being appended.

**Worker.** Neutral spec with `capabilities: read, write, shell` (no `spawn`, no MCP), a short body
("carry out the assigned slice, return evidence, never spawn"), effort inherited. The run workflow
gains a `## Tier L Fan-Out` clause: the orchestrator spawns `implementer` (single milestone) or
`worker` (mechanical slice) leaves at depth one under the existing one-writer-per-tree rule —
this replaces `manager-lead` role A, which was the only agent holding `Agent`.

**Preload probe (measured AC).** In M6 the run phase creates a throwaway agent whose
`skills:` names a throwaway skill carrying a unique nonce line, hidden by the candidate mode, and
asks it to repeat that line. Evidence file `.moai/reports/t1649/preload-probe.md` records the
mode, the Claude Code version, the prompt, the reply, and the debug-log line (a skipped skill logs
a warning per the documentation). Candidates in order: **H1** explicit-name `skillOverrides: off`;
**H2** `user-invocable: false` + `skillOverrides: name-only`; **H3** inline the skill bodies at
render and publish no skill at all (the Codex behavior already). The shipped mode is the first
whose probe passes. If no live probe can run in the run phase, H3 ships and the hiding-settings
deliverable is recorded as satisfied-by-absence (decision-index Q9).

## §6 Retirement

**Single table.** `internal/template/agentspec/retired_names.yaml` (name, class, replacement text)
drives: the refusal hook, the generated check of the rule table, the reference-sweep allowlist
test, and the doctor message. Classes: `renamed` (spec, develop, docs), `phased` (the two
auditors → `auditor` + `phase:` line), `command` (todo, git, lead).

| Retired | Replacement | Class |
|---|---|---|
| manager-spec | `planner` | renamed |
| manager-develop | `implementer` | renamed |
| manager-docs | `documenter` | renamed |
| plan-auditor | `auditor`, first prompt line `phase: plan` | phased |
| sync-auditor | `auditor`, first prompt line `phase: sync` | phased |
| manager-lead | role B: `moai cc -f` leader session (role injection) + `moai-factory-foreman`; role A: the Tier L Fan-Out clause (`implementer`/`worker` leaves) | command |
| manager-todo | `moai todo` verbs and `/moai:todo` | command |
| manager-git | `moai factory complete` (landing edge); Route B branch/PR steps become orchestrator-direct per `moai-ref-git-workflow` | command |

**Refusal (the backward-compatibility decision).** After cutover (M8) a spawn naming a retired
name is **refused**, not aliased: the PreToolUse Agent handler denies with
`RETIRED_AGENT_REJECTED: <old> is retired; <replacement text>`. Justification: (1) an alias would
keep eight names alive in the roster, the registry and the docs, defeating the retirement and the
generated-copy invariant; (2) for the five renamed/phased names the replacement is unambiguous, so
the orchestrator's existing single-owner rule (`archived-agent-rejection.md` §D) lets it re-spawn
with the new name under an info log without a question — one retry round, no stall; (3) for the
three command-class names no agent can substitute, so a refusal that names the command is the only
honest answer; (4) the deny is testable offline from the table (AC-AC4-012). The hook does not
rewrite the spawn (the guard family's no-payload-rewrite property is preserved). A same-named
definition the user owns outside the `.claude/agents/moai/` namespace is honored (the update
namespace-protect rule already treats that directory as user-owned), which also means a stale
user-scope copy under `~/.claude/agents/` is not silently revived by the hook's absence of an
opinion: it is refused like any retired name unless it sits in a project's own non-`moai/` path. Running lanes that
still carry old names in context therefore degrade to one refused spawn plus a retry, not a hang.
Codex has no equivalent spawn interception wired for roles (`AGENTS.md` §9 lists the events); the
Codex roster simply no longer contains the files, and the failure is the runtime's own
unknown-role error — recorded as a Gap, not claimed as handled.

**Ordering (REQ-AC4-011).** Replacement paths first (M5), reference flip second (M7), removal last
(M8). M8 is the only milestone that deletes definition files, and its commit is a descendant of
the M5 verification commit — checkable from the commit graph (AC-AC4-021).

**Trailers.** The `Authored-By-Agent` enumeration and the ownership lint gain `planner`,
`implementer`, `documenter` (and `auditor` where an auditor commit is ever attributed) as accepted
values for the transitions their predecessors owned, and keep accepting the legacy values forever
(history is immutable). `manager-git` stays a recognized-but-ownerless value, as today.

## §7 Hiding settings

Generated from the union of `skills:` lists whose names start with `moai-agent-`: a settings
fragment merged by the template's settings builder (`settings.json.tmpl`), listing each name
explicitly with the selected state (§5 ladder). The doctor check (`moai doctor --check skills`,
today a no-op line) reports `moai-agent exposure: <n> exposed, <m> hidden`; any exposed count above
zero is a Fail. Codex: nothing published under `.agents/skills` for these names and no
`[[skills.config]]` entry (an entry for an unpublished path would be a ghost the existing prune
removes); the Codex render inlines the body (§2).

## §8 Roster guard

All 65 registry entries that name retired agents change in the milestone that changes the
definition set (M3 adds names; M8 removes them), and `TestCanonicalSourceIsTheDefinitionFileSetPlusExplore`
keeps comparing the roster literal to the definition file set. The enumeration threshold becomes
`ceil(0.7 × universe)`: 13 → 10 (today's value, reproduced) and 10 → 7, so the sweep still finds a
file listing seven of ten names. A fixture test plants such a file and expects the sweep to name it.

## §9 Reference sweep (M7)

Method: for each class in `research.md` §3, apply the rewrite, then run the sweep test, which reads
`retired_names.yaml` and fails on any retired name outside the allowlist file
(`agentspec/retired_names_allowlist.yaml`: path or glob + reason; docs-site and README entries
carry `expires: <docs rewrite card>`). Order inside M7: Go production and registries (so tests
compile), rules/skills/hooks/config and their template mirrors in the same commit pair, tests and
fixtures last. Mechanical 1:1 renames (spec/develop/docs) are scripted per file class; phased and
command-class names are rewritten by hand with the replacement text from the table. The
template-neutrality guard (no SPEC ids, dates, SHAs in templates) and the `moai:role-rules-required`
marker sweep apply to every new or rewritten template file.
