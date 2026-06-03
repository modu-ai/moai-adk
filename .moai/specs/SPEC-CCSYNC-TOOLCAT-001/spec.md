---
id: SPEC-CCSYNC-TOOLCAT-001
title: "Claude Code tool-catalog sync (remove MultiEdit, migrate TodoWrite→Task*)"
version: "0.1.1"
status: in-progress
created: 2026-06-02
updated: 2026-06-03
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents + internal/template/templates"
lifecycle: spec-anchored
tags: "ccsync, tool-catalog, agents, template, multiedit, todowrite, task-family"
---

# SPEC-CCSYNC-TOOLCAT-001 — Claude Code tool-catalog sync

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-02 | manager-spec | Initial plan-phase authoring. Sibling of SPEC-CCSYNC-CLAUDEMD-001. Derived from a Claude Code official-docs gap audit. Two verified tool-catalog drift findings (H3 retired `MultiEdit` still declared, H4 `TodoWrite` disabled by default since v2.1.142 — migrate to Task* family) plus a low-priority bundle (foundation-cc authoring-kit scrub + a new CI guard test). |
| 0.1.1 | 2026-06-03 | manager-spec | plan-auditor PASS-WITH-DEBT defect remediation (AC command portability + mirror-test filename + REQ-011 AC coverage): replaced the non-portable BSD/macOS-failing sum segment (the legacy `bc`-piped form) with the POSIX-portable `awk '{s+=$1} END{print s}'` summation in AC-T-006/T-007/T-008; made AC-CCSYNC-T-017 binary (`make build && git diff --quiet …; echo $?` → expected `0`). REQ/scope/exclusions unchanged. |

## A. Context and Motivation

The agent + skill instruction layer declares each agent's tool surface in YAML
`tools:` frontmatter. Two tools named there have been changed in the current
Claude Code tool catalog:

- **`MultiEdit`** was removed from the catalog. Only `Edit` (with `replace_all`)
  and `Write` remain for file modification. An agent that declares `MultiEdit`
  declares a tool that no longer exists.
