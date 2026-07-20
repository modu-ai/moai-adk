---
id: SPEC-MOAI-WORKFLOW-SCHEDULE-001
title: "MoAI-Workflow: User-Defined Workflow Save + Scheduled Execution"
version: "0.1.0"
status: completed
created: 2026-07-17
updated: 2026-07-17
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".moai/workflows"
lifecycle: spec-anchored
tags: "workflow, scheduler, cron, loop, cadence, template"
tier: M
---

# SPEC-MOAI-WORKFLOW-SCHEDULE-001 — MoAI-Workflow: User-Defined Workflow Save + Scheduled Execution

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-07-17 | manager-spec | Initial plan-phase authoring (draft). GEARS format, Tier M. |

## §A Context and Motivation

MoAI-ADK users repeatedly perform the same recurring, read-mostly tasks — drift watching, nightly lean reviews, backlog re-discovery, per-domain health scans — but nothing lets them **capture such a task once and have it recur on a schedule**. The runtime already ships a native scheduler (Cron tools + the `/loop` interval scheduler), and the `cadence-bridge.md` doctrine already sanctions composing that scheduler with read-only `/moai` entry points. What is missing is a **user-facing, user-authored layer**: a simple way to write a recurring task as a file and register it on the schedule.

MoAI-Workflow fills that gap. A user saves a recurring task as a Markdown file with YAML frontmatter under `.moai/workflows/`, declares when it should run, and the workflow runs on the native scheduler. The feature is deliberately the **simple markdown layer** — it does NOT introduce a new execution engine, a daemon, or OS-level cron; it consumes the Claude Code native scheduler as-is and inherits the `cadence-bridge.md` read-only safety invariant.

### §A.1 Fixed design decisions (user-confirmed — not reopened by this SPEC)

1. **Execution engine**: Claude Code native scheduler ONLY — Cron tools (CronCreate / CronList / CronDelete) plus the `/loop` interval scheduler. NO Go daemon, NO OS-level cron/launchd is introduced by this SPEC.
2. **File format**: Markdown body + YAML frontmatter at `.moai/workflows/<name>.md`. The frontmatter declares at minimum `name`, `description`, `schedule` (interval-or-cron expression + a `mechanism: cron|loop` discriminator), and `safety` (`read-only` default | `write` explicit opt-in). The body is natural-language step instructions the orchestrator follows.
3. **Safety boundary**: default `read-only`/advisory, per the `cadence-bridge.md` catalog invariant — scheduled runs never commit, never push, never enter run-phase. `safety: write` is an explicit per-workflow opt-in; even then, Implementation Kickoff Approval-class human gates remain cadence-unsatisfiable (a scheduled run cannot substitute a human gate).

## §B Requirements (GEARS)

### §B.1 Workflow file format and storage

- **REQ-MWS-001** (Ubiquitous): The system shall store each user-defined workflow as a single Markdown file with a YAML frontmatter block at the path `.moai/workflows/<name>.md`.
- **REQ-MWS-002** (Ubiquitous): The workflow frontmatter shall declare at minimum the fields `name`, `description`, `schedule`, and `safety`.
- **REQ-MWS-003** (Ubiquitous): The `schedule` field shall carry an interval-or-cron expression together with a `mechanism` discriminator whose value is `cron` or `loop`.
- **REQ-MWS-004** (Ubiquitous): The `safety` field shall carry the value `read-only` or `write`.
- **REQ-MWS-005** (Ubiquitous): The workflow body (the Markdown content below the frontmatter) shall be natural-language step instructions that the orchestrator follows when the workflow runs.
- **REQ-MWS-006** (Where): Where the frontmatter omits the `safety` field, the system shall treat the workflow as `read-only`.

### §B.2 Creation entry surface

- **REQ-MWS-007** (Event-driven): When a user invokes the `/moai workflow` creation entry surface, the orchestrator shall conduct an `AskUserQuestion`-guided capture of `name`, `description`, `schedule` (expression + mechanism), `safety`, and the step body, then write the resulting workflow file to `.moai/workflows/<name>.md`.
- **REQ-MWS-008** (Event-driven): When the creation flow assembles a candidate workflow, the orchestrator shall validate the `schedule` expression and `mechanism` value, and the `safety` value, before writing the file.
- **REQ-MWS-009** (Where): Where a natural-language workflow-capture request arrives without invoking the `/moai workflow` surface, the orchestrator shall route it to the same guided-capture path so a single creation code-path serves both entry forms.

