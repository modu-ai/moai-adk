# Acceptance — SPEC-CODEX-HOOK-ADAPTER-001

> One falsifiable AC per REQ. Each names a command or a file and an observable outcome. Where
> an AC requires a live Codex run, it names the marker to grep for — a passing exit code is
> never the evidence, because every measured failure also exited 0.

## §D AC Matrix

| AC ID | REQ | Severity | Summary |
|---|---|---|---|
| AC-REQ-1a | REQ-1 | MUST-PASS | Each of the 8 Codex event names maps to its MoAI dispatcher name |
| AC-REQ-1b | REQ-1 | MUST-PASS | Unrecognized event name → non-zero exit + diagnostic, no handler runs |
| AC-REQ-2a | REQ-2 | MUST-PASS | `continue:false` + `stopReason` → `decision:block` + that reason |
| AC-REQ-2b | REQ-2 | MUST-PASS | `continue:false` with no `stopReason` → `decision:block` + non-empty default reason |
| AC-REQ-2c | REQ-2 | MUST-PASS | `systemMessage` on UserPromptSubmit → emitted as `additionalContext` |
| AC-REQ-2d | REQ-2 | MUST-PASS | Keys measured working pass through byte-identical |
| AC-REQ-3a | REQ-3 | MUST-PASS | Undeliverable message produces a diagnostic naming event + key + length |
| AC-REQ-3b | REQ-3 | MUST-PASS | No discard path exits without emitting a diagnostic |
| AC-REQ-4a | REQ-4 | MUST-PASS | PreToolUse exit 2 → stderr classified as blocking reason |
| AC-REQ-4b | REQ-4 | MUST-PASS | Stop exit 2 → stderr classified as continuation prompt |
| AC-REQ-4c | REQ-4 | SHOULD | Unmeasured events marked as unmeasured in the table itself |
| AC-REQ-5a | REQ-5 | MUST-PASS | No emitted config carries a top-level `version` |
| AC-REQ-5b | REQ-5 | MUST-PASS | Emitted config keys ⊆ measured-accepted set |
| AC-REQ-6 | REQ-6 | MUST-PASS | Golden tests read the captured dumps, 6 events covered |
| AC-REQ-7 | REQ-7 | MUST-PASS | `internal/hook` decision logic unchanged |

## §D.1 Severity / Traceability

Every REQ has at least one MUST-PASS AC. AC-REQ-4c is SHOULD because it constrains how a
gap is documented rather than how code behaves. A MUST-PASS failure blocks the run-phase
close regardless of the others.

## §D.2 Given-When-Then Scenarios

### AC-REQ-1a — the eight pairs map

**Given** the mapping table, **when** each Codex event name is looked up, **then** it returns
the MoAI dispatcher argument for: PreToolUse, PostToolUse, SessionStart, SessionEnd, Stop,
UserPromptSubmit, and the two the dispatcher distinguishes internally.
**Evidence**: a table-driven test enumerating all pairs; the count asserted explicitly so a
dropped row fails rather than silently shrinking the table.

### AC-REQ-1b — unknown event is rejected, not defaulted

**Given** the adapter, **when** invoked with an event name not in the table, **then** it exits
non-zero and writes a diagnostic naming the received value, and no MoAI handler runs.
**Why it matters**: Codex silently ignores unknown event names in its own config, so an
adapter that also defaults quietly would leave a hook that appears installed and never fires.

### AC-REQ-2a — `continue:false` + `stopReason` is rewritten

**Given** a hook emitting `{"continue": false, "stopReason": "X"}`, **when** the adapter
processes it, **then** the emitted object is `{"decision": "block", "reason": "X"}`.
**Evidence**: unit assertion on the transformed JSON.

### AC-REQ-2b — a reason is always non-empty

**Given** a hook emitting `{"continue": false}` with no `stopReason`, **when** the adapter
processes it, **then** `reason` is present and non-empty.
**Why it matters**: Codex rejects `decision:block` without a non-empty reason — the binary
carries that error string per event. An empty reason would turn a block into a no-op.

