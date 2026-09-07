# SPEC-CODEX-TEST-GAPS-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-CODEX-TEST-GAPS-001
status: draft
tier: M
cycle_type_recommendation: ddd
harness: standard
authored: 2026-09-07
fix_rounds:
  - "v0.2.0 (2026-09-07): plan-audit iter-1 FAIL 0.875 -> D1 tier S->M, D2 non-vacuous test selectors, D3 base-pinned union diff gate, D4 pid un-skipped + REQ-CTG-012 + zero-0.0% retarget, D5 REQ-CTG-011 quality gates, D6 background-run recipe note, D7 census denominators"
worktree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t501
branch: WT-codex-uncovered
base: origin/develop @ ace1c5440
id_check: "PASS (regex ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$, verbatim Bash output)"
uniqueness_check: "PASS (no SPEC-CODEX-TEST-GAPS in catalog; 16 SPEC-CODEX-* siblings)"
premise_verification: "8/8 named functions located at cited files (grep -n, this run, this tree); pid 3-branch body re-read at mcp_codex.go:494-499 in fix round"
refuted_premise_recorded: true
evidence_paths:
  - .moai/reports/t501/namegrep-counts.txt
  - .moai/reports/t501/coverage-perfunc.txt
  - .moai/reports/t501/coverage-run.log
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

```yaml
phase: run
logged_at: 2026-09-07
logged_after: "Implementation Kickoff Approval obtained 2026-09-07 (operator, via lead-session AskUserQuestion; lead relayed with HEAD cd855f296 measured)"
input_parameters:
  tier: M
  scope_files: "<5 test files extended (codex_job_control_test.go, mcp_codex_test.go, codex_contract_test.go, codex_init_test.go)"
  domain_count: 1
  file_language_mix: "100% Go test code, single package internal/cli"
  concurrency_benefit: "LOW — coding-heavy (Anthropic coding-task parallelism caveat); milestones M1+M2 both edit-adjacent test files, fan-out would create write contention on shared test files"
  agent_teams_prereqs: "not requested (no operator --team)"
mode_evaluation:
  direct: "not selected — 8 test items across 4+ files exceeds a single-response edit"
  serial: "selected — one manager-develop carries M1-M8 in audited order with the mutant-RED discipline interleaved"
  fanout: "not selected — coding-heavy work, 1 domain, <10 files; shared test files would collide across concurrent writers"
  sweep: "not selected — semantic test authoring, not mechanical-uniform; <30 files"
  agent-team: "not selected — explicit-request-only; no request"
decision: serial
justification: >
  Tests-only work confined to one package with per-milestone edits landing in
  shared test files makes concurrency a write-conflict risk rather than a speed
  win. The audited milestone order (M1=U3 terminateCodexProcess, M2=U2
  codexIDMatches, M3=U2 ctx-cancel arm) already fronts the card's priority
  clusters, so a single sequential spawn preserves both the priority directive
  and the mutant-RED evidence chain (each test shown RED under a named mutant
  before GREEN on the pristine tree). Serial is the default fallback for
  coding-heavy work per the Anthropic parallelism caveat.
boundary_case: "none — no threshold-adjacent inputs"
sweep_confirmation: "N/A (sweep not selected)"
```

