# Acceptance — SPEC-CODEX-PHASE2-001

> Every criterion is binary-testable: a Go test, a command exit code, or a grep with a stated expected count. Criteria are written as Given-When-Then; the GEARS requirements they verify live in `spec.md` §D.

## Severity legend

- **MUST** — blocks Definition of Done.
- **SHOULD** — non-blocking; a miss is recorded as debt with a rationale.

## §A. AC Matrix

### M1 — Reusable session + model/effort SSOT

**AC-CX2-001** (MUST, REQ-CX2-001) — *Given* a canned session runner that answers `initialize`, `thread/start`, and two successive turns, *When* the caller issues a second turn on the returned session handle, *Then* the exchange contains exactly one `initialize` and one `thread/start`, both turns carry the same `threadId`, and both return their result.

**AC-CX2-002** (MUST, REQ-CX2-001) — *Given* the refactored session, *When* `go test ./internal/cli/... -run 'Codex|ReviewGate'` runs, *Then* every pre-existing codex and review-gate test passes unmodified in its assertions, demonstrating the two existing `runCodexReviewRPC` callers are behaviorally unchanged.

**AC-CX2-003** (MUST, REQ-CX2-002) — *Given* a codex request with no explicit model, *When* the request params are captured at the transport seam, *Then* they carry the model resolved by `template.ResolveAgentModelEffort`; and *Given* an explicit `model` argument, *Then* the transmitted params carry that value — the current drop at `buildCodexReviewParams` (`mcp_codex.go:433`) is closed.

**AC-CX2-004** (MUST, REQ-CX2-002) — *Given* the M1 tree, *When* `go test ./internal/cli/ -run TestMCPAudit_NoDirectFrontmatterRead` runs, *Then* it passes, and a companion test asserts the resolver is invoked on the codex path, so the guard can no longer pass vacuously.

### M2 — Job registry

**AC-CX2-005** (MUST, REQ-CX2-003) — *Given* a writable temporary project root, *When* a background task is started, *Then* exactly one file appears under `.moai/state/codex-jobs/`, and its JSON decodes with a non-empty job id, a status, creation and update timestamps, a `threadId`, a mode, and a request summary.

**AC-CX2-006** (MUST, REQ-CX2-004) — *Given* a job progressing through its lifecycle, *When* each transition is written, *Then* the recorded status is one of `queued`/`running`/`completed`/`failed`/`cancelled` at every read, and no read observes a truncated or partially-written file.

**AC-CX2-007** (MUST, REQ-CX2-004) — *Given* a project root whose `.moai/state/` cannot be written, *When* a background task is requested, *Then* the tool returns a structured error result, the test process does not panic, and the server remains able to serve a subsequent tool call.

**AC-CX2-008** (MUST, REQ-CX2-005) — *Given* a job record produced with credential-shaped values present in the environment and in the request, *When* the record file is read, *Then* it contains no API key, no token, and no serialized environment block, verified by asserting the absence of the seeded sentinel values.

### M3 — `codex_task`

**AC-CX2-009** (MUST, REQ-CX2-006) — *Given* a canned codex session that completes a turn, *When* `codex_task` is invoked with `background` false, *Then* the result carries the task output; and *When* invoked with `background` true, *Then* the result carries a job id that resolves to an existing job record.

**AC-CX2-010** (MUST, REQ-CX2-007) — *Given* a project with the write opt-in absent (the distributed default), *When* `codex_task` is invoked with `write` true, *Then* the turn is not run in a writing mode and the result states that the write request was not honored; and *Given* the opt-in enabled, *Then* the writing mode is requested.

**AC-CX2-011** (MUST, REQ-CX2-008) — *Given* a recorded prior thread for the project, *When* `codex_task` runs with `resume_last`, *Then* the transmitted `threadId` equals the recorded one and no new `thread/start` is issued; and *Given* no recorded thread, *Then* a new thread is opened and the result states that no prior thread was resumed.

