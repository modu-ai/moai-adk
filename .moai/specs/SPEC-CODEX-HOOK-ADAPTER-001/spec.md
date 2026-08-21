---
id: SPEC-CODEX-HOOK-ADAPTER-001
title: "Codex Dual Harness M3 — Hook Adapter: event-name normalization + partial output-dialect mapping"
version: 0.2.0
status: draft
created: 2026-08-22
updated: 2026-08-22
author: lane-7
priority: medium
phase: "v3.2"
module: hook
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "codex, dual-harness, hook, adapter, event-mapping, output-dialect"
related_specs: []
---

# SPEC-CODEX-HOOK-ADAPTER-001

## §A Problem / Motivation

MoAI's hook layer is written against Claude Code. Running the same harness under
`codex-cli` requires a translation seam. The card that motivated this work (t83) assumed
three differences to absorb — event names, an output dialect, and operational limits — and
assumed `internal/hook`'s existing files stay untouched behind a mapping layer.

Two rounds of direct measurement against `codex-cli 0.147.0` changed what that seam has to
do. The measurements are recorded in `.moai/reports/t83/precondition-measurement.md` and
`.moai/reports/t83/precondition-measurement-round3.md`, with raw event streams under
`.moai/reports/t83/probe/`.

### Finding A — the output dialect is shared, not divergent

The card framed the difference as `decision` versus `permissionDecision`/`continue`. That
framing does not hold: Codex 0.147.0 implements Claude Code's hook output schema wholesale —
`decision`/`reason`, `continue`/`stopReason`, `suppressOutput`, `systemMessage`, and
`hookSpecificOutput` carrying `hookEventName`, `permissionDecision`,
`permissionDecisionReason`, `additionalContext`, `updatedMCPToolOutput`. Those are not two
harnesses' rival dialects; they are coexisting mechanisms inside one schema both implement.

### Finding B — but three keys are declared and inert, and they are the ones MoAI uses

Declaration is not behavior. Measured per key:

| Key | Event measured | Result |
|---|---|---|
| `hookSpecificOutput.permissionDecision: "deny"` + reason | PreToolUse | works — call blocked, reason reaches the model |
| `hookSpecificOutput.additionalContext` | UserPromptSubmit | works — content reaches the model |
| `decision: "block"` + `reason` | Stop | works — turn continues, model follows the reason |
| **`continue: false`** | PreToolUse, PostToolUse | **no observable effect** |
| **`stopReason`** | PreToolUse, PostToolUse | **never surfaces** (event stream and stderr both zero) |
| **`systemMessage`** | PreToolUse, PostToolUse | **never surfaces** |

The inert three are exactly what MoAI's own hooks emit: `team-ac-verify.sh` rejects a task
with `{"continue": false, "stopReason": …}`, and the sync-phase quality gate emits an
advisory `systemMessage`. Under Codex both would do nothing.

### Finding C — exit 2 works, but stderr means different things per event

Exit 2 blocks on PreToolUse and continues on Stop, and the stderr text is treated
differently: PreToolUse surfaces it to the model as the blocking reason, Stop injects it as
a continuation prompt and never displays the text. An adapter handling stderr uniformly
across events would be wrong in one of the two directions.

### Finding D — a top-level `version` key disables the hooks file; the report is easy to miss

A top-level `"version"` key disables the entire hooks file (single-variable A/B: the
identical file fires without it and fires nothing with it). The failure **is** reported, but
only as an item in the `--json` event stream:

```
failed to parse hooks config <path>/hooks.json:
  unknown field `version`, expected `description` or `hooks` at line 2 column 11
```

There is no interactive warning and the process still exits 0, so a caller that checks the
exit code or greps for its own markers sees nothing wrong — which is how this was initially
mis-recorded as a silent failure. The error names file, field, line, and column, and it also
establishes the accepted top-level key set: `description` and `hooks`.

### Finding E — project-level hook wiring never fired, and that failure is genuinely silent

Project-level hook discovery was not observed at all: two paths, four config shapes, and
explicit project trust all produced no hook execution, while the identical file placed at
`$CODEX_HOME/hooks.json` fired on the first attempt. Unlike Finding D, this one emits
nothing — `probe/run-projectlevel-nofire.jsonl` contains zero error items. This contradicts
what M0 recorded and the cause was not established.

## §B Scope

**In Scope**:

- An adapter seam in front of the `moai hook` dispatcher that normalizes Codex hook event
  names to MoAI's dispatcher argument names, per the table in REQ-1.
- A **partial** output mapping covering the three inert keys, and only those.
- Per-event stderr semantics, so a blocking reason and a continuation prompt are not
  conflated.
- A validator expressing the measured hook-config constraints, consumed by the wiring
  generator card rather than invoked here.
- Golden-file tests over the payload dumps vendored into this SPEC's `testdata/`.

### Out of Scope — `internal/hook` decision logic

- The parsing and decision logic under `internal/hook/` is not modified. Payload field names
  are snake_case-identical between the harnesses and the observed key sets match the vendored
  goldens exactly, so the adapter sits in front of the dispatcher rather than inside it.
