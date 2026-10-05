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

### M6 — Template default and documentation (cycle tdd)

Commit: `027a84f68`. Same env-scrub prefix. Raw outputs: `.moai/reports/t1235/run-m6/` (gitignored).
`acceptance.md` unchanged. No `.claude/agents/moai/*.md` under the template was edited, so `make agents-emit`
was not run. The local dogfood `.moai/config/sections/workflow.yaml` was left untouched.

| Item | Test (package) | Command | Actual output | HEAD | Status |
|---|---|---|---|---|---|
| Template ships `new_api_detector: graph` explicitly, resolves warning-free, autonomy block neutral | `TestTemplateAutonomyEscalationNewAPIDetector` (`internal/template`) | `go test -count=1 -v -run 'TestTemplateAutonomyEscalationNewAPIDetector\|TestAC_CONTRACT_020\|TestTemplateNoInternalContentLeak\|TestTemplateLearnedWorkflowBlockNeutral\|TestMCPNeutralityTemplateShape\|TestTemplateNeutralityAuditC8Preserve\|TestTemplateNeutralityAudit$\|TestLanguageNeutrality' ./internal/template/` | `--- PASS: TestTemplateAutonomyEscalationNewAPIDetector (0.00s)` | `027a84f68` | PASS |
| A1 template defaults unchanged | `TestAC_CONTRACT_020` (`internal/template`) | same invocation | `--- PASS: TestAC_CONTRACT_020 (0.00s)` | `027a84f68` | PASS |
| Template neutrality (CI-guard pair + 16-language) | `TestTemplateNeutralityAudit`, `TestTemplateNoInternalContentLeak`, `TestTemplateNeutralityAuditC8Preserve`, `TestTemplateLearnedWorkflowBlockNeutral`, `TestMCPNeutralityTemplateShape`, `TestLanguageNeutrality` | same invocation | all six `--- PASS`; `ok github.com/modu-ai/moai-adk/internal/template 1.369s` | `027a84f68` | PASS |
| Shipped key has a reader | `TestShippedConfigKeysHaveReaders` (`internal/config`) | `go test -count=1 -run 'TestShippedConfigKeysHaveReaders' -v ./internal/config/` | `--- PASS: TestShippedConfigKeysHaveReaders (1.47s)` / `ok github.com/modu-ai/moai-adk/internal/config 1.778s` | `027a84f68` | PASS |

RED before GREEN (E8): `autonomy_escalation_template_test.go:39: workflow.autonomy.escalation.new_api_detector = <nil>, want graph written explicitly`
(exit 1), captured before the template edit. After the edit, `internal/config` failed once with
`1 shipped config key(s) are NOT in the triage inventory`; fixed by the inventory entry (class W, evidence reader).

