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
- GREEN: `go test ./internal/cli -run 'TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestFactoryLaneJoin|TestFactoryLeadFlagLegacyRefused|TestFactoryLeadFlagSurfaceGates' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 12.302s`.
- Files: `internal/cli/factory.go` (enterFactoryLaneRun shared gate + discoverFactoryLeader seam + AMBIGUOUS_FACTORY_LEADER sentinel + resumeDiscoveredRun + `-l/--lead` parse with legacy refusal and surface gates + @MX:ANCHOR/@MX:NOTE), `internal/cli/kanban.go` (kanbanEntryParse.FactoryLead), `internal/cli/cc.go` + `internal/cli/glm.go` (lane branch → enterFactoryLaneRun; help text `-l, --lead` line + record-absence tolerance), `internal/cli/codex_factory.go` (parse --lead + lane branch → enterFactoryLaneRun + lead-name forwarding in codexFactoryEnv), `internal/discovery/factory_discovery.go` (DescribeVerifiedLeaders exported), `internal/cli/factory_test.go` (capture.leadName field — additive), `internal/cli/factory_join_discovery_test.go` (new, 10 tests), `internal/hook/factory_resumed_bind_test.go` (new — AC-012's hook half: the unchanged bind chain binds a lane peer generation-1 into a RESUMED run's broker).
- Build/vet: `go build ./...` + `go vet ./internal/cli ./internal/discovery` clean.

### M5 — docs and help (REQ-008/001; AC-016)

- Help text: `internal/cli/glm.go` + `internal/cli/cc.go` — `-f lane` block gains the record-absence tolerance line and the `-l, --lead <name>` row (twins in parity).
- docs-site 12 files: `content/{en,ja,zh,ko}/advanced/factory-mode.md` (new paragraph 「실행 기록이 없어도 리더가 살아 있으면 합류합니다」 and locale equivalents), `content/{en,ja,zh,ko}/advanced/kanban-mode.md` (factory entry bullet extended), `content/{en,ja,zh,ko}/cli-reference/launchers.md` (new `-l, --lead <name>` table row).
- README 4 occurrences: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` — factory paragraph gains the inverse-direction coverage sentence (verify + restore + fail-closed-on-many) with the `--lead` flag.
- Parity check: `grep -c -- '--lead'` → 1 per locale in each of factory-mode.md / kanban-mode.md / launchers.md / README (12+4 files all 1).

### M4 — regression, parity matrix, mutant verification

- Full-suite classification: 3 local `go test ./internal/cli` attempts (600s default timeout / 1401s / 1800s -timeout 30m) all died to the global time ceiling with ZERO test-level failures — `grep -cE "^--- FAIL" /tmp/t1330-cli-full-suite.log` → `0` over 5250 lines; `panic: test timed out after 30m0s` fired while `TestUpdateForce_InitOriginNamesSurvive` (UPDATE subsystem, untouched by this card) was 1m41s in. Classification: load/time event on a machine running parallel factory lanes; the package-wide verdict is CI's (lane-local discipline, CLAUDE.local §4/§6/§8).
- Targeted evidence (this run, this tree @ HEAD below): regression pair + new join/parse/codex-twin cluster `go test ./internal/cli -run '^(TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestCodexFactoryLaneJoinsDiscoveredLeader|TestCodexFactoryLeadFlagSurface|TestFactoryLaneJoin.*|TestFactoryLeadFlagLegacyRefused|TestFactoryLeadFlagSurfaceGates|TestFactoryRunSelectionAtomicSlotsAndArgv|TestGLM_FactoryLeadRunIsJoinableByLane)$' -count=1 -timeout 10m` → `ok github.com/modu-ai/moai-adk/internal/cli 16.198s`; `go test ./internal/discovery -count=1` → `ok ... 2.188s`; `go test ./internal/homestate -count=1` → `ok ... 47.736s` (85.0% coverage).
- Mutant probes (each: patch → single test → verbatim RED → revert → package re-green):
  - AC-005 (liveness check deleted): `--- FAIL: TestDiscoverLeaderDeclinesDeadPidDespiteRecordPresence ... dead-pid candidate verified = [{RunID:runlead1 ... ProcessStart:mutant-stamp ...}], want none`.
  - AC-007 (resume stamps the caller — recordFactoryRunStart semantics): `--- FAIL: TestCCFactoryLaneJoinsDiscoveredLeader ... resumed row owner = (24446,"1790701220.017243"), want the verdict's verified identity (424242,1700000000.004242)` — the lane's own pid, exactly the hazard REQ-004 forbids.
  - AC-017 leg a (membership check deleted): `--- FAIL: TestDiscoverLeaderDeclinesCrossProjectCandidate ... cross-project candidate verified = [...], want none`.
  - AC-017 leg b (readable-identity requirement deleted; run id minted): `--- FAIL: TestDiscoverLeaderDeclinesUnreadableEnv ... verified = [{RunID:mutant00 ...}], want none`.
