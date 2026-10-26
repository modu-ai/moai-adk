---
id: SPEC-JEV-LEDGER-001
title: "Jev decision ledger recorder and per-consumer config surface"
version: "0.2.1"
status: in-progress
created: 2026-10-26
updated: 2026-10-27
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/jevledger (new) + internal/config + internal/template"
lifecycle: spec-anchored
tags: "jev, typesafe, ledger, config, fail-open, display-only, split-a"
tier: M
related_specs: [SPEC-JEV-CORE-001, SPEC-JEV-CONSUMERS-001, SPEC-JEV-RESOLVED-MODEL-001, SPEC-JEV-SKILL-SUGGESTION-001]
---

# SPEC-JEV-LEDGER-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-10-26 | 0.1.0 | Initial plan-phase draft for card t1762 (Class C, P1, "Jev 1/6"). Scope: a local-only decision ledger, a shadow-mode hook adapter on `pre-tool` / `stop` / `subagent-stop` under both harnesses, the `workflow.jev.consumers.<name>` config surface, and a `moai jev cost` report. | manager-spec |
| 2026-10-26 | 0.1.1 | Plan-audit iter-2 repair (FAIL 0.6875 at `d1aa4ed656`). | manager-spec |
| 2026-10-26 | 0.1.2 | Plan-audit iter-3 repair (FAIL 0.81 at `82aff00a6d`, N1-N6). | manager-spec |
| 2026-10-26 | 0.2.0 | **Scope reduction by factory-leader ruling (operator-delegated)** after the Tier M audit ceiling ended in FAIL 0.88 at `0c2d2aa4c2` (`.moai/reports/t1762/plan-audit-verdict.md`, R1-R4; `.moai/reports/t1762/iter-3-adversarial-verdict.md`; hold record `.moai/reports/t1762/hold-and-split.md`). Card t1762 becomes split "a" and keeps only M1 (consumer config map, budget constants, template block) and M2 (ledger recorder with the bounded, non-following, regular-file-checked write open). The hook adapter, its call site, shadow byte identity, the deadline anchor and its tolerance, the M3 wall-clock measurement, and the credential wiring move to SPEC-JEV-LEDGER-002 (card t1970). `moai jev cost` moves to SPEC-JEV-LEDGER-003 (card t1971). Surviving REQ and AC ids keep their numbers; every migrated id is listed in §D.1. REQ-JEVL-017/018 and AC-JEVL-017/018 are new. R1 and R2 are repaired inside AC-JEVL-012. | manager-spec |
| 2026-10-26 | 0.2.1 | Plan-audit repair of the reduced scope (FAIL 0.81 at `954adab25e`, `.moai/reports/t1762/plan-audit-verdict.md`, D1-D9). D1: AC-JEVL-014 permits the yaml library in `internal/config/jev_consumers.go` through a forbidden-import list; `internal/jevledger` stays standard-library only. D2: `threshold` joins the REQ-JEVL-009 field list and AC-JEVL-009. D3: §D.1 gains the REQ-JEVL-009 (part) row (per-attempt rows, `model_id` never the request alias). D4: the §D.1 REQ-JEVL-008 (part) row carries the warning passthrough. D5-D9: binding constant stem in AC-JEVL-013, test-deletion residual risk, REQ-JEVL-014 in Ubiquitous positive form, platform-split pointer on ledger rows E11/E12, directory case (v) in AC-JEVL-012. REQ and AC ids and counts unchanged (12 REQ, 11 AC). | manager-spec |

## Position in the chain

This SPEC builds on the completed chain `SPEC-JEV-CORE-001` → `SPEC-JEV-OPTIN-MEASURE-001` → `SPEC-JEV-CONSUMERS-001` → `SPEC-JEV-GOAL-DIST-001`, and on `SPEC-JEV-RESOLVED-MODEL-001` (card t1768, completed), which is why a row keeps the response model and the request alias in two separate fields. It is split "a" of the "Jev 1/6" infrastructure card. Splits "b" (hook adapter, card t1970) and "c" (cost report, card t1971) land after it and consume what it ships. This SPEC ships no caller of the recorder: nothing in the shipped build writes a ledger row.