Other E-items at the `027a84f68` tree (template/config tests run on the same bytes before the commit, E1 re-run after):
- `make build` exit 0; `catalog.yaml` unchanged in git.
- E2: `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `GOOS=darwin GOARCH=arm64 go build ./...` exit 0; `GOOS=linux GOARCH=amd64 go build ./...` exit 0.
- E3: `go test -count=1 -cover ./internal/template/...` → template `ok … 62.396s coverage: 82.0%`, agentemit `88.5%`, commandemit `90.0%`
  (package-wide; the template package's pre-M6 figure was not measured, so no delta is claimed).
- E5: `golangci-lint run --new-from-rev=ad7a2404b ./internal/template/... ./internal/config/...` → `0 issues.`
- `internal/cli` whole package (slot lease `go-test-cli`): first run `go test -count=1 ./internal/cli/` hit the go
  default 10m test timeout with zero `--- FAIL` lines (`panic: test timed out after 10m0s`, running
  `TestT1013_PremiseTargetDivergesFromBackup (2s)`, host load ~10-13 with another lane's heavy suite); rerun
  `go test -count=1 -timeout 22m ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 1010.350s`, exit 0.
  The rerun started before the commit on the same Go/template bytes; only progress.md changed after.

M6 decisions:
- The template comment names `graph | off`, what each does, and that the check runs "at an on-demand checkpoint whose
  invocation surface is pending, never inside a tool-call hook". No CLI verb is named (Q2 open); the new test
  refuses `moai escalation` in the autonomy block.
- The inventory header "Total entries: 977" was already stale before M6 (1005 entries); left unchanged (out of scope).

### Addendum — post-sync-audit fixes F2, F1, F10 (lead ruling on `.moai/reports/t1235/sync-audit.md`)

Base `222f91b36`. Commits: `b21474d3f` (F2), `11f14fe21` (F1), `b7c8c97c8` (F10). Same env-scrub prefix.

- **F2 — credentials in class 6/8 commands.** RED (pass-through `MaskCommand` stub):
  `redact_test.go:58: secret written to: [record irreversible-action-….md]` and
  `redact_test.go:80: secret written to: [record same-diagnostic-repeat-….md state file]` (×3), exit 1.
  The state file leaked through the class 8 streak key (raw command); the card log did not.
  Fix: `MaskCommand` (URL userinfo, Authorization values, token/password-style assignments and
  flags, known token prefixes) applied where class 6/8 derive the command and before the class 8
  diagnostic key. GREEN: `--- PASS: TestCommandCredentialsAreMasked`, `--- PASS: TestMaskCommandIsStable`.
  Diagnosis check: a mutant leaving the diagnostic unmasked first PASSED — the diagnostic key rewrites
  digit runs, so the literal secret never survived intact; the probe was tightened to the digit-free
  core and the mutant then failed (`secret written to: [state file record same-diagnostic-repeat-….md]`).
- **F1 — resolved class 5/9 records reopening.** AC-AE-023 governs the writer on a trip ("a third time
  after the record has been set to `status: resolved` … writes `<class>-<fingerprint>-2.md`"); it does not
  make an unchanged evidence file a new observation, so the fix does not conflict with it. RED:
  `evidence_retrip_test.go:64: open_after_resolve_and_recommit=1, want 0` (class 9) and `:86` (class 5), exit 1.
  Fix: `CardState.ConsumedEvidence` (class:fingerprint:file:content-digest, kept across re-arming);
  classes 5 and 9 trip only on unconsumed evidence. The writer is unchanged. GREEN:
  `--- PASS: TestResolvedEvidenceRecordReopensOnlyOnNewEvidence`, with `TestRecordDedupAndRetripAfterResolve`,
  `TestAuditFailAtRetryCapTrips`, `TestContradictoryEvidenceTrips`, `TestStateTamperJudgedFromCardLog` PASS.
  Diagnosis check: a mutant dropping the digest from the key failed the new-evidence assertion
  (`new evidence did not re-trip as -2`), so the gate keys on evidence change.
- **F10** — CHANGELOG now names PreToolUse, PostToolUse, PostToolUseFailure, and Stop (call sites:
  `pre_tool.go:419`, `post_tool.go:156`, `post_tool_failure.go:85`, `stop.go:42`).

Verification at `b7c8c97c8`: `go test -count=1 -race -cover ./internal/escalation/...` → 50 top-level
`--- PASS`, 0 FAIL, `ok … 10.412s coverage: 88.5%`; hook subset (`TestEscalationGuidedGolden`,
`TestFirstObservationVerifiedAtPreToolUse`, `TestFrozenFileUnionTrips`, `TestInvariantFailureReachesDetectorFromHooks`,
`TestIrreversibleActionTrips`, `TestDetectorNeverAltersToolCall`) all PASS, `ok … internal/hook 3.887s`;
`golangci-lint run --new-from-rev=222f91b36 ./internal/escalation/... ./internal/hook/...` → `0 issues.`;
builds windows/amd64, linux/amd64, darwin/arm64 exit 0.

Debt recorded, no code change (lead ruling): F3 (`.moai/reports/<card>/escalation/` inside the class 3
exemption — spec amendment), F4 (Windows process-local lock), F5 (recover registered late), F6 (whole-log
chain per call), F7 (state replaced before its log entry), F8 (empty `write` return before `effective_never`),
F9 (Q2 — no production caller of `Checkpoint`), F11 (sync commit lacks `Authored-By-Agent`), F12
(PostToolUseFailure passes `PostToolUse` as hook name, by design).

### Addendum 2 — re-audit fixes N1, N2, N4 (+N7) (`.moai/reports/t1235/sync-audit-rerun.md`, pinned `7dbd81588`)

Commits: `19b138c60` (N1), `f06de71c9` (N2), `4d93b46c1` (N4, N7). Same env-scrub prefix.

- **N1 — class 7 `audit_retries` budget reopening.** RED:
  `evidence_retrip_test.go:113: open_after_resolve_and_recommit=1, want 0 (records [budget-exceeded-10723a1185de3c70-2.md budget-exceeded-10723a1185de3c70.md])`.
  Fix: the audit_retries trip is gated on class + kind + observed per-kind count through `freshEvidence`;
  turns and operations unchanged. GREEN: `--- PASS: TestResolvedAuditRetriesBudgetReopensOnlyOnHigherCount`
  (resolve → recommit → 0 open; a third iteration re-trips as -2). Challenge: a mutant keying on kind only
  failed `higher count did not re-trip as -2`.
- **N2 — class 6 classified the masked command.** RED:
  ``redact_test.go:106: authorized push tripped class 6: Bash `git push origin WT-token-rotation:***` pushes WT-token-rotation:***``.
  Fix: classify the raw command; mask only the rendered observation and the hashed fingerprint. GREEN:
  `--- PASS: TestAuthorizedPushNotMisclassifiedByMask` (0 records, 0 leaks, including an authorized push with a
  URL credential). Challenge: a mutant rendering the raw command failed the class 6 leak test
  (`secret written to: [record irreversible-action-….md]`).
- **N4 + N7 — mask coverage.** RED: all seven cases of `TestMaskCommandCoversCredentialShapes` failed, among them
  `MaskCommand("mysql -uroot -pFAKEPASSWORDVALUE db") = "mysql -uroot -pFAKEPASSWORDVALUE ***"`. Fix: new rules for
  mysql-family `-p<secret>` (ordered before the flag rule, which read the attached value as a credential-named
  flag and masked the next argument), `-u user:pass`, `login -p <secret>`, JSON credential fields, `*_KEY=`, and
  `sk-`/`sk-ant-`/`AKIA`/`ASIA` prefixes; godoc states the list is best-effort. Fixtures concatenate vendor
  prefixes (N7). GREEN: `--- PASS: TestMaskCommandCoversCredentialShapes`. Challenge: moving the mysql rule after
  the flag rule failed with `"mysql -uroot -p*** ***", lost the non-secret part " db"`.

Verification at `4d93b46c1`: `go test -count=1 -race -cover ./internal/escalation/...` → 53 top-level PASS, 0 FAIL,
`ok … 8.940s coverage: 88.5%`; hook escalation subset incl. `TestEscalationGuidedGolden` all PASS, `ok … internal/hook 2.625s`;
`golangci-lint run --new-from-rev=7dbd81588 ./internal/escalation/... ./internal/hook/...` → `0 issues.`; builds
windows/amd64, linux/amd64, darwin/arm64 exit 0.

Debt recorded, no code change: N3 (a deleted open class 5/9 record stays gone until the evidence changes —
widens the F3 surface; fold into the F3 amendment), N5 (`freshEvidence` marks evidence consumed before the
write, so a failed `WriteRecord` loses the trip), N6 (`ConsumedEvidence` unbounded, O(n) lookup, never pruned).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 027a84f68            # last code commit (M6); this §E.3 lands in the following docs commit
run_status: audit-ready-with-open-items
ac_total: 25
ac_pass_count: 24
ac_partial_count: 1                  # AC-AE-012: (a)(b)(d)(e) PASS, (c) blocked
ac_fail_count: 0
new_warnings_or_lints_introduced: 0  # golangci-lint --new-from-rev per milestone, all "0 issues."
cross_platform_build:
  windows_amd64: exit 0
  darwin_arm64: exit 0
  linux_amd64: exit 0
total_run_phase_files: 58            # git diff --name-only b4f798dcc..HEAD (card base = merge-base with develop), incl. plan-phase files
m1_to_mN_commit_strategy: one code commit + one progress.md evidence commit per milestone, on WT-escalation-detector, unpushed
run_commits: [b1e2d163e, b712799ea, be8e03897, f56e9c28d, 7ec8e9e12, dc76e9f55, 1bdd2ebad, 263d3e685, 13a94c311, 51088d547, ad7a2404b, 027a84f68]
l44_pre_commit_fetch: not-run        # lane does not push (lead batch push)
l44_post_push_fetch: not-applicable
spec_status: in-progress             # in-progress -> implemented is manager-docs' transition
```

