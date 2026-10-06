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

All evidence captured this run against this tree (branch `WT-ast-template-guard`).
Two measurement passes: the original matrix at M2 HEAD `b82f8429a` (m3-*.txt) and the
repair-pass matrix at post-repair state over base `27cf29e95` (rp-*.txt) — both
line-for-line reproducible at their own HEAD. Primary evidence files:
`.moai/reports/t1377/run/` (gitignored local-primary; hoist to the card evidence path is
the lane/leader concern). GUARD shorthand = the anchored alternation of
`TestWorkflowWorktreeKeyHonesty|TestWorkflowWorktreeKeyHonestyAliasFixture`.

```yaml
run_evidence:
  measured_at: 2026-10-03
  head: df828f92d (rp-*.txt are the binding set; base tree = post-sync close fa107a019)
  m1_scanner:
    file: internal/template/workflow_worktree_key_honesty_test.go
    first_run: "reader index ([./...]): 165 packages, 1423 files scanned, 0 type errors"
    repair_pass_run: "reader index ([./...]): 165 packages, 1423 files scanned, 0 load errors (Fset attribution keeps the map identical)"
    computed_map: matches audited expectation table exactly (no reconciliation needed)
    evidence: .moai/reports/t1377/run/m1-gate.txt, rp-clean-final.txt
  guard_runtime_c6: >
    production test 8.92s first run, 2.21s warm; full GUARD 10.813s — within C-6's
    ~24s guidance (scanner precedent measured 23.997s).
  ac_matrix:
    AC-001: PASS — both named tests execute with --- PASS (rp-clean-final.txt)
    AC-002: PASS — deciding line: "REQ-003 dropped reader: key \"auto_cleanup\" (field AutoCleanup) — expected reader internal/cli/session_worktree.go no longer reads the field" (rp-ac002.txt)
    AC-003: PASS — deciding line: "REQ-003 unnamed reader: key \"auto_merge\" (field AutoMerge) — internal/cli/doctor.go reads the field but the expectation table does not name it" (rp-ac003.txt)
    AC-004: PASS — three arms, each RED then clean (rp-ac004-arm1/2/3.txt); arm 3 deciding line: "REQ-004: auto_cleanup expectation must name mandatory reader internal/cli/session_worktree_prmerge.go (both auto-cleanup sites gate worktree removal)"
    AC-005: PASS — a/b/c/d fixture characterization (rp-clean-final.txt + m2-guard-both-tests.txt)
    AC-006: PASS — deciding line: "REQ-006: field TmuxPreferred (key \"tmux_preferred\") has no expectation-table entry — add one when the field lands" (rp-ac006.txt, re-captured at final HEAD)
    AC-007a: PASS — deciding line: "REQ-003 unnamed reader: key \"tmux_preferred\" (field TmuxPreferred) — internal/cli/doctor.go reads the field but the expectation table does not name it" (rp-ac007a.txt)
    AC-007b: PASS — deciding line: "REQ-005: reserved key \"tmux_preferred\" must have an empty expectation; the table names [internal/cli/worktree_advisory.go]" (rp-ac007b.txt)
    AC-008: PASS — deciding line: "REQ-007: load errors (type/parse/list) across scanned packages — … internal/cli/session_worktree.go:990:22: undefined: undefinedWorktreeProbeIdentifier" (rp-ac008.txt)
    AC-009: PASS — package gate ok (pre-repair 262.586s; post-repair 161.682s) (rp-package-gate.txt)
    AC-010: PASS — golangci-lint run ./internal/template/... → "0 issues." (ABSOLUTE; the pre-change baseline was also 0, so no NEW issues) (rp-lint.txt)
    AC-011: PASS — gofmt -l empty on all four new files (post-repair re-run)
    AC-012: PASS — negative control, captured PASS (exit 0, no finding) with a _test.go-only read present (rp-ac012.txt)
  builds:
    go_build: exit 0 (post-repair re-run)
    windows_cross_build: "GOOS=windows GOARCH=amd64 go build ./... → exit 0 (post-repair re-run)"
  coverage: "go test -cover ./internal/template/... → internal/template 84.4% (reported, not gated)"
  boundary_grep: "AskUserQuestion|mcp__askuser over the 4 new files → no matches"
  scope_proof: "git diff --name-only against card base → exactly the 4 new source paths + SPEC artifacts"
  red_record: >
    The deliverable IS a test (C-5): RED is the mutation matrix — 10 prior arms + 3 new
    arms, each applied → GUARD → verbatim capture → revert → clean re-observe. The
    repair-pass set (rp-*.txt) is the binding record; the original m3-*.txt set is kept
    for its own HEAD.
```

