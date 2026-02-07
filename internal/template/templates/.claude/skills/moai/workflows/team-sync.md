# Workflow: Team Sync - Documentation Phase

Purpose: Explain why sync phase always uses sub-agent mode and document the rationale.

Flow: manager-docs subagent (always single agent, no team mode)

## Rationale

The sync phase always uses sub-agent mode (manager-docs) because:

1. Documentation generation is sequential by nature (README depends on code analysis, CHANGELOG depends on git history, PR depends on both)
2. File outputs are few (3-5 files) with heavy interdependency
3. Token budget is small (40K) making team overhead wasteful
4. Single coherent voice produces better documentation quality

## When Team Mode Might Help

Future consideration for team mode if:
- Documentation spans multiple languages (i18n)
- API docs and user docs need simultaneous generation
- Documentation site has 10+ pages requiring parallel rendering

## Current Behavior

When /moai sync is invoked with --team flag:
- Log informational message: "Sync phase uses sub-agent mode for optimal coherence"
- Execute standard sync workflow via manager-docs subagent
- --team flag is acknowledged but not applied for this phase

For standard sync workflow details: See workflows/sync.md

---

Version: 1.0.0
