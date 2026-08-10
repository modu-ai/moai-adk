# SPEC-HARNESS-LEARNING-EVO-001 — Acceptance Criteria (L1)

Every criterion below names the command that decides it, and that command actually decides it. Budget: 16 criteria (Tier M ceiling 16).

## §A. Verification conventions

- **Go-test criteria** are decided by the named test function's pass/fail, run with `-count=1`.
- **Grep criteria** are decided by the exact command's match count; a criterion asserting absence names the expected empty output.
- No criterion reads `.moai/state/routing-ledger.jsonl` or `.moai/evolution/telemetry/*.jsonl` as live content. Those files are gitignored runtime state, gate-dependent, and empty in CI. See `plan.md` §G AP-5.
- A criterion that inspects source text rather than behavior is marked as such and is paired with a behavioral assertion; source inspection alone never decides a criterion.

## §B. Store lifecycle

### AC-HLE-001 — create-if-absent creates once, preserves content, and never reroutes; Record still does
- **Given** a temp state dir with no pending row for session `S`
- **When** `Store.RecordIfAbsent` is called twice for `S`, then once more after two delegations and one evidence ref have been appended
- **Then** exactly one pending file exists for `S`, its delegations and evidence refs are unchanged by the third call, and the ledger file has zero rows; and **when** the pre-existing `Store.Record` is called for `S`, exactly one ledger row with `outcome: "reroute"` is appended, as before this SPEC
- **Command**: `go test ./internal/harness/routing/ -run 'TestRecordIfAbsent_Lifecycle|TestRecord_StillReroutesSelf' -count=1`
- **Covers**: REQ-HLE-003, REQ-HLE-004

### AC-HLE-002 — seam I/O stays bounded, and the sweep is not on the per-prompt path
- **Given** a temp state dir holding 5 foreign pending files, and a temp telemetry dir holding day-files for today, yesterday, and three earlier days
- **When** `Store.RecordIfAbsent` runs for a fresh session, and separately the session evidence load runs
- **Then** no foreign pending file is read, modified, or removed and no ledger row is appended by `RecordIfAbsent` (the sweep did not run on this path), while the same fixture under `Store.Record` does sweep; and the evidence load returns records from exactly the today and yesterday files, never from the three earlier ones
- **Command**: `go test ./internal/harness/routing/ -run TestRecordIfAbsent_NoSweepOnCreatePath -count=1` and `go test ./internal/telemetry/ -run TestLoadBySession_TwoDayWindow -count=1`
- **Covers**: REQ-HLE-015
- **Note**: "bounded" is decided by these two behavioral proxies, not by a syscall counter. See §F R4.

### AC-HLE-003 — the create-if-absent path inherits both sweep guards unchanged
- **Given** two foreign pending rows — one aged past `routing.StalenessThreshold` (24h) with its session absent from `active-sessions.json`, and one aged past the threshold with its session listed live
- **When** the dispatch path that does sweep (`Store.Record`) runs for a third session
- **Then** the first foreign row is finalized as `abort` and removed, and the live-listed row is left untouched, demonstrating that both the age guard and the liveness guard still gate the sweep
- **Command**: `go test ./internal/harness/routing/ -run TestSweepStale_AgeAndLivenessGuards -count=1`
- **Covers**: REQ-HLE-004

### AC-HLE-004 — annotate patches without creating or finalizing, and is reachable from the CLI
- **Given** a pending row for `S` with an empty `matched_subcommand`
- **When** `Store.Annotate` is called with subcommand `plan`, tier `M`, and an empty mode; and separately `moai harness ledger annotate --session S --subcommand run --tier L` is invoked against a temp root
- **Then** the pending row shows the patched values, its `mode_selected` is unchanged by the empty field, no ledger row is appended, annotating a session with no pending row is a no-op that creates nothing, and the CLI invocation exits 0 with the annotated values present
- **Command**: `go test ./internal/harness/routing/ -run TestAnnotate -count=1` and `go test ./internal/cli/ -run TestHarnessLedgerAnnotateCmd -count=1`
- **Covers**: REQ-HLE-005

### AC-HLE-005 — schema version is unchanged and pre-existing rows still parse
- **Given** a fixture ledger of rows written before this SPEC
- **When** the routing reader reads them
- **Then** every row parses, `schema_version` is 1 on every row the store writes, and no field was removed or renamed
- **Command**: `go test ./internal/harness/routing/ -run 'TestSchemaVersionStable|TestReader' -count=1`
- **Covers**: REQ-HLE-014

