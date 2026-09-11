---
description: "Direct-consumer contract for the shared sync diagnostic snapshot"
paths:
  - ".claude/hooks/moai/sync-phase-quality-gate.sh"
  - ".claude/workflows/sync-audit-4dim.js"
  - ".claude/agents/moai/sync-auditor.md"
  - ".claude/skills/moai/workflows/sync/**"
---

# Shared snapshot consumer contract

The shared diagnostic snapshot is evidence only when the consumer runs
`moai verify check --key-current` for the current tree and preserves the
command result. A consumer records `hit`, `miss`, or `unavailable`; an
unavailable/miss result is an evidence gap and never a PASS.

Required consumers:

1. The sync Stop hook queries the key before its fast checks and records the
   status with the HEAD SHA in the gate log.
2. The 4-dimension audit Context step queries the key once, includes the exact
   command and output in its typed context, and passes that context to Judges.
3. The fallback `sync-auditor` runs the same read-only query before scoring and
   cites the exact output (or reports an evidence gap).

Consumers must not invent a snapshot hit from a path or cached report alone.
If the tree key changes, the previous result is not reusable.
