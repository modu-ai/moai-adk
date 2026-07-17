---
name: example
description: Example read-only workflow — run the quality gate and queue any drift.
schedule:
  expression: "nightly"
  mechanism: loop
safety: read-only
---

# Steps

This is an example workflow you can copy, rename, and adapt. Delete it once you
have created your own.

1. Run the read-only quality gate (lint + format + type-check + test).
2. If the gate reports any failure, record the finding to the queue.
3. Surface the queued findings at the next interactive session. Do not apply
   any fix — a scheduled run is read-only and never commits or pushes.
