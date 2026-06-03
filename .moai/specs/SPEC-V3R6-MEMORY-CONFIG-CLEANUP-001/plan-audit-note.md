# plan-auditor note — SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001 (iteration 1)

Verdict: PASS-WITH-DEBT (overall 0.88, Tier S threshold 0.75 — skip-eligible NOT reached < 0.90)

Live-code verification (2026-06-03, HEAD on main):
- MemoryConfig non-test consumers: types.go:315 (field), types.go:365 (struct), defaults.go:365 (default block) — ZERO functional consumers confirmed. Premise REAL.
- BOTH workflow.yaml copies contain memory block: local L18-22, template L142-150. Confirmed.
- const block defaults.go:74-76 (comment L73) present. Confirmed — KEEP target valid.
- AuditIndex/AuditDuplicates: only definition sites in audit.go (no production caller). Gap B no-touch baseline confirmed.
- resolveMemoryDir (session_end.go:184) + projectSlug (session_end.go:206) + resolve_memory_dir_test.go present. Gap C baseline confirmed.

DEFECTS:
- D1 [SHOULD-FIX] types_test.go claim is FALSE. spec REQ-MCC-005, plan §B L18 ("types_test.go 4 reflection assertions on Memory.*"), plan M1 step 5 ("types_test.go (~329-332)"): live grep of internal/config/types_test.go = ZERO Memory.* hits. Only workflow_nested_test.go:89 + defaults_test.go:396/419-421 reference Memory. The 3-file claim is really a 2-file reality. Remediation: correct REQ-MCC-005 + plan §B + M1-step5 to enumerate the 2 actual files; if types_test.go is genuinely clean it must not be listed as a required edit.
- D2 [MINOR] struct line range off-by-one. spec §B.1 + plan M1 say MemoryConfig "364-370"; live is comment L364 + struct L365-370. Cosmetic; AC-MCC-001 grep gate is line-agnostic so non-blocking.
- D3 [MINOR] symmetry-guard assertion claim unverifiable. plan §B + AC-MCC-002/edge-case assert audit_struct_yaml_symmetry_test.go guards the memory key; live grep shows zero memory refs there (reflection/table-driven). Guarantee is softer than stated but AC-MCC-002 (go test green) still covers it. No action required beyond awareness.

MUST-PASS: MP-1 PASS, MP-2 PASS, MP-3 PASS, MP-4 N/A (single-language Go cleanup).

---

## Orchestrator addendum (2026-06-03) — D1 REJECTED as false positive

D1 is a plan-auditor grep error and is REJECTED. Ground-truth re-verification:
`grep -n 'Memory' internal/config/types_test.go` returns 4 hits at lines 329-332:

    329: {wfType, []string{"Memory", "AuditEnabled"}},
    330: {wfType, []string{"Memory", "IndexLineCap"}},
    331: {wfType, []string{"Memory", "StaleAggregateThreshold"}},
    332: {wfType, []string{"Memory", "StalenessThresholdHours"}},

The auditor grepped for `Memory.` (dot field access) and missed the reflection-table
`[]string{"Memory", ...}` quoted-string form. Therefore SPEC REQ-MCC-005's 3-file
inventory (workflow_nested_test.go + types_test.go + defaults_test.go) is CORRECT, and
`types_test.go:329-332` MUST be edited in run-phase (remove the 4 Memory rows) —
otherwise removing the `WorkflowConfig.Memory` field breaks `types_test.go` compilation.

Net: D1 docking on Traceability (0.80) is invalid → effective Traceability ~0.95,
effective overall ~0.92. D2 (struct line 364 comment + 365-370) and D3 (symmetry guard
generic) remain MINOR / non-blocking. No SPEC body change required; SPEC is run-ready.
