# SPEC-TEMPLATE-RULES-CLEANUP-001 — progress.md

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-09T06:33:35Z
artifacts: [spec.md, plan.md, acceptance.md, research.md, design.md, progress.md]
tier: L
req_count: 28
ac_count: 33
```

## §E.2 Run-phase Evidence

### M1 — CI Guards (TDD RED baseline)

M1 completed: 2 new guard files authored, 4 guard classes active.

**Guard files:**
- `internal/template/rule_provenance_audit_test.go` (guards a+b+d)
- `internal/template/rule_date_provenance_audit_test.go` (guard c)

**RED evidence (pre-cleanup, captured M1):**

Guard (a) REQ/AC-token + (b) lessons/W# + (d) governance-token:
`go test ./internal/template/ -run TestRuleProvenanceAudit$ -count=1` → **FAIL** (145 occurrences)
- REQ/AC: AC-ADM-005 (askuser-protocol.md:177), REQ-MIG003-006/REQ-WF006-006/015/011 (settings-management.md), AC-LR-009 (sprint-round-naming.md:23,94)
- lessons/W#: W3 meta-analysis (agent-common-protocol.md:274), lessons #21 W0 fix / W1/W2 / W3 케이스 / W3에서 (manager-develop-prompt-template.md), lessons #13/#12/#14 (session-handoff.md, spec-workflow.md)
- governance: 127 CONST-V3R* (zone-registry.md 121+, manager-develop-prompt-template.md 4, worktree-integration.md 2) + MIG-003 (settings-management.md:174)

Guard (c) date-provenance:
`go test ./internal/template/ -run TestRuleDateProvenance$ -count=1` → **FAIL** (12 occurrences)
- zone-registry.md:629,630 (removed in M3), constitution.md:12-15 (HISTORY dates), session-handoff.md:330, spec-workflow.md:175,269,427, worktree-integration.md:44,46 (2026-05-17 policy)

**Sentinel output format:** `path | sentinel=SENTINEL | class=NAME | line=N | match=TOKEN`
- RULE_REQ_AC_TOKEN_LEAK, RULE_PROVENANCE_LEAK, RULE_GOVERNANCE_TOKEN_LEAK, RULE_DATE_PROVENANCE_LEAK

**Recurrence backstop self-tests:** all PASS (TestRuleProvenanceRecurrenceBackstop, TestRuleDateProvenanceRecurrenceBackstop)
**Tier-ownership contract:** TestRuleDateProvenanceNotInDefaultTier PASS (date NOT in default-tier leakClasses)
**3 contract tests:** TestRuleTemplateMirrorDrift, TestSanitizedPairParity, TestLeakClassNoDateShaInDefaultTier — all PASS
**Cross-platform build:** `go build ./...` exit=0, `GOOS=windows GOARCH=amd64 go build ./...` exit=0

RED evidence logs: `.moai/state/verify/cleanup-001/m1-red-provenance.log`, `m1-red-date.log`

### M2-M7 — Complete (GREEN)

All 4 guards PASS. Full test suite GREEN. make build exit=0. Cross-platform build exit=0.
GREEN evidence: `.moai/state/verify/cleanup-001/m7-green-guards.log`

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-09T08:42:00Z
run_commit_sha: pending-push
run_status: complete
ac_pass_count: 33
ac_fail_count: 0
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-09T09:15:00Z
sync_commit_sha: "0d90fa594"
sync_status: complete
frontmatter_status_transitions:
  spec_md: "in-progress → completed (single sync commit, 3-phase close)"
ac_pass_count: 33
ac_fail_count: 0
changelog_entry_position: "[Unreleased] > Changed"
canary_compliance_check:
  b12_self_test_a: "grep -c 'SPEC-TEMPLATE-RULES-CLEANUP-001' CHANGELOG.md == 1 (pre-emission 0, no duplicate)"
  b12_self_test_b: "AC count 33 / REQ count 28 verified against acceptance.md + spec.md SSOT"
  b12_self_test_c: "file paths in CHANGELOG verified via ls before commit"
```

## §F Phase 0.95 Mode Selection

Decision: sub-agent (Mode 5, sequential)

Tier L, coding-heavy Go guard authoring (M1) + sequential milestone dependencies (M1→M2-M5→M6→M7). Per Anthropic coding-task parallelism caveat, sequential sub-agent is the correct default. Agent Teams unavailable (team.enabled default false).
