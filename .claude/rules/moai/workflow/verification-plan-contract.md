---
description: "De-duplicate sync verification commands by attributable execution key"
paths:
  - ".claude/skills/moai/workflows/run/**"
  - ".claude/skills/moai/workflows/sync/**"
  - ".claude/rules/moai/workflow/**"
---

# Verification plan and de-duplication contract

Every test, coverage, lint, type-check, and build invocation registers a
verification key:

```text
(tree_key, command, toolchain, environment, scope)
```

`tree_key` includes the commit and working-tree state. A verification plan has
one owner per key. Before starting a command, the owner queries the shared
snapshot/ledger; an existing COMPLETE result with the exact same key may be
reused and cited instead of running the command again.

Reuse is denied when any key field differs, the result is incomplete/failed,
the command generated or changed files, or the recorded environment/toolchain
is unknown. A writer changes `tree_key` and invalidates prior results for the
affected scope. A coverage run may be combined with a test run only when the
toolchain produces both attributable results; otherwise they remain separate
keys. The report records `owner`, `key`, `reused_from`, or `rerun_reason` so a
reviewer can see why a command ran once or again.
