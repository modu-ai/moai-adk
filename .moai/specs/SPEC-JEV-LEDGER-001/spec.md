---
id: SPEC-JEV-LEDGER-001
title: "Jev decision ledger, shadow-mode hook adapter, and cost-per-correct-decision report"
version: "0.1.2"
status: draft
created: 2026-10-26
updated: 2026-10-26
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/jev + internal/cli (hook dispatch, jev cost) + internal/config"
lifecycle: spec-anchored
tags: "jev, typesafe, ledger, hook, shadow-mode, cost, fail-open, display-only"
tier: M
related_specs: [SPEC-JEV-CORE-001, SPEC-JEV-CONSUMERS-001, SPEC-JEV-RESOLVED-MODEL-001, SPEC-JEV-SKILL-SUGGESTION-001]
---

# SPEC-JEV-LEDGER-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-10-26 | 0.1.0 | Initial plan-phase draft for card t1762 (Class C, P1, "Jev 1/6"). Predecessor of Jev 2, 3 and 5. Scope: a local-only decision ledger, a common hook adapter that runs in shadow mode on `pre-tool` / `stop` / `subagent-stop` under both harnesses, the `workflow.jev.consumers.<name>` config surface, and a `moai jev cost` report. It ships no compiled consumer, so the shipped build sends zero requests even with the gate on. | manager-spec |
| 2026-10-26 | 0.1.1 | Plan-audit iter-2 repair (FAIL 0.6875 at `d1aa4ed656`). REQ-JEVL-002/006/007 restated as subject-naming GEARS sentences. REQ-JEVL-013 fixes the adapter deadline at a 5 s budget minus a 1 s margin (4 s) on every event and harness. REQ-JEVL-016 keeps aggregating valid rows and reports `skipped: N`. REQ-JEVL-007 requires a per-consumer tolerant decode. | manager-spec |
| 2026-10-26 | 0.1.2 | Plan-audit iter-3 repair (FAIL 0.81 at `82aff00a6d`, N1-N6). REQ-JEVL-015 and §D name the cost unit "estimated input tokens" (a byte-length estimate) and label it so in the report. REQ-JEVL-011/012 require a ledger open that follows no symbolic link, never blocks, and writes only to a regular file. The AC, plan, and decision-index changes are recorded in `acceptance.md`, `plan.md`, and `decision-index.md`. | manager-spec |

## Position in the chain

This SPEC builds on the completed chain `SPEC-JEV-CORE-001` → `SPEC-JEV-OPTIN-MEASURE-001` → `SPEC-JEV-CONSUMERS-001` → `SPEC-JEV-GOAL-DIST-001`, and on `SPEC-JEV-RESOLVED-MODEL-001` (card t1768, completed). That last one is why this SPEC sends the alias `jev-latest` and records the interpreted model the response reports. It is the infrastructure card of the "Jev 1/6" series. Later cards (Jev 2, 3, 5) register real consumers on top of it. This one registers none.

---

## §A. Problem Statement

Today every Jev call is in-memory accounting only. `Client.UsageLog()` keeps per-call input tokens and the interpreted model for the lifetime of one process (`internal/jev/jev.go`, `type Usage` and `func (c *Client) record`). Nothing is kept across processes. A hook process exits in milliseconds, so a judgment made inside a hook leaves no record anywhere. Nobody can later ask whether the judgment was right or what it cost.

The later Jev cards want to put typed judgments next to hook events (a tool call about to run, a turn ending, a subagent ending). Such a judgment can only be trusted, tuned, or retired once three things exist:

1. A durable, local, per-question record of what was asked and what came back.
2. A way to run a consumer **in shadow**: asked and recorded, with zero effect on what the hook prints or returns.
3. A way to read the record back as a cost figure ("input tokens per correct decision") without asking the model anything.

None of these exists, and every later card would otherwise build its own.

## §B. Goal

One recorder, one adapter seam on the three hook events, one config surface, one read-only report. All four are inert at the shipped default, display-only by construction, and fail-open on every path.

---

## §C. Requirements (GEARS)

### §C.1 The gate and the empty default

**REQ-JEVL-001** (Capability gate) **Where** `workflow.jev.enabled` is false, the hook adapter shall construct no Jev request and write no ledger row on any `pre-tool`, `stop`, or `subagent-stop` dispatch, under either harness.

**REQ-JEVL-002** (Ubiquitous) The consumer registry shall ship with no compiled consumer. The hook adapter shall run a consumer only when a compiled registration of that name exists **and** `workflow.jev.consumers.<name>` names it, and shall ignore a configured name that has no compiled registration, sending no request and writing no row for it.

### §C.2 The adapter

