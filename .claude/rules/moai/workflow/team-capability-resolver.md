---
description: "Single resolver for current Agent Teams capability versus genealogy"
paths:
  - ".claude/rules/moai/workflow/orchestration-mode-selection.md"
  - ".claude/skills/moai/workflows/run/mode-orchestration.md"
  - ".claude/rules/moai/workflow/spec-workflow.md"
---

# Team capability resolver

`team` is a dispatch-axis request, not proof that the current runtime can
create and return named teammates. Resolve it once per session with these
fields:

```text
team_requested: true|false
feature_flag: on|off|unset
runtime_probe: passed|failed|indeterminate
result: TEAM_AVAILABLE|MODE_TEAM_UNAVAILABLE|TEAM_NOT_REQUESTED
```

The resolver rules are:

1. Without an explicit `--team`, `--mode team`, or `Team` scale request, return
   `TEAM_NOT_REQUESTED`; never auto-select team from harness tier or a stale
   configuration flag.
2. For an explicit request, probe the current runtime's documented team
   surface and perform a bounded named-worker/result-return capability check.
   The environment flag alone is not evidence of availability.
3. `passed` returns `TEAM_AVAILABLE` and records the probe/version. `failed` or
   `indeterminate` returns `MODE_TEAM_UNAVAILABLE` with a blocker report; do
   not silently claim the retired fallback or silently downgrade to autopilot.

The former team catalog entry, removed files, and historical sentinel are
genealogy only. They must be labelled `historical`, while the resolver's
current probe owns all operational decisions. Native `moai cg` panes and
`moai cc -w --spawn` teammate windows are separate sanctioned runtimes and do
not prove Agent Teams availability.
