# Progress — SPEC-HOOK-MATCHER-POWERSHELL-001

## §E.1 Plan-phase Audit-Ready Signal

- Card t1224. Artifacts (Tier M): spec.md, plan.md, acceptance.md, progress.md. Status: draft. Version 0.1.0.
- Requirements: 16 GEARS (REQ-HMP-001..016). ACs: 12 matrix rows, 9 Given-When-Then scenarios.
- SPEC ID self-check: `SPEC-HOOK-MATCHER-POWERSHELL-001` → PASS; uniqueness: 0 existing directories.
- Base: `efc807b81` (local develop had advanced to `044fb91c7` at authoring; M0 absorbs).
- Open decisions: D1 matcher form (operator), D2 unclassifiable-command policy (operator), D3 PostToolUse dead branches (lead), D4 sibling cards (operator), D5 LIVE arm C (lead) — plan.md § Open Decisions.
- Unverified premise: PowerShell hook payload carries `tool_input.command` and a Bash-shaped `tool_response` — captured by M5 arm B before live claims.
- Collision watch: t1099 (`WT-dual-harness-parity-rebuild`) and t1152 (`WT-hook-stdin-failclosed`) edit `internal/cli/hook.go` and `internal/codexadapter/**`; neither edits the `internal/hook` files or the template this SPEC touches (measured 2026-09-26).
- Awaiting plan-audit and Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
