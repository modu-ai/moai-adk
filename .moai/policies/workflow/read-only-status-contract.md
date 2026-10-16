---
description: "Contract for side-effect-free /moai sync status mode"
paths:
  - ".claude/skills/moai/workflows/sync/**"
  - ".claude/hooks/moai/**"
  - ".claude/rules/moai/workflow/**"
---

# Read-only status mode contract

`/moai sync status` is a diagnostic mode, not a shortened write mode. At mode
entry the orchestrator MUST establish the run context
`mode=status`, `read_only=true`, and `writer_policy=deny` before dispatching
any gate or auditor.

## Permitted work

Status may read the working-tree state, current branch and commit, existing
diagnostic snapshots, configuration, and already-produced reports. It may
compute and display findings, but it MUST NOT repair or normalize them.

## Denied work

While `read_only=true`, the workflow MUST reject or skip all writers:

- formatter/linter/type-check auto-fix or any command with a write flag;
- @MX tag insertion, generated documentation, or other file edits;
- `git add`, `git commit`, `git tag`, `git push`, branch changes, or merges;
- write-capable manager-develop/manager-docs delegation;
- cache, lock, or report updates that alter tracked or user source files.

A tool that cannot prove read-only behavior is denied in status mode. A failed
read-only probe is a status error, not permission to fall through to auto-fix.

## Mutation proof and reporting

The status path captures `before_tree_key` (HEAD plus porcelain-v2 and diff
hash) before diagnostics and `after_tree_key` after diagnostics. It may say
"no changes" only when the keys are equal and the writer-deny counter is zero.
If they differ, the report MUST name the changed paths and classify the event
as an unexpected mutation; it MUST NOT claim a clean read-only run.
