---
description: "Resolve sync 4-dimension judge model effort from the verify-judge policy"
paths: ".claude/workflows/sync-audit-4dim.js,.claude/skills/moai/workflows/sync/**"
---

# Verify-Judge Effort Contract

The orchestrator resolves the `verify-judge` model profile and passes its
`judge_effort` to `sync-audit-4dim.js`. The workflow accepts only the runtime
policy values `low`, `medium`, and `high`; an absent or unsupported value
falls back to `high`. `xhigh` is not a valid judge input even if another
surface uses it for a different purpose. The resolved value is recorded with
the judge evidence so cost and quality comparisons use the actual profile.
