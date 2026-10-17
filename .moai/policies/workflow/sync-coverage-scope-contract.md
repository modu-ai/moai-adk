---
description: "Mode-aware scope contract for sync diagnostics and coverage"
paths:
  - ".claude/skills/moai/workflows/sync/**"
  - ".claude/rules/moai/workflow/**"
---

# Sync diagnostic scope contract

The sync mode selects the verification scope before launching diagnostics. The
scope is recorded as `sync_scope` and `coverage_scope` in the report; a command
must not silently widen a selective run to a full-repository run.

| Mode | Source/document scope | Coverage and test scope |
|---|---|---|
| `auto` | Changed files plus their directly impacted docs/API surface | Changed packages and the dependency/import closure needed to interpret them |
| `force` | All synchronizable documentation and source surfaces | Full repository, including every package selected by the toolchain |
| `project` | All project documentation and codemap surfaces | Full repository baseline and project-wide test/coverage checks |
| `status` | Read-only health inputs only | No writer or coverage-generation path; report existing evidence only |

An auto run may widen to `force`/`project` scope only after recording the
reason (for example, an unresolved impact graph or an explicit user flag). A
pre-existing coverage gap is reported separately from the changed-scope
result; it must not force an unannounced full scan or make an auto run claim
that the entire repository was revalidated.
