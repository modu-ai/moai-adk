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

### M2 — leader discovery primitive with seams (REQ-002/003/005; AC-004/005/011/017 unit level)

- New package `internal/discovery` (6 files): `factory_discovery.go` (DiscoverLeader + seams + verifyCandidate 4-fact proof), `leader_readers_{darwin,linux,other}.go` (ps eww + lsof / /proc / decline-stubs), `factory_discovery_test.go` (11-test classifier matrix, pure fakes), `factory_discovery_live_test.go` (darwin real-process smoke via the helper-child pattern).
- RED (verifyCandidate stubbed to decline-all): `go test ./internal/discovery -count=1` → 4 `--- FAIL` verbatim (VerifiesLiveLeader `verified = 0, want 1`; ReturnsBothOnMultiLeaders; MembershipFallbackOnUnreadableCwd; DeduplicatesCandidates).
- GREEN (real implementation): `go test ./internal/discovery -count=1` → caught a real double-count defect (duplicate candidate pid verified twice); fixed by deduping inside DiscoverLeader's loop → `ok github.com/modu-ai/moai-adk/internal/discovery 1.864s`.
- Live-process smoke (REAL ps/lsof/probe against a real spawned helper): `--- PASS: TestDiscoverLeaderVerifiesLiveProcessWithPlatformReaders (0.20s)`.
- Live-population latency measurement (design.md §G, throwaway test, deleted after): 97 live candidates enumerated, `DiscoverLeader over live population: 0 verified, elapsed 468.096834ms` — 3% of the 15s deadline.
- Build: `go build ./...` + `GOOS=linux` + `GOOS=windows` on internal/discovery — all exit 0; `go vet ./internal/discovery` clean.
- Design note: REQ-002's fingerprint predicate in the discovery direction has no recorded side to mismatch (the measured fingerprint IS the resume stamp); the "together with" leg is pinned by TestDiscoverLeaderDeclinesIndeterminateIdentity (live pid + unreadable fingerprint declines).

### M3 — join-point integration + flag + env (REQ-001/004/005/006/008/009/010; AC-001..AC-004, AC-009..AC-013)

- RED (`enterFactoryLaneRun` stubbed to the plain gate, discovery absent): `go test ./internal/cli -run '...' -count=1` → 6 `--- FAIL` blocks verbatim (TestCCFactoryLaneJoinsDiscoveredLeader / TestGLMFactoryLaneJoinsDiscoveredLeader `NO_ACTIVE_FACTORY`; TestFactoryLaneJoinZeroLeadersRefuses `discovery asked 0 time(s)`; TestFactoryLaneJoinMultiLeaderFailsClosed; TestFactoryLaneJoinLeadTargeting ×2; TestFactoryLaneJoinMirrorParity `does not carry "discoverFactoryLeader("`).
- GREEN: `go test ./internal/cli -run 'TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestFactoryLaneJoin|TestCCFactoryLeadFlagLegacyRefused|TestFactoryLeadFlagSurfaceGates' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 12.302s`.
- Files: `internal/cli/factory.go` (enterFactoryLaneRun shared gate + discoverFactoryLeader seam + AMBIGUOUS_FACTORY_LEADER sentinel + resumeDiscoveredRun + `-l/--lead` parse with legacy refusal and surface gates + @MX:ANCHOR/@MX:NOTE), `internal/cli/kanban.go` (kanbanEntryParse.FactoryLead), `internal/cli/cc.go` + `internal/cli/glm.go` (lane branch → enterFactoryLaneRun; help text `-l, --lead` line + record-absence tolerance), `internal/cli/codex_factory.go` (parse --lead + lane branch → enterFactoryLaneRun + lead-name forwarding in codexFactoryEnv), `internal/discovery/factory_discovery.go` (DescribeVerifiedLeaders exported), `internal/cli/factory_test.go` (capture.leadName field — additive), `internal/cli/factory_join_discovery_test.go` (new, 10 tests), `internal/hook/factory_resumed_bind_test.go` (new — AC-012's hook half: the unchanged bind chain binds a lane peer generation-1 into a RESUMED run's broker).
- Build/vet: `go build ./...` + `go vet ./internal/cli ./internal/discovery` clean.

### M5 — docs and help (REQ-008/001; AC-016)

- Help text: `internal/cli/glm.go` + `internal/cli/cc.go` — `-f lane` block gains the record-absence tolerance line and the `-l, --lead <name>` row (twins in parity).
- docs-site 12 files: `content/{en,ja,zh,ko}/advanced/factory-mode.md` (new paragraph 「실행 기록이 없어도 리더가 살아 있으면 합류합니다」 and locale equivalents), `content/{en,ja,zh,ko}/advanced/kanban-mode.md` (factory entry bullet extended), `content/{en,ja,zh,ko}/cli-reference/launchers.md` (new `-l, --lead <name>` table row).
- README 4 occurrences: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` — factory paragraph gains the inverse-direction coverage sentence (verify + restore + fail-closed-on-many) with the `--lead` flag.
- Parity check: `grep -c -- '--lead'` → 1 per locale in each of factory-mode.md / kanban-mode.md / launchers.md / README (12+4 files all 1).

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase — owned by manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase — owned by manager-docs; sync_commit_sha: )_
