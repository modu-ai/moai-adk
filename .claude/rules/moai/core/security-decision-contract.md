---
description: "Single severity and exception contract for security gates"
paths: ".claude/skills/moai/workflows/sync/**,.claude/agents/moai/*auditor.md"
---

# Security Decision Contract

This is the single severity-to-decision contract for plan, run, sync, and
auditor outputs. A caller MUST reference this rule instead of redefining a
severity threshold locally.

| Severity | Default decision | Required handling |
|---|---|---|
| Critical | BLOCK | stop the owning phase and record the finding |
| High | BLOCK | stop the owning phase and record the finding |
| Medium | ADVISORY | continue only after recording the finding |
| Low | ADVISORY | record the finding; continuation is allowed |

`blocking` is a property of the decision contract, not a synonym for the
severity text. A finding may be downgraded only through an explicit,
user-approved exception record containing: finding ID, severity, rationale,
scope, approver, expiry, and the next review condition. The exception is
carried into the final report and does not rewrite the original finding.

The following rules are mandatory:

1. Critical and High findings never become a warning merely because they were
   discovered in a different phase. Phase 8, the sync-auditor, and downstream
   delivery all consume the same decision.
2. `Continue with warning` is not a default option for a blocking finding. It
   is offered only after the exception record is created and the user approves
   that exact finding.
3. A missing or stale scan is `UNVERIFIED`, not PASS. A manifest-change
   observation is not a vulnerability scan and cannot satisfy this contract.
4. The final verdict reports both the raw severity and the resolved decision,
   including the exception ID when one exists.