## §C. Mechanical seams

### AC-HLE-006 — a prompt creates a pending row mechanically
- **Given** a temp project root with both harness observation gates open and no pending row
- **When** the UserPromptSubmit handler is invoked with a `HookInput` carrying a prompt and a session id
- **Then** a pending row exists for that session whose `request_digest` is the digest of the prompt and whose `request_class` is the classifier's output
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_UserPromptSubmit_CreatesPending -count=1`
- **Covers**: REQ-HLE-001, REQ-HLE-002

### AC-HLE-007 — no verbatim prompt text is persisted
- **Given** the seam-A test above, run with the distinctive prompt literal `ZZQX-CANARY-PROMPT`
- **When** the handler completes
- **Then** no file anywhere under the temp state dir contains that literal
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_NoRawPromptPersisted -count=1`
- **Covers**: REQ-HLE-016

### AC-HLE-008 — repeated prompts do not reroute
- **Given** an open pending row for session `S`
- **When** the UserPromptSubmit handler is invoked three more times for `S`
- **Then** the ledger has zero rows and the pending row still exists
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_MultiTurnNoReroute -count=1`
- **Covers**: REQ-HLE-003

### AC-HLE-009 — subcommand is set from a literal prefix only, and only once
- **Given** three prompts submitted to one session in order — `"/moai plan add auth"`, `"please run the auth work"`, `"/moai run SPEC-X"` — and separately a fresh session receiving only `"please plan the auth work"`
- **When** each is submitted to the handler
- **Then** the first session's `matched_subcommand` is `"plan"` after all three prompts (the natural-language prompt set nothing; the second literal did not overwrite the first), and the fresh session's `matched_subcommand` is empty
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_LiteralSubcommandFirstWriterWins -count=1`
- **Covers**: REQ-HLE-006

### AC-HLE-010 — the delegation records `agent_type` verbatim, including non-catalog values
- **Given** two SubagentStop inputs for an open pending row — one with `agent_type: "manager-develop"` and one with `agent_type: "audit-hle"` (the observed named-spawn shape from `spec.md` §A.5 caveat 2)
- **When** the seam runs for each
- **Then** the pending row's `delegations` has length 2 with `agent` values `"manager-develop"` and `"audit-hle"` respectively — the second stored verbatim and neither normalized to, nor replaced by, the input's `subject` field
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_SubagentStop_AgentTypeVerbatim -count=1`
- **Covers**: REQ-HLE-007

### AC-HLE-011 — an absent identity is recorded under a distinguishable marker, not an empty string
- **Given** a SubagentStop input whose `agent_type` is absent (the shape of the 842 rows measured in `spec.md` §A.5 caveat 1)
- **When** the seam runs
- **Then** a delegation entry is appended whose `agent` value is a non-empty marker distinct from every retained-catalog name, and the entry is countable as unattributed rather than indistinguishable from an attributed one
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_AbsentAgentTypeMarker -count=1`
- **Covers**: REQ-HLE-008

### AC-HLE-012 — an unobservable outcome is recorded as unknown, not success
- **Given** a SubagentStop input carrying no outcome-bearing signal
- **When** the seam runs
- **Then** the appended delegation's outcome is the explicit unknown marker and its blocker is null; the string `success` does not appear in the entry
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_UnknownOutcomeNotSuccess -count=1`
- **Covers**: REQ-HLE-009

### AC-HLE-013 — the Bash evidence record is reachable in production wiring, and written exactly once
- **Given** a temp project root with the learning gate open
- **When** the harness-observe handler is invoked with a Bash PostToolUse payload carrying a test command and a passing result, and separately with an Edit payload that the `Write|Edit|MultiEdit` post-tool handler also processes
- **Then** the Bash invocation produces exactly one telemetry record with `is_test_pass` set (the path `spec.md` §A.6 measured as unreachable now executes), and the Edit invocation produces exactly one telemetry record in total across both handlers — no duplicate
- **Command**: `go test ./internal/cli/ -run TestHarnessObserve_BashEvidenceRecorded -count=1` and `go test ./internal/hook/ -run TestEvidenceRecord_NoDoubleWriteOnEdit -count=1`
- **Covers**: REQ-HLE-011

### AC-HLE-014 — a terminal signal closes the row, and a row accumulates delegations end to end
- **Given** a temp project root with gates open
- **When** the seam sequence is driven in order — UserPromptSubmit, two SubagentStop invocations, then the Stop path with an observed test-pass signal on the session evidence path; and again with an observed test-fail signal
- **Then** the ledger gains exactly one finalized row per run, with `outcome: "success"` and `"fail"` respectively, a non-empty `evidence_refs`, a `delegations` array of length 2, and the pending file removed
- **Command**: `go test ./internal/cli/ -run TestRoutingLedger_TerminalCloseEndToEnd -count=1`
- **Covers**: REQ-HLE-010
- **Note**: this criterion drives the production handler functions against a temp root; it does not prove that Claude Code invokes those handlers in a live session. See §F R1.

### AC-HLE-015 — every seam is fail-open
- **Given** a state dir made unwritable, and separately a malformed pending file
- **When** each of the three seams runs
- **Then** each returns without error, the handler's own output is unaffected, and a diagnostic was written to the sink
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_FailOpen -count=1`
- **Covers**: REQ-HLE-012

