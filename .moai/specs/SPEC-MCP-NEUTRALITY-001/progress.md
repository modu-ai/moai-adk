# SPEC-MCP-NEUTRALITY-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-14
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L) + decision-index.md (decision_gate: on)
plan_audit_verdict: PASS-WITH-DEBT (r3, overall_score 0.91; trajectory 0.625 → 0.825 → 0.91)
plan_audit_sha: 9be172fce7
plan_artifact_hash: 61225de45bf1ed864b5af169632dca8db6499c1c4ba6216aaaee3937f599270c
plan_audit_receipt: rcpt-7323ca7bef2c885633529458 (codex required-gate pass, 0 findings)
plan_audit_verdict_file: .moai/reports/t1632/plan-audit-r3.md
debts: N3 (serial-slot instance red at M1 RED), N4 (AC-MN-023 reuse-refusal + restart-persistence depth, M6), N5 (AC-MN-012 "without stalling" quantification, M3)

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

Input parameters:
- tier: L (25 REQ / 25 AC, 6 milestones M1-M6)
- scope (file count): ~15-25 Go source + test files across internal/cli (mcp_*), internal/hook, internal/factorymsg, internal/auditreceipt, internal/codexwiring
- domain count: 3+ (MCP server layer, hook observer, messaging broker, audit receipts) — but coding-heavy
- file language mix: ~100% Go (+ markdown SPEC artifacts)
- concurrency benefit: LOW (coding-heavy — sequential milestone dependencies; M1 refusal contract is consumed by M2-M6)
- Agent Teams prereqs: not applicable (no operator request)

Mode evaluation:
- direct: not selected — semantic multi-file implementation, not trivial
- fanout: not selected — coding-heavy work violates Anthropic's coding-task parallelism caveat; milestones share the refusal contract and have inter-file dependencies
- sweep: not selected — semantic new-code work, not mechanical-uniform transformation
- serial: **selected** — one manager-develop spawn per milestone, sequential M1→M6

Decision: serial

Justification: the milestone chain is a layered contract build — M1's structured-refusal + actor-argument result shape is consumed by every later milestone (M2 project_root, M3 gate sessions, M6 audit integrity). Parallel spawns would each need the finished M1 contract and would contend for the same files (mcp_server.go dispatch tables). Anthropic's guidance: coding tasks involve fewer truly parallelizable tasks than research; serial is the safe default. Kickoff gate met in autonomous form (audit-cross: plan-audit-r3 PASS-WITH-DEBT 0.91 + codex pass rcpt-7323ca7bef2c885633529458; decision record .moai/reports/t1632/kickoff-decision.md; artifacts unchanged since verdict).
