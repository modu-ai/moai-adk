---
description: >
  MoAI-Workflow — save a recurring, read-mostly task as a Markdown file under
  .moai/workflows/ and register it on the Claude Code native scheduler (Cron
  tools or the /loop interval scheduler). The simple user-facing markdown layer
  for scheduled discovery work; scheduled runs are cadence-bounded (read-only /
  advisory) and never commit, push, or enter run-phase.
  Use when creating, listing, editing, or removing a recurring workflow, or when
  a natural-language request asks to "save this as a recurring/nightly workflow".
user-invocable: false
metadata:
  category: "workflow"
  status: "active"
---

# /moai workflow — User-Defined Workflow Save + Scheduled Execution

> MoAI-Workflow is the **simple user-facing markdown layer** for recurring tasks.
> A user saves a recurring task as a Markdown file with YAML frontmatter under
> `.moai/workflows/<name>.md`, declares when it should run, and the workflow runs
> on the Claude Code **native scheduler** (Cron tools for `cron` mechanism, the
> native `/loop` interval scheduler for `loop` mechanism). This layer introduces
> **no new execution engine, no daemon, and no OS-level cron** — it consumes the
> native scheduler as-is and inherits the read-only cadence invariant from
> `.claude/rules/moai/workflow/cadence-bridge.md`.

## What It Is (and what it is NOT)

MoAI-Workflow is the user-authored recurring-task layer. It is distinct from four
sibling assets and MUST NOT duplicate any of them — it reuses their vocabulary
rather than re-forking it (see § Boundary — Non-Duplication):

- NOT the harness-v4 manifest `schedule` (harness-scoped, maintainer-facing).
- NOT the `.claude/workflows/*.js` dynamic-workflow scripts (executable JS orchestration).
- NOT the native `/loop` scheduler itself (this layer *consumes* `/loop` as the `loop` mechanism).
- NOT the `cadence-bridge.md` fixed recipe catalog (this layer lets users author *their own* recurring tasks; cadence-bridge only offers a fixed sanctioned set).

## Workflow File Data Model

A MoAI-Workflow is a single Markdown file at `.moai/workflows/<name>.md`: a YAML
frontmatter block followed by a Markdown body of natural-language step
instructions.

### Frontmatter Schema

The frontmatter declares at minimum four fields:

| Field | Required | Values | Meaning |
|-------|----------|--------|---------|
| `name` | yes | slug matching the file basename `<name>` | Workflow identifier; MUST equal `<name>.md` |
| `description` | yes | one-line string | What the workflow does |
| `schedule` | yes | mapping (see below) | When and how it runs |
| `safety` | no (default `read-only`) | `read-only` \| `write` | Interactive-invocation permission tier |

The `schedule` mapping carries an interval-or-cron expression together with a
`mechanism` discriminator:

```yaml
schedule:
  expression: "30m"        # an interval (loop) or a cron expression (cron)
  mechanism: loop          # exactly one of: cron | loop
```

- `mechanism: cron` — the `expression` is a cron expression; the schedule is
  registered persistently with the native Cron tool (CronCreate).
- `mechanism: loop` — the `expression` is a `/loop` interval (e.g. `30m`,
  `nightly`); the schedule is session-scoped and recorded record-only (a `/loop`
  dies with its session — see § Schedule Registration).

`mechanism` MUST be exactly `cron` or `loop`. Any other value (e.g. `daemon`) is
a validation failure at creation time (§ Creation — Validation).

### safety field

- `safety: read-only` (default when the field is omitted) — the workflow is
  advisory: it inspects, reports, and queues findings, and makes no write beyond
  the Level-1 uncommitted carve-out.
- `safety: write` — grants working-tree write permission **for interactive
  (user-present) invocation only**. It does NOT relax the scheduled-run cadence
  invariant (§ Safety Boundary): a scheduled run of a `safety: write` workflow is
  still cadence-bounded.

### Body

The Markdown body below the frontmatter is **natural-language step instructions**
that the orchestrator follows when the workflow runs. The body MUST be non-empty
— a frontmatter-only file (no steps) is a validation failure (§ Edge Cases).

### Example workflow file

```markdown
---
name: drift-watch
description: Run the quality gate and surface any new drift to the queue.
schedule:
  expression: "30m"
  mechanism: loop
safety: read-only
---

# Steps

1. Run `/moai gate` (lint + format + type-check + test, read-only validation).
2. If the gate reports any failure, record it to the cadence queue.
3. Surface the queued findings at the next interactive session. Do not fix.
```

## Safety Boundary (cadence invariant — HARD)

[HARD] A scheduled workflow run inherits the `cadence-bridge.md` catalog-level
HARD invariant verbatim. The single governing sentence, binding every scheduled
run of every workflow:

