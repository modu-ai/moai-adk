# Acceptance — SPEC-CODEX-HOOK-ADAPTER-001

> One falsifiable AC per REQ. Each names a command or a file and an observable outcome. Where
> an AC requires a live Codex run, it names the marker to grep for — a passing exit code is
> never the evidence, because every measured failure also exited 0.

## §D AC Matrix

| AC ID | REQ | Severity | Summary |
|---|---|---|---|
| AC-REQ-1a | REQ-1 | MUST-PASS | All 11 table rows map to their dispatcher argument; count asserted |
| AC-REQ-1b | REQ-1 | MUST-PASS | Unrecognized event name → non-zero exit + diagnostic, no handler runs |
| AC-REQ-2a | REQ-2 | MUST-PASS | `continue:false` + `stopReason` → `decision:block` + that reason |
| AC-REQ-2b | REQ-2 | MUST-PASS | `continue:false` with no `stopReason` → `decision:block` + non-empty default reason |
| AC-REQ-2c | REQ-2 | MUST-PASS | `systemMessage` on UserPromptSubmit → emitted as `additionalContext` |
| AC-REQ-2d | REQ-2 | MUST-PASS | Keys measured working pass through byte-identical |
| AC-REQ-3a | REQ-3 | MUST-PASS | Undeliverable message → diagnostic with event + key + length, content absent |
| AC-REQ-3b | REQ-3 | MUST-PASS | Discard-branch count constant matches the tested branches |
| AC-REQ-3c | REQ-3 | MUST-PASS | Hook exited 2 → no diagnostic on stderr |
| AC-REQ-4a | REQ-4 | MUST-PASS | PreToolUse exit 2 → stderr classified as blocking reason |
| AC-REQ-4b | REQ-4 | MUST-PASS | Stop exit 2 → stderr classified as continuation prompt |
| AC-REQ-4c | REQ-4 | SHOULD | Declared-not-measured classes annotated; excluded events absent |
| AC-REQ-5a | REQ-5 | MUST-PASS | Validator rejects a config with a top-level `version` |
| AC-REQ-5b | REQ-5 | MUST-PASS | Validator accepts the measured-good shape, rejects unknown keys per level |
| AC-REQ-6 | REQ-6 | MUST-PASS | Golden tests read the vendored `testdata/`, 6 events covered |
| AC-REQ-7 | REQ-7 | MUST-PASS | Adapter package outside `internal/hook/`; zero pre-existing files modified |

## §D.1 Severity / Traceability

Every REQ has at least one MUST-PASS AC. AC-REQ-4c is SHOULD because it constrains how a gap
is documented rather than how code behaves. A MUST-PASS failure blocks the run-phase close
regardless of the others.

## §D.2 Given-When-Then Scenarios

### AC-REQ-1a — the eleven pairs map, and the count is asserted

**Given** REQ-1's table, **when** each Codex event name is looked up, **then** it returns the
listed dispatcher argument, for all eleven rows.
**And** every dispatcher argument in the table is found among the subcommands registered in
`internal/cli/hook.go`.
**Evidence**: a table-driven test whose case count is asserted against a constant (`11`), so a
dropped row fails rather than silently shrinking coverage. The six rows marked adapted are
additionally asserted to route to a live handler; the five marked not-adapted are asserted to
be recognized-but-refused, distinguishing them from AC-REQ-1b's unknown names. The
registration cross-check is part of this test rather than a separate AC: the first draft
asserted four events had no counterpart, which the registration list contradicts, so the
table's right-hand side must be mechanically checkable rather than re-assertable.

### AC-REQ-1b — unknown event is rejected, not defaulted

**Given** the adapter, **when** invoked with an event name absent from the table, **then** it
exits non-zero and writes a diagnostic naming the received value, and no MoAI handler runs.
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
carries that error string for `Stop` and `SubagentStop`, the two events where it was located.
An empty reason would turn a block into a no-op.

### AC-REQ-2c — `systemMessage` uses the working channel where one exists

**Given** a hook emitting `systemMessage` on UserPromptSubmit, **when** the adapter processes
it, **then** the content is emitted as `hookSpecificOutput.additionalContext`.
**Live check**: a Codex run whose hook emits it, grepping the event stream for the marker —
the injection is only proven by the model receiving content absent from the prompt.