---

## §A. Problem Statement

Every Jev call today is in-memory accounting only. `Client.UsageLog()` keeps per-call input tokens and the interpreted model for the lifetime of one process (`internal/jev/jev.go`). A hook process exits in milliseconds, so a judgment made inside a hook leaves no record. The later Jev cards need two foundations before any hook can carry a judgment:

1. A durable, local, per-question record that can be written safely from a short-lived process, even when the ledger path has been replaced by a link, a pipe, or a directory.
2. A per-consumer configuration surface that a malformed entry cannot break, so that one bad consumer block never resets the rest of the `workflow` configuration.

## §B. Goal

One recorder and one config surface, both inert in the shipped build, both fail-open, and neither able to send a request.

---

## §C. Requirements (GEARS)

### §C.1 Nothing happens until asked

**REQ-JEVL-001** (Ubiquitous) The consumer-config resolver and the ledger recorder shall send no Jev request. They shall create no file or directory under `<project root>/.moai/state/jev/` except when the recorder is asked to append a row: loading or resolving the configuration, with `workflow.jev.enabled` true or false, and constructing a recorder shall create nothing on disk.

### §C.2 The config surface

**REQ-JEVL-006** (Ubiquitous) The consumer-config resolver shall accept exactly `shadow`, `advise`, `steer`, and `confirm` as values of `workflow.jev.consumers.<name>.mode`, shall resolve an absent or unrecognized value to `shadow`, and shall return `shadow` as the effective mode for every accepted value in this SPEC, returning the configured value alongside the effective one. The ledger recorder shall write a supplied configured value as `configured_mode` and a supplied effective value as `mode`.

**REQ-JEVL-007** (Ubiquitous) The consumer-config resolver shall treat `workflow.jev.consumers.<name>.threshold` as an optional number in the closed interval [0, 1], and shall resolve a value outside that interval, or of any other YAML type, to absent. The config loader shall decode each consumer entry independently, so that a malformed entry is quarantined to that consumer while the rest of the `workflow` section still loads.

**REQ-JEVL-008** (Event-detected) **When** a consumer's configuration is invalid (an unrecognized mode, an out-of-range threshold, a threshold of a non-numeric type, or a `consumers` value that is not a mapping), the consumer-config resolver shall return the condition as data, and neither the config loader nor the resolver shall write it to stdout, stderr, or a log record. The ledger recorder shall write a supplied condition as `config_warning` on the row.

**REQ-JEVL-013** (Ubiquitous) The configuration defaults shall declare the hook Jev time budget as 5 s and the safety margin as 1 s, and shall derive the deadline offset as budget minus margin (4 s at the defaults). The offset shall be one value, not parameterized by hook event or harness, and shall be derived from the two declared values rather than declared separately.

**REQ-JEVL-017** (Ubiquitous) The distributed template `workflow.yaml` shall carry `workflow.jev.enabled: false` and an empty `consumers` mapping under `workflow.jev`, and the build that embeds the template shall succeed after the edit with no tracked-file change beyond the files this SPEC names.

### §C.3 The ledger recorder

**REQ-JEVL-009** (Ubiquitous) The ledger recorder shall append one JSON Lines row per question to `<project root>/.moai/state/jev/decisions.jsonl`. Each row shall carry `model_id`, `request_model`, `rubric_version`, `question_id`, `probability` (the raw answer probability, `null` when unanswered), `mode`, `action`, and `outcome_label`, and also `ts`, `consumer`, `event`, `harness`, `availability`, `input_tokens`, `call_id`, and `threshold` (the supplied threshold value, `null` when none is supplied, with `0` kept distinct from `null`). The recorder shall write `model_id` and `request_model` from two separate inputs and shall never fill one from the other.