- The adapter's own package lives outside `internal/hook/` (REQ-7), so "no change" is
  mechanically checkable rather than a judgment about which files count as decision logic.

### Out of Scope — the wiring generator (t88 / M4)

- This SPEC ships the constraint validator (REQ-5) but does not build the thing that writes a
  Codex hooks file into a project or a home directory.
- The generator card consumes REQ-1's table and REQ-5's validator as inputs.

### Out of Scope — events registered but not adapted

- MoAI's dispatcher registers a counterpart for **every** Codex event (see REQ-1's table), so
  excluding an event here is a scoping decision about measurement coverage, **not** an absence
  of a counterpart. An earlier draft asserted the absence; that was false.
- Excluded by lack of measurement: `PreCompact`, `PostCompact`, `PermissionRequest`,
  `SubagentStart`. No payload capture and no behavioral observation exists for any of them, so
  adapting them would mean guessing at their contracts.
- Excluded by measurement: `SubagentStop`. M0 measured that delegation surfaces as
  `PostToolUse` with `tool_name` beginning `collaboration`, and that `SubagentStart` /
  `SubagentStop` never fire in this build. Mapping it would wire a dead path.

### Out of Scope — unmeasured output keys

- `suppressOutput`, `updatedMCPToolOutput`, and the `PermissionRequest` behavior fields
  (`behavior`, `updatedInput`, `updatedPermissions`, `interrupt`) are declared in the binary
  but never exercised.
- A MoAI hook that begins emitting one of them falls through to the discard path, which is why
  REQ-3 is a MUST-PASS rather than a convenience.

## §C Requirements (GEARS)

### REQ-1 — Event-name normalization

WHERE a Codex hook invokes the adapter, the adapter SHALL map the Codex event name to the
MoAI dispatcher argument name using exactly this table, and SHALL reject an unrecognized
event name with a non-zero exit and a diagnostic rather than falling through to any handler.

| Codex event | MoAI dispatcher argument | Adapted by M3 |
|---|---|---|
| `PreToolUse` | `pre-tool` | yes |
| `PostToolUse` | `post-tool` | yes |
| `SessionStart` | `session-start` | yes |
| `SessionEnd` | `session-end` | yes |
| `Stop` | `stop` | yes |
| `UserPromptSubmit` | `user-prompt-submit` | yes |
| `PreCompact` | `compact` | no — unmeasured (§B) |
| `PostCompact` | `post-compact` | no — unmeasured (§B) |
| `PermissionRequest` | `permission-request` | no — unmeasured (§B) |
| `SubagentStart` | `subagent-start` | no — unmeasured (§B) |
| `SubagentStop` | `subagent-stop` | no — dead path, measured (§B) |

Eleven Codex events, eleven dispatcher counterparts, six adapted. The dispatcher argument
names are the registered subcommands in `internal/cli/hook.go`.

### REQ-2 — Partial output mapping for the three inert keys

WHERE a MoAI hook emits `continue: false` (with or without `stopReason`), the adapter SHALL
rewrite it to `{"decision": "block", "reason": <stopReason, or a default non-empty string>}`
for events where `decision: "block"` is honored, because Codex rejects a `decision:block`
carrying an empty reason.

WHERE a MoAI hook emits `systemMessage` on an event with a working delivery channel, the
adapter SHALL route it to that channel — `additionalContext` on `UserPromptSubmit`.

WHERE a MoAI hook emits a key measured to work natively, the adapter SHALL pass it through
unmodified.

### REQ-3 — Silent no-op is prohibited

WHERE the adapter cannot deliver a message on the event it was emitted for, the adapter SHALL
discard it AND SHALL record the discard as a diagnostic naming the event, the key, and the
content length, written to the adapter's own log sink at `.moai/logs/codex-adapter.jsonl`.

WHERE the underlying hook exited 2, the adapter SHALL NOT write the diagnostic to stderr,
because stderr on that path carries the blocking reason or continuation prompt REQ-4 requires
passing through unmodified.

### REQ-4 — Per-event stderr semantics

WHERE the adapter passes a hook's stderr through on exit 2, it SHALL treat the text as a
blocking reason for `PreToolUse` and as a continuation prompt for `Stop`, and SHALL NOT apply
one class's treatment to the other.

WHERE the event is one whose stderr class is declared in the Codex binary but never observed
(`UserPromptSubmit` — blocking reason), the adapter SHALL carry the classification with an
explicit declared-not-measured annotation in the table, so the untested basis stays visible at
the point of use. Events excluded from adaptation by §B receive no stderr class here.

### REQ-5 — Hook config constraint validator

WHERE a Codex hooks config object is presented to the validator this SPEC ships, the validator
SHALL reject it when it carries a top-level key outside `{description, hooks}` — `version`
being the measured instance — and SHALL reject a hook entry carrying a key outside
`{matcher, hooks}` or `{type, command, timeout}` at their respective levels.

### REQ-6 — Golden-file tests over captured payloads