### AC-REQ-2d — working keys are not touched

**Given** hooks emitting `permissionDecision`, `additionalContext`, or `decision:block`,
**when** the adapter processes them, **then** output is byte-identical to input.
**Why it matters**: every unnecessary translation is a drift point between harnesses.

### AC-REQ-3a — discard is announced, without echoing content

**Given** a `systemMessage` on an event with no delivery channel, **when** the adapter discards
it, **then** `.moai/logs/codex-adapter.jsonl` gains a record naming the event, the key, and the
content length.
**Evidence**: assertion on the log record; the discarded content itself must NOT appear in it.

### AC-REQ-3b — the discard branches are counted, not eyeballed

**Given** the adapter's discard paths, **when** the test enumerates them, **then** the number
tested equals a branch-count constant declared beside the implementation.
**Why it matters**: the earlier phrasing claimed a later-added branch would fail the AC, which
nothing enforced. A shared constant makes adding a branch without a test a compile-or-assert
failure rather than a silent gap.

### AC-REQ-3c — the diagnostic never corrupts a blocking reason

**Given** an underlying hook that exited 2, **when** the adapter emits a discard diagnostic,
**then** nothing is written to stderr.
**Why it matters**: stderr on that path is the blocking reason or continuation prompt REQ-4
requires passing through unmodified; a diagnostic line appended there would alter what the
model receives.

### AC-REQ-4a / 4b — stderr means different things

**Given** a PreToolUse hook exiting 2 with stderr text, **then** the adapter classifies it as
a blocking reason. **Given** a Stop hook exiting 2 with stderr text, **then** it classifies it
as a continuation prompt.
**Evidence**: the measured runs are the baseline — `probe/run-exit2.jsonl` shows PreToolUse
stderr surfaced to the model verbatim; `probe/run-stop-exit2.jsonl` shows Stop stderr acted on
but never displayed. The classification must match those two observations.

### AC-REQ-4c — the untested basis stays visible

**Given** the stderr classification table, **when** read, **then** `UserPromptSubmit` carries an
explicit declared-not-measured annotation, and no event excluded by §B appears in it at all.

### AC-REQ-5a — the validator rejects the measured killer

**Given** the REQ-5 validator, **when** handed a config object carrying a top-level `version`,
**then** it rejects with an error naming that key.
**Evidence**: unit test with the exact shape from `probe/run-versionkey-kills-file.jsonl`,
which Codex answered with `unknown field 'version', expected 'description' or 'hooks'`.

### AC-REQ-5b — accepted shape in, unknown keys out

**Given** the validator, **when** handed the measured-working config shape, **then** it
accepts; **when** handed an object with an unknown key at the top level, the entry level, or
the hook level, **then** it rejects and names the offending key and its level.
**Why it matters**: the constraint exists because unknown keys are not rejected loudly by
Codex at every level — the top-level case surfaces only in the `--json` stream, and the
project-level case surfaces not at all.

### AC-REQ-6 — goldens, tracked and reachable

**Given** the payload tests, **when** inspected, **then** each reads a file under
`.moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/testdata/hook-payloads/`, covering PreToolUse,
PostToolUse, SessionStart, SessionEnd, Stop, UserPromptSubmit.
**Evidence**: `git ls-files` shows all six tracked on the branch, so the test runs in a
worktree and in CI. A hand-written literal fixture, or a path into an untracked report
directory, fails this AC.

### AC-REQ-7 — placement, checked mechanically

**Given** the run-phase diff, **when** filtered to files that existed under `internal/hook/`
at this SPEC's base commit, **then** the count of modified files is zero.
**Evidence**: `git diff --name-only <base>..HEAD -- internal/hook/` intersected with the base
commit's file list, cited in the completion report. New files elsewhere are unconstrained;
this is a placement check, not a judgment about which code counts as decision logic.

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
  unmeasured and deliberately unmapped; a MoAI hook that starts emitting them falls through to
  the discard path, which is why REQ-3 is MUST-PASS.
- The five events REQ-1 recognizes but does not adapt have no behavioral coverage at all. They
  are refused rather than silently defaulted, which bounds the damage but does not test them.