- AC-013 mirror parity: TestFactoryLaneJoinMirrorParity (source-level: cc.go/glm.go/codex_factory.go all route through enterFactoryLaneRun; no launcher carries DiscoverLeader/ResumeRun/discoverFactoryLeader) — green in the targeted cluster run; AC-001/AC-002 shape verified through cc AND glm AND the codex twin (TestCodexFactoryLaneJoinsDiscoveredLeader).
- AC-012 evidence note: the AC's "generation 1" parenthetical reflects `want.Generation:1` at factory_messages.go:96; the MEASURED ordinary chain binds the peer at launch-pending+1 (staged pending generation 1 → bound generation 2). The test pins the substance — "the bind an ordinary join produces" — asserted as pending+1 against the resumed run. acceptance.md body untouched (not run-phase scope); flagged for manager-spec if the auditor wants the parenthetical updated.
- Coverage (targeted packages, this run): `internal/discovery` 87.3% (after adding DescribeVerifiedLeaders + broker-peer enumeration tests); `internal/homestate` 85.0% (full package); `internal/cli` 7.1% and `internal/hook` 4.3% under SELECTIVE selectors only (the packages are far larger than the selector sweeps; package-wide figures are CI's to report — recorded as a Gap, not a pass).
- Lint: `golangci-lint run --timeout=2m` over internal/{cli,homestate,discovery,hook,kanban} → `0 issues.` (baseline was 0; no NEW findings).
- gofmt: `gofmt -l` over the four packages → clean after formatting factory_discovery_test.go.
- Preserved-invariant closure (AC-014/015): `go test ./internal/kanban -run 'TestNextFactoryLaneNumber|TestClaimFactoryLane|TestPruneFactoryDeadClaims|TestFactoryLaneLabel|TestSplitFactoryLegacyLabel' -count=1` → `ok github.com/modu-ai/moai-adk/internal/kanban 4.087s`; the ENTIRE cli factory test family `go test ./internal/cli -run 'TestFactory' -count=1 -timeout 20m` → `ok github.com/modu-ai/moai-adk/internal/cli 55.587s` (entry truth tables, legacy refusals, mixed-slot selection, and the new join tests together).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-30
run_commit_sha: f1e5c11bf
run_status: complete (M1-M5 landed; no blockers)
ac_pass_count: 17
ac_fail_count: 0
preserve_list_post_run_count: 7 (factorymsg/store.go resolver+refusals, homestate/factory_run_retire.go, kanban/bootstrap.go, kanban/factory_slots.go, cli/factory.go parse truth table, hook/factory_messages.go, existing regression assertions — untouched; ResolveActiveRun/ReconcileActiveRuns/RetireRunIfDead read-only from discovery)
l44_pre_commit_fetch: (lane worktree — no fetch performed; integration is the lead's window)
l44_post_push_fetch: (lane never pushes — lead batch-pushes develop)
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues over internal/{cli,homestate,discovery,hook,kanban}; gofmt clean)
cross_platform_build.darwin: pass (go build ./... exit 0)
cross_platform_build.windows: pass (GOOS=windows GOARCH=amd64 go build ./... exit 0, pre-flight and post-M3)
cross_platform_build.linux: pass (GOOS=linux go build ./internal/discovery/ exit 0)
total_run_phase_files: 29 (Go 12 + tests 4 + SPEC artifacts 2 + docs-site 12 + README 4, minus overlaps — see commits)
m1_to_m5_commit_strategy: one commit per milestone (M1 757255ff4, M2 c8a0d9137, M3 f892907c1, M5 c8c671a68; M4 was verification — its test additions ride the final commit)


## §E.4 Sync-phase Audit-Ready Signal

- sync_complete_at: 2026-09-30
- sync_commit_sha: 8883ed975
- sync_status: complete
- sync_phase_scope: artifact-only close (spec/plan/acceptance frontmatter, §E.3 run_commit_sha backfill per the D3 placeholder exemption, this §E.4, CHANGELOG entry) — no code, no template source, no docs edits; the documentation surface was already cleared in-run by M5 (c8c671a68) and the sync-phase spot-check found no drift
- b12_self_test_a: pass — `grep -c 'SPEC-FACTORY-LANE-JOIN-SOCKET-001' CHANGELOG.md` = 0 pre-emission
- b12_self_test_b: pass — 17 distinct live AC identifiers in acceptance.md (AC-001..AC-017, zero [RETIRED]/[REF] markers); the CHANGELOG entry references the same 17
- b12_self_test_c: pass — every file path cited in the CHANGELOG entry verified with `ls` (internal/homestate/factory_run_resume.go, internal/discovery/factory_discovery.go, internal/cli/factory.go, internal/cli/factory_join_discovery_test.go, internal/hook/factory_resumed_bind_test.go) and the load-bearing claims verified by grep (AMBIGUOUS_FACTORY_LEADER at factory.go:360, DiscoverLeader at factory_discovery.go:113, ResumeRun at factory_run_resume.go:39, run.resumed {basis,outcome} payload at factory_run_resume.go:70-71)
- changelog_entry_position: [Unreleased] → `### Added`, first entry
- frontmatter_status_transitions.spec_md: in-progress → implemented → completed (merged into the single sync commit; `updated:` refreshed 2026-09-30)
- frontmatter_status_transitions.plan_acceptance: no `status:` field authored (status-axis statelessness per spec-frontmatter-schema.md § Artifact Statelessness); `updated:` refreshed to 2026-09-30
- canary_compliance_check.mx_tags: no MX tag surface touched — sync edits are SPEC artifacts + CHANGELOG only (M3's @MX:ANCHOR/@MX:NOTE in factory.go landed in-run)
- carried_gap: package-wide internal/cli + internal/hook suite verdicts remain CI's (run-phase full-suite attempts died to load with 0 assertion failures, /tmp/t1330-cli-full-suite.log) — unchanged by this artifact-only sync; targeted clusters all green