**REQ-JEVL-010** (Ubiquitous) The ledger recorder shall write `outcome_label` as `unlabeled` at record time, and shall write `action` as `none` on every row whose `mode` is `shadow`, whatever value the caller supplies for either field. Labelling an outcome after the fact is not part of this SPEC.

**REQ-JEVL-011** (Ubiquitous) The ledger path shall stay ignored by git (`.moai/state/` in the repository `.gitignore` and in the template `.gitignore`). The recorder shall create the directory at mode 0700 and the file at mode 0600 on POSIX systems, and shall write each row with a single append so that concurrent writers never interleave a row. The recorder shall open the ledger file without following a symbolic link at the ledger path and without blocking on the open, shall write a row only when the opened handle is a regular file, and shall never open the ledger for reading.

**REQ-JEVL-012** (Event-detected) **When** the ledger directory or file cannot be created or written, or the opened ledger handle is not a regular file (a symbolic link, a FIFO with or without a reader, a device, or a directory at the ledger path), the recorder shall drop the row, leave the object at that path and any link target unchanged, deliver no byte to that object, return to its caller without blocking and without panicking, and report the drop only as a returned value, writing nothing to stdout or stderr.

**REQ-JEVL-018** (Event-detected) **When** the ledger file already holds one or more malformed lines, each terminated by a newline, the recorder shall append the new row as one complete line after them, leave every existing byte unchanged, and not drop the row on account of them.

### §C.4 Authority and the boundary

**REQ-JEVL-014** (Ubiquitous) The ledger recorder and the consumer-config resolver shall limit their effects to returning resolved configuration values and appending ledger rows: every Jev answer shall stay outside the inputs of permission decisions, completion verdicts, merge approvals, queue mutations, operator gates, and hook output, and every endpoint, credential read, and network call shall stay outside both components.

---

## §D. Exclusions

### Out of Scope — migrated to SPEC-JEV-LEDGER-002 (card t1970, lands after t1762)

- The hook adapter and its call site in the hook dispatcher, the compiled consumer registry, and the `Ask` seam.
- Shadow byte identity of hook output (the 24-case matrix), the run-after-output ordering, and hook-level silence on invalid configuration.
- The deadline anchored at hook process start, the return tolerance T, the timeout, intermediate-elapsed, exhausted, and exit-2 cases.
- The M3 wrapper-start-to-exit wall-clock measurement, its sample-validity precondition, and its S = 0.5 s gate (audit finding R3).
- Credential wiring of the production client and the credential clause of the boundary check (audit finding R4).
- The hook-level ledger-failure test (the adapter-driven twin of AC-JEVL-012).

### Out of Scope — migrated to SPEC-JEV-LEDGER-003 (card t1971, lands after t1762)

- `moai jev cost`: aggregation, de-duplication by `call_id`, undefined-when-zero, malformed-line skip with `skipped: N`, and the report labels.
- The "estimated input tokens" unit label. It only matters where a cost figure is printed, and nothing here prints one. The recorder stores `input_tokens` as the caller supplies it.

### Out of Scope — the internal/jev client (card t1968)

- The `internal/jev` client cache and retry belong to card t1968 in another lane. This SPEC does not edit `internal/jev/jev.go` or any other file in `internal/jev`. The recorder lives in its own package and imports nothing from `internal/jev` in production code.

### Out of Scope — real consumers and non-shadow behaviour

- Any compiled consumer, question set, or rubric. Jev 2, 3, and 5 add those.
- Executing `advise`, `steer`, or `confirm`. The resolver accepts the values and returns `shadow` as the effective mode. Acting on `threshold` is likewise out of scope.

### Out of Scope — ledger maintenance