### AC-REQ-2c — `systemMessage` uses the working channel where one exists

**Given** a hook emitting `systemMessage` on UserPromptSubmit, **when** the adapter processes
it, **then** the content is emitted as `hookSpecificOutput.additionalContext`.
**Live check**: a Codex run whose hook emits it, grepping the event stream for the marker —
the injection is only proven by the model receiving content absent from the prompt.

### AC-REQ-2d — working keys are not touched

**Given** hooks emitting `permissionDecision`, `additionalContext`, or `decision:block`,
**when** the adapter processes them, **then** output is byte-identical to input.
**Why it matters**: every unnecessary translation is a drift point between harnesses.

### AC-REQ-3a — discard is announced

**Given** a `systemMessage` on an event with no delivery channel, **when** the adapter
discards it, **then** a diagnostic records the event name, the key, and the content length.
**Evidence**: assertion on the diagnostic sink; content itself must NOT appear.

### AC-REQ-3b — no silent path

**Given** the adapter's discard paths, **when** enumerated, **then** each emits a diagnostic.
**Evidence**: a test that drives every discard branch and asserts a non-empty diagnostic for
each; a branch added later without one fails this AC.

### AC-REQ-4a / 4b — stderr means different things

**Given** a PreToolUse hook exiting 2 with stderr text, **then** the adapter classifies it as
a blocking reason. **Given** a Stop hook exiting 2 with stderr text, **then** it classifies it
as a continuation prompt.
**Evidence**: the measured runs are the baseline — `probe/run-exit2.jsonl` shows PreToolUse
stderr surfaced to the model verbatim; `probe/run-stop-exit2.jsonl` shows Stop stderr acted on
but never displayed. The classification must match those two observations.

### AC-REQ-4c — unmeasured events say so

**Given** the stderr classification table, **when** read, **then** events with no measurement
carry an explicit unmeasured marker rather than an inherited default.

### AC-REQ-5a — no `version` key

**Given** any Codex hooks config this repository emits, **when** parsed, **then** it has no
top-level `version`.
**Why it matters**: measured single-variable A/B — the identical file fires with the key
absent and fires nothing with it present, silently.

### AC-REQ-5b — field whitelist

**Given** any emitted config, **when** its keys are collected, **then** they are a subset of
`{hooks, <event>, matcher, type, command, timeout}`.

### AC-REQ-6 — goldens, not fixtures

**Given** the payload tests, **when** inspected, **then** each reads a captured dump under
`.moai/reports/t91/hook-payloads/` or `.moai/reports/t83/probe/`, covering PreToolUse,
PostToolUse, SessionStart, SessionEnd, Stop, UserPromptSubmit.
**Evidence**: the test names the file it reads; a hand-written literal fixture fails this AC.

### AC-REQ-7 — `internal/hook` untouched

**Given** the run-phase diff, **when** filtered to `internal/hook/`, **then** it contains no
change to decision logic.
**Evidence**: `git diff --stat -- internal/hook/` cited in the completion report. A non-empty
result requires justification against REQ-7 rather than being reported as incidental.

## §D.3 Residual Risk (not covered by any AC)

These are stated rather than tested, because the evidence to test them does not exist yet:

- The three inert keys were measured on PreToolUse and PostToolUse only. If one of them is
  live on Stop or SessionStart, AC-REQ-2's event coverage is narrower than reality and the
  mapping does unnecessary work there — wasteful, not incorrect.
- All measurements ran with approvals and hook trust bypassed. The normal approval path is
  unverified.
- One Codex build, one platform, one account. A build that makes an inert key live would make
  part of REQ-2 redundant.
- `suppressOutput`, `updatedMCPToolOutput`, and the `PermissionRequest` behavior fields are
  unmeasured and deliberately unmapped; a MoAI hook that starts emitting them would fall
  through to the discard path, which is why REQ-3 is a MUST-PASS rather than a nicety.
