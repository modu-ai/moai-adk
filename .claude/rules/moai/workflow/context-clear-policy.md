---
description: "Evidence-based clear versus warm-context handoff policy"
paths:
  - ".claude/skills/moai/workflows/run/**"
  - ".claude/rules/moai/workflow/cache-aware-execution.md"
  - ".claude/rules/moai/workflow/**"
---

# Context clear policy

Choose `warm` or `clear` explicitly and record the decision in the handoff
ledger with `clear_reason`, `context_snapshot_id`, `plan_artifact_hash`,
`tree_key`, pending tasks, approvals, and the last evidence references.

| State | Use when | Required action |
|---|---|---|
| `warm` | Plan→run handoff has the same approved hash/tree, the follow-up is small, and the prefix remains valid | Preserve context; load only missing files and keep the existing cache |
| `clear` | Context is bloated with unrelated work, a phase needs audit independence, the prefix/rules changed, a model/effort switch is needed, or the context threshold is reached | Persist the handoff ledger, clear, reload the minimal context, and verify the handoff identity before acting |

Never clear across an approval gate without persisting the approval and its
scope. Never keep a warm context after a prefix edit or tree/plan identity
change. A `/clear` is not itself evidence of a clean handoff; the resumed
session must report which ledger fields it loaded and which it could not
verify.