- **`TodoWrite`** is disabled by default as of Claude Code v2.1.142 in favor of
  the `TaskCreate` / `TaskGet` / `TaskList` / `TaskUpdate` family. An agent that
  declares only `TodoWrite` for task tracking gets NO task-tracking tool unless an
  environment override is set — yet CLAUDE.md §3 (Phase 3 Execute) and §14
  (Parallel Execution Safeguards: "All implementation agents MUST include …
  TaskCreate, TaskUpdate, TaskList, TaskGet") already require the Task* family.
  The frontmatter therefore contradicts the orchestrator rules.

Both findings are corroborated by THIS session's own tool environment: no
`MultiEdit` tool is available, and the deferred-tool list exposes
`TaskCreate`/`TaskUpdate`/`TaskList`/`TaskGet`, not `TodoWrite`.

This SPEC scopes the WHAT and WHY of the remediation. The HOW (the actual
frontmatter edits, body-instruction rewrite, settings.json.tmpl edits,
foundation-cc scrub, the new CI guard test, `make build`, and embedded-file
regeneration) is run-phase work owned by manager-develop. This SPEC describes
observable behaviors and mechanically checkable acceptance criteria so the run
phase has an unambiguous target.

Baseline: HEAD `5042e309c`. A parallel session has renamed the sync-phase auditor
agent `evaluator-active` → `sync-auditor` across the agent catalog. The 7 retained
MoAI agents at this baseline are: manager-spec, manager-develop, manager-docs,
manager-git, plan-auditor, **sync-auditor**, builder-harness. The two read-only
evaluators (`plan-auditor`, `sync-auditor`) correctly do NOT declare `TodoWrite`
and are therefore out of the H4 migration scope.

### Why this matters

- **Correctness for downstream users**: every `moai init` / `moai update` user
  receives the template mirror of these agents. Shipping a non-existent tool name
  (`MultiEdit`) and a default-disabled tool (`TodoWrite`) in the deployed catalog
  means new agents inherit a broken tool surface.
- **Self-consistency**: CLAUDE.md already mandates the Task* family for
  implementation agents (§3, §14), but the agent frontmatter declares `TodoWrite`.
  The instruction layer contradicts itself.
- **Regression prevention**: there is currently no CI guard validating agent /
  skill `tools:` names against the live Claude Code tool catalog. A guard test
  closes the gap so a future edit cannot re-introduce `MultiEdit` or a bare
  `TodoWrite`.

## B. Verified Findings (run-phase remediation scope)

All line numbers verified at HEAD `5042e309c` during plan authoring. The run
phase MUST Grep/Read each file to re-confirm exact lines before editing (a
parallel session may shift lines by a few).

### B.1 Finding H3 — retired `MultiEdit` tool still declared

`MultiEdit` appears in the following verified locations:

| Location | Line(s) | Kind |
|----------|---------|------|
| `.claude/agents/moai/manager-spec.md` | 13 | `tools:` frontmatter |
| `.claude/agents/moai/manager-develop.md` | 21 | `tools:` frontmatter |
| `.claude/agents/moai/manager-develop.md` | 39, 45 | hook `matcher: "Write\|Edit\|MultiEdit"` |
| `.claude/agents/moai/manager-spec.md` | 130 | body: "[HARD] Use MultiEdit for simultaneous 3-file creation" |
| `.claude/agents/moai/manager-spec.md` | 140, 153, 207 | body: incidental "Write/MultiEdit" mentions in the SPEC ID self-check protocol |
| `internal/template/templates/.claude/agents/moai/manager-spec.md` | (mirror of all the above) | template mirror |
| `internal/template/templates/.claude/agents/moai/manager-develop.md` | (mirror of all the above) | template mirror |
| `internal/template/templates/.claude/settings.json.tmpl` | 81 | PostToolUse `matcher: "Write\|Edit\|MultiEdit"` |
| `internal/template/templates/.claude/settings.json.tmpl` | 383 | `permissions` array entry `"MultiEdit"` |

Required run-phase change:

- Remove `MultiEdit` from the `tools:` frontmatter of manager-spec and
  manager-develop (source + template mirror).
- Replace the manager-spec body "[HARD] Use MultiEdit for simultaneous 3-file
  creation (60% faster than sequential)" instruction with a "make parallel
  Edit/Write calls in a single turn" instruction (source + template mirror).
- Correct the manager-spec body incidental "Write/MultiEdit" mentions (lines 140,
  153, 207) to "Write" (or "Write/Edit") so the SPEC ID self-check protocol no
  longer references the retired tool (source + template mirror).
- Drop `MultiEdit` from the manager-develop hook matchers (lines 39, 45 → `"Write|Edit"`)
  and from the settings.json.tmpl matcher (line 81 → `"Write|Edit"`) and
  permissions array (line 383, remove the `"MultiEdit"` entry).
- Run `make build`.

### B.2 Finding H4 — `TodoWrite` disabled by default (v2.1.142+); migrate to Task* family

All 5 implementation-capable retained agents declare `TodoWrite` in `tools:`
(verified at HEAD `5042e309c`):

| Agent | Source line | Template-mirror line |
|-------|-------------|----------------------|
| `manager-spec` | 13 | 13 |
| `manager-develop` | 21 | 21 |
| `manager-docs` | 13 | 13 |
| `manager-git` | 12 | 12 |
| `builder-harness` | 11 | 11 |

The two read-only evaluators (`plan-auditor`, `sync-auditor`) correctly do NOT
declare `TodoWrite` and are NOT in scope.

`.claude/rules/moai/development/agent-authoring.md` line 212 (and its template
mirror at the same line) recommends `TodoWrite` in the manager-agent tool list:
"Manager agents: Read, Write, Edit, Grep, Glob, Bash, Skill, TodoWrite (…)".

Required run-phase change:

- Replace `TodoWrite` with `TaskCreate, TaskUpdate, TaskList, TaskGet` in the
  `tools:` frontmatter of all 5 implementation-capable agents (source + template
  mirror).
- Update the agent-authoring.md line-212 recommendation to the Task* family
  (source + template mirror). See §E CON-5 — this line is ALSO a target of the
  sibling SPEC-CCSYNC-CLAUDEMD-001 (which fixes the co-located teammates-spawn +
  v2.1.50 claim); sibling runs FIRST and this SPEC rebases onto its commit.
- Run `make build`.

### B.3 Low-priority bundle

#### B.3.1 — foundation-cc authoring-kit scrub

The `moai-foundation-cc` skill bundles two kinds of reference material that
mention `MultiEdit` / `TodoWrite`:

1. **MoAI-authored authoring-kit examples** (teach how to write an agent's
   `tools:` line — these SHOULD be scrubbed so the authoring reference does not
   teach retired tool names):
   - `reference/skill-formatting-guide.md` (line 144: `allowed-tools: … MultiEdit`)
   - `reference/sub-agents/sub-agent-examples.md` (lines 28, 196, 437, 934)
   - `reference/sub-agents/sub-agent-formatting-guide.md` (lines 154, 390, 848, 892)
   - `reference/sub-agents/sub-agent-integration-patterns.md` (line 768)
   - (source + template mirror for each)

2. **Verbatim official-docs reproductions** (document what the Claude Code
   catalog historically contained — these are OUT OF SCOPE; see EXC-3):
   - `reference/claude-code-iam-official.md` (lines 24, 89)
   - `reference/claude-code-sub-agents-official.md` (line 334)
   - `reference/claude-code-plugins-official.md` (line 190)

Required run-phase change: scrub `MultiEdit` and bare `TodoWrite` from the
MoAI-authored authoring-kit examples only (category 1), source + template mirror.
For `TodoWrite` in authoring examples, replace with the Task* family or remove;
for `MultiEdit`, remove from the example `tools:`/`allowed-tools:` lines. Do NOT
edit the verbatim official-docs reproductions (category 2).

#### B.3.2 — new CI guard test

There is no Go CI guard validating agent / skill `tools:` / `allowed-tools:`
names against the current Claude Code tool catalog. Required run-phase change: add
a Go test under `internal/template/` (alongside the existing
`agent_frontmatter_audit_test.go`, `commands_audit_test.go`,
`rule_template_mirror_test.go` audit tests) that:

- Walks the embedded agent files (and SHOULD also cover skill `allowed-tools:`).
- REJECTS any `tools:` / `allowed-tools:` entry equal to `MultiEdit` (hard fail —
  retired tool).
- FLAGS `TodoWrite` as default-disabled (hard fail when it appears in a retained
  agent's `tools:`, since the migration target is the Task* family).
- Documents the allowlist source (the current Claude Code tool catalog) in a
  code comment so a future maintainer can update it.

## C. Requirements (GEARS)

### Finding H3 requirements

- **REQ-CCSYNC-T-001** (Ubiquitous): The `tools:` frontmatter of `manager-spec`
  and `manager-develop` (source + template mirror) shall not contain `MultiEdit`.
- **REQ-CCSYNC-T-002** (Ubiquitous): The manager-spec body shall instruct the
  agent to make parallel `Edit`/`Write` calls in a single turn for multi-file
  creation, and shall not instruct the agent to "Use MultiEdit" (source + template
  mirror).
- **REQ-CCSYNC-T-003** (Ubiquitous): No agent file and no `settings.json.tmpl`
  (source + template mirror) shall reference `MultiEdit` in a hook matcher or a
  permissions array.
- **REQ-CCSYNC-T-004** (When the manager-spec SPEC ID self-check protocol mentions
  a write tool): When the manager-spec body references the write tool in its SPEC
  ID self-check protocol, the reference shall name `Write` (or `Write`/`Edit`), not
  `MultiEdit` (source + template mirror).

### Finding H4 requirements

- **REQ-CCSYNC-T-005** (Ubiquitous): The `tools:` frontmatter of every
  implementation-capable retained agent (`manager-spec`, `manager-develop`,
  `manager-docs`, `manager-git`, `builder-harness`) shall declare the Task* family
  (`TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`) and shall not declare
  `TodoWrite` (source + template mirror).
- **REQ-CCSYNC-T-006** (Ubiquitous): The read-only evaluator agents
  (`plan-auditor`, `sync-auditor`) shall remain unchanged with respect to
  task-tracking tools (they declare neither `TodoWrite` nor the Task* family).
- **REQ-CCSYNC-T-007** (Ubiquitous): The agent-authoring.md manager-agent tool-list
  recommendation (source + template mirror) shall recommend the Task* family and
  shall not recommend `TodoWrite`.

### Low-priority bundle requirements

- **REQ-CCSYNC-T-008** (Where a foundation-cc reference is a MoAI-authored
  authoring-kit example): Where a `moai-foundation-cc` reference file teaches how
  to author an agent's `tools:`/`allowed-tools:` line, that example shall not name
  `MultiEdit` and shall not name a bare `TodoWrite` (source + template mirror).
- **REQ-CCSYNC-T-009** (Ubiquitous): A Go CI guard test under `internal/template/`
  shall validate agent (and SHOULD validate skill) `tools:`/`allowed-tools:` names
  against the current Claude Code tool catalog allowlist — failing on any
  `MultiEdit` entry and on any retained-agent `TodoWrite` entry — and shall
  document the allowlist source in a code comment.

### Build + mirror requirements (cross-cutting)

- **REQ-CCSYNC-T-010** (When the template source changes): When any file under
  `internal/template/templates/` is modified, the build shall regenerate
  `internal/template/embedded.go` via `make build` so the embedded copy matches
  the source.
- **REQ-CCSYNC-T-011** (Where a file exists in both `.claude/**` and its
  `internal/template/templates/` mirror): Where a remediated file has a template
  mirror (every agent file, agent-authoring.md, each foundation-cc reference),
  both copies shall be edited in the same commit so byte-parity invariants
  (`internal/template/rule_template_mirror_test.go` and the embedded mirror
  guards) hold.

### Neutrality requirement (cross-cutting)

- **REQ-CCSYNC-T-012** (Where content is written into the template): Where any
  text is written into `internal/template/templates/`, the orchestrator shall
  ensure it contains no moai-adk-internal SPEC IDs, REQ tokens, AC tokens, audit
  citations, internal dates, commit SHAs, archive paths, or memory-hash references
  (per CLAUDE.local.md §25 Template Internal-Content Isolation).

## D. Exclusions (What NOT to Build)

- **EXC-1 — Read-only evaluators are NOT migrated.** `plan-auditor` and
  `sync-auditor` declare no task-tracking tool; this SPEC does not add the Task*
  family to them. They are read-only and do not need task tracking.
- **EXC-2 — No agent catalog restructuring.** This SPEC does not add, remove,
  rename, or re-scope any agent. It edits only the `tools:` line + the related
  body/matcher references of existing agents.
- **EXC-3 — Verbatim official-docs reproductions are OUT OF SCOPE.** The
  foundation-cc `claude-code-iam-official.md`, `claude-code-sub-agents-official.md`,
  and `claude-code-plugins-official.md` reference files document the historical
  Claude Code catalog verbatim and MUST NOT be edited. Only the MoAI-authored
  authoring-kit examples (B.3.1 category 1) are scrubbed.
- **EXC-4 — No edit to the agent-authoring.md teammates-spawn / version claim.**
  That co-located claim on line 212 ("teammates CAN spawn other teammates …
  v2.1.50+") belongs to the sibling SPEC-CCSYNC-CLAUDEMD-001 (M3.2). This SPEC
  edits ONLY the `TodoWrite` → Task* recommendation portion of line 212. See
  CON-5 sequencing.
- **EXC-5 — No CLAUDE.md edits.** The CLAUDE.md §3/§14 Task*-family mandate is
  already correct; this SPEC aligns the agent frontmatter TO it. CLAUDE.md itself
  is out of scope here (the sibling owns CLAUDE.md drift).
- **EXC-6 — No broad tool-surface redesign.** This SPEC does not add new tools
  beyond the Task* family, does not remove tools other than `MultiEdit`, and does
  not reorder existing tool entries except as required to drop/add the named tools.
- **EXC-7 — This SPEC does not perform the edits.** Plan-phase authoring only. All
  file modifications, the new test file, `make build`, and embedded regeneration
  are run-phase work owned by manager-develop.

## E. Constraints

- **CON-1**: Template-First Rule (CLAUDE.local.md §2). Template source under
  `internal/template/templates/` is edited; `make build` regenerates
  `embedded.go`. `embedded.go` is never hand-edited.
- **CON-2**: Mirror parity (`internal/template/rule_template_mirror_test.go` +
  embedded mirror guards). Every agent file and rule file with a `templates/`
  mirror must be edited in both copies in the same commit.
- **CON-3**: Template Internal-Content Isolation (CLAUDE.local.md §25). Any text
  written into the template stays generic.
- **CON-4**: Hybrid Trunk Tier M (CLAUDE.local.md §23.9). main-direct, no PR,
  unless the user explicitly escalates with `--pr`.
- **CON-5 — Shared-file sequencing with SPEC-CCSYNC-CLAUDEMD-001.**
  `agent-authoring.md` line ~212 is a run-phase target of BOTH SPECs:
  - The sibling SPEC-CCSYNC-CLAUDEMD-001 (M3.2) reconciles the co-located
    "teammates CAN spawn other teammates … v2.1.50+" claim.
  - THIS SPEC edits the `TodoWrite` → Task* recommendation on the same line.

  **The sibling SPEC-CCSYNC-CLAUDEMD-001 runs FIRST; this SPEC rebases onto its
  agent-authoring.md commit** before making its own edit. Running them
  concurrently would cause an intra-batch edit collision on the same line in the
  same file (source + mirror). The orchestrator MUST confirm the sibling is
  merged/committed (not mid-run) at this SPEC's run-phase pre-flight check. See
  plan.md §H for the detailed sequencing language (aligned with the sibling's
  plan.md §H).

## F. Verification Strategy

Every requirement maps to a mechanically checkable acceptance criterion in
acceptance.md: `MultiEdit`-absence greps returning 0, Task*-presence greps,
`TodoWrite`-absence greps on the 5 migrated agents, settings.json.tmpl
`MultiEdit`-absence greps, the new guard test passing
(`go test ./internal/template/... -run TestToolCatalog…`), `make build` clean,
`git diff --exit-code internal/template/embedded.go` showing the regenerated
mirror, and `go test ./internal/template/...` green (mirror-drift + neutrality +
leak). See acceptance.md for the full Given-When-Then matrix.

## G. Cross-References

- Baseline HEAD: `5042e309c` (sync-auditor rename).
- Sibling SPEC sharing `agent-authoring.md`: `SPEC-CCSYNC-CLAUDEMD-001` (see CON-5
  + plan.md §H). That SPEC runs first.
- CLAUDE.md §3 Phase 3 Execute + §14 Parallel Execution Safeguards — the
  Task*-family mandate this SPEC aligns the frontmatter to.
- Template-First doctrine: CLAUDE.local.md §2.
- Template isolation doctrine: CLAUDE.local.md §25.
- Tier-based PR routing: CLAUDE.local.md §23.9.
- Existing audit-test neighbors: `internal/template/agent_frontmatter_audit_test.go`,
  `internal/template/commands_audit_test.go`,
  `internal/template/rule_template_mirror_test.go`.