**REQ-JEVL-003** (Event-driven) **Where** `workflow.jev.enabled` is true, **When** a `pre-tool`, `stop`, or `subagent-stop` hook dispatch completes under the Claude Code harness or the Codex harness (`--harness codex`), the hook adapter shall offer the event to each runnable consumer (REQ-JEVL-002). Each consumer's questions for that event shall travel as one batched request over one state, per `SPEC-JEV-CORE-001` REQ-JEVC-010.

**REQ-JEVL-004** (Ubiquitous) The hook adapter shall run only after the hook's output bytes and exit decision are already fixed. It shall have no channel through which an answer reaches the hook's stdout, stderr, exit code, or permission decision. Display-only is therefore a property of the call order, not a convention.

**REQ-JEVL-005** (State-driven) **While** a consumer's effective mode is `shadow` (in this SPEC, for every configured mode value: `shadow`, `advise`, `steer`, `confirm`, absent, or unrecognized), the hook's stdout bytes, stderr bytes, and exit code shall be byte-identical to the same dispatch with `workflow.jev.enabled` false, under both harnesses.

### §C.3 The config surface

**REQ-JEVL-006** (Ubiquitous) The consumer-config resolver shall accept exactly `shadow`, `advise`, `steer`, and `confirm` as values of `workflow.jev.consumers.<name>.mode`, shall resolve an absent or unrecognized value to `shadow`, and shall return `shadow` as the effective mode for every accepted value in this SPEC. The hook adapter shall branch only on the effective mode, never on the configured value. The ledger recorder shall record the configured value as `configured_mode` and the effective value as `mode` on each row.

**REQ-JEVL-007** (Ubiquitous) The consumer-config resolver shall treat `workflow.jev.consumers.<name>.threshold` as an optional number in the closed interval [0, 1], and shall resolve a value outside that interval, or of any other YAML type, to absent. The config loader shall decode each consumer entry independently, so that a malformed entry is quarantined to that consumer while the rest of the `workflow` section still loads. The ledger recorder shall record the threshold on each row, and the hook adapter shall select no action on it in this SPEC.

**REQ-JEVL-008** (Event-detected) **When** a consumer's configuration is invalid (an unrecognized mode, an out-of-range threshold, or a threshold of a non-numeric type), the hook adapter shall record the condition on that consumer's ledger rows and shall emit nothing to the hook's stdout or stderr.

### §C.4 The ledger