### §B.3 Schedule registration and unregistration

- **REQ-MWS-010** (Event-driven): When a workflow declaring `mechanism: cron` is created, the orchestrator shall register its schedule with the native Cron tool (CronCreate) using the declared interval-or-cron expression.
- **REQ-MWS-011** (Event-driven): When a workflow declaring `mechanism: loop` is created, the orchestrator shall guide the user to arm the session-scoped native `/loop` for that workflow, disclosing that a `loop`-mechanism schedule is session-scoped and must be re-armed each session; the workflow file shall record the declared loop schedule (record-only), and at session start the orchestrator shall surface a re-arm reminder for any `mechanism: loop` workflow present while the user re-arms `/loop` manually (no orchestrator auto-arm), per DECISION-MWS-D2 (plan.md §D).
- **REQ-MWS-012** (Event-driven): When a workflow is removed, the orchestrator shall unregister its schedule — CronDelete for a `cron` mechanism, session-scoped loop cancellation for a `loop` mechanism — before deleting the workflow file.
- **REQ-MWS-013** (Ubiquitous): The registration and unregistration notices the orchestrator surfaces shall reuse the harness-v4 schedule vocabulary (interval + mechanism naming; CronDelete for cron; session-scoped loop cancellation for loop) rather than forking a new vocabulary.

### §B.4 Safety boundary (cadence invariant)

- **REQ-MWS-014** (Ubiquitous, HARD): A scheduled workflow run shall not commit, shall not push, and shall not enter run-phase, per the `cadence-bridge.md` catalog invariant.
- **REQ-MWS-015** (Ubiquitous): A scheduled workflow run shall be bounded to read-only or advisory actions, or at most Level-1 uncommitted working-tree edits left uncommitted and unpushed, regardless of the workflow's declared `safety` tier.
- **REQ-MWS-016** (Ubiquitous): The `safety: write` tier shall govern interactive (manual, user-present) invocation permissions only; it shall not relax the scheduled-run cadence invariant of REQ-MWS-014 and REQ-MWS-015.
- **REQ-MWS-017** (When): When a workflow step would require an Implementation Kickoff Approval-class human gate, a scheduled run shall not satisfy that gate; it shall surface the pending decision to the queue for the next interactive session instead.

### §B.5 Discovery and lifecycle

- **REQ-MWS-018** (Event-driven): When a user requests the workflow list, the orchestrator shall enumerate the `.moai/workflows/*.md` files and render each one's `name`, `description`, and `schedule` (interval + mechanism); a workflow that declares no schedule shall render without schedule detail.
- **REQ-MWS-019** (Event-driven): When a user requests to edit a workflow, the orchestrator shall open the workflow Markdown file for direct modification, treating the frontmatter as the single source of truth for that workflow.
- **REQ-MWS-020** (Event-driven): When a user requests to remove a workflow, the orchestrator shall first unregister its schedule (REQ-MWS-012) and then delete the workflow file, and shall report the outcome; when the file is absent the removal shall report a no-op rather than fail hard.

### §B.6 Template deployment

- **REQ-MWS-021** (Ubiquitous): The template source shall provide a `internal/template/templates/.moai/workflows/` scaffold containing a `README.md` that explains the workflow file format and one neutral example workflow file.
- **REQ-MWS-022** (Where): Where the scaffold content is authored, it shall contain no internal SPEC IDs, no internal dates, and no commit SHAs, so that the 16-language template distribution stays neutral to MoAI-ADK internal development state.
- **REQ-MWS-023** (Ubiquitous): Template scaffold changes shall be made in the template source first and then embedded via `make build`, per the Template-First rule.

### §B.7 Boundary and non-duplication

- **REQ-MWS-024** (Ubiquitous): This feature shall be the user-facing simple markdown-workflow layer only, distinct and non-duplicative of the harness-v4 manifest schedule, the `.claude/workflows/*.js` dynamic-workflow scripts, the native `/loop` scheduler, and the `cadence-bridge.md` recipe catalog.

## §C Boundary Definition vs Existing Assets

This feature MUST NOT duplicate any existing scheduling or workflow asset. The layer boundaries are:

