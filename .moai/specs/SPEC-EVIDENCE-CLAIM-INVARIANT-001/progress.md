# SPEC-EVIDENCE-CLAIM-INVARIANT-001 — Progress

> Tier S (minimal). 4-phase lifecycle: plan → run → sync → Mx.

## §F.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-06-15
plan_status: audit-ready
plan_artifacts:
  - spec.md          # 12-field frontmatter + §A intent/provenance + §3 inline AC + §X.1 h3 Out of Scope
  - plan.md          # M1-M3 milestones, run-phase constraints (template neutrality), PRESERVE notes
  - progress.md      # this file — §F.1 signal
tier: S
artifact_count: 2    # spec.md + plan.md (acceptance.md inline in spec.md §3 per Tier S)
ac_count: 7          # AC-ECI-001..007, all MUST, all grep/file-existence falsifiable
spec_id_self_check: "decomposition: SPEC | EVIDENCE | CLAIM | INVARIANT | 001 → PASS"
frontmatter_schema: "12-field canonical PASS (created/updated/tags canonical names; +optional tier:S)"
out_of_scope_h3: present   # §X.1 '### X.1 Out of Scope —' + list items → MissingExclusions PASS
provenance: "IMP-06 (fable-ish 13-agent roadmap final adopted item)"
predecessors:
  - SPEC-HOOK-DISCIPLINE-WIRING-001   # IMP-01 (completed)
  - SPEC-STOP-EVIDENCE-GATE-001       # IMP-02/03 (completed); deferred IMP-06 OUT OF SCOPE at spec.md §B.2 line 185
defect_class_motivation: "L_manager_docs_false_backfill_report (claiming an unobserved verification)"
run_phase_deliverable: "1 doctrine rule file + template mirror (.claude/rules/moai/core/verification-claim-integrity.md)"
biggest_run_risk: "template mirror neutrality (internal_content_leak_test.go CI guard)"
```

## §F.1.1 Plan Audit Gate (plan-auditor iter-1)

```yaml
verdict: PASS-WITH-DEBT
score: 0.84              # Tier S threshold 0.75 (+0.09)
dimensions:
  clarity: 0.92
  completeness: 0.95
  testability: 0.62      # the debt — falsifiability
  traceability: 0.96
must_pass: "MP-1 AC consistency PASS / MP-2 N/A(doctrine matrix) / MP-3 frontmatter 12+tier:S PASS / MP-4 neutrality N/A->auto"
cross_ref_integrity: "all 5 cited files + line numbers EXACT (Skeptical@113, Verify-Don't-Assume@262, SectionE@167, verification-batch-pattern present, Verification Matrix@368 / Completion Report@574)"
defects:
  D1: "SHOULD-FIX -> REMEDIATED (orchestrator-direct). AC-ECI-006 vacuous grep: ERE backslash-pipe = literal pipe -> always 0-match (would pass on a real leak). Fix: verification commands moved table-cell -> fenced code block (§3.2) with real | alternation + escaped dot. Empirically proven: old form 0-match on 'SPEC-EVIDENCE' leak; new form matches; neutral string 0-match."
  D2: "MINOR -> REMEDIATED. AC-ECI-003 '또는 동치' synonym hatch -> §3.2 C3 concrete tokens (output-styles/moai/moai.md + manager-develop §E / E1-E7)."
  D3: "MINOR -> REMEDIATED. AC-ECI-007 human-tally grep -c -> §3.2 C7 single 'grep -cE ... -ge 4'."
  D4: "NO-ACTION. StatusGitConsistency WARNING (status draft vs git-implied implemented) = expected Hybrid-Trunk plan-on-main heuristic over-inference, NOT a SPEC defect; frontmatter draft is correct. lint.skip PROHIBITED (would suppress a useful signal on other SPECs). Independently confirmed by orchestrator Trust-but-verify V4 AND plan-auditor re-run."
remediation: "orchestrator-direct §3.2 code-block fix, committed with this plan-audit remediation commit (chore scope, status stays draft)"
report: ".moai/reports/plan-audit/SPEC-EVIDENCE-CLAIM-INVARIANT-001-2026-06-15.md"
falsifiability_post_fix: "testability debt addressed — all 7 AC now executable verbatim (C1-C7 in §3.2); run-phase §E runs the corrected idiom (running the vacuous form + reporting false 0-match PASS would itself violate this SPEC's invariant)"
```

## §F.2 Run-phase Evidence

(run-phase에서 manager-develop가 채움)

## §F.3 Sync-phase Audit-Ready Signal

(sync-phase에서 manager-docs가 채움)

## §F.4 Mx-phase Audit-Ready Signal

(Mx-phase에서 채움 — mx_commit_sha 등)