AC matrix across milestones (evidence rows in §E.2):

| AC | Milestone | Test | Status |
|---|---|---|---|
| AC-AE-001 | M1 (golden), live guard from M2 | `TestEscalationGuidedGolden` | PASS |
| AC-AE-002 | M1 | `TestResolveContractByCardField` | PASS |
| AC-AE-003 | M1 | `TestResolverNotArmedCases` | PASS |
| AC-AE-004 | M5 | `TestDetectorNeverAltersToolCall` | PASS |
| AC-AE-005 | M2 | `TestFaultIsNotChecked` | PASS (skipped on Windows) |
| AC-AE-006 | M2 | `TestAcceptanceChangeTrips` | PASS |
| AC-AE-007 | M3 | `TestInvariantCommandFailureTrips` | PASS |
| AC-AE-008 | M3 | `TestFrozenFileUnionTrips` | PASS (harness-learner deny clause: no such test exists) |
| AC-AE-009 | M3 | `TestOwnershipMoveTrips` | PASS |
| AC-AE-010 | M3 | `TestOwnershipExemptionsAndOutsideRoot` | PASS |
| AC-AE-011 | M4 | `TestNewAPIAdditionsTrip`, `TestNewAPINotObservedCases` | PASS (library level; no production caller) |
| AC-AE-012 | M5 | `TestContradictoryEvidenceTrips` | PARTIAL — (c) BLOCKED: no on-disk CI verdict producer |
| AC-AE-013 | M5 | `TestIrreversibleActionTrips` | PASS |
| AC-AE-014 | M5 | `TestBudgetExceededTrips` | PASS |
| AC-AE-015 | M5 | `TestSameDiagnosticRepeatTrips`, `TestAuditFailAtRetryCapTrips` | PASS |
| AC-AE-016 | M5 | `TestDisarmContractLossWritesOneRecord` | PASS |
| AC-AE-017 | M5 | `TestDisarmTransitions` | PASS |
| AC-AE-018 | M5 | `TestStateTamperJudgedFromCardLog` | PASS |
| AC-AE-019 | M5 | `TestOtherCardLogDoesNotAffectThisCard`, `TestSameCardConcurrentHooksKeepChain` | PASS |
| AC-AE-020 | M5 | `TestMissingCardLogReading` | PASS |
| AC-AE-021 | M1 (path + freeze), M2 (writer) | `TestRecordPathAndQueueUntouched`, `TestTodoHistoryAddsNoSchemaChange` | PASS |
| AC-AE-022 | M2 | `TestRecordFrontmatterParses` | PASS |
| AC-AE-023 | M2 | `TestRecordDedupAndRetripAfterResolve` | PASS |
| AC-AE-024 | M3 | `TestUnreadableContractFieldIsNotObserved` | PASS (premise reinterpreted, see M3 decisions) |
| AC-AE-025 | M2 | `TestFirstObservationVerifiedAtPreToolUse` | PASS (subprocess clause measured as zero added invocations) |