| Asset | What it is | Audience | Relationship to this feature |
|-------|-----------|----------|------------------------------|
| harness-v4 manifest `schedule` | A schedule field declared inside a **built harness's** manifest (`/moai:harness`), harness-scoped; list/remove verbs print its interval + mechanism | Harness authors (maintainers) | This feature **reuses its vocabulary** (interval + mechanism, CronDelete / session-scoped cancellation) but is NOT harness-scoped — a MoAI-Workflow is a standalone user file, not part of a generated harness. |
| `.claude/workflows/*.js` | Dynamic-Workflow **scripts** (JS orchestrating dozens-to-hundreds of subagents) | Advanced orchestration authors | Disjoint: this feature is markdown natural-language steps, not executable JS orchestration. |
| native `/loop <interval>` | The runtime's raw interval scheduler primitive | Any user, ad hoc | This feature **consumes** `/loop` as the `mechanism: loop` execution primitive; it adds the persisted file + guided capture on top. |
| `cadence-bridge.md` recipes | A fixed, doctrine-authored **read-only recipe catalog** composing `/loop`/Cron with read-only `/moai` entry points | Doctrine / orchestrator | This feature **inherits its HARD safety invariant** and lets users author their OWN recurring tasks, where cadence-bridge only offers a fixed sanctioned set. |
| **This feature** (`.moai/workflows/*.md`) | User-authored recurring-task markdown files, registered on the native scheduler | End users | The simple user-facing layer. |

The distinguishing axis: **who authors it and where it lives.** cadence-bridge recipes are fixed and doctrine-owned; harness schedules live inside a generated harness manifest; `.js` workflows are executable scripts. MoAI-Workflow files are user-authored plain-markdown recurring tasks in a dedicated `.moai/workflows/` directory — nothing else occupies that niche.

## §D Acceptance Criteria

The full Given-When-Then acceptance matrix, edge cases, and Definition of Done are enumerated in `acceptance.md` (AC-MWS-001 … AC-MWS-0NN). Each REQ-MWS-xxx above maps to at least one AC-MWS-xxx.

## §E Exclusions

The following are deliberately NOT built by this SPEC. Each is expressed as an `### Out of Scope` sub-heading so the exclusion set is explicit.

### Out of Scope — Execution engine internals

- No Go daemon, no OS-level `cron`/`launchd`, no background process. The Claude Code native scheduler (Cron tools + `/loop`) is the ONLY execution engine and is consumed as-is.
- No modification to how the native scheduler injects a scheduled turn or evaluates a `/loop` interval — that is runtime behavior this feature sits on top of.

### Out of Scope — `moai workflow` Go CLI subcommand

- A Go `moai workflow` CLI subcommand (list/remove filesystem operations) is deferred to a follow-up SPEC. In this SPEC, list/edit/remove are orchestrator-driven filesystem operations (Read/Glob/Edit/Write), because registration itself (Cron tools) is an orchestrator-only capability, so the whole lifecycle is orchestrator-driven and a Go subcommand would add code without covering the registration path.

### Out of Scope — Mechanical frontmatter validation

- No Go lint rule or `doctor`-style mechanical validator for `.moai/workflows/*.md` frontmatter. Schedule/mechanism/safety validation is orchestrator-side at creation time only (REQ-MWS-008). A mechanical validator is deferred.

### Out of Scope — Scheduling mutating subcommands

- Scheduling any write-capable or git-mutating `/moai` entry point (`/moai run`, `/moai sync`, `/moai loop`, `/moai fix` beyond the Level-1-no-commit carve-out) is explicitly unsanctioned per the `cadence-bridge.md` eligibility table and is NOT enabled by this feature, even for a `safety: write` workflow.

### Out of Scope — Cross-project / global workflows

- Workflows are project-local under `.moai/workflows/`. A global (`~/.claude`-level) or cross-project workflow registry is not in scope.

## §F Cross-References

- `.claude/rules/moai/workflow/cadence-bridge.md` — the read-only cadence invariant this feature inherits (catalog-level HARD invariant + eligibility table + Discovery-to-Queue contract).
- `.claude/skills/moai/SKILL.md` § Branch A.1 — the harness-v4 schedule vocabulary this feature reuses (interval + mechanism; CronDelete / session-scoped loop cancellation).
- `.claude/rules/moai/workflow/goal-directive.md` § Comparing Autonomous-Continuation Approaches — native `/loop` vs `/moai loop` distinction.
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 — template-neutrality forbidden/allowed content classes.
- `CLAUDE.local.md` §2 Template-First Rule — template-source-first + `make build` discipline.
