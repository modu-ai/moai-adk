# SPEC-TPL-AST-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-03
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md, research.md]
tier_artifact_set: [spec.md, plan.md, acceptance.md]
operator_added: [research.md]
code_baseline: 7c7c84b5c
branch: WT-ast-template-guard
related_specs: [SPEC-CONFIG-KEY-HONESTY-001, SPEC-WORKTREE-KEY-WIRING-001]
decisions:
  D1_contract: A-template-as-is (recommendation; leader may override at plan review)
  D2_scope: a-minimal-single-guard (recommendation)
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.87
    threshold: 0.80
    blocking: [D1-vacuous-selector, D2-PR1707-narrative, D3-REQ004-enforcement-hole, D4-reader-definition]
    folded_optional: [D5, D6, D7, D8, D9]
    verdict_file: .moai/reports/t1377/plan-audit.md
    corrections_applied: 2026-10-03
  iteration_2:
    verdict: FAIL
    score: 0.85
    blocking: [N1-reader-definition-incomplete, N2-file-count-contradiction, N3-PR-narrative-residual, N4-AC005-invocation-vacuity]
    folded_optional: [N5, N6, N7, N8, N9, N10]
    verdict_file: .moai/reports/t1377/plan-audit-iter2.md
    corrections_applied: 2026-10-03 (iteration 3 = confirming pass)
  iteration_3:
    verdict: PASS-WITH-DEBT
    score: 0.94
    verdict_file: .moai/reports/t1377/plan-audit-iter3.md
kickoff:
  decision: APPROVED (PASS-WITH-DEBT 0.94 accepted as entry verdict — leader gate, 2026-10-03)
  decision_record: .moai/reports/t1377/kickoff-decision.md
  plan_audit_verdict: "plan-audit 판정 = .moai/reports/t1377/plan-audit-iter3.md + 델타 확인 파일, 영수증 rcpt-4b191496ee0b52affcd21a00"
  run_phase_1: BYPASSED per leader decision (t1344 known cache-miss mismatch)
gate_debt:
  F3: M1-gate ordering notation — notation-only, carried per leader decision 2026-10-03 (no text fix)
  F4: REQ-005 reserved-key source notation — notation-only, carried per leader decision 2026-10-03 (no text fix)
```

## §E.2 Run-phase Evidence

All evidence captured this run against this tree (branch `WT-ast-template-guard`),
HEAD M2 commit `b82f8429a` (mutations observed at M2 state; final clean re-observation
identical). Primary evidence files: `.moai/reports/t1377/run/` (gitignored local-primary;
hoist to the card evidence path is the lane/leader concern). GUARD shorthand = the
anchored alternation of `TestWorkflowWorktreeKeyHonesty|TestWorkflowWorktreeKeyHonestyAliasFixture`.

```yaml
run_evidence:
  measured_at: 2026-10-03
  head: b82f8429a
  m1_scanner:
    file: internal/template/workflow_worktree_key_honesty_test.go
    first_run: "reader index ([./...]): 165 packages, 1423 files scanned, 0 type errors"
    computed_map: matches audited expectation table exactly (no reconciliation needed)
    evidence: .moai/reports/t1377/run/m1-gate.txt
  ac_matrix:
    AC-001: PASS — both named tests execute with --- PASS (m2-guard-both-tests.txt)
    AC-002: PASS — dropped-reader RED then clean (m3-ac002.txt)
    AC-003: PASS — unnamed-reader RED then clean (m3-ac003.txt)
    AC-004: PASS — three arms, each RED then clean (m3-ac004-arm1/2/3.txt; arm 3 fires REQ-004)
    AC-005: PASS — a/b/c/d fixture characterization (m2-guard-both-tests.txt + m3-clean-final.txt)
    AC-006: PASS — table-entry deletion RED (m3-ac006.txt)
    AC-007a: PASS — reserved-key reader RED (m3-ac007a.txt)
    AC-007b: PASS — table-invalid under reserved key RED (m3-ac007b.txt)
    AC-008: PASS — type-error RED with named cause (m3-ac008.txt)
    AC-009: PASS — package gate ok 262.586s (m4-ac009-package-gate.txt)
    AC-010: PASS — golangci-lint 0 issues, no NEW vs baseline (m4-lint-after.txt)
    AC-011: PASS — gofmt -l empty on all four new files
    AC-012: PASS — negative control, captured PASS with a _test.go-only read (m3-ac012.txt)
  builds:
    go_build: exit 0
    windows_cross_build: "GOOS=windows GOARCH=amd64 go build ./... → exit 0"
  coverage: "go test -cover ./internal/template/... → internal/template 84.4% (reported, not gated)"
  boundary_grep: "AskUserQuestion|mcp__askuser over the 4 new files → no matches"
  scope_proof: "git diff --name-only against card base → exactly the 4 new source paths + SPEC artifacts"
  red_record: >
    The deliverable IS a test (C-5): RED is the mutation matrix above — 9 mutation
    observations each carrying verbatim failing output, every mutation reverted and the
    clean pass re-observed (m3-clean-after-arm3.txt, m3-clean-final.txt).
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-10-03
run_commits:
  - 568d8907c M1 AST scanner + expectation table + production honesty test
  - b82f8429a M2 alias characterization fixture + fixture-mode test
  - (M4 progress close commit — this commit)