Open item (operator decision):
- **Q2 — invocation surface of the on-demand checkpoint.** `escalation.Checkpoint` (class 4, new
  architecture/API) has no production caller. REQ-AE-009 and AC-AE-011 name only "the on-demand checkpoint";
  design.md §C.1 calls `moai escalation check` a proposed verb, and plan.md lists Q2 as "Is a new CLI verb
  for on-demand checkpoints acceptable, given it is itself a class-4 event?". The option set was returned to the
  lead in the M4 blocker report; the template documents the surface as pending.

Sync-audit carry items:
1. **Per-call not-armed lines.** One `not-armed` line per processed hook call, not per state change
   (REQ-AE-002, REQ-AE-023, AC-AE-003 as written). Log growth on long unarmed sessions is unbounded by design.
2. **Outside-worktree writes not observed.** No runtime field or environment variable supplies the session
   scratchpad root, so an outside-root write no other root covers is listed not-observed, never tripped (REQ-AE-013).
3. **Two unidentified baseline git calls.** PreToolUse Write in guided mode already starts git twice; which
   steps start them was not identified. AC-AE-025 was measured as zero invocations added by the detector.
4. **Windows.** The per-card lock rides `internal/lockfile`, which on Windows is a process-local mutex, so two
   hook processes of the same card are not serialized there; `TestFaultIsNotChecked`,
   `TestResolverUnreadableContractIsError`, and the AC-AE-025 signed-valid git trap are skipped there.