> **Scheduled runs never commit, never push, never enter run-phase; Level-1
> uncommitted working-tree edits are the sole permitted exception.**

Concretely, for any scheduled workflow run, regardless of the workflow's declared
`safety` tier:

- **No commit.** A scheduled run performs no `git commit`.
- **No push.** A scheduled run performs no `git push`.
- **No run-phase entry.** A scheduled run never enters `/moai run` or any
  write-capable / git-mutating `/moai` entry point.
- **Level-1 edits only, left uncommitted.** At most Level-1 working-tree edits
  (the `fix.md` Level 1 "immediate, no approval required" class — formatting,
  import-sorting) may be made, and they are left uncommitted and unpushed.

### `safety: write` governs interactive invocation only

The `safety: write` tier does NOT relax the cadence invariant above. `write`
grants working-tree write permission to an **interactive (user-present)**
invocation of the workflow; it has no effect on a scheduled (unattended) run,
which remains bounded exactly as a `read-only` workflow's scheduled run is
bounded. There is no workflow configuration that lets a scheduled run commit,
push, or enter run-phase.

### Human gates are cadence-unsatisfiable

When a workflow step would require an Implementation Kickoff Approval-class human
gate, a scheduled run does NOT satisfy that gate. It queues the pending decision
(§ Discovery-to-Queue) and surfaces it at the next interactive session. A
scheduled discovery is *input* to a human decision; it is never itself a decision.
This mirrors the `cadence-bridge.md` clause that the plan→run human gate is
human-only and cadence-unsatisfiable.

### Discovery-to-Queue

When a scheduled run finds work, it persists the finding to the active TaskList
when a session ledger is live, otherwise to a backlog record at
`.moai/reports/cadence/<date>.md`, and surfaces it at the next interactive
session. A scheduled run never auto-executes remediation.

## Creation — Guided Capture

### Entry surface

`/moai workflow` (creation intent) enters an `AskUserQuestion`-guided capture.
A natural-language capture request that does NOT type `/moai workflow` (e.g.
"save this as a recurring nightly workflow") routes to the **same** guided-capture
path — there is a single creation code-path serving both entry forms (the
`/moai:workflow` wrapper command and natural-language capture converge here).

### Guided capture flow (orchestrator)

The orchestrator captures the following via `AskUserQuestion` rounds
(`ToolSearch(query: "select:AskUserQuestion")` preload before each call, per the
AskUserQuestion protocol), then assembles and validates the candidate before
writing:

1. `name` — the workflow slug (also the file basename).
2. `description` — one line.
3. `schedule` — the `expression` and the `mechanism` (`cron` or `loop`).
4. `safety` — `read-only` (recommended default) or `write`.
5. Step body — the natural-language steps.

### Validation (before writing)

The orchestrator validates the candidate BEFORE writing the file. On any failure
the file is NOT written and the orchestrator surfaces the failure for correction:

- `mechanism` is exactly `cron` or `loop` (else validation failure, e.g. `daemon`).
- The `schedule.expression` is well-formed for its mechanism (a malformed cron
  expression for `cron`, or a malformed interval for `loop`, fails).
- `safety`, when present, is exactly `read-only` or `write`.
- The body is non-empty (frontmatter-only is rejected).
- `name` has no path separators or `..` (no directory traversal outside
  `.moai/workflows/`).

### Name collision — reject and re-prompt (DECISION-MWS-D1)

When `<name>.md` already exists, the orchestrator **rejects** the creation and
re-prompts the user to choose a different name. The existing file and its
registered schedule are NEVER silently overwritten and NEVER auto-suffixed
(`<name>-2.md` is not used). Overwrite would silently destroy a registered
schedule (data loss); auto-suffix would produce an unintended name while leaving
the collision unresolved. Reject-and-re-prompt keeps the user in control of
naming and preserves the existing workflow.

### Write

On successful validation, the orchestrator writes the assembled workflow file to
`.moai/workflows/<name>.md`, then registers its schedule (§ Schedule
Registration).

## Schedule Registration and Unregistration

Registration and unregistration notices reuse the harness-v4 schedule vocabulary
(interval + mechanism naming; CronDelete for cron; session-scoped loop
cancellation for loop) — this layer does NOT fork a new vocabulary.

### mechanism: cron — CronCreate (persistent)

When a `mechanism: cron` workflow is created, the orchestrator registers its
schedule with the native Cron tool (CronCreate) using the declared cron
expression, and reports the registration using interval + mechanism vocabulary.

### mechanism: loop — record-only + session-start reminder (DECISION-MWS-D2)

