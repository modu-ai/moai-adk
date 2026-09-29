# progress.md — SPEC-FACTORY-LANE-JOIN-SOCKET-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: clarifications-resolved-awaiting-delta-reaudit
plan_complete_at: (pending delta re-audit verdict)
authored_at: 2026-09-29
authored_by: manager-spec (card t1330, Tier L, 5 plan-phase artifacts + progress)
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md
baseline: WT-join-lead-socket @ 68e37864a
plan_audit: CONDITIONAL PASS 0.88 iter-1 (.moai/reports/t1330/plan-audit.md); D1-D4 annotation fixes applied (spec.md v0.2.0 §G History); scoped delta re-audit owed per auditor recommendation 1
clarifications_open: 0 — all three RESOLVED 2026-09-29 (operator confirmed --lead default `leader`; plan.md §F carries the resolution records)
red_ledger: acceptance.md §C (tree pin 68e37864a, cells R-1/R-2/R-3; auditor re-measured all three at HEAD 67a0a67dc, match)

## §E.2 Run-phase Evidence

Baseline (pre-flight, this tree @ c961c4d4a): `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0; regression pair `go test ./internal/cli -run '^(TestFactoryRunSelectionAtomicSlotsAndArgv|TestGLM_FactoryLeadRunIsJoinableByLane)$' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 4.297s`; `golangci-lint run --timeout=2m ./internal/cli/... ./internal/homestate/...` → `0 issues`.

### M1 — homestate resume writer (REQ-004; AC-006/007/008)

- RED (stub `factory_run_resume.go` returning nil): `go test ./internal/homestate -run 'TestResumeRun' -count=1` → 5 `--- FAIL` blocks verbatim (TestResumeRunCreatesAbsentRowActiveWithSuppliedOwner `sql: no rows in result set`; TestResumeRunReactivatesRetiredRowWithLeaderOwnerStamp `status = "retired", want active` + owner `(111,...)`; TestResumeRunAppendsAuditableResumedEvent / ...RecordsAlreadyActiveOutcome `run.resumed events = 0, want 1`; TestResumeRunRequiresVerifiedIdentity 3 subtests `= <nil>, want an error`).
- GREEN: `go test ./internal/homestate -run 'TestResumeRun' -count=1` → `ok github.com/modu-ai/moai-adk/internal/homestate 1.529s`.
- Full package: `go test ./internal/homestate -count=1` → `ok ... 45.068s`; `go vet ./internal/homestate` clean.
- Files: `internal/homestate/factory_run_resume.go` (new — `(*FactoryDB).ResumeRun`, one transaction, `run.resumed` event with `{basis,outcome}` payload, already-active no-op via `WHERE runs.status!='active'`), `internal/homestate/factory_run_resume_test.go` (new — 6 tests / 3 subtests).
- Status transition draft → in-progress recorded on this commit (M1).

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase — owned by manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase — owned by manager-docs; sync_commit_sha: )_
