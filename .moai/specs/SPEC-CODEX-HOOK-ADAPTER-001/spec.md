---
id: SPEC-CODEX-HOOK-ADAPTER-001
title: "Codex Dual Harness M3 — Hook Adapter: event-name normalization + partial output-dialect mapping"
version: 0.1.0
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
assumed `internal/hook`'s 105 files / 22,659 LOC stay untouched behind a mapping layer.

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
advisory `systemMessage`. Under Codex both would do nothing, with no error and no signal —
the failure mode this SPEC exists to prevent.

### Finding C — exit 2 works, but stderr means different things per event

Exit 2 blocks on PreToolUse and continues on Stop, and the stderr text is treated
differently: PreToolUse surfaces it to the model as the blocking reason, Stop injects it as
a continuation prompt and never displays the text. An adapter handling stderr uniformly
across events would be wrong in one of the two directions.

### Finding D — hook config is silently strict, and project-level wiring did not fire

A top-level `"version"` key disables the entire hooks file with no warning or error
(single-variable A/B: identical file fires without it, fires nothing with it). Separately,
project-level hook discovery was not observed at all: two paths, four config shapes, and
explicit project trust all produced no hook execution, while the identical file placed at
`$CODEX_HOME/hooks.json` fired on the first attempt. This contradicts what M0 recorded and
the cause was not established.

## §B Scope

**In Scope**:

- An adapter seam in front of the `moai hook` dispatcher that normalizes Codex hook event
  names to MoAI's dispatcher argument names.
- A **partial** output mapping covering the three inert keys, and only those.
- Per-event stderr semantics, so a blocking reason and a continuation prompt are not
  conflated.
- Hook config emission constraints for whatever writes a Codex hooks file: no top-level
  `version`, and a field whitelist.
- Golden-file tests keyed to the payload dumps already captured under
  `.moai/reports/t91/hook-payloads/` and `.moai/reports/t83/probe/`.

**Out of Scope — `internal/hook` logic**: the parsing and decision logic stays unchanged.
Payload field names are snake_case-identical between the harnesses and the observed key sets
match the t91 goldens exactly, so the adapter sits in front of the dispatcher rather than
inside it.

**Out of Scope — the wiring generator (t88 / M4)**: this SPEC states the config constraints
it measured but does not build the generator that emits the file.

**Out of Scope — `SubagentStop` mapping**: retired. M0 measured that delegation surfaces as
`PostToolUse` with `tool_name` beginning `collaboration`, and that `SubagentStart`/`SubagentStop`
never fire in this build.

**Out of Scope — Codex-only events**: `PermissionRequest`, `PreCompact`, `PostCompact`,
`SubagentStart` have no MoAI counterpart and are not wired here.

## §C Requirements (GEARS)

### REQ-1 — Event-name normalization

WHERE a Codex hook invokes the adapter, the adapter SHALL map the Codex event name to the
MoAI dispatcher's argument name for the eight events with a counterpart, and SHALL reject an
unrecognized event name with a non-zero exit and a diagnostic rather than defaulting to any
handler.

### REQ-2 — Partial output mapping for the three inert keys

WHERE a MoAI hook emits `continue: false` (with or without `stopReason`), the adapter SHALL
rewrite it to `{"decision": "block", "reason": <stopReason, or a default non-empty string>}`
for events where `decision: "block"` is honored, because Codex rejects a `decision:block`
carrying an empty reason.

WHERE a MoAI hook emits `systemMessage` on an event with a working delivery channel, the
adapter SHALL route it to that channel — `additionalContext` on `UserPromptSubmit`.

### REQ-3 — Silent no-op is prohibited

WHERE the adapter cannot deliver a message on the event it was emitted for, the adapter SHALL
discard it AND SHALL record the discard as a diagnostic naming the event, the key, and the
content length. An advisory that cannot be delivered is not silently dropped.

### REQ-4 — Per-event stderr semantics

WHERE the adapter passes a hook's stderr through on exit 2, it SHALL treat the text as a
blocking reason for `PreToolUse`, `PermissionRequest`, and `UserPromptSubmit`, and as a
continuation prompt for `Stop` and `SubagentStop`, and SHALL NOT apply one event's treatment
to the other class.

### REQ-5 — Hook config emission constraints

WHERE anything in this repository emits a Codex hooks config, it SHALL NOT write a top-level
`version` key, and SHALL restrict emitted keys to the measured-accepted set
(`hooks`, per-event arrays, `matcher`, `hooks[].type`, `hooks[].command`, `hooks[].timeout`).

### REQ-6 — Golden-file tests over both harnesses' payloads

WHERE the adapter parses a hook payload, its tests SHALL assert against the captured payload
dumps rather than hand-written fixtures, covering at minimum `PreToolUse`, `PostToolUse`,
`SessionStart`, `SessionEnd`, `Stop`, `UserPromptSubmit`.

### REQ-7 — `internal/hook` invariance

The adapter SHALL NOT modify decision logic in `internal/hook`. A change there is a scope
violation, not an implementation detail.

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
- `version` key disabling the file: `probe/run-versionkey-kills-file.jsonl`
- project-level non-firing: `probe/run-projectlevel-nofire.jsonl`

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
  on `Stop`, `SessionStart`, or `UserPromptSubmit` is unmeasured, so REQ-2's event coverage
  is stated as a range to confirm at run-phase, not as an established fact.
- `suppressOutput`, `updatedMCPToolOutput`, and the `PermissionRequest` behavior fields are
  unmeasured and unmapped.

## §F Blocker Candidate — project-level wiring (carried for M4)

Project-level hook discovery produced no execution under every combination tried. If the
wiring generator (t88 / M4) is designed to write `.codex/hooks.json` into a project, the
generated file would install and then do nothing, with nothing reporting that. Two items are
preconditions for that card rather than this one:

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
- `.claude/rules/moai/core/verification-claim-integrity.md` — why §E separates measured from
  declared

## §H HISTORY

- 0.1.0 (2026-08-22) — initial draft. Scope set by two rounds of measurement that refuted the
  card's output-dialect premise and replaced it with a three-key partial mapping.
