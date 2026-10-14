---
description: "Mutation-aware abort and graceful-exit reporting for sync"
paths:
  - ".claude/skills/moai/workflows/sync/**"
  - ".claude/rules/moai/workflow/**"
---

# Graceful exit and mutation contract

An abort is a control-flow outcome, not proof that no files changed. At sync
entry capture `before_tree_key`; after an abort or failed writer capture
`after_tree_key` and a path-level mutation list.

- **No mutation observed:** say “no changes” only when the keys are equal and
  the mutation list is empty.
- **Mutation observed:** report each applied operation/path, each unapplied
  operation, and whether a rollback was attempted and verified. Say
  “partial changes remain” when the keys differ, even if a rollback was
  requested but not verified.
- **Verified rollback:** may report that the affected paths were restored only
  when the post-rollback key and path list match the pre-mutation snapshot.

Every exit report includes the abort reason, current status, retry command, and
the evidence gap for any writer or rollback that could not be observed. Exit
code 0 does not downgrade a mutation into “no changes.”