- Rewriting `outcome_label` after the fact, ledger rotation, retention, or pruning.
- Repairing an unterminated final fragment in the ledger. The recorder never reads the ledger (REQ-JEVL-011), so it cannot see one; a row appended after such a fragment joins it on one line (plan.md §H).
- Protecting against a symbolic link at an intermediate directory (`.moai/state/jev` as a link). The no-follow open covers the final path element only (plan.md §H).

### §D.1 Migrated requirement and criterion ids

Nothing is silently dropped. Each id below left the active set of this SPEC at 0.2.0. "Part" means the id survives here in a narrowed form and only the named part moved.

| Destination | Id | What moved |
|---|---|---|
| -002 | REQ-JEVL-001 (part) | Zero requests and zero rows on `pre-tool` / `stop` / `subagent-stop` dispatch with the gate off, under both harnesses. |
| -002 | REQ-JEVL-002 | Empty compiled consumer registry; configured name without a registration ignored. |
| -002 | REQ-JEVL-003 | Offering hook events to consumers, one batched request per consumer, both harnesses. |
| -002 | REQ-JEVL-004 | Adapter runs after output and exit decision are fixed. |
| -002 | REQ-JEVL-005 | Shadow byte identity of stdout, stderr, exit code. |
| -002 | REQ-JEVL-006 (part) | The hook adapter branches only on the effective mode. |
| -002 | REQ-JEVL-007 (part) | The hook adapter selects no action on `threshold`. |
| -002 | REQ-JEVL-008 (part) | Invalid configuration emits nothing to the hook's stdout or stderr, and the adapter passes each consumer's returned warning as `config_warning` on every row of that consumer. |
| -002 | REQ-JEVL-009 (part) | One row per question per attempted call, including unanswered and timed-out attempts; the adapter supplies the response's interpreted model id, never the request alias, as model_id. |
| -002 | REQ-JEVL-012 (part) | Ledger failure leaves the hook's exit code unchanged. |
| -002 | REQ-JEVL-013 (part) | Deadline anchored at process start, timeout row, skip when exhausted, return tolerance. |
| -002 / -003 | REQ-JEVL-014 (part) | The boundary for the hook adapter (-002) and the cost report (-003); every request through `jev.Client.Ask` (-002). |
| -003 | REQ-JEVL-015 | The `moai jev cost` report. |
| -003 | REQ-JEVL-016 | Absent, empty, and malformed ledgers in the cost report. |
| -002 | AC-JEVL-001 (part) | Hook-level builder, `Ask`, and transport counts, including the builder-call mutant whose subject is the adapter. |
| -002 | AC-JEVL-002 | Shipped registry empty. |
| -002 | AC-JEVL-003 | Both harnesses, three events, one batched request. |
| -002 | AC-JEVL-004 | 24-case shadow byte identity. |
| -002 | AC-JEVL-005 | Adapter runs after output. |
| -002 | AC-JEVL-006 (part) | Part (ii), the call-site effective-mode check. |
| -002 | AC-JEVL-008 | Hook-level silence on invalid configuration. REQ-JEVL-008 is now covered here by AC-JEVL-007 and AC-JEVL-009. |
| -002 | AC-JEVL-012 (part) | `TestHookJev_LedgerFailureSilent` and the 200 ms + T bound. |
| -002 | AC-JEVL-013 (part) | Cases (a)-(d), the deadline equality, T = 50 ms, and the per-event and per-harness assertion. |
| -002 / -003 | AC-JEVL-014 (part) | By-name allow-list entries for `hook_jev.go` (-002) and `jev_cost.go` (-003), and the credential clause (-002). |
| -003 | AC-JEVL-015 | Cost report aggregates. |
| -003 | AC-JEVL-016 | Absent and malformed ledgers in the report. Its template partial moved to AC-JEVL-017 here. |

---

## §E. Open Questions

Every open question carries a row in `decision-index.md`. Q1, Q4, and Q7 bind this SPEC. Q2's measurement half, Q3, and Q5 migrated with their subjects. Q6 was settled by the split ruling.