5. **Harness-learner deny untested.** AC-AE-008's "existing harness-learner deny tests still pass" has no test
   to run: none in `internal/hook` names it; whole `internal/hook` and `internal/harness` passed.
6. **AC-AE-012(c) blocked** (listed above): the CI limb is not-observed at every commit checkpoint.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-26
sync_commit_sha: pending-backfill    # backfilled once this commit's own SHA is known
sync_status: audit-ready-with-open-items
b12_self_test_a: 0        # grep -c 'SPEC-AUTONOMY-ESCALATION-001' CHANGELOG.md before emission
b12_self_test_b: 25       # grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l
b12_self_test_c: pass     # every path cited in the CHANGELOG entry verified with ls (below)
changelog_entry_position: "[Unreleased] > Added, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented"
  plan_md: "unchanged (frontmatter field not touched)"
  acceptance_md: "unchanged (frontmatter field not touched)"
  progress_md: "N/A (this file; §E.4 authored in this commit)"
canary_compliance_check: not-applicable   # this SPEC defines no forward-looking canary policy of its own
```

Open items carried to the sync-audit and to follow-up work (not resolved here, per manager-docs' CHANGELOG-only
authoring boundary):

- **AC-AE-012 PARTIAL** — clause (c) (a recorded CI verdict) is not-observed: no on-disk CI-verdict producer exists
  in this codebase. Follow-up card: **t1268**.
- **Q2 unresolved** — the on-demand new-API checkpoint's invocation surface (a possible new CLI verb) was never
  answered by the operator; only the detection library landed, with no production caller. This is a genuine
  operator decision, not something manager-docs can resolve during sync.
- No docs-site page was touched: neither `spec.md` nor `acceptance.md` names a docs-site requirement for this
  SPEC, and the shipped `workflow.yaml` comment (landed in M6, `027a84f68`) already documents the new key and its
  pending-invocation caveat for template users. `docs-site/content/ko/cli-reference/contract.md` mentions
  `escalate_on` generically but was not modified — a change there is optional editorial follow-up, not a gap this
  SPEC's acceptance criteria require.