When a `mechanism: loop` workflow is created, the orchestrator guides the user to
arm the session-scoped native `/loop` for that workflow, disclosing that a
`loop`-mechanism schedule is **session-scoped and must be re-armed each session**
(a `/loop` dies with its session; there is no persistent registration). The
workflow file records the declared loop schedule (record-only).

At session start, the orchestrator surfaces a **re-arm reminder** for any
`mechanism: loop` workflow present, via natural-language status guidance (not
`AskUserQuestion` — it is an announcement), while the **user re-arms `/loop`
manually**. The orchestrator does NOT auto-arm the loop — auto-arming would resume
unbounded background scheduling the user did not re-consent to.

### Cron-unavailable fallback

Where the native Cron tools are unavailable in the runtime version, a
`mechanism: cron` creation degrades to guiding the user toward the session-scoped
`/loop` form rather than failing, consistent with the `cadence-bridge.md`
fallback clause.

### Unregistration (on removal)

When a workflow is removed, the orchestrator unregisters its schedule BEFORE
deleting the file:

- `mechanism: cron` → CronDelete (the unregister notice names the mechanism).
- `mechanism: loop` → session-scoped loop cancellation (the notice names it).

If the Cron registration was already externally deleted, CronDelete is a no-op;
the file is still deleted and the outcome is reported cleanly.

## Discovery and Lifecycle

Lifecycle operations (list / edit / remove) are orchestrator-driven filesystem
operations (Glob / Read / Edit / Write) — there is no Go `moai workflow` CLI
subcommand in this layer (registration itself is an orchestrator-only Cron-tool
capability, so the whole lifecycle stays orchestrator-driven).

### list

When the user requests the list, the orchestrator enumerates `.moai/workflows/*.md`
and renders each workflow's `name`, `description`, and `schedule` (interval +
mechanism). A workflow that declares no schedule renders without schedule detail
(no error).

### edit

When the user requests to edit a workflow, the orchestrator opens the workflow
Markdown file for direct modification, treating the frontmatter as the single
source of truth for that workflow.

### remove

When the user requests to remove a workflow, the orchestrator first unregisters
its schedule (§ Unregistration) and then deletes the workflow file, reporting the
outcome. When the file is absent, the removal reports a **no-op** rather than
failing hard.

## Edge Cases

- **Name with path separators or `..`** — rejected at creation (no directory
  traversal outside `.moai/workflows/`).
- **Empty body (frontmatter only)** — rejected; a workflow must carry step
  instructions.
- **`mechanism: cron` with a malformed cron expression** — validation failure.
- **Removing a workflow whose Cron registration was already externally deleted**
  — CronDelete no-op, file still deleted, reported cleanly.
- **Name collision on create** — rejected and re-prompted; the existing file and
  its registered schedule are never overwritten or auto-suffixed.

## Boundary — Non-Duplication

MoAI-Workflow is the distinct user-facing markdown layer. The distinguishing
axis is **who authors it and where it lives**:

| Asset | Audience | Relationship |
|-------|----------|--------------|
| harness-v4 manifest `schedule` | Harness authors (maintainers) | Vocabulary reused (interval + mechanism; CronDelete / session-scoped cancellation); NOT harness-scoped. |
| `.claude/workflows/*.js` | Advanced orchestration authors | Disjoint — markdown natural-language steps, not executable JS. |
| native `/loop <interval>` | Any user, ad hoc | Consumed as the `loop` mechanism primitive; this layer adds the persisted file + guided capture on top. |
| `cadence-bridge.md` recipes | Doctrine / orchestrator | HARD safety invariant inherited; this layer lets users author their OWN recurring tasks (cadence-bridge offers only a fixed set). |
| **MoAI-Workflow** (`.moai/workflows/*.md`) | End users | The simple user-facing layer — nothing else occupies this niche. |

For the common recurring cases (drift watch, nightly lean review, backlog
re-discovery), point users at the fixed `cadence-bridge.md` recipes rather than
re-authoring them as a MoAI-Workflow.

## Cross-References

- `.claude/rules/moai/workflow/cadence-bridge.md` — the read-only cadence
  invariant this layer inherits (catalog-level HARD invariant + eligibility table
  + Discovery-to-Queue contract).
- `.claude/skills/moai/SKILL.md` § harness Branch A.1 — the harness-v4 schedule
  vocabulary this layer reuses (interval + mechanism; CronDelete / session-scoped
  loop cancellation).
- `.claude/rules/moai/workflow/goal-directive.md` § Comparing
  Autonomous-Continuation Approaches — native `/loop` vs `/moai loop` distinction.
- `.claude/rules/moai/core/askuser-protocol.md` — the guided-capture
  `AskUserQuestion` preload + channel-monopoly rules.
