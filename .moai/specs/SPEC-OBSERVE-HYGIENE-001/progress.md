# SPEC-OBSERVE-HYGIENE-001 — Progress

> Plan-phase artifact. Run-phase evidence (§E.2, §E.3) and sync-phase evidence (§E.4) are populated by manager-develop and manager-docs respectively. This file is parser-load-bearing: `internal/spec/era.go` `hasAnyProgressMarker` greps for the literal `§E.2` / `§E.3` / `§E.4` substrings — the headings below MUST be preserved verbatim.

---

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: S
epic: Workflow-Reflex (6 of 6)
artifacts: 4 (spec.md + plan.md + acceptance.md + progress.md)
gears_requirements: 6
acceptance_criteria: 7
out_of_scope_topics: 6
audit_findings_traced: H4, H5, H6, L7
related_specs: SPEC-HARNESS-RATCHET-REWIRE-001 (harness boundary), SPEC-TOKEN-BUDGET-STOP-001 (verify-dir boundary)
open_decision_points: D1 (audit-log consumer vs document — (a) recommended), D2 (retention threshold + task-metrics disposition), D3 (sync-gate blocking promotion — user confirmation required before M3)
brief_correction: model_upgrade_review HAS a consumer (emitModelUpgradeReminder, REQ-HRN-001-016); dormancy is at the never-set CLAUDE_MODEL_PREVIOUS trigger env — annotation wording adjusted accordingly
sibling_surface_coordination: loop.md (three-way with SPEC-LOOP-VERDICT-CONTRACT-001 + SPEC-ADVISOR-RUNG-001 — orchestrator sequences)
spec_id_self_check: PASS (SPEC-OBSERVE-HYGIENE-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

**Plan-phase self-verification**:
- [x] All GEARS requirements use valid patterns (Ubiquitous / When / Where; no IF/THEN).
- [x] All gap claims cite measured source + observed pattern (vci §2; measured 2026-07-09; log sizes, zero-consumer greps, dormancy layers, and the model_upgrade_review consumer re-verified — brief claim corrected rather than propagated per vci §1.1 surface 3).
- [x] §Out of Scope has 6 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.
- [x] 12 canonical frontmatter fields + era + tier + related_specs; no snake_case aliases.
- [x] Per-sink disposition rule (exactly one of consumer/document/prune) encoded in the requirements header + DoD #3.
- [x] Hook signaling contract (exit-0 stdout block channel) + runtime-recovery §4 carve-out stated as preserved invariants.
- [x] Template mirrors verified present for all mirrored edit surfaces (sunset.yaml, harness.yaml, both hook scripts, loop.md).

---

## §E.2 Run-phase Evidence

```
spec_id: SPEC-OBSERVE-HYGIENE-001
tier: S
cycle_type: tdd
mode: sub-agent (Mode 5, sequential)
base: 280c9dd71 (race-absorbed — 2 parallel sessions landed scope-overlap-zero)
commits:
  M1: bcde87d24 feat(SPEC-OBSERVE-HYGIENE-001): M1 spec-audit log consumer (REQ-OBH-001)
  M2: 5c11080bb feat(SPEC-OBSERVE-HYGIENE-001): M2 telemetry pruning (REQ-OBH-002)
  M3: <this commit> feat(SPEC-OBSERVE-HYGIENE-001): M3 decisions + annotations + template sync (REQ-OBH-003..006)
d3_decision: Promote (sync-gate go vet/go build → default-blocking; MOAI_SYNC_GATE_BLOCKING → opt-out)
manager-develop: aborted mid-M3 (API 429 usage limit); orchestrator-direct M3 completion (user "이어서 계속 진행")
```

D1/D2/D3 decisions recorded:
- D1: (a) spec-audit consumer (recommended) — M1 implemented (moai spec audit parses status-transition-audit.log, correlates against Ownership Matrix, INFO findings, graceful no-op on absent/corrupt)
- D2: 30-day retention (DefaultTraceRetentionDays in internal/config/defaults.go); task-metrics.jsonl disposition = document write-only + age-out (writer path post_tool.go:~166; root-cause suspected upstream Agent/Task event shape change 2026-05-30; inconclusive → default document+age-out)
- D3: Promote — M3 implemented (vet/build default-blocking; env var opt-out; tests/coverage advisory; stdout-JSON exit-0 + --skip-hook + §4 carve-out preserved)

Per-sink disposition (6 groups, exactly one each):
1. status-transition-audit.log → (a) consumer [M1]
2. trace-*.jsonl zero-byte + age → (c) prune [M2, SessionEnd path, 30-day]
3. task-metrics.jsonl → (b) documented write-only + age-out [M2, root-cause note]
4. fact-force-skip.log → (b) documented write-only [M3, header line]
5. sync-gate vet/build → D3 Promote [M3, default-blocking]
6. loop-snapshots + sunset.yaml + harness.yaml model_upgrade_review → (b) dormancy annotations [M3]

## §E.3 Run-phase Audit-Ready Signal

```
run_status: audit-ready (PASS-WITH-DEBT)
run_complete_at: 2026-07-09
ac_matrix:
  AC-OBH-001: PASS — moai spec audit consumes status-transition-audit.log (fixture tests green; internal/spec coverage 88.3%)
  AC-OBH-002: PASS — SessionEnd prunes zero-byte trace + 30-day age retention; task-metrics disposition recorded (internal/hook 83.5%)
  AC-OBH-003: PASS — fact-force-skip.log write-only line in gateguard-fact-force.sh header (live + template mirror)
  AC-OBH-004: PASS — D3=Promote; go vet/go build default-blocking; env var opt-out; stdout-JSON exit-0 + --skip-hook + §4 carve-out preserved (live + template mirror)
  AC-OBH-005: PASS — loop.md best-effort marker on snapshot/resume sections (live + template mirror)
  AC-OBH-006: PASS — sunset.yaml + harness.yaml model_upgrade_review dormancy comments (live + template mirror); config tests bit-identical green
  AC-OBH-007: PASS-WITH-DEBT — cross-platform build green (linux+windows exit 0); make build green; new paths ≥85% (spec 88.3%); lint 0 issues; subagent-boundary grep 0; template-first mirrors synced; no .moai/logs/ committed.
     DEBT (2 pre-existing FAIL, OBSERVE 유발 0건):
       (1) internal/cli TestDoctor/Status 6 golden mismatch — TEST-003 AC-006 external debt (stale golden, ARCH-001 M0 소관)
       (2) internal/template TestOutputStylesTemplateLiveParity moai-easy.md — 별도 chore (RATCHET-REWIRE memory; moai-easy.md template-only, live mirror + Doctor/Status golden update 미완료)
evidence_dir: /tmp/moai-verify-obh/ (logs 1-9; persist to .moai/state/verify/ pre-sync)
```

## §E.4 Sync-phase Audit-Ready Signal

```
sync_status: complete
sync_complete_at: 2026-07-09
sync_commit_sha: 343b32287
changelog_entry_position: CHANGELOG.md [Unreleased] > ### Fixed (top entry)
frontmatter_status_transitions:
  spec.md: in-progress → completed (single sync commit, merged 3-phase close)
  plan.md: in-progress → completed (single sync commit)
  acceptance.md: in-progress → completed (single sync commit)
  progress.md: §E.4 populated (this block); §E.1-§E.3 preserved verbatim (era.go parse-load-bearing)
b12_self_test_a: PASS — grep -c 'OBSERVE-HYGIENE-001' CHANGELOG.md = 0 pre-emission (no duplicate from parallel BATCH-SYNC)
b12_self_test_b: PASS — AC count 7 matches acceptance.md §D SSOT (grep -cE '^\| AC-OBH-[0-9]+' acceptance.md = 7; NOT progress.md)
b12_self_test_c: PASS — all cited file paths verified via ls: internal/spec/audit.go, internal/cli/spec_audit.go, internal/hook/post_tool.go, internal/config/defaults.go, .claude/hooks/moai/gateguard-fact-force.sh, .claude/hooks/moai/sync-phase-quality-gate.sh, .moai/config/sections/sunset.yaml, .moai/config/sections/harness.yaml, .claude/skills/moai/workflows/loop.md
canary_compliance_check: N/A (SPEC does not define a forward-looking policy that its own sync tests)
route: A (Hybrid Trunk main-direct, Tier S — no PR, no worktree)
ac_matrix_summary: AC-OBH-001..006 PASS, AC-OBH-007 PASS-WITH-DEBT (2 pre-existing FAIL unrelated to OBSERVE — cli Doctor/Status golden TEST-003 AC-006 external debt + template moai-easy.md parity chore)
d3_decision: Promote (sync-gate go vet/go build → default-blocking; MOAI_SYNC_GATE_BLOCKING → opt-out)
push_state: origin/main (sync commit + follow-up backfill chore commit)
note: sync_commit_sha backfill is a separate amend-free chore commit per SPEC-V3R6-LIFECYCLE-REDESIGN-001 merged 3-phase close convention (close-subject full-ID mandate: SPEC-OBSERVE-HYGIENE-001)
```

---

## §F Phase 0.95 Mode Selection

```
tier: S
scope_files: ~7 logical surfaces (Go source internal/spec + internal/hook; 2 hook scripts; 2 config blocks; loop.md; template mirrors; _test.go)
domain_count: 2 (Go internal/{spec,hook} code + hook/config/doc markdown surfaces)
file_language_mix: Go + bash + YAML + markdown
concurrency_benefit: LOW (coding-heavy; milestone dependency — M3 gates depend on D3 decision)
agent_teams_prereqs: NOT MET (workflow.team.enabled default false since Sonnet 5 re-design)
```

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | 6 REQ / 7 AC / 3 milestones — non-trivial implementation |
| 2 background | no | write-heavy (Go code + commits + template mirrors); not read-only |
| 3 agent-team | no | team.enabled default false; all three prereqs unmet |
| 4 parallel | no | coding-heavy — Anthropic coding-task parallelism caveat; M1→M2→M3 sequential dependency |
| 5 sub-agent | YES | Tier S coding-heavy work; sequential per-milestone delegation; M3 depends on D3 decision (Promote) |
| 6 workflow | no | not mechanical-uniform transform; new code + semantic edits; <30 files |

Decision: sub-agent

Justification: Tier S coding-heavy SPEC (Go TDD across internal/spec audit consumer + internal/hook SessionEnd pruning, ~200-280 LOC) with sequential milestone dependency (M3's gate promotion blocks on the D3 user decision = Promote, already captured at Implementation Kickoff Approval). Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research") routes coding-heavy work to the sequential sub-agent path (Mode 5), not parallel fan-out (Mode 4) nor workflow (Mode 6 — admits only genuinely-parallel high-volume mechanical transforms). A single manager-develop delegation (cycle_type=tdd) covers M1→M2→M3 in dependency order. Implementation Kickoff Approval: PASSED (user confirmed run-phase entry + D3=Promote). All preferences collected. plan-audit iter-1 PASS 0.96 skip-eligible.
