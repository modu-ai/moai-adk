---
description: "Shared concurrency budget for sync fan-out and verification work"
paths: ".claude/skills/moai/workflows/sync/**,.claude/workflows/sync-audit-4dim.js"
---

# Workflow Resource Budget Contract

The orchestrator owns one `max_concurrency` budget for a phase turn. It is the
minimum of the runtime subagent cap, the configured workflow cap, and the
available slots reported at launch. Drafters, MX shards, four-dimensional
judges, and other read-only work all consume the same budget; their individual
fan-out sizes are not additive reservations.

Before launch, record `resource_plan` with `max_concurrency`, `queued`, and
`active`. Schedule work through a bounded queue. A task that cannot fit waits
in `queued` rather than being launched optimistically and relying on runtime
failure. Record `peak_concurrency`, queue wait, and any admission rejection in
the phase evidence.

Priority order when slots are scarce is: (1) safety/structure gates, (2)
verification needed to make the sync decision, (3) documentation drafts, and
(4) advisory judges. This order limits peak load without weakening a blocking
gate or pretending that an unstarted check passed.