m3_note: M3 is evidence-only (mutation matrix, no tracked-file change) — no commit.
push_state: DEFERRED (factory leader batch-pushes local develop after integration)
slot: go-test-internal-template released after the heavy-run batch
gate_debt_carry: F3/F4 notation-only items unchanged (leader decision 2026-10-03)
residuals:
  - M1 commit (568d8907c) trailer reads "MoAI" — the 🗿 glyph was dropped by the heredoc;
    --amend is prohibited (B9), so the defect is recorded here. M2+ commits carry the
    correct trailer.
  - template-claim ↔ expectation-table divergence remains unguarded under contract A
    (spec §A.3 known residual; follow-up candidate).
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-10-03
sync_commit_sha: "pending-backfill-sync"
sync_changes:
  - CHANGELOG.md: exactly one [Unreleased]/### Added entry for the AST-based
    workflow.worktree.* key-honesty guard + testdata characterization fixtures
    (duplicate-guard grep -c "SPEC-TPL-AST-GUARD-001" CHANGELOG.md = 0 before append)
  - spec.md frontmatter: status in-progress → completed (both transitions ride
    this single sync commit — the 3-phase close); updated already 2026-10-03,
    left unchanged; NO body edits
  - progress.md: this §E.4 fill; the sync_commit_sha placeholder is backfilled
    with the real SHA in the immediately following commit (D3 backfill exemption)
skip_records:
  mx_scan: >
    Test-only change. Mechanical declaration scan of the four new files: the guard
    test file declares only unexported identifiers plus the two Test* entry points
    (test functions are not exported API); the three fixture files declare
    AliasRead / PlainWriteOnly / CompoundAppend under testdata/, which the go tool
    structurally excludes from builds and the import graph — no symbol reaches any
    consumer, so no MX obligations fire. The @MX:NOTE authored by the run phase in
    workflow_worktree_key_honesty_test.go (shared-helper extraction deferral, plan
    D2) is retained unchanged.
  readme_docs_site: skipped — internal test-only change, no user-facing surface
```

## §F Phase 4 Mode Selection

```yaml
logged_by: lane-14 orchestrator (card t1377)
logged_at: 2026-10-03
inputs:
  tier: M
  scope_files: 4 (1 guard test + 3 testdata fixtures)
  domain_count: 1 (internal/template Go test)
  language_mix: "100% Go"
  concurrency_benefit: LOW (coding-heavy, inter-file dependency scanner->fixtures->mutations)
  agent_teams_prereqs: not requested
mode_evaluation:
  direct: not selected (non-trivial multi-file deliverable)
  serial: SELECTED (coding-heavy per Anthropic coding-task parallelism caveat; single manager-develop over milestones M1-M4)
  fanout: not selected (not multi-domain research)
  sweep: not selected (not >=30-file mechanical transform)
decision: serial
justification: >
  The deliverable is one interdependent Go test file plus fixtures with a strict
  mutation-evidence sequence; parallel spawns add reconciliation cost with no
  research fan-out benefit. Serial single-agent delegation per Milestone is the
  default fallback and matches Anthropic's coding-task parallelism caveat.
kickoff_gate: MET (leader decision .moai/reports/t1377/kickoff-decision.md — PASS-WITH-DEBT 0.94 accepted; delta-PASS re-pin 72e196d7 @ 73524bb35; Phase 1 BYPASSED per leader item 4)
```
