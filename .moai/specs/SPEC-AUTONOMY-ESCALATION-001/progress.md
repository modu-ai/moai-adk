# progress.md — SPEC-AUTONOMY-ESCALATION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- Tier: L. Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (+ this file).
- Requirements: 23 (REQ-AE-001 … REQ-AE-023). Acceptance criteria: 25 (AC-AE-001 … AC-AE-025).
- v0.4.0: plan-audit iteration 3 (FAIL 0.83) repaired; lead rulings 09-26 (3) folded in
  (spec.md §H): card-field resolver without the queue `spec_id`, unified disarm rule with class 10
  renamed `detection-disarmed`, in-process verify at PreToolUse, state file carved out of the
  `.moai/state/` exemption with a hash-chain tamper check, B1-B6. §F re-pinned to A1 v0.5.0 at
  `67a2f55cb` (lead instruction; decider surfaces re-checked). New A1 requests R8-R10.
- v0.4.1: lead ruling 09-26 (4): decider rules removed (value follows the A1 schema; judgment
  rules are A3's); R10 closed — card state and detector audit log move to
  `$MOAI_HOME/db/<project-key>/contract/`. Counts unchanged.
- v0.4.2: plan-audit iteration 4 (FAIL 0.83) Q3-Q5 and m2 repaired; lead ruling 09-26 (5)
  folded in (Q1, Q2): one audit log per card, authoritative for arming. §F pinned to A1 v0.5.2
  `25283ebf8` (single pin); R8, R9, R10 closed. Requirements 23, criteria 25.
- v0.4.3: plan-audit iteration 5 (PASS-WITH-DEBT 0.87) R1, R2, n1, n2 closed in plan: per-card
  lock, widened "log does not show armed" reading, digest always compared, §G corrected. R1 and
  R2 are derived from the design text; first observed in run milestone M5. Counts unchanged.
- v0.3.0: plan-audit iteration 2 (FAIL 0.82) repaired; lead rulings 09-26 (2) folded in
  (spec.md §H): two-layer resolver, contract-void before resolution, Markdown record with YAML
  frontmatter and revoke kinds; A1 request R7 added. The v0.2.1 A3 preconditions, their
  criteria, R5-R6, and the mission-validator projection moved to card t1245 (spec.md §K).
- v0.2.1: two A3 preconditions assigned by the lead (since moved to card t1245).
- v0.2.0: plan-audit iteration 1 (FAIL 0.79) lane-owned defects D4-D16 repaired; lead rulings
  09-26 #1-#6 folded in (spec.md §H); A1 requests R1-R4 listed in spec.md §F.2.
- Base tree: `develop` at `ca1d5dc43`. Card: t1235.
- Run-phase blocked on card t1234 (A1 contract schema) landing on `develop`.
- v0.1.1: contract field names aligned to the A1 draft at `8f77d9a33` (not plan-audited);
  dependent requirements tagged 「A1 plan-audit 통과본으로 재확인」; open items O1-O9 in spec.md §F.

## §E.2 Run-phase Evidence

### M1 — Record path, resolver, activation gate (cycle tdd)

Commits: `b1e2d163e` (plan §C step 3: AC-AE-001 golden captured before any detector code),
`b712799ea` (M1 code + tests, spec.md `draft → in-progress`). Package as proposed:
`internal/escalation` (no rename). Pre-flight §C steps 1-2 were run by the orchestrator before
this delegation (A1 merge `b1a62fb2b` is an ancestor; §F.1 rows match the landed A1).

Every command below ran as one invocation prefixed with
`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY MOAI_FACTORY_RUN MOAI_FACTORY_SLOT GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE &&`.
Raw outputs: `.moai/reports/t1235/run-m1/` (gitignored, card worktree).

| AC | Test (package) | Command | Actual output | HEAD | Status |
|---|---|---|---|---|---|
| AC-AE-001 | `TestEscalationGuidedGolden` (`internal/hook`) | `go test -count=1 ./internal/hook -run '^TestEscalationGuidedGolden$' -v` | `--- PASS: TestEscalationGuidedGolden (0.37s)` / `ok github.com/modu-ai/moai-adk/internal/hook 0.732s` | `b712799ea` | PASS |
| AC-AE-002 | `TestResolveContractByCardField` (`internal/escalation`) | `go test -count=1 ./internal/escalation -run '^(TestResolveContractByCardField\|TestResolverNotArmedCases\|TestRecordPathAndQueueUntouched)$' -v` | `--- PASS: TestResolveContractByCardField (0.04s)` | `b712799ea` | PASS |
| AC-AE-003 | `TestResolverNotArmedCases` (`internal/escalation`) | same invocation | `--- PASS: TestResolverNotArmedCases (0.08s)` | `b712799ea` | PASS |
| AC-AE-021 (path half) | `TestRecordPathAndQueueUntouched` (`internal/escalation`) | same invocation | `--- PASS: TestRecordPathAndQueueUntouched (0.05s)` / `ok github.com/modu-ai/moai-adk/internal/escalation 0.523s` | `b712799ea` | PASS (path half; writer is M2) |
| AC-AE-021 (freeze clause) | `TestTodoHistoryAddsNoSchemaChange` (`internal/kanban`) | `go test -count=1 ./internal/kanban/ -run '^TestTodoHistoryAddsNoSchemaChange$' -v` | `--- PASS: TestTodoHistoryAddsNoSchemaChange (0.02s)` | `b712799ea` tree (run before commit, same bytes) | PASS |

Golden mutant control (AC-AE-001 comparator fires): one golden line edited to
`"pre-bash-ls","output":{"systemMessage":"mutant"}` → `--- FAIL: TestEscalationGuidedGolden`,
`escalation_guided_golden_test.go:193: event 3 differs from golden`, exit 1; file restored.

RED before GREEN (E8), tests written first, then zero-value stubs, then implementation:
- compile RED: `internal/config/autonomy_escalation_test.go:23:9: s.NewAPIDetector undefined` /
  `internal/escalation: no non-test Go files` (exit 1).
- assertion RED against stubs: `gate_test.go:25: Active(mode "contract") = false, want true`;
  `record_test.go:75: fingerprint "" is not 16 lowercase hex characters`;
  `resolver_test.go:106: baseline: armed=false spec="", want armed against SPEC-A-001 (lines [])`;
  `--- FAIL: TestResolveAutonomy_NewAPIDetector`, `--- FAIL: TestAutonomy_CacheSchemaBumpedForNewAPIDetector` (exit 1).

Other E-items at `b712799ea` tree:
- E2: `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `GOOS=darwin GOARCH=arm64 go build ./...` exit 0; `GOOS=linux GOARCH=amd64 go build ./...` exit 0.
- E3: `go test -count=1 -cover ./internal/escalation/` → `coverage: 91.4% of statements`;
  `go test -count=1 -cover ./internal/config/` → `ok … coverage: 82.8% of statements` (package-wide; baseline before M1 not measured).
- E4: `grep -rn "AskUserQuestion\|mcp__askuser" internal/escalation/` exit 1 (0 lines); control on `internal/hook/pre_tool.go` exit 0.
- E5: `golangci-lint run --new-from-rev=3ea951b0f ./internal/escalation/... ./internal/config/... ./internal/hook/...` → `0 issues.`; `go vet` on the three packages exit 0.

M1 decisions taken from the SPEC/plan (no user decision needed): no hook wiring in M1 (the gate
and resolver land as library code; M2 wires them with the writer and audit log, where AC-AE-025
exercises the hook path); resolver audit lines are returned, not persisted (the per-card log is
M2); an unreadable `card` field is a warning, not a claimant; an out-of-set `new_api_detector`
falls back to `graph` with a warning; config cache schema bumped 7 → 8.

### M2 — Record writer, state file, fault handling, acceptance-change, first observation (cycle tdd)

Commit: `f56e9c28d`. Same env-scrub prefix on every command. Raw outputs:
`.moai/reports/t1235/run-m2/` (gitignored). `acceptance.md` unchanged, so `./internal/spec` was
not in scope.

| AC | Test (package) | Command | Actual output | HEAD | Status |
|---|---|---|---|---|---|
| AC-AE-005 | `TestFaultIsNotChecked` (`internal/escalation`) | `go test -count=1 ./internal/escalation -run '^(TestFaultIsNotChecked\|TestAcceptanceChangeTrips\|TestRecordPathAndQueueUntouched\|TestRecordFrontmatterParses\|TestRecordDedupAndRetripAfterResolve)$' -v` | `--- PASS: TestFaultIsNotChecked (0.01s)` | `f56e9c28d` | PASS (skipped on Windows: file modes) |
| AC-AE-006 | `TestAcceptanceChangeTrips` (`internal/escalation`) | same invocation | `--- PASS: TestAcceptanceChangeTrips (0.09s)` | `f56e9c28d` | PASS |
| AC-AE-021 | `TestRecordPathAndQueueUntouched` (`internal/escalation`) | same invocation | `--- PASS: TestRecordPathAndQueueUntouched (0.05s)` | `f56e9c28d` | PASS (now through `WriteRecord`) |
| AC-AE-022 | `TestRecordFrontmatterParses` (`internal/escalation`) | same invocation | `--- PASS: TestRecordFrontmatterParses (0.01s)` | `f56e9c28d` | PASS |
| AC-AE-023 | `TestRecordDedupAndRetripAfterResolve` (`internal/escalation`) | same invocation | `--- PASS: TestRecordDedupAndRetripAfterResolve (0.00s)` / `ok github.com/modu-ai/moai-adk/internal/escalation 0.521s` | `f56e9c28d` | PASS |
| AC-AE-025 | `TestFirstObservationVerifiedAtPreToolUse` (`internal/hook`) | `go test -count=1 ./internal/hook -run '^(TestFirstObservationVerifiedAtPreToolUse\|TestEscalationGuidedGolden)$' -v` | `--- PASS: TestFirstObservationVerifiedAtPreToolUse (0.32s)` | `f56e9c28d` | PASS (subprocess clause measured as "no process beyond the guided baseline"; see decisions) |
| AC-AE-001 | `TestEscalationGuidedGolden` (`internal/hook`) | same invocation | `--- PASS: TestEscalationGuidedGolden (0.41s)` / `ok github.com/modu-ai/moai-adk/internal/hook 1.606s` | `f56e9c28d` | PASS, golden file unchanged since `b1e2d163e` |

AC-AE-001 is now a live guard: with `Active` mutated to return true, the run failed —
`escalation_guided_golden_test.go:150: escalation directory present: …/db/t9001-74c053d1/contract/escalation`
(exit 1); file restored before commit.

RED before GREEN (E8): compile RED `internal/escalation/detector_test.go:30:67: undefined: escalation.CardLog`,
`internal/hook/escalation_first_observation_test.go:47:10: undefined: WithEscalationConfig` (exit 1);
hook assertion RED with the package implemented but not wired —
`escalation_first_observation_test.go:123: log = [], want one not-armed and one warning carrying plan_audit_not_passing`,
`:136: records = [], want one budget-exceeded`, `:165: card not armed on first observation: []`,
`--- FAIL: TestFirstObservationVerifiedAtPreToolUse` (exit 1).

Other E-items at `f56e9c28d` tree:
- E2: `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `GOOS=darwin GOARCH=arm64 go build ./...` exit 0; `GOOS=linux GOARCH=amd64 go build ./...` exit 0.
- E3: `go test -count=1 -race -cover ./internal/escalation/` → `coverage: 85.5% of statements`;
  `go test -count=1 -cover ./internal/hook/` (whole package, under `moai slot` lease `go-test-hook`) → `ok … 267.668s coverage: 86.1% of statements`.
- Neighbours: `./internal/homestate/` ok, `./internal/config/` ok, `internal/kanban` `TestTodoHistoryAddsNoSchemaChange` PASS, 15 targeted `internal/cli` hook/deps tests PASS.
- E4: `grep -rn "AskUserQuestion\|mcp__askuser" internal/escalation/ internal/hook/escalation_observe.go` exit 1.
- E5: `golangci-lint run --new-from-rev=be8e03897 ./internal/escalation/... ./internal/hook/... ./internal/homestate/... ./internal/cli/...` → `0 issues.`

M2 decisions (SPEC/plan-recommended options; none user-visible under the default `guided`):
- AC-AE-025 "no subprocess started by the hook process": existing PreToolUse Write steps already
  start git (guided baseline measured: 2 invocations), so the literal clause is red for reasons
  this SPEC does not touch. Measured instead as zero invocations added by the detector (contract
  count − guided count = 0) plus a direct `escalation.Observe` call starting none (C4/B-5 intent).
- AC-AE-025 (a) and (b) need class 7 and class 3 before M5/M3: M2 ships the operations
  dimension of class 7 and the inside-root class 3 predicate (exempting `.moai/reports/<card>/`,
  `.moai/state/`, `ownership.scratch`); outside-root writes, the contract-store rule, and the
  remaining exemptions stay M3.
- AC-AE-006 needs a `detection-disarmed` record: M2 implements disarm for `contract-absent` and
  `signature-invalid`; `terminal-status`, `card-mismatch`, `state-tamper` stay M5.
- Contract store resolved from `.git` files without git; separate/bare/submodule layouts are a
  fault (nothing arms). `homestate.ProjectKeyForCanonicalRoot` added so the key formula is shared.
- Production PostToolUse had no config provider; `WithEscalationConfig` supplies one without
  changing the handler's `lint_as_instruction` nil-config default (one line in `internal/cli/deps.go`).
- Q5: class 5 is M5; no CI producer is consumed in M2.

### M3 — Path and command classes (cycle tdd)

Commit: `dc76e9f55`. Same env-scrub prefix on every command. Raw outputs:
`.moai/reports/t1235/run-m3/` (gitignored). `acceptance.md` unchanged.

| AC | Test (package) | Command | Actual output | HEAD | Status |
|---|---|---|---|---|---|
| AC-AE-007 | `TestInvariantCommandFailureTrips` (`internal/escalation`) | `go test -count=1 ./internal/escalation -run '^(TestInvariantCommandFailureTrips\|TestOwnershipMoveTrips\|TestOwnershipExemptionsAndOutsideRoot\|TestUnreadableContractFieldIsNotObserved)$' -v` | `--- PASS: TestInvariantCommandFailureTrips (0.06s)` | `dc76e9f55` | PASS |
| AC-AE-009 | `TestOwnershipMoveTrips` (`internal/escalation`) | same invocation | `--- PASS: TestOwnershipMoveTrips (0.05s)` | `dc76e9f55` | PASS |
| AC-AE-010 | `TestOwnershipExemptionsAndOutsideRoot` (`internal/escalation`) | same invocation | `--- PASS: TestOwnershipExemptionsAndOutsideRoot (0.07s)` | `dc76e9f55` | PASS |
| AC-AE-024 | `TestUnreadableContractFieldIsNotObserved` (`internal/escalation`) | same invocation | `--- PASS: TestUnreadableContractFieldIsNotObserved (0.02s)` / `ok github.com/modu-ai/moai-adk/internal/escalation 0.563s` | `dc76e9f55` | PASS |
| AC-AE-008 | `TestFrozenFileUnionTrips` (`internal/hook`) | `go test -count=1 ./internal/hook -run '^(TestFrozenFileUnionTrips\|TestInvariantFailureReachesDetectorFromHooks\|TestEscalationGuidedGolden\|TestFirstObservationVerifiedAtPreToolUse)$' -v` | `--- PASS: TestFrozenFileUnionTrips (0.83s)` / `ok github.com/modu-ai/moai-adk/internal/hook 3.126s` | `dc76e9f55` | PASS |
| AC-AE-001 / AC-AE-025 (regression) | `TestEscalationGuidedGolden`, `TestFirstObservationVerifiedAtPreToolUse` (`internal/hook`) | same invocation | `--- PASS: TestEscalationGuidedGolden (0.75s)`, `--- PASS: TestFirstObservationVerifiedAtPreToolUse (0.55s)` | `dc76e9f55` | PASS |

AC-AE-008 "existing harness-learner deny tests still pass": no test in `internal/hook` names the
harness-learner deny (grep for `HARNESS_FROZEN` / `harness-learner` over `internal/hook/*_test.go`
found none); the whole `internal/hook` package passed (`ok … 388.807s coverage: 86.2%`) and
`go test -count=1 ./internal/harness/` → `ok … 0.891s`.

RED before GREEN (E8): compile RED `classes_m3_test.go:60:45: undefined: escalation.LineNotObserved`,
`:66:35: unknown field Failed in struct literal of type escalation.Event` (exit 1), and in the same
run the hook assertion RED `escalation_frozen_test.go:80: frozen-file records = 0 (…), want 3`;
assertion RED after adding the API fields only: `classes_m3_test.go:98: invariant-violation records = []`,
`:158: … record does not name post-signing immutability`, `:207: ownership-move records = 0 ([]), want 3`,
`:274: unreadable ownership still tripped` (exit 1); wiring RED
`escalation_failure_wiring_test.go:65: invariant-violation records = 0 ([]), want 1` for both subtests (exit 1).

Other E-items at `dc76e9f55` tree:
- E2: `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `GOOS=darwin GOARCH=arm64 go build ./...` exit 0; `GOOS=linux GOARCH=amd64 go build ./...` exit 0.
- E3: `go test -count=1 -race -cover ./internal/escalation/` → `coverage: 87.1% of statements`; whole `internal/hook` (under `moai slot` lease `go-test-hook`) → `coverage: 86.2% of statements`; 6 targeted `internal/cli` deps/hook tests PASS.
- E4: `grep -rn "AskUserQuestion\|mcp__askuser" internal/escalation/ internal/hook/escalation_observe.go` exit 1.
- E5: `golangci-lint run --new-from-rev=7ec8e9e12 ./internal/escalation/... ./internal/hook/... ./internal/cli/...` → `0 issues.`

M3 decisions:
- Session scratchpad root: no runtime field or environment variable supplies it (none found in
  `internal/`); the hook passes none, so in production it is undeterminable and an outside-root
  write that no other root covers is listed not-observed, never tripped (REQ-AE-013 as written).
  Tests supply the root through `Event.ScratchpadDir`.
- Auto-memory roots mirror `moai memory doctor`'s candidate set; the slug rule is duplicated in
  `escalation.MemorySlug` (the CLI's is unexported), tagged `@MX:NOTE` to keep them equal.
- AC-AE-024's premise (a signed contract whose ownership cannot be decoded while the rest verifies)
  is unreachable through A1 verify, which rejects undecodable ownership; M3 reads "unreadable" as
  the arming snapshot's ownership globs being unreadable (absent/empty `write`).
- Invariant command failure is observed from PostToolUseFailure and from a PostToolUse
  `tool_response.exit_code != 0`; the failure handler gets the config through `WithEscalationConfig`.
- Not-observed items are card-log lines of kind `not-observed` (the "checkpoint output").
- Not-armed log growth (orchestrator question): the SPEC specifies one line per hook firing, not
  dedup — REQ-AE-002: "when a hook or checkpoint fires, … zero appends one `not-armed` line";
  REQ-AE-023: "on `unsigned` it shall append a `not-armed` line"; AC-AE-003: "each writes exactly
  one `not-armed` line" per processed call. Behavior left per-call; carried as a sync-audit item.

### M4 — New-architecture/API detector (cycle tdd)

Commit: `263d3e685`. Same env-scrub prefix. Raw outputs: `.moai/reports/t1235/run-m4/` (gitignored).
`acceptance.md` unchanged.

| AC | Test (package) | Command | Actual output | HEAD | Status |
|---|---|---|---|---|---|
| AC-AE-011 | `TestNewAPIAdditionsTrip` and `TestNewAPINotObservedCases` (`internal/escalation`) | `go test -count=1 ./internal/escalation -run '^(TestNewAPIAdditionsTrip\|TestNewAPINotObservedCases)$' -v` | `--- PASS: TestNewAPIAdditionsTrip (4.04s)`, `--- PASS: TestNewAPINotObservedCases (1.12s)`, `ok github.com/modu-ai/moai-adk/internal/escalation 5.657s` | `263d3e685` | PASS (library level; no CLI surface) |

Python fixture (c) both branches observed: with CGO `python main: observed as exported declaration=true, listed not-observed=false`;
with `CGO_ENABLED=0 go test … -run 'TestNewAPIAdditionsTrip|TestNewAPINotObservedCases' -v` →
`python main: observed as exported declaration=false, listed not-observed=true`, both tests PASS.

RED before GREEN (E8): compile RED `newapi_test.go:62:17: undefined: escalation.AdditionExportedDecl`,
`:81:22: undefined: escalation.Checkpoint` (exit 1); assertion RED against stubs
`newapi_test.go:88: additions = [], want exported-declaration Added` (and the four other kinds),
`:148: not_observed = [], want the card base`, `:165: not_observed = [], want "cli-verb (python)"` (exit 1).

Other E-items at `263d3e685` tree:
- E2: `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `GOOS=darwin GOARCH=arm64 go build ./...` exit 0; `GOOS=linux GOARCH=amd64 go build ./...` exit 0; `CGO_ENABLED=0 go vet ./internal/escalation/` exit 0.
- E3: `go test -count=1 -race -cover ./internal/escalation/` → `coverage: 87.2% of statements`; hook regression subset (`TestEscalationGuidedGolden`, `TestFirstObservationVerifiedAtPreToolUse`, `TestFrozenFileUnionTrips`, `TestInvariantFailureReachesDetectorFromHooks`) PASS.
- E4: `grep -rn "AskUserQuestion\|mcp__askuser" internal/escalation/` exit 1.
- E5: `golangci-lint run --new-from-rev=1bdd2ebad ./internal/escalation/...` → `0 issues.`

M4 decisions and the open item:
- Q2 (CLI verb) is NOT mandated: REQ-AE-009 and AC-AE-011 name only "the on-demand checkpoint";
  design.md §C.1 calls `moai escalation check` "a proposed CLI verb (open question Q2)". No verb was
  added. `escalation.Checkpoint` is the checkpoint body with no production caller until the operator
  decides the surface — returned to the orchestrator as a blocker.
- Card base: configured develop branch (git-flow), otherwise `origin/HEAD`, then `main`, `master`.
- Go declarations use `go/parser` (no CGO dependency); other languages use the navigator extractor
  (`astx.Extract`), unsupported → not-observed. One record per addition (design.md §C.10 fingerprint:
  kind + qualified name). Only added/modified source files count; a new package is a directory with
  source added at HEAD and no file at base.

### M5 — Evidence, irreversible action, operational trips, unified disarm (cycle tdd)

Commit: `51088d547`. Same env-scrub prefix. Raw outputs: `.moai/reports/t1235/run-m5/` (gitignored).
`acceptance.md` unchanged.

| AC | Test (package) | Actual output | HEAD | Status |
|---|---|---|---|---|
| AC-AE-004 | `TestDetectorNeverAltersToolCall` (`internal/hook`) | `--- PASS: TestDetectorNeverAltersToolCall (0.81s)` | `51088d547` | PASS (class 4 tripped via `escalation.Checkpoint`, the only class-4 path) |
| AC-AE-012 | `TestContradictoryEvidenceTrips` (`internal/escalation`) | `--- PASS: TestContradictoryEvidenceTrips (0.05s)` | `51088d547` | PARTIAL — (a)(b)(d)(e) PASS; (c) not satisfiable: no on-disk CI verdict producer exists, CI limb not-observed per Q5 ruling |
| AC-AE-013 | `TestIrreversibleActionTrips` (`internal/hook`) | `--- PASS: TestIrreversibleActionTrips (0.16s)` | `51088d547` | PASS |
| AC-AE-014 | `TestBudgetExceededTrips` (`internal/escalation`) | `--- PASS: TestBudgetExceededTrips (0.07s)` | `51088d547` | PASS |
| AC-AE-015 | `TestSameDiagnosticRepeatTrips`, `TestAuditFailAtRetryCapTrips` (`internal/escalation`) | `--- PASS: TestSameDiagnosticRepeatTrips (0.03s)`, `--- PASS: TestAuditFailAtRetryCapTrips (0.02s)` | `51088d547` | PASS |
| AC-AE-016 | `TestDisarmContractLossWritesOneRecord` | `--- PASS: TestDisarmContractLossWritesOneRecord (0.06s)` | `51088d547` | PASS |
| AC-AE-017 | `TestDisarmTransitions` | `--- PASS: TestDisarmTransitions (0.04s)` | `51088d547` | PASS |
| AC-AE-018 | `TestStateTamperJudgedFromCardLog` | `--- PASS: TestStateTamperJudgedFromCardLog (0.05s)` | `51088d547` | PASS (cases a-d) |
| AC-AE-019 | `TestOtherCardLogDoesNotAffectThisCard`, `TestSameCardConcurrentHooksKeepChain` | `--- PASS: … (0.21s)`, `--- PASS: … (0.01s)` | `51088d547` | PASS |
| AC-AE-020 | `TestMissingCardLogReading` | `--- PASS: TestMissingCardLogReading (0.04s)`, `ok github.com/modu-ai/moai-adk/internal/escalation 0.875s` | `51088d547` | PASS (cases a-d) |

Commands: `go test -count=1 ./internal/escalation -run '^(TestContradictoryEvidenceTrips|…|TestMissingCardLogReading)$' -v`;
`go test -count=1 ./internal/hook -run '^(TestDetectorNeverAltersToolCall|TestIrreversibleActionTrips|TestEscalationGuidedGolden)$' -v`
(`--- PASS: TestEscalationGuidedGolden (0.37s)`, `ok … internal/hook 1.712s`).

Per-card lock mutant (R1): lock replaced by a no-op, `go test -count=40 … -run '^TestSameCardConcurrentHooksKeepChain$'`
→ 40/40 FAIL with `disarm_m5_test.go:353: chain broken by concurrent hooks` and `:360: operations 0 -> 1, want +2`;
restored → `ok` at `-count=40`.

RED before GREEN (E8): compile RED `operational_m5_test.go:37:51: unknown field Diagnostic in struct literal`,
`:92:59: undefined: escalation.HookStop`; assertion RED `disarm_m5_test.go:108: detection-disarmed records = [], want one disarm:terminal-status`,
`:192: … want one disarm:state-tamper` (×4), `:257: state-tamper records = 0, want 1`, `operational_m5_test.go:96: turns record = []`,
`:116: same-diagnostic records = []`; hook RED `escalation_m5_test.go:75: irreversible-action records = 0 ([]), want 5`,
`:220: class contradictory-evidence did not trip` (and irreversible, same-diagnostic, audit-cap) (exit 1).
`TestIrreversibleActionRecognizer` (escalation-level table) was added after GREEN for coverage; its behavior's RED is the hook-level one above.

Other E-items at `51088d547` tree:
- E2: `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `GOOS=darwin GOARCH=arm64 go build ./...` exit 0; `GOOS=linux GOARCH=amd64 go build ./...` exit 0.
- E3: `go test -count=1 -race -cover ./internal/escalation/` → `coverage: 88.3% of statements`; whole `internal/hook` (slot lease `go-test-hook`) → `ok … 278.472s coverage: 86.2% of statements`; 6 targeted `internal/cli` deps/hook tests PASS.
- E4: `grep -rn "AskUserQuestion\|mcp__askuser" internal/escalation/*.go internal/hook/escalation_observe.go | grep -v _test` exit 1.
- E5: `golangci-lint run --new-from-rev=13a94c311 ./internal/escalation/... ./internal/hook/... ./internal/cli/...` → `0 issues.` (after fixing 3 test-file findings).

M5 decisions:
- Q5: no on-disk producer of recorded CI verdicts exists (`grep` for check-runs / statusCheckRollup / ci-verdict producers
  in `internal`, `cmd`, `pkg`, `scripts` found none; `scripts/ci-watch` writes nothing under `.moai/`). The CI limb is listed
  not-observed at every commit checkpoint; AC-AE-012 clause (c) cannot pass as written — open item.
- Class 5 reads persisted `audit_multi` results at `<worktree>/.moai/state/audit-multi/*.json` (the path
  `internal/cli/mcp_convergence.go` persists); the "first verdict" is the first pass/fail entry of a backend other than the
  contract's `review.second_model`.
- Class 9 reads the audit-artifact convention files `.moai/reports/<card>/{plan,sync}-audit[-iterN].md` (verdict line
  `Verdict: …`); FAIL at iteration ≥ `audit_retries+1` trips.
- State-tamper judgments record the evidence they consumed (`accounted`) in the disarmed entry; a chain break before the
  last state-tamper entry is treated as judged. After a disarm the resolver may re-arm the same contract in the same event
  (design.md §C.6 step 7), starting a new episode.
- Class 8 diagnostic key: first non-empty failure line (excluding "Exit code"), digits normalized.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