### M4 — Job control

**AC-CX2-012** (MUST, REQ-CX2-009) — *Given* an existing job, *When* `codex_job_status` is invoked with its id, *Then* the record is returned; and *Given* an unknown id, *Then* a structured not-found result is returned with `IsError` set rather than a Go error that aborts the call.

**AC-CX2-013** (MUST, REQ-CX2-010) — *Given* a job in a terminal status, *When* `codex_job_result` is invoked, *Then* the recorded output is returned; and *Given* a running job, *Then* the current status is returned and the call completes without waiting for the turn.

**AC-CX2-014** (MUST, REQ-CX2-011) — *Given* a running job backed by a canned session, *When* `codex_job_cancel` is invoked, *Then* the M0-confirmed interrupt request is sent on that job's session, the job's recorded status becomes `cancelled`, and *When* the process does not exit within the grace window, *Then* it is terminated and the call still returns within a bounded time.

**AC-CX2-015** (MUST, REQ-CX2-012) — *Given* a job record naming a pid the server did not spawn, *When* `codex_job_cancel` is invoked for it, *Then* no signal is sent to that pid and the tool returns a structured refusal — asserted by a test that records signal attempts rather than by observing a live process.

### Cross-cutting

**AC-CX2-016** (MUST, REQ-CX2-013/014/015) — *Given* the completed tree, *When* the following batch runs, *Then* every command meets its stated expectation:

```bash
# tool registration + schema (expect 4 new tool names present)
grep -c 'codex_task\|codex_job_status\|codex_job_result\|codex_job_cancel' internal/cli/mcp_server.go   # >= 4

# subagent boundary (expect no output)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go internal/cli/codex_*.go \
  | grep -v '_test.go' | grep -v '//'

# no template surface touched (expect empty)
git status --porcelain internal/template/templates/

# cross-platform build (expect exit 0 twice)
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
```

## §B. Edge cases (negative tests, MUST)

- Codex binary absent from PATH: every new tool returns a structured fail-open result; none panics, and none blocks the caller.
- Codex spawns but the session closes before the handshake completes: the tool returns a structured result naming the failure; no goroutine leaks past the test.
- Two concurrent background tasks in the same project: two distinct job ids, two distinct record files, neither overwriting the other.
- A job record file that is present but malformed JSON: `codex_job_status` reports it as unreadable rather than crashing the server.
- Cancel invoked on an already-terminal job: returns the terminal status and sends nothing.

## §C. Definition of Done

- Every MUST criterion above passes with cited command output.
- `go test ./...` passes; `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.
- `internal/cli` package coverage is at or above its pre-change level.
- `golangci-lint run` introduces no new findings against the pre-change baseline.
- Both `[NEEDS CLARIFICATION]` markers in `plan.md` §D M0 are resolved and the resolutions are recorded before run-phase entry.
- `git status --porcelain internal/template/templates/` is empty.

## §D. Traceability

| REQ | AC |
|---|---|
| REQ-CX2-001 | AC-CX2-001, AC-CX2-002 |
| REQ-CX2-002 | AC-CX2-003, AC-CX2-004 |
| REQ-CX2-003 | AC-CX2-005 |
| REQ-CX2-004 | AC-CX2-006, AC-CX2-007 |
| REQ-CX2-005 | AC-CX2-008 |
| REQ-CX2-006 | AC-CX2-009 |
| REQ-CX2-007 | AC-CX2-010 |
| REQ-CX2-008 | AC-CX2-011 |
| REQ-CX2-009 | AC-CX2-012 |
| REQ-CX2-010 | AC-CX2-013 |
| REQ-CX2-011 | AC-CX2-014 |
| REQ-CX2-012 | AC-CX2-015 |
| REQ-CX2-013 | AC-CX2-016 |
| REQ-CX2-014 | AC-CX2-016 |
| REQ-CX2-015 | AC-CX2-016 |
