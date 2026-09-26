# Progress — SPEC-HOOK-MATCHER-POWERSHELL-001

## §E.1 Plan-phase Audit-Ready Signal

- Card t1224. Artifacts (Tier M): spec.md, plan.md, acceptance.md, progress.md. Status: draft. Version 0.2.0 (plan-audit iter-1 FAIL 0.77 resolved; report `.moai/reports/plan-audit/SPEC-HOOK-MATCHER-POWERSHELL-001-review-1.md`).
- Requirements: 16 GEARS (REQ-HMP-001..016; REQ-HMP-002 redefined to the pre-tool wrapper in 0.2.0). ACs: 13 matrix rows, 13 Given-When-Then scenarios.
- SPEC ID self-check: `SPEC-HOOK-MATCHER-POWERSHELL-001` → PASS; uniqueness: 0 existing directories at authoring.
- Base: `efc807b81`; local develop at `35ab8cff3` when 0.2.0 was written; M0 absorbs.
- Open decisions: D1 matcher form (operator), D2 unclassifiable-command policy with per-guard log destinations (operator), D3 PostToolUse dead branches (lead), D4 sibling cards (operator), D5 LIVE arm C with binary provenance (lead), D6 wrapper Risk-Amplifier on PowerShell (operator) — plan.md § Open Decisions.
- Unverified premise: PowerShell hook payload carries `tool_input.command` and a Bash-shaped `tool_response`. Transcript-level t1211 evidence supports it; hook-level capture runs in M0 (arms A/B) before M1.
- Collision state: t1152 (SPEC-HOOK-STDIN-FAILCLOSED-001) landed in develop, `status: completed`; REQ-HMP-009 defers to its fail-closed answer. t1099 (`WT-dual-harness-parity-rebuild` `e722a1493`) edits `internal/cli/hook.go`, `hook_test.go`, `internal/codexadapter/**`, `internal/template/obligations*`; it does not edit `internal/hook/*`, the settings template, or `.claude/hooks/*` (measured 2026-09-26).
- Awaiting plan-audit iter-2 and Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