WHERE the adapter parses a hook payload, its tests SHALL assert against the payload dumps
vendored at `.moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/testdata/hook-payloads/`, covering
`PreToolUse`, `PostToolUse`, `SessionStart`, `SessionEnd`, `Stop`, and `UserPromptSubmit`.

### REQ-7 — Adapter package placement

The adapter SHALL live in a package outside `internal/hook/`, and the change set SHALL contain
zero modifications to files that existed under `internal/hook/` before this SPEC.

## §D Evidence (Observed)

Every claim in §A is attributed to a recorded run. Commands, isolation, and raw output are in
the two measurement reports; the event streams are preserved under `.moai/reports/t83/probe/`
with absolute paths masked.

- exit-2 block, JSON deny, additionalContext injection: `probe/run-exit2.jsonl`,
  `probe/run-jsondeny-addctx.jsonl`
- `decision:block` on Stop, Stop exit 2 continuation: `probe/run-stop-decisionblock.jsonl`,
  `probe/run-stop-exit2.jsonl`
- `continue:false` inert: `probe/run-continuefalse-inert-pretool.jsonl`,
  `probe/run-continuefalse-inert-posttool.jsonl`
- `version` key disabling the file, and the parse-error item that reports it:
  `probe/run-versionkey-kills-file.jsonl` line 4
- project-level non-firing, with zero error items: `probe/run-projectlevel-nofire.jsonl`
- dispatcher counterpart registrations: `internal/cli/hook.go`

Measurement isolation: a scratch `CODEX_HOME`, with the real `~/.codex/hooks.json` mtime
unchanged across the whole session and the probe project never entering the real
`config.toml`. Auth was borrowed by symlink rather than copied, and the link was removed
afterwards.

## §E Constraints / Non-Goals

- **The `systemMessage` loss is a real reduction, not a lossless translation.** The
  `continue:false` → `decision:block` rewrite preserves meaning; discarding an undeliverable
  `systemMessage` removes an advisory the Claude Code side shows. This SPEC does not claim
  parity there, and REQ-3 exists so the reduction is visible rather than silent.
- **The measurements bind one build.** `codex-cli 0.147.0`, one platform, one account. An
  inert key may become live in a later build, which would make part of REQ-2 unnecessary
  rather than wrong.
- **Everything was measured with approvals and hook trust bypassed.** Behavior under the
  normal approval path is unmeasured, and `permission_mode` was only ever observed as
  `bypassPermissions`.
- **The inert three were measured on `PreToolUse` and `PostToolUse` only.** Whether they work
  on `Stop`, `SessionStart`, or `UserPromptSubmit` is unmeasured, so REQ-2's event coverage is
  a range to confirm at run-phase, not an established fact.
- **Two failure modes report differently.** The `version` key produces a machine-readable error
  item (Finding D); project-level non-firing produces nothing at all (Finding E). Only the
  second is silent, and only it justifies a "nothing reports this" framing.

## §F Blocker Candidate — project-level wiring (carried for M4)

Project-level hook discovery produced no execution under every combination tried, and produced
no diagnostic of any kind. If the wiring generator (t88 / M4) is designed to write
`.codex/hooks.json` into a project, the generated file would install and then do nothing, with
nothing reporting that. Two items are preconditions for that card rather than this one:

1. Establish whether project-level hook discovery is supported in the target build.
2. If it is not, either retarget generation at `$CODEX_HOME/hooks.json` or redesign the
   user-facing installation path.

M0 §6 recorded the opposite result and could not be reproduced here; re-measurement is
recommended over trusting either record.

## §G Cross-References

- `.moai/reports/t91/README.md` — M0 premise measurement (event enum, payload goldens,
  subagent hook refutation)
- `.moai/reports/t83/precondition-measurement.md` — round 2 (exit 2, stdout JSON), carrying a
  correction banner for a retracted inference
- `.moai/reports/t83/precondition-measurement-round3.md` — round 3 (MoAI's own keys, Stop
  exit 2, discovery path, matcher)
- `.moai/reports/t83/plan-audit.md` — iteration-1 audit; §H records what it changed
- `internal/cli/hook.go` — the dispatcher subcommand registrations REQ-1's table maps onto
- `.claude/rules/moai/core/verification-claim-integrity.md` — why §E separates measured from
  declared

## §H HISTORY

- 0.2.0 (2026-08-22) — revised after iteration-1 audit (FAIL 0.63). Two factual corrections:
  the dispatcher does register a counterpart for all eleven Codex events, so "no MoAI
  counterpart" was false and the exclusions are restated as measurement-coverage scoping; and
  the `version` key is not silent — the probe's own event stream carries a parse error naming
  file, field, line, and column, so the silence claim now applies only to project-level
  non-firing. Also: REQ-1's table enumerated, goldens vendored into `testdata/` so REQ-6 is
  executable from the branch, exclusions converted to the project's heading convention, REQ-5
  re-aimed at a validator this SPEC ships, REQ-3's sink named, REQ-4's unmeasured classes
  annotated, and REQ-7 pinned to package placement.
- 0.1.0 (2026-08-22) — initial draft. Scope set by two rounds of measurement that refuted the
  card's output-dialect premise and replaced it with a three-key partial mapping.