### §E.2.1 Post-sync audit repair record (pre-merge)

Three review surfaces converged on the same defects — **claude audit_multi leg,
codex audit_multi leg, codex card-review (scope=card)**; codex executed overlay
reproductions for the code repairs. All applied in one repair commit:
`df828f92d` (full: `df828f92d39b2e3659a0184f4a8915bc45eac5eb`).

| # | Repair | Surface |
|---|--------|---------|
| 1 | compoundassign.go: `+=` is now the file's ONLY tracked-field access (the trailing `return w.SessionNamePattern` plain read removed — the fixture satisfies its own acceptance §D.5 only-as precondition) | P2, 3 surfaces, codex-executed |
| 2 | REQ-007 path fails closed on pkg.Errors (parse/list) in addition to pkg.TypeErrors, deduped, package + cause named — production and fixture scans alike | P2, 3 surfaces, codex-executed |
| 3 | Orphan table keys (a key naming no live struct field) now t.Errorf — bidirectional table↔struct completeness; the silent `continue` removed | P2, 2 surfaces + 4dim Security judge |
| 4 | File attribution via pkg.Fset.File(file.Pos()).Name() instead of pkg.Syntax↔pkg.GoFiles index alignment (diverges under cgo); _test.go exclusion applied to the same name; reader map verified identical (165/1423) | P3 + Security judge |
| 5 | Load-mode split: production scan restored to the precedent's exact mode set (NeedName\|NeedTypes\|NeedSyntax\|NeedTypesInfo\|NeedFiles); NeedDeps\|NeedImports restricted to the fixture-mode loader (loadWorktreeScanWithDeps); @MX:NOTE corrected; guard runtime measured against C-6 (8.92s first run / 2.21s warm, guidance ~24s) | P2 claude |
| 6 | sortedStringKeys generic → slices.Sorted(maps.Keys(m)) — the sibling-guard idiom (internal/cli/huh_v1_guard_test.go:45) | Craft judge |
| 7 | One-line comment on the ast.Inspect parent-stack idiom (push-on-node / pop-on-trailing-nil) | Craft judge |
| 8 | AC-005d strengthened to a real comparison: the t682-style accessor text match is run over the fixture source and asserted to report NO reader (both legacy forms), while the AST scan is asserted to attribute the read | P3 claude |

New mutation arms (repair pass):

- **(a) AC-005c scanner mutation** — the read-exclusion mutated to accept ANY assignment
  token → RED: `AC-005c: compoundassign.go consumes SessionNamePattern via += and must be classified as a reader` (rp-arm-a.txt)
- **(b) parse-error mutation** — `@` appended to worktree_advisory.go → RED: `REQ-007: load errors (type/parse/list) across scanned packages — … worktree_advisory.go:69:1: expected declaration, found 'ILLEGAL'` (rp-arm-b.txt)
- **(c) orphan-key mutation** — bogus table entry → RED: `REQ-006 orphan key: "bogus_reserved_key" names no field of the live WorkflowWorktreeConfig struct — remove the entry or add the field it expected` (rp-arm-c.txt)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-10-03
run_commits:
  - 568d8907c M1 AST scanner + expectation table + production honesty test
  - b82f8429a M2 alias characterization fixture + fixture-mode test
  - 27cf29e95 M4 run-phase gates green + progress evidence close
  - df828f92d post-sync audit repair commit — repairs 1-8 + mutation arms a-c
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
sync_commit_sha: "a1692c380"
sync_changes:
  - CHANGELOG.md: exactly one [Unreleased]/### Added entry for the AST-based
    workflow.worktree.* key-honesty guard + testdata characterization fixtures
    (duplicate-guard grep -c "SPEC-TPL-AST-GUARD-001" CHANGELOG.md = 0 before append)
  - spec.md frontmatter: status in-progress → implemented → completed (single
    sync commit — the 3-phase close); updated already 2026-10-03,
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