### AC-HLE-016 — a closed gate writes nothing
- **Given** a temp project root with the hook opt-in gate closed, and separately with the learning gate closed
- **When** all three seams run
- **Then** the state dir contains no ledger file and no pending file in both configurations
- **Command**: `go test ./internal/hook/ -run TestRoutingSeam_GatedOffWritesNothing -count=1`
- **Covers**: REQ-HLE-013

## §D. Quality gate

- `go build ./...` exits 0, and `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- `go test ./... -count=1` exits 0 (full suite, not the affected packages alone).
- `golangci-lint run` reports no new finding against the pre-change baseline.
- `moai spec lint .moai/specs/SPEC-HARNESS-LEARNING-EVO-001/spec.md --strict` exits 0.
- `git diff --name-only origin/main...HEAD | grep -cE '^(internal/template/templates/|\.claude/skills/|\.claude/lsel/)'` returns `0` — no template, frozen-skill, or frozen-allowlist path is touched.
- `grep -rn 'delegation.yaml' internal/ --include='*.go' | grep -v _test` returns no output — this SPEC introduces neither a reader nor a writer for the map.
- `grep -rn 'AskUserQuestion' internal/hook/routing_ledger.go` returns no output.

## §E. Definition of Done

- All 16 criteria pass, each cited with its command and observed output.
- The traceability matrix in §G shows every requirement covered.
- The falsification review named in `spec.md` §E is scheduled, not performed — it needs 50 finalized rows, which do not exist at merge time.

## §F. Residual risk

- **R1 — the live-session path is not mechanically verifiable. → DISCHARGED (manually), but NOT converted into a gate.** AC-HLE-014 drives the production handler functions directly against a temp root. It cannot prove that Claude Code invokes those hooks in a real session, because that depends on the runtime's hook dispatch, on `settings.json` registration, and on the fail-closed opt-in gate being enabled locally. **No CI-runnable criterion covers that link, and none was added** — the check remains manual, as stated here originally rather than dressed up as an automated criterion.

  The manual dogfood was performed against a throwaway project (own `system.yaml` with `hook.opt_in.enabled: true`, `.claude/settings.json` pointing at the worktree's `bin/moai`, a minimal Go module so a real `go test` ran). Observed final ledger row from a live `claude -p` session:

  ```json
  {"session_id":"69f536d5","request_class":"other",
   "delegations":[{"agent":"Explore","outcome":"unknown","blocker":null}],
   "outcome":"success",
   "evidence_refs":[{"kind":"gate_exit","value":"0","ref":"session test evidence","terminal":true}]}
  ```

  One row, delegations populated with the correct agent identity, terminal outcome, pending cleared — the shape the original four-row ledger never produced. The M3 path is confirmed independently: `{"is_test_pass":true,"outcome":"success"}` in telemetry, the first such signal against a 37,107-record corpus containing zero. Seams A, B, C and the Bash-evidence path all fired under real dispatch, so **AC-HLE-014's live-invocation caveat is satisfied**. The trace, the twice-firing UserPromptSubmit/Stop finding, and the hook-wrapper naming correction are in `progress.md` §E.2. R8 below remains open.
- **R2 — the gate ships closed.** Every L1 path is inert for distributed users by default. "No rows appeared" is therefore ambiguous until the gate state is checked; a reader must not infer breakage from an empty ledger.
- **R3 — the terminal signal source rests on one unobserved premise.** REQ-HLE-011's chosen restoration path assumes the Bash PostToolUse payload delivered on the matcher-null observe wrapper carries `tool_input.command` and `tool_response`. The wrapper's registration for Bash is verified (`matcher: null`); the payload content on that specific channel is not runtime-observed in this repository. `plan.md` §F M0 gates the milestone on a probe that settles it, with a declared fallback.
- **R4 — "bounded I/O" is decided by proxies.** AC-HLE-002 asserts sweep absence on the create path and a two-day evidence read window. Neither counts syscalls, so a future edit could add a read on a seam path without failing this criterion.
- **R5 — first-writer-wins mislabels multi-subcommand sessions.** A session running `/moai plan` then `/moai run` finalizes as one row labelled `plan` (REQ-HLE-006). Every delegation spawned during the run phase is then attributed to `plan`. The alternative (last-writer-wins) relabels delegations that were observed under the first label, so both are approximations; the split into per-subcommand rows is deferred (`spec.md` §G).
- **R6 — agent identity is a mixed population.** 30.3% of observed `agent_type` values are absent and a further large share are spawn names rather than agent types (`spec.md` §A.5). This SPEC records the field verbatim and pushes the discrimination downstream; that is a deliberate boundary, and it is why `spec.md` §E F1/F2 exist.
- **R7 — the terminal signal rests on an output-text heuristic, not a structured exit code.** The M0 probe measured the observe channel's `tool_response` as carrying no `exit_code` field, so `deriveFromExitCode` can never fire there and pass/fail detection falls entirely to `deriveFromOutputText`'s marker matching (`ok  \t`, `--- FAIL`, `test result:`, and a ` passed`/` failed` word count). A test runner whose output matches none of those yields no signal, the row stays pending, and it closes later as `abort` via the stale sweep rather than as `fail`. The live yield of terminal `success`/`fail` rows may therefore be lower than the code path suggests, and that is only measurable once rows accumulate.
- **R8 — delegations and the terminal outcome can land on different rows (ordering hazard, OPEN).** One dogfood run produced **two** ledger rows for a single session, both with `delegations: []` and `outcome: "success"`. It did **not** reproduce across two subsequent traced runs, each of which produced one correct row — so the following is a hypothesis, not a diagnosis: if a terminal test signal exists before the first subagent stops, the mid-session Stop closes the row early with no delegation, and the later delegation lands on a freshly created row.

  This is an ordering hazard rather than a coding error — each seam does what it is specified to do, and the trigger is that a single `claude -p` invocation fires UserPromptSubmit and Stop **twice each** (trace in `progress.md` §E.2). It is not fixed in 001, by decision.

  **Downstream consequence, which is why it is filed rather than noted:** a session's delegations can be split across rows, or absent from the row that carries the outcome — and that row is the one `SPEC-HARNESS-LEARNING-EVO-002` counts as qualifying. An analyzer that assumes one session yields one row bearing both its delegations and its outcome will under-count support and over-count `delegations: []` rows. Cross-referenced into `SPEC-HARNESS-LEARNING-EVO-002` `plan.md` §B.

## §G. Traceability matrix

| REQ | Criteria |
|---|---|
| REQ-HLE-001 | AC-HLE-006 |
| REQ-HLE-002 | AC-HLE-006 |
| REQ-HLE-003 | AC-HLE-001, AC-HLE-008 |
| REQ-HLE-004 | AC-HLE-001, AC-HLE-003 |
| REQ-HLE-005 | AC-HLE-004 |
| REQ-HLE-006 | AC-HLE-009 |
| REQ-HLE-007 | AC-HLE-010 |
| REQ-HLE-008 | AC-HLE-011 |
| REQ-HLE-009 | AC-HLE-012 |
| REQ-HLE-010 | AC-HLE-014 |
| REQ-HLE-011 | AC-HLE-013 |
| REQ-HLE-012 | AC-HLE-015 |
| REQ-HLE-013 | AC-HLE-016 |
| REQ-HLE-014 | AC-HLE-005 |
| REQ-HLE-015 | AC-HLE-002 |
| REQ-HLE-016 | AC-HLE-007 |