**REQ-JEVL-009** (Ubiquitous) The ledger recorder shall append one JSON Lines row per question per attempted call to `<project root>/.moai/state/jev/decisions.jsonl`. Each row shall carry `model_id` (the response's interpreted model id, empty when nothing was answered, and never the request alias), `request_model`, `rubric_version`, `question_id`, `probability` (the raw answer probability, `null` when unanswered), `mode`, `action`, `outcome_label`, and also `ts`, `consumer`, `event`, `harness`, `availability`, and `input_tokens`.

**REQ-JEVL-010** (Ubiquitous) The recorder shall write `outcome_label` as `unlabeled` at record time and `action` as `none` while the effective mode is `shadow`. Labelling an outcome after the fact is not part of this SPEC.

**REQ-JEVL-011** (Ubiquitous) The ledger path shall stay ignored by git (`.moai/state/` in the repository `.gitignore` and in the template `.gitignore`). The recorder shall create the directory at mode 0700 and the file at mode 0600 on POSIX systems, and shall write each row with a single append so that concurrent hook processes never interleave a row. The recorder shall open the ledger file without following a symbolic link at the ledger path and without blocking on the open, and shall write a row only when the opened handle is a regular file.

**REQ-JEVL-012** (Event-detected) **When** the ledger directory or file cannot be created or written, or the opened ledger handle is not a regular file (a symbolic link, a FIFO, a device, or a directory at the ledger path), the recorder shall drop the row and leave the object at that path, and any link target, unchanged. It shall emit nothing to the hook's stdout or stderr and shall leave the hook's exit code unchanged.

### §C.5 The time budget

**REQ-JEVL-013** (Event-detected) **When** the adapter's Jev deadline elapses before an answer arrives, the hook adapter shall treat the call as no answer. The deadline shall be the same on all three events and both harnesses: a 5 s budget (the tightest host timeout, Claude Code `Stop` / `SubagentStop`) minus a 1 s safety margin, measured from hook process start, so no Jev call outlives 4 s of process time. The adapter shall write one row per question with the observed availability, `deadline_exceeded: true`, and `probability: null`, leave the hook's stdout, stderr, and exit code exactly as they were, and return before the deadline plus a bounded tolerance. **When** the deadline has already passed at adapter entry, the hook adapter shall skip the call entirely, sending no request and writing no row.

### §C.6 Authority and the boundary

**REQ-JEVL-014** (Unwanted) The hook adapter, the ledger, and the cost report shall not make a Jev answer an input to a permission decision, a completion verdict, a merge approval, a queue mutation, an operator gate, or any hook output. They shall add no endpoint, no credential, no API key, and no cost-incurring call path of their own. Every request shall go through `internal/jev.Client.Ask`.

### §C.7 The cost report

**REQ-JEVL-015** (Event-driven) **When** an operator runs `moai jev cost`, the command shall read the ledger and print, without any network call, the number of calls, the total estimated input tokens (`Usage.InputTokens`, a byte-length estimate), the number of rows labelled `correct`, and the estimated input tokens per correct decision. Each figure shall be labelled as an aggregate over model signals, and each token figure shall be labelled "estimated". With zero correct rows, the per-correct figure shall be reported as undefined rather than as a number. The command shall exit 0.

**REQ-JEVL-016** (Event-detected) **When** the ledger file is absent or empty, `moai jev cost` shall report zero calls and exit 0. **When** the ledger carries one or more malformed lines, `moai jev cost` shall skip each malformed line, aggregate the remaining valid rows exactly as REQ-JEVL-015 describes, print `skipped: N` where N is the number of malformed lines, and exit 0. In every case the command shall make no Jev call.

---

## §D. Exclusions

### Out of Scope — real consumers

- Any compiled consumer: no question set, no rubric, no state composition for a real hook event. Jev 2, 3, and 5 add those. Each one carries its own measurement gate under `SPEC-JEV-CONSUMERS-001` REQ-JEVN-015/016, and none is needed here because nothing here is reachable.
- Changing the `jev_ask` MCP tool, `todo --auto` ranking, the near-duplicate finding, or skill suggestion. Each keeps its own call path.

### Out of Scope — non-shadow behaviour

- Executing `advise`, `steer`, or `confirm`. This SPEC accepts the values and runs them as `shadow`. Any surface that prints an advisory, alters a hook's output, or asks for confirmation is a user-surface behaviour change or an operator gate under `SPEC-JEV-CORE-001` REQ-JEVC-012. The card that activates such a mode carries that SPEC's amendment.
- Acting on `threshold`. The value is recorded and never compared against an answer here.

### Out of Scope — outcome labelling

- Any command, hook, or agent that rewrites `outcome_label` after the fact. Rows are written as `unlabeled`. Until a later card supplies labels, `moai jev cost` reports the per-correct figure as undefined on real data.

### Out of Scope — money

- Price constants or currency conversion. The cost unit is estimated input tokens (`Usage.InputTokens`, a byte-length estimate), the only cost quantity `internal/jev` records. It is an estimate of request size, not a billed amount. A price table would be new, unverified data in the binary.
- Reading the response's server-reported `usage.input_tokens` field. This SPEC neither reads nor records it, so no figure in the report is presented as a billed token count.

### Out of Scope — other events and surfaces

- Hook events other than `pre-tool`, `stop`, and `subagent-stop`.
- `moai doctor`, `moai init`, and `moai web`. None of them changes.
- Ledger rotation, retention, or pruning.

---

## §E. Open Questions

Each open question carries a row in `decision-index.md`, with the default the plan assumes.

| Id | Question | Plan default |
|---|---|---|
| Q1 | Does the ledger write rely on the `SPEC-JEV-CORE-001` REQ-JEVC-011 inert-sibling-record carve-out by name, or does it need a CORE amendment? | Relies on the carve-out by name. Evidence is in `plan.md` §C. |
| Q2 | How is the adapter's Jev deadline derived inside the 5 s / 10 s budget? | Resolved (lane decision, iter-2): 5 s budget minus 1 s margin = 4 s from process start, on every event and harness. A deadline already passed skips the call. The 1 s margin is a decision until M3 measures the worst-case wrapper-start-to-exit wall clock (plan.md §B). |
| Q6 | One card or three slices? | Resolved (lane decision, iter-2): one card, carried serially M1 → M2 → M3 → M4 with one commit per milestone. |
| Q3 | What unit does "cost per correct decision" use? | Estimated input tokens (`Usage.InputTokens`, a byte-length estimate), labelled "estimated" in the report. |
| Q4 | Do non-shadow modes run as shadow in this SPEC? | Yes. Accept, record, and run as shadow. |
| Q5 | Is the new `moai jev cost` subcommand a "user-visible surface change" in the sense of card acceptance item 5? | No. It is a new, explicitly invoked, read-only report that consults no Jev. Existing surfaces stay byte-identical. |
