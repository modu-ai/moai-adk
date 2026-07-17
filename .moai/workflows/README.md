# MoAI-Workflows

This directory holds **user-defined workflows** — recurring, read-mostly tasks
you save once and run on a schedule. Each workflow is a single Markdown file:
a YAML frontmatter block plus a Markdown body of natural-language steps.

Create, list, edit, or remove workflows with the `/moai workflow` command
(or just describe a recurring task in natural language — "save this as a
recurring nightly workflow" — and it routes to the same guided flow).

## File format

Save each workflow as `.moai/workflows/<name>.md`:

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

1. Run the read-only quality gate.
2. If it reports any failure, record it to the queue.
3. Surface the queued findings at the next interactive session. Do not fix.
```

### Frontmatter fields

| Field | Required | Values | Meaning |
|-------|----------|--------|---------|
| `name` | yes | slug equal to the file basename | Workflow identifier |
| `description` | yes | one line | What the workflow does |
| `schedule` | yes | mapping (see below) | When and how it runs |
| `safety` | no (default `read-only`) | `read-only` \| `write` | Interactive-invocation permission tier |

The `schedule` mapping carries an `expression` and a `mechanism` discriminator:

```yaml
schedule:
  expression: "30m"   # a /loop interval (loop) or a cron expression (cron)
  mechanism: loop     # exactly one of: cron | loop
```

- `mechanism: cron` — the `expression` is a cron expression, registered
  persistently on the native scheduler.
- `mechanism: loop` — the `expression` is a `/loop` interval; the schedule is
  session-scoped and must be re-armed each session (a `/loop` dies with its
  session). At session start you are reminded to re-arm any `loop` workflow.

The Markdown body below the frontmatter is the natural-language steps the
workflow follows when it runs. It must be non-empty.

## Safety — scheduled runs are read-only

Scheduled workflow runs are **cadence-bounded**: a scheduled run never commits,
never pushes, and never enters run-phase. At most it makes small uncommitted
working-tree edits and leaves them uncommitted. This holds for every workflow,
regardless of its `safety` tier.

The `safety: write` tier grants working-tree write permission for an
**interactive** (you-present) invocation only — it does NOT let a scheduled
(unattended) run commit, push, or enter run-phase. A scheduled discovery is
input to a decision you make; it is never itself a decision.

## This is the simple layer

MoAI-Workflows is the simple, user-authored recurring-task layer. For common
recurring cases (drift watching, nightly lean review, backlog re-discovery),
prefer the fixed sanctioned recipes in the cadence-bridge doctrine rather than
re-authoring them here.
